// Package logging provides logging and auditing services for OweHost
package logging

import (
	"strconv"
	"sync"
	"time"

	"github.com/iSundram/OweHost/pkg/models"
	"github.com/iSundram/OweHost/pkg/utils"
)

// Service provides logging and auditing functionality
type Service struct {
	logs    []*models.LogEntry
	audits  []*models.AuditEntry
	metrics map[string]*models.Metric
	mu      sync.RWMutex
}

// NewService creates a new logging service
func NewService() *Service {
	return &Service{
		logs:    make([]*models.LogEntry, 0),
		audits:  make([]*models.AuditEntry, 0),
		metrics: make(map[string]*models.Metric),
	}
}

// Log creates a log entry
func (s *Service) Log(level models.LogLevel, service, message string, userID, requestID, traceID *string, metadata map[string]interface{}) *models.LogEntry {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry := &models.LogEntry{
		ID:        utils.GenerateID("log"),
		Timestamp: time.Now(),
		Level:     level,
		Service:   service,
		Message:   message,
		UserID:    userID,
		RequestID: requestID,
		TraceID:   traceID,
		Metadata:  metadata,
	}

	s.logs = append(s.logs, entry)

	// Keep only last 100000 entries
	if len(s.logs) > 100000 {
		s.logs = s.logs[len(s.logs)-100000:]
	}

	return entry
}

// Debug logs a debug message
func (s *Service) Debug(service, message string) *models.LogEntry {
	return s.Log(models.LogLevelDebug, service, message, nil, nil, nil, nil)
}

// Info logs an info message
func (s *Service) Info(service, message string) *models.LogEntry {
	return s.Log(models.LogLevelInfo, service, message, nil, nil, nil, nil)
}

// Warn logs a warning message
func (s *Service) Warn(service, message string) *models.LogEntry {
	return s.Log(models.LogLevelWarn, service, message, nil, nil, nil, nil)
}

// Error logs an error message
func (s *Service) Error(service, message string) *models.LogEntry {
	return s.Log(models.LogLevelError, service, message, nil, nil, nil, nil)
}

// QueryLogs queries log entries
func (s *Service) QueryLogs(req *models.LogQueryRequest) []*models.LogEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()

	results := make([]*models.LogEntry, 0)

	for _, entry := range s.logs {
		// Filter by time range
		if entry.Timestamp.Before(req.StartTime) || entry.Timestamp.After(req.EndTime) {
			continue
		}

		// Filter by level
		if req.Level != nil && entry.Level != *req.Level {
			continue
		}

		// Filter by service
		if req.Service != nil && entry.Service != *req.Service {
			continue
		}

		// Filter by user
		if req.UserID != nil && (entry.UserID == nil || *entry.UserID != *req.UserID) {
			continue
		}

		results = append(results, entry)
	}

	// Apply pagination
	start := req.Offset
	if start > len(results) {
		start = len(results)
	}
	end := start + req.Limit
	if end > len(results) {
		end = len(results)
	}

	return results[start:end]
}

// Audit creates an audit entry
func (s *Service) Audit(userID, action, resourceType, resourceID, ipAddress, userAgent string, oldValue, newValue map[string]interface{}, success bool) *models.AuditEntry {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry := &models.AuditEntry{
		ID:           utils.GenerateID("audit"),
		Timestamp:    time.Now(),
		UserID:       userID,
		Action:       action,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		OldValue:     oldValue,
		NewValue:     newValue,
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
		Success:      success,
	}

	s.audits = append(s.audits, entry)

	// Keep only last 100000 entries (immutable, append-only)
	if len(s.audits) > 100000 {
		s.audits = s.audits[len(s.audits)-100000:]
	}

	return entry
}

// QueryAudits queries audit entries
func (s *Service) QueryAudits(req *models.AuditQueryRequest) []*models.AuditEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()

	results := make([]*models.AuditEntry, 0)

	for _, entry := range s.audits {
		// Filter by time range
		if entry.Timestamp.Before(req.StartTime) || entry.Timestamp.After(req.EndTime) {
			continue
		}

		// Filter by user
		if req.UserID != nil && entry.UserID != *req.UserID {
			continue
		}

		// Filter by action
		if req.Action != nil && entry.Action != *req.Action {
			continue
		}

		// Filter by resource type
		if req.ResourceType != nil && entry.ResourceType != *req.ResourceType {
			continue
		}

		// Filter by resource ID
		if req.ResourceID != nil && entry.ResourceID != *req.ResourceID {
			continue
		}

		results = append(results, entry)
	}

	// Apply pagination
	start := req.Offset
	if start > len(results) {
		start = len(results)
	}
	end := start + req.Limit
	if end > len(results) {
		end = len(results)
	}

	return results[start:end]
}

// RecordMetric records a metric
func (s *Service) RecordMetric(name, metricType string, value float64, labels map[string]string) *models.Metric {
	s.mu.Lock()
	defer s.mu.Unlock()

	metric := &models.Metric{
		Name:      name,
		Type:      metricType,
		Value:     value,
		Labels:    labels,
		Timestamp: time.Now(),
	}

	// Create key from name and labels
	key := name
	for k, v := range labels {
		key += "_" + k + "_" + v
	}

	s.metrics[key] = metric
	return metric
}

// GetMetric gets a metric by name
func (s *Service) GetMetric(name string) *models.Metric {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.metrics[name]
}

// ListMetrics lists all metrics
func (s *Service) ListMetrics() []*models.Metric {
	s.mu.RLock()
	defer s.mu.RUnlock()

	metrics := make([]*models.Metric, 0, len(s.metrics))
	for _, m := range s.metrics {
		metrics = append(metrics, m)
	}
	return metrics
}

// ExportPrometheus exports metrics in Prometheus format
func (s *Service) ExportPrometheus() string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	output := ""
	for _, metric := range s.metrics {
		// Format: metric_name{label1="value1",label2="value2"} value
		output += metric.Name
		if len(metric.Labels) > 0 {
			output += "{"
			first := true
			for k, v := range metric.Labels {
				if !first {
					output += ","
				}
				output += k + "=\"" + v + "\""
				first = false
			}
			output += "}"
		}
		output += " " + formatFloat(metric.Value) + "\n"
	}
	return output
}

// formatFloat formats a float64 for output
func formatFloat(f float64) string {
	return strconv.FormatFloat(f, 'f', -1, 64)
}

// IncrementCounter increments a counter metric
func (s *Service) IncrementCounter(name string, labels map[string]string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := name
	for k, v := range labels {
		key += "_" + k + "_" + v
	}

	if metric, exists := s.metrics[key]; exists {
		metric.Value++
		metric.Timestamp = time.Now()
	} else {
		s.metrics[key] = &models.Metric{
			Name:      name,
			Type:      "counter",
			Value:     1,
			Labels:    labels,
			Timestamp: time.Now(),
		}
	}
}

// SetGauge sets a gauge metric
func (s *Service) SetGauge(name string, value float64, labels map[string]string) {
	s.RecordMetric(name, "gauge", value, labels)
}

// ObserveHistogram observes a value for a histogram
func (s *Service) ObserveHistogram(name string, value float64, labels map[string]string) {
	// Simplified - just record the last value
	s.RecordMetric(name, "histogram", value, labels)
}
