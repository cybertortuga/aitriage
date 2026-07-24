package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	rt "github.com/dodobrands/aitriage/internal/runtime"
)

func TestE2E_DefaultWebUsesFullContainerBundle(t *testing.T) {
	if testing.Short() || !dockerImageAvailable(t) {
		t.Skip("requires Docker and the resolved AITriage runtime image")
	}
	bin := buildBinary(t)
	project := t.TempDir()
	writeFile(t, filepath.Join(project, "main.py"), "from fastapi import FastAPI\napp = FastAPI()\n@app.get('/ping')\ndef ping(): return 'pong'\n")
	writeFile(t, filepath.Join(project, "requirements.txt"), "fastapi\n")

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	_ = listener.Close()

	cmd := exec.Command(bin, "web", "--project", project, "--port", fmt.Sprint(port))
	cmd.Env = append(os.Environ(), "AITRIAGE_IMAGE="+rt.ResolveImage(Version))
	var logs bytes.Buffer
	cmd.Stdout, cmd.Stderr = &logs, &logs
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	finished := false
	t.Cleanup(func() {
		if !finished {
			_ = exec.Command("docker", "rm", "-f", fmt.Sprintf("aitriage-web-%d", port)).Run()
			_ = cmd.Process.Kill()
		}
	})

	baseURL := fmt.Sprintf("http://127.0.0.1:%d", port)
	client := &http.Client{Timeout: 4 * time.Minute}
	deadline := time.Now().Add(45 * time.Second)
	for {
		resp, requestErr := client.Get(baseURL + "/api/health")
		if requestErr == nil && resp.StatusCode == http.StatusOK {
			_ = resp.Body.Close()
			break
		}
		if resp != nil {
			_ = resp.Body.Close()
		}
		if time.Now().After(deadline) {
			t.Fatalf("web did not become healthy: %v\n%s", requestErr, logs.String())
		}
		time.Sleep(500 * time.Millisecond)
	}

	resp, err := client.Post(baseURL+"/api/scan", "application/json", bytes.NewBufferString(`{"path":"."}`))
	if err != nil {
		t.Fatalf("full Web scan: %v\n%s", err, logs.String())
	}
	defer func() { _ = resp.Body.Close() }()
	var result struct {
		OK              bool   `json:"ok"`
		ScannerCoverage string `json:"scanner_coverage"`
		Scanners        []struct {
			Scanner string `json:"scanner"`
			Status  string `json:"status"`
		} `json:"scanners"`
		ManifestPath string `json:"manifest_path"`
		Error        string `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK || !result.OK || result.ScannerCoverage != "full" {
		t.Fatalf("web full scan status=%d result=%+v\n%s", resp.StatusCode, result, logs.String())
	}
	statuses := map[string]string{}
	for _, scanner := range result.Scanners {
		statuses[scanner.Scanner] = scanner.Status
	}
	for _, scanner := range []string{"aitriage", "semgrep", "trivy_fs", "trivy_config", "gitleaks", "bandit"} {
		if statuses[scanner] != "completed" {
			t.Errorf("Web scanner %s = %q, want completed (all=%v)", scanner, statuses[scanner], statuses)
		}
	}
	if result.ManifestPath == "" {
		t.Fatal("Web full scan did not expose its scanner manifest path")
	}
	manifestPath := filepath.Join(project, "aitriage-reports", filepath.FromSlash(strings.TrimPrefix(result.ManifestPath, "aitriage-reports/")))
	if info, statErr := os.Stat(manifestPath); statErr != nil || info.Mode().Perm() != 0o600 {
		t.Fatalf("Web manifest missing or not private at %s: info=%v err=%v", manifestPath, info, statErr)
	}

	// The host launcher owns lifecycle: Ctrl-C must stop and remove the named
	// container without an extra `docker rm` from the user.
	if err := cmd.Process.Signal(os.Interrupt); err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Web host launcher returned an error after Ctrl-C: %v\n%s", err, logs.String())
		}
	case <-time.After(20 * time.Second):
		t.Fatal("Web host launcher did not exit after Ctrl-C")
	}
	finished = true
	name := fmt.Sprintf("aitriage-web-%d", port)
	if err := exec.Command("docker", "container", "inspect", name).Run(); err == nil {
		_ = exec.Command("docker", "rm", "-f", name).Run()
		t.Fatalf("managed Web container %s was orphaned after Ctrl-C", name)
	}
}
