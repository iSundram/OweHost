// Package backup provides backup and snapshot services for OweHost
package backup

import (
	"errors"
	"sync"
	"time"

	"github.com/iSundram/OweHost/pkg/models"
	"github.com/iSundram/OweHost/pkg/utils"
)

// Service provides backup functionality
type Service struct {
	backups    map[string]*models.Backup
	schedules  map[string]*models.BackupSchedule
	restores   map[string]*models.RestoreStatus
	byUser     map[string][]*models.Backup
	queue      []*BackupTask
	mu         sync.RWMutex
}

// BackupTask represents a backup task in the queue
type BackupTask struct {
	BackupID  string
	Priority  int
	CreatedAt time.Time
}

// NewService creates a new backup service
func NewService() *Service {
	return &Service{
		backups:   make(map[string]*models.Backup),
		schedules: make(map[string]*models.BackupSchedule),
		restores:  make(map[string]*models.RestoreStatus),
		byUser:    make(map[string][]*models.Backup),
		queue:     make([]*BackupTask, 0),
	}
}

// Create creates a new backup
func (s *Service) Create(userID string, req *models.BackupCreateRequest) (*models.Backup, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	backup := &models.Backup{
		ID:               utils.GenerateID("bkp"),
		UserID:           userID,
		Type:             req.Type,
		Status:           models.BackupStatusPending,
		SizeMB:           0,
		StoragePath:      "/backups/" + userID + "/" + time.Now().Format("20060102150405"),
		IncludeFiles:     req.IncludeFiles,
		IncludeDatabases: req.IncludeDatabases,
		CreatedAt:        time.Now(),
	}

	// For incremental/differential, find parent
	if req.Type != models.BackupTypeFull {
		latestFull := s.findLatestBackup(userID, models.BackupTypeFull)
		if latestFull == nil {
			return nil, errors.New("no full backup found for incremental/differential backup")
		}
		backup.ParentBackupID = &latestFull.ID
	}

	s.backups[backup.ID] = backup
	s.byUser[userID] = append(s.byUser[userID], backup)

	// Add to queue
	task := &BackupTask{
		BackupID:  backup.ID,
		Priority:  req.Priority,
		CreatedAt: time.Now(),
	}
	s.addToQueue(task)

	return backup, nil
}

// Get gets a backup by ID
func (s *Service) Get(id string) (*models.Backup, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	backup, exists := s.backups[id]
	if !exists {
		return nil, errors.New("backup not found")
	}
	return backup, nil
}

// ListByUser lists backups for a user
func (s *Service) ListByUser(userID string) []*models.Backup {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.byUser[userID]
}

// Delete deletes a backup
func (s *Service) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	backup, exists := s.backups[id]
	if !exists {
		return errors.New("backup not found")
	}

	// Check if any backup depends on this one
	for _, b := range s.backups {
		if b.ParentBackupID != nil && *b.ParentBackupID == id {
			return errors.New("cannot delete backup with dependent backups")
		}
	}

	// Remove from user's backups
	userBackups := s.byUser[backup.UserID]
	for i, b := range userBackups {
		if b.ID == id {
			s.byUser[backup.UserID] = append(userBackups[:i], userBackups[i+1:]...)
			break
		}
	}

	delete(s.backups, id)
	return nil
}

// CreateSchedule creates a backup schedule
func (s *Service) CreateSchedule(userID string, req *models.BackupScheduleCreateRequest) (*models.BackupSchedule, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	schedule := &models.BackupSchedule{
		ID:               utils.GenerateID("sched"),
		UserID:           userID,
		Type:             req.Type,
		CronExpression:   req.CronExpression,
		RetentionDays:    req.RetentionDays,
		IncludeFiles:     req.IncludeFiles,
		IncludeDatabases: req.IncludeDatabases,
		Priority:         req.Priority,
		Enabled:          true,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	// Calculate next run
	nextRun := time.Now().Add(24 * time.Hour) // Simplified
	schedule.NextRunAt = &nextRun

	s.schedules[schedule.ID] = schedule
	return schedule, nil
}

// GetSchedule gets a backup schedule
func (s *Service) GetSchedule(id string) (*models.BackupSchedule, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	schedule, exists := s.schedules[id]
	if !exists {
		return nil, errors.New("schedule not found")
	}
	return schedule, nil
}

// ListSchedules lists schedules for a user
func (s *Service) ListSchedules(userID string) []*models.BackupSchedule {
	s.mu.RLock()
	defer s.mu.RUnlock()

	schedules := make([]*models.BackupSchedule, 0)
	for _, sched := range s.schedules {
		if sched.UserID == userID {
			schedules = append(schedules, sched)
		}
	}
	return schedules
}

// UpdateSchedule updates a backup schedule
func (s *Service) UpdateSchedule(id string, enabled bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	schedule, exists := s.schedules[id]
	if !exists {
		return errors.New("schedule not found")
	}

	schedule.Enabled = enabled
	schedule.UpdatedAt = time.Now()
	return nil
}

// DeleteSchedule deletes a backup schedule
func (s *Service) DeleteSchedule(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.schedules[id]; !exists {
		return errors.New("schedule not found")
	}

	delete(s.schedules, id)
	return nil
}

// Restore initiates a restore operation
func (s *Service) Restore(userID string, req *models.RestoreRequest) (*models.RestoreStatus, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	backup, exists := s.backups[req.BackupID]
	if !exists {
		return nil, errors.New("backup not found")
	}

	if backup.UserID != userID {
		return nil, errors.New("backup belongs to another user")
	}

	if backup.Status != models.BackupStatusCompleted {
		return nil, errors.New("backup is not completed")
	}

	restore := &models.RestoreStatus{
		ID:       utils.GenerateID("rest"),
		BackupID: req.BackupID,
		Status:   models.BackupStatusPending,
		Progress: 0,
	}

	now := time.Now()
	restore.StartedAt = &now

	s.restores[restore.ID] = restore
	return restore, nil
}

// GetRestoreStatus gets restore status
func (s *Service) GetRestoreStatus(id string) (*models.RestoreStatus, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	restore, exists := s.restores[id]
	if !exists {
		return nil, errors.New("restore not found")
	}
	return restore, nil
}

// StartBackup starts processing a backup
func (s *Service) StartBackup(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	backup, exists := s.backups[id]
	if !exists {
		return errors.New("backup not found")
	}

	now := time.Now()
	backup.Status = models.BackupStatusRunning
	backup.StartedAt = &now

	return nil
}

// CompleteBackup marks a backup as completed
func (s *Service) CompleteBackup(id string, sizeMB int64, checksum string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	backup, exists := s.backups[id]
	if !exists {
		return errors.New("backup not found")
	}

	now := time.Now()
	backup.Status = models.BackupStatusCompleted
	backup.CompletedAt = &now
	backup.SizeMB = sizeMB
	backup.Checksum = checksum

	return nil
}

// FailBackup marks a backup as failed
func (s *Service) FailBackup(id, errorMessage string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	backup, exists := s.backups[id]
	if !exists {
		return errors.New("backup not found")
	}

	backup.Status = models.BackupStatusFailed
	backup.ErrorMessage = &errorMessage

	return nil
}

// findLatestBackup finds the latest backup of a specific type
func (s *Service) findLatestBackup(userID string, backupType models.BackupType) *models.Backup {
	var latest *models.Backup
	for _, backup := range s.byUser[userID] {
		if backup.Type == backupType && backup.Status == models.BackupStatusCompleted {
			if latest == nil || backup.CreatedAt.After(latest.CreatedAt) {
				latest = backup
			}
		}
	}
	return latest
}

// addToQueue adds a backup task to the queue
func (s *Service) addToQueue(task *BackupTask) {
	// Insert based on priority (higher priority first)
	inserted := false
	for i, t := range s.queue {
		if task.Priority > t.Priority {
			s.queue = append(s.queue[:i], append([]*BackupTask{task}, s.queue[i:]...)...)
			inserted = true
			break
		}
	}
	if !inserted {
		s.queue = append(s.queue, task)
	}
}

// GetNextTask gets the next backup task from the queue
func (s *Service) GetNextTask() *BackupTask {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.queue) == 0 {
		return nil
	}

	task := s.queue[0]
	s.queue = s.queue[1:]
	return task
}

// DeleteAllByUser deletes all backups for a user
func (s *Service) DeleteAllByUser(userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	for id, backup := range s.backups {
		if backup.UserID == userID {
			delete(s.backups, id)
		}
	}
	delete(s.byUser, userID)
	return nil
}
