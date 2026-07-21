<div align="center">
  <img src="web/public/favicon.svg" width="88" height="88" alt="AITriage logo">
  <h1>AITriage</h1>
  <p><strong>Security scan → SecureCoder AI triage → human approval → verified fixes</strong></p>
  <p>Use it from an AI coding agent, a browser, CI/CD, or a terminal.</p>

  <p>
    <a href="https://github.com/cybertortuga/aitriage/releases"><img src="https://img.shields.io/github/v/release/cybertortuga/aitriage?style=flat-square&color=2563eb" alt="Latest release"></a>
    <a href="https://github.com/cybertortuga/aitriage/actions"><img src="https://img.shields.io/github/actions/workflow/status/cybertortuga/aitriage/ci.yml?branch=main&style=flat-square&label=build" alt="Build status"></a>
    <a href="LICENSE"><img src="https://img.shields.io/github/license/cybertortuga/aitriage?style=flat-square" alt="MIT license"></a>
    <a href="https://github.com/cybertortuga/aitriage/pkgs/container/aitriage"><img src="https://img.shields.io/badge/GHCR-container-2496ED?style=flat-square&logo=docker&logoColor=white" alt="GHCR container"></a>
  </p>

  <p>
    <a href="#full-installation"><strong>Installation</strong></a>&nbsp;&nbsp;&nbsp;
    <a href="#ai-ide"><strong>AI IDE</strong></a>&nbsp;&nbsp;&nbsp;
    <a href="#web-ui"><strong>Web UI</strong></a>&nbsp;&nbsp;&nbsp;
    <a href="#cicd"><strong>CI/CD</strong></a>&nbsp;&nbsp;&nbsp;
    <a href="#cli"><strong>CLI</strong></a>
  </p>
</div>

---

AITriage scans source code, runs every finding through the bundled SecureCoder prompts, separates confirmed vulnerabilities from false positives, and prepares fix instructions. It never treats a failed security check as permission to change source code.

## Full installation

AITriage is installed **once on the computer**. After that, the same installation powers AI IDE, Web UI, and CLI usage. It is not copied into every project and must never be cloned inside the project being checked.

```text
GitHub/
├── aitriage/          # optional: source checkout for AITriage development
└── my-project/        # the project being checked
```

Most users do not need the `aitriage/` source directory at all.

### Install through your AI IDE

Paste this into Codex or Claude Code from any working directory:

```text
Install the official AITriage release once on this computer.
Official repository: https://github.com/cybertortuga/aitriage

AITriage must be installed as a system command. Never clone or copy its source
repository inside my current project.

1. Check whether `aitriage` is already available. If it is, print
   `aitriage version` and do not reinstall it.
2. If it is missing and Homebrew is available, run
   `brew install cybertortuga/aitriage/aitriage`.
3. If Homebrew is unavailable but Go 1.25.5+ is installed, run
   `go install github.com/cybertortuga/aitriage/cmd/aitriage@latest`.
4. If neither method is available, stop and explain what must be installed.
5. Verify `aitriage version` and `aitriage --help`.
6. Report the installed version and executable path. Do not connect a project,
   start Web, run a scan, or modify source code during this installation step.
```

### Install manually

With Homebrew:

```bash
brew install cybertortuga/aitriage/aitriage
```

Or with Go 1.25.5+:

```bash
go install github.com/cybertortuga/aitriage/cmd/aitriage@latest
```

Verify it once:

```bash
aitriage version
```

### What the installation contains

The released executable is self-contained:

- AITriage CLI commands;
- built-in security rules and deterministic scanners;
- the MCP server used by Codex and Claude Code;
- the local Web UI and API server;
- report generation and policy checks.

External tools such as Semgrep, Trivy, Gitleaks, and Bandit are not embedded in the host executable. Use the published Docker image when you want those tools preinstalled. Provider credentials are never bundled.

## Use AITriage

Choose one interface:

| Interface | Best for | AI credentials |
| :--- | :--- | :--- |
| [AI IDE](#ai-ide) | Checking and fixing code from Codex or Claude Code | Existing Codex or Claude subscription |
| [Web UI](#web-ui) | Visual scanning and report review in a browser | Provider API key for AI features |
| [CI/CD](#cicd) | Automatic checks for pushes and pull requests | Provider key in GitHub Secrets |
| [CLI](#cli) | Scripts, terminal use, and local automation | None for `scan`; provider key for `agent` |

> [!TIP]
> Already using Codex or Claude Code? Start with [AI IDE](#ai-ide). It uses your existing subscription and does not require another LLM API key.

---

## AI IDE

Use this mode when you want Codex or Claude Code to run the complete AITriage pipeline with the model included in your existing subscription. No separate LLM API key is required.

Connecting a project is not another installation. It only writes the small project-local MCP configuration needed to confine AITriage to that repository.

> [!IMPORTANT]
> An audit stops before source changes. The agent may fix only the confirmed finding IDs that you explicitly approve.

### Connect through your AI IDE

Open your project in Codex or Claude Code and paste:

```text
Configure AITriage MCP for the repository currently open in this AI IDE. Work
from the repository root.

Rules:
- AITriage is a system CLI, not a dependency of this repository. Never clone or
  copy https://github.com/cybertortuga/aitriage inside the current repository.
- Preserve all source files, existing agent instructions, MCP servers, and
  `.gitignore` entries.
- Do not scan the repository and do not modify source code during setup.

Do the following:
1. Check whether `aitriage` is available and print `aitriage version`.
2. If it is missing, stop and tell me to complete the one-time installation at
   https://github.com/cybertortuga/aitriage#full-installation.
3. Detect the current AI IDE and run exactly one project connector:
   - Codex: `aitriage install-codex .`
   - Claude Code: `aitriage install-claude-code .`
4. Verify the generated project-local MCP configuration and managed instruction
   block. Ensure `/aitriage-reports/` appears once in `.gitignore`.
5. Report the installed AITriage version, command executed, files changed, and
   verification result. Do not claim the MCP server is connected if only its
   configuration was written.
6. Stop. Tell me to open a new task/session so the AI IDE loads the MCP server.
```

After setup, open a new task/session in the same project and paste:

```text
Check this project with AITriage through its MCP workflow. Do not use a raw
`aitriage scan` as the final result. Do not fix anything yet. Complete the AI
triage, explain the verdict in plain language, and show me where the reports
were saved.
```

To approve selected fixes later:

```text
Fix confirmed AITriage findings CS-AUTH-001 and CS-AUTHZ-001. Do not change
false positives or uncertain findings. Run the required tests and verify the
fixes with AITriage.
```

### Connect manually

Open a terminal in the root of the project you want to check.

For Codex:

```bash
aitriage install-codex .
```

For Claude Code:

```bash
aitriage install-claude-code .
```

The command name contains `install`, but it does not reinstall AITriage. It adds only the project-local MCP configuration and managed agent instructions. It preserves unrelated MCP servers and existing instruction text. Open a **new** Codex task or Claude Code session after connection, then use the audit request above.

### AI IDE files

Depending on the client, the project connector may update:

- `.codex/config.toml` and `AGENTS.md` for Codex;
- the project-local Claude MCP configuration and `CLAUDE.md` for Claude Code;
- `.gitignore` with `/aitriage-reports/`.

Update the integration by running the install command again. Remove only AITriage-managed configuration with:

```bash
aitriage install-codex . --uninstall
aitriage install-claude-code . --uninstall
```

---

## Web UI

Use this mode when you want to select projects, start scans, and read reports in a browser. Web AI features use a provider API key; the Web UI cannot use a Codex or Claude subscription.

Web uses the same installed `aitriage` executable. Starting the server does not install another copy.

### Start through your AI IDE

Paste this into Codex or Claude Code:

```text
Start the AITriage Web UI for the repository currently open in this AI IDE.

Do not clone AITriage inside this repository. Check whether the system command
`aitriage` is installed. If it is missing, stop and tell me to complete the
one-time installation documented at
https://github.com/cybertortuga/aitriage#full-installation.

Run `aitriage web --port 8080` locally. Do not expose the server to the network
and do not write API keys to files. If no supported provider key is already in
the environment, start Web without AI features and tell me which environment
variable enables them. Confirm that http://localhost:8080 responds, then show
me the URL and the exact command needed to stop the server.
```

### Start manually

No source checkout is required:

```bash
aitriage web --port 8080
```

Open `http://localhost:8080`.

To enable AI features, provide a supported key before startup:

```bash
export GEMINI_API_KEY="your-key"
aitriage web --port 8080
```

Supported provider variables include `GEMINI_API_KEY`, `ANTHROPIC_API_KEY`, `OPENAI_API_KEY`, and `GROQ_API_KEY`.

<details>
<summary><strong>Docker: run with Semgrep, Trivy, Gitleaks, and Bandit included</strong></summary>

<br>

The published container includes Semgrep, Trivy, Gitleaks, and Bandit. Run this command from the project you want to inspect:

```bash
docker run --rm -p 8080:8080 \
  -v "$PWD:/host:ro" \
  -v aitriage-data:/app/data \
  -e DB_PATH=/app/data/aitriage.db \
  ghcr.io/cybertortuga/aitriage:latest \
  web --port 8080 --host-prefix /host
```

Open `http://localhost:8080` and select the mounted project at `/host`. Pass a provider key with an additional `-e GEMINI_API_KEY` or equivalent only when AI features are needed.

</details>

The current Web UI has no enforced login. Keep it on a trusted local machine or isolated network; do not expose it directly to the Internet.

---

## CI/CD

Use this mode to run AITriage automatically on pushes, pull requests, and manual GitHub Actions runs. The canonical organization setup keeps scan logic and SecureCoder prompts in one centrally maintained reusable workflow.

CI/CD runs its own copy inside a GitHub Actions runner. The developer's local AITriage installation is not used by CI.

### Configure through your AI IDE

Paste this into Codex or Claude Code:

```text
Configure AITriage CI/CD for this repository using the canonical reusable
workflow documented at https://github.com/cybertortuga/aitriage#cicd.

Preserve every existing workflow. Create or update only
`.github/workflows/aitriage.yml`. First identify the approved reusable workflow
repository and the allowed GitHub Secret name from existing repository or
organization configuration. If either value cannot be confirmed, stop and ask
me instead of guessing.

Never put an API key in YAML or source code. Validate the completed workflow,
show me the exact diff, and explain which GitHub Secret must exist. Do not
commit, push, or run the workflow until I explicitly approve those actions.
```

### Configure manually

Create `.github/workflows/aitriage.yml` in the application repository, then replace the workflow repository and secret names with values approved by your organization.

<details>
<summary><strong>Canonical reusable workflow</strong></summary>

<br>

```yaml
name: AITriage Security

on:
  pull_request:
    branches: [main]
  push:
    branches: [main]
  workflow_dispatch:
    inputs:
      llm-provider:
        description: "LLM provider: gemini, openai"
        required: false
        default: gemini
        type: choice
        options:
          - gemini
          - openai
      llm-model:
        description: "Model override (leave empty for provider default)"
        required: false
        default: ""
        type: string
      llm-base-url:
        description: "OpenAI-compatible base URL"
        required: false
        default: ""
        type: string
      llm-secret:
        description: "Which organization secret to use"
        required: false
        default: GEMINI_API_KEY
        type: choice
        options:
          - GEMINI_API_KEY
          - GLM_CI_KEY
          - XIAOMI_API_KEY

permissions:
  contents: read
  security-events: write

jobs:
  aitriage:
    uses: your-org/security-workflows/.github/workflows/aitriage.yml@main
    with:
      llm-provider: ${{ inputs.llm-provider || 'gemini' }}
      llm-model: ${{ inputs.llm-model || '' }}
      llm-base-url: ${{ inputs.llm-base-url || '' }}
    secrets:
      llm_api_key: ${{ inputs.llm-secret == 'GLM_CI_KEY' && secrets.GLM_CI_KEY || inputs.llm-secret == 'XIAOMI_API_KEY' && secrets.XIAOMI_API_KEY || secrets.GEMINI_API_KEY }}
```

</details>

Then:

1. replace `your-org/security-workflows` with the approved reusable-workflow repository;
2. add the selected provider key to organization or repository GitHub Secrets;
3. allow the application repository to call the reusable workflow;
4. run `workflow_dispatch` once and verify the report artifacts and SARIF upload;
5. make the post-AI AITriage result a required branch-protection check.

Standalone workflow examples are available in [`examples/github-actions/`](examples/github-actions/).

The AI IDE may need you to provide the organization’s reusable-workflow repository and allowed secret name. Those values must not be guessed.

---

## CLI

Use this mode for direct terminal commands, scripts, and local automation.

### Run through your AI IDE

For a raw scan without AI triage:

```text
Run `aitriage scan .` in this repository. Do not modify source code. Explain
the findings and clearly label them as raw, deterministic results that have not
been verified by AI triage.
```

For full CLI triage with a provider key already available in the environment:

```text
Run the full AITriage CLI agent for this repository. Use only provider
credentials already present in the environment. Never ask me to paste a secret
into chat and never write credentials to files.

Save `triage-findings.json`, `summary.md`, `report.md`, and `fixspec.md` under
`aitriage-reports/`. Do not modify source code. Explain the final AI-triaged
gate and link every generated artifact. If no supported provider key is
available, stop and tell me the required environment variable.
```

### Run manually without AI

```bash
aitriage scan .
```

No API key is required. This is a deterministic pre-scan, not a final AI-triaged verdict.

Useful variants:

```bash
aitriage scan ./service
aitriage scan . --staged
aitriage scan . --diff origin/main
aitriage scan . --format sarif -o aitriage.sarif
```

### Run manually with AI

Set a provider key and run the full SecureCoder pipeline:

```bash
export GEMINI_API_KEY="your-key"
aitriage agent .
```

<details>
<summary><strong>Non-interactive run with explicit artifacts</strong></summary>

<br>

```bash
mkdir -p aitriage-reports
aitriage agent . --no-chat \
  --triage-out aitriage-reports/triage-findings.json \
  --summary-out aitriage-reports/summary.md \
  --report-out aitriage-reports/report.md \
  --fixspec-out aitriage-reports/fixspec.md \
  --fail-on critical
```

</details>

If you want to use the Codex or Claude subscription instead of a provider API key, use the [AI IDE](#ai-ide) integration, not `aitriage agent`.

---

## Reports

AI IDE runs store all state and artifacts under the repository root:

```text
aitriage-reports/
└── run-.../
    ├── summary.md           # short result; start here
    ├── report.md            # full security report
    ├── fixspec.md           # proposed remediation instructions
    ├── triage-findings.json # machine-readable AI verdicts
    ├── aitriage.sarif       # standard security-tool output
    ├── scan.json            # deterministic scan input
    ├── manifest.json        # run state and integrity metadata
    └── audit.log            # run events
```

Add `/aitriage-reports/` to `.gitignore`. AITriage excludes this directory from later scans and model context so generated reports cannot create an analysis loop.

## Safety

- AITriage uses its bundled SecureCoder prompts; the host AI must not invent a replacement triage flow.
- AI IDE tools are confined to the configured repository root and its real subdirectories.
- Path traversal and symlink escapes outside the project are rejected.
- Source changes require an explicit user request.
- False positives and uncertain findings are not fixed automatically.
- A raw `aitriage scan` result is never presented as completed AI triage.

## Troubleshooting

| Problem | What to do |
| :--- | :--- |
| Codex or Claude cannot see AITriage | Run the matching project connector from the repository root, then open a new task/session |
| Claude shows MCP approval as pending | Open the project in Claude Code and approve the `aitriage` MCP server |
| A subfolder is rejected | Confirm it is a real folder inside the repository root used during connection |
| The agent runs only `aitriage scan` | Ask it to use the AITriage MCP workflow; raw scan is not full triage |
| Web AI features are unavailable | Set a supported provider API key before starting `aitriage web` |
| Full triage fails | Keep the failure visible; do not substitute raw scan output as the verdict |

## Project documentation

- [Configuration example](.aitriage.yaml.example)
- [Security rules](rules/README.md)
- [GitHub Actions examples](examples/github-actions/)
- [Contributing](CONTRIBUTING.md)
- [Security policy](SECURITY.md)

## License

MIT. See [LICENSE](LICENSE).
