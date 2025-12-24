// Package ftp provides FTP account management for OweHost
package ftp

import (
	"errors"
	"sync"
	"time"

	"github.com/iSundram/OweHost/pkg/models"
	"github.com/iSundram/OweHost/pkg/utils"
)

// Service provides FTP management functionality
type Service struct {
	accounts   map[string]*models.FTPAccount
	byUser     map[string][]*models.FTPAccount
	byUsername map[string]*models.FTPAccount
	sessions   map[string]*models.FTPSession
	config     *models.FTPConfig
	mu         sync.RWMutex
}

// NewService creates a new FTP service
func NewService() *Service {
	return &Service{
		accounts:   make(map[string]*models.FTPAccount),
		byUser:     make(map[string][]*models.FTPAccount),
		byUsername: make(map[string]*models.FTPAccount),
		sessions:   make(map[string]*models.FTPSession),
		config: &models.FTPConfig{
			Port:           21,
			PassivePortMin: 30000,
			PassivePortMax: 30100,
			MaxConnections: 100,
			MaxPerIP:       10,
			IdleTimeout:    300,
			TLSRequired:    true,
		},
	}
}

// Create creates a new FTP account
func (s *Service) Create(userID string, req *models.FTPAccountCreateRequest) (*models.FTPAccount, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check for duplicate username
	if _, exists := s.byUsername[req.Username]; exists {
		return nil, errors.New("username already exists")
	}

	// Validate password
	if len(req.Password) < 8 {
		return nil, errors.New("password must be at least 8 characters")
	}

	// Hash password
	passwordHash, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	account := &models.FTPAccount{
		ID:            utils.GenerateID("ftp"),
		UserID:        userID,
		Username:      req.Username,
		PasswordHash:  passwordHash,
		HomeDirectory: req.HomeDirectory,
		QuotaMB:       req.QuotaMB,
		UsedMB:        0,
		Status:        "active",
		ReadOnly:      req.ReadOnly,
		IPWhitelist:   req.IPWhitelist,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	s.accounts[account.ID] = account
	s.byUser[userID] = append(s.byUser[userID], account)
	s.byUsername[account.Username] = account

	return account, nil
}

// Get gets an FTP account by ID
func (s *Service) Get(id string) (*models.FTPAccount, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	account, exists := s.accounts[id]
	if !exists {
		return nil, errors.New("FTP account not found")
	}
	return account, nil
}

// GetByUsername gets an FTP account by username
func (s *Service) GetByUsername(username string) (*models.FTPAccount, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	account, exists := s.byUsername[username]
	if !exists {
		return nil, errors.New("FTP account not found")
	}
	return account, nil
}

// ListByUser lists FTP accounts for a user
func (s *Service) ListByUser(userID string) []*models.FTPAccount {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.byUser[userID]
}

// ListAll lists all FTP accounts (admin)
func (s *Service) ListAll() []*models.FTPAccount {
	s.mu.RLock()
	defer s.mu.RUnlock()

	accounts := make([]*models.FTPAccount, 0, len(s.accounts))
	for _, account := range s.accounts {
		accounts = append(accounts, account)
	}
	return accounts
}

// Update updates an FTP account
func (s *Service) Update(id string, req *models.FTPAccountUpdateRequest) (*models.FTPAccount, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	account, exists := s.accounts[id]
	if !exists {
		return nil, errors.New("FTP account not found")
	}

	if req.Password != nil && *req.Password != "" {
		if len(*req.Password) < 8 {
			return nil, errors.New("password must be at least 8 characters")
		}
		hash, err := utils.HashPassword(*req.Password)
		if err != nil {
			return nil, err
		}
		account.PasswordHash = hash
	}

	if req.HomeDirectory != nil {
		account.HomeDirectory = *req.HomeDirectory
	}
	if req.QuotaMB != nil {
		account.QuotaMB = *req.QuotaMB
	}
	if req.Status != nil {
		account.Status = *req.Status
	}
	if req.ReadOnly != nil {
		account.ReadOnly = *req.ReadOnly
	}
	if req.IPWhitelist != nil {
		account.IPWhitelist = *req.IPWhitelist
	}

	account.UpdatedAt = time.Now()
	return account, nil
}

// Delete deletes an FTP account
func (s *Service) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	account, exists := s.accounts[id]
	if !exists {
		return errors.New("FTP account not found")
	}

	// Remove from user's accounts
	userAccounts := s.byUser[account.UserID]
	for i, a := range userAccounts {
		if a.ID == id {
			s.byUser[account.UserID] = append(userAccounts[:i], userAccounts[i+1:]...)
			break
		}
	}

	delete(s.byUsername, account.Username)
	delete(s.accounts, id)
	return nil
}

// Suspend suspends an FTP account
func (s *Service) Suspend(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	account, exists := s.accounts[id]
	if !exists {
		return errors.New("FTP account not found")
	}

	account.Status = "suspended"
	account.UpdatedAt = time.Now()
	return nil
}

// Unsuspend unsuspends an FTP account
func (s *Service) Unsuspend(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	account, exists := s.accounts[id]
	if !exists {
		return errors.New("FTP account not found")
	}

	account.Status = "active"
	account.UpdatedAt = time.Now()
	return nil
}

// ValidateCredentials validates FTP login credentials
func (s *Service) ValidateCredentials(username, password, remoteIP string) (*models.FTPAccount, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	account, exists := s.byUsername[username]
	if !exists {
		return nil, errors.New("invalid credentials")
	}

	if account.Status != "active" {
		return nil, errors.New("account is not active")
	}

	// Check IP whitelist
	if len(account.IPWhitelist) > 0 {
		allowed := false
		for _, ip := range account.IPWhitelist {
			if ip == remoteIP {
				allowed = true
				break
			}
		}
		if !allowed {
			return nil, errors.New("IP not allowed")
		}
	}

	if !utils.CheckPassword(password, account.PasswordHash) {
		return nil, errors.New("invalid credentials")
	}

	return account, nil
}

// RecordLogin records a successful login
func (s *Service) RecordLogin(id, remoteIP string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	account, exists := s.accounts[id]
	if !exists {
		return errors.New("FTP account not found")
	}

	now := time.Now()
	account.LastLoginAt = &now
	account.LastLoginIP = remoteIP
	return nil
}

// GetActiveSessions returns active FTP sessions
func (s *Service) GetActiveSessions() []*models.FTPSession {
	s.mu.RLock()
	defer s.mu.RUnlock()

	sessions := make([]*models.FTPSession, 0, len(s.sessions))
	for _, session := range s.sessions {
		sessions = append(sessions, session)
	}
	return sessions
}

// GetConfig returns FTP server configuration
func (s *Service) GetConfig() *models.FTPConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.config
}

// UpdateConfig updates FTP server configuration
func (s *Service) UpdateConfig(cfg *models.FTPConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.config = cfg
	return nil
}

// DeleteAllByUser deletes all FTP accounts for a user
func (s *Service) DeleteAllByUser(userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, account := range s.byUser[userID] {
		delete(s.byUsername, account.Username)
		delete(s.accounts, account.ID)
	}
	delete(s.byUser, userID)
	return nil
}
