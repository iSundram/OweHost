package models

import "time"

// LicenseStatus represents the status of a license
type LicenseStatus string

const (
	LicenseStatusValid     LicenseStatus = "valid"
	LicenseStatusExpired   LicenseStatus = "expired"
	LicenseStatusInvalid   LicenseStatus = "invalid"
	LicenseStatusGrace     LicenseStatus = "grace"
)

// License represents a license
type License struct {
	ID             string            `json:"id"`
	Key            string            `json:"-"`
	Status         LicenseStatus     `json:"status"`
	ProductID      string            `json:"product_id"`
	CustomerID     string            `json:"customer_id"`
	Plan           string            `json:"plan"`
	Features       []string          `json:"features"`
	MaxUsers       int               `json:"max_users"`
	MaxDomains     int               `json:"max_domains"`
	MaxNodes       int               `json:"max_nodes"`
	ExpiresAt      time.Time         `json:"expires_at"`
	GraceExpiresAt *time.Time        `json:"grace_expires_at,omitempty"`
	LastValidated  time.Time         `json:"last_validated"`
	CreatedAt      time.Time         `json:"created_at"`
	UpdatedAt      time.Time         `json:"updated_at"`
}

// FeatureFlag represents a feature flag
type FeatureFlag struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Enabled     bool                   `json:"enabled"`
	Plans       []string               `json:"plans"`
	Conditions  map[string]interface{} `json:"conditions,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// LicenseValidationRequest represents a request to validate a license
type LicenseValidationRequest struct {
	Key        string `json:"key" validate:"required"`
	ProductID  string `json:"product_id" validate:"required"`
	MachineID  string `json:"machine_id" validate:"required"`
}

// LicenseValidationResponse represents a license validation response
type LicenseValidationResponse struct {
	Valid          bool          `json:"valid"`
	Status         LicenseStatus `json:"status"`
	ExpiresAt      time.Time     `json:"expires_at"`
	Features       []string      `json:"features"`
	Plan           string        `json:"plan"`
	MaxUsers       int           `json:"max_users"`
	MaxDomains     int           `json:"max_domains"`
	MaxNodes       int           `json:"max_nodes"`
	Message        string        `json:"message,omitempty"`
}

// FeatureFlagCheckRequest represents a request to check a feature flag
type FeatureFlagCheckRequest struct {
	Feature string            `json:"feature" validate:"required"`
	Context map[string]string `json:"context,omitempty"`
}

// FeatureFlagCheckResponse represents a feature flag check response
type FeatureFlagCheckResponse struct {
	Enabled bool   `json:"enabled"`
	Reason  string `json:"reason,omitempty"`
}
