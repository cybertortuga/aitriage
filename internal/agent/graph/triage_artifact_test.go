package graph

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestBuildTriageFindingsArtifactPreservesEveryFindingAndDisposition(t *testing.T) {
	state := &AgentState{
		EnrichedFindings: []EnrichedFinding{
			{ID: "secret", VulnID: "CS-SECRETS-001", Type: "core", Severity: "CRITICAL", File: "config.py", Line: 10, Message: "Hardcoded secret", Snippet: "token = 'x'"},
			{ID: "random", VulnID: "CS-CRYPTO-002", Type: "core", Severity: "LOW", File: "snow.ts", Line: 3, Message: "Visual randomness"},
			{ID: "cors", VulnID: "CS-CONFIG-003", Type: "core", Severity: "MEDIUM", File: "main.py", Line: 21, Message: "Wildcard CORS"},
		},
		// Deliberately out of original order: the exported list must still retain
		// the original scanner ordering by finding index.
		FindingDispositions: []FindingDisposition{
			{FindingIndex: 2, FindingID: "CS-CONFIG-003", Disposition: "Needs Manual Review", Rationale: "Deployment context required", Confidence: "low", DispositionSource: "llm", Fingerprint: "fp-cors"},
			{FindingIndex: 0, FindingID: "CS-SECRETS-001", Disposition: "True Positive", Rationale: "Secret committed in source", Confidence: "high", DispositionSource: "cache", Fingerprint: "fp-secret"},
			{FindingIndex: 1, FindingID: "CS-CRYPTO-002", Disposition: "False Positive", Rationale: "Only controls animation", Confidence: "high", DispositionSource: "deterministic", Fingerprint: "fp-random"},
		},
		ClassificationAudit: []ClassificationAuditEntry{{
			Attempt:                0,
			UniqueFindingIndices:   []int{0},
			FindingIDs:             []string{"CS-SECRETS-001"},
			Fingerprints:           []string{"fp-secret"},
			RawResponse:            `{"finding_dispositions":[...]}`,
			AcceptedFindingIndices: []int{0},
		}},
	}

	artifact, err := BuildTriageFindingsArtifact(state)
	if err != nil {
		t.Fatalf("BuildTriageFindingsArtifact() error = %v", err)
	}
	if artifact.SchemaVersion != TriageArtifactSchemaVersion {
		t.Fatalf("schema_version = %d, want %d", artifact.SchemaVersion, TriageArtifactSchemaVersion)
	}
	if artifact.TriageStatus != "complete" {
		t.Fatalf("triage_status = %q, want complete", artifact.TriageStatus)
	}
	if artifact.TotalFindings != 3 || len(artifact.Findings) != 3 {
		t.Fatalf("total/findings = %d/%d, want 3/3", artifact.TotalFindings, len(artifact.Findings))
	}
	if len(artifact.ClassificationAudit) != 1 || artifact.ClassificationAudit[0].RawResponse == "" {
		t.Fatalf("classification audit = %+v, want persisted raw response", artifact.ClassificationAudit)
	}

	for index, want := range []struct {
		vulnID      string
		disposition string
		fingerprint string
	}{
		{vulnID: "CS-SECRETS-001", disposition: "True Positive", fingerprint: "fp-secret"},
		{vulnID: "CS-CRYPTO-002", disposition: "False Positive", fingerprint: "fp-random"},
		{vulnID: "CS-CONFIG-003", disposition: "Needs Manual Review", fingerprint: "fp-cors"},
	} {
		got := artifact.Findings[index]
		if got.FindingIndex != index || got.Finding.VulnID != want.vulnID || got.Disposition.Disposition != want.disposition || got.Disposition.Fingerprint != want.fingerprint {
			t.Fatalf("finding %d = %+v, want id=%s disposition=%s fingerprint=%s", index, got, want.vulnID, want.disposition, want.fingerprint)
		}
	}

	data, err := json.Marshal(artifact)
	if err != nil {
		t.Fatalf("marshal artifact: %v", err)
	}
	jsonText := string(data)
	for _, want := range []string{"\"schema_version\":1", "\"triage_status\":\"complete\"", "\"code_snippet\":\"token = 'x'\"", "\"rationale\":\"Only controls animation\"", "\"disposition_source\":\"deterministic\"", "\"classification_audit\"", "\"raw_response\""} {
		if !strings.Contains(jsonText, want) {
			t.Errorf("serialized artifact does not contain %s: %s", want, jsonText)
		}
	}
}

func TestBuildTriageFindingsArtifactRejectsIncompleteOrAmbiguousDispositions(t *testing.T) {
	findings := []EnrichedFinding{{VulnID: "CS-1"}, {VulnID: "CS-2"}}
	tests := []struct {
		name         string
		dispositions []FindingDisposition
		want         string
	}{
		{
			name:         "missing disposition",
			dispositions: []FindingDisposition{{FindingIndex: 0, Disposition: "True Positive"}},
			want:         "classified 1 of 2",
		},
		{
			name: "duplicate disposition",
			dispositions: []FindingDisposition{
				{FindingIndex: 0, Disposition: "True Positive"},
				{FindingIndex: 0, Disposition: "False Positive"},
			},
			want: "more than once",
		},
		{
			name: "out of range disposition",
			dispositions: []FindingDisposition{
				{FindingIndex: 0, Disposition: "True Positive"},
				{FindingIndex: 2, Disposition: "False Positive"},
			},
			want: "out-of-range",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := BuildTriageFindingsArtifact(&AgentState{
				EnrichedFindings:    findings,
				FindingDispositions: tt.dispositions,
			})
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("BuildTriageFindingsArtifact() error = %v, want substring %q", err, tt.want)
			}
		})
	}
}

func TestBuildTriageFindingsArtifactAllowsCompletedEmptyScan(t *testing.T) {
	artifact, err := BuildTriageFindingsArtifact(&AgentState{})
	if err != nil {
		t.Fatalf("BuildTriageFindingsArtifact() error = %v", err)
	}
	if artifact.TriageStatus != "complete" || artifact.TotalFindings != 0 || len(artifact.Findings) != 0 {
		t.Fatalf("empty artifact = %+v, want completed empty inventory", artifact)
	}
}

func TestBuildTriageFindingsArtifactRejectsDispositionForEmptyScan(t *testing.T) {
	_, err := BuildTriageFindingsArtifact(&AgentState{
		FindingDispositions: []FindingDisposition{{FindingIndex: 0, Disposition: "True Positive"}},
	})
	if err == nil || !strings.Contains(err.Error(), "classified 1 of 0") {
		t.Fatalf("BuildTriageFindingsArtifact() error = %v, want empty-scan disposition rejection", err)
	}
}
