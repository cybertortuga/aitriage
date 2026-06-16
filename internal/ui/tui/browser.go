package tui

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// BrowserEntry represents a single file or directory in the browser view.
type BrowserEntry struct {
	Name        string
	Path        string // full absolute path
	IsDir       bool
	Size        int64
	ModTime     time.Time
	Children    int    // count of items inside (for dirs only)
	ProjectType string // detected: "Go", "Node.js", "Python", "Rust", "PHP", "Ruby", ""
	GitStatus   string // "M", "A", "??", etc
}

// scanDirectory reads a directory and returns sorted BrowserEntry items.
// Directories come first, then files. ".." is always first if not at root.
func scanDirectory(dir string) ([]BrowserEntry, error) {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}

	dirEntries, err := os.ReadDir(absDir)
	if err != nil {
		return nil, err
	}

	var entries []BrowserEntry

	// Add parent directory entry (unless at filesystem root)
	if absDir != "/" {
		entries = append(entries, BrowserEntry{
			Name:  "..",
			Path:  filepath.Dir(absDir),
			IsDir: true,
		})
	}

	var dirs, files []BrowserEntry

	statusMap := getGitStatus(absDir)

	for _, de := range dirEntries {
		// Skip hidden files/dirs (starting with .)
		if strings.HasPrefix(de.Name(), ".") {
			continue
		}

		fullPath := filepath.Join(absDir, de.Name())
		info, err := de.Info()
		if err != nil {
			continue
		}

		entry := BrowserEntry{
			Name:    de.Name(),
			Path:    fullPath,
			IsDir:   de.IsDir(),
			Size:    info.Size(),
			ModTime: info.ModTime(),
		}

		// Map git status
		if status, ok := statusMap[de.Name()]; ok {
			entry.GitStatus = strings.TrimSpace(status)
		} else if de.IsDir() {
			// Check if any child is modified
			prefix := de.Name() + "/"
			for k, v := range statusMap {
				if strings.HasPrefix(k, prefix) {
					entry.GitStatus = strings.TrimSpace(v)
					break
				}
			}
		}

		if de.IsDir() {
			entry.Children = countChildren(fullPath)
			entry.ProjectType = detectProjectType(fullPath)
			entry.Size = dirSize(fullPath)
			dirs = append(dirs, entry)
		} else {
			files = append(files, entry)
		}
	}

	// Sort dirs and files alphabetically
	sort.Slice(dirs, func(i, j int) bool {
		return strings.ToLower(dirs[i].Name) < strings.ToLower(dirs[j].Name)
	})
	sort.Slice(files, func(i, j int) bool {
		return strings.ToLower(files[i].Name) < strings.ToLower(files[j].Name)
	})

	entries = append(entries, dirs...)
	entries = append(entries, files...)

	return entries, nil
}

// detectProjectType checks for manifest files to determine project type.
func detectProjectType(dir string) string {
	checks := []struct {
		file    string
		project string
	}{
		{"go.mod", "Go"},
		{"package.json", "Node.js"},
		{"requirements.txt", "Python"},
		{"Pipfile", "Python"},
		{"pyproject.toml", "Python"},
		{"Cargo.toml", "Rust"},
		{"composer.json", "PHP"},
		{"Gemfile", "Ruby"},
		{"pom.xml", "Java"},
		{"build.gradle", "Java"},
		{"Dockerfile", "Docker"},
		{"docker-compose.yml", "Docker"},
		{"terraform.tf", "Terraform"},
		{"main.tf", "Terraform"},
	}

	for _, c := range checks {
		if _, err := os.Stat(filepath.Join(dir, c.file)); err == nil {
			return c.project
		}
	}
	return ""
}

// countChildren returns the number of non-hidden items in a directory.
// Quick scan — does not recurse.
func countChildren(dir string) int {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0
	}
	count := 0
	for _, e := range entries {
		if !strings.HasPrefix(e.Name(), ".") {
			count++
		}
	}
	return count
}

// dirSize estimates total size of a directory (non-recursive, top-level only for speed).
func dirSize(dir string) int64 {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0
	}
	var total int64
	for _, e := range entries {
		if info, err := e.Info(); err == nil {
			total += info.Size()
		}
	}
	return total
}

// formatSize converts bytes to human-readable format.
func formatSize(bytes int64) string {
	switch {
	case bytes >= 1<<30:
		return fmt.Sprintf("%.1f GB", float64(bytes)/float64(1<<30))
	case bytes >= 1<<20:
		return fmt.Sprintf("%.1f MB", float64(bytes)/float64(1<<20))
	case bytes >= 1<<10:
		return fmt.Sprintf("%.1f KB", float64(bytes)/float64(1<<10))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

// detectKeyFiles returns a list of notable files in the directory for the preview panel.
func detectKeyFiles(dir string) []string {
	important := []string{
		"go.mod", "package.json", "requirements.txt", "Cargo.toml",
		"Dockerfile", "docker-compose.yml", ".env", ".env.example",
		"Makefile", "README.md", ".gitignore", "nginx.conf",
		"tsconfig.json", "webpack.config.js", "vite.config.ts",
	}

	var found []string
	for _, f := range important {
		if _, err := os.Stat(filepath.Join(dir, f)); err == nil {
			found = append(found, f)
		}
	}
	return found
}

func getGitStatus(dir string) map[string]string {
	statusMap := make(map[string]string)
	cmd := exec.Command("git", "status", "--porcelain=v1", "-z")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return statusMap
	}

	parts := strings.Split(string(out), "\x00")
	for i := 0; i < len(parts); i++ {
		if len(parts[i]) < 4 {
			continue
		}
		status := parts[i][:2]
		filePath := parts[i][3:]

		// If it's a rename, skip the next part (which is the original filename)
		if status[0] == 'R' || status[0] == 'C' {
			i++
		}

		statusMap[filePath] = status
	}
	return statusMap
}
