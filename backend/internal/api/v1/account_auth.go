package v1

import (
	"encoding/json"
	"net/http"

	"github.com/iSundram/OweHost/internal/accountsvc"
	"github.com/iSundram/OweHost/internal/api/middleware"
	"github.com/iSundram/OweHost/internal/auth"
	"github.com/iSundram/OweHost/internal/user"
	"github.com/iSundram/OweHost/pkg/models"
	"github.com/iSundram/OweHost/pkg/utils"
)

// AccountAuthHandler handles account authentication
type AccountAuthHandler struct {
	accountService *accountsvc.Service
	authService    *auth.Service
	userService    *user.Service
}

// NewAccountAuthHandler creates a new account auth handler
func NewAccountAuthHandler(accountService *accountsvc.Service, authService *auth.Service, userService *user.Service) *AccountAuthHandler {
	return &AccountAuthHandler{
		accountService: accountService,
		authService:    authService,
		userService:    userService,
	}
}

// LoginRequest represents account login request
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Login handles account login
func (h *AccountAuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Invalid request body")
		return
	}

	if req.Username == "" || req.Password == "" {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Username and password required")
		return
	}

	// Authenticate account
	identity, err := h.accountService.Authenticate(r.Context(), req.Username, req.Password)
	if err != nil {
		utils.WriteError(w, http.StatusUnauthorized, utils.ErrCodeUnauthorized, "Invalid credentials")
		return
	}

	// Create user model for JWT
	user := &models.User{
		ID:       identity.Name,
		Username: identity.Name,
		Email:    "", // Would need to get from metadata
		Role:     "account",
		TenantID: identity.Owner,
	}

	// Generate tokens
	tokens, err := h.authService.GenerateTokens(user)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, utils.ErrCodeInternalError, "Failed to generate tokens")
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"message":    "Login successful",
		"account_id": identity.ID,
		"username":   identity.Name,
		"plan":       identity.Plan,
		"tokens":     tokens,
	})
}

// ChangePassword handles password change
func (h *AccountAuthHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	// Get authenticated user ID from context
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		utils.WriteError(w, http.StatusUnauthorized, utils.ErrCodeUnauthorized, "Not authenticated")
		return
	}

	// Get user from database
	dbUser, err := h.userService.Get(userID)
	if err != nil {
		utils.WriteError(w, http.StatusUnauthorized, utils.ErrCodeUnauthorized, "User not found")
		return
	}

	var req struct {
		CurrentPassword string `json:"current_password"`
		NewPassword     string `json:"new_password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Invalid request body")
		return
	}

	// Verify current password
	_, err = h.accountService.Authenticate(r.Context(), dbUser.Username, req.CurrentPassword)
	if err != nil {
		utils.WriteError(w, http.StatusUnauthorized, utils.ErrCodeUnauthorized, "Current password incorrect")
		return
	}

	// Get account to find ID
	account, err := h.accountService.GetByUsername(r.Context(), dbUser.Username)
	if err != nil {
		utils.WriteError(w, http.StatusNotFound, utils.ErrCodeNotFound, "Account not found")
		return
	}

	// Change password
	if err := h.accountService.ChangePassword(r.Context(), account.Identity.ID, req.NewPassword, dbUser.Username, "account"); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, utils.ErrCodeInternalError, "Failed to change password")
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]string{
		"message": "Password changed successfully",
	})
}
