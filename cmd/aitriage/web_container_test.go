package main

import (
	"path/filepath"
	"strings"
	"testing"
)

// TestWebContainerArgs locks the Web container launch: published port, source
// read-only, reports read-write, and no docker socket / privileged.
func TestWebContainerArgs(t *testing.T) {
	hostRoot := t.TempDir()
	cache := t.TempDir()
	args := webContainerArgs("ghcr.io/dodobrands/aitriage:v1", hostRoot, cache, 8080)
	joined := strings.Join(args, " ")
	for _, want := range []string{
		"--name aitriage-web-8080",
		"-p 127.0.0.1:8080:8080",
		hostRoot + ":/workspace:ro",
		filepath.Join(hostRoot, "aitriage-reports") + ":/workspace/aitriage-reports:rw",
		cache + ":/home/aitriage/.cache:rw",
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
