package network

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestProbeHost(t *testing.T) {
	// Start a local TCP server to test port probing
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to start local TCP server: %v", err)
	}
	defer func() { _ = listener.Close() }()

	// Get the port it was assigned
	addr := listener.Addr().String()
	parts := strings.Split(addr, ":")
	portStr := parts[len(parts)-1]

	// Create a goroutine to accept connection so it doesn't block indefinitely
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			_ = conn.Close()
		}
	}()

	// We can't guarantee `ProbeHost` will scan our random port if we just do fullScan=false,
	// because `ProbeHost` scans a specific set of common ports or full 65535 ports.
	// We should test `probePort` directly, or patch commonPorts for testing.
	// Since `probePort` is unexported, we can call it.

	var p int
	_, _ = fmt.Sscanf(portStr, "%d", &p)

	finding := probePort("127.0.0.1", p, "TestService", 1*time.Second)
	if finding == nil {
		t.Fatalf("Expected probePort to find the open port %d, got nil", p)
	}

	if finding.Port != p {
		t.Errorf("Expected finding port %d, got %d", p, finding.Port)
	}
	if finding.Target != "127.0.0.1" {
		t.Errorf("Expected target 127.0.0.1, got %s", finding.Target)
	}
}

func TestProbeDockerCompose(t *testing.T) {
	tempDir := t.TempDir()

	composeContent := []byte("version: '3'\nservices:\n  db:\n    image: postgres\n    ports:\n      - \"5432:5432\"\n  web:\n    image: nginx\n    ports:\n      - \"8080:80\"\n")
	err := os.WriteFile(filepath.Join(tempDir, "docker-compose.yml"), composeContent, 0644)
	if err != nil {
		t.Fatalf("Failed to write docker-compose.yml: %v", err)
	}

	findings := ProbeDockerCompose(tempDir, false)

	// ProbeDockerCompose will attempt to dial 'db' and 'web' which will fail on localhost.
	// It should return DNS failure findings for 'db' and 'web'.

	foundDB := false
	foundWeb := false

	for _, f := range findings {
		if f.Target == "db" && f.Service == "DNS" {
			foundDB = true
		}
		if f.Target == "web" && f.Service == "DNS" {
			foundWeb = true
		}
	}

	if !foundDB {
		t.Errorf("Expected DNS finding for service 'db'")
	}
	if !foundWeb {
		t.Errorf("Expected DNS finding for service 'web'")
	}
}
