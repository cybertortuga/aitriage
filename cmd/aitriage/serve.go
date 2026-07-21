package main

import (
	"context"
	"fmt"

	mcpserver "github.com/cybertortuga/aitriage/internal/agent/mcp"
	"github.com/spf13/cobra"
)

var (
	serveTransport string
	servePort      int
	serveProfile   string
	serveScanRoot  string
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start AITriage as an MCP server",
	Long: `Expose AITriage security tools via Model Context Protocol for Claude Code, Cursor, etc.

Profiles:
  full       All tools, including the mutating securecoder_ignore. No path
             confinement unless --scan-root is set. (default)
  safe  Safe-by-default mode for safe-profile installs: only read-only scan/analysis
             tools; securecoder_ignore is not exposed. Requires --scan-root and
             confines every path (including symlinks) to that directory.`,
	RunE: runServe,
}

func init() {
	rootCmd.AddCommand(serveCmd)
	serveCmd.Flags().StringVar(&serveTransport, "transport", "stdio", "Transport type: stdio or sse")
	serveCmd.Flags().IntVar(&servePort, "port", 8080, "Port for SSE transport")
	serveCmd.Flags().StringVar(&serveProfile, "profile", "full", "Tool profile: full or safe")
	serveCmd.Flags().StringVar(&serveScanRoot, "scan-root", "", "Confine all path arguments to this directory (required for safe profile)")
}

func runServe(cmd *cobra.Command, args []string) error {
	profile, err := mcpserver.ParseProfile(serveProfile)
	if err != nil {
		return err
	}
	if profile == mcpserver.ProfileSafe && serveScanRoot == "" {
		return fmt.Errorf("the safe profile requires --scan-root to confine scans to the project directory")
	}

	srv, err := mcpserver.NewServerWithConfig(Version, mcpserver.Config{
		Profile:  profile,
		ScanRoot: serveScanRoot,
	})
	if err != nil {
		return err
	}
	return srv.Run(context.Background(), serveTransport, servePort)
}
