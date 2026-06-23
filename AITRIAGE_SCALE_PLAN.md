# AITriage — Enterprise-Scale AI Triage: План работ и прогресс

> Единый источник правды для рефакторинга AI-триажа под enterprise PROD-нагрузку.
> Здесь хранятся: контекст, исследование, целевая архитектура, разбивка задач,
> чеклист и **живой лог прогресса**. Обновляется по мере работы.

- **Статус:** � Фазы 0-6 завершены и зелёные. Осталась Фаза 7 (Docker/SHA — внешние шаги)
- **Владелец задачи:** Cascade + @cybertortuga
- **Создан:** 2026-06-22
- **Последнее обновление:** 2026-06-22

---

## 1. Контекст и проблема

### 1.1. P0-баг (что упало в CI)
Реальный прогон `accrual-ai` PR #2 упал с ошибкой:

```
LLM analysis failed: threat-model analysis failed:
threat-model response classified 148 of 218 findings
```

Корень: в `internal/agent/graph/orchestrator.go` функция `buildThreatModel`
отправляла в LLM максимум 150 находок (`findingsToSend[:150]`), но:
1. сообщала модели полное число (218);
2. требовала диспозицию (TP/FP/NR) для всех 218.

Модель вернула 148 → строгая валидация `validateFindingDispositions` упала.
Это **гарантированный провал для любого репозитория с >150 находок**.

### 1.2. Проблема масштаба
При 3000 находок наивный батчинг = ~20+ последовательных LLM-вызовов с retry:
дорого, медленно, упирается в rate limits и per-request token cap, и тем чаще
ловит «модель вернула не все».

### 1.3. Цель
Сделать AI-триаж **enterprise PROD-уровня**: устойчивый, дешёвый, быстрый,
аудируемый, не теряющий находки и не помечающий непроверенное как безопасное.

---

## 2. Исследование (на чём основаны решения)

| Источник | Ключевой вывод | Применяем |
|---|---|---|
| **Datadog Bits AI** ([blog](https://www.datadoghq.com/blog/using-llms-to-filter-out-false-positives/)) | LLM запускается только на curated подмножестве CWE (OWASP Top Ten/Benchmark); результат — confidence-уровни, а не бинарь; есть feedback-loop | Слой 3 (gating), confidence в диспозициях |
| **ZeroFalse** ([arXiv 2510.02534](https://arxiv.org/html/2510.02534)) | Canonicalization находок (rule/CWE id + локация + dataflow trace); **каждая находка оценивается независимо, строго 1:1** с audit trail; точный срез контекста, а не «весь файл» | Слой 1 (canonical), 1:1 mapping, аудит-лог |
| **GitHub Code Scanning / SARIF** ([docs](https://docs.github.com/en/enterprise-cloud@latest/code-security/reference/code-scanning/sarif-files/sarif-support)) | Дедуп через `partialFingerprints` (`primaryLocationLineHash`) — стабильный отпечаток, чтобы одна проблема не дублировалась между прогонами | Слой 1 (fingerprint), Слой 2 (cache по отпечатку) |

**Важный вывод:** объединять «один rule в разных файлах» в одну группу и
пропагировать диспозицию — **НЕ best practice** (эксплуатируемость зависит от
потока данных в конкретном месте). Поэтому дедупим только **идентичные** находки
(одинаковый fingerprint), а не «похожие по правилу».

---

## 3. Целевая архитектура (5 слоёв поверх батчинга)

Батчинг/retry/NR — это транспортный слой (safety net). Поверх него:

```
findings (raw, напр. 3000)
   │
   ▼  Слой 1: Canonicalize + Fingerprint + Dedup (детерминированно, без ИИ)
   │         → схлопывание идентичных → напр. 800 уникальных
   ▼  Слой 2: Cache по fingerprint (между прогонами CI)
   │         → попадание в кэш → 0 LLM-вызовов на повтор → напр. 120 новых
   ▼  Слой 3: Category/Severity gating (как Datadog)
   │         → LLM только на значимые CWE + HIGH/CRITICAL → напр. 60 в ИИ
   │         → остальное: детерминированный disposition (без FP-индульгенции)
   ▼  Слой 4: Structured-output классификация (строгая JSON-схема, 1:1)
   │         → заставляет вернуть запись на КАЖДУЮ находку батча (лечит P0)
   ▼  Слой 5: Bounded concurrency + backoff + budget
             → параллельные воркеры, rate-limit, лимит; остаток → NR
   │
   ▼  dispositions (ровно по 1 на исходную находку, через fingerprint-проекцию)
```

### Инварианты (не нарушать)
- Каждая исходная находка получает ровно один disposition: TP / FP / NR.
- Непроверенное → **Needs Manual Review**, НИКОГДА не False Positive.
- NR штрафует Health Check score (не даёт PR пройти зелёным).
- Transport/provider failure → pipeline падает (не маскируем сбой LLM).
- Никакого silent-discard находок.
- Полный audit trail: для каждой находки — источник диспозиции (LLM / cache /
  deterministic / NR-fallback) и rationale.

### 3.1. SecureCoder — методологический каркас (НЕ нарушать)
SecureCoder — это **единая личность, методология и ruleset**, вшитые во ВСЕ
LLM-промпты (`prompts.SecureCoderFramework` + `SecureCodingGuidelines`):
ThreatModel, PoC, Report, FixSpec. Это не отдельный шаг, а основа всего пайплайна.
7-шаговая методология: repo → threat model → оценка КАЖДОЙ находки против кода и
модели → TP/FP/NR → PoC-трейс для TP → ремедиация → CS-XXX-NNN ID.

Как слои масштабирования обязаны сохранить SecureCoder:
- **Threat model строится ОДИН раз** (из repo-context), затем находки
  классифицируются ПРОТИВ неё. Не пересоздавать модель в каждом батче.
  Threat model передаётся в каждый батч/воркер классификации как контекст.
- **«Каждая находка» остаётся святой.** Gating (Слой 3) не выкидывает находки
  мимо методологии: не-gated находки проходят через **детерминированный
  SecureCoder-ruleset** (rule → MUST/MUST NOT), при недостатке контекста → NR.
  Никогда не молчаливый FP.
- **PoC Verification — тоже SecureCoder** и подчиняется тем же слоям (dedup,
  cache, concurrency, budget). Текущий cap 75 TP с silent-drop — устранить.
- **Report/FixSpec** потребляют сгруппированные диспозиции и сохраняют
  CS-XXX-NNN ID и FP-rationale (audit trail), несмотря на дедуп.
- **Ruleset — единый источник.** Детерминированная классификация (gating) и
  LLM-классификация ссылаются на один и тот же `SecureCodingGuidelines`.

---

## 4. Затрагиваемые файлы

| Файл | Изменение |
|---|---|
| `internal/agent/graph/orchestrator.go` | ✅ батчинг+retry+NR (готово); интеграция слоёв 1-5 |
| `internal/agent/graph/state.go` | новые поля: `Fingerprint`, `Confidence`, `DispositionSource` в `FindingDisposition`; счётчики |
| `internal/agent/graph/fingerprint.go` (новый) | canonical fingerprint + dedup |
| `internal/agent/graph/cache.go` (новый) | кэш вердиктов по fingerprint (файловый, для CI) |
| `internal/agent/graph/gating.go` (новый) | category/severity gating, deterministic disposition |
| `internal/agent/graph/classify.go` (новый или вынос из orchestrator) | threat-model-once + structured-output классификация + concurrency |
| `internal/agent/graph/orchestrator.go` (`runPoCVerification`) | PoC scaling (фаза 5b): убрать cap 75, dedup/cache/concurrency |
| `internal/agent/prompts/templates.go` | строгая JSON-схема классификации + сохранить SecureCoder ruleset (MUST/MUST NOT) |
| `internal/agent/graph/gating.go` → SecureCoder ruleset | детерминированная классификация по единому `SecureCodingGuidelines` |
| `internal/server/pipeline_handler.go` | веб-режим переводим на общий `ClassifyFindings` |
| `internal/report/healthcheck/healthcheck.go` | сверить дедуп-ключ с fingerprint |
| тесты `*_test.go` | покрытие всех слоёв |

---

## 5. Разбивка задач (фазы, подзадачи, чеклист)

### Фаза 0 — Стабилизация P0 (safety net) — ✅ ГОТОВО
- [x] 0.1 Батчинг по ВСЕМ находкам (не дропать >150)
- [x] 0.2 Bounded retry для пропущенных индексов
- [x] 0.3 NR-fallback для непроклассифицированного (не FP)
- [x] 0.4 `ClassifyFindings` как общая функция; transport-fail валит pipeline
- [x] 0.5 Перевести веб-обработчик `runWebThreatModel` на `ClassifyFindings`
- [x] 0.6 Тесты Фазы 0: >150, partial, dup/out-of-range, retry/NR, no-FP, transport-fail
- [x] 0.7 `go test -p 1 ./...` зелёный

### Фаза 1 — Canonicalize + Fingerprint + Dedup — ✅ ГОТОВО (`fingerprint.go`)
- [x] 1.1 `Fingerprint(f EnrichedFinding) string` = sha256(ruleId+type+normPath+line+message)
- [x] 1.2 `normalizePath` (стрип /src/, ./, ведущий /, lowercase, backslashes)
- [x] 1.3 `dedupFindings(findings) -> (unique, groups [][]int)` — без drop
- [x] 1.4 `projectDispositions` — проекция на все дубли (свой index/VulnID/fp)
- [x] 1.5 Дедуп только идентичных (location-sensitive), не объединяет rule×файлы
- [x] 1.6 Тесты дедупа (`fingerprint_test.go`)

### Фаза 2 — Cache вердиктов по fingerprint — ✅ ГОТОВО (`cache.go`)
- [x] 2.1 `verdictCache` (Get/Set/Save по fingerprint)
- [x] 2.2 Файловая реализация (`AITRIAGE_CACHE_DIR/triage_cache.json`), версия схемы
- [x] 2.3 Инвалидация: ключ = model|vN|fingerprint (смена модели/схемы → miss)
- [x] 2.4 Подключено в `classifyUnique`: cache-hit → пропуск LLM (off без env)
- [x] 2.5 Тесты cache hit/miss/invalidation (`cache_test.go`)

### Фаза 3 — Category/Severity gating (SecureCoder-aware) — ✅ ГОТОВО (`gating.go`)
- [x] 3.1 `llmSeverities` (CRITICAL/HIGH) — основа gating, расширяемо
- [x] 3.2 `shouldTriageWithLLM(f)` (gating off → всё в LLM; через AITRIAGE_GATING=on)
- [x] 3.3 `deterministicDisposition` для не-gated → NR (никогда не FP), source=deterministic
- [x] 3.4 Gating OFF по умолчанию (= «каждая находка через LLM»)
- [x] 3.5 Тесты gating (gated-out → NR, не FP) (`gating_test.go`)

### Фаза 4 — SecureCoder threat-model-once + structured classification (1:1) — ✅ ГОТОВО (`classify.go`)
- [x] 4.1 Threat model строится ОДИН раз из sample, затем `classifyBatchLLM(tmSummary, batch)`
- [x] 4.2 Жёсткая JSON-схема `ClassificationSystemPrompt` (запись на КАЖДУЮ)
- [x] 4.3 Robust-парсер + retry пропущенных + NR-fallback
- [x] 4.4 Провайдер-агностично (строгий JSON в промпте; native JSON-mode — позже)
- [x] 4.5 `confidence` и `disposition_source` в каждой записи
- [x] 4.6 `ClassificationSystemPrompt` = SecureCoderFramework + ruleset (MUST/MUST NOT)
- [x] 4.7 Тесты парсинга + TM-once (`classify_findings_test.go`)

### Фаза 5 — Concurrency + backoff + budget — ✅ ГОТОВО (`classify.go`)
- [x] 5.1 Пул воркеров (`AITRIAGE_CONCURRENCY`, default 4) через semaphore
- [x] 5.2 Backoff на rate-limit — через существующий `llm.RetryClient` (оборачивает client)
- [x] 5.3 Бюджет `AITRIAGE_LLM_BUDGET` (default -1); остаток → NR с budgetRationale
- [x] 5.4 Результаты собираются по unique-index → детерминированный порядок
- [x] 5.5 Тесты concurrency(=1)/budget (`classify_findings_test.go`)

### Фаза 5b — PoC Verification scaling (SecureCoder шаг 5) — ✅ ГОТОВО (`poc.go`)
- [x] 5b.1 Убран silent-drop cap 75: батчинг по ВСЕМ True Positive
- [x] 5b.2 Dedup PoC по fingerprint (один трейс на идентичную уязвимость)
- [x] 5b.3 Батчинг (`pocBatchSize=25`) + concurrency + budget `AITRIAGE_POC_BUDGET`
- [x] 5b.4 Оверфлоу/непокрытые TP → conclusion "Needs Manual Review" (не drop)
- [x] 5b.5 Тесты PoC scaling (`poc_test.go`)

### Фаза 6 — Аудит, отчётность, интеграция (SecureCoder-faithful) — ✅ ГОТОВО
- [x] 6.1 `DispositionSource` в отчёте (блок «Disposition sources»)
- [x] 6.2 Лог метрик `logTriageMetrics` (findings→unique, deduped, по источникам)
- [x] 6.3 Report потребляет спроецированные диспозиции, CS-ID/FP-rationale сохранены
- [x] 6.4 Веб-режим и CLI используют один `ClassifyFindings`
- [x] 6.5 Обновить документацию/README (env-переменные ниже в этом файле)

### Фаза 7 — Release — 🟡 (локально готово; внешние шаги требуют доступа)
- [x] 7.1 `go test -p 1 ./...` + `go build ./...` — зелёные
- [ ] 7.2 Собрать новый Docker image, новый digest
- [ ] 7.3 Обновить immutable SHA в `security-workflows`
- [ ] 7.4 Обновить pinned SHA в caller `accrual-ai`
- [ ] 7.5 Перезапустить реальный workflow, убедиться что зелёный

---

## 6. План тестирования

Команда прогона (из ТЗ):
```bash
GOCACHE=/private/tmp/aitriage-go-build-cache go test -p 1 ./...
```

Обязательные сценарии:
- >150 находок проходят без падения;
- partial LLM response → retry → NR fallback;
- duplicate / out-of-range индексы игнорируются;
- непроклассифицированное → NR, НИКОГДА не FP;
- transport error → pipeline падает;
- дедуп: идентичные находки схлопываются, диспозиция проецируется;
- cache hit пропускает LLM; invalidation при смене модели;
- gating: LOW/INFO не идут в LLM, FP не выдаётся им автоматически;
- budget: превышение → NR.

---

## 7. Риски и открытые вопросы

- **FP-проекция при дедупе:** проецируем диспозицию только на идентичные
  fingerprint, не на «похожие». Подтвердить нормализацию путей.
- **Cache корректность:** ключ обязан включать модель + версию схемы + версию
  правил, иначе риск устаревших вердиктов. (По умолчанию инвалидация при апгрейде.)
- **Gating и комплаенс:** какие CWE-классы считаем «значимыми»? Нужен ли режим
  «всё через ИИ» для строгого профиля? → согласовать список.
- **Function-calling:** интерфейс `llm.Client` сейчас только `Chat(string)`.
  Решение: строгая JSON-схема в промпте + robust-парсер (провайдер-агностично);
  нативный JSON-mode — опционально позже.

---

## 8. Лог прогресса (живой)

- **2026-06-22:** Диагностирован P0 (`148 of 218`). Реализован батчинг+retry+NR
  в `ClassifyFindings`/`buildThreatModel` (`orchestrator.go`). Проведено
  исследование (Datadog, ZeroFalse, SARIF). Составлена 5-слойная архитектура и
  этот план.
- **2026-06-22 (Фаза 0 ✅):** Веб-обработчик `runWebThreatModel` переведён на
  общий `graph.ClassifyFindings` (убраны silent-drop >150 и опасный
  default-to-TP). Добавлен `classify_findings_test.go` (9 тестов: >150 батчинг,
  partial→NR, dup/out-of-range, unsupported→NR, malformed-retry толерантно,
  malformed-first-pass фатально, transport-fail в pass1 и retry). Полный прогон
  `go test -p 1 ./...` зелёный.
- **2026-06-22 (SecureCoder в план):** Добавлена секция 3.1 (SecureCoder —
  методологический каркас). Учтены 3 риска: (1) gating должен
  применять детерминированный ruleset (фаза 3), (2) threat model строится
  ОДИН раз и переиспользуется (фаза 4.1), (3) PoC — тоже SecureCoder с тем
  же silent-drop багом (новая фаза 5b). Дальше: Фаза 1 (fingerprint + dedup).
- **2026-06-22 (Фазы 1-6 ✅):** Реализованы все слои масштабирования:
  - `fingerprint.go` — fingerprint + dedup идентичных + проекция диспозиций.
  - `cache.go` — verdict cache по fingerprint (off без `AITRIAGE_CACHE_DIR`),
    инвалидация по model|schema.
  - `gating.go` — severity gating (off по умолчанию) + детерминированный NR.
  - `classify.go` — новый `ClassifyFindings`: dedup → cache → gating →
    threat-model-once → structured classification → concurrency + budget; остаток
    → NR (никогда FP). `templates.go` получил `ClassificationSystemPrompt`
    (SecureCoder ruleset + строгая JSON-схема 1:1).
  - `poc.go` — Фаза 5b: убран cap 75, dedup + батчи + concurrency + budget,
    оверфлоу → "Needs Manual Review".
  - Отчёт: блок «Disposition sources» (audit trail). Метрики в stderr.
  Тесты: `fingerprint_test.go`, `gating_test.go`, `cache_test.go`,
  `poc_test.go`, переписан `classify_findings_test.go` (content-aware mock,
  concurrency=1). **`go build ./...` и `go test -p 1 ./...` — зелёные.**
  Осталась только Фаза 7 (Docker image + обновление pinned SHA в
  `security-workflows` и `accrual-ai` — внешние репозитории/доступ).

---

## 9. Конфигурация (ENV)

Все слои масштабирования управляются переменными окружения и **по умолчанию
выключены/безопасны** (поведение не меняется без явного включения):

| Переменная | Default | Назначение |
|---|---|---|
| `AITRIAGE_CONCURRENCY` | `4` | Число параллельных воркеров классификации/PoC |
| `AITRIAGE_LLM_BUDGET` | `-1` (∞) | Макс. число уникальных находок в LLM; остаток → NR |
| `AITRIAGE_POC_BUDGET` | `-1` (∞) | Макс. число уникальных TP в PoC; остаток → NR |
| `AITRIAGE_GATING` | off | `on` → LLM только для CRITICAL/HIGH; остальное → детерминированный NR |
| `AITRIAGE_CACHE_DIR` | unset | Каталог для verdict-кэша; unset → кэш выключен |
| `AITRIAGE_MODEL` | "" | Метка модели в ключе кэша (для инвалидации при апгрейде) |

**Пример для большого репо (3000 находок) в CI:**
```bash
export AITRIAGE_CONCURRENCY=8
export AITRIAGE_GATING=on
export AITRIAGE_CACHE_DIR=.aitriage-cache
export AITRIAGE_MODEL=gemini-2.0-flash
export AITRIAGE_LLM_BUDGET=400
export AITRIAGE_POC_BUDGET=100
```

## 10. Новые файлы и точки входа

| Файл | Роль |
|---|---|
| `internal/agent/graph/fingerprint.go` | Слой 1: `Fingerprint`, `dedupFindings`, `projectDispositions` |
| `internal/agent/graph/cache.go` | Слой 2: `verdictCache` |
| `internal/agent/graph/gating.go` | Слой 3: `gatingConfig`, `deterministicDisposition` |
| `internal/agent/graph/classify.go` | Слой 4+5: `ClassifyFindings`, `classifyUnique`, concurrency/budget |
| `internal/agent/graph/poc.go` | Фаза 5b: `verifyPoCs` |
| `internal/agent/prompts/templates.go` | `ClassificationSystemPrompt` / `ClassificationUserPromptTemplate` |
