package telemetry

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestRecord(t *testing.T) {
	t.Run("Disabled", func(t *testing.T) {
		tmpHome := t.TempDir()
		t.Setenv("HOME", tmpHome)
		t.Setenv("USERPROFILE", tmpHome)
		t.Setenv("AITRIAGE_TELEMETRY", "off")

		metric := ScanMetric{
			FilesScanned: 10,
		}
		Record(metric)

		store := Load()
		if store.TotalScans != 0 {
			t.Errorf("Expected 0 scans, got %d", store.TotalScans)
		}

		if _, err := os.Stat(filepath.Join(tmpHome, telemetryDir, telemetryFile)); !os.IsNotExist(err) {
			t.Errorf("Telemetry file should not exist")
		}
	})

	t.Run("Enabled_AggregatesAndHash", func(t *testing.T) {
		tmpHome := t.TempDir()
		t.Setenv("HOME", tmpHome)
		t.Setenv("USERPROFILE", tmpHome)
		t.Setenv("AITRIAGE_TELEMETRY", "")

		blankMetric := ScanMetric{
			FilesScanned:  5,
			FindingsTotal: 2,
			SecurityScore: 90,
		}

		Record(blankMetric)

		store := Load()
		if store.TotalScans != 1 {
			t.Fatalf("Expected 1 scan, got %d", store.TotalScans)
		}
		if store.TotalFiles != 5 {
			t.Errorf("Expected 5 total files, got %d", store.TotalFiles)
		}
		if store.TotalFindings != 2 {
			t.Errorf("Expected 2 total findings, got %d", store.TotalFindings)
		}
		if store.AvgScore != 90.0 {
			t.Errorf("Expected AvgScore 90.0, got %f", store.AvgScore)
		}

		expectedHashBytes := sha256.Sum256([]byte(""))
		expectedHash := fmt.Sprintf("%x", expectedHashBytes[:4])
		if store.Scans[0].ProjectHash != expectedHash {
			t.Errorf("Expected project hash %q, got %q", expectedHash, store.Scans[0].ProjectHash)
		}
		if store.Scans[0].Timestamp.IsZero() {
			t.Errorf("Timestamp should be populated")
		}

		metric2 := ScanMetric{
			ProjectHash:   "my-project",
			FilesScanned:  10,
			FindingsTotal: 5,
			SecurityScore: 80,
		}
		Record(metric2)

		store = Load()
		if store.TotalScans != 2 {
			t.Fatalf("Expected 2 scans, got %d", store.TotalScans)
		}
		if store.TotalFiles != 15 {
			t.Errorf("Expected 15 total files, got %d", store.TotalFiles)
		}
		if store.TotalFindings != 7 {
			t.Errorf("Expected 7 total findings, got %d", store.TotalFindings)
		}
		if store.AvgScore != 85.0 {
			t.Errorf("Expected AvgScore 85.0, got %f", store.AvgScore)
		}
		if store.Scans[1].ProjectHash != "my-project" {
			t.Errorf("Expected project hash 'my-project', got %q", store.Scans[1].ProjectHash)
		}
	})

	t.Run("RingBuffer", func(t *testing.T) {
		tmpHome := t.TempDir()
		t.Setenv("HOME", tmpHome)
		t.Setenv("USERPROFILE", tmpHome)
		t.Setenv("AITRIAGE_TELEMETRY", "")

		for i := 0; i < maxScans+5; i++ {
			Record(ScanMetric{
				FilesScanned:  1,
				SecurityScore: 100,
			})
		}

		store := Load()
		if store.TotalScans != maxScans+5 {
			t.Errorf("Expected TotalScans to be %d, got %d", maxScans+5, store.TotalScans)
		}
		if len(store.Scans) != maxScans {
			t.Errorf("Expected len(Scans) to be bounded by %d, got %d", maxScans, len(store.Scans))
		}
	})
}
