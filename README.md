<div align="center">
  <img src="web/public/favicon.svg" width="88" height="88" alt="AITriage logo">
  <h1>AITriage</h1>
  <p><strong>Security scan → SecureCoder AI triage → human approval → verified fixes</strong></p>
  <p>Use it from an AI coding agent, a browser, CI/CD, or a terminal.</p>

  <p>
    <a href="https://github.com/cybertortuga/aitriage/releases"><img src="https://img.shields.io/github/v/release/cybertortuga/aitriage?style=for-the-badge&color=2563eb" alt="Latest release"></a>
    <a href="https://github.com/cybertortuga/aitriage/actions"><img src="https://img.shields.io/github/actions/workflow/status/cybertortuga/aitriage/ci.yml?branch=main&style=for-the-badge&label=build" alt="Build status"></a>
    <a href="LICENSE"><img src="https://img.shields.io/github/license/cybertortuga/aitriage?style=for-the-badge" alt="MIT license"></a>
    <a href="https://github.com/cybertortuga/aitriage/pkgs/container/aitriage"><img src="https://img.shields.io/badge/GHCR-container-2496ED?style=for-the-badge&logo=docker&logoColor=white" alt="GHCR container"></a>
  </p>

  <p>
    <a href="#ai-ide"><strong>AI IDE</strong></a> ·
    <a href="#web-ui"><strong>Web UI</strong></a> ·
    <a href="#cicd"><strong>CI/CD</strong></a> ·
    <a href="#cli"><strong>CLI</strong></a>
  </p>
</div>

---

AITriage scans source code, runs every finding through the bundled SecureCoder prompts, separates confirmed vulnerabilities from false positives, and prepares fix instructions. It never treats a failed security check as permission to change source code.

## Before you start

AITriage is installed **once on the computer**. It is not copied into every project and must never be cloned inside the project being checked.

```text
GitHub/
├── aitriage/          # optional: source checkout for AITriage development
└── my-project/        # the project being checked
```

Most users do not need the `aitriage/` source directory at all. Install the released CLI:

```bash
brew install cybertortuga/aitriage/aitriage
```

Or, with Go 1.25.5+:

```bash
go install github.com/cybertortuga/aitriage/cmd/aitriage@latest
```

Verify it once:

```bash
aitriage version
```

Choose one interface:

| Interface | Best for | AI credentials |
| :--- | :--- | :--- |
| [AI IDE](#ai-ide) | Checking and fixing code from Codex or Claude Code | Existing Codex or Claude subscription |
| [Web UI](#web-ui) | Visual scanning and report review in a browser | Provider API key for AI features |
| [CI/CD](#cicd) | Automatic checks for pushes and pull requests | Provider key in GitHub Secrets |
| [CLI](#cli) | Scripts, terminal use, and local automation | None for `scan`; provider key for `agent` |

## AI IDE

Use this mode when you want Codex or Claude Code to run the complete AITriage pipeline with the model included in your existing subscription. No separate LLM API key is required.

### Set up manually

Open a terminal in the root of the project you want to check.

For Codex:

```bash
aitriage install-codex .
```

For Claude Code:

```bash
aitriage install-claude-code .
```

The installer adds only the project-local MCP configuration and managed agent instructions. It preserves unrelated MCP servers and existing instruction text. Open a **new** Codex task or Claude Code session after installation.

Then ask:

```text
Check this project with AITriage. Do not fix anything yet. Explain the result
in plain language and show me where the report was saved.
```

To approve selected fixes later:

```text
Fix confirmed AITriage findings CS-AUTH-001 and CS-AUTHZ-001. Do not change
false positives or uncertain findings. Run the required tests and verify the
fixes with AITriage.
```

### Set up through your AI IDE

Open your project in Codex or Claude Code and paste:

```text
Set up AITriage for the project currently open in this AI IDE using the official
instructions: https://github.com/cybertortuga/aitriage

AITriage is a system tool, not a project dependency. Never clone or copy its
repository inside my project. Install the released CLI separately if needed,
then run the installer for this AI IDE against the current project. Preserve
all existing settings and source files. Do not scan or modify source code yet.
Verify the project-local MCP connection and tell me when to open a new session.
```

In the new session, use the normal audit request shown above. The managed agent instructions contain the internal MCP tool sequence; users do not need to call those tools manually.

### AI IDE files

Depending on the client, the installer may update:

- `.codex/config.toml` and `AGENTS.md` for Codex;
- the project-local Claude MCP configuration and `CLAUDE.md` for Claude Code;
- `.gitignore` with `/aitriage-reports/`.

Update the integration by running the install command again. Remove only AITriage-managed configuration with:

```bash
aitriage install-codex . --uninstall
aitriage install-claude-code . --uninstall
```

## Web UI

Use this mode when you want to select projects, start scans, and read reports in a browser. Web AI features use a provider API key; the Web UI cannot use a Codex or Claude subscription.

### Start manually with the installed CLI

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

### Start manually with Docker

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

### Start through your AI IDE

Paste this into Codex or Claude Code:

```text
Start the AITriage Web UI for the project currently open in this AI IDE.
Official instructions: https://github.com/cybertortuga/aitriage#web-ui

Do not clone AITriage inside my project. Use the installed CLI; install the
released CLI first if it is missing. Start the Web UI locally on port 8080,
do not expose it to the network, and do not store API keys in project files.
Tell me the local URL and how to stop the server.
```

The current Web UI has no enforced login. Keep it on a trusted local machine or isolated network; do not expose it directly to the Internet.

## CI/CD

Use this mode to run AITriage automatically on pushes, pull requests, and manual GitHub Actions runs. The canonical organization setup keeps scan logic and SecureCoder prompts in one centrally maintained reusable workflow.

### Configure manually

Create `.github/workflows/aitriage.yml` in the application repository:

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

Then:

1. replace `your-org/security-workflows` with the approved reusable-workflow repository;
2. add the selected provider key to organization or repository GitHub Secrets;
3. allow the application repository to call the reusable workflow;
4. run `workflow_dispatch` once and verify the report artifacts and SARIF upload;
5. make the post-AI AITriage result a required branch-protection check.

Standalone workflow examples are available in [`examples/github-actions/`](examples/github-actions/).

### Configure through your AI IDE

Paste this into Codex or Claude Code:

```text
Configure AITriage CI/CD for this repository using the canonical reusable
workflow documented at https://github.com/cybertortuga/aitriage#cicd

Preserve existing workflows. Create or update only
`.github/workflows/aitriage.yml`. Use our approved reusable security-workflow
repository and existing GitHub Secret names; never put an API key in YAML or
source code. Validate the workflow syntax and show me the exact diff. Do not
commit, push, or run the workflow until I approve the changes.
```

The AI IDE may need you to provide the organization’s reusable-workflow repository and allowed secret name. Those values must not be guessed.

## CLI

Use this mode for direct terminal commands, scripts, and local automation.

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

For non-interactive automation with explicit artifacts:

```bash
mkdir -p aitriage-reports
aitriage agent . --no-chat \
  --triage-out aitriage-reports/triage-findings.json \
  --summary-out aitriage-reports/summary.md \
  --report-out aitriage-reports/report.md \
  --fixspec-out aitriage-reports/fixspec.md \
  --fail-on critical
```

### Run through your AI IDE

For a raw scan without connecting MCP:

```text
Run `aitriage scan .` in this project. Do not change source code. Explain the
raw findings and clearly state that they have not been AI-triaged.
```

For full CLI triage with an existing provider key:

```text
Run the full AITriage CLI agent for this project using the provider credentials
already present in my environment. Do not ask me to paste a secret into chat
and do not write secrets to project files. Save the canonical outputs under
`aitriage-reports/`, make no source changes, and explain the final gate.
If no supported provider key is available, stop and tell me which environment
variable is required.
```

If you want to use the Codex or Claude subscription instead of a provider API key, use the [AI IDE](#ai-ide) integration, not `aitriage agent`.

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
| Codex or Claude cannot see AITriage | Run the matching installer from the project root, then open a new task/session |
| Claude shows MCP approval as pending | Open the project in Claude Code and approve the `aitriage` MCP server |
| A subfolder is rejected | Confirm it is a real folder inside the root used during installation |
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
