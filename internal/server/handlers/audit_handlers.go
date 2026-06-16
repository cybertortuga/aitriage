package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/cybertortuga/aitriage/internal/server/repositories"
	"github.com/cybertortuga/aitriage/internal/server/utils"
)

type AuditHandler struct {
	repo *repositories.AuditRepository
}

func NewAuditHandler(repo *repositories.AuditRepository) *AuditHandler {
	return &AuditHandler{repo: repo}
}

func (h *AuditHandler) HandleListAuditLogs(w http.ResponseWriter, r *http.Request) {
	entityType := r.URL.Query().Get("entity_type")

	logs, err := h.repo.List(r.Context(), entityType, 100, 0)
	if err != nil {
		utils.JSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"ok": true, "audit_logs": logs})
}
