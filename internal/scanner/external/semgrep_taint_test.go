package external_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/cybertortuga/aitriage/internal/scanner/external"
	"github.com/cybertortuga/aitriage/rules"
)

// runXSS runs the trusted taint config (as a full audit would) over one fixture
// directory and returns the FAST-XSS findings, keyed by base filename.
func runXSS(t *testing.T, sub string) map[string][]external.UnifiedFinding {
	t.Helper()
	if !external.IsInstalled("semgrep") {
		t.Skip("semgrep not installed; skipping real-Semgrep taint test")
	}

	dir := t.TempDir()
	cfg, ids, err := rules.WriteTaintConfig(dir)
	if err != nil {
		t.Fatalf("WriteTaintConfig: %v", err)
	}

	target := filepath.Join("testdata", "xss", sub)
	findings, err := external.RunSemgrepConfigs(context.Background(), target, ids, cfg)
	if err != nil {
		t.Fatalf("RunSemgrepConfigs(%s): %v", sub, err)
	}

	byFile := map[string][]external.UnifiedFinding{}
	for _, f := range findings {
		if f.RuleID != "FAST-XSS" {
			continue
		}
		byFile[filepath.Base(f.File)] = append(byFile[filepath.Base(f.File)], f)
	}
	return byFile
}

// TestFastXSS_Unsafe_RealSemgrep drives the real bundled Semgrep over the unsafe
// fixtures and asserts every reflected-XSS variant is caught with the canonical
// FAST-XSS id: direct, through a local variable / f-string / % / .format(), and
// via query_params / Query() / Form().
func TestFastXSS_Unsafe_RealSemgrep(t *testing.T) {
	byFile := runXSS(t, "unsafe")

	cases := []struct {
		file string
		min  int
		what string
	}{
		{"direct.py", 1, "direct reflected XSS in HTMLResponse f-string"},
		{"via_var.py", 1, "taint through local variable, concat and %-format"},
		{"query_params.py", 1, "request.query_params.get via .format()"},
		{"query_dep.py", 2, "Query() and Form() parameters into HTML sinks"},
		{"custom_escape.py", 1, "a user-defined function named escape is not a trusted sanitizer"},
		{"response_class.py", 1, "route decorator response_class=HTMLResponse"},
		{"response_charset.py", 1, "HTML Response media type with charset parameter"},
	}
	for _, c := range cases {
		got := len(byFile[c.file])
		if got < c.min {
			t.Errorf("%s: expected >=%d FAST-XSS finding(s) (%s), got %d", c.file, c.min, c.what, got)
		}
	}
	for _, f := range byFile {
		for _, finding := range f {
			if finding.Source != "semgrep" {
				t.Errorf("expected source 'semgrep', got %q", finding.Source)
			}
			if finding.RuleID != "FAST-XSS" {
				t.Errorf("expected canonical rule id FAST-XSS, got %q", finding.RuleID)
			}
		}
	}
}

// TestFastXSS_Safe_RealSemgrep asserts no false positives: html.escape /
// markupsafe.escape sanitized flows, a static HTML page, and a JSONResponse
// carrying a route parameter must NOT be reported.
func TestFastXSS_Safe_RealSemgrep(t *testing.T) {
	byFile := runXSS(t, "safe")
	if len(byFile) != 0 {
		t.Errorf("expected zero FAST-XSS findings on safe fixtures, got %v", byFile)
	}
}
