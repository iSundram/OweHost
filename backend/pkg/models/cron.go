package models

import "time"

// CronJobStatus represents the status of a cron job
type CronJobStatus string

const (
	CronJobStatusActive   CronJobStatus = "active"
	CronJobStatusPaused   CronJobStatus = "paused"
	CronJobStatusDisabled CronJobStatus = "disabled"
)

// CronJob represents a cron job
type CronJob struct {
	ID             string        `json:"id"`
	UserID         string        `json:"user_id"`
	Name           string        `json:"name"`
	Command        string        `json:"command"`
	CronExpression string        `json:"cron_expression"`
	Status         CronJobStatus `json:"status"`
	Timeout        int           `json:"timeout"`
	LastRunAt      *time.Time    `json:"last_run_at,omitempty"`
	NextRunAt      *time.Time    `json:"next_run_at,omitempty"`
	LastExitCode   *int          `json:"last_exit_code,omitempty"`
	CreatedAt      time.Time     `json:"created_at"`
	UpdatedAt      time.Time     `json:"updated_at"`
}

// CronJobExecution represents a cron job execution
type CronJobExecution struct {
	ID         string     `json:"id"`
	CronJobID  string     `json:"cron_job_id"`
	StartedAt  time.Time  `json:"started_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	ExitCode   *int       `json:"exit_code,omitempty"`
	Stdout     string     `json:"stdout"`
	Stderr     string     `json:"stderr"`
	Duration   *int       `json:"duration_ms,omitempty"`
}

// CronJobCreateRequest represents a request to create a cron job
type CronJobCreateRequest struct {
	Name           string `json:"name" validate:"required,min=1,max=64"`
	Command        string `json:"command" validate:"required"`
	CronExpression string `json:"cron_expression" validate:"required"`
	Timeout        int    `json:"timeout,omitempty"`
}

// CronJobUpdateRequest represents a request to update a cron job
type CronJobUpdateRequest struct {
	Name           *string        `json:"name,omitempty" validate:"omitempty,min=1,max=64"`
	Command        *string        `json:"command,omitempty"`
	CronExpression *string        `json:"cron_expression,omitempty"`
	Status         *CronJobStatus `json:"status,omitempty"`
	Timeout        *int           `json:"timeout,omitempty"`
}
