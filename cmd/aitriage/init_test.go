package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestInitWritesModernCursorRuleNotLegacy proves `aitriage init` generates a
// modern Cursor project rule (.cursor/rules/aitriage.mdc) with rule frontmatter
// and NOT the deprecated .cursorrules, and that the rule carries no embedded MCP
// server block (MCP lives in .cursor/mcp.json, written by install-cursor).
func TestInitWritesModernCursorRuleNotLegacy(t *testing.T) {
	prevForce, prevMCP := initForce, initMCP
	initForce, initMCP = false, false
	t.Cleanup(func() { initForce, initMCP = prevForce, prevMCP })

	root := t.TempDir()
	if err := runInit(nil, []string{root}); err != nil {
		t.Fatalf("init: %v", err)
	}

	if _, err := os.Stat(filepath.Join(root, ".cursorrules")); !os.IsNotExist(err) {
		t.Errorf("init must not write the deprecated .cursorrules (err=%v)", err)
	}

	mdcPath := filepath.Join(root, ".cursor", "rules", "aitriage.mdc")
	data, err := os.ReadFile(mdcPath)
	if err != nil {
		t.Fatalf("modern Cursor rule not written: %v", err)
	}
	rule := string(data)
	if !strings.HasPrefix(rule, "---\n") || !strings.Contains(rule, "alwaysApply: true") {
		t.Errorf("Cursor rule missing .mdc frontmatter:\n%s", rule)
	}
	if !strings.Contains(rule, agentContractHeading) {
		t.Errorf("Cursor rule missing the shared security contract:\n%s", rule)
	}
	// The rule must NOT embed an MCP server config, and must not carry the stale
	// `serve sse` invocation that the old .cursorrules shipped.
	for _, banned := range []string{`"mcp"`, "mcpServers", `"sse"`, "serve\", \"sse"} {
		if strings.Contains(rule, banned) {
			t.Errorf("Cursor rule must not embed MCP config (%q found):\n%s", banned, rule)
		}
	}
}
