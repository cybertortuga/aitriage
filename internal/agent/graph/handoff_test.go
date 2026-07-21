package graph

import (
	"strings"
	"testing"
	"time"

	"github.com/cybertortuga/aitriage/internal/scanner/external"
)

func TestBuildAgentHandoffKeepsCIAndWebDataInParity(t *testing.T) {
	state := &AgentState{
		ScannerCoverage: "full",
		ScannerExecutions: []external.ScannerExecution{
			{Scanner: "aitriage", Status: external.StatusCompleted, Findings: 1},
			{Scanner: "semgrep", Status: external.StatusCompleted},
		},
		EnrichedFindings: []EnrichedFinding{
			{VulnID: "CS-AUTH-001", Severity: "CRITICAL", File: "auth.go", Line: 10, Message: "Hardcoded secret"},
			{VulnID: "CS-MISC-002", Severity: "MEDIUM", File: "app.go", Line: 20, Message: "Manual review"},
			{VulnID: "CS-MISC-003", Severity: "LOW", File: "visual.ts", Line: 30, Message: "Suppressed visual randomness"},
		},
		FindingDispositions: []FindingDisposition{
			{FindingIndex: 0, FindingID: "CS-AUTH-001", Disposition: "True Positive", Rationale: "Confirmed"},
			{FindingIndex: 1, FindingID: "CS-MISC-002", Disposition: "Needs Manual Review", Rationale: "Owner decision required"},
			{FindingIndex: 2, FindingID: "CS-MISC-003", Disposition: "False Positive", Rationale: "Not security-sensitive"},
		},
	}

	generatedAt := time.Date(2026, 7, 15, 10, 0, 0, 0, time.UTC)
	handoff, err := BuildAgentHandoff(state, generatedAt)
	if err != nil {
		t.Fatalf("BuildAgentHandoff() error = %v", err)
	}
	if handoff.SchemaVersion != AgentHandoffSchemaVersion {
		t.Fatalf("schema version = %d", handoff.SchemaVersion)
	}
	if len(handoff.AgentData.Findings) != 2 {
		t.Fatalf("actionable findings = %d, want 2", len(handoff.AgentData.Findings))
	}
	for _, id := range []string{"CS-AUTH-001", "CS-MISC-002"} {
		if !strings.Contains(handoff.RemediationPromptMarkdown, id) || !strings.Contains(handoff.SummaryMarkdown, id) {
			t.Fatalf("actionable finding %s missing from prompt or summary", id)
		}
	}
	if strings.Contains(handoff.RemediationPromptMarkdown, "CS-MISC-003") || strings.Contains(handoff.SummaryMarkdown, "CS-MISC-003") {
		t.Fatal("false positive leaked into actionable handoff")
	}
	if handoff.AgentData.ScanDate != "2026-07-15" {
		t.Fatalf("scan date = %q", handoff.AgentData.ScanDate)
	}
	if handoff.AgentData.ScannerCoverage != "full" || len(handoff.AgentData.Scanners) != 2 {
		t.Fatalf("scanner evidence missing from structured handoff: %+v", handoff.AgentData)
	}
	if !strings.Contains(handoff.SummaryMarkdown, "**Scanner coverage**: `FULL`") || !strings.Contains(handoff.SummaryMarkdown, "`semgrep`") {
		t.Fatal("summary is missing scanner coverage evidence")
	}
}

func TestBuildAgentHandoffRejectsIncompleteDispositions(t *testing.T) {
	_, err := BuildAgentHandoff(&AgentState{
		EnrichedFindings: []EnrichedFinding{{VulnID: "CS-AUTH-001", Severity: "HIGH"}},
	}, time.Now())
	if err == nil {
		t.Fatal("expected incomplete dispositions error")
	}
}

func TestBuildAgentHandoffHandlesEmptyAndAllFalsePositiveScans(t *testing.T) {
	generatedAt := time.Date(2026, 7, 15, 10, 0, 0, 0, time.UTC)
	tests := []struct {
		name  string
		state *AgentState
	}{
		{name: "empty", state: &AgentState{}},
		{
			name: "all false positive",
			state: &AgentState{
				EnrichedFindings: []EnrichedFinding{{VulnID: "CS-MISC-001", Severity: "LOW", Message: "Non-security visual state"}},
				FindingDispositions: []FindingDisposition{{
					FindingIndex: 0, FindingID: "CS-MISC-001", Disposition: "False Positive", Rationale: "Expected behavior",
				}},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			handoff, err := BuildAgentHandoff(test.state, generatedAt)
			if err != nil {
				t.Fatal(err)
			}
			if len(handoff.AgentData.Findings) != 0 {
				t.Fatalf("actionable findings = %d, want 0", len(handoff.AgentData.Findings))
			}
			if handoff.RemediationPromptMarkdown != "" {
				t.Fatal("empty actionable scan generated a remediation prompt")
			}
			if !strings.Contains(handoff.SummaryMarkdown, "No actionable security findings") {
				t.Fatal("summary is missing the no-actionable state")
			}
			if handoff.AgentData.GateStatus != "FAILED" {
				t.Fatalf("default failed gate was hidden: %q", handoff.AgentData.GateStatus)
			}
		})
	}
}
