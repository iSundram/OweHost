// Package middleware provides request validation middleware
package middleware

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/iSundram/OweHost/pkg/utils"
)

// ValidationRule represents a validation rule
type ValidationRule struct {
	Field     string
	Required  bool
	MinLength int
	MaxLength int
	Pattern   string
	Validator func(interface{}) bool
}

// ValidateRequest validates request body against rules
func ValidateRequest(rules []ValidationRule) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var body map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Invalid JSON")
				return
			}

			errors := make(map[string]string)

			for _, rule := range rules {
				value, exists := body[rule.Field]

				// Check required
				if rule.Required && !exists {
					errors[rule.Field] = "This field is required"
					continue
				}

				if !exists {
					continue
				}

				// Check type and length for strings
				if str, ok := value.(string); ok {
					str = strings.TrimSpace(str)

					if rule.Required && str == "" {
						errors[rule.Field] = "This field cannot be empty"
						continue
					}

					if rule.MinLength > 0 && len(str) < rule.MinLength {
						errors[rule.Field] = "Must be at least " + string(rune(rule.MinLength)) + " characters"
						continue
					}

					if rule.MaxLength > 0 && len(str) > rule.MaxLength {
						errors[rule.Field] = "Must be no more than " + string(rune(rule.MaxLength)) + " characters"
						continue
					}
				}

				// Custom validator
				if rule.Validator != nil && !rule.Validator(value) {
					errors[rule.Field] = "Invalid value"
				}
			}

			if len(errors) > 0 {
				utils.WriteValidationError(w, errors)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// ValidateDomainName validates domain name format
func ValidateDomainName(value interface{}) bool {
	str, ok := value.(string)
	if !ok {
		return false
	}
	return utils.IsValidDomain(str)
}

// ValidateEmail validates email format
func ValidateEmail(value interface{}) bool {
	str, ok := value.(string)
	if !ok {
		return false
	}
	return utils.IsValidEmail(str)
}
