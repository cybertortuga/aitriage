package runtime

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// execDocker is the real Docker implementation backed by the docker CLI.
type execDocker struct{}

// NewDocker returns the real docker-CLI-backed Docker.
func NewDocker() Docker { return execDocker{} }

// dockerBin resolves the docker executable. Normally it is on PATH. On Windows,
// a freshly installed Docker Desktop puts docker.exe on the system PATH, but a
// terminal/AI IDE started before that install has a stale PATH; in that case we
// fall back to Docker Desktop's official install location ONLY. We never search
// the current project or arbitrary writable directories for a docker binary.
func dockerBin() string {
	if p, err := exec.LookPath("docker"); err == nil {
		return p
	}
	if runtime.GOOS == "windows" {
		if p := firstExistingFile(windowsDockerCandidates()); p != "" {
			return p
		}
	}
	return "docker"
}

// DockerExecutable returns the trusted Docker CLI path used by all host-side
// container launches. Keeping this resolution in one place prevents setup from
// succeeding via Docker Desktop's standard Windows location while MCP, Web or
// agent launch later fails because an older terminal has a stale PATH.
func DockerExecutable() string { return dockerBin() }

// windowsDockerCandidates lists the official Docker Desktop docker.exe locations
// under the trusted Program Files roots. Order is deterministic; only absolute
// system paths are considered.
func windowsDockerCandidates() []string {
	var out []string
	seen := map[string]bool{}
	for _, env := range []string{"ProgramFiles", "ProgramW6432", "ProgramFiles(x86)"} {
		root := strings.TrimSpace(os.Getenv(env))
		if root == "" || seen[root] {
			continue
		}
		seen[root] = true
		out = append(out, filepath.Join(root, "Docker", "Docker", "resources", "bin", "docker.exe"))
	}
	return out
}

// firstExistingFile returns the first path that exists and is a regular file.
func firstExistingFile(paths []string) string {
	for _, p := range paths {
		if fi, err := os.Stat(p); err == nil && fi.Mode().IsRegular() {
			return p
		}
	}
	return ""
}

func (execDocker) Installed() bool {
	if _, err := exec.LookPath("docker"); err == nil {
		return true
	}
	if runtime.GOOS == "windows" {
		return firstExistingFile(windowsDockerCandidates()) != ""
	}
	return false
}

func (execDocker) Info(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, dockerBin(), "info", "--format", "{{.ServerVersion}}")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		if stderr.Len() > 0 {
			return fmt.Errorf("%s", strings.TrimSpace(stderr.String()))
		}
		return err
	}
	return nil
}

func (execDocker) ImageExists(ctx context.Context, image string) bool {
	return exec.CommandContext(ctx, dockerBin(), "image", "inspect", image).Run() == nil
}

func (execDocker) Pull(ctx context.Context, image string) error {
	cmd := exec.CommandContext(ctx, dockerBin(), "pull", image)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		if stderr.Len() > 0 {
			return fmt.Errorf("%s", strings.TrimSpace(stderr.String()))
		}
		return err
	}
	return nil
}

func (execDocker) Digest(ctx context.Context, image string) string {
	out, err := exec.CommandContext(ctx, dockerBin(), "image", "inspect",
		"--format", "{{index .RepoDigests 0}}", image).Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func (execDocker) RemoveImage(ctx context.Context, image string) error {
	cmd := exec.CommandContext(ctx, dockerBin(), "image", "rm", image)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		if stderr.Len() > 0 {
			return fmt.Errorf("%s", strings.TrimSpace(stderr.String()))
		}
		return err
	}
	return nil
}

const warmScannerScript = `set -eu
trivy image --download-db-only --quiet >/dev/null
semgrep --config auto --validate >/dev/null 2>&1`

// bundleCheckScript runs each version probe and prints a parseable status. It
// is network-free, so `setup --status` stays read-only.
const bundleCheckScript = `for t in aitriage:version semgrep:--version trivy:--version gitleaks:version bandit:--version; do
  name="${t%%:*}"; arg="${t#*:}";
  if v="$("$name" "$arg" 2>/dev/null | head -n1)"; then echo "$name	ok	$v"; else echo "$name	missing	"; fi
done`

func warmRuntimeContainerCommand(ctx context.Context, image, script string) ([]byte, error) {
	cache, err := EnsureScannerCacheDir()
	if err != nil {
		return nil, err
	}
	args := []string{"run", "--rm"}
	if hostUser := HostUser(); hostUser != "" {
		args = append(args, "--user", hostUser)
	}
	args = append(args, "-v", cache+":"+containerCache+":rw", "-e", "HOME=/home/aitriage/.cache", "--entrypoint", "sh", image, "-c", script)
	return exec.CommandContext(ctx, dockerBin(), args...).CombinedOutput()
}

func (execDocker) Warm(ctx context.Context, image string) error {
	out, err := warmRuntimeContainerCommand(ctx, image, warmScannerScript)
	if err == nil {
		return nil
	}
	detail := strings.TrimSpace(string(out))
	if len(detail) > 2000 {
		detail = detail[:2000] + "…"
	}
	if detail == "" {
		detail = err.Error()
	}
	return fmt.Errorf("scanner preparation failed: %s", detail)
}

func (execDocker) Verify(ctx context.Context, image string) ([]BundleStatus, error) {
	// Verification is deliberately network-free and mount-free: `setup
	// --status` must not create or update scanner caches.
	args := []string{"run", "--rm", "--security-opt", "no-new-privileges", "--cap-drop", "ALL", "--entrypoint", "sh", image, "-c", bundleCheckScript}
	out, err := exec.CommandContext(ctx, dockerBin(), args...).CombinedOutput()
	if err != nil {
		detail := strings.TrimSpace(string(out))
		if len(detail) > 2000 {
			detail = detail[:2000] + "…"
		}
		if detail == "" {
			detail = err.Error()
		}
		return nil, fmt.Errorf("container bundle check failed: %s", detail)
	}
	return parseBundleOutput(string(out)), nil
}

// parseBundleOutput parses the tab-separated bundle-check output.
func parseBundleOutput(out string) []BundleStatus {
	var res []BundleStatus
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		if line == "" {
			continue
		}
		f := strings.SplitN(line, "\t", 3)
		if len(f) < 2 {
			continue
		}
		b := BundleStatus{Tool: strings.TrimSpace(f[0]), OK: strings.TrimSpace(f[1]) == "ok"}
		if len(f) == 3 {
			b.Version = strings.TrimSpace(f[2])
		}
		res = append(res, b)
	}
	return res
}
