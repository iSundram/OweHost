package models

import "time"

// Reseller represents a reseller in the system
type Reseller struct {
	ID               string    `json:"id"`
	UserID           string    `json:"user_id"`
	ParentResellerID *string   `json:"parent_reseller_id,omitempty"`
	Name             string    `json:"name"`
	ResourcePool     ResourcePool `json:"resource_pool"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// ResourcePool represents the resource allocation for a reseller
type ResourcePool struct {
	MaxUsers       int   `json:"max_users"`
	MaxDomains     int   `json:"max_domains"`
	MaxDiskMB      int64 `json:"max_disk_mb"`
	MaxBandwidthMB int64 `json:"max_bandwidth_mb"`
	MaxDatabases   int   `json:"max_databases"`
	MaxCPUQuota    int   `json:"max_cpu_quota"`
	MaxMemoryMB    int64 `json:"max_memory_mb"`
}

// ResellerCreateRequest represents a request to create a reseller
type ResellerCreateRequest struct {
	UserID           string       `json:"user_id" validate:"required"`
	ParentResellerID *string      `json:"parent_reseller_id,omitempty"`
	Name             string       `json:"name" validate:"required,min=2,max=64"`
	ResourcePool     ResourcePool `json:"resource_pool" validate:"required"`
}

// ResellerUpdateRequest represents a request to update a reseller
type ResellerUpdateRequest struct {
	Name         *string       `json:"name,omitempty" validate:"omitempty,min=2,max=64"`
	ResourcePool *ResourcePool `json:"resource_pool,omitempty"`
}

// OwnershipNode represents a node in the ownership graph
type OwnershipNode struct {
	ID       string          `json:"id"`
	Type     string          `json:"type"`
	ParentID *string         `json:"parent_id,omitempty"`
	Children []OwnershipNode `json:"children,omitempty"`
}
