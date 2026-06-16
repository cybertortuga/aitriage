package mcp

import (
	"context"
	"fmt"

	"github.com/cybertortuga/aitriage/internal/scanner/deployaudit"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type deployInput struct {
	Path string `json:"path"`
}

type deployResult struct {
	Findings []deployaudit.DeployFinding `json:"findings"`
	Count    int                         `json:"count"`
}

func registerDeployTool(srv *mcp.Server) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "aitriage_deploy_audit",
		Description: "Analyze Dockerfile and docker-compose configurations for security vulnerabilities like root execution, privileged mode, and hardcoded secrets.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input deployInput) (*mcp.CallToolResult, deployResult, error) {
		findings, err := deployaudit.AuditDeployFiles(input.Path)
		if err != nil {
			return nil, deployResult{}, fmt.Errorf("deploy audit error: %v", err)
		}
		return nil, deployResult{Findings: findings, Count: len(findings)}, nil
	})
}
