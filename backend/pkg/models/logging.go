package models

import "time"

// LogLevel represents the log level
type LogLevel string

const (
	LogLevelDebug LogLevel = "debug"
	LogLevelInfo  LogLevel = "info"
	LogLevelWarn  LogLevel = "warn"
	LogLevelError LogLevel = "error"
)

// LogEntry represents a structured log entry
type LogEntry struct {
	ID         string                 `json:"id"`
	Timestamp  time.Time              `json:"timestamp"`
	Level      LogLevel               `json:"level"`
	Service    string                 `json:"service"`
	Message    string                 `json:"message"`
	UserID     *string                `json:"user_id,omitempty"`
	RequestID  *string                `json:"request_id,omitempty"`
	TraceID    *string                `json:"trace_id,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// AuditEntry represents an audit trail entry
type AuditEntry struct {
	ID           string                 `json:"id"`
	Timestamp    time.Time              `json:"timestamp"`
	UserID       string                 `json:"user_id"`
	Action       string                 `json:"action"`
	ResourceType string                 `json:"resource_type"`
	ResourceID   string                 `json:"resource_id"`
	OldValue     map[string]interface{} `json:"old_value,omitempty"`
	NewValue     map[string]interface{} `json:"new_value,omitempty"`
	IPAddress    string                 `json:"ip_address"`
	UserAgent    string                 `json:"user_agent"`
	Success      bool                   `json:"success"`
}

// Metric represents a metric for observability
type Metric struct {
	Name      string            `json:"name"`
	Type      string            `json:"type"`
	Value     float64           `json:"value"`
	Labels    map[string]string `json:"labels"`
	Timestamp time.Time         `json:"timestamp"`
}

// MetricDefinition represents a metric definition for Prometheus
type MetricDefinition struct {
	Name        string   `json:"name"`
	Help        string   `json:"help"`
	Type        string   `json:"type"`
	LabelNames  []string `json:"label_names"`
}

// LogQueryRequest represents a request to query logs
type LogQueryRequest struct {
	StartTime time.Time  `json:"start_time" validate:"required"`
	EndTime   time.Time  `json:"end_time" validate:"required"`
	Level     *LogLevel  `json:"level,omitempty"`
	Service   *string    `json:"service,omitempty"`
	UserID    *string    `json:"user_id,omitempty"`
	Query     string     `json:"query,omitempty"`
	Limit     int        `json:"limit" validate:"min=1,max=1000"`
	Offset    int        `json:"offset" validate:"min=0"`
}

// AuditQueryRequest represents a request to query audit entries
type AuditQueryRequest struct {
	StartTime    time.Time `json:"start_time" validate:"required"`
	EndTime      time.Time `json:"end_time" validate:"required"`
	UserID       *string   `json:"user_id,omitempty"`
	Action       *string   `json:"action,omitempty"`
	ResourceType *string   `json:"resource_type,omitempty"`
	ResourceID   *string   `json:"resource_id,omitempty"`
	Limit        int       `json:"limit" validate:"min=1,max=1000"`
	Offset       int       `json:"offset" validate:"min=0"`
}
