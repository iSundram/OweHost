// Package middleware provides rate limiting middleware for OweHost API
package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/iSundram/OweHost/pkg/utils"
)

// RateLimiter provides rate limiting functionality
type RateLimiter struct {
	requests map[string]*rateLimitEntry
	mu       sync.RWMutex
	
	// Configuration
	requestsPerMinute int
	burstSize         int
	cleanupInterval   time.Duration
}

type rateLimitEntry struct {
	tokens     float64
	lastUpdate time.Time
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(requestsPerMinute, burstSize int) *RateLimiter {
	rl := &RateLimiter{
		requests:          make(map[string]*rateLimitEntry),
		requestsPerMinute: requestsPerMinute,
		burstSize:         burstSize,
		cleanupInterval:   5 * time.Minute,
	}
	
	// Start cleanup goroutine
	go rl.cleanup()
	
	return rl
}

// Allow checks if a request should be allowed
func (rl *RateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	entry, exists := rl.requests[key]
	
	if !exists {
		rl.requests[key] = &rateLimitEntry{
			tokens:     float64(rl.burstSize - 1),
			lastUpdate: now,
		}
		return true
	}

	// Calculate tokens to add based on time passed
	elapsed := now.Sub(entry.lastUpdate).Seconds()
	tokensPerSecond := float64(rl.requestsPerMinute) / 60.0
	entry.tokens += elapsed * tokensPerSecond
	
	// Cap at burst size
	if entry.tokens > float64(rl.burstSize) {
		entry.tokens = float64(rl.burstSize)
	}
	
	entry.lastUpdate = now

	if entry.tokens >= 1 {
		entry.tokens--
		return true
	}

	return false
}

// cleanup removes old entries
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(rl.cleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		cutoff := time.Now().Add(-10 * time.Minute)
		for key, entry := range rl.requests {
			if entry.lastUpdate.Before(cutoff) {
				delete(rl.requests, key)
			}
		}
		rl.mu.Unlock()
	}
}

// TokenBucketRateLimitMiddleware creates rate limiting middleware using token bucket
func TokenBucketRateLimitMiddleware(rl *RateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get client identifier (IP + optional user ID)
			clientIP := getClientIP(r)
			key := clientIP
			
			// Add user ID if authenticated
			if userID := r.Context().Value(ContextKeyUserID); userID != nil {
				key = userID.(string)
			}

			if !rl.Allow(key) {
				w.Header().Set("Retry-After", "60")
				utils.WriteError(w, http.StatusTooManyRequests, "rate_limit_exceeded", "Too many requests, please try again later")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// UserRateLimiter provides per-user rate limiting with different tiers
type UserRateLimiter struct {
	limiters map[string]*RateLimiter
	mu       sync.RWMutex
	
	// Tier configurations
	defaultLimit int
	adminLimit   int
	apiKeyLimit  int
}

// NewUserRateLimiter creates a new per-user rate limiter
func NewUserRateLimiter() *UserRateLimiter {
	return &UserRateLimiter{
		limiters:     make(map[string]*RateLimiter),
		defaultLimit: 60,   // 60 requests per minute for regular users
		adminLimit:   300,  // 300 for admins
		apiKeyLimit:  120,  // 120 for API key access
	}
}

// GetLimiter gets or creates a rate limiter for a user
func (url *UserRateLimiter) GetLimiter(userID, role string, isAPIKey bool) *RateLimiter {
	url.mu.Lock()
	defer url.mu.Unlock()

	key := userID
	if rl, exists := url.limiters[key]; exists {
		return rl
	}

	// Determine limit based on role
	limit := url.defaultLimit
	if role == "admin" {
		limit = url.adminLimit
	} else if isAPIKey {
		limit = url.apiKeyLimit
	}

	rl := NewRateLimiter(limit, limit/2)
	url.limiters[key] = rl
	return rl
}

// UserRateLimitMiddleware creates per-user rate limiting middleware
func UserRateLimitMiddleware(url *UserRateLimiter) func(http.Handler) http.Handler {
	// Fallback limiter for unauthenticated requests
	fallbackLimiter := NewRateLimiter(30, 10)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var rl *RateLimiter
			var key string

			userID := r.Context().Value(ContextKeyUserID)
			userRole := r.Context().Value(ContextKeyUserRole)

			if userID != nil {
				role := ""
				if userRole != nil {
					role = userRole.(string)
				}
				rl = url.GetLimiter(userID.(string), role, false)
				key = userID.(string)
			} else {
				rl = fallbackLimiter
				key = getClientIP(r)
			}

			if !rl.Allow(key) {
				w.Header().Set("Retry-After", "60")
				utils.WriteError(w, http.StatusTooManyRequests, "rate_limit_exceeded", "Too many requests, please try again later")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// getClientIP extracts the client IP from the request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// Take the first IP
		for i := 0; i < len(xff); i++ {
			if xff[i] == ',' {
				return xff[:i]
			}
		}
		return xff
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	ip := r.RemoteAddr
	// Remove port if present
	for i := len(ip) - 1; i >= 0; i-- {
		if ip[i] == ':' {
			return ip[:i]
		}
	}
	return ip
}
