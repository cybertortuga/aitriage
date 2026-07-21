package hostrun

import (
	"github.com/cybertortuga/aitriage/internal/config"
	"github.com/cybertortuga/aitriage/internal/healthpolicy"
	"github.com/cybertortuga/aitriage/internal/report/healthcheck"
)

// AuditPolicy resolves the health-check policy for a local host-agent run so it
// matches the CI/CD reference command `aitriage agent --fail-on any` exactly:
// load the project's .aitriage.yaml config, derive the base policy, and force
// FailOn=any. It never invents a policy — it reuses the same config + override
// path the CLI uses, so the local gate verdict is identical to CI.
func AuditPolicy(projectRoot string) healthcheck.Policy {
	cfg := config.LoadConfig(projectRoot)
	policy := healthpolicy.FromConfig(cfg)
	if !healthpolicy.HasConfiguredGate(cfg) {
		policy.FailOn = healthcheck.FailOnNever
		policy.MinimumScore = 0
	}
	policy = healthpolicy.ApplyOverrides(policy, healthpolicy.Overrides{
		FailOn:    "any",
		FailOnSet: true,
	})
	return healthcheck.NormalizePolicy(policy)
}
