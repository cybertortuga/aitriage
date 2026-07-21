package mcp

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSanitizeSemgrepConfigSafe(t *testing.T) {
	// A real local file that must still be rejected under the safe profile.
	localCfg := filepath.Join(t.TempDir(), "rules.yml")
	if err := os.WriteFile(localCfg, []byte("rules: []\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	allowed := []string{"", "auto", "p/security-audit", "r/python.flask", "s/team.rules"}
	for _, c := range allowed {
		if _, err := sanitizeSemgrepConfig(c, true); err != nil {
			t.Errorf("safe should allow config %q, got error: %v", c, err)
		}
	}

	rejected := []string{
		localCfg,              // absolute path to a real file
		"/etc/passwd",         // absolute path
		"../../rules.yml",     // traversal
		"./rules.yml",         // relative path
		"https://evil.test/r", // URL
		"http://x/y",          // URL
		"file:///etc/passwd",  // file URL
		"rules.yml",           // bare filename
	}
	for _, c := range rejected {
		if _, err := sanitizeSemgrepConfig(c, true); err == nil {
			t.Errorf("safe should reject config %q, but it was allowed", c)
		}
	}
}

func TestSanitizeSemgrepConfigFullPassthrough(t *testing.T) {
	// With restrict=false (full profile) anything passes through unchanged.
	for _, c := range []string{"/abs/path.yml", "https://x/y", "../x", "auto"} {
		got, err := sanitizeSemgrepConfig(c, false)
		if err != nil {
			t.Errorf("full profile should pass %q through, got error: %v", c, err)
		}
		if got != c {
			t.Errorf("full profile altered config %q -> %q", c, got)
		}
	}
}
