package runtime

import (
	"context"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
)

// ResolveImage returns the container image reference for a host CLI version.
// Precedence: explicit AITRIAGE_IMAGE override > version-pinned release tag >
// dev "v1" lineage. `latest` is used only via the explicit override.
func ResolveImage(version string) string {
	if img := strings.TrimSpace(os.Getenv("AITRIAGE_IMAGE")); img != "" {
		return img
	}
	v := strings.TrimSpace(version)
	v = strings.TrimPrefix(v, "v")
	// A clean semantic version pins to that release; anything else (dev builds)
	// uses the rolling major lineage.
	if isSemver(v) {
		return ImageRepo + ":v" + v
	}
	return ImageRepo + ":v1"
}

func isSemver(v string) bool {
	parts := strings.Split(v, ".")
	if len(parts) != 3 {
		return false
	}
	for _, p := range parts {
		if p == "" {
			return false
		}
		for _, r := range p {
			if r < '0' || r > '9' {
				return false
			}
		}
	}
	return true
}

// ScannerCacheDir returns the machine-wide cache used by containerized scanner
// databases and registries. It is runtime state, not an audit artifact, so it
// lives outside every checked repository. Tests and managed environments may
// override it with AITRIAGE_SCANNER_CACHE_DIR.
func ScannerCacheDir() (string, error) {
	if override := strings.TrimSpace(os.Getenv("AITRIAGE_SCANNER_CACHE_DIR")); override != "" {
		abs, err := filepath.Abs(override)
		if err != nil {
			return "", fmt.Errorf("resolve scanner cache: %w", err)
		}
		return abs, nil
	}
	base, err := os.UserCacheDir()
	if err != nil {
		return "", fmt.Errorf("resolve user cache directory: %w", err)
	}
	return filepath.Join(base, "aitriage", "scanners"), nil
}

// EnsureScannerCacheDir creates the cache with owner-only permissions.
func EnsureScannerCacheDir() (string, error) {
	dir, err := ScannerCacheDir()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", fmt.Errorf("create scanner cache: %w", err)
	}
	if err := os.Chmod(dir, 0o700); err != nil {
		return "", fmt.Errorf("secure scanner cache: %w", err)
	}
	return dir, nil
}

// HostUser returns the numeric uid:gid used for bind-mounted container runs.
// Matching the caller prevents root-owned artifacts and makes owner-only report
// and cache directories writable on native Linux Docker hosts.
func HostUser() string {
	current, err := user.Current()
	if err != nil {
		return ""
	}
	if _, err := strconv.ParseUint(current.Uid, 10, 32); err != nil {
		return ""
	}
	if _, err := strconv.ParseUint(current.Gid, 10, 32); err != nil {
		return ""
	}
	return current.Uid + ":" + current.Gid
}

// Detect classifies Docker availability into an actionable setup error, or nil.
func Detect(ctx context.Context, d Docker) *SetupError {
	if !d.Installed() {
		return errDockerNotInstalled()
	}
	if err := d.Info(ctx); err != nil {
		if isPermissionErr(err.Error()) {
			return errDockerPermission(err.Error())
		}
		return errDockerNotRunning(err.Error())
	}
	return nil
}

// StepStatus is the outcome label for a setup step.
type StepStatus string

const (
	StepOK     StepStatus = "OK"
	StepCheck  StepStatus = "CHECK"
	StepAction StepStatus = "ACTION REQUIRED"
	StepError  StepStatus = "ERROR"
)

// Step is one line of setup progress.
type Step struct {
	Label  string     `json:"label"`
	Status StepStatus `json:"status"`
}

// Report is the full result of a setup/status/repair run.
type Report struct {
	OK       bool           `json:"ok"`
	Image    string         `json:"image,omitempty"`
	Digest   string         `json:"digest,omitempty"`
	Bundle   []BundleStatus `json:"bundle,omitempty"`
	Steps    []Step         `json:"-"`
	Err      *SetupError    `json:"error,omitempty"`
	NextHint string         `json:"next_hint,omitempty"`
}

// Setup runs the full machine setup: detect Docker, resolve/pull the compatible
// image, warm scanner databases, and verify the scanner bundle. It is
// idempotent — a present image is not re-pulled unless forcePull is true.
func Setup(ctx context.Context, d Docker, version string, forcePull bool) *Report {
	r := &Report{}
	if serr := Detect(ctx, d); serr != nil {
		r.Err = serr
		if serr.Code == "docker_not_installed" || serr.Code == "docker_not_running" {
			r.Steps = append(r.Steps, Step{"Docker is installed and running", StepAction})
		} else {
			r.Steps = append(r.Steps, Step{"Docker is accessible", StepError})
		}
		return r
	}
	r.Steps = append(r.Steps, Step{"Docker is installed", StepCheck}, Step{"Docker is running", StepCheck})

	image := ResolveImage(version)
	r.Image = image
	if forcePull || !d.ImageExists(ctx, image) {
		if err := d.Pull(ctx, image); err != nil {
			r.Err = classifyPull(image, err)
			r.Steps = append(r.Steps, Step{"AITriage scanner image downloaded", StepError})
			return r
		}
		r.Steps = append(r.Steps, Step{"AITriage scanner image downloaded", StepOK})
	} else {
		r.Steps = append(r.Steps, Step{"AITriage scanner image already present", StepOK})
	}
	r.Digest = d.Digest(ctx, image)
	if err := d.Warm(ctx, image); err != nil {
		r.Err = &SetupError{Code: "bundle_prepare_failed", Message: "Could not prepare the scanner databases and registry configuration. Check the network connection, then repair: aitriage setup --repair", RetryCommand: "aitriage setup --repair", Detail: err.Error()}
		r.Steps = append(r.Steps, Step{"Scanner databases and registry configuration prepared", StepError})
		return r
	}
	r.Steps = append(r.Steps, Step{"Scanner databases and registry configuration prepared", StepOK})

	bundle, err := d.Verify(ctx, image)
	if err != nil {
		r.Err = &SetupError{Code: "bundle_verify_failed", Message: "Could not verify the AITriage scanner package. Repair the installation: aitriage setup --repair", RetryCommand: "aitriage setup --repair", Detail: err.Error()}
		r.Steps = append(r.Steps, Step{"Scanner bundle verified", StepError})
		return r
	}
	r.Bundle = bundle
	if missing := unhealthyTools(bundle); len(missing) > 0 {
		r.Err = &SetupError{
			Code:         "bundle_incomplete",
			Message:      "The AITriage scanner package is incomplete. Missing or unhealthy: " + strings.Join(missing, ", ") + ". AITriage will not run a partial security audit. Repair: aitriage setup --repair",
			RetryCommand: "aitriage setup --repair",
		}
		r.Steps = append(r.Steps, Step{"Scanner bundle verified", StepError})
		return r
	}
	if !bundleVersionCompatible(version, bundle) {
		r.Err = &SetupError{
			Code:         "bundle_version_mismatch",
			Message:      "The downloaded AITriage scanner image does not match this CLI version. Repair the installation: aitriage setup --repair",
			RetryCommand: "aitriage setup --repair",
		}
		r.Steps = append(r.Steps, Step{"AITriage CLI and scanner image versions match", StepError})
		return r
	}
	r.Steps = append(r.Steps,
		Step{"Semgrep, Trivy, Gitleaks and Bandit verified", StepOK},
		Step{"Full scanner runtime is ready", StepOK})
	r.OK = true
	r.NextHint = "Connect a project:\n  Codex:       aitriage install-codex .\n  Claude Code: aitriage install-claude-code .\nOr start the Web UI:\n  aitriage web"
	return r
}

// Status reports the current runtime state WITHOUT pulling: Docker health, then
// whether the compatible image is present and its bundle healthy.
func Status(ctx context.Context, d Docker, version string) *Report {
	r := &Report{}
	if serr := Detect(ctx, d); serr != nil {
		r.Err = serr
		r.Steps = append(r.Steps, Step{"Docker is accessible", StepAction})
		return r
	}
	image := ResolveImage(version)
	r.Image = image
	if !d.ImageExists(ctx, image) {
		r.Err = &SetupError{Code: "image_missing", Message: "The AITriage scanner image is not downloaded yet. Run: aitriage setup --full", RetryCommand: retryFull}
		r.Steps = append(r.Steps, Step{"Scanner image downloaded", StepAction})
		return r
	}
	r.Digest = d.Digest(ctx, image)
	bundle, err := d.Verify(ctx, image)
	if err != nil {
		r.Err = &SetupError{Code: "bundle_verify_failed", Message: "Could not verify the AITriage scanner package. Repair the installation: aitriage setup --repair", RetryCommand: "aitriage setup --repair", Detail: err.Error()}
		r.Steps = append(r.Steps, Step{"Scanner bundle healthy", StepError})
		return r
	}
	r.Bundle = bundle
	if len(unhealthyTools(bundle)) > 0 {
		r.Err = &SetupError{Code: "bundle_incomplete", Message: "The scanner package is incomplete. Repair: aitriage setup --repair", RetryCommand: "aitriage setup --repair"}
		r.Steps = append(r.Steps, Step{"Scanner bundle healthy", StepError})
		return r
	}
	if !bundleVersionCompatible(version, bundle) {
		r.Err = &SetupError{Code: "bundle_version_mismatch", Message: "The installed AITriage scanner image does not match this CLI version. Repair the installation: aitriage setup --repair", RetryCommand: "aitriage setup --repair"}
		r.Steps = append(r.Steps, Step{"AITriage CLI and scanner image versions match", StepError})
		return r
	}
	r.OK = true
	r.Steps = append(r.Steps, Step{"Full scanner runtime is ready", StepOK})
	return r
}

// bundleVersionCompatible enforces exact release parity. Development binaries
// intentionally accept development images because they share the rolling v1
// tag and can additionally use AITRIAGE_IMAGE for local E2E.
func bundleVersionCompatible(hostVersion string, bundle []BundleStatus) bool {
	host := strings.TrimPrefix(strings.TrimSpace(hostVersion), "v")
	if !isSemver(host) {
		return true
	}
	for _, tool := range bundle {
		if tool.Tool != "aitriage" || !tool.OK {
			continue
		}
		fields := strings.Fields(strings.TrimSpace(tool.Version))
		for _, field := range fields {
			candidate := strings.TrimPrefix(strings.Trim(field, "v,;()[]"), "v")
			if isSemver(candidate) {
				return candidate == host
			}
		}
		return false
	}
	return false
}

// unhealthyTools returns the required tools that are absent or unhealthy in the
// verified bundle.
func unhealthyTools(bundle []BundleStatus) []string {
	ok := map[string]bool{}
	for _, b := range bundle {
		if b.OK {
			ok[b.Tool] = true
		}
	}
	var missing []string
	for _, t := range RequiredTools {
		if !ok[t] {
			missing = append(missing, t)
		}
	}
	return missing
}

// ── docker run argument construction (shared by CLI / MCP / Web) ──────────────

// RunSpec describes one container invocation.
type RunSpec struct {
	Image       string
	Name        string   // optional Docker container name (used for managed Web lifecycle)
	User        string   // numeric uid:gid; empty preserves the image USER
	HostRoot    string   // absolute host repository root, mounted read-only at /workspace
	ReportsDir  string   // absolute host <root>/aitriage-reports, mounted read-write
	CacheDir    string   // absolute host cache dir, mounted read-write
	Argv        []string // the aitriage command + args to run inside the container
	EnvPassed   []string // names of env vars to forward (values pulled from os.Env by caller)
	EnvSet      []string // explicit KEY=VALUE env set inside the container
	Ports       []string // published port mappings, e.g. "8080:8080" (Web only)
	TTY         bool     // allocate a TTY (interactive CLI only; never MCP/CI)
	Interactive bool     // keep stdin open (-i)
}

const containerWorkspace = "/workspace"
const containerReports = "/workspace/aitriage-reports"
const containerCache = "/home/aitriage/.cache"

// DockerRunArgs builds the exact `docker run` argument vector. Security posture:
// source read-only, reports read-write, no docker socket, no --privileged,
// no-new-privileges, and only an explicit env allowlist is forwarded.
func DockerRunArgs(spec RunSpec) []string {
	args := []string{"run", "--rm"}
	if spec.Name != "" {
		args = append(args, "--name", spec.Name)
	}
	if spec.User != "" {
		args = append(args, "--user", spec.User)
	}
	if spec.Interactive {
		args = append(args, "-i")
	}
	if spec.TTY {
		args = append(args, "-t")
	}
	args = append(args,
		"--security-opt", "no-new-privileges",
		"--cap-drop", "ALL",
		"-v", spec.HostRoot+":"+containerWorkspace+":ro",
		"-v", spec.ReportsDir+":"+containerReports+":rw",
		"-w", containerWorkspace,
	)
	if spec.CacheDir != "" {
		args = append(args, "-v", spec.CacheDir+":"+containerCache+":rw")
	}
	if spec.User != "" {
		// Docker defaults HOME to / for a numeric uid unknown to the image.
		// Keep scanner caches on the dedicated writable mount instead.
		args = append(args, "-e", "HOME=/home/aitriage/.cache")
	}
	for _, p := range spec.Ports {
		args = append(args, "-p", p)
	}
	for _, name := range spec.EnvPassed {
		// Pass only the variable name. Docker reads its value from the docker
		// process environment, so secrets never appear in argv or process lists.
		args = append(args, "-e", name)
	}
	for _, kv := range spec.EnvSet {
		args = append(args, "-e", kv)
	}
	args = append(args, spec.Image)
	args = append(args, spec.Argv...)
	return args
}

// ContainerScanRoot maps a host path (repository root or a nested subfolder) to
// its path inside the container workspace. It rejects any path that escapes the
// mounted root.
func ContainerScanRoot(hostRoot, target string) (string, error) {
	hr, err := filepath.Abs(hostRoot)
	if err != nil {
		return "", err
	}
	if resolved, resolveErr := filepath.EvalSymlinks(hr); resolveErr == nil {
		hr = resolved
	}
	tp := target
	if !filepath.IsAbs(tp) {
		tp = filepath.Join(hr, tp)
	}
	tp = filepath.Clean(tp)
	if resolved, resolveErr := filepath.EvalSymlinks(tp); resolveErr == nil {
		tp = resolved
	}
	rel, err := filepath.Rel(hr, tp)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", &SetupError{Code: "path_escape", Message: "the requested path is outside the connected repository root", RetryCommand: retryFull}
	}
	if rel == "." {
		return containerWorkspace, nil
	}
	return containerWorkspace + "/" + filepath.ToSlash(rel), nil
}
