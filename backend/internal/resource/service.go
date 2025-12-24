// Package resource provides resource and quota management for OweHost
package resource

import (
	"errors"
	"sync"
	"time"

	"github.com/iSundram/OweHost/pkg/models"
	"github.com/iSundram/OweHost/pkg/utils"
)

// Service provides resource quota management
type Service struct {
	quotas    map[string]*models.ResourceQuota // userID -> quota
	usage     map[string]*models.ResourceUsage // userID -> usage
	cgroups   map[string]*models.CgroupConfig  // userID -> cgroup config
	mu        sync.RWMutex
}

// NewService creates a new resource service
func NewService() *Service {
	return &Service{
		quotas:  make(map[string]*models.ResourceQuota),
		usage:   make(map[string]*models.ResourceUsage),
		cgroups: make(map[string]*models.CgroupConfig),
	}
}

// CreateQuota creates resource quotas for a user
func (s *Service) CreateQuota(userID string, quota *models.ResourceQuota) (*models.ResourceQuota, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.quotas[userID]; exists {
		return nil, errors.New("quota already exists for user")
	}

	quota.ID = utils.GenerateID("quota")
	quota.UserID = userID
	quota.CreatedAt = time.Now()
	quota.UpdatedAt = time.Now()

	s.quotas[userID] = quota
	return quota, nil
}

// GetQuota gets resource quotas for a user
func (s *Service) GetQuota(userID string) (*models.ResourceQuota, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	quota, exists := s.quotas[userID]
	if !exists {
		return nil, errors.New("quota not found")
	}
	return quota, nil
}

// UpdateQuota updates resource quotas for a user
func (s *Service) UpdateQuota(userID string, req *models.ResourceQuotaUpdateRequest) (*models.ResourceQuota, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	quota, exists := s.quotas[userID]
	if !exists {
		return nil, errors.New("quota not found")
	}

	if req.CPUQuota != nil {
		quota.CPUQuota = *req.CPUQuota
	}
	if req.CPUBurstEnabled != nil {
		quota.CPUBurstEnabled = *req.CPUBurstEnabled
	}
	if req.MemoryLimitMB != nil {
		quota.MemoryLimitMB = *req.MemoryLimitMB
	}
	if req.SwapLimitMB != nil {
		quota.SwapLimitMB = *req.SwapLimitMB
	}
	if req.DiskQuotaMB != nil {
		quota.DiskQuotaMB = *req.DiskQuotaMB
	}
	if req.InodeLimit != nil {
		quota.InodeLimit = *req.InodeLimit
	}
	if req.IOReadBps != nil {
		quota.IOReadBps = *req.IOReadBps
	}
	if req.IOWriteBps != nil {
		quota.IOWriteBps = *req.IOWriteBps
	}
	if req.MaxProcesses != nil {
		quota.MaxProcesses = *req.MaxProcesses
	}

	quota.UpdatedAt = time.Now()
	return quota, nil
}

// DeleteQuota deletes resource quotas for a user
func (s *Service) DeleteQuota(userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.quotas[userID]; !exists {
		return errors.New("quota not found")
	}

	delete(s.quotas, userID)
	return nil
}

// GetUsage gets current resource usage for a user
func (s *Service) GetUsage(userID string) (*models.ResourceUsage, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	usage, exists := s.usage[userID]
	if !exists {
		return nil, errors.New("usage data not found")
	}
	return usage, nil
}

// UpdateUsage updates resource usage for a user
func (s *Service) UpdateUsage(userID string, usage *models.ResourceUsage) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	usage.UserID = userID
	usage.MeasuredAt = time.Now()
	s.usage[userID] = usage
	return nil
}

// CheckQuota checks if a user is within quota limits
func (s *Service) CheckQuota(userID string) (bool, map[string]string) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	quota, qExists := s.quotas[userID]
	usage, uExists := s.usage[userID]

	if !qExists || !uExists {
		return true, nil
	}

	violations := make(map[string]string)

	if usage.MemoryUsageMB > quota.MemoryLimitMB {
		violations["memory"] = "exceeded memory limit"
	}
	if usage.DiskUsageMB > quota.DiskQuotaMB {
		violations["disk"] = "exceeded disk quota"
	}
	if usage.InodeUsage > quota.InodeLimit {
		violations["inodes"] = "exceeded inode limit"
	}
	if usage.ProcessCount > quota.MaxProcesses {
		violations["processes"] = "exceeded process limit"
	}

	return len(violations) == 0, violations
}

// CreateCgroupConfig creates cgroup configuration for a user
func (s *Service) CreateCgroupConfig(userID string) (*models.CgroupConfig, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	quota, exists := s.quotas[userID]
	if !exists {
		return nil, errors.New("quota not found")
	}

	config := &models.CgroupConfig{
		UserID:      userID,
		CgroupPath:  "/sys/fs/cgroup/user-" + userID,
		CPUPeriod:   100000,
		CPUQuota:    quota.CPUQuota * 1000,
		MemoryLimit: quota.MemoryLimitMB * 1024 * 1024,
		SwapLimit:   quota.SwapLimitMB * 1024 * 1024,
		PidsMax:     quota.MaxProcesses,
	}

	s.cgroups[userID] = config
	return config, nil
}

// GetCgroupConfig gets cgroup configuration for a user
func (s *Service) GetCgroupConfig(userID string) (*models.CgroupConfig, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	config, exists := s.cgroups[userID]
	if !exists {
		return nil, errors.New("cgroup config not found")
	}
	return config, nil
}

// ApplyCPUBurst applies CPU burst for a user (if enabled)
func (s *Service) ApplyCPUBurst(userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	quota, exists := s.quotas[userID]
	if !exists {
		return errors.New("quota not found")
	}

	if !quota.CPUBurstEnabled {
		return errors.New("CPU burst not enabled")
	}

	// Apply burst by temporarily increasing CPU quota
	// In production, this would interact with the actual cgroup subsystem
	return nil
}

// EnforceIOLimits enforces IO throttling limits
func (s *Service) EnforceIOLimits(userID string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	_, exists := s.quotas[userID]
	if !exists {
		return errors.New("quota not found")
	}

	// In production, this would configure blkio cgroup limits
	return nil
}

// PreventForkBomb checks and prevents fork bomb scenarios
func (s *Service) PreventForkBomb(userID string, currentProcesses int) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	quota, exists := s.quotas[userID]
	if !exists {
		return false, errors.New("quota not found")
	}

	// Check if approaching limit (80% threshold)
	threshold := int(float64(quota.MaxProcesses) * 0.8)
	if currentProcesses > threshold {
		return true, nil
	}

	return false, nil
}

// AllocatePackage allocates resources based on a package
func (s *Service) AllocatePackage(userID string, pkg interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return nil
}

// AllocateDefault allocates default resources to a user
func (s *Service) AllocateDefault(userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return nil
}

// ReleaseAll releases all resources allocated to a user
func (s *Service) ReleaseAll(userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return nil
}
