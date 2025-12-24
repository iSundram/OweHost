package models

import "time"

// CertificateStatus represents the status of a certificate
type CertificateStatus string

const (
	CertificateStatusPending   CertificateStatus = "pending"
	CertificateStatusActive    CertificateStatus = "active"
	CertificateStatusExpired   CertificateStatus = "expired"
	CertificateStatusRevoked   CertificateStatus = "revoked"
)

// CertificateType represents the type of certificate
type CertificateType string

const (
	CertificateTypeLetsEncrypt CertificateType = "letsencrypt"
	CertificateTypeCustom      CertificateType = "custom"
	CertificateTypeSelfSigned  CertificateType = "selfsigned"
)

// Certificate represents an SSL certificate
type Certificate struct {
	ID              string            `json:"id"`
	DomainID        string            `json:"domain_id"`
	Type            CertificateType   `json:"type"`
	Status          CertificateStatus `json:"status"`
	CommonName      string            `json:"common_name"`
	SANs            []string          `json:"sans"`
	Issuer          string            `json:"issuer"`
	SerialNumber    string            `json:"serial_number"`
	CertPath        string            `json:"cert_path"`
	KeyPath         string            `json:"key_path"`
	ChainPath       *string           `json:"chain_path,omitempty"`
	IssuedAt        time.Time         `json:"issued_at"`
	ExpiresAt       time.Time         `json:"expires_at"`
	AutoRenew       bool              `json:"auto_renew"`
	RenewalAttempts int               `json:"renewal_attempts"`
	LastRenewalAt   *time.Time        `json:"last_renewal_at,omitempty"`
	CreatedAt       time.Time         `json:"created_at"`
	UpdatedAt       time.Time         `json:"updated_at"`
}

// CSR represents a Certificate Signing Request
type CSR struct {
	ID           string    `json:"id"`
	DomainID     string    `json:"domain_id"`
	CommonName   string    `json:"common_name"`
	Organization string    `json:"organization"`
	Country      string    `json:"country"`
	State        string    `json:"state"`
	City         string    `json:"city"`
	CSRData      string    `json:"csr_data"`
	PrivateKey   string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
}

// RenewalSchedule represents a certificate renewal schedule
type RenewalSchedule struct {
	CertificateID   string     `json:"certificate_id"`
	ScheduledAt     time.Time  `json:"scheduled_at"`
	MaxRetries      int        `json:"max_retries"`
	CurrentRetries  int        `json:"current_retries"`
	LastAttemptAt   *time.Time `json:"last_attempt_at,omitempty"`
	ErrorMessage    *string    `json:"error_message,omitempty"`
}

// CSRCreateRequest represents a request to create a CSR
type CSRCreateRequest struct {
	DomainID     string   `json:"domain_id" validate:"required"`
	CommonName   string   `json:"common_name" validate:"required"`
	SANs         []string `json:"sans,omitempty"`
	Organization string   `json:"organization,omitempty"`
	Country      string   `json:"country,omitempty"`
	State        string   `json:"state,omitempty"`
	City         string   `json:"city,omitempty"`
}

// CertificateUploadRequest represents a request to upload a custom certificate
type CertificateUploadRequest struct {
	DomainID    string `json:"domain_id" validate:"required"`
	Certificate string `json:"certificate" validate:"required"`
	PrivateKey  string `json:"private_key" validate:"required"`
	Chain       string `json:"chain,omitempty"`
}

// LetsEncryptRequest represents a request to issue a Let's Encrypt certificate
type LetsEncryptRequest struct {
	DomainID      string   `json:"domain_id" validate:"required"`
	Domains       []string `json:"domains" validate:"required,min=1"`
	Email         string   `json:"email" validate:"required,email"`
	AutoRenew     bool     `json:"auto_renew"`
	Wildcard      bool     `json:"wildcard"`
	ChallengeType string   `json:"challenge_type"` // http-01, dns-01
}

// ACMEChallenge represents an ACME challenge for domain validation
type ACMEChallenge struct {
	ID           string    `json:"id"`
	DomainID     string    `json:"domain_id"`
	Domain       string    `json:"domain"`
	Type         string    `json:"type"` // http-01, dns-01
	Token        string    `json:"token"`
	KeyAuth      string    `json:"key_auth"`
	Status       string    `json:"status"` // pending, valid, invalid
	ValidatedAt  *time.Time `json:"validated_at,omitempty"`
	ExpiresAt    time.Time  `json:"expires_at"`
	CreatedAt    time.Time  `json:"created_at"`
}

// ACMEOrder represents an ACME order for certificate issuance
type ACMEOrder struct {
	ID            string          `json:"id"`
	DomainID      string          `json:"domain_id"`
	Domains       []string        `json:"domains"`
	Status        string          `json:"status"` // pending, ready, valid, invalid
	Challenges    []ACMEChallenge `json:"challenges"`
	CertificateID *string         `json:"certificate_id,omitempty"`
	ExpiresAt     time.Time       `json:"expires_at"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
}

// SSLSettings represents SSL/TLS settings for a domain
type SSLSettings struct {
	DomainID          string   `json:"domain_id"`
	ForceHTTPS        bool     `json:"force_https"`
	HSTSEnabled       bool     `json:"hsts_enabled"`
	HSTSMaxAge        int      `json:"hsts_max_age"`
	HSTSIncludeSubs   bool     `json:"hsts_include_subdomains"`
	HSTSPreload       bool     `json:"hsts_preload"`
	MinTLSVersion     string   `json:"min_tls_version"` // TLS1.2, TLS1.3
	CipherSuites      []string `json:"cipher_suites"`
	OCSPStapling      bool     `json:"ocsp_stapling"`
}

// SSLSettingsUpdateRequest represents a request to update SSL settings
type SSLSettingsUpdateRequest struct {
	ForceHTTPS        *bool     `json:"force_https"`
	HSTSEnabled       *bool     `json:"hsts_enabled"`
	HSTSMaxAge        *int      `json:"hsts_max_age"`
	HSTSIncludeSubs   *bool     `json:"hsts_include_subdomains"`
	HSTSPreload       *bool     `json:"hsts_preload"`
	MinTLSVersion     *string   `json:"min_tls_version"`
	CipherSuites      *[]string `json:"cipher_suites"`
	OCSPStapling      *bool     `json:"ocsp_stapling"`
}
