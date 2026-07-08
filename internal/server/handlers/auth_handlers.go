package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/cybertortuga/aitriage/internal/models"
	"github.com/cybertortuga/aitriage/internal/server/middleware"
	"github.com/cybertortuga/aitriage/internal/server/repositories"
	"github.com/cybertortuga/aitriage/internal/server/utils"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"sync"
)

type loginAttempt struct {
	count       int
	lastAttempt time.Time
}

type AuthHandler struct {
	userRepo      *repositories.UserRepository
	loginAttempts map[string]*loginAttempt
	mu            sync.Mutex
}

func NewAuthHandler(userRepo *repositories.UserRepository) *AuthHandler {
	return &AuthHandler{
		userRepo:      userRepo,
		loginAttempts: make(map[string]*loginAttempt),
	}
}

func (h *AuthHandler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	ip := r.RemoteAddr
	if forward := r.Header.Get("X-Forwarded-For"); forward != "" {
		ip = forward
	}

	h.mu.Lock()
	attempt, exists := h.loginAttempts[ip]
	if exists {
		if time.Since(attempt.lastAttempt) > 10*time.Minute {
			attempt.count = 0
		}
		if attempt.count >= 5 {
			h.mu.Unlock()
			utils.JSONError(w, "Too many login attempts. Please try again in 10 minutes.", http.StatusTooManyRequests)
			return
		}
	} else {
		attempt = &loginAttempt{}
		h.loginAttempts[ip] = attempt
	}
	h.mu.Unlock()

	var creds struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		utils.JSONError(w, "invalid credentials", http.StatusBadRequest)
		return
	}

	user, err := h.userRepo.GetByUsername(r.Context(), creds.Username)
	if err != nil {
		utils.JSONError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if !user.IsActive {
		utils.JSONError(w, "Account disabled", http.StatusUnauthorized)
		return
	}

	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(creds.Password)) != nil {
		h.mu.Lock()
		attempt.count++
		attempt.lastAttempt = time.Now()
		h.mu.Unlock()
		utils.JSONError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	h.mu.Lock()
	delete(h.loginAttempts, ip) // Reset on success
	h.mu.Unlock()

	// Update last login
	_ = h.userRepo.UpdateLastLogin(r.Context(), user.ID)

	expirationTime := time.Now().Add(24 * time.Hour)
	isAdmin := user.GlobalRole == "superadmin" || user.GlobalRole == "admin"

	claims := &middleware.Claims{
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.GlobalRole,
		IsAdmin:  isAdmin,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	// Sign token using middleware key
	tokenString, err := middleware.SignToken(claims)
	if err != nil {
		utils.JSONError(w, "Internal error", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    tokenString,
		Expires:  expirationTime,
		Path:     "/",
		HttpOnly: true,
	})

	_ = json.NewEncoder(w).Encode(map[string]any{
		"ok":       true,
		"user_id":  user.ID,
		"username": user.Username,
		"role":     user.GlobalRole,
		"is_admin": isAdmin,
		"token":    tokenString,
	})
}

func (h *AuthHandler) HandleMe(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie("token")
	if err != nil {
		// Auth disabled — return default admin user
		_ = json.NewEncoder(w).Encode(map[string]any{
			"ok":          true,
			"id":          1,
			"username":    "admin",
			"global_role": "superadmin",
			"is_admin":    true,
		})
		return
	}

	claims := &middleware.Claims{}
	_, err = jwt.ParseWithClaims(c.Value, claims, func(token *jwt.Token) (interface{}, error) {
		return middleware.JwtKey, nil
	})

	if err != nil {
		// Invalid token — still return default admin (auth disabled)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"ok":          true,
			"id":          1,
			"username":    "admin",
			"global_role": "superadmin",
			"is_admin":    true,
		})
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]any{
		"ok":          true,
		"id":          claims.UserID,
		"username":    claims.Username,
		"global_role": claims.Role,
		"is_admin":    claims.IsAdmin,
	})
}

func (h *AuthHandler) HandleListUsers(w http.ResponseWriter, r *http.Request) {
	limit := 100
	offset := 0

	users, err := h.userRepo.List(r.Context(), limit, offset)
	if err != nil {
		utils.JSONError(w, "Internal error", http.StatusInternalServerError)
		return
	}

	var res []map[string]any
	for _, u := range users {
		res = append(res, map[string]any{
			"id":          u.ID,
			"username":    u.Username,
			"email":       u.Email,
			"full_name":   u.FullName,
			"global_role": u.GlobalRole,
			"is_active":   u.IsActive,
			"is_admin":    u.GlobalRole == "superadmin" || u.GlobalRole == "admin",
		})
	}

	_ = json.NewEncoder(w).Encode(map[string]any{"users": res})
}

func (h *AuthHandler) HandleCreateUser(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username   string `json:"username"`
		Password   string `json:"password"`
		Email      string `json:"email"`
		FullName   string `json:"full_name"`
		GlobalRole string `json:"global_role"`
		IsActive   bool   `json:"is_active"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.JSONError(w, "Invalid data", http.StatusBadRequest)
		return
	}

	if req.Username == "" {
		utils.JSONError(w, "Username is required", http.StatusBadRequest)
		return
	}

	if req.GlobalRole == "" {
		req.GlobalRole = "viewer"
	}

	hash, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)

	emailPtr := &req.Email
	if req.Email == "" {
		emailPtr = nil
	}

	fullNamePtr := &req.FullName
	if req.FullName == "" {
		fullNamePtr = nil
	}

	user := &models.User{
		Username:     req.Username,
		Email:        emailPtr,
		FullName:     fullNamePtr,
		PasswordHash: string(hash),
		GlobalRole:   req.GlobalRole,
		IsActive:     req.IsActive,
	}

	id, err := h.userRepo.Create(r.Context(), user)
	if err != nil {
		utils.JSONError(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]any{"ok": true, "id": id})
}

func (h *AuthHandler) HandleDeleteUser(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		utils.JSONError(w, "id is required", http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.JSONError(w, "invalid id", http.StatusBadRequest)
		return
	}

	user, err := h.userRepo.GetByID(r.Context(), id)
	if err != nil {
		utils.JSONError(w, "User not found", http.StatusNotFound)
		return
	}

	if user.GlobalRole == "superadmin" {
		utils.JSONError(w, "Cannot delete superadmin", http.StatusBadRequest)
		return
	}

	err = h.userRepo.Delete(r.Context(), id)
	if err != nil {
		utils.JSONError(w, "Failed to delete user", http.StatusInternalServerError)
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
}
