package external_test

import (
	"context"
	"errors"
	"runtime"
	"testing"
	"time"

	"github.com/cybertortuga/aitriage/internal/scanner/external"
)

func TestIsInstalled_Go(t *testing.T) {
	// Go is always installed in the test environment
	if !external.IsInstalled("go") {
		t.Error("Expected go to be installed")
	}
}

func TestRunTool_CancellationFailsClosed(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("uses the POSIX sleep command")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Millisecond)
	defer cancel()
	result, err := external.RunTool(ctx, "sh", "-c", "sleep 10")
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("cancelled scanner error = %v, want deadline exceeded", err)
	}
	if result.ExitCode != -1 {
		t.Fatalf("cancelled scanner exit code = %d, want -1", result.ExitCode)
	}
}

func TestRunTool_Echo(t *testing.T) {
	result, err := external.RunTool(context.Background(), "echo", "hello")
	if err != nil {
		t.Fatalf("RunTool failed: %v", err)
	}
	if result.ExitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", result.ExitCode)
	}
}
