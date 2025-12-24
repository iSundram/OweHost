// Package user provides user management services for OweHost
package user

import (
	"errors"
	"sync"
	"time"

	"github.com/iSundram/OweHost/pkg/config"
	"github.com/iSundram/OweHost/pkg/models"
	"github.com/iSundram/OweHost/pkg/utils"
)

// Service provides user management functionality
type Service struct {
	config  *config.Config
	users   map[string]*models.User
	byEmail map[string]*models.User
	byName  map[string]*models.User
	nextUID int
	mu      sync.RWMutex
}

// NewService creates a new user service
func NewService(cfg *config.Config) *Service {
	s := &Service{
		config:  cfg,
		users:   make(map[string]*models.User),
		byEmail: make(map[string]*models.User),
		byName:  make(map[string]*models.User),
		nextUID: 1000,
	}
	
	// Seed default admin user if none exists
	s.seedDefaultAdmin()
	
	return s
}

// seedDefaultAdmin creates default admin user with credentials: admin / admin@123
func (s *Service) seedDefaultAdmin() {
	// Check if any users exist
	if len(s.users) > 0 {
		return
	}
	
	// Hash the default password: admin@123
	passwordHash, err := utils.HashPassword("admin@123")
	if err != nil {
		return
	}
	
	adminUser := &models.User{
		ID:            utils.GenerateID("usr"),
		TenantID:      "default",
		Username:      "admin",
		Email:         "admin@owehost.local",
		PasswordHash:  passwordHash,
		Role:          models.UserRoleAdmin,
		Status:        models.UserStatusActive,
		UID:           1000,
		GID:           1000,
		HomeDirectory: "/home/admin",
		Namespace:     "admin",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	
	s.users[adminUser.ID] = adminUser
	s.byEmail[adminUser.Email] = adminUser
	s.byName[adminUser.Username] = adminUser
	s.nextUID = 1001
}

// Create creates a new user
func (s *Service) Create(req *models.UserCreateRequest) (*models.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check for duplicate email
	if _, exists := s.byEmail[req.Email]; exists {
		return nil, errors.New("email already in use")
	}

	// Check for duplicate username
	if _, exists := s.byName[req.Username]; exists {
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

	user := &models.User{
		ID:            utils.GenerateID("usr"),
		TenantID:      req.TenantID,
		Username:      req.Username,
		Email:         req.Email,
		PasswordHash:  passwordHash,
		Role:          role,
		Status:        models.UserStatusActive,
		UID:           s.nextUID,
		GID:           s.nextUID,
		HomeDirectory: "/home/" + req.Username,
		Namespace:     "user-" + req.Username,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	s.nextUID++
	s.users[user.ID] = user
	s.byEmail[user.Email] = user
	s.byName[user.Username] = user

	return user, nil
}

// Get gets a user by ID
func (s *Service) Get(id string) (*models.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, exists := s.users[id]
	if !exists {
		return nil, errors.New("user not found")
	}
	return user, nil
}

// GetByEmail gets a user by email
func (s *Service) GetByEmail(email string) (*models.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, exists := s.byEmail[email]
	if !exists {
		return nil, errors.New("user not found")
	}
	return user, nil
}

// GetByUsername gets a user by username
func (s *Service) GetByUsername(username string) (*models.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, exists := s.byName[username]
	if !exists {
		return nil, errors.New("user not found")
	}
	return user, nil
}

// List lists all users
func (s *Service) List(tenantID *string) []*models.User {
	s.mu.RLock()
	defer s.mu.RUnlock()

	users := make([]*models.User, 0)
	for _, user := range s.users {
		if tenantID == nil || user.TenantID == *tenantID {
			users = append(users, user)
		}
	}
	return users
}

// Update updates a user
func (s *Service) Update(id string, req *models.UserUpdateRequest) (*models.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	user, exists := s.users[id]
	if !exists {
		return nil, errors.New("user not found")
	}

	if req.Email != nil && *req.Email != user.Email {
		if _, exists := s.byEmail[*req.Email]; exists {
			return nil, errors.New("email already in use")
		}
		delete(s.byEmail, user.Email)
		user.Email = *req.Email
		s.byEmail[user.Email] = user
	}

	if req.Role != nil {
		user.Role = *req.Role
	}

	if req.Status != nil {
		user.Status = *req.Status
	}

	user.UpdatedAt = time.Now()
	return user, nil
}

// Suspend suspends a user
func (s *Service) Suspend(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	user, exists := s.users[id]
	if !exists {
		return errors.New("user not found")
	}

	user.Status = models.UserStatusSuspended
	user.UpdatedAt = time.Now()
	return nil
}

// Terminate terminates a user
func (s *Service) Terminate(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	user, exists := s.users[id]
	if !exists {
		return errors.New("user not found")
	}

	user.Status = models.UserStatusTerminated
	user.UpdatedAt = time.Now()
	return nil
}

// Clone clones a user account
func (s *Service) Clone(sourceID string, newUsername, newEmail string) (*models.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	source, exists := s.users[sourceID]
	if !exists {
		return nil, errors.New("source user not found")
	}

	// Check for duplicates
	if _, exists := s.byEmail[newEmail]; exists {
		return nil, errors.New("email already in use")
	}
	if _, exists := s.byName[newUsername]; exists {
		return nil, errors.New("username already in use")
	}

	clone := &models.User{
		ID:            utils.GenerateID("usr"),
		TenantID:      source.TenantID,
		Username:      newUsername,
		Email:         newEmail,
		PasswordHash:  source.PasswordHash,
		Role:          source.Role,
		Status:        models.UserStatusActive,
		UID:           s.nextUID,
		GID:           s.nextUID,
		HomeDirectory: "/home/" + newUsername,
		Namespace:     "user-" + newUsername,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	s.nextUID++
	s.users[clone.ID] = clone
	s.byEmail[clone.Email] = clone
	s.byName[clone.Username] = clone

	return clone, nil
}

// UpdatePassword updates a user's password
func (s *Service) UpdatePassword(id, newPassword string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	user, exists := s.users[id]
	if !exists {
		return errors.New("user not found")
	}

	if len(newPassword) < s.config.Auth.PasswordMinLength {
		return errors.New("password too short")
	}

	hash, err := utils.HashPassword(newPassword)
	if err != nil {
		return err
	}

	user.PasswordHash = hash
	user.UpdatedAt = time.Now()
	return nil
}

// ValidateCredentials validates user credentials
func (s *Service) ValidateCredentials(username, password string) (*models.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, exists := s.byName[username]
	if !exists {
		return nil, errors.New("invalid credentials")
	}

	if user.Status != models.UserStatusActive {
		return nil, errors.New("account is not active")
	}

	if !utils.CheckPassword(password, user.PasswordHash) {
		return nil, errors.New("invalid credentials")
	}

	return user, nil
}

// Delete deletes a user
func (s *Service) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	user, exists := s.users[id]
	if !exists {
		return errors.New("user not found")
	}

	delete(s.users, id)
	delete(s.byEmail, user.Email)
	delete(s.byName, user.Username)
	return nil
}
