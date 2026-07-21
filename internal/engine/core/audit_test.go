package core

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestAuditStoreWritesOnlyUnderReports(t *testing.T) {
	root := t.TempDir()
	store := NewAuditStore(root)
	if err := store.SetStatus("RULE-1", "src/app.go", AuditStatusIgnored, "confirmed false positive"); err != nil {
		t.Fatal(err)
	}

	path := filepath.Join(root, "aitriage-reports", "history", "audit.json")
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("audit state not written under reports: %v", err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Errorf("audit mode = %o, want 600", info.Mode().Perm())
	}
	if _, err := os.Stat(filepath.Join(root, ".aitriage-audit.json")); !os.IsNotExist(err) {
		t.Fatalf("legacy root audit file must not be created: %v", err)
	}
}

func TestAuditStoreReadsLegacyButMigratesOnWrite(t *testing.T) {
	root := t.TempDir()
	legacy := filepath.Join(root, ".aitriage-audit.json")
	payload, err := json.Marshal(AuditStore{Entries: map[string]AuditEntry{
		GetAuditKey("OLD", "main.go"): {Status: AuditStatusTriage},
	}})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(legacy, payload, 0o600); err != nil {
		t.Fatal(err)
	}

	store := NewAuditStore(root)
	if got := store.GetStatus("OLD", "main.go"); got != AuditStatusTriage {
		t.Fatalf("legacy status = %q, want TRIAGE", got)
	}
	if err := store.SetStatus("NEW", "main.go", AuditStatusIgnored, "migrated"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, "aitriage-reports", "history", "audit.json")); err != nil {
		t.Fatalf("new audit state not created: %v", err)
	}
}
