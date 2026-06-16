package reporter

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/cybertortuga/aitriage/internal/engine/core"
	"github.com/cybertortuga/aitriage/internal/scanner"
	"github.com/cybertortuga/aitriage/internal/scanner/detector"
)

func createMockReport() scanner.ScanReport {
	return scanner.ScanReport{
		ProjectPath: "/test/path",
		Stacks:      []detector.Stack{detector.Go},
		Results: []core.CheckResult{
			{
				ID:           "TEST-001",
				Name:         "Test Check",
				Severity:     "CRITICAL",
				Status:       core.Absent,
				Suggestion:   "Fix the test",
				Evidence:     "Test evidence",
				File:         "main.go",
				Line:         10,
				OWASPMapping: "A01",
			},
			{
				ID:         "TEST-002",
				Name:       "Test Check 2",
				Severity:   "LOW",
				Status:     core.Present,
				Suggestion: "All good",
				Evidence:   "Found it",
			},
		},
		HasCriticalFailures: true,
		SecurityScore:       40,
		SecurityGrade:       "F",
		TotalFiles:          10,
		RulesApplied:        2,
		ScanDuration:        100 * time.Millisecond,
	}
}

func TestPrintJSON(t *testing.T) {
	report := createMockReport()

	var buf bytes.Buffer
	PrintJSON(report, &buf)
	output := buf.String()

	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Fatalf("PrintJSON did not output valid JSON: %v", err)
	}

	if parsed["project_path"] != "/test/path" {
		t.Errorf("Expected project_path /test/path, got %v", parsed["project_path"])
	}
	if parsed["security_score"].(float64) != 40 {
		t.Errorf("Expected security_score 40, got %v", parsed["security_score"])
	}
}

func TestPrintSARIF(t *testing.T) {
	report := createMockReport()

	var buf bytes.Buffer
	PrintSARIF(report, &buf)
	output := buf.String()

	var sarif SarifReport
	if err := json.Unmarshal([]byte(output), &sarif); err != nil {
		t.Fatalf("PrintSARIF did not output valid JSON: %v", err)
	}

	if sarif.Version != "2.1.0" {
		t.Errorf("Expected SARIF version 2.1.0, got %v", sarif.Version)
	}
	if len(sarif.Runs) == 0 {
		t.Fatalf("Expected at least one run in SARIF")
	}

	results := sarif.Runs[0].Results
	if len(results) != 1 {
		// Only Absent status checks should be in SARIF
		t.Errorf("Expected 1 result in SARIF, got %d", len(results))
	} else {
		if results[0].RuleID != "TEST-001" {
			t.Errorf("Expected rule ID TEST-001, got %v", results[0].RuleID)
		}
		if results[0].Level != "error" {
			t.Errorf("Expected severity error for CRITICAL, got %v", results[0].Level)
		}
	}
}

func TestGenerateHTMLReport(t *testing.T) {
	report := createMockReport()
	tempFile := t.TempDir() + "/test-report.html"

	stackNames := make([]string, len(report.Stacks))
	for i, s := range report.Stacks {
		stackNames[i] = string(s)
	}

	data := ReportData{
		SecurityGrade: report.SecurityGrade,
		CriticalCount: 1,
		Stacks:        stackNames,
		Results:       report.Results,
	}

	err := GenerateHTMLReport(tempFile, data)
	if err != nil {
		t.Fatalf("GenerateHTMLReport failed: %v", err)
	}

	content, err := os.ReadFile(tempFile)
	if err != nil {
		t.Fatalf("Could not read generated HTML file: %v", err)
	}

	htmlStr := string(content)
	if !strings.Contains(htmlStr, "AITriage Security Report") {
		t.Errorf("HTML report missing title")
	}
	if !strings.Contains(htmlStr, "TEST-001") {
		t.Errorf("HTML report missing finding TEST-001")
	}
	if !strings.Contains(htmlStr, "gF") { // Grade F class
		t.Errorf("HTML report missing grade class")
	}
}

func TestGenerateHTMLReport_XSS(t *testing.T) {
	tempFile := t.TempDir() + "/test-report-xss.html"

	xssPayload := "<script>alert('XSS')</script>"

	data := ReportData{
		SecurityGrade: "A",
		CriticalCount: 0,
		Stacks:        []string{"Go"},
		Results: []core.CheckResult{
			{
				ID:         "TEST-XSS",
				Name:       xssPayload,
				Severity:   "CRITICAL",
				Status:     core.Absent,
				Suggestion: xssPayload,
				Evidence:   xssPayload,
			},
		},
	}

	err := GenerateHTMLReport(tempFile, data)
	if err != nil {
		t.Fatalf("GenerateHTMLReport failed: %v", err)
	}

	content, err := os.ReadFile(tempFile)
	if err != nil {
		t.Fatalf("Could not read generated HTML file: %v", err)
	}

	htmlStr := string(content)

	if strings.Contains(htmlStr, xssPayload) {
		t.Errorf("HTML report contains unescaped XSS payload")
	}

	if !strings.Contains(htmlStr, "&lt;script&gt;alert(&#39;XSS&#39;)&lt;/script&gt;") {
		t.Errorf("HTML report does not contain escaped XSS payload. HTML content: %s", htmlStr)
	}
}
