package v1

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/iSundram/OweHost/internal/logging"
	"github.com/iSundram/OweHost/pkg/models"
	"github.com/iSundram/OweHost/pkg/utils"
)

// LoggingHandler handles logging and metrics endpoints
type LoggingHandler struct {
	loggingService *logging.Service
}

// NewLoggingHandler creates a new logging handler
func NewLoggingHandler(loggingSvc *logging.Service) *LoggingHandler {
	return &LoggingHandler{
		loggingService: loggingSvc,
	}
}

// QueryLogs handles querying log entries
func (h *LoggingHandler) QueryLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	var req models.LogQueryRequest

	if r.Method == http.MethodPost {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Invalid request body")
			return
		}
	} else {
		// Parse query params for GET request
		query := r.URL.Query()

		// Default time range: last 24 hours
		req.StartTime = time.Now().Add(-24 * time.Hour)
		req.EndTime = time.Now()

		if startStr := query.Get("start_time"); startStr != "" {
			if t, err := time.Parse(time.RFC3339, startStr); err == nil {
				req.StartTime = t
			}
		}
		if endStr := query.Get("end_time"); endStr != "" {
			if t, err := time.Parse(time.RFC3339, endStr); err == nil {
				req.EndTime = t
			}
		}

		if levelStr := query.Get("level"); levelStr != "" {
			level := models.LogLevel(levelStr)
			req.Level = &level
		}

		if service := query.Get("service"); service != "" {
			req.Service = &service
		}

		if userID := query.Get("user_id"); userID != "" {
			req.UserID = &userID
		}

		// Pagination
		req.Limit = 100
		req.Offset = 0
	}

	// Set defaults if not provided
	if req.Limit == 0 {
		req.Limit = 100
	}
	if req.StartTime.IsZero() {
		req.StartTime = time.Now().Add(-24 * time.Hour)
	}
	if req.EndTime.IsZero() {
		req.EndTime = time.Now()
	}

	logs := h.loggingService.QueryLogs(&req)
	utils.WriteSuccess(w, map[string]interface{}{
		"logs":   logs,
		"count":  len(logs),
		"offset": req.Offset,
		"limit":  req.Limit,
	})
}

// CreateLog handles creating a log entry
func (h *LoggingHandler) CreateLog(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	var req struct {
		Level    string                 `json:"level"`
		Service  string                 `json:"service"`
		Message  string                 `json:"message"`
		Metadata map[string]interface{} `json:"metadata"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Invalid request body")
		return
	}

	var level models.LogLevel
	switch req.Level {
	case "debug":
		level = models.LogLevelDebug
	case "info":
		level = models.LogLevelInfo
	case "warn", "warning":
		level = models.LogLevelWarn
	case "error":
		level = models.LogLevelError
	default:
		level = models.LogLevelInfo
	}

	entry := h.loggingService.Log(level, req.Service, req.Message, nil, nil, nil, req.Metadata)
	utils.WriteCreated(w, entry)
}

// QueryAudits handles querying audit entries
func (h *LoggingHandler) QueryAudits(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	var req models.AuditQueryRequest

	if r.Method == http.MethodPost {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Invalid request body")
			return
		}
	} else {
		// Parse query params for GET request
		query := r.URL.Query()

		// Default time range: last 24 hours
		req.StartTime = time.Now().Add(-24 * time.Hour)
		req.EndTime = time.Now()

		if startStr := query.Get("start_time"); startStr != "" {
			if t, err := time.Parse(time.RFC3339, startStr); err == nil {
				req.StartTime = t
			}
		}
		if endStr := query.Get("end_time"); endStr != "" {
			if t, err := time.Parse(time.RFC3339, endStr); err == nil {
				req.EndTime = t
			}
		}

		if userID := query.Get("user_id"); userID != "" {
			req.UserID = &userID
		}

		if action := query.Get("action"); action != "" {
			req.Action = &action
		}

		if resourceType := query.Get("resource_type"); resourceType != "" {
			req.ResourceType = &resourceType
		}

		if resourceID := query.Get("resource_id"); resourceID != "" {
			req.ResourceID = &resourceID
		}

		// Pagination
		req.Limit = 100
		req.Offset = 0
	}

	// Set defaults if not provided
	if req.Limit == 0 {
		req.Limit = 100
	}
	if req.StartTime.IsZero() {
		req.StartTime = time.Now().Add(-24 * time.Hour)
	}
	if req.EndTime.IsZero() {
		req.EndTime = time.Now()
	}

	audits := h.loggingService.QueryAudits(&req)
	utils.WriteSuccess(w, map[string]interface{}{
		"audits": audits,
		"count":  len(audits),
		"offset": req.Offset,
		"limit":  req.Limit,
	})
}

// ListMetrics handles listing all metrics
func (h *LoggingHandler) ListMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	metrics := h.loggingService.ListMetrics()
	utils.WriteSuccess(w, metrics)
}

// RecordMetric handles recording a metric
func (h *LoggingHandler) RecordMetric(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	var req struct {
		Name   string            `json:"name"`
		Type   string            `json:"type"`
		Value  float64           `json:"value"`
		Labels map[string]string `json:"labels"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Invalid request body")
		return
	}

	metric := h.loggingService.RecordMetric(req.Name, req.Type, req.Value, req.Labels)
	utils.WriteCreated(w, metric)
}

// IncrementCounter handles incrementing a counter metric
func (h *LoggingHandler) IncrementCounter(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	var req struct {
		Name   string            `json:"name"`
		Labels map[string]string `json:"labels"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Invalid request body")
		return
	}

	h.loggingService.IncrementCounter(req.Name, req.Labels)
	utils.WriteSuccess(w, map[string]string{"message": "Counter incremented"})
}

// SetGauge handles setting a gauge metric
func (h *LoggingHandler) SetGauge(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	var req struct {
		Name   string            `json:"name"`
		Value  float64           `json:"value"`
		Labels map[string]string `json:"labels"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Invalid request body")
		return
	}

	h.loggingService.SetGauge(req.Name, req.Value, req.Labels)
	utils.WriteSuccess(w, map[string]string{"message": "Gauge set"})
}

// ExportPrometheus handles exporting metrics in Prometheus format
func (h *LoggingHandler) ExportPrometheus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	output := h.loggingService.ExportPrometheus()
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(output))
}
