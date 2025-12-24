// Package errors provides structured error handling for OweHost
package errors

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"strings"
)

// Error codes for categorizing errors
const (
	// Client errors (4xx)
	CodeBadRequest          = "BAD_REQUEST"
	CodeUnauthorized        = "UNAUTHORIZED"
	CodeForbidden           = "FORBIDDEN"
	CodeNotFound            = "NOT_FOUND"
	CodeConflict            = "CONFLICT"
	CodeValidationFailed    = "VALIDATION_FAILED"
	CodeRateLimited         = "RATE_LIMITED"
	CodePayloadTooLarge     = "PAYLOAD_TOO_LARGE"
	CodeUnsupportedMedia    = "UNSUPPORTED_MEDIA_TYPE"

	// Server errors (5xx)
	CodeInternalError       = "INTERNAL_ERROR"
	CodeNotImplemented      = "NOT_IMPLEMENTED"
	CodeServiceUnavailable  = "SERVICE_UNAVAILABLE"
	CodeDatabaseError       = "DATABASE_ERROR"
	CodeExternalServiceError = "EXTERNAL_SERVICE_ERROR"

	// Domain-specific errors
	CodeUserNotFound        = "USER_NOT_FOUND"
	CodeUserExists          = "USER_ALREADY_EXISTS"
	CodeInvalidCredentials  = "INVALID_CREDENTIALS"
	CodeAccountSuspended    = "ACCOUNT_SUSPENDED"
	CodeAccountTerminated   = "ACCOUNT_TERMINATED"
	CodeDomainNotFound      = "DOMAIN_NOT_FOUND"
	CodeDomainExists        = "DOMAIN_ALREADY_EXISTS"
	CodeDatabaseNotFound    = "DATABASE_NOT_FOUND"
	CodeDatabaseExists      = "DATABASE_ALREADY_EXISTS"
	CodeQuotaExceeded       = "QUOTA_EXCEEDED"
	CodeInvalidToken        = "INVALID_TOKEN"
	CodeTokenExpired        = "TOKEN_EXPIRED"
	CodeTwoFactorRequired   = "TWO_FACTOR_REQUIRED"
	CodeTwoFactorFailed     = "TWO_FACTOR_FAILED"
	CodePermissionDenied    = "PERMISSION_DENIED"
	CodeResourceLocked      = "RESOURCE_LOCKED"
	CodeOperationFailed     = "OPERATION_FAILED"
)

// AppError represents an application-level error with context
type AppError struct {
	Code       string            `json:"code"`
	Message    string            `json:"message"`
	Details    string            `json:"details,omitempty"`
	HTTPStatus int               `json:"-"`
	Cause      error             `json:"-"`
	Stack      string            `json:"-"`
	Metadata   map[string]string `json:"metadata,omitempty"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the underlying error
func (e *AppError) Unwrap() error {
	return e.Cause
}

// WithDetails adds details to the error
func (e *AppError) WithDetails(details string) *AppError {
	e.Details = details
	return e
}

// WithCause adds a cause to the error
func (e *AppError) WithCause(cause error) *AppError {
	e.Cause = cause
	return e
}

// WithMetadata adds metadata to the error
func (e *AppError) WithMetadata(key, value string) *AppError {
	if e.Metadata == nil {
		e.Metadata = make(map[string]string)
	}
	e.Metadata[key] = value
	return e
}

// ToJSON converts the error to JSON
func (e *AppError) ToJSON() []byte {
	data, _ := json.Marshal(e)
	return data
}

// New creates a new AppError
func New(code string, message string, httpStatus int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
		Stack:      captureStack(2),
	}
}

// Wrap wraps an existing error with an AppError
func Wrap(err error, code string, message string, httpStatus int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
		Cause:      err,
		Stack:      captureStack(2),
	}
}

// captureStack captures the call stack
func captureStack(skip int) string {
	var pcs [32]uintptr
	n := runtime.Callers(skip+1, pcs[:])
	frames := runtime.CallersFrames(pcs[:n])

	var sb strings.Builder
	for {
		frame, more := frames.Next()
		// Skip runtime and testing frames
		if !strings.Contains(frame.File, "runtime/") && !strings.Contains(frame.File, "testing/") {
			sb.WriteString(fmt.Sprintf("%s\n\t%s:%d\n", frame.Function, frame.File, frame.Line))
		}
		if !more {
			break
		}
	}
	return sb.String()
}

// Common error constructors

// BadRequest creates a 400 Bad Request error
func BadRequest(message string) *AppError {
	return New(CodeBadRequest, message, http.StatusBadRequest)
}

// Unauthorized creates a 401 Unauthorized error
func Unauthorized(message string) *AppError {
	return New(CodeUnauthorized, message, http.StatusUnauthorized)
}

// Forbidden creates a 403 Forbidden error
func Forbidden(message string) *AppError {
	return New(CodeForbidden, message, http.StatusForbidden)
}

// NotFound creates a 404 Not Found error
func NotFound(resource string) *AppError {
	return New(CodeNotFound, fmt.Sprintf("%s not found", resource), http.StatusNotFound)
}

// Conflict creates a 409 Conflict error
func Conflict(message string) *AppError {
	return New(CodeConflict, message, http.StatusConflict)
}

// ValidationFailed creates a validation error
func ValidationFailed(message string) *AppError {
	return New(CodeValidationFailed, message, http.StatusBadRequest)
}

// RateLimited creates a 429 Too Many Requests error
func RateLimited(message string) *AppError {
	return New(CodeRateLimited, message, http.StatusTooManyRequests)
}

// InternalError creates a 500 Internal Server Error
func InternalError(message string) *AppError {
	return New(CodeInternalError, message, http.StatusInternalServerError)
}

// DatabaseError creates a database error
func DatabaseError(err error) *AppError {
	return Wrap(err, CodeDatabaseError, "database operation failed", http.StatusInternalServerError)
}

// ExternalServiceError creates an external service error
func ExternalServiceError(service string, err error) *AppError {
	return Wrap(err, CodeExternalServiceError, 
		fmt.Sprintf("external service error: %s", service), 
		http.StatusBadGateway)
}

// User-related errors

// UserNotFound creates a user not found error
func UserNotFound(id string) *AppError {
	return New(CodeUserNotFound, "user not found", http.StatusNotFound).
		WithMetadata("user_id", id)
}

// UserExists creates a user already exists error
func UserExists(identifier string) *AppError {
	return New(CodeUserExists, "user already exists", http.StatusConflict).
		WithMetadata("identifier", identifier)
}

// InvalidCredentials creates an invalid credentials error
func InvalidCredentials() *AppError {
	return New(CodeInvalidCredentials, "invalid email or password", http.StatusUnauthorized)
}

// AccountSuspended creates an account suspended error
func AccountSuspended(reason string) *AppError {
	return New(CodeAccountSuspended, "account is suspended", http.StatusForbidden).
		WithDetails(reason)
}

// AccountTerminated creates an account terminated error
func AccountTerminated() *AppError {
	return New(CodeAccountTerminated, "account has been terminated", http.StatusForbidden)
}

// Domain-related errors

// DomainNotFound creates a domain not found error
func DomainNotFound(id string) *AppError {
	return New(CodeDomainNotFound, "domain not found", http.StatusNotFound).
		WithMetadata("domain_id", id)
}

// DomainExists creates a domain already exists error
func DomainExists(name string) *AppError {
	return New(CodeDomainExists, "domain already exists", http.StatusConflict).
		WithMetadata("domain", name)
}

// Auth-related errors

// InvalidToken creates an invalid token error
func InvalidToken() *AppError {
	return New(CodeInvalidToken, "invalid or malformed token", http.StatusUnauthorized)
}

// TokenExpired creates a token expired error
func TokenExpired() *AppError {
	return New(CodeTokenExpired, "token has expired", http.StatusUnauthorized)
}

// TwoFactorRequired creates a 2FA required error
func TwoFactorRequired() *AppError {
	return New(CodeTwoFactorRequired, "two-factor authentication required", http.StatusUnauthorized)
}

// TwoFactorFailed creates a 2FA failed error
func TwoFactorFailed() *AppError {
	return New(CodeTwoFactorFailed, "two-factor authentication failed", http.StatusUnauthorized)
}

// Resource errors

// QuotaExceeded creates a quota exceeded error
func QuotaExceeded(resource string, limit int64) *AppError {
	return New(CodeQuotaExceeded, 
		fmt.Sprintf("%s quota exceeded (limit: %d)", resource, limit), 
		http.StatusForbidden)
}

// PermissionDenied creates a permission denied error
func PermissionDenied(action string) *AppError {
	return New(CodePermissionDenied, 
		fmt.Sprintf("permission denied: %s", action), 
		http.StatusForbidden)
}

// ResourceLocked creates a resource locked error
func ResourceLocked(resource string) *AppError {
	return New(CodeResourceLocked, 
		fmt.Sprintf("%s is currently locked", resource), 
		http.StatusConflict)
}

// OperationFailed creates a generic operation failed error
func OperationFailed(operation string, err error) *AppError {
	return Wrap(err, CodeOperationFailed, 
		fmt.Sprintf("operation failed: %s", operation), 
		http.StatusInternalServerError)
}

// IsAppError checks if an error is an AppError
func IsAppError(err error) bool {
	_, ok := err.(*AppError)
	return ok
}

// AsAppError attempts to convert an error to an AppError
func AsAppError(err error) (*AppError, bool) {
	appErr, ok := err.(*AppError)
	return appErr, ok
}

// GetHTTPStatus returns the HTTP status code for an error
func GetHTTPStatus(err error) int {
	if appErr, ok := AsAppError(err); ok {
		return appErr.HTTPStatus
	}
	return http.StatusInternalServerError
}

// ErrorResponse represents the API error response format
type ErrorResponse struct {
	Success bool      `json:"success"`
	Error   *AppError `json:"error"`
}

// NewErrorResponse creates an error response
func NewErrorResponse(err *AppError) *ErrorResponse {
	return &ErrorResponse{
		Success: false,
		Error:   err,
	}
}

// WriteError writes an error response to the HTTP response writer
func WriteError(w http.ResponseWriter, err *AppError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(err.HTTPStatus)
	json.NewEncoder(w).Encode(NewErrorResponse(err))
}

// WriteErrorFromError writes an error response from a standard error
func WriteErrorFromError(w http.ResponseWriter, err error) {
	if appErr, ok := AsAppError(err); ok {
		WriteError(w, appErr)
		return
	}
	WriteError(w, InternalError(err.Error()))
}
