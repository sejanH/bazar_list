package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/sejan/bazarlist/internal/auth"
	"github.com/sejan/bazarlist/internal/models"
	"github.com/sejan/bazarlist/internal/storage"
)

// AuthHandler handles authentication requests
type AuthHandler struct {
	storage *storage.MySQLStorage
	authMgr  *auth.AuthManager
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(store *storage.MySQLStorage) *AuthHandler {
	return &AuthHandler{
		storage: store,
		authMgr:  auth.NewAuthManager(),
	}
}

// RegisterRoutes registers authentication routes
func (h *AuthHandler) RegisterRoutes(router interface{}) {
	// This will be registered in main.go
}

// Register handles user registration
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Check if user already exists by email
	_, err := h.storage.GetUserByEmail(req.Email)
	if err == nil {
		respondError(w, http.StatusConflict, "Email already exists")
		return
	}

	// Hash password
	hashedPassword, err := h.authMgr.HashPassword(req.Password)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to hash password")
		return
	}

	// Create user
	user := &models.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: hashedPassword,
	}

	if err := h.storage.CreateUser(user); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to create user")
		return
	}

	// Generate token
	token, err := h.authMgr.GenerateToken(user.ID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	respondJSON(w, http.StatusCreated, models.TokenResponse{
		Token: token,
		User:  *user,
	})
}

// Login handles user login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Find user by email
	user, err := h.storage.GetUserByEmail(req.Email)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	// Check password
	if !h.authMgr.CheckPassword(req.Password, user.Password) {
		respondError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	// Generate token
	token, err := h.authMgr.GenerateToken(user.ID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	respondJSON(w, http.StatusOK, models.TokenResponse{
		Token: token,
		User:  *user,
	})
}

// AuthMiddleware validates JWT tokens
func (h *AuthHandler) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		tokenString := auth.GetTokenFromHeader(authHeader)

		if tokenString == "" {
			tokenString = r.URL.Query().Get("token")
		}

		if tokenString == "" {
			respondError(w, http.StatusUnauthorized, "Authorization header required")
			return
		}

		userID, err := h.authMgr.ValidateToken(tokenString)
		if err != nil {
			respondError(w, http.StatusUnauthorized, "Invalid or expired token")
			return
		}

		// Add user ID to request context
		r.Header.Set("X-User-ID", strconv.FormatUint(uint64(userID), 10))

		next.ServeHTTP(w, r)
	})
}

// GetUserID extracts user ID from request headers
func GetUserID(r *http.Request) uint {
	userIDStr := r.Header.Get("X-User-ID")
	if userIDStr == "" {
		return 0
	}
	var userID uint
	if _, err := fmt.Sscanf(userIDStr, "%d", &userID); err != nil {
		return 0
	}
	return userID
}
