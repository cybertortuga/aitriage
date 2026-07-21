package llm

import (
	"testing"

	"github.com/cybertortuga/aitriage/internal/scanner/external"
)

func TestMissingRequiredScanners(t *testing.T) {
	tests := []struct {
		name string
		exec []external.ScannerExecution
		want []string
	}{
		{
			name: "all completed",
			exec: []external.ScannerExecution{
				{Scanner: "aitriage", Status: external.StatusCompleted},
				{Scanner: "semgrep", Status: external.StatusCompleted},
				{Scanner: "gitleaks", Status: external.StatusCompleted},
				{Scanner: "bandit", Status: external.StatusCompleted},
				{Scanner: "trivy_fs", Status: external.StatusCompleted},
				{Scanner: "trivy_config", Status: external.StatusCompleted},
			},
			want: nil,
		},
		{
			name: "failed trivy config is not hidden by filesystem success",
			exec: []external.ScannerExecution{
				{Scanner: "aitriage", Status: external.StatusCompleted},
				{Scanner: "semgrep", Status: external.StatusCompleted},
				{Scanner: "gitleaks", Status: external.StatusCompleted},
				{Scanner: "bandit", Status: external.StatusNotApplicable},
				{Scanner: "trivy_fs", Status: external.StatusCompleted},
				{Scanner: "trivy_config", Status: external.StatusFailed},
			},
			want: []string{"trivy_config"},
		},
		{
			name: "missing and failed reported",
			exec: []external.ScannerExecution{
				{Scanner: "aitriage", Status: external.StatusCompleted},
				{Scanner: "semgrep", Status: external.StatusMissing},
				{Scanner: "gitleaks", Status: external.StatusCompleted},
				{Scanner: "bandit", Status: external.StatusFailed},
				{Scanner: "trivy_fs", Status: external.StatusMissing},
				{Scanner: "trivy_config", Status: external.StatusMissing},
			},
			want: []string{"semgrep", "trivy_fs", "trivy_config", "bandit"},
		},
		{
			name: "nothing ran",
			exec: nil,
			want: []string{"aitriage", "semgrep", "trivy_fs", "trivy_config", "gitleaks", "bandit"},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := RichScanResult{ScannerExecutions: tc.exec}.MissingRequiredScanners()
			set := map[string]bool{}
			for _, g := range got {
				set[g] = true
			}
			if len(got) != len(tc.want) {
				t.Fatalf("missing = %v, want %v", got, tc.want)
			}
			for _, w := range tc.want {
				if !set[w] {
					t.Errorf("expected %q in missing set %v", w, got)
				}
			}
		})
	}
}

func TestEveryLogicalScannerFailsClosedIndependently(t *testing.T) {
	for _, failed := range external.RequiredScannerExecutions {
		t.Run(failed, func(t *testing.T) {
			executions := make([]external.ScannerExecution, 0, len(external.RequiredScannerExecutions))
			for _, scanner := range external.RequiredScannerExecutions {
				status := external.StatusCompleted
				if scanner == failed {
					status = external.StatusFailed
				}
				executions = append(executions, external.ScannerExecution{Scanner: scanner, Status: status})
			}
			missing := RichScanResult{ScannerExecutions: executions}.MissingRequiredScanners()
			if len(missing) != 1 || missing[0] != failed {
				t.Fatalf("failed %s -> missing %v", failed, missing)
			}
		})
	}
}
