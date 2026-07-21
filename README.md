<div align="center">

<pre>
    _     ___  _____  ____   ___     _      ____  _____ 
   / \   |_ _||_   _||  _ \ |_ _|   / \    / ___|| ____|
  / _ \   | |   | |  | |_) | | |   / _ \  | |  _ |  _|  
 / ___ \  | |   | |  |  _ <  | |  / ___ \ | |_| || |___ 
/_/   \_\|___|  |_|  |_| \_\|___|/_/   \_\ \____||_____|

                    Security Audit Engine
</pre>

**Deterministic Security Scanner & AI-Powered Triage for Modern Codebases**

[![GitHub Release](https://img.shields.io/github/v/release/cybertortuga/aitriage?style=flat-square&color=blue)](https://github.com/cybertortuga/aitriage/releases)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg?style=flat-square)](LICENSE)
[![Go Version](https://img.shields.io/github/go-mod/go-version/cybertortuga/aitriage?style=flat-square)](https://go.dev/)
[![CI](https://img.shields.io/github/actions/workflow/status/cybertortuga/aitriage/ci.yml?style=flat-square&label=CI)](https://github.com/cybertortuga/aitriage/actions)
[![GHCR](https://img.shields.io/badge/GHCR-cybertortuga%2Faitriage-2496ED?style=flat-square&logo=github)](https://github.com/cybertortuga/aitriage/pkgs/container/aitriage)

</div>

---

## At a Glance

- **Deterministic evidence first:** 187 built-in rules and integrated external scanners collect findings, SARIF, annotations, and artifacts without failing a trusted CI run on untriaged results.
- **Hardcoded-secret & leak blocking:** deterministic signatures for common credential formats (AWS/GitHub/Slack/Stripe/OpenAI/Google/SendGrid/npm/Twilio tokens, private-key blocks, JWTs) plus credentialed connection strings and a keyword-guarded assignment heuristic — all reported as CRITICAL so the `--staged` pre-commit hook and CI gate block them before push. Evidence is redacted; the full secret is never re-emitted.
- **Mandatory AI gate in the primary CI workflow:** AI triage removes false positives, writes the authoritative summary, then is the sole policy gate. If the AI provider or agent fails, the workflow fails closed.
- **Deterministic AI caching:** verdict and artifact-bundle caches make repeat CI runs on the same code cheap — cached verdicts skip per-finding LLM classification, and an exact artifact hit skips PoC/report/fixspec generation entirely.
- **Built for local development and CI:** run a deterministic scan locally without an LLM, use the hardened GitHub Actions workflow for trusted code, or expose security context through MCP.
- **Go 1.25.5 for source builds:** released binaries and the Homebrew cask do not require a local Go toolchain.

---

## Why AITriage?

AI coding assistants generate code at light speed — but they also propagate **security vulnerabilities** just as fast. AITriage is a hybrid security scanner designed specifically for the post-AI software development era. It bridges the gap between deterministic pattern matching and intelligent context analysis by catching what traditional SAST tools often miss:

- **Hardcoded secrets** hidden in complex AI-generated structures.
- **Unreviewed LLM scaffold residue** and boilerplate left in production code.
- **Happy-path logic** generated with zero error handling or input validation.
- **Hallucinated dependencies** and packages that could lead to supply-chain attacks.

---

## How It Works

AITriage uses a **single-pass O(N) concurrent audit engine** written in Go. Code files are loaded and streamed simultaneously through the AST, Entropy, and Config engines. There is zero redundant disk I/O, allowing scans to complete in seconds.

```
Files ──► Loader ──► [ AST Engine + Entropy Engine + Config Auditor ] ──► Scorer ──► Report
                              (concurrent, single pass)
```

For AI-powered triage, the agent orchestrator runs a multi-stage pipeline:

```
Scan ──► Context Enrichment ──► Threat Model (TP/FP/NR) ──► PoC Verification
                                                              │
                                                              ▼
                          Health Check ◄── Report ◄── Fix Specification
                                                              │
                                                              ▼
                                          Summary + Canonical JSON Inventory
```

---

## Core Capabilities

| Capability | Description |
| :--- | :--- |
| **AST Analysis** | Tree-sitter powered scanning for Go, Python, and TypeScript/JavaScript. Tracks SQLi, XSS, CSRF, and path traversal at the syntax level. |
| **Entropy Engine** | Shannon Entropy analysis catches high-entropy variables, hardcoded keys, and AI chat remnants. |
| **Secret & Leak Detection** | Vendor-format signatures (AWS, GitHub, Slack, Stripe, OpenAI, Anthropic, Google, SendGrid, npm, Twilio), private-key blocks, JWTs, and credentialed connection strings (postgres/mysql/mongodb/redis/…), reported as CRITICAL with redacted evidence. Detection escalates when a secret is also written to a log/print sink. |
| **Interactive TUI** | Professional terminal dashboard for audit triage, code browsing, and real-time review. |
| **MCP Server** | Model Context Protocol server exposing 18 security tools and 2 resources to AI assistants (Cursor, Claude, Windsurf). |
| **External Orchestration** | Wraps and unifies findings from Semgrep, Trivy, Gitleaks, and Bandit into a single consolidated stream. |
| **AI Agent Mode** | LLM-driven triage that classifies every finding, suppresses false positives, and produces a full report, actionable summary, fix specification, and canonical JSON inventory. |
| **AI Response Caching** | Two deterministic cache layers keyed by content fingerprints: per-finding verdict reuse plus exact reuse of the PoC/report/fixspec bundle on identical re-runs. |
| **AI IDE Remediation Brief** | Gives an AI IDE the verified finding context and a secure operating contract: audit and plan first, implement only confirmed true positives, then verify; manual-review items are not changed speculatively. |
| **Web Dashboard** | Browser-based React/TypeScript dashboard with multi-project scanning, RBAC, and persistent SQLite storage. |
| **SBOM Generation** | CycloneDX 1.5 and SPDX 2.3 Software Bill of Materials exports from the dependency graph. |
| **Git Baseline** | Accept current findings as a baseline to suppress legacy alerts and track only new regressions. |
| **Sentinel Watch** | Background file-watcher that runs incremental scans on every edit with configurable debouncing. |

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

# See all supported commands and flags
aitriage --help
```

### LLM Provider Configuration

AI-powered commands (`agent`, `fix`, `preaudit`) auto-detect the provider from environment variables:

| Provider | Env Variable | Default Model |
| :--- | :--- | :--- |
| **Google Gemini** (default) | `GEMINI_API_KEY` | `gemini-2.5-flash` |
| **Anthropic Claude** | `ANTHROPIC_API_KEY` | `claude-sonnet-4-5` |
| **OpenAI** | `OPENAI_API_KEY` | `gpt-4o` |
| **Ollama** (local) | — (no key needed) | configured via `--model` |
| **Groq** | `GROQ_API_KEY` | configured via `--model` |

Override with `--provider` and `--model` flags, or configure permanently in [`.aitriage.yaml`](.aitriage.yaml.example).

---

## Commands Reference

### Scanning
```bash
aitriage scan .                              # Basic scan
aitriage scan . --format json                # Structured JSON output
aitriage scan . --format sarif               # SARIF 2.1 stream for CI platforms
aitriage scan . --format sarif -o results.sarif  # Write SARIF to a file
aitriage scan . --format html                # HTML report
aitriage scan . --health-profile standard    # Apply the standard policy profile
aitriage scan . -i                           # Interactive TUI dashboard
```

### Incremental Scanning
```bash
aitriage scan . --diff HEAD~1                # Scan files changed since the previous commit
aitriage scan . --diff origin/main           # Scan files changed compared to the main branch
aitriage scan . --staged                     # Scan git-staged changes (pre-commit hooks)
```

### Baseline Management
Suppress alert fatigue on legacy codebases. Accepting current findings as a baseline hides old alerts, so the scanner notifies you only about new regressions.
```bash
aitriage baseline create .                   # Save current security status as baseline
aitriage baseline show .                     # Show current baseline statistics
aitriage baseline update .                   # Re-scan and update the baseline
aitriage baseline clear .                    # Remove the baseline file
aitriage scan . --baseline                   # Scan and hide baseline findings
```

### AI Agent Mode
```bash
aitriage agent .                             # Scan + LLM triage + report + fix spec (interactive Q&A)
aitriage agent . --no-chat                   # Skip interactive Q&A (CI/CD)
aitriage agent . --provider gemini --model gemini-2.5-pro
aitriage agent . --health-profile standard --fail-on any \
  --summary-out summary.md --report-out report.md \
  --fixspec-out fixspec.md --triage-out triage-findings.json
```

### AI-Powered Remediation
```bash
aitriage fix .                               # Generate fix specifications for issues
aitriage fix . --dry-run                     # Preview changes without editing files
aitriage fix . --severity high               # Only generate fixes for high+ issues
aitriage fix . --auto                        # Auto-apply safe fixes (LOW/MEDIUM)

# Deterministic fixes: dry run by default; writes only with --apply
aitriage autofix .                           # Preview deterministic fixes
aitriage autofix . --apply                   # Write fixes to disk
```

### Pre-Audit & Spec Generation
```bash
aitriage preaudit --arch "Next.js, Go API, Postgres"  # NFR check before writing code
aitriage generate-spec .                     # Generate CLAUDE.md AI-agent spec from scan results
```

### Sentinel (Watch Mode)
```bash
aitriage watch .                             # Background sentinel watching file edits
aitriage watch . --debounce 500              # Custom debounce (ms)
aitriage watch . --fail-on critical          # Exit non-zero on critical findings
```

### SBOM Generation
```bash
aitriage sbom .                              # CycloneDX 1.5 to stdout
aitriage sbom . --format spdx                # SPDX 2.3 format
aitriage sbom . --format cyclonedx -o sbom.json  # Save to file
```

### Rule Packs
```bash
aitriage rules list                          # List installed and available rule packs
aitriage rules install owasp-api-2025        # Install from registry
aitriage rules install ./my-rules/           # Install from local directory
aitriage rules remove owasp-api-2025         # Remove a rule pack
aitriage rules info owasp-api-2025           # Show details about a rule pack
```

### Setup & IDE Integration
```bash
aitriage init                                # Onboarding wizard: config + IDE settings
aitriage init --ci --pre-commit              # Generate config + pre-commit hook + GHA workflow
aitriage init --mcp                          # Configure MCP server for Claude Desktop
aitriage install-mcp                         # Install AITriage as MCP server in Claude Desktop
aitriage serve                               # Run MCP server over stdio
aitriage serve --transport sse --port 9090   # Run MCP server over SSE
```

### Web Dashboard
```bash
# Start from the repository checkout
make up                                      # Build frontend + backend, launch in Docker

# Or scan a mounted host filesystem with the published image
docker run --rm -p 8080:8080 -v /:/host:ro \
  ghcr.io/cybertortuga/aitriage:v1 web

# Full enterprise stack (Web UI + API + SQLite)
make enterprise-up
```

Open `http://localhost:8080` after the service starts. Default login: `admin` / `admin`.

---

## AI Response Caching

Repeat CI runs on the same commit should not pay for the same LLM analysis twice. The agent ships two deterministic, opt-in cache layers (both disabled unless a cache directory is configured):

| Layer | File | What it saves |
| :--- | :--- | :--- |
| **Verdict cache** | `triage_cache.json` | The TP/FP/Needs-Review verdict per finding fingerprint. On a warm cache, per-finding LLM classification drops to zero calls. |
| **Artifact bundle cache** | `artifact_bundle_cache.json` | The complete PoC-verification results, `report.md`, and `fixspec.md` for an exact repeat run. On an exact hit those LLM stages are skipped entirely. |

Configuration:

```bash
AITRIAGE_CACHE_DIR=.aitriage-cache            # Enables both layers in one directory
AITRIAGE_VERDICT_CACHE_DIR=...                # Optional: separate verdict cache location
AITRIAGE_ARTIFACT_CACHE_DIR=...               # Optional: separate artifact bundle location
```

In GitHub Actions, pass the directories via the `verdict-cache-dir` and `artifact-cache-dir` action inputs and persist them with `actions/cache` (see the [canonical workflow](examples/github-actions/aitriage-security.yml) for the exact restore/save wiring, including the save condition that requires both cache files to exist).

Safety properties, by design:

- **Exact matches only.** Cache keys are content hashes over provider, model, prompt versions, embedded-rules digest, health policy, ordered finding fingerprints, and ordered disposition hashes. Any material change invalidates the cache; a partial GitHub cache restore can never be treated as a hit.
- **A full artifact hit requires a full verdict hit** — stale report/fixspec text cannot survive a changed verdict.
- **The security gate is never cached.** The health check, policy verdict, and summary are recomputed on every run from current dispositions.
- **No secrets persisted.** Bundles or verdicts containing secret-shaped values (API keys, PATs, JWTs, AWS keys) are skipped, never written to cache.
- **Fully auditable.** `triage-findings.json` records `verdict_cache` and `artifact_cache` telemetry (hits, misses, stores, computed key, miss reason), and the run log prints the cache key and hit/miss status.

---

## Built-in Rules Ecosystem

AITriage ships with **187 security rules** across **11 categories**, loaded directly from the embedded `default_rules.yaml` at compile time. The browsable rule catalog is maintained in [rules/README.md](rules/README.md).

| Category | Rules | Key Detections |
| :--- | :--- | :--- |
| **[Universal](rules/universal/)** | 26 | Plaintext keys, weak cryptography, SSRF, prototype pollution, AI residue |
| **[Next.js / React](rules/nextjs/)** | 28 | Cross-site scripting (XSS), server-side injection, raw DOM nodes |
| **[FastAPI](rules/fastapi/)** | 25 | Unsafe pickle loaders, SSTI, synchronous DB calls inside async handlers |
| **[Django](rules/django/)** | 16 | Missing CSRF middleware, raw SQL execution, DEBUG mode enabled |
| **[Flask](rules/flask/)** | 14 | Dev debug flags, SSTI, unescaped templates, insecure cookies |
| **[Express.js](rules/express/)** | 14 | Missing helmet protection, NoSQL injection, shell child processes |
| **[Go](rules/golang/)** | 14 | SSRF, unsafe pointers, crypto/rand omission, error swallowing |
| **[Python](rules/python/)** | 12 | YAML unsafe loading, subprocess shells, eval/exec execution |
| **[LLM / AI Security](rules/llm/)** | 10 | OWASP LLM Top 10: prompt injection, execution flows, excessive agency |
| **[Docker / IaC](rules/docker/)** | 11 | Root user configs, privileged containers, secret leakage in env keys |
| **[ASP.NET Core](rules/aspnetcore/)** | 10 | Deserialization flaws, XXE, CORS wildcards |
| **Shannon Entropy** | Built-in | High-entropy variables, hardcoded keys, binary analysis (all languages) |

Rules map to **OWASP Top 10:2025** and **OWASP LLM Top 10:2025**. See [rules/README.md](rules/README.md) for the full mapping, rule schema, and instructions for writing custom rules.

---

## Docker & External Scanners

The published Docker image (`ghcr.io/cybertortuga/aitriage`) is an all-in-one runtime with all external scanners pre-installed:

| Tool | Version | Purpose |
| :--- | :--- | :--- |
| **Semgrep** | latest (pipx) | Multi-language static analysis |
| **Bandit** | latest (pipx) | Python-specific security linter |
| **Gitleaks** | v8.30.1 | Secret detection in code and Git history |
| **Trivy** | v0.70.0 | Container, dependency, and IaC scanning |

### Docker Auto-Escalation

If external scanners are **missing locally** but a Docker daemon is active, AITriage transparently re-launches itself in a container. The local fallback follows `ghcr.io/cybertortuga/aitriage:latest`; this is separate from the published GitHub Action, which references the maintained `:v1` image.

### Docker Make Targets

```bash
make docker-build                            # Build the local Docker image
make docker-tui                              # TUI in Docker (scans current directory)
make docker-tui PROJECT=/path/to/app         # TUI in Docker (specific project)
make docker-web                              # Web dashboard in Docker
make docker-scan                             # CLI scan with JSON output (CI/CD)
```

---

## CI/CD Pipeline Architecture

AITriage is published as a pre-built Docker Action. Consumers use `cybertortuga/aitriage@v1`; the Action references the maintained `ghcr.io/cybertortuga/aitriage:v1` image, which the `docker-publish` workflow rebuilds and pushes on every merge to `main` and on tagged releases. Because the image is built once and pulled at run time, Action runtime drops from a full Dockerfile build to a few seconds.

The primary workflow has a deliberate trust boundary: raw scanner output is evidence, not a merge decision. Mandatory AI triage is the only security-policy gate.

```
trusted same-repository PR / main push / manual dispatch
                         │
                         ▼
       deterministic collection: scan --no-summary --fail-on never
                         │
       SARIF + annotations + artifact (evidence only; never blocks)
                         │
                         ▼
     restore AI caches (verdict + artifact bundle, exact keys only)
                         │
                         ▼
     mandatory AI triage: agent --health-profile standard --fail-on any
                         │
       authoritative three-block Job Summary after completed triage
                         │
       save AI caches · fails on any remaining True Positive or score below 70
```

### Install the Primary Workflow

Copy the [canonical workflow](examples/github-actions/aitriage-security.yml) to `.github/workflows/aitriage.yml`. It pins third-party Actions to reviewed commits and contains the complete static evidence, SARIF, artifact, AI-cache, and mandatory AI-triage flow. Do not copy an abbreviated workflow from an old issue or README snippet.

Before the first run:

1. Create the repository variable `AITRIAGE_ALLOWED_ACTOR_ID` with the numeric GitHub account ID permitted to start jobs. An empty or mismatched value skips all jobs before checkout or secret access.
2. Create the `ai-triage` environment and add `GEMINI_API_KEY` as an **environment secret**. Restrict eligible branches and use required reviewers when your GitHub plan supports them.
3. Protect the workflow file and repository access. YAML cannot stop an administrator or writer from changing the allowlist.
4. Run `workflow_dispatch` once from the permitted account, then make **AI Triage & Fix Specs** the required branch-protection check. Do not make deterministic collection a required check: it is evidence-only.

### Example Workflows

| Workflow | Description |
| :--- | :--- |
| [aitriage-security.yml](examples/github-actions/aitriage-security.yml) | **Primary:** deterministic evidence + mandatory AI triage gate (recommended) |
| [aitriage-pr-gate.yml](examples/github-actions/aitriage-pr-gate.yml) | Fast deterministic-only gate for PRs (no LLM required) |
| [aitriage-ai-advisor.yml](examples/github-actions/aitriage-ai-advisor.yml) | Non-blocking AI triage with report and fix specs (advisory) |
| [aitriage-manual-html-report.yml](examples/github-actions/aitriage-manual-html-report.yml) | On-demand HTML security report via manual dispatch |

### Dual Output: Actionable Summary vs Full Report

The completed AI agent produces separate outputs to maximise signal-to-noise ratio:

| Output | Contains | Destination |
| :--- | :--- | :--- |
| **Job Summary / `summary.md`** | Security assessment, AI IDE implementation brief, structured TP/Needs Review data | `$GITHUB_STEP_SUMMARY` and optional artifact file |
| **Full Report** (`report.md`) | All findings, including false-positive rationale | AI-triage artifact on a successful agent run |
| **Fix Specification** (`fixspec.md`) | Detailed remediation specification | AI-triage artifact on a successful agent run |
| **Canonical Inventory** (`triage-findings.json`) | Machine-readable record of every finding and disposition, health check, LLM usage per stage, and cache telemetry | AI-triage artifact on a successful agent run |

- The scanner never writes a raw Job Summary. The agent writes the only authoritative summary after all required AI stages complete, even when the resulting policy verdict fails.
- False positives are counted in the assessment but excluded from the actionable prompt and structured AI data. The full report retains their rationale as an audit trail.
- The AI IDE brief requires a scoped audit and written plan before changes, implements only confirmed true positives, and leaves `Needs Manual Review` to a human decision.

---

## Information Security Policy Gates

AITriage calculates a comprehensive Security Score and evaluates a policy verdict (`health_check.verdict.passed`). In the canonical GitHub Actions workflow, that verdict is applied only after AI triage has removed false positives.

### Built-in Security Profiles

Configure a profile via the `health-profile` action parameter or `.aitriage.yaml`:

- **`baseline`** (Default): Blocks only active `CRITICAL` and `HIGH` findings. General codebase score is informational.
- **`standard`** (Sensitive/Business apps): Enforces a minimum codebase score of `70` and blocks any active `CRITICAL` or `HIGH` vulnerabilities.
- **`strict`** (High-assurance systems): Blocks on *any* active vulnerability (critical, high, or medium) and requires a minimum score of `90`.

### Configuration

Copy [`.aitriage.yaml.example`](.aitriage.yaml.example) to `.aitriage.yaml` and adjust:

```yaml
health_check:
  profile: baseline       # baseline | standard | strict
  fail_on: critical       # critical | any | never
  minimum_score: 0        # Fail if Health Check score < this value (0 disables score gating)
  max_critical: -1        # -1 = unlimited, 0 = block any active critical finding
  max_high: -1            # -1 = unlimited, 0 = block any active high finding
  max_medium: -1          # -1 = unlimited, 0 = block any active medium finding
  block_sources: []       # e.g. ["gitleaks"] — fail if a listed scanner finds active issues
  block_classes: []       # e.g. ["hardcoded-secret"] — block a finding class regardless of severity

llm:
  provider: gemini        # gemini | openai | anthropic | ollama
  model: gemini-2.5-flash
  api_key: $GEMINI_API_KEY  # Never hardcode — reference an environment variable
```

> [!TIP]
> **Baseline Gating (`--baseline`)**: If your codebase has legacy technical debt, run `aitriage baseline create .` locally. When `--baseline` is enabled in CI, AITriage suppresses old findings and recalculates the policy verdict on new changes only. Legacy issues will not fail your build.

---

## MCP Server

AITriage exposes a Model Context Protocol server with 18 security tools and 2 resources, allowing AI assistants like Claude, Cursor, and Windsurf to query security context directly.

### Tools

All tools perform read-only security analysis except:
- `securecoder_ignore` — mutates suppression state (suppresses a finding); requires explicit approval.
- `aitriage_diff` — runs a scan and persists a snapshot under the project's own `.aitriage/history/` so it can diff against next time. It never edits your source or suppresses findings; the only write is that local scan-history cache.

The `safe` server profile (see `aitriage serve --profile safe`) forbids finding suppression and any change to your source: the mutating `securecoder_ignore` is not registered. It still permits the benign local scan-history cache used by `aitriage_diff` (written only under the project's own `.aitriage/history/`). Every path argument is confined to a single project directory.

| Tool | Description |
| :--- | :--- |
| `aitriage_scan` | Run a full deterministic security scan (AST + Shannon entropy). No LLM required; returns structured JSON |
| `aitriage_secrets` | Detect hardcoded secrets via entropy analysis (API keys, tokens, passwords) |
| `aitriage_entropy_check` | Detect chat residue in comments, missing error handling, God Files, TODO stubs, `.cursorrules` tampering |
| `aitriage_architecture` | Analyze project structure and tech stacks (call this first) |
| `generate_fix_plan` | Generate a markdown fix plan with actionable prompts per finding |
| `list_available_scanners` | Report which external scanners are installed in PATH |
| `run_semgrep` | Run Semgrep (requires it installed) |
| `run_gitleaks` | Run Gitleaks to detect hardcoded secrets |
| `run_trivy` | Run Trivy for dependency/IaC vulnerabilities (`fs` or `config`) |
| `run_bandit` | Run Bandit on Python projects |
| `run_securecoder` | Full SecureCoder scan with unified findings |
| `run_securecoder_deps` | Check packages for known vulnerabilities before importing |
| `securecoder_ignore` | **(mutating)** Suppress a SecureCoder finding as FP / accepted risk / won't fix |
| `aitriage_deploy_audit` | Audit Dockerfile / docker-compose for root, privileged mode, hardcoded secrets |
| `aitriage_nfr_check` | Check NFR compliance: rate limiting, CORS, unprotected routes |
| `aitriage_diagram` | Generate a Mermaid architecture diagram |
| `aitriage_last_scan` | Show the most recent saved scan result and SecurityScore |
| `aitriage_diff` | Run a fresh scan and diff it against the previous saved scan |

### Resources

- **Security Playbook** — step-by-step remediation guidance
- **Secure Coding Guidelines** — language-specific best practices

### Usage

```bash
# stdio (default, for Claude Desktop and IDE integration)
aitriage serve

# SSE (for remote clients and LangGraph agent)
aitriage serve --transport sse --port 9090

# Auto-install in Claude Desktop config
aitriage install-mcp
```

---

## LangGraph Agent (Optional Companion)

An optional Python companion agent runs a stateful, human-in-the-loop remediation workflow on top of the MCP server. It is wired into [`docker-compose.yaml`](docker-compose.yaml) under the `agent` profile and expects an `agent/` build context (provided separately from the core Go binary).

```bash
# Start the MCP server (SSE, port 9090) + LangGraph agent
docker compose --profile agent up -d
```

Behaviour is driven by environment variables:

| Variable | Default | Purpose |
| :--- | :--- | :--- |
| `AITRIAGE_MCP_URL` | `http://aitriage-mcp:9090/sse` | MCP server endpoint the agent connects to |
| `AGENT_MODEL` | `google-genai:gemini-2.5-flash` | LLM used by the agent |
| `AGENT_HITL` | `0` | `1` pauses before applying fixes; `0` is full autopilot |
| `LANGCHAIN_TRACING_V2` | `false` | Enable LangSmith observability and tracing |

---

## Pre-Commit Hooks

AITriage supports the [pre-commit framework](https://pre-commit.com). Add to your `.pre-commit-config.yaml`:

```yaml
repos:
  - repo: https://github.com/cybertortuga/aitriage
    rev: v1.0.0
    hooks:
      - id: aitriage           # Full security scan on changed files
      - id: aitriage-secrets   # Secret detection via Shannon Entropy
```

---

## Enterprise Deployment

AITriage Enterprise provides multi-repo dashboards, role-based access controls (RBAC), JWT authentication, and persistent audit logs via SQLite.

### Environment Configuration

```bash
JWT_SECRET=your-32-character-secret-key-for-api-authentication
GEMINI_API_KEY=your-gemini-key
DB_PATH=/var/lib/aitriage/production.db
```

### Role-Based Access Controls

| Role | Permissions |
| :--- | :--- |
| `superadmin` / `admin` | Full system configuration and team management |
| `security_lead` | Audit policy sign-offs and report reviews |
| `analyst` | Finding triage, validating AI fixes, false-positive marking |
| `developer` | Viewing project findings and applying security fixes |
| `viewer` | Read-only reporting dashboards |

### Startup

```bash
# Web UI + API + SQLite (default)
make up

# Full enterprise stack
make enterprise-up

# Stop
make down
```

---

## Development

### Prerequisites

- Go 1.25.5+
- Node.js 22+ (for web frontend)
- CGO enabled (tree-sitter requires C compilation)

### Build & Test

```bash
make build          # Build the binary to bin/aitriage
make test           # Run the full test suite with verbose output
make lint           # Run golangci-lint
make format         # Format Go + web sources
make build-web      # Build the web frontend
make sync-web       # Build frontend and sync assets into Go binary
make release        # Run GoReleaser snapshot (local test)
make clean          # Remove build artifacts
```

### Project Structure

```
aitriage/
├── cmd/aitriage/          # CLI commands (scan, agent, fix, serve, web, etc.)
├── internal/
│   ├── agent/             # AI agent: graph orchestrator, LLM clients, MCP server
│   │   ├── graph/         # Multi-stage triage pipeline (classify, PoC, report, fixspec)
│   │   ├── llm/           # Provider-agnostic LLM client (Gemini, OpenAI, Anthropic, Ollama, Groq)
│   │   ├── mcp/           # MCP server with 18 tools and 2 resources
│   │   ├── remedy/        # Deterministic autofix engine and spec generation
│   │   └── architect/     # Threat modeling and architecture diagrams
│   ├── engine/            # Core audit engine, baseline, history, orchestrator
│   ├── scanner/           # AST, entropy, external, NFR, deploy, network scanners
│   ├── server/            # Web API server, handlers, repositories, SQLite
│   ├── healthpolicy/      # Security policy profiles and verdict computation
│   ├── report/            # Health check and reporter
│   ├── telemetry/         # Usage tracking and telemetry
│   └── ui/tui/            # Terminal UI dashboard
├── rules/                 # 187 security rules across 11 categories
├── web/                   # React 19 + TypeScript + Vite + TailwindCSS 4
├── testdata/              # Sample vulnerable repos for engine testing
├── examples/              # Example GitHub Actions workflows
├── Dockerfile             # Multi-stage build (web + Go + runtime with scanners)
├── docker-compose.yaml    # Web, MCP, and LangGraph agent services
├── action.yml             # GitHub Docker Action definition
└── .goreleaser.yaml       # Cross-platform release + Homebrew cask config
```

---

## Roadmap

- [x] Concurrent O(N) scanning architecture
- [x] Interactive TUI dashboard
- [x] Model Context Protocol (MCP) server
- [x] Git baseline support (`--baseline`)
- [x] Incremental git-diff audits (`--diff`, `--staged`)
- [x] Information Security Policy Gate & verdict system
- [x] Sentinel watch engine (`aitriage watch`)
- [x] Rule pack package management (`aitriage rules`)
- [x] CycloneDX / SPDX SBOM exports (`aitriage sbom`)
- [x] AI triage & remediation engine (`aitriage agent`, `aitriage fix`)
- [x] Deterministic AI verdict & artifact caching for CI re-runs
- [x] Web dashboard with RBAC and SQLite storage
- [x] LangGraph Python agent with human-in-the-loop
- [ ] Compliance mappings (SOC 2, ISO 27001, OWASP Top 10)

---

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feat/my-feature`)
3. Run tests (`make test`) and linting (`make lint`)
4. Commit with conventional commits (`feat:`, `fix:`, `docs:`, etc.)
5. Open a pull request

For rule contributions, see [rules/README.md](rules/README.md) for the rule schema and guidelines.

---

## License

Distributed under the MIT License. See [LICENSE](LICENSE) for details.

---

<div align="center">
  <sub>Designed and built for high-assurance security triaging. &copy; 2026 cybertortuga.</sub>
</div>
