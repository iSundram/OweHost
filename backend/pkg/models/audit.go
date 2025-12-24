// Package models provides audit logging data models for OweHost
package models

import "time"

// AuditAction represents the type of audited action
type AuditAction string

const (
	AuditActionCreate   AuditAction = "create"
	AuditActionRead     AuditAction = "read"
	AuditActionUpdate   AuditAction = "update"
	AuditActionDelete   AuditAction = "delete"
	AuditActionLogin    AuditAction = "login"
	AuditActionLogout   AuditAction = "logout"
	AuditActionExecute  AuditAction = "execute"
	AuditActionDownload AuditAction = "download"
	AuditActionUpload   AuditAction = "upload"
)

// AuditSeverity represents the severity level of an audit event
type AuditSeverity string

const (
	AuditSeverityInfo     AuditSeverity = "info"
	AuditSeverityWarning  AuditSeverity = "warning"
	AuditSeverityCritical AuditSeverity = "critical"
)

// AuditLog represents an audit log entry
type AuditLog struct {
	ID           string                 `json:"id"`
	Timestamp    time.Time              `json:"timestamp"`
	UserID       string                 `json:"user_id,omitempty"`
	Username     string                 `json:"username,omitempty"`
	TenantID     string                 `json:"tenant_id,omitempty"`
	Action       AuditAction            `json:"action"`
	Resource     string                 `json:"resource"`      // e.g., "domain", "database", "user"
	ResourceID   string                 `json:"resource_id,omitempty"`
	ResourceName string                 `json:"resource_name,omitempty"`
	Severity     AuditSeverity          `json:"severity"`
	IPAddress    string                 `json:"ip_address"`
	UserAgent    string                 `json:"user_agent,omitempty"`
	RequestID    string                 `json:"request_id,omitempty"`
	Description  string                 `json:"description"`
	OldValue     map[string]interface{} `json:"old_value,omitempty"`
	NewValue     map[string]interface{} `json:"new_value,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	Success      bool                   `json:"success"`
	ErrorMessage string                 `json:"error_message,omitempty"`
}

// AuditLogQuery represents query parameters for audit log search
type AuditLogQuery struct {
	UserID       string        `json:"user_id,omitempty"`
	TenantID     string        `json:"tenant_id,omitempty"`
	Action       AuditAction   `json:"action,omitempty"`
	Resource     string        `json:"resource,omitempty"`
	ResourceID   string        `json:"resource_id,omitempty"`
	Severity     AuditSeverity `json:"severity,omitempty"`
	IPAddress    string        `json:"ip_address,omitempty"`
	StartTime    *time.Time    `json:"start_time,omitempty"`
	EndTime      *time.Time    `json:"end_time,omitempty"`
	SuccessOnly  *bool         `json:"success_only,omitempty"`
	Limit        int           `json:"limit,omitempty"`
	Offset       int           `json:"offset,omitempty"`
	SortBy       string        `json:"sort_by,omitempty"`
	SortOrder    string        `json:"sort_order,omitempty"` // asc, desc
}

// AuditLogResponse represents paginated audit log results
type AuditLogResponse struct {
	Logs       []*AuditLog `json:"logs"`
	Total      int         `json:"total"`
	Limit      int         `json:"limit"`
	Offset     int         `json:"offset"`
	HasMore    bool        `json:"has_more"`
}

// AuditStats represents audit statistics
type AuditStats struct {
	TotalEvents      int            `json:"total_events"`
	EventsByAction   map[string]int `json:"events_by_action"`
	EventsBySeverity map[string]int `json:"events_by_severity"`
	EventsByResource map[string]int `json:"events_by_resource"`
	FailedEvents     int            `json:"failed_events"`
	UniqueUsers      int            `json:"unique_users"`
	UniqueIPs        int            `json:"unique_ips"`
	Period           string         `json:"period"` // e.g., "last_24h", "last_7d"
}

// SecurityEvent represents a security-related event
type SecurityEvent struct {
	ID          string    `json:"id"`
	Timestamp   time.Time `json:"timestamp"`
	EventType   string    `json:"event_type"` // brute_force, unauthorized_access, suspicious_activity
	UserID      string    `json:"user_id,omitempty"`
	IPAddress   string    `json:"ip_address"`
	Description string    `json:"description"`
	Severity    string    `json:"severity"`
	Resolved    bool      `json:"resolved"`
	ResolvedAt  *time.Time `json:"resolved_at,omitempty"`
	ResolvedBy  string    `json:"resolved_by,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}
