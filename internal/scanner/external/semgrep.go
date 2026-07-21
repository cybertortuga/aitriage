package external

import (
	"context"
	"encoding/json"
	"fmt"
)

type semgrepOutput struct {
	Results []struct {
		RuleID string `json:"check_id"`
		Extra  struct {
			Message  string `json:"message"`
			Severity string `json:"severity"`
		} `json:"extra"`
		Path  string `json:"path"`
		Start struct {
			Line int `json:"line"`
		} `json:"start"`
	} `json:"results"`
}

// RunSemgrep runs semgrep and returns unified findings.
// config: semgrep rules, e.g. "auto" or a path to a yaml file
func RunSemgrep(ctx context.Context, path, config string) ([]UnifiedFinding, error) {
	if !IsInstalled("semgrep") {
		return nil, fmt.Errorf("semgrep not installed")
	}
	if config == "" {
		config = "auto"
	}
	result, err := RunTool(ctx, "semgrep", "scan", "--json", "--config", config,
		"--exclude", "aitriage-reports", "--exclude", ".aitriage", "--exclude", ".aitriage-cache", path)
	if err != nil {
		return nil, fmt.Errorf("semgrep execution failed: %w", err)
	}
	var output semgrepOutput
	if err := json.Unmarshal([]byte(result.Stdout), &output); err != nil {
		return nil, fmt.Errorf("failed to parse semgrep output: %w", err)
	}
	findings := make([]UnifiedFinding, 0, len(output.Results))
	for _, r := range output.Results {
		if isGeneratedArtifactPath(r.Path) {
			continue
		}
		findings = append(findings, UnifiedFinding{
			Source:   "semgrep",
			RuleID:   r.RuleID,
			Severity: normalizeSeverity(r.Extra.Severity),
			Message:  r.Extra.Message,
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
