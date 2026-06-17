package graph

import (
	"testing"

	"github.com/cybertortuga/aitriage/internal/report/healthcheck"
	"github.com/cybertortuga/aitriage/internal/scanner/deployaudit"
	"github.com/cybertortuga/aitriage/internal/scanner/network"
)

func TestComputeHealthCheckHonorsDeployAndNetworkFalsePositives(t *testing.T) {
	state := &AgentState{
		DeployFindings: []deployaudit.DeployFinding{
			{Issue: "dockerfile_root_user", Severity: "HIGH", File: "Dockerfile", Line: 3},
		},
		NetworkFindings: []network.NetworkFinding{
			{Port: 5432, Severity: "MEDIUM", Service: "postgres"},
		},
	}

	enrichFindings(state)
	if len(state.EnrichedFindings) != 2 {
		t.Fatalf("enriched findings = %d; want 2", len(state.EnrichedFindings))
	}
	if state.EnrichedFindings[0].ID != "dockerfile_root_user" {
		t.Fatalf("deploy enriched ID = %q; want dockerfile_root_user", state.EnrichedFindings[0].ID)
	}
	if state.EnrichedFindings[1].ID != "port-5432" {
		t.Fatalf("network enriched ID = %q; want port-5432", state.EnrichedFindings[1].ID)
	}

	state.FindingDispositions = []FindingDisposition{
		{FindingIndex: 0, FindingID: state.EnrichedFindings[0].VulnID, Disposition: "False Positive"},
		{FindingIndex: 1, FindingID: state.EnrichedFindings[1].VulnID, Disposition: "False Positive"},
	}

	computeHealthCheck(state)
	if state.HealthCheck.Breakdown.ActiveFindings != 0 {
		t.Fatalf("active findings = %d; want 0", state.HealthCheck.Breakdown.ActiveFindings)
	}
	if state.HealthCheck.Breakdown.IgnoredFindings != 2 {
		t.Fatalf("ignored findings = %d; want 2", state.HealthCheck.Breakdown.IgnoredFindings)
	}
	if state.HealthCheck.Score != 100 {
		t.Fatalf("score = %d; want 100", state.HealthCheck.Score)
	}
	if !state.HealthCheck.Verdict.Passed {
		t.Fatalf("verdict failed for false positives: %+v", state.HealthCheck.Verdict)
	}
}

func TestComputeHealthCheckAppliesAgentPolicyToUndisposedFindings(t *testing.T) {
	state := &AgentState{
		Policy: healthcheck.PolicyForProfile(healthcheck.PolicyStrict),
		DeployFindings: []deployaudit.DeployFinding{
			{Issue: "dockerfile_root_user", Severity: "HIGH", File: "Dockerfile", Line: 3},
		},
	}

	enrichFindings(state)
	computeHealthCheck(state)

	if state.HealthCheck.Breakdown.ActiveFindings != 1 {
		t.Fatalf("active findings = %d; want 1", state.HealthCheck.Breakdown.ActiveFindings)
	}
	if state.HealthCheck.Verdict.Passed {
		t.Fatalf("strict verdict passed; want failure: %+v", state.HealthCheck.Verdict)
	}
	if len(state.HealthCheck.Verdict.BlockingReasons) == 0 {
		t.Fatal("strict verdict has no blocking reasons")
	}
}
