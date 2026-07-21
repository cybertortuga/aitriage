package mcp

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestParseProfile(t *testing.T) {
	cases := map[string]struct {
		want    Profile
		wantErr bool
	}{
		"":       {ProfileFull, false},
		"full":   {ProfileFull, false},
		"FULL":   {ProfileFull, false},
		"safe":   {ProfileSafe, false},
		" safe ": {ProfileSafe, false},
		"bogus":  {"", true},
	}
	for in, c := range cases {
		got, err := ParseProfile(in)
		if c.wantErr {
			if err == nil {
				t.Errorf("ParseProfile(%q): expected error, got %q", in, got)
			}
			continue
		}
		if err != nil {
			t.Errorf("ParseProfile(%q): unexpected error %v", in, err)
		}
		if got != c.want {
			t.Errorf("ParseProfile(%q) = %q, want %q", in, got, c.want)
		}
	}
}

func TestProfileAllowsMutation(t *testing.T) {
	if !ProfileFull.allowsMutation() {
		t.Error("full profile must allow mutation")
	}
	if ProfileSafe.allowsMutation() {
		t.Error("safe profile must NOT allow mutation")
	}
}

func TestNilPathGuardPassthrough(t *testing.T) {
	var g *PathGuard
	got, err := g.Resolve("/anywhere/at/all")
	if err != nil {
		t.Fatalf("nil guard should not error: %v", err)
	}
	if got != "/anywhere/at/all" {
		t.Errorf("nil guard passthrough = %q, want unchanged", got)
	}
	if g.Root() != "" {
		t.Errorf("nil guard Root() = %q, want empty", g.Root())
	}
}

func TestNewPathGuardRejectsNonDir(t *testing.T) {
	if _, err := NewPathGuard(""); err == nil {
		t.Error("empty root must error")
	}
	file := filepath.Join(t.TempDir(), "f.txt")
	if err := os.WriteFile(file, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := NewPathGuard(file); err == nil {
		t.Error("file root must error")
	}
	if _, err := NewPathGuard(filepath.Join(t.TempDir(), "does-not-exist")); err == nil {
		t.Error("missing root must error")
	}
}

func TestPathGuardResolveWithinRoot(t *testing.T) {
	root := t.TempDir()
	sub := filepath.Join(root, "src", "pkg")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	g, err := NewPathGuard(root)
	if err != nil {
		t.Fatal(err)
	}

	// empty and "." resolve to the root itself.
	for _, in := range []string{"", "."} {
		got, err := g.Resolve(in)
		if err != nil {
			t.Fatalf("Resolve(%q): %v", in, err)
		}
		if got != g.Root() {
			t.Errorf("Resolve(%q) = %q, want root %q", in, got, g.Root())
		}
	}

	// relative subdir resolves inside root.
	got, err := g.Resolve("src/pkg")
	if err != nil {
		t.Fatalf("Resolve subdir: %v", err)
	}
	if !within(g.Root(), got) {
		t.Errorf("resolved subdir %q not within root %q", got, g.Root())
	}
}

func TestPathGuardRejectsTraversal(t *testing.T) {
	root := filepath.Join(t.TempDir(), "project")
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatal(err)
	}
	g, err := NewPathGuard(root)
	if err != nil {
		t.Fatal(err)
	}

	// "../" traversal to the parent (which exists) must be rejected.
	if got, err := g.Resolve(".."); err == nil {
		t.Errorf("Resolve(..) should be rejected, got %q", got)
	}
	// absolute path outside root must be rejected.
	if got, err := g.Resolve(os.TempDir()); err == nil {
		t.Errorf("Resolve(absolute-outside) should be rejected, got %q", got)
	}
}

func TestPathGuardRejectsSymlinkEscape(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink semantics differ on Windows")
	}
	base := t.TempDir()
	root := filepath.Join(base, "project")
	outside := filepath.Join(base, "secrets")
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(outside, 0o755); err != nil {
		t.Fatal(err)
	}

	// A symlink inside the root pointing outside must be rejected.
	escape := filepath.Join(root, "escape")
	if err := os.Symlink(outside, escape); err != nil {
		t.Fatal(err)
	}
	g, err := NewPathGuard(root)
	if err != nil {
		t.Fatal(err)
	}
	if got, err := g.Resolve("escape"); err == nil {
		t.Errorf("symlink escaping root should be rejected, got %q", got)
	}

	// A symlink inside the root pointing to another in-root dir is allowed.
	inner := filepath.Join(root, "inner")
	if err := os.MkdirAll(inner, 0o755); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(root, "link-inner")
	if err := os.Symlink(inner, link); err != nil {
		t.Fatal(err)
	}
	got, err := g.Resolve("link-inner")
	if err != nil {
		t.Fatalf("in-root symlink should resolve: %v", err)
	}
	if !within(g.Root(), got) {
		t.Errorf("resolved in-root symlink %q not within root %q", got, g.Root())
	}
}
