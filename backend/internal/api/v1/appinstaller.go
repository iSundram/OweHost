package v1

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/iSundram/OweHost/internal/appinstaller"
	"github.com/iSundram/OweHost/internal/user"
	"github.com/iSundram/OweHost/pkg/models"
)

type AppInstallerHandler struct {
	appService  *appinstaller.Service
	userService *user.Service
}

func NewAppInstallerHandler(appService *appinstaller.Service, userService *user.Service) *AppInstallerHandler {
	return &AppInstallerHandler{
		appService:  appService,
		userService: userService,
	}
}

// ListAvailableApps lists all available applications
func (h *AppInstallerHandler) ListAvailableApps(w http.ResponseWriter, r *http.Request) {
	category := r.URL.Query().Get("category")

	apps := h.appService.ListApps(category)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(apps)
}

// GetApp retrieves details about a specific application
func (h *AppInstallerHandler) GetApp(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.TrimSuffix(r.URL.Path, "/"), "/")
	if len(parts) < 5 {
		http.Error(w, "Invalid app ID", http.StatusBadRequest)
		return
	}
	appID := parts[4]

	app, err := h.appService.GetApp(appID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(app)
}

// InstallApp installs an application
func (h *AppInstallerHandler) InstallApp(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)

	var req models.AppInstallRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	installation, err := h.appService.Install(userID, &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(installation)
}

// ListInstalledApps lists user's installed applications
func (h *AppInstallerHandler) ListInstalledApps(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)

	installations := h.appService.ListInstalledByUser(userID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(installations)
}

// GetInstalledApp retrieves details about an installed application
func (h *AppInstallerHandler) GetInstalledApp(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)
	parts := strings.Split(strings.TrimSuffix(r.URL.Path, "/"), "/")
	if len(parts) < 6 {
		http.Error(w, "Invalid installation ID", http.StatusBadRequest)
		return
	}
	installationID := parts[5]

	installation, err := h.appService.GetInstalled(installationID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// Authorization check
	if installation.UserID != userID {
		role := r.Context().Value("role").(string)
		if role != "admin" {
			http.Error(w, "Unauthorized", http.StatusForbidden)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(installation)
}

// UninstallApp uninstalls an application
func (h *AppInstallerHandler) UninstallApp(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)
	parts := strings.Split(strings.TrimSuffix(r.URL.Path, "/"), "/")
	if len(parts) < 6 {
		http.Error(w, "Invalid installation ID", http.StatusBadRequest)
		return
	}
	installationID := parts[5]

	installation, err := h.appService.GetInstalled(installationID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// Authorization check
	if installation.UserID != userID {
		role := r.Context().Value("role").(string)
		if role != "admin" {
			http.Error(w, "Unauthorized", http.StatusForbidden)
			return
		}
	}

	if err := h.appService.Uninstall(installationID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// UpdateInstalledApp updates an installed application
func (h *AppInstallerHandler) UpdateInstalledApp(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)
	parts := strings.Split(strings.TrimSuffix(r.URL.Path, "/"), "/")
	if len(parts) < 6 {
		http.Error(w, "Invalid installation ID", http.StatusBadRequest)
		return
	}
	installationID := parts[5]

	installation, err := h.appService.GetInstalled(installationID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// Authorization check
	if installation.UserID != userID {
		role := r.Context().Value("role").(string)
		if role != "admin" {
			http.Error(w, "Unauthorized", http.StatusForbidden)
			return
		}
	}

	var req models.AppUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		req = models.AppUpdateRequest{} // use default
	}

	updated, err := h.appService.Update(installationID, &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updated)
}

// GetInstallationStatus retrieves the status of an app installation
func (h *AppInstallerHandler) GetInstallationStatus(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)
	parts := strings.Split(strings.TrimSuffix(r.URL.Path, "/"), "/")
	if len(parts) < 7 {
		http.Error(w, "Invalid installation ID", http.StatusBadRequest)
		return
	}
	installationID := parts[5]

	installation, err := h.appService.GetInstalled(installationID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// Authorization check
	if installation.UserID != userID {
		role := r.Context().Value("role").(string)
		if role != "admin" {
			http.Error(w, "Unauthorized", http.StatusForbidden)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":        installation.Status,
		"version":       installation.Version,
		"error_message": installation.ErrorMessage,
	})
}
