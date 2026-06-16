package main

import (
	"context"
	mcpserver "github.com/cybertortuga/aitriage/internal/agent/mcp"
	"github.com/spf13/cobra"
)

var (
	serveTransport string
	servePort      int
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start AITriage as an MCP server",
	Long:  "Expose AITriage security tools via Model Context Protocol for Claude Code, Cursor, etc.",
	RunE:  runServe,
}

func init() {
	rootCmd.AddCommand(serveCmd)
	serveCmd.Flags().StringVar(&serveTransport, "transport", "stdio", "Transport type: stdio or sse")
	serveCmd.Flags().IntVar(&servePort, "port", 8080, "Port for SSE transport")
}

func runServe(cmd *cobra.Command, args []string) error {
	srv := mcpserver.NewServer(Version)
	return srv.Run(context.Background(), serveTransport, servePort)
}
