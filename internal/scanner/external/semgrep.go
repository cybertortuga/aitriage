package external

import (
	"context"
	"encoding/json"
	"fmt"
)

type semgrepOutput struct {
	Results []struct {
		RuleID  string `json:"check_id"`
		Message struct {
			Text string `json:"text"`
		} `json:"extra"`
		Path  string `json:"path"`
		Start struct {
			Line int `json:"line"`
		} `json:"start"`
		Severity string `json:"severity"`
	} `json:"results"`
}

// RunSemgrep запускает semgrep и возвращает унифицированные находки.
// config: правила для semgrep, например "auto" или путь к yaml файлу
func RunSemgrep(ctx context.Context, path, config string) ([]UnifiedFinding, error) {
	if !IsInstalled("semgrep") {
		return nil, fmt.Errorf("semgrep not installed")
	}
	if config == "" {
		config = "auto"
	}
	result, err := RunTool(ctx, "semgrep", "scan", "--json", "--config", config, path)
	if err != nil {
		return nil, fmt.Errorf("semgrep execution failed: %w", err)
	}
	var output semgrepOutput
	if err := json.Unmarshal([]byte(result.Stdout), &output); err != nil {
		return nil, fmt.Errorf("failed to parse semgrep output: %w", err)
	}
	findings := make([]UnifiedFinding, 0, len(output.Results))
	for _, r := range output.Results {
		findings = append(findings, UnifiedFinding{
			Source:   "semgrep",
			RuleID:   r.RuleID,
			Severity: normalizeSeverity(r.Severity),
			Message:  r.Message.Text,
			File:     r.Path,
			Line:     r.Start.Line,
		})
	}
	return findings, nil
}

func normalizeSeverity(s string) string {
	switch s {
	case "ERROR":
		return "HIGH"
	case "WARNING":
		return "MEDIUM"
	case "INFO":
		return "LOW"
	default:
		return s
	}
}
