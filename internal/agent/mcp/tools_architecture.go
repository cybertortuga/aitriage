package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/dodobrands/aitriage/internal/agent/architect"
	"github.com/dodobrands/aitriage/internal/engine/core"
	"github.com/dodobrands/aitriage/internal/scanner/detector"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type archInput struct {
	Path string `json:"path"`
}

type archResult struct {
	Stacks          []string        `json:"stacks"`
	TotalFiles      int             `json:"total_files"`
	FilesByExt      map[string]int  `json:"files_by_extension"`
	KeyFilesPresent map[string]bool `json:"key_files_present"`
	Summary         string          `json:"summary"`
}

func registerArchitectureTool(srv *mcp.Server, guard *PathGuard) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "aitriage_architecture",
		Description: "Analyze project structure: detect tech stacks, count files by extension, check for key files (Dockerfile, docker-compose.yml, .env, Makefile, nginx.conf, terraform/*.tf, go.mod, package.json, requirements.txt). Call this FIRST before any scan.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input archInput) (*mcp.CallToolResult, archResult, error) {
		path, err := guard.Resolve(input.Path)
		if err != nil {
			return nil, archResult{}, err
		}
		ws, err := core.NewWorkspace(path)
		if err != nil {
			return nil, archResult{}, fmt.Errorf("failed to read workspace: %w", err)
		}

		diag, err := architect.GenerateMermaidDiagram(path)
		if err != nil {
			return nil, archResult{}, fmt.Errorf("failed to generate architecture diagram: %v", err)
		}

		components := architect.DetectComponents(path)
		tm := architect.GenerateThreatModel(components, path)
		tmJSON, _ := json.MarshalIndent(tm, "", "  ")

		stacks := detector.DetectProjects(ws)
		byExt := make(map[string]int)
		for _, f := range ws.Files {
			ext := filepath.Ext(f.Path)
			byExt[ext]++
		}
		keyFiles := map[string]bool{
			"Dockerfile":         fileExists(filepath.Join(path, "Dockerfile")),
			"docker-compose.yml": fileExists(filepath.Join(path, "docker-compose.yml")),
			".env":               fileExists(filepath.Join(path, ".env")),
			".env.example":       fileExists(filepath.Join(path, ".env.example")),
			"Makefile":           fileExists(filepath.Join(path, "Makefile")),
			"nginx.conf":         fileExists(filepath.Join(path, "nginx.conf")),
			"go.mod":             fileExists(filepath.Join(path, "go.mod")),
			"package.json":       fileExists(filepath.Join(path, "package.json")),
			"requirements.txt":   fileExists(filepath.Join(path, "requirements.txt")),
			"terraform":          dirExists(filepath.Join(path, "terraform")),
		}
		stackNames := make([]string, 0, len(stacks))
		for _, s := range stacks {
			stackNames = append(stackNames, string(s.Stack)) // Fixed string(s) to string(s.Stack)
		}
		res := archResult{
			Stacks:          stackNames,
			TotalFiles:      len(ws.Files),
			FilesByExt:      byExt,
			KeyFilesPresent: keyFiles,
			Summary:         fmt.Sprintf("Detected stacks: %v. Total files: %d.\n\nArchitecture Diagram:\n%s\n\nThreat Model:\n%s", stackNames, len(ws.Files), diag, string(tmJSON)),
		}
		return nil, res, nil
	})
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
