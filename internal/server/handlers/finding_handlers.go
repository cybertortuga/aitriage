package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/cybertortuga/aitriage/internal/models"
	"github.com/cybertortuga/aitriage/internal/server/repositories"
	"github.com/cybertortuga/aitriage/internal/server/utils"
)

type FindingHandler struct {
	repo *repositories.FindingRepository
}

func NewFindingHandler(repo *repositories.FindingRepository) *FindingHandler {
	return &FindingHandler{repo: repo}
}

func (h *FindingHandler) HandleListFindings(w http.ResponseWriter, r *http.Request) {
	engagementIDStr := r.URL.Query().Get("engagement_id")

	w.Header().Set("Content-Type", "application/json")
	if engagementIDStr == "" {
		findings, err := h.repo.ListAll(r.Context())
		if err != nil {
			utils.JSONError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if findings == nil {
			findings = []models.Finding{}
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"findings": findings, "ok": true})
		return
	}

	engagementID, err := strconv.ParseInt(engagementIDStr, 10, 64)
	if err != nil {
		utils.JSONError(w, "invalid engagement_id", http.StatusBadRequest)
		return
	}

	findings, err := h.repo.List(r.Context(), engagementID)
	if err != nil {
		utils.JSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if findings == nil {
		findings = []models.Finding{}
	}
	_ = json.NewEncoder(w).Encode(map[string]any{"findings": findings, "ok": true})
}

func (h *FindingHandler) HandleUpdateFinding(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/findings/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		utils.JSONError(w, "missing finding id", http.StatusBadRequest)
		return
	}
	findingID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		utils.JSONError(w, "invalid finding id", http.StatusBadRequest)
		return
	}

	var req struct {
		Action       string `json:"action"`        // "status" or "kanban"
		Status       string `json:"status"`        // e.g. "open", "in_progress", "fixed"
		KanbanColumn string `json:"kanban_column"` // e.g. "backlog", "todo", "in_progress", "done"
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.JSONError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Action == "status" {
		if err := h.repo.UpdateStatus(r.Context(), findingID, req.Status); err != nil {
			utils.JSONError(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else if req.Action == "kanban" {
		if err := h.repo.UpdateKanbanColumn(r.Context(), findingID, req.KanbanColumn); err != nil {
			utils.JSONError(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		utils.JSONError(w, "invalid action", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
}
