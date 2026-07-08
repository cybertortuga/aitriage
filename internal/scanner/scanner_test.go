package scanner_test

import (
	"context"
	"os"
	"testing"

	"github.com/cybertortuga/aitriage/internal/report/healthcheck"
	"github.com/cybertortuga/aitriage/internal/scanner"
)

func TestScanReturnsReport(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a dummy file that might trigger a finding, e.g., a hardcoded secret pattern
	err := os.WriteFile(tmpDir+"/main.go", []byte("package main\n\nfunc main() {\n\tsecret := \"AKIAIOSFODNN7EXAMPLE\"\n}\n"), 0644)
	if err != nil {
		t.Fatalf("Failed to create dummy file: %v", err)
	}

	report, err := scanner.Scan(context.Background(), tmpDir, scanner.ScanOptions{})
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}
	if report.TotalFiles == 0 {
		t.Error("Expected TotalFiles > 0")
	}
	if len(report.Results) == 0 {
		t.Error("Expected findings in the dummy project")
	}
}

func TestScanAppliesHealthCheckPolicyConfig(t *testing.T) {
	tmpDir := t.TempDir()
	if err := os.WriteFile(tmpDir+"/main.go", []byte("package main\n\nfunc main() {\n\tsecret := \"AKIAIOSFODNN7EXAMPLE\"\n}\n"), 0644); err != nil {
		t.Fatalf("Failed to create dummy file: %v", err)
	}
	if err := os.WriteFile(tmpDir+"/.aitriage.yaml", []byte(`
health_check:
  profile: strict
  fail_on: never
  minimum_score: 100
`), 0644); err != nil {
		t.Fatalf("Failed to create config: %v", err)
	}

	report, err := scanner.Scan(context.Background(), tmpDir, scanner.ScanOptions{})
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	if report.HealthCheck.Policy.Profile != healthcheck.PolicyStrict {
		t.Fatalf("profile = %q; want strict", report.HealthCheck.Policy.Profile)
	}
	if report.HealthCheck.Policy.FailOn != healthcheck.FailOnNever {
		t.Fatalf("fail_on = %q; want never", report.HealthCheck.Policy.FailOn)
	}
	if !report.HealthCheck.Verdict.Passed {
		t.Fatalf("verdict failed despite fail_on=never: %+v", report.HealthCheck.Verdict)
	}
}
