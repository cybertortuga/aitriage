package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/cybertortuga/aitriage/internal/server/repositories"
)

type APIKeyHandler struct {
	repo *repositories.APIKeyRepository
}

func NewAPIKeyHandler(repo *repositories.APIKeyRepository) *APIKeyHandler {
	return &APIKeyHandler{repo: repo}
}

func (h *APIKeyHandler) HandleListKeys(w http.ResponseWriter, r *http.Request) {
	keys, err := h.repo.List()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"ok": true, "keys": keys})
}

func (h *APIKeyHandler) HandleCreateKey(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	// Generate a random token
	b := make([]byte, 24)
	rand.Read(b)
	token := "ait_" + hex.EncodeToString(b)
	prefix := token[:8] + "..."

	// For simplicity, we store the token as is for now, but in real life we'd hash it.
	// The user needs the raw token once.
	userID := 1 // Default admin for now, should get from context
	err := h.repo.Create(req.Name, prefix, token, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"ok":    true, 
		"token": token,
		"name":  req.Name,
	})
}

func (h *APIKeyHandler) HandleRevokeKey(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/admin/keys/")
	id, _ := strconv.Atoi(idStr)
	if id == 0 {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	err := h.repo.Revoke(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"ok": true})
}
