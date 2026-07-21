package runtime

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// fakeDocker is a scriptable Docker for tests — no real Docker required.
type fakeDocker struct {
	installed bool
	infoErr   error
	exists    bool
	pullErr   error
	pulled    int
	removed   int
	removeErr error
	warmed    int
	warmErr   error
	bundle    []BundleStatus
	verifyErr error
}

func (f *fakeDocker) Installed() bool                           { return f.installed }
func (f *fakeDocker) Info(context.Context) error                { return f.infoErr }
func (f *fakeDocker) ImageExists(context.Context, string) bool  { return f.exists }
func (f *fakeDocker) Pull(context.Context, string) error        { f.pulled++; return f.pullErr }
func (f *fakeDocker) Digest(context.Context, string) string     { return "sha256:deadbeef" }
func (f *fakeDocker) RemoveImage(context.Context, string) error { f.removed++; return f.removeErr }
func (f *fakeDocker) Warm(context.Context, string) error        { f.warmed++; return f.warmErr }
func (f *fakeDocker) Verify(context.Context, string) ([]BundleStatus, error) {
	return f.bundle, f.verifyErr
}

func fullBundle() []BundleStatus {
	return []BundleStatus{
		{"aitriage", true, "aitriage version 1.7.0"}, {"semgrep", true, "1.2"}, {"trivy", true, "0.5"},
		{"gitleaks", true, "8.1"}, {"bandit", true, "1.7"},
	}
}

func TestDetect(t *testing.T) {
	ctx := context.Background()
	if e := Detect(ctx, &fakeDocker{installed: false}); e == nil || e.Code != "docker_not_installed" {
		t.Fatalf("not installed: %+v", e)
	}
	if e := Detect(ctx, &fakeDocker{installed: true, infoErr: errors.New("Cannot connect to the Docker daemon")}); e == nil || e.Code != "docker_not_running" {
		t.Fatalf("daemon down: %+v", e)
	}
	if e := Detect(ctx, &fakeDocker{installed: true, infoErr: errors.New("permission denied while trying to connect")}); e == nil || e.Code != "docker_permission_denied" {
		t.Fatalf("permission: %+v", e)
	}
	if e := Detect(ctx, &fakeDocker{installed: true}); e != nil {
		t.Fatalf("healthy docker should detect nil, got %+v", e)
	}
}

func TestSetupPassAndIdempotent(t *testing.T) {
	ctx := context.Background()
	// First setup: image absent -> pull once -> verify OK.
	f := &fakeDocker{installed: true, exists: false, bundle: fullBundle()}
	r := Setup(ctx, f, "1.7.0", false)
	if !r.OK || r.Err != nil {
		t.Fatalf("setup should pass: %+v", r.Err)
	}
	if f.pulled != 1 {
		t.Fatalf("expected exactly one pull, got %d", f.pulled)
	}
	if f.warmed != 1 {
		t.Fatalf("expected scanner preparation once, got %d", f.warmed)
	}
	if r.Image != "ghcr.io/cybertortuga/aitriage:v1.7.0" {
		t.Fatalf("image = %q", r.Image)
	}
	if r.Digest == "" {
		t.Fatal("digest should be recorded")
	}

	// Second setup: image present -> no re-pull.
	f2 := &fakeDocker{installed: true, exists: true, bundle: fullBundle()}
	r2 := Setup(ctx, f2, "1.7.0", false)
	if !r2.OK || f2.pulled != 0 {
		t.Fatalf("idempotent setup must not re-pull: pulled=%d ok=%v", f2.pulled, r2.OK)
	}
}

func TestStatusIsReadOnlyAndDoesNotWarmOrPull(t *testing.T) {
	f := &fakeDocker{installed: true, exists: true, bundle: fullBundle()}
	r := Status(context.Background(), f, "1.7.0")
	if !r.OK {
		t.Fatalf("status should pass: %+v", r.Err)
	}
	if f.pulled != 0 || f.warmed != 0 {
		t.Fatalf("read-only status mutated runtime: pulled=%d warmed=%d", f.pulled, f.warmed)
	}
}

func TestSetupPullFailureClassified(t *testing.T) {
	ctx := context.Background()
	cases := map[string]string{
		"dial tcp: lookup ghcr.io: no such host": "image_pull_network",
		"unauthorized: authentication required":  "image_pull_denied",
		"no space left on device":                "image_pull_no_space",
		"context deadline exceeded":              "image_pull_timeout",
		"no matching manifest for linux/arm64":   "image_pull_arch",
	}
	for msg, code := range cases {
		f := &fakeDocker{installed: true, exists: false, pullErr: errors.New(msg)}
		r := Setup(ctx, f, "1.7.0", false)
		if r.OK || r.Err == nil || r.Err.Code != code {
			t.Errorf("pull %q -> code %v, want %s", msg, r.Err, code)
		}
		if r.Err != nil && r.Err.RetryCommand == "" {
			t.Errorf("pull error must carry a retry command")
		}
	}
}

func TestSetupBundleIncomplete(t *testing.T) {
	ctx := context.Background()
	partial := []BundleStatus{
		{"aitriage", true, "1"}, {"semgrep", true, "1"}, {"trivy", false, ""},
		{"gitleaks", true, "8"}, {"bandit", true, "1"},
	}
	f := &fakeDocker{installed: true, exists: true, bundle: partial}
	r := Setup(ctx, f, "1.7.0", false)
	if r.OK || r.Err == nil || r.Err.Code != "bundle_incomplete" {
		t.Fatalf("incomplete bundle must fail: %+v", r.Err)
	}
	if !strings.Contains(r.Err.Message, "trivy") {
		t.Errorf("message should name the unhealthy tool: %s", r.Err.Message)
	}
}

func TestBundleVersionCompatibility(t *testing.T) {
	matching := fullBundle()
	matching[0].Version = "aitriage version v1.7.0"
	if !bundleVersionCompatible("1.7.0", matching) {
		t.Fatal("matching release versions must pass")
	}
	mismatch := fullBundle()
	mismatch[0].Version = "aitriage version 1.6.9"
	if bundleVersionCompatible("v1.7.0", mismatch) {
		t.Fatal("mismatched release image must fail")
	}
	if !bundleVersionCompatible("dev", mismatch) {
		t.Fatal("development host intentionally accepts a local development image")
	}
}

func TestStatusPreservesBundleVerificationFailure(t *testing.T) {
	f := &fakeDocker{installed: true, exists: true, verifyErr: errors.New("container could not start")}
	r := Status(context.Background(), f, "dev")
	if r.OK || r.Err == nil || r.Err.Code != "bundle_verify_failed" {
		t.Fatalf("verification error was masked: %+v", r)
	}
}

func TestResolveImage(t *testing.T) {
	t.Setenv("AITRIAGE_IMAGE", "")
	if got := ResolveImage("1.7.0"); got != "ghcr.io/cybertortuga/aitriage:v1.7.0" {
		t.Errorf("semver pin = %q", got)
	}
	if got := ResolveImage("dev"); got != "ghcr.io/cybertortuga/aitriage:v1" {
		t.Errorf("dev fallback = %q", got)
	}
	t.Setenv("AITRIAGE_IMAGE", "local/custom:latest")
	if got := ResolveImage("1.7.0"); got != "local/custom:latest" {
		t.Errorf("override = %q", got)
	}
}

func TestEnsureScannerCacheDirOutsideProject(t *testing.T) {
	cache := filepath.Join(t.TempDir(), "machine-cache")
	t.Setenv("AITRIAGE_SCANNER_CACHE_DIR", cache)
	got, err := EnsureScannerCacheDir()
	if err != nil {
		t.Fatal(err)
	}
	if got != cache {
		t.Fatalf("cache = %q, want %q", got, cache)
	}
	info, err := os.Stat(cache)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o700 {
		t.Errorf("cache mode = %o, want 700", info.Mode().Perm())
	}
}

func TestDockerRunArgsSecurity(t *testing.T) {
	t.Setenv("GEMINI_API_KEY", "super-secret-value")
	args := DockerRunArgs(RunSpec{
		Image:       "img",
		User:        "501:20",
		HostRoot:    "/host/repo",
		ReportsDir:  "/host/repo/aitriage-reports",
		CacheDir:    "/host/cache",
		Argv:        []string{"serve", "--profile", "safe"},
		Interactive: true,
		TTY:         false,
		EnvPassed:   []string{"GEMINI_API_KEY"},
	})
	joined := strings.Join(args, " ")
	for _, want := range []string{
		"--user 501:20",
		"-e HOME=/home/aitriage/.cache",
		"--security-opt no-new-privileges",
		"--cap-drop ALL",
		"/host/repo:/workspace:ro",
		"/host/repo/aitriage-reports:/workspace/aitriage-reports:rw",
		"/host/cache:/home/aitriage/.cache:rw",
	} {
		if !strings.Contains(joined, want) {
			t.Errorf("run args missing %q:\n%s", want, joined)
		}
	}
	if strings.Contains(joined, " -t ") || strings.HasSuffix(joined, " -t") {
		t.Error("MCP/non-interactive must not allocate a TTY")
	}
	if strings.Contains(joined, "docker.sock") || strings.Contains(joined, "--privileged") {
		t.Error("must never mount the docker socket or use --privileged")
	}
	if !strings.Contains(joined, "-e GEMINI_API_KEY") || strings.Contains(joined, "super-secret-value") {
		t.Fatalf("env must be forwarded by name without exposing its value: %s", joined)
	}
}

func TestContainerScanRoot(t *testing.T) {
	if got, err := ContainerScanRoot("/host/repo", "."); err != nil || got != "/workspace" {
		t.Fatalf("root = %q err=%v", got, err)
	}
	if got, err := ContainerScanRoot("/host/repo", "synthetic/app"); err != nil || got != "/workspace/synthetic/app" {
		t.Fatalf("nested = %q err=%v", got, err)
	}
	if _, err := ContainerScanRoot("/host/repo", "../escape"); err == nil {
		t.Fatal("escape above root must be rejected")
	}
	if _, err := ContainerScanRoot("/host/repo", "/etc/passwd"); err == nil {
		t.Fatal("absolute path outside root must be rejected")
	}
}

func TestContainerScanRootRejectsSymlinkEscape(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	link := filepath.Join(root, "escape")
	if err := os.Symlink(outside, link); err != nil {
		t.Fatal(err)
	}
	if _, err := ContainerScanRoot(root, link); err == nil {
		t.Fatal("symlink escape above root must be rejected")
	}
}

func TestContainerScanRootPreservesSpacesAndUnicode(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "service with spaces", "данные")
	if err := os.MkdirAll(target, 0o755); err != nil {
		t.Fatal(err)
	}
	got, err := ContainerScanRoot(root, target)
	if err != nil {
		t.Fatal(err)
	}
	if got != "/workspace/service with spaces/данные" {
		t.Fatalf("container path = %q", got)
	}
}
