// Package models provides 2FA/MFA data models for OweHost
package models

import "time"

// TwoFactorType represents the type of 2FA
type TwoFactorType string

const (
	TwoFactorTypeTOTP     TwoFactorType = "totp"
	TwoFactorTypeWebAuthn TwoFactorType = "webauthn"
	TwoFactorTypeBackup   TwoFactorType = "backup"
)

// TwoFactorConfig represents 2FA configuration for a user
type TwoFactorConfig struct {
	ID              string        `json:"id"`
	UserID          string        `json:"user_id"`
	Enabled         bool          `json:"enabled"`
	Type            TwoFactorType `json:"type"`
	Secret          string        `json:"-"`                    // TOTP secret (encrypted)
	BackupCodes     []string      `json:"-"`                    // hashed backup codes
	BackupCodesUsed int           `json:"backup_codes_used"`
	VerifiedAt      *time.Time    `json:"verified_at,omitempty"`
	LastUsedAt      *time.Time    `json:"last_used_at,omitempty"`
	CreatedAt       time.Time     `json:"created_at"`
	UpdatedAt       time.Time     `json:"updated_at"`
}

// TOTPSetupResponse represents the response when setting up TOTP
type TOTPSetupResponse struct {
	Secret      string   `json:"secret"`       // base32 encoded secret
	QRCodeURL   string   `json:"qr_code_url"`  // otpauth:// URL for QR code
	BackupCodes []string `json:"backup_codes"` // one-time backup codes
}

// TOTPVerifyRequest represents a request to verify TOTP setup
type TOTPVerifyRequest struct {
	Code string `json:"code"`
}

// TwoFactorLoginRequest represents a 2FA login verification request
type TwoFactorLoginRequest struct {
	Token string `json:"token"`  // temporary token from initial login
	Code  string `json:"code"`   // TOTP code or backup code
}

// TwoFactorDisableRequest represents a request to disable 2FA
type TwoFactorDisableRequest struct {
	Password string `json:"password"`
	Code     string `json:"code"` // current TOTP code to confirm
}

// BackupCodesRegenerateRequest represents a request to regenerate backup codes
type BackupCodesRegenerateRequest struct {
	Password string `json:"password"`
	Code     string `json:"code"` // current TOTP code
}

// WebAuthnCredential represents a WebAuthn credential (security key)
type WebAuthnCredential struct {
	ID           string    `json:"id"`
	UserID       string    `json:"user_id"`
	Name         string    `json:"name"`
	CredentialID []byte    `json:"-"`
	PublicKey    []byte    `json:"-"`
	AAGUID       []byte    `json:"-"`
	SignCount    uint32    `json:"sign_count"`
	LastUsedAt   *time.Time `json:"last_used_at,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

// LoginAttempt represents a login attempt for security tracking
type LoginAttempt struct {
	ID           string    `json:"id"`
	UserID       string    `json:"user_id,omitempty"`
	Username     string    `json:"username"`
	IPAddress    string    `json:"ip_address"`
	UserAgent    string    `json:"user_agent"`
	Success      bool      `json:"success"`
	FailReason   string    `json:"fail_reason,omitempty"`
	TwoFactorUsed bool     `json:"two_factor_used"`
	CreatedAt    time.Time `json:"created_at"`
}

// SecuritySettings represents user security settings
type SecuritySettings struct {
	UserID              string    `json:"user_id"`
	TwoFactorEnabled    bool      `json:"two_factor_enabled"`
	TwoFactorType       string    `json:"two_factor_type,omitempty"`
	SessionTimeout      int       `json:"session_timeout"`       // minutes
	IPRestriction       bool      `json:"ip_restriction"`
	AllowedIPs          []string  `json:"allowed_ips,omitempty"`
	LoginNotifications  bool      `json:"login_notifications"`
	PasswordLastChanged time.Time `json:"password_last_changed"`
	RequirePasswordChange bool    `json:"require_password_change"`
}
