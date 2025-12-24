// Package models provides data models for OweHost
package models

import (
	"time"
)

// UserStatus represents the status of a user account
type UserStatus string

const (
	UserStatusActive    UserStatus = "active"
	UserStatusSuspended UserStatus = "suspended"
	UserStatusTerminated UserStatus = "terminated"
)

// UserRole represents the role of a user account
type UserRole string

const (
	UserRoleAdmin    UserRole = "admin"
	UserRoleReseller UserRole = "reseller"
	UserRoleUser     UserRole = "user"
)

// User represents a user account in the system
type User struct {
	ID            string     `json:"id"`
	TenantID      string     `json:"tenant_id"`
	Username      string     `json:"username"`
	Email         string     `json:"email"`
	PasswordHash  string     `json:"-"`
	Role          UserRole   `json:"role"`
	Status        UserStatus `json:"status"`
	UID           int        `json:"uid"`
	GID           int        `json:"gid"`
	HomeDirectory string     `json:"home_directory"`
	Namespace     string     `json:"namespace"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	LastLoginAt   *time.Time `json:"last_login_at,omitempty"`
}

// Session represents a user session
type Session struct {
	ID              string    `json:"id"`
	UserID          string    `json:"user_id"`
	RefreshToken    string    `json:"-"`
	DeviceFingerprint string  `json:"device_fingerprint"`
	IPAddress       string    `json:"ip_address"`
	UserAgent       string    `json:"user_agent"`
	CreatedAt       time.Time `json:"created_at"`
	ExpiresAt       time.Time `json:"expires_at"`
	LastAccessedAt  time.Time `json:"last_accessed_at"`
}

// APIKey represents an API key for authentication
type APIKey struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	Name        string    `json:"name"`
	KeyHash     string    `json:"-"`
	Prefix      string    `json:"prefix"`
	Scopes      []string  `json:"scopes"`
	IPBindings  []string  `json:"ip_bindings,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	LastUsedAt  *time.Time `json:"last_used_at,omitempty"`
}

// PasswordReset represents a password reset request
type PasswordReset struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Token     string    `json:"-"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
	Used      bool      `json:"used"`
}

// UserCreateRequest represents a request to create a user
type UserCreateRequest struct {
	Username string   `json:"username" validate:"required,min=3,max=32"`
	Email    string   `json:"email" validate:"required,email"`
	Password string   `json:"password" validate:"required,min=8"`
	Role     UserRole `json:"role,omitempty" validate:"omitempty,oneof=admin reseller user"`
	TenantID string   `json:"tenant_id,omitempty"`
}

// UserUpdateRequest represents a request to update a user
type UserUpdateRequest struct {
	Email    *string     `json:"email,omitempty" validate:"omitempty,email"`
	Role     *UserRole   `json:"role,omitempty" validate:"omitempty,oneof=admin reseller user"`
	Status   *UserStatus `json:"status,omitempty"`
}

// LoginRequest represents a login request
type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// LoginResponse represents a login response
type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

// RefreshRequest represents a token refresh request
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}
