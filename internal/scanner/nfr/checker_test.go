package nfr_test

import (
	"github.com/cybertortuga/aitriage/internal/scanner/nfr"
	"os"
	"path/filepath"
	"testing"
)

func TestCheckNFR_FindsMissingDotEnvExample(t *testing.T) {
	// Создать временный проект без .env.example
	tmpDir := t.TempDir()
	_ = os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte("package main"), 0644)

	findings, err := nfr.CheckNFR(tmpDir)
	if err != nil {
		t.Skipf("Rules dir not found: %v", err)
	}

	found := false
	for _, f := range findings {
		if f.RuleID == "NFR-ENV-001" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected NFR-ENV-001 (.env.example missing) to trigger")
	}
}

func BenchmarkCheckNFR(b *testing.B) {
	tmpDir := b.TempDir()
	_ = os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte("package main"), 0644)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = nfr.CheckNFR(tmpDir)
	}
}

func BenchmarkGetAllRulesAsText(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = nfr.GetAllRulesAsText()
	}
}
