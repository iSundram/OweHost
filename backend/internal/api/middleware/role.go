// Package middleware provides HTTP middleware for OweHost API
package middleware

import (
	"context"
	"net/http"

	"github.com/iSundram/OweHost/internal/user"
	"github.com/iSundram/OweHost/pkg/models"
	"github.com/iSundram/OweHost/pkg/utils"
)

// RoleMiddleware provides role-based access control middleware
func RoleMiddleware(userService *user.Service, allowedRoles ...models.UserRole) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID := GetUserID(r.Context())
			if userID == "" {
				utils.WriteError(w, http.StatusUnauthorized, utils.ErrCodeUnauthorized, "Not authenticated")
				return
			}

			user, err := userService.Get(userID)
			if err != nil {
				utils.WriteError(w, http.StatusUnauthorized, utils.ErrCodeUnauthorized, "User not found")
				return
			}

			// Admin can access everything
			if user.Role == models.UserRoleAdmin {
				ctx := context.WithValue(r.Context(), ContextKeyUserRole, user.Role)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			// Check if user's role is in allowed roles
			allowed := false
			for _, role := range allowedRoles {
				if user.Role == role {
					allowed = true
					break
				}
			}

			if !allowed {
				utils.WriteError(w, http.StatusForbidden, utils.ErrCodeForbidden, "Insufficient permissions")
				return
			}

			ctx := context.WithValue(r.Context(), ContextKeyUserRole, user.Role)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// AdminOnlyMiddleware allows only admin users
func AdminOnlyMiddleware(userService *user.Service) func(http.Handler) http.Handler {
	return RoleMiddleware(userService, models.UserRoleAdmin)
}

// AdminOrResellerMiddleware allows admin and reseller users
func AdminOrResellerMiddleware(userService *user.Service) func(http.Handler) http.Handler {
	return RoleMiddleware(userService, models.UserRoleAdmin, models.UserRoleReseller)
}

// GetUserRole gets user role from context
func GetUserRole(ctx context.Context) models.UserRole {
	if role, ok := ctx.Value(ContextKeyUserRole).(models.UserRole); ok {
		return role
	}
	return ""
}
