package diff

import (
	"os/exec"
	"path/filepath"
	"strings"
)

// ──────────────────────────────────────────────────────────────────────────────
// Git-based file filtering for incremental scanning.
//
// Usage:
//   aitriage scan . --diff HEAD~1       → Only files changed since last commit
//   aitriage scan . --diff origin/main  → Only files changed vs main branch
//   aitriage scan . --staged            → Only files staged for commit
// ──────────────────────────────────────────────────────────────────────────────

// ChangedFiles returns absolute paths of files changed between the working tree and a git ref.
// The ref can be any valid git reference: HEAD~1, origin/main, a commit SHA, etc.
func ChangedFiles(projectPath, ref string) ([]string, error) {
	cmd := exec.Command("git", "diff", "--name-only", "--diff-filter=ACMR", ref)
	cmd.Dir = projectPath
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return parseFileList(projectPath, string(out)), nil
}

// StagedFiles returns absolute paths of files currently staged in the git index.
func StagedFiles(projectPath string) ([]string, error) {
	cmd := exec.Command("git", "diff", "--cached", "--name-only", "--diff-filter=ACMR")
	cmd.Dir = projectPath
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return parseFileList(projectPath, string(out)), nil
}

// parseFileList converts newline-separated relative paths to absolute paths.
func parseFileList(projectPath, output string) []string {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	var files []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		files = append(files, filepath.Join(projectPath, line))
	}
	return files
}
