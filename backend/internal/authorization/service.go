// Package authorization provides RBAC and policy-based authorization for OweHost
package authorization

import (
	"errors"
	"sync"
	"time"

	"github.com/iSundram/OweHost/pkg/models"
	"github.com/iSundram/OweHost/pkg/utils"
)

// Service provides authorization functionality
type Service struct {
	roles           map[string]*models.Role
	permissions     map[string]*models.Permission
	roleAssignments map[string][]*models.RoleAssignment // userID -> assignments
	policyRules     map[string]*models.PolicyRule
	mu              sync.RWMutex
}

// NewService creates a new authorization service
func NewService() *Service {
	svc := &Service{
		roles:           make(map[string]*models.Role),
		permissions:     make(map[string]*models.Permission),
		roleAssignments: make(map[string][]*models.RoleAssignment),
		policyRules:     make(map[string]*models.PolicyRule),
	}
	svc.initDefaultRoles()
	return svc
}

// initDefaultRoles initializes default roles
func (s *Service) initDefaultRoles() {
	// Admin role
	adminRole := &models.Role{
		ID:          utils.GenerateID("role"),
		Name:        "admin",
		Description: "Full system administrator",
		Permissions: []string{"*"},
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	s.roles[adminRole.ID] = adminRole

	// Reseller role
	resellerRole := &models.Role{
		ID:          utils.GenerateID("role"),
		Name:        "reseller",
		Description: "Reseller with limited admin capabilities",
		Permissions: []string{
			"users:create", "users:read", "users:update",
			"domains:*", "databases:*", "files:*",
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	s.roles[resellerRole.ID] = resellerRole

	// User role
	userRole := &models.Role{
		ID:          utils.GenerateID("role"),
		Name:        "user",
		Description: "Standard user",
		Permissions: []string{
			"domains:read", "domains:create",
			"databases:read", "databases:create",
			"files:*",
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	s.roles[userRole.ID] = userRole
}

// CreateRole creates a new role
func (s *Service) CreateRole(req *models.RoleCreateRequest) (*models.Role, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check for duplicate name
	for _, r := range s.roles {
		if r.Name == req.Name {
			return nil, errors.New("role with this name already exists")
		}
	}

	role := &models.Role{
		ID:          utils.GenerateID("role"),
		Name:        req.Name,
		Description: req.Description,
		ParentID:    req.ParentID,
		Permissions: req.Permissions,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	s.roles[role.ID] = role
	return role, nil
}

// GetRole gets a role by ID
func (s *Service) GetRole(id string) (*models.Role, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	role, exists := s.roles[id]
	if !exists {
		return nil, errors.New("role not found")
	}
	return role, nil
}

// GetRoleByName gets a role by name
func (s *Service) GetRoleByName(name string) (*models.Role, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, role := range s.roles {
		if role.Name == name {
			return role, nil
		}
	}
	return nil, errors.New("role not found")
}

// ListRoles lists all roles
func (s *Service) ListRoles() []*models.Role {
	s.mu.RLock()
	defer s.mu.RUnlock()

	roles := make([]*models.Role, 0, len(s.roles))
	for _, role := range s.roles {
		roles = append(roles, role)
	}
	return roles
}

// AssignRole assigns a role to a user
func (s *Service) AssignRole(userID, roleID string, scope *string) (*models.RoleAssignment, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.roles[roleID]; !exists {
		return nil, errors.New("role not found")
	}

	assignment := &models.RoleAssignment{
		ID:        utils.GenerateID("ra"),
		UserID:    userID,
		RoleID:    roleID,
		Scope:     scope,
		CreatedAt: time.Now(),
	}

	s.roleAssignments[userID] = append(s.roleAssignments[userID], assignment)
	return assignment, nil
}

// RemoveRoleAssignment removes a role assignment
func (s *Service) RemoveRoleAssignment(userID, assignmentID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	assignments := s.roleAssignments[userID]
	for i, a := range assignments {
		if a.ID == assignmentID {
			s.roleAssignments[userID] = append(assignments[:i], assignments[i+1:]...)
			return nil
		}
	}
	return errors.New("assignment not found")
}

// GetUserRoles gets all roles for a user
func (s *Service) GetUserRoles(userID string) []*models.Role {
	s.mu.RLock()
	defer s.mu.RUnlock()

	assignments := s.roleAssignments[userID]
	roles := make([]*models.Role, 0, len(assignments))
	for _, a := range assignments {
		if role, exists := s.roles[a.RoleID]; exists {
			roles = append(roles, role)
		}
	}
	return roles
}

// GetUserPermissions gets all permissions for a user (including inherited)
func (s *Service) GetUserPermissions(userID string) []string {
	roles := s.GetUserRoles(userID)
	permSet := make(map[string]bool)

	for _, role := range roles {
		s.collectPermissions(role, permSet)
	}

	permissions := make([]string, 0, len(permSet))
	for perm := range permSet {
		permissions = append(permissions, perm)
	}
	return permissions
}

// collectPermissions collects permissions from a role and its parents
func (s *Service) collectPermissions(role *models.Role, permSet map[string]bool) {
	for _, perm := range role.Permissions {
		permSet[perm] = true
	}

	// Get parent role permissions
	if role.ParentID != nil {
		if parentRole, exists := s.roles[*role.ParentID]; exists {
			s.collectPermissions(parentRole, permSet)
		}
	}
}

// CheckPermission checks if a user has a specific permission
func (s *Service) CheckPermission(userID, resource, action string) bool {
	permissions := s.GetUserPermissions(userID)
	requiredPerm := resource + ":" + action

	for _, perm := range permissions {
		if perm == "*" || perm == requiredPerm || perm == resource+":*" {
			return true
		}
	}
	return false
}

// CreatePolicyRule creates a new policy rule
func (s *Service) CreatePolicyRule(req *models.PolicyRuleCreateRequest) (*models.PolicyRule, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	rule := &models.PolicyRule{
		ID:              utils.GenerateID("policy"),
		Name:            req.Name,
		Description:     req.Description,
		Resource:        req.Resource,
		Action:          req.Action,
		Conditions:      req.Conditions,
		Effect:          req.Effect,
		Priority:        req.Priority,
		TimeRestriction: req.TimeRestriction,
		Enabled:         true,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	s.policyRules[rule.ID] = rule
	return rule, nil
}

// EvaluatePolicy evaluates policy rules for an authorization request
func (s *Service) EvaluatePolicy(req *models.AuthorizationCheckRequest) *models.AuthorizationCheckResponse {
	// First check RBAC
	if s.CheckPermission(req.UserID, req.Resource, req.Action) {
		// Then check policy rules
		s.mu.RLock()
		defer s.mu.RUnlock()

		for _, rule := range s.policyRules {
			if !rule.Enabled {
				continue
			}
			if rule.Resource == req.Resource && rule.Action == req.Action {
				if s.evaluateConditions(rule, req.Context) {
					if rule.Effect == "deny" {
						return &models.AuthorizationCheckResponse{
							Allowed: false,
							Reason:  "Denied by policy: " + rule.Name,
						}
					}
				}
			}
		}

		return &models.AuthorizationCheckResponse{
			Allowed: true,
			Reason:  "Allowed by RBAC",
		}
	}

	return &models.AuthorizationCheckResponse{
		Allowed: false,
		Reason:  "Permission denied",
	}
}

// evaluateConditions evaluates policy conditions
func (s *Service) evaluateConditions(rule *models.PolicyRule, context map[string]interface{}) bool {
	// Check time restrictions
	if rule.TimeRestriction != nil {
		// Time-based evaluation would go here
		// For now, return true (no time restriction applied)
	}

	// Evaluate other conditions
	for key, expected := range rule.Conditions {
		if actual, exists := context[key]; exists {
			if actual != expected {
				return false
			}
		} else {
			return false
		}
	}

	return true
}

// CheckOwnership checks if a user owns a resource
func (s *Service) CheckOwnership(userID, resourceType, resourceID string, resourceOwnerID string) bool {
	return userID == resourceOwnerID
}
