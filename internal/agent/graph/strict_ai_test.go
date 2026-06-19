package graph

import (
	"encoding/json"
	"testing"
)

func TestValidateFindingDispositionsRequiresOneValidDecisionPerFinding(t *testing.T) {
	tests := []struct {
		name         string
		dispositions []FindingDisposition
		findingCount int
		wantErr      bool
	}{
		{
			name: "complete valid classification",
			dispositions: []FindingDisposition{
				{FindingIndex: 0, Disposition: "True Positive"},
				{FindingIndex: 1, Disposition: "False Positive"},
			},
			findingCount: 2,
		},
		{
			name:         "missing classification",
			dispositions: []FindingDisposition{{FindingIndex: 0, Disposition: "True Positive"}},
			findingCount: 2,
			wantErr:      true,
		},
		{
			name: "duplicate classification",
			dispositions: []FindingDisposition{
				{FindingIndex: 0, Disposition: "True Positive"},
				{FindingIndex: 0, Disposition: "False Positive"},
			},
			findingCount: 2,
			wantErr:      true,
		},
		{
			name:         "unsupported disposition",
			dispositions: []FindingDisposition{{FindingIndex: 0, Disposition: "Unknown"}},
			findingCount: 1,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateFindingDispositions(tt.dispositions, tt.findingCount)
			if (err != nil) != tt.wantErr {
				t.Fatalf("validateFindingDispositions() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPoCStepAcceptsDecimalStepLabels(t *testing.T) {
	var results []PoCResult
	err := json.Unmarshal([]byte(`[
		{"reasoning_steps":[{"step":2.1,"description":"trace","result":"reachable"}]}
	]`), &results)
	if err != nil {
		t.Fatalf("unmarshal PoC results: %v", err)
	}
	if got := results[0].ReasoningSteps[0].Step.String(); got != "2.1" {
		t.Fatalf("step = %q; want 2.1", got)
	}
}
