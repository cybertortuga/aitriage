package telemetry

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

func TestLoad(t *testing.T) {
	// Create a valid telemetry store for testing
	now := time.Now().UTC().Truncate(time.Second)
	validStore := TelemetryStore{
		TotalScans:    10,
		TotalFiles:    100,
		TotalFindings: 5,
		AvgScore:      8.5,
		Scans: []ScanMetric{
			{
				Timestamp:     now,
				ProjectHash:   "test-hash",
				Duration:      1000,
				FilesScanned:  10,
				FindingsTotal: 1,
			},
		},
	}
	validJSON, err := json.Marshal(validStore)
	if err != nil {
		t.Fatalf("Failed to marshal valid store: %v", err)
	}

	// Unmarshal back to get exact expected store (handles nil vs empty slice and time formatting differences)
	var expectValidStore TelemetryStore
	if err := json.Unmarshal(validJSON, &expectValidStore); err != nil {
		t.Fatalf("Failed to unmarshal valid store: %v", err)
	}

	tests := []struct {
		name        string
		setupFs     func(t *testing.T, homeDir string)
		expectStore TelemetryStore
	}{
		{
			name: "valid JSON",
			setupFs: func(t *testing.T, homeDir string) {
				dir := filepath.Join(homeDir, telemetryDir)
				err := os.MkdirAll(dir, 0o755)
				if err != nil {
					t.Fatalf("Failed to create dir: %v", err)
				}
				err = os.WriteFile(filepath.Join(dir, telemetryFile), validJSON, 0o644)
				if err != nil {
					t.Fatalf("Failed to write file: %v", err)
				}
			},
			expectStore: expectValidStore,
		},
		{
			name: "file missing",
			setupFs: func(t *testing.T, homeDir string) {
				// Do not create the file
			},
			expectStore: TelemetryStore{},
		},
		{
			name: "invalid JSON",
			setupFs: func(t *testing.T, homeDir string) {
				dir := filepath.Join(homeDir, telemetryDir)
				err := os.MkdirAll(dir, 0o755)
				if err != nil {
					t.Fatalf("Failed to create dir: %v", err)
				}
				err = os.WriteFile(filepath.Join(dir, telemetryFile), []byte("{invalid json}"), 0o644)
				if err != nil {
					t.Fatalf("Failed to write file: %v", err)
				}
			},
			expectStore: TelemetryStore{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			homeDir := t.TempDir()
			t.Setenv("HOME", homeDir)
			t.Setenv("USERPROFILE", homeDir)

			tt.setupFs(t, homeDir)

			store := Load()

			// Use reflect.DeepEqual to compare structs
			if !reflect.DeepEqual(store, tt.expectStore) {
				t.Errorf("Load() = %+v, want %+v", store, tt.expectStore)
			}
		})
	}
}
