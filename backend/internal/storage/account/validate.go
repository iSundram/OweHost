// Package account provides filesystem-based account state management
package account

import (
	"errors"
	"regexp"
	"strings"
)

// Validation errors
var (
	ErrInvalidAccountID   = errors.New("invalid account ID: must be positive integer")
	ErrInvalidAccountName = errors.New("invalid account name: must be 3-32 lowercase alphanumeric characters starting with a letter")
	ErrInvalidEmail       = errors.New("invalid email address")
	ErrInvalidPlan        = errors.New("invalid plan: must be one of starter, standard, premium, enterprise")
	ErrInvalidState       = errors.New("invalid state: must be one of active, suspended, terminated, pending")
	ErrInvalidUID         = errors.New("invalid UID: must be >= 1000")
	ErrInvalidGID         = errors.New("invalid GID: must be >= 1000")
	ErrInvalidNode        = errors.New("invalid node identifier")
	ErrInvalidOwner       = errors.New("invalid owner identifier")
	ErrInvalidDiskLimit   = errors.New("invalid disk limit: must be at least 100 MB or -1 for unlimited")
	ErrInvalidCPULimit    = errors.New("invalid CPU limit: must be between 1 and 400 percent or -1 for unlimited")
	ErrInvalidRAMLimit    = errors.New("invalid RAM limit: must be at least 128 MB or -1 for unlimited")
	ErrInvalidDomainLimit = errors.New("invalid domain limit: must be at least 1 or -1 for unlimited")
)

// Regex patterns for validation
var (
	validAccountName = regexp.MustCompile(`^[a-z][a-z0-9_]{2,31}$`)
	validEmail       = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	validNodeID      = regexp.MustCompile(`^[a-z][a-z0-9-]{0,62}$`)
	validOwnerID     = regexp.MustCompile(`^(admin|reseller-[0-9]+|partner-[0-9]+)$`)
)

// Valid values
var (
	validPlans  = map[string]bool{"starter": true, "standard": true, "premium": true, "enterprise": true}
	validStates = map[string]bool{"active": true, "suspended": true, "terminated": true, "pending": true}
)

// ValidateIdentity validates account identity
func ValidateIdentity(identity *AccountIdentity) error {
	if identity == nil {
		return errors.New("identity is nil")
	}

	if identity.ID <= 0 {
		return ErrInvalidAccountID
	}

	if !validAccountName.MatchString(identity.Name) {
		return ErrInvalidAccountName
	}

	if identity.UID < 1000 {
		return ErrInvalidUID
	}

	if identity.GID < 1000 {
		return ErrInvalidGID
	}

	if !validPlans[strings.ToLower(identity.Plan)] {
		return ErrInvalidPlan
	}

	if !validStates[strings.ToLower(identity.State)] {
		return ErrInvalidState
	}

	// Node validation (optional but if present must be valid)
	if identity.Node != "" && !validNodeID.MatchString(identity.Node) {
		return ErrInvalidNode
	}

	// Owner validation
	if identity.Owner != "" && !validOwnerID.MatchString(identity.Owner) {
		return ErrInvalidOwner
	}

	return nil
}

// ValidateLimits validates resource limits
func ValidateLimits(limits *ResourceLimits) error {
	if limits == nil {
		return errors.New("limits is nil")
	}

	// -1 means unlimited, otherwise must meet minimums
	if limits.DiskMB != -1 && limits.DiskMB < 100 {
		return ErrInvalidDiskLimit
	}

	if limits.CPUPercent != -1 && (limits.CPUPercent < 1 || limits.CPUPercent > 400) {
		return ErrInvalidCPULimit
	}

	if limits.RAMMB != -1 && limits.RAMMB < 128 {
		return ErrInvalidRAMLimit
	}

	if limits.Domains != -1 && limits.Domains < 1 {
		return ErrInvalidDomainLimit
	}

	return nil
}

// ValidateStatus validates account status
func ValidateStatus(status *AccountStatus) error {
	if status == nil {
		return errors.New("status is nil")
	}

	// If suspended, should have a reason
	if status.Suspended && (status.Reason == nil || *status.Reason == "") {
		// This is a warning, not an error - we allow it but it's not ideal
	}

	return nil
}

// ValidateMetadata validates account metadata
func ValidateMetadata(metadata *AccountMetadata) error {
	if metadata == nil {
		return nil // Metadata is optional
	}

	if metadata.Email != "" && !validEmail.MatchString(metadata.Email) {
		return ErrInvalidEmail
	}

	return nil
}

// ValidateApplyConfig validates the complete apply configuration
func ValidateApplyConfig(config *ApplyConfig) error {
	if config == nil {
		return errors.New("config is nil")
	}

	if config.Identity != nil {
		if err := ValidateIdentity(config.Identity); err != nil {
			return err
		}
	}

	if config.Limits != nil {
		if err := ValidateLimits(config.Limits); err != nil {
			return err
		}
	}

	if config.Status != nil {
		if err := ValidateStatus(config.Status); err != nil {
			return err
		}
	}

	if config.Metadata != nil {
		if err := ValidateMetadata(config.Metadata); err != nil {
			return err
		}
	}

	return nil
}

// SanitizeAccountName sanitizes an account name for use
func SanitizeAccountName(name string) string {
	// Convert to lowercase
	name = strings.ToLower(name)

	// Replace invalid characters with underscore
	result := make([]byte, 0, len(name))
	for i, c := range name {
		if i == 0 {
			// First character must be a letter
			if c >= 'a' && c <= 'z' {
				result = append(result, byte(c))
			}
		} else {
			// Subsequent characters can be alphanumeric or underscore
			if (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '_' {
				result = append(result, byte(c))
			}
		}
	}

	// Ensure minimum length
	if len(result) < 3 {
		return ""
	}

	// Truncate to max length
	if len(result) > 32 {
		result = result[:32]
	}

	return string(result)
}

// IsValidPlan checks if a plan name is valid
func IsValidPlan(plan string) bool {
	return validPlans[strings.ToLower(plan)]
}

// IsValidState checks if a state is valid
func IsValidState(state string) bool {
	return validStates[strings.ToLower(state)]
}
