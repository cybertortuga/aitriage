package core_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cybertortuga/aitriage/internal/engine/core"
)

func TestNewWorkspace(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a regular go file
	if err := os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte("package main"), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	// Create a test file
	if err := os.WriteFile(filepath.Join(tmpDir, "main_test.go"), []byte("package main_test"), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	// Create an ignored directory and file
	nodeModulesDir := filepath.Join(tmpDir, "node_modules")
	if err := os.Mkdir(nodeModulesDir, 0755); err != nil {
		t.Fatalf("failed to create node_modules dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(nodeModulesDir, "ignored.js"), []byte("console.log('ignored')"), 0644); err != nil {
		t.Fatalf("failed to write ignored.js: %v", err)
	}

	// Create an ignored file by extension
	if err := os.WriteFile(filepath.Join(tmpDir, "image.png"), []byte("fake image data"), 0644); err != nil {
		t.Fatalf("failed to write image.png: %v", err)
	}

	ws, err := core.NewWorkspace(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create workspace: %v", err)
	}
	defer ws.Close()

	if len(ws.Files) != 2 {
		t.Errorf("Expected 2 files (main.go, main_test.go), got %d", len(ws.Files))
	}

	hasMain := false
	hasTest := false
	for _, f := range ws.Files {
		if strings.HasSuffix(f.Path, "main.go") {
			hasMain = true
			if f.IsTest {
				t.Error("main.go should not be marked as test")
			}
		}
		if strings.HasSuffix(f.Path, "main_test.go") {
			hasTest = true
			if !f.IsTest {
				t.Error("main_test.go should be marked as test")
			}
		}
	}

	if !hasMain || !hasTest {
		t.Error("Expected to find main.go and main_test.go in workspace files")
	}
}

func TestFileInfo_GetContent(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.go")
	contentStr := "package main\n\nfunc main() {}"
	if err := os.WriteFile(filePath, []byte(contentStr), 0644); err != nil {
		t.Fatalf("failed to write snippet test file: %v", err)
	}

	fi := &core.FileInfo{
		Path:      filePath,
		Extension: ".go",
	}

	content, err := fi.GetContent()
	if err != nil {
		t.Fatalf("GetContent failed: %v", err)
	}

	if string(content) != contentStr {
		t.Errorf("Expected content %q, got %q", contentStr, string(content))
	}

	if fi.IsBinary {
		t.Error("Expected IsBinary to be false for a text file")
	}
}

func TestProjectContext_FindFilesByExtension(t *testing.T) {
	ctx := &core.ProjectContext{
		Files: []*core.FileInfo{
			{Path: "test.go", Extension: ".go"},
			{Path: "test.js", Extension: ".js"},
			{Path: "test.py", Extension: ".py"},
			{Path: "test2.go", Extension: ".go"},
			{Path: "test3.Go", Extension: ".Go"},
		},
	}

	tests := []struct {
		name       string
		extensions []string
		wantCount  int
		wantPaths  []string
	}{
		{
			name:       "Single extension",
			extensions: []string{".go"},
			wantCount:  3,
			wantPaths:  []string{"test.go", "test2.go", "test3.Go"},
		},
		{
			name:       "Multiple extensions",
			extensions: []string{".go", ".js"},
			wantCount:  4,
			wantPaths:  []string{"test.go", "test.js", "test2.go", "test3.Go"},
		},
		{
			name:       "Case insensitive search",
			extensions: []string{".GO"},
			wantCount:  3,
			wantPaths:  []string{"test.go", "test2.go", "test3.Go"},
		},
		{
			name:       "Case insensitive file extension",
			extensions: []string{".go"},
			wantCount:  3,
			wantPaths:  []string{"test.go", "test2.go", "test3.Go"},
		},
		{
			name:       "Unmatched extension",
			extensions: []string{".rb"},
			wantCount:  0,
			wantPaths:  []string{},
		},
		{
			name:       "Empty extensions list",
			extensions: []string{},
			wantCount:  0,
			wantPaths:  []string{},
		},
		{
			name:       "Duplicate extensions provided",
			extensions: []string{".go", ".go"},
			wantCount:  3,
			wantPaths:  []string{"test.go", "test2.go", "test3.Go"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ctx.FindFilesByExtension(tt.extensions...)

			// If tt.wantCount is 0, we can accept either nil or empty slice
			if tt.wantCount == 0 {
				if len(got) != 0 {
					t.Errorf("FindFilesByExtension() got %v, want empty or nil", got)
				}
				return
			}

			if len(got) != tt.wantCount {
				t.Errorf("FindFilesByExtension() returned %v files, want %v", len(got), tt.wantCount)
			}

			// Verification of paths
			if len(got) == len(tt.wantPaths) {
				for i, path := range tt.wantPaths {
					if got[i].Path != path {
						t.Errorf("FindFilesByExtension() got path %v, want %v", got[i].Path, path)
					}
				}
			}
		})
	}

	t.Run("Empty files in context", func(t *testing.T) {
		emptyCtx := &core.ProjectContext{Files: []*core.FileInfo{}}
		got := emptyCtx.FindFilesByExtension(".go")
		if len(got) != 0 {
			t.Errorf("Expected empty result for empty context files, got %v", got)
		}
	})

	t.Run("Nil files in context", func(t *testing.T) {
		nilCtx := &core.ProjectContext{Files: nil}
		got := nilCtx.FindFilesByExtension(".go")
		if len(got) != 0 {
			t.Errorf("Expected empty result for nil context files, got %v", got)
		}
	})

}

func TestProjectContext_GetFile(t *testing.T) {
	ctx := &core.ProjectContext{
		RootPath: "/project",
		Files: []*core.FileInfo{
			{Path: "/project/src/main.go", Extension: ".go"},
		},
	}

	f := ctx.GetFile("src/main.go")
	if f == nil {
		t.Fatal("Expected to find file, got nil")
	}

	if f.Path != "/project/src/main.go" {
		t.Errorf("Expected path /project/src/main.go, got %s", f.Path)
	}

	// test caching
	f2 := ctx.GetFile("src/main.go")
	if f2 != f {
		t.Error("Expected same pointer from cache")
	}
}

func TestFileInfo_ParsingMethods(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.go")

	contentStr := `package main

// This is a comment
func main() {
	x := "this is a string"
	_ = x
}`
	if err := os.WriteFile(filePath, []byte(contentStr), 0644); err != nil {
		t.Fatalf("failed to write test file for snippet formatting: %v", err)
	}

	fi := &core.FileInfo{
		Path:      filePath,
		Extension: ".go",
	}
	defer fi.Close()

	tree, err := fi.GetTree()
	if err != nil {
		t.Fatalf("GetTree failed: %v", err)
	}
	if tree == nil {
		t.Error("Expected tree to be non-nil")
	}

	// Test caching of tree
	tree2, _ := fi.GetTree()
	if tree != tree2 {
		t.Error("Expected tree to be cached")
	}

	stripped, err := fi.GetStrippedContent()
	if err != nil {
		t.Fatalf("GetStrippedContent failed: %v", err)
	}

	if strings.Contains(string(stripped), "This is a comment") || strings.Contains(string(stripped), "this is a string") {
		t.Errorf("Stripped content still contains comments/strings: %s", string(stripped))
	}
	if !strings.Contains(string(stripped), "func main()") {
		t.Errorf("Stripped content missing code: %s", string(stripped))
	}

	docs, err := fi.GetDocsOnlyContent()
	if err != nil {
		t.Fatalf("GetDocsOnlyContent failed: %v", err)
	}

	if !strings.Contains(string(docs), "This is a comment") || !strings.Contains(string(docs), "this is a string") {
		t.Errorf("Docs content missing comments/strings: %s", string(docs))
	}
}
