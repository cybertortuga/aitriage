package mcp

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const playbookContent = `# AITriage Full Security Audit Playbook
**Objective**: Perform a comprehensive security audit of an unknown codebase.

1. **Understand Architecture**: Call 'aitriage_architecture' FIRST to understand the stack and key files.
2. **Setup Tools**: Call 'list_available_scanners' to see what tools are installed.
3. **Core Scan**: Call 'aitriage_scan' to run standard AST checks.
4. **Deep Scans** (if available): 
   - If semgrep is installed, use it for deep static analysis.
   - If gitleaks is installed, run it against the repo.
5. **Entropy Check**: Call 'aitriage_entropy_check' to find AI-residue and bad coding practices.
6. **Secrets**: Call 'aitriage_secrets' to find hidden entropy secrets.
7. **Action Plan**: After gathering all data, use 'generate_fix_plan' to structure your output.
`

func registerPlaybookResource(srv *mcp.Server) {
	srv.AddResource(&mcp.Resource{
		URI:         "aitriage://playbook",
		Name:        "Full Security Audit Playbook",
		Description: "A step-by-step methodology for auditing unknown repositories using AITriage tools.",
		MIMEType:    "text/markdown",
	}, func(ctx context.Context, request *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		return &mcp.ReadResourceResult{
			Contents: []*mcp.ResourceContents{
				{
					URI:      "aitriage://playbook",
					MIMEType: "text/markdown",
					Text:     playbookContent,
				},
			},
		}, nil
	})
}
