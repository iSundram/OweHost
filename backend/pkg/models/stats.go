package models

import "time"

// BandwidthStat represents bandwidth usage statistics
type BandwidthStat struct {
	ID       string    `json:"id"`
	DomainID string    `json:"domain_id"`
	Date     time.Time `json:"date"`
	BytesIn  int64     `json:"bytes_in"`
	BytesOut int64     `json:"bytes_out"`
	Requests int64     `json:"requests"`
}

// VisitorStat represents visitor statistics
type VisitorStat struct {
	ID             string    `json:"id"`
	DomainID       string    `json:"domain_id"`
	Date           time.Time `json:"date"`
	UniqueVisitors int       `json:"unique_visitors"`
	PageViews      int       `json:"page_views"`
	Hits           int       `json:"hits"`
	BounceRate     float64   `json:"bounce_rate"`
	AvgSessionTime int       `json:"avg_session_time"`
}

// ErrorLogEntry represents an error log entry
type ErrorLogEntry struct {
	ID        string    `json:"id"`
	DomainID  string    `json:"domain_id"`
	Timestamp time.Time `json:"timestamp"`
	Level     string    `json:"level"`
	Message   string    `json:"message"`
	File      string    `json:"file"`
	Line      int       `json:"line"`
}

// AccessLogEntry represents an access log entry
type AccessLogEntry struct {
	ID           string    `json:"id"`
	DomainID     string    `json:"domain_id"`
	Timestamp    time.Time `json:"timestamp"`
	IPAddress    string    `json:"ip_address"`
	Method       string    `json:"method"`
	Path         string    `json:"path"`
	StatusCode   int       `json:"status_code"`
	BytesSent    int64     `json:"bytes_sent"`
	ResponseTime int       `json:"response_time"`
	UserAgent    string    `json:"user_agent"`
	Referer      string    `json:"referer,omitempty"`
}

// ResourceStat represents resource usage statistics
type ResourceStat struct {
	UserID           string    `json:"user_id"`
	DiskUsedMB       int64     `json:"disk_used_mb"`
	DiskLimitMB      int64     `json:"disk_limit_mb"`
	InodeUsed        int64     `json:"inode_used"`
	InodeLimit       int64     `json:"inode_limit"`
	BandwidthUsedMB  int64     `json:"bandwidth_used_mb"`
	BandwidthLimitMB int64     `json:"bandwidth_limit_mb"`
	DomainsUsed      int       `json:"domains_used"`
	DomainsLimit     int       `json:"domains_limit"`
	DatabasesUsed    int       `json:"databases_used"`
	DatabasesLimit   int       `json:"databases_limit"`
	CPUUsagePercent  float64   `json:"cpu_usage_percent"`
	MemoryUsedMB     int64     `json:"memory_used_mb"`
	MemoryLimitMB    int64     `json:"memory_limit_mb"`
	MeasuredAt       time.Time `json:"measured_at"`
}

// DomainStatsSummary represents a summary of domain statistics
type DomainStatsSummary struct {
	DomainID        string `json:"domain_id"`
	Period          string `json:"period"`
	TotalBandwidth  int64  `json:"total_bandwidth"`
	TotalVisitors   int64  `json:"total_visitors"`
	TotalPageViews  int64  `json:"total_page_views"`
	TotalHits       int64  `json:"total_hits"`
	AverageLoadTime int    `json:"average_load_time"`
	ErrorCount      int    `json:"error_count"`
}

// UserStatsSummary represents a summary of user statistics
type UserStatsSummary struct {
	UserID           string  `json:"user_id"`
	DiskUsedMB       int64   `json:"disk_used_mb"`
	DiskLimitMB      int64   `json:"disk_limit_mb"`
	DiskPercent      float64 `json:"disk_percent"`
	BandwidthUsedMB  int64   `json:"bandwidth_used_mb"`
	BandwidthLimitMB int64   `json:"bandwidth_limit_mb"`
	BandwidthPercent float64 `json:"bandwidth_percent"`
	InodeUsed        int64   `json:"inode_used"`
	InodeLimit       int64   `json:"inode_limit"`
	DomainsUsed      int     `json:"domains_used"`
	DomainsLimit     int     `json:"domains_limit"`
	DatabasesUsed    int     `json:"databases_used"`
	DatabasesLimit   int     `json:"databases_limit"`
}

// TopPage represents a top visited page
type TopPage struct {
	Path      string `json:"path"`
	Views     int    `json:"views"`
	UniqueIPs int    `json:"unique_ips"`
}

// TopReferrer represents a top referrer
type TopReferrer struct {
	Referrer string `json:"referrer"`
	Count    int    `json:"count"`
}

// GeoStat represents geographical statistics
type GeoStat struct {
	Country   string `json:"country"`
	CountryCode string `json:"country_code"`
	Visitors  int    `json:"visitors"`
	Percent   float64 `json:"percent"`
}

// BrowserStat represents browser statistics
type BrowserStat struct {
	Browser string  `json:"browser"`
	Version string  `json:"version"`
	Count   int     `json:"count"`
	Percent float64 `json:"percent"`
}

// OSStat represents operating system statistics
type OSStat struct {
	OS      string  `json:"os"`
	Version string  `json:"version"`
	Count   int     `json:"count"`
	Percent float64 `json:"percent"`
}
