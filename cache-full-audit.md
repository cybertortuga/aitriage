# CACHE FULL AUDIT

Дата: 2026-07-03
Ветка: `codex/artifact-reuse-cache`
Контекст: продолжение `artifact-cache-key-determinism.md` (проблемы №1–9) и `artifact-reuse-cache.md`. Полный аудит всей кэш-поверхности по требованию пользователя.
Ограничения: GHA не запускать, Kaiten не трогать, не коммитить/пушить без явной команды, public repo — нейтральные комментарии.

## Обзор

Полный read-only аудит verdict cache + artifact bundle cache: код (`internal/agent/graph/*`), wiring (`cmd/aitriage/agent.go`, `action.yml`, `entrypoint.sh`, `Dockerfile`, workflows), env-инвентарь, secrets-поверхность, конкурентность, рост файлов, атомарность записи. Инструменты: filesystem (Read/Grep), Explore-агент (breadth sweep), WebSearch/WebFetch (GitHub docs), gh CLI, git.

## Результаты аудита

### Проверено и ПОДТВЕРЖДЕНО КОРРЕКТНЫМ

- **Policy в ключе** — `withVerdictCachePolicy(state.Policy)` переопределяет env-значения реальной политикой из флагов (`orchestrator.go` → `classify.go` → `cache.go`). Ключ отражает фактическую политику. (Гипотеза breadth-свипа о рассинхроне опровергнута.)
- **Secrets-скан кэшируемых значений** — свободный текст в кэше: `Rationale` (проверяется) и `Evidence.Observed` (проверяется). `ClassificationAuditEntry.RawResponse` в кэш НЕ попадает (только в triage-findings.json артефакт — осознанная политика retention 14d).
- **Конкурентность** — `cache.Get/Set` вызываются только из основной горутины `classifyUnique` (после `wg.Wait()`); audit collector под мьютексом. Гонок нет.
- **Fork-защита CI** — restore-step скипается для fork PR, save требует `restore.outcome == 'success'` → форки не пишут кэш.
- **Path traversal в evidence** — `readProjectLine` отклоняет абсолютные пути и выход за project root.
- **Самовосстановление после corruption** — битый JSON игнорируется (`CorruptCacheIgnored`), следующий warm перезаписывает файл.
- **Retry-клиент** — ретраит deadline/429/5xx с backoff, уважает ctx.Done.
- **Env-propagation** — step-env доходит до Docker action (доказано `AITRIAGE_LLM_*` в реальных ранах).

### НОВЫЕ ПРОБЛЕМЫ (продолжение реестра №1–9 из artifact-cache-key-determinism.md)

- **№10 (критично, надёжность): LLM request timeout отсутствует.** `client.go` обещает «default 120», но дефолт нигде не выставлен: `Timeout=0` → `WithRequestTimeout` не применяется ни для одного провайдера (`factory.go:176,202,221` — все под `if cfg.Timeout > 0`). Зависший вызов висит до job-timeout; ретраи не срабатывают (первый вызов не возвращается).
- **№11 (критично, корректность): identity модели в verdict-ключе — только из env.** `defaultVerdictCacheKeyContext` читает `AITRIAGE_LLM_PROVIDER/MODEL`, `AITRIAGE_MODEL`, `AITRIAGE_LLM_BASE_URL`, но реальный клиент конфигурируется флагами `--provider/--model`, `.aitriage.yaml` или автодетектом по API-ключу (`config.go:117-125`). Если env не выставлены (как в public example — там только `GEMINI_API_KEY`), namespace получает `provider=default, model=default` → вердикты одной модели могут быть переиспользованы другой. Наш reusable workflow не подвержен (env выставлены), но CLI/example — подвержены.
- **№12 (средне): `DisableThinking` меняет поведение модели, но не входит в ключ.** Прочие env (`AITRIAGE_GATING`, `AITRIAGE_BATCH_SIZE`, `AITRIAGE_LLM_BUDGET`, `AITRIAGE_CONCURRENCY`, `AITRIAGE_CLASSIFY_RETRIES`) в ключ сознательно НЕ включаем: они влияют на охват/латентность/батчинг, но не на валидность конкретного кэшированного вердикта (фиксируется в этом документе как контракт).
- **№13 (низко): небезопасная запись cache-файлов.** `os.WriteFile` неатомарен: kill посреди записи → обрезанный JSON → на следующем ране весь кэш выброшен как corrupt. Лечится temp-file + `os.Rename` (атомарен в пределах одной FS).
- **№14 (низко): неограниченный рост artifact_bundle_cache.json.** Entries никогда не удаляются; при restore-keys chaining файл копит ~100–200KB на каждый набор dispositions. Verdict cache растёт медленнее (fingerprint-scoped) — отложен, 7-дневная eviction ограничивает.
- **№15 (информационно): `AITRIAGE_MODEL` — фантомный alias.** Читается только кэшем/classify, но не config → не влияет на клиент. После фикса №11 identity едет через state; env остаётся fallback-ом.

## Затронутые файлы и области

- `internal/agent/graph/state.go` — поля LLM identity в AgentState
- `internal/agent/graph/cache.go` — identity-option, DisableThinking в контексте ключа, атомарная запись
- `internal/agent/graph/classify.go` — прокидывание identity-option
- `internal/agent/graph/orchestrator.go` — identity в buildThreatModel/buildArtifactCacheKey
- `internal/agent/graph/artifact_cache.go` — identity в ключе, атомарная запись, prune
- `internal/agent/llm/factory.go`, `client.go` — дефолт timeout
- `cmd/aitriage/agent.go` — заполнение identity в state
- `security-workflows/.github/workflows/aitriage.yml` — `AITRIAGE_LLM_TIMEOUT`
- тесты соответствующих пакетов

## Задача 1: LLM request timeout по умолчанию (№10)

### Описание
Выставить дефолт `Timeout` в `llm.NewClient` (600s — щадяще для медленных local/ollama сценариев), исправить лживый комментарий, задать явный `AITRIAGE_LLM_TIMEOUT: '300'` в reusable workflow.

### Затронутые файлы
- `internal/agent/llm/factory.go`, `internal/agent/llm/client.go`
- `security-workflows/.github/workflows/aitriage.yml`

### Подзадачи
- [x] Дефолт в `NewClient`: `if cfg.Timeout <= 0 { cfg.Timeout = 600 }`.
- [x] Комментарий `client.go` → default 600.
- [x] `AITRIAGE_LLM_TIMEOUT: '300'` в reusable workflow env.
- [x] Тест: конфиг без timeout получает дефолт.

### Критерии приёмки
- [x] Зависший вызов провайдера обрывается request-timeout-ом и уходит в retry.
- [x] Локальные сценарии (ollama) не ломаются: дефолт 600s, переопределяем.

### Риски
- Слишком низкий дефолт оборвёт легитимные долгие батчи → выбран 600s + явные 300s только в CI.

## Задача 2: Identity модели в ключе из resolved config (№11, №12, №15)

### Описание
Прокинуть фактические provider/model/baseURL/disableThinking из `cmd/aitriage/agent.go` через `AgentState` в verdict-namespace и artifact-key. Env остаётся fallback-ом (обратная совместимость). Bump `verdictCacheSchemaVersion` 3→4 (чистая инвалидация: старые namespace-ы с «default» identity умирают).

### Затронутые файлы
- `internal/agent/graph/state.go`, `cache.go`, `classify.go`, `orchestrator.go`, `artifact_cache.go`
- `cmd/aitriage/agent.go`

### Подзадачи
- [x] Поля `LLMProvider/LLMModel/LLMBaseURL/LLMDisableThinking` в `AgentState` (без API-ключей).
- [x] `withVerdictCacheLLMIdentity(...)` option: непустые значения из state переопределяют env.
- [x] `DisableThinking bool` в `verdictCacheKeyContext` + bump schema 3→4.
- [x] `agent.go` заполняет identity из resolved `llmCfg`.
- [x] `buildArtifactCacheKey` использует ту же identity.
- [x] Тесты: ключ меняется при state-identity, state побеждает env, пустой state → env fallback.

### Критерии приёмки
- [x] Смена `--model`/YAML-модели меняет verdict namespace и artifact key без env.
- [x] Наш CI-путь (env-driven) даёт тот же namespace, что и state-driven (значения совпадают).

### Риски
- Инвалидация всех существующих кэшей — приемлемо (pilot, и v4-кэша ещё нет).

## Задача 3: Атомарная запись cache-файлов (№13)

### Подзадачи
- [x] `writeFileAtomic(path, data)`: temp-файл в той же директории + `os.Rename`.
- [x] Использовать в `verdictCache.Save`, `artifactBundleCache.Save`, `ensureFile`.
- [x] Тест: Save создаёт файл без temp-мусора.

### Критерии приёмки
- [x] Обрыв процесса не может оставить обрезанный JSON на месте валидного кэша.

### Риски
- Нет: rename в пределах одной директории атомарен на POSIX.

## Задача 4: Prune artifact bundle entries (№14)

### Подзадачи
- [x] На `Save` оставлять не более 8 самых свежих entries по `created_at`.
- [x] Тест: 10 stores → в файле 8 новейших.

### Критерии приёмки
- [x] Файл не растёт неограниченно при restore-keys chaining.

### Риски
- Слишком агрессивный prune снизит hit-rate у активных SHA — 8 достаточно (по entry на набор dispositions).

## Задача 5: Верификация

### Подзадачи
- [x] `go test ./...` полный.
- [x] `git diff --check`, gofmt по изменённым файлам, golangci-lint (не хуже baseline: 5 pre-existing).
- [x] YAML parse reusable workflow.
- [x] Обновить реестр проблем в `artifact-cache-key-determinism.md`.

### Критерии приёмки
- [x] Все тесты зелёные, ничего не запушено, GHA не запускались.

## Риски и зависимости

- Все фиксы локальные; remote proof (warm/hit) по-прежнему gated на явную команду.
- Bump schema инвалидирует кэши — согласовано с pilot-фазой.
- `AITRIAGE_VERSION` в namespace (№8) остаётся осознанным tradeoff.

## Progress Log

- [x] 2026-07-03: Шаги 1–3 (инструменты, read-only аудит, research) завершены; breadth-свип Explore-агентом; policy-гипотеза опровергнута; реестр пополнен №10–15.

## Финальный отчёт

Дата завершения: 2026-07-04. Все задачи 1–5 выполнены, все критерии приёмки закрыты.

**Реализовано (локально, без push/GHA):**

1. **№10** — `defaultedTimeout()` в `llm.NewClient`: дефолт 600s вместо «без таймаута»; `AITRIAGE_LLM_TIMEOUT: '300'` в reusable workflow. Зависший провайдер теперь обрывается и ретраится.
2. **№11/12/15** — resolved LLM identity (`LLMProvider/LLMModel/LLMBaseURL/LLMDisableThinking`) идёт из `cmd/aitriage/agent.go` через `AgentState` в verdict namespace и artifact key (`withVerdictCacheLLMIdentity`); env — fallback. `DisableThinking` включён в контекст ключа. `verdictCacheSchemaVersion` 3→4.
3. **№13** — `writeCacheFileAtomic` (temp + rename) для обоих кэшей и `ensureFile`.
4. **№14** — prune artifact bundle до 8 новейших entries по `created_at` на Save.

**Верификация:** `go build ./...` ок; `go test ./...` — 20 пакетов ок (новые тесты: identity namespace/key, prune+atomic, defaultedTimeout); `git diff --check` чисто; gofmt чисто по изменённым файлам; golangci-lint — 0 замечаний в моих строках (все остальные pre-existing); YAML reusable workflow парсится.

**Изменённые файлы:** `internal/agent/graph/{state,cache,classify,orchestrator,artifact_cache}.go` + тесты, `internal/agent/llm/{factory,client}.go` + тест, `cmd/aitriage/agent.go`, `security-workflows/.github/workflows/aitriage.yml`.

**Незакрытых TODO нет.** Remote proof (warm/hit) остаётся gated на явную команду пользователя. Суммарный реестр по всей работе: №1–15, из них исправлено 13, №8 — документированный tradeoff, verdict-cache growth — отложено осознанно.
