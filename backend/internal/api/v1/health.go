package v1

import (
	"net/http"

	"github.com/iSundram/OweHost/internal/logging"
	"github.com/iSundram/OweHost/pkg/utils"
)

// HealthHandler handles health check endpoints
type HealthHandler struct {
	loggingService *logging.Service
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(logSvc *logging.Service) *HealthHandler {
	return &HealthHandler{
		loggingService: logSvc,
	}
}

// Health handles health check
func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	utils.WriteSuccess(w, map[string]interface{}{
		"status":  "healthy",
		"version": "1.0.0",
	})
}

// Ready handles readiness check
func (h *HealthHandler) Ready(w http.ResponseWriter, r *http.Request) {
	utils.WriteSuccess(w, map[string]interface{}{
		"ready": true,
	})
}

// Metrics handles metrics endpoint
func (h *HealthHandler) Metrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	metrics := h.loggingService.ExportPrometheus()
	w.Write([]byte(metrics))
}
