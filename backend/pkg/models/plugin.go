package models

import "time"

// PluginStatus represents the status of a plugin
type PluginStatus string

const (
	PluginStatusActive   PluginStatus = "active"
	PluginStatusInactive PluginStatus = "inactive"
	PluginStatusFailed   PluginStatus = "failed"
)

// Plugin represents a plugin
type Plugin struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	Slug        string       `json:"slug"`
	Version     string       `json:"version"`
	Author      string       `json:"author"`
	Description string       `json:"description"`
	Status      PluginStatus `json:"status"`
	Signature   string       `json:"signature"`
	Verified    bool         `json:"verified"`
	APIScopes   []string     `json:"api_scopes"`
	ConfigSchema map[string]interface{} `json:"config_schema"`
	Config      map[string]interface{} `json:"config"`
	InstalledAt time.Time    `json:"installed_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

// PluginHook represents a plugin hook
type PluginHook struct {
	ID        string    `json:"id"`
	PluginID  string    `json:"plugin_id"`
	Event     string    `json:"event"`
	Handler   string    `json:"handler"`
	Priority  int       `json:"priority"`
	CreatedAt time.Time `json:"created_at"`
}

// PluginAPIScope represents an API scope for plugins
type PluginAPIScope struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Permissions []string `json:"permissions"`
}

// PluginInstallRequest represents a request to install a plugin
type PluginInstallRequest struct {
	PackageURL string `json:"package_url" validate:"required,url"`
	Signature  string `json:"signature" validate:"required"`
}

// PluginConfigRequest represents a request to configure a plugin
type PluginConfigRequest struct {
	Config map[string]interface{} `json:"config" validate:"required"`
}
