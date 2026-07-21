<div align="center">
  <img src="web/public/favicon.svg" width="88" height="88" alt="AITriage logo">
  <h1>AITriage</h1>
  <p><strong>Security scan → SecureCoder AI triage → human approval → verified fixes</strong></p>
  <p>Use Codex or Claude Code subscriptions locally. Run the same analysis pipeline in CI/CD.</p>

  <p>
    <a href="https://github.com/cybertortuga/aitriage/releases"><img src="https://img.shields.io/github/v/release/cybertortuga/aitriage?style=for-the-badge&color=2563eb" alt="Latest release"></a>
    <a href="https://github.com/cybertortuga/aitriage/actions"><img src="https://img.shields.io/github/actions/workflow/status/cybertortuga/aitriage/ci.yml?branch=main&style=for-the-badge&label=build" alt="Build status"></a>
    <a href="LICENSE"><img src="https://img.shields.io/github/license/cybertortuga/aitriage?style=for-the-badge" alt="MIT license"></a>
    <a href="https://github.com/cybertortuga/aitriage/pkgs/container/aitriage"><img src="https://img.shields.io/badge/GHCR-container-2496ED?style=for-the-badge&logo=docker&logoColor=white" alt="GHCR container"></a>
  </p>

  <p>
    <a href="#the-easiest-way-let-your-ai-ide-set-it-up"><strong>AI IDE setup</strong></a> ·
    <a href="#quick-start"><strong>Quick start</strong></a> ·
    <a href="#where-are-the-results"><strong>Reports</strong></a> ·
    <a href="#cicd"><strong>CI/CD</strong></a>
  </p>
</div>

---

AITriage checks a codebase for security problems, asks an AI coding agent to verify every finding, and prepares safe fix instructions.

It works with:

- **Codex** using your Codex subscription;
- **Claude Code** using your Claude subscription;
- **GitHub Actions** using the same SecureCoder analysis pipeline in CI/CD;
- **CLI** as a deterministic scan without AI or full AI triage with your provider key;
- **Web UI** for local, visual project scanning and report review.

AITriage does not silently fix code. It first shows the result. Source changes begin only after the user explicitly asks to fix selected findings.

## The easiest way: let your AI IDE set it up

Open the project root in Codex or Claude Code and paste this exact request:

```text
Set up AITriage for this project using the official instructions:
https://github.com/cybertortuga/aitriage

AITriage is a system tool, not a dependency of my project. Never clone or copy
the AITriage repository inside the project currently open in this AI IDE.
Install the released CLI separately with Homebrew or `go install`, then connect
that CLI to the current project. If source code must be cloned for development,
put it outside this project, for example as a sibling directory:

GitHub/
├── aitriage/          # the AITriage source repository
└── my-vibe-project/   # the project being checked

Preserve my existing settings and source files. Only add the project-local MCP
configuration, managed AI instructions, `/aitriage-reports/` gitignore entry,
and generated reports required by the official installer. Do not start a scan
or modify source code yet. Verify the connection and tell me when I must open a
new task/session.
```

After setup, open a **new** task/session in the same project and paste:

```text
Check this project with AITriage. Do not fix anything yet. Explain the result
in plain language and show me where the report was saved.
```

That is enough. The installer adds the exact MCP workflow contract for the AI IDE; the user does not need to know internal tool names.

AITriage itself is installed outside the checked project. The installer writes only the small project-local integration files needed by Codex or Claude Code; it never vendors the AITriage source tree into the project.

## Quick start

### 1. Install AITriage

macOS or Linux with Homebrew:

```bash
brew install cybertortuga/aitriage/aitriage
```

Or with Go 1.25.5+:

```bash
go install github.com/cybertortuga/aitriage/cmd/aitriage@latest
```

Check the installation:

```bash
aitriage version
```

### 2. Connect your AI IDE once

Open a terminal in the root of your project.

For Codex:

```bash
aitriage install-codex .
```

For Claude Code:

```bash
aitriage install-claude-code .
```

Then open a **new** Codex task or Claude Code session in that project. The connection is project-local: AITriage can inspect the root and any folder inside it, but cannot access paths outside it.

### 3. Ask for a security check

Write this in Codex or Claude Code:

```text
Check this repository with AITriage. Do not fix anything yet. Show me confirmed problems, uncertain findings, and false positives separately.
```

AITriage will:

1. scan the project with deterministic security rules;
2. send findings through the built-in SecureCoder prompts;
3. use the AI model from your Codex or Claude subscription;
4. verify and classify the findings;
5. create a readable report and a fix specification;
6. stop and wait for your decision.

### 4. Decide what to fix

To fix specific findings:

```text
Fix confirmed findings CS-AUTH-001 and CS-AUTHZ-001 from the latest AITriage report. Run the required tests and verify the fixes with AITriage.
```

To fix every confirmed True Positive:

```text
Fix all confirmed True Positives from the latest AITriage report. Do not fix uncertain findings or false positives. Run the required tests and verify the result with AITriage.
```

No finding is approved for fixing until you explicitly say so.

## Where are the results?

Everything produced by a local AI-assisted run is stored in:

```text
aitriage-reports/
└── run-.../
    ├── summary.md           # start here: short result for a human
    ├── report.md            # full security report
    ├── fixspec.md           # exact instructions for approved fixes
    ├── triage-findings.json # machine-readable AI verdicts
    ├── aitriage.sarif       # result for security tools
    ├── scan.json            # raw scan input
    ├── manifest.json        # run state and integrity metadata
    └── audit.log            # run events
```

`aitriage-reports/` is ignored during later scans so generated reports cannot create an analysis loop. Add this line to the project `.gitignore`:

```gitignore
/aitriage-reports/
```

## Check only one folder

If your repository contains several projects, install AITriage once at the repository root and name the folder in your request:

```text
Check services/payments with AITriage. Do not fix anything yet.
```

You do not need a separate MCP configuration for every subfolder.

## Scan without AI

For a fast raw scan:

```bash
aitriage scan .
```

This does not require an API key, Codex, or Claude. It is useful for a quick check, but it is **not** the final AI-triaged verdict.

Useful variants:

```bash
aitriage scan ./service
aitriage scan . --staged
aitriage scan . --diff origin/main
aitriage scan . --format sarif -o aitriage.sarif
```

## Run AI triage from the CLI

Use this when you want full SecureCoder triage without Codex or Claude Code. It requires an API key for your selected model provider:

```bash
export GEMINI_API_KEY="your-key"
aitriage agent .
```

Other supported providers use `ANTHROPIC_API_KEY`, `OPENAI_API_KEY`, or `GROQ_API_KEY`. The AI IDE workflow does not need these keys because it uses the active Codex or Claude subscription.

## CI/CD

The canonical setup keeps a small workflow in each application repository and calls one centrally maintained reusable workflow. CI uses the same scan, SecureCoder prompts, triage artifacts, and final policy gate as the local AI-assisted flow.

Create `.github/workflows/aitriage.yml`:

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

Replace `your-org/security-workflows` with your reusable-workflow repository and add the selected provider key to GitHub Actions secrets.

## Update or remove the AI IDE connection

Run the same install command again to update it:

```bash
aitriage install-codex .
aitriage install-claude-code .
```

Remove only AITriage-managed settings:

```bash
aitriage install-codex . --uninstall
aitriage install-claude-code . --uninstall
```

Uninstall preserves unrelated MCP servers and your own instruction text.

## Safety guarantees

- AITriage uses its bundled SecureCoder prompts; the AI IDE must not invent a replacement analysis flow.
- The MCP server is confined to the configured project root and its real subdirectories.
- Path traversal and symlink escapes outside the project are rejected.
- Source changes require an explicit user request.
- False Positives and uncertain findings are not fixed automatically.
- Reports and run state are kept under `aitriage-reports/` and excluded from analysis.

## Web UI

From a cloned AITriage repository, run:

```bash
make up
```

Open `http://localhost:8080`. The current Web UI has no enforced login and is intended for a trusted local machine or isolated network. Do not expose it directly to the Internet.

## Troubleshooting

| Problem | What to do |
| :--- | :--- |
| Codex or Claude cannot see AITriage | Run the matching install command from the project root, then open a new task/session |
| Claude shows MCP approval as pending | Open the project in Claude Code and approve the `aitriage` MCP server |
| A subfolder is rejected | Confirm it is a real folder inside the root used during installation |
| The AI runs only `aitriage scan` | Ask it to use the AITriage MCP workflow; raw scan is not full triage |
| The run takes several minutes | SecureCoder performs several analysis stages; wait unless the task reports an error |
| Full AI triage fails | Keep the failure visible; do not treat raw scan output as the final verdict |

## More documentation

- [Configuration example](.aitriage.yaml.example)
- [Security rules](rules/README.md)
- [GitHub Actions examples](examples/github-actions/)
- [Contributing](CONTRIBUTING.md)

## License

MIT. See [LICENSE](LICENSE).
