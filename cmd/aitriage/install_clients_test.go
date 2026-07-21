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

// ── Codex TOML ───────────────────────────────────────────────────────────────

func TestTomlSetServerOnEmpty(t *testing.T) {
	block := codexServerBlock("aitriage", "/usr/local/bin/aitriage", safeProfileArgs("/proj"))
	got := tomlSetServer("", "aitriage", block)
	if !strings.Contains(got, "[mcp_servers.aitriage]") {
		t.Fatalf("missing header:\n%s", got)
	}
	if !strings.Contains(got, `"serve", "--runtime", "container", "--profile", "safe", "--scan-root", "/proj"`) {
		t.Fatalf("missing safe args:\n%s", got)
	}
	if !strings.Contains(got, "enabled = true") {
		t.Fatalf("project-local install must override a disabled user-level entry:\n%s", got)
	}
}

func TestClientInstallersDefaultToContainerRuntime(t *testing.T) {
	for _, cmd := range []*cobra.Command{installCodexCmd, installClaudeCodeCmd} {
		flag := cmd.Flags().Lookup("runtime")
		if flag == nil || flag.DefValue != "container" {
			t.Fatalf("%s --runtime default = %v, want container", cmd.Name(), flag)
		}
	}

	previous := clientRuntime
	clientRuntime = "native"
	t.Cleanup(func() { clientRuntime = previous })
	got := strings.Join(safeProfileArgs("/proj"), " ")
	if !strings.Contains(got, "--runtime native") {
		t.Fatalf("explicit native opt-in lost: %s", got)
	}
}

func TestTomlSetServerPreservesOthersAndIsIdempotent(t *testing.T) {
	existing := `conversationDetailMode = "STEPS_COMMANDS"

[mcp_servers.node_repl]
command = "/bin/node_repl"
args = []

[mcp_servers.node_repl.env]
FOO = "bar"
`
	block := codexServerBlock("aitriage", "/bin/aitriage", safeProfileArgs("/proj"))
	once := tomlSetServer(existing, "aitriage", block)

	// Unrelated content preserved verbatim.
	for _, want := range []string{
		`conversationDetailMode = "STEPS_COMMANDS"`,
		"[mcp_servers.node_repl]",
		"[mcp_servers.node_repl.env]",
		`FOO = "bar"`,
		"[mcp_servers.aitriage]",
	} {
		if !strings.Contains(once, want) {
			t.Fatalf("result missing %q:\n%s", want, once)
		}
	}

	// Re-running install must not duplicate the aitriage section.
	twice := tomlSetServer(once, "aitriage", block)
	if strings.Count(twice, "[mcp_servers.aitriage]") != 1 {
		t.Fatalf("aitriage section duplicated on re-install:\n%s", twice)
	}
	if once != twice {
		t.Fatalf("install not idempotent:\n--- once ---\n%s\n--- twice ---\n%s", once, twice)
	}
}

func TestTomlRemoveServerDropsSubtables(t *testing.T) {
	existing := `[mcp_servers.node_repl]
command = "/bin/node_repl"

[mcp_servers.aitriage]
command = "/bin/aitriage"
args = ["serve"]

[mcp_servers.aitriage.env]
TOKEN = "x"

[other]
keep = true
`
	got := tomlRemoveServer(existing, "aitriage")
	if strings.Contains(got, "mcp_servers.aitriage") {
		t.Fatalf("aitriage sections not fully removed:\n%s", got)
	}
	for _, want := range []string{"[mcp_servers.node_repl]", "[other]", "keep = true"} {
		if !strings.Contains(got, want) {
			t.Fatalf("removal dropped unrelated content %q:\n%s", want, got)
		}
	}

	// Removing again is a no-op (returns identical content).
	if again := tomlRemoveServer(got, "aitriage"); again != got {
		t.Fatalf("second remove changed content:\n%s", again)
	}
}

func TestTomlQuoteEscapes(t *testing.T) {
	got := tomlQuote(`C:\a"b`)
	if got != `"C:\\a\"b"` {
		t.Fatalf("tomlQuote escape = %s", got)
	}
}

func TestCodexConfigPathIsProjectLocal(t *testing.T) {
	root := t.TempDir()
	want := filepath.Join(root, ".codex", "config.toml")
	if got := codexConfigPath(root); got != want {
		t.Fatalf("codexConfigPath = %q, want project-local %q", got, want)
	}
}

func TestManagedAgentContractPreservesProjectInstructions(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "AGENTS.md")
	original := "# Project Instructions\n\nKeep this text.\n"
	if err := os.WriteFile(path, []byte(original), 0o640); err != nil {
		t.Fatal(err)
	}

	changed, err := syncAgentContract(root, "AGENTS.md", agentsMdContent, false)
	if err != nil || !changed {
		t.Fatalf("install managed contract: changed=%v err=%v", changed, err)
	}
	once := readConfig(t, path)
	if !strings.Contains(once, original) || strings.Count(once, agentContractBegin) != 1 {
		t.Fatalf("project instructions were lost or contract missing:\n%s", once)
	}
	for _, required := range []string{"aitriage_run_approve", "FIRST action", "before planning or editing anything", "status=fixing and fix_context"} {
		if !strings.Contains(once, required) {
			t.Errorf("managed contract missing approval-ordering rule %q", required)
		}
	}
	if info, err := os.Stat(path); err != nil || info.Mode().Perm() != 0o640 {
		t.Fatalf("AGENTS.md mode changed: info=%v err=%v", info, err)
	}

	changed, err = syncAgentContract(root, "AGENTS.md", agentsMdContent, false)
	if err != nil || changed {
		t.Fatalf("second install must be idempotent: changed=%v err=%v", changed, err)
	}
	if twice := readConfig(t, path); twice != once {
		t.Fatal("second install changed AGENTS.md")
	}

	changed, err = syncAgentContract(root, "AGENTS.md", agentsMdContent, true)
	if err != nil || !changed {
		t.Fatalf("remove managed contract: changed=%v err=%v", changed, err)
	}
	if got := readConfig(t, path); got != original {
		t.Fatalf("uninstall did not restore project instructions:\n%s", got)
	}
}

func TestEnsureGitignoreEntryPreservesContentAndIsIdempotent(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, ".gitignore")
	if err := os.WriteFile(path, []byte("node_modules/\n"), 0o640); err != nil {
		t.Fatal(err)
	}
	changed, err := ensureGitignoreEntry(root, reportsGitignoreEntry)
	if err != nil || !changed {
		t.Fatalf("first ensure changed=%v err=%v", changed, err)
	}
	changed, err = ensureGitignoreEntry(root, reportsGitignoreEntry)
	if err != nil || changed {
		t.Fatalf("second ensure changed=%v err=%v", changed, err)
	}
	content := readConfig(t, path)
	if content != "node_modules/\n/aitriage-reports/\n" {
		t.Fatalf("unexpected .gitignore:\n%s", content)
	}
	if info, err := os.Stat(path); err != nil || info.Mode().Perm() != 0o640 {
		t.Fatalf(".gitignore mode changed: info=%v err=%v", info, err)
	}
}

func TestCodexInstallAddsAgentContract(t *testing.T) {
	prevUninstall, prevRuntime := clientUninstall, clientRuntime
	clientUninstall, clientRuntime = false, "container"
	t.Cleanup(func() { clientUninstall, clientRuntime = prevUninstall, prevRuntime })

	root := t.TempDir()
	if err := runInstallCodex(nil, []string{root}); err != nil {
		t.Fatal(err)
	}
	agents := readConfig(t, filepath.Join(root, "AGENTS.md"))
	config := readConfig(t, filepath.Join(root, ".codex", "config.toml"))
	if !strings.Contains(config, `"--runtime", "container"`) {
		t.Fatalf("default installer did not create container-backed MCP config:\n%s", config)
	}
	for _, want := range []string{agentContractBegin, "aitriage_run_start", "never fall back to raw", agentContractEnd} {
		if !strings.Contains(agents, want) {
			t.Fatalf("managed AGENTS.md missing %q:\n%s", want, agents)
		}
	}

	clientUninstall = true
	if err := runInstallCodex(nil, []string{root}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, "AGENTS.md")); !os.IsNotExist(err) {
		t.Fatalf("installer-created AGENTS.md should be removed on uninstall (err=%v)", err)
	}
}

func TestCodexReinstallMigratesNativeEntryToContainer(t *testing.T) {
	previousUninstall, previousRuntime := clientUninstall, clientRuntime
	clientUninstall, clientRuntime = false, "container"
	t.Cleanup(func() { clientUninstall, clientRuntime = previousUninstall, previousRuntime })

	root := t.TempDir()
	configPath := filepath.Join(root, ".codex", "config.toml")
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		t.Fatal(err)
	}
	legacy := codexServerBlock(mcpServerName, "/old/aitriage", []string{"serve", "--runtime", "native", "--profile", "safe", "--scan-root", root})
	if err := os.WriteFile(configPath, []byte(legacy), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := runInstallCodex(nil, []string{root}); err != nil {
		t.Fatal(err)
	}
	updated := readConfig(t, configPath)
	if strings.Contains(updated, `"--runtime", "native"`) || !strings.Contains(updated, `"--runtime", "container"`) {
		t.Fatalf("native entry was not migrated:\n%s", updated)
	}
}

// TestCodexInstallTwoProjectsNoLeak proves each project gets its own
// project-scoped config with its own scan-root, and installing for a second
// project neither overwrites the first nor leaks its scan-root.
func TestCodexInstallTwoProjectsNoLeak(t *testing.T) {
	prev := clientUninstall
	clientUninstall = false
	t.Cleanup(func() { clientUninstall = prev })

	projA := t.TempDir()
	projB := t.TempDir()

	if err := runInstallCodex(nil, []string{projA}); err != nil {
		t.Fatalf("install A: %v", err)
	}
	if err := runInstallCodex(nil, []string{projB}); err != nil {
		t.Fatalf("install B: %v", err)
	}

	cfgA := readConfig(t, filepath.Join(projA, ".codex", "config.toml"))
	cfgB := readConfig(t, filepath.Join(projB, ".codex", "config.toml"))

	if !strings.Contains(cfgA, projA) || strings.Contains(cfgA, projB) {
		t.Errorf("project A config leaked scan-root:\n%s", cfgA)
	}
	if !strings.Contains(cfgB, projB) || strings.Contains(cfgB, projA) {
		t.Errorf("project B config leaked scan-root:\n%s", cfgB)
	}
	if strings.Count(cfgA, "[mcp_servers.aitriage]") != 1 {
		t.Errorf("project A should have exactly one aitriage section:\n%s", cfgA)
	}

	// Uninstall A leaves B untouched.
	clientUninstall = true
	if err := runInstallCodex(nil, []string{projA}); err != nil {
		t.Fatalf("uninstall A: %v", err)
	}
	clientUninstall = false
	if strings.Contains(readConfig(t, filepath.Join(projA, ".codex", "config.toml")), "mcp_servers.aitriage") {
		t.Error("uninstall A did not remove aitriage from project A")
	}
	if !strings.Contains(readConfig(t, filepath.Join(projB, ".codex", "config.toml")), "mcp_servers.aitriage") {
		t.Error("uninstall A wrongly affected project B")
	}
}

func readConfig(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(b)
}

// TestResolveProjectRootCanonicalizesSymlink proves the project root is resolved
// through symlinks, so configs and scan-root use the canonical path clients
// (e.g. Codex trust) key on.
func TestResolveProjectRootCanonicalizesSymlink(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink semantics differ on Windows")
	}
	real := t.TempDir()
	realResolved, err := filepath.EvalSymlinks(real)
	if err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(t.TempDir(), "link")
	if err := os.Symlink(real, link); err != nil {
		t.Fatal(err)
	}

	got, err := resolveProjectRoot([]string{link})
	if err != nil {
		t.Fatalf("resolveProjectRoot: %v", err)
	}
	if got != realResolved {
		t.Errorf("resolveProjectRoot(%q) = %q, want canonical %q", link, got, realResolved)
	}
}

// TestCodexInstallViaSymlinkUsesCanonicalPath proves install-codex writes its
// config under the canonical directory and bakes the canonical scan-root, even
// when invoked through a symlinked path.
func TestCodexInstallViaSymlinkUsesCanonicalPath(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink semantics differ on Windows")
	}
	prev := clientUninstall
	clientUninstall = false
	t.Cleanup(func() { clientUninstall = prev })

	real := t.TempDir()
	realResolved, err := filepath.EvalSymlinks(real)
	if err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(t.TempDir(), "proj-link")
	if err := os.Symlink(real, link); err != nil {
		t.Fatal(err)
	}

	if err := runInstallCodex(nil, []string{link}); err != nil {
		t.Fatalf("install via symlink: %v", err)
	}

	// Config must live under the canonical dir, not the symlink dir.
	if _, err := os.Stat(filepath.Join(realResolved, ".codex", "config.toml")); err != nil {
		t.Fatalf("config not at canonical path: %v", err)
	}
	cfg := readConfig(t, filepath.Join(realResolved, ".codex", "config.toml"))
	if !strings.Contains(cfg, realResolved) {
		t.Errorf("scan-root is not canonical:\n%s", cfg)
	}
}

// TestInitMcpFailingClientExitsNonZero proves that with --mcp, a failing
// mandatory client makes `init` return an error (non-zero exit) instead of a
// green success.
func TestInitMcpFailingClientExitsNonZero(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell fake-binary setup is POSIX-specific")
	}
	// Fake `claude` that always fails, on PATH so runInstallClaudeCode takes the
	// CLI path and gets a non-zero exit.
	binDir := t.TempDir()
	fake := filepath.Join(binDir, "claude")
	if err := os.WriteFile(fake, []byte("#!/bin/sh\nexit 42\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	prevMCP, prevUninstall := initMCP, clientUninstall
	initMCP, clientUninstall = true, false
	t.Cleanup(func() { initMCP, clientUninstall = prevMCP, prevUninstall })

	proj := t.TempDir()
	err := runInit(nil, []string{proj})
	if err == nil {
		t.Fatal("init --mcp with a failing client must return an error (non-zero exit)")
	}
	if !strings.Contains(err.Error(), "claude-code") {
		t.Errorf("error should name the failing client, got: %v", err)
	}
}

// ── Claude Code JSON ─────────────────────────────────────────────────────────

func serversOf(t *testing.T, data []byte) map[string]any {
	t.Helper()
	var doc map[string]any
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, data)
	}
	servers, _ := doc["mcpServers"].(map[string]any)
	return servers
}

func TestJSONSetServerPreservesOthersAndIsIdempotent(t *testing.T) {
	existing := []byte(`{"mcpServers":{"other":{"command":"x"}},"someKey":42}`)
	entry := map[string]any{"command": "/bin/aitriage", "args": safeProfileArgs("/proj")}

	once, err := jsonSetServer(existing, "aitriage", entry)
	if err != nil {
		t.Fatal(err)
	}
	servers := serversOf(t, once)
	if servers["other"] == nil {
		t.Error("existing server 'other' was dropped")
	}
	if servers["aitriage"] == nil {
		t.Error("aitriage server not added")
	}

	var doc map[string]any
	_ = json.Unmarshal(once, &doc)
	if doc["someKey"] == nil {
		t.Error("top-level key 'someKey' was dropped")
	}

	twice, err := jsonSetServer(once, "aitriage", entry)
	if err != nil {
		t.Fatal(err)
	}
	if string(once) != string(twice) {
		t.Fatalf("json install not idempotent:\n%s\nvs\n%s", once, twice)
	}
}

func TestJSONSetServerOnEmpty(t *testing.T) {
	out, err := jsonSetServer(nil, "aitriage", map[string]any{"command": "/bin/aitriage"})
	if err != nil {
		t.Fatal(err)
	}
	if serversOf(t, out)["aitriage"] == nil {
		t.Errorf("empty-input install did not create server:\n%s", out)
	}
}

func TestJSONRemoveServer(t *testing.T) {
	existing := []byte(`{"mcpServers":{"aitriage":{"command":"x"},"other":{"command":"y"}}}`)
	out, changed, err := jsonRemoveServer(existing, "aitriage")
	if err != nil {
		t.Fatal(err)
	}
	if !changed {
		t.Error("expected changed=true when server present")
	}
	servers := serversOf(t, out)
	if servers["aitriage"] != nil {
		t.Error("aitriage not removed")
	}
	if servers["other"] == nil {
		t.Error("unrelated server dropped")
	}

	// Removing an absent server reports no change.
	_, changed, err = jsonRemoveServer(out, "aitriage")
	if err != nil {
		t.Fatal(err)
	}
	if changed {
		t.Error("expected changed=false when server absent")
	}
}

func TestJSONInvalidInputErrors(t *testing.T) {
	if _, err := jsonSetServer([]byte("not json"), "aitriage", nil); err == nil {
		t.Error("expected error on invalid JSON")
	}
}
