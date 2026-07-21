package runtime

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// execDocker is the real Docker implementation backed by the docker CLI.
type execDocker struct{}

// NewDocker returns the real docker-CLI-backed Docker.
func NewDocker() Docker { return execDocker{} }

func (execDocker) Installed() bool {
	_, err := exec.LookPath("docker")
	return err == nil
}

func (execDocker) Info(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "docker", "info", "--format", "{{.ServerVersion}}")
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
	return exec.CommandContext(ctx, "docker", "image", "inspect", image).Run() == nil
}

func (execDocker) Pull(ctx context.Context, image string) error {
	cmd := exec.CommandContext(ctx, "docker", "pull", image)
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
	out, err := exec.CommandContext(ctx, "docker", "image", "inspect",
		"--format", "{{index .RepoDigests 0}}", image).Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func (execDocker) RemoveImage(ctx context.Context, image string) error {
	cmd := exec.CommandContext(ctx, "docker", "image", "rm", image)
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
	return exec.CommandContext(ctx, "docker", args...).CombinedOutput()
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
	out, err := exec.CommandContext(ctx, "docker", args...).CombinedOutput()
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
