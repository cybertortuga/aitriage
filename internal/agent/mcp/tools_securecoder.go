package mcp

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/cybertortuga/aitriage/internal/scanner/external"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ── Input / Output Types ─────────────────────────────────────────────────────

type secureCoderScanInput struct {
	Path string `json:"path"`
}

type secureCoderScanResult struct {
	Findings []external.UnifiedFinding `json:"findings"`
	Count    int                       `json:"count"`
	Backend  string                    `json:"backend"`
}

type secureCoderDepsInput struct {
	Registry string                       `json:"registry"`
	Packages []external.DepPackageRequest `json:"packages"`
}

type secureCoderDepsResult struct {
	UnsafeDependencies []external.DepFinding `json:"unsafe_dependencies"`
	Count              int                   `json:"count"`
}

type secureCoderIgnoreInput struct {
	FilePath           string `json:"file_path"`
	RuleID             string `json:"rule_id"`
	CodeSnippet        string `json:"code_snippet"`
	LineNumber         int    `json:"line_number"`
	VulnerabilityClass string `json:"vulnerability_class"`
	Reason             string `json:"reason"`
}

type secureCoderIgnoreResult struct {
	Success     bool   `json:"success"`
	VulnID      string `json:"vuln_id"`
	ContentHash string `json:"content_hash"`
}

// ── Registration ─────────────────────────────────────────────────────────────

// registerSecureCoderTools registers the SecureCoder tools. The read-only scan
// tools (run_securecoder, run_securecoder_deps) are always registered. The
// mutating securecoder_ignore is registered only when allowMutation is true;
// the safe profile passes false so safe-profile installs cannot silently suppress
// findings.
func registerSecureCoderTools(srv *mcp.Server, guard *PathGuard, allowMutation bool) {
	// run_securecoder — scan file or directory
	mcp.AddTool(srv, &mcp.Tool{
		Name: "run_securecoder",
		Description: "Scan a file or directory using SecureCoder (Antigravity IDE's built-in security scanner). " +
			"Uses semgrep or Wiz CLI as backend. Requires Antigravity IDE to be running with SecureCoder enabled. " +
			"Call list_available_scanners first to check if SecureCoder is available.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input secureCoderScanInput) (*mcp.CallToolResult, secureCoderScanResult, error) {
		if !external.IsSecureCoderRunning() {
			return nil, secureCoderScanResult{}, fmt.Errorf("SecureCoder is not running. Ensure Antigravity IDE is open with SecureCoder enabled")
		}

		scanPath, err := guard.Resolve(input.Path)
		if err != nil {
			return nil, secureCoderScanResult{}, err
		}

		var allFindings []external.UnifiedFinding

		// Check if path is a file or directory
		info, err := statPath(scanPath)
		if err != nil {
			return nil, secureCoderScanResult{}, fmt.Errorf("cannot access path: %w", err)
		}

		if !info.IsDir() {
			// Single file scan
			findings, err := external.RunSecureCoder(ctx, scanPath)
			if err != nil {
				return nil, secureCoderScanResult{}, fmt.Errorf("securecoder scan error: %w", err)
			}
			allFindings = findings
		} else {
			// Directory scan — walk and scan each file
			err := filepath.WalkDir(scanPath, func(path string, d fs.DirEntry, err error) error {
				if err != nil {
					return nil // skip inaccessible files
				}
				if d.IsDir() {
					// Skip common non-source directories
					name := d.Name()
					if name == ".git" || name == "node_modules" || name == "vendor" || name == "__pycache__" || name == ".next" {
						return filepath.SkipDir
					}
					return nil
				}
				// Only scan known source files
				if !isSourceFile(path) {
					return nil
				}
				findings, err := external.RunSecureCoder(ctx, path)
				if err != nil {
					return nil // skip files that fail, continue scanning
				}
				allFindings = append(allFindings, findings...)
				return nil
			})
			if err != nil {
				return nil, secureCoderScanResult{}, fmt.Errorf("directory walk error: %w", err)
			}
		}

		backend := external.SecureCoderBackend()
		return nil, secureCoderScanResult{
			Findings: allFindings,
			Count:    len(allFindings),
			Backend:  backend,
		}, nil
	})

	// run_securecoder_deps — dependency vulnerability check
	mcp.AddTool(srv, &mcp.Tool{
		Name: "run_securecoder_deps",
		Description: "Check packages for known vulnerabilities BEFORE importing them. " +
			"Uses SecureCoder's dependency scanner (Antigravity IDE). " +
			"Supported registries: npm, pypi, gomodproxy, rubygems, crates.io, maven, nuget. " +
			"Call list_available_scanners first to check if SecureCoder is available.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input secureCoderDepsInput) (*mcp.CallToolResult, secureCoderDepsResult, error) {
		if !external.IsSecureCoderRunning() {
			return nil, secureCoderDepsResult{}, fmt.Errorf("SecureCoder is not running. Ensure Antigravity IDE is open with SecureCoder enabled")
		}

		findings, err := external.RunSecureCoderDeps(ctx, input.Registry, input.Packages)
		if err != nil {
			return nil, secureCoderDepsResult{}, fmt.Errorf("securecoder dependency scan error: %w", err)
		}
		return nil, secureCoderDepsResult{
			UnsafeDependencies: findings,
			Count:              len(findings),
		}, nil
	})

	// securecoder_ignore — suppress a finding (MUTATING).
	// Not registered under the safe profile: safe-profile installs must not be able to
	// silently mark security findings as ignored.
	if !allowMutation {
		return
	}
	mcp.AddTool(srv, &mcp.Tool{
		Name: "securecoder_ignore",
		Description: "Suppress a SecureCoder finding as false positive, accepted risk, or won't fix. " +
			"The suppression uses content-hash and survives line shifts. " +
			"Valid reasons: 'False Positive', 'Accepted Risk', 'Won't Fix'.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input secureCoderIgnoreInput) (*mcp.CallToolResult, secureCoderIgnoreResult, error) {
		if !external.IsSecureCoderRunning() {
			return nil, secureCoderIgnoreResult{}, fmt.Errorf("SecureCoder is not running. Ensure Antigravity IDE is open with SecureCoder enabled")
		}

		filePath, err := guard.Resolve(input.FilePath)
		if err != nil {
			return nil, secureCoderIgnoreResult{}, err
		}

		result, err := external.IgnoreSecureCoderFinding(ctx, external.IgnoreRequest{
			FilePath:           filePath,
			RuleID:             input.RuleID,
			CodeSnippet:        input.CodeSnippet,
			LineNumber:         input.LineNumber,
			VulnerabilityClass: input.VulnerabilityClass,
			Reason:             input.Reason,
		})
		if err != nil {
			return nil, secureCoderIgnoreResult{}, fmt.Errorf("securecoder ignore error: %w", err)
		}
		return nil, secureCoderIgnoreResult{
			Success:     result.Success,
			VulnID:      result.VulnID,
			ContentHash: result.ContentHash,
		}, nil
	})
}

// ── Helpers ──────────────────────────────────────────────────────────────────

// statPath wraps os.Stat for testability.
var statPath = func(path string) (fs.FileInfo, error) {
	return os.Stat(path)
}

// isSourceFile checks if a file is a scannable source file by extension.
func isSourceFile(path string) bool {
	ext := filepath.Ext(path)
	switch ext {
	case ".go", ".py", ".js", ".ts", ".jsx", ".tsx",
		".java", ".rb", ".rs", ".cs", ".php",
		".yaml", ".yml", ".json", ".toml",
		".html", ".vue", ".svelte",
		".sql", ".sh", ".bash",
		".dockerfile", ".tf", ".hcl":
		return true
	}
	// Also match Dockerfile without extension
	base := filepath.Base(path)
	return base == "Dockerfile" || base == "docker-compose.yaml" || base == "docker-compose.yml"
}
