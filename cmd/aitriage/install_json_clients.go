package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// This file wires the AITriage MCP server into the AI IDEs that configure MCP
// through a project-local JSON file: Cursor, Antigravity, and VS Code (Copilot
// agent mode).
//
// Unlike Claude Code, none of these clients expose a scriptable "add and trust"
// command such as `claude mcp add`. The most we can do honestly is write the
// project-local config; the human must then ENABLE the server and APPROVE its
// tools in the client UI. Every installer says so explicitly instead of
// pretending the server is live.
//
// All three write project-scoped configs (never a global one), so each project
// keeps its own safe scan-root and one project's server is never exposed to
// another — the same isolation guarantee as the Codex and Claude Code
// installers. Windsurf is intentionally not here: it only supports a single
// global ~/.codeium/windsurf/mcp_config.json, which cannot preserve per-project
// isolation.
//
// Each client's agent contract goes to the file that client actually reads:
// Cursor → project-root AGENTS.md (shared with Codex), VS Code →
// .github/copilot-instructions.md, Antigravity → .agents/rules/aitriage.md
// (an AITriage-owned rule file with `trigger: always_on` frontmatter, since
// Antigravity reads workspace rules from .agents/rules/, not AGENTS.md).
//
// Every installer is idempotent (install == update) and never rewrites unrelated
// servers or keys. --uninstall removes only the aitriage entry and its managed
// agent-contract block.

// jsonClient describes a client whose MCP config is a project-local JSON file.
type jsonClient struct {
	// name is the human-facing client label, e.g. "Cursor".
	name string
	// cmdWord is the CLI verb suffix, e.g. "cursor" for `aitriage install-cursor`.
	cmdWord string
	// configPath returns the absolute path of the project-local config for root.
	configPath func(root string) string
	// serversKey is the top-level JSON key servers nest under: "mcpServers" for
	// Cursor/Antigravity, "servers" for VS Code.
	serversKey string
	// stdioType, when true, adds "type":"stdio" to the server entry. VS Code
	// wants an explicit transport type; Cursor and Antigravity infer it from the
	// presence of "command".
	stdioType bool
	// contractFile is the agent-instruction file this client reads, relative to
	// the project root, e.g. "AGENTS.md".
	contractFile string
	// contract is the managed instruction block content.
	contract string
	// ruleFrontmatter, when non-empty, marks contractFile as an AITriage-OWNED
	// rule file rather than a shared instructions file. The file is written as
	// ruleFrontmatter + contract (frontmatter at the very first byte, as YAML
	// parsing requires) and is removed wholesale on uninstall. Antigravity needs
	// this: its workspace rules live in .agents/rules/*.md and must start with a
	// `trigger:` frontmatter block. Shared files (AGENTS.md,
	// copilot-instructions.md) leave this empty and use the managed-block writer.
	ruleFrontmatter string
	// enableHint is the client-specific sentence telling the user how to enable
	// the freshly written server in the UI.
	enableHint string
}

func cursorClient() jsonClient {
	return jsonClient{
		name:    "Cursor",
		cmdWord: "cursor",
		configPath: func(root string) string {
			return filepath.Join(root, ".cursor", "mcp.json")
		},
		serversKey:   "mcpServers",
		stdioType:    false,
		contractFile: "AGENTS.md",
		contract:     agentsMdContent,
		enableHint:   "Open Cursor → Settings → MCP (or the MCP list), toggle the 'aitriage' server ON, and approve its tools when prompted.",
	}
}

func antigravityClient() jsonClient {
	return jsonClient{
		name:    "Antigravity",
		cmdWord: "antigravity",
		configPath: func(root string) string {
			return filepath.Join(root, ".agents", "mcp_config.json")
		},
		serversKey: "mcpServers",
		stdioType:  false,
		// Antigravity reads workspace rules from .agents/rules/*.md (NOT a
		// project-root AGENTS.md), and each rule file must open with a `trigger`
		// frontmatter block. A security contract is exactly the kind of rule that
		// should always apply.
		contractFile:    filepath.Join(".agents", "rules", "aitriage.md"),
		contract:        agentsMdContent,
		ruleFrontmatter: "---\ntrigger: always_on\n---\n\n",
		enableHint:      "In Antigravity, open the Agent panel → ... → MCP Servers → Manage MCP Servers, enable 'aitriage', and approve its tools.",
	}
}

func vscodeClient() jsonClient {
	return jsonClient{
		name:    "VS Code",
		cmdWord: "vscode",
		configPath: func(root string) string {
			return filepath.Join(root, ".vscode", "mcp.json")
		},
		serversKey:   "servers",
		stdioType:    true,
		contractFile: filepath.Join(".github", "copilot-instructions.md"),
		contract:     agentsMdContent,
		enableHint:   "Open the repo in VS Code, switch Copilot Chat to Agent mode, then Start the 'aitriage' server from .vscode/mcp.json and allow its tools.",
	}
}

var installCursorCmd = newJSONClientCommand(cursorClient(),
	"Install the AITriage MCP server into Cursor (project-local .cursor/mcp.json)")
var installAntigravityCmd = newJSONClientCommand(antigravityClient(),
	"Install the AITriage MCP server into Antigravity (workspace .agents/mcp_config.json)")
var installVSCodeCmd = newJSONClientCommand(vscodeClient(),
	"Install the AITriage MCP server into VS Code Copilot agent mode (project-local .vscode/mcp.json)")

func init() {
	for _, cmd := range []*cobra.Command{installCursorCmd, installAntigravityCmd, installVSCodeCmd} {
		rootCmd.AddCommand(cmd)
		cmd.Flags().BoolVar(&clientUninstall, "uninstall", false, "Remove the AITriage entry instead of installing it")
		cmd.Flags().StringVar(&clientRuntime, "runtime", "container", "MCP server runtime: container (default, full scanner bundle) or native (development only)")
	}
}

// newJSONClientCommand builds the cobra command for a JSON-configured client.
func newJSONClientCommand(client jsonClient, short string) *cobra.Command {
	return &cobra.Command{
		Use:   "install-" + client.cmdWord + " [project-path]",
		Short: short,
		Long: fmt.Sprintf(`Wire AITriage into %s as a local stdio MCP server using the safe profile
(read-only tools, scans confined to the project directory).

Writes the project-local %s (never a global config), so each project keeps its
own scan-root and one project's server is never exposed to another. Unrelated
servers and keys in the file are preserved.

IMPORTANT: %s cannot approve an MCP server from the command line the way
Claude Code can. This command only writes the config; you must then enable the
'aitriage' server and approve its tools in the %s UI, then open a new
task/session so the client loads it. Until you do, do NOT accept a raw
`+"`aitriage scan`"+` as the AITriage audit — it is an untriaged pre-scan, not a verdict.

Idempotent: re-running updates the existing entry. Use --uninstall to remove it.`,
			client.name, client.configPath("<project>"), client.name, client.name),
		Example: fmt.Sprintf(`  aitriage install-%s
  aitriage install-%s ./my-project
  aitriage install-%s --uninstall`, client.cmdWord, client.cmdWord, client.cmdWord),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInstallJSONClient(client, args)
		},
	}
}

// sharesAgentsMd reports whether this client's contract lives in the shared
// project-root AGENTS.md. Only Codex and Cursor do; Antigravity has its own
// .agents/rules/aitriage.md and VS Code its own .github/copilot-instructions.md,
// so neither collides.
func (c jsonClient) sharesAgentsMd() bool {
	return c.contractFile == "AGENTS.md"
}

// syncContract installs or removes this client's agent contract, dispatching to
// the dedicated-rule writer for AITriage-owned rule files (Antigravity) and to
// the shared managed-block writer for shared instruction files (AGENTS.md,
// copilot-instructions.md).
func (c jsonClient) syncContract(root string, uninstall bool) (bool, error) {
	if c.ruleFrontmatter != "" {
		return syncDedicatedRule(root, c.contractFile, c.ruleFrontmatter, c.contract, uninstall)
	}
	return syncAgentContract(root, c.contractFile, c.contract, uninstall)
}

// syncDedicatedRule manages an AITriage-owned rule file whose entire contents
// (frontmatter + contract) belong to AITriage. It never merges with unrelated
// content: on install it writes the file only if absent or already AITriage's;
// on uninstall it removes the file only if AITriage authored it. A same-named
// file that is not ours is left untouched (install errors rather than clobber).
func syncDedicatedRule(root, file, frontmatter, contract string, uninstall bool) (bool, error) {
	path := filepath.Join(root, file)
	existing, err := readFileOrEmpty(path)
	if err != nil {
		return false, err
	}
	ours := strings.Contains(existing, agentContractHeading)

	if uninstall {
		if existing == "" || !ours {
			return false, nil
		}
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			return false, err
		}
		return true, nil
	}

	if existing != "" && !ours {
		return false, fmt.Errorf("%s already exists and is not AITriage-managed; not overwriting", path)
	}
	want := ensureTrailingNewline(frontmatter + strings.TrimSpace(contract))
	if existing == want {
		return false, nil
	}
	return true, writeFilePreservingMode(path, []byte(want), 0o644)
}

// jsonConfigHasServer reports whether the JSON config at path already declares
// the named server under key. Missing/empty/invalid files count as absent.
func jsonConfigHasServer(path, key, name string) bool {
	data, err := readBytesOrEmpty(path)
	if err != nil || len(data) == 0 {
		return false
	}
	doc, err := decodeJSONObject(data)
	if err != nil {
		return false
	}
	servers, _ := doc[key].(map[string]any)
	if servers == nil {
		return false
	}
	_, ok := servers[name]
	return ok
}

// agentsMdSiblingInstalled reports whether an AGENTS.md-consuming client OTHER
// than exceptCmdWord still has the aitriage server configured under root. It
// guards uninstall so removing one client's connector never strips the shared
// AGENTS.md contract another client still relies on.
func agentsMdSiblingInstalled(root, exceptCmdWord string) bool {
	if exceptCmdWord != "codex" {
		if data, _ := readFileOrEmpty(codexConfigPath(root)); strings.Contains(data, "[mcp_servers."+mcpServerName+"]") {
			return true
		}
	}
	if exceptCmdWord != "cursor" && jsonConfigHasServer(cursorClient().configPath(root), "mcpServers", mcpServerName) {
		return true
	}
	// Antigravity is intentionally not consulted: it owns .agents/rules/aitriage.md
	// and never shares the project-root AGENTS.md, so it cannot keep it alive.
	return false
}

func runInstallJSONClient(client jsonClient, args []string) error {
	root, err := resolveProjectRoot(args)
	if err != nil {
		return err
	}
	configPath := client.configPath(root)
	existing, err := readBytesOrEmpty(configPath)
	if err != nil {
		return err
	}

	if clientUninstall {
		updated, changed, err := jsonRemoveServerWithKey(existing, client.serversKey, mcpServerName)
		if err != nil {
			return err
		}
		if changed {
			if err := writeFilePreservingMode(configPath, updated, 0o644); err != nil {
				return err
			}
		}
		// Keep the shared AGENTS.md contract if another AGENTS.md-based client
		// (Codex/Cursor/Antigravity) is still connected here — removing it would
		// strip a contract that client still relies on.
		contractShared := client.sharesAgentsMd() && agentsMdSiblingInstalled(root, client.cmdWord)
		contractChanged := false
		if !contractShared {
			contractChanged, err = client.syncContract(root, true)
			if err != nil {
				return err
			}
		}
		if !changed && !contractChanged {
			if contractShared {
				fmt.Printf("Removed nothing: another client still uses the shared %s here.\n", client.contractFile)
			} else {
				fmt.Printf("AITriage was not present in %s or %s — nothing to remove.\n", configPath, client.contractFile)
			}
			return nil
		}
		if contractShared {
			fmt.Printf("✅ Removed AITriage from %s config. Kept %s — another client still uses it.\n", client.name, client.contractFile)
		} else {
			fmt.Printf("✅ Removed AITriage from %s and managed %s instructions.\n", client.name, client.contractFile)
		}
		return nil
	}

	if err := validateClientRuntime(); err != nil {
		return err
	}
	bin, err := aitriageBinaryPath()
	if err != nil {
		return err
	}

	entry := map[string]any{
		"command": bin,
		"args":    safeProfileArgs(root),
		"env":     map[string]any{},
	}
	if client.stdioType {
		entry["type"] = "stdio"
	}
	updated, err := jsonSetServerWithKey(existing, client.serversKey, mcpServerName, entry)
	if err != nil {
		return err
	}
	if err := writeFilePreservingMode(configPath, updated, 0o644); err != nil {
		return err
	}
	if _, err := client.syncContract(root, false); err != nil {
		return fmt.Errorf("install %s agent instructions: %w", client.name, err)
	}
	if _, err := ensureGitignoreEntry(root, reportsGitignoreEntry); err != nil {
		return fmt.Errorf("ignore AITriage reports: %w", err)
	}

	fmt.Printf("✅ AITriage written to %s config (safe profile):\n   %s\n", client.name, configPath)
	fmt.Printf("   scan root: %s\n", root)
	fmt.Printf("   agent contract: %s\n", filepath.Join(root, client.contractFile))
	fmt.Printf("⚠️  Not active yet — %s has no scriptable approval.\n", client.name)
	fmt.Printf("   %s\n", client.enableHint)
	fmt.Println("   Then open a NEW task/session so the client loads the server, and ask it to run")
	fmt.Println("   an AITriage audit. If the server is not enabled, the agent must NOT present a raw")
	fmt.Println("   `aitriage scan` as the audit.")
	fmt.Println("   Note: this config contains machine-specific paths; add it to .gitignore if the repo is shared.")
	return nil
}
