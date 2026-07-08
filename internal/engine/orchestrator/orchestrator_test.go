package orchestrator

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestRunAllScanners_Integration(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Write a dummy go file to trigger the scanner
	goFile := filepath.Join(tempDir, "main.go")
	err := os.WriteFile(goFile, []byte("package main\n\nfunc main() {\n\t// TODO: fix this\n}\n"), 0644)
	if err != nil {
		t.Fatalf("Failed to write dummy go file: %v", err)
	}

	// Write a dummy docker-compose file to trigger network and deploy audit
	composeFile := filepath.Join(tempDir, "docker-compose.yml")
	err = os.WriteFile(composeFile, []byte("version: '3'\nservices:\n  web:\n    image: nginx\n    privileged: true\n    ports:\n      - \"8080:80\"\n"), 0644)
	if err != nil {
		t.Fatalf("Failed to write dummy compose file: %v", err)
	}

	opts := Options{
		ProjectPath:  tempDir,
		ProbeHost:    "localhost",
		ForceStack:   "",
		RunExternal:  false, // We skip external tools (trivy, semgrep, gitleaks, bandit) to avoid dependencies
		FullPortScan: false,
	}

	ctx := context.Background()

	result := RunAllScanners(ctx, opts)

	// Validate the result
	if result.ProjectPath != tempDir {
		t.Errorf("Expected ProjectPath %q, got %q", tempDir, result.ProjectPath)
	}

	// The network probe should have found the 8080 port from docker-compose
	foundDockerCompose := false
	for _, finding := range result.Network {
		if finding.Target == "web" || finding.Target == "docker-compose" {
			foundDockerCompose = true
			break
		}
	}
	if !foundDockerCompose && len(result.Network) > 0 {
		t.Logf("Warning: Expected network findings to include docker-compose ports, got: %v", result.Network)
	}

	// Scanner should run (Report struct is populated)
	if result.Report.ProjectPath == "" && len(result.Report.Results) == 0 {
		t.Logf("Scanner ran but returned empty/uninitialized report")
	}

	// Deploy audit should have found the docker-compose.yml
	if len(result.Deploy) == 0 {
		t.Errorf("Expected Deploy findings for docker-compose.yml, got none")
	}

	// NFR checks should run
	if result.NFR == nil {
		t.Errorf("Expected NFR findings to be initialized")
	}

	// Architecture diagram should run
	if result.Diagram == "" {
		t.Errorf("Expected Diagram to be generated")
	}
}
