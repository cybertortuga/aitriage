package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/cybertortuga/aitriage/internal/server/repositories"
	"github.com/cybertortuga/aitriage/internal/server/utils"
)

type MetricsHandler struct {
	repo *repositories.MetricsRepository
}

func NewMetricsHandler(repo *repositories.MetricsRepository) *MetricsHandler {
	return &MetricsHandler{repo: repo}
}

func (h *MetricsHandler) HandleGetDashboardMetrics(w http.ResponseWriter, r *http.Request) {
	metrics, err := h.repo.GetDashboardMetrics(r.Context())
	if err != nil {
		utils.JSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"ok": true, "metrics": metrics})
}
