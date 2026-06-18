package graph

import (
	"strings"
	"testing"
)

func TestGenerateSummaryExcludesFalsePositives(t *testing.T) {
	state := &AgentState{
		EnrichedFindings: []EnrichedFinding{
			{ID: "hardcoded-secret", VulnID: "CS-SECRETS-001", Type: "core", Severity: "CRITICAL", File: "config.py", Line: 10, Message: "Hardcoded password found"},
			{ID: "math-random", VulnID: "CS-CRYPTO-001", Type: "core", Severity: "HIGH", File: "snowflake.ts", Line: 5, Message: "Math.random() used for visual effects"},
			{ID: "insecure-proxy", VulnID: "CS-CONFIG-001", Type: "core", Severity: "MEDIUM", File: "vite.config.ts", Line: 20, Message: "Proxy secure=false for docker"},
		},
		FindingDispositions: []FindingDisposition{
			{FindingIndex: 0, FindingID: "CS-SECRETS-001", Disposition: "True Positive", Rationale: "Real password in config"},
			{FindingIndex: 1, FindingID: "CS-CRYPTO-001", Disposition: "False Positive", Rationale: "Math.random used for visual snowflakes, not security"},
			{FindingIndex: 2, FindingID: "CS-CONFIG-001", Disposition: "Needs Manual Review", Rationale: "Docker local dev only, but verify"},
		},
	}

	generateSummary(state)

	summary := state.SummaryMarkdown
	if summary == "" {
		t.Fatal("SummaryMarkdown is empty")
	}

	// Must contain TP finding
	if !strings.Contains(summary, "CS-SECRETS-001") {
		t.Error("Summary should contain True Positive finding CS-SECRETS-001")
	}

	// Must contain NR finding
	if !strings.Contains(summary, "CS-CONFIG-001") {
		t.Error("Summary should contain Needs Manual Review finding CS-CONFIG-001")
	}

	// Must NOT contain FP finding in table
	if strings.Contains(summary, "CS-CRYPTO-001") {
		t.Error("Summary should NOT contain False Positive finding CS-CRYPTO-001")
	}

	// Must mention suppressed count in footer
	if !strings.Contains(summary, "1 false positive(s) suppressed") {
		t.Error("Summary footer should mention 1 suppressed false positive")
	}

	// Must contain stats table
	if !strings.Contains(summary, "True Positives | 1") {
		t.Error("Summary should show 1 True Positive in stats")
	}
	if !strings.Contains(summary, "False Positives (suppressed) | 1") {
		t.Error("Summary should show 1 suppressed FP in stats")
	}
}

func TestGenerateSummaryNoFindings(t *testing.T) {
	state := &AgentState{
		EnrichedFindings:    nil,
		FindingDispositions: nil,
	}

	generateSummary(state)

	summary := state.SummaryMarkdown
	if summary == "" {
		t.Fatal("SummaryMarkdown is empty")
	}
	if !strings.Contains(summary, "No actionable security findings") {
		t.Error("Summary should show 'No actionable security findings' when there are no findings")
	}
}

func TestGenerateSummaryAllFalsePositives(t *testing.T) {
	state := &AgentState{
		EnrichedFindings: []EnrichedFinding{
			{ID: "r1", VulnID: "CS-XSS-001", Type: "core", Severity: "HIGH", File: "app.ts", Line: 1, Message: "XSS in safe context"},
			{ID: "r2", VulnID: "CS-XSS-002", Type: "core", Severity: "MEDIUM", File: "app.ts", Line: 5, Message: "Another safe XSS"},
		},
		FindingDispositions: []FindingDisposition{
			{FindingIndex: 0, FindingID: "CS-XSS-001", Disposition: "False Positive", Rationale: "React auto-escapes"},
			{FindingIndex: 1, FindingID: "CS-XSS-002", Disposition: "False Positive", Rationale: "Template is safe"},
		},
	}

	generateSummary(state)

	summary := state.SummaryMarkdown
	if !strings.Contains(summary, "No actionable security findings") {
		t.Error("Summary should show 'No actionable security findings' when all are FP")
	}
	if !strings.Contains(summary, "2 false positive(s) suppressed") {
		t.Error("Summary footer should mention 2 suppressed false positives")
	}
}

func TestGenerateSummaryUndisposedFindingsTreatedAsActionable(t *testing.T) {
	// If threat model fails and no dispositions exist, all findings should be actionable
	state := &AgentState{
		EnrichedFindings: []EnrichedFinding{
			{ID: "r1", VulnID: "CS-AUTH-001", Type: "core", Severity: "HIGH", File: "api.go", Line: 42, Message: "Missing auth"},
		},
		FindingDispositions: nil, // No dispositions — threat model failed
	}

	generateSummary(state)

	summary := state.SummaryMarkdown
	if !strings.Contains(summary, "CS-AUTH-001") {
		t.Error("Undisposed finding should appear in summary as actionable")
	}
	if !strings.Contains(summary, "Needs Manual Review") {
		t.Error("Undisposed finding should be labelled 'Needs Manual Review'")
	}
}

func TestGenerateSummaryPipesInMessagesAreEscaped(t *testing.T) {
	state := &AgentState{
		EnrichedFindings: []EnrichedFinding{
			{ID: "r1", VulnID: "CS-MISC-001", Type: "core", Severity: "HIGH", File: "x.go", Line: 1, Message: "Issue with | pipe char"},
		},
		FindingDispositions: []FindingDisposition{
			{FindingIndex: 0, FindingID: "CS-MISC-001", Disposition: "True Positive", Rationale: "Real issue"},
		},
	}

	generateSummary(state)

	// The message "Issue with | pipe char" should become "Issue with \| pipe char"
	// in the markdown table to avoid breaking the table structure.
	if !strings.Contains(state.SummaryMarkdown, `Issue with \| pipe char`) {
		t.Error("Pipe characters in messages should be escaped as \\| for valid markdown tables")
	}
}
