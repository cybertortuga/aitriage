package healthpolicy

import (
	"testing"

	"github.com/cybertortuga/aitriage/internal/config"
	"github.com/cybertortuga/aitriage/internal/report/healthcheck"
)

func TestFromConfigKeepsLegacyFallback(t *testing.T) {
	policy := FromConfig(&config.Config{
		StrictMode: true,
		FailScore:  75,
	})

	if policy.Profile != healthcheck.PolicyBaseline {
		t.Fatalf("profile = %q; want baseline", policy.Profile)
	}
	if policy.FailOn != healthcheck.FailOnAny {
		t.Fatalf("fail_on = %q; want any", policy.FailOn)
	}
	if policy.MinimumScore != 75 {
		t.Fatalf("minimum_score = %d; want 75", policy.MinimumScore)
	}
}

func TestFromConfigNewBlockOverridesLegacy(t *testing.T) {
	zero := 0
	policy := FromConfig(&config.Config{
		StrictMode: true,
		FailScore:  75,
		HealthCheck: config.HealthCheckPolicyConfig{
			Profile:      healthcheck.PolicyStrict,
			FailOn:       healthcheck.FailOnNever,
			MinimumScore: &zero,
		},
	})

	if policy.Profile != healthcheck.PolicyStrict {
		t.Fatalf("profile = %q; want strict", policy.Profile)
	}
	if policy.FailOn != healthcheck.FailOnNever {
		t.Fatalf("fail_on = %q; want never", policy.FailOn)
	}
	if policy.MinimumScore != 0 {
		t.Fatalf("minimum_score = %d; want explicit 0", policy.MinimumScore)
	}
}

func TestFromConfigNormalizesBlockLists(t *testing.T) {
	policy := FromConfig(&config.Config{
		HealthCheck: config.HealthCheckPolicyConfig{
			BlockSources: []string{" Gitleaks ", "gitleaks"},
			BlockClasses: []string{" Hardcoded-Secret ", ""},
		},
	})

	if len(policy.BlockSources) != 1 || policy.BlockSources[0] != "gitleaks" {
		t.Fatalf("block_sources = %#v", policy.BlockSources)
	}
	if len(policy.BlockClasses) != 1 || policy.BlockClasses[0] != "hardcoded-secret" {
		t.Fatalf("block_classes = %#v", policy.BlockClasses)
	}
}

func TestApplyOverridesUsesOnlyExplicitInputs(t *testing.T) {
	policy := FromConfig(&config.Config{
		HealthCheck: config.HealthCheckPolicyConfig{
			Profile:      healthcheck.PolicyStandard,
			FailOn:       healthcheck.FailOnNever,
			BlockSources: []string{"gitleaks"},
		},
	})

	policy = ApplyOverrides(policy, Overrides{
		Profile:         healthcheck.PolicyStrict,
		ProfileSet:      true,
		MinimumScore:    0,
		MinimumScoreSet: true,
	})

	if policy.Profile != healthcheck.PolicyStrict {
		t.Fatalf("profile = %q; want strict", policy.Profile)
	}
	if policy.FailOn != healthcheck.FailOnAny {
		t.Fatalf("fail_on = %q; want strict profile default any", policy.FailOn)
	}
	if policy.MinimumScore != 0 {
		t.Fatalf("minimum_score = %d; want explicit 0 override", policy.MinimumScore)
	}
	if len(policy.BlockSources) != 1 || policy.BlockSources[0] != "gitleaks" {
		t.Fatalf("block_sources = %#v", policy.BlockSources)
	}
}

func TestHasConfiguredGate(t *testing.T) {
	if HasConfiguredGate(&config.Config{}) {
		t.Fatal("empty config should not be treated as an explicit gate")
	}
	if !HasConfiguredGate(&config.Config{FailScore: 80}) {
		t.Fatal("fail_score should be treated as an explicit gate")
	}
	if !HasConfiguredGate(&config.Config{HealthCheck: config.HealthCheckPolicyConfig{FailOn: healthcheck.FailOnNever}}) {
		t.Fatal("health_check.fail_on should be treated as an explicit gate")
	}
}
