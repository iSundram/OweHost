package v1

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/iSundram/OweHost/internal/user"
	"github.com/iSundram/OweHost/internal/webserver"
	"github.com/iSundram/OweHost/pkg/models"
)

type WebServerHandler struct {
	webserverService *webserver.Service
	userService      *user.Service
}

func NewWebServerHandler(webserverService *webserver.Service, userService *user.Service) *WebServerHandler {
	return &WebServerHandler{
		webserverService: webserverService,
		userService:      userService,
	}
}

// GetConfig retrieves web server configuration
func (h *WebServerHandler) GetConfig(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)

	config, err := h.webserverService.GetUserConfig(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(config)
}

// UpdateConfig updates web server configuration
func (h *WebServerHandler) UpdateConfig(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)

	var req models.WebServerConfigUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	config, err := h.webserverService.UpdateUserConfig(userID, &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(config)
}

// ListPHPVersions lists available PHP versions
func (h *WebServerHandler) ListPHPVersions(w http.ResponseWriter, r *http.Request) {
	versions, err := h.webserverService.ListPHPVersions()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(versions)
}

// SwitchPHPVersion switches PHP version for user
func (h *WebServerHandler) SwitchPHPVersion(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)

	var req struct {
		Version string `json:"version"`
		Domain  string `json:"domain,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.webserverService.SwitchPHPVersion(userID, req.Domain, req.Version); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"version": req.Version,
	})
}

// ListModules lists web server modules
func (h *WebServerHandler) ListModules(w http.ResponseWriter, r *http.Request) {
	modules, err := h.webserverService.ListModules()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(modules)
}

// EnableModule enables a web server module
func (h *WebServerHandler) EnableModule(w http.ResponseWriter, r *http.Request) {
	role := r.Context().Value("role").(string)
	if role != "admin" {
		http.Error(w, "Admin access required", http.StatusForbidden)
		return
	}

	var req struct {
		Module string `json:"module"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.webserverService.EnableModule(req.Module); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "enabled"})
}

// DisableModule disables a web server module
func (h *WebServerHandler) DisableModule(w http.ResponseWriter, r *http.Request) {
	role := r.Context().Value("role").(string)
	if role != "admin" {
		http.Error(w, "Admin access required", http.StatusForbidden)
		return
	}

	var req struct {
		Module string `json:"module"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.webserverService.DisableModule(req.Module); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "disabled"})
}

// RestartServer restarts the web server
func (h *WebServerHandler) RestartServer(w http.ResponseWriter, r *http.Request) {
	role := r.Context().Value("role").(string)
	if role != "admin" {
		http.Error(w, "Admin access required", http.StatusForbidden)
		return
	}

	if err := h.webserverService.Restart(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "restarted"})
}

// GetErrorLogs retrieves error logs
func (h *WebServerHandler) GetErrorLogs(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)
	role := r.Context().Value("role").(string)

	domain := r.URL.Query().Get("domain")
	lines := 100
	if l := r.URL.Query().Get("lines"); l != "" {
		if _, err := fmt.Sscanf(l, "%d", &lines); err != nil {
			lines = 100
		}
	}

	var logs []string
	var err error

	if role == "admin" && domain == "" {
		logs, err = h.webserverService.GetSystemErrorLogs(lines)
	} else {
		logs, err = h.webserverService.GetUserErrorLogs(userID, domain, lines)
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"logs": logs,
	})
}

// GetAccessLogs retrieves access logs
func (h *WebServerHandler) GetAccessLogs(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)
	domain := r.URL.Query().Get("domain")

	lines := 100
	if l := r.URL.Query().Get("lines"); l != "" {
		if _, err := fmt.Sscanf(l, "%d", &lines); err != nil {
			lines = 100
		}
	}

	logs, err := h.webserverService.GetUserAccessLogs(userID, domain, lines)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"logs": logs,
	})
}

// GetServerStatus retrieves web server status
func (h *WebServerHandler) GetServerStatus(w http.ResponseWriter, r *http.Request) {
	status, err := h.webserverService.GetStatus()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// TestConfiguration tests the web server configuration
func (h *WebServerHandler) TestConfiguration(w http.ResponseWriter, r *http.Request) {
	role := r.Context().Value("role").(string)
	if role != "admin" {
		http.Error(w, "Admin access required", http.StatusForbidden)
		return
	}

	result, err := h.webserverService.TestConfig()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
