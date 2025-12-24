package models

import "time"

// InstallationStatus represents the installation status
type InstallationStatus string

const (
	InstallationStatusPending    InstallationStatus = "pending"
	InstallationStatusInProgress InstallationStatus = "in_progress"
	InstallationStatusCompleted  InstallationStatus = "completed"
	InstallationStatusFailed     InstallationStatus = "failed"
)

// Installation represents system installation state
type Installation struct {
	ID                 string             `json:"id"`
	Status             InstallationStatus `json:"status"`
	DatabaseEngine     DatabaseType       `json:"database_engine"`
	DatabaseHost       string             `json:"database_host"`
	DatabasePort       int                `json:"database_port"`
	DatabaseName       string             `json:"database_name"`
	DatabaseUser       string             `json:"database_user"`
	AdminUsername      string             `json:"admin_username"`
	AdminEmail         string             `json:"admin_email"`
	InstallationStep   int                `json:"installation_step"`
	TotalSteps         int                `json:"total_steps"`
	ErrorMessage       string             `json:"error_message,omitempty"`
	CreatedAt          time.Time          `json:"created_at"`
	CompletedAt        *time.Time         `json:"completed_at,omitempty"`
}

// InstallationRequest represents installation configuration
type InstallationRequest struct {
	DatabaseEngine   DatabaseType `json:"database_engine" validate:"required,oneof=mysql postgresql mariadb mongodb redis sqlite"`
	DatabaseHost     string       `json:"database_host" validate:"required"`
	DatabasePort     int          `json:"database_port" validate:"required,min=1,max=65535"`
	DatabaseName     string       `json:"database_name" validate:"required"`
	DatabaseUser     string       `json:"database_user" validate:"required"`
	DatabasePassword string       `json:"database_password" validate:"required"`
	AdminUsername    string       `json:"admin_username" validate:"required,min=3,max=32"`
	AdminPassword    string       `json:"admin_password" validate:"required,min=8"`
	AdminEmail       string       `json:"admin_email" validate:"required,email"`
}

// InstallationCheckResponse represents system installation check
type InstallationCheckResponse struct {
	IsInstalled      bool               `json:"is_installed"`
	RequiresSetup    bool               `json:"requires_setup"`
	SupportedEngines []DatabaseEngine   `json:"supported_engines"`
}

// DatabaseEngine represents available database engine info
type DatabaseEngine struct {
	Type        DatabaseType `json:"type"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	DefaultPort int          `json:"default_port"`
	IsInstalled bool         `json:"is_installed"`
}
