<div align="center">

```
     _    ___ _____ ____  ___    _    ____ _____ 
    / \  |_ _|_   _|  _ \|_ _|  / \  / ___| ____|
   / _ \  | |  | | | |_) || |  / _ \| |  _|  _|  
  / ___ \ | |  | | |  _ < | | / ___ \ |_| | |___ 
 /_/   \_\___|_|_| |_| \_\___/_/   \_\____|_____|
              Security Audit Engine
```

**Deterministic Security Scanner for AI-Generated Codebases**

[![Go Report Card](https://goreportcard.com/badge/github.com/cybertortuga/aitriage?style=flat-square)](https://goreportcard.com/report/github.com/cybertortuga/aitriage)
[![GitHub Release](https://img.shields.io/github/v/release/cybertortuga/aitriage?style=flat-square&color=blue)](https://github.com/cybertortuga/aitriage/releases)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg?style=flat-square)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/github/go-mod/go-version/cybertortuga/aitriage?style=flat-square)](https://go.dev/)
[![CI](https://img.shields.io/github/actions/workflow/status/cybertortuga/aitriage/ci.yml?style=flat-square&label=CI)](https://github.com/cybertortuga/aitriage/actions)
[![OpenSSF Scorecard](https://img.shields.io/badge/OpenSSF-Scorecard-brightgreen?style=flat-square)](https://securityscorecards.dev/viewer/?uri=github.com/cybertortuga/aitriage)
[![Docker Pulls](https://img.shields.io/docker/pulls/cybertortuga/aitriage?style=flat-square)](https://hub.docker.com/r/cybertortuga/aitriage)
[![Contributor Covenant](https://img.shields.io/badge/Contributor%20Covenant-2.1-4baaaa.svg?style=flat-square)](CODE_OF_CONDUCT.md)

</div>

---

## Why AITriage?

AI coding assistants generate code fast — but they also generate **vulnerabilities fast**. AITriage is a security scanner specifically designed for the post-AI development era. It catches the patterns that traditional SAST tools miss: hardcoded secrets disguised as variables, unreviewed LLM scaffolding, chat residue in production code, and happy-path logic with zero error handling.

<!-- DEMO: Record with `vhs` or `asciinema rec` and replace this comment with:
![demo](docs/demo.gif)
-->

## Quick Start

```bash
# Install via Homebrew (macOS / Linux)
brew install cybertortuga/aitriage/aitriage

# Install via Go
go install github.com/cybertortuga/aitriage/cmd/aitriage@latest

# Initialize project (generates .aitriage.yaml, CI workflow, IDE config)
aitriage init

# Scan your project
aitriage scan .

# Launch interactive TUI dashboard
aitriage scan . -i
```

## Core Capabilities

| Capability | Description |
|---|---|
| **AST Analysis** | Tree-sitter powered scanning for Go, Python, TypeScript/JavaScript. Finds SQLi, XSS, CSRF, path traversal at the syntax level. |
| **Entropy Detection** | Shannon Entropy analysis catches hardcoded secrets even with non-obvious variable names. Detects AI chat residue and scaffolding artifacts. |
| **Silent Luxury TUI** | Professional interactive dashboard for real-time audit triage, project navigation, and vulnerability review. |
| **MCP Native** | Model Context Protocol server provides deep security context directly to AI assistants (Claude, Cursor, Windsurf). |
| **Orchestration** | Wraps and unifies reports from Semgrep, Trivy, Gitleaks, and Bandit into a single SARIF 2.1 stream. |
| **AI Agent Mode** | LLM-driven triage with automated prioritization and fix generation. |

## How It Works

AITriage uses a **single-pass O(N) concurrent engine**. Files are streamed through a unified pipeline — AST queries, entropy checks, and configuration audits run simultaneously. Zero redundant I/O.

```
Files ──► Loader ──► [ AST Engine + Entropy Engine + Config Auditor ] ──► Scorer ──► Report
                              (concurrent, single pass)
```

## Commands

```bash
# Core
aitriage scan .                    # Deterministic security scan
aitriage scan . --format sarif     # SARIF 2.1 output for CI/CD
aitriage scan . -o results.sarif   # Write SARIF to file, TUI to stdout
aitriage scan . -i                 # Interactive TUI dashboard

# Incremental Scanning
aitriage scan . --diff HEAD~1      # Only files changed since last commit
aitriage scan . --diff origin/main # Only files changed vs main branch
aitriage scan . --staged           # Only git-staged files (pre-commit)

# Baseline Management (suppress alert fatigue on legacy codebases)
aitriage baseline create .         # Accept current findings as baseline
aitriage baseline show .           # Show baseline statistics
aitriage scan . --baseline         # Report ONLY new regressions

# AI-Powered Remediation
aitriage fix .                     # Generate fix diffs for all findings
aitriage fix . --dry-run           # Preview fixes without applying
aitriage fix . --severity high     # Only fix HIGH+ severity
aitriage fix . --auto              # Auto-apply safe fixes (LOW/MEDIUM)

# Watch Mode (Sentinel)
aitriage watch .                   # Real-time file watcher + incremental scan
aitriage watch . --debounce 500    # Custom debounce (ms)
aitriage watch . --quiet           # Only show findings

# SBOM Generation
aitriage sbom .                    # CycloneDX 1.5 to stdout
aitriage sbom . --format spdx      # SPDX 2.3 format
aitriage sbom . -o sbom.json       # Save to file

# Rule Packs (Plugin System)
aitriage rules list                # List installed & available packs
aitriage rules install owasp-api-2025  # Install from registry
aitriage rules install ./my-rules/ # Install from local directory
aitriage rules remove owasp-api-2025   # Remove a rule pack

# Setup & Integration
aitriage init                      # Generate config, CI, IDE integration
aitriage init --ci --pre-commit    # + GitHub Actions workflow + git hook
aitriage install-mcp               # Configure as MCP server for your IDE
aitriage agent .                   # AI-powered audit with remediation
```

## Built-in Rules Ecosystem

AITriage ships with **180+ security rules** across 11 technology stacks, loaded directly from [`rules/`](rules/) at compile time. Add a YAML file — it's automatically included.

| Stack | Rules | Key Detections |
|---|---|---|
| [Universal](rules/universal/) | 28 | Secrets, weak crypto, SSRF, NoSQL injection, AI residue, prototype pollution |
| [Next.js / React](rules/nextjs/) | 28 | XSS, SQLi, CSRF, SSRF, CSP, JWT abuse, command injection |
| [FastAPI](rules/fastapi/) | 22 | SSTI, eval/exec, pickle, SSRF, sync-in-async, path traversal |
| [Flask](rules/flask/) | 14 | Debug mode, SSTI, raw SQL, pickle, open redirect, unsafe cookies |
| [Django](rules/django/) | 16 | DEBUG, SECRET_KEY, CSRF, XSS, raw SQL, mass assignment, middleware |
| [Express.js](rules/express/) | 14 | Helmet, NoSQL injection, child_process, SSRF, session security |
| [Go](rules/golang/) | 14 | SSRF, command injection, TLS skip, math/rand, error swallowing |
| [Python](rules/python/) | 13 | subprocess shell=True, pickle, yaml.load, eval/exec, weak hashing |
| [LLM / AI Security](rules/llm/) | 10 | OWASP LLM Top 10: prompt injection, output exec, excessive agency |
| [Docker / IaC](rules/docker/) | 11 | Running as root, privileged, secrets in ENV, curl pipe to shell |
| [ASP.NET Core](rules/aspnetcore/) | 10 | XXE, deserialization, path traversal, insecure CORS |

Custom rules can be added in `.aitriage.yaml`. See the [Rules Documentation](rules/README.md) for the full schema reference.

## CI/CD Integration

### GitHub Actions

AITriage ships as a **pre-built Docker Action** (image published to GHCR), so runs take seconds — not minutes. Use a **two-layer model**: a deterministic SARIF gate that blocks merges, plus an optional non-blocking AI advisor that posts triage + fix suggestions on PRs.

```yaml
name: AITriage Security
on: [push, pull_request]
permissions:
  contents: read
  security-events: write   # upload SARIF to the Security tab
  pull-requests: write     # AI advisor comment (optional)

jobs:
  # ── Layer 1: deterministic gate (blocks merge) ──
  gate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with: { fetch-depth: 0 }
      - uses: cybertortuga/aitriage@v1
        with:
          health-profile: standard   # baseline | standard | strict
          fail-on: critical
          fail-score: 70            # optional minimum Health Check score
          baseline: 'true'          # legacy debt does not block; only new regressions
          format: sarif
          output-file: aitriage-results.sarif
      - uses: github/codeql-action/upload-sarif@v3
        if: always()
        with:
          sarif_file: aitriage-results.sarif

  # ── Layer 2: AI advisor (does NOT block) ──
  ai-advisor:
    if: github.event_name == 'pull_request'
    needs: gate
    continue-on-error: true
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: cybertortuga/aitriage@v1
        with:
          command: agent
          args: '--no-chat --report-out report.md --fixspec-out FIXSPEC.md'
        env:
          GEMINI_API_KEY: ${{ secrets.GEMINI_API_KEY }}
      # → post report.md / FIXSPEC.md as a PR comment or upload as an artifact
```

**Action inputs:** `command`, `project-dir`, `format`, `output-file`, `fail-on`, `fail-score`, `health-profile`, `stack`, `diff`, `baseline`, `args`.

**Health Check gate:** `security_score` and `security_grade` stay in JSON for
compatibility. The CI/CD pass/fail decision is `health_check.verdict.passed`.
Use `health-profile` or `.aitriage.yaml` `health_check:` to choose the IB policy:
`baseline` keeps the compatible default, `standard` enforces a score threshold
for sensitive/business apps, and `strict` blocks any active finding. Legacy
`strict_mode` and `fail_score` still work when `health_check` is not configured.

### Docker

```bash
# Interactive TUI
make docker-tui

# CI/CD scan with JSON output
make docker-scan

# Web dashboard at http://localhost:8080
make docker-web
```

## Enterprise Setup

AITriage Enterprise provides high-authority auditing, RBAC, and executive reporting.

### 1. Environment Configuration

Ensure the following environment variables are set for production:

```bash
# Security
JWT_SECRET=your-enterprise-secret-key-min-32-chars
# AI Analysis
GEMINI_API_KEY=your-gemini-key
# Database (Optional, default: ~/.aitriage/aitriage.db)
AITRIAGE_DB_PATH=/path/to/enterprise.db
```

### 2. Role-Based Access Control (RBAC)

AITriage enforces strict role-based access. Default roles include:

- `superadmin`: Full system access, team management.
- `admin`: Team-level administration, engagement management.
- `security_lead`: Audit approval, report generation.
- `analyst`: Findings triage, AI fix validation.
- `developer`: View findings, implement fixes.
- `viewer`: Read-only access to dashboards.

### 3. Deployment

Start the full enterprise stack (Web UI, API, SQLite persistence) using Docker Compose:

```bash
make enterprise-up
# Or manually: docker compose up -d
```
See [docs/DEPLOYMENT.md](docs/DEPLOYMENT.md) for the complete production deployment guide.

### 4. Hardening & Security

- **Rate Limiting**: Integrated per-IP protection on sensitive endpoints (Login, Reports).
- **Security Headers**: HSTS, CSP, and X-Frame-Options enforced by default.
- **Audit Logging**: Every write operation is recorded in the `audit_logs` table for compliance.

### 4. Executive Reporting

## Project Structure

AITriage organizes its codebase into distinct layers separating Go business logic, AST checkers, Web UI layers, and mockup structures:

- [`cmd/`](cmd/) — Executable entry points for `aitriage` CLI, Triage testing tool, and `test_rules` utility.
- [`internal/`](internal/) — Core Go implementation including scanner engine, database logic, telemetry, and API handlers.
- [`rules/`](rules/) — Static security YAML rule files structured by technology stack (Django, Go, FastAPI, Next.js, etc.).
- [`web/`](web/) — React/TypeScript Vite web application (Simple & Advanced UI modes).
- [`docs/`](docs/) — Manual documentation, structural schemas, and guidelines.
  - [`docs/design_mockups/`](docs/design_mockups/) — Interactive design mockups, components, screenshots, and visual specifications.
- [`testdata/`](testdata/) — Synthetic and third-party benchmark applications used to verify rule checkers.

## Audit Scope

- **Source Code** — AST-level vulnerability detection (SQLi, Taint, CSRF, XSS, SSTI).
- **Entropy** — AI scaffolding, chat residue, hallucinated packages, sensitive data leaks.
- **Containers** — Dockerfile and docker-compose audits for root execution and privileged modes.
- **Architecture** — Mermaid diagram generation and NFR compliance checks.
- **Git History** — Secret leak detection across the entire commit log.

## Roadmap

- [x] O(N) Single-Pass Engine Architecture
- [x] "Silent Luxury" TUI Redesign
- [x] MCP Server Native Implementation
- [x] SARIF 2.1 Output with `--out` flag
- [x] `aitriage init` — Full Onboarding Wizard
- [x] Baseline Management (`--baseline`)
- [x] Incremental Scanning (`--diff`, `--staged`)
- [x] Inline Suppression (`// aitriage-ignore: RULE-ID`)
- [x] Pre-commit Hook Support
- [x] GoReleaser with Homebrew Tap
- [x] Watch Mode / Sentinel (`aitriage watch .`)
- [x] Plugin System — Downloadable Rule Packs (`aitriage rules install`)
- [x] SBOM Generation — CycloneDX 1.5 / SPDX 2.3 (`aitriage sbom .`)
- [x] AI-Powered Remediation (`aitriage fix .`)
- [ ] Compliance mapping (SOC 2, ISO 27001, NIST)

## License

MIT — see [LICENSE](LICENSE) for details.

---

<div align="center">
  <sub>Built for high-authority security auditing. &copy; 2026 Tortuga Co.</sub>
</div>
