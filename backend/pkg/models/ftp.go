// Package models provides FTP-related data models for OweHost
package models

import "time"

// FTPAccount represents an FTP account
type FTPAccount struct {
	ID            string    `json:"id"`
	UserID        string    `json:"user_id"`
	Username      string    `json:"username"`
	PasswordHash  string    `json:"-"`
	HomeDirectory string    `json:"home_directory"`
	QuotaMB       int       `json:"quota_mb"`
	UsedMB        int       `json:"used_mb"`
	Status        string    `json:"status"` // active, suspended, disabled
	ReadOnly      bool      `json:"read_only"`
	IPWhitelist   []string  `json:"ip_whitelist,omitempty"`
	LastLoginAt   *time.Time `json:"last_login_at,omitempty"`
	LastLoginIP   string    `json:"last_login_ip,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// FTPAccountCreateRequest represents a request to create an FTP account
type FTPAccountCreateRequest struct {
	Username      string   `json:"username"`
	Password      string   `json:"password"`
	HomeDirectory string   `json:"home_directory"`
	QuotaMB       int      `json:"quota_mb"`
	ReadOnly      bool     `json:"read_only"`
	IPWhitelist   []string `json:"ip_whitelist,omitempty"`
}

// FTPAccountUpdateRequest represents a request to update an FTP account
type FTPAccountUpdateRequest struct {
	Password      *string   `json:"password,omitempty"`
	HomeDirectory *string   `json:"home_directory,omitempty"`
	QuotaMB       *int      `json:"quota_mb,omitempty"`
	Status        *string   `json:"status,omitempty"`
	ReadOnly      *bool     `json:"read_only,omitempty"`
	IPWhitelist   *[]string `json:"ip_whitelist,omitempty"`
}

// FTPSession represents an active FTP session
type FTPSession struct {
	ID          string    `json:"id"`
	AccountID   string    `json:"account_id"`
	Username    string    `json:"username"`
	RemoteIP    string    `json:"remote_ip"`
	StartedAt   time.Time `json:"started_at"`
	CurrentPath string    `json:"current_path"`
	BytesIn     int64     `json:"bytes_in"`
	BytesOut    int64     `json:"bytes_out"`
}

// FTPConfig represents FTP server configuration
type FTPConfig struct {
	Port            int    `json:"port"`
	PassivePortMin  int    `json:"passive_port_min"`
	PassivePortMax  int    `json:"passive_port_max"`
	MaxConnections  int    `json:"max_connections"`
	MaxPerIP        int    `json:"max_per_ip"`
	IdleTimeout     int    `json:"idle_timeout"` // seconds
	TLSRequired     bool   `json:"tls_required"`
	AnonymousAccess bool   `json:"anonymous_access"`
}
