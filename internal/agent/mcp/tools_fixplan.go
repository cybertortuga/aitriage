package mcp

import (
	"context"
	"fmt"

	"github.com/dodobrands/aitriage/internal/agent/remedy"
	"github.com/dodobrands/aitriage/internal/scanner"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type fixPlanInput struct {
	Path string `json:"path"`
}

type fixPlanResult struct {
	Markdown string `json:"markdown"`
}

func registerFixPlanTool(srv *mcp.Server, guard *PathGuard) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "generate_fix_plan",
		Description: "Scan the project and generate a structured fix plan with actionable prompts for each finding. Output is a markdown document ready to paste into Claude Code or Cursor as a task.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input fixPlanInput) (*mcp.CallToolResult, fixPlanResult, error) {
		path, err := guard.Resolve(input.Path)
		if err != nil {
			return nil, fixPlanResult{}, err
		}
		report, err := scanner.Scan(ctx, path, scanner.ScanOptions{})
		if err != nil {
			return nil, fixPlanResult{}, fmt.Errorf("scan failed: %w", err)
		}
		plan := remedy.GenerateFixPlan(report.Results)
		return nil, fixPlanResult{Markdown: plan.ToMarkdown()}, nil
	})
}
