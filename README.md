<div align="center">
  <img src="web/public/favicon.svg" width="88" height="88" alt="AITriage logo">
  <h1>AITriage</h1>
  <p><strong>Security scanners → SecureCoder AI triage → human decision → verified fixes</strong></p>
  <p>One security workflow for AI coding agents, Web, CI/CD, and the terminal.</p>

  <p>
    <a href="https://github.com/cybertortuga/aitriage/releases"><img src="https://img.shields.io/github/v/release/cybertortuga/aitriage?style=flat-square&color=2563eb" alt="Latest release"></a>
    <a href="https://github.com/cybertortuga/aitriage/actions/workflows/ci.yml"><img src="https://img.shields.io/github/actions/workflow/status/cybertortuga/aitriage/ci.yml?branch=main&style=flat-square&label=tests" alt="Test status"></a>
    <a href="https://github.com/cybertortuga/aitriage/pkgs/container/aitriage"><img src="https://img.shields.io/badge/scanners-Docker-2496ED?style=flat-square&logo=docker&logoColor=white" alt="Docker scanner bundle"></a>
    <a href="LICENSE"><img src="https://img.shields.io/github/license/cybertortuga/aitriage?style=flat-square" alt="MIT license"></a>
  </p>

  <p>
    <a href="#1-install-once"><strong>Install</strong></a> ·
    <a href="#2-ai-ide-codex-or-claude-code"><strong>AI IDE</strong></a> ·
    <a href="#3-web-ui"><strong>Web</strong></a> ·
    <a href="#4-cicd"><strong>CI/CD</strong></a> ·
    <a href="#5-cli"><strong>CLI</strong></a> ·
    <a href="#reports"><strong>Reports</strong></a>
  </p>
</div>

---

AITriage runs deterministic checks plus Semgrep, Trivy, Gitleaks, and Bandit. Its SecureCoder workflow then validates the findings, separates confirmed vulnerabilities from uncertain results and false positives, and prepares remediation instructions. Source code changes require a user decision.

## 1. Install once

AITriage is a system tool, not a dependency of the application being checked. Install it once on the computer; do not clone or copy AITriage into every project.

### Set up through your AI IDE

Open Codex or Claude Code and paste this request:

```text
Install the official released AITriage CLI on this computer and prepare its
complete scanner bundle. Official project: https://github.com/cybertortuga/aitriage

Do not clone AITriage into my current repository and do not change its source.
If `aitriage` is missing, install the latest official release with
`curl -fsSL https://github.com/cybertortuga/aitriage/releases/latest/download/install.sh | sh`.
This installer verifies the release checksum and prepares the complete scanner
bundle. Do not download a binary from any other source.

Then run `aitriage version` and `aitriage setup --status --json`.
If setup returns `action_required`, show me its message, official Docker URL,
and retry command, then stop. Never install Docker from another source.
If setup returns `ok`, run `aitriage setup --status --json` and tell me the
AITriage version, executable path, container image, and status of every bundled
scanner. Do not connect a project or run an audit yet.
```

The result is one host CLI plus one verified Docker image. The source repository is unnecessary for normal use.

### Set up manually

Requirements: macOS or Linux, and a running [Docker Desktop or Docker Engine](https://docs.docker.com/get-started/get-docker/).

```bash
curl -fsSL https://github.com/cybertortuga/aitriage/releases/latest/download/install.sh | sh
aitriage setup --status
```

The installer downloads the matching official GitHub Release, verifies its
SHA-256 checksum, installs the CLI, and runs `aitriage setup --full`. To inspect
the script before running it:

```bash
curl -fsSLO https://github.com/cybertortuga/aitriage/releases/latest/download/install.sh
less install.sh
sh install.sh
```

Go 1.25.12 or newer remains a developer fallback:

```bash
go install github.com/cybertortuga/aitriage/cmd/aitriage@latest
aitriage setup --full
```

If Docker is missing or stopped, setup exits without scanning and prints one official installation link plus the same command to retry. It does not silently install Docker or individual scanners.

### What is installed

| Part | Where | Purpose |
| :--- | :--- | :--- |
| `aitriage` CLI | system executable path | setup, project connectors, CLI, Web, MCP launcher |
| scanner image | local Docker image cache | AITriage, Semgrep, Trivy, Gitleaks, Bandit |
| scanner cache | user cache directory | reusable scanner databases; never stored in source |
| reports | `<project>/aitriage-reports/` | results and local run state for that project |

`aitriage setup --full` is safe to repeat after an upgrade. Remove only the downloaded AITriage image with `aitriage setup --remove-runtime`.

## Choose how to use it

| Interface | Use it when | AI access |
| :--- | :--- | :--- |
| [AI IDE](#2-ai-ide-codex-or-claude-code) | Codex or Claude Code should audit and help fix the open project | existing Codex or Claude subscription |
| [Web](#3-web-ui) | a person wants a local browser dashboard | provider API key for AI triage |
| [CI/CD](#4-cicd) | pushes and pull requests need an automatic gate | provider key in CI secrets |
| [CLI](#5-cli) | terminal or scripts | none for raw scan; provider key for full AI triage |

## 2. AI IDE: Codex or Claude Code

This mode uses the model included in the current Codex or Claude subscription. It does not need a separate LLM API key.

### Connect the open project through your AI IDE

First finish the [one-time installation](#1-install-once). Then open the **root of the project to audit** and paste:

```text
Connect AITriage to the repository currently open in this AI IDE.

Run `aitriage setup --status --json` first. Stop if the complete scanner bundle
is not ready. Detect this client and run exactly one command from the open
repository root:
- Codex: `aitriage install-codex .`
- Claude Code: `aitriage install-claude-code .`

Do not clone AITriage here. Preserve existing source, instructions, MCP servers,
and `.gitignore` entries. Do not audit or edit source during setup. Verify the
project-local MCP configuration, list the files changed, and tell me to open a
new task or session so the client loads the server.
```

Open a new task/session in the same project. From then on, ordinary requests are enough:

```text
Проверь этот проект через AITriage. Пока ничего не исправляй.
```

AITriage—not the AI IDE—supplies the SecureCoder prompts, controls the workflow, and creates the final artifacts. The agent returns each model answer to AITriage until triage completes.

When you decide to fix findings, say which ones or explicitly approve all confirmed findings:

```text
Исправь подтверждённые AITriage уязвимости. Не трогай сомнительные находки и
false positives. Запусти тесты и проверь исправления через AITriage.
```

### Connect manually

Run one command from the project root:

```bash
aitriage install-codex .
# or
aitriage install-claude-code .
```

Both connectors use the verified container runtime by default, preserve unrelated client settings, add `/aitriage-reports/` to `.gitignore`, and confine tools to the opened root and its subdirectories. Open a new task/session after connecting.

Update by running the same command again. Remove only the AITriage integration with:

```bash
aitriage install-codex . --uninstall
aitriage install-claude-code . --uninstall
```

## 3. Web UI

Web provides a local dashboard for findings, evidence, reports, and run history. It uses the same verified scanner image.

### Start through your AI IDE

```text
Start AITriage Web for the repository currently open in this AI IDE.
Run `aitriage setup --status --json` and stop if the scanner bundle is not ready.
Then run `aitriage web --project . --port 8080`, wait until it responds, and
give me the local URL and the command to stop it. Do not expose it to the
network, write credentials to files, or change source code.
```

### Start manually

```bash
aitriage web --project . --port 8080
```

Open [http://localhost:8080](http://localhost:8080). Stop with `Ctrl-C`.

Web AI triage cannot use a Codex or Claude subscription. Set a supported provider key in the process environment when AI features are needed, for example:

```bash
export GEMINI_API_KEY="..."
aitriage web --project . --port 8080
```

The current Web server has no enforced login. It binds through the managed container to localhost; keep it local and do not expose it directly to the Internet.

## 4. CI/CD

CI uses the same AITriage pipeline and SecureCoder prompts, but its model is authenticated by a secret in the CI platform. A developer's local installation and subscription are not used.

### Configure through your AI IDE

```text
Add AITriage CI/CD to this repository using an official example from
https://github.com/cybertortuga/aitriage/tree/main/examples/github-actions.
Preserve existing workflows. Never put a provider key in YAML or source code;
reference a GitHub Secret. Show me the workflow diff and required secret before
committing or pushing. Do not invent an organization-specific reusable workflow
URL: ask me if the approved URL is not already documented in this repository.
```

### Configure manually

Choose the closest template in [`examples/github-actions/`](examples/github-actions/):

- [`aitriage-security.yml`](examples/github-actions/aitriage-security.yml) — standard security run;
- [`aitriage-pr-gate.yml`](examples/github-actions/aitriage-pr-gate.yml) — pull-request gate;
- [`aitriage-ai-advisor.yml`](examples/github-actions/aitriage-ai-advisor.yml) — AI advisory output;
- [`aitriage-manual-html-report.yml`](examples/github-actions/aitriage-manual-html-report.yml) — manually triggered report.

Copy the selected file to `.github/workflows/`, add the referenced provider key in GitHub Secrets, and run `workflow_dispatch` once. Verify the uploaded reports, SARIF, and post-AI gate before making the check required.

For organizations, the application workflow may call one centrally managed reusable workflow. Use the exact approved repository and secret names; AITriage documentation intentionally does not guess them.

## 5. CLI

### Raw deterministic scan

```bash
aitriage scan .
```

This fast pre-scan uses built-in deterministic checks only. It requires neither Docker nor an AI key, but it is **not** a completed SecureCoder verdict.

Useful variants:

```bash
aitriage scan ./service
aitriage scan . --staged
aitriage scan . --diff origin/main
aitriage scan . --format sarif -o aitriage.sarif
```

### Full AI audit

Provide a supported provider key and run:

```bash
export GEMINI_API_KEY="..."
aitriage agent . --no-chat
```

`agent` uses the full container scanner bundle by default and runs the same SecureCoder pipeline used by CI. Provider variables include `GEMINI_API_KEY`, `ANTHROPIC_API_KEY`, `OPENAI_API_KEY`, and `GROQ_API_KEY`.

To use a Codex or Claude subscription instead of an API key, use the [AI IDE integration](#2-ai-ide-codex-or-claude-code).

## Reports

All project-owned output lives under one ignored directory:

```text
aitriage-reports/
├── run-.../
│   ├── summary.md            # start here
│   ├── report.md             # full SecureCoder report
│   ├── fixspec.md            # remediation instructions
│   ├── triage-findings.json  # all AI verdicts
│   ├── aitriage.sarif        # standard scanner output
│   ├── scan.json             # deterministic input to triage
│   ├── manifest.json         # state, scanner coverage, integrity metadata
│   └── audit.log             # run events
├── history/                  # deterministic scan history
└── web/                      # local Web state
```

Connectors add `/aitriage-reports/` to `.gitignore`. AITriage excludes this directory from every scanner and AI context so its own output cannot trigger another audit loop.

## Safety contract

- A full run fails closed if Docker or any required scanner cannot run.
- Every report records whether each scanner completed or was deterministically not applicable.
- The source tree is mounted read-only inside the scanner container; only `aitriage-reports/` is writable there.
- MCP tools are confined to the opened repository root and real subdirectories; traversal and symlink escapes are rejected.
- AITriage supplies the canonical SecureCoder prompts. The host agent must not replace them with an ad-hoc review.
- No source change is authorized by a scan. The user chooses whether and what to fix.

## Troubleshooting

| Problem | Action |
| :--- | :--- |
| Docker is absent or stopped | Follow the official URL printed by `aitriage setup --full`, start Docker, then repeat the same command |
| Runtime is incomplete after an upgrade | Run `aitriage setup --repair`, then `aitriage setup --status` |
| Codex or Claude cannot see AITriage | Re-run the matching connector from the repository root and open a new task/session |
| Claude shows MCP approval pending | Open Claude Code in the project and approve the local `aitriage` server |
| A real subdirectory is rejected | Reconnect from the intended repository root; paths outside it and symlink escapes are deliberately blocked |
| The agent ran only `aitriage scan` | Ask it to use the AITriage MCP workflow; raw scan is not AI triage |
| Web AI is unavailable | Export a supported provider API key before starting Web |
| Full audit failed | Keep the error visible and repair the runtime; do not present a reduced scan as the final verdict |

## Project documentation

- [Configuration example](.aitriage.yaml.example)
- [Security rules](rules/README.md)
- [GitHub Actions examples](examples/github-actions/)
- [Contributing](CONTRIBUTING.md)
- [Security policy](SECURITY.md)

## License

MIT. See [LICENSE](LICENSE).
