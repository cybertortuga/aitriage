# Contributing to AITriage

Thank you for your interest in contributing to AITriage. This guide covers the
basics of setting up a development environment, running tests, and submitting
changes.

## Prerequisites

- **Go 1.25.12+** — [go.dev/dl](https://go.dev/dl/)
- **Node.js 22+** — for the web frontend
- **CGO enabled** — tree-sitter requires C compilation (`CGO_ENABLED=1`)
- **git** — for version control and diff-based scanning tests

## Getting Started

```bash
git clone https://github.com/cybertortuga/aitriage.git
cd aitriage
go mod download
make build          # Build the binary to bin/aitriage
make test           # Run the full test suite
```

## Development Workflow

### Build & Test

```bash
make build          # Build binary to bin/aitriage
make test           # Run all Go tests with verbose output
make lint           # Run golangci-lint
make format         # Format Go + web sources
make clean          # Remove build artifacts
```

### Web Frontend

```bash
make build-web      # Build the frontend (npm install + npm run build)
make sync-web       # Build frontend and sync assets into internal/server/ui/dist
cd web && npm run dev   # Start Vite dev server with HMR on :5173
```

The dev server proxies `/api` to the Go backend on `http://localhost:8080`.

### Docker

```bash
make docker-build              # Build the local Docker image
make docker-tui                # Run TUI in Docker (scans current directory)
make docker-web                # Web dashboard in Docker on :8080
make up                        # Full stack via docker compose
```

### Release (local snapshot)

```bash
make release        # Run GoReleaser snapshot (does not publish)
```

## Coding Standards

### Go

- Run `make format` before committing — it applies `go fmt`, `goimports`,
  `gofumpt`, and `gci` (if installed).
- Run `make lint` — CI enforces `golangci-lint` with a 5-minute timeout.
- Follow standard Go conventions: effective error handling, context propagation,
  and table-driven tests.
- Comments should be in English, on exported identifiers and non-obvious logic.

### Web (TypeScript / React)

- Run `cd web && npm run format` — applies Prettier + ESLint auto-fix.
- Follow the existing component structure: `pages/` for route views,
  `components/` for reusable UI, `hooks/` for custom hooks, `store/` for
  Zustand state.

### Rules

Security rules live in `rules/<stack>/security.yaml` and are mirrored in the
embedded `default_rules.yaml`. See [rules/README.md](rules/README.md) for the
rule schema, target modes, conditions, and guidelines for writing custom rules.

When adding or modifying a rule:

1. Add the rule to the appropriate `rules/<stack>/security.yaml`.
2. Test the regex with `grep -P` before committing.
3. Add a test case in the relevant `testdata/` directory.
4. Run `aitriage scan ./testdata` to verify the rule fires.
5. Map the rule to an OWASP category in the suggestion text.

## Commit Messages

Use [Conventional Commits](https://www.conventionalcommits.org/):

```
feat(agent): add retry logic for transient 429 errors
fix(scanner): correct regex for Flask debug flag
docs(readme): update rule count to 187
chore(deps): bump tree-sitter to v0.22
```

- `feat:` — new feature
- `fix:` — bug fix
- `docs:` — documentation only
- `refactor:` — code change that neither fixes a bug nor adds a feature
- `test:` — adding or correcting tests
- `chore:` — tooling, deps, CI

## Pull Request Process

1. **Fork** the repository and create a branch from `main`:
   ```bash
   git checkout -b feat/my-feature
   ```
2. **Write tests** for your changes. We do not merge features without test
   coverage.
3. **Run checks locally**:
   ```bash
   make format
   make lint
   make test
   ```
4. **Commit** with conventional commit messages.
5. **Open a pull request** — fill in the PR template completely.
6. **Address review feedback** — push additional commits (do not squash until
   requested).

### CI Checks

All PRs run through GitHub Actions:

- `go build` with `CGO_ENABLED=1`
- `go vet`
- `golangci-lint`
- `go test -race -coverprofile`
- GoReleaser config validation

PRs must pass all checks before merge.

## Project Structure

```
cmd/aitriage/          CLI commands
internal/agent/        AI agent: graph orchestrator, LLM clients, MCP server
internal/engine/       Core audit engine, baseline, history
internal/scanner/      AST, entropy, external, NFR, deploy, network scanners
internal/server/       Web API server, handlers, SQLite
internal/healthpolicy/ Security policy profiles and verdict computation
internal/report/       Health check and reporter
internal/telemetry/    Usage tracking
internal/ui/tui/       Terminal UI dashboard
rules/                 187 security rules across 11 categories
web/                   React 19 + TypeScript + Vite frontend
testdata/              Sample vulnerable repos for engine testing
examples/              Example GitHub Actions workflows
```

## Reporting Issues

- **Bug reports** — use the GitHub issue template. Include OS, AITriage version
  (`aitriage version`), and a minimal reproduction.
- **Feature requests** — describe the use case and expected behaviour.
- **Security vulnerabilities** — see [SECURITY.md](SECURITY.md). Do not open a
  public issue for security vulnerabilities.

## License

By contributing, you agree that your contributions will be licensed under the
[MIT License](LICENSE).
