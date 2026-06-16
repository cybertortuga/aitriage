package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/cybertortuga/aitriage/internal/server"
	"github.com/spf13/cobra"
)

var (
	webPort       int
	webHostPrefix string
)

var webCmd = &cobra.Command{
	Use:   "web",
	Short: "Start AITriage Web UI (browser-based security dashboard)",
	Long: `Start a web server with a browser UI for scanning projects.

Designed to run inside Docker with host filesystem mounted:

  docker run -p 8080:8080 -v /:/host:ro ghcr.io/cybertortuga/aitriage web

Then open http://localhost:8080 and enter any path on your Mac.
The container automatically maps /Users/... → /host/Users/...

Flags:
  --port        Port to listen on (default 8080)
  --host-prefix Path prefix for host volume mount (default /host)`,
	RunE: func(cmd *cobra.Command, args []string) error {
		addr := fmt.Sprintf("0.0.0.0:%d", webPort)
		fmt.Printf("\n  AITriage Web UI\n")
		fmt.Printf("  ──────────────────────────────────────\n")
		fmt.Printf("  Open → http://localhost:%d\n", webPort)
		if webHostPrefix != "" {
			fmt.Printf("  Host prefix: %s (scanning host paths transparently)\n", webHostPrefix)
		}
		fmt.Printf("  ──────────────────────────────────────\n\n")

		// Removed auto-open browser to prevent opening :8080 during frontend dev on :5174

		prefix := webHostPrefix
		if prefix == "" {
			prefix = os.Getenv("HOST_PREFIX")
		}

		dbPath := os.Getenv("DB_PATH")
		if dbPath == "" {
			dbDir := filepath.Join(os.Getenv("HOME"), ".aitriage")
			if err := os.MkdirAll(dbDir, 0755); err != nil {
				return fmt.Errorf("failed to create config dir: %w", err)
			}
			dbPath = filepath.Join(dbDir, "aitriage.db")
		} else {
			dbDir := filepath.Dir(dbPath)
			if err := os.MkdirAll(dbDir, 0755); err != nil {
				return fmt.Errorf("failed to create config dir: %w", err)
			}
		}

		db, err := server.InitDB(dbPath)
		if err != nil {
			return fmt.Errorf("failed to init db: %w", err)
		}
		defer db.Close()

		srv := server.NewServer(prefix, db)
		return srv.Listen(addr)
	},
}

func openBrowser(url string) {
	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		fmt.Printf("  (Could not automatically open browser: %v)\n", err)
	}
}

func init() {
	webCmd.Flags().IntVar(&webPort, "port", 8080, "Port to listen on")
	webCmd.Flags().StringVar(&webHostPrefix, "host-prefix", "", "Prefix added to scan paths (empty = paths used as-is)")
	rootCmd.AddCommand(webCmd)
}
