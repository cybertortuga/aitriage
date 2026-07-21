package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

var binaryPath string

func TestMain(m *testing.M) {
	buildDir, err := os.MkdirTemp("", "aitriage-test-binary-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not create test binary directory: %v\n", err)
		os.Exit(1)
	}

	binaryName := "aitriage-test-binary"
	if runtime.GOOS == "windows" {
		binaryName += ".exe"
	}
	binaryPath = filepath.Join(buildDir, binaryName)

	// Build into a process-unique directory. A fixed package-local path makes
	// concurrent `go test` and `go test -race` invocations delete each other's
	// executable and creates a false test failure unrelated to product code.
	build := exec.Command("go", "build", "-o", binaryPath, ".")
	if err := build.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Could not build binary for testing: %v\n", err)
		os.Exit(1)
	}

	// Run tests
	exitCode := m.Run()
	if err := os.RemoveAll(buildDir); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to remove test binary directory: %v\n", err)
	}

	os.Exit(exitCode)
}

func TestScanCommand(t *testing.T) {
	// Create a temporary directory with some dummy code
	tempDir := t.TempDir()

	// Write a Go file that should trigger VIBE checks (e.g. TODO without author)
	goCode := []byte("package main\n\n// TODO: fix this\nfunc main() {}\n")
	err := os.WriteFile(filepath.Join(tempDir, "main.go"), goCode, 0644)
	if err != nil {
		t.Fatalf("failed to write dummy code: %v", err)
	}

	// We run `scan <tempDir> --format json`
	// We might get a non-zero exit code if fail-on is triggered, but we mainly care about output parsing.
	cmd := exec.Command(binaryPath, "scan", tempDir, "--format", "json", "--no-summary")

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Logf("Scan command exited with error (expected if findings exist): %v", err)
	}
	// It may return an error if it exits with status 1 due to critical failures or security score threshold
	// We don't fail the test solely on err != nil, we just check the output JSON

	output := out.String()

	if !strings.Contains(output, "project_path") {
		t.Errorf("Expected JSON output containing project_path, got: %s\nStderr: %s", output, stderr.String())
	}
}

func TestWebCommandHelp(t *testing.T) {
	// Just test if `web --help` runs without panicking
	cmd := exec.Command(binaryPath, "web", "--help")

	var out bytes.Buffer
	cmd.Stdout = &out

	err := cmd.Run()
	if err != nil {
		t.Fatalf("Failed to run web --help: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "Start the local AITriage browser dashboard") ||
		!strings.Contains(output, "aitriage-reports") {
		t.Errorf("Expected web help output, got: %s", output)
	}
}
