package rules

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// TestTaintSemgrepConfig verifies the trusted taint config is well-formed and
// contains the mandatory FAST-XSS rule with sources and sinks.
func TestTaintSemgrepConfig(t *testing.T) {
	cfg, ids, err := TaintSemgrepConfig()
	if err != nil {
		t.Fatalf("TaintSemgrepConfig: %v", err)
	}
	found := false
	for _, id := range ids {
		if id == "FAST-XSS" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected FAST-XSS in emitted ids, got %v", ids)
	}

	// The config must be valid YAML that Semgrep can consume: one rule per id,
	// each with mode:taint, sources and sinks.
	var parsed struct {
		Rules []struct {
			ID       string      `yaml:"id"`
			Mode     string      `yaml:"mode"`
			Sources  []yaml.Node `yaml:"pattern-sources"`
			Sinks    []yaml.Node `yaml:"pattern-sinks"`
			Severity string      `yaml:"severity"`
		} `yaml:"rules"`
	}
	if err := yaml.Unmarshal(cfg, &parsed); err != nil {
		t.Fatalf("generated config is not valid YAML: %v", err)
	}
	if len(parsed.Rules) != len(ids) {
		t.Fatalf("config has %d rules, expected %d", len(parsed.Rules), len(ids))
	}
	for _, r := range parsed.Rules {
		if r.Mode != "taint" {
			t.Errorf("rule %q mode=%q, want taint", r.ID, r.Mode)
		}
		if len(r.Sources) == 0 || len(r.Sinks) == 0 {
			t.Errorf("rule %q missing sources or sinks", r.ID)
		}
		if r.Severity != "ERROR" && r.Severity != "WARNING" && r.Severity != "INFO" {
			t.Errorf("rule %q has non-Semgrep severity %q", r.ID, r.Severity)
		}
	}

	// The config must NOT leak engine-only keys that Semgrep would reject.
	for _, bad := range []string{"stack:", "extensions:", "target:", "condition:", "suggestion:", "exclude_tests:"} {
		if strings.Contains(string(cfg), bad) {
			t.Errorf("generated Semgrep config must not contain engine-only key %q", bad)
		}
	}
}

// TestWriteTaintConfig verifies the config is written with owner-only perms.
func TestWriteTaintConfig(t *testing.T) {
	dir := t.TempDir()
	path, ids, err := WriteTaintConfig(dir)
	if err != nil {
		t.Fatalf("WriteTaintConfig: %v", err)
	}
	if len(ids) == 0 {
		t.Fatal("expected at least one taint rule id")
	}
	if filepath.Dir(path) != dir {
		t.Fatalf("config written outside provided dir: %s", path)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if perm := info.Mode().Perm(); runtime.GOOS != "windows" && perm != 0o600 {
		t.Errorf("taint config perms = %o, want 600", perm)
	}
}

// TestTaintRulesArePresent checks that the mandatory taint rule exists in the
// catalog with the exact canonical id and taint mode.
func TestTaintRulesArePresent(t *testing.T) {
	trs, err := TaintRules()
	if err != nil {
		t.Fatalf("TaintRules: %v", err)
	}
	byID := map[string]bool{}
	for _, r := range trs {
		if !r.IsTaint() {
			t.Errorf("TaintRules returned non-taint rule %q", r.ID)
		}
		byID[r.ID] = true
	}
	for _, req := range RequiredTaintRuleIDs {
		if !byID[req] {
			t.Errorf("mandatory taint rule %q missing from catalog", req)
		}
	}
}
