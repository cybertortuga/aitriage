package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/cybertortuga/aitriage/internal/server/repositories"
	"github.com/cybertortuga/aitriage/internal/server/utils"
)

type ConfigHandler struct {
	configRepo *repositories.ConfigRepository
}

func NewConfigHandler(configRepo *repositories.ConfigRepository) *ConfigHandler {
	return &ConfigHandler{configRepo: configRepo}
}

func (h *ConfigHandler) HandleGetConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.JSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	config, err := h.configRepo.GetAll(r.Context())
	if err != nil {
		utils.JSONError(w, "Failed to fetch configuration", http.StatusInternalServerError)
		return
	}

	utils.JSONResponse(w, config)
}

func (h *ConfigHandler) HandleUpdateConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodPut {
		utils.JSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var newConfig map[string]string
	if err := json.NewDecoder(r.Body).Decode(&newConfig); err != nil {
		utils.JSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.configRepo.SetMany(r.Context(), newConfig); err != nil {
		utils.JSONError(w, "Failed to update configuration", http.StatusInternalServerError)
		return
	}

	utils.JSONResponse(w, map[string]string{"status": "success"})
}
