// Package cron provides cron job management for OweHost
package cron

import (
	"errors"
	"sync"
	"time"

	"github.com/iSundram/OweHost/pkg/models"
	"github.com/iSundram/OweHost/pkg/utils"
)

// Service provides cron job functionality
type Service struct {
	jobs       map[string]*models.CronJob
	executions map[string][]*models.CronJobExecution
	byUser     map[string][]*models.CronJob
	mu         sync.RWMutex
}

// NewService creates a new cron service
func NewService() *Service {
	return &Service{
		jobs:       make(map[string]*models.CronJob),
		executions: make(map[string][]*models.CronJobExecution),
		byUser:     make(map[string][]*models.CronJob),
	}
}

// Create creates a new cron job
func (s *Service) Create(userID string, req *models.CronJobCreateRequest) (*models.CronJob, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	timeout := req.Timeout
	if timeout == 0 {
		timeout = 3600 // 1 hour default
	}

	job := &models.CronJob{
		ID:             utils.GenerateID("cron"),
		UserID:         userID,
		Name:           req.Name,
		Command:        req.Command,
		CronExpression: req.CronExpression,
		Status:         models.CronJobStatusActive,
		Timeout:        timeout,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// Calculate next run
	nextRun := s.calculateNextRun(job.CronExpression)
	job.NextRunAt = nextRun

	s.jobs[job.ID] = job
	s.byUser[userID] = append(s.byUser[userID], job)
	s.executions[job.ID] = make([]*models.CronJobExecution, 0)

	return job, nil
}

// Get gets a cron job by ID
func (s *Service) Get(id string) (*models.CronJob, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	job, exists := s.jobs[id]
	if !exists {
		return nil, errors.New("cron job not found")
	}
	return job, nil
}

// ListByUser lists cron jobs for a user
func (s *Service) ListByUser(userID string) []*models.CronJob {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.byUser[userID]
}

// Update updates a cron job
func (s *Service) Update(id string, req *models.CronJobUpdateRequest) (*models.CronJob, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	job, exists := s.jobs[id]
	if !exists {
		return nil, errors.New("cron job not found")
	}

	if req.Name != nil {
		job.Name = *req.Name
	}
	if req.Command != nil {
		job.Command = *req.Command
	}
	if req.CronExpression != nil {
		job.CronExpression = *req.CronExpression
		job.NextRunAt = s.calculateNextRun(*req.CronExpression)
	}
	if req.Status != nil {
		job.Status = *req.Status
	}
	if req.Timeout != nil {
		job.Timeout = *req.Timeout
	}

	job.UpdatedAt = time.Now()
	return job, nil
}

// Delete deletes a cron job
func (s *Service) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	job, exists := s.jobs[id]
	if !exists {
		return errors.New("cron job not found")
	}

	// Remove from user's jobs
	userJobs := s.byUser[job.UserID]
	for i, j := range userJobs {
		if j.ID == id {
			s.byUser[job.UserID] = append(userJobs[:i], userJobs[i+1:]...)
			break
		}
	}

	delete(s.jobs, id)
	delete(s.executions, id)
	return nil
}

// Pause pauses a cron job
func (s *Service) Pause(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	job, exists := s.jobs[id]
	if !exists {
		return errors.New("cron job not found")
	}

	job.Status = models.CronJobStatusPaused
	job.UpdatedAt = time.Now()
	return nil
}

// Resume resumes a cron job
func (s *Service) Resume(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	job, exists := s.jobs[id]
	if !exists {
		return errors.New("cron job not found")
	}

	job.Status = models.CronJobStatusActive
	job.NextRunAt = s.calculateNextRun(job.CronExpression)
	job.UpdatedAt = time.Now()
	return nil
}

// Execute executes a cron job
func (s *Service) Execute(id string) (*models.CronJobExecution, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	job, exists := s.jobs[id]
	if !exists {
		return nil, errors.New("cron job not found")
	}

	execution := &models.CronJobExecution{
		ID:        utils.GenerateID("exec"),
		CronJobID: id,
		StartedAt: time.Now(),
		Stdout:    "",
		Stderr:    "",
	}

	s.executions[id] = append(s.executions[id], execution)

	// Update job
	now := time.Now()
	job.LastRunAt = &now
	job.NextRunAt = s.calculateNextRun(job.CronExpression)

	return execution, nil
}

// CompleteExecution completes a cron job execution
func (s *Service) CompleteExecution(execID string, exitCode int, stdout, stderr string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, execs := range s.executions {
		for _, exec := range execs {
			if exec.ID == execID {
				now := time.Now()
				exec.CompletedAt = &now
				exec.ExitCode = &exitCode
				exec.Stdout = stdout
				exec.Stderr = stderr
				duration := int(now.Sub(exec.StartedAt).Milliseconds())
				exec.Duration = &duration

				// Update job's last exit code
				job := s.jobs[exec.CronJobID]
				if job != nil {
					job.LastExitCode = &exitCode
					job.UpdatedAt = now
				}

				return nil
			}
		}
	}

	return errors.New("execution not found")
}

// GetExecutions gets executions for a cron job
func (s *Service) GetExecutions(jobID string, limit int) []*models.CronJobExecution {
	s.mu.RLock()
	defer s.mu.RUnlock()

	execs := s.executions[jobID]
	if limit <= 0 || limit > len(execs) {
		limit = len(execs)
	}

	// Return most recent
	start := len(execs) - limit
	if start < 0 {
		start = 0
	}
	return execs[start:]
}

// GetDueJobs gets jobs that are due to run
func (s *Service) GetDueJobs() []*models.CronJob {
	s.mu.RLock()
	defer s.mu.RUnlock()

	now := time.Now()
	dueJobs := make([]*models.CronJob, 0)

	for _, job := range s.jobs {
		if job.Status != models.CronJobStatusActive {
			continue
		}
		if job.NextRunAt != nil && job.NextRunAt.Before(now) {
			dueJobs = append(dueJobs, job)
		}
	}

	return dueJobs
}

// calculateNextRun calculates the next run time for a cron expression
func (s *Service) calculateNextRun(expr string) *time.Time {
	// Simplified implementation - in production, use a cron parser library
	// For now, schedule 1 minute from now
	next := time.Now().Add(time.Minute)
	return &next
}

// CleanupOldExecutions removes old execution records
func (s *Service) CleanupOldExecutions(maxAge time.Duration) int {
	s.mu.Lock()
	defer s.mu.Unlock()

	threshold := time.Now().Add(-maxAge)
	removed := 0

	for jobID, execs := range s.executions {
		newExecs := make([]*models.CronJobExecution, 0)
		for _, exec := range execs {
			if exec.CompletedAt != nil && exec.CompletedAt.Before(threshold) {
				removed++
			} else {
				newExecs = append(newExecs, exec)
			}
		}
		s.executions[jobID] = newExecs
	}

	return removed
}
