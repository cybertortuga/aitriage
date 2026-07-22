package main

import (
	"context"
	"errors"
	"os/exec"
	"runtime"
	"testing"
)

func TestRunManagedContainerPreservesChildExitCode(t *testing.T) {
	previous := dockerExecutable
	t.Cleanup(func() { dockerExecutable = previous })

	var command string
	var args []string
	if runtime.GOOS == "windows" {
		var err error
		command, err = exec.LookPath("cmd.exe")
		if err != nil {
			t.Fatal(err)
		}
		args = []string{"/d", "/c", "exit 42"}
	} else {
		command = "/bin/sh"
		args = []string{"-c", "exit 42"}
	}
	dockerExecutable = func() string { return command }

	err := runManagedContainer(context.Background(), args, "")
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("error = %T %v, want *exec.ExitError", err, err)
	}
	if exitErr.ExitCode() != 42 {
		t.Fatalf("exit code = %d, want 42", exitErr.ExitCode())
	}
}
