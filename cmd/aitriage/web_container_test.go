package main

import (
	"strings"
	"testing"
)

// TestWebContainerArgs locks the Web container launch: published port, source
// read-only, reports read-write, and no docker socket / privileged.
func TestWebContainerArgs(t *testing.T) {
	args := webContainerArgs("ghcr.io/cybertortuga/aitriage:v1", "/host/repo", "/host/scanner-cache", 8080)
	joined := strings.Join(args, " ")
	for _, want := range []string{
		"--name aitriage-web-8080",
		"-p 127.0.0.1:8080:8080",
		"/host/repo:/workspace:ro",
		"/host/repo/aitriage-reports:/workspace/aitriage-reports:rw",
		"/host/scanner-cache:/home/aitriage/.cache:rw",
		"web --runtime native --port 8080 --host-prefix /workspace",
		"AITRIAGE_RUNTIME=container",
		"AITRIAGE_REPORTS_DIR=/workspace/aitriage-reports",
		"DB_PATH=/workspace/aitriage-reports/web/aitriage.db",
	} {
		if !strings.Contains(joined, want) {
			t.Errorf("web container args missing %q:\n%s", want, joined)
		}
	}
	if strings.Contains(joined, "docker.sock") || strings.Contains(joined, "--privileged") {
		t.Error("must never mount the docker socket or use --privileged")
	}
}

func TestWebDefaultsToContainerRuntime(t *testing.T) {
	flag := webCmd.Flags().Lookup("runtime")
	if flag == nil || flag.DefValue != "container" {
		t.Fatalf("web --runtime default = %v, want container", flag)
	}
}
