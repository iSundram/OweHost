package v1

import (
	"encoding/json"
	"net/http"

	"github.com/iSundram/OweHost/internal/installation"
	"github.com/iSundram/OweHost/pkg/models"
	"github.com/iSundram/OweHost/pkg/utils"
)

// InstallationHandler handles installation endpoints
type InstallationHandler struct {
	installationService *installation.Service
}

// NewInstallationHandler creates a new installation handler
func NewInstallationHandler(installSvc *installation.Service) *InstallationHandler {
	return &InstallationHandler{
		installationService: installSvc,
	}
}

// Check handles checking installation status
func (h *InstallationHandler) Check(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	response := models.InstallationCheckResponse{
		IsInstalled:      h.installationService.IsInstalled(),
		RequiresSetup:    !h.installationService.IsInstalled(),
		SupportedEngines: h.installationService.GetSupportedEngines(),
	}

	utils.WriteSuccess(w, response)
}

// Install handles the installation process
func (h *InstallationHandler) Install(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	// Check if already installed
	if h.installationService.IsInstalled() {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "System is already installed")
		return
	}

	var req models.InstallationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Invalid request body")
		return
	}

	// Validate required fields
	if req.DatabaseEngine == "" {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeValidation, "Database engine is required")
		return
	}

	if req.AdminEmail == "" {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeValidation, "Admin email is required")
		return
	}

	installation, err := h.installationService.Install(&req)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, err.Error())
		return
	}

	utils.WriteCreated(w, installation)
}
