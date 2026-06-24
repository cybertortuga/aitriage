# AITriage — Enterprise-Scale AI Triage: План работ и прогресс

> Единый источник правды для рефакторинга AI-триажа под enterprise PROD-нагрузку.
> Здесь хранятся: контекст, исследование, целевая архитектура, разбивка задач,
> чеклист и **живой лог прогресса**. Обновляется по мере работы.

- **Статус:** ✅ Фазы 0-7 завершены. Production image, Action pin и binary release опубликованы; существующий caller-пин в `accrual-ai` отсутствует и не изменялся.
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

### Фаза 7 — Release — ✅ ГОТОВО
- [x] 7.1 `go test -p 1 ./...` + `go build ./...` — зелёные
- [x] 7.2 Собрать новый Docker image, новый digest (`sha256:be007aa4ed9b6e8ac719818a96974d340de468593c3f269134386e87925b9088`)
- [x] 7.3 Обновить immutable SHA в `security-workflows` (`cc9c04667112f5b9dc395cc6d71e059954ee86ba`)
- [x] 7.4 Проверить caller `accrual-ai`: current `main` не содержит AITriage/reusable-workflow pin, поэтому обновлять нечего
- [x] 7.5 Проверить релизные workflows: CI, GHCR и Release — зелёные. `security-workflows` не запускался: в нём нет секрета и он имеет только `workflow_call`

### Фаза 8 — Канонический triage artifact — ✅ ГОТОВО

**Цель:** после *завершённого* AI triage сохранять независимый от policy gate
машиночитаемый audit trail всех исходных находок. `report.md`, `fixspec.md` и
`summary.md` остаются производными представлениями; `triage-findings.json` —
канонический экспорт уже валидированного `AgentState`.

**Контракт `triage-findings.json` v1:**
- top-level `schema_version`, `triage_status: "complete"`, итоговый
  `health_check` и массив `findings`;
- каждая запись массива содержит исходную `finding` и её ровно одну
  `disposition` (индекс, ID, TP/FP/NR, rationale, confidence, source,
  fingerprint);
- порядок — исходный порядок находок; дубликаты не скрываются, поскольку
  экспорт предназначен для аудита всех записей, а не для отображения;
- exporter отказывается писать неполный/двусмысленный набор (missing,
  duplicate, out-of-range или unsupported disposition).

- [x] 8.1 Добавить версионированную модель/построитель artifact в `graph`,
  использующий только уже валидированные `EnrichedFindings` и
  `FindingDispositions`.
- [x] 8.2 Добавить CLI-флаг `--triage-out`; записывать JSON после трёх
  Markdown-файлов и **до** GitHub summary и возврата `ErrPolicyViolation`.
- [x] 8.3 Покрыть контракт тестами: все TP/FP/NR, исходный порядок,
  fingerprints/rationale и отказ для неполной либо двусмысленной проекции.
- [x] 8.4 Обновить reusable workflow: передать `--triage-out
  triage-findings.json`, загрузить вместе `triage-findings.json`, `report.md`,
  `fixspec.md`, `summary.md` через `if: always()` до отдельного enforcement
  шага. Статический `aitriage.sarif` остаётся отдельным artifact.
- [x] 8.5 Выпустить новый immutable Action image/commit, repin reusable
  workflow и caller.
- [x] 8.6 Запустить **один** caller workflow и проверить скачанный artifact:
  количество JSON-записей равно числу triaged findings, в том числе
  suppressed FP; policy может быть FAILED, artifact обязан быть доступен.

### Фаза 9 — Evidence-bound LLM suppression — ✅ ВЫПОЛНЕНО

**Проблема:** транспортный контракт TP/FP/NR и 1:1 mapping надёжен, но
слабая модель может выдать корректный index с rationale от другой находки.
Нельзя позволять одному свободному тексту модели снять finding с policy gate.

**Контракт:**
- модель возвращает `finding_index`, `finding_id`, `fingerprint`, disposition
  и structured `evidence`;
- identity должна совпасть с finding, отправленным в конкретный batch;
- LLM-FP принимается только с deterministic evidence:
  `test_only` (действительно test path) или `code_mitigation` (существующий
  file/line и literal observed evidence в source);
- несовпадение identity, unsupported evidence или непроверяемый FP →
  `Needs Manual Review`, никогда не suppressed FP;
- raw response, batch-local→deduplicated-global mapping и принятые/отклонённые ответы
  сохраняются в `triage-findings.json` как `classification_audit`. Основной
  четырёхфайловый artifact contract не меняется.

- [x] 9.1 Добавить версионированный structured-output contract с identity и
  evidence в classification prompt/parser.
- [x] 9.2 Реализовать deterministic validator и NR fallback для invalid FP.
- [x] 9.3 Сохранять classification audit в canonical JSON artifact.
- [x] 9.4 Тесты: 218 batch mapping, wrong identity, invalid FP, valid
  test-only/code-mitigation FP, raw-audit persistence.
- [x] 9.5 Выпустить GHCR image, repin Action/reusable workflow/caller и
  выполнить один Gemini caller run.

**Rollout (2026-06-23):** image
`sha256:960a67e1178993c69f4423a85692ef36f36eb4de0e660925310fd52c80cf7064`
проверен и закреплён в Action commit `6138abb`; reusable workflow —
`2d5df9a`; caller run `28015119014` с Gemini завершил AI triage и загрузил
все четыре файла до ожидаемого FAILED policy gate. `triage-findings.json`
содержит 218 findings, шесть raw-response audit entries и 40 отклонённых
неподтверждённых ответов False Positive.

### Фаза 10 — Scanner signal quality — ✅ ВЫПОЛНЕНО

**Проблема:** live scan `accrual-ai` показал deterministic ложные findings до
LLM triage: extensionless `Dockerfile` не попадал под filename-rule, наличие
неиспользуемой npm dependency включало Express rules, а `not_contains:*` в
YAML-правиле не исполнялся.

- [x] 10.1 Матчить extensionless filenames (в частности `Dockerfile`) в
  project и file rules.
- [x] 10.2 Включать Express rules только при runtime import/require, а не по
  одной dependency в `package.json`.
- [x] 10.3 Реализовать `not_contains:<token>` и покрыть его LLM token-limit
  сценарием.
- [x] 10.4 Прогнать unit suite и повторить deterministic scan `accrual-ai`.

**Validation (2026-06-24):** полный `go test -p 1 ./...`, `go build ./...` и
`go vet ./...` прошли. Повторный deterministic scan `accrual-ai` снизился с
17 до 8 active findings: остались rate limiting, security logging, LLM timeout
и UI-specific findings; ложные Dockerfile/Express/CSRF/ORM/token-limit findings
не генерируются.

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
- канонический artifact: все исходные findings (TP/FP/NR) экспортируются 1:1;
  missing/duplicate/out-of-range disposition не создаёт misleading JSON;
  policy `FAILED` не мешает его upload после завершённого triage.
- evidence-bound suppression: LLM не может suppress finding без совпадающих
  identity и deterministic proof; любой непроверяемый ответ становится NR.

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

- **2026-06-23 (Фаза 7.1–7.2 ✅):** полный локальный набор (`go test -p 1`,
  `go build`, `go vet`) и Docker smoke-test зелёные. Изменение прошло PR #15
  и основной CI. Production workflow `Publish Docker Image (GHCR)` собрал
  linux/amd64 image и опубликовал immutable digest
  `sha256:be007aa4ed9b6e8ac719818a96974d340de468593c3f269134386e87925b9088`.
- **2026-06-23 (Фаза 7.3–7.5 ✅):** Action pin обновлён через PR #16,
  затем опубликован release `v1.5.4` с Linux/macOS/Windows binaries и
  `checksums.txt`. Reusable workflow `dodo-ai-platform/security-workflows`
  обновлён до immutable Action commit
  `d777cd590184e0ab1c673e9eeb5106f187c09654` (PR #4). В
  `dodo-ai-platform/accrual-ai` нет существующей ссылки на AITriage или этот
  reusable workflow; новый gate намеренно не добавлялся. Внешний workflow не
  запускался, поскольку в `security-workflows` нет нужного секрета.
- **2026-06-23 (Фаза 8 ✅):** В PR #18 добавлен versioned
  `triage-findings.json` (`schema_version: 1`) и CLI-флаг `--triage-out`.
  Exporter строит 1:1 inventory из валидированных `EnrichedFindings` и
  `FindingDispositions` до GitHub summary/policy gate и отказывается писать
  неполный или двусмысленный результат. GHCR image
  `sha256:fd29c6a8750cef46a94964b8c29ee092d69c6ff35f8fd01e30de91a40499c75c`
  закреплён в Action commit `947f278088925b46a31a5814c1638b88c6232a18`
  (PR #19). Reusable workflow обновлён через
  `dodo-ai-platform/security-workflows` PR #6 (`8341a4a27525a6ad1e9c7bfb3100be19e7f2fd04`):
  upload всех четырёх AI-файлов выполняется `if: always()` до отдельного
  gate step. Caller `accrual-ai` repin `7b1d12a` создал один PR run
  [#28011270090](https://github.com/dodo-ai-platform/accrual-ai/actions/runs/28011270090):
  AI triage и upload прошли, затем policy gate ожидаемо failed. Скачанный
  JSON: 218/218 записей, 218 уникальных `finding_index`, 45 TP, 173 FP,
  0 NR; bundle также содержит `summary.md`, `report.md`, `fixspec.md`.

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
