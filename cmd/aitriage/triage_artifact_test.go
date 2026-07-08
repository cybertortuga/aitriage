package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cybertortuga/aitriage/internal/agent/graph"
)

func TestWriteTriageArtifactWritesCanonicalJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "triage-findings.json")
	state := &graph.AgentState{
		EnrichedFindings: []graph.EnrichedFinding{{VulnID: "CS-TEST-001", Message: "test finding"}},
		FindingDispositions: []graph.FindingDisposition{{
			FindingIndex: 0,
			FindingID:    "CS-TEST-001",
			Disposition:  "False Positive",
			Rationale:    "Test fixture",
		}},
	}

	if err := writeTriageArtifact(path, state); err != nil {
		t.Fatalf("writeTriageArtifact() error = %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read artifact: %v", err)
	}
	text := string(data)
	for _, want := range []string{"\"schema_version\": 1", "\"triage_status\": \"complete\"", "\"disposition\": \"False Positive\""} {
		if !strings.Contains(text, want) {
			t.Errorf("artifact does not contain %q: %s", want, text)
		}
	}
}
