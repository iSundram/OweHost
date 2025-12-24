// Package recovery provides failure handling and recovery services for OweHost
package recovery

import (
	"errors"
	"sync"
	"time"

	"github.com/iSundram/OweHost/pkg/models"
	"github.com/iSundram/OweHost/pkg/utils"
)

// Service provides recovery functionality
type Service struct {
	snapshots      map[string]*models.ConfigSnapshot
	rollbacks      map[string]*models.RollbackOperation
	healthChecks   map[string]*models.HealthCheck
	checkResults   map[string][]*models.HealthCheckResult
	healingTriggers map[string]*models.SelfHealingTrigger
	mu             sync.RWMutex
}

// NewService creates a new recovery service
func NewService() *Service {
	return &Service{
		snapshots:       make(map[string]*models.ConfigSnapshot),
		rollbacks:       make(map[string]*models.RollbackOperation),
		healthChecks:    make(map[string]*models.HealthCheck),
		checkResults:    make(map[string][]*models.HealthCheckResult),
		healingTriggers: make(map[string]*models.SelfHealingTrigger),
	}
}

// CreateSnapshot creates a configuration snapshot
func (s *Service) CreateSnapshot(resourceType, resourceID string, data map[string]interface{}) (*models.ConfigSnapshot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Get next version number
	version := 1
	for _, snap := range s.snapshots {
		if snap.ResourceType == resourceType && snap.ResourceID == resourceID {
			if snap.Version >= version {
				version = snap.Version + 1
			}
		}
	}

	snapshot := &models.ConfigSnapshot{
		ID:           utils.GenerateID("snap"),
		ResourceType: resourceType,
		ResourceID:   resourceID,
		Version:      version,
		Data:         data,
		Checksum:     "sha256:dummy",
		CreatedAt:    time.Now(),
	}

	s.snapshots[snapshot.ID] = snapshot
	return snapshot, nil
}

// GetSnapshot gets a snapshot by ID
func (s *Service) GetSnapshot(id string) (*models.ConfigSnapshot, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	snapshot, exists := s.snapshots[id]
	if !exists {
		return nil, errors.New("snapshot not found")
	}
	return snapshot, nil
}

// ListSnapshots lists snapshots for a resource
func (s *Service) ListSnapshots(resourceType, resourceID string) []*models.ConfigSnapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()

	snapshots := make([]*models.ConfigSnapshot, 0)
	for _, snap := range s.snapshots {
		if snap.ResourceType == resourceType && snap.ResourceID == resourceID {
			snapshots = append(snapshots, snap)
		}
	}
	return snapshots
}

// GetLatestSnapshot gets the latest snapshot for a resource
func (s *Service) GetLatestSnapshot(resourceType, resourceID string) (*models.ConfigSnapshot, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var latest *models.ConfigSnapshot
	for _, snap := range s.snapshots {
		if snap.ResourceType == resourceType && snap.ResourceID == resourceID {
			if latest == nil || snap.Version > latest.Version {
				latest = snap
			}
		}
	}

	if latest == nil {
		return nil, errors.New("no snapshots found")
	}
	return latest, nil
}

// DeleteSnapshot deletes a snapshot
func (s *Service) DeleteSnapshot(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.snapshots[id]; !exists {
		return errors.New("snapshot not found")
	}

	delete(s.snapshots, id)
	return nil
}

// InitiateRollback initiates a rollback operation
func (s *Service) InitiateRollback(userID string, req *models.RollbackRequest) (*models.RollbackOperation, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	snapshotID := req.SnapshotID
	if snapshotID == "" {
		// Find latest snapshot
		var latest *models.ConfigSnapshot
		for _, snap := range s.snapshots {
			if snap.ResourceType == req.ResourceType && snap.ResourceID == req.ResourceID {
				if latest == nil || snap.Version > latest.Version {
					latest = snap
				}
			}
		}
		if latest == nil {
			return nil, errors.New("no snapshots available for rollback")
		}
		snapshotID = latest.ID
	}

	rollback := &models.RollbackOperation{
		ID:           utils.GenerateID("rb"),
		Type:         req.Type,
		Status:       models.RollbackStatusPending,
		ResourceType: req.ResourceType,
		ResourceID:   req.ResourceID,
		SnapshotID:   snapshotID,
		InitiatedBy:  userID,
		CreatedAt:    time.Now(),
	}

	s.rollbacks[rollback.ID] = rollback

	// Start rollback process
	go s.performRollback(rollback.ID)

	return rollback, nil
}

// performRollback performs the actual rollback
func (s *Service) performRollback(rollbackID string) {
	s.mu.Lock()
	rollback := s.rollbacks[rollbackID]
	if rollback == nil {
		s.mu.Unlock()
		return
	}

	now := time.Now()
	rollback.Status = models.RollbackStatusRunning
	rollback.StartedAt = &now
	s.mu.Unlock()

	// Simulate rollback
	time.Sleep(100 * time.Millisecond)

	s.mu.Lock()
	defer s.mu.Unlock()

	rollback = s.rollbacks[rollbackID]
	if rollback == nil {
		return
	}

	completedAt := time.Now()
	rollback.Status = models.RollbackStatusCompleted
	rollback.CompletedAt = &completedAt
}

// GetRollback gets a rollback operation
func (s *Service) GetRollback(id string) (*models.RollbackOperation, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rollback, exists := s.rollbacks[id]
	if !exists {
		return nil, errors.New("rollback not found")
	}
	return rollback, nil
}

// ListRollbacks lists rollback operations
func (s *Service) ListRollbacks(resourceType, resourceID string) []*models.RollbackOperation {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rollbacks := make([]*models.RollbackOperation, 0)
	for _, rb := range s.rollbacks {
		if resourceType == "" || (rb.ResourceType == resourceType && rb.ResourceID == resourceID) {
			rollbacks = append(rollbacks, rb)
		}
	}
	return rollbacks
}

// CreateHealthCheck creates a health check
func (s *Service) CreateHealthCheck(req *models.HealthCheckCreateRequest) (*models.HealthCheck, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	check := &models.HealthCheck{
		ID:            utils.GenerateID("hc"),
		Name:          req.Name,
		Type:          req.Type,
		Target:        req.Target,
		Interval:      req.Interval,
		Timeout:       req.Timeout,
		Retries:       req.Retries,
		Status:        models.HealthCheckStatusUnknown,
		SelfHealing:   req.SelfHealing,
		HealingAction: req.HealingAction,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	s.healthChecks[check.ID] = check
	s.checkResults[check.ID] = make([]*models.HealthCheckResult, 0)

	return check, nil
}

// GetHealthCheck gets a health check
func (s *Service) GetHealthCheck(id string) (*models.HealthCheck, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	check, exists := s.healthChecks[id]
	if !exists {
		return nil, errors.New("health check not found")
	}
	return check, nil
}

// ListHealthChecks lists all health checks
func (s *Service) ListHealthChecks() []*models.HealthCheck {
	s.mu.RLock()
	defer s.mu.RUnlock()

	checks := make([]*models.HealthCheck, 0, len(s.healthChecks))
	for _, check := range s.healthChecks {
		checks = append(checks, check)
	}
	return checks
}

// UpdateHealthCheck updates a health check
func (s *Service) UpdateHealthCheck(id string, enabled bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, exists := s.healthChecks[id]
	if !exists {
		return errors.New("health check not found")
	}

	// Would update enabled status
	return nil
}

// DeleteHealthCheck deletes a health check
func (s *Service) DeleteHealthCheck(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.healthChecks[id]; !exists {
		return errors.New("health check not found")
	}

	delete(s.healthChecks, id)
	delete(s.checkResults, id)
	return nil
}

// RecordHealthCheckResult records a health check result
func (s *Service) RecordHealthCheckResult(checkID string, status models.HealthCheckStatus, responseTime int, details string) (*models.HealthCheckResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	check, exists := s.healthChecks[checkID]
	if !exists {
		return nil, errors.New("health check not found")
	}

	result := &models.HealthCheckResult{
		ID:            utils.GenerateID("hcr"),
		HealthCheckID: checkID,
		Status:        status,
		ResponseTime:  responseTime,
		Details:       details,
		CheckedAt:     time.Now(),
	}

	s.checkResults[checkID] = append(s.checkResults[checkID], result)

	// Keep only last 100 results
	if len(s.checkResults[checkID]) > 100 {
		s.checkResults[checkID] = s.checkResults[checkID][len(s.checkResults[checkID])-100:]
	}

	// Update health check status
	now := time.Now()
	check.Status = status
	check.LastCheckAt = &now
	if status == models.HealthCheckStatusHealthy {
		check.LastHealthyAt = &now
		check.ErrorMessage = nil
	} else {
		check.ErrorMessage = &details
	}
	check.UpdatedAt = now

	// Trigger self-healing if enabled
	if check.SelfHealing && status == models.HealthCheckStatusUnhealthy {
		s.triggerSelfHealing(check)
	}

	return result, nil
}

// GetHealthCheckResults gets results for a health check
func (s *Service) GetHealthCheckResults(checkID string, limit int) []*models.HealthCheckResult {
	s.mu.RLock()
	defer s.mu.RUnlock()

	results := s.checkResults[checkID]
	if limit <= 0 || limit > len(results) {
		limit = len(results)
	}

	start := len(results) - limit
	if start < 0 {
		start = 0
	}
	return results[start:]
}

// triggerSelfHealing triggers self-healing for a health check
func (s *Service) triggerSelfHealing(check *models.HealthCheck) {
	trigger := &models.SelfHealingTrigger{
		ID:            utils.GenerateID("heal"),
		HealthCheckID: check.ID,
		TriggeredAt:   time.Now(),
		Action:        check.HealingAction,
		Success:       true,
		Details:       "Self-healing action triggered",
	}

	s.healingTriggers[trigger.ID] = trigger
}

// ListSelfHealingTriggers lists self-healing triggers
func (s *Service) ListSelfHealingTriggers(checkID string) []*models.SelfHealingTrigger {
	s.mu.RLock()
	defer s.mu.RUnlock()

	triggers := make([]*models.SelfHealingTrigger, 0)
	for _, trigger := range s.healingTriggers {
		if checkID == "" || trigger.HealthCheckID == checkID {
			triggers = append(triggers, trigger)
		}
	}
	return triggers
}

// RunHealthChecks runs all pending health checks
func (s *Service) RunHealthChecks() []*models.HealthCheckResult {
	s.mu.RLock()
	checks := make([]*models.HealthCheck, 0)
	for _, check := range s.healthChecks {
		checks = append(checks, check)
	}
	s.mu.RUnlock()

	results := make([]*models.HealthCheckResult, 0)
	for _, check := range checks {
		// Simulate health check execution
		status := models.HealthCheckStatusHealthy
		responseTime := 50

		result, err := s.RecordHealthCheckResult(check.ID, status, responseTime, "Check passed")
		if err == nil {
			results = append(results, result)
		}
	}

	return results
}
