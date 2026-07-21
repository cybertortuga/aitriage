package agentcontext

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestBuildRepoContextIgnoresRuntimeState(t *testing.T) {
	root := t.TempDir()
	mustWriteContextTestFile(t, filepath.Join(root, "package.json"), `{"name":"fixture"}`)
	mustWriteContextTestFile(t, filepath.Join(root, "auth.ts"), "export function auth() {}\n")

	serviceDirs := []string{
		"aitriage-reports",
		".aitriage",
		".aitriage-cache",
		".codex",
		".claude",
	}
	for _, dir := range serviceDirs {
		mustWriteContextTestFile(t, filepath.Join(root, dir, "security-handler.ts"), "runtime state\n")
	}

	before := BuildRepoContext(root)
	if before == nil {
		t.Fatal("BuildRepoContext returned nil")
	}
	assertNoRuntimeState(t, before, serviceDirs)

	// Model a deferred host turn: request/response/audit files appear while the
	// same graph.Run is replayed. Repository context must remain byte-stable.
	mustWriteContextTestFile(t, filepath.Join(root, "aitriage-reports", "run-1", "requests", "request.json"), `{"request":"one"}`)
	mustWriteContextTestFile(t, filepath.Join(root, "aitriage-reports", "run-1", "responses", "response.json"), `{"response":"one"}`)
	mustWriteContextTestFile(t, filepath.Join(root, "aitriage-reports", "run-1", "audit.log"), "turn completed\n")

	after := BuildRepoContext(root)
	if !reflect.DeepEqual(before, after) {
		t.Fatalf("repository context changed after runtime-state write\nbefore: %#v\nafter:  %#v", before, after)
	}
	assertNoRuntimeState(t, after, serviceDirs)
}

func assertNoRuntimeState(t *testing.T, ctx *RepoContext, serviceDirs []string) {
	t.Helper()
	for _, dir := range serviceDirs {
		if strings.Contains(ctx.ProjectTree, dir) {
			t.Errorf("project tree contains runtime directory %q", dir)
		}
		for _, file := range ctx.KeyFiles {
			if file.Path == dir || strings.HasPrefix(file.Path, dir+string(os.PathSeparator)) {
				t.Errorf("key files contain runtime path %q", file.Path)
			}
		}
	}
}

func mustWriteContextTestFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}
