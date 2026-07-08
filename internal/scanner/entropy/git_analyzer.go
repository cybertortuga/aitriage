package entropy

import (
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// GitFileStats contains risk metrics from version control.
type GitFileStats struct {
	FilePath     string
	CommitCount  int
	AuthorCount  int
	AgenticScore float64 // 0..1 based on change patterns
}

// AnalyzeGitHistory runs 'git log' to detect risk patterns.
func AnalyzeGitHistory(filePath string) (GitFileStats, error) {
	// Formula for "Agentic Intensity":
	// If a file has massive additions in very few commits, it's likely an AI dump.

	dir := filepath.Dir(filePath)
	base := filepath.Base(filePath)

	cmd := exec.Command("git", "log", "--follow", "--pretty=format:%ae", "--", base)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return GitFileStats{}, err
	}

	authors := strings.Split(string(out), "\n")
	authorMap := make(map[string]bool)
	for _, a := range authors {
		if a != "" {
			authorMap[a] = true
		}
	}

	// Get stats on line changes
	cmd = exec.Command("git", "log", "--numstat", "--pretty=format:", "--", base)
	cmd.Dir = dir
	out, err = cmd.Output()
	if err != nil {
		return GitFileStats{}, err
	}

	lines := strings.Split(string(out), "\n")
	totalAdditions := 0
	for _, line := range lines {
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			add, _ := strconv.Atoi(parts[0])
			totalAdditions += add
		}
	}

	commitCount := len(authors)
	authorCount := len(authorMap)

	// Heuristic: If additions per commit > 500, it's risky
	agenticScore := 0.0
	if commitCount > 0 {
		avgAdd := float64(totalAdditions) / float64(commitCount)
		if avgAdd > 500 {
			agenticScore = 0.8
		} else if avgAdd > 200 {
			agenticScore = 0.4
		}
	}

	return GitFileStats{
		FilePath:     filePath,
		CommitCount:  commitCount,
		AuthorCount:  authorCount,
		AgenticScore: agenticScore,
	}, nil
}
