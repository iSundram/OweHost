package v1

import (
	"encoding/json"
	"net/http"

	"github.com/iSundram/OweHost/internal/provisioning"
	"github.com/iSundram/OweHost/internal/user"
)

type ProvisioningHandler struct {
	provisioningService *provisioning.Service
	userService         *user.Service
}

func NewProvisioningHandler(provisioningService *provisioning.Service, userService *user.Service) *ProvisioningHandler {
	return &ProvisioningHandler{
		provisioningService: provisioningService,
		userService:         userService,
	}
}

// ProvisionAccount provisions a complete account with all features
func (h *ProvisioningHandler) ProvisionAccount(w http.ResponseWriter, r *http.Request) {
	role := r.Context().Value("role").(string)
	if role != "admin" && role != "reseller" {
		http.Error(w, "Admin or reseller access required", http.StatusForbidden)
		return
	}

	var req provisioning.AccountProvisionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	result, err := h.provisioningService.ProvisionAccount(&req)
	if err != nil {
		// Return partial result even on error
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(result)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(result)
}

// GetProvisioningStatus retrieves the status of a provisioning operation
func (h *ProvisioningHandler) GetProvisioningStatus(w http.ResponseWriter, r *http.Request) {
	provisioningID := r.URL.Query().Get("id")
	if provisioningID == "" {
		http.Error(w, "Provisioning ID required", http.StatusBadRequest)
		return
	}

	status, err := h.provisioningService.GetProvisioningStatus(provisioningID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// DeprovisionAccount completely removes an account and all resources
func (h *ProvisioningHandler) DeprovisionAccount(w http.ResponseWriter, r *http.Request) {
	role := r.Context().Value("role").(string)
	if role != "admin" && role != "reseller" {
		http.Error(w, "Admin or reseller access required", http.StatusForbidden)
		return
	}

	var req struct {
		UserID string `json:"user_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.provisioningService.DeprovisionAccount(req.UserID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "deprovisioned"})
}
