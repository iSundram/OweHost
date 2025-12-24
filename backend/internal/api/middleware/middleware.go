// Package middleware provides HTTP middleware for OweHost API
package middleware

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/iSundram/OweHost/internal/auth"
	"github.com/iSundram/OweHost/internal/logging"
	"github.com/iSundram/OweHost/pkg/utils"
)

// ContextKey type for context keys
type ContextKey string

const (
	// ContextKeyUserID is the context key for user ID
	ContextKeyUserID ContextKey = "user_id"
	// ContextKeyTenantID is the context key for tenant ID
	ContextKeyTenantID ContextKey = "tenant_id"
	// ContextKeyRequestID is the context key for request ID
	ContextKeyRequestID ContextKey = "request_id"
	// ContextKeyUserRole is the context key for user role
	ContextKeyUserRole ContextKey = "user_role"
)

// AuthMiddleware provides authentication middleware
func AuthMiddleware(authService *auth.Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				utils.WriteError(w, http.StatusUnauthorized, utils.ErrCodeUnauthorized, "Authorization header required")
				return
			}

			var userID, tenantID string

			// Check for Bearer token
			if strings.HasPrefix(authHeader, "Bearer ") {
				token := strings.TrimPrefix(authHeader, "Bearer ")
				claims, err := authService.ValidateToken(token)
				if err != nil {
					utils.WriteError(w, http.StatusUnauthorized, utils.ErrCodeUnauthorized, "Invalid token")
					return
				}
				userID = claims.UserID
				tenantID = claims.TenantID
			} else if strings.HasPrefix(authHeader, "ApiKey ") {
				// Check for API key
				apiKey := strings.TrimPrefix(authHeader, "ApiKey ")
				key, err := authService.ValidateAPIKey(apiKey)
				if err != nil {
					utils.WriteError(w, http.StatusUnauthorized, utils.ErrCodeUnauthorized, "Invalid API key")
					return
				}
				userID = key.UserID
			} else {
				utils.WriteError(w, http.StatusUnauthorized, utils.ErrCodeUnauthorized, "Invalid authorization format")
				return
			}

			ctx := context.WithValue(r.Context(), ContextKeyUserID, userID)
			ctx = context.WithValue(ctx, ContextKeyTenantID, tenantID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequestIDMiddleware adds a request ID to each request
func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = utils.GenerateID("req")
		}

		ctx := context.WithValue(r.Context(), ContextKeyRequestID, requestID)
		w.Header().Set("X-Request-ID", requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// LoggingMiddleware logs requests
func LoggingMiddleware(logService *logging.Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Wrap response writer to capture status
			wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			next.ServeHTTP(wrapped, r)

			duration := time.Since(start)

			metadata := map[string]interface{}{
				"method":      r.Method,
				"path":        r.URL.Path,
				"status":      wrapped.statusCode,
				"duration_ms": duration.Milliseconds(),
				"ip":          r.RemoteAddr,
				"user_agent":  r.UserAgent(),
			}

			requestID, _ := r.Context().Value(ContextKeyRequestID).(string)
			var reqIDPtr *string
			if requestID != "" {
				reqIDPtr = &requestID
			}

			logService.Log("info", "api", "HTTP request", nil, reqIDPtr, nil, metadata)
		})
	}
}

// CORSMiddleware handles CORS
func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Request-ID")
		w.Header().Set("Access-Control-Max-Age", "86400")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// ContentTypeMiddleware sets JSON content type
func ContentTypeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

// RateLimitMiddleware provides rate limiting
func RateLimitMiddleware(requestsPerSecond int) func(http.Handler) http.Handler {
	// Simplified rate limiting - in production use a proper rate limiter
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Rate limiting logic would go here
			next.ServeHTTP(w, r)
		})
	}
}

// RecoveryMiddleware recovers from panics
func RecoveryMiddleware(logService *logging.Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					logService.Error("api", "Panic recovered: "+r.URL.Path)
					utils.WriteError(w, http.StatusInternalServerError, utils.ErrCodeInternalError, "Internal server error")
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// GetUserID gets user ID from context
func GetUserID(ctx context.Context) string {
	if userID, ok := ctx.Value(ContextKeyUserID).(string); ok {
		return userID
	}
	return ""
}

// GetTenantID gets tenant ID from context
func GetTenantID(ctx context.Context) string {
	if tenantID, ok := ctx.Value(ContextKeyTenantID).(string); ok {
		return tenantID
	}
	return ""
}

// GetRequestID gets request ID from context
func GetRequestID(ctx context.Context) string {
	if requestID, ok := ctx.Value(ContextKeyRequestID).(string); ok {
		return requestID
	}
	return ""
}
