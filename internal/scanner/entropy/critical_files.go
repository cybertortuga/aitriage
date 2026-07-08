package entropy

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// CriticalFile represents a file identified as security-critical based on git metrics.
type CriticalFile struct {
	Path         string    `json:"path"`
	Risk         string    `json:"risk"`   // "HIGH", "MEDIUM", "LOW"
	Reason       string    `json:"reason"` // human-readable explanation
	CommitCount  int       `json:"commit_count"`
	AuthorCount  int       `json:"author_count"`
	LastModified time.Time `json:"last_modified"`
	LinesChanged int       `json:"lines_changed"`
}

// securityPatterns are filename patterns that indicate security-critical files.
var securityPatterns = []struct {
	pattern string
	reason  string
}{
	{".env", "Environment secrets file"},
	{"secret", "Contains 'secret' in name"},
	{"credential", "Contains 'credential' in name"},
	{"password", "Contains 'password' in name"},
	{"private", "Contains 'private' in name"},
	{".pem", "PEM certificate/key file"},
	{".key", "Private key file"},
	{".p12", "PKCS12 certificate file"},
	{".pfx", "PFX certificate file"},
	{"auth", "Authentication module"},
	{"middleware", "Middleware (auth/security)"},
	{"jwt", "JWT token handling"},
	{"session", "Session management"},
	{"nginx.conf", "Nginx configuration"},
	{"Dockerfile", "Container configuration"},
	{"docker-compose", "Container orchestration"},
}

// FindCriticalFiles scans the project for security-critical files using git metadata.
func FindCriticalFiles(projectPath string) []CriticalFile {
	// Check if git is available
	if _, err := exec.LookPath("git"); err != nil {
		return nil
	}
	// Check if this is a git repo
	if _, err := os.Stat(filepath.Join(projectPath, ".git")); os.IsNotExist(err) {
		return nil
	}

	var results []CriticalFile

	// 1. Find files matching security patterns
	results = append(results, findByPatterns(projectPath)...)

	// 2. Find high-churn files with few authors (potential AI dumps)
	results = append(results, findHighChurn(projectPath)...)

	// Deduplicate by path
	seen := make(map[string]bool)
	var deduped []CriticalFile
	for _, cf := range results {
		if !seen[cf.Path] {
			seen[cf.Path] = true
			deduped = append(deduped, cf)
		}
	}

	return deduped
}

// findByPatterns checks tracked files against security-sensitive name patterns.
func findByPatterns(projectPath string) []CriticalFile {
	cmd := exec.Command("git", "ls-files")
	cmd.Dir = projectPath
	out, err := cmd.Output()
	if err != nil {
		return nil
	}

	var results []CriticalFile
	for _, file := range strings.Split(string(out), "\n") {
		file = strings.TrimSpace(file)
		if file == "" {
			continue
		}
		lower := strings.ToLower(filepath.Base(file))
		for _, p := range securityPatterns {
			if strings.Contains(lower, strings.ToLower(p.pattern)) {
				cf := CriticalFile{
					Path:   file,
					Risk:   "MEDIUM",
					Reason: p.reason,
				}
				// Elevate risk for actual key/secret files
				if strings.HasSuffix(lower, ".pem") || strings.HasSuffix(lower, ".key") ||
					strings.HasSuffix(lower, ".p12") || lower == ".env" {
					cf.Risk = "HIGH"
				}
				// Get git stats
				stats := getFileGitStats(projectPath, file)
				cf.CommitCount = stats.commitCount
				cf.AuthorCount = stats.authorCount
				cf.LinesChanged = stats.linesChanged
				cf.LastModified = stats.lastModified
				results = append(results, cf)
				break
			}
		}
	}
	return results
}

type fileGitStats struct {
	commitCount  int
	authorCount  int
	linesChanged int
	lastModified time.Time
}

func getFileGitStats(projectPath, file string) fileGitStats {
	var stats fileGitStats

	// Commit count and authors
	cmd := exec.Command("git", "log", "--follow", "--pretty=format:%ae|%aI", "--", file)
	cmd.Dir = projectPath
	out, err := cmd.Output()
	if err != nil {
		return stats
	}

	authors := make(map[string]bool)
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		stats.commitCount++
		parts := strings.SplitN(line, "|", 2)
		if len(parts) >= 1 {
			authors[parts[0]] = true
		}
		if len(parts) >= 2 && stats.lastModified.IsZero() {
			stats.lastModified, _ = time.Parse(time.RFC3339, parts[1])
		}
	}
	stats.authorCount = len(authors)

	// Lines changed
	cmd = exec.Command("git", "log", "--numstat", "--pretty=format:", "--", file)
	cmd.Dir = projectPath
	out, err = cmd.Output()
	if err != nil {
		return stats
	}
	for _, line := range strings.Split(string(out), "\n") {
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			add, _ := strconv.Atoi(parts[0])
			del, _ := strconv.Atoi(parts[1])
			stats.linesChanged += add + del
		}
	}

	return stats
}

// findHighChurn finds files with excessive changes from few authors (potential AI dumps).
func findHighChurn(projectPath string) []CriticalFile {
	// Get shortlog to find files with most commits
	cmd := exec.Command("git", "log", "--pretty=format:", "--name-only")
	cmd.Dir = projectPath
	out, err := cmd.Output()
	if err != nil {
		return nil
	}

	fileCounts := make(map[string]int)
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			fileCounts[line]++
		}
	}

	var results []CriticalFile
	for file, count := range fileCounts {
		if count < 30 { // Only flag files with 30+ commits
			continue
		}
		stats := getFileGitStats(projectPath, file)
		if stats.authorCount <= 1 && stats.linesChanged > 1000 {
			results = append(results, CriticalFile{
				Path:         file,
				Risk:         "MEDIUM",
				Reason:       "High churn + single author — potential AI-generated code",
				CommitCount:  stats.commitCount,
				AuthorCount:  stats.authorCount,
				LastModified: stats.lastModified,
				LinesChanged: stats.linesChanged,
			})
		}
	}

	return results
}
