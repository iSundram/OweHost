package v1

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/iSundram/OweHost/internal/api/middleware"
	"github.com/iSundram/OweHost/internal/runtime"
	"github.com/iSundram/OweHost/pkg/models"
	"github.com/iSundram/OweHost/pkg/utils"
)

// RuntimeHandler handles runtime endpoints
type RuntimeHandler struct {
	runtimeService *runtime.Service
}

// NewRuntimeHandler creates a new runtime handler
func NewRuntimeHandler(runtimeSvc *runtime.Service) *RuntimeHandler {
	return &RuntimeHandler{
		runtimeService: runtimeSvc,
	}
}

// ListVersions handles listing available runtime versions
func (h *RuntimeHandler) ListVersions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	// Get runtime type from path: /api/v1/runtimes/{type}/versions
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Runtime type required")
		return
	}
	runtimeType := parts[len(parts)-2]

	var rtType models.RuntimeType
	switch runtimeType {
	case "php":
		rtType = models.RuntimeTypePHP
	case "nodejs":
		rtType = models.RuntimeTypeNodeJS
	case "python":
		rtType = models.RuntimeTypePython
	case "go":
		rtType = models.RuntimeTypeGo
	case "java":
		rtType = models.RuntimeTypeJava
	default:
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Invalid runtime type")
		return
	}

	versions := h.runtimeService.ListVersions(rtType)
	utils.WriteSuccess(w, versions)
}

// CreatePHPPool handles PHP pool creation
func (h *RuntimeHandler) CreatePHPPool(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	userID := middleware.GetUserID(r.Context())

	var req models.PHPPoolCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Invalid request body")
		return
	}

	pool, err := h.runtimeService.CreatePHPPool(userID, &req)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, err.Error())
		return
	}

	utils.WriteCreated(w, pool)
}

// GetPHPPool handles getting a PHP pool
func (h *RuntimeHandler) GetPHPPool(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Pool ID required")
		return
	}
	poolID := parts[len(parts)-1]

	pool, err := h.runtimeService.GetPHPPool(poolID)
	if err != nil {
		utils.WriteError(w, http.StatusNotFound, utils.ErrCodeNotFound, err.Error())
		return
	}

	utils.WriteSuccess(w, pool)
}

// ListPHPPools handles listing PHP pools for a user
func (h *RuntimeHandler) ListPHPPools(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	userID := middleware.GetUserID(r.Context())
	pools := h.runtimeService.ListPHPPoolsByUser(userID)
	utils.WriteSuccess(w, pools)
}

// UpdatePHPPool handles updating a PHP pool
func (h *RuntimeHandler) UpdatePHPPool(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut && r.Method != http.MethodPatch {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Pool ID required")
		return
	}
	poolID := parts[len(parts)-1]

	var req models.PHPPoolCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Invalid request body")
		return
	}

	pool, err := h.runtimeService.UpdatePHPPool(poolID, &req)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, err.Error())
		return
	}

	utils.WriteSuccess(w, pool)
}

// DeletePHPPool handles deleting a PHP pool
func (h *RuntimeHandler) DeletePHPPool(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Pool ID required")
		return
	}
	poolID := parts[len(parts)-1]

	if err := h.runtimeService.DeletePHPPool(poolID); err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// EnablePHPExtension handles enabling a PHP extension
func (h *RuntimeHandler) EnablePHPExtension(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Pool ID required")
		return
	}
	poolID := parts[len(parts)-2]

	var req struct {
		Extension string `json:"extension"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Invalid request body")
		return
	}

	if err := h.runtimeService.EnablePHPExtension(poolID, req.Extension); err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, err.Error())
		return
	}

	utils.WriteSuccess(w, map[string]string{"message": "Extension enabled"})
}

// DisablePHPExtension handles disabling a PHP extension
func (h *RuntimeHandler) DisablePHPExtension(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Pool ID and extension required")
		return
	}
	poolID := parts[len(parts)-3]
	extension := parts[len(parts)-1]

	if err := h.runtimeService.DisablePHPExtension(poolID, extension); err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// CreateNodeJSApp handles Node.js app creation
func (h *RuntimeHandler) CreateNodeJSApp(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	userID := middleware.GetUserID(r.Context())

	var req models.NodeJSAppCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Invalid request body")
		return
	}

	app, err := h.runtimeService.CreateNodeJSApp(userID, &req)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, err.Error())
		return
	}

	utils.WriteCreated(w, app)
}

// GetNodeJSApp handles getting a Node.js app
func (h *RuntimeHandler) GetNodeJSApp(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "App ID required")
		return
	}
	appID := parts[len(parts)-1]

	app, err := h.runtimeService.GetNodeJSApp(appID)
	if err != nil {
		utils.WriteError(w, http.StatusNotFound, utils.ErrCodeNotFound, err.Error())
		return
	}

	utils.WriteSuccess(w, app)
}

// StartNodeJSApp handles starting a Node.js app
func (h *RuntimeHandler) StartNodeJSApp(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "App ID required")
		return
	}
	appID := parts[len(parts)-2]

	if err := h.runtimeService.StartNodeJSApp(appID); err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, err.Error())
		return
	}

	app, _ := h.runtimeService.GetNodeJSApp(appID)
	utils.WriteSuccess(w, app)
}

// StopNodeJSApp handles stopping a Node.js app
func (h *RuntimeHandler) StopNodeJSApp(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "App ID required")
		return
	}
	appID := parts[len(parts)-2]

	if err := h.runtimeService.StopNodeJSApp(appID); err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, err.Error())
		return
	}

	app, _ := h.runtimeService.GetNodeJSApp(appID)
	utils.WriteSuccess(w, app)
}

// RestartNodeJSApp handles restarting a Node.js app
func (h *RuntimeHandler) RestartNodeJSApp(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "App ID required")
		return
	}
	appID := parts[len(parts)-2]

	// Restart = stop + start
	_ = h.runtimeService.StopNodeJSApp(appID)
	if err := h.runtimeService.StartNodeJSApp(appID); err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, err.Error())
		return
	}

	app, _ := h.runtimeService.GetNodeJSApp(appID)
	utils.WriteSuccess(w, app)
}

// DeleteNodeJSApp handles deleting a Node.js app
func (h *RuntimeHandler) DeleteNodeJSApp(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "App ID required")
		return
	}
	appID := parts[len(parts)-1]

	if err := h.runtimeService.DeleteNodeJSApp(appID); err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// CreatePythonApp handles Python app creation
func (h *RuntimeHandler) CreatePythonApp(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	userID := middleware.GetUserID(r.Context())

	var req models.PythonAppCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Invalid request body")
		return
	}

	app, err := h.runtimeService.CreatePythonApp(userID, &req)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, err.Error())
		return
	}

	utils.WriteCreated(w, app)
}

// GetPythonApp handles getting a Python app
func (h *RuntimeHandler) GetPythonApp(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "App ID required")
		return
	}
	appID := parts[len(parts)-1]

	app, err := h.runtimeService.GetPythonApp(appID)
	if err != nil {
		utils.WriteError(w, http.StatusNotFound, utils.ErrCodeNotFound, err.Error())
		return
	}

	utils.WriteSuccess(w, app)
}

// ProvisionVirtualenv handles provisioning a Python virtualenv
func (h *RuntimeHandler) ProvisionVirtualenv(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "App ID required")
		return
	}
	appID := parts[len(parts)-2]

	if err := h.runtimeService.ProvisionVirtualenv(appID); err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, err.Error())
		return
	}

	utils.WriteSuccess(w, map[string]string{"message": "Virtualenv provisioned"})
}

// DeletePythonApp handles deleting a Python app
func (h *RuntimeHandler) DeletePythonApp(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "App ID required")
		return
	}
	appID := parts[len(parts)-1]

	if err := h.runtimeService.DeletePythonApp(appID); err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
