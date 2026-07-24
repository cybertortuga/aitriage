package mcp

import (
	"context"
	"fmt"
	"regexp"

	"github.com/dodobrands/aitriage/internal/scanner/external"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// approvedSemgrepConfig matches the only Semgrep --config values allowed under
// the safe profile: the built-in "auto" ruleset and Semgrep registry
// references (p/…, r/…, s/…). Anything else — an absolute path, a relative path,
// "..", or a URL — is rejected so a safe client cannot point Semgrep at an
// arbitrary local file or a remote config outside the project.
var approvedSemgrepConfig = regexp.MustCompile(`^(auto|[prs]/[A-Za-z0-9][A-Za-z0-9._/-]*)$`)

// sanitizeSemgrepConfig enforces the safe Semgrep config allowlist. When
// restrict is false (full profile) the config is passed through unchanged.
func sanitizeSemgrepConfig(cfg string, restrict bool) (string, error) {
	if !restrict || cfg == "" {
		return cfg, nil
	}
	if !approvedSemgrepConfig.MatchString(cfg) {
		return "", fmt.Errorf("safe profile allows only the built-in Semgrep config (\"auto\") or a registry ref (p/…, r/…, s/…); %q is not permitted", cfg)
	}
	return cfg, nil
}

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

func registerExternalTools(srv *mcp.Server, guard *PathGuard, restrictConfig bool) {
	// Semgrep
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "run_semgrep",
		Description: "Run semgrep on the project. Requires semgrep to be installed.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input semgrepInput) (*mcp.CallToolResult, externalResult, error) {
		path, err := guard.Resolve(input.Path)
		if err != nil {
			return nil, externalResult{}, err
		}
		cfg, err := sanitizeSemgrepConfig(input.Config, restrictConfig)
		if err != nil {
			return nil, externalResult{}, err
		}
		findings, err := external.RunSemgrep(ctx, path, cfg)
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
		path, err := guard.Resolve(input.Path)
		if err != nil {
			return nil, externalResult{}, err
		}
		findings, err := external.RunGitleaks(ctx, path)
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
		path, err := guard.Resolve(input.Path)
		if err != nil {
			return nil, externalResult{}, err
		}
		findings, err := external.RunTrivy(ctx, path, input.ScanType)
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
		path, err := guard.Resolve(input.Path)
		if err != nil {
			return nil, externalResult{}, err
		}
		findings, err := external.RunBandit(ctx, path)
		if err != nil {
			return nil, externalResult{}, fmt.Errorf("bandit error: %v, please ensure it is installed", err)
		}
		return nil, externalResult{Findings: findings, Count: len(findings)}, nil
	})
}
