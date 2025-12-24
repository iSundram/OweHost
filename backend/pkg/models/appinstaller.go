package models

import "time"

// AppInstallStatus represents the status of an app installation
type AppInstallStatus string

const (
	AppInstallStatusPending   AppInstallStatus = "pending"
	AppInstallStatusInstalling AppInstallStatus = "installing"
	AppInstallStatusCompleted AppInstallStatus = "completed"
	AppInstallStatusFailed    AppInstallStatus = "failed"
	AppInstallStatusUpdating  AppInstallStatus = "updating"
)

// AppDefinition represents an application definition
type AppDefinition struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Slug         string   `json:"slug"`
	Version      string   `json:"version"`
	Category     string   `json:"category"`
	Description  string   `json:"description"`
	Requirements []string `json:"requirements"`
	ManifestURL  string   `json:"manifest_url"`
	IconURL      string   `json:"icon_url"`
	InstallScript string  `json:"install_script"`
	UpdateScript  string  `json:"update_script"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// InstalledApp represents an installed application
type InstalledApp struct {
	ID           string           `json:"id"`
	UserID       string           `json:"user_id"`
	DomainID     string           `json:"domain_id"`
	AppID        string           `json:"app_id"`
	Version      string           `json:"version"`
	InstallPath  string           `json:"install_path"`
	Status       AppInstallStatus `json:"status"`
	Settings     map[string]string `json:"settings"`
	ErrorMessage *string          `json:"error_message,omitempty"`
	InstalledAt  time.Time        `json:"installed_at"`
	UpdatedAt    time.Time        `json:"updated_at"`
}

// AppInstallRequest represents a request to install an application
type AppInstallRequest struct {
	AppID       string            `json:"app_id" validate:"required"`
	DomainID    string            `json:"domain_id" validate:"required"`
	InstallPath string            `json:"install_path,omitempty"`
	Settings    map[string]string `json:"settings,omitempty"`
}

// AppUpdateRequest represents a request to update an application
type AppUpdateRequest struct {
	TargetVersion string `json:"target_version,omitempty"`
}

// AppManifest represents an application manifest
type AppManifest struct {
	Name         string            `json:"name"`
	Version      string            `json:"version"`
	Description  string            `json:"description"`
	Author       string            `json:"author"`
	Website      string            `json:"website"`
	License      string            `json:"license"`
	Requirements map[string]string `json:"requirements"`
	Database     *DatabaseRequirement `json:"database,omitempty"`
	Files        []FileRequirement `json:"files"`
	PostInstall  []string          `json:"post_install"`
}

// DatabaseRequirement represents database requirements for an app
type DatabaseRequirement struct {
	Type     string `json:"type"`
	Name     string `json:"name"`
	User     string `json:"user"`
}

// FileRequirement represents file requirements for an app
type FileRequirement struct {
	Source string `json:"source"`
	Dest   string `json:"dest"`
	Mode   string `json:"mode"`
}
