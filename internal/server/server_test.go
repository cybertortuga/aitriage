package server

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

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
