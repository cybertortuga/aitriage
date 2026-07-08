package telemetry

import (
	"crypto/sha256"
	"fmt"
	"os"
	"time"
)

// IsEnabled returns true if telemetry collection is enabled.
// Disabled when AITRIAGE_TELEMETRY=off
func IsEnabled() bool {
	return os.Getenv("AITRIAGE_TELEMETRY") != "off"
}

// Record adds a new scan metric to the telemetry store.
// Respects opt-out: does nothing if AITRIAGE_TELEMETRY=off.
func Record(metric ScanMetric) {
	if !IsEnabled() {
		return
	}

	// Hash project path for privacy
	if metric.ProjectHash == "" {
		h := sha256.Sum256([]byte(metric.ProjectHash))
		metric.ProjectHash = fmt.Sprintf("%x", h[:4]) // first 8 hex chars
	}
	if metric.Timestamp.IsZero() {
		metric.Timestamp = time.Now()
	}

	store := Load()

	// Update aggregates
	store.TotalScans++
	store.TotalFiles += metric.FilesScanned
	store.TotalFindings += metric.FindingsTotal

	// Recalculate running average score
	if store.TotalScans > 0 {
		store.AvgScore = (store.AvgScore*float64(store.TotalScans-1) + float64(metric.SecurityScore)) / float64(store.TotalScans)
	}

	// Append to ring buffer (keep last maxScans)
	store.Scans = append(store.Scans, metric)
	if len(store.Scans) > maxScans {
		store.Scans = store.Scans[len(store.Scans)-maxScans:]
	}

	// Best-effort save — don't crash on errors
	_ = Save(store)
}

// Summary returns the current telemetry aggregates.
func Summary() TelemetryStore {
	return Load()
}

// HashProjectPath creates a privacy-safe hash of a project path.
func HashProjectPath(path string) string {
	h := sha256.Sum256([]byte(path))
	return fmt.Sprintf("%x", h[:4])
}

// ScoreTrend calculates the score trend from recent scans.
// Returns the delta between the average of the last 5 scans and the previous 5.
func ScoreTrend(store TelemetryStore) int {
	scans := store.Scans
	n := len(scans)
	if n < 2 {
		return 0
	}

	// Recent: last 5 (or fewer)
	recentCount := 5
	if recentCount > n {
		recentCount = n
	}
	recentSum := 0
	for _, s := range scans[n-recentCount:] {
		recentSum += s.SecurityScore
	}
	recentAvg := recentSum / recentCount

	// Previous: the 5 before that
	prevEnd := n - recentCount
	if prevEnd <= 0 {
		return 0
	}
	prevCount := 5
	if prevCount > prevEnd {
		prevCount = prevEnd
	}
	prevSum := 0
	for _, s := range scans[prevEnd-prevCount : prevEnd] {
		prevSum += s.SecurityScore
	}
	prevAvg := prevSum / prevCount

	return recentAvg - prevAvg
}
