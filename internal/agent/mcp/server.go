package mcp

import (
	"context"
	"fmt"
	"net/http"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"golang.org/x/time/rate"
)

type Server struct {
	version string
	srv     *mcp.Server
}

func NewServer(version string) *Server {
	s := &Server{version: version}
	s.srv = mcp.NewServer(&mcp.Implementation{
		Name:    "aitriage",
		Version: version,
	}, nil)
	s.registerTools()
	s.registerResources()
	return s
}

func (s *Server) Run(ctx context.Context, transport string, port int) error {
	switch transport {
	case "stdio":
		return s.srv.Run(ctx, &mcp.StdioTransport{})
	case "sse":
		addr := fmt.Sprintf("0.0.0.0:%d", port)
		handler := mcp.NewSSEHandler(func(r *http.Request) *mcp.Server {
			return s.srv
		}, nil)
		mux := http.NewServeMux()

		securedHandler := corsMiddleware(rateLimitMiddleware(handler))
		mux.Handle("/sse", securedHandler)
		mux.Handle("/", securedHandler)
		httpSrv := &http.Server{Addr: addr, Handler: mux}

		fmt.Printf("  AITriage MCP Server (SSE)\n")
		fmt.Printf("  ─────────────────────────────────────\n")
		fmt.Printf("  Listening on http://%s\n", addr)
		fmt.Printf("  Add to your AI client:\n")
		fmt.Printf("    URL: http://localhost:%d/sse\n", port)
		fmt.Printf("  ─────────────────────────────────────\n\n")

		go func() {
			<-ctx.Done()
			httpSrv.Shutdown(context.Background()) //nolint:errcheck
		}()
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			return fmt.Errorf("SSE server error: %w", err)
		}
		return nil
	default:
		return fmt.Errorf("unknown transport: %s (supported: stdio, sse)", transport)
	}
}

func (s *Server) registerTools() {
	registerScanTool(s.srv)
	registerSecretsTool(s.srv)
	registerEntropyCheckTool(s.srv)
	registerArchitectureTool(s.srv)
	registerFixPlanTool(s.srv)
	registerScannersListTool(s.srv)
	registerExternalTools(s.srv)
	registerSecureCoderTools(s.srv)
	registerDeployTool(s.srv)
	registerNFRTool(s.srv)
	registerDiagramTool(s.srv)
	registerHistoryTool(s.srv)
}

func (s *Server) registerResources() {
	registerPlaybookResource(s.srv)
	registerGuidelinesResource(s.srv)
}

// ── Middleware ───────────────────────────────────────────────────────────────

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

var globalLimiter = rate.NewLimiter(rate.Limit(10), 20)

func rateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !globalLimiter.Allow() {
			http.Error(w, "429 Too Many Requests", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}
