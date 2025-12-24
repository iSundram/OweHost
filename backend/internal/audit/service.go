// Package audit provides audit logging for OweHost
package audit

import (
	"sync"
	"time"

	"github.com/iSundram/OweHost/pkg/models"
	"github.com/iSundram/OweHost/pkg/utils"
)

// Service provides audit logging functionality
type Service struct {
	logs           []*models.AuditLog
	securityEvents []*models.SecurityEvent
	maxLogs        int
	mu             sync.RWMutex
}

// NewService creates a new audit service
func NewService() *Service {
	return &Service{
		logs:           make([]*models.AuditLog, 0),
		securityEvents: make([]*models.SecurityEvent, 0),
		maxLogs:        100000, // Keep last 100k logs in memory
	}
}

// Log creates an audit log entry
func (s *Service) Log(log *models.AuditLog) {
	s.mu.Lock()
	defer s.mu.Unlock()

	log.ID = utils.GenerateID("audit")
	log.Timestamp = time.Now()

	s.logs = append(s.logs, log)

	// Trim if exceeds max
	if len(s.logs) > s.maxLogs {
		s.logs = s.logs[len(s.logs)-s.maxLogs:]
	}
}

// LogAction is a convenience method to log an action
func (s *Service) LogAction(
	userID, username, tenantID string,
	action models.AuditAction,
	resource, resourceID, resourceName string,
	severity models.AuditSeverity,
	ipAddress, userAgent, requestID, description string,
	success bool,
	errorMessage string,
	oldValue, newValue, metadata map[string]interface{},
) {
	log := &models.AuditLog{
		UserID:       userID,
		Username:     username,
		TenantID:     tenantID,
		Action:       action,
		Resource:     resource,
		ResourceID:   resourceID,
		ResourceName: resourceName,
		Severity:     severity,
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
		RequestID:    requestID,
		Description:  description,
		OldValue:     oldValue,
		NewValue:     newValue,
		Metadata:     metadata,
		Success:      success,
		ErrorMessage: errorMessage,
	}
	s.Log(log)
}

// LogCreate logs a create action
func (s *Service) LogCreate(userID, username, resource, resourceID, resourceName, ipAddress, requestID string) {
	s.LogAction(
		userID, username, "",
		models.AuditActionCreate,
		resource, resourceID, resourceName,
		models.AuditSeverityInfo,
		ipAddress, "", requestID,
		"Created "+resource+": "+resourceName,
		true, "",
		nil, nil, nil,
	)
}

// LogUpdate logs an update action
func (s *Service) LogUpdate(userID, username, resource, resourceID, resourceName, ipAddress, requestID string, oldValue, newValue map[string]interface{}) {
	s.LogAction(
		userID, username, "",
		models.AuditActionUpdate,
		resource, resourceID, resourceName,
		models.AuditSeverityInfo,
		ipAddress, "", requestID,
		"Updated "+resource+": "+resourceName,
		true, "",
		oldValue, newValue, nil,
	)
}

// LogDelete logs a delete action
func (s *Service) LogDelete(userID, username, resource, resourceID, resourceName, ipAddress, requestID string) {
	s.LogAction(
		userID, username, "",
		models.AuditActionDelete,
		resource, resourceID, resourceName,
		models.AuditSeverityWarning,
		ipAddress, "", requestID,
		"Deleted "+resource+": "+resourceName,
		true, "",
		nil, nil, nil,
	)
}

// LogLogin logs a login action
func (s *Service) LogLogin(userID, username, ipAddress, userAgent string, success bool, failReason string) {
	severity := models.AuditSeverityInfo
	if !success {
		severity = models.AuditSeverityWarning
	}
	
	description := "User logged in"
	if !success {
		description = "Login failed: " + failReason
	}

	s.LogAction(
		userID, username, "",
		models.AuditActionLogin,
		"session", "", "",
		severity,
		ipAddress, userAgent, "",
		description,
		success, failReason,
		nil, nil, nil,
	)
}

// LogSecurityEvent logs a security event
func (s *Service) LogSecurityEvent(eventType, userID, ipAddress, description, severity string, metadata map[string]interface{}) *models.SecurityEvent {
	s.mu.Lock()
	defer s.mu.Unlock()

	event := &models.SecurityEvent{
		ID:          utils.GenerateID("sec"),
		Timestamp:   time.Now(),
		EventType:   eventType,
		UserID:      userID,
		IPAddress:   ipAddress,
		Description: description,
		Severity:    severity,
		Metadata:    metadata,
	}

	s.securityEvents = append(s.securityEvents, event)

	// Keep last 10000 security events
	if len(s.securityEvents) > 10000 {
		s.securityEvents = s.securityEvents[len(s.securityEvents)-10000:]
	}

	return event
}

// Query queries audit logs
func (s *Service) Query(query *models.AuditLogQuery) *models.AuditLogResponse {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Filter logs
	filtered := make([]*models.AuditLog, 0)
	for _, log := range s.logs {
		if s.matchesQuery(log, query) {
			filtered = append(filtered, log)
		}
	}

	// Sort (newest first by default)
	if query.SortOrder == "asc" {
		// Already in ascending order (oldest first)
	} else {
		// Reverse for descending (newest first)
		for i, j := 0, len(filtered)-1; i < j; i, j = i+1, j-1 {
			filtered[i], filtered[j] = filtered[j], filtered[i]
		}
	}

	// Apply pagination
	total := len(filtered)
	limit := query.Limit
	if limit <= 0 {
		limit = 50
	}
	offset := query.Offset
	if offset < 0 {
		offset = 0
	}

	start := offset
	if start > len(filtered) {
		start = len(filtered)
	}
	end := start + limit
	if end > len(filtered) {
		end = len(filtered)
	}

	return &models.AuditLogResponse{
		Logs:    filtered[start:end],
		Total:   total,
		Limit:   limit,
		Offset:  offset,
		HasMore: end < total,
	}
}

// matchesQuery checks if a log matches the query
func (s *Service) matchesQuery(log *models.AuditLog, query *models.AuditLogQuery) bool {
	if query.UserID != "" && log.UserID != query.UserID {
		return false
	}
	if query.TenantID != "" && log.TenantID != query.TenantID {
		return false
	}
	if query.Action != "" && log.Action != query.Action {
		return false
	}
	if query.Resource != "" && log.Resource != query.Resource {
		return false
	}
	if query.ResourceID != "" && log.ResourceID != query.ResourceID {
		return false
	}
	if query.Severity != "" && log.Severity != query.Severity {
		return false
	}
	if query.IPAddress != "" && log.IPAddress != query.IPAddress {
		return false
	}
	if query.StartTime != nil && log.Timestamp.Before(*query.StartTime) {
		return false
	}
	if query.EndTime != nil && log.Timestamp.After(*query.EndTime) {
		return false
	}
	if query.SuccessOnly != nil && *query.SuccessOnly && !log.Success {
		return false
	}
	return true
}

// GetStats returns audit statistics
func (s *Service) GetStats(period string) *models.AuditStats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var cutoff time.Time
	switch period {
	case "last_24h":
		cutoff = time.Now().Add(-24 * time.Hour)
	case "last_7d":
		cutoff = time.Now().Add(-7 * 24 * time.Hour)
	case "last_30d":
		cutoff = time.Now().Add(-30 * 24 * time.Hour)
	default:
		cutoff = time.Now().Add(-24 * time.Hour)
		period = "last_24h"
	}

	stats := &models.AuditStats{
		EventsByAction:   make(map[string]int),
		EventsBySeverity: make(map[string]int),
		EventsByResource: make(map[string]int),
		Period:           period,
	}

	uniqueUsers := make(map[string]bool)
	uniqueIPs := make(map[string]bool)

	for _, log := range s.logs {
		if log.Timestamp.Before(cutoff) {
			continue
		}

		stats.TotalEvents++
		stats.EventsByAction[string(log.Action)]++
		stats.EventsBySeverity[string(log.Severity)]++
		stats.EventsByResource[log.Resource]++

		if !log.Success {
			stats.FailedEvents++
		}
		if log.UserID != "" {
			uniqueUsers[log.UserID] = true
		}
		if log.IPAddress != "" {
			uniqueIPs[log.IPAddress] = true
		}
	}

	stats.UniqueUsers = len(uniqueUsers)
	stats.UniqueIPs = len(uniqueIPs)

	return stats
}

// GetSecurityEvents returns security events
func (s *Service) GetSecurityEvents(limit int, unresolvedOnly bool) []*models.SecurityEvent {
	s.mu.RLock()
	defer s.mu.RUnlock()

	events := make([]*models.SecurityEvent, 0)
	for i := len(s.securityEvents) - 1; i >= 0 && len(events) < limit; i-- {
		event := s.securityEvents[i]
		if unresolvedOnly && event.Resolved {
			continue
		}
		events = append(events, event)
	}
	return events
}

// ResolveSecurityEvent marks a security event as resolved
func (s *Service) ResolveSecurityEvent(id, resolvedBy string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, event := range s.securityEvents {
		if event.ID == id {
			event.Resolved = true
			now := time.Now()
			event.ResolvedAt = &now
			event.ResolvedBy = resolvedBy
			return nil
		}
	}
	return nil
}

// GetRecentLogsForResource gets recent logs for a specific resource
func (s *Service) GetRecentLogsForResource(resource, resourceID string, limit int) []*models.AuditLog {
	s.mu.RLock()
	defer s.mu.RUnlock()

	logs := make([]*models.AuditLog, 0)
	for i := len(s.logs) - 1; i >= 0 && len(logs) < limit; i-- {
		log := s.logs[i]
		if log.Resource == resource && (resourceID == "" || log.ResourceID == resourceID) {
			logs = append(logs, log)
		}
	}
	return logs
}

// GetUserActivity gets activity logs for a user
func (s *Service) GetUserActivity(userID string, limit int) []*models.AuditLog {
	s.mu.RLock()
	defer s.mu.RUnlock()

	logs := make([]*models.AuditLog, 0)
	for i := len(s.logs) - 1; i >= 0 && len(logs) < limit; i-- {
		if s.logs[i].UserID == userID {
			logs = append(logs, s.logs[i])
		}
	}
	return logs
}
