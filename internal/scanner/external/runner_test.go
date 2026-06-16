package external_test

import (
	"context"
	"github.com/cybertortuga/aitriage/internal/scanner/external"
	"testing"
)

func TestIsInstalled_Go(t *testing.T) {
	// Go всегда установлен в тест-окружении
	if !external.IsInstalled("go") {
		t.Error("Expected go to be installed")
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
