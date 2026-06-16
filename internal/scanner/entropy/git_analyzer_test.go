package entropy

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func runCmd(t *testing.T, dir string, name string, args ...string) {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to run %s %v: %v", name, args, err)
	}
}

func TestAnalyzeGitHistory(t *testing.T) {
	dir := t.TempDir()

	// Initialize git repo
	runCmd(t, dir, "git", "init")
	runCmd(t, dir, "git", "config", "user.email", "test@example.com")
	runCmd(t, dir, "git", "config", "user.name", "Test User")

	filePath := filepath.Join(dir, "test.txt")

	// Create a file with many lines to trigger the heuristic
	content := ""
	for i := 0; i < 600; i++ {
		content += "line\n"
	}
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	// Commit 1
	runCmd(t, dir, "git", "add", "test.txt")
	runCmd(t, dir, "git", "commit", "-m", "Initial commit")

	stats, err := AnalyzeGitHistory(filePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if stats.CommitCount != 1 {
		t.Errorf("expected 1 commit, got %d", stats.CommitCount)
	}
	if stats.AuthorCount != 1 {
		t.Errorf("expected 1 author, got %d", stats.AuthorCount)
	}
	// 600 additions in 1 commit -> avgAdd > 500 -> score 0.8
	if stats.AgenticScore != 0.8 {
		t.Errorf("expected score 0.8, got %f", stats.AgenticScore)
	}
}
