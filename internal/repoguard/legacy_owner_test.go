// Package repoguard holds regression guards that protect distribution-level
// invariants of the repository. It contains no runtime code.
//
// The legacy-owner guard defends the dodobrands migration: after the module,
// image and public URLs moved from the previous personal owner to the
// dodobrands organization, no product-owned file may reintroduce the old
// "<owner>/aitriage" repository path (Go module, GHCR image, GitHub URL,
// Action reference or raw path all share that substring).
package repoguard

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// legacyOwnerRepo is assembled from fragments so this guard file does not itself
// contain the forbidden literal and therefore does not match its own scan.
var legacyOwnerRepo = "cyber" + "tortuga" + "/aitriage"

// allowedFiles lists product-owned files that are permitted to keep a
// documented backward-compatibility reference to the legacy repository path.
//
// It is intentionally empty: the migration removed every "<owner>/aitriage"
// reference, so no compatibility exception is currently required. Add an entry
// here (with a comment explaining why) only for a deliberately preserved,
// documented compatibility link.
var allowedFiles = map[string]bool{}

// repoRoot walks up from the test's working directory until it finds the module
// go.mod, so the guard works regardless of where `go test` is invoked from.
func repoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("could not locate repository root (go.mod not found)")
		}
		dir = parent
	}
}

// TestNoLegacyOwnerReferences fails if any tracked, product-owned file contains
// the legacy "<owner>/aitriage" repository path. This rejects an accidental
// revert of the dodobrands migration during review or future edits.
func TestNoLegacyOwnerReferences(t *testing.T) {
	root := repoRoot(t)

	// Enumerate tracked files via git so vendored/build/output directories that
	// are not part of the product are naturally excluded.
	cmd := exec.Command("git", "ls-files", "-z")
	cmd.Dir = root
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("git ls-files failed (is this a git checkout?): %v", err)
	}

	files := strings.Split(strings.TrimRight(string(out), "\x00"), "\x00")
	var offenders []string

	for _, rel := range files {
		if rel == "" || allowedFiles[rel] {
			continue
		}
		// Skip vendored third-party trees that are not product-owned.
		if strings.HasPrefix(rel, "web/node_modules/") {
			continue
		}
		// The guard file itself never contains the literal, but skip it anyway
		// to keep the intent obvious.
		if strings.HasSuffix(rel, "legacy_owner_test.go") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(root, rel))
		if err != nil {
			// A tracked path that cannot be read (e.g. a symlink target) is not
			// a source of a legacy reference; skip it rather than fail.
			continue
		}
		if strings.Contains(string(data), legacyOwnerRepo) {
			offenders = append(offenders, rel)
		}
	}

	if len(offenders) > 0 {
		t.Fatalf("legacy owner path %q must not appear in product-owned files; "+
			"the repository migrated to dodobrands/aitriage. Offending files:\n  %s",
			legacyOwnerRepo, strings.Join(offenders, "\n  "))
	}
}
