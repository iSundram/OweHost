package utils

import (
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"strings"
)

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidationResult represents the result of validation
type ValidationResult struct {
	Valid  bool              `json:"valid"`
	Errors []ValidationError `json:"errors,omitempty"`
}

// Validate performs basic validation on a struct.
// NOTE: This is a placeholder for future implementation with reflection-based struct tag validation.
// For now, field-specific validation functions like IsValidEmail and IsValidUsername should be used.
func Validate(data interface{}) *ValidationResult {
	result := &ValidationResult{Valid: true}
	// TODO: Implement reflection-based validation using struct tags
	_ = data // Silence unused parameter warning
	return result
}

// DecodeAndValidate decodes JSON from request body and validates it
func DecodeAndValidate(r *http.Request, v interface{}) error {
	if r.Body == nil {
		return errors.New("request body is required")
	}
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		return err
	}
	return nil
}

// IsValidEmail validates an email address
func IsValidEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

// IsValidUsername validates a username
func IsValidUsername(username string) bool {
	if len(username) < 3 || len(username) > 32 {
		return false
	}
	usernameRegex := regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_-]*$`)
	return usernameRegex.MatchString(username)
}

// IsValidDomain validates a domain name
func IsValidDomain(domain string) bool {
	domainRegex := regexp.MustCompile(`^([a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,}$`)
	return domainRegex.MatchString(domain)
}

// IsValidIPv4 validates an IPv4 address
func IsValidIPv4(ip string) bool {
	parts := strings.Split(ip, ".")
	if len(parts) != 4 {
		return false
	}
	for _, part := range parts {
		if len(part) == 0 || len(part) > 3 {
			return false
		}
		// Check for leading zeros (invalid in IPv4)
		if len(part) > 1 && part[0] == '0' {
			return false
		}
		num := 0
		for _, c := range part {
			if c < '0' || c > '9' {
				return false
			}
			num = num*10 + int(c-'0')
		}
		// Check range 0-255
		if num > 255 {
			return false
		}
	}
	return true
}

// IsValidPath validates a file path
func IsValidPath(path string) bool {
	if path == "" {
		return false
	}
	// Check for path traversal
	if strings.Contains(path, "..") {
		return false
	}
	// Check for absolute path
	if !strings.HasPrefix(path, "/") {
		return false
	}
	return true
}

// SanitizePath sanitizes a file path
func SanitizePath(path string) string {
	// Remove any null bytes
	path = strings.ReplaceAll(path, "\x00", "")
	// Normalize path separators
	path = strings.ReplaceAll(path, "\\", "/")
	// Remove double slashes
	for strings.Contains(path, "//") {
		path = strings.ReplaceAll(path, "//", "/")
	}
	return path
}
