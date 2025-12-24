package models

import "time"

// NodeStatus represents the status of a cluster node
type NodeStatus string

const (
	NodeStatusOnline      NodeStatus = "online"
	NodeStatusOffline     NodeStatus = "offline"
	NodeStatusMaintenance NodeStatus = "maintenance"
	NodeStatusDraining    NodeStatus = "draining"
)

// Node represents a cluster node
type Node struct {
	ID            string                 `json:"id"`
	Name          string                 `json:"name"`
	Hostname      string                 `json:"hostname"`
	IPAddress     string                 `json:"ip_address"`
	Status        NodeStatus             `json:"status"`
	Version       string                 `json:"version"`
	Capabilities  []string               `json:"capabilities"`
	Resources     NodeResources          `json:"resources"`
	Metadata      map[string]interface{} `json:"metadata"`
	LastHeartbeat time.Time              `json:"last_heartbeat"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
}

// NodeResources represents the resources of a node
type NodeResources struct {
	CPUCores      int   `json:"cpu_cores"`
	MemoryMB      int64 `json:"memory_mb"`
	DiskMB        int64 `json:"disk_mb"`
	CPUUsage      float64 `json:"cpu_usage"`
	MemoryUsage   float64 `json:"memory_usage"`
	DiskUsage     float64 `json:"disk_usage"`
	UserCount     int   `json:"user_count"`
	DomainCount   int   `json:"domain_count"`
}

// NodeHeartbeat represents a heartbeat from a node
type NodeHeartbeat struct {
	NodeID      string        `json:"node_id"`
	Timestamp   time.Time     `json:"timestamp"`
	Status      NodeStatus    `json:"status"`
	Resources   NodeResources `json:"resources"`
	Version     string        `json:"version"`
	Errors      []string      `json:"errors,omitempty"`
}

// CloudProvider represents a cloud provider abstraction
type CloudProvider struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Type         string                 `json:"type"`
	Credentials  map[string]string      `json:"-"`
	Region       string                 `json:"region"`
	Enabled      bool                   `json:"enabled"`
	Settings     map[string]interface{} `json:"settings"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
}

// VMLifecycleAction represents a VM lifecycle action
type VMLifecycleAction string

const (
	VMActionCreate  VMLifecycleAction = "create"
	VMActionStart   VMLifecycleAction = "start"
	VMActionStop    VMLifecycleAction = "stop"
	VMActionRestart VMLifecycleAction = "restart"
	VMActionDestroy VMLifecycleAction = "destroy"
)

// VMInstance represents a cloud VM instance
type VMInstance struct {
	ID           string            `json:"id"`
	ProviderID   string            `json:"provider_id"`
	Name         string            `json:"name"`
	ExternalID   string            `json:"external_id"`
	Status       string            `json:"status"`
	IPAddress    string            `json:"ip_address"`
	InstanceType string            `json:"instance_type"`
	Region       string            `json:"region"`
	Zone         string            `json:"zone"`
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
}

// PlacementRequest represents a request to place a resource on a node
type PlacementRequest struct {
	ResourceType string            `json:"resource_type" validate:"required"`
	ResourceID   string            `json:"resource_id" validate:"required"`
	Requirements map[string]int64  `json:"requirements"`
	Preferences  map[string]string `json:"preferences,omitempty"`
}

// PlacementResult represents the result of a placement decision
type PlacementResult struct {
	NodeID     string `json:"node_id"`
	Score      int    `json:"score"`
	Reason     string `json:"reason"`
}

// NodeRegisterRequest represents a request to register a node
type NodeRegisterRequest struct {
	Name         string                 `json:"name" validate:"required"`
	Hostname     string                 `json:"hostname" validate:"required"`
	IPAddress    string                 `json:"ip_address" validate:"required,ip"`
	Capabilities []string               `json:"capabilities,omitempty"`
	Resources    NodeResources          `json:"resources" validate:"required"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}
