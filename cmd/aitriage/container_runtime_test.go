package main

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestRunManagedContainerPreservesChildExitCode(t *testing.T) {
	binDir := t.TempDir()
	docker := filepath.Join(binDir, "docker")
	if err := os.WriteFile(docker, []byte("#!/bin/sh\nexit 42\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	err := runManagedContainer(context.Background(), []string{"run", "image"}, "")
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("error = %T %v, want *exec.ExitError", err, err)
	}
	if exitErr.ExitCode() != 42 {
		t.Fatalf("exit code = %d, want 42", exitErr.ExitCode())
	}
}
