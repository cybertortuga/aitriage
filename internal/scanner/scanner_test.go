package scanner_test

import (
	"context"
	"github.com/cybertortuga/aitriage/internal/scanner"
	"os"
	"testing"
)

func TestScanReturnsReport(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a dummy file that might trigger a finding, e.g., a hardcoded secret pattern
	err := os.WriteFile(tmpDir+"/main.go", []byte("package main\n\nfunc main() {\n\tsecret := \"AKIAIOSFODNN7EXAMPLE\"\n}\n"), 0644)
	if err != nil {
		t.Fatalf("Failed to create dummy file: %v", err)
	}

	report, err := scanner.Scan(context.Background(), tmpDir, scanner.ScanOptions{})
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}
	if report.TotalFiles == 0 {
		t.Error("Expected TotalFiles > 0")
	}
	if len(report.Results) == 0 {
		t.Error("Expected findings in the dummy project")
	}
}
