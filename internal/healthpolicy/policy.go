package healthpolicy

import (
	"github.com/cybertortuga/aitriage/internal/config"
	"github.com/cybertortuga/aitriage/internal/report/healthcheck"
)

type Overrides struct {
	Profile         string
	ProfileSet      bool
	FailOn          string
	FailOnSet       bool
	MinimumScore    int
	MinimumScoreSet bool
}

// FromConfig converts .aitriage.yaml policy settings plus legacy top-level
// fields into the single runtime IB gate policy used by scan/agent/CI.
func FromConfig(cfg *config.Config) healthcheck.Policy {
	if cfg == nil {
		return healthcheck.DefaultPolicy()
	}

	policy := healthcheck.PolicyForProfile(cfg.HealthCheck.Profile)

	if cfg.HealthCheck.FailOn != "" {
		policy.FailOn = cfg.HealthCheck.FailOn
	} else if cfg.StrictMode {
		policy.FailOn = healthcheck.FailOnAny
	}

	if cfg.HealthCheck.MinimumScore != nil {
		policy.MinimumScore = *cfg.HealthCheck.MinimumScore
	} else if cfg.FailScore > 0 {
		policy.MinimumScore = cfg.FailScore
	}
	if cfg.HealthCheck.MaxCritical != nil {
		policy.MaxCritical = *cfg.HealthCheck.MaxCritical
	}
	if cfg.HealthCheck.MaxHigh != nil {
		policy.MaxHigh = *cfg.HealthCheck.MaxHigh
	}
	if cfg.HealthCheck.MaxMedium != nil {
		policy.MaxMedium = *cfg.HealthCheck.MaxMedium
	}
	if len(cfg.HealthCheck.BlockSources) > 0 {
		policy.BlockSources = cfg.HealthCheck.BlockSources
	}
	if len(cfg.HealthCheck.BlockClasses) > 0 {
		policy.BlockClasses = cfg.HealthCheck.BlockClasses
	}

	return healthcheck.NormalizePolicy(policy)
}

func HasConfiguredGate(cfg *config.Config) bool {
	if cfg == nil {
		return false
	}
	hc := cfg.HealthCheck
	return cfg.StrictMode ||
		cfg.FailScore > 0 ||
		hc.Profile != "" ||
		hc.FailOn != "" ||
		hc.MinimumScore != nil ||
		hc.MaxCritical != nil ||
		hc.MaxHigh != nil ||
		hc.MaxMedium != nil ||
		len(hc.BlockSources) > 0 ||
		len(hc.BlockClasses) > 0
}

// ApplyOverrides applies explicit CLI/action inputs on top of the config policy.
func ApplyOverrides(policy healthcheck.Policy, overrides Overrides) healthcheck.Policy {
	if overrides.ProfileSet {
		next := healthcheck.PolicyForProfile(overrides.Profile)
		next.BlockSources = policy.BlockSources
		next.BlockClasses = policy.BlockClasses
		policy = next
	}
	if overrides.FailOnSet {
		policy.FailOn = overrides.FailOn
	}
	if overrides.MinimumScoreSet {
		policy.MinimumScore = overrides.MinimumScore
	}
	return healthcheck.NormalizePolicy(policy)
}
