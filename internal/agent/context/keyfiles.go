package agentcontext

import (
	"os"
	"path/filepath"
	"strings"
)

// KeyFile holds the path, content, and role of a key project file.
type KeyFile struct {
	Path    string `json:"path"`
	Content string `json:"content"`
	Role    string `json:"role"` // "manifest", "config", "entrypoint", "security", "routing"
}

// manifestFiles are package manager manifest files.
var manifestFiles = []struct {
	name string
	role string
}{
	{"package.json", "manifest"},
	{"requirements.txt", "manifest"},
	{"go.mod", "manifest"},
	{"Cargo.toml", "manifest"},
	{"Gemfile", "manifest"},
	{"composer.json", "manifest"},
	{"pyproject.toml", "manifest"},
	{"Pipfile", "manifest"},
	{"pom.xml", "manifest"},
	{"build.gradle", "manifest"},
}

// configFiles are infrastructure and build config files.
var configFiles = []struct {
	name string
	role string
}{
	{"Dockerfile", "config"},
	{"docker-compose.yml", "config"},
	{"docker-compose.yaml", "config"},
	{".env.example", "config"},
	{"tsconfig.json", "config"},
	{"vite.config.ts", "config"},
	{"vite.config.js", "config"},
	{"webpack.config.js", "config"},
	{"next.config.js", "config"},
	{"next.config.ts", "config"},
	{"nginx.conf", "config"},
	{".aitriage.yaml", "security"},
	{".aitriage.yml", "security"},
}

// entrypointPatterns are common entrypoint file names per language.
var entrypointPatterns = []string{
	"main.go", "main.py", "app.py", "asgi.py", "wsgi.py", "manage.py",
	"index.ts", "index.js", "server.ts", "server.js", "app.ts", "app.js",
	"main.ts", "main.js",
}

// securityKeywords in filenames indicate auth/security/routing files.
var securityKeywords = []string{
	"auth", "login", "session", "middleware", "permission",
	"security", "guard", "policy", "rbac", "acl",
	"route", "router", "handler", "controller", "endpoint",
}

// ReadKeyFiles discovers and reads important project files.
// Files are categorized by role and truncated to maxFileSize bytes.
// Returns files ordered by priority: manifests > entrypoints > security > configs.
func ReadKeyFiles(projectPath string, maxFileSize int) []KeyFile {
	if maxFileSize <= 0 {
		maxFileSize = 8192
	}

	absRoot, err := filepath.Abs(projectPath)
	if err != nil {
		return nil
	}

	var result []KeyFile

	// 1. Manifest files (top-level).
	for _, mf := range manifestFiles {
		if content := readFileCapped(filepath.Join(absRoot, mf.name), maxFileSize); content != "" {
			result = append(result, KeyFile{
				Path:    mf.name,
				Content: content,
				Role:    mf.role,
			})
		}
	}

	// 2. Config files (top-level).
	for _, cf := range configFiles {
		if content := readFileCapped(filepath.Join(absRoot, cf.name), maxFileSize); content != "" {
			result = append(result, KeyFile{
				Path:    cf.name,
				Content: content,
				Role:    cf.role,
			})
		}
	}

	// 3. Entrypoint files (search up to 3 levels deep).
	found := findEntrypoints(absRoot, 3)
	for _, ep := range found {
		relPath, _ := filepath.Rel(absRoot, ep)
		if content := readFileCapped(ep, maxFileSize); content != "" {
			result = append(result, KeyFile{
				Path:    relPath,
				Content: content,
				Role:    "entrypoint",
			})
		}
	}

	// 4. Security/auth/routing files (search up to 4 levels deep, max 10 files).
	secFiles := findSecurityFiles(absRoot, 4, 10)
	for _, sf := range secFiles {
		relPath, _ := filepath.Rel(absRoot, sf)
		// Skip if already added as entrypoint.
		if containsPath(result, relPath) {
			continue
		}
		if content := readFileCapped(sf, maxFileSize); content != "" {
			result = append(result, KeyFile{
				Path:    relPath,
				Content: content,
				Role:    "security",
			})
		}
	}

	return result
}

// findEntrypoints searches for common entrypoint files up to maxDepth levels.
func findEntrypoints(root string, maxDepth int) []string {
	var found []string
	entrySet := make(map[string]bool)
	for _, ep := range entrypointPatterns {
		entrySet[ep] = true
	}

	_ = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			if defaultSkipDirs[info.Name()] {
				return filepath.SkipDir
			}
			relPath, _ := filepath.Rel(root, path)
			depth := strings.Count(relPath, string(os.PathSeparator)) + 1
			if relPath != "." && depth > maxDepth {
				return filepath.SkipDir
			}
			return nil
		}

		if entrySet[info.Name()] {
			found = append(found, path)
		}

		// Also check for cmd/*/main.go pattern (Go convention).
		relPath, _ := filepath.Rel(root, path)
		if strings.HasPrefix(relPath, "cmd"+string(os.PathSeparator)) && info.Name() == "main.go" {
			if !containsStr(found, path) {
				found = append(found, path)
			}
		}

		return nil
	})

	return found
}

// findSecurityFiles searches for files with security-related names.
func findSecurityFiles(root string, maxDepth, maxResults int) []string {
	var found []string

	codeExts := map[string]bool{
		".go": true, ".py": true, ".js": true, ".ts": true,
		".jsx": true, ".tsx": true, ".rb": true, ".java": true,
		".rs": true, ".php": true, ".cs": true,
	}

	_ = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || len(found) >= maxResults {
			return nil
		}
		if info.IsDir() {
			if defaultSkipDirs[info.Name()] {
				return filepath.SkipDir
			}
			relPath, _ := filepath.Rel(root, path)
			depth := strings.Count(relPath, string(os.PathSeparator)) + 1
			if relPath != "." && depth > maxDepth {
				return filepath.SkipDir
			}
			return nil
		}

		ext := strings.ToLower(filepath.Ext(info.Name()))
		if !codeExts[ext] {
			return nil
		}

		baseLower := strings.ToLower(info.Name())
		for _, kw := range securityKeywords {
			if strings.Contains(baseLower, kw) {
				found = append(found, path)
				break
			}
		}

		return nil
	})

	return found
}

// readFileCapped reads a file and returns at most maxBytes of content.
func readFileCapped(path string, maxBytes int) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	if len(data) > maxBytes {
		return string(data[:maxBytes]) + "\n... (truncated)"
	}
	return string(data)
}

func containsPath(files []KeyFile, path string) bool {
	for _, f := range files {
		if f.Path == path {
			return true
		}
	}
	return false
}

func containsStr(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}
