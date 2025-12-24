package models

import "time"

// RuntimeType represents the type of runtime
type RuntimeType string

const (
	RuntimeTypePHP    RuntimeType = "php"
	RuntimeTypeNodeJS RuntimeType = "nodejs"
	RuntimeTypePython RuntimeType = "python"
	RuntimeTypeGo     RuntimeType = "go"
	RuntimeTypeJava   RuntimeType = "java"
)

// PHPPool represents a PHP-FPM pool configuration
type PHPPool struct {
	ID            string    `json:"id"`
	UserID        string    `json:"user_id"`
	Version       string    `json:"version"`
	PoolName      string    `json:"pool_name"`
	SocketPath    string    `json:"socket_path"`
	MaxChildren   int       `json:"max_children"`
	StartServers  int       `json:"start_servers"`
	MinSpareServers int     `json:"min_spare_servers"`
	MaxSpareServers int     `json:"max_spare_servers"`
	Extensions    []string  `json:"extensions"`
	INIOverrides  map[string]string `json:"ini_overrides"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// NodeJSApp represents a Node.js application
type NodeJSApp struct {
	ID           string    `json:"id"`
	UserID       string    `json:"user_id"`
	DomainID     string    `json:"domain_id"`
	Name         string    `json:"name"`
	Version      string    `json:"version"`
	AppRoot      string    `json:"app_root"`
	StartFile    string    `json:"start_file"`
	Port         int       `json:"port"`
	Environment  map[string]string `json:"environment"`
	Running      bool      `json:"running"`
	PID          *int      `json:"pid,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// PythonApp represents a Python application
type PythonApp struct {
	ID           string    `json:"id"`
	UserID       string    `json:"user_id"`
	DomainID     string    `json:"domain_id"`
	Name         string    `json:"name"`
	Version      string    `json:"version"`
	VenvPath     string    `json:"venv_path"`
	AppRoot      string    `json:"app_root"`
	WSGIFile     string    `json:"wsgi_file"`
	Port         int       `json:"port"`
	Environment  map[string]string `json:"environment"`
	Running      bool      `json:"running"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// RuntimeVersion represents an available runtime version
type RuntimeVersion struct {
	Type      RuntimeType `json:"type"`
	Version   string      `json:"version"`
	Path      string      `json:"path"`
	Default   bool        `json:"default"`
	Available bool        `json:"available"`
}

// PHPExtension represents a PHP extension
type PHPExtension struct {
	Name         string   `json:"name"`
	Enabled      bool     `json:"enabled"`
	Dependencies []string `json:"dependencies"`
}

// PHPPoolCreateRequest represents a request to create a PHP pool
type PHPPoolCreateRequest struct {
	Version       string            `json:"version" validate:"required"`
	MaxChildren   int               `json:"max_children" validate:"min=1,max=100"`
	Extensions    []string          `json:"extensions,omitempty"`
	INIOverrides  map[string]string `json:"ini_overrides,omitempty"`
}

// NodeJSAppCreateRequest represents a request to create a Node.js app
type NodeJSAppCreateRequest struct {
	DomainID    string            `json:"domain_id" validate:"required"`
	Name        string            `json:"name" validate:"required"`
	Version     string            `json:"version" validate:"required"`
	AppRoot     string            `json:"app_root" validate:"required"`
	StartFile   string            `json:"start_file" validate:"required"`
	Environment map[string]string `json:"environment,omitempty"`
}

// PythonAppCreateRequest represents a request to create a Python app
type PythonAppCreateRequest struct {
	DomainID    string            `json:"domain_id" validate:"required"`
	Name        string            `json:"name" validate:"required"`
	Version     string            `json:"version" validate:"required"`
	AppRoot     string            `json:"app_root" validate:"required"`
	WSGIFile    string            `json:"wsgi_file" validate:"required"`
	Environment map[string]string `json:"environment,omitempty"`
}
