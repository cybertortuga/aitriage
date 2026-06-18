package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"github.com/spf13/cobra"
)

// Version is set via -ldflags at build time (GoReleaser).
// Falls back to "dev" for local builds.
var Version = "dev"

// ErrPolicyViolation is returned when the security policy gate fails.
// Commands should return this instead of calling os.Exit(1) directly,
// so that defers run and the error is testable.
var ErrPolicyViolation = errors.New("security policy violation: see blocking reasons above")

var rootCmd = &cobra.Command{
	Use:   "aitriage [path]",
	Short: "AITriage — AI-Powered Security Audit Engine",
	Long: `AITriage is a deterministic security presence-checker with AI-powered triage.

It scans your project for missing security practices, hardcoded secrets,
misconfigured deployments, and architectural risks — then optionally uses
an LLM to prioritize findings and generate fix specifications.

Running "aitriage" without a subcommand is equivalent to "aitriage scan .".`,
	Example: `  aitriage                    # Scan current directory
  aitriage ./my-project       # Scan specific path
  aitriage scan --format json # Scan with JSON output
  aitriage agent              # AI-powered audit with LLM
  aitriage agent --no-chat    # Non-interactive (CI/CD)`,
	// When called without a subcommand, run scan on the given path (or .)
	RunE: func(cmd *cobra.Command, args []string) error {
		path := "."
		if len(args) > 0 {
			path = args[0]
		}
		// Call RunE directly instead of Execute() to avoid infinite recursion
		// through cobra's command chain (root → scan → root → ...)
		return scanCmd.RunE(scanCmd, []string{path})
	},
	Args:              cobra.MaximumNArgs(1),
	DisableAutoGenTag: true,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		if errors.Is(err, ErrPolicyViolation) {
			// Policy failure already printed by printPolicyFailure.
			os.Exit(1)
		}
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	// Root-level persistent flags can be added here
	lipgloss.SetColorProfile(termenv.TrueColor)
}
