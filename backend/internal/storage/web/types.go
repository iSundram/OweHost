// Package web provides filesystem-based web/site state management
package web

// SiteDescriptor represents site.json - configuration for a website
type SiteDescriptor struct {
	Domain       string            `json:"domain"`
	Runtime      string            `json:"runtime"`       // php-8.2, php-8.1, nodejs-20, python-3.12
	SSL          bool              `json:"ssl"`
	SSLRedirect  bool              `json:"ssl_redirect"`  // Force HTTPS redirect
	DocumentRoot string            `json:"document_root"` // Relative to site directory (e.g., "public")
	Aliases      []string          `json:"aliases,omitempty"`
	Redirects    []Redirect        `json:"redirects,omitempty"`
	ErrorPages   map[string]string `json:"error_pages,omitempty"` // e.g., {"404": "404.html"}
	Headers      map[string]string `json:"headers,omitempty"`     // Custom headers
	PHPSettings  *PHPSettings      `json:"php_settings,omitempty"`
	NodeSettings *NodeSettings     `json:"node_settings,omitempty"`
	CreatedAt    string            `json:"created_at"`
	UpdatedAt    string            `json:"updated_at"`
}

// Redirect represents a URL redirect configuration
type Redirect struct {
	Source     string `json:"source"`      // Source path or pattern
	Target     string `json:"target"`      // Target URL
	Code       int    `json:"code"`        // HTTP status code (301, 302, 307, 308)
	IsRegex    bool   `json:"is_regex"`    // Whether source is a regex pattern
	IsWildcard bool   `json:"is_wildcard"` // Whether to match wildcards
}

// PHPSettings represents PHP-specific configuration
type PHPSettings struct {
	Version           string            `json:"version"`            // e.g., "8.2"
	MaxExecutionTime  int               `json:"max_execution_time"` // seconds
	MemoryLimit       string            `json:"memory_limit"`       // e.g., "256M"
	PostMaxSize       string            `json:"post_max_size"`      // e.g., "64M"
	UploadMaxFilesize string            `json:"upload_max_filesize"`
	MaxInputVars      int               `json:"max_input_vars"`
	DisplayErrors     bool              `json:"display_errors"`
	Extensions        []string          `json:"extensions,omitempty"` // Additional extensions
	CustomINI         map[string]string `json:"custom_ini,omitempty"` // Custom php.ini directives
}

// NodeSettings represents Node.js-specific configuration
type NodeSettings struct {
	Version        string            `json:"version"`         // e.g., "20"
	Port           int               `json:"port"`            // Application port
	StartCommand   string            `json:"start_command"`   // e.g., "npm start"
	Environment    map[string]string `json:"environment"`     // Environment variables
	AutoRestart    bool              `json:"auto_restart"`
	MaxRestarts    int               `json:"max_restarts"`
	WatchFiles     bool              `json:"watch_files"`     // Auto-restart on file changes
	PassengerMode  bool              `json:"passenger_mode"`  // Use Passenger for process management
}

// PythonSettings represents Python-specific configuration
type PythonSettings struct {
	Version      string            `json:"version"`       // e.g., "3.12"
	AppPath      string            `json:"app_path"`      // Path to WSGI/ASGI app
	Framework    string            `json:"framework"`     // django, flask, fastapi
	Environment  map[string]string `json:"environment"`
	VirtualEnv   string            `json:"virtualenv"`    // Path to virtualenv
	WorkerCount  int               `json:"worker_count"`
}

// SSLMeta represents SSL certificate metadata (stored in ssl/{domain}/meta.json)
type SSLMeta struct {
	Domain       string   `json:"domain"`
	Type         string   `json:"type"`          // letsencrypt, custom, self-signed
	Issuer       string   `json:"issuer"`
	ValidFrom    string   `json:"valid_from"`
	ValidUntil   string   `json:"valid_until"`
	AutoRenew    bool     `json:"auto_renew"`
	SANs         []string `json:"sans,omitempty"` // Subject Alternative Names
	LastRenewed  string   `json:"last_renewed,omitempty"`
	RenewalError string   `json:"renewal_error,omitempty"`
}

// AccessLog represents a log entry
type AccessLog struct {
	Timestamp    string `json:"timestamp"`
	IP           string `json:"ip"`
	Method       string `json:"method"`
	Path         string `json:"path"`
	Status       int    `json:"status"`
	Size         int64  `json:"size"`
	Referer      string `json:"referer"`
	UserAgent    string `json:"user_agent"`
	ResponseTime int    `json:"response_time_ms"`
}

// Valid runtimes
var ValidRuntimes = map[string]bool{
	"php-7.4":     true,
	"php-8.0":     true,
	"php-8.1":     true,
	"php-8.2":     true,
	"php-8.3":     true,
	"nodejs-18":   true,
	"nodejs-20":   true,
	"nodejs-22":   true,
	"python-3.10": true,
	"python-3.11": true,
	"python-3.12": true,
	"static":      true, // Static files only
}

// IsValidRuntime checks if a runtime is valid
func IsValidRuntime(runtime string) bool {
	return ValidRuntimes[runtime]
}

// DefaultPHPSettings returns default PHP settings
func DefaultPHPSettings(version string) *PHPSettings {
	return &PHPSettings{
		Version:           version,
		MaxExecutionTime:  300,
		MemoryLimit:       "256M",
		PostMaxSize:       "64M",
		UploadMaxFilesize: "64M",
		MaxInputVars:      5000,
		DisplayErrors:     false,
		Extensions:        []string{},
	}
}

// DefaultNodeSettings returns default Node.js settings
func DefaultNodeSettings(version string) *NodeSettings {
	return &NodeSettings{
		Version:       version,
		Port:          3000,
		StartCommand:  "npm start",
		AutoRestart:   true,
		MaxRestarts:   10,
		WatchFiles:    false,
		PassengerMode: true,
	}
}
