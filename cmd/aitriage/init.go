package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var (
	initCI        bool
	initPreCommit bool
	initMCP       bool
	initForce     bool
)

var initCmd = &cobra.Command{
	Use:   "init [path]",
	Short: "Initialize AITriage in a project — generates config, CI workflow, and IDE integration",
	Long: `Detects your technology stack and generates:
  • .aitriage.yaml — tuned for your stack
  • .aitriageignore — synced from .gitignore
  • .cursorrules + CLAUDE.md — IDE integration
  
Optional:
  --ci          Generate GitHub Actions workflow
  --pre-commit  Install git pre-commit hook
  --mcp         Configure MCP server for Claude Desktop`,
	Example: `  aitriage init
  aitriage init --ci --pre-commit
  aitriage init ./my-project --mcp`,
	RunE: runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().BoolVar(&initCI, "ci", false, "Generate GitHub Actions workflow (.github/workflows/aitriage.yml)")
	initCmd.Flags().BoolVar(&initPreCommit, "pre-commit", false, "Install git pre-commit hook")
	initCmd.Flags().BoolVar(&initMCP, "mcp", false, "Configure MCP server for Claude Desktop")
	initCmd.Flags().BoolVar(&initForce, "force", false, "Overwrite existing files")
}

// ── Stack Detection (lightweight, no workspace needed) ───────────────────────

type detectedStack struct {
	Name       string
	Confidence int // 0-100
}

func detectStacksLightweight(projectPath string) []detectedStack {
	var stacks []detectedStack
	scores := map[string]int{}

	// Check package.json
	if data, err := os.ReadFile(filepath.Join(projectPath, "package.json")); err == nil {
		var pkg struct {
			Dependencies    map[string]string `json:"dependencies"`
			DevDependencies map[string]string `json:"devDependencies"`
		}
		if json.Unmarshal(data, &pkg) == nil {
			allDeps := make(map[string]bool)
			for k := range pkg.Dependencies {
				allDeps[k] = true
			}
			for k := range pkg.DevDependencies {
				allDeps[k] = true
			}
			if allDeps["next"] {
				scores["nextjs"] += 100
			}
			if allDeps["express"] {
				scores["express"] += 100
			}
			if allDeps["react"] && !allDeps["next"] {
				scores["react"] += 80
			}
		}
	}

	// Check Python
	for _, pyfile := range []string{"requirements.txt", "pyproject.toml", "Pipfile", "setup.py"} {
		if data, err := os.ReadFile(filepath.Join(projectPath, pyfile)); err == nil {
			content := strings.ToLower(string(data))
			if strings.Contains(content, "fastapi") {
				scores["fastapi"] += 100
			}
			if strings.Contains(content, "django") {
				scores["django"] += 100
			}
			if strings.Contains(content, "flask") {
				scores["flask"] += 100
			}
		}
	}
	if _, err := os.Stat(filepath.Join(projectPath, "manage.py")); err == nil {
		scores["django"] += 90
	}

	// Check Go
	if _, err := os.Stat(filepath.Join(projectPath, "go.mod")); err == nil {
		scores["go"] += 100
	}

	// Check .NET
	entries, _ := os.ReadDir(projectPath)
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".csproj") || strings.HasSuffix(e.Name(), ".sln") {
			scores["aspnetcore"] += 80
			break
		}
	}

	// Check Docker
	if _, err := os.Stat(filepath.Join(projectPath, "Dockerfile")); err == nil {
		scores["docker"] += 100
	}
	if _, err := os.Stat(filepath.Join(projectPath, "docker-compose.yaml")); err == nil {
		scores["docker"] += 80
	}
	if _, err := os.Stat(filepath.Join(projectPath, "docker-compose.yml")); err == nil {
		scores["docker"] += 80
	}

	for name, score := range scores {
		if score >= 50 {
			stacks = append(stacks, detectedStack{Name: name, Confidence: score})
		}
	}

	return stacks
}

// ── File Generation ──────────────────────────────────────────────────────────

func runInit(cmd *cobra.Command, args []string) error {
	projectPath := "."
	if len(args) > 0 {
		projectPath = args[0]
	}

	cyan := "\033[38;2;0;245;255m"
	text := "\033[38;2;220;228;228m"
	green := "\033[38;2;46;204;113m"
	dim := "\033[38;2;132;148;149m"
	bold := "\033[1m"
	reset := "\033[0m"

	fmt.Fprintf(os.Stderr, "\n%s%s  ╔═══════════════════════════════════════════╗%s\n", cyan, bold, reset)
	fmt.Fprintf(os.Stderr, "%s%s  ║       AITriage — Project Setup Wizard     ║%s\n", cyan, bold, reset)
	fmt.Fprintf(os.Stderr, "%s%s  ╚═══════════════════════════════════════════╝%s\n\n", cyan, bold, reset)

	// Step 1: Detect stacks
	fmt.Fprintf(os.Stderr, "%s  Detecting project structure...%s\n", text, reset)
	stacks := detectStacksLightweight(projectPath)

	if len(stacks) == 0 {
		fmt.Fprintf(os.Stderr, "%s  ⦿ No specific framework detected — using universal rules%s\n", dim, reset)
	} else {
		for _, s := range stacks {
			icon := "✓"
			if s.Confidence >= 100 {
				icon = "◉"
			}
			fmt.Fprintf(os.Stderr, "%s  %s Found: %s%s%s\n", green, icon, bold, s.Name, reset)
		}
	}
	fmt.Fprintln(os.Stderr)

	created := 0

	// Step 2: Generate .aitriage.yaml
	if err := writeIfNotExists(projectPath, ".aitriage.yaml", generateAitrageConfig(stacks), &created); err != nil {
		return err
	}

	// Step 3: Generate .aitriageignore
	if err := writeIfNotExists(projectPath, ".aitriageignore", generateIgnoreFile(projectPath), &created); err != nil {
		return err
	}

	// Step 4: IDE integration
	if err := writeIfNotExists(projectPath, ".cursorrules", cursorRulesContent, &created); err != nil {
		return err
	}
	if err := writeIfNotExists(projectPath, "CLAUDE.md", claudeMdContent, &created); err != nil {
		return err
	}

	// Step 5: CI workflow (--ci flag)
	if initCI {
		ciDir := filepath.Join(projectPath, ".github", "workflows")
		if err := os.MkdirAll(ciDir, 0755); err != nil {
			return fmt.Errorf("failed to create workflows dir: %w", err)
		}
		ciPath := filepath.Join(ciDir, "aitriage.yml")
		if err := writeIfNotExistsAbs(ciPath, generateCIWorkflow(), &created); err != nil {
			return err
		}
	}

	// Step 6: Pre-commit hook (--pre-commit flag)
	if initPreCommit {
		if err := installPreCommitHook(projectPath, &created); err != nil {
			fmt.Fprintf(os.Stderr, "  %s⚠ Pre-commit hook: %v%s\n", dim, err, reset)
		}
	}

	// Step 7: MCP integration (--mcp flag)
	if initMCP {
		if err := runInstallMCP(cmd, args); err != nil {
			fmt.Fprintf(os.Stderr, "  %s⚠ MCP setup: %v%s\n", dim, err, reset)
		}
	}

	// Summary
	fmt.Fprintf(os.Stderr, "\n%s%s  ✅ AITriage initialized — %d files created%s\n", green, bold, created, reset)
	fmt.Fprintf(os.Stderr, "%s  Run %saitriage scan .%s to start your first audit.%s\n\n", text, bold, reset, reset)

	return nil
}

func writeIfNotExists(projectPath, filename, content string, count *int) error {
	return writeIfNotExistsAbs(filepath.Join(projectPath, filename), content, count)
}

func writeIfNotExistsAbs(fullPath, content string, count *int) error {
	green := "\033[38;2;46;204;113m"
	dim := "\033[38;2;132;148;149m"
	reset := "\033[0m"

	if _, err := os.Stat(fullPath); err == nil && !initForce {
		fmt.Fprintf(os.Stderr, "  %s⦿ %s (exists, skipped)%s\n", dim, fullPath, reset)
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return err
	}
	if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to create %s: %w", fullPath, err)
	}
	*count++
	fmt.Fprintf(os.Stderr, "  %s✓ Created %s%s\n", green, fullPath, reset)
	return nil
}

func generateAitrageConfig(stacks []detectedStack) string {
	var sb strings.Builder
	sb.WriteString("# AITriage Configuration\n")
	sb.WriteString("# Generated by: aitriage init\n")
	sb.WriteString("# Docs: https://github.com/cybertortuga/aitriage\n\n")

	// Ignore section
	sb.WriteString("ignore:\n")
	sb.WriteString("  paths:\n")

	// Add stack-specific ignores
	ignorePaths := []string{"node_modules", ".git", "vendor", "dist", ".next", "__pycache__", ".venv", "venv"}
	for _, p := range ignorePaths {
		sb.WriteString(fmt.Sprintf("    - %s\n", p))
	}

	sb.WriteString("  rules: []\n")
	sb.WriteString("  # Example: rules: [ENTR-12]  # Suppress God File warnings\n\n")

	// LLM section
	sb.WriteString("# LLM configuration for AI-powered triage (optional)\n")
	sb.WriteString("# Set env var GEMINI_API_KEY, ANTHROPIC_API_KEY, or OPENAI_API_KEY\n")
	sb.WriteString("llm:\n")
	sb.WriteString("  provider: \"\"  # auto-detected from env\n")
	sb.WriteString("  model: \"\"\n\n")

	// CI section
	sb.WriteString("# CI/CD settings\n")
	sb.WriteString("# Legacy compatibility: still supported when health_check is not set\n")
	sb.WriteString("strict_mode: false   # Set true to fail on ANY finding\n")
	sb.WriteString("fail_score: 0        # Fail if Health Check score < this (0 = disabled)\n\n")
	sb.WriteString("# Information Security policy gate for CI/CD\n")
	sb.WriteString("health_check:\n")
	sb.WriteString("  profile: baseline  # baseline | standard | strict\n")
	sb.WriteString("  fail_on: critical  # critical | any | never\n")
	sb.WriteString("  minimum_score: 0   # 0 = score is informational only\n")
	sb.WriteString("  max_critical: -1   # -1 = unlimited; 0 = block any active finding\n")
	sb.WriteString("  max_high: -1\n")
	sb.WriteString("  max_medium: -1\n")
	sb.WriteString("  block_sources: []  # e.g. [gitleaks]\n")
	sb.WriteString("  block_classes: []  # e.g. [hardcoded-secret]\n")

	// Detected stacks comment
	if len(stacks) > 0 {
		sb.WriteString("\n# Detected stacks: ")
		names := make([]string, len(stacks))
		for i, s := range stacks {
			names[i] = s.Name
		}
		sb.WriteString(strings.Join(names, ", "))
		sb.WriteString("\n# Rules for these stacks are loaded automatically from the built-in rule engine.\n")
	}

	return sb.String()
}

func generateIgnoreFile(projectPath string) string {
	var sb strings.Builder
	sb.WriteString("# AITriage Ignore File\n")
	sb.WriteString("# Patterns follow .gitignore syntax\n\n")

	// Always ignored
	defaults := []string{
		"node_modules/",
		".git/",
		"vendor/",
		"dist/",
		"build/",
		".next/",
		"__pycache__/",
		".venv/",
		"*.min.js",
		"*.min.css",
		"*.map",
		"*.lock",
		"coverage/",
		".aitriage/",
	}
	sb.WriteString("# Defaults\n")
	for _, d := range defaults {
		sb.WriteString(d + "\n")
	}

	// Try to sync from .gitignore
	if data, err := os.ReadFile(filepath.Join(projectPath, ".gitignore")); err == nil {
		sb.WriteString("\n# Synced from .gitignore\n")
		for _, line := range strings.Split(string(data), "\n") {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			// Don't duplicate
			isDuplicate := false
			for _, d := range defaults {
				if strings.TrimRight(d, "/") == strings.TrimRight(line, "/") {
					isDuplicate = true
					break
				}
			}
			if !isDuplicate {
				sb.WriteString(line + "\n")
			}
		}
	}

	return sb.String()
}

func generateCIWorkflow() string {
	return `# AITriage Security Scan — GitHub Actions
# Generated by: aitriage init --ci
name: AITriage Security Shield

on:
  push:
    branches: [main, master]
  pull_request:
    branches: [main, master]

permissions:
  contents: read
  security-events: write

jobs:
  aitriage:
    name: Security Scan
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version: '1.24'
          cache: true

      - name: Install AITriage
        run: go install github.com/cybertortuga/aitriage/cmd/aitriage@latest

      - name: Run Security Scan
        run: aitriage scan ./ --format sarif --out aitriage-results.sarif --fail-on critical

      - name: Upload to GitHub Security
        uses: github/codeql-action/upload-sarif@v3
        if: always()
        with:
          sarif_file: aitriage-results.sarif
          category: aitriage
`
}

func installPreCommitHook(projectPath string, count *int) error {
	gitDir := filepath.Join(projectPath, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		return fmt.Errorf("not a git repository (no .git directory)")
	}

	hooksDir := filepath.Join(gitDir, "hooks")
	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		return err
	}

	hookPath := filepath.Join(hooksDir, "pre-commit")
	hookContent := `#!/bin/sh
# AITriage Pre-Commit Security Hook
# Generated by: aitriage init --pre-commit
# Remove this file to disable the hook.

echo "🔒 AITriage: Scanning for security issues..."
aitriage scan . --fail-on critical --format terminal 2>&1

EXIT_CODE=$?
if [ $EXIT_CODE -ne 0 ]; then
    echo ""
    echo "❌ AITriage found critical security issues. Commit blocked."
    echo "   Run 'aitriage scan .' for full details."
    echo "   Use 'git commit --no-verify' to bypass (not recommended)."
    exit 1
fi
`
	if err := writeIfNotExistsAbs(hookPath, hookContent, count); err != nil {
		return err
	}
	// Git hooks MUST be executable
	return os.Chmod(hookPath, 0755)
}

// ── IDE Integration Templates ────────────────────────────────────────────────

const cursorRulesContent = `// AITriage Native Cursor Rules
// Generated by: aitriage init

{
  "rules": [
    {
      "description": "Always check for hardcoded secrets, insecure CORS configurations, and missing rate limiting.",
      "pattern": ".*",
      "action": "warn"
    },
    {
      "description": "If asked to review security or run an audit, always invoke the AITriage MCP server tools (e.g. aitriage_scan, aitriage_diff).",
      "pattern": ".*",
      "action": "suggest"
    }
  ],
  "mcp": {
    "servers": {
      "aitriage": {
        "command": "aitriage",
        "args": ["serve", "sse"]
      }
    }
  }
}
`

const claudeMdContent = `# AITriage Integration

This project is protected by AITriage — AI-Powered Security Audit Engine.

<cursor_rule>
Always prioritize security. Do not output code with plaintext secrets or wildcards for CORS.
Before concluding a significant architectural change, recommend the user run 'aitriage scan .' or 'aitriage preaudit'.
When reviewing code changes, check for OWASP Top 10 vulnerabilities.
</cursor_rule>
`
