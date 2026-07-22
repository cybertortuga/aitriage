// Package runtime is the single, transport-independent orchestrator for the
// AITriage container runtime. The host CLI does only orchestration here: detect
// Docker, resolve/pull a compatible image, verify the scanner bundle, and build
// the exact `docker run` arguments (mounts, env allowlist, security flags) used
// identically by CLI, MCP and Web. It never runs a reduced native scan in place
// of the full container scan.
//
// Everything is testable without a real Docker: the Docker dependency is an
// interface, and setup/verification/arg-building are pure over it.
package runtime

import (
	"context"
	"fmt"
	"runtime"
	"strings"
)

// ImageRepo is the release image lineage. The compatible tag is derived from the
// host CLI version; `latest` is only used via an explicit override.
const ImageRepo = "ghcr.io/cybertortuga/aitriage"

// RequiredTools is the scanner bundle the image must contain (plus aitriage).
var RequiredTools = []string{"aitriage", "semgrep", "trivy", "gitleaks", "bandit"}

// ── Docker abstraction ────────────────────────────────────────────────────────

// BundleStatus is the verification result for one tool inside the image.
type BundleStatus struct {
	Tool    string `json:"tool"`
	OK      bool   `json:"ok"`
	Version string `json:"version,omitempty"`
}

// Docker abstracts the docker CLI so setup is unit-testable with a fake.
type Docker interface {
	// Installed reports whether the docker binary is on PATH.
	Installed() bool
	// Info returns nil when the daemon is reachable; otherwise a classified error.
	Info(ctx context.Context) error
	// ImageExists reports whether the image is present locally.
	ImageExists(ctx context.Context, image string) bool
	// Pull downloads the image; the error message is classified by the caller.
	Pull(ctx context.Context, image string) error
	// Digest returns the resolved image digest, or "" if unknown.
	Digest(ctx context.Context, image string) string
	// RemoveImage deletes only the named AITriage image from the local cache.
	RemoveImage(ctx context.Context, image string) error
	// Warm prepares network-backed scanner databases/config without scanning a
	// user project. Status checks never call it.
	Warm(ctx context.Context, image string) error
	// Verify runs the bundle check inside the image and returns per-tool status.
	Verify(ctx context.Context, image string) ([]BundleStatus, error)
}

// ── Typed setup errors ────────────────────────────────────────────────────────

// SetupError is a user-actionable setup failure with a stable machine code, an
// official action URL and the exact command to retry. It never carries a stack
// trace or raw Docker stderr (those go to --verbose only).
type SetupError struct {
	Code         string `json:"code"`
	Message      string `json:"message"`
	ActionURL    string `json:"action_url,omitempty"`
	RetryCommand string `json:"retry_command"`
	// Detail is verbose-only technical context; never shown by default.
	Detail string `json:"-"`
}

func (e *SetupError) Error() string { return e.Message }

const retryFull = "aitriage setup --full"

// dockerInstallURL returns the official Docker install URL for the current OS.
func dockerInstallURL() string {
	switch runtime.GOOS {
	case "darwin":
		return "https://docs.docker.com/desktop/setup/install/mac-install/"
	case "windows":
		return "https://docs.docker.com/desktop/setup/install/windows-install/"
	default:
		return "https://docs.docker.com/engine/install/"
	}
}

func errDockerNotInstalled() *SetupError {
	msg := "Docker is not installed. AITriage uses Docker to run the complete scanner bundle safely and consistently."
	if runtime.GOOS == "windows" {
		msg += " If you just installed Docker Desktop, close and reopen your terminal or AI IDE so it picks up the new PATH, then run the same command again."
	}
	return &SetupError{
		Code:         "docker_not_installed",
		Message:      msg,
		ActionURL:    dockerInstallURL(),
		RetryCommand: retryFull,
	}
}

func errDockerNotRunning(detail string) *SetupError {
	msg := "Docker is installed but not running. Start Docker and wait until it reports running, then run the same command again."
	if runtime.GOOS == "linux" {
		msg = "Docker is installed but not running. Start the Docker service (see the official docs), then run the same command again."
	}
	return &SetupError{Code: "docker_not_running", Message: msg, ActionURL: dockerInstallURL(), RetryCommand: retryFull, Detail: detail}
}

func errDockerPermission(detail string) *SetupError {
	return &SetupError{
		Code:         "docker_permission_denied",
		Message:      "AITriage cannot access Docker: the current user lacks permission. Follow the official Docker post-installation instructions, then run the same command again.",
		ActionURL:    "https://docs.docker.com/engine/install/linux-postinstall/",
		RetryCommand: retryFull,
		Detail:       detail,
	}
}

// classifyPull maps a docker pull failure to a specific setup error code.
func classifyPull(image string, err error) *SetupError {
	detail := ""
	if err != nil {
		detail = err.Error()
	}
	low := strings.ToLower(detail)
	base := &SetupError{RetryCommand: retryFull, Detail: detail}
	switch {
	case containsAny(low, "no such host", "temporary failure", "network is unreachable", "dial tcp", "lookup"):
		base.Code = "image_pull_network"
		base.Message = "Could not download the AITriage scanner image: no network connection. Check your internet connection and run the same command again."
	case containsAny(low, "unauthorized", "denied", "authentication", "forbidden"):
		base.Code = "image_pull_denied"
		base.Message = "Could not download the AITriage scanner image: access was denied by the registry. Run the same command again after checking registry access."
	case containsAny(low, "no space", "disk", "insufficient"):
		base.Code = "image_pull_no_space"
		base.Message = "Could not download the AITriage scanner image: not enough disk space. Free some space and run the same command again."
	case containsAny(low, "timeout", "timed out", "deadline exceeded"):
		base.Code = "image_pull_timeout"
		base.Message = "Downloading the AITriage scanner image timed out. Run the same command again."
	case containsAny(low, "no matching manifest", "platform", "architecture"):
		base.Code = "image_pull_arch"
		base.Message = "The AITriage scanner image is not available for this computer's architecture. Please report this with your OS and CPU."
	default:
		base.Code = "image_pull_failed"
		base.Message = fmt.Sprintf("Could not download the AITriage scanner image %q. Run the same command again; use --verbose for details.", image)
	}
	return base
}

func containsAny(s string, subs ...string) bool {
	for _, sub := range subs {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}

// isPermissionErr / isDaemonDownErr classify a `docker info` failure.
func isPermissionErr(s string) bool {
	l := strings.ToLower(s)
	return containsAny(l, "permission denied", "got permission denied", "dial unix", "connect: permission denied")
}
