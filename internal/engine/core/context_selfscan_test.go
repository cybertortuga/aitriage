package core

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsAITriageGeneratedArtifact(t *testing.T) {
	dir := t.TempDir()

	// Runway report matched by its stable filename, regardless of content.
	runway := filepath.Join(dir, "runway-report-11-2026-06-19.md")
	if err := os.WriteFile(runway, []byte("# report\n[gitleaks] jwt example\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	info, _ := os.Stat(runway)
	if !isAITriageGeneratedArtifact(runway, info.Size()) {
		t.Fatal("runway-report-*.md must be recognized as an AITriage artifact")
	}

	// Arbitrary markdown carrying the content marker is also recognized.
	marked := filepath.Join(dir, "audit.md")
	if err := os.WriteFile(marked, []byte(AITriageArtifactMarker+"\n# Security Audit\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	mInfo, _ := os.Stat(marked)
	if !isAITriageGeneratedArtifact(marked, mInfo.Size()) {
		t.Fatal("markdown with the artifact marker must be recognized")
	}

	// A normal source/markdown file is not an artifact.
	normal := filepath.Join(dir, "README.md")
	if err := os.WriteFile(normal, []byte("# My Project\nsome docs\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	nInfo, _ := os.Stat(normal)
	if isAITriageGeneratedArtifact(normal, nInfo.Size()) {
		t.Fatal("a normal README must not be treated as an AITriage artifact")
	}
}

func TestNewWorkspaceSkipsGeneratedReports(t *testing.T) {
	root := t.TempDir()
	// AITriage's canonical output directory must be excluded as a whole even if
	// a generated file has no content marker.
	aiDir := filepath.Join(root, "aitriage-reports", "run-1")
	if err := os.MkdirAll(aiDir, 0o755); err != nil {
		t.Fatal(err)
	}
	report := filepath.Join(aiDir, "runway-report-11-2026-06-19.md")
	if err := os.WriteFile(report, []byte("unmarked generated content\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	src := filepath.Join(root, "main.py")
	if err := os.WriteFile(src, []byte("print('hi')\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	ws, err := NewWorkspace(root)
	if err != nil {
		t.Fatalf("NewWorkspace: %v", err)
	}
	for _, f := range ws.Files {
		if filepath.Base(f.Path) == "runway-report-11-2026-06-19.md" {
			t.Fatalf("generated report must not be indexed for scanning: %s", f.Path)
		}
	}
	// Sanity: the real source file is still indexed.
	found := false
	for _, f := range ws.Files {
		if filepath.Base(f.Path) == "main.py" {
			found = true
		}
	}
	if !found {
		t.Fatal("real source file should still be scanned")
	}
}
