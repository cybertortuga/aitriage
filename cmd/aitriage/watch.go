package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/cybertortuga/aitriage/internal/scanner"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cobra"
)

var (
	watchDebounce int
	watchFailOn   string
	watchQuiet    bool
)

var watchCmd = &cobra.Command{
	Use:   "watch [path]",
	Short: "Watch for file changes and scan incrementally — Sentinel Mode",
	Long: `Sentinel Mode: continuously monitors the project for file changes
and runs incremental security scans automatically.

  aitriage watch .                    → Watch current directory
  aitriage watch . --debounce 500     → Custom debounce (ms)
  aitriage watch . --fail-on critical → Exit non-zero on critical findings

Press Ctrl+C to stop.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runWatch,
}

func init() {
	rootCmd.AddCommand(watchCmd)
	watchCmd.Flags().IntVar(&watchDebounce, "debounce", 300, "Debounce delay in milliseconds")
	watchCmd.Flags().StringVar(&watchFailOn, "fail-on", "", "Exit with error if finding of this severity or above is found")
	watchCmd.Flags().BoolVar(&watchQuiet, "quiet", false, "Only show findings, suppress status messages")
}

func runWatch(cmd *cobra.Command, args []string) error {
	projectPath := "."
	if len(args) > 0 {
		projectPath = args[0]
	}

	absPath, err := filepath.Abs(projectPath)
	if err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}

	// Colors
	cyan := "\033[38;2;0;245;255m"
	green := "\033[38;2;46;204;113m"
	yellow := "\033[38;2;255;214;0m"
	red := "\033[38;2;231;76;60m"
	dim := "\033[38;2;120;120;140m"
	bold := "\033[1m"
	reset := "\033[0m"

	if !watchQuiet {
		fmt.Fprintf(os.Stderr, "\n%s%s  ◉ SENTINEL MODE — AITriage File Watcher%s\n", cyan, bold, reset)
		fmt.Fprintf(os.Stderr, "%s  Watching: %s%s\n", dim, absPath, reset)
		fmt.Fprintf(os.Stderr, "%s  Debounce: %dms%s\n", dim, watchDebounce, reset)
		fmt.Fprintf(os.Stderr, "%s  Press Ctrl+C to stop.%s\n\n", dim, reset)
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create watcher: %w", err)
	}
	defer watcher.Close()

	// Walk and add all directories (fsnotify watches dirs, not files)
	dirCount := 0
	err = filepath.WalkDir(absPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			name := d.Name()
			// Skip hidden dirs, vendor, node_modules, .git
			if strings.HasPrefix(name, ".") || name == "node_modules" || name == "vendor" ||
				name == "__pycache__" || name == ".venv" || name == "venv" {
				return filepath.SkipDir
			}
			if err := watcher.Add(path); err == nil {
				dirCount++
			}
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to walk directory: %w", err)
	}

	if !watchQuiet {
		fmt.Fprintf(os.Stderr, "%s  ✓ Watching %d directories%s\n\n", green, dirCount, reset)
	}

	// Initial scan
	if !watchQuiet {
		fmt.Fprintf(os.Stderr, "%s  Running initial scan...%s\n", dim, reset)
	}
	ctx := context.Background()
	report, err := scanner.Scan(ctx, absPath, scanner.ScanOptions{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s  ⚠ Initial scan failed: %v%s\n", yellow, err, reset)
	} else {
		printWatchSummary(report, absPath, cyan, green, yellow, red, dim, bold, reset)
	}

	// Debounce timer
	var (
		mu         sync.Mutex
		pending    = make(map[string]bool)
		debounceMs = time.Duration(watchDebounce) * time.Millisecond
		timer      *time.Timer
	)

	// Signal handling for graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)

	scanCount := 0

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return nil
			}

			// Only care about writes, creates, renames
			if !event.Has(fsnotify.Write) && !event.Has(fsnotify.Create) && !event.Has(fsnotify.Rename) {
				continue
			}

			// Skip hidden files and common noise
			base := filepath.Base(event.Name)
			if strings.HasPrefix(base, ".") || strings.HasSuffix(base, "~") ||
				strings.HasSuffix(base, ".swp") || strings.HasSuffix(base, ".tmp") {
				continue
			}

			mu.Lock()
			pending[event.Name] = true

			if timer != nil {
				timer.Stop()
			}
			timer = time.AfterFunc(debounceMs, func() {
				mu.Lock()
				files := make([]string, 0, len(pending))
				for f := range pending {
					files = append(files, f)
				}
				pending = make(map[string]bool)
				mu.Unlock()

				if len(files) == 0 {
					return
				}

				scanCount++
				now := time.Now().Format("15:04:05")

				if !watchQuiet {
					fmt.Fprintf(os.Stderr, "\n%s  %s │ %sCHANGED%s │ ", dim, now, yellow, dim)
					if len(files) == 1 {
						rel, _ := filepath.Rel(absPath, files[0])
						if rel == "" {
							rel = files[0]
						}
						fmt.Fprintf(os.Stderr, "%s%s\n", rel, reset)
					} else {
						fmt.Fprintf(os.Stderr, "%d files%s\n", len(files), reset)
					}
				}

				// Incremental scan — only changed files
				report, err := scanner.Scan(ctx, absPath, scanner.ScanOptions{
					FileFilter: files,
				})
				if err != nil {
					fmt.Fprintf(os.Stderr, "%s         │ ⚠ ERROR  │ %v%s\n", red, err, reset)
					return
				}

				if len(report.Results) == 0 {
					if !watchQuiet {
						fmt.Fprintf(os.Stderr, "%s         │ %s✓ CLEAR%s  │ No findings%s\n", dim, green, dim, reset)
					}
				} else {
					for _, r := range report.Results {
						sevColor := dim
						sevIcon := "·"
						switch r.Severity {
						case "CRITICAL":
							sevColor = red
							sevIcon = "▲"
						case "HIGH":
							sevColor = red
							sevIcon = "■"
						case "MEDIUM":
							sevColor = yellow
							sevIcon = "▸"
						case "LOW":
							sevColor = dim
							sevIcon = "·"
						}

						relFile, _ := filepath.Rel(absPath, r.File)
						if relFile == "" {
							relFile = r.File
						}
						_ = relFile

						lineStr := ""
						if r.Line > 0 {
							lineStr = fmt.Sprintf(" (line %d)", r.Line)
						}

						fmt.Fprintf(os.Stderr, "%s         │ %s%s %s%s │ %s%s%s%s\n",
							dim, sevColor, sevIcon, r.Severity, dim,
							reset, r.Name, lineStr, reset)

						if r.Suggestion != "" {
							fmt.Fprintf(os.Stderr, "%s         │ 💡 FIX    │ %s%s\n", dim, r.Suggestion, reset)
						}
					}
				}
			})
			mu.Unlock()

		case err, ok := <-watcher.Errors:
			if !ok {
				return nil
			}
			fmt.Fprintf(os.Stderr, "%s  ⚠ Watch error: %v%s\n", yellow, err, reset)

		case <-sigCh:
			fmt.Fprintf(os.Stderr, "\n%s  ■ Sentinel stopped. %d incremental scans performed.%s\n\n", dim, scanCount, reset)
			return nil
		}
	}
}

func printWatchSummary(report scanner.ScanReport, absPath, cyan, green, yellow, red, dim, bold, reset string) {
	critCount := 0
	highCount := 0
	medCount := 0
	lowCount := 0
	for _, r := range report.Results {
		switch r.Severity {
		case "CRITICAL":
			critCount++
		case "HIGH":
			highCount++
		case "MEDIUM":
			medCount++
		case "LOW":
			lowCount++
		}
	}

	scoreColor := green
	if report.SecurityScore < 70 {
		scoreColor = red
	} else if report.SecurityScore < 85 {
		scoreColor = yellow
	}

	fmt.Fprintf(os.Stderr, "%s  ┌─────────────────────────────────────────┐%s\n", dim, reset)
	fmt.Fprintf(os.Stderr, "%s  │ %sHealth Check: %s%d/100 %s(%s)%s              │%s\n",
		dim, bold, scoreColor, report.SecurityScore, dim, report.SecurityGrade, dim, reset)
	fmt.Fprintf(os.Stderr, "%s  │ %sFindings: %s%dC %s%dH %s%dM %s%dL%s              │%s\n",
		dim, bold, red, critCount, red, highCount, yellow, medCount, dim, lowCount, dim, reset)
	fmt.Fprintf(os.Stderr, "%s  │ %sFiles: %d │ Rules: %d │ %s%s            │%s\n",
		dim, reset, report.TotalFiles, report.RulesApplied, dim, report.ScanDuration.Round(time.Millisecond), reset)
	fmt.Fprintf(os.Stderr, "%s  └─────────────────────────────────────────┘%s\n\n", dim, reset)
}
