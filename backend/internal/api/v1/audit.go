// Package v1 provides Audit API handlers for OweHost
package v1

import (
	"net/http"
	"strconv"
	"time"

	"github.com/iSundram/OweHost/internal/api/middleware"
	"github.com/iSundram/OweHost/internal/audit"
	"github.com/iSundram/OweHost/internal/user"
	"github.com/iSundram/OweHost/pkg/models"
	"github.com/iSundram/OweHost/pkg/utils"
)

// AuditHandler handles audit-related API requests
type AuditHandler struct {
	auditService *audit.Service
	userService  *user.Service
}

// NewAuditHandler creates a new audit handler
func NewAuditHandler(auditService *audit.Service, userService *user.Service) *AuditHandler {
	return &AuditHandler{
		auditService: auditService,
		userService:  userService,
	}
}

// ListLogs lists audit logs (admin only)
func (h *AuditHandler) ListLogs(w http.ResponseWriter, r *http.Request) {
	query := h.parseQuery(r)
	result := h.auditService.Query(query)
	utils.WriteJSON(w, http.StatusOK, result)
}

// GetStats returns audit statistics
func (h *AuditHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	period := r.URL.Query().Get("period")
	if period == "" {
		period = "last_24h"
	}

	stats := h.auditService.GetStats(period)
	utils.WriteJSON(w, http.StatusOK, stats)
}

// GetUserActivity returns activity logs for a user
func (h *AuditHandler) GetUserActivity(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.ContextKeyUserID).(string)
	userRole := r.Context().Value(middleware.ContextKeyUserRole)

	// Check if requesting another user's activity
	targetUserID := r.URL.Query().Get("user_id")
	if targetUserID != "" && targetUserID != userID {
		if userRole == nil || userRole.(string) != "admin" {
			utils.WriteError(w, http.StatusForbidden, utils.ErrCodeForbidden, "Access denied")
			return
		}
		userID = targetUserID
	}

	limit := 50
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	logs := h.auditService.GetUserActivity(userID, limit)
	utils.WriteJSON(w, http.StatusOK, logs)
}

// GetResourceLogs returns logs for a specific resource
func (h *AuditHandler) GetResourceLogs(w http.ResponseWriter, r *http.Request) {
	resource := r.URL.Query().Get("resource")
	resourceID := r.URL.Query().Get("resource_id")

	if resource == "" {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Resource type is required")
		return
	}

	limit := 50
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	logs := h.auditService.GetRecentLogsForResource(resource, resourceID, limit)
	utils.WriteJSON(w, http.StatusOK, logs)
}

// GetSecurityEvents returns security events (admin only)
func (h *AuditHandler) GetSecurityEvents(w http.ResponseWriter, r *http.Request) {
	limit := 50
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	unresolvedOnly := r.URL.Query().Get("unresolved") == "true"

	events := h.auditService.GetSecurityEvents(limit, unresolvedOnly)
	utils.WriteJSON(w, http.StatusOK, events)
}

// ResolveSecurityEvent marks a security event as resolved
func (h *AuditHandler) ResolveSecurityEvent(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.ContextKeyUserID).(string)

	id := extractIDFromPath(r.URL.Path, "security-events")

	user, _ := h.userService.Get(userID)
	resolvedBy := ""
	if user != nil {
		resolvedBy = user.Username
	}

	if err := h.auditService.ResolveSecurityEvent(id, resolvedBy); err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, err.Error())
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]string{"message": "Security event resolved"})
}

// parseQuery parses query parameters into AuditLogQuery
func (h *AuditHandler) parseQuery(r *http.Request) *models.AuditLogQuery {
	query := &models.AuditLogQuery{
		UserID:     r.URL.Query().Get("user_id"),
		TenantID:   r.URL.Query().Get("tenant_id"),
		Resource:   r.URL.Query().Get("resource"),
		ResourceID: r.URL.Query().Get("resource_id"),
		IPAddress:  r.URL.Query().Get("ip"),
		SortBy:     r.URL.Query().Get("sort_by"),
		SortOrder:  r.URL.Query().Get("sort_order"),
	}

	if action := r.URL.Query().Get("action"); action != "" {
		query.Action = models.AuditAction(action)
	}

	if severity := r.URL.Query().Get("severity"); severity != "" {
		query.Severity = models.AuditSeverity(severity)
	}

	if limit := r.URL.Query().Get("limit"); limit != "" {
		if parsed, err := strconv.Atoi(limit); err == nil {
			query.Limit = parsed
		}
	}
	if query.Limit <= 0 {
		query.Limit = 50
	}

	if offset := r.URL.Query().Get("offset"); offset != "" {
		if parsed, err := strconv.Atoi(offset); err == nil {
			query.Offset = parsed
		}
	}

	if startTime := r.URL.Query().Get("start_time"); startTime != "" {
		if t, err := time.Parse(time.RFC3339, startTime); err == nil {
			query.StartTime = &t
		}
	}

	if endTime := r.URL.Query().Get("end_time"); endTime != "" {
		if t, err := time.Parse(time.RFC3339, endTime); err == nil {
			query.EndTime = &t
		}
	}

	if successOnly := r.URL.Query().Get("success_only"); successOnly == "true" {
		val := true
		query.SuccessOnly = &val
	}

	return query
}
