package v1

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/iSundram/OweHost/internal/api/middleware"
	"github.com/iSundram/OweHost/internal/stats"
	"github.com/iSundram/OweHost/pkg/utils"
)

// StatsHandler handles statistics endpoints
type StatsHandler struct {
	statsService *stats.Service
}

// NewStatsHandler creates a new stats handler
func NewStatsHandler(statsSvc *stats.Service) *StatsHandler {
	return &StatsHandler{
		statsService: statsSvc,
	}
}

// GetBandwidthStats handles getting bandwidth statistics
func (h *StatsHandler) GetBandwidthStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Domain ID required")
		return
	}
	domainID := parts[len(parts)-2]

	// Parse date range from query params
	startDate, endDate := parseDateRange(r)

	stats := h.statsService.GetBandwidthStats(domainID, startDate, endDate)
	utils.WriteSuccess(w, stats)
}

// GetVisitorStats handles getting visitor statistics
func (h *StatsHandler) GetVisitorStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Domain ID required")
		return
	}
	domainID := parts[len(parts)-2]

	startDate, endDate := parseDateRange(r)

	stats := h.statsService.GetVisitorStats(domainID, startDate, endDate)
	utils.WriteSuccess(w, stats)
}

// GetErrorLogs handles getting error logs
func (h *StatsHandler) GetErrorLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Domain ID required")
		return
	}
	domainID := parts[len(parts)-2]

	limit := parseLimit(r, 100)

	logs := h.statsService.GetErrorLogs(domainID, limit)
	utils.WriteSuccess(w, map[string]interface{}{
		"logs":  logs,
		"count": len(logs),
	})
}

// GetAccessLogs handles getting access logs
func (h *StatsHandler) GetAccessLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Domain ID required")
		return
	}
	domainID := parts[len(parts)-2]

	limit := parseLimit(r, 100)

	logs := h.statsService.GetAccessLogs(domainID, limit)
	utils.WriteSuccess(w, map[string]interface{}{
		"logs":  logs,
		"count": len(logs),
	})
}

// GetDomainSummary handles getting domain statistics summary
func (h *StatsHandler) GetDomainSummary(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Domain ID required")
		return
	}
	domainID := parts[len(parts)-2]

	summary := h.statsService.GetDomainSummary(domainID)
	utils.WriteSuccess(w, summary)
}

// GetResourceStats handles getting resource statistics for the current user
func (h *StatsHandler) GetResourceStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		utils.WriteError(w, http.StatusUnauthorized, utils.ErrCodeUnauthorized, "Not authenticated")
		return
	}

	stats := h.statsService.GetResourceStats(userID)
	utils.WriteSuccess(w, stats)
}

// GetUserSummary handles getting user statistics summary
func (h *StatsHandler) GetUserSummary(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		utils.WriteError(w, http.StatusUnauthorized, utils.ErrCodeUnauthorized, "Not authenticated")
		return
	}

	summary := h.statsService.GetUserSummary(userID)
	utils.WriteSuccess(w, summary)
}

// GetUserResourceStats handles getting resource stats for a specific user (admin)
func (h *StatsHandler) GetUserResourceStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "User ID required")
		return
	}
	userID := parts[len(parts)-2]

	stats := h.statsService.GetResourceStats(userID)
	utils.WriteSuccess(w, stats)
}

// Helper functions

func parseDateRange(r *http.Request) (time.Time, time.Time) {
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -30) // Default: last 30 days

	if startStr := r.URL.Query().Get("start_date"); startStr != "" {
		if t, err := time.Parse("2006-01-02", startStr); err == nil {
			startDate = t
		}
	}

	if endStr := r.URL.Query().Get("end_date"); endStr != "" {
		if t, err := time.Parse("2006-01-02", endStr); err == nil {
			endDate = t
		}
	}

	// Handle period param (e.g., 7d, 30d, 90d)
	if period := r.URL.Query().Get("period"); period != "" {
		switch period {
		case "7d":
			startDate = endDate.AddDate(0, 0, -7)
		case "30d":
			startDate = endDate.AddDate(0, 0, -30)
		case "90d":
			startDate = endDate.AddDate(0, 0, -90)
		case "1y":
			startDate = endDate.AddDate(-1, 0, 0)
		}
	}

	return startDate, endDate
}

func parseLimit(r *http.Request, defaultLimit int) int {
	limitStr := r.URL.Query().Get("limit")
	if limitStr == "" {
		return defaultLimit
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		return defaultLimit
	}

	if limit > 1000 {
		return 1000
	}

	return limit
}
