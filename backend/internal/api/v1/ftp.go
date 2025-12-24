// Package v1 provides FTP API handlers for OweHost
package v1

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/iSundram/OweHost/internal/api/middleware"
	"github.com/iSundram/OweHost/internal/ftp"
	"github.com/iSundram/OweHost/internal/user"
	"github.com/iSundram/OweHost/pkg/models"
	"github.com/iSundram/OweHost/pkg/utils"
)

// FTPHandler handles FTP-related API requests
type FTPHandler struct {
	ftpService  *ftp.Service
	userService *user.Service
}

// NewFTPHandler creates a new FTP handler
func NewFTPHandler(ftpService *ftp.Service, userService *user.Service) *FTPHandler {
	return &FTPHandler{
		ftpService:  ftpService,
		userService: userService,
	}
}

// ListAccounts lists FTP accounts
func (h *FTPHandler) ListAccounts(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.ContextKeyUserID).(string)
	userRole := r.Context().Value(middleware.ContextKeyUserRole)

	var accounts []*models.FTPAccount

	if userRole != nil && userRole.(string) == "admin" {
		accounts = h.ftpService.ListAll()
	} else {
		accounts = h.ftpService.ListByUser(userID)
	}

	utils.WriteJSON(w, http.StatusOK, accounts)
}

// CreateAccount creates a new FTP account
func (h *FTPHandler) CreateAccount(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.ContextKeyUserID).(string)

	var req models.FTPAccountCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Invalid request body")
		return
	}

	if req.Username == "" || req.Password == "" {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Username and password are required")
		return
	}

	account, err := h.ftpService.Create(userID, &req)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, err.Error())
		return
	}

	utils.WriteJSON(w, http.StatusCreated, account)
}

// GetAccount gets an FTP account by ID
func (h *FTPHandler) GetAccount(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.ContextKeyUserID).(string)
	userRole := r.Context().Value(middleware.ContextKeyUserRole)

	id := extractIDFromPath(r.URL.Path, "ftp/accounts")

	account, err := h.ftpService.Get(id)
	if err != nil {
		utils.WriteError(w, http.StatusNotFound, utils.ErrCodeNotFound, err.Error())
		return
	}

	// Check ownership
	if account.UserID != userID && (userRole == nil || userRole.(string) != "admin") {
		utils.WriteError(w, http.StatusForbidden, utils.ErrCodeForbidden, "Access denied")
		return
	}

	utils.WriteJSON(w, http.StatusOK, account)
}

// UpdateAccount updates an FTP account
func (h *FTPHandler) UpdateAccount(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.ContextKeyUserID).(string)
	userRole := r.Context().Value(middleware.ContextKeyUserRole)

	id := extractIDFromPath(r.URL.Path, "ftp/accounts")

	account, err := h.ftpService.Get(id)
	if err != nil {
		utils.WriteError(w, http.StatusNotFound, utils.ErrCodeNotFound, err.Error())
		return
	}

	// Check ownership
	if account.UserID != userID && (userRole == nil || userRole.(string) != "admin") {
		utils.WriteError(w, http.StatusForbidden, utils.ErrCodeForbidden, "Access denied")
		return
	}

	var req models.FTPAccountUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Invalid request body")
		return
	}

	updated, err := h.ftpService.Update(id, &req)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, err.Error())
		return
	}

	utils.WriteJSON(w, http.StatusOK, updated)
}

// DeleteAccount deletes an FTP account
func (h *FTPHandler) DeleteAccount(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.ContextKeyUserID).(string)
	userRole := r.Context().Value(middleware.ContextKeyUserRole)

	id := extractIDFromPath(r.URL.Path, "ftp/accounts")

	account, err := h.ftpService.Get(id)
	if err != nil {
		utils.WriteError(w, http.StatusNotFound, utils.ErrCodeNotFound, err.Error())
		return
	}

	// Check ownership
	if account.UserID != userID && (userRole == nil || userRole.(string) != "admin") {
		utils.WriteError(w, http.StatusForbidden, utils.ErrCodeForbidden, "Access denied")
		return
	}

	if err := h.ftpService.Delete(id); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, utils.ErrCodeInternalError, err.Error())
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]string{"message": "FTP account deleted"})
}

// SuspendAccount suspends an FTP account
func (h *FTPHandler) SuspendAccount(w http.ResponseWriter, r *http.Request) {
	id := extractIDFromPath(r.URL.Path, "ftp/accounts")
	id = strings.TrimSuffix(id, "/suspend")

	if err := h.ftpService.Suspend(id); err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, err.Error())
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]string{"message": "FTP account suspended"})
}

// UnsuspendAccount unsuspends an FTP account
func (h *FTPHandler) UnsuspendAccount(w http.ResponseWriter, r *http.Request) {
	id := extractIDFromPath(r.URL.Path, "ftp/accounts")
	id = strings.TrimSuffix(id, "/unsuspend")

	if err := h.ftpService.Unsuspend(id); err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, err.Error())
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]string{"message": "FTP account unsuspended"})
}

// GetConfig returns FTP server configuration
func (h *FTPHandler) GetConfig(w http.ResponseWriter, r *http.Request) {
	config := h.ftpService.GetConfig()
	utils.WriteJSON(w, http.StatusOK, config)
}

// UpdateConfig updates FTP server configuration (admin only)
func (h *FTPHandler) UpdateConfig(w http.ResponseWriter, r *http.Request) {
	var config models.FTPConfig
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Invalid request body")
		return
	}

	if err := h.ftpService.UpdateConfig(&config); err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, err.Error())
		return
	}

	utils.WriteJSON(w, http.StatusOK, config)
}

// GetSessions returns active FTP sessions (admin only)
func (h *FTPHandler) GetSessions(w http.ResponseWriter, r *http.Request) {
	sessions := h.ftpService.GetActiveSessions()
	utils.WriteJSON(w, http.StatusOK, sessions)
}
