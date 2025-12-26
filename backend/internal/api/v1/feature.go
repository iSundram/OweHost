package v1

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/iSundram/OweHost/internal/feature"
	"github.com/iSundram/OweHost/pkg/utils"
)

// FeatureHandler handles feature flag management
type FeatureHandler struct {
	featureService *feature.Service
}

// NewFeatureHandler creates a new feature handler
func NewFeatureHandler(featureSvc *feature.Service) *FeatureHandler {
	return &FeatureHandler{
		featureService: featureSvc,
	}
}

// List returns all features grouped by category
func (h *FeatureHandler) List(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	categories, err := h.featureService.List()
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, utils.ErrCodeInternalError, err.Error())
		return
	}

	utils.WriteSuccess(w, categories)
}

// Get returns a specific feature
func (h *FeatureHandler) Get(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	parts := strings.Split(strings.TrimSuffix(r.URL.Path, "/"), "/")
	if len(parts) < 5 {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Feature name required")
		return
	}

	featureName := parts[len(parts)-1]

	feat, err := h.featureService.Get(featureName)
	if err != nil {
		utils.WriteError(w, http.StatusNotFound, utils.ErrCodeNotFound, err.Error())
		return
	}

	utils.WriteSuccess(w, feat)
}

// Update enables/disables a feature
func (h *FeatureHandler) Update(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut && r.Method != http.MethodPatch {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	parts := strings.Split(strings.TrimSuffix(r.URL.Path, "/"), "/")
	if len(parts) < 5 {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Feature name required")
		return
	}

	featureName := parts[len(parts)-1]

	var req struct {
		Enabled bool `json:"enabled"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Invalid request body")
		return
	}

	feat, err := h.featureService.Update(featureName, req.Enabled)
	if err != nil {
		utils.WriteError(w, http.StatusNotFound, utils.ErrCodeNotFound, err.Error())
		return
	}

	utils.WriteSuccess(w, feat)
}
