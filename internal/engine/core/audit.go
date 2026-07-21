package core

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// AuditStatus represents the human-driven triage state of a finding
type AuditStatus string

const (
	AuditStatusOpen    AuditStatus = "OPEN"
	AuditStatusIgnored AuditStatus = "IGNORED"
	AuditStatusTriage  AuditStatus = "TRIAGE"
)

// AuditEntry records a single triage decision
type AuditEntry struct {
	Status    AuditStatus `json:"status"`
	Reason    string      `json:"reason,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
	Author    string      `json:"author,omitempty"`
}

// AuditStore manages the persistence of triage decisions
type AuditStore struct {
	mu         sync.RWMutex
	filePath   string
	legacyPath string
	Entries    map[string]AuditEntry `json:"entries"`
}

// NewAuditStore initializes a store, optionally loading from disk
func NewAuditStore(workspaceRoot string) *AuditStore {
	store := &AuditStore{
		filePath:   filepath.Join(workspaceRoot, "aitriage-reports", "history", "audit.json"),
		legacyPath: filepath.Join(workspaceRoot, ".aitriage-audit.json"),
		Entries:    make(map[string]AuditEntry),
	}
	_ = store.Load() // Best-effort: missing file is not an error
	return store
}

// Load reads the audit state from disk
func (s *AuditStore) Load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			// Read the old root-level state once for backwards compatibility,
			// but all future writes go to aitriage-reports/history/audit.json.
			data, err = os.ReadFile(s.legacyPath)
			if os.IsNotExist(err) {
				return nil
			}
		}
		if err != nil {
			return err
		}
	}

	if err := json.Unmarshal(data, s); err != nil {
		return err
	}
	if s.Entries == nil {
		s.Entries = make(map[string]AuditEntry)
	}
	return nil
}

// Save writes the audit state to disk
func (s *AuditStore) Save() error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}

	dir := filepath.Dir(s.filePath)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	if err := os.Chmod(dir, 0o700); err != nil {
		return err
	}

	tmp, err := os.CreateTemp(dir, ".audit-*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer func() { _ = os.Remove(tmpName) }()
	if err := tmp.Chmod(0o600); err != nil {
		_ = tmp.Close()
		return err
	}
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpName, s.filePath)
}

// GetKey creates a deterministic key for a finding
func GetAuditKey(ruleID, file string) string {
	return ruleID + ":" + file
}

// SetStatus updates the status of a finding and saves to disk
func (s *AuditStore) SetStatus(ruleID, file string, status AuditStatus, reason string) error {
	s.mu.Lock()

	key := GetAuditKey(ruleID, file)

	// If it's OPEN, we just remove it to save space
	if status == AuditStatusOpen {
		delete(s.Entries, key)
	} else {
		s.Entries[key] = AuditEntry{
			Status:    status,
			Reason:    reason,
			Timestamp: time.Now(),
		}
	}
	s.mu.Unlock()

	return s.Save()
}

// GetStatus returns the current status, defaulting to OPEN
func (s *AuditStore) GetStatus(ruleID, file string) AuditStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entry, exists := s.Entries[GetAuditKey(ruleID, file)]
	if !exists {
		return AuditStatusOpen
	}
	return entry.Status
}
