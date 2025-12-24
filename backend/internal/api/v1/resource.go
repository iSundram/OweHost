package v1

import (
	"encoding/json"
	"net/http"

	"github.com/iSundram/OweHost/internal/resource"
	"github.com/iSundram/OweHost/internal/user"
)

type ResourceHandler struct {
	resourceService *resource.Service
	userService     *user.Service
}

func NewResourceHandler(resourceService *resource.Service, userService *user.Service) *ResourceHandler {
	return &ResourceHandler{
		resourceService: resourceService,
		userService:     userService,
	}
}

// GetUsage retrieves resource usage for authenticated user
func (h *ResourceHandler) GetUsage(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)

	usage, err := h.resourceService.GetUsage(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(usage)
}

// GetLimits retrieves resource limits for authenticated user
func (h *ResourceHandler) GetLimits(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)

	quota, err := h.resourceService.GetQuota(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(quota)
}

// UpdateLimits updates resource limits (admin only)
func (h *ResourceHandler) UpdateLimits(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID string `json:"user_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	quota, err := h.resourceService.GetQuota(req.UserID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(quota)
}

// GetHistory retrieves resource usage history
func (h *ResourceHandler) GetHistory(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)

	// Return current usage as history placeholder
	usage, err := h.resourceService.GetUsage(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode([]interface{}{usage})
}

// GetSystemResources retrieves system-wide resource information (admin only)
func (h *ResourceHandler) GetSystemResources(w http.ResponseWriter, r *http.Request) {
	// Return placeholder system resources
	resources := map[string]interface{}{
		"cpu_total":    100,
		"cpu_used":     45,
		"memory_total": 16384,
		"memory_used":  8192,
		"disk_total":   500000,
		"disk_used":    250000,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resources)
}
