package mcp

import (
	"context"
	"fmt"

	"github.com/cybertortuga/aitriage/internal/agent/architect"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type diagramInput struct {
	Path string `json:"path"`
}

type diagramResult struct {
	Mermaid string `json:"mermaid"`
}

func registerDiagramTool(srv *mcp.Server) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "aitriage_diagram",
		Description: "Generate a Mermaid diagram representing the architecture of the project based on detected files and docker-compose configurations.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input diagramInput) (*mcp.CallToolResult, diagramResult, error) {
		diagram, err := architect.GenerateMermaidDiagram(input.Path)
		if err != nil {
			return nil, diagramResult{}, fmt.Errorf("diagram generation error: %v", err)
		}
		return nil, diagramResult{Mermaid: diagram}, nil
	})
}
