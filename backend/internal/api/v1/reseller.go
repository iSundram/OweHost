package v1

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/iSundram/OweHost/internal/api/middleware"
	"github.com/iSundram/OweHost/internal/reseller"
	"github.com/iSundram/OweHost/internal/user"
	"github.com/iSundram/OweHost/pkg/models"
	"github.com/iSundram/OweHost/pkg/utils"
)

// ResellerHandler handles reseller endpoints
type ResellerHandler struct {
	resellerService *reseller.Service
	userService     *user.Service
}

// NewResellerHandler creates a new reseller handler
func NewResellerHandler(resellerSvc *reseller.Service, userSvc *user.Service) *ResellerHandler {
	return &ResellerHandler{
		resellerService: resellerSvc,
		userService:     userSvc,
	}
}

// Create handles reseller creation
func (h *ResellerHandler) Create(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	var req models.ResellerCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Invalid request body")
		return
	}

	// Verify user exists and set role to reseller
	_, err := h.userService.Get(req.UserID)
	if err != nil {
		utils.WriteError(w, http.StatusNotFound, utils.ErrCodeNotFound, "User not found")
		return
	}

	// Update user role to reseller
	role := models.UserRoleReseller
	_, err = h.userService.Update(req.UserID, &models.UserUpdateRequest{
		Role: &role,
	})
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, err.Error())
		return
	}

	reseller, err := h.resellerService.Create(&req)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, err.Error())
		return
	}

	utils.WriteCreated(w, reseller)
}

// Get handles getting a reseller
func (h *ResellerHandler) Get(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Reseller ID required")
		return
	}
	resellerID := parts[len(parts)-1]

	reseller, err := h.resellerService.Get(resellerID)
	if err != nil {
		utils.WriteError(w, http.StatusNotFound, utils.ErrCodeNotFound, err.Error())
		return
	}

	utils.WriteSuccess(w, reseller)
}

// List handles listing resellers
func (h *ResellerHandler) List(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	resellers := h.resellerService.List()
	utils.WriteSuccess(w, resellers)
}

// Update handles updating a reseller
func (h *ResellerHandler) Update(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch && r.Method != http.MethodPut {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Reseller ID required")
		return
	}
	resellerID := parts[len(parts)-1]

	var req models.ResellerUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Invalid request body")
		return
	}

	reseller, err := h.resellerService.Update(resellerID, &req)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, err.Error())
		return
	}

	utils.WriteSuccess(w, reseller)
}

// Delete handles deleting a reseller
func (h *ResellerHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Reseller ID required")
		return
	}
	resellerID := parts[len(parts)-1]

	reseller, err := h.resellerService.Get(resellerID)
	if err != nil {
		utils.WriteError(w, http.StatusNotFound, utils.ErrCodeNotFound, err.Error())
		return
	}

	// Update user role back to user
	role := models.UserRoleUser
	_, err = h.userService.Update(reseller.UserID, &models.UserUpdateRequest{
		Role: &role,
	})
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, err.Error())
		return
	}

	if err := h.resellerService.Delete(resellerID); err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetByUserID handles getting a reseller by user ID
func (h *ResellerHandler) GetByUserID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		utils.WriteError(w, http.StatusUnauthorized, utils.ErrCodeUnauthorized, "Not authenticated")
		return
	}

	reseller, err := h.resellerService.GetByUserID(userID)
	if err != nil {
		utils.WriteError(w, http.StatusNotFound, utils.ErrCodeNotFound, err.Error())
		return
	}

	utils.WriteSuccess(w, reseller)
}
