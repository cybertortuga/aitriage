package main

import (
	"encoding/json"
	"strings"
	"testing"
)

const winBin = `C:\Users\Renée\AppData\Local\Programs\AITriage\bin\aitriage.exe`
const winRoot = `C:\Users\Renée\My Project`

// TestCodexBlockWindowsPathEscaping proves the Codex TOML writer escapes Windows
// backslashes (and preserves them for the reader), embeds the absolute .exe and
// the Windows scan root, and re-applies idempotently while preserving unrelated
// servers verbatim. These are the Windows-path/config concerns from TASK 4/11.
func TestCodexBlockWindowsPathEscaping(t *testing.T) {
	prev := clientRuntime
	clientRuntime = "container"
	t.Cleanup(func() { clientRuntime = prev })

	block := codexServerBlock(mcpServerName, winBin, safeProfileArgs(winRoot))

	// Backslashes must be TOML-escaped (\\), never emitted raw.
	if strings.Contains(block, `bin\aitriage.exe`) {
		t.Errorf("raw backslash leaked into TOML:\n%s", block)
	}
	if !strings.Contains(block, `\\aitriage.exe"`) {
		t.Errorf("expected escaped backslashes for the binary path:\n%s", block)
	}
	if !strings.Contains(block, `"--scan-root"`) || !strings.Contains(block, `My Project`) {
		t.Errorf("scan root not embedded:\n%s", block)
	}

	// Preserve a foreign server and be idempotent.
	existing := "[mcp_servers.other]\ncommand = \"other\"\n"
	once := tomlSetServer(existing, mcpServerName, block)
	twice := tomlSetServer(once, mcpServerName, block)
	if once != twice {
		t.Errorf("re-applying the Codex block was not idempotent:\n--- once ---\n%s\n--- twice ---\n%s", once, twice)
	}
	if !strings.Contains(twice, "[mcp_servers.other]") || !strings.Contains(twice, `command = "other"`) {
		t.Errorf("foreign server section was not preserved:\n%s", twice)
	}
	if strings.Count(twice, "[mcp_servers.aitriage]") != 1 {
		t.Errorf("aitriage section must appear exactly once:\n%s", twice)
	}
}

// TestClaudeMcpJSONWindowsPath proves the Claude .mcp.json writer stores the
// Windows binary path (JSON-escaped), preserves other servers, and round-trips
// back to the exact host path a reader would use. (TASK 5/11.)
func TestClaudeMcpJSONWindowsPath(t *testing.T) {
	prev := clientRuntime
	clientRuntime = "container"
	t.Cleanup(func() { clientRuntime = prev })

	existing := []byte(`{"mcpServers":{"other":{"command":"other"}}}`)
	entry := map[string]any{"command": winBin, "args": safeProfileArgs(winRoot), "env": map[string]any{}}
	out, err := jsonSetServer(existing, mcpServerName, entry)
	if err != nil {
		t.Fatal(err)
	}

	// Raw JSON must contain escaped backslashes, never a single raw backslash run.
	if !strings.Contains(string(out), `\\aitriage.exe`) {
		t.Errorf("expected JSON-escaped backslashes in the command path:\n%s", out)
	}

	var doc struct {
		McpServers map[string]struct {
			Command string   `json:"command"`
			Args    []string `json:"args"`
		} `json:"mcpServers"`
	}
	if err := json.Unmarshal(out, &doc); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	if _, ok := doc.McpServers["other"]; !ok {
		t.Error("existing 'other' server was dropped")
	}
	got := doc.McpServers[mcpServerName]
	if got.Command != winBin {
		t.Errorf("command round-trip = %q, want %q", got.Command, winBin)
	}
	if len(got.Args) == 0 || got.Args[len(got.Args)-1] != winRoot {
		t.Errorf("scan-root arg round-trip = %v, want last element %q", got.Args, winRoot)
	}
}
