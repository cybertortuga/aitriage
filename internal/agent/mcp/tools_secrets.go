package mcp

import (
	"context"
	"fmt"

	"github.com/cybertortuga/aitriage/internal/engine/core"
	"github.com/cybertortuga/aitriage/internal/scanner"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type secretsInput struct {
	Path string `json:"path"`
}

type secretsResult struct {
	Found   []core.CheckResult `json:"found"`
	Count   int                `json:"count"`
	Summary string             `json:"summary"`
}

func registerSecretsTool(srv *mcp.Server) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "aitriage_secrets",
		Description: "Scan for hardcoded secrets using Shannon Entropy analysis. Finds API keys, tokens, passwords even with non-obvious variable names. Returns only secret-related findings.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input secretsInput) (*mcp.CallToolResult, secretsResult, error) {
		report, err := scanner.Scan(ctx, input.Path, scanner.ScanOptions{})
		if err != nil {
			return nil, secretsResult{}, fmt.Errorf("scan failed: %w", err)
		}
		var secrets []core.CheckResult
		for _, r := range report.Results {
			if r.ID == "ENTROPY-SECRET" {
				secrets = append(secrets, r)
			}
		}
		res := secretsResult{
			Found:   secrets,
			Count:   len(secrets),
			Summary: fmt.Sprintf("Found %d potential secrets via Shannon Entropy analysis", len(secrets)),
		}
		return nil, res, nil
	})
}
