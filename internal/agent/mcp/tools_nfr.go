package mcp

import (
	"context"
	"fmt"

	"github.com/cybertortuga/aitriage/internal/scanner/nfr"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type nfrInput struct {
	Path string `json:"path"`
}

type nfrResult struct {
	Findings []nfr.NFRFinding `json:"findings"`
	Count    int              `json:"count"`
}

func registerNFRTool(srv *mcp.Server) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "aitriage_nfr_check",
		Description: "Check Non-Functional Requirements (NFR) compliance, such as missing Rate Limiting, CORS, or unprotected routes.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input nfrInput) (*mcp.CallToolResult, nfrResult, error) {
		findings, err := nfr.CheckNFR(input.Path)
		if err != nil {
			return nil, nfrResult{}, fmt.Errorf("could not run NFR check: %v", err)
		}
		return nil, nfrResult{Findings: findings, Count: len(findings)}, nil
	})
}
