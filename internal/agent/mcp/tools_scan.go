package mcp

import (
	"context"
	"fmt"

	"github.com/dodobrands/aitriage/internal/scanner"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type scanInput struct {
	Path          string `json:"path"`
	Stack         string `json:"stack,omitempty"`
	UniversalOnly bool   `json:"universal_only,omitempty"`
}

func registerScanTool(srv *mcp.Server, guard *PathGuard) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "aitriage_scan",
		Description: "RAW deterministic pre-scan only (AST, Shannon Entropy for secrets, Entropy Code detection). No LLM, NO triage, NOT a final security verdict — untriaged findings include false positives. For a real security review use aitriage_run_start, which runs the full AI-triage pipeline and gate. Returns a structured JSON report.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input scanInput) (*mcp.CallToolResult, scanner.ScanReport, error) {
		var empty scanner.ScanReport
		path, err := guard.Resolve(input.Path)
		if err != nil {
			return nil, empty, err
		}
		opts := scanner.ScanOptions{
			ForceStack:    input.Stack,
			UniversalOnly: input.UniversalOnly,
		}
		report, err := scanner.Scan(ctx, path, opts)
		if err != nil {
			return nil, empty, fmt.Errorf("scan failed: %w", err)
		}
		return nil, report, nil
	})
}
