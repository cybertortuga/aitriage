package scorer

import (
	"github.com/cybertortuga/aitriage/internal/engine/core"
)

// Calculate processes checking results to determine if CI should fail and returns a Security Score (0-100).
func Calculate(results []core.CheckResult) (hasCriticalFailures bool, securityScore int) {
	securityScore = 100
	for _, r := range results {
		// Do not penalize if the issue has been manually triaged as ignored or checked
		if r.AuditStatus == core.AuditStatusIgnored || r.AuditStatus == core.AuditStatusTriage {
			continue
		}
		if r.Status == core.Absent {
			penalty := 0
			switch r.Severity {
			case "CRITICAL":
				penalty = 25
				hasCriticalFailures = true
			case "HIGH":
				penalty = 15
				hasCriticalFailures = true
			case "MEDIUM":
				penalty = 5
			case "LOW":
				penalty = 1
			}
			securityScore -= penalty
		}
	}

	if securityScore < 0 {
		securityScore = 0
	}

	return hasCriticalFailures, securityScore
}
