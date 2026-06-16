package remedy_test

import (
	"testing"
	"time"

	"github.com/cybertortuga/aitriage/internal/agent/remedy"
	"github.com/cybertortuga/aitriage/internal/engine/core"
)

func TestGenerateFixPlan_Critical(t *testing.T) {
	results := []core.CheckResult{
		{ID: "ENTROPY-SECRET", Name: "Hardcoded Secret", Severity: "CRITICAL", File: "main.go", Line: 42},
	}
	plan := remedy.GenerateFixPlan(results)
	if plan.TotalFindings != 1 {
		t.Errorf("Expected 1 finding, got %d", plan.TotalFindings)
	}
	if len(plan.CriticalActions) != 1 {
		t.Errorf("Expected 1 critical action, got %d", len(plan.CriticalActions))
	}
	md := plan.ToMarkdown()
	if md == "" {
		t.Error("Expected non-empty markdown")
	}
}

func TestFixPlan_ToMarkdown(t *testing.T) {
	tests := []struct {
		name     string
		plan     remedy.FixPlan
		expected string
	}{
		{
			name: "fully populated",
			plan: remedy.FixPlan{
				GeneratedAt:   time.Date(2023, 10, 27, 10, 0, 0, 0, time.UTC),
				TotalFindings: 3,
				CriticalActions: []remedy.FixItem{
					{
						RuleID:     "RULE-1",
						Name:       "Critical Vulnerability",
						File:       "app.go",
						Line:       10,
						FixPrompt:  "Fix this critical issue.",
						FixExample: "example fix critical",
						References: []string{"http://ref1.com", "http://ref2.com"},
					},
				},
				HighActions: []remedy.FixItem{
					{
						RuleID:    "RULE-2",
						Name:      "High Vulnerability",
						File:      "app.go",
						Line:      20,
						FixPrompt: "Fix this high issue.",
					},
				},
				MediumActions: []remedy.FixItem{
					{
						RuleID:    "RULE-3",
						Name:      "Medium Vulnerability",
						File:      "app.go",
						Line:      30,
						FixPrompt: "Fix this medium issue.",
					},
				},
			},
			expected: `# AITriage Security Fix Plan
Generated: 2023-10-27 10:00

**Total findings: 3**

## 🔴 CRITICAL — Fix Immediately

### Critical Vulnerability
**File:** ` + "`app.go`" + ` line 10

**Fix prompt for AI agent:**
` + "```\nFix this critical issue.\n```" + `

**Example:**
` + "```\nexample fix critical\n```" + `

**Reference:** http://ref1.com, http://ref2.com

## 🟠 HIGH

- **High Vulnerability** (` + "`app.go:20`" + `) — Fix this high issue.

## 🟡 MEDIUM

- **Medium Vulnerability** (` + "`app.go:30`" + `) — Fix this medium issue.
`,
		},
		{
			name: "missing optional fields in critical",
			plan: remedy.FixPlan{
				GeneratedAt:   time.Date(2023, 10, 27, 10, 0, 0, 0, time.UTC),
				TotalFindings: 1,
				CriticalActions: []remedy.FixItem{
					{
						RuleID:    "RULE-1",
						Name:      "Critical Vulnerability",
						File:      "app.go",
						Line:      10,
						FixPrompt: "Fix this critical issue.",
					},
				},
			},
			expected: `# AITriage Security Fix Plan
Generated: 2023-10-27 10:00

**Total findings: 1**

## 🔴 CRITICAL — Fix Immediately

### Critical Vulnerability
**File:** ` + "`app.go`" + ` line 10

**Fix prompt for AI agent:**
` + "```\nFix this critical issue.\n```" + `

`,
		},
		{
			name: "empty",
			plan: remedy.FixPlan{
				GeneratedAt:   time.Date(2023, 10, 27, 10, 0, 0, 0, time.UTC),
				TotalFindings: 0,
			},
			expected: `# AITriage Security Fix Plan
Generated: 2023-10-27 10:00

**Total findings: 0**

`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			md := tt.plan.ToMarkdown()
			if md != tt.expected {
				t.Errorf("Expected markdown:\n%s\n\nGot:\n%s", tt.expected, md)
			}
		})
	}
}
