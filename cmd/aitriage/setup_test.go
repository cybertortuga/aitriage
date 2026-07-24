package main

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	rt "github.com/dodobrands/aitriage/internal/runtime"
)

type removeDockerFake struct {
	installed bool
	exists    bool
	removed   int
	removeErr error
}

func (f *removeDockerFake) Installed() bool                          { return f.installed }
func (f *removeDockerFake) Info(context.Context) error               { return nil }
func (f *removeDockerFake) ImageExists(context.Context, string) bool { return f.exists }
func (f *removeDockerFake) Pull(context.Context, string) error       { return nil }
func (f *removeDockerFake) Digest(context.Context, string) string    { return "" }
func (f *removeDockerFake) RemoveImage(context.Context, string) error {
	f.removed++
	return f.removeErr
}
func (f *removeDockerFake) Warm(context.Context, string) error { return nil }
func (f *removeDockerFake) Verify(context.Context, string) ([]rt.BundleStatus, error) {
	return nil, nil
}

// TestBuildSetupOutContract locks the stable --json contract: status mapping,
// required fields, and that a valid JSON document is produced for each outcome.
func TestBuildSetupOutContract(t *testing.T) {
	cases := []struct {
		name       string
		report     *rt.Report
		wantStatus string
		wantRetry  bool
	}{
		{
			name:       "ready",
			report:     &rt.Report{OK: true, Image: "img", Digest: "sha256:x"},
			wantStatus: "ok",
		},
		{
			name:       "docker missing is action_required",
			report:     &rt.Report{Err: &rt.SetupError{Code: "docker_not_installed", Message: "m", ActionURL: "https://x", RetryCommand: "aitriage setup --full"}},
			wantStatus: "action_required",
			wantRetry:  true,
		},
		{
			name:       "bundle incomplete is error",
			report:     &rt.Report{Err: &rt.SetupError{Code: "bundle_incomplete", Message: "m", RetryCommand: "aitriage setup --repair"}},
			wantStatus: "error",
			wantRetry:  true,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			out := buildSetupOut(c.report)
			if out.Status != c.wantStatus {
				t.Errorf("status = %q, want %q", out.Status, c.wantStatus)
			}
			if out.Message == "" {
				t.Error("message must never be empty")
			}
			if c.wantRetry && out.RetryCommand == "" {
				t.Error("recoverable outcome must carry retry_command")
			}
			// Must marshal to exactly one valid JSON document.
			b, err := json.Marshal(out)
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			var round map[string]any
			if err := json.Unmarshal(b, &round); err != nil {
				t.Fatalf("output is not valid JSON: %v", err)
			}
		})
	}
}

func TestBuildRemoveRuntimeReport(t *testing.T) {
	f := &removeDockerFake{installed: true, exists: true}
	report := buildRemoveRuntimeReport(context.Background(), f, "example/image:v1")
	if !report.OK || f.removed != 1 {
		t.Fatalf("remove report = %+v, removed=%d", report, f.removed)
	}

	absent := &removeDockerFake{installed: true, exists: false}
	report = buildRemoveRuntimeReport(context.Background(), absent, "example/image:v1")
	if !report.OK || absent.removed != 0 {
		t.Fatalf("absent image must be an idempotent success: %+v", report)
	}

	failed := &removeDockerFake{installed: true, exists: true, removeErr: errors.New("image is in use")}
	report = buildRemoveRuntimeReport(context.Background(), failed, "example/image:v1")
	if report.OK || report.Err == nil || report.Err.Code != "runtime_remove_failed" {
		t.Fatalf("remove failure not classified: %+v", report)
	}
}

func TestSetupColorContract(t *testing.T) {
	cases := []struct {
		name      string
		noColor   string
		term      string
		stderrTTY bool
		want      bool
	}{
		{name: "interactive terminal", stderrTTY: true, term: "xterm-256color", want: true},
		{name: "redirected output", stderrTTY: false, term: "xterm-256color", want: false},
		{name: "NO_COLOR", stderrTTY: true, term: "xterm-256color", noColor: "1", want: false},
		{name: "dumb terminal", stderrTTY: true, term: "dumb", want: false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := setupColorsEnabled(tc.noColor, tc.term, tc.stderrTTY); got != tc.want {
				t.Fatalf("setupColorsEnabled() = %v, want %v", got, tc.want)
			}
		})
	}
	if got := (setupColors{enabled: true}).status(rt.StepOK); got == string(rt.StepOK) {
		t.Fatal("interactive status should include ANSI color")
	}
	if got := (setupColors{enabled: false}).status(rt.StepOK); got != string(rt.StepOK) {
		t.Fatalf("plain status = %q, want %q", got, rt.StepOK)
	}
}
