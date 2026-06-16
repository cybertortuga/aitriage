package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/cybertortuga/aitriage/internal/server/middleware"
	"github.com/cybertortuga/aitriage/internal/server/repositories"
	"github.com/cybertortuga/aitriage/internal/server/utils"
)

type NotificationHandler struct {
	repo *repositories.NotificationRepository
}

func NewNotificationHandler(repo *repositories.NotificationRepository) *NotificationHandler {
	return &NotificationHandler{repo: repo}
}

func (h *NotificationHandler) HandleListNotifications(w http.ResponseWriter, r *http.Request) {
	claims, err := middleware.ExtractClaims(r)
	if err != nil {
		utils.JSONError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	userID := claims.UserID
	unreadOnly := r.URL.Query().Get("unread") == "true"

	notifs, err := h.repo.ListByUser(r.Context(), userID, unreadOnly)
	if err != nil {
		utils.JSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"ok": true, "notifications": notifs})
}

func (h *NotificationHandler) HandleMarkAsRead(w http.ResponseWriter, r *http.Request) {
	userID := int64(1)

	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/notifications/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		utils.JSONError(w, "missing notification id", http.StatusBadRequest)
		return
	}

	if parts[0] == "read-all" {
		err := h.repo.MarkAllAsRead(r.Context(), userID)
		if err != nil {
			utils.JSONError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		utils.JSONResponse(w, map[string]any{"ok": true})
		return
	}

	id, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		utils.JSONError(w, "invalid notification id", http.StatusBadRequest)
		return
	}

	err = h.repo.MarkAsRead(r.Context(), id, userID)
	if err != nil {
		utils.JSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.JSONResponse(w, map[string]any{"ok": true})
}
