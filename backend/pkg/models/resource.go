package models

import "time"

// ResourceQuota represents resource quotas for a user
type ResourceQuota struct {
	ID              string    `json:"id"`
	UserID          string    `json:"user_id"`
	CPUQuota        int       `json:"cpu_quota"`
	CPUBurstEnabled bool      `json:"cpu_burst_enabled"`
	MemoryLimitMB   int64     `json:"memory_limit_mb"`
	SwapLimitMB     int64     `json:"swap_limit_mb"`
	DiskQuotaMB     int64     `json:"disk_quota_mb"`
	InodeLimit      int64     `json:"inode_limit"`
	IOReadBps       int64     `json:"io_read_bps"`
	IOWriteBps      int64     `json:"io_write_bps"`
	MaxProcesses    int       `json:"max_processes"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// ResourceUsage represents current resource usage
type ResourceUsage struct {
	UserID          string    `json:"user_id"`
	CPUUsage        float64   `json:"cpu_usage"`
	MemoryUsageMB   int64     `json:"memory_usage_mb"`
	SwapUsageMB     int64     `json:"swap_usage_mb"`
	DiskUsageMB     int64     `json:"disk_usage_mb"`
	InodeUsage      int64     `json:"inode_usage"`
	ProcessCount    int       `json:"process_count"`
	MeasuredAt      time.Time `json:"measured_at"`
}

// CgroupConfig represents cgroup configuration
type CgroupConfig struct {
	UserID       string `json:"user_id"`
	CgroupPath   string `json:"cgroup_path"`
	CPUPeriod    int    `json:"cpu_period"`
	CPUQuota     int    `json:"cpu_quota"`
	MemoryLimit  int64  `json:"memory_limit"`
	SwapLimit    int64  `json:"swap_limit"`
	PidsMax      int    `json:"pids_max"`
}

// ResourceQuotaUpdateRequest represents a request to update resource quotas
type ResourceQuotaUpdateRequest struct {
	CPUQuota        *int   `json:"cpu_quota,omitempty"`
	CPUBurstEnabled *bool  `json:"cpu_burst_enabled,omitempty"`
	MemoryLimitMB   *int64 `json:"memory_limit_mb,omitempty"`
	SwapLimitMB     *int64 `json:"swap_limit_mb,omitempty"`
	DiskQuotaMB     *int64 `json:"disk_quota_mb,omitempty"`
	InodeLimit      *int64 `json:"inode_limit,omitempty"`
	IOReadBps       *int64 `json:"io_read_bps,omitempty"`
	IOWriteBps      *int64 `json:"io_write_bps,omitempty"`
	MaxProcesses    *int   `json:"max_processes,omitempty"`
}
