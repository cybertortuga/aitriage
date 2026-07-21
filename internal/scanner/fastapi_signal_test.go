package scanner

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestFastAPIDoesNotInventMissingArchitectureComponents(t *testing.T) {
	root := t.TempDir()
	source := `from fastapi import FastAPI

app = FastAPI()

@app.get("/health")
def health():
    return {"ok": True}
`
	if err := os.WriteFile(filepath.Join(root, "main.py"), []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	report, err := Scan(context.Background(), root, ScanOptions{ForceStack: "fastapi"})
	if err != nil {
		t.Fatal(err)
	}
	invalidAbsenceRules := map[string]bool{
		"FAST-VAL": true, "FAST-ERROR": true, "FAST-HEADERS": true,
		"FAST-CORS": true, "FAST-SECRETS": true, "FAST-LOGGING": true,
	}
	for _, finding := range report.Results {
		if invalidAbsenceRules[finding.ID] {
			t.Fatalf("invented architecture finding %s: %+v", finding.ID, finding)
		}
	}
}

func TestFastAPIStillDetectsConcreteWildcardCORS(t *testing.T) {
	root := t.TempDir()
	source := `from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware
app = FastAPI()
app.add_middleware(CORSMiddleware, allow_origins=["*"])
`
	if err := os.WriteFile(filepath.Join(root, "main.py"), []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	report, err := Scan(context.Background(), root, ScanOptions{ForceStack: "fastapi"})
	if err != nil {
		t.Fatal(err)
	}
	for _, finding := range report.Results {
		if finding.ID == "FAST-CORS-WILDCARD" {
			return
		}
	}
	t.Fatal("concrete wildcard CORS vulnerability was not detected")
}
