package telemetry

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const (
	maxScans      = 100
	telemetryDir  = ".aitriage"
	telemetryFile = "telemetry.json"
)

// storePath returns the full path to the telemetry file: ~/.aitriage/telemetry.json
func storePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, telemetryDir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return filepath.Join(dir, telemetryFile), nil
}

// Load reads the telemetry store from disk. Returns an empty store if not found.
func Load() TelemetryStore {
	path, err := storePath()
	if err != nil {
		return TelemetryStore{}
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return TelemetryStore{}
	}

	var store TelemetryStore
	if err := json.Unmarshal(data, &store); err != nil {
		return TelemetryStore{}
	}
	return store
}

// Save writes the telemetry store to disk atomically (write-to-temp + rename).
func Save(store TelemetryStore) error {
	path, err := storePath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(store, "", "  ")
	if err != nil {
		return err
	}

	// Atomic write: temp file → rename
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
