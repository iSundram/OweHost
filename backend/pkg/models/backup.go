package models

import "time"

// BackupStatus represents the status of a backup
type BackupStatus string

const (
	BackupStatusPending    BackupStatus = "pending"
	BackupStatusRunning    BackupStatus = "running"
	BackupStatusCompleted  BackupStatus = "completed"
	BackupStatusFailed     BackupStatus = "failed"
)

// BackupType represents the type of backup
type BackupType string

const (
	BackupTypeFull         BackupType = "full"
	BackupTypeIncremental  BackupType = "incremental"
	BackupTypeDifferential BackupType = "differential"
)

// Backup represents a backup
type Backup struct {
	ID             string       `json:"id"`
	UserID         string       `json:"user_id"`
	Type           BackupType   `json:"type"`
	Status         BackupStatus `json:"status"`
	SizeMB         int64        `json:"size_mb"`
	StoragePath    string       `json:"storage_path"`
	ParentBackupID *string      `json:"parent_backup_id,omitempty"`
	IncludeFiles   bool         `json:"include_files"`
	IncludeDatabases bool       `json:"include_databases"`
	Checksum       string       `json:"checksum"`
	ErrorMessage   *string      `json:"error_message,omitempty"`
	StartedAt      *time.Time   `json:"started_at,omitempty"`
	CompletedAt    *time.Time   `json:"completed_at,omitempty"`
	CreatedAt      time.Time    `json:"created_at"`
}

// BackupSchedule represents a backup schedule
type BackupSchedule struct {
	ID               string     `json:"id"`
	UserID           string     `json:"user_id"`
	Type             BackupType `json:"type"`
	CronExpression   string     `json:"cron_expression"`
	RetentionDays    int        `json:"retention_days"`
	IncludeFiles     bool       `json:"include_files"`
	IncludeDatabases bool       `json:"include_databases"`
	Priority         int        `json:"priority"`
	Enabled          bool       `json:"enabled"`
	LastRunAt        *time.Time `json:"last_run_at,omitempty"`
	NextRunAt        *time.Time `json:"next_run_at,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

// RestoreRequest represents a restore request
type RestoreRequest struct {
	BackupID          string   `json:"backup_id" validate:"required"`
	RestoreFiles      bool     `json:"restore_files"`
	RestoreDatabases  bool     `json:"restore_databases"`
	ConflictResolution string  `json:"conflict_resolution" validate:"oneof=overwrite skip rename"`
	TargetNodeID      *string  `json:"target_node_id,omitempty"`
}

// RestoreStatus represents the status of a restore operation
type RestoreStatus struct {
	ID           string       `json:"id"`
	BackupID     string       `json:"backup_id"`
	Status       BackupStatus `json:"status"`
	Progress     int          `json:"progress"`
	ErrorMessage *string      `json:"error_message,omitempty"`
	StartedAt    *time.Time   `json:"started_at,omitempty"`
	CompletedAt  *time.Time   `json:"completed_at,omitempty"`
}

// BackupCreateRequest represents a request to create a backup
type BackupCreateRequest struct {
	Type             BackupType `json:"type" validate:"required,oneof=full incremental differential"`
	IncludeFiles     bool       `json:"include_files"`
	IncludeDatabases bool       `json:"include_databases"`
	Priority         int        `json:"priority,omitempty"`
}

// BackupScheduleCreateRequest represents a request to create a backup schedule
type BackupScheduleCreateRequest struct {
	Type             BackupType `json:"type" validate:"required,oneof=full incremental differential"`
	CronExpression   string     `json:"cron_expression" validate:"required"`
	RetentionDays    int        `json:"retention_days" validate:"required,min=1"`
	IncludeFiles     bool       `json:"include_files"`
	IncludeDatabases bool       `json:"include_databases"`
	Priority         int        `json:"priority,omitempty"`
}
