package main

import (
	"strings"
	"testing"
)

// TestContainerServeArgs proves the MCP container launch is built correctly:
// stdio (no TTY), read-only source, read-write reports, /workspace scan-root,
// and an inner native server (the scanner bundle lives in the image).
func TestContainerServeArgs(t *testing.T) {
	t.Setenv("AITRIAGE_IMAGE", "ghcr.io/cybertortuga/aitriage:v1.7.0")
	args := containerServeArgs("ghcr.io/cybertortuga/aitriage:v1.7.0", "/host/repo", "safe", "/host/scanner-cache", "aitriage-mcp-test")
	joined := strings.Join(args, " ")

	for _, want := range []string{
		"run --rm",
		"--name aitriage-mcp-test",
		"-i --security-opt",
		"--security-opt no-new-privileges",
		"/host/repo:/workspace:ro",
		"/host/repo/aitriage-reports:/workspace/aitriage-reports:rw",
		"/host/scanner-cache:/home/aitriage/.cache:rw",
		"AITRIAGE_CACHE_DIR=/workspace/aitriage-reports/cache",
		"ghcr.io/cybertortuga/aitriage:v1.7.0 serve --runtime native --profile safe --scan-root /workspace",
	} {
		if !strings.Contains(joined, want) {
			t.Errorf("container serve args missing %q:\n%s", want, joined)
		}
	}
	if strings.Contains(joined, " -t") {
		t.Error("MCP container serve must not allocate a TTY")
	}
}
