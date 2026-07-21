package mcp

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// connectClient wires an in-memory MCP client to a safe/full server and
// returns a live client session.
func connectClient(t *testing.T, cfg Config) *mcp.ClientSession {
	t.Helper()
	s, err := NewServerWithConfig("test", cfg)
	if err != nil {
		t.Fatalf("NewServerWithConfig: %v", err)
	}
	ctx := context.Background()
	clientT, serverT := mcp.NewInMemoryTransports()

	ss, err := s.srv.Connect(ctx, serverT, nil)
	if err != nil {
		t.Fatalf("server connect: %v", err)
	}
	t.Cleanup(func() { _ = ss.Close() })

	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "0"}, nil)
	cs, err := client.Connect(ctx, clientT, nil)
	if err != nil {
		t.Fatalf("client connect: %v", err)
	}
	t.Cleanup(func() { _ = cs.Close() })
	return cs
}

func listToolNames(t *testing.T, cs *mcp.ClientSession) map[string]bool {
	t.Helper()
	res, err := cs.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListTools: %v", err)
	}
	names := map[string]bool{}
	for _, tool := range res.Tools {
		names[tool.Name] = true
	}
	return names
}

// TestSafeProfileHidesMutatingTool proves the safe profile exposes the
// read-only scan tools but never registers the mutating securecoder_ignore,
// while the full profile does.
func TestSafeProfileHidesMutatingTool(t *testing.T) {
	root := t.TempDir()

	safe := listToolNames(t, connectClient(t, Config{Profile: ProfileSafe, ScanRoot: root}))
	if safe["securecoder_ignore"] {
		t.Error("safe profile must NOT expose securecoder_ignore")
	}
	for _, want := range []string{"aitriage_scan", "aitriage_secrets", "run_securecoder", "run_securecoder_deps"} {
		if !safe[want] {
			t.Errorf("safe profile is missing read-only tool %q", want)
		}
	}

	full := listToolNames(t, connectClient(t, Config{Profile: ProfileFull}))
	if !full["securecoder_ignore"] {
		t.Error("full profile must still expose securecoder_ignore")
	}
}

// TestSafeScanConfinement proves that, end-to-end over MCP, a scan inside
// the root succeeds while a path escaping the root is rejected as a tool error.
func TestSafeScanConfinement(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "app.js"), []byte("const x = 1;\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	cs := connectClient(t, Config{Profile: ProfileSafe, ScanRoot: root})
	ctx := context.Background()

	// A scan of the project root (".") succeeds.
	ok, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "aitriage_scan",
		Arguments: map[string]any{"path": "."},
	})
	if err != nil {
		t.Fatalf("in-root scan transport error: %v", err)
	}
	if ok.IsError {
		t.Fatalf("in-root scan should succeed, got tool error: %+v", ok.Content)
	}

	// A scan that tries to climb out of the root is rejected.
	escape, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "aitriage_scan",
		Arguments: map[string]any{"path": "../"},
	})
	if err != nil {
		t.Fatalf("escape scan transport error: %v", err)
	}
	if !escape.IsError {
		t.Error("scan of '../' must be rejected by the scan-root guard")
	}
}
