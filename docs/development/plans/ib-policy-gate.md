# IB Policy Gate

## Обзор

Делаем `health_check` не просто score/breakdown, а полноценный уровень ИБ проекта:
можно ли пропускать проект через CI/CD по требованиям ИБ, почему можно/нельзя,
какой профиль требований применён, какие причины блокируют релиз.

Цель: единый deterministic policy/verdict слой поверх Health Check:

- `health_check.score/grade/breakdown` остаются уровнем состояния ИБ.
- Новый `health_check.verdict` отвечает на вопрос CI/CD: `passed: true/false`.
- CLI `scan`, CLI `agent`, GitHub Action, JSON/API и summary используют один
  verdict вместо разрозненной логики.
- Старые поля `security_score/security_grade/has_critical_failures` остаются для
  совместимости.
- Изменения делаются точечно, без массового переименования публичного API.

Research sources:

- OWASP ASVS: стандарт задаёт уровни security verification и может быть
  основой профилей требований. Источник:
  https://owasp.org/www-project-application-security-verification-standard/
- OWASP ASVS usage: Level 1 minimum, Level 2 for sensitive/business apps, Level 3
  for high-value/high-assurance systems. Источник:
  https://github.com/OWASP/ASVS/blob/master/4.0/en/0x03-Using-ASVS.md
- NIST SSDF: secure software practices должны быть risk-based и adaptable, а не
  просто чеклистом. Источник: https://csrc.nist.gov/projects/ssdf
- GitHub Actions: exit code `0` means success, non-zero means failure and blocks
  dependent work. Источник:
  https://docs.github.com/en/actions/how-tos/create-and-publish-actions/set-exit-codes
- GitHub Actions workflow commands: `::error`, `::warning`, `::notice` создают
  аннотации в CI. Источник:
  https://docs.github.com/en/actions/reference/workflows-and-actions/workflow-commands
- GitHub SARIF/code scanning: SARIF должен иметь стабильные `ruleId`,
  locations и severity metadata; GitHub использует fingerprints для дедупа.
  Источник:
  https://docs.github.com/en/code-security/reference/code-scanning/sarif-files/sarif-support

## Затронутые файлы и области

Планируемые изменения:

- `internal/report/healthcheck/healthcheck.go`
- `internal/report/healthcheck/core_adapter.go`
- `internal/report/healthcheck/healthcheck_test.go`
- возможно новый файл `internal/report/healthcheck/policy.go`
- возможно новый файл `internal/report/healthcheck/policy_test.go`
- `internal/config/config.go`
- `internal/scanner/scanner.go`
- `internal/agent/graph/state.go`
- `internal/agent/graph/orchestrator.go`
- `cmd/aitriage/scan.go`
- `cmd/aitriage/agent.go`
- `cmd/aitriage/init.go`
- `action.yml`
- `entrypoint.sh`
- `internal/server/server.go`
- `web/src/types.ts`
- `web/src/services/securityService.ts`
- `README.md`
- `docs/aitriage.yaml.example`
- tests under `cmd/aitriage`, `internal/scanner`, `internal/report/healthcheck`,
  `internal/agent/graph`

Read-only audited areas that should not be changed unless a criterion requires it:

- `internal/scanner/sarif.go`
- `internal/report/reporter/reporter.go`
- `internal/ui/tui/update.go`
- `internal/ui/tui/view.go`
- `internal/engine/history/history.go`
- `internal/telemetry/*`
- `internal/server/repositories/metrics_repo.go`

## Задача 1: Ввести модель IB policy verdict

### Описание

Добавить в `healthcheck` явную модель policy gate:

- профиль требований (`baseline`, `standard`, `strict`)
- pass/fail verdict
- machine-readable blocking reasons
- настройки порогов: минимальный score, max critical/high/medium,
  block-on sources/classes
- default policy должна сохранять текущее поведение `scan --fail-on critical`
  как близкое к baseline, но verdict должен быть явным.

Предлагаемый публичный JSON внутри `health_check`:

```json
{
  "score": 82,
  "grade": "B",
  "has_critical_failures": true,
  "breakdown": {},
  "policy": {
    "profile": "baseline",
    "minimum_score": 0,
    "fail_on": "critical",
    "max_critical": 0,
    "max_high": -1,
    "max_medium": -1,
    "block_sources": ["gitleaks"],
    "block_classes": []
  },
  "verdict": {
    "passed": false,
    "status": "failed",
    "summary": "Blocked by 1 active CRITICAL finding",
    "blocking_reasons": [
      {
        "code": "critical_findings",
        "message": "Active CRITICAL findings exceed allowed threshold",
        "severity": "CRITICAL",
        "count": 1,
        "threshold": 0
      }
    ]
  }
}
```

### Затронутые файлы

- `internal/report/healthcheck/healthcheck.go`
- `internal/report/healthcheck/policy.go`
- `internal/report/healthcheck/policy_test.go`
- `internal/report/healthcheck/healthcheck_test.go`

### Подзадачи

- [x] Добавить типы `Policy`, `PolicyProfile`, `Verdict`, `BlockingReason`.
- [x] Добавить `Policy` и `Verdict` в `healthcheck.Result`.
- [x] Реализовать `EvaluatePolicy(result, policy)` без зависимости от CLI.
- [x] Добавить default policies: `baseline`, `standard`, `strict`.
- [x] Поддержать `fail-on`: `critical`, `any`, `never`.
- [x] Поддержать score gate как часть verdict, а не отдельную CLI-логику.
- [x] Добавить tests для pass/fail по critical/high/any/score/source/class.

### Критерии приёмки

- [x] `healthcheck.Evaluate` возвращает `Result` с заполненным default verdict.
- [x] Verdict fail содержит хотя бы одну blocking reason.
- [x] Ignored/FP findings не создают blocking reasons.
- [x] `fail-on never` может дать `passed=true`, но score/breakdown остаются
      честными.
- [x] Все новые типы сериализуются в стабильный JSON.

### Риски

- Можно случайно сломать совместимость JSON, если переименовать старые поля.
- Можно сделать слишком мягкий default и перестать блокировать critical.
- Можно сделать слишком жёсткий default и начать валить существующие CI без
  явного opt-in.

## Задача 2: Расширить конфиг требований ИБ

### Описание

Расширить `.aitriage.yaml` так, чтобы требования ИБ задавались явно, но старые
поля продолжили работать:

```yaml
health_check:
  profile: baseline
  fail_on: critical
  minimum_score: 70
  max_critical: 0
  max_high: -1
  max_medium: -1
  block_sources:
    - gitleaks
  block_classes: []
```

Backcompat:

- `strict_mode: true` мапится в policy `fail_on: any`, если новый блок не задан.
- `fail_score` мапится в `minimum_score`, если новый блок не задан.
- CLI flags могут override только явные CLI параметры.

### Затронутые файлы

- `internal/config/config.go`
- `cmd/aitriage/init.go`
- `docs/aitriage.yaml.example`
- `README.md`

### Подзадачи

- [x] Добавить `HealthCheckPolicyConfig` в `config.Config`.
- [x] Реализовать преобразование config -> `healthcheck.Policy`.
- [x] Сохранить поддержку `strict_mode` и `fail_score`.
- [x] Обновить generated config в `aitriage init`.
- [x] Обновить docs/example config.

### Критерии приёмки

- [x] Старый `.aitriage.yaml` без `health_check` работает как раньше.
- [x] Новый `health_check` блок управляет verdict.
- [x] `strict_mode` и `fail_score` не конфликтуют с новым блоком, а имеют
      понятный приоритет.
- [x] YAML parsing tests покрывают старый и новый формат.

### Риски

- Конфликт приоритетов CLI/config может сделать CI непредсказуемым.
- Неверные default значения (`0` vs `-1`) могут изменить смысл threshold.

## Задача 3: Подключить verdict к scanner.Scan

### Описание

`scanner.Scan` должен возвращать `ScanReport.HealthCheck` уже с policy verdict,
а не только score. Это главный deterministic путь для CI/CD.

### Затронутые файлы

- `internal/scanner/scanner.go`
- `internal/scanner/scanner_test.go`
- `internal/report/healthcheck/core_adapter.go`

### Подзадачи

- [x] Собрать policy из `ws.Config` и передать в Health Check.
- [x] После baseline/diff filtering в CLI пересчитывать verdict или явно
      решить, что baseline фильтр должен происходить до Health Check.
- [x] Уточнить поведение `--baseline`: legacy debt не должен создавать blocking
      reasons.
- [x] Добавить scanner tests для `health_check.verdict`.

### Критерии приёмки

- [x] `ScanReport.HealthCheck.Verdict.Passed` отражает требования ИБ.
- [x] `ScanReport.HasCriticalFailures` остаётся backcompat alias для active
      critical/high.
- [x] `security_score/security_grade` продолжают совпадать с
      `health_check.score/grade`.

### Риски

- Сейчас baseline filtering происходит после `scanner.Scan`; если не
  пересчитать verdict после baseline, CI может блокировать legacy debt.

## Задача 4: Подключить verdict к CLI scan и GitHub Action

### Описание

`aitriage scan` должен принимать решение exit code только по
`report.HealthCheck.Verdict.Passed`.

CLI flags должны стать policy inputs:

- `--fail-on`
- `--fail-score`
- новый флаг `--policy` или `--health-profile` (выбрать одно название перед
  реализацией)

GitHub summary должен показывать:

- PASS/FAIL
- policy profile
- score/grade
- blocking reasons
- active/ignored/deduped breakdown

### Затронутые файлы

- `cmd/aitriage/scan.go`
- `action.yml`
- `entrypoint.sh`
- `.github/workflows/test-action.yml`
- `README.md`

### Подзадачи

- [x] Добавить CLI flag для policy profile.
- [x] Убрать ручную `shouldFail` логику из `scan.go` в пользу verdict.
- [x] Сохранить `--fail-on never` как explicit non-blocking mode.
- [x] Обновить GitHub Action input для policy profile.
- [x] Обновить `$GITHUB_STEP_SUMMARY`.
- [x] Добавить/обновить cmd tests.

### Критерии приёмки

- [x] Exit code `0` если verdict passed.
- [x] Exit code non-zero если verdict failed.
- [x] В failure output есть blocking reasons.
- [x] GitHub Action может настроить `fail-score` и policy profile без `args`.

### Риски

- GitHub Action default может измениться для пользователей. Default должен
  остаться максимально совместимым.

## Задача 5: Подключить verdict к agent path

### Описание

`aitriage agent` должен использовать тот же policy/verdict слой после AI FP/TP
классификации. AI advisor остаётся non-blocking по default, но если пользователь
включил gate, он должен видеть те же blocking reasons.

### Затронутые файлы

- `cmd/aitriage/agent.go`
- `internal/agent/graph/state.go`
- `internal/agent/graph/orchestrator.go`
- `internal/agent/graph/healthcheck_test.go`

### Подзадачи

- [x] Передать policy config в `AgentState` или в `computeHealthCheck`.
- [x] Убрать ручную `agentShouldFail` логику в пользу verdict.
- [x] Обновить agent report metadata: Health Check verdict + blocking reasons.
- [x] Добавить tests для agent verdict с FP и без dispositions.

### Критерии приёмки

- [x] Agent Health Check использует тот же policy engine.
- [x] FP не блокируют.
- [x] Findings без AI disposition остаются active/conservative.
- [x] Default `agent --no-chat` без gate не ломает AI advisor workflows.

### Риски

- AI classification недетерминированна; hard gate по agent должен быть opt-in.

## Задача 6: API/Web additive propagation

### Описание

API и web-типы должны видеть verdict, но существующие UI не должны ломаться.

### Затронутые файлы

- `internal/server/server.go`
- `web/src/types.ts`
- `web/src/services/securityService.ts`
- возможно страницы `web/src/pages/TerminalPage.tsx`,
  `web/src/pages/DashboardPage.tsx`

### Подзадачи

- [x] Убедиться, что `/scan` отдаёт `health_check.verdict`.
- [x] Обновить TypeScript types.
- [x] Минимально отобразить pass/fail в Terminal/scan UI, если это не требует
      большого frontend refactor.
- [x] Не менять enterprise metrics formula без отдельной задачи.

### Критерии приёмки

- [x] Старые UI поля `security_score/security_grade` работают.
- [x] Новые клиенты могут читать `health_check.verdict`.
- [x] Web build проходит.

### Риски

- Enterprise metrics сейчас считают portfolio risk отдельно; нельзя молча
  заменить их scan-level Health Check.

## Задача 7: Документация и миграция

### Описание

Документировать Health Check как ИБ gate:

- что значит score
- что значит verdict
- какие есть profiles
- как настроить CI/CD
- как migration работает со старыми `fail_score`/`strict_mode`

### Затронутые файлы

- `README.md`
- `docs/aitriage.yaml.example`
- `docs/development/healthcheck_ci_cd_plan.md`
- возможно новый раздел в `docs/INTEGRATION.md`

### Подзадачи

- [x] Обновить README CI/CD example.
- [x] Обновить config example.
- [x] Добавить migration notes.
- [x] Обновить текущий план-файл финальным отчётом.

### Критерии приёмки

- [x] Документация объясняет, что `security_score` — compatibility alias.
- [x] Документация объясняет, что `health_check.verdict` решает pass/fail.
- [x] Нет противоречия между action inputs, CLI flags и YAML config.

### Риски

- Старые historical docs могут всё ещё упоминать scorer/SecurityScore. Не
  править всё подряд, только активные пользовательские docs.

## Задача 8: Финальная проверка

### Описание

Проверить весь путь end-to-end.

### Затронутые файлы

- все изменённые файлы

### Подзадачи

- [x] `go test ./internal/report/healthcheck ./internal/scanner ./internal/agent/graph ./internal/server ./cmd/aitriage`
- [x] `go test -p 1 ./...`
- [x] `go vet ./...`
- [x] `sh -n entrypoint.sh scripts/entrypoint.sh`
- [x] `npm ci && npm run build`
- [x] удалить `web/node_modules` и `web/dist` после web build, если они появились
      локально
- [x] smoke JSON: `aitriage scan ... --format json --no-history --no-summary`
- [x] smoke SARIF: `aitriage scan ... --format sarif --out /tmp/... --no-history --no-summary`
- [x] проверить exit codes для pass/fail policy fixture
- [x] проверить `git status --short`

### Критерии приёмки

- [x] Все проверки зелёные.
- [x] В финальном JSON есть `health_check.verdict`.
- [x] CI exit code управляется verdict.
- [x] Нет новых generated artifacts в git status.

### Риски

- `npm ci` может сообщить dependency audit vulnerabilities. Не запускать
  `npm audit fix` в рамках этой задачи без отдельного решения.

## Риски и зависимости

- В рабочей копии уже есть незакоммиченные изменения предыдущего Health Check
  этапа. Их нельзя откатывать без явного разрешения.
- Самый высокий риск: baseline filtering сейчас происходит после Health Check;
  для полноценного policy gate нужно пересчитать Health Check/verdict после
  baseline или перенести baseline до оценки.
- `scan` сейчас core-only, `agent` multi-source. Решение: policy engine общий,
  источники остаются как есть до отдельного решения о multi-source scan.
- `agent` использует AI dispositions; hard gate по agent должен быть opt-in.
- GitHub Action failure должен быть простым non-zero exit code, но summary и
  annotations должны объяснять blocking reasons до выхода.
- Не вводить внешнюю policy dependency без необходимости. Go stdlib + текущий
  YAML parser достаточно для deterministic local policy.

## Финальный отчёт

Выполнено:

- Добавлен единый `health_check.policy` + `health_check.verdict` слой для ответа
  “можно ли пропускать проект по требованиям ИБ”.
- Добавлены профили `baseline`, `standard`, `strict`, `fail_on` режимы
  `critical`, `any`, `never`, score gate, severity thresholds,
  `block_sources`, `block_classes`.
- `.aitriage.yaml` получил новый `health_check:` блок; legacy
  `strict_mode`/`fail_score` сохранены как fallback.
- `scanner.Scan`, `aitriage scan`, GitHub Action и `agent` используют один
  verdict engine вместо ручной разрозненной логики.
- `--baseline` теперь пересчитывает Health Check/verdict после фильтрации, так
  что legacy debt не создаёт blocking reasons.
- `agent` остаётся advisory/non-blocking по default, но config/CLI gate включает
  тот же policy engine после AI FP/TP dispositions.
- API/Web получили additive `health_check.policy/verdict`; старые
  `security_score/security_grade` оставлены.
- README, config example и migration notes обновлены.

Проверки:

- `go test ./internal/report/healthcheck ./internal/scanner ./internal/agent/graph ./internal/server ./cmd/aitriage`
- `go test -p 1 ./...`
- `go vet ./...`
- `sh -n entrypoint.sh scripts/entrypoint.sh`
- `npm ci && npm run build`
- smoke JSON: exit `0`, `health_check.verdict` присутствует
- smoke SARIF: exit `0`, SARIF `2.1.0` файл создан
- smoke strict policy: exit `1`, blocking reasons напечатаны

Замечания:

- `npm ci` сообщает 3 audit vulnerabilities (1 low, 2 high). `npm audit fix` не
  запускался, потому что это отдельное изменение зависимостей.
- `web/node_modules` и `web/dist` после build удалены.
