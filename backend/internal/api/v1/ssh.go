// Package v1 provides SSH API handlers for OweHost
package v1

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/iSundram/OweHost/internal/api/middleware"
	"github.com/iSundram/OweHost/internal/ssh"
	"github.com/iSundram/OweHost/internal/user"
	"github.com/iSundram/OweHost/pkg/models"
	"github.com/iSundram/OweHost/pkg/utils"
)

// SSHHandler handles SSH-related API requests
type SSHHandler struct {
	sshService  *ssh.Service
	userService *user.Service
}

// NewSSHHandler creates a new SSH handler
func NewSSHHandler(sshService *ssh.Service, userService *user.Service) *SSHHandler {
	return &SSHHandler{
		sshService:  sshService,
		userService: userService,
	}
}

// ListKeys lists SSH keys for a user
func (h *SSHHandler) ListKeys(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.ContextKeyUserID).(string)
	
	keys := h.sshService.ListKeys(userID)
	utils.WriteJSON(w, http.StatusOK, keys)
}

// AddKey adds a new SSH key
func (h *SSHHandler) AddKey(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.ContextKeyUserID).(string)

	var req models.SSHKeyCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Invalid request body")
		return
	}

	if req.Name == "" || req.PublicKey == "" {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Name and public key are required")
		return
	}

	key, err := h.sshService.AddKey(userID, &req)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, err.Error())
		return
	}

	utils.WriteJSON(w, http.StatusCreated, key)
}

// GenerateKey generates a new SSH key pair
func (h *SSHHandler) GenerateKey(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.ContextKeyUserID).(string)

	var req models.SSHKeyGenerateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Invalid request body")
		return
	}

	if req.Name == "" {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Name is required")
		return
	}

	if req.KeyType == "" {
		req.KeyType = "ed25519"
	}

	key, keyPair, err := h.sshService.GenerateKey(userID, &req)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, err.Error())
		return
	}

	utils.WriteJSON(w, http.StatusCreated, map[string]interface{}{
		"key":      key,
		"key_pair": keyPair,
	})
}

// GetKey gets an SSH key by ID
func (h *SSHHandler) GetKey(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.ContextKeyUserID).(string)
	userRole := r.Context().Value(middleware.ContextKeyUserRole)

	id := extractIDFromPath(r.URL.Path, "ssh/keys")

	key, err := h.sshService.GetKey(id)
	if err != nil {
		utils.WriteError(w, http.StatusNotFound, utils.ErrCodeNotFound, err.Error())
		return
	}

	// Check ownership
	if key.UserID != userID && (userRole == nil || userRole.(string) != "admin") {
		utils.WriteError(w, http.StatusForbidden, utils.ErrCodeForbidden, "Access denied")
		return
	}

	utils.WriteJSON(w, http.StatusOK, key)
}

// DeleteKey deletes an SSH key
func (h *SSHHandler) DeleteKey(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.ContextKeyUserID).(string)
	userRole := r.Context().Value(middleware.ContextKeyUserRole)

	id := extractIDFromPath(r.URL.Path, "ssh/keys")

	key, err := h.sshService.GetKey(id)
	if err != nil {
		utils.WriteError(w, http.StatusNotFound, utils.ErrCodeNotFound, err.Error())
		return
	}

	// Check ownership
	if key.UserID != userID && (userRole == nil || userRole.(string) != "admin") {
		utils.WriteError(w, http.StatusForbidden, utils.ErrCodeForbidden, "Access denied")
		return
	}

	if err := h.sshService.DeleteKey(id); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, utils.ErrCodeInternalError, err.Error())
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]string{"message": "SSH key deleted"})
}

// GetAccess gets SSH access configuration
func (h *SSHHandler) GetAccess(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.ContextKeyUserID).(string)

	access, err := h.sshService.GetAccess(userID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, utils.ErrCodeInternalError, err.Error())
		return
	}

	utils.WriteJSON(w, http.StatusOK, access)
}

// UpdateAccess updates SSH access configuration
func (h *SSHHandler) UpdateAccess(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.ContextKeyUserID).(string)

	var req models.SSHAccessUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Invalid request body")
		return
	}

	access, err := h.sshService.UpdateAccess(userID, &req)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, err.Error())
		return
	}

	utils.WriteJSON(w, http.StatusOK, access)
}

// EnableAccess enables SSH access
func (h *SSHHandler) EnableAccess(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.ContextKeyUserID).(string)

	if err := h.sshService.EnableAccess(userID); err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, err.Error())
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]string{"message": "SSH access enabled"})
}

// DisableAccess disables SSH access
func (h *SSHHandler) DisableAccess(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.ContextKeyUserID).(string)

	if err := h.sshService.DisableAccess(userID); err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, err.Error())
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]string{"message": "SSH access disabled"})
}

// GetSessions returns active SSH sessions (admin only)
func (h *SSHHandler) GetSessions(w http.ResponseWriter, r *http.Request) {
	sessions := h.sshService.GetActiveSessions()
	utils.WriteJSON(w, http.StatusOK, sessions)
}

// extractSSHKeyIDFromPath extracts the key ID from path
func extractSSHKeyIDFromPath(path string) string {
	parts := strings.Split(strings.TrimSuffix(path, "/"), "/")
	for i, part := range parts {
		if part == "keys" && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return ""
}
