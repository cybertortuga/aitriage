package external

import (
	"context"
	"encoding/json"
	"fmt"
)

type trivyOutput struct {
	Results []struct {
		Target          string `json:"Target"`
		Vulnerabilities []struct {
			VulnerabilityID string `json:"VulnerabilityID"`
			Severity        string `json:"Severity"`
			Title           string `json:"Title"`
			Description     string `json:"Description"`
		} `json:"Vulnerabilities"`
	} `json:"Results"`
}

// RunTrivy запускает trivy и возвращает унифицированные находки.
// scanType: "fs" (filesystem) или "config" (IaC конфиги)
func RunTrivy(ctx context.Context, path, scanType string) ([]UnifiedFinding, error) {
	if !IsInstalled("trivy") {
		return nil, fmt.Errorf("trivy not installed")
	}
	if scanType == "" {
		scanType = "fs"
	}
	result, err := RunTool(ctx, "trivy", scanType, "--format", "json", "--quiet", path)
	if err != nil {
		return nil, fmt.Errorf("trivy execution failed: %w", err)
	}
	var output trivyOutput
	if err := json.Unmarshal([]byte(result.Stdout), &output); err != nil {
		return nil, fmt.Errorf("failed to parse trivy output: %w", err)
	}
	var findings []UnifiedFinding
	for _, res := range output.Results {
		for _, v := range res.Vulnerabilities {
			findings = append(findings, UnifiedFinding{
				Source:   "trivy",
				RuleID:   v.VulnerabilityID,
				Severity: v.Severity,
				Message:  fmt.Sprintf("%s: %s", v.Title, v.Description),
				File:     res.Target,
			})
		}
	}
	return findings, nil
}
