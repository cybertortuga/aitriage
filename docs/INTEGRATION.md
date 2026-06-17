# AITriage Workflow Integration Guide 🛡️

Integrate AITriage into your continuous integration pipelines to automatically audit code, catch security regressions, and prevent insecure configurations from reaching production.

---

## 🐙 GitHub Actions

AITriage is published as a pre-built Docker action. Running the pre-compiled image skips build stages and finishes in seconds.

### Two-Layer Pipeline Workflow Example

Create `.github/workflows/aitriage.yml`:

```yaml
name: AITriage Security Scan
on: [push, pull_request]

permissions:
  contents: read
  security-events: write   # Required to upload SARIF report to the Security tab
  pull-requests: write     # Required for AI Advisor pull request comments

jobs:
  # ── Layer 1: Deterministic Gate (Blocks PR merges) ──
  gate:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout Repository
        uses: actions/checkout@v4
        with:
          fetch-depth: 0   # Fetch entire history so diff/baseline checks work

      - name: Run AITriage Security Scan
        uses: cybertortuga/aitriage@v1
        with:
          health-profile: standard   # baseline | standard | strict
          fail-on: critical          # critical | any | never
          fail-score: 70             # Exit non-zero if score is below 70 (0 = disabled)
          baseline: 'true'           # Suppress legacy baseline findings
          format: sarif
          output-file: aitriage-results.sarif

      - name: Upload Report to GitHub Security Tab
        uses: github/codeql-action/upload-sarif@v3
        if: always()
        with:
          sarif_file: aitriage-results.sarif
          category: aitriage

  # ── Layer 2: AI Advisor / PR Agent (Non-blocking triage & fixes) ──
  ai-advisor:
    if: github.event_name == 'pull_request'
    needs: gate
    continue-on-error: true
    runs-on: ubuntu-latest
    steps:
      - name: Checkout Repository
        uses: actions/checkout@v4

      - name: Run AITriage AI Triage Agent
        uses: cybertortuga/aitriage@v1
        with:
          command: agent
          args: '--no-chat --report-out report.md --fixspec-out FIXSPEC.md'
        env:
          GEMINI_API_KEY: ${{ secrets.GEMINI_API_KEY }}
```

> [!IMPORTANT]
> **AI Keys & Provider Auto-Detection**:
> - **LLM Key Storage**: Never hardcode API keys. Store them securely in your repository secrets (e.g. `secrets.GEMINI_API_KEY`, `secrets.OPENAI_API_KEY`, or `secrets.ANTHROPIC_API_KEY`) and map them under the `env:` block of your action step.
> - **Provider Auto-Detection**: The AITriage Agent automatically detects the LLM provider based on which API key environment variable is set (`GEMINI_API_KEY` for Google Gemini, `OPENAI_API_KEY` for OpenAI, `ANTHROPIC_API_KEY` for Anthropic).

---

## 🦊 GitLab CI/CD

For GitLab CI/CD, use the pre-built Docker image hosted on the GitHub Container Registry. The runner maps the build path and runs the scanner based on the exit verdict.

Create `.gitlab-ci.yml`:

```yaml
stages:
  - security

aitriage_security_scan:
  stage: security
  image:
    name: ghcr.io/cybertortuga/aitriage:latest
    entrypoint: [""]
  variables:
    # Set this in your GitLab CI/CD variables dashboard
    GEMINI_API_KEY: $GEMINI_API_KEY
  script:
    # Run deterministic scanner and evaluate policy
    - aitriage scan . --format json --out results.json --health-profile standard --fail-on critical
  artifacts:
    name: "aitriage-report"
    when: always
    paths:
      - results.json
```

---

## 🔐 Integration Best Practices

### 1. Leverage the Baseline Feature
Avoid blockages from pre-existing technical debt. Run AITriage locally and create a baseline file:
```bash
aitriage baseline create .
```
Commit `.aitriage/baseline.json`. In your CI workflows, specify `--baseline` or `baseline: 'true'`. The policy verdict will ignore old findings and only fail if a new regression is introduced.

### 2. Choose the Right Policy Profile
- Use **`baseline`** for legacy codebases where you only want to stop new critical/high issues.
- Use **`standard`** for active application pipelines. This enforces a minimum score of `70` and blocks critical findings.
- Use **`strict`** for critical infrastructure, APIs, and high-assurance domains, blocking any active vulnerability and requiring a minimum score of `90`.

### 3. Add Pre-commit Hooks
Catch security mistakes before they are pushed. Initialize git pre-commit hooks locally:
```bash
aitriage init --pre-commit
```
This runs a fast incremental scan (`--staged`) on staged files during git commit.
