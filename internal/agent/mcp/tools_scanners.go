package mcp

import (
	"context"
	"os/exec"

	"github.com/cybertortuga/aitriage/internal/scanner/external"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type scannersResult struct {
	Semgrep      bool   `json:"semgrep"`
	Gitleaks     bool   `json:"gitleaks"`
	Trivy        bool   `json:"trivy"`
	Bandit       bool   `json:"bandit"`
	SecureCoder  bool   `json:"securecoder"`
	SecureCoderBackend string `json:"securecoder_backend,omitempty"`
}

func registerScannersListTool(srv *mcp.Server) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "list_available_scanners",
		Description: "Check which external security scanners are installed and available in PATH. ALWAYS call this before calling run_semgrep, run_gitleaks, run_trivy, or run_bandit to avoid errors.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, scannersResult, error) {
		res := scannersResult{
			Semgrep:            isInstalled("semgrep"),
			Gitleaks:           isInstalled("gitleaks"),
			Trivy:              isInstalled("trivy"),
			Bandit:             isInstalled("bandit"),
			SecureCoder:        external.IsSecureCoderRunning(),
			SecureCoderBackend: external.SecureCoderBackend(),
		}
		return nil, res, nil
	})
}

func isInstalled(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}
