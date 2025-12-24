package models

import "time"

// RollbackType represents the type of rollback
type RollbackType string

const (
	RollbackTypeConfig   RollbackType = "config"
	RollbackTypeDatabase RollbackType = "database"
	RollbackTypeFiles    RollbackType = "files"
	RollbackTypeFull     RollbackType = "full"
)

// RollbackStatus represents the status of a rollback
type RollbackStatus string

const (
	RollbackStatusPending   RollbackStatus = "pending"
	RollbackStatusRunning   RollbackStatus = "running"
	RollbackStatusCompleted RollbackStatus = "completed"
	RollbackStatusFailed    RollbackStatus = "failed"
)

// ConfigSnapshot represents a configuration snapshot
type ConfigSnapshot struct {
	ID           string                 `json:"id"`
	ResourceType string                 `json:"resource_type"`
	ResourceID   string                 `json:"resource_id"`
	Version      int                    `json:"version"`
	Data         map[string]interface{} `json:"data"`
	Checksum     string                 `json:"checksum"`
	CreatedAt    time.Time              `json:"created_at"`
}

// RollbackOperation represents a rollback operation
type RollbackOperation struct {
	ID             string         `json:"id"`
	Type           RollbackType   `json:"type"`
	Status         RollbackStatus `json:"status"`
	ResourceType   string         `json:"resource_type"`
	ResourceID     string         `json:"resource_id"`
	SnapshotID     string         `json:"snapshot_id"`
	InitiatedBy    string         `json:"initiated_by"`
	ErrorMessage   *string        `json:"error_message,omitempty"`
	StartedAt      *time.Time     `json:"started_at,omitempty"`
	CompletedAt    *time.Time     `json:"completed_at,omitempty"`
	CreatedAt      time.Time      `json:"created_at"`
}

// HealthCheckType represents the type of health check
type HealthCheckType string

const (
	HealthCheckTypeHTTP    HealthCheckType = "http"
	HealthCheckTypeTCP     HealthCheckType = "tcp"
	HealthCheckTypeProcess HealthCheckType = "process"
	HealthCheckTypeCustom  HealthCheckType = "custom"
)

// HealthCheckStatus represents the status of a health check
type HealthCheckStatus string

const (
	HealthCheckStatusHealthy   HealthCheckStatus = "healthy"
	HealthCheckStatusUnhealthy HealthCheckStatus = "unhealthy"
	HealthCheckStatusDegraded  HealthCheckStatus = "degraded"
	HealthCheckStatusUnknown   HealthCheckStatus = "unknown"
)

// HealthCheck represents a health check configuration
type HealthCheck struct {
	ID            string          `json:"id"`
	Name          string          `json:"name"`
	Type          HealthCheckType `json:"type"`
	Target        string          `json:"target"`
	Interval      int             `json:"interval_seconds"`
	Timeout       int             `json:"timeout_seconds"`
	Retries       int             `json:"retries"`
	Status        HealthCheckStatus `json:"status"`
	LastCheckAt   *time.Time      `json:"last_check_at,omitempty"`
	LastHealthyAt *time.Time      `json:"last_healthy_at,omitempty"`
	ErrorMessage  *string         `json:"error_message,omitempty"`
	SelfHealing   bool            `json:"self_healing"`
	HealingAction string          `json:"healing_action,omitempty"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
}

// HealthCheckResult represents a health check result
type HealthCheckResult struct {
	ID            string            `json:"id"`
	HealthCheckID string            `json:"health_check_id"`
	Status        HealthCheckStatus `json:"status"`
	ResponseTime  int               `json:"response_time_ms"`
	Details       string            `json:"details,omitempty"`
	CheckedAt     time.Time         `json:"checked_at"`
}

// SelfHealingTrigger represents a self-healing trigger
type SelfHealingTrigger struct {
	ID            string    `json:"id"`
	HealthCheckID string    `json:"health_check_id"`
	TriggeredAt   time.Time `json:"triggered_at"`
	Action        string    `json:"action"`
	Success       bool      `json:"success"`
	Details       string    `json:"details,omitempty"`
}

// RollbackRequest represents a request to rollback
type RollbackRequest struct {
	ResourceType string `json:"resource_type" validate:"required"`
	ResourceID   string `json:"resource_id" validate:"required"`
	SnapshotID   string `json:"snapshot_id,omitempty"`
	Type         RollbackType `json:"type" validate:"required,oneof=config database files full"`
}

// HealthCheckCreateRequest represents a request to create a health check
type HealthCheckCreateRequest struct {
	Name          string          `json:"name" validate:"required,min=1,max=64"`
	Type          HealthCheckType `json:"type" validate:"required,oneof=http tcp process custom"`
	Target        string          `json:"target" validate:"required"`
	Interval      int             `json:"interval_seconds" validate:"required,min=10"`
	Timeout       int             `json:"timeout_seconds" validate:"required,min=1"`
	Retries       int             `json:"retries" validate:"min=0,max=10"`
	SelfHealing   bool            `json:"self_healing"`
	HealingAction string          `json:"healing_action,omitempty"`
}
