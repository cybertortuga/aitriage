package pipeline_test

import (
	"strings"
	"testing"

	"github.com/dodobrands/aitriage/internal/agent/llm"
	"github.com/dodobrands/aitriage/internal/agent/pipeline"
	"github.com/dodobrands/aitriage/internal/scanner/external"
)

// TestFastXSS_ReachesSecureCoderInput proves the canonical FAST-XSS finding
// travels the ordinary SecureCoder flow: external finding → pipeline.BuildState
// → graph.AgentState.ExternalFindings, and renders into the analysis input the
// host agent hands to SecureCoder. The real Semgrep boundary is covered by the
// scanner and container E2E tests.
func TestFastXSS_ReachesSecureCoderInput(t *testing.T) {
	const fixtures = "testdata/fastapi-xss"
	rich := llm.RichScanResult{
		ProjectPath: fixtures,
		External: []external.UnifiedFinding{{
			Source: "semgrep", RuleID: "FAST-XSS", Severity: "HIGH",
			Message: "Reflected XSS", File: "app.py", Line: 8,
		}},
	}

	// 1. It flows into the shared graph state the host-agent runner consumes.
	state := pipeline.BuildState(pipeline.Options{ProjectPath: fixtures}, rich)
	var inState bool
	for _, f := range state.ExternalFindings {
		if f.RuleID == "FAST-XSS" {
			inState = true
		}
	}
	if !inState {
		t.Error("FAST-XSS did not reach graph.AgentState.ExternalFindings")
	}

	// 2. It renders into the analysis input handed to SecureCoder / the triage LLM.
	prompt := llm.BuildAnalysisPrompt(rich)
	if !strings.Contains(prompt, "FAST-XSS") {
		t.Error("FAST-XSS is not present in the SecureCoder analysis input")
	}
}
