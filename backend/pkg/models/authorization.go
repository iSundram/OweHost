package models

import "time"

// Role represents a role in RBAC
type Role struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	ParentID    *string   `json:"parent_id,omitempty"`
	Permissions []string  `json:"permissions"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Permission represents a permission
type Permission struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Resource    string    `json:"resource"`
	Action      string    `json:"action"`
	CreatedAt   time.Time `json:"created_at"`
}

// RoleAssignment represents a role assignment to a user
type RoleAssignment struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	RoleID    string    `json:"role_id"`
	Scope     *string   `json:"scope,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// PolicyRule represents a policy rule
type PolicyRule struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Resource    string                 `json:"resource"`
	Action      string                 `json:"action"`
	Conditions  map[string]interface{} `json:"conditions"`
	Effect      string                 `json:"effect"`
	Priority    int                    `json:"priority"`
	TimeRestriction *TimeRestriction   `json:"time_restriction,omitempty"`
	Enabled     bool                   `json:"enabled"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// TimeRestriction represents time-based restrictions for a policy
type TimeRestriction struct {
	StartTime string   `json:"start_time"`
	EndTime   string   `json:"end_time"`
	Days      []string `json:"days"`
	Timezone  string   `json:"timezone"`
}

// RoleCreateRequest represents a request to create a role
type RoleCreateRequest struct {
	Name        string   `json:"name" validate:"required,min=2,max=64"`
	Description string   `json:"description,omitempty"`
	ParentID    *string  `json:"parent_id,omitempty"`
	Permissions []string `json:"permissions" validate:"required"`
}

// RoleUpdateRequest represents a request to update a role
type RoleUpdateRequest struct {
	Name        *string   `json:"name,omitempty" validate:"omitempty,min=2,max=64"`
	Description *string   `json:"description,omitempty"`
	Permissions *[]string `json:"permissions,omitempty"`
}

// PolicyRuleCreateRequest represents a request to create a policy rule
type PolicyRuleCreateRequest struct {
	Name            string                 `json:"name" validate:"required,min=2,max=64"`
	Description     string                 `json:"description,omitempty"`
	Resource        string                 `json:"resource" validate:"required"`
	Action          string                 `json:"action" validate:"required"`
	Conditions      map[string]interface{} `json:"conditions,omitempty"`
	Effect          string                 `json:"effect" validate:"required,oneof=allow deny"`
	Priority        int                    `json:"priority"`
	TimeRestriction *TimeRestriction       `json:"time_restriction,omitempty"`
}

// AuthorizationCheckRequest represents an authorization check request
type AuthorizationCheckRequest struct {
	UserID   string                 `json:"user_id" validate:"required"`
	Resource string                 `json:"resource" validate:"required"`
	Action   string                 `json:"action" validate:"required"`
	Context  map[string]interface{} `json:"context,omitempty"`
}

// AuthorizationCheckResponse represents an authorization check response
type AuthorizationCheckResponse struct {
	Allowed bool   `json:"allowed"`
	Reason  string `json:"reason,omitempty"`
}
