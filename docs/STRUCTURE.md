# AITriage — Структура проекта и план реализации

## Язык: Go

```
go 1.23+
module: github.com/cybertortuga/aitriage
```

---

## Структура директорий

```
aitriage/
├── cmd/
│   ├── aitriage/
│   │   └── main.go                  # Точка входа, CLI (cobra)
│   ├── test_rules/
│   │   └── main.go                  # Утилита для тестирования движка правил
│   └── test_triage/
│       └── main.go                  # Тестовый запуск триажа
│
├── internal/
│   ├── detector/
│   │   └── detector.go              # Авто-определение стека (Next.js, FastAPI, etc.)
│   │
│   ├── checker/
│   │   ├── checker.go               # Интерфейс Checker + типы (PRESENT/ABSENT/UNKNOWN)
│   │   ├── registry.go              # Реестр всех чеков по стекам
│   │   │
│   │   ├── universal/               # Чеки не привязанные к стеку
│   │   │   ├── lockfile.go           # P-11: Dependency Lockfile
│   │   │   └── secrets.go            # P-08: Secrets Management (.env in .gitignore)
│   │   │
│   │   ├── nextjs/                   # Маркеры для Next.js
│   │   │   ├── auth.go               # P-01: Authentication
│   │   │   ├── authz.go              # P-02: Authorization
│   │   │   ├── validation.go         # P-03: Input Validation
│   │   │   ├── ratelimit.go          # P-04: Rate Limiting
│   │   │   ├── errorhandling.go      # P-05: Error Handling
│   │   │   ├── headers.go            # P-06: Security Headers
│   │   │   ├── cors.go               # P-07: CORS Policy
│   │   │   ├── queries.go            # P-09: Parameterized Queries
│   │   │   ├── https.go              # P-10: HTTPS Enforcement
│   │   │   ├── logging.go            # P-12: Logging
│   │   │   └── csrf.go               # P-13: CSRF Protection
│   │   │
│   │   └── fastapi/                  # Маркеры для FastAPI
│   │       ├── auth.go
│   │       ├── authz.go
│   │       ├── validation.go
│   │       ├── ratelimit.go
│   │       ├── errorhandling.go
│   │       ├── headers.go
│   │       ├── cors.go
│   │       ├── queries.go
│   │       ├── https.go
│   │       ├── logging.go
│   │       └── csrf.go
│   │
│   ├── scanner/
│   │   └── scanner.go               # Оркестрация: detect stack → run checks → collect results
│   │
│   ├── scorer/
│   │   └── scorer.go                # Score + Grade (A-F)
│   │
│   └── reporter/
│       ├── reporter.go              # Интерфейс Reporter
│       ├── terminal.go              # Красивый вывод в терминал (цвета, emoji)
│       ├── json.go                  # JSON output
│       └── sarif.go                 # SARIF для GitHub Security tab
│
├── markers/                         # YAML-файлы с маркерами (опционально, для расширяемости)
│   ├── nextjs.yaml
│   └── fastapi.yaml
│
├── testdata/                        # Тестовые проекты для проверки чеков
│   ├── nextjs-secure/               # Проект где всё есть
│   ├── nextjs-insecure/             # Проект где ничего нет
│   ├── fastapi-secure/
│   └── fastapi-insecure/
│
├── .github/
│   └── workflows/
│       └── ci.yml                   # CI для самого AITriage
│
├── .goreleaser.yaml                 # Сборка бинарников для всех платформ
├── go.mod
├── go.sum
├── README.md
├── bin/                             # Бинарные файлы
├── docs/                            # Документация проекта
│   ├── design_mockups/              # Дизайн макеты и прототипы
│   ├── ARCHITECTURE.md
│   ├── DESIGN.md
│   ├── INTEGRATION.md
│   ├── PLAN.md
│   ├── REFACTORING_PLAN.md
│   ├── REDESIGN_PLAN.md
│   ├── ROADMAP_V2.md
│   ├── RULES_EXPANSION_PLAN.md
│   ├── STRUCTURE.md
│   ├── SUMMARY.md
│   ├── max.md
│   └── implementation_plan.md
└── reports/                         # Отчеты сканирования
    └── aitriage-report.html
```

---

## Ключевые интерфейсы

### CheckResult

```go
type Status string

const (
    Present Status = "PRESENT"
    Absent  Status = "ABSENT"
    Unknown Status = "UNKNOWN"
)

type CheckResult struct {
    ID          string   // "P-01"
    Name        string   // "Authentication"
    Status      Status   // PRESENT / ABSENT / UNKNOWN
    Evidence    string   // "next-auth detected in package.json"
    Suggestion  string   // "Consider: NextAuth.js, Clerk"
}
```

### Checker

```go
type Checker interface {
    ID() string
    Name() string
    Check(projectPath string) CheckResult
}
```

### Detector

```go
type Stack string

const (
    NextJS  Stack = "nextjs"
    FastAPI Stack = "fastapi"
    Express Stack = "express"
    Django  Stack = "django"
    UnknownStack Stack = "unknown"
)

func Detect(projectPath string) Stack
```

### Scanner (оркестратор)

```go
func Scan(projectPath string, opts ScanOptions) ScanReport

type ScanReport struct {
    Stack    Stack
    Results  []CheckResult
    Score    int           // 0-100
    Grade    string        // A/B/C/D/F
}
```

---

## Зависимости (минимум)

```
github.com/spf13/cobra       # CLI
github.com/fatih/color        # Цветной вывод в терминал
gopkg.in/yaml.v3              # Парсинг YAML конфигов
```

Больше ничего не нужно. Файлы, regex, grep — всё в стандартной библиотеке Go.

---

## Как работает один чек (пример)

### P-01 Authentication для Next.js

```go
// internal/checker/nextjs/auth.go

func (c *AuthChecker) Check(projectPath string) checker.CheckResult {
    // 1. Проверить package.json на auth-зависимости
    deps := []string{"next-auth", "@clerk/nextjs", "@auth/core", "lucia"}
    if found := findInPackageJSON(projectPath, deps); found != "" {
        return checker.CheckResult{
            Status:   checker.Present,
            Evidence: fmt.Sprintf("%s found in package.json", found),
        }
    }

    // 2. Проверить наличие middleware.ts с auth-паттернами
    if hasFileWithPattern(projectPath, "middleware.ts", authPatterns) {
        return checker.CheckResult{
            Status:   checker.Present,
            Evidence: "Auth logic detected in middleware.ts",
        }
    }

    // 3. Ничего не найдено
    return checker.CheckResult{
        Status:     checker.Absent,
        Suggestion: "Consider: NextAuth.js, Clerk, or custom middleware.ts",
    }
}
```

---

## Хелперы (internal/checker/)

Общие функции, которые используют все чеки:

```go
// Найти зависимость в package.json (Node.js стеки)
func findInPackageJSON(projectPath string, deps []string) string

// Найти зависимость в requirements.txt / pyproject.toml (Python стеки)
func findInPythonDeps(projectPath string, deps []string) string

// Проверить существование файла
func fileExists(projectPath string, relativePath string) bool

// Найти файлы по glob-паттерну
func findFiles(projectPath string, pattern string) []string

// Grep: найти regex в файлах
func grepFiles(projectPath string, pattern string, fileGlob string) []GrepMatch

// Проверить есть ли строка в .gitignore
func isInGitignore(projectPath string, entry string) bool
```

---

## Порядок реализации

### Шаг 1: Скелет

```
1. go mod init github.com/cybertortuga/aitriage
2. cmd/aitriage/main.go — cobra CLI с командой `scan`
3. Флаги: --path, --stack, --format (terminal/json)
4. Проверить что `aitriage scan .` запускается и печатает заглушку
```

### Шаг 2: Detector

```
1. internal/detector/detector.go
2. Detect() возвращает Stack по маркерам:
   - package.json + "next" в dependencies → NextJS
   - requirements.txt + "fastapi" → FastAPI
3. Тесты на testdata/
```

### Шаг 3: Хелперы

```
1. findInPackageJSON()
2. findInPythonDeps()
3. fileExists(), findFiles()
4. grepFiles()
5. isInGitignore()
6. Тесты на каждый хелпер
```

### Шаг 4: Универсальные чеки

```
1. P-08 Secrets Management — .env в .gitignore
2. P-11 Dependency Lockfile — package-lock.json / poetry.lock / etc.
Тесты.
```

### Шаг 5: Next.js чеки

```
По одному, каждый с тестом:
P-01 Auth → P-02 Authz → P-03 Validation → P-04 RateLimit →
P-05 ErrorHandling → P-06 Headers → P-07 CORS → P-09 Queries →
P-10 HTTPS → P-12 Logging → P-13 CSRF
```

### Шаг 6: FastAPI чеки

```
Аналогично шагу 5, но с Python-маркерами.
```

### Шаг 7: Scorer

```
1. internal/scorer/scorer.go
2. Score = (present / total) * 100
3. Grade: A/B/C/D/F
```

### Шаг 8: Reporter — Terminal

```
1. Красивый вывод с цветами и emoji
2. Группировка: ABSENT сверху, PRESENT снизу
3. Итог: Score + Grade
```

### Шаг 9: Reporter — JSON

```
1. Структурированный JSON для CI-пайплайнов
```

### Шаг 10: Сборка и релиз

```
1. .goreleaser.yaml — бинарники для linux/mac/windows
2. GitHub Releases
3. Homebrew formula (опционально)
```

---

## Тестирование

Каждый чек тестируется на двух тестовых проектах в `testdata/`:

- `testdata/nextjs-secure/` — минимальный Next.js проект со всеми практиками
- `testdata/nextjs-insecure/` — минимальный Next.js проект без ничего
- `testdata/fastapi-secure/` — аналогично
- `testdata/fastapi-insecure/` — аналогично

Тестовые проекты — не настоящие приложения. Достаточно файлов-заглушек: `package.json` с нужными зависимостями, `middleware.ts` с auth-паттерном, и т.д.

```go
func TestAuthChecker_Present(t *testing.T) {
    result := authChecker.Check("../../testdata/nextjs-secure")
    assert.Equal(t, checker.Present, result.Status)
}

func TestAuthChecker_Absent(t *testing.T) {
    result := authChecker.Check("../../testdata/nextjs-insecure")
    assert.Equal(t, checker.Absent, result.Status)
}
```

---

## CI для самого AITriage

```yaml
# .github/workflows/ci.yml
name: CI
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.23'
      - run: go test ./...
      - run: go vet ./...
      - run: go build ./cmd/aitriage
```
