// Package v1 provides the v1 API handlers for OweHost
package v1

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/iSundram/OweHost/internal/api/middleware"
	"github.com/iSundram/OweHost/internal/auth"
	"github.com/iSundram/OweHost/internal/user"
	"github.com/iSundram/OweHost/pkg/models"
	"github.com/iSundram/OweHost/pkg/utils"
)

// AuthHandler handles authentication endpoints
type AuthHandler struct {
	authService *auth.Service
	userService *user.Service
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(authSvc *auth.Service, userSvc *user.Service) *AuthHandler {
	return &AuthHandler{
		authService: authSvc,
		userService: userSvc,
	}
}

// Login handles user login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Invalid request body")
		return
	}

	user, err := h.userService.ValidateCredentials(req.Username, req.Password)
	if err != nil {
		utils.WriteError(w, http.StatusUnauthorized, utils.ErrCodeUnauthorized, "Invalid credentials")
		return
	}

	// Register user with auth service
	h.authService.RegisterUser(user)

	tokens, err := h.authService.GenerateTokens(user)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, utils.ErrCodeInternalError, "Failed to generate tokens")
		return
	}

	utils.WriteSuccess(w, tokens)
}

// Refresh handles token refresh
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	var req models.RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Invalid request body")
		return
	}

	tokens, err := h.authService.RefreshTokens(req.RefreshToken)
	if err != nil {
		utils.WriteError(w, http.StatusUnauthorized, utils.ErrCodeUnauthorized, err.Error())
		return
	}

	utils.WriteSuccess(w, tokens)
}

// Logout handles user logout
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	var req models.RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Invalid request body")
		return
	}

	if err := h.authService.InvalidateSession(req.RefreshToken); err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, err.Error())
		return
	}

	utils.WriteSuccess(w, map[string]string{"message": "Logged out successfully"})
}

// UserHandler handles user endpoints
type UserHandler struct {
	userService *user.Service
}

// NewUserHandler creates a new user handler
func NewUserHandler(userSvc *user.Service) *UserHandler {
	return &UserHandler{
		userService: userSvc,
	}
}

// Create handles user creation
func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	var req models.UserCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Invalid request body")
		return
	}

	// Validate
	if !utils.IsValidUsername(req.Username) {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeValidation, "Invalid username")
		return
	}
	if !utils.IsValidEmail(req.Email) {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeValidation, "Invalid email")
		return
	}

	user, err := h.userService.Create(&req)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, err.Error())
		return
	}

	utils.WriteCreated(w, user)
}

// Get handles getting a user
func (h *UserHandler) Get(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	// Extract user ID from path
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "User ID required")
		return
	}
	userID := parts[len(parts)-1]

	user, err := h.userService.Get(userID)
	if err != nil {
		utils.WriteError(w, http.StatusNotFound, utils.ErrCodeNotFound, err.Error())
		return
	}

	utils.WriteSuccess(w, user)
}

// List handles listing users
func (h *UserHandler) List(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	tenantID := r.URL.Query().Get("tenant_id")
	var tenantPtr *string
	if tenantID != "" {
		tenantPtr = &tenantID
	}

	users := h.userService.List(tenantPtr)
	utils.WriteSuccess(w, users)
}

// Update handles updating a user
func (h *UserHandler) Update(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch && r.Method != http.MethodPut {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "User ID required")
		return
	}
	userID := parts[len(parts)-1]

	var req models.UserUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Invalid request body")
		return
	}

	user, err := h.userService.Update(userID, &req)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, err.Error())
		return
	}

	utils.WriteSuccess(w, user)
}

// Delete handles deleting a user
func (h *UserHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "User ID required")
		return
	}
	userID := parts[len(parts)-1]

	if err := h.userService.Delete(userID); err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Me handles getting the current user
func (h *UserHandler) Me(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		utils.WriteError(w, http.StatusUnauthorized, utils.ErrCodeUnauthorized, "Not authenticated")
		return
	}

	user, err := h.userService.Get(userID)
	if err != nil {
		utils.WriteError(w, http.StatusNotFound, utils.ErrCodeNotFound, err.Error())
		return
	}

	utils.WriteSuccess(w, user)
}

// Suspend handles suspending a user
func (h *UserHandler) Suspend(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "User ID required")
		return
	}
	userID := parts[len(parts)-2]

	if err := h.userService.Suspend(userID); err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, err.Error())
		return
	}

	user, _ := h.userService.Get(userID)
	utils.WriteSuccess(w, user)
}

// Terminate handles terminating a user
func (h *UserHandler) Terminate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "User ID required")
		return
	}
	userID := parts[len(parts)-2]

	if err := h.userService.Terminate(userID); err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, err.Error())
		return
	}

	user, _ := h.userService.Get(userID)
	utils.WriteSuccess(w, user)
}
