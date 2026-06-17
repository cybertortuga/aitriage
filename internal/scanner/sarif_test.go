package scanner_test

import (
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/cybertortuga/aitriage/internal/engine/core"
	"github.com/cybertortuga/aitriage/internal/scanner"
)

func TestToSARIFExportsOnlyActiveLocatedFindings(t *testing.T) {
	projectPath := t.TempDir()
	report := scanner.ScanReport{
		ProjectPath: projectPath,
		Results: []core.CheckResult{
			{
				ID:         "ACTIVE-HIGH",
				Name:       "Active high",
				Severity:   "HIGH",
				Status:     core.Absent,
				File:       filepath.Join(projectPath, "main.go"),
				Line:       42,
				Suggestion: "Fix it",
			},
			{
				ID:          "IGNORED-HIGH",
				Name:        "Ignored high",
				Severity:    "HIGH",
				Status:      core.Absent,
				File:        filepath.Join(projectPath, "ignored.go"),
				Line:        10,
				AuditStatus: core.AuditStatusIgnored,
			},
			{
				ID:       "PRESENT",
				Name:     "Present check",
				Severity: "LOW",
				Status:   core.Present,
				File:     filepath.Join(projectPath, "present.go"),
				Line:     1,
			},
			{
				ID:       "PROJECT-LEVEL",
				Name:     "Project-level issue",
				Severity: "MEDIUM",
				Status:   core.Absent,
			},
		},
	}

	data, err := report.ToSARIF()
	if err != nil {
		t.Fatalf("ToSARIF failed: %v", err)
	}

	var sarif scanner.SarifLog
	if err := json.Unmarshal(data, &sarif); err != nil {
		t.Fatalf("SARIF JSON did not parse: %v", err)
	}
	if len(sarif.Runs) != 1 {
		t.Fatalf("runs = %d; want 1", len(sarif.Runs))
	}
	results := sarif.Runs[0].Results
	if len(results) != 1 {
		t.Fatalf("results = %d; want 1 active located finding", len(results))
	}
	if results[0].RuleID != "ACTIVE-HIGH" {
		t.Fatalf("ruleId = %q; want ACTIVE-HIGH", results[0].RuleID)
	}
	if got := results[0].Locations[0].PhysicalLocation.ArtifactLocation.Uri; got != "main.go" {
		t.Fatalf("uri = %q; want main.go", got)
	}
	if results[0].Level != "error" {
		t.Fatalf("level = %q; want error", results[0].Level)
	}
}
