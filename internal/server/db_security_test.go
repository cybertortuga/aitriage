package server

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInitDBDoesNotInterpretProjectUsersJSON(t *testing.T) {
	project := t.TempDir()
	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(project); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(oldWD) })
	projectUsers := filepath.Join(project, "users.json")
	if err := os.WriteFile(projectUsers, []byte(`[{"username":"source-file-must-not-be-auth-config","password":"x","is_admin":true}]`), 0o644); err != nil {
		t.Fatal(err)
	}

	dbDir := filepath.Join(t.TempDir(), "aitriage-reports", "web")
	if err := os.MkdirAll(dbDir, 0o700); err != nil {
		t.Fatal(err)
	}
	db, err := InitDB(filepath.Join(dbDir, "aitriage.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	var count int
	if err := db.QueryRow(`SELECT count(*) FROM users WHERE username = ?`, "source-file-must-not-be-auth-config").Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatal("project users.json was incorrectly imported as Web authentication state")
	}
	if _, err := os.Stat(projectUsers); err != nil {
		t.Fatalf("project users.json was modified or renamed: %v", err)
	}
}
