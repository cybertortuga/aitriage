package mcp

import (
	"context"
	"fmt"

	"github.com/cybertortuga/aitriage/internal/scanner/external"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type externalInput struct {
	Path string `json:"path"`
}

type trivyInput struct {
	Path     string `json:"path"`
	ScanType string `json:"scan_type,omitempty"`
}

type semgrepInput struct {
	Path   string `json:"path"`
	Config string `json:"config,omitempty"`
}

type externalResult struct {
	Findings []external.UnifiedFinding `json:"findings"`
	Count    int                       `json:"count"`
}

func registerExternalTools(srv *mcp.Server) {
	// Semgrep
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "run_semgrep",
		Description: "Run semgrep on the project. Requires semgrep to be installed.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input semgrepInput) (*mcp.CallToolResult, externalResult, error) {
		findings, err := external.RunSemgrep(ctx, input.Path, input.Config)
		if err != nil {
			return nil, externalResult{}, fmt.Errorf("semgrep error: %v, please ensure it is installed", err)
		}
		return nil, externalResult{Findings: findings, Count: len(findings)}, nil
	})

	// Gitleaks
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "run_gitleaks",
		Description: "Run gitleaks on the project to detect hardcoded secrets. Requires gitleaks to be installed.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input externalInput) (*mcp.CallToolResult, externalResult, error) {
		findings, err := external.RunGitleaks(ctx, input.Path)
		if err != nil {
			return nil, externalResult{}, fmt.Errorf("gitleaks error: %v, please ensure it is installed", err)
		}
		return nil, externalResult{Findings: findings, Count: len(findings)}, nil
	})

	// Trivy
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "run_trivy",
		Description: "Run trivy on the project to detect vulnerabilities in dependencies or IaC. Requires trivy to be installed. scan_type can be 'fs' or 'config'.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input trivyInput) (*mcp.CallToolResult, externalResult, error) {
		findings, err := external.RunTrivy(ctx, input.Path, input.ScanType)
		if err != nil {
			return nil, externalResult{}, fmt.Errorf("trivy error: %v, please ensure it is installed", err)
		}
		return nil, externalResult{Findings: findings, Count: len(findings)}, nil
	})

	// Bandit
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "run_bandit",
		Description: "Run bandit on Python projects to detect security issues. Requires bandit to be installed.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input externalInput) (*mcp.CallToolResult, externalResult, error) {
		findings, err := external.RunBandit(ctx, input.Path)
		if err != nil {
			return nil, externalResult{}, fmt.Errorf("bandit error: %v, please ensure it is installed", err)
		}
		return nil, externalResult{Findings: findings, Count: len(findings)}, nil
	})
}
