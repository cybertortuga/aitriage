package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// serversUnderKey returns the servers map nested under an arbitrary top-level
// key ("mcpServers" for most clients, "servers" for VS Code).
func serversUnderKey(t *testing.T, data []byte, key string) map[string]any {
	t.Helper()
	var doc map[string]any
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, data)
	}
	servers, _ := doc[key].(map[string]any)
	return servers
}

// withCleanInstallState resets the shared install flags for a test.
func withCleanInstallState(t *testing.T) {
	t.Helper()
	prevUninstall, prevRuntime := clientUninstall, clientRuntime
	clientUninstall, clientRuntime = false, "container"
	t.Cleanup(func() { clientUninstall, clientRuntime = prevUninstall, prevRuntime })
}

func TestJSONClientCommandsRegisteredWithFlags(t *testing.T) {
	for _, cmd := range []*cobra.Command{installCursorCmd, installAntigravityCmd, installVSCodeCmd} {
		if flag := cmd.Flags().Lookup("runtime"); flag == nil || flag.DefValue != "container" {
			t.Fatalf("%s --runtime default = %v, want container", cmd.Name(), flag)
		}
		if flag := cmd.Flags().Lookup("uninstall"); flag == nil {
			t.Fatalf("%s missing --uninstall flag", cmd.Name())
		}
	}
}

// TestJSONClientInstallWritesProjectLocalConfig proves each connector writes its
// own project-local file, under the correct top-level key, with the safe
// profile and the project scan-root baked in.
func TestJSONClientInstallWritesProjectLocalConfig(t *testing.T) {
	cases := []struct {
		client          jsonClient
		relPath         string
		serversKey      string
		wantType        bool   // expects "type":"stdio" on the entry
		wantFrontmatter string // non-empty ⇒ dedicated rule file must start with it
	}{
		{cursorClient(), ".cursor/mcp.json", "mcpServers", false, ""},
		{antigravityClient(), ".agents/mcp_config.json", "mcpServers", false, "---\ntrigger: always_on\n---"},
		{vscodeClient(), ".vscode/mcp.json", "servers", true, ""},
	}
	for _, tc := range cases {
		t.Run(tc.client.cmdWord, func(t *testing.T) {
			withCleanInstallState(t)
			projDir := t.TempDir()
			root, err := filepath.EvalSymlinks(projDir)
			if err != nil {
				t.Fatal(err)
			}

			if err := runInstallJSONClient(tc.client, []string{projDir}); err != nil {
				t.Fatalf("install: %v", err)
			}

			cfgPath := filepath.Join(root, filepath.FromSlash(tc.relPath))
			data, err := os.ReadFile(cfgPath)
			if err != nil {
				t.Fatalf("config not written at %s: %v", tc.relPath, err)
			}
			servers := serversUnderKey(t, data, tc.serversKey)
			entry, ok := servers[mcpServerName].(map[string]any)
			if !ok {
				t.Fatalf("aitriage entry missing under %q:\n%s", tc.serversKey, data)
			}
			args, _ := json.Marshal(entry["args"])
			for _, want := range []string{"serve", "--profile", "safe", "--runtime", "container", "--scan-root", root} {
				if !strings.Contains(string(args), want) {
					t.Errorf("safe-profile arg %q missing:\n%s", want, args)
				}
			}
			if tc.wantType && entry["type"] != "stdio" {
				t.Errorf("%s entry must declare \"type\":\"stdio\":\n%s", tc.client.name, data)
			}
			if !tc.wantType {
				if _, present := entry["type"]; present {
					t.Errorf("%s entry should not carry a type field:\n%s", tc.client.name, data)
				}
			}

			// Agent contract + gitignore side-effects.
			contract, err := os.ReadFile(filepath.Join(root, tc.client.contractFile))
			if err != nil {
				t.Fatalf("contract %s not written: %v", tc.client.contractFile, err)
			}
			if !strings.Contains(string(contract), agentContractHeading) ||
				!strings.Contains(string(contract), "never fall back to raw") {
				t.Errorf("contract missing expected content:\n%s", contract)
			}
			if tc.wantFrontmatter != "" {
				// Dedicated rule file: frontmatter must be the very first bytes
				// (YAML parsing requires it) and there must be no managed-block
				// markers wrapping it.
				if !strings.HasPrefix(string(contract), tc.wantFrontmatter) {
					t.Errorf("%s rule must start with frontmatter %q:\n%s", tc.client.name, tc.wantFrontmatter, contract)
				}
				if strings.Contains(string(contract), agentContractBegin) {
					t.Errorf("dedicated rule file should not carry managed-block markers:\n%s", contract)
				}
			} else if !strings.Contains(string(contract), agentContractBegin) {
				t.Errorf("shared instruction file must use managed-block markers:\n%s", contract)
			}
			gi, _ := os.ReadFile(filepath.Join(root, ".gitignore"))
			if !strings.Contains(string(gi), reportsGitignoreEntry) {
				t.Errorf(".gitignore missing reports entry: %s", gi)
			}
		})
	}
}

// TestJSONClientInstallPreservesUnrelatedServersAndIsIdempotent proves the
// connector never clobbers other servers or top-level keys, and re-running is a
// no-op update.
func TestJSONClientInstallPreservesUnrelatedServersAndIsIdempotent(t *testing.T) {
	withCleanInstallState(t)
	client := cursorClient()
	root := t.TempDir()
	cfgPath := client.configPath(root)
	if err := os.MkdirAll(filepath.Dir(cfgPath), 0o755); err != nil {
		t.Fatal(err)
	}
	seed := []byte(`{"mcpServers":{"other":{"command":"keepme"}},"topKey":7}`)
	if err := os.WriteFile(cfgPath, seed, 0o644); err != nil {
		t.Fatal(err)
	}

	if err := runInstallJSONClient(client, []string{root}); err != nil {
		t.Fatalf("install: %v", err)
	}
	once, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	servers := serversUnderKey(t, once, "mcpServers")
	if servers["other"] == nil {
		t.Error("unrelated server 'other' was dropped")
	}
	if servers[mcpServerName] == nil {
		t.Error("aitriage server not added")
	}
	var doc map[string]any
	_ = json.Unmarshal(once, &doc)
	if doc["topKey"] == nil {
		t.Error("unrelated top-level key 'topKey' was dropped")
	}

	if err := runInstallJSONClient(client, []string{root}); err != nil {
		t.Fatalf("re-install: %v", err)
	}
	twice, _ := os.ReadFile(cfgPath)
	if string(once) != string(twice) {
		t.Fatalf("install not idempotent:\n--- once ---\n%s\n--- twice ---\n%s", once, twice)
	}
}

// TestJSONClientUninstallRemovesOnlyAITriage proves --uninstall strips the
// aitriage server and its managed contract block while leaving unrelated
// servers and the file itself intact.
func TestJSONClientUninstallRemovesOnlyAITriage(t *testing.T) {
	withCleanInstallState(t)
	client := vscodeClient() // exercises the "servers" key path too
	root := t.TempDir()
	cfgPath := client.configPath(root)
	if err := os.MkdirAll(filepath.Dir(cfgPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(cfgPath, []byte(`{"servers":{"other":{"command":"keepme"}}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	// Project already had its own instructions we must not delete.
	contractPath := filepath.Join(root, client.contractFile)
	if err := os.MkdirAll(filepath.Dir(contractPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(contractPath, []byte("# House rules\n\nKeep me.\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := runInstallJSONClient(client, []string{root}); err != nil {
		t.Fatalf("install: %v", err)
	}

	clientUninstall = true
	if err := runInstallJSONClient(client, []string{root}); err != nil {
		t.Fatalf("uninstall: %v", err)
	}
	clientUninstall = false

	data, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Fatalf("config file should survive uninstall: %v", err)
	}
	servers := serversUnderKey(t, data, "servers")
	if servers[mcpServerName] != nil {
		t.Errorf("aitriage server not removed:\n%s", data)
	}
	if servers["other"] == nil {
		t.Errorf("unrelated server dropped by uninstall:\n%s", data)
	}
	contract, _ := os.ReadFile(contractPath)
	if !strings.Contains(string(contract), "Keep me.") {
		t.Errorf("project instructions lost on uninstall:\n%s", contract)
	}
	if strings.Contains(string(contract), agentContractBegin) {
		t.Errorf("managed contract block not removed on uninstall:\n%s", contract)
	}
}

// TestJSONClientInstallTwoProjectsNoLeak proves installing for a second project
// neither overwrites the first nor leaks its scan-root — the per-project
// isolation guarantee.
func TestJSONClientInstallTwoProjectsNoLeak(t *testing.T) {
	withCleanInstallState(t)
	client := cursorClient()
	projA, projB := t.TempDir(), t.TempDir()
	rootA, err := filepath.EvalSymlinks(projA)
	if err != nil {
		t.Fatal(err)
	}
	rootB, err := filepath.EvalSymlinks(projB)
	if err != nil {
		t.Fatal(err)
	}

	if err := runInstallJSONClient(client, []string{projA}); err != nil {
		t.Fatalf("install A: %v", err)
	}
	if err := runInstallJSONClient(client, []string{projB}); err != nil {
		t.Fatalf("install B: %v", err)
	}

	cfgA, _ := os.ReadFile(client.configPath(rootA))
	cfgB, _ := os.ReadFile(client.configPath(rootB))
	if !strings.Contains(string(cfgA), rootA) || strings.Contains(string(cfgA), rootB) {
		t.Errorf("project A config leaked or missed a scan-root:\n%s", cfgA)
	}
	if !strings.Contains(string(cfgB), rootB) || strings.Contains(string(cfgB), rootA) {
		t.Errorf("project B config leaked or missed a scan-root:\n%s", cfgB)
	}
}

// TestJSONClientPathsAreProjectLocal locks in the config file locations so a
// refactor cannot silently move them to a global path.
func TestJSONClientPathsAreProjectLocal(t *testing.T) {
	root := t.TempDir()
	cases := map[string]string{
		cursorClient().configPath(root):      filepath.Join(root, ".cursor", "mcp.json"),
		antigravityClient().configPath(root): filepath.Join(root, ".agents", "mcp_config.json"),
		vscodeClient().configPath(root):      filepath.Join(root, ".vscode", "mcp.json"),
	}
	for got, want := range cases {
		if got != want {
			t.Errorf("config path = %q, want project-local %q", got, want)
		}
	}
}

// TestJSONSetServerWithKeyRespectsKey proves the VS Code "servers" key and the
// default "mcpServers" key are handled independently within one document.
func TestJSONSetServerWithKeyRespectsKey(t *testing.T) {
	out, err := jsonSetServerWithKey(nil, "servers", "aitriage", map[string]any{"command": "x"})
	if err != nil {
		t.Fatal(err)
	}
	if serversUnderKey(t, out, "servers")["aitriage"] == nil {
		t.Errorf("aitriage not written under 'servers':\n%s", out)
	}
	if serversUnderKey(t, out, "mcpServers") != nil {
		t.Errorf("unexpected 'mcpServers' key created:\n%s", out)
	}

	// A document can carry both keys without cross-contamination.
	both, err := jsonSetServerWithKey(out, "mcpServers", "aitriage", map[string]any{"command": "y"})
	if err != nil {
		t.Fatal(err)
	}
	if serversUnderKey(t, both, "servers")["aitriage"] == nil ||
		serversUnderKey(t, both, "mcpServers")["aitriage"] == nil {
		t.Errorf("both keys should coexist:\n%s", both)
	}
	_, changed, err := jsonRemoveServerWithKey(both, "servers", "aitriage")
	if err != nil || !changed {
		t.Fatalf("remove under 'servers': changed=%v err=%v", changed, err)
	}
}

// TestSharedAgentsMdSurvivesSiblingUninstall proves that when two AGENTS.md
// clients are connected (Codex + Cursor), uninstalling one keeps the shared
// AGENTS.md contract the other still relies on, and only the last uninstall
// removes it.
func TestSharedAgentsMdSurvivesSiblingUninstall(t *testing.T) {
	withCleanInstallState(t)
	root := t.TempDir()
	agentsMd := filepath.Join(root, "AGENTS.md")

	if err := runInstallCodex(nil, []string{root}); err != nil {
		t.Fatalf("install codex: %v", err)
	}
	if err := runInstallJSONClient(cursorClient(), []string{root}); err != nil {
		t.Fatalf("install cursor: %v", err)
	}

	// Uninstall Cursor: Codex still connected → AGENTS.md must survive.
	clientUninstall = true
	if err := runInstallJSONClient(cursorClient(), []string{root}); err != nil {
		t.Fatalf("uninstall cursor: %v", err)
	}
	data, err := os.ReadFile(agentsMd)
	if err != nil {
		t.Fatalf("AGENTS.md deleted while Codex still relies on it: %v", err)
	}
	if !strings.Contains(string(data), agentContractBegin) {
		t.Fatalf("shared contract block stripped while Codex still installed:\n%s", data)
	}
	// Cursor's own MCP config is gone.
	if _, err := os.Stat(cursorClient().configPath(root)); err == nil {
		if jsonConfigHasServer(cursorClient().configPath(root), "mcpServers", mcpServerName) {
			t.Error("cursor MCP server not removed on uninstall")
		}
	}

	// Now uninstall Codex too: no sibling left → AGENTS.md is removed.
	if err := runInstallCodex(nil, []string{root}); err != nil {
		t.Fatalf("uninstall codex: %v", err)
	}
	clientUninstall = false
	if _, err := os.Stat(agentsMd); !os.IsNotExist(err) {
		t.Fatalf("AGENTS.md should be removed once no AGENTS.md client remains (err=%v)", err)
	}
}

// TestAntigravityRuleFileLifecycle proves Antigravity's contract is an
// AITriage-owned rule at .agents/rules/aitriage.md (frontmatter first), that
// uninstall removes the whole file, and that a same-named file NOT authored by
// AITriage is never clobbered or deleted.
func TestAntigravityRuleFileLifecycle(t *testing.T) {
	withCleanInstallState(t)
	client := antigravityClient()
	root := t.TempDir()
	rulePath := filepath.Join(root, client.contractFile)

	if err := runInstallJSONClient(client, []string{root}); err != nil {
		t.Fatalf("install: %v", err)
	}
	data, err := os.ReadFile(rulePath)
	if err != nil {
		t.Fatalf("rule file not written: %v", err)
	}
	if !strings.HasPrefix(string(data), "---\ntrigger: always_on\n---") {
		t.Errorf("rule must open with trigger frontmatter:\n%s", data)
	}
	if strings.Contains(string(data), agentContractBegin) {
		t.Errorf("dedicated rule must not carry managed-block markers:\n%s", data)
	}

	clientUninstall = true
	if err := runInstallJSONClient(client, []string{root}); err != nil {
		t.Fatalf("uninstall: %v", err)
	}
	clientUninstall = false
	if _, err := os.Stat(rulePath); !os.IsNotExist(err) {
		t.Errorf("uninstall must remove the AITriage rule file (err=%v)", err)
	}

	// A user's own rule of the same name must be refused on install and left
	// intact on uninstall.
	if err := os.MkdirAll(filepath.Dir(rulePath), 0o755); err != nil {
		t.Fatal(err)
	}
	foreign := "---\ntrigger: model_decision\n---\n\n# My own rule\n"
	if err := os.WriteFile(rulePath, []byte(foreign), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := runInstallJSONClient(client, []string{root}); err == nil {
		t.Error("install must refuse to overwrite a non-AITriage rule file")
	}
	clientUninstall = true
	if err := runInstallJSONClient(client, []string{root}); err != nil {
		t.Fatalf("uninstall with foreign file present: %v", err)
	}
	clientUninstall = false
	if got, _ := os.ReadFile(rulePath); string(got) != foreign {
		t.Errorf("uninstall must not touch a user's own rule file:\n%s", got)
	}
}

func TestJSONClientContractFilePathHasNoSurprises(t *testing.T) {
	if runtime.GOOS != "windows" && vscodeClient().contractFile != ".github/copilot-instructions.md" {
		t.Errorf("VS Code contract file = %q", vscodeClient().contractFile)
	}
}
