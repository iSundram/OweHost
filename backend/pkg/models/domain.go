package models

import "time"

// DomainStatus represents the status of a domain
type DomainStatus string

const (
	DomainStatusActive    DomainStatus = "active"
	DomainStatusPending   DomainStatus = "pending"
	DomainStatusSuspended DomainStatus = "suspended"
)

// DomainType represents the type of domain
type DomainType string

const (
	DomainTypePrimary DomainType = "primary"
	DomainTypeAddon   DomainType = "addon"
	DomainTypeParked  DomainType = "parked"
	DomainTypeAlias   DomainType = "alias"
)

// Domain represents a domain in the system
type Domain struct {
	ID            string       `json:"id"`
	UserID        string       `json:"user_id"`
	Name          string       `json:"name"`
	Type          DomainType   `json:"type"`
	Status        DomainStatus `json:"status"`
	DocumentRoot  string       `json:"document_root"`
	TargetDomain  *string      `json:"target_domain,omitempty"`
	Validated     bool         `json:"validated"`
	ValidationKey string       `json:"-"`
	CreatedAt     time.Time    `json:"created_at"`
	UpdatedAt     time.Time    `json:"updated_at"`
}

// Subdomain represents a subdomain
type Subdomain struct {
	ID           string    `json:"id"`
	DomainID     string    `json:"domain_id"`
	Name         string    `json:"name"`
	FullName     string    `json:"full_name"`
	DocumentRoot string    `json:"document_root"`
	PathMapping  *string   `json:"path_mapping,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// DomainCreateRequest represents a request to create a domain
type DomainCreateRequest struct {
	Name         string     `json:"name" validate:"required,fqdn"`
	Type         DomainType `json:"type" validate:"required,oneof=primary addon parked alias"`
	DocumentRoot string     `json:"document_root,omitempty"`
	TargetDomain *string    `json:"target_domain,omitempty"`
}

// SubdomainCreateRequest represents a request to create a subdomain
type SubdomainCreateRequest struct {
	Name         string  `json:"name" validate:"required,alphanum"`
	DocumentRoot string  `json:"document_root,omitempty"`
	PathMapping  *string `json:"path_mapping,omitempty"`
}

// RedirectType represents the type of redirect
type RedirectType string

const (
	RedirectTypePermanent RedirectType = "permanent" // 301
	RedirectTypeTemporary RedirectType = "temporary" // 302
)

// DomainRedirect represents a URL redirect configuration
type DomainRedirect struct {
	ID          string       `json:"id"`
	DomainID    string       `json:"domain_id"`
	SourcePath  string       `json:"source_path"`
	TargetURL   string       `json:"target_url"`
	Type        RedirectType `json:"type"`
	PreservePath bool        `json:"preserve_path"`
	Enabled     bool         `json:"enabled"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

// DomainRedirectCreateRequest represents a request to create a redirect
type DomainRedirectCreateRequest struct {
	SourcePath   string       `json:"source_path" validate:"required"`
	TargetURL    string       `json:"target_url" validate:"required,url"`
	Type         RedirectType `json:"type" validate:"required,oneof=permanent temporary"`
	PreservePath bool         `json:"preserve_path"`
}

// ErrorPageType represents the HTTP error code
type ErrorPageType string

const (
	ErrorPage400 ErrorPageType = "400"
	ErrorPage401 ErrorPageType = "401"
	ErrorPage403 ErrorPageType = "403"
	ErrorPage404 ErrorPageType = "404"
	ErrorPage500 ErrorPageType = "500"
	ErrorPage502 ErrorPageType = "502"
	ErrorPage503 ErrorPageType = "503"
)

// DomainErrorPage represents a custom error page configuration
type DomainErrorPage struct {
	ID        string        `json:"id"`
	DomainID  string        `json:"domain_id"`
	ErrorCode ErrorPageType `json:"error_code"`
	PagePath  string        `json:"page_path"`
	Content   string        `json:"content,omitempty"`
	Enabled   bool          `json:"enabled"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`
}

// DomainErrorPageCreateRequest represents a request to create an error page
type DomainErrorPageCreateRequest struct {
	ErrorCode ErrorPageType `json:"error_code" validate:"required,oneof=400 401 403 404 500 502 503"`
	PagePath  string        `json:"page_path"`
	Content   string        `json:"content"`
}

// DomainSettings represents domain-specific settings
type DomainSettings struct {
	DomainID       string `json:"domain_id"`
	ForceHTTPS     bool   `json:"force_https"`
	WWWRedirect    string `json:"www_redirect"` // "none", "add_www", "remove_www"
	HSTSEnabled    bool   `json:"hsts_enabled"`
	HSTSMaxAge     int    `json:"hsts_max_age"`
	IndexFiles     string `json:"index_files"` // comma-separated list
	DirectoryIndex bool   `json:"directory_index"`
}

// DomainSettingsUpdateRequest represents a request to update domain settings
type DomainSettingsUpdateRequest struct {
	ForceHTTPS     *bool   `json:"force_https"`
	WWWRedirect    *string `json:"www_redirect"`
	HSTSEnabled    *bool   `json:"hsts_enabled"`
	HSTSMaxAge     *int    `json:"hsts_max_age"`
	IndexFiles     *string `json:"index_files"`
	DirectoryIndex *bool   `json:"directory_index"`
}
