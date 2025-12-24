package models

import "time"

// EventType represents the type of event
type EventType string

// Event represents an internal event for pub/sub
type Event struct {
	ID        string                 `json:"id"`
	Type      EventType              `json:"type"`
	Source    string                 `json:"source"`
	Subject   string                 `json:"subject"`
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
}

// EventSubscription represents a subscription to events
type EventSubscription struct {
	ID          string      `json:"id"`
	EventType   EventType   `json:"event_type"`
	Handler     string      `json:"handler"`
	Filter      string      `json:"filter,omitempty"`
	CreatedAt   time.Time   `json:"created_at"`
}

// Webhook represents a webhook configuration
type Webhook struct {
	ID           string    `json:"id"`
	UserID       string    `json:"user_id"`
	Name         string    `json:"name"`
	URL          string    `json:"url"`
	Secret       string    `json:"-"`
	Events       []string  `json:"events"`
	Enabled      bool      `json:"enabled"`
	RetryCount   int       `json:"retry_count"`
	RetryDelay   int       `json:"retry_delay_seconds"`
	LastCalledAt *time.Time `json:"last_called_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// WebhookDelivery represents a webhook delivery attempt
type WebhookDelivery struct {
	ID           string     `json:"id"`
	WebhookID    string     `json:"webhook_id"`
	EventID      string     `json:"event_id"`
	Payload      string     `json:"payload"`
	StatusCode   int        `json:"status_code"`
	Response     string     `json:"response"`
	Attempt      int        `json:"attempt"`
	Success      bool       `json:"success"`
	DeliveredAt  time.Time  `json:"delivered_at"`
	NextRetryAt  *time.Time `json:"next_retry_at,omitempty"`
}

// WebhookCreateRequest represents a request to create a webhook
type WebhookCreateRequest struct {
	Name       string   `json:"name" validate:"required,min=1,max=64"`
	URL        string   `json:"url" validate:"required,url"`
	Events     []string `json:"events" validate:"required,min=1"`
	RetryCount int      `json:"retry_count,omitempty"`
	RetryDelay int      `json:"retry_delay_seconds,omitempty"`
}

// WebhookUpdateRequest represents a request to update a webhook
type WebhookUpdateRequest struct {
	Name       *string   `json:"name,omitempty" validate:"omitempty,min=1,max=64"`
	URL        *string   `json:"url,omitempty" validate:"omitempty,url"`
	Events     *[]string `json:"events,omitempty"`
	Enabled    *bool     `json:"enabled,omitempty"`
	RetryCount *int      `json:"retry_count,omitempty"`
	RetryDelay *int      `json:"retry_delay_seconds,omitempty"`
}
