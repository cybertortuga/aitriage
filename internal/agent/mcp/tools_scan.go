package mcp

import (
	"context"
	"fmt"

	"github.com/cybertortuga/aitriage/internal/scanner"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type scanInput struct {
	Path          string `json:"path"`
	Stack         string `json:"stack,omitempty"`
	UniversalOnly bool   `json:"universal_only,omitempty"`
}

func registerScanTool(srv *mcp.Server) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "aitriage_scan",
		Description: "Run a full deterministic security scan on a project directory. Uses AST analysis, Shannon Entropy for secrets, and Entropy Code detection. No LLM required. Returns structured JSON report.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input scanInput) (*mcp.CallToolResult, scanner.ScanReport, error) {
		opts := scanner.ScanOptions{
			ForceStack:    input.Stack,
			UniversalOnly: input.UniversalOnly,
		}
		report, err := scanner.Scan(ctx, input.Path, opts)
		if err != nil {
			var empty scanner.ScanReport
			return nil, empty, fmt.Errorf("scan failed: %w", err)
		}
		return nil, report, nil
	})
}
