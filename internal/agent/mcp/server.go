package mcp

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"golang.org/x/time/rate"
)

type Server struct {
	version string
	srv     *mcp.Server
	profile Profile
	guard   *PathGuard
}

// NewServer builds a server with the full profile and no path confinement.
// Backward-compatible entry point (e.g. Claude Desktop via `serve`).
func NewServer(version string) *Server {
	s, err := NewServerWithConfig(version, Config{Profile: ProfileFull})
	if err != nil {
		// The full profile builds no guard, so this cannot fail; panic guards
		// against a future regression rather than a runtime condition.
		panic(fmt.Sprintf("NewServer: %v", err))
	}
	return s
}

// NewServerWithConfig builds a server with an explicit profile and optional
// scan-root confinement.
func NewServerWithConfig(version string, cfg Config) (*Server, error) {
	if cfg.Profile == "" {
		cfg.Profile = ProfileFull
	}
	var guard *PathGuard
	if strings.TrimSpace(cfg.ScanRoot) != "" {
		g, err := NewPathGuard(cfg.ScanRoot)
		if err != nil {
			return nil, err
		}
		guard = g
	}

	s := &Server{version: version, profile: cfg.Profile, guard: guard}
	s.srv = mcp.NewServer(&mcp.Implementation{
		Name:    "aitriage",
		Version: version,
	}, nil)
	s.registerTools()
	s.registerResources()
	return s, nil
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
	registerScanTool(s.srv, s.guard)
	registerSecretsTool(s.srv, s.guard)
	registerEntropyCheckTool(s.srv, s.guard)
	registerArchitectureTool(s.srv, s.guard)
	registerFixPlanTool(s.srv, s.guard)
	registerScannersListTool(s.srv)
	registerExternalTools(s.srv, s.guard, s.profile == ProfileSafe)
	registerSecureCoderTools(s.srv, s.guard, s.profile.allowsMutation())
	registerDeployTool(s.srv, s.guard)
	registerNFRTool(s.srv, s.guard)
	registerDiagramTool(s.srv, s.guard)
	registerHistoryTool(s.srv, s.guard)
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
