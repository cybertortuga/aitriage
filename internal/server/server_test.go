package server

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/cybertortuga/aitriage/internal/agent/llm"
	"github.com/cybertortuga/aitriage/internal/server/middleware"
	"github.com/golang-jwt/jwt/v5"
	_ "modernc.org/sqlite"
	"time"
)

func setupTestServer(t *testing.T) *Server {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open in-memory db: %v", err)
	}
	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("failed to run schema: %v", err)
	}
	return NewServer("", db)
}

func addAuthCookie(req *http.Request) {
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &middleware.Claims{
		UserID:   1,
		Username: "admin",
		Role:     "superadmin",
		IsAdmin:  true,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}
	tokenString, _ := middleware.SignToken(claims)
	req.AddCookie(&http.Cookie{
		Name:  "token",
		Value: tokenString,
	})
}

type summaryCaptureLLM struct {
	messages []llm.Message
	response string
}

func (m *summaryCaptureLLM) Chat(ctx context.Context, messages []llm.Message) (string, llm.Usage, error) {
	m.messages = append([]llm.Message(nil), messages...)
	return m.response, llm.Usage{}, nil
}

func TestNewServer(t *testing.T) {
	hostPrefix := "/custom-host"
	db, _ := sql.Open("sqlite", ":memory:")
	s := NewServer(hostPrefix, db)

	if s == nil {
		t.Fatal("expected NewServer to return a non-nil Server")
	}

	if s.hostPrefix != hostPrefix {
		t.Errorf("expected hostPrefix to be %q, got %q", hostPrefix, s.hostPrefix)
	}

	// Verify routes are registered by hitting an API endpoint
	req, err := http.NewRequest("GET", "/api/health", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	rr := httptest.NewRecorder()
	s.ServeHTTP(rr, req)

	// We expect a 200 OK because the route is registered and the health handler responds successfully.
	if rr.Code != http.StatusOK {
		t.Errorf("expected /api/health to return 200 OK, got %d", rr.Code)
	}

	// Verify static UI route is registered
	reqUI, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	rrUI := httptest.NewRecorder()
	s.ServeHTTP(rrUI, reqUI)

	if rrUI.Code != http.StatusOK {
		t.Errorf("expected / to return 200 OK, got %d", rrUI.Code)
	}

	// Verify unknown route
	reqUnknown, err := http.NewRequest("GET", "/api/unknown", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	addAuthCookie(reqUnknown)
	rrUnknown := httptest.NewRecorder()
	s.ServeHTTP(rrUnknown, reqUnknown)

	if rrUnknown.Code != http.StatusOK {
		t.Errorf("expected unknown route to return 200 OK (fallback), got %d", rrUnknown.Code)
	}
}

func TestHandleCreateUserValidation(t *testing.T) {
	s := setupTestServer(t)

	// Admin token
	req, _ := http.NewRequest("POST", "/api/admin/users", bytes.NewBuffer([]byte(`{}`)))
	addAuthCookie(req)
	rr := httptest.NewRecorder()
	s.ServeHTTP(rr, req)

	// Since we added validation for Username, empty username returns 400.
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for missing username (validation), got %d", rr.Code)
	}

	// Invalid ID query for delete
	reqDel, _ := http.NewRequest("DELETE", "/api/admin/users?id=abc", nil)
	addAuthCookie(reqDel)
	rrDel := httptest.NewRecorder()
	s.ServeHTTP(rrDel, reqDel)

	if rrDel.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid ID, got %d", rrDel.Code)
	}
}

func TestHandleSummaryUsesSecureCoderEvidenceContract(t *testing.T) {
	s := setupTestServer(t)
	projectDir := t.TempDir()

	if err := os.WriteFile(filepath.Join(projectDir, "package.json"), []byte(`{
  "name": "ait-dodo-landing",
  "scripts": {"build": "tsc -b && vite build"},
  "dependencies": {"react": "19.2.6", "react-dom": "19.2.6"},
  "devDependencies": {"vite": "8.0.12", "@vitejs/plugin-react": "6.0.1"}
}`), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(projectDir, "vite.config.ts"), []byte("export default {}\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(projectDir, "src"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(projectDir, "src", "main.tsx"), []byte("import React from 'react'\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(projectDir, "src", "App.tsx"), []byte("export default function App() { return null }\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(projectDir, "SECURITY_TRIAGE_REPORT.md"), []byte(`# Security triage report

Current application shape: static Vite + React landing page. The repository has no backend server, API routes, auth middleware, database client, form submission endpoint, cookies, or user-controlled HTML rendering.

Authentication Middleware Missing: false positive for the current repository.
Rate Limiting Missing: false positive for the current repository.
`), 0644); err != nil {
		t.Fatal(err)
	}

	res, err := s.db.Exec(`INSERT INTO products (name, repo_url, tech_stack, business_criticality, lifecycle) VALUES (?, ?, ?, ?, ?)`, "ait-dodo-landing", projectDir, "Vite + React", "medium", "production")
	if err != nil {
		t.Fatalf("failed to insert product: %v", err)
	}
	productID, _ := res.LastInsertId()
	res, err = s.db.Exec(`INSERT INTO engagements (product_id, name, scan_path, status) VALUES (?, ?, ?, ?)`, productID, "scan", projectDir, "completed")
	if err != nil {
		t.Fatalf("failed to insert engagement: %v", err)
	}
	engagementID, _ := res.LastInsertId()

	_, err = s.db.Exec(`
		INSERT INTO findings (engagement_id, product_id, rule_id, title, severity, file_path, line_number, description, fix_suggestion, status, kanban_column, stack, is_false_positive, fp_reason)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, engagementID, productID, "AUTH-MIDDLEWARE", "Authentication Middleware Missing", "CRITICAL", "", 0, "Project-level auth rule", "No backend/API route evidence in this static landing page.", "false_positive", "done", "nfr", 1, "static landing page")
	if err != nil {
		t.Fatalf("failed to insert false positive finding: %v", err)
	}
	_, err = s.db.Exec(`
		INSERT INTO findings (engagement_id, product_id, rule_id, title, severity, file_path, line_number, description, fix_suggestion, status, kanban_column, stack)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, engagementID, productID, "SEC-HEADERS", "Missing Security Headers", "HIGH", "vite.config.ts", 10, "Preview server headers need verification.", "Set CSP and related headers.", "open", "backlog", "deploy")
	if err != nil {
		t.Fatalf("failed to insert active finding: %v", err)
	}

	capture := &summaryCaptureLLM{response: "**Статус безопасности:** evidence-based"}
	s.llmClient = capture

	req, _ := http.NewRequest("GET", "/api/ai-summary?product_id="+strconv.FormatInt(productID, 10)+"&lang=ru", nil)
	addAuthCookie(req)
	rr := httptest.NewRecorder()
	s.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	var resp map[string]any
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp["summary"] != capture.response {
		t.Fatalf("unexpected summary response: %#v", resp)
	}
	if len(capture.messages) != 2 {
		t.Fatalf("expected two LLM messages, got %d", len(capture.messages))
	}

	systemPrompt := capture.messages[0].Content
	for _, want := range []string{
		"Do not infer backend/API/auth/rate-limit risk",
		"Scanner titles are hypotheses, not proof",
		"Respond in Russian",
	} {
		if !strings.Contains(systemPrompt, want) {
			t.Fatalf("system prompt missing %q:\n%s", want, systemPrompt)
		}
	}

	userPrompt := capture.messages[1].Content
	for _, want := range []string{
		"Product: ait-dodo-landing",
		"Active findings: 1",
		"Non-active/suppressed findings: 1",
		"Detected project markers: package.json, vite.config.ts, src/main.tsx, src/App.tsx",
		"Backend dependency markers: none",
		"Server/API route file markers: none",
		"[non-active][CRITICAL][AUTH-MIDDLEWARE] Authentication Middleware Missing",
		"status=false_positive",
		"[active][HIGH][SEC-HEADERS] Missing Security Headers at `vite.config.ts:10`",
		"Current application shape: static Vite + React landing page",
	} {
		if !strings.Contains(userPrompt, want) {
			t.Fatalf("user prompt missing %q:\n%s", want, userPrompt)
		}
	}
}

func TestHandleHealth(t *testing.T) {
	s := setupTestServer(t)

	req, err := http.NewRequest("GET", "/api/health", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	s.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if ok, exists := resp["ok"].(bool); !exists || !ok {
		t.Errorf("expected ok: true, got %v", resp)
	}
}

func TestHandleScan(t *testing.T) {
	s := setupTestServer(t)
	tempDir := t.TempDir()

	// 1. Test empty body
	req, _ := http.NewRequest("POST", "/api/scan", bytes.NewBuffer([]byte{}))
	addAuthCookie(req)
	rr := httptest.NewRecorder()
	s.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 Bad Request for empty body, got %d", rr.Code)
	}

	// 2. Test valid request
	body := []byte(`{"path":"` + tempDir + `", "external": false}`)
	req, _ = http.NewRequest("POST", "/api/scan", bytes.NewBuffer(body))
	addAuthCookie(req)
	rr = httptest.NewRecorder()
	s.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var resp scanResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	// Ok field is lowercase in the struct (Ok, not OK)
	if !resp.Ok {
		t.Errorf("expected Ok to be true, got error: %s", resp.Error)
	}
	if resp.ScanID == "" {
		t.Errorf("expected ScanID to be populated")
	}
	// scanResponse does not include Path — verify via scan_id presence
	if resp.ScanID == "" {
		t.Errorf("expected ScanID to be non-empty, got %q", resp.ScanID)
	}
}

func seedFindingForRemediation(t *testing.T, s *Server, scanPath string) int64 {
	t.Helper()

	res, err := s.db.Exec(`INSERT INTO products (name, repo_url, business_criticality) VALUES (?, ?, ?)`, "Demo", scanPath, "high")
	if err != nil {
		t.Fatalf("failed to insert product: %v", err)
	}
	productID, _ := res.LastInsertId()

	res, err = s.db.Exec(`INSERT INTO engagements (product_id, name, scan_path, status) VALUES (?, ?, ?, ?)`, productID, "Demo scan", scanPath, "completed")
	if err != nil {
		t.Fatalf("failed to insert engagement: %v", err)
	}
	engagementID, _ := res.LastInsertId()

	desc := "Unsanitized user input reaches a sink."
	fix := "Validate input at the boundary."
	res, err = s.db.Exec(`
		INSERT INTO findings (engagement_id, product_id, rule_id, title, severity, file_path, line_number, description, fix_suggestion, status, kanban_column, stack)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, engagementID, productID, "TEST-RULE", "Unsafe input handling", "HIGH", "src/app.go", 2, desc, fix, "open", "backlog", "go")
	if err != nil {
		t.Fatalf("failed to insert finding: %v", err)
	}
	findingID, _ := res.LastInsertId()
	return findingID
}

func TestFindingAgentPromptMarksSentToAgent(t *testing.T) {
	s := setupTestServer(t)
	tempDir := t.TempDir()
	srcDir := filepath.Join(tempDir, "src")
	if err := os.MkdirAll(srcDir, 0755); err != nil {
		t.Fatal(err)
	}
	jwt := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9." +
		"eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkRlbW8ifQ." +
		"fake_signature_here"
	if err := os.WriteFile(filepath.Join(srcDir, "app.go"), []byte("package main\n// token: "+jwt+"\nfunc handler(input string) {}\n"), 0644); err != nil {
		t.Fatal(err)
	}
	findingID := seedFindingForRemediation(t, s, tempDir)

	req, _ := http.NewRequest("POST", "/api/findings/"+strconv.FormatInt(findingID, 10)+"/agent-prompt", nil)
	addAuthCookie(req)
	rr := httptest.NewRecorder()
	s.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var resp map[string]any
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp["ok"] != true {
		t.Fatalf("expected ok response, got %#v", resp)
	}
	prompt, _ := resp["prompt"].(string)
	if !strings.Contains(prompt, "Unsafe input handling") || !strings.Contains(prompt, "src/app.go:2") {
		t.Fatalf("prompt missing finding context: %s", prompt)
	}
	if !strings.Contains(prompt, "AITriage generated this AGENT PROMPT") ||
		!strings.Contains(prompt, "## Remediation Guidance") ||
		!strings.Contains(prompt, "## Source Excerpt (redacted)") {
		t.Fatalf("prompt missing expected sections: %s", prompt)
	}
	if strings.Contains(prompt, jwt) || !strings.Contains(prompt, "[REDACTED]") {
		t.Fatalf("prompt did not redact source secret: %s", prompt)
	}

	finding, err := s.findingRepo.GetByID(req.Context(), findingID)
	if err != nil {
		t.Fatal(err)
	}
	if finding.Status != "sent_to_agent" || finding.KanbanColumn != "in_progress" {
		t.Fatalf("unexpected finding lifecycle: status=%q kanban=%q", finding.Status, finding.KanbanColumn)
	}
	if finding.AgentPrompt == nil || !strings.Contains(*finding.AgentPrompt, "Required Workflow") {
		t.Fatalf("expected stored agent prompt, got %#v", finding.AgentPrompt)
	}
}

func TestLocalPathForHostPrefixUsesDockerMountInfo(t *testing.T) {
	mountInfo := "315 306 0:43 /example/workspace /host rw,nosuid,nodev,relatime - fakeowner /run/host_mark/home rw,fakeowner"
	got, ok := localPathForHostPrefix("/host", "/host/demo-app/app.py", mountInfo)
	if !ok {
		t.Fatal("expected host path to be derived from mount info")
	}
	want := "/home/example/workspace/demo-app/app.py"
	if got != want {
		t.Fatalf("unexpected local path: got %q want %q", got, want)
	}
}

func TestRepositoryRelativePathPrefersPathInsideScanRoot(t *testing.T) {
	got := repositoryRelativePath("/host/demo-app", "/host/demo-app/thirdparty/VAmPI/app.py", "/host/demo-app/thirdparty/VAmPI/app.py")
	want := "thirdparty/VAmPI/app.py"
	if got != want {
		t.Fatalf("unexpected repository-relative path: got %q want %q", got, want)
	}
}

func TestFindingVerificationTransitions(t *testing.T) {
	s := setupTestServer(t)
	findingID := seedFindingForRemediation(t, s, t.TempDir())

	if err := s.findingRepo.MarkPendingVerification(context.Background(), findingID); err != nil {
		t.Fatalf("MarkPendingVerification failed: %v", err)
	}
	finding, err := s.findingRepo.GetByID(context.Background(), findingID)
	if err != nil {
		t.Fatal(err)
	}
	if finding.Status != "pending_verification" || finding.VerificationStatus == nil || *finding.VerificationStatus != "running" {
		t.Fatalf("unexpected pending state: status=%q verification=%v", finding.Status, finding.VerificationStatus)
	}

	if err := s.findingRepo.MarkVerificationResult(context.Background(), findingID, false, "still detected"); err != nil {
		t.Fatalf("MarkVerificationResult(false) failed: %v", err)
	}
	finding, err = s.findingRepo.GetByID(context.Background(), findingID)
	if err != nil {
		t.Fatal(err)
	}
	if finding.Status != "verification_failed" || finding.IsVerified {
		t.Fatalf("unexpected failed verification state: status=%q verified=%v", finding.Status, finding.IsVerified)
	}

	if err := s.findingRepo.MarkVerificationResult(context.Background(), findingID, true, "not detected"); err != nil {
		t.Fatalf("MarkVerificationResult(true) failed: %v", err)
	}
	finding, err = s.findingRepo.GetByID(context.Background(), findingID)
	if err != nil {
		t.Fatal(err)
	}
	if finding.Status != "resolved" || !finding.IsVerified || finding.ResolvedAt == nil || finding.VerifiedAt == nil {
		t.Fatalf("unexpected resolved state: status=%q verified=%v resolved_at=%v verified_at=%v", finding.Status, finding.IsVerified, finding.ResolvedAt, finding.VerifiedAt)
	}
}

func TestServeUI(t *testing.T) {
	s := setupTestServer(t)

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	s.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	if !strings.Contains(rr.Header().Get("Content-Type"), "text/html") && !strings.Contains(rr.Header().Get("Content-Type"), "text/plain") {
		t.Errorf("expected Content-Type text/html or text/plain, got %v", rr.Header().Get("Content-Type"))
	}
}
