package v1

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/iSundram/OweHost/internal/plugin"
	"github.com/iSundram/OweHost/pkg/models"
	"github.com/iSundram/OweHost/pkg/utils"
)

// PluginHandler handles plugin endpoints
type PluginHandler struct {
	pluginService *plugin.Service
}

// NewPluginHandler creates a new plugin handler
func NewPluginHandler(pluginSvc *plugin.Service) *PluginHandler {
	return &PluginHandler{
		pluginService: pluginSvc,
	}
}

// List handles listing all plugins
func (h *PluginHandler) List(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	plugins := h.pluginService.List()
	utils.WriteSuccess(w, plugins)
}

// Get handles getting a plugin by ID
func (h *PluginHandler) Get(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Plugin ID required")
		return
	}
	pluginID := parts[len(parts)-1]

	plugin, err := h.pluginService.Get(pluginID)
	if err != nil {
		utils.WriteError(w, http.StatusNotFound, utils.ErrCodeNotFound, err.Error())
		return
	}

	utils.WriteSuccess(w, plugin)
}

// Install handles plugin installation
func (h *PluginHandler) Install(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	var req models.PluginInstallRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Invalid request body")
		return
	}

	plugin, err := h.pluginService.Install(&req)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, err.Error())
		return
	}

	utils.WriteCreated(w, plugin)
}

// Uninstall handles plugin uninstallation
func (h *PluginHandler) Uninstall(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Plugin ID required")
		return
	}
	pluginID := parts[len(parts)-1]

	if err := h.pluginService.Uninstall(pluginID); err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Activate handles plugin activation
func (h *PluginHandler) Activate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Plugin ID required")
		return
	}
	pluginID := parts[len(parts)-2]

	if err := h.pluginService.Activate(pluginID); err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, err.Error())
		return
	}

	plugin, _ := h.pluginService.Get(pluginID)
	utils.WriteSuccess(w, plugin)
}

// Deactivate handles plugin deactivation
func (h *PluginHandler) Deactivate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Plugin ID required")
		return
	}
	pluginID := parts[len(parts)-2]

	if err := h.pluginService.Deactivate(pluginID); err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, err.Error())
		return
	}

	plugin, _ := h.pluginService.Get(pluginID)
	utils.WriteSuccess(w, plugin)
}

// Configure handles plugin configuration
func (h *PluginHandler) Configure(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut && r.Method != http.MethodPatch {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Plugin ID required")
		return
	}
	pluginID := parts[len(parts)-2]

	var req models.PluginConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Invalid request body")
		return
	}

	if err := h.pluginService.Configure(pluginID, &req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, err.Error())
		return
	}

	plugin, _ := h.pluginService.Get(pluginID)
	utils.WriteSuccess(w, plugin)
}

// ListScopes handles listing available API scopes
func (h *PluginHandler) ListScopes(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	scopes := h.pluginService.GetScopes()
	utils.WriteSuccess(w, scopes)
}

// GrantScope handles granting a scope to a plugin
func (h *PluginHandler) GrantScope(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Plugin ID required")
		return
	}
	pluginID := parts[len(parts)-2]

	var req struct {
		Scope string `json:"scope"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Invalid request body")
		return
	}

	if err := h.pluginService.GrantScope(pluginID, req.Scope); err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, err.Error())
		return
	}

	utils.WriteSuccess(w, map[string]string{"message": "Scope granted"})
}

// RevokeScope handles revoking a scope from a plugin
func (h *PluginHandler) RevokeScope(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Plugin ID and scope required")
		return
	}
	pluginID := parts[len(parts)-3]
	scope := parts[len(parts)-1]

	if err := h.pluginService.RevokeScope(pluginID, scope); err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ListHooks handles listing hooks for a plugin
func (h *PluginHandler) ListHooks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Plugin ID required")
		return
	}
	pluginID := parts[len(parts)-2]

	hooks := h.pluginService.GetHooks(pluginID)
	utils.WriteSuccess(w, hooks)
}

// RegisterHook handles registering a plugin hook
func (h *PluginHandler) RegisterHook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Plugin ID required")
		return
	}
	pluginID := parts[len(parts)-2]

	var req struct {
		Event    string `json:"event"`
		Handler  string `json:"handler"`
		Priority int    `json:"priority"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Invalid request body")
		return
	}

	hook, err := h.pluginService.RegisterHook(pluginID, req.Event, req.Handler, req.Priority)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, err.Error())
		return
	}

	utils.WriteCreated(w, hook)
}

// UnregisterHook handles unregistering a plugin hook
func (h *PluginHandler) UnregisterHook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Hook ID required")
		return
	}
	hookID := parts[len(parts)-1]

	if err := h.pluginService.UnregisterHook(hookID); err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
