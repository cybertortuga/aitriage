package remedy

import (
	"fmt"
	"strings"
	"time"

	"github.com/cybertortuga/aitriage/internal/engine/core"
)

type FixItem struct {
	RuleID     string   `json:"rule_id"`
	Name       string   `json:"name"`
	Severity   string   `json:"severity"`
	File       string   `json:"file"`
	Line       int      `json:"line"`
	FixPrompt  string   `json:"fix_prompt"`
	FixExample string   `json:"fix_example"`
	References []string `json:"references"`
}

type FixPlan struct {
	GeneratedAt     time.Time `json:"generated_at"`
	TotalFindings   int       `json:"total_findings"`
	CriticalActions []FixItem `json:"critical_actions"`
	HighActions     []FixItem `json:"high_actions"`
	MediumActions   []FixItem `json:"medium_actions"`
}

// GenerateFixPlan создаёт структурированный план исправлений без LLM.
// Использует маппинг Rule ID → шаблон из templates.go
func GenerateFixPlan(results []core.CheckResult) FixPlan {
	plan := FixPlan{
		GeneratedAt:   time.Now(),
		TotalFindings: len(results),
	}
	for _, r := range results {
		tmpl, ok := fixTemplates[r.ID]
		if !ok {
			tmpl = defaultTemplate
		}
		item := FixItem{
			RuleID:     r.ID,
			Name:       r.Name,
			Severity:   r.Severity,
			File:       r.File,
			Line:       r.Line,
			FixPrompt:  fmt.Sprintf(tmpl.Prompt, r.File, r.Line),
			FixExample: tmpl.Example,
			References: tmpl.References,
		}
		switch r.Severity {
		case "CRITICAL":
			plan.CriticalActions = append(plan.CriticalActions, item)
		case "HIGH":
			plan.HighActions = append(plan.HighActions, item)
		default:
			plan.MediumActions = append(plan.MediumActions, item)
		}
	}
	return plan
}

// ToMarkdown конвертирует план в Markdown для вставки в Claude Code / Cursor
func (p FixPlan) ToMarkdown() string {
	var sb strings.Builder
	sb.WriteString("# AITriage Security Fix Plan\n")
	sb.WriteString(fmt.Sprintf("Generated: %s\n\n", p.GeneratedAt.Format("2006-01-02 15:04")))
	sb.WriteString(fmt.Sprintf("**Total findings: %d**\n\n", p.TotalFindings))

	if len(p.CriticalActions) > 0 {
		sb.WriteString("## 🔴 CRITICAL — Fix Immediately\n\n")
		for _, item := range p.CriticalActions {
			sb.WriteString(fmt.Sprintf("### %s\n", item.Name))
			sb.WriteString(fmt.Sprintf("**File:** `%s` line %d\n\n", item.File, item.Line))
			sb.WriteString(fmt.Sprintf("**Fix prompt for AI agent:**\n```\n%s\n```\n\n", item.FixPrompt))
			if item.FixExample != "" {
				sb.WriteString(fmt.Sprintf("**Example:**\n```\n%s\n```\n\n", item.FixExample))
			}
			if len(item.References) > 0 {
				sb.WriteString(fmt.Sprintf("**Reference:** %s\n\n", strings.Join(item.References, ", ")))
			}
		}
	}

	if len(p.HighActions) > 0 {
		sb.WriteString("## 🟠 HIGH\n\n")
		for _, item := range p.HighActions {
			sb.WriteString(fmt.Sprintf("- **%s** (`%s:%d`) — %s\n", item.Name, item.File, item.Line, item.FixPrompt))
		}
		sb.WriteString("\n")
	}

	if len(p.MediumActions) > 0 {
		sb.WriteString("## 🟡 MEDIUM\n\n")
		for _, item := range p.MediumActions {
			sb.WriteString(fmt.Sprintf("- **%s** (`%s:%d`) — %s\n", item.Name, item.File, item.Line, item.FixPrompt))
		}
	}

	return sb.String()
}
