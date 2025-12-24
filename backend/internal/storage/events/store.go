// Package events provides immutable event/audit system for OweHost
package events

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/iSundram/OweHost/pkg/utils"
)

const (
	// EventsBasePath is the base path for event storage
	EventsBasePath = "/opt/owehost/logs/events"
	// AlertsBasePath is the base path for security alerts
	AlertsBasePath = "/opt/owehost/logs/alerts"
)

// Store handles event persistence
type Store struct {
	basePath   string
	alertsPath string
	mu         sync.RWMutex
}

// NewStore creates a new event store
func NewStore() *Store {
	s := &Store{
		basePath:   EventsBasePath,
		alertsPath: AlertsBasePath,
	}
	s.ensureDirectories()
	return s
}

// NewStoreWithPath creates an event store with custom paths
func NewStoreWithPath(eventsPath, alertsPath string) *Store {
	s := &Store{
		basePath:   eventsPath,
		alertsPath: alertsPath,
	}
	s.ensureDirectories()
	return s
}

// ensureDirectories creates required directories
func (s *Store) ensureDirectories() {
	os.MkdirAll(s.basePath, 0755)
	os.MkdirAll(s.alertsPath, 0755)
}

// Save saves an event to the filesystem
func (s *Store) Save(event *Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Generate ID if not set
	if event.ID == "" {
		event.ID = utils.GenerateID("evt")
	}

	// Set timestamp if not set
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	// Create date-based directory
	dateDir := event.Timestamp.Format("2006/01/02")
	dirPath := filepath.Join(s.basePath, dateDir)
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return fmt.Errorf("failed to create event directory: %w", err)
	}

	// Generate filename
	filename := fmt.Sprintf("%s-%s-%s.json",
		event.Timestamp.Format("150405"),
		event.Type,
		event.ID,
	)
	path := filepath.Join(dirPath, filename)

	// Marshal event
	data, err := json.MarshalIndent(event, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// Write as read-only (immutable)
	if err := os.WriteFile(path, data, 0444); err != nil {
		return fmt.Errorf("failed to write event: %w", err)
	}

	return nil
}

// Query queries events based on filters
func (s *Store) Query(filters EventFilters) ([]Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var events []Event

	// Determine date range
	startDate := time.Now().AddDate(0, 0, -30) // Default last 30 days
	endDate := time.Now()

	if filters.StartTime != nil {
		startDate = *filters.StartTime
	}
	if filters.EndTime != nil {
		endDate = *filters.EndTime
	}

	// Walk through date directories
	for date := startDate; !date.After(endDate); date = date.AddDate(0, 0, 1) {
		dateDir := filepath.Join(s.basePath, date.Format("2006/01/02"))
		entries, err := os.ReadDir(dateDir)
		if err != nil {
			continue // Directory doesn't exist
		}

		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
				continue
			}

			event, err := s.readEvent(filepath.Join(dateDir, entry.Name()))
			if err != nil {
				continue
			}

			// Apply filters
			if s.matchesFilters(event, filters) {
				events = append(events, *event)
			}
		}
	}

	// Sort by timestamp descending
	sort.Slice(events, func(i, j int) bool {
		return events[i].Timestamp.After(events[j].Timestamp)
	})

	// Apply pagination
	if filters.Offset > 0 && filters.Offset < len(events) {
		events = events[filters.Offset:]
	}
	if filters.Limit > 0 && filters.Limit < len(events) {
		events = events[:filters.Limit]
	}

	return events, nil
}

// readEvent reads an event from a file
func (s *Store) readEvent(path string) (*Event, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var event Event
	if err := json.Unmarshal(data, &event); err != nil {
		return nil, err
	}

	return &event, nil
}

// matchesFilters checks if an event matches the given filters
func (s *Store) matchesFilters(event *Event, filters EventFilters) bool {
	if filters.AccountID != nil && event.AccountID != *filters.AccountID {
		return false
	}
	if filters.Type != nil && event.Type != *filters.Type {
		return false
	}
	if filters.Actor != nil && event.Actor != *filters.Actor {
		return false
	}
	if filters.ActorType != nil && event.ActorType != *filters.ActorType {
		return false
	}
	if filters.Result != nil && event.Result != *filters.Result {
		return false
	}
	if filters.RequestID != nil && event.RequestID != *filters.RequestID {
		return false
	}
	if filters.StartTime != nil && event.Timestamp.Before(*filters.StartTime) {
		return false
	}
	if filters.EndTime != nil && event.Timestamp.After(*filters.EndTime) {
		return false
	}
	return true
}

// GetByID retrieves an event by ID
func (s *Store) GetByID(id string) (*Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Search in recent directories (last 30 days)
	for i := 0; i < 30; i++ {
		date := time.Now().AddDate(0, 0, -i)
		dateDir := filepath.Join(s.basePath, date.Format("2006/01/02"))
		entries, err := os.ReadDir(dateDir)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			if strings.Contains(entry.Name(), id) {
				return s.readEvent(filepath.Join(dateDir, entry.Name()))
			}
		}
	}

	return nil, fmt.Errorf("event not found: %s", id)
}

// GetStats returns event statistics
func (s *Store) GetStats(days int) (*EventStats, error) {
	filters := EventFilters{
		StartTime: func() *time.Time { t := time.Now().AddDate(0, 0, -days); return &t }(),
	}

	events, err := s.Query(filters)
	if err != nil {
		return nil, err
	}

	stats := &EventStats{
		TotalEvents: int64(len(events)),
		ByType:      make(map[string]int64),
		ByResult:    make(map[string]int64),
		ByActorType: make(map[string]int64),
	}

	for _, event := range events {
		stats.ByType[string(event.Type)]++
		stats.ByResult[string(event.Result)]++
		stats.ByActorType[event.ActorType]++

		if stats.LastEventTime == nil || event.Timestamp.After(*stats.LastEventTime) {
			t := event.Timestamp
			stats.LastEventTime = &t
		}
	}

	return stats, nil
}

// GetAccountEvents returns events for a specific account
func (s *Store) GetAccountEvents(accountID int, limit int) ([]Event, error) {
	return s.Query(EventFilters{
		AccountID: &accountID,
		Limit:     limit,
	})
}

// GetSecurityEvents returns security-related events
func (s *Store) GetSecurityEvents(limit int) ([]Event, error) {
	allEvents, err := s.Query(EventFilters{Limit: limit * 3}) // Get more to filter
	if err != nil {
		return nil, err
	}

	var securityEvents []Event
	for _, event := range allEvents {
		if strings.HasPrefix(string(event.Type), "security.") {
			securityEvents = append(securityEvents, event)
			if len(securityEvents) >= limit {
				break
			}
		}
	}

	return securityEvents, nil
}

// SaveAlert saves a security alert
func (s *Store) SaveAlert(alert *SecurityAlert) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if alert.ID == "" {
		alert.ID = utils.GenerateID("alt")
	}
	if alert.Timestamp.IsZero() {
		alert.Timestamp = time.Now()
	}

	filename := fmt.Sprintf("%s-%s.json",
		alert.Timestamp.Format("2006-01-02-150405"),
		alert.ID,
	)
	path := filepath.Join(s.alertsPath, filename)

	data, err := json.MarshalIndent(alert, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// GetUnresolvedAlerts returns all unresolved security alerts
func (s *Store) GetUnresolvedAlerts() ([]SecurityAlert, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entries, err := os.ReadDir(s.alertsPath)
	if err != nil {
		return nil, err
	}

	var alerts []SecurityAlert
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		data, err := os.ReadFile(filepath.Join(s.alertsPath, entry.Name()))
		if err != nil {
			continue
		}

		var alert SecurityAlert
		if err := json.Unmarshal(data, &alert); err != nil {
			continue
		}

		if !alert.Resolved {
			alerts = append(alerts, alert)
		}
	}

	return alerts, nil
}

// ResolveAlert marks an alert as resolved
func (s *Store) ResolveAlert(alertID, resolvedBy string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	entries, err := os.ReadDir(s.alertsPath)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if strings.Contains(entry.Name(), alertID) {
			path := filepath.Join(s.alertsPath, entry.Name())
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			var alert SecurityAlert
			if err := json.Unmarshal(data, &alert); err != nil {
				return err
			}

			alert.Resolved = true
			now := time.Now()
			alert.ResolvedAt = &now
			alert.ResolvedBy = &resolvedBy

			newData, err := json.MarshalIndent(alert, "", "  ")
			if err != nil {
				return err
			}

			return os.WriteFile(path, newData, 0644)
		}
	}

	return fmt.Errorf("alert not found: %s", alertID)
}

// Purge removes events older than the specified number of days
func (s *Store) Purge(daysToKeep int) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	cutoff := time.Now().AddDate(0, 0, -daysToKeep)
	count := 0

	// Walk through all date directories
	err := filepath.Walk(s.basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if info.IsDir() {
			// Check if directory is older than cutoff
			rel, _ := filepath.Rel(s.basePath, path)
			if len(rel) >= 10 { // yyyy/mm/dd format
				dirDate, err := time.Parse("2006/01/02", rel[:10])
				if err == nil && dirDate.Before(cutoff) {
					entries, _ := os.ReadDir(path)
					count += len(entries)
					os.RemoveAll(path)
				}
			}
		}
		return nil
	})

	return count, err
}
