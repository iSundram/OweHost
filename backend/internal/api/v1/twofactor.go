// Package v1 provides 2FA API handlers for OweHost
package v1

import (
	"encoding/json"
	"net/http"

	"github.com/iSundram/OweHost/internal/api/middleware"
	"github.com/iSundram/OweHost/internal/twofactor"
	"github.com/iSundram/OweHost/internal/user"
	"github.com/iSundram/OweHost/pkg/models"
	"github.com/iSundram/OweHost/pkg/utils"
)

// TwoFactorHandler handles 2FA-related API requests
type TwoFactorHandler struct {
	tfService   *twofactor.Service
	userService *user.Service
}

// NewTwoFactorHandler creates a new 2FA handler
func NewTwoFactorHandler(tfService *twofactor.Service, userService *user.Service) *TwoFactorHandler {
	return &TwoFactorHandler{
		tfService:   tfService,
		userService: userService,
	}
}

// GetStatus gets 2FA status for current user
func (h *TwoFactorHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.ContextKeyUserID).(string)

	enabled := h.tfService.IsEnabled(userID)
	
	response := map[string]interface{}{
		"enabled": enabled,
	}

	if enabled {
		config, _ := h.tfService.GetConfig(userID)
		if config != nil {
			response["type"] = config.Type
			response["verified_at"] = config.VerifiedAt
			response["last_used_at"] = config.LastUsedAt
			
			remaining, _ := h.tfService.GetBackupCodesRemaining(userID)
			response["backup_codes_remaining"] = remaining
		}
	}

	utils.WriteJSON(w, http.StatusOK, response)
}

// SetupTOTP initiates TOTP setup
func (h *TwoFactorHandler) SetupTOTP(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.ContextKeyUserID).(string)

	// Get username
	u, err := h.userService.Get(userID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, utils.ErrCodeInternalError, err.Error())
		return
	}

	setup, err := h.tfService.SetupTOTP(userID, u.Username)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, err.Error())
		return
	}

	utils.WriteJSON(w, http.StatusOK, setup)
}

// VerifySetup verifies TOTP setup and enables 2FA
func (h *TwoFactorHandler) VerifySetup(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.ContextKeyUserID).(string)

	var req models.TOTPVerifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Invalid request body")
		return
	}

	if req.Code == "" {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Code is required")
		return
	}

	if err := h.tfService.VerifySetup(userID, req.Code); err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, err.Error())
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"enabled": true,
		"message": "Two-factor authentication enabled successfully",
	})
}

// Disable disables 2FA
func (h *TwoFactorHandler) Disable(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.ContextKeyUserID).(string)

	var req models.TwoFactorDisableRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Invalid request body")
		return
	}

	if req.Code == "" {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Current 2FA code is required")
		return
	}

	if err := h.tfService.Disable(userID, req.Code); err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, err.Error())
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"enabled": false,
		"message": "Two-factor authentication disabled",
	})
}

// RegenerateBackupCodes regenerates backup codes
func (h *TwoFactorHandler) RegenerateBackupCodes(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.ContextKeyUserID).(string)

	var req models.BackupCodesRegenerateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Invalid request body")
		return
	}

	if req.Code == "" {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Current 2FA code is required")
		return
	}

	codes, err := h.tfService.RegenerateBackupCodes(userID, req.Code)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, err.Error())
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"backup_codes": codes,
		"message":      "New backup codes generated. Store them securely.",
	})
}

// Verify verifies a 2FA code (used during login)
func (h *TwoFactorHandler) Verify(w http.ResponseWriter, r *http.Request) {
	var req models.TwoFactorLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Invalid request body")
		return
	}

	// In production, token would contain encrypted user info
	// For now, we expect userID in context from initial login
	userID := r.Context().Value(middleware.ContextKeyUserID)
	if userID == nil {
		utils.WriteError(w, http.StatusUnauthorized, utils.ErrCodeUnauthorized, "Invalid or expired token")
		return
	}

	valid, err := h.tfService.VerifyCode(userID.(string), req.Code)
	if err != nil || !valid {
		utils.WriteError(w, http.StatusUnauthorized, utils.ErrCodeUnauthorized, "Invalid 2FA code")
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"verified": true,
		"message":  "Two-factor authentication verified",
	})
}

// GetLoginAttempts returns recent login attempts
func (h *TwoFactorHandler) GetLoginAttempts(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.ContextKeyUserID).(string)

	attempts := h.tfService.GetLoginAttempts(userID, 20)
	utils.WriteJSON(w, http.StatusOK, attempts)
}
