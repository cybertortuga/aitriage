//go:build windows

package runtime

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestContainerScanRootWindows verifies drive-letter paths map into the Linux
// container workspace with forward slashes, nested subfolders are allowed, and
// parent/other-drive escapes are rejected. Runs only on Windows, where
// filepath has Windows semantics.
func TestContainerScanRootWindows(t *testing.T) {
	root := `C:\repo`

	if got, err := ContainerScanRoot(root, "."); err != nil || got != "/workspace" {
		t.Fatalf("root: got %q err=%v", got, err)
	}
	if got, err := ContainerScanRoot(root, "app"); err != nil || got != "/workspace/app" {
		t.Fatalf("nested rel: got %q err=%v", got, err)
	}
	if got, err := ContainerScanRoot(root, `C:\repo\app\sub`); err != nil || got != "/workspace/app/sub" {
		t.Fatalf("nested abs: got %q err=%v", got, err)
	}
	if strings.Contains(mustScanRoot(t, root, `C:\repo\a\b`), `\`) {
		t.Error("container path must use forward slashes, not backslashes")
	}
	if _, err := ContainerScanRoot(root, `..\other`); err == nil {
		t.Error("parent-directory escape must be rejected")
	}
	if _, err := ContainerScanRoot(root, `D:\other`); err == nil {
		t.Error("a path on another drive must be rejected")
	}
}

func mustScanRoot(t *testing.T, root, target string) string {
	t.Helper()
	got, err := ContainerScanRoot(root, target)
	if err != nil {
		t.Fatalf("ContainerScanRoot(%q,%q): %v", root, target, err)
	}
	return got
}

func TestContainerScanRootWindowsSpacesUnicode(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "My Project", "данные")
	if err := os.MkdirAll(target, 0o755); err != nil {
		t.Fatal(err)
	}
	got, err := ContainerScanRoot(root, target)
	if err != nil {
		t.Fatal(err)
	}
	if got != "/workspace/My Project/данные" {
		t.Fatalf("container path = %q", got)
	}
}

// TestHostUserEmptyOnWindows: a Windows account SID is not a Docker uid:gid, so
// no --user mapping must be produced.
func TestHostUserEmptyOnWindows(t *testing.T) {
	if got := HostUser(); got != "" {
		t.Errorf("HostUser() on Windows = %q, want empty (no uid:gid)", got)
	}
}

func TestEnsureScannerCacheDirWindows(t *testing.T) {
	cache := filepath.Join(t.TempDir(), "scanners")
	t.Setenv("AITRIAGE_SCANNER_CACHE_DIR", cache)
	got, err := EnsureScannerCacheDir()
	if err != nil {
		t.Fatal(err)
	}
	if got != cache {
		t.Fatalf("cache = %q, want %q", got, cache)
	}
	if _, err := os.Stat(cache); err != nil {
		t.Fatalf("cache dir not created: %v", err)
	}
}

func TestDockerExecutableFallsBackToDockerDesktopInstall(t *testing.T) {
	programFiles := t.TempDir()
	docker := filepath.Join(programFiles, "Docker", "Docker", "resources", "bin", "docker.exe")
	if err := os.MkdirAll(filepath.Dir(docker), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(docker, []byte("stub"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", t.TempDir())
	t.Setenv("ProgramFiles", programFiles)
	t.Setenv("ProgramW6432", "")
	t.Setenv("ProgramFiles(x86)", "")

	if got := DockerExecutable(); got != docker {
		t.Fatalf("DockerExecutable() = %q, want Docker Desktop fallback %q", got, docker)
	}
}

func TestDockerMissingErrorIsActionableOnWindows(t *testing.T) {
	err := errDockerNotInstalled()
	if err.ActionURL != "https://docs.docker.com/desktop/setup/install/windows-install/" {
		t.Fatalf("ActionURL = %q", err.ActionURL)
	}
	if err.RetryCommand != "aitriage setup --full" {
		t.Fatalf("RetryCommand = %q", err.RetryCommand)
	}
	if !strings.Contains(err.Message, "close and reopen your terminal or AI IDE") {
		t.Fatalf("Windows PATH recovery hint missing: %q", err.Message)
	}
}
