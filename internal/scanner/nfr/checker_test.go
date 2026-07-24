package nfr_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/dodobrands/aitriage/internal/scanner/nfr"
)

func TestCheckNFR_FindsMissingDotEnvExampleOnlyWhenEnvironmentIsUsed(t *testing.T) {
	tmpDir := t.TempDir()
	_ = os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte("package main\nimport \"os\"\nvar token = os.Getenv(\"TOKEN\")\n"), 0644)

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

func TestCheckNFR_DoesNotInventEnvironmentContract(t *testing.T) {
	tmpDir := t.TempDir()
	_ = os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte("package main\nfunc main() {}\n"), 0o644)

	findings, err := nfr.CheckNFR(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	for _, finding := range findings {
		if finding.RuleID == "NFR-ENV-001" {
			t.Fatalf(".env.example finding emitted without env consumption: %+v", finding)
		}
	}
}

func TestCheckNFR_RecursesIntoDeepSourceDirectories(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "src", "deep", "main.py")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	_ = os.WriteFile(path, []byte("from fastapi import FastAPI\napp = FastAPI()\n@app.get('/x')\ndef x(): return 1\n"), 0o644)

	findings, err := nfr.CheckNFR(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	foundAuth := false
	for _, finding := range findings {
		if finding.RuleID == "NFR-API-003" {
			foundAuth = true
		}
		if finding.RuleID == "NFR-API-002" {
			t.Fatal("removed unconditional CORS rule must not reappear")
		}
	}
	if !foundAuth {
		t.Fatal("deep API source was not evaluated")
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
