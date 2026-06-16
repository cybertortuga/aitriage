# AITriage — Полный план реализации
## AI Security Agent для Dodo Pizza и внутренних проектов

> **ВАЖНО ДЛЯ AI-АГЕНТА:** Следуй плану строго. Не добавляй ничего от себя.
> Каждый пункт — конкретное действие в конкретном файле. Не переходи к следующей фазе
> пока не проставлены все `[x]` в текущей. Если что-то непонятно — спроси, не угадывай.

---

## Архитектура целиком

```
aitriage scan ./         → Go-движок (AST + Entropy + Security), без LLM, для CI/CD
aitriage agent ./        → AI-агент (оркестрирует сканеры + LLM анализ + Q&A)
aitriage serve           → MCP-сервер (для Claude Code / Cursor / любого MCP-клиента)
```

```
┌─────────────────────────────────────────────────────┐
│                    aitriage agent                   │
│           (LLM-powered, натуральный язык)           │
│  ┌──────────┐  ┌──────────┐  ┌──────────────────┐  │
│  │  scan    │  │ external │  │ infra+NFR+git+   │  │
│  │ (Go AST) │  │ scanners │  │ diagram          │  │
│  └──────────┘  └──────────┘  └──────────────────┘  │
│                      ↓                              │
│              LLM Layer (AI-lab agnostic)            │
│        OpenAI │ Anthropic │ Ollama │ Groq           │
│                      ↓                              │
│    Анализ → Отчёт → Fix Plan → Q&A консультация    │
└─────────────────────────────────────────────────────┘
```

---

## Фаза 0: Подготовка кодовой базы
> Цель: сделать существующий код готовым к расширению. Не менять логику, только структуры и сигнатуры.

### 0.1 — Обновить ScanReport в `internal/scanner/scanner.go`

- [ ] Открыть файл `internal/scanner/scanner.go`
- [ ] Найти struct `ScanReport`
- [x] Добавить json-теги к **каждому** существующему полю (не удалять поля, только добавить теги)
- [x] Добавить **новые поля** в конец struct:
  ```go
  ProjectPath  string        `json:"project_path"`
  TotalFiles   int           `json:"total_files"`
  RulesApplied int           `json:"rules_applied"`
  ScanDuration time.Duration `json:"scan_duration_ms"`
  ```
- [x] Добавить import `"time"` если его нет
- [x] Добавить метод после объявления struct:
  ```go
  func (r ScanReport) ToJSON() ([]byte, error) {
      return json.Marshal(r)
  }
  ```
- [x] Добавить import `"encoding/json"` если его нет
- [x] Запустить `go build ./...` — должно компилироваться без ошибок

### 0.2 — Обновить CheckResult в `internal/core/context.go`

- [x] Открыть файл `internal/core/context.go`
- [x] Найти struct `CheckResult`
- [x] Добавить json-теги к **каждому** существующему полю (не удалять поля)
- [x] Добавить **новые поля** в конец struct:
  ```go
  OWASPMapping   string   `json:"owasp_mapping,omitempty"`
  ReasoningChain []string `json:"reasoning_chain,omitempty"`
  ```
- [x] Запустить `go build ./...` — должно компилироваться без ошибок

### 0.3 — Добавить context.Context в scanner.Scan()

- [x] Открыть файл `internal/scanner/scanner.go`
- [x] Найти функцию `Scan()` (или аналог — основную точку входа сканера)
- [x] Изменить сигнатуру: добавить `ctx context.Context` первым параметром
- [x] Изменить возврат: вместо `ScanReport` вернуть `(ScanReport, error)`
- [x] Добавить import `"context"` если его нет
- [x] Заменить все `os.Exit(...)` внутри функции на `return ScanReport{}, fmt.Errorf("...")`
- [x] Открыть файл `cmd/aitriage/scan.go`
- [x] Обновить вызов `scanner.Scan(...)` — добавить `context.Background()` первым аргументом, обработать возвращаемую ошибку
- [x] Запустить `go build ./...` — должно компилироваться без ошибок

### 0.4 — Добавить версионирование

- [x] Открыть файл `cmd/aitriage/root.go`
- [x] Добавить переменную в package-scope:
  ```go
  var Version = "dev"
  ```
- [x] Создать файл `cmd/aitriage/version.go` со следующим содержимым:
  ```go
  package main

  import (
      "fmt"
      "github.com/spf13/cobra"
  )

  var versionCmd = &cobra.Command{
      Use:   "version",
      Short: "Print AITriage version",
      Run: func(cmd *cobra.Command, args []string) {
          fmt.Printf("aitriage %s\n", Version)
      },
  }

  func init() {
      rootCmd.AddCommand(versionCmd)
  }
  ```
- [x] Запустить `go build ./...` — должно компилироваться без ошибок
- [x] Запустить `go run ./cmd/aitriage version` — должно вывести `aitriage dev`

---

## Фаза 1: MCP-сервер
> Цель: запустить `aitriage serve` чтобы Claude Code и Cursor могли вызывать AITriage как инструмент.

### 1.1 — Добавить зависимость Go MCP SDK

- [x] Выполнить: `go get github.com/modelcontextprotocol/go-sdk@latest`
- [x] Выполнить: `go mod tidy`
- [x] Убедиться что `go.mod` содержит строку с `modelcontextprotocol/go-sdk`

### 1.2 — Создать команду `serve`

- [x] Создать файл `cmd/aitriage/serve.go` со следующим содержимым:
  ```go
  package main

  import (
      "context"
      "github.com/spf13/cobra"
      mcpserver "github.com/cybertortuga/aitriage/internal/mcp"
  )

  var (
      serveTransport string
      servePort      int
  )

  var serveCmd = &cobra.Command{
      Use:   "serve",
      Short: "Start AITriage as an MCP server",
      Long:  "Expose AITriage security tools via Model Context Protocol for Claude Code, Cursor, etc.",
      RunE:  runServe,
  }

  func init() {
      rootCmd.AddCommand(serveCmd)
      serveCmd.Flags().StringVar(&serveTransport, "transport", "stdio", "Transport type: stdio or sse")
      serveCmd.Flags().IntVar(&servePort, "port", 8080, "Port for SSE transport")
  }

  func runServe(cmd *cobra.Command, args []string) error {
      srv := mcpserver.NewServer(Version)
      return srv.Run(context.Background(), serveTransport, servePort)
  }
  ```

### 1.3 — Создать директорию и файл сервера MCP

- [x] Создать директорию `internal/mcp/`
- [x] Создать файл `internal/mcp/server.go`:
  ```go
  package mcp

  import (
      "context"
      "fmt"

      "github.com/modelcontextprotocol/go-sdk/mcp"
  )

  type Server struct {
      version string
      srv     *mcp.Server
  }

  func NewServer(version string) *Server {
      s := &Server{version: version}
      s.srv = mcp.NewServer(&mcp.Implementation{
          Name:    "aitriage",
          Version: version,
      }, nil)
      s.registerTools()
      s.registerResources()
      return s
  }

  func (s *Server) Run(ctx context.Context, transport string, port int) error {
      switch transport {
      case "stdio":
          return mcp.NewStdioTransport().Run(ctx, s.srv)
      case "sse":
          return fmt.Errorf("SSE transport not yet implemented")
      default:
          return fmt.Errorf("unknown transport: %s", transport)
      }
  }

  func (s *Server) registerTools() {
      registerScanTool(s.srv)
      registerSecretsTool(s.srv)
      registerEntropyCheckTool(s.srv)
      registerArchitectureTool(s.srv)
      registerFixPlanTool(s.srv)
      registerScannersListTool(s.srv)
  }

  func (s *Server) registerResources() {
      registerPlaybookResource(s.srv)
  }
  ```

### 1.4 — Создать tool `aitriage_scan`

- [x] Создать файл `internal/mcp/tools_scan.go`:
  ```go
  package mcp

  import (
      "context"
      "encoding/json"
      "fmt"

      "github.com/modelcontextprotocol/go-sdk/mcp"
      "github.com/cybertortuga/aitriage/internal/scanner"
  )

  type scanInput struct {
      Path          string `json:"path"`
      Stack         string `json:"stack,omitempty"`
      UniversalOnly bool   `json:"universal_only,omitempty"`
  }

  func registerScanTool(srv *mcp.Server) {
      mcp.AddTool(srv, &mcp.Tool{
          Name:        "aitriage_scan",
          Description: "Run a full deterministic security scan on a project directory. Uses AST analysis, Shannon Entropy for secrets, and Entropy anomaly detection. No LLM required. Returns structured JSON report.",
          InputSchema: mcp.MustParseInputSchema(`{
              "type": "object",
              "properties": {
                  "path": {"type": "string", "description": "Absolute path to the project directory"},
                  "stack": {"type": "string", "description": "Force stack: nextjs, fastapi, go, etc. Leave empty for auto-detect"},
                  "universal_only": {"type": "boolean", "description": "Run only universal checks, skip stack-specific rules"}
              },
              "required": ["path"]
          }`),
      }, func(ctx context.Context, req *mcp.CallToolRequest, input scanInput) (*mcp.CallToolResult, error) {
          opts := scanner.ScanOptions{
              ForceStack:    input.Stack,
              UniversalOnly: input.UniversalOnly,
          }
          report, err := scanner.Scan(ctx, input.Path, opts)
          if err != nil {
              return nil, fmt.Errorf("scan failed: %w", err)
          }
          data, err := json.Marshal(report)
          if err != nil {
              return nil, err
          }
          return mcp.NewToolResultText(string(data)), nil
      })
  }
  ```

### 1.5 — Создать tool `aitriage_secrets`

- [x] Создать файл `internal/mcp/tools_secrets.go`:
  ```go
  package mcp

  import (
      "context"
      "encoding/json"
      "fmt"

      "github.com/modelcontextprotocol/go-sdk/mcp"
      "github.com/cybertortuga/aitriage/internal/scanner"
      "github.com/cybertortuga/aitriage/internal/core"
  )

  type secretsInput struct {
      Path string `json:"path"`
  }

  type secretsResult struct {
      Found   []core.CheckResult `json:"found"`
      Count   int                `json:"count"`
      Summary string             `json:"summary"`
  }

  func registerSecretsTool(srv *mcp.Server) {
      mcp.AddTool(srv, &mcp.Tool{
          Name:        "aitriage_secrets",
          Description: "Scan for hardcoded secrets using Shannon Entropy analysis. Finds API keys, tokens, passwords even with non-obvious variable names. Returns only secret-related findings.",
          InputSchema: mcp.MustParseInputSchema(`{
              "type": "object",
              "properties": {
                  "path": {"type": "string", "description": "Absolute path to the project directory"}
              },
              "required": ["path"]
          }`),
      }, func(ctx context.Context, req *mcp.CallToolRequest, input secretsInput) (*mcp.CallToolResult, error) {
          report, err := scanner.Scan(ctx, input.Path, scanner.ScanOptions{})
          if err != nil {
              return nil, fmt.Errorf("scan failed: %w", err)
          }
          var secrets []core.CheckResult
          for _, r := range report.Results {
              if r.ID == "ENTROPY-SECRET" {
                  secrets = append(secrets, r)
              }
          }
          res := secretsResult{
              Found:   secrets,
              Count:   len(secrets),
              Summary: fmt.Sprintf("Found %d potential secrets via Shannon Entropy analysis", len(secrets)),
          }
          data, _ := json.Marshal(res)
          return mcp.NewToolResultText(string(data)), nil
      })
  }
  ```

### 1.6 — Создать tool `aitriage_entropy_check`

- [x] Создать файл `internal/mcp/tools_entropy.go`:
  ```go
  package mcp

  import (
      "context"
      "encoding/json"
      "fmt"

      "github.com/modelcontextprotocol/go-sdk/mcp"
      "github.com/cybertortuga/aitriage/internal/scanner"
      "github.com/cybertortuga/aitriage/internal/core"
  )

  type entropyInput struct {
      Path string `json:"path"`
  }

  type entropyResult struct {
      Score    int                `json:"security_score"`
      Grade    string             `json:"security_grade"`
      Issues   []core.CheckResult `json:"issues"`
      Count    int                `json:"entropy_issue_count"`
      Summary  string             `json:"summary"`
  }

  func registerEntropyCheckTool(srv *mcp.Server) {
      mcp.AddTool(srv, &mcp.Tool{
          Name:        "aitriage_entropy_check",
          Description: "Check for AI-generated code quality issues: chat residue in comments, missing error handling, God Files (>1500 lines), TODO stubs, and .cursorrules manipulation attempts.",
          InputSchema: mcp.MustParseInputSchema(`{
              "type": "object",
              "properties": {
                  "path": {"type": "string", "description": "Absolute path to the project directory"}
              },
              "required": ["path"]
          }`),
      }, func(ctx context.Context, req *mcp.CallToolRequest, input entropyInput) (*mcp.CallToolResult, error) {
          report, err := scanner.Scan(ctx, input.Path, scanner.ScanOptions{})
          if err != nil {
              return nil, fmt.Errorf("scan failed: %w", err)
          }
          var entropyIssues []core.CheckResult
          for _, r := range report.Results {
              if len(r.ID) >= 5 && r.ID[:5] == "ENTR-" {
                  entropyIssues = append(entropyIssues, r)
              }
          }
          res := entropyResult{
              Score:   report.SecurityScore,
              Grade:   report.SecurityGrade,
              Issues:  entropyIssues,
              Count:   len(entropyIssues),
              Summary: fmt.Sprintf("Security Grade: %s (%d/100). Found %d AI-code issues.", report.SecurityGrade, report.SecurityScore, len(entropyIssues)),
          }
          data, _ := json.Marshal(res)
          return mcp.NewToolResultText(string(data)), nil
      })
  }
  ```

### 1.7 — Создать tool `aitriage_architecture`

- [x] Создать файл `internal/mcp/tools_architecture.go`:
  ```go
  package mcp

  import (
      "context"
      "encoding/json"
      "fmt"
      "os"
      "path/filepath"

      "github.com/modelcontextprotocol/go-sdk/mcp"
      "github.com/cybertortuga/aitriage/internal/detector"
      "github.com/cybertortuga/aitriage/internal/core"
  )

  type archInput struct {
      Path string `json:"path"`
  }

  type archResult struct {
      Stacks          []string          `json:"stacks"`
      TotalFiles      int               `json:"total_files"`
      FilesByExt      map[string]int    `json:"files_by_extension"`
      KeyFilesPresent map[string]bool   `json:"key_files_present"`
      Summary         string            `json:"summary"`
  }

  func registerArchitectureTool(srv *mcp.Server) {
      mcp.AddTool(srv, &mcp.Tool{
          Name:        "aitriage_architecture",
          Description: "Analyze project structure: detect tech stacks, count files by extension, check for key files (Dockerfile, docker-compose.yml, .env, Makefile, nginx.conf, terraform/*.tf, go.mod, package.json, requirements.txt). Call this FIRST before any scan.",
          InputSchema: mcp.MustParseInputSchema(`{
              "type": "object",
              "properties": {
                  "path": {"type": "string", "description": "Absolute path to the project directory"}
              },
              "required": ["path"]
          }`),
      }, func(ctx context.Context, req *mcp.CallToolRequest, input archInput) (*mcp.CallToolResult, error) {
          ws, err := core.NewWorkspace(input.Path)
          if err != nil {
              return nil, fmt.Errorf("failed to read workspace: %w", err)
          }
          stacks := detector.DetectProjects(ws)
          byExt := make(map[string]int)
          for _, f := range ws.Files {
              ext := filepath.Ext(f.Path)
              byExt[ext]++
          }
          keyFiles := map[string]bool{
              "Dockerfile":          fileExists(filepath.Join(input.Path, "Dockerfile")),
              "docker-compose.yml":  fileExists(filepath.Join(input.Path, "docker-compose.yml")),
              ".env":                fileExists(filepath.Join(input.Path, ".env")),
              ".env.example":        fileExists(filepath.Join(input.Path, ".env.example")),
              "Makefile":            fileExists(filepath.Join(input.Path, "Makefile")),
              "nginx.conf":          fileExists(filepath.Join(input.Path, "nginx.conf")),
              "go.mod":              fileExists(filepath.Join(input.Path, "go.mod")),
              "package.json":        fileExists(filepath.Join(input.Path, "package.json")),
              "requirements.txt":    fileExists(filepath.Join(input.Path, "requirements.txt")),
              "terraform":           dirExists(filepath.Join(input.Path, "terraform")),
          }
          stackNames := make([]string, 0, len(stacks))
          for _, s := range stacks {
              stackNames = append(stackNames, string(s))
          }
          res := archResult{
              Stacks:          stackNames,
              TotalFiles:      len(ws.Files),
              FilesByExt:      byExt,
              KeyFilesPresent: keyFiles,
              Summary:         fmt.Sprintf("Detected stacks: %v. Total files: %d.", stackNames, len(ws.Files)),
          }
          data, _ := json.Marshal(res)
          return mcp.NewToolResultText(string(data)), nil
      })
  }

  func fileExists(path string) bool {
      _, err := os.Stat(path)
      return err == nil
  }

  func dirExists(path string) bool {
      info, err := os.Stat(path)
      return err == nil && info.IsDir()
  }
  ```

### 1.8 — Создать tool `list_available_scanners`

- [x] Создать файл `internal/mcp/tools_scanners.go`:
  ```go
  package mcp

  import (
      "context"
      "encoding/json"
      "os/exec"

      "github.com/modelcontextprotocol/go-sdk/mcp"
  )

  type scannersResult struct {
      Semgrep  bool `json:"semgrep"`
      Gitleaks bool `json:"gitleaks"`
      Trivy    bool `json:"trivy"`
      Bandit   bool `json:"bandit"`
  }

  func registerScannersListTool(srv *mcp.Server) {
      mcp.AddTool(srv, &mcp.Tool{
          Name:        "list_available_scanners",
          Description: "Check which external security scanners are installed and available in PATH. ALWAYS call this before calling run_semgrep, run_gitleaks, run_trivy, or run_bandit to avoid errors.",
          InputSchema: mcp.MustParseInputSchema(`{"type": "object", "properties": {}}`),
      }, func(ctx context.Context, req *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, error) {
          res := scannersResult{
              Semgrep:  isInstalled("semgrep"),
              Gitleaks: isInstalled("gitleaks"),
              Trivy:    isInstalled("trivy"),
              Bandit:   isInstalled("bandit"),
          }
          data, _ := json.Marshal(res)
          return mcp.NewToolResultText(string(data)), nil
      })
  }

  func isInstalled(name string) bool {
      _, err := exec.LookPath(name)
      return err == nil
  }
  ```

### 1.9 — Создать tool `generate_fix_plan` (заглушка, реализация в Фазе 6)

- [x] Создать файл `internal/mcp/tools_fixplan.go`:
  ```go
  package mcp

  import (
      "context"
      "fmt"

      "github.com/modelcontextprotocol/go-sdk/mcp"
      "github.com/cybertortuga/aitriage/internal/scanner"
      "github.com/cybertortuga/aitriage/internal/remedy"
  )

  type fixPlanInput struct {
      Path string `json:"path"`
  }

  func registerFixPlanTool(srv *mcp.Server) {
      mcp.AddTool(srv, &mcp.Tool{
          Name:        "generate_fix_plan",
          Description: "Scan the project and generate a structured fix plan with actionable prompts for each finding. Output is a markdown document ready to paste into Claude Code or Cursor as a task.",
          InputSchema: mcp.MustParseInputSchema(`{
              "type": "object",
              "properties": {
                  "path": {"type": "string", "description": "Absolute path to the project directory"}
              },
              "required": ["path"]
          }`),
      }, func(ctx context.Context, req *mcp.CallToolRequest, input fixPlanInput) (*mcp.CallToolResult, error) {
          report, err := scanner.Scan(ctx, input.Path, scanner.ScanOptions{})
          if err != nil {
              return nil, fmt.Errorf("scan failed: %w", err)
          }
          plan := remedy.GenerateFixPlan(report.Results)
          return mcp.NewToolResultText(plan.ToMarkdown()), nil
      })
  }
  ```

### 1.10 — Создать MCP Resource: Playbook

- [x] Создать файл `internal/mcp/resources.go`:
  ```go
  package mcp

  import (
      "context"

      "github.com/modelcontextprotocol/go-sdk/mcp"
  )

  const playbookContent = `# AITriage Full Security Audit Playbook

  ## Шаг 1: Понять структуру проекта
  Вызови: aitriage_architecture с путём к проекту.
  Изучи стеки, ключевые файлы, количество файлов.

  ## Шаг 2: Проверить доступные сканеры
  Вызови: list_available_scanners
  Запомни какие инструменты есть в системе.

  ## Шаг 3: Запустить основной скан
  Вызови: aitriage_scan с путём к проекту.
  Получишь структурированный JSON с находками.

  ## Шаг 4: Запустить внешние сканеры (параллельно, если доступны)
  - Если semgrep=true: run_semgrep
  - Если gitleaks=true: run_gitleaks
  - Если trivy=true: run_trivy
  - Если bandit=true: run_bandit

  ## Шаг 5: Сгенерировать план исправлений
  Вызови: generate_fix_plan
  Получишь готовый промпт для агента.

  ## Шаг 6: Проанализировать и доложить
  Консолидируй все находки. Сгруппируй по severity.
  Объясни пользователю: что критично, что подождёт, как чинить.
  `

  func registerPlaybookResource(srv *mcp.Server) {
      mcp.AddResource(srv, &mcp.Resource{
          URI:         "playbook://security-audit",
          Name:        "Security Audit Playbook",
          Description: "Step-by-step instructions for AI agent to perform a full security audit using AITriage tools",
          MimeType:    "text/markdown",
      }, func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
          return &mcp.ReadResourceResult{
              Contents: []mcp.ResourceContents{
                  mcp.TextResourceContents{
                      URI:      "playbook://security-audit",
                      MimeType: "text/markdown",
                      Text:     playbookContent,
                  },
              },
          }, nil
      })
  }
  ```

### 1.11 — Проверить что всё компилируется

- [ ] `go build ./...`
- [ ] `go vet ./...`
- [ ] `go run ./cmd/aitriage serve --help` — должна появиться команда serve

---

## Фаза 2: Внешние сканеры
> Цель: AITriage умеет запускать semgrep, gitleaks, trivy, bandit и возвращать унифицированный формат.
> НЕ устанавливать эти инструменты — только обёртки для вызова если они уже есть в системе.

### 2.1 — Создать директорию и базовый runner

- [x] Создать директорию `internal/external/`
- [x] Создать файл `internal/external/runner.go`:
  ```go
  package external

  import (
      "bytes"
      "context"
      "os/exec"
  )

  // RunResult содержит результат запуска внешнего инструмента
  type RunResult struct {
      Stdout   string
      Stderr   string
      ExitCode int
  }

  // RunTool запускает внешний CLI-инструмент и возвращает его вывод.
  // Не паникует при ненулевом exit code — просто возвращает его в ExitCode.
  func RunTool(ctx context.Context, name string, args ...string) (RunResult, error) {
      cmd := exec.CommandContext(ctx, name, args...)
      var outBuf, errBuf bytes.Buffer
      cmd.Stdout = &outBuf
      cmd.Stderr = &errBuf
      err := cmd.Run()
      exitCode := 0
      if exitErr, ok := err.(*exec.ExitError); ok {
          exitCode = exitErr.ExitCode()
          err = nil // ненулевой exit code — не ошибка, инструмент так работает
      } else if err != nil {
          return RunResult{}, err
      }
      return RunResult{
          Stdout:   outBuf.String(),
          Stderr:   errBuf.String(),
          ExitCode: exitCode,
      }, nil
  }

  // IsInstalled проверяет доступность инструмента в PATH
  func IsInstalled(name string) bool {
      _, err := exec.LookPath(name)
      return err == nil
  }
  ```

### 2.2 — Создать унифицированную модель Finding

- [x] Создать файл `internal/external/finding.go`:
  ```go
  package external

  // UnifiedFinding — общая структура для результатов от всех сканеров
  type UnifiedFinding struct {
      Source     string `json:"source"`       // "aitriage" | "semgrep" | "gitleaks" | "trivy" | "bandit"
      RuleID     string `json:"rule_id"`
      Severity   string `json:"severity"`     // "CRITICAL" | "HIGH" | "MEDIUM" | "LOW" | "INFO"
      Message    string `json:"message"`
      File       string `json:"file"`
      Line       int    `json:"line"`
      Suggestion string `json:"suggestion,omitempty"`
      OWASP      string `json:"owasp,omitempty"`
  }
  ```

### 2.3 — Создать обёртку Semgrep

- [x] Создать файл `internal/external/semgrep.go`:
  ```go
  package external

  import (
      "context"
      "encoding/json"
      "fmt"
  )

  type semgrepOutput struct {
      Results []struct {
          RuleID  string `json:"check_id"`
          Message struct {
              Text string `json:"text"`
          } `json:"extra"`
          Path  string `json:"path"`
          Start struct {
              Line int `json:"line"`
          } `json:"start"`
          Severity string `json:"severity"`
      } `json:"results"`
  }

  // RunSemgrep запускает semgrep и возвращает унифицированные находки.
  // config: правила для semgrep, например "auto" или путь к yaml файлу
  func RunSemgrep(ctx context.Context, path, config string) ([]UnifiedFinding, error) {
      if !IsInstalled("semgrep") {
          return nil, fmt.Errorf("semgrep not installed")
      }
      if config == "" {
          config = "auto"
      }
      result, err := RunTool(ctx, "semgrep", "scan", "--json", "--config", config, path)
      if err != nil {
          return nil, fmt.Errorf("semgrep execution failed: %w", err)
      }
      var output semgrepOutput
      if err := json.Unmarshal([]byte(result.Stdout), &output); err != nil {
          return nil, fmt.Errorf("failed to parse semgrep output: %w", err)
      }
      findings := make([]UnifiedFinding, 0, len(output.Results))
      for _, r := range output.Results {
          findings = append(findings, UnifiedFinding{
              Source:   "semgrep",
              RuleID:   r.RuleID,
              Severity: normalizeSeverity(r.Severity),
              Message:  r.Message.Text,
              File:     r.Path,
              Line:     r.Start.Line,
          })
      }
      return findings, nil
  }

  func normalizeSeverity(s string) string {
      switch s {
      case "ERROR":
          return "HIGH"
      case "WARNING":
          return "MEDIUM"
      case "INFO":
          return "LOW"
      default:
          return s
      }
  }
  ```

### 2.4 — Создать обёртку Gitleaks

- [x] Создать файл `internal/external/gitleaks.go`:
  ```go
  package external

  import (
      "context"
      "encoding/json"
      "fmt"
  )

  type gitleaksOutput []struct {
      RuleID      string `json:"RuleID"`
      Description string `json:"Description"`
      File        string `json:"File"`
      StartLine   int    `json:"StartLine"`
      Secret      string `json:"Secret"`
  }

  // RunGitleaks запускает gitleaks и возвращает унифицированные находки.
  func RunGitleaks(ctx context.Context, path string) ([]UnifiedFinding, error) {
      if !IsInstalled("gitleaks") {
          return nil, fmt.Errorf("gitleaks not installed")
      }
      result, err := RunTool(ctx, "gitleaks", "detect", "--source", path,
          "--report-format", "json", "--report-path", "-", "--no-git")
      if err != nil {
          return nil, fmt.Errorf("gitleaks execution failed: %w", err)
      }
      if result.Stdout == "" || result.Stdout == "null" {
          return []UnifiedFinding{}, nil
      }
      var output gitleaksOutput
      if err := json.Unmarshal([]byte(result.Stdout), &output); err != nil {
          return nil, fmt.Errorf("failed to parse gitleaks output: %w", err)
      }
      findings := make([]UnifiedFinding, 0, len(output))
      for _, r := range output {
          findings = append(findings, UnifiedFinding{
              Source:   "gitleaks",
              RuleID:   r.RuleID,
              Severity: "CRITICAL",
              Message:  fmt.Sprintf("%s: %s", r.Description, maskSecret(r.Secret)),
              File:     r.File,
              Line:     r.StartLine,
          })
      }
      return findings, nil
  }

  func maskSecret(s string) string {
      if len(s) <= 4 {
          return "****"
      }
      return s[:2] + "****" + s[len(s)-2:]
  }
  ```

### 2.5 — Создать обёртку Trivy

- [x] Создать файл `internal/external/trivy.go`:
  ```go
  package external

  import (
      "context"
      "encoding/json"
      "fmt"
  )

  type trivyOutput struct {
      Results []struct {
          Target          string `json:"Target"`
          Vulnerabilities []struct {
              VulnerabilityID string `json:"VulnerabilityID"`
              Severity        string `json:"Severity"`
              Title           string `json:"Title"`
              Description     string `json:"Description"`
          } `json:"Vulnerabilities"`
      } `json:"Results"`
  }

  // RunTrivy запускает trivy и возвращает унифицированные находки.
  // scanType: "fs" (filesystem) или "config" (IaC конфиги)
  func RunTrivy(ctx context.Context, path, scanType string) ([]UnifiedFinding, error) {
      if !IsInstalled("trivy") {
          return nil, fmt.Errorf("trivy not installed")
      }
      if scanType == "" {
          scanType = "fs"
      }
      result, err := RunTool(ctx, "trivy", scanType, "--format", "json", "--quiet", path)
      if err != nil {
          return nil, fmt.Errorf("trivy execution failed: %w", err)
      }
      var output trivyOutput
      if err := json.Unmarshal([]byte(result.Stdout), &output); err != nil {
          return nil, fmt.Errorf("failed to parse trivy output: %w", err)
      }
      var findings []UnifiedFinding
      for _, res := range output.Results {
          for _, v := range res.Vulnerabilities {
              findings = append(findings, UnifiedFinding{
                  Source:   "trivy",
                  RuleID:   v.VulnerabilityID,
                  Severity: v.Severity,
                  Message:  fmt.Sprintf("%s: %s", v.Title, v.Description),
                  File:     res.Target,
              })
          }
      }
      return findings, nil
  }
  ```

### 2.6 — Создать обёртку Bandit

- [x] Создать файл `internal/external/bandit.go`:
  ```go
  package external

  import (
      "context"
      "encoding/json"
      "fmt"
  )

  type banditOutput struct {
      Results []struct {
          TestID    string `json:"test_id"`
          TestName  string `json:"test_name"`
          Severity  string `json:"issue_severity"`
          Text      string `json:"issue_text"`
          Filename  string `json:"filename"`
          LineRange []int  `json:"line_range"`
      } `json:"results"`
  }

  // RunBandit запускает bandit (Python SAST) и возвращает унифицированные находки.
  func RunBandit(ctx context.Context, path string) ([]UnifiedFinding, error) {
      if !IsInstalled("bandit") {
          return nil, fmt.Errorf("bandit not installed")
      }
      result, err := RunTool(ctx, "bandit", "-r", path, "-f", "json", "-q")
      if err != nil {
          return nil, fmt.Errorf("bandit execution failed: %w", err)
      }
      var output banditOutput
      if err := json.Unmarshal([]byte(result.Stdout), &output); err != nil {
          return nil, fmt.Errorf("failed to parse bandit output: %w", err)
      }
      findings := make([]UnifiedFinding, 0, len(output.Results))
      for _, r := range output.Results {
          line := 0
          if len(r.LineRange) > 0 {
              line = r.LineRange[0]
          }
          findings = append(findings, UnifiedFinding{
              Source:   "bandit",
              RuleID:   r.TestID,
              Severity: r.Severity,
              Message:  fmt.Sprintf("%s: %s", r.TestName, r.Text),
              File:     r.Filename,
              Line:     line,
          })
      }
      return findings, nil
  }
  ```

### 2.7 — Создать MCP tools для внешних сканеров

- [x] Создать файл `internal/mcp/tools_external.go` с tools: `run_semgrep`, `run_gitleaks`, `run_trivy`, `run_bandit`
- [x] Каждый tool: проверяет IsInstalled → запускает → возвращает []UnifiedFinding как JSON
- [x] Если инструмент не установлен — возвращать понятную ошибку: `"semgrep is not installed. Run: brew install semgrep"`

### 2.8 — Проверить компиляцию

- [x] `go build ./...`
- [x] `go vet ./...`

---

## Фаза 3: LLM Layer (AI-lab agnostic)
> Цель: AITriage умеет говорить с любым LLM-провайдером через единый интерфейс.
> НЕ использовать LLM в режиме `scan`. Только в режиме `agent`.

### 3.1 — Создать директорию и интерфейс

- [x] Создать директорию `internal/llm/`
- [x] Создать файл `internal/llm/client.go`:
  ```go
  package llm

  import "context"

  // Message — одно сообщение в чате
  type Message struct {
      Role    string `json:"role"`    // "system" | "user" | "assistant"
      Content string `json:"content"`
  }

  // Client — интерфейс для любого LLM провайдера
  type Client interface {
      // Chat отправляет массив сообщений и возвращает ответ модели
      Chat(ctx context.Context, messages []Message) (string, error)
  }

  // Config — конфигурация LLM провайдера из .aitriage.yaml
  type Config struct {
      Provider string `yaml:"provider"` // "anthropic" | "openai" | "ollama"
      Model    string `yaml:"model"`
      APIKey   string `yaml:"api_key"`
      BaseURL  string `yaml:"base_url"` // для ollama и openai-compatible
      Timeout  int    `yaml:"timeout"`  // секунды, default 120
  }
  ```

### 3.2 — Создать OpenAI-совместимый клиент

> Этот клиент работает с: OpenAI, Ollama, Groq, Together, Mistral, LM Studio — у всех OpenAI-совместимый API.

- [x] Создать файл `internal/llm/openai.go`:
  ```go
  package llm

  import (
      "bytes"
      "context"
      "encoding/json"
      "fmt"
      "io"
      "net/http"
      "time"
  )

  type openAIClient struct {
      cfg        Config
      httpClient *http.Client
  }

  type openAIRequest struct {
      Model    string    `json:"model"`
      Messages []Message `json:"messages"`
  }

  type openAIResponse struct {
      Choices []struct {
          Message struct {
              Content string `json:"content"`
          } `json:"message"`
      } `json:"choices"`
      Error *struct {
          Message string `json:"message"`
      } `json:"error,omitempty"`
  }

  func newOpenAIClient(cfg Config) *openAIClient {
      timeout := time.Duration(cfg.Timeout) * time.Second
      if timeout == 0 {
          timeout = 120 * time.Second
      }
      return &openAIClient{
          cfg:        cfg,
          httpClient: &http.Client{Timeout: timeout},
      }
  }

  func (c *openAIClient) Chat(ctx context.Context, messages []Message) (string, error) {
      baseURL := c.cfg.BaseURL
      if baseURL == "" {
          baseURL = "https://api.openai.com"
      }
      url := baseURL + "/v1/chat/completions"

      body, _ := json.Marshal(openAIRequest{
          Model:    c.cfg.Model,
          Messages: messages,
      })

      req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
      if err != nil {
          return "", err
      }
      req.Header.Set("Content-Type", "application/json")
      if c.cfg.APIKey != "" {
          req.Header.Set("Authorization", "Bearer "+c.cfg.APIKey)
      }

      resp, err := c.httpClient.Do(req)
      if err != nil {
          return "", fmt.Errorf("LLM request failed: %w", err)
      }
      defer resp.Body.Close()

      data, _ := io.ReadAll(resp.Body)
      var result openAIResponse
      if err := json.Unmarshal(data, &result); err != nil {
          return "", fmt.Errorf("failed to parse LLM response: %w", err)
      }
      if result.Error != nil {
          return "", fmt.Errorf("LLM API error: %s", result.Error.Message)
      }
      if len(result.Choices) == 0 {
          return "", fmt.Errorf("empty response from LLM")
      }
      return result.Choices[0].Message.Content, nil
  }
  ```

### 3.3 — Создать Anthropic клиент

- [x] Создать файл `internal/llm/anthropic.go`:
  ```go
  package llm

  import (
      "bytes"
      "context"
      "encoding/json"
      "fmt"
      "io"
      "net/http"
      "time"
  )

  type anthropicClient struct {
      cfg        Config
      httpClient *http.Client
  }

  type anthropicRequest struct {
      Model     string             `json:"model"`
      MaxTokens int                `json:"max_tokens"`
      System    string             `json:"system,omitempty"`
      Messages  []anthropicMessage `json:"messages"`
  }

  type anthropicMessage struct {
      Role    string `json:"role"`
      Content string `json:"content"`
  }

  type anthropicResponse struct {
      Content []struct {
          Text string `json:"text"`
      } `json:"content"`
      Error *struct {
          Message string `json:"message"`
      } `json:"error,omitempty"`
  }

  func newAnthropicClient(cfg Config) *anthropicClient {
      timeout := time.Duration(cfg.Timeout) * time.Second
      if timeout == 0 {
          timeout = 120 * time.Second
      }
      return &anthropicClient{
          cfg:        cfg,
          httpClient: &http.Client{Timeout: timeout},
      }
  }

  func (c *anthropicClient) Chat(ctx context.Context, messages []Message) (string, error) {
      var systemPrompt string
      var anthropicMsgs []anthropicMessage

      for _, m := range messages {
          if m.Role == "system" {
              systemPrompt = m.Content
              continue
          }
          anthropicMsgs = append(anthropicMsgs, anthropicMessage{
              Role:    m.Role,
              Content: m.Content,
          })
      }

      body, _ := json.Marshal(anthropicRequest{
          Model:     c.cfg.Model,
          MaxTokens: 8096,
          System:    systemPrompt,
          Messages:  anthropicMsgs,
      })

      req, err := http.NewRequestWithContext(ctx, "POST", "https://api.anthropic.com/v1/messages", bytes.NewReader(body))
      if err != nil {
          return "", err
      }
      req.Header.Set("Content-Type", "application/json")
      req.Header.Set("x-api-key", c.cfg.APIKey)
      req.Header.Set("anthropic-version", "2023-06-01")

      resp, err := c.httpClient.Do(req)
      if err != nil {
          return "", fmt.Errorf("Anthropic API request failed: %w", err)
      }
      defer resp.Body.Close()

      data, _ := io.ReadAll(resp.Body)
      var result anthropicResponse
      if err := json.Unmarshal(data, &result); err != nil {
          return "", fmt.Errorf("failed to parse Anthropic response: %w", err)
      }
      if result.Error != nil {
          return "", fmt.Errorf("Anthropic API error: %s", result.Error.Message)
      }
      if len(result.Content) == 0 {
          return "", fmt.Errorf("empty response from Anthropic")
      }
      return result.Content[0].Text, nil
  }
  ```

### 3.4 — Создать фабрику клиентов

- [x] Создать файл `internal/llm/factory.go`:
  ```go
  package llm

  import "fmt"

  // NewClient создаёт LLM клиент нужного провайдера на основе конфига.
  // provider: "anthropic" | "openai" | "ollama" | "groq"
  // Для ollama: указать base_url: "http://localhost:11434", api_key не нужен
  // Для groq: указать base_url: "https://api.groq.com/openai", api_key: groq ключ
  func NewClient(cfg Config) (Client, error) {
      switch cfg.Provider {
      case "anthropic":
          if cfg.APIKey == "" {
              return nil, fmt.Errorf("anthropic provider requires api_key")
          }
          return newAnthropicClient(cfg), nil
      case "openai", "ollama", "groq", "":
          // OpenAI-совместимый API — один клиент для всех
          return newOpenAIClient(cfg), nil
      default:
          return nil, fmt.Errorf("unknown LLM provider: %q. Supported: anthropic, openai, ollama, groq", cfg.Provider)
      }
  }
  ```

### 3.5 — Создать системные промпты

- [x] Создать файл `internal/llm/prompts.go`:
  ```go
  package llm

  import (
      "fmt"
      "strings"
      "github.com/cybertortuga/aitriage/internal/core"
  )

  // SystemPrompt — базовая роль агента
  const SystemPrompt = `You are AITriage, an expert security engineer AI assistant.
  You help developers find and fix security vulnerabilities in their code.
  You have access to deterministic scan results from static analysis tools.
  Always be specific: cite file names and line numbers.
  Prioritize critical findings. Be concise but thorough.
  Respond in the same language as the user's question.`

  // BuildAnalysisPrompt строит промпт для анализа результатов сканирования
  func BuildAnalysisPrompt(results []core.CheckResult, projectPath string) string {
      if len(results) == 0 {
          return fmt.Sprintf("Project at %s has been scanned. No security issues found.", projectPath)
      }
      var sb strings.Builder
      sb.WriteString(fmt.Sprintf("I've scanned the project at: %s\n\n", projectPath))
      sb.WriteString(fmt.Sprintf("Found %d security findings:\n\n", len(results)))
      for i, r := range results {
          sb.WriteString(fmt.Sprintf("%d. [%s] %s\n", i+1, r.Severity, r.Name))
          sb.WriteString(fmt.Sprintf("   File: %s, Line: %d\n", r.File, r.Line))
          sb.WriteString(fmt.Sprintf("   Evidence: %s\n", r.Evidence))
          sb.WriteString(fmt.Sprintf("   Rule ID: %s\n\n", r.ID))
      }
      sb.WriteString("\nPlease analyze these findings, prioritize them, and provide specific remediation recommendations.")
      return sb.String()
  }

  // BuildConsultationPrompt строит промпт для режима Q&A
  func BuildConsultationPrompt(question string, previousAnalysis string) string {
      return fmt.Sprintf(`Based on the security analysis above, answer the following question:

  %s

  Be specific, cite findings and line numbers where relevant.`, question)
  }
  ```

### 3.6 — Проверить компиляцию

- [x] `go build ./...`
- [x] `go vet ./...`

---

## Фаза 4: Agent Mode (`aitriage agent`)
> Цель: команда `aitriage agent ./` запускает полный воркфлоу из task.md.
> Параллельное сканирование → LLM-анализ → интерактивная консультация.

### 4.1 — Обновить конфиг для поддержки LLM

- [x] Открыть файл `internal/config/config.go` (или найти файл конфигурации проекта)
- [x] Добавить в struct конфига:
  ```go
  LLM struct {
      Provider string `yaml:"provider"`
      Model    string `yaml:"model"`
      APIKey   string `yaml:"api_key"`
      BaseURL  string `yaml:"base_url"`
      Timeout  int    `yaml:"timeout"`
  } `yaml:"llm"`
  ```
- [x] Добавить в README.md пример `.aitriage.yaml` с LLM-конфигом:
  ```yaml
  llm:
    provider: anthropic   # anthropic | openai | ollama | groq
    model: claude-sonnet-4-5
    api_key: $ANTHROPIC_API_KEY  # или явный ключ
    timeout: 120
  ```

### 4.2 — Создать команду `agent`

- [x] Создать файл `cmd/aitriage/agent.go`:
  ```go
  package main

  import (
      "bufio"
      "context"
      "fmt"
      "os"
      "sync"

      "github.com/spf13/cobra"
      "github.com/cybertortuga/aitriage/internal/config"
      "github.com/cybertortuga/aitriage/internal/external"
      "github.com/cybertortuga/aitriage/internal/llm"
      "github.com/cybertortuga/aitriage/internal/scanner"
  )

  var (
      agentProvider string
      agentModel    string
      agentAPIKey   string
      agentNoChat   bool
      agentOutput   string
  )

  var agentCmd = &cobra.Command{
      Use:   "agent [path]",
      Short: "Run AI-powered security audit with LLM analysis and Q&A",
      Args:  cobra.ExactArgs(1),
      RunE:  runAgent,
  }

  func init() {
      rootCmd.AddCommand(agentCmd)
      agentCmd.Flags().StringVar(&agentProvider, "provider", "", "LLM provider: anthropic, openai, ollama, groq")
      agentCmd.Flags().StringVar(&agentModel, "model", "", "LLM model name")
      agentCmd.Flags().StringVar(&agentAPIKey, "api-key", "", "LLM API key (or set via env)")
      agentCmd.Flags().BoolVar(&agentNoChat, "no-chat", false, "Skip interactive Q&A (for CI/CD)")
      agentCmd.Flags().StringVar(&agentOutput, "output", "text", "Output format: text | json | md")
  }

  func runAgent(cmd *cobra.Command, args []string) error {
      projectPath := args[0]
      ctx := context.Background()

      // Загрузить конфиг
      cfg, _ := config.Load(projectPath)

      // Флаги командной строки перебивают конфиг
      llmCfg := cfg.LLM
      if agentProvider != "" { llmCfg.Provider = agentProvider }
      if agentModel != ""    { llmCfg.Model = agentModel }
      if agentAPIKey != ""   { llmCfg.APIKey = agentAPIKey }

      // Создать LLM клиент
      client, err := llm.NewClient(llm.Config{
          Provider: llmCfg.Provider,
          Model:    llmCfg.Model,
          APIKey:   llmCfg.APIKey,
          BaseURL:  llmCfg.BaseURL,
          Timeout:  llmCfg.Timeout,
      })
      if err != nil {
          return fmt.Errorf("LLM setup failed: %w\nSet provider via --provider flag or .aitriage.yaml", err)
      }

      fmt.Fprintf(os.Stderr, "🔍 AITriage Agent starting...\n\n")

      // ШАГ 1: ПАРАЛЛЕЛЬНОЕ СКАНИРОВАНИЕ
      fmt.Fprintf(os.Stderr, "📡 Step 1/3: Scanning (parallel)...\n")
      report, allFindings := runParallelScan(ctx, projectPath)

      fmt.Fprintf(os.Stderr, "   ✓ AITriage: %d findings\n", len(report.Results))
      fmt.Fprintf(os.Stderr, "   ✓ External: %d findings\n", len(allFindings))
      fmt.Fprintf(os.Stderr, "   SecurityGrade: %s (%d/100)\n\n", report.SecurityGrade, report.SecurityScore)

      // ШАГ 2: LLM АНАЛИЗ
      fmt.Fprintf(os.Stderr, "🤖 Step 2/3: LLM Analysis...\n")
      messages := []llm.Message{
          {Role: "system", Content: llm.SystemPrompt},
          {Role: "user", Content: llm.BuildAnalysisPrompt(report.Results, projectPath)},
      }
      analysis, err := client.Chat(ctx, messages)
      if err != nil {
          return fmt.Errorf("LLM analysis failed: %w", err)
      }
      messages = append(messages, llm.Message{Role: "assistant", Content: analysis})
      fmt.Println(analysis)
      fmt.Fprintf(os.Stderr, "\n")

      // ШАГ 3: ИНТЕРАКТИВНАЯ КОНСУЛЬТАЦИЯ
      if !agentNoChat {
          fmt.Fprintf(os.Stderr, "💬 Step 3/3: Consultation mode (type 'exit' to quit)\n")
          runConsultation(ctx, client, messages)
      }

      return nil
  }

  func runParallelScan(ctx context.Context, path string) (scanner.ScanReport, []external.UnifiedFinding) {
      var wg sync.WaitGroup
      var mu sync.Mutex
      var report scanner.ScanReport
      var allFindings []external.UnifiedFinding

      // Горутина 1: основной скан AITriage
      wg.Add(1)
      go func() {
          defer wg.Done()
          r, err := scanner.Scan(ctx, path, scanner.ScanOptions{})
          if err == nil {
              mu.Lock()
              report = r
              mu.Unlock()
          }
      }()

      // Горутина 2: внешние сканеры (запускает только установленные)
      wg.Add(1)
      go func() {
          defer wg.Done()
          var scanners [][]external.UnifiedFinding
          var swg sync.WaitGroup

          if external.IsInstalled("semgrep") {
              swg.Add(1)
              go func() {
                  defer swg.Done()
                  findings, err := external.RunSemgrep(ctx, path, "auto")
                  if err == nil { mu.Lock(); scanners = append(scanners, findings); mu.Unlock() }
              }()
          }
          if external.IsInstalled("gitleaks") {
              swg.Add(1)
              go func() {
                  defer swg.Done()
                  findings, err := external.RunGitleaks(ctx, path)
                  if err == nil { mu.Lock(); scanners = append(scanners, findings); mu.Unlock() }
              }()
          }
          if external.IsInstalled("bandit") {
              swg.Add(1)
              go func() {
                  defer swg.Done()
                  findings, err := external.RunBandit(ctx, path)
                  if err == nil { mu.Lock(); scanners = append(scanners, findings); mu.Unlock() }
              }()
          }

          swg.Wait()
          mu.Lock()
          for _, f := range scanners {
              allFindings = append(allFindings, f...)
          }
          mu.Unlock()
      }()

      wg.Wait()
      return report, allFindings
  }

  func runConsultation(ctx context.Context, client llm.Client, history []llm.Message) {
      scanner := bufio.NewScanner(os.Stdin)
      fmt.Print("\n> ")
      for scanner.Scan() {
          question := scanner.Text()
          if question == "exit" || question == "quit" {
              break
          }
          if question == "" {
              fmt.Print("> ")
              continue
          }
          history = append(history, llm.Message{
              Role:    "user",
              Content: llm.BuildConsultationPrompt(question, ""),
          })
          answer, err := client.Chat(ctx, history)
          if err != nil {
              fmt.Fprintf(os.Stderr, "Error: %v\n", err)
          } else {
              history = append(history, llm.Message{Role: "assistant", Content: answer})
              fmt.Println(answer)
          }
          fmt.Print("\n> ")
      }
  }
  ```

### 4.3 — Проверить компиляцию и работу

- [x] `go build ./...`
- [x] `go vet ./...`
- [x] `go run ./cmd/aitriage agent --help` — должны появиться флаги
- [x] `go run ./cmd/aitriage agent . --no-chat --provider openai --model gpt-4o-mini --api-key $OPENAI_API_KEY` — запустить тест

---

## Фаза 5: Расширенные инструменты
> Цель: добавить инструменты аудита которых не хватает: git-анализ, структура деплоя, NFR, схема архитектуры.

### 5.1 — Git-анализ

- [x] Создать директорию `internal/gitanalysis/`
- [x] Создать файл `internal/gitanalysis/git.go`:
  ```go
  package gitanalysis

  import (
      "context"
      "fmt"
      "path/filepath"
      "strings"
  )

  // критические паттерны имён файлов которые не должны быть в git
  var criticalPatterns = []string{
      ".env", ".env.local", ".env.production",
      "*.pem", "*.key", "*.p12", "*.pfx",
      "*secret*", "*password*", "*credential*",
      "id_rsa", "id_ed25519",
  }

  type GitFinding struct {
      Commit   string `json:"commit"`
      File     string `json:"file"`
      Issue    string `json:"issue"`    // "secret_in_history" | "env_committed" | "key_file"
      Severity string `json:"severity"` // "CRITICAL" | "HIGH"
      Message  string `json:"message"`
  }

  // AnalyzeGitHistory проверяет git-историю на наличие критических файлов.
  // Запускает git log и ищет коммиты с чувствительными файлами.
  func AnalyzeGitHistory(ctx context.Context, repoPath string) ([]GitFinding, error) {
      // Проверить что это git-репозиторий
      result, err := runGit(ctx, repoPath, "rev-parse", "--git-dir")
      if err != nil || strings.TrimSpace(result) == "" {
          return nil, fmt.Errorf("not a git repository: %s", repoPath)
      }

      // Получить список всех файлов в истории (включая удалённые)
      histResult, err := runGit(ctx, repoPath, "log", "--all", "--name-only", "--format=%H", "--diff-filter=A")
      if err != nil {
          return nil, fmt.Errorf("git log failed: %w", err)
      }

      var findings []GitFinding
      lines := strings.Split(histResult, "\n")
      currentCommit := ""

      for _, line := range lines {
          line = strings.TrimSpace(line)
          if line == "" {
              continue
          }
          // Строки из 40 hex символов — это commit hash
          if isCommitHash(line) {
              currentCommit = line
              continue
          }
          // Остальное — имена файлов
          filename := filepath.Base(line)
          for _, pattern := range criticalPatterns {
              matched, _ := filepath.Match(pattern, filename)
              if matched || strings.Contains(strings.ToLower(filename), "secret") ||
                  strings.Contains(strings.ToLower(filename), "password") {
                  findings = append(findings, GitFinding{
                      Commit:   currentCommit[:8],
                      File:     line,
                      Issue:    "secret_in_history",
                      Severity: "CRITICAL",
                      Message:  fmt.Sprintf("File %q was committed and may contain secrets (commit %s)", line, currentCommit[:8]),
                  })
                  break
              }
          }
      }

      return findings, nil
  }

  func isCommitHash(s string) bool {
      if len(s) != 40 {
          return false
      }
      for _, c := range s {
          if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
              return false
          }
      }
      return true
  }
  ```

- [x] Создать файл `internal/gitanalysis/runner.go`:
  ```go
  package gitanalysis

  import (
      "bytes"
      "context"
      "os/exec"
  )

  func runGit(ctx context.Context, repoPath string, args ...string) (string, error) {
      cmd := exec.CommandContext(ctx, "git", args...)
      cmd.Dir = repoPath
      var out bytes.Buffer
      cmd.Stdout = &out
      cmd.Run() // игнорируем ошибку — git возвращает ненулевой код при пустом результате
      return out.String(), nil
  }
  ```

- [x] Добавить MCP tool `aitriage_git_analysis` в `internal/mcp/tools_git.go`

### 5.2 — Аудит файлов деплоя

- [x] Создать директорию `internal/deployaudit/`
- [x] Создать файл `internal/deployaudit/audit.go`:
  ```go
  package deployaudit

  import (
      "bufio"
      "fmt"
      "os"
      "path/filepath"
      "strings"
  )

  type DeployFinding struct {
      File     string `json:"file"`
      Line     int    `json:"line"`
      Issue    string `json:"issue"`
      Severity string `json:"severity"`
      Evidence string `json:"evidence"`
      Advice   string `json:"advice"`
  }

  // AuditDeployFiles проверяет файлы деплоя на security-проблемы.
  func AuditDeployFiles(projectPath string) ([]DeployFinding, error) {
      var findings []DeployFinding

      // Проверить Dockerfile
      dockerFile := filepath.Join(projectPath, "Dockerfile")
      if _, err := os.Stat(dockerFile); err == nil {
          f, _ := auditDockerfile(dockerFile)
          findings = append(findings, f...)
      }

      // Проверить docker-compose.yml
      for _, name := range []string{"docker-compose.yml", "docker-compose.yaml"} {
          composeFile := filepath.Join(projectPath, name)
          if _, err := os.Stat(composeFile); err == nil {
              f, _ := auditCompose(composeFile)
              findings = append(findings, f...)
          }
      }

      return findings, nil
  }

  func auditDockerfile(path string) ([]DeployFinding, error) {
      var findings []DeployFinding
      file, err := os.Open(path)
      if err != nil {
          return nil, err
      }
      defer file.Close()

      scanner := bufio.NewScanner(file)
      lineNum := 0
      for scanner.Scan() {
          lineNum++
          line := strings.TrimSpace(scanner.Text())

          // Проверка: запуск от root
          if strings.HasPrefix(line, "USER root") || (strings.HasPrefix(line, "USER") && !strings.Contains(line, "USER ")) {
              findings = append(findings, DeployFinding{
                  File: path, Line: lineNum,
                  Issue: "dockerfile_root_user", Severity: "HIGH",
                  Evidence: line,
                  Advice:   "Add USER nonroot or create a dedicated user. Running as root in containers is a security risk.",
              })
          }

          // Проверка: ADD вместо COPY (ADD может раскрывать URLs)
          if strings.HasPrefix(line, "ADD ") && !strings.Contains(line, ".tar") {
              findings = append(findings, DeployFinding{
                  File: path, Line: lineNum,
                  Issue: "dockerfile_add_instead_of_copy", Severity: "LOW",
                  Evidence: line,
                  Advice:   "Prefer COPY over ADD unless you need URL fetching or tar extraction.",
              })
          }

          // Проверка: hardcoded secrets в ENV
          if strings.HasPrefix(line, "ENV ") && (strings.Contains(line, "PASSWORD") ||
              strings.Contains(line, "SECRET") || strings.Contains(line, "API_KEY")) {
              findings = append(findings, DeployFinding{
                  File: path, Line: lineNum,
                  Issue: "dockerfile_hardcoded_secret", Severity: "CRITICAL",
                  Evidence: fmt.Sprintf("%.50s...", line),
                  Advice:   "Never hardcode secrets in Dockerfile ENV. Use --secret or runtime env injection.",
              })
          }
      }
      return findings, nil
  }

  func auditCompose(path string) ([]DeployFinding, error) {
      var findings []DeployFinding
      file, err := os.Open(path)
      if err != nil {
          return nil, err
      }
      defer file.Close()

      scanner := bufio.NewScanner(file)
      lineNum := 0
      for scanner.Scan() {
          lineNum++
          line := strings.TrimSpace(scanner.Text())
          lower := strings.ToLower(line)

          // Проверка: hardcoded пароли
          if (strings.Contains(lower, "password") || strings.Contains(lower, "secret")) &&
              strings.Contains(line, ":") && !strings.Contains(line, "${") {
              findings = append(findings, DeployFinding{
                  File: path, Line: lineNum,
                  Issue: "compose_hardcoded_secret", Severity: "CRITICAL",
                  Evidence: fmt.Sprintf("%.60s", line),
                  Advice:   "Use environment variable substitution: ${MY_PASSWORD}",
              })
          }

          // Проверка: privileged mode
          if strings.Contains(lower, "privileged: true") {
              findings = append(findings, DeployFinding{
                  File: path, Line: lineNum,
                  Issue: "compose_privileged_mode", Severity: "HIGH",
                  Evidence: line,
                  Advice:   "Remove privileged: true. Grant only specific capabilities with cap_add instead.",
              })
          }
      }
      return findings, nil
  }
  ```

- [x] Добавить MCP tool `aitriage_deploy_audit` в `internal/mcp/tools_deploy.go`

### 5.3 — NFR Compliance

- [x] Создать директорию `internal/nfr/`
- [x] Создать директорию `internal/nfr/rules/`
- [x] Создать файл `internal/nfr/checker.go`:
  ```go
  package nfr

  import (
      "fmt"
      "os"
      "path/filepath"
      "regexp"
      "strings"

      _ "embed"
      "gopkg.in/yaml.v3"
  )

  type Rule struct {
      ID       string   `yaml:"id"`
      Name     string   `yaml:"name"`
      Severity string   `yaml:"severity"`
      Message  string   `yaml:"message"`
      Advice   string   `yaml:"advice"`
      Check    string   `yaml:"check"`   // "file_contains" | "file_exists" | "file_not_exists"
      Pattern  string   `yaml:"pattern"` // regex для file_contains
      Files    []string `yaml:"files"`   // glob паттерны файлов для проверки
  }

  type NFRFinding struct {
      RuleID   string `json:"rule_id"`
      Name     string `json:"name"`
      Severity string `json:"severity"`
      Message  string `json:"message"`
      Advice   string `json:"advice"`
  }

  // CheckNFR проверяет проект на соответствие NFR правилам из директории rulesDir
  func CheckNFR(projectPath, rulesDir string) ([]NFRFinding, error) {
      var allRules []Rule

      entries, err := os.ReadDir(rulesDir)
      if err != nil {
          return nil, fmt.Errorf("cannot read NFR rules dir %s: %w", rulesDir, err)
      }

      for _, e := range entries {
          if !strings.HasSuffix(e.Name(), ".yaml") && !strings.HasSuffix(e.Name(), ".yml") {
              continue
          }
          data, err := os.ReadFile(filepath.Join(rulesDir, e.Name()))
          if err != nil {
              continue
          }
          var rules []Rule
          if err := yaml.Unmarshal(data, &rules); err != nil {
              continue
          }
          allRules = append(allRules, rules...)
      }

      var findings []NFRFinding
      for _, rule := range allRules {
          triggered, err := evaluateRule(projectPath, rule)
          if err != nil {
              continue
          }
          if triggered {
              findings = append(findings, NFRFinding{
                  RuleID:   rule.ID,
                  Name:     rule.Name,
                  Severity: rule.Severity,
                  Message:  rule.Message,
                  Advice:   rule.Advice,
              })
          }
      }

      return findings, nil
  }

  func evaluateRule(projectPath string, rule Rule) (bool, error) {
      switch rule.Check {
      case "file_contains":
          re, err := regexp.Compile(rule.Pattern)
          if err != nil {
              return false, err
          }
          found := false
          for _, glob := range rule.Files {
              matches, _ := filepath.Glob(filepath.Join(projectPath, "**", glob))
              matches2, _ := filepath.Glob(filepath.Join(projectPath, glob))
              matches = append(matches, matches2...)
              for _, path := range matches {
                  data, err := os.ReadFile(path)
                  if err != nil {
                      continue
                  }
                  if re.Match(data) {
                      found = true
                      break
                  }
              }
          }
          return !found, nil // NFR нарушено если паттерн НЕ найден
      case "file_exists":
          _, err := os.Stat(filepath.Join(projectPath, rule.Pattern))
          return os.IsNotExist(err), nil // нарушено если файл НЕ существует
      case "file_not_exists":
          _, err := os.Stat(filepath.Join(projectPath, rule.Pattern))
          return err == nil, nil // нарушено если файл СУЩЕСТВУЕТ
      default:
          return false, nil
      }
  }
  ```

- [x] Создать файл `internal/nfr/rules/web_api.yaml`:
  ```yaml
  - id: NFR-API-001
    name: Rate Limiting Missing
    severity: HIGH
    check: file_contains
    pattern: "rate.?limit|ratelimit|RateLimit|throttle"
    files: ["*.go", "*.py", "*.ts", "*.js"]
    message: "No rate limiting detected in source files"
    advice: "Implement rate limiting middleware. For Go: golang.org/x/time/rate. For Express: express-rate-limit."

  - id: NFR-API-002
    name: CORS Configuration Missing
    severity: MEDIUM
    check: file_contains
    pattern: "cors|CORS|Access-Control-Allow"
    files: ["*.go", "*.py", "*.ts", "*.js"]
    message: "No CORS configuration detected"
    advice: "Configure CORS explicitly. Never use wildcard (*) in production."

  - id: NFR-API-003
    name: Authentication Middleware Missing
    severity: HIGH
    check: file_contains
    pattern: "auth|Auth|jwt|JWT|bearer|Bearer|middleware"
    files: ["*.go", "*.py", "*.ts", "*.js"]
    message: "No authentication middleware detected"
    advice: "Implement authentication middleware for all protected routes."

  - id: NFR-ENV-001
    name: .env.example Missing
    severity: MEDIUM
    check: file_exists
    pattern: ".env.example"
    message: ".env.example file is missing"
    advice: "Create .env.example with all required env variables (without values) for documentation."

  - id: NFR-ENV-002
    name: .env Committed to Git
    severity: CRITICAL
    check: file_not_exists
    pattern: ".env"
    message: ".env file exists in project root — may be committed to git"
    advice: "Add .env to .gitignore immediately. Rotate all secrets if it was ever committed."
  ```

- [x] Добавить MCP tool `aitriage_nfr_check` в `internal/mcp/tools_nfr.go`

### 5.4 — Схема архитектуры (Mermaid)

- [x] Создать директорию `internal/architect/`
- [x] Создать файл `internal/architect/diagram.go`:
  ```go
  package architect

  import (
      "fmt"
      "os"
      "path/filepath"
      "strings"
  )

  type Component struct {
      Name string
      Type string // "app" | "db" | "cache" | "proxy" | "storage"
  }

  // GenerateMermaidDiagram анализирует проект и генерирует Mermaid-диаграмму.
  // Определяет компоненты по наличию ключевых файлов и зависимостей.
  func GenerateMermaidDiagram(projectPath string) (string, error) {
      components := detectComponents(projectPath)
      if len(components) == 0 {
          return "", fmt.Errorf("could not detect project components")
      }

      var sb strings.Builder
      sb.WriteString("graph TD\n")

      // Основное приложение всегда есть
      mainApp := findMainApp(components)
      sb.WriteString(fmt.Sprintf("    %s[\"%s\"]\n", sanitize(mainApp.Name), mainApp.Name))

      for _, c := range components {
          if c.Name == mainApp.Name {
              continue
          }
          nodeID := sanitize(c.Name)
          var icon string
          switch c.Type {
          case "db":
              icon = "🗄️"
          case "cache":
              icon = "⚡"
          case "proxy":
              icon = "🔀"
          case "storage":
              icon = "📦"
          default:
              icon = "🔧"
          }
          sb.WriteString(fmt.Sprintf("    %s[\"%s %s\"]\n", nodeID, icon, c.Name))
          sb.WriteString(fmt.Sprintf("    %s --> %s\n", sanitize(mainApp.Name), nodeID))
      }

      return sb.String(), nil
  }

  func detectComponents(projectPath string) []Component {
      var components []Component

      // Определить тип основного приложения
      if fileExists(filepath.Join(projectPath, "go.mod")) {
          components = append(components, Component{Name: "Go App", Type: "app"})
      } else if fileExists(filepath.Join(projectPath, "package.json")) {
          components = append(components, Component{Name: "Node.js App", Type: "app"})
      } else if fileExists(filepath.Join(projectPath, "requirements.txt")) {
          components = append(components, Component{Name: "Python App", Type: "app"})
      } else {
          components = append(components, Component{Name: "Application", Type: "app"})
      }

      // Определить зависимости по наличию файлов и строк в docker-compose
      composeFiles := []string{"docker-compose.yml", "docker-compose.yaml"}
      for _, cf := range composeFiles {
          data, err := os.ReadFile(filepath.Join(projectPath, cf))
          if err != nil {
              continue
          }
          content := strings.ToLower(string(data))
          if strings.Contains(content, "postgres") || strings.Contains(content, "postgresql") {
              components = append(components, Component{Name: "PostgreSQL", Type: "db"})
          }
          if strings.Contains(content, "mysql") || strings.Contains(content, "mariadb") {
              components = append(components, Component{Name: "MySQL", Type: "db"})
          }
          if strings.Contains(content, "redis") {
              components = append(components, Component{Name: "Redis", Type: "cache"})
          }
          if strings.Contains(content, "nginx") {
              components = append(components, Component{Name: "Nginx", Type: "proxy"})
          }
          if strings.Contains(content, "minio") {
              components = append(components, Component{Name: "MinIO", Type: "storage"})
          }
      }

      return components
  }

  func findMainApp(components []Component) Component {
      for _, c := range components {
          if c.Type == "app" {
              return c
          }
      }
      return Component{Name: "Application", Type: "app"}
  }

  func sanitize(s string) string {
      return strings.ReplaceAll(strings.ReplaceAll(s, " ", "_"), ".", "_")
  }

  func fileExists(path string) bool {
      _, err := os.Stat(path)
      return err == nil
  }
  ```

- [x] Добавить MCP tool `aitriage_diagram` в `internal/mcp/tools_diagram.go`

### 5.5 — Проверить компиляцию

- [x] `go build ./...`
- [x] `go vet ./...`

---

## Фаза 6: Remediation Engine
> Цель: детерминированные fix-шаблоны. НЕ использовать LLM. Работает автономно.

### 6.1 — Создать Fix Plan Generator

- [x] Создать файл `internal/remedy/fix_plan.go`:
  ```go
  package remedy

  import (
      "fmt"
      "strings"
      "time"

      "github.com/cybertortuga/aitriage/internal/core"
  )

  type FixItem struct {
      RuleID     string `json:"rule_id"`
      Name       string `json:"name"`
      Severity   string `json:"severity"`
      File       string `json:"file"`
      Line       int    `json:"line"`
      FixPrompt  string `json:"fix_prompt"`
      FixExample string `json:"fix_example"`
      References []string `json:"references"`
  }

  type FixPlan struct {
      GeneratedAt     time.Time  `json:"generated_at"`
      TotalFindings   int        `json:"total_findings"`
      CriticalActions []FixItem  `json:"critical_actions"`
      HighActions     []FixItem  `json:"high_actions"`
      MediumActions   []FixItem  `json:"medium_actions"`
  }

  // GenerateFixPlan создаёт структурированный план исправлений без LLM.
  // Использует маппинг Rule ID → шаблон из templates.go
  func GenerateFixPlan(results []core.CheckResult) FixPlan {
      plan := FixPlan{
          GeneratedAt:   time.Now(),
          TotalFindings: len(results),
      }
      for _, r := range results {
          tmpl, ok := fixTemplates[r.ID]
          if !ok {
              tmpl = defaultTemplate
          }
          item := FixItem{
              RuleID:     r.ID,
              Name:       r.Name,
              Severity:   r.Severity,
              File:       r.File,
              Line:       r.Line,
              FixPrompt:  fmt.Sprintf(tmpl.Prompt, r.File, r.Line),
              FixExample: tmpl.Example,
              References: tmpl.References,
          }
          switch r.Severity {
          case "CRITICAL":
              plan.CriticalActions = append(plan.CriticalActions, item)
          case "HIGH":
              plan.HighActions = append(plan.HighActions, item)
          default:
              plan.MediumActions = append(plan.MediumActions, item)
          }
      }
      return plan
  }

  // ToMarkdown конвертирует план в Markdown для вставки в Claude Code / Cursor
  func (p FixPlan) ToMarkdown() string {
      var sb strings.Builder
      sb.WriteString("# AITriage Security Fix Plan\n")
      sb.WriteString(fmt.Sprintf("Generated: %s\n\n", p.GeneratedAt.Format("2006-01-02 15:04")))
      sb.WriteString(fmt.Sprintf("**Total findings: %d**\n\n", p.TotalFindings))

      if len(p.CriticalActions) > 0 {
          sb.WriteString("## 🔴 CRITICAL — Fix Immediately\n\n")
          for _, item := range p.CriticalActions {
              sb.WriteString(fmt.Sprintf("### %s\n", item.Name))
              sb.WriteString(fmt.Sprintf("**File:** `%s` line %d\n\n", item.File, item.Line))
              sb.WriteString(fmt.Sprintf("**Fix prompt for AI agent:**\n```\n%s\n```\n\n", item.FixPrompt))
              if item.FixExample != "" {
                  sb.WriteString(fmt.Sprintf("**Example:**\n```\n%s\n```\n\n", item.FixExample))
              }
              if len(item.References) > 0 {
                  sb.WriteString(fmt.Sprintf("**Reference:** %s\n\n", strings.Join(item.References, ", ")))
              }
          }
      }

      if len(p.HighActions) > 0 {
          sb.WriteString("## 🟠 HIGH\n\n")
          for _, item := range p.HighActions {
              sb.WriteString(fmt.Sprintf("- **%s** (`%s:%d`) — %s\n", item.Name, item.File, item.Line, item.FixPrompt))
          }
          sb.WriteString("\n")
      }

      if len(p.MediumActions) > 0 {
          sb.WriteString("## 🟡 MEDIUM\n\n")
          for _, item := range p.MediumActions {
              sb.WriteString(fmt.Sprintf("- **%s** (`%s:%d`) — %s\n", item.Name, item.File, item.Line, item.FixPrompt))
          }
      }

      return sb.String()
  }
  ```

### 6.2 — Создать шаблоны исправлений

- [x] Создать файл `internal/remedy/templates.go`:
  ```go
  package remedy

  type FixTemplate struct {
      Prompt     string   // %s = file, %d = line
      Example    string
      References []string
  }

  var defaultTemplate = FixTemplate{
      Prompt:     "Fix the security issue at %s line %d. Review the code and apply appropriate security controls.",
      References: []string{"https://owasp.org/www-project-top-ten/"},
  }

  var fixTemplates = map[string]FixTemplate{
      "ENTROPY-SECRET": {
          Prompt:  "Open %s line %d. The hardcoded secret must be moved to an environment variable. Replace the literal value with os.Getenv(\"VAR_NAME\") in Go, process.env.VAR_NAME in Node.js, or os.environ.get(\"VAR_NAME\") in Python. Add the variable to .env.example without the value.",
          Example: "// Before:\nconst apiKey = \"sk-abc123...\"\n\n// After:\nconst apiKey = os.Getenv(\"API_KEY\")\nif apiKey == \"\" {\n    log.Fatal(\"API_KEY env variable is required\")\n}",
          References: []string{"https://12factor.net/config", "https://owasp.org/www-project-top-ten/2021/A02_2021-Cryptographic_Failures"},
      },
      "ENTR-FRAGILE": {
          Prompt:  "Add error handling to %s (line %d area). The file has many lines but missing error handling patterns. Wrap risky operations in error checks.",
          Example: "// Before:\nresult := riskyOperation()\n\n// After:\nresult, err := riskyOperation()\nif err != nil {\n    return fmt.Errorf(\"operation failed: %w\", err)\n}",
          References: []string{"https://go.dev/blog/error-handling-and-go"},
      },
      "ENTR-04": {
          Prompt:  "Remove AI assistant chat residue from %s (line %d). Comments like \"As an AI\" or \"I cannot\" should not appear in production code. Clean up all such comments.",
          Example: "// Remove comments like:\n// 'As an AI language model, I cannot...'\n// 'I apologize, but as an AI...'\n// These are artifacts from AI code generation sessions.",
          References: []string{},
      },
  }
  ```

### 6.3 — Проверить компиляцию

- [x] `go build ./...`
- [x] `go vet ./...`

---

## Фаза 7: OWASP маппинг
> Цель: каждая находка имеет ссылку на категорию OWASP Top 10 2021.

### 7.1 — Создать файл маппинга

- [x] Создать файл `internal/scorer/owasp.go`:
  ```go
  package scorer

  // OWASPMap — маппинг Rule ID → OWASP Top 10 2021 категория
  var OWASPMap = map[string]string{
      "ENTROPY-SECRET":   "A02:2021 – Cryptographic Failures",
      "ENTR-FRAGILE":     "A04:2021 – Insecure Design",
      "ENTR-04":          "A04:2021 – Insecure Design",
      "ENTR-12":          "A04:2021 – Insecure Design",
      "missing_lockfile": "A06:2021 – Vulnerable and Outdated Components",
      "NFR-API-001":      "A04:2021 – Insecure Design",
      "NFR-API-002":      "A05:2021 – Security Misconfiguration",
      "NFR-API-003":      "A01:2021 – Broken Access Control",
      "NFR-ENV-002":      "A02:2021 – Cryptographic Failures",
  }

  // GetOWASP возвращает OWASP категорию для rule ID.
  // Если маппинга нет — возвращает пустую строку.
  func GetOWASP(ruleID string) string {
      return OWASPMap[ruleID]
  }
  ```

### 7.2 — Применить маппинг при сканировании

- [x] Открыть файл `internal/scanner/scanner.go`
- [x] После получения результатов сканирования добавить цикл:
  ```go
  for i := range report.Results {
      report.Results[i].OWASPMapping = scorer.GetOWASP(report.Results[i].ID)
  }
  ```
- [x] `go build ./...`

---

## Фаза 8: Distribution & DX
> Цель: установка за 10 секунд, интеграция с Claude Desktop одной командой.

### 8.1 — Обновить .goreleaser.yaml

- [ ] Открыть файл `.goreleaser.yaml`
- [ ] Заменить содержимое на:
  ```yaml
  before:
    hooks:
      - go mod tidy

  builds:
    - id: aitriage
      main: ./cmd/aitriage
      binary: aitriage
      env:
        - CGO_ENABLED=0
      goos:
        - linux
        - darwin
        - windows
      goarch:
        - amd64
        - arm64
      ldflags:
        - -s -w -X main.Version={{.Version}}

  archives:
    - format: tar.gz
      name_template: >-
        {{ .ProjectName }}_
        {{- .Version }}_
        {{- title .Os }}_
        {{- if eq .Arch "amd64" }}x86_64
        {{- else if eq .Arch "arm64" }}arm64
        {{- else }}{{ .Arch }}{{ end }}
      format_overrides:
        - goos: windows
          format: zip

  checksum:
    name_template: 'checksums.txt'

  snapshot:
    name_template: "{{ incpatch .Version }}-next"

  changelog:
    sort: asc
    filters:
      exclude:
        - '^docs:'
        - '^test:'
        - '^chore:'
  ```

### 8.2 — Создать команду `install-mcp`

- [x] Создать файл `cmd/aitriage/install_mcp.go`:
  ```go
  package main

  import (
      "encoding/json"
      "fmt"
      "os"
      "path/filepath"
      "runtime"

      "github.com/spf13/cobra"
  )

  var installMCPCmd = &cobra.Command{
      Use:   "install-mcp",
      Short: "Install AITriage as MCP server in Claude Desktop",
      Long:  "Automatically adds AITriage to Claude Desktop MCP configuration.",
      RunE:  runInstallMCP,
  }

  func init() {
      rootCmd.AddCommand(installMCPCmd)
  }

  func runInstallMCP(cmd *cobra.Command, args []string) error {
      binaryPath, err := os.Executable()
      if err != nil {
          return fmt.Errorf("cannot determine binary path: %w", err)
      }

      configPath, err := claudeDesktopConfigPath()
      if err != nil {
          // Если не нашли — вывести инструкцию для ручной установки
          printManualInstall(binaryPath)
          return nil
      }

      // Читаем существующий конфиг или создаём новый
      data, err := os.ReadFile(configPath)
      var config map[string]interface{}
      if err != nil {
          config = map[string]interface{}{}
      } else {
          json.Unmarshal(data, &config)
      }

      // Добавляем AITriage
      if config["mcpServers"] == nil {
          config["mcpServers"] = map[string]interface{}{}
      }
      mcpServers := config["mcpServers"].(map[string]interface{})
      mcpServers["aitriage"] = map[string]interface{}{
          "command": binaryPath,
          "args":    []string{"serve"},
      }

      // Сохраняем
      out, _ := json.MarshalIndent(config, "", "  ")
      if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
          return err
      }
      if err := os.WriteFile(configPath, out, 0644); err != nil {
          return err
      }

      fmt.Printf("✅ AITriage added to Claude Desktop MCP config:\n%s\n\n", configPath)
      fmt.Println("Restart Claude Desktop to apply changes.")
      return nil
  }

  func claudeDesktopConfigPath() (string, error) {
      var configDir string
      switch runtime.GOOS {
      case "darwin":
          configDir = filepath.Join(os.Getenv("HOME"), "Library", "Application Support", "Claude")
      case "linux":
          configDir = filepath.Join(os.Getenv("HOME"), ".config", "Claude")
      case "windows":
          configDir = filepath.Join(os.Getenv("APPDATA"), "Claude")
      default:
          return "", fmt.Errorf("unsupported OS: %s", runtime.GOOS)
      }
      return filepath.Join(configDir, "claude_desktop_config.json"), nil
  }

  func printManualInstall(binaryPath string) {
      snippet := map[string]interface{}{
          "mcpServers": map[string]interface{}{
              "aitriage": map[string]interface{}{
                  "command": binaryPath,
                  "args":    []string{"serve"},
              },
          },
      }
      data, _ := json.MarshalIndent(snippet, "", "  ")
      fmt.Println("Add this to your Claude Desktop config manually:")
      fmt.Println(string(data))
  }
  ```

### 8.3 — Создать GitHub Release workflow

- [x] Создать файл `.github/workflows/release.yml`:
  ```yaml
  name: Release

  on:
    push:
      tags:
        - 'v*'

  permissions:
    contents: write

  jobs:
    release:
      runs-on: ubuntu-latest
      steps:
        - uses: actions/checkout@v4
          with:
            fetch-depth: 0

        - uses: actions/setup-go@v5
          with:
            go-version: '1.24'
            cache: true

        - uses: goreleaser/goreleaser-action@v6
          with:
            version: latest
            args: release --clean
          env:
            GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
  ```

---

## Фаза 9: Тесты & CI
> Цель: автоматическая проверка что ничего не сломалось при каждом PR.

### 9.1 — Обновить ci.yml

- [x] Открыть файл `.github/workflows/ci.yml`
- [x] Заменить содержимое на:
  ```yaml
  name: CI

  on:
    push:
      branches: [main]
    pull_request:
      branches: [main]

  jobs:
    test:
      runs-on: ubuntu-latest
      steps:
        - uses: actions/checkout@v4

        - uses: actions/setup-go@v5
          with:
            go-version: '1.24'
            cache: true

        - name: Download dependencies
          run: go mod download

        - name: Build
          run: go build -v ./...

        - name: Vet
          run: go vet ./...

        - name: Test
          run: go test -v -race -coverprofile=coverage.out ./...

        - name: Check goreleaser
          uses: goreleaser/goreleaser-action@v6
          with:
            version: latest
            args: check
  ```

### 9.2 — Починить aitriage-shield.yml

- [x] Открыть файл `.github/workflows/aitriage-shield.yml`
- [x] Заменить содержимое на:
  ```yaml
  name: AITriage Self-Scan

  on:
    push:
      branches: [main, master]
    pull_request:
      branches: [main, master]

  jobs:
    shield:
      name: AITriage Security Scan
      runs-on: ubuntu-latest
      steps:
        - uses: actions/checkout@v4

        - uses: actions/setup-go@v5
          with:
            go-version: '1.24'
            cache: true

        - name: Install AITriage
          run: go install ./cmd/aitriage

        - name: Run AITriage
          run: aitriage scan ./
          # exit code 1 = critical findings → PR blocked автоматически
  ```

### 9.3 — Написать unit-тесты

#### Тест scanner
- [x] Создать файл `internal/scanner/scanner_test.go`:
  ```go
  package scanner_test

  import (
      "context"
      "testing"
      "github.com/cybertortuga/aitriage/internal/scanner"
  )

  func TestScanReturnsReport(t *testing.T) {
      report, err := scanner.Scan(context.Background(), "../../examples/nextjs-terrible", scanner.ScanOptions{})
      if err != nil {
          t.Fatalf("Scan failed: %v", err)
      }
      if report.TotalFiles == 0 {
          t.Error("Expected TotalFiles > 0")
      }
      if len(report.Results) == 0 {
          t.Error("Expected findings in a terrible example project")
      }
  }
  ```

#### Тест external runner
- [x] Создать файл `internal/external/runner_test.go`:
  ```go
  package external_test

  import (
      "context"
      "testing"
      "github.com/cybertortuga/aitriage/internal/external"
  )

  func TestIsInstalled_Go(t *testing.T) {
      // Go всегда установлен в тест-окружении
      if !external.IsInstalled("go") {
          t.Error("Expected go to be installed")
      }
  }

  func TestRunTool_Echo(t *testing.T) {
      result, err := external.RunTool(context.Background(), "echo", "hello")
      if err != nil {
          t.Fatalf("RunTool failed: %v", err)
      }
      if result.ExitCode != 0 {
          t.Errorf("Expected exit code 0, got %d", result.ExitCode)
      }
  }
  ```

#### Тест remedy
- [x] Создать файл `internal/remedy/fix_plan_test.go`:
  ```go
  package remedy_test

  import (
      "testing"
      "github.com/cybertortuga/aitriage/internal/core"
      "github.com/cybertortuga/aitriage/internal/remedy"
  )

  func TestGenerateFixPlan_Critical(t *testing.T) {
      results := []core.CheckResult{
          {ID: "ENTROPY-SECRET", Name: "Hardcoded Secret", Severity: "CRITICAL", File: "main.go", Line: 42},
      }
      plan := remedy.GenerateFixPlan(results)
      if plan.TotalFindings != 1 {
          t.Errorf("Expected 1 finding, got %d", plan.TotalFindings)
      }
      if len(plan.CriticalActions) != 1 {
          t.Errorf("Expected 1 critical action, got %d", len(plan.CriticalActions))
      }
      md := plan.ToMarkdown()
      if md == "" {
          t.Error("Expected non-empty markdown")
      }
  }
  ```

#### Тест NFR
- [x] Создать файл `internal/nfr/checker_test.go`:
  ```go
  package nfr_test

  import (
      "os"
      "path/filepath"
      "testing"
      "github.com/cybertortuga/aitriage/internal/nfr"
  )

  func TestCheckNFR_FindsMissingDotEnvExample(t *testing.T) {
      // Создать временный проект без .env.example
      tmpDir := t.TempDir()
      os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte("package main"), 0644)

      rulesDir := "rules"
      findings, err := nfr.CheckNFR(tmpDir, rulesDir)
      if err != nil {
          t.Skipf("Rules dir not found: %v", err)
      }

      found := false
      for _, f := range findings {
          if f.RuleID == "NFR-ENV-001" {
              found = true
              break
          }
      }
      if !found {
          t.Error("Expected NFR-ENV-001 (.env.example missing) to trigger")
      }
  }
  ```

### 9.4 — Финальная проверка

- [x] `go build ./...` — без ошибок
- [x] `go vet ./...` — без ошибок
- [x] `go test ./...` — все тесты проходят
- [x] `go run ./cmd/aitriage --help` — все команды видны: scan, agent, serve, install-mcp, version
- [x] `go run ./cmd/aitriage scan ./examples/nextjs-terrible` — выводит находки
- [x] `go run ./cmd/aitriage serve --help` — описание команды есть

---

## Покрытие task.md

| Скоуп (task.md) | Фаза | Файлы | Статус |
|---|---|---|---|
| Сканеры: semgrep | 2 | `internal/external/semgrep.go` + MCP tool | ✅ |
| Сканеры: bandit | 2 | `internal/external/bandit.go` + MCP tool | ✅ |
| Сканеры: gitleaks | 2 | `internal/external/gitleaks.go` + MCP tool | ✅ |
| Сканеры: trivy | 2 | `internal/external/trivy.go` + MCP tool | ✅ |
| NFR compliance | 5 | `internal/nfr/` + `internal/nfr/rules/web_api.yaml` | ✅ |
| Аудит инфры: VM/IaC | 5 | `internal/deployaudit/audit.go` | ✅ |
| Прощупать домены/порты | — | Out of scope (требует сеть) | ⛔ |
| Построить схему | 5 | `internal/architect/diagram.go` | ✅ |
| Секреты (где/как) | 0+1 | entropy в движке + MCP tool | ✅ |
| Файлы деплоя | 5 | `internal/deployaudit/audit.go` | ✅ |
| Git-анализ | 5 | `internal/gitanalysis/git.go` | ✅ |
| Натуральный язык | 4 | `cmd/aitriage/agent.go` | ✅ |
| OpenSource | ✅ | MIT licence | ✅ |
| Fix instructions для AI | 4+6 | `internal/remedy/` + agent prompt | ✅ |
| Прожарка архитектуры | 0+1 | engine rules + `aitriage_architecture` tool | ✅ |
| AI-lab agnostic | 3 | `internal/llm/` (OpenAI + Anthropic + Ollama) | ✅ |
| Простая установка | 8 | `go install` + goreleaser | ✅ |
| Плагины Claude Code/Cursor | 1+8 | MCP server + `install-mcp` команда | ✅ |
| Параллельное сканирование | 4 | `runParallelScan()` в agent.go | ✅ |
| Сборка артефактов | 4 | UnifiedFinding + FixPlan | ✅ |
| Консолидация результатов | 4+6 | GenerateFixPlan + LLM анализ | ✅ |
| Сборка отчёта | 4+6 | ToMarkdown() + LLM summary | ✅ |
| Рекомендации пользователю | 4 | LLM analysis output | ✅ |
| Промпт для AI-агента | 4 | fix_plan.md = готовый промпт | ✅ |
| Консультация Q&A по NFR | 4 | `runConsultation()` в agent.go | ✅ |
| Консультация Q&A по спеке | 4 | `runConsultation()` в agent.go | ✅ |
| Консультация Q&A по уязвимостям | 4 | `runConsultation()` в agent.go | ✅ |

> **⛔ Out of scope:** сканирование живых доменов и портов — это active recon, требует сеть и выходит за рамки SAST инструмента.

---

## Структура файлов после реализации

```
aitriage/
├── cmd/aitriage/
│   ├── main.go
│   ├── root.go           ← + var Version = "dev"
│   ├── scan.go           ← + context.Context
│   ├── serve.go          ← НОВЫЙ
│   ├── agent.go          ← НОВЫЙ
│   ├── version.go        ← НОВЫЙ
│   ├── remedy.go         ← обновлён
│   └── install_mcp.go    ← НОВЫЙ
├── internal/
│   ├── llm/              ← НОВАЯ
│   │   ├── client.go
│   │   ├── openai.go
│   │   ├── anthropic.go
│   │   ├── factory.go
│   │   └── prompts.go
│   ├── mcp/              ← НОВАЯ
│   │   ├── server.go
│   │   ├── tools_scan.go
│   │   ├── tools_secrets.go
│   │   ├── tools_entropy.go
│   │   ├── tools_architecture.go
│   │   ├── tools_fixplan.go
│   │   ├── tools_scanners.go
│   │   ├── tools_external.go
│   │   ├── tools_git.go
│   │   ├── tools_deploy.go
│   │   ├── tools_nfr.go
│   │   ├── tools_diagram.go
│   │   └── resources.go
│   ├── external/         ← НОВАЯ
│   │   ├── runner.go
│   │   ├── finding.go
│   │   ├── semgrep.go
│   │   ├── gitleaks.go
│   │   ├── trivy.go
│   │   └── bandit.go
│   ├── gitanalysis/      ← НОВАЯ
│   │   ├── git.go
│   │   └── runner.go
│   ├── architect/        ← НОВАЯ
│   │   └── diagram.go
│   ├── deployaudit/      ← НОВАЯ
│   │   └── audit.go
│   ├── nfr/              ← НОВАЯ
│   │   ├── checker.go
│   │   └── rules/
│   │       └── web_api.yaml
│   ├── remedy/           ← обновлена
│   │   ├── fix_plan.go
│   │   └── templates.go
│   ├── scorer/
│   │   ├── scorer.go     (без изменений)
│   │   └── owasp.go      ← НОВЫЙ
│   ├── scanner/scanner.go ← + context, + json tags
│   ├── core/context.go   ← + json tags, + OWASP field
│   ├── ast/              (без изменений)
│   ├── config/           (без изменений — только добавить LLM поле)
│   ├── detector/         (без изменений)
│   ├── engine/           (без изменений)
│   ├── loader/           (без изменений)
│   ├── models/           (без изменений)
│   └── entropy/             (без изменений)
├── .github/workflows/
│   ├── ci.yml            ← обновлён
│   ├── release.yml       ← НОВЫЙ
│   └── aitriage-shield.yml ← починен
├── .goreleaser.yaml      ← обновлён
└── README.md             ← обновить после всех фаз
```

---

## Оценка трудозатрат

| Фаза | Файлов | Оценка |
|---|---|---|
| Фаза 0: Подготовка | 3 | 2 часа |
| Фаза 1: MCP-сервер | 9 | 2-3 дня |
| Фаза 2: Внешние сканеры | 7 | 1-2 дня |
| Фаза 3: LLM Layer | 5 | 1-2 дня |
| Фаза 4: Agent Mode | 1 | 1-2 дня |
| Фаза 5: Инструменты | 7 | 2-3 дня |
| Фаза 6: Remediation | 2 | 1 день |
| Фаза 7: OWASP | 1 | 2 часа |
| Фаза 8: Distribution | 3 | 1 день |
| Фаза 9: Тесты & CI | 7 | 1 день |
| **ИТОГО** | **~45 файлов** | **~15-18 дней** |
