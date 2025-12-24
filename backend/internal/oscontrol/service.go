// Package oscontrol provides OS-level control for OweHost
package oscontrol

import (
	"errors"
	"sync"
	"time"

	"github.com/iSundram/OweHost/pkg/models"
	"github.com/iSundram/OweHost/pkg/utils"
)

// Service provides OS control functionality
type Service struct {
	baseConfig      *models.ImmutableBaseConfig
	updates         map[string]*models.OSUpdate
	protectionRules map[string]*models.WriteProtectionRule
	mu              sync.RWMutex
}

// NewService creates a new OS control service
func NewService() *Service {
	svc := &Service{
		updates:         make(map[string]*models.OSUpdate),
		protectionRules: make(map[string]*models.WriteProtectionRule),
	}
	svc.initBaseConfig()
	return svc
}

// GetServiceStatuses returns a summarized list of core daemon statuses.
// This is a lightweight, non-invasive approximation (no service restarts).
func (s *Service) GetServiceStatuses() []*models.ServiceStatus {
	// In a real implementation you'd query systemd or the process table.
	// Here we provide mocked-but-structured data so the UI can render.
	return []*models.ServiceStatus{
		{Name: "HTTP Server (Nginx)", Status: "running", Uptime: "3d 4h", Load: "low"},
		{Name: "SQL Server (PostgreSQL)", Status: "running", Uptime: "2d 21h", Load: "medium"},
		{Name: "DNS Server", Status: "running", Uptime: "7d 12h", Load: "low"},
		{Name: "Mail Server (Exim)", Status: "warning", Uptime: "1d 2h", Load: "medium"},
		{Name: "FTP Server", Status: "running", Uptime: "6d 3h", Load: "low"},
		{Name: "SSH Server", Status: "running", Uptime: "14d", Load: "low"},
	}
}

// initBaseConfig initializes the immutable base configuration
func (s *Service) initBaseConfig() {
	s.baseConfig = &models.ImmutableBaseConfig{
		ID:             utils.GenerateID("base"),
		Version:        "1.0.0",
		WriteProtected: true,
		ProtectedPaths: []string{
			"/usr",
			"/bin",
			"/sbin",
			"/lib",
			"/lib64",
			"/etc/passwd",
			"/etc/shadow",
			"/etc/group",
		},
		AllowedWritePaths: []string{
			"/home",
			"/var/log",
			"/var/lib",
			"/tmp",
			"/var/tmp",
		},
		Checksum:  "sha256:dummy",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// GetBaseConfig gets the immutable base configuration
func (s *Service) GetBaseConfig() *models.ImmutableBaseConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.baseConfig
}

// IsPathProtected checks if a path is write-protected
func (s *Service) IsPathProtected(path string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.baseConfig.WriteProtected {
		return false
	}

	for _, protectedPath := range s.baseConfig.ProtectedPaths {
		if len(path) >= len(protectedPath) && path[:len(protectedPath)] == protectedPath {
			return true
		}
	}

	return false
}

// IsPathWritable checks if a path is writable
func (s *Service) IsPathWritable(path string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Check if in allowed write paths
	for _, allowedPath := range s.baseConfig.AllowedWritePaths {
		if len(path) >= len(allowedPath) && path[:len(allowedPath)] == allowedPath {
			return true
		}
	}

	return !s.IsPathProtected(path)
}

// EnableWriteProtection enables write protection
func (s *Service) EnableWriteProtection() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.baseConfig.WriteProtected = true
	s.baseConfig.UpdatedAt = time.Now()
	return nil
}

// DisableWriteProtection disables write protection (for maintenance)
func (s *Service) DisableWriteProtection() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.baseConfig.WriteProtected = false
	s.baseConfig.UpdatedAt = time.Now()
	return nil
}

// AddProtectedPath adds a path to the protected list
func (s *Service) AddProtectedPath(path string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, p := range s.baseConfig.ProtectedPaths {
		if p == path {
			return nil // Already protected
		}
	}

	s.baseConfig.ProtectedPaths = append(s.baseConfig.ProtectedPaths, path)
	s.baseConfig.UpdatedAt = time.Now()
	return nil
}

// RemoveProtectedPath removes a path from the protected list
func (s *Service) RemoveProtectedPath(path string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, p := range s.baseConfig.ProtectedPaths {
		if p == path {
			s.baseConfig.ProtectedPaths = append(s.baseConfig.ProtectedPaths[:i], s.baseConfig.ProtectedPaths[i+1:]...)
			s.baseConfig.UpdatedAt = time.Now()
			return nil
		}
	}

	return errors.New("path not in protected list")
}

// CreateWriteProtectionRule creates a write protection rule
func (s *Service) CreateWriteProtectionRule(path string, recursive bool) (*models.WriteProtectionRule, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	rule := &models.WriteProtectionRule{
		ID:        utils.GenerateID("wpr"),
		Path:      path,
		Recursive: recursive,
		Enabled:   true,
		CreatedAt: time.Now(),
	}

	s.protectionRules[rule.ID] = rule
	return rule, nil
}

// GetWriteProtectionRule gets a write protection rule
func (s *Service) GetWriteProtectionRule(id string) (*models.WriteProtectionRule, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rule, exists := s.protectionRules[id]
	if !exists {
		return nil, errors.New("rule not found")
	}
	return rule, nil
}

// ListWriteProtectionRules lists all write protection rules
func (s *Service) ListWriteProtectionRules() []*models.WriteProtectionRule {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rules := make([]*models.WriteProtectionRule, 0, len(s.protectionRules))
	for _, rule := range s.protectionRules {
		rules = append(rules, rule)
	}
	return rules
}

// DeleteWriteProtectionRule deletes a write protection rule
func (s *Service) DeleteWriteProtectionRule(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.protectionRules[id]; !exists {
		return errors.New("rule not found")
	}

	delete(s.protectionRules, id)
	return nil
}

// CheckForUpdates checks for available OS updates
func (s *Service) CheckForUpdates() []*models.OSUpdate {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Simulate available update
	update := &models.OSUpdate{
		ID:              utils.GenerateID("upd"),
		Version:         "1.1.0",
		PreviousVersion: s.baseConfig.Version,
		Status:          models.OSUpdateStatusPending,
		DownloadURL:     "https://updates.owehost.com/v1.1.0.tar.gz",
		Checksum:        "sha256:dummy",
		Size:            104857600, // 100MB
		ReleaseNotes:    "Bug fixes and improvements",
		Atomic:          true,
		CreatedAt:       time.Now(),
	}

	s.updates[update.ID] = update
	return []*models.OSUpdate{update}
}

// GetUpdate gets an OS update by ID
func (s *Service) GetUpdate(id string) (*models.OSUpdate, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	update, exists := s.updates[id]
	if !exists {
		return nil, errors.New("update not found")
	}
	return update, nil
}

// ListUpdates lists all OS updates
func (s *Service) ListUpdates() []*models.OSUpdate {
	s.mu.RLock()
	defer s.mu.RUnlock()

	updates := make([]*models.OSUpdate, 0, len(s.updates))
	for _, update := range s.updates {
		updates = append(updates, update)
	}
	return updates
}

// ScheduleUpdate schedules an OS update
func (s *Service) ScheduleUpdate(req *models.OSUpdateScheduleRequest) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	update, exists := s.updates[req.UpdateID]
	if !exists {
		return errors.New("update not found")
	}

	update.ScheduledAt = req.ScheduledAt
	return nil
}

// ApplyUpdate applies an OS update
func (s *Service) ApplyUpdate(updateID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	update, exists := s.updates[updateID]
	if !exists {
		return errors.New("update not found")
	}

	now := time.Now()
	update.Status = models.OSUpdateStatusApplying
	update.StartedAt = &now

	// Simulate atomic update
	go func() {
		time.Sleep(100 * time.Millisecond)

		s.mu.Lock()
		defer s.mu.Unlock()

		update.Status = models.OSUpdateStatusCompleted
		completedAt := time.Now()
		update.CompletedAt = &completedAt

		s.baseConfig.Version = update.Version
		s.baseConfig.UpdatedAt = completedAt
	}()

	return nil
}

// RollbackUpdate rolls back an OS update
func (s *Service) RollbackUpdate(req *models.OSUpdateRollbackRequest) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	update, exists := s.updates[req.UpdateID]
	if !exists {
		return errors.New("update not found")
	}

	if update.Status != models.OSUpdateStatusCompleted {
		return errors.New("can only rollback completed updates")
	}

	update.Status = models.OSUpdateStatusRolledBack

	s.baseConfig.Version = update.PreviousVersion
	s.baseConfig.UpdatedAt = time.Now()

	return nil
}

// GetCurrentVersion gets the current OS version
func (s *Service) GetCurrentVersion() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.baseConfig.Version
}

// CreateSystemUser creates a Linux system user (mock implementation)
func (s *Service) CreateSystemUser(username string, uid, gid int) error {
	// In production, this would execute: useradd -u {uid} -g {gid} -m -s /bin/bash {username}
	return nil
}

// CreateHomeDirectory creates and initializes a user's home directory (mock implementation)
func (s *Service) CreateHomeDirectory(homePath string, uid, gid int) error {
	// In production, this would: mkdir -p {homePath} && chown {uid}:{gid} {homePath}
	return nil
}

// DeleteSystemUser deletes a Linux system user (mock implementation)
func (s *Service) DeleteSystemUser(username string) error {
	// In production, this would execute: userdel -r {username}
	return nil
}

// DeleteHomeDirectory deletes a user's home directory (mock implementation)
func (s *Service) DeleteHomeDirectory(homePath string) error {
	// In production, this would execute: rm -rf {homePath}
	return nil
}
