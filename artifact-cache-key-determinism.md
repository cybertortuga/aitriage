# ARTIFACT CACHE KEY DETERMINISM

Дата: 2026-07-03
Ветка: `codex/artifact-reuse-cache`
Контекст: продолжение `artifact-reuse-cache.md` (Task 7 blocker) и `artifact-reuse-cache-handoff.md`.
Назначение: локальный план работ. Не коммитить в public repo без отдельного решения.

## Проблема (root cause, доказан локально)

На v3 hit run (`28654594215`) verdict cache сработал 212/212, GitHub cache восстановился по exact key, но artifact bundle дал `key_miss`, и PoC/report/fixspec пошли в LLM заново.

Причина: `classifyVulnCode` в `internal/agent/graph/orchestrator.go` выбирает код класса уязвимости циклом `for key, code := range prompts.VulnClassCodes`. Порядок итерации Go map рандомизирован при каждом проходе, а функция возвращает первый совпавший ключ. Если message матчится на несколько разных кодов, код выбирается случайно в каждом run.

Доказательства:

- В warm-артефакте `/tmp/aitriage-28653128012/triage-findings.json` **6 из 213** findings матчатся на >1 код (например PyJWT «Authentication bypass … JSON Web Tokens … secret» → `AUTH`/`JWT`/`SECRETS`; Vite «Information disclosure via path traversal» → `INFO`/`PATH`).
- Локальный тест на этих сообщениях: 200 вызовов `classifyVulnCode` дают 3/2/2 разных кода соответственно.

Каскад: случайный код → другой `VulnID` у finding **и сдвиг нумерации `CS-XXX-NNN` у всех последующих findings того же класса** → `FindingID` входит в `hashArtifactDisposition` → другие `DispositionHashes` → другой artifact key → `key_miss`.

Почему прошлый fix не помог: `08a4bf4` стабилизировал только *порядок* findings перед `assignVulnIDs`. Рандом сидит не в порядке, а в выборе кода внутри `classifyVulnCode`.

Закрытые гипотезы из handoff:

- Гипотеза 1 (файл не попал в кэш) — исключена: save-условие workflow требует `hashFiles('.aitriage-cache/artifact_bundle_cache.json') != ''`, и cache был сохранён.
- Гипотезы 4/5 (env/namespace mismatch) — исключены: verdict cache ключуется `namespace|fingerprint` и попал 212/212, значит namespace и fingerprints идентичны warm/hit.
- Гипотеза 2 подтверждена в уточнённом виде: ключ отличался из-за `FindingID` (CS-* ID), а не из-за порядка или fingerprints.
- «31 vs 30 unique TPs» в логах — красная селёдка: строка печатает `len(pocResults)` (число результатов PoC, проседает от refusal/parse failure), а не число входных TP.

## Ограничения

- Не запускать GitHub Actions без явной команды пользователя.
- Не писать в Kaiten без явного подтверждения.
- Не коммитить/пушить без явной команды.
- Только локальная разработка и локальные тесты.

## Задача 1: Детерминированный classifyVulnCode

### Описание

Заменить итерацию по map на детерминированный порядок ключей: длинные ключи первыми (более специфичный класс побеждает), при равной длине — алфавитный порядок.

### Затронутые файлы

- `internal/agent/graph/orchestrator.go`
- `internal/agent/graph/orchestrator_test.go`

### Подзадачи

- [x] Построить отсортированный список ключей `VulnClassCodes` один раз (longest-first, затем alpha).
- [x] Переписать `classifyVulnCode` на итерацию по этому списку.
- [x] Тест: неоднозначные сообщения дают один и тот же код на сотнях вызовов.
- [x] Тест: специфичный ключ побеждает («session cookie» → `SESSION`, «authentication … jwt» → `AUTH`).

### Критерии приёмки

- [x] `classifyVulnCode` детерминирован для всех сообщений.
- [x] `assignVulnIDs` даёт одинаковые `CS-*` ID между запусками при одинаковом входе.
- [x] Web pipeline (`AssignVulnIDsPublic`) получает фикс автоматически (общий код).

### Риски

- `CS-*` ID в отчётах изменятся относительно прошлых runs (для неоднозначных findings). Это ожидаемо и требует bump cache namespace (Задача 5).

## Задача 2: Убрать FindingID из artifact disposition hash

### Описание

Защита в глубину: `Fingerprint` уже однозначно идентифицирует finding, `FindingID` — производная величина и не должна влиять на ключ. Одновременно bump `artifactCacheSchemaVersion` 1 → 2 для чистой инвалидации старых entries.

### Затронутые файлы

- `internal/agent/graph/artifact_cache.go`
- `internal/agent/graph/artifact_cache_test.go`

### Подзадачи

- [x] Убрать `FindingID` из `hashArtifactDisposition`.
- [x] Bump `artifactCacheSchemaVersion` до 2.
- [x] Тест: ключ не меняется, если отличаются только `VulnID`/`FindingID`.
- [x] Тест: существующие инварианты инвалидации (disposition/policy/model/provider) не сломаны.

### Критерии приёмки

- [x] Два состояния, идентичные кроме generated ID, дают одинаковый artifact key.
- [x] Старые entries (schema 1) дают miss `stale_or_mismatched_entry`, не ломают run.

### Риски

- Restored report.md содержит `CS-*` ID из warm run. С Задачей 1 ID детерминированы, поэтому совпадают с текущим `triage-findings.json`. Без Задачи 1 применять Задачу 2 нельзя.

## Задача 3: Диагностика key_miss

### Описание

Сделать то, чего не хватило handoff для доказательства root cause по remote-логам: логировать computed key, число загруженных entries и различать «кэш пуст» от «ключ не совпал».

### Затронутые файлы

- `internal/agent/graph/artifact_cache.go`
- `internal/agent/graph/orchestrator.go`
- `internal/agent/graph/artifact_cache_test.go`

### Подзадачи

- [x] `Restore`: при пустом entries map ставить `miss_reason=no_entries_loaded` вместо `key_miss`.
- [x] Добавить `key` в `ArtifactCacheStats` (попадает в `triage-findings.json` автоматически).
- [x] stderr log: computed artifact key + `loaded_entries` на warm и hit.
- [x] Тест: пустой кэш → `no_entries_loaded`; непустой без ключа → `key_miss`.

### Критерии приёмки

- [x] По одним только stderr-логам CI можно сравнить warm/hit key и наличие entries.
- [x] Ключ — это sha256-хеш, не содержит sensitive данных.

### Риски

- Нет. Изменения аддитивные.

## Задача 4: Локальная верификация

### Подзадачи

- [x] `go test ./internal/agent/graph`
- [x] `go test ./cmd/aitriage`
- [x] `go test ./...`
- [x] `git diff --check`

### Критерии приёмки

- [x] Все локальные тесты проходят.
- [x] Нет незапланированных изменений в диффе.

## Задача 5: Remote proof (GATED — только по явной команде пользователя)

### Описание

Повторить warm/hit эксперимент с исправленным image. НЕ ВЫПОЛНЯТЬ без явной команды.

### Подзадачи

- [ ] Bump namespace `aitriage-ai-v3` → `aitriage-ai-v4` в `security-workflows` reusable workflow.
- [ ] Опубликовать обновлённый pilot image `ghcr.io/cybertortuga/aitriage:codex-artifact-reuse-cache`.
- [ ] Warm run на accrual-ai: проверить `artifact_cache.key` в артефакте, `stores=1 saved=true`.
- [ ] Hit run на том же SHA: ожидать `Artifact cache exact hit`, нулевой LLM usage по poc/report/fixspec.
- [ ] Сравнить warm/hit key по логам (теперь видны).
- [ ] Записать run IDs и метрики сюда и в `artifact-reuse-cache.md` Task 7.

### Критерии приёмки

- [ ] Exact artifact hit доказан в реальном CI.
- [ ] Runtime hit run существенно ниже ~19m warm floor.
- [ ] Гейт/exit code не изменились.

## Задача 6: Cleanup / решение

- [ ] Обновить `artifact-reuse-cache.md` (Task 7/8) результатами.
- [ ] Решить судьбу root MD-файлов (не коммитить без решения).
- [ ] Kaiten — только по подтверждению, кратко, по-русски, по задачам.

## Цель (goal recap)

Artifact reuse cache для AITriage CI: при повторном run на том же SHA/provider/model/policy/findings/verdicts восстанавливать `poc_results`/`report.md`/`fixspec.md` из кэша без LLM-вызовов, не ослабляя security gate. Verdict cache уже работает; artifact exact hit в реальном CI ещё не доказан. Remote proof (Задача 5) — только по явной команде пользователя.

## Аудит и ресерч (2026-07-03, без запуска GHA)

### Проверенные допущения дизайна (подтверждены документацией)

- **restore-keys**: при нескольких partial matches восстанавливается самый свежий кэш — наш run-scoped дизайн (`...-<sha>-run-<run_id>` + prefix restore-keys) корректен. Источники: GitHub docs dependency-caching, actions/cache README.
- **Branch scope кэша**: run видит кэши своей ветки + default branch; sibling-ветки изолированы. Warm/hit на одной ветке — ок. Для будущих PR: кэш с main будет доступен PR-ам (прогрев с main — рабочая стратегия после мержа).
- **Лимиты**: key ≤ 512 chars (наш ~110), кэш живёт 7 дней без обращений, 10GB LRU на репо — run-scoped ключи (~50KiB) не проблема.
- **`always()`**: по docs выполняется и при cancellation; для timeout-случая рекомендация docs — не полагаться слепо (проверить на следующем pilot run, что save-step реально отработал после отмены). `!cancelled()` НЕ подходит — он выключает шаг при отмене.
- **Env-propagation в Docker action**: step-level env доходит до контейнера (доказано работой `AITRIAGE_LLM_*` в текущих ранах) — `AITRIAGE_CLASSIFY_RETRIES`/`AITRIAGE_BATCH_SIZE` доедут.

### Новые находки

- **Проблема №8 (design tradeoff): `AITRIAGE_VERSION` инвалидирует все кэши на каждый image.** Dockerfile передаёт `AITRIAGE_VERSION=<ref>-<sha>` в ENV, а verdict namespace включает это значение. Каждая публикация pilot image = полный cold start кэшей (объясняет `0 cache` на первом v4 run). Для production: каждый релиз `@v1` инвалидирует кэши всех потребителей. Пока оставляем (безопасно по умолчанию), зафиксировано как осознанный tradeoff; вариант на будущее — включать в ключ только major/schema-версии.
- **Проблема №9 (перфоманс/надёжность): batch size 150 — источник и латентности, и nr-fallback.** 212 findings → 2 гигантских промпта; конкурентность 4 не используется, а модель теряет findings из огромных промптов (замечены 9 nr-fallback). Фикс: `AITRIAGE_BATCH_SIZE: '50'` в reusable workflow (5 батчей параллельно, меньше prompt, меньше потерь). Ожидаемое ускорение classification в 2-3 раза — атакует и timeout (№6).
- **Наблюдение: рост `triage_cache.json`.** Записи не имеют timestamp и не прюнятся; при restore-keys chaining файл будет копить fingerprints старых SHA. Скорость роста мала (~100KB на полный набор), 7-дневная eviction кэша ограничивает. Отложено как будущее улучшение (LRU-cap по числу записей).
- **Наблюдение: restored report содержит дату warm run** (`time.Now()` при генерации). Это честное поведение («это тот самый отчёт»), зафиксировано как известное.

### Актуальный реестр проблем (№1–9)

> Продолжение реестра (№10–15) и полный аудит кэш-поверхности: см. `cache-full-audit.md` (2026-07-04, все фиксы локально).

1. ✅ map iteration в `classifyVulnCode` → рандомные CS-ID (fix: 9c2cbfa).
2. ✅ `FindingID` в disposition hash (fix: 9c2cbfa).
3. ✅ nr-fallback не кэшируется → телеметрия `uncached_verdicts` + `AITRIAGE_CLASSIFY_RETRIES=4` (локально).
4. ✅ Save() не писал файл без stores → `ensureFile()` при init (локально).
5. ✅ Иммутабельный GitHub cache + `cache-hit != true` замораживал деградированный кэш → run-scoped keys (локально).
6. ✅ `timeout-minutes: 25` убивает warm run (GLM classification 7.9→15.4 мин за день) → 45 минут (локально).
7. ✅ Timeout терял готовый verdict cache → ensureFile (локально, = №4).
8. 📝 `AITRIAGE_VERSION` в namespace — осознанный tradeoff, задокументирован.
9. ✅ Batch 150 → `AITRIAGE_BATCH_SIZE: '50'` в workflow (локально).

## Progress Log

- [x] 2026-07-03: Root cause доказан локально (map iteration в `classifyVulnCode`, 6/213 неоднозначных findings в warm-артефакте).
- [x] 2026-07-03: План создан.
- [x] 2026-07-03: Задача 1 выполнена: детерминированный `classifyVulnCode` (longest-first порядок ключей) + тесты детерминизма и специфичности.
- [x] 2026-07-03: Задача 2 выполнена: `FindingID` убран из `hashArtifactDisposition`, schema version 1→2, тест стабильности ключа при разных `VulnID`.
- [x] 2026-07-03: Задача 3 выполнена: `no_entries_loaded` miss reason, `artifact_cache.key` в telemetry, stderr-лог ключа и loaded_entries.
- [x] 2026-07-03: Задача 4 выполнена: `go test ./...` полностью зелёный, `git diff --check` чистый.
- [ ] Задача 5 ожидает явной команды пользователя (GHA запрещены).
- [x] 2026-07-03: Codex re-check: `go test -count=1 ./internal/agent/graph ./cmd/aitriage ./internal/agent/llm`, `go test -count=1 ./...`, `gofmt`, `git diff --check` прошли локально.
- [ ] 2026-07-03: Codex scope note: в рабочем дереве есть out-of-scope diff в `README.md`; не включать в тестовый commit/remote proof без отдельного решения.
- [x] 2026-07-03: AITriage commit `9c2cbfa` pushed; pilot image published by `https://github.com/cybertortuga/aitriage/actions/runs/28662416816`, digest `sha256:7b566ad6e4a4ade3f023c79521883efd844ecb9a8fbe7cf4ee107128e5951e75`.
- [x] 2026-07-03: `security-workflows` cache namespace bumped to `aitriage-ai-v4` in commit `06bb9b6`.
- [x] 2026-07-03: Remote warm run `https://github.com/dodo-ai-platform/accrual-ai/actions/runs/28662830891` timed out/cancelled in `Run AI triage` after AI job `25m3s`; cache was not saved, hit run was not started.
- [x] 2026-07-03: Warm run timing from logs: scan ~13s, threat model/classification ~15m24s (`203 llm`, `9 nr-fallback`), PoC ~3m27s, report ~3m37s, cancelled during fixspec. Artifact key was `v2:145c519201802535ec18adac4ccbb1145dba44e0386b276af79ad996ad18233f`, miss `no_entries_loaded`.
- [x] 2026-07-03: Найдена проблема №3 (по логам warm run): `9 nr-fallback` вердиктов не сохраняются в verdict cache → на hit run эти findings уйдут в LLM → rationale изменится → artifact `key_miss`. Warm run с nr-fallback > 0 не может быть seed-ом artifact-кэша.
- [x] 2026-07-03: Найдена проблема №4: `artifactBundleCache.Save()` не писал файл без stores (sensitive-skip) → workflow save-условие `hashFiles(artifact_bundle_cache.json)` блокировало сохранение И verdict cache.
- [x] 2026-07-03: Найдена проблема №5: immutable GitHub cache + условие `cache-hit != 'true'` → hit run, дообогативший verdict cache (классифицировав nr-fallback findings), не мог его сохранить — деградированный кэш замораживался навсегда для SHA.
- [x] 2026-07-03: Фиксы (локально, без push): (1) телеметрия `artifact_cache.uncached_verdicts` + warning в stderr/summary; (2) env-knob `AITRIAGE_CLASSIFY_RETRIES` (default 2, в workflows выставлен 4); (3) `Save()` всегда пишет cache-файл; (4) run-scoped cache keys (`...-<sha>-run-<run_id>`) + restore-keys в reusable workflow и public example (example bumped v2→v3), save-условие без `cache-hit`.
- [x] 2026-07-03: Верификация: `go test ./...` — 20 пакетов ok; `git diff --check` чисто; gofumpt чисто; golangci-lint — только 5 pre-existing замечаний; YAML parse ok (`action.yml`, example, reusable workflow). Новых GHA runs не запускалось.
- [x] 2026-07-03: Анализ всех ранов accrual-ai. Найдена проблема №6: run `28662830891` убит **job timeout** (`timeout-minutes: 25`, kill ровно на 25m02s), а не отменён вручную. Тренд деградации GLM classification на одних и тех же 212 findings: 7m51s (02.07) → 12m03s (03.07 утро) → 15m24s (03.07 день); warm run больше не влезает в 25 минут.
- [x] 2026-07-03: Найдена проблема №7: при timeout терялся уже сохранённый verdict cache — save-step требует `artifact_bundle_cache.json`, который создавался только в конце run. Фикс: `ensureFile()` создаёт пустой cache-файл при инициализации; timeout после classification больше не теряет вердикты (save-step с `always()` выполняется при отмене job).
- [x] 2026-07-03: Фикс проблемы №6: `timeout-minutes` 25 → 45 в reusable workflow (локально, без push). Save() возвращён к простой dirty-семантике (ensureFile покрывает существование файла).
- [x] 2026-07-03: Повторная верификация: `go test ./...` 20 ok, `git diff --check` чисто, gofmt чисто по изменённым файлам, YAML ok.
