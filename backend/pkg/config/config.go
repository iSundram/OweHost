// Package config provides configuration management for OweHost
package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds all configuration for the application
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Auth     AuthConfig
	Cluster  ClusterConfig
	License  LicenseConfig
}

// ServerConfig holds server-related configuration
type ServerConfig struct {
	Host         string
	Port         int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	EnableGRPC   bool
	GRPCPort     int
	
	// Panel ports
	UserPanelPort     int // Port 2083 - User cPanel-like interface
	AdminPanelPort    int // Port 2087 - WHM/Admin interface
	ResellerPanelPort int // Port 2086 - Reseller interface
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Driver   string
	Host     string
	Port     int
	User     string
	Password string
	Name     string
	SSLMode  string
}

// AuthConfig holds authentication configuration
type AuthConfig struct {
	JWTSecret           string
	JWTExpiry           time.Duration
	RefreshTokenExpiry  time.Duration
	APIKeyPrefix        string
	PasswordMinLength   int
	MaxConcurrentSessions int
}

// ClusterConfig holds cluster configuration
type ClusterConfig struct {
	NodeID            string
	HeartbeatInterval time.Duration
	DiscoveryEnabled  bool
}

// LicenseConfig holds licensing configuration
type LicenseConfig struct {
	Key               string
	OfflineGraceDays  int
	ValidationURL     string
}

// Load loads configuration from environment variables with defaults
func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Host:              getEnv("OWEHOST_HOST", "0.0.0.0"),
			Port:              getEnvInt("OWEHOST_PORT", 8080),
			ReadTimeout:       time.Duration(getEnvInt("OWEHOST_READ_TIMEOUT", 30)) * time.Second,
			WriteTimeout:      time.Duration(getEnvInt("OWEHOST_WRITE_TIMEOUT", 30)) * time.Second,
			EnableGRPC:        getEnvBool("OWEHOST_ENABLE_GRPC", false),
			GRPCPort:          getEnvInt("OWEHOST_GRPC_PORT", 9090),
			UserPanelPort:     getEnvInt("OWEHOST_USER_PANEL_PORT", 2083),
			AdminPanelPort:    getEnvInt("OWEHOST_ADMIN_PANEL_PORT", 2087),
			ResellerPanelPort: getEnvInt("OWEHOST_RESELLER_PANEL_PORT", 2086),
		},
		Database: DatabaseConfig{
			Driver:   getEnv("OWEHOST_DB_DRIVER", "postgres"),
			Host:     getEnv("OWEHOST_DB_HOST", "localhost"),
			Port:     getEnvInt("OWEHOST_DB_PORT", 5432),
			User:     getEnv("OWEHOST_DB_USER", "owehost"),
			Password: getEnv("OWEHOST_DB_PASSWORD", ""),
			Name:     getEnv("OWEHOST_DB_NAME", "owehost"),
			SSLMode:  getEnv("OWEHOST_DB_SSLMODE", "disable"),
		},
		Auth: AuthConfig{
			JWTSecret:             getEnv("OWEHOST_JWT_SECRET", "change-me-in-production"),
			JWTExpiry:             time.Duration(getEnvInt("OWEHOST_JWT_EXPIRY_MINUTES", 15)) * time.Minute,
			RefreshTokenExpiry:    time.Duration(getEnvInt("OWEHOST_REFRESH_EXPIRY_DAYS", 7)) * 24 * time.Hour,
			APIKeyPrefix:          getEnv("OWEHOST_API_KEY_PREFIX", "owh_"),
			PasswordMinLength:     getEnvInt("OWEHOST_PASSWORD_MIN_LENGTH", 8),
			MaxConcurrentSessions: getEnvInt("OWEHOST_MAX_SESSIONS", 5),
		},
		Cluster: ClusterConfig{
			NodeID:            getEnv("OWEHOST_NODE_ID", "node-1"),
			HeartbeatInterval: time.Duration(getEnvInt("OWEHOST_HEARTBEAT_SECONDS", 30)) * time.Second,
			DiscoveryEnabled:  getEnvBool("OWEHOST_DISCOVERY_ENABLED", false),
		},
		License: LicenseConfig{
			Key:              getEnv("OWEHOST_LICENSE_KEY", ""),
			OfflineGraceDays: getEnvInt("OWEHOST_OFFLINE_GRACE_DAYS", 7),
			ValidationURL:    getEnv("OWEHOST_LICENSE_URL", "https://license.owehost.com/validate"),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return boolVal
		}
	}
	return defaultValue
}
