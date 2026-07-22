package runtime

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFirstExistingFile(t *testing.T) {
	dir := t.TempDir()
	real := filepath.Join(dir, "docker.exe")
	if err := os.WriteFile(real, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	missing := filepath.Join(dir, "nope.exe")

	if got := firstExistingFile([]string{missing, real}); got != real {
		t.Errorf("firstExistingFile = %q, want %q", got, real)
	}
	if got := firstExistingFile([]string{missing}); got != "" {
		t.Errorf("firstExistingFile on all-missing = %q, want empty", got)
	}
	// A directory must not be treated as the docker binary.
	if got := firstExistingFile([]string{dir}); got != "" {
		t.Errorf("firstExistingFile must skip directories, got %q", got)
	}
}

func TestWindowsDockerCandidates(t *testing.T) {
	t.Setenv("ProgramFiles", `C:\Program Files`)
	t.Setenv("ProgramW6432", `C:\Program Files`) // duplicate root must be de-duplicated
	t.Setenv("ProgramFiles(x86)", `C:\Program Files (x86)`)

	got := windowsDockerCandidates()
	if len(got) != 2 {
		t.Fatalf("expected 2 de-duplicated candidates, got %d: %v", len(got), got)
	}
	want := filepath.Join(`C:\Program Files`, "Docker", "Docker", "resources", "bin", "docker.exe")
	if got[0] != want {
		t.Errorf("candidate[0] = %q, want %q", got[0], want)
	}
	for _, c := range got {
		if !strings.HasSuffix(c, "docker.exe") {
			t.Errorf("candidate %q must end in docker.exe", c)
		}
		if !filepath.IsAbs(c) && !strings.HasPrefix(c, `C:\`) {
			t.Errorf("candidate %q must be an absolute system path", c)
		}
	}
}

func TestWindowsDockerCandidatesEmptyWhenNoProgramFiles(t *testing.T) {
	t.Setenv("ProgramFiles", "")
	t.Setenv("ProgramW6432", "")
	t.Setenv("ProgramFiles(x86)", "")
	if got := windowsDockerCandidates(); len(got) != 0 {
		t.Errorf("expected no candidates without Program Files env, got %v", got)
	}
}
