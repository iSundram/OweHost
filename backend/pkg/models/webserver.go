package models

import "time"

// WebServerType represents the type of web server
type WebServerType string

const (
	WebServerTypeNginx  WebServerType = "nginx"
	WebServerTypeApache WebServerType = "apache"
	WebServerTypeHybrid WebServerType = "hybrid"
)

// VirtualHost represents a virtual host configuration
type VirtualHost struct {
	ID             string        `json:"id"`
	DomainID       string        `json:"domain_id"`
	ServerType     WebServerType `json:"server_type"`
	DocumentRoot   string        `json:"document_root"`
	SSLEnabled     bool          `json:"ssl_enabled"`
	SSLCertID      *string       `json:"ssl_cert_id,omitempty"`
	PHPEnabled     bool          `json:"php_enabled"`
	PHPVersion     *string       `json:"php_version,omitempty"`
	ProxyPass      *string       `json:"proxy_pass,omitempty"`
	ConfigPath     string        `json:"config_path"`
	ConfigChecksum string        `json:"config_checksum"`
	CreatedAt      time.Time     `json:"created_at"`
	UpdatedAt      time.Time     `json:"updated_at"`
}

// WebServerConfig represents web server configuration
type WebServerConfig struct {
	ID           string    `json:"id"`
	VHostID      string    `json:"vhost_id"`
	ConfigData   string    `json:"config_data"`
	Version      int       `json:"version"`
	Active       bool      `json:"active"`
	CreatedAt    time.Time `json:"created_at"`
}

// ConfigReloadStatus represents the status of a config reload
type ConfigReloadStatus struct {
	ServerType   WebServerType `json:"server_type"`
	Success      bool          `json:"success"`
	ErrorMessage *string       `json:"error_message,omitempty"`
	ReloadedAt   time.Time     `json:"reloaded_at"`
}

// VirtualHostCreateRequest represents a request to create a virtual host
type VirtualHostCreateRequest struct {
	DomainID     string        `json:"domain_id" validate:"required"`
	ServerType   WebServerType `json:"server_type" validate:"required,oneof=nginx apache hybrid"`
	DocumentRoot string        `json:"document_root,omitempty"`
	SSLEnabled   bool          `json:"ssl_enabled"`
	PHPEnabled   bool          `json:"php_enabled"`
	PHPVersion   *string       `json:"php_version,omitempty"`
	ProxyPass    *string       `json:"proxy_pass,omitempty"`
}

// WebServerConfigUpdateRequest represents a request to update web server config
type WebServerConfigUpdateRequest struct {
	PHPVersion   *string `json:"php_version,omitempty"`
	SSLEnabled   *bool   `json:"ssl_enabled,omitempty"`
	DocumentRoot *string `json:"document_root,omitempty"`
	ProxyPass    *string `json:"proxy_pass,omitempty"`
}

// UserWebServerConfig represents a user's web server configuration
type UserWebServerConfig struct {
	UserID       string   `json:"user_id"`
	PHPVersion   string   `json:"php_version"`
	SSLEnabled   bool     `json:"ssl_enabled"`
	DocumentRoot string   `json:"document_root"`
	Domains      []string `json:"domains"`
}

// PHPVersion represents an available PHP version
type PHPVersion struct {
	Version   string `json:"version"`
	Available bool   `json:"available"`
	Default   bool   `json:"default"`
}

// WebServerModule represents a web server module
type WebServerModule struct {
	Name        string `json:"name"`
	Enabled     bool   `json:"enabled"`
	Description string `json:"description"`
}

// WebServerStatus represents the web server status
type WebServerStatus struct {
	Running     bool   `json:"running"`
	ServerType  string `json:"server_type"`
	Version     string `json:"version"`
	Uptime      int64  `json:"uptime_seconds"`
	Connections int    `json:"active_connections"`
}

// ConfigTestResult represents the result of a configuration test
type ConfigTestResult struct {
	Valid   bool   `json:"valid"`
	Message string `json:"message"`
}
