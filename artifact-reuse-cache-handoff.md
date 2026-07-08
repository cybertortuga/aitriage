# Artifact Reuse Cache Handoff

Дата handoff: 2026-07-03  
Репозиторий: `/Users/afedotov/Documents/GitHub/aitriage`  
Текущая ветка: `codex/artifact-reuse-cache`  
Назначение файла: локальный контекст для другого AI агента. Это не финальный отчет, не PR description и не публичный документ для open-source репозитория.

## Коротко

Мы пытались добавить второй слой кэша для AITriage CI: не только кэшировать verdict/classification по finding, но и переиспользовать готовые secondary artifacts:

- `poc_results`
- `report.md`
- `fixspec.md`

Цель была простая: если повторный CI run идет по тому же SHA/ref/provider/model и те же findings уже классифицированы, AITriage не должен снова тратить LLM на PoC/report/fixspec. Он должен взять прошлый bundle из cache, сохранить security gate behavior, загрузить артефакты и завершиться быстрее/дешевле.

Текущее состояние: verdict cache работает, GitHub cache exact restore работает, но artifact bundle reuse в реальном CI не доказан и сейчас фактически ломается с `Artifact cache miss: key_miss`. Из-за этого hit-run опять уходит в PoC/report generation и тратит время/токены.

Самая важная проблема: по текущим логам я не могу доказать точную причину `key_miss`. Логи не показывают computed artifact key, saved artifact key, loaded artifact entries и наличие/содержимое `artifact_bundle_cache.json` после restore. Поэтому я не знаю точный root cause без дополнительной диагностики.

## Жесткие ограничения от пользователя

- Не запускать GitHub Actions вообще, пока пользователь явно не скажет.
- Не продолжать прожигать CI minutes и LLM tokens.
- Разбирать только существующие логи/артефакты.
- Не писать в Kaiten без явного подтверждения.
- Если обновлять Kaiten после подтверждения, писать кратко, по-русски, по задачам и без размытого английского прогресса.
- Не коммитить root MD/planning/handoff файлы в public repo без отдельного решения.

## Карточки Kaiten по кэшу

### Карточка 66752084

URL: `https://dodopizza.kaiten.ru/space/39142/boards/card/66752084`

Это была первая задача/исследование по LLM caching: понять, что такое AI cache, поможет ли он AITriage CI/CD, и провести первый pilot. Карточка перенесена пользователем в Done.

Что было доказано в той фазе:

- verdict/classification cache может убрать повторную LLM-классификацию findings;
- на повторном run classification действительно перестает дергать LLM по 212 findings;
- но после этого остаются тяжелые secondary stages: PoC/report/fixspec.

### Карточка 66823497

URL: `https://dodopizza.kaiten.ru/space/39142/boards/card/66823497`

Текущая карточка: artifact reuse cache для AITriage CI. Именно по ней велась работа с файлом плана `artifact-reuse-cache.md`.

Комментарии, которые уже были добавлены:

- `85145529`: зафиксирован blocker перед Task 7A, что нельзя запускать новые hit/Xiaomi runs, пока локально не починен false positive в sensitive guard.
- `85146951`: зафиксирован локальный результат Task 7A.

Важно: пользователь был недоволен, когда прогресс-комментарии были на английском и не по задачам. Дальше в карточку нельзя писать размыто; только коротко, по-русски, с привязкой к Task/статусу/проверкам.

### Parent / исходный контекст

Пользователь также указывал parent/context card:

- `https://dodopizza.kaiten.ru/space/39142/boards/card/63350717`

## Ветки и репозитории

### AITriage

Путь: `/Users/afedotov/Documents/GitHub/aitriage`  
Текущая ветка: `codex/artifact-reuse-cache`  
Текущее состояние на момент handoff:

```text
## codex/artifact-reuse-cache...origin/codex/artifact-reuse-cache
3426004 (HEAD -> codex/artifact-reuse-cache, origin/codex/artifact-reuse-cache) fix: avoid artifact cache secret false positive
08a4bf4 fix: stabilize artifact cache finding order
4c3ff26 feat: reuse cached AI triage artifacts
e8a4692 (origin/codex/llm-caching-pilot, codex/llm-caching-pilot) feat: pilot deterministic LLM verdict caching
fe8974b (origin/main, origin/HEAD, main) Add finding remediation handoff workflow
```

Ветка до artifact reuse была `codex/llm-caching-pilot`. До нее актуальный upstream main указывал на `fe8974b`.

Изменения относительно main затрагивают 28 файлов, примерно `1939 insertions(+), 106 deletions(-)`.

Основные области diff:

- `internal/agent/graph/artifact_cache.go`
- `internal/agent/graph/orchestrator.go`
- `internal/agent/graph/cache.go`
- `internal/agent/graph/classify.go`
- `internal/agent/graph/poc.go`
- `internal/agent/graph/usage.go`
- `internal/agent/graph/triage_artifact.go`
- `internal/agent/llm/*`
- `internal/agent/prompts/templates.go`
- `action.yml`
- `examples/github-actions/aitriage-security.yml`
- Docker/workflow files

### security-workflows

Путь: `/Users/afedotov/Documents/GitHub/security-workflows`  
Текущая ветка: `codex/artifact-reuse-cache`

```text
## codex/artifact-reuse-cache...origin/codex/artifact-reuse-cache
5481121 (HEAD -> codex/artifact-reuse-cache, origin/codex/artifact-reuse-cache) ci: refresh AITriage artifact cache namespace
6a7a1c4 ci: bump AITriage artifact cache namespace
cc7fa29 ci: pilot AITriage artifact cache workflow
49db71c (origin/codex/llm-caching-pilot, codex/llm-caching-pilot) ci: test AITriage cache pilot workflow
ff0dec8 (origin/main, origin/HEAD, main) Merge pull request #9 from dodo-ai-platform/codex/use-aitriage-v1-action
```

Изменен reusable workflow `.github/workflows/aitriage.yml`:

- добавлен restore/save path `.aitriage-cache`;
- key namespace сейчас `aitriage-ai-v3`;
- AITriage Action вызывается с `verdict-cache-dir: .aitriage-cache` и `artifact-cache-dir: .aitriage-cache`;
- save cache требует наличие и `triage_cache.json`, и `artifact_bundle_cache.json`;
- workflow использует branch Action ref `cybertortuga/aitriage@codex/artifact-reuse-cache`.

### accrual-ai

Путь: `/Users/afedotov/Documents/GitHub/accrual-ai`  
Текущая ветка: `codex/artifact-reuse-cache`

```text
## codex/artifact-reuse-cache...origin/codex/artifact-reuse-cache
a5be9ae (HEAD -> codex/artifact-reuse-cache, origin/codex/artifact-reuse-cache) ci: pilot AITriage artifact cache
874176b (origin/codex/llm-caching-pilot, codex/llm-caching-pilot) ci: route AITriage pilot workflow
```

Caller workflow `.github/workflows/aitriage.yml` переключен на reusable workflow branch:

```text
dodo-ai-platform/security-workflows/.github/workflows/aitriage.yml@codex/artifact-reuse-cache
```

Пилотные ручные runs запускались на ветке `codex/artifact-reuse-cache` с GLM/Z.ai параметрами:

- `llm-provider=openai`
- `llm-model=glm-5.2`
- `llm-base-url=https://api.z.ai/api/coding/paas/v4`
- secret selector: `GLM_CI_KEY`

Значение секрета нигде не фиксировалось.

## Плановый файл

Основной план: `/Users/afedotov/Documents/GitHub/aitriage/artifact-reuse-cache.md`

Важное замечание: шапка этого файла сейчас частично устарела. Там все еще написано `Planning only. Код не менялся`, но фактически код уже менялся, коммиты были запушены, image публиковался, remote CI runs выполнялись. Ниже в progress log этого же файла есть более актуальные записи до 2026-07-03.

По плану:

- Task 0 закрыта: audit/plan gate.
- Task 1 закрыта: artifact bundle cache contract.
- Task 2 закрыта локально: graph full-hit restore path.
- Task 3 закрыта локально: PoC refusal-safe miss path.
- Task 4 закрыта локально: telemetry in artifacts/summary.
- Task 5 закрыта локально/workflow: trusted CI cache path.
- Task 6 закрыта локально: tests and local verification.
- Task 7 не закрыта: remote CI pilot and measurement.
- Task 7A закрыта: sensitive guard false positive fix.
- Task 8 не закрыта: final decision/cleanup.

## Что уже сделали в AITriage

### 1. Добавили artifact bundle cache

Файл: `internal/agent/graph/artifact_cache.go`

Добавлен отдельный cache file:

```text
artifact_bundle_cache.json
```

Директория берется из:

- `AITRIAGE_ARTIFACT_CACHE_DIR`
- fallback: `AITRIAGE_CACHE_DIR`

Cache entry содержит:

- schema version;
- key;
- created_at;
- `PoCResults`;
- `ReportMarkdown`;
- `FixSpecMarkdown`;
- counts.

### 2. Встроили restore point в graph

Файл: `internal/agent/graph/orchestrator.go`

Текущий порядок:

```text
gatherRepoContext
enrichFindings
buildThreatModel
buildArtifactCacheKey / Restore
if miss: runPoCVerification
computeHealthCheck
if miss: generateReport
if miss: generateAIFixSpec
if miss: Store/Save artifact bundle
generateSummary
```

Важно: artifact cache сейчас не может пропустить scanner/context/threat model. Но verdict cache делает threat model быстрым, если classification hit полный.

### 3. Добавили telemetry

В `triage-findings.json` добавлен блок `artifact_cache`.

Пример expected fields:

- `enabled`
- `loaded_entries`
- `exact_hit`
- `miss_reason`
- `restored_poc`
- `restored_report`
- `restored_fixspec`
- `stores`
- `skipped_sensitive`
- `corrupt_cache_ignored`
- `saved`

В stderr добавлены строки вида:

```text
Artifact cache miss: key_miss
Artifact cache exact hit: restored PoC/report/fixspec bundle
Artifact cache store: stores=1 saved=true skipped_sensitive=0
```

### 4. Добавили LLM usage accounting

Добавлена stage usage telemetry, чтобы видеть tokens по classification/threat_model/poc/report/fixspec.

### 5. Сделали PoC miss path отказоустойчивее

PoC prompt был переписан в defensive exploitability assessment. Provider refusal должен превращаться в Needs Manual Review PoC result, а не ломать весь run молча.

### 6. Исправили false positive sensitive guard

Был баг: safe package name `flask-limiter` матчился на raw substring `sk-`, из-за чего artifact bundle не сохранялся.

Коммит:

```text
3426004 fix: avoid artifact cache secret false positive
```

Sensitive guard стал token-shaped / boundary-aware. Локальные тесты подтвердили:

- `flask-limiter` больше не блокирует artifact bundle cache;
- реальные secret-like markers все еще блокируются.

## Что менялось в CI/workflows

### AITriage action

Файл: `action.yml`

Добавлен input:

```text
artifact-cache-dir
```

Pilot branch runtime image:

```text
docker://ghcr.io/cybertortuga/aitriage:codex-artifact-reuse-cache
```

В файле есть комментарий, что этот image tag не должен попасть в stable main release.

### Reusable workflow

Файл: `/Users/afedotov/Documents/GitHub/security-workflows/.github/workflows/aitriage.yml`

Cache restore:

```text
path: .aitriage-cache
key: aitriage-ai-v3-${{ runner.os }}-${{ github.repository }}-${{ inputs.project-dir }}-${{ inputs.llm-provider }}-${{ inputs.llm-model || 'default' }}-${{ github.sha }}
restore-keys:
  aitriage-ai-v3-${{ runner.os }}-${{ github.repository }}-${{ inputs.project-dir }}-${{ inputs.llm-provider }}-${{ inputs.llm-model || 'default' }}-
```

AITriage call:

```text
verdict-cache-dir: .aitriage-cache
artifact-cache-dir: .aitriage-cache
```

Cache save condition:

```text
hashFiles('.aitriage-cache/triage_cache.json') != '' &&
hashFiles('.aitriage-cache/artifact_bundle_cache.json') != ''
```

## Локальные проверки, которые проходили

В разные моменты проходили:

```text
go test ./internal/agent/graph
go test ./internal/agent/llm
go test ./cmd/aitriage
go test ./...
git diff --check
```

Также парсились YAML-файлы:

- `action.yml`
- `examples/github-actions/aitriage-security.yml`
- reusable workflow `security-workflows/.github/workflows/aitriage.yml`

## Remote runs и результаты

Ниже только уже выполненные runs. Новые runs запускать нельзя без явной команды пользователя.

### Старые baseline/reference runs

Пользователь давал ссылки как baseline для LLM cache:

- GLM warm: `https://github.com/dodo-ai-platform/accrual-ai/actions/runs/28587174328`
- GLM hit: `https://github.com/dodo-ai-platform/accrual-ai/actions/runs/28588542116`
- Xiaomi warm: `https://github.com/dodo-ai-platform/accrual-ai/actions/runs/28589434824`
- Xiaomi hit: `https://github.com/dodo-ai-platform/accrual-ai/actions/runs/28590824936`

Они относятся к первой фазе LLM verdict caching, не к текущему artifact bundle reuse.

### Artifact reuse v1

#### Warm run 28647107728

URL: `https://github.com/dodo-ai-platform/accrual-ai/actions/runs/28647107728`

Результат:

- expected workflow failure из-за security gate;
- artifact cache miss `key_miss`;
- artifact cache saved;
- `stores=1`;
- `skipped_sensitive=0`;
- GitHub cache saved with key `aitriage-ai-v1-Linux-dodo-ai-platform/accrual-ai-.-openai-glm-5.2-a5be9ae3ef12e9d9480b38ab4ddfab8a265d67e1`.

#### Hit run 28647927953

URL: `https://github.com/dodo-ai-platform/accrual-ai/actions/runs/28647927953`

Результат:

- GitHub cache restored exact primary key;
- verdict cache hit worked: `212/212`;
- artifact bundle missed: `Artifact cache miss: key_miss`;
- PoC/report/fixspec regenerated;
- secondary stages consumed about `59,733` LLM tokens.

Вывод тогда: artifact key менялся между warm/hit. Была найдена вероятная причина: нестабильный порядок scanner findings и generated `CS-*` IDs.

Fix:

```text
08a4bf4 fix: stabilize artifact cache finding order
```

После этого reusable workflow cache namespace bumped:

```text
aitriage-ai-v1 -> aitriage-ai-v2
```

### Artifact reuse v2

#### Run 28649365358

URL: `https://github.com/dodo-ai-platform/accrual-ai/actions/runs/28649365358`

Результат:

- discarded from measurement;
- manually cancelled during image pull before useful deterministic/AI stages.

#### Warm run 28649448924

URL: `https://github.com/dodo-ai-platform/accrual-ai/actions/runs/28649448924`

Результат:

- expected workflow failure;
- `Run AI triage` succeeded;
- took about `19m26s`;
- artifact bundle was not saved;
- `artifact_cache.miss_reason=sensitive_bundle_skipped`;
- `stores=0`;
- `skipped_sensitive=1`;
- `saved=false`;
- verdict cache saved true.

Root cause found locally: report contained safe text/package name `flask-limiter`, but sensitive guard matched raw substring `sk-`.

Fix:

```text
3426004 fix: avoid artifact cache secret false positive
```

Then image was published:

Run: `https://github.com/cybertortuga/aitriage/actions/runs/28652631738`

Image:

```text
ghcr.io/cybertortuga/aitriage:codex-artifact-reuse-cache
sha256:2ab8281f450417682ebea76b0609ff30fa86bc56ba319ab47e6d81b74332cf9b
version codex/artifact-reuse-cache-34260044999792babcf31b3589be7ac7fa9d4e97
```

Reusable workflow cache namespace bumped:

```text
aitriage-ai-v2 -> aitriage-ai-v3
```

Commit:

```text
5481121 ci: refresh AITriage artifact cache namespace
```

### Artifact reuse v3

#### Warm run 28653128012

URL: `https://github.com/dodo-ai-platform/accrual-ai/actions/runs/28653128012`

Результат:

- expected workflow failure;
- `Run AI triage` succeeded;
- fresh `aitriage-ai-v3` cache missed first;
- artifact bundle cache stored successfully:
  - `stores=1`
  - `saved=true`
  - `skipped_sensitive=0`
- GitHub cache saved.

Key log lines:

```text
Cache not found for input keys: aitriage-ai-v3-Linux-dodo-ai-platform/accrual-ai-.-openai-glm-5.2-a5be9ae3ef12e9d9480b38ab4ddfab8a265d67e1, ...
Triage scale: 213 findings -> 212 unique (1 deduped) | sources: 212 llm, 0 cache, 0 deterministic, 0 nr-fallback
Threat model: 32 True Positives, 155 False Positives, 26 Needs Review
Artifact cache miss: key_miss
PoC: 31 unique TPs -> 16 exploitable, 15 blocked/unknown
Artifact cache store: stores=1 saved=true skipped_sensitive=0
LLM usage: 472765 total · 346160 prompt · 126605 completion · cache telemetry: 98816 cached prompt
Cache saved with key: aitriage-ai-v3-Linux-dodo-ai-platform/accrual-ai-.-openai-glm-5.2-a5be9ae3ef12e9d9480b38ab4ddfab8a265d67e1
```

Downloaded warm artifact existed locally at:

```text
/tmp/aitriage-28653128012
```

Files observed:

```text
77495 report.md
24360 fixspec.md
712261 triage-findings.json
48512 summary.md
```

Telemetry observed in `triage-findings.json`:

- `artifact_cache.enabled=true`
- `artifact_cache.exact_hit=false`
- `artifact_cache.miss_reason=key_miss`
- `artifact_cache.stores=1`
- `artifact_cache.skipped_sensitive=0`
- `artifact_cache.saved=true`
- `verdict_cache.loaded_entries=0`
- `verdict_cache.hits=0`
- `verdict_cache.misses=212`
- `verdict_cache.stores=212`
- `llm_usage.total_tokens=472765`
- stage usage included:
  - classification: `279640`
  - threat_model: `94525`
  - poc: `23264`
  - report: `48387`
  - fixspec: `26949`

GitHub cache list showed:

```text
5465203734  aitriage-ai-v3-...  48.13 KiB  2026-07-03T10:23:30Z  2026-07-03T10:28:33Z
```

48.13 KiB looks suspiciously small for `triage_cache.json + artifact_bundle_cache.json`, but it is not conclusive because compression may be strong.

#### Hit run 28654594215

URL: `https://github.com/dodo-ai-platform/accrual-ai/actions/runs/28654594215`

Результат:

- run был force-cancelled after user stop request;
- exact GitHub cache restore succeeded;
- verdict cache worked;
- artifact cache missed again with `key_miss`;
- after miss it started PoC/report generation again;
- cancelled during report generation.

Key log lines:

```text
Cache hit for: aitriage-ai-v3-Linux-dodo-ai-platform/accrual-ai-.-openai-glm-5.2-a5be9ae3ef12e9d9480b38ab4ddfab8a265d67e1
Received 49286 of 49286 (100.0%)
Cache Size: ~0 MB (49286 B)
Cache restored successfully
Cache restored from key: aitriage-ai-v3-Linux-dodo-ai-platform/accrual-ai-.-openai-glm-5.2-a5be9ae3ef12e9d9480b38ab4ddfab8a265d67e1
Triage scale: 213 findings -> 212 unique (1 deduped) | sources: 0 llm, 212 cache, 0 deterministic, 0 nr-fallback
Threat model: 32 True Positives, 155 False Positives, 26 Needs Review
Artifact cache miss: key_miss
PoC Verification...
PoC: 30 unique TPs -> 12 exploitable, 18 blocked/unknown
Computing Security Health Check...
Generating Security Report...
The operation was canceled.
```

Именно этот run вызвал текущий кризис доверия: он снова начал тратить время/токены, хотя warm run уже сохранил cache.

Последний известный статус active runs после force cancel:

- `28654594215`: completed/cancelled
- `28653128012`: completed/failure
- `28649448924`: completed/failure
- `28649365358`: cancelled
- `28647927953`: completed/failure

На тот момент активных AITriage Security runs по ветке не оставалось.

## Что точно работает

### GitHub Actions cache exact restore

На v3 hit run GitHub Actions восстановил cache по точному key:

```text
Cache restored from key: aitriage-ai-v3-Linux-dodo-ai-platform/accrual-ai-.-openai-glm-5.2-a5be9ae3ef12e9d9480b38ab4ddfab8a265d67e1
```

Это не выглядит как partial restore.

### Verdict cache

На v3 hit run verdict cache сработал полностью:

```text
sources: 0 llm, 212 cache, 0 deterministic, 0 nr-fallback
```

То есть classification по findings уже не дергал LLM.

### Warm artifact store

На v3 warm run AITriage сам сообщил:

```text
Artifact cache store: stores=1 saved=true skipped_sensitive=0
```

И workflow сохранил GitHub cache после этого.

## Что точно не работает

### Artifact bundle exact hit

На v3 hit run artifact cache не восстановил PoC/report/fixspec:

```text
Artifact cache miss: key_miss
```

После этого AITriage пошел в:

- PoC Verification;
- Security Report generation;
- FixSpec would likely follow, но run был cancelled during report generation.

Именно это противоречит цели artifact reuse cache.

## Главная текущая проблема, которую я не знаю как решить по имеющимся логам

Artifact bundle cache пишет `stores=1 saved=true` на warm run, GitHub cache потом exact-restores на hit run, verdict cache внутри этого restore работает, но artifact bundle lookup возвращает `key_miss`.

Я не знаю точную причину, потому что текущие remote logs не показывают достаточно данных:

- не логируется computed artifact key на warm;
- не логируется computed artifact key на hit;
- не логируется список loaded artifact cache entries;
- не логируется, был ли физически найден `.aitriage-cache/artifact_bundle_cache.json` после restore;
- не логируется saved entry key внутри `artifact_bundle_cache.json`;
- `artifact_cache.loaded_entries` был доступен в artifacts warm run, но hit run был cancelled до artifact upload, поэтому нет hit `triage-findings.json` для сравнения.

Из-за этого нельзя честно сказать, что root cause уже найден.

## Возможные причины, но они не доказаны

Это не план действий и не рекомендация. Это список гипотез, которые появились из фактов.

### Гипотеза 1: `artifact_bundle_cache.json` не попал в restored GitHub cache

За:

- cache size в GitHub list около `48.13 KiB` / log `49286 B`, что выглядит маленьким для bundle с report/fixspec/PoC.
- если файла нет, `newArtifactBundleCache().load()` просто вернет пустой entries map, а `Restore` даст `key_miss`.

Против / неизвестно:

- workflow save condition требует `hashFiles('.aitriage-cache/artifact_bundle_cache.json') != ''`.
- warm AITriage логировал `saved=true`.
- compression может сделать размер сильно меньше, поэтому размер не доказательство.

### Гипотеза 2: файл есть, но artifact key отличается между warm и hit

За:

- уже был v1 bug с нестабильным order / generated `CS-*` IDs.
- текущий artifact key очень чувствительный.

Текущий `buildArtifactCacheKey` включает:

- `VerdictNamespace`
- `PoCPromptVersion`
- `ReportPromptVersion`
- `FixSpecVersion`
- ordered `FindingFingerprints`
- ordered `DispositionHashes`

`hashArtifactDisposition` включает:

- `FindingID`
- `Fingerprint`
- `Disposition`
- `Confidence`
- `Rationale`
- hash of `Evidence`

Любая разница в FindingID/order/rationale/evidence даст другой key.

Против / неизвестно:

- после fix `08a4bf4` order должен быть стабильнее.
- v3 hit имел full verdict cache hit, но мы не видим exact projected dispositions из hit artifacts, потому что run был cancelled до upload.

### Гипотеза 3: artifact key слишком хрупкий из-за LLM text fields

За:

- `Rationale` и `Evidence` являются частью disposition hash.
- Даже если verdict из cache, projected/generated state может отличаться в мелких полях, если где-то пересобирается, нормализуется или меняется order.

Против / неизвестно:

- verdict cache должен возвращать cached disposition data, но это не доказано по hit artifact because no artifact uploaded.

### Гипотеза 4: path/env contract mismatch

За:

- artifact cache dir uses `AITRIAGE_ARTIFACT_CACHE_DIR`, fallback `AITRIAGE_CACHE_DIR`.
- workflow passes both verdict and artifact dir as `.aitriage-cache`.

Против:

- verdict cache from same `.aitriage-cache` restored and worked.
- значит сама папка restore доступна.

### Гипотеза 5: Save и Restore используют разные logical namespace

За:

- namespace bumped across v1/v2/v3 due immutable cache.
- artifact key embeds verdict namespace from env/provider/model/version/policy/rules.

Против:

- warm/hit v3 used same GitHub cache key, same SHA/provider/model according to logs.

## Важная деталь про current code

`newArtifactBundleCache().load()` при отсутствии файла или пустых entries не печатает отдельный warning. Потом `Restore` просто дает:

```text
miss_reason = key_miss
```

Поэтому `key_miss` сейчас может означать два разных класса проблем:

1. файл cache физически не был загружен / не содержал entries;
2. файл был загружен, но конкретный computed key не совпал.

Существующий лог не различает эти случаи.

## Почему стало долго

На v3 hit run classification уже не тратила LLM по 212 findings. Но artifact cache не сработал, поэтому AITriage пошел в дорогие secondary stages:

- PoC Verification;
- Report generation;
- FixSpec generation.

Run был остановлен во время report generation. Если бы его не остановили, он мог снова дойти до fixspec и потратить еще больше tokens.

Иначе говоря: мы убрали повторное "думание по каждой finding", но не добились переиспользования "готового отчета/PoC/fixspec". Поэтому экономия есть только частичная, а цель artifact reuse пока не достигнута.

## Ошибки текущего агента

Это важно для следующего агента, чтобы не повторять.

- Я слишком долго продолжал remote validation через GitHub Actions.
- Я запускал warm/hit retries, когда нужно было раньше остановиться и расширить диагностику.
- Я обновлял Kaiten местами недостаточно аккуратно: один комментарий был на английском и не по ожидаемому стилю пользователя.
- Я не сразу синхронизировал шапку `artifact-reuse-cache.md` с фактом, что работа уже не planning only.
- Я не должен был продолжать запускать GHA после первых признаков, что artifact hit не доказан.

## Что нельзя утверждать

- Нельзя утверждать, что artifact reuse cache работает в реальном CI.
- Нельзя утверждать, что v3 hit был успешным artifact hit.
- Нельзя утверждать, что проблема точно в GitHub cache.
- Нельзя утверждать, что проблема точно в artifact key.
- Нельзя утверждать, что достаточно еще одного run, чтобы все стало ясно.

Доказано только:

- GitHub cache exact restore был;
- verdict cache hit 212/212 был;
- artifact cache hit не был;
- AITriage ушел в PoC/report generation;
- run был отменен.

## Текущая цель проекта

Финальная цель artifact reuse cache:

- при exact повторе same SHA/ref/provider/model/policy/rules/findings/verdicts восстановить `poc_results`, `report.md`, `fixspec.md`;
- не вызывать LLM для PoC/report/fixspec на full hit;
- не ухудшать security gate;
- не скрывать stale findings;
- не сохранять secrets/raw prompts;
- сохранять нормальные `summary.md`, `report.md`, `fixspec.md`, `triage-findings.json`;
- сделать telemetry достаточно явной, чтобы по artifact/logs было видно hit/miss/stores/skip reason.

## Текущее состояние работы

Локально/в ветках уже есть значительный pilot implementation.

Но remote Task 7 не закрыт:

- exact artifact hit не доказан;
- runtime/token reduction не доказаны;
- Xiaomi повтор не запускался после v3;
- final cleanup не делался;
- PR/merge readiness отсутствует.

Кодовая ветка содержит pilot image tag и workflow refs. Это нельзя слепо мержить в stable/main.

## Последняя команда пользователя перед этим handoff

Пользователь остановил работу и потребовал:

- не запускать GHA без явной команды;
- разобрать и анализировать логи;
- создать MD в корне проекта;
- выписать все попытки, результаты, ветки, карточки, что сделано и что хотим сделать;
- особенно честно описать текущую проблему, которую я не знаю как решать.

Этот файл создан именно под эту задачу.
