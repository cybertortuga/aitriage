<div align="center">

```
     _    ___ _____ ____  ___    _    ____ _____ 
    / \  |_ _|_   _|  _ \|_ _|  / \  / ___| ____|
   / _ \  | |  | | | |_) || |  / _ \| |  _|  _|  
  / ___ \ | |  | | |  _ < | | / ___ \ |_| | |___ 
 /_/   \_\___|_|_| |_| \_\___/_/   \_\____|_____|
              Security Audit Engine
```

**Deterministic Security Scanner & AI-Powered Triage for Modern Codebases**

[![Go Report Card](https://goreportcard.com/badge/github.com/cybertortuga/aitriage?style=flat-square)](https://goreportcard.com/report/github.com/cybertortuga/aitriage)
[![GitHub Release](https://img.shields.io/github/v/release/cybertortuga/aitriage?style=flat-square&color=blue)](https://github.com/cybertortuga/aitriage/releases)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg?style=flat-square)](LICENSE)
[![Go Version](https://img.shields.io/github/go-mod/go-version/cybertortuga/aitriage?style=flat-square)](https://go.dev/)
[![CI](https://img.shields.io/github/actions/workflow/status/cybertortuga/aitriage/ci.yml?style=flat-square&label=CI)](https://github.com/cybertortuga/aitriage/actions)
[![Docker Pulls](https://img.shields.io/docker/pulls/cybertortuga/aitriage?style=flat-square)](https://hub.docker.com/r/cybertortuga/aitriage)

</div>

---

## Why AITriage?

AI coding assistants generate code at light speed — but they also propagate **security vulnerabilities** just as fast. AITriage is a hybrid security scanner designed specifically for the post-AI software development era. It bridges the gap between deterministic pattern matching and intelligent context analysis by catching what traditional SAST tools often miss:

*   **Hardcoded secrets** hidden in complex AI structures.
*   **Unreviewed LLM scaffold residue** and boilerplate left in production.
*   **Happy-path logic** generated with zero error handling.
*   **Hallucinated dependencies** and packages that could lead to supply-chain attacks.

---

## How It Works

AITriage utilizes a **single-pass O(N) concurrent audit engine** written in Go. Code files are loaded and streamed simultaneously through the AST, Entropy, and Config engines. There is zero redundant disk I/O, allowing you to run scans in seconds.

```
Files ──► Loader ──► [ AST Engine + Entropy Engine + Config Auditor ] ──► Scorer ──► Report
                              (concurrent, single pass)
```

---

## Core Capabilities

| Capability | Description |
| :--- | :--- |
| **AST Analysis** | Tree-sitter powered scanning for Go, Python, and TypeScript/JavaScript. Tracks SQLi, XSS, CSRF, and path traversal at the syntax level. |
| **Entropy Engine** | Shannon Entropy analysis catches high-entropy variables and hardcoded keys, plus AI chat remnants. |
| **Silent Luxury TUI** | Professional interactive terminal dashboard for audit triage, code browsing, and real-time review. |
| **MCP Native** | Model Context Protocol server exposing security context tools directly to AI assistants (Cursor, Claude, Windsurf). |
| **Orchestration** | Wraps and unifies findings from Semgrep, Trivy, Gitleaks, and Bandit into a single consolidated stream. |
| **AI Agent Mode** | LLM-driven map-reduce analysis that automates false-positive filtering and compiles prioritizations. |
| **Auto-Remediation** | Generates fix diffs for detected vulnerabilities using local policies or LLM models. |

---

## Quick Start

```bash
# Install via Homebrew (macOS / Linux)
brew install cybertortuga/aitriage/aitriage

# Install via Go
go install github.com/cybertortuga/aitriage/cmd/aitriage@latest

# Initialize your project configuration, CI workflows, and IDE settings
aitriage init

# Run a deterministic security scan
aitriage scan .

# Run the interactive TUI dashboard
aitriage scan . -i
```

---

## Commands Reference

### Core & Scanning
```bash
aitriage scan .                    # Basic scan
aitriage scan . --format json      # Structured JSON output
aitriage scan . --format sarif     # SARIF 2.1 stream for CI platforms
aitriage scan . -o results.sarif   # Write SARIF to file and display TUI logs
```

### Incremental Scanning
```bash
aitriage scan . --diff HEAD~1      # Scan files changed since the previous commit
aitriage scan . --diff origin/main # Scan files changed compared to the main branch
aitriage scan . --staged           # Scan git-staged changes (ideal for pre-commit hooks)
```

### Baseline Management
Avoid alert fatigue on legacy codebases. Accepting current findings as a baseline suppresses old alerts, allowing the scanner to notify you only about new regressions.
```bash
aitriage baseline create .         # Save current security status as baseline
aitriage baseline show .           # Show current baseline statistics
aitriage scan . --baseline         # Scan and hide baseline findings (fails only on new code)
```

### AI-Powered Remediation
```bash
aitriage fix .                     # Generate fix specifications for issues
aitriage fix . --dry-run           # Preview changes without editing files
aitriage fix . --severity high     # Only generate fixes for high+ issues
```

### Sentinel (Watch Mode)
```bash
aitriage watch .                   # Run background sentinel that watches file edits
aitriage watch . --debounce 500    # Set debouncing timeout in milliseconds
```

### SBOM Generation
```bash
aitriage sbom .                    # Generate CycloneDX 1.5 SBOM format
aitriage sbom . --format spdx      # Generate SPDX 2.3 format
```

### Plugin & Rule Packs
```bash
aitriage rules list                # List all built-in and external rules
aitriage rules install owasp-2025  # Install specific package from registry
```

### Setup & IDE Integration
```bash
aitriage init                      # Launch onboarding setup wizard
aitriage init --ci --pre-commit    # Generate config + pre-commit hook + GHA workflow
aitriage install-mcp               # Install AITriage as an MCP Server
```

---

## Built-in Rules Ecosystem

AITriage ships with **180+ static security rules** across 11 technology stacks, loaded directly from the [rules/](file:///Users/afedotov/Documents/GitHub/aitriage/rules/) directory at compile time.

| Technology | Rules | Key Detections |
| :--- | :--- | :--- |
| **[Universal](file:///Users/afedotov/Documents/GitHub/aitriage/rules/universal/)** | 28 | Plaintext keys, weak cryptography, SSRF, prototype pollution, AI residue |
| **[Next.js / React](file:///Users/afedotov/Documents/GitHub/aitriage/rules/nextjs/)** | 28 | Cross-site Scripting (XSS), server-side injection, raw DOM nodes |
| **[FastAPI](file:///Users/afedotov/Documents/GitHub/aitriage/rules/fastapi/)** | 22 | Unsafe pickle loaders, SSTI, synchronous database calls inside async handlers |
| **[Flask](file:///Users/afedotov/Documents/GitHub/aitriage/rules/flask/)** | 14 | Dev debug flags, SSTI, unescaped templates, insecure cookies |
| **[Django](file:///Users/afedotov/Documents/GitHub/aitriage/rules/django/)** | 16 | Missing CSRF middleware, raw SQL exec, DEBUG mode enabled |
| **[ExpressJS](file:///Users/afedotov/Documents/GitHub/aitriage/rules/express/)** | 14 | Missing helmet protection, NoSQL injection patterns, shell child processes |
| **[Go](file:///Users/afedotov/Documents/GitHub/aitriage/rules/golang/)** | 14 | SSRF, unsafe pointers, crypto/rand package omission, error swallowing |
| **[Python](file:///Users/afedotov/Documents/GitHub/aitriage/rules/python/)** | 13 | YAML unsafe loading, subprocess shells, eval/exec execution |
| **[LLM / AI Security](file:///Users/afedotov/Documents/GitHub/aitriage/rules/llm/)** | 10 | OWASP Top 10 for LLMs: prompt injections, execution flows, excessive agency |
| **[Docker / IaC](file:///Users/afedotov/Documents/GitHub/aitriage/rules/docker/)** | 11 | Root user configurations, privileged containers, secret leakage in env keys |
| **[ASP.NET Core](file:///Users/afedotov/Documents/GitHub/aitriage/rules/aspnetcore/)** | 10 | Deserialization flaws, unsafe XML parsing (XXE), CORS wildcards |

---

## Docker Auto-Escalation

To run comprehensive audits, AITriage orchestrates external scanners (e.g., `semgrep`, `trivy`, `gitleaks`, `bandit`). 
If these utilities are **missing locally** but a Docker daemon is active, AITriage **transparently re-launches itself in a container** (using the pre-built GHCR image `ghcr.io/cybertortuga/aitriage:latest`). This process is completely seamless and ensures you get full AST and secret audits without manually installing dependencies.

---

## CI/CD Pipeline Architecture

AITriage uses a **Two-Layer Pipeline Model** for GitHub Actions. It is published as a pre-compiled Docker action, bypassing build times and running in seconds instead of minutes.

```
                  ┌──────────────────────────────┐
                  │      GitHub Actions Run      │
                  └──────────────┬───────────────┘
                                 │
                 ┌───────────────┴───────────────┐
                 │       actions/checkout        │
                 └───────────────┬───────────────┘
                                 │
                ┌────────────────┴────────────────┐
                │                                 │
  ┌─────────────▼─────────────┐     ┌─────────────▼─────────────┐
  │ Layer 1: Deterministic    │     │ Layer 2: AI Advisor       │
  │ Gate (Blocks Pull Request)│     │ (Non-blocking review)     │
  └─────────────┬─────────────┘     └─────────────┬─────────────┘
                │                                 │
        aitriage scan .                   aitriage agent --no-chat
   verdict = health_check.passed                  │
  SARIF → Security Dashboard                      │
  Annotations → PR diff             Post triage details & fixspecs
  Summary → Job step summary        as a comments on the pull request
```

### GitHub Actions Workflow Example

Create `.github/workflows/aitriage.yml` in your repository:

```yaml
name: AITriage Security Pipeline

on:
  workflow_dispatch:

permissions:
  contents: read

jobs:
  static-scan:
    name: Static Security Scan
    runs-on: ubuntu-latest
    steps:
      - name: Checkout Code
        uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - name: Run AITriage Scanner
        uses: cybertortuga/aitriage@v1
        with:
          command: 'scan'
          args: '--no-summary'
          format: 'html'
          output-file: 'report.html'
          fail-on: never

      - name: Upload HTML Security Report
        if: always()
        uses: actions/upload-artifact@v4
        with:
          name: aitriage-security-report
          path: report.html

  ai-triage:
    name: AI Triage & Fix Specs
    needs: static-scan
    runs-on: ubuntu-latest
    steps:
      - name: Checkout Code
        uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - name: Run AI Triage Agent (SecureCoder Rules)
        env:
          GEMINI_API_KEY: ${{ secrets.GEMINI_API_KEY }}
        uses: cybertortuga/aitriage@v1
        with:
          command: 'agent'
          args: '--no-chat --report-out report.md --fixspec-out fixspec.md'

      - name: Publish AI Triage Report & Fix Specs to GitHub Summary
        if: always()
        run: |
          if [ -f report.md ]; then
            echo "### AITriage AI Audit Report" >> $GITHUB_STEP_SUMMARY
            cat report.md >> $GITHUB_STEP_SUMMARY
          fi
          if [ -f fixspec.md ]; then
            echo "### AI IDE Fix Prompt" >> $GITHUB_STEP_SUMMARY
            echo '```markdown' >> $GITHUB_STEP_SUMMARY
            cat fixspec.md >> $GITHUB_STEP_SUMMARY
            echo '```' >> $GITHUB_STEP_SUMMARY
          fi

      - name: Upload AI Triage Artifacts
        if: always()
        uses: actions/upload-artifact@v4
        with:
          name: aitriage-ai-triage-results
          path: |
            report.md
            fixspec.md
```

> [!IMPORTANT]
> **AI Keys & Provider Auto-Detection**:
> - **LLM Key Storage**: Never hardcode API keys. Store them securely in your repository secrets (e.g. `secrets.GEMINI_API_KEY`, `secrets.OPENAI_API_KEY`, or `secrets.ANTHROPIC_API_KEY`) and map them under the `env:` block of your action step.
> - **Provider Auto-Detection**: The AITriage Agent automatically detects the LLM provider based on which API key environment variable is set (`GEMINI_API_KEY` for Google Gemini, `OPENAI_API_KEY` for OpenAI, `ANTHROPIC_API_KEY` for Anthropic).

---

## Information Security Policy Gates

Instead of simple pass/fail checks, AITriage calculates a comprehensive Security Score and evaluates a deterministic policy verdict (`health_check.verdict.passed`).

### 1. Built-in Security Profiles
You can configure a profile via the `health-profile` action parameter or `.aitriage.yaml`:

*   **`baseline`** (Default): Blocks only active `CRITICAL` and `HIGH` findings. General codebase score is informational.
*   **`standard`** (Sensitive/Business apps): Enforces a minimum codebase score of `70` and blocks any active `CRITICAL` or `HIGH` vulnerabilities.
*   **`strict`** (High-assurance systems): Blocks on *any* active vulnerability (critical, high, or medium) and requires a minimum score of `90`.

### 2. Configuration Options
Configure your security policy details in [.aitriage.yaml](file:///Users/afedotov/Documents/GitHub/aitriage/.aitriage.yaml.example):

```yaml
health_check:
  profile: baseline
  fail_on: critical       # critical | any | never
  minimum_score: 70       # Fail if general score falls below this value
  max_critical: 0         # Max allowed active critical findings
  max_high: 2            # Max allowed active high findings
  max_medium: 5
  block_sources:
    - gitleaks            # Explicitly fail if gitleaks finds active secrets
  block_classes:
    - hardcoded-secret    # Block any hardcoded secrets regardless of severity
```

> [!TIP]
> **Baseline Gating (`--baseline`)**: If your codebase has legacy technical debt, run `aitriage baseline create .` locally. When `--baseline` is enabled in CI, AITriage suppresses old findings and recalculates the policy verdict on new changes only. Legacy issues will not fail your build.

---

## Enterprise Deployment

AITriage Enterprise provides multi-repo dashboards, role-based access controls (RBAC), and persistent audit logs.

### 1. Environment Configuration
Set the following keys for production enterprise nodes:

```bash
JWT_SECRET=your-32-character-secret-key-for-api-authentication
GEMINI_API_KEY=your-gemini-key
DB_PATH=/var/lib/aitriage/production.db
```

### 2. Role-Based Access Controls (RBAC)
AITriage enforces granular roles:
*   `superadmin` / `admin`: Full system configurations and team setups.
*   `security_lead`: Audit policy sign-offs and report reviews.
*   `analyst`: Finding triaging, validating AI-generated fixes, and false-positive marking.
*   `developer`: Viewing project findings and applying security fixes.
*   `viewer`: Read-only reporting dashboards.

### 3. Startup Stack
Start the enterprise stack (Web UI, API server, and SQLite storage) via Docker Compose:

```bash
# Start the stack in background daemon mode
docker compose up -d
```
See [docs/DEPLOYMENT.md](file:///Users/afedotov/Documents/GitHub/aitriage/docs/DEPLOYMENT.md) for full system deployment details.

---

## Project Structure

*   [cmd/](file:///Users/afedotov/Documents/GitHub/aitriage/cmd/) — CLI commands and sub-command definitions.
*   [internal/](file:///Users/afedotov/Documents/GitHub/aitriage/internal/) — Core Go library, AST query processing engines, scoring, and telemetry logic.
*   [rules/](file:///Users/afedotov/Documents/GitHub/aitriage/rules/) — Static security rule patterns grouped by stack.
*   [web/](file:///Users/afedotov/Documents/GitHub/aitriage/web/) — Vite-powered React/TypeScript web app.
*   [docs/](file:///Users/afedotov/Documents/GitHub/aitriage/docs/) — Architecture details, guidelines, and manuals.
*   [testdata/](file:///Users/afedotov/Documents/GitHub/aitriage/testdata/) — Standard sample repositories containing security flaws for engine testing.

---

## Roadmap

- [x] Concurrent O(N) Scanning Architecture
- [x] Premium TUI Dashboard
- [x] Model Context Protocol (MCP) Server
- [x] Full Git Baseline support (`--baseline`)
- [x] Incremental git-diff audits (`--diff`, `--staged`)
- [x] Information Security Policy Gate & Verdict system
- [x] Watch Sentinel engine (`aitriage watch`)
- [x] Rule Pack Package Management (`aitriage rules`)
- [x] CycloneDX / SPDX SBOM exports (`aitriage sbom`)
- [x] AI-Triage & Remediation engine (`aitriage fix`)
- [ ] Compliance Mappings (SOC 2, ISO 27001, OWASP Top 10)

---

## License

Distributed under the MIT License. See [LICENSE](LICENSE) for details.

---

<div align="center">
  <sub>Designed and built for high-assurance security triaging. &copy; 2026 Tortuga Co.</sub>
</div>
