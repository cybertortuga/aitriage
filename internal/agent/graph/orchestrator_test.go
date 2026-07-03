package graph

import (
	"reflect"
	"testing"

	"github.com/cybertortuga/aitriage/internal/engine/core"
)

func TestEnrichFindingsCanonicalizesOrderBeforeVulnIDs(t *testing.T) {
	findings := []core.CheckResult{
		{ID: "FAST-RATELIMIT", Name: "Missing Rate Limiting", Evidence: "Required pattern not found", Severity: "HIGH"},
		{ID: "ENTR-INSECURE-RANDOM", Name: "Insecure Randomness", Evidence: "Math.random", Severity: "HIGH", File: "frontend/src/Confetti.tsx", Line: 21},
		{ID: "FAST-CORS-WILDCARD", Name: "CORS Allow All Origins", Evidence: "allow_origins=['*']", Severity: "HIGH", File: "backend/main.py", Line: 24},
		{ID: "FAST-LOGGING", Name: "Missing Security Logging", Evidence: "Required pattern not found", Severity: "MEDIUM"},
	}

	stateA := &AgentState{CoreFindings: findings}
	stateB := &AgentState{CoreFindings: []core.CheckResult{findings[1], findings[3], findings[2], findings[0]}}
	enrichFindings(stateA)
	enrichFindings(stateB)

	gotA := enrichedOrderSignature(stateA.EnrichedFindings)
	gotB := enrichedOrderSignature(stateB.EnrichedFindings)
	if !reflect.DeepEqual(gotA, gotB) {
		t.Fatalf("canonical enriched order differs:\nA=%v\nB=%v", gotA, gotB)
	}
}

func enrichedOrderSignature(findings []EnrichedFinding) []string {
	out := make([]string, 0, len(findings))
	for _, f := range findings {
		out = append(out, f.VulnID+"|"+f.ID+"|"+Fingerprint(f))
	}
	return out
}
