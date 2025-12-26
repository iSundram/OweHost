// Package user provides user management services for OweHost
package user

import (
	"context"
	"errors"
	"time"

	"github.com/iSundram/OweHost/pkg/config"
	"github.com/iSundram/OweHost/pkg/database"
	"github.com/iSundram/OweHost/pkg/models"
	"github.com/iSundram/OweHost/pkg/utils"
)

// Service provides user management functionality
type Service struct {
	config *config.Config
	repo   *Repository
}

// NewService creates a new user service
func NewService(cfg *config.Config, repo *Repository) *Service {
	s := &Service{
		config: cfg,
		repo:   repo,
	}

	// Seed default admin user if none exists
	s.seedDefaultAdmin()

	return s
}

// seedDefaultAdmin creates default admin user with credentials from config
func (s *Service) seedDefaultAdmin() {
	ctx := context.Background()

	// Skip if repo is nil (database not connected)
	if s.repo == nil {
		return
	}

	// Check if admin exists
	exists, _ := s.repo.ExistsByUsername(ctx, s.config.Admin.Username)
	if exists {
		return
	}

	// Check if any users exist at all
	users, _, _ := s.repo.List(ctx, database.Pagination{PerPage: 1})
	if len(users) > 0 {
		return
	}

	// Hash the default password from config
	passwordHash, err := utils.HashPassword(s.config.Admin.Password)
	if err != nil {
		return
	}

	adminUser := &models.User{
		ID:            utils.GenerateID("usr"),
		TenantID:      "default",
		Username:      s.config.Admin.Username,
		Email:         s.config.Admin.Email,
		PasswordHash:  passwordHash,
		Role:          models.UserRoleAdmin,
		Status:        models.UserStatusActive,
		UID:           1000,
		GID:           1000,
		HomeDirectory: "/home/" + s.config.Admin.Username,
		Namespace:     s.config.Admin.Username,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	s.repo.Create(ctx, adminUser)
}

// Create creates a new user
func (s *Service) Create(req *models.UserCreateRequest) (*models.User, error) {
	ctx := context.Background()

	// Check for duplicate email
	exists, _ := s.repo.ExistsByEmail(ctx, req.Email)
	if exists {
		return nil, errors.New("email already in use")
	}

	// Check for duplicate username
	exists, _ = s.repo.ExistsByUsername(ctx, req.Username)
	if exists {
		return nil, errors.New("username already in use")
	}

	// Validate password length
	if len(req.Password) < s.config.Auth.PasswordMinLength {
		return nil, errors.New("password too short")
	}

	// Hash password
	passwordHash, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	// Set role, default to user if not specified
	role := req.Role
	if role == "" {
		role = models.UserRoleUser
	}

	// Get next UID (simplified for now, ideally should query max UID from DB)
	// For now, we'll let the system manage UIDs or query it, but to keep it simple we'll use a placeholder
	// In a real implementation we'd need a sequence or max query
	nextUID := 1001 // Fallback

	user := &models.User{
		ID:            utils.GenerateID("usr"),
		TenantID:      req.TenantID,
		Username:      req.Username,
		Email:         req.Email,
		PasswordHash:  passwordHash,
		Role:          role,
		Status:        models.UserStatusActive,
		UID:           nextUID,
		GID:           nextUID,
		HomeDirectory: "/home/" + req.Username,
		Namespace:     "user-" + req.Username,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// Get gets a user by ID
func (s *Service) Get(id string) (*models.User, error) {
	return s.repo.GetByID(context.Background(), id)
}

// GetByEmail gets a user by email
func (s *Service) GetByEmail(email string) (*models.User, error) {
	return s.repo.GetByEmail(context.Background(), email)
}

// GetByUsername gets a user by username
func (s *Service) GetByUsername(username string) (*models.User, error) {
	return s.repo.GetByUsername(context.Background(), username)
}

// List lists all users
func (s *Service) List(tenantID *string) []*models.User {
	// Note: Pagination should be passed down, but preserving signature
	users, _, err := s.repo.List(context.Background(), database.Pagination{
		PerPage:  1000,
		OrderBy:  "created_at",
		OrderDir: "DESC",
	})
	if err != nil {
		return []*models.User{}
	}

	// Convert to pointers for compatibility
	result := make([]*models.User, len(users))
	for i := range users {
		// Filter by tenant if needed (though DB should handle this)
		if tenantID != nil && users[i].TenantID != *tenantID {
			continue
		}
		result[i] = &users[i]
	}
	return result
}

// Update updates a user
func (s *Service) Update(id string, req *models.UserUpdateRequest) (*models.User, error) {
	ctx := context.Background()
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Email != nil && *req.Email != user.Email {
		exists, _ := s.repo.ExistsByEmail(ctx, *req.Email)
		if exists {
			return nil, errors.New("email already in use")
		}
		user.Email = *req.Email
	}

	if req.Role != nil {
		user.Role = *req.Role
	}

	if req.Status != nil {
		user.Status = *req.Status
	}

	if err := s.repo.Update(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// Suspend suspends a user
func (s *Service) Suspend(id string) error {
	return s.repo.SuspendUser(context.Background(), id, "Suspended via API")
}

// Terminate terminates a user
func (s *Service) Terminate(id string) error {
	return s.repo.TerminateUser(context.Background(), id)
}

// Clone clones a user account
func (s *Service) Clone(sourceID string, newUsername, newEmail string) (*models.User, error) {
	ctx := context.Background()
	source, err := s.repo.GetByID(ctx, sourceID)
	if err != nil {
		return nil, errors.New("source user not found")
	}

	// Check for duplicates
	if exists, _ := s.repo.ExistsByEmail(ctx, newEmail); exists {
		return nil, errors.New("email already in use")
	}
	if exists, _ := s.repo.ExistsByUsername(ctx, newUsername); exists {
		return nil, errors.New("username already in use")
	}

	// Determine next UID (simplified)
	nextUID := 0 // Should query DB

	clone := &models.User{
		ID:            utils.GenerateID("usr"),
		TenantID:      source.TenantID,
		Username:      newUsername,
		Email:         newEmail,
		PasswordHash:  source.PasswordHash,
		Role:          source.Role,
		Status:        models.UserStatusActive,
		UID:           nextUID,
		GID:           nextUID,
		HomeDirectory: "/home/" + newUsername,
		Namespace:     "user-" + newUsername,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := s.repo.Create(ctx, clone); err != nil {
		return nil, err
	}

	return clone, nil
}

// UpdatePassword updates a user's password
func (s *Service) UpdatePassword(id, newPassword string) error {
	ctx := context.Background()
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if len(newPassword) < s.config.Auth.PasswordMinLength {
		return errors.New("password too short")
	}

	hash, err := utils.HashPassword(newPassword)
	if err != nil {
		return err
	}

	user.PasswordHash = hash
	return s.repo.Update(ctx, user)
}

// ValidateCredentials validates user credentials
func (s *Service) ValidateCredentials(username, password string) (*models.User, error) {
	ctx := context.Background()
	user, err := s.repo.GetByUsername(ctx, username)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	if user.Status != models.UserStatusActive {
		return nil, errors.New("account is not active")
	}

	if !utils.CheckPassword(password, user.PasswordHash) {
		s.repo.IncrementFailedLogin(ctx, user.ID)
		return nil, errors.New("invalid credentials")
	}

	s.repo.UpdateLastLogin(ctx, user.ID)
	return user, nil
}

// Delete deletes a user
func (s *Service) Delete(id string) error {
	return s.repo.Delete(context.Background(), id)
}
