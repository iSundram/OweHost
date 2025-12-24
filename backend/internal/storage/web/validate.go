// Package web provides filesystem-based web/site state management
package web

import (
	"errors"
	"net"
	"regexp"
	"strings"
)

// Validation errors
var (
	ErrInvalidDomain       = errors.New("invalid domain name")
	ErrInvalidRuntime      = errors.New("invalid runtime")
	ErrInvalidDocumentRoot = errors.New("invalid document root: must not contain path traversal")
	ErrInvalidRedirectCode = errors.New("invalid redirect code: must be 301, 302, 307, or 308")
	ErrInvalidRedirectURL  = errors.New("invalid redirect target URL")
	ErrInvalidPHPVersion   = errors.New("invalid PHP version")
	ErrInvalidNodeVersion  = errors.New("invalid Node.js version")
	ErrInvalidPort         = errors.New("invalid port number: must be between 1024 and 65535")
)

// Domain validation regex
var validDomainRegex = regexp.MustCompile(`^([a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,}$`)
var validSubdomainRegex = regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?$`)

// Valid redirect codes
var validRedirectCodes = map[int]bool{
	301: true, // Permanent
	302: true, // Found
	307: true, // Temporary Redirect
	308: true, // Permanent Redirect
}

// ValidateSite validates a site descriptor
func ValidateSite(site *SiteDescriptor) error {
	if site == nil {
		return errors.New("site is nil")
	}

	// Validate domain
	if err := ValidateDomain(site.Domain); err != nil {
		return err
	}

	// Validate runtime
	if !IsValidRuntime(site.Runtime) {
		return ErrInvalidRuntime
	}

	// Validate document root (prevent path traversal)
	if err := ValidateDocumentRoot(site.DocumentRoot); err != nil {
		return err
	}

	// Validate aliases
	for _, alias := range site.Aliases {
		if err := ValidateDomain(alias); err != nil {
			return errors.New("invalid alias: " + alias)
		}
	}

	// Validate redirects
	for _, redirect := range site.Redirects {
		if err := ValidateRedirect(&redirect); err != nil {
			return err
		}
	}

	// Validate PHP settings if present
	if site.PHPSettings != nil {
		if err := ValidatePHPSettings(site.PHPSettings); err != nil {
			return err
		}
	}

	// Validate Node settings if present
	if site.NodeSettings != nil {
		if err := ValidateNodeSettings(site.NodeSettings); err != nil {
			return err
		}
	}

	return nil
}

// ValidateDomain validates a domain name
func ValidateDomain(domain string) error {
	if domain == "" {
		return ErrInvalidDomain
	}

	// Check length
	if len(domain) > 253 {
		return ErrInvalidDomain
	}

	// Check if it's a valid domain format
	if !validDomainRegex.MatchString(domain) {
		return ErrInvalidDomain
	}

	return nil
}

// ValidateSubdomain validates a subdomain name
func ValidateSubdomain(subdomain string) error {
	if subdomain == "" {
		return errors.New("subdomain cannot be empty")
	}

	if len(subdomain) > 63 {
		return errors.New("subdomain too long: max 63 characters")
	}

	if !validSubdomainRegex.MatchString(subdomain) {
		return errors.New("invalid subdomain format")
	}

	return nil
}

// ValidateDocumentRoot validates a document root path
func ValidateDocumentRoot(docRoot string) error {
	if docRoot == "" {
		return nil // Empty is valid, defaults to "public"
	}

	// Check for path traversal attempts
	if strings.Contains(docRoot, "..") {
		return ErrInvalidDocumentRoot
	}

	// Check for absolute paths
	if strings.HasPrefix(docRoot, "/") {
		return ErrInvalidDocumentRoot
	}

	// Check for invalid characters
	invalidChars := []string{"\\", ":", "*", "?", "\"", "<", ">", "|"}
	for _, char := range invalidChars {
		if strings.Contains(docRoot, char) {
			return ErrInvalidDocumentRoot
		}
	}

	return nil
}

// ValidateRedirect validates a redirect configuration
func ValidateRedirect(redirect *Redirect) error {
	if redirect == nil {
		return nil
	}

	// Validate source
	if redirect.Source == "" {
		return errors.New("redirect source cannot be empty")
	}

	// Validate target
	if redirect.Target == "" {
		return ErrInvalidRedirectURL
	}

	// Validate code
	if !validRedirectCodes[redirect.Code] {
		return ErrInvalidRedirectCode
	}

	return nil
}

// ValidatePHPSettings validates PHP settings
func ValidatePHPSettings(settings *PHPSettings) error {
	if settings == nil {
		return nil
	}

	// Validate version
	validPHPVersions := map[string]bool{
		"7.4": true, "8.0": true, "8.1": true, "8.2": true, "8.3": true,
	}
	if !validPHPVersions[settings.Version] {
		return ErrInvalidPHPVersion
	}

	// Validate execution time
	if settings.MaxExecutionTime < 0 || settings.MaxExecutionTime > 3600 {
		return errors.New("max_execution_time must be between 0 and 3600")
	}

	// Validate max input vars
	if settings.MaxInputVars < 0 || settings.MaxInputVars > 100000 {
		return errors.New("max_input_vars must be between 0 and 100000")
	}

	return nil
}

// ValidateNodeSettings validates Node.js settings
func ValidateNodeSettings(settings *NodeSettings) error {
	if settings == nil {
		return nil
	}

	// Validate version
	validNodeVersions := map[string]bool{
		"18": true, "20": true, "22": true,
	}
	if !validNodeVersions[settings.Version] {
		return ErrInvalidNodeVersion
	}

	// Validate port (must be unprivileged)
	if settings.Port < 1024 || settings.Port > 65535 {
		return ErrInvalidPort
	}

	// Validate max restarts
	if settings.MaxRestarts < 0 || settings.MaxRestarts > 100 {
		return errors.New("max_restarts must be between 0 and 100")
	}

	return nil
}

// ValidateSSLMeta validates SSL metadata
func ValidateSSLMeta(meta *SSLMeta) error {
	if meta == nil {
		return nil
	}

	// Validate domain
	if err := ValidateDomain(meta.Domain); err != nil {
		return err
	}

	// Validate type
	validTypes := map[string]bool{
		"letsencrypt": true, "custom": true, "self-signed": true,
	}
	if !validTypes[meta.Type] {
		return errors.New("invalid SSL type")
	}

	// Validate SANs
	for _, san := range meta.SANs {
		// SANs can be domains or IPs
		if ValidateDomain(san) != nil && net.ParseIP(san) == nil {
			return errors.New("invalid SAN: " + san)
		}
	}

	return nil
}

// SanitizeDomain sanitizes a domain name
func SanitizeDomain(domain string) string {
	// Convert to lowercase
	domain = strings.ToLower(domain)

	// Remove leading/trailing whitespace and dots
	domain = strings.Trim(domain, " .\t\n\r")

	// Remove any protocol prefix
	domain = strings.TrimPrefix(domain, "http://")
	domain = strings.TrimPrefix(domain, "https://")

	// Remove path and query string
	if idx := strings.Index(domain, "/"); idx != -1 {
		domain = domain[:idx]
	}
	if idx := strings.Index(domain, "?"); idx != -1 {
		domain = domain[:idx]
	}

	// Remove port
	if host, _, err := net.SplitHostPort(domain); err == nil {
		domain = host
	}

	return domain
}
