package deployaudit

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAuditDockerfile(t *testing.T) {
	content := `
FROM ubuntu:latest
USER root
ADD http://example.com/file /tmp/
ENV PASSWORD=secret123
`
	dir := t.TempDir()
	path := filepath.Join(dir, "Dockerfile")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	findings, err := auditDockerfile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(findings) != 3 {
		t.Fatalf("expected 3 findings, got %d", len(findings))
	}

	issues := map[string]bool{}
	for _, f := range findings {
		issues[f.Issue] = true
	}

	expected := []string{"dockerfile_root_user", "dockerfile_add_instead_of_copy", "dockerfile_hardcoded_secret"}
	for _, exp := range expected {
		if !issues[exp] {
			t.Errorf("missing expected issue: %s", exp)
		}
	}
}

func TestAuditCompose(t *testing.T) {
	content := `
version: '3'
services:
  web:
    image: nginx
    privileged: true
    environment:
      DATABASE_PASSWORD: supersecret
      SAFE_VAR: ${SAFE_VAR}
`
	dir := t.TempDir()
	path := filepath.Join(dir, "docker-compose.yml")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	findings, err := auditCompose(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(findings) != 2 {
		t.Fatalf("expected 2 findings, got %d", len(findings))
	}

	issues := map[string]bool{}
	for _, f := range findings {
		issues[f.Issue] = true
	}

	expected := []string{"compose_privileged_mode", "compose_hardcoded_secret"}
	for _, exp := range expected {
		if !issues[exp] {
			t.Errorf("missing expected issue: %s", exp)
		}
	}
}

func TestAuditKubernetes(t *testing.T) {
	content := `
apiVersion: v1
kind: Pod
metadata:
  name: mypod
spec:
  securityContext:
    runAsUser: 0
  containers:
  - name: mycontainer
    image: nginx
    securityContext:
      privileged: true
`
	dir := t.TempDir()
	path := filepath.Join(dir, "deployment.yaml")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	findings, err := auditKubernetes(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(findings) != 2 {
		t.Fatalf("expected 2 findings, got %d", len(findings))
	}

	issues := map[string]bool{}
	for _, f := range findings {
		issues[f.Issue] = true
	}

	expected := []string{"k8s_run_as_root", "k8s_privileged_container"}
	for _, exp := range expected {
		if !issues[exp] {
			t.Errorf("missing expected issue: %s", exp)
		}
	}
}

func TestAuditNginx(t *testing.T) {
	content := `
user root;
worker_processes auto;
http {
    # does not have tokens off
    # does not have x-frame-opts
    # does not have csp
}
`
	dir := t.TempDir()
	path := filepath.Join(dir, "nginx.conf")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	findings, err := auditNginx(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(findings) != 4 {
		t.Fatalf("expected 4 findings, got %d", len(findings))
	}

	issues := map[string]bool{}
	for _, f := range findings {
		issues[f.Issue] = true
	}

	expected := []string{
		"nginx_runs_as_root",
		"nginx_missing_server_tokens_off",
		"nginx_missing_xframe_options",
		"nginx_missing_csp",
	}
	for _, exp := range expected {
		if !issues[exp] {
			t.Errorf("missing expected issue: %s", exp)
		}
	}
}

func TestAuditMakefile(t *testing.T) {
	content := `
build:
	echo "Building..."
DB_PASSWORD = mysecretpass
SAFE_TOKEN = $(MY_ENV_TOKEN)
`
	dir := t.TempDir()
	path := filepath.Join(dir, "Makefile")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	findings, err := auditMakefile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}

	if findings[0].Issue != "makefile_hardcoded_secret" {
		t.Errorf("expected makefile_hardcoded_secret, got %s", findings[0].Issue)
	}
}
