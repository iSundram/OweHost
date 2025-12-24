// Package stats provides statistics and analytics services for OweHost
package stats

import (
	"math/rand"
	"sync"
	"time"

	"github.com/iSundram/OweHost/pkg/models"
	"github.com/iSundram/OweHost/pkg/utils"
)

// Service provides statistics and analytics functionality
type Service struct {
	bandwidthStats  map[string][]*models.BandwidthStat
	visitorStats    map[string][]*models.VisitorStat
	errorLogs       map[string][]*models.ErrorLogEntry
	accessLogs      map[string][]*models.AccessLogEntry
	resourceStats   map[string]*models.ResourceStat
	mu              sync.RWMutex
}

// NewService creates a new stats service
func NewService() *Service {
	return &Service{
		bandwidthStats: make(map[string][]*models.BandwidthStat),
		visitorStats:   make(map[string][]*models.VisitorStat),
		errorLogs:      make(map[string][]*models.ErrorLogEntry),
		accessLogs:     make(map[string][]*models.AccessLogEntry),
		resourceStats:  make(map[string]*models.ResourceStat),
	}
}

// GetBandwidthStats gets bandwidth statistics for a domain
func (s *Service) GetBandwidthStats(domainID string, startDate, endDate time.Time) []*models.BandwidthStat {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stats := s.bandwidthStats[domainID]
	if stats == nil {
		// Generate mock data for demonstration
		stats = s.generateMockBandwidthStats(domainID, startDate, endDate)
	}

	// Filter by date range
	filtered := make([]*models.BandwidthStat, 0)
	for _, stat := range stats {
		if !stat.Date.Before(startDate) && !stat.Date.After(endDate) {
			filtered = append(filtered, stat)
		}
	}

	return filtered
}

// GetVisitorStats gets visitor statistics for a domain
func (s *Service) GetVisitorStats(domainID string, startDate, endDate time.Time) []*models.VisitorStat {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stats := s.visitorStats[domainID]
	if stats == nil {
		// Generate mock data for demonstration
		stats = s.generateMockVisitorStats(domainID, startDate, endDate)
	}

	// Filter by date range
	filtered := make([]*models.VisitorStat, 0)
	for _, stat := range stats {
		if !stat.Date.Before(startDate) && !stat.Date.After(endDate) {
			filtered = append(filtered, stat)
		}
	}

	return filtered
}

// GetErrorLogs gets error logs for a domain
func (s *Service) GetErrorLogs(domainID string, limit int) []*models.ErrorLogEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()

	logs := s.errorLogs[domainID]
	if logs == nil {
		// Generate mock data
		logs = s.generateMockErrorLogs(domainID)
	}

	if limit > 0 && limit < len(logs) {
		return logs[len(logs)-limit:]
	}
	return logs
}

// GetAccessLogs gets access logs for a domain
func (s *Service) GetAccessLogs(domainID string, limit int) []*models.AccessLogEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()

	logs := s.accessLogs[domainID]
	if logs == nil {
		// Generate mock data
		logs = s.generateMockAccessLogs(domainID)
	}

	if limit > 0 && limit < len(logs) {
		return logs[len(logs)-limit:]
	}
	return logs
}

// GetResourceStats gets resource statistics for a user
func (s *Service) GetResourceStats(userID string) *models.ResourceStat {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stat := s.resourceStats[userID]
	if stat == nil {
		// Generate mock data
		stat = s.generateMockResourceStats(userID)
	}
	return stat
}

// GetDomainSummary gets a summary of statistics for a domain
func (s *Service) GetDomainSummary(domainID string) *models.DomainStatsSummary {
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -30) // Last 30 days

	bandwidthStats := s.GetBandwidthStats(domainID, startDate, endDate)
	visitorStats := s.GetVisitorStats(domainID, startDate, endDate)

	var totalBandwidth int64
	var totalVisitors int64
	var totalPageViews int64
	var totalHits int64

	for _, stat := range bandwidthStats {
		totalBandwidth += stat.BytesIn + stat.BytesOut
	}

	for _, stat := range visitorStats {
		totalVisitors += int64(stat.UniqueVisitors)
		totalPageViews += int64(stat.PageViews)
		totalHits += int64(stat.Hits)
	}

	return &models.DomainStatsSummary{
		DomainID:         domainID,
		Period:           "30d",
		TotalBandwidth:   totalBandwidth,
		TotalVisitors:    totalVisitors,
		TotalPageViews:   totalPageViews,
		TotalHits:        totalHits,
		AverageLoadTime:  250, // ms
		ErrorCount:       len(s.GetErrorLogs(domainID, 0)),
	}
}

// GetUserSummary gets a summary of statistics for a user
func (s *Service) GetUserSummary(userID string) *models.UserStatsSummary {
	resourceStats := s.GetResourceStats(userID)

	return &models.UserStatsSummary{
		UserID:          userID,
		DiskUsedMB:      resourceStats.DiskUsedMB,
		DiskLimitMB:     resourceStats.DiskLimitMB,
		DiskPercent:     float64(resourceStats.DiskUsedMB) / float64(resourceStats.DiskLimitMB) * 100,
		BandwidthUsedMB: resourceStats.BandwidthUsedMB,
		BandwidthLimitMB: resourceStats.BandwidthLimitMB,
		BandwidthPercent: float64(resourceStats.BandwidthUsedMB) / float64(resourceStats.BandwidthLimitMB) * 100,
		InodeUsed:       resourceStats.InodeUsed,
		InodeLimit:      resourceStats.InodeLimit,
		DomainsUsed:     resourceStats.DomainsUsed,
		DomainsLimit:    resourceStats.DomainsLimit,
		DatabasesUsed:   resourceStats.DatabasesUsed,
		DatabasesLimit:  resourceStats.DatabasesLimit,
	}
}

// RecordBandwidth records bandwidth usage
func (s *Service) RecordBandwidth(domainID string, bytesIn, bytesOut int64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	today := time.Now().Truncate(24 * time.Hour)

	stats := s.bandwidthStats[domainID]
	for _, stat := range stats {
		if stat.Date.Equal(today) {
			stat.BytesIn += bytesIn
			stat.BytesOut += bytesOut
			stat.Requests++
			return
		}
	}

	// Create new stat for today
	stat := &models.BandwidthStat{
		ID:       utils.GenerateID("bw"),
		DomainID: domainID,
		Date:     today,
		BytesIn:  bytesIn,
		BytesOut: bytesOut,
		Requests: 1,
	}
	s.bandwidthStats[domainID] = append(s.bandwidthStats[domainID], stat)
}

// RecordVisitor records a visitor
func (s *Service) RecordVisitor(domainID, ipAddress, userAgent, path string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	today := time.Now().Truncate(24 * time.Hour)

	stats := s.visitorStats[domainID]
	for _, stat := range stats {
		if stat.Date.Equal(today) {
			stat.Hits++
			stat.PageViews++
			return
		}
	}

	// Create new stat for today
	stat := &models.VisitorStat{
		ID:             utils.GenerateID("vs"),
		DomainID:       domainID,
		Date:           today,
		UniqueVisitors: 1,
		PageViews:      1,
		Hits:           1,
	}
	s.visitorStats[domainID] = append(s.visitorStats[domainID], stat)
}

// RecordError records an error log entry
func (s *Service) RecordError(domainID string, level, message, file string, line int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry := &models.ErrorLogEntry{
		ID:        utils.GenerateID("err"),
		DomainID:  domainID,
		Timestamp: time.Now(),
		Level:     level,
		Message:   message,
		File:      file,
		Line:      line,
	}

	s.errorLogs[domainID] = append(s.errorLogs[domainID], entry)

	// Keep only last 1000 entries
	if len(s.errorLogs[domainID]) > 1000 {
		s.errorLogs[domainID] = s.errorLogs[domainID][len(s.errorLogs[domainID])-1000:]
	}
}

// Helper functions for generating mock data

func (s *Service) generateMockBandwidthStats(domainID string, startDate, endDate time.Time) []*models.BandwidthStat {
	stats := make([]*models.BandwidthStat, 0)
	for d := startDate; !d.After(endDate); d = d.AddDate(0, 0, 1) {
		stat := &models.BandwidthStat{
			ID:       utils.GenerateID("bw"),
			DomainID: domainID,
			Date:     d,
			BytesIn:  int64(rand.Intn(1000000000)),
			BytesOut: int64(rand.Intn(5000000000)),
			Requests: int64(rand.Intn(10000)),
		}
		stats = append(stats, stat)
	}
	return stats
}

func (s *Service) generateMockVisitorStats(domainID string, startDate, endDate time.Time) []*models.VisitorStat {
	stats := make([]*models.VisitorStat, 0)
	for d := startDate; !d.After(endDate); d = d.AddDate(0, 0, 1) {
		stat := &models.VisitorStat{
			ID:             utils.GenerateID("vs"),
			DomainID:       domainID,
			Date:           d,
			UniqueVisitors: rand.Intn(500),
			PageViews:      rand.Intn(2000),
			Hits:           rand.Intn(5000),
			BounceRate:     float64(rand.Intn(80)),
			AvgSessionTime: rand.Intn(300),
		}
		stats = append(stats, stat)
	}
	return stats
}

func (s *Service) generateMockErrorLogs(domainID string) []*models.ErrorLogEntry {
	levels := []string{"error", "warning", "notice"}
	messages := []string{
		"PHP Fatal error: Class 'App\\Controller' not found",
		"PHP Warning: Invalid argument supplied for foreach()",
		"PHP Notice: Undefined variable: user",
		"PHP Parse error: syntax error, unexpected '}'",
	}

	logs := make([]*models.ErrorLogEntry, 0)
	for i := 0; i < 20; i++ {
		log := &models.ErrorLogEntry{
			ID:        utils.GenerateID("err"),
			DomainID:  domainID,
			Timestamp: time.Now().Add(-time.Duration(i) * time.Hour),
			Level:     levels[rand.Intn(len(levels))],
			Message:   messages[rand.Intn(len(messages))],
			File:      "/home/user/public_html/index.php",
			Line:      rand.Intn(500),
		}
		logs = append(logs, log)
	}
	return logs
}

func (s *Service) generateMockAccessLogs(domainID string) []*models.AccessLogEntry {
	methods := []string{"GET", "POST", "PUT", "DELETE"}
	paths := []string{"/", "/api/users", "/products", "/about", "/contact"}
	statuses := []int{200, 200, 200, 200, 301, 404, 500}
	userAgents := []string{
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)",
		"curl/7.68.0",
	}

	logs := make([]*models.AccessLogEntry, 0)
	for i := 0; i < 50; i++ {
		log := &models.AccessLogEntry{
			ID:           utils.GenerateID("acc"),
			DomainID:     domainID,
			Timestamp:    time.Now().Add(-time.Duration(i) * time.Minute),
			IPAddress:    "192.168.1." + string(rune(rand.Intn(255))),
			Method:       methods[rand.Intn(len(methods))],
			Path:         paths[rand.Intn(len(paths))],
			StatusCode:   statuses[rand.Intn(len(statuses))],
			BytesSent:    int64(rand.Intn(50000)),
			ResponseTime: rand.Intn(500),
			UserAgent:    userAgents[rand.Intn(len(userAgents))],
		}
		logs = append(logs, log)
	}
	return logs
}

func (s *Service) generateMockResourceStats(userID string) *models.ResourceStat {
	return &models.ResourceStat{
		UserID:           userID,
		DiskUsedMB:       int64(rand.Intn(5000)),
		DiskLimitMB:      10240,
		InodeUsed:        int64(rand.Intn(100000)),
		InodeLimit:       250000,
		BandwidthUsedMB:  int64(rand.Intn(50000)),
		BandwidthLimitMB: 102400,
		DomainsUsed:      rand.Intn(5),
		DomainsLimit:     10,
		DatabasesUsed:    rand.Intn(3),
		DatabasesLimit:   5,
		CPUUsagePercent:  float64(rand.Intn(80)),
		MemoryUsedMB:     int64(rand.Intn(512)),
		MemoryLimitMB:    1024,
		MeasuredAt:       time.Now(),
	}
}
