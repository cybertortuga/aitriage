package main

import (
	"testing"

	"github.com/cybertortuga/aitriage/internal/config"
	"github.com/cybertortuga/aitriage/internal/report/healthcheck"
	"github.com/spf13/cobra"
)

func TestAgentPolicyDefaultsToNonBlockingWithoutConfiguredGate(t *testing.T) {
	cmd := newAgentPolicyTestCommand()
	agentFailOn = healthcheck.FailOnNever
	agentFailScore = 0
	agentHealthProfile = ""

	policy := agentPolicyFromFlags(cmd, &config.Config{})
	if policy.FailOn != healthcheck.FailOnNever {
		t.Fatalf("fail_on = %q; want never", policy.FailOn)
	}
	if policy.MinimumScore != 0 {
		t.Fatalf("minimum_score = %d; want 0", policy.MinimumScore)
	}
}

func TestAgentPolicyUsesConfiguredGate(t *testing.T) {
	cmd := newAgentPolicyTestCommand()
	agentFailOn = healthcheck.FailOnNever
	agentFailScore = 0
	agentHealthProfile = ""

	policy := agentPolicyFromFlags(cmd, &config.Config{
		HealthCheck: config.HealthCheckPolicyConfig{
			Profile: healthcheck.PolicyStrict,
		},
	})
	if policy.Profile != healthcheck.PolicyStrict {
		t.Fatalf("profile = %q; want strict", policy.Profile)
	}
	if policy.FailOn != healthcheck.FailOnAny {
		t.Fatalf("fail_on = %q; want strict default any", policy.FailOn)
	}
}

func TestAgentPolicyFailScoreEnablesGateUnlessFailOnExplicit(t *testing.T) {
	cmd := newAgentPolicyTestCommand()
	if err := cmd.Flags().Set("fail-score", "90"); err != nil {
		t.Fatal(err)
	}
	agentFailOn = healthcheck.FailOnNever
	agentFailScore = 90
	agentHealthProfile = ""

	policy := agentPolicyFromFlags(cmd, &config.Config{})
	if policy.FailOn != healthcheck.FailOnCritical {
		t.Fatalf("fail_on = %q; want critical when fail-score enables gate", policy.FailOn)
	}
	if policy.MinimumScore != 90 {
		t.Fatalf("minimum_score = %d; want 90", policy.MinimumScore)
	}
}

func newAgentPolicyTestCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "agent"}
	cmd.Flags().String("health-profile", "", "")
	cmd.Flags().String("fail-on", healthcheck.FailOnNever, "")
	cmd.Flags().Int("fail-score", 0, "")
	return cmd
}
