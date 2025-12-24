package models

import "time"

// OSUpdateStatus represents the status of an OS update
type OSUpdateStatus string

const (
	OSUpdateStatusPending    OSUpdateStatus = "pending"
	OSUpdateStatusDownloading OSUpdateStatus = "downloading"
	OSUpdateStatusApplying   OSUpdateStatus = "applying"
	OSUpdateStatusCompleted  OSUpdateStatus = "completed"
	OSUpdateStatusFailed     OSUpdateStatus = "failed"
	OSUpdateStatusRolledBack OSUpdateStatus = "rolled_back"
)

// ImmutableBaseConfig represents the immutable base configuration
type ImmutableBaseConfig struct {
	ID                string    `json:"id"`
	Version           string    `json:"version"`
	WriteProtected    bool      `json:"write_protected"`
	ProtectedPaths    []string  `json:"protected_paths"`
	AllowedWritePaths []string  `json:"allowed_write_paths"`
	Checksum          string    `json:"checksum"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// OSUpdate represents an OS update
type OSUpdate struct {
	ID             string         `json:"id"`
	Version        string         `json:"version"`
	PreviousVersion string        `json:"previous_version"`
	Status         OSUpdateStatus `json:"status"`
	DownloadURL    string         `json:"download_url"`
	Checksum       string         `json:"checksum"`
	Size           int64          `json:"size"`
	ReleaseNotes   string         `json:"release_notes"`
	Atomic         bool           `json:"atomic"`
	ScheduledAt    *time.Time     `json:"scheduled_at,omitempty"`
	StartedAt      *time.Time     `json:"started_at,omitempty"`
	CompletedAt    *time.Time     `json:"completed_at,omitempty"`
	ErrorMessage   *string        `json:"error_message,omitempty"`
	CreatedAt      time.Time      `json:"created_at"`
}

// WriteProtectionRule represents a write protection rule
type WriteProtectionRule struct {
	ID        string    `json:"id"`
	Path      string    `json:"path"`
	Recursive bool      `json:"recursive"`
	Enabled   bool      `json:"enabled"`
	CreatedAt time.Time `json:"created_at"`
}

// OSUpdateScheduleRequest represents a request to schedule an OS update
type OSUpdateScheduleRequest struct {
	UpdateID    string     `json:"update_id" validate:"required"`
	ScheduledAt *time.Time `json:"scheduled_at,omitempty"`
}

// OSUpdateRollbackRequest represents a request to rollback an OS update
type OSUpdateRollbackRequest struct {
	UpdateID string `json:"update_id" validate:"required"`
}
