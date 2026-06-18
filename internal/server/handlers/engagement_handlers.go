package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/cybertortuga/aitriage/internal/models"
	"github.com/cybertortuga/aitriage/internal/server/repositories"
	"github.com/cybertortuga/aitriage/internal/server/utils"
)

type EngagementHandler struct {
	repo *repositories.EngagementRepository
}

func NewEngagementHandler(repo *repositories.EngagementRepository) *EngagementHandler {
	return &EngagementHandler{repo: repo}
}

func (h *EngagementHandler) HandleListEngagements(w http.ResponseWriter, r *http.Request) {
	productIDStr := r.URL.Query().Get("product_id")
	if productIDStr == "" {
		engagements, err := h.repo.ListAll(r.Context())
		if err != nil {
			utils.JSONError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(engagements)
		return
	}

	productID, err := strconv.ParseInt(productIDStr, 10, 64)
	if err != nil {
		utils.JSONError(w, "invalid product_id", http.StatusBadRequest)
		return
	}

	engagements, err := h.repo.List(r.Context(), productID)
	if err != nil {
		utils.JSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(engagements)
}

func (h *EngagementHandler) HandleCreateEngagement(w http.ResponseWriter, r *http.Request) {
	var e models.Engagement
	if err := json.NewDecoder(r.Body).Decode(&e); err != nil {
		utils.JSONError(w, "invalid engagement data", http.StatusBadRequest)
		return
	}
	if err := h.repo.Create(r.Context(), &e); err != nil {
		utils.JSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(e)
}
