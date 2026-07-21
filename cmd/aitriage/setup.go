package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	rt "github.com/cybertortuga/aitriage/internal/runtime"
	"github.com/spf13/cobra"
)

var (
	setupFull    bool
	setupStatus  bool
	setupRepair  bool
	setupRemove  bool
	setupJSON    bool
	setupVerbose bool
)

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Install and verify the AITriage container scanner runtime (Docker)",
	Long: `Prepare this computer to run the full AITriage scanner bundle.

aitriage setup --full     Check Docker, download the compatible scanner image and
                          verify Semgrep, Trivy, Gitleaks and Bandit.
aitriage setup --status   Report the current runtime state without downloading.
aitriage setup --repair   Re-download the image and re-verify the bundle.
aitriage setup --remove-runtime  Remove only AITriage's downloaded image (keeps Docker).`,
	RunE: runSetup,
}

func init() {
	rootCmd.AddCommand(setupCmd)
	setupCmd.Flags().BoolVar(&setupFull, "full", false, "Install and verify the full scanner runtime")
	setupCmd.Flags().BoolVar(&setupStatus, "status", false, "Report runtime status without downloading")
	setupCmd.Flags().BoolVar(&setupRepair, "repair", false, "Re-download the image and re-verify the bundle")
	setupCmd.Flags().BoolVar(&setupRemove, "remove-runtime", false, "Remove AITriage's downloaded image (does not touch Docker or your projects)")
	setupCmd.Flags().BoolVar(&setupJSON, "json", false, "Emit a single machine-readable JSON document")
	setupCmd.Flags().BoolVar(&setupVerbose, "verbose", false, "Show raw technical details on failure")
}

func runSetup(cmd *cobra.Command, args []string) error {
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true
	ctx := cmd.Context()
	docker := rt.NewDocker()

	var report *rt.Report
	switch {
	case setupStatus:
		report = rt.Status(ctx, docker, Version)
	case setupRepair:
		report = rt.Setup(ctx, docker, Version, true) // force pull
	case setupRemove:
		return runSetupRemove(ctx, docker)
	default: // --full is the default action
		report = rt.Setup(ctx, docker, Version, false)
	}

	if setupJSON {
		return emitSetupJSON(report)
	}
	renderSetup(report)
	if !report.OK {
		return errSilent
	}
	return nil
}

// errSilent is the package-local alias for the shared silent-exit sentinel.
var errSilent = errSilentExit

func runSetupRemove(ctx context.Context, docker rt.Docker) error {
	image := rt.ResolveImage(Version)
	report := buildRemoveRuntimeReport(ctx, docker, image)

	if setupJSON {
		return emitSetupJSON(report)
	}
	renderSetup(report)
	if !report.OK {
		return errSilent
	}
	return nil
}

func buildRemoveRuntimeReport(ctx context.Context, docker rt.Docker, image string) *rt.Report {
	report := &rt.Report{Image: image}
	if setupErr := rt.Detect(ctx, docker); setupErr != nil {
		report.Err = setupErr
		report.Steps = append(report.Steps, rt.Step{Label: "Docker is accessible", Status: rt.StepAction})
	} else if !docker.ImageExists(ctx, image) {
		report.OK = true
		report.Steps = append(report.Steps, rt.Step{Label: "AITriage scanner image is already absent", Status: rt.StepOK})
		report.NextHint = "Docker, your projects and aitriage-reports were not changed."
	} else if err := docker.RemoveImage(ctx, image); err != nil {
		report.Err = &rt.SetupError{
			Code:         "runtime_remove_failed",
			Message:      "Could not remove the AITriage scanner image. No project files were touched.",
			RetryCommand: "aitriage setup --remove-runtime",
			Detail:       err.Error(),
		}
		report.Steps = append(report.Steps, rt.Step{Label: "AITriage scanner image removed", Status: rt.StepError})
	} else {
		report.OK = true
		report.Steps = append(report.Steps, rt.Step{Label: "AITriage scanner image removed", Status: rt.StepOK})
		report.NextHint = "Docker, your projects and aitriage-reports were not changed. Reinstall later with: aitriage setup --full"
	}
	return report
}

// ── rendering ────────────────────────────────────────────────────────────────

func renderSetup(r *rt.Report) {
	c := newSetupColors()
	if len(r.Steps) > 0 {
		fmt.Fprintln(os.Stderr, c.bold("AITriage full setup")+"\n")
		total := len(r.Steps)
		for i, s := range r.Steps {
			fmt.Fprintf(os.Stderr, "[%d/%d] %s  %s\n", i+1, total, c.status(s.Status), s.Label)
		}
		fmt.Fprintln(os.Stderr)
	}
	if r.Err != nil {
		fmt.Fprintf(os.Stderr, "%s: %s\n", c.status(errLabel(r.Err.Code)), r.Err.Message)
		if r.Err.ActionURL != "" {
			fmt.Fprintf(os.Stderr, "  %s\n", r.Err.ActionURL)
		}
		fmt.Fprintf(os.Stderr, "\nThen run:\n  %s\n", r.Err.RetryCommand)
		if setupVerbose && r.Err.Detail != "" {
			fmt.Fprintf(os.Stderr, "\n%s\n", c.muted(r.Err.Detail))
		}
		return
	}
	fmt.Fprintln(os.Stderr, c.ok("AITriage is ready."))
	if r.Digest != "" {
		fmt.Fprintf(os.Stderr, "  %s\n", c.muted("image "+r.Image+" @ "+r.Digest))
	}
	if r.NextHint != "" {
		fmt.Fprintf(os.Stderr, "\n%s\n", r.NextHint)
	}
}

func errLabel(code string) rt.StepStatus {
	switch code {
	case "docker_not_installed", "docker_not_running", "image_missing":
		return rt.StepAction
	default:
		return rt.StepError
	}
}

// setupOut is the stable --json contract for AI IDE / automation.
type setupOut struct {
	Status       string            `json:"status"` // ok | action_required | error
	Code         string            `json:"code,omitempty"`
	Message      string            `json:"message"`
	ActionURL    string            `json:"action_url,omitempty"`
	RetryCommand string            `json:"retry_command,omitempty"`
	Image        string            `json:"image,omitempty"`
	Digest       string            `json:"digest,omitempty"`
	Bundle       []rt.BundleStatus `json:"bundle,omitempty"`
}

// buildSetupOut maps a runtime Report to the stable JSON contract (pure, tested).
func buildSetupOut(r *rt.Report) setupOut {
	out := setupOut{Image: r.Image, Digest: r.Digest, Bundle: r.Bundle}
	if r.OK {
		out.Status = "ok"
		out.Message = "Full scanner runtime is ready."
		return out
	}
	if r.Err != nil {
		out.Code = r.Err.Code
		out.Message = r.Err.Message
		out.ActionURL = r.Err.ActionURL
		out.RetryCommand = r.Err.RetryCommand
		if errLabel(r.Err.Code) == rt.StepAction {
			out.Status = "action_required"
		} else {
			out.Status = "error"
		}
	}
	return out
}

func emitSetupJSON(r *rt.Report) error {
	out := buildSetupOut(r)
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(out); err != nil {
		return err
	}
	if !r.OK {
		return errSilent
	}
	return nil
}

// ── color handling (TTY only; NO_COLOR / TERM=dumb / redirected disable) ──────

type setupColors struct{ enabled bool }

func newSetupColors() setupColors {
	fi, err := os.Stderr.Stat()
	tty := err == nil && (fi.Mode()&os.ModeCharDevice) != 0
	return setupColors{enabled: setupColorsEnabled(os.Getenv("NO_COLOR"), os.Getenv("TERM"), tty)}
}

func setupColorsEnabled(noColor, term string, stderrTTY bool) bool {
	return stderrTTY && noColor == "" && term != "dumb"
}

func (c setupColors) wrap(code, s string) string {
	if !c.enabled {
		return s
	}
	return "\033[" + code + "m" + s + "\033[0m"
}
func (c setupColors) bold(s string) string  { return c.wrap("1", s) }
func (c setupColors) ok(s string) string    { return c.wrap("38;2;46;204;113", s) }
func (c setupColors) muted(s string) string { return c.wrap("38;2;132;148;149", s) }

// status returns the colored, always-present text prefix for a step status.
func (c setupColors) status(s rt.StepStatus) string {
	switch s {
	case rt.StepOK:
		return c.ok(string(s))
	case rt.StepCheck:
		return c.wrap("38;2;0;245;255", string(s))
	case rt.StepAction:
		return c.wrap("38;2;231;196;39", string(s))
	case rt.StepError:
		return c.wrap("38;2;231;76;60", string(s))
	default:
		return string(s)
	}
}
