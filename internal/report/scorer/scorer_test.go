package scorer

import (
	"github.com/cybertortuga/aitriage/internal/engine/core"
	"testing"
)

func TestCalculate(t *testing.T) {
	tests := []struct {
		name                string
		results             []core.CheckResult
		hasCriticalFailures bool
		expectedScore       int
	}{
		{
			name: "Clean report",
			results: []core.CheckResult{
				{ID: "STD-01", Status: core.Present},
				{ID: "STD-02", Status: core.Present},
			},
			hasCriticalFailures: false,
			expectedScore:       100,
		},
		{
			name: "One absent LOW issue",
			results: []core.CheckResult{
				{ID: "STD-01", Status: core.Absent, Severity: "LOW"},
				{ID: "STD-02", Status: core.Present},
			},
			hasCriticalFailures: false,
			expectedScore:       99, // -1
		},
		{
			name: "One absent HIGH issue",
			results: []core.CheckResult{
				{ID: "ENTR-01", Status: core.Absent, Severity: "HIGH"},
			},
			hasCriticalFailures: true,
			expectedScore:       85, // -15
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			hasCritical, securityScore := Calculate(tc.results)
			if hasCritical != tc.hasCriticalFailures {
				t.Errorf("%s: expected hasCriticalFailures %v, got %v", tc.name, tc.hasCriticalFailures, hasCritical)
			}
			if securityScore != tc.expectedScore {
				t.Errorf("%s: expected securityScore %d, got %d", tc.name, tc.expectedScore, securityScore)
			}
		})
	}
}
