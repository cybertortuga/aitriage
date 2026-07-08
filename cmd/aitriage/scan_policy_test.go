package main

import (
	"testing"

	"github.com/cybertortuga/aitriage/internal/config"
	"github.com/cybertortuga/aitriage/internal/report/healthcheck"
	"github.com/cybertortuga/aitriage/internal/scanner"
	"github.com/spf13/cobra"
)

func TestScanPolicyFromFlagsDoesNotOverrideConfigWithDefaults(t *testing.T) {
	cmd := newScanPolicyTestCommand()
	failOn = healthcheck.FailOnCritical
	failScore = 0
	healthProfile = ""

	report := scanner.ScanReport{
		Config: &config.Config{
			HealthCheck: config.HealthCheckPolicyConfig{
				Profile: healthcheck.PolicyStandard,
				FailOn:  healthcheck.FailOnNever,
			},
		},
	}

	policy := scanPolicyFromFlags(cmd, report)
	if policy.Profile != healthcheck.PolicyStandard {
		t.Fatalf("profile = %q; want standard", policy.Profile)
	}
	if policy.FailOn != healthcheck.FailOnNever {
		t.Fatalf("fail_on = %q; want config never", policy.FailOn)
	}
}

func TestScanPolicyFromFlagsAppliesExplicitOverrides(t *testing.T) {
	cmd := newScanPolicyTestCommand()
	if err := cmd.Flags().Set("health-profile", healthcheck.PolicyStrict); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("fail-on", healthcheck.FailOnNever); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("fail-score", "0"); err != nil {
		t.Fatal(err)
	}
	failOn = healthcheck.FailOnNever
	failScore = 0
	healthProfile = healthcheck.PolicyStrict

	report := scanner.ScanReport{
		Config: &config.Config{
			HealthCheck: config.HealthCheckPolicyConfig{
				Profile:      healthcheck.PolicyBaseline,
				BlockSources: []string{"gitleaks"},
			},
		},
	}

	policy := scanPolicyFromFlags(cmd, report)
	if policy.Profile != healthcheck.PolicyStrict {
		t.Fatalf("profile = %q; want strict", policy.Profile)
	}
	if policy.FailOn != healthcheck.FailOnNever {
		t.Fatalf("fail_on = %q; want explicit never", policy.FailOn)
	}
	if policy.MinimumScore != 0 {
		t.Fatalf("minimum_score = %d; want explicit 0", policy.MinimumScore)
	}
	if len(policy.BlockSources) != 1 || policy.BlockSources[0] != "gitleaks" {
		t.Fatalf("block_sources = %#v", policy.BlockSources)
	}
}

func newScanPolicyTestCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "scan"}
	cmd.Flags().String("health-profile", "", "")
	cmd.Flags().String("fail-on", healthcheck.FailOnCritical, "")
	cmd.Flags().Int("fail-score", 0, "")
	return cmd
}
