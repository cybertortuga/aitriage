package mcp

import (
	"context"
	"fmt"

	"github.com/cybertortuga/aitriage/internal/engine/core"
	"github.com/cybertortuga/aitriage/internal/scanner"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type entropyInput struct {
	Path string `json:"path"`
}

type entropyResult struct {
	Score   int                `json:"security_score"`
	Grade   string             `json:"security_grade"`
	Issues  []core.CheckResult `json:"issues"`
	Count   int                `json:"entropy_issue_count"`
	Summary string             `json:"summary"`
}

func registerEntropyCheckTool(srv *mcp.Server) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "aitriage_entropy_check",
		Description: "Check for entropy issues: chat residue in comments, missing error handling, God Files (>1500 lines), TODO stubs, and .cursorrules manipulation attempts.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input entropyInput) (*mcp.CallToolResult, entropyResult, error) {
		report, err := scanner.Scan(ctx, input.Path, scanner.ScanOptions{})
		if err != nil {
			return nil, entropyResult{}, fmt.Errorf("scan failed: %w", err)
		}
		var entropyIssues []core.CheckResult
		for _, r := range report.Results {
			if len(r.ID) >= 5 && r.ID[:5] == "ENTR-" {
				entropyIssues = append(entropyIssues, r)
			}
		}
		res := entropyResult{
			Score:   report.SecurityScore,
			Grade:   report.SecurityGrade,
			Issues:  entropyIssues,
			Count:   len(entropyIssues),
			Summary: fmt.Sprintf("Security Grade: %s (%d/100). Found %d entropy issues.", report.SecurityGrade, report.SecurityScore, len(entropyIssues)),
		}
		return nil, res, nil
	})
}
