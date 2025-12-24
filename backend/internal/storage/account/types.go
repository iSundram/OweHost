// Package account provides filesystem-based account state management
package account

import "time"

const (
	// BaseAccountPath is the root directory for all accounts
	BaseAccountPath = "/srv/accounts"
	// AccountPrefix is the prefix for account directories
	AccountPrefix = "a-"
	// OweHostBasePath is the base path for OweHost system files
	OweHostBasePath = "/opt/owehost"
)

// AccountIdentity represents account.json - the core account identity
type AccountIdentity struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	UID       int    `json:"uid"`
	GID       int    `json:"gid"`
	Owner     string `json:"owner"`      // Parent owner (e.g., "reseller-22", "admin")
	Plan      string `json:"plan"`       // Package plan (e.g., "starter", "standard", "premium")
	Node      string `json:"node"`       // Assigned node (e.g., "node-1")
	CreatedAt string `json:"created_at"` // RFC3339 timestamp
	State     string `json:"state"`      // active, suspended, terminated, pending
}

// ResourceLimits represents limits.json - resource quotas for the account
type ResourceLimits struct {
	DiskMB       int `json:"disk_mb"`
	CPUPercent   int `json:"cpu_percent"`
	RAMMB        int `json:"ram_mb"`
	Databases    int `json:"databases"`
	Domains      int `json:"domains"`
	Subdomains   int `json:"subdomains"`
	EmailAccounts int `json:"email_accounts"`
	FTPAccounts  int `json:"ftp_accounts"`
	Bandwidth    int `json:"bandwidth_gb"` // Monthly bandwidth in GB
	Inodes       int `json:"inodes"`
}

// AccountStatus represents status.json - current account state
type AccountStatus struct {
	Suspended    bool    `json:"suspended"`
	Locked       bool    `json:"locked"`
	Reason       *string `json:"reason"`
	SuspendedAt  *string `json:"suspended_at,omitempty"`
	SuspendedBy  *string `json:"suspended_by,omitempty"`
	LockedAt     *string `json:"locked_at,omitempty"`
	LockedReason *string `json:"locked_reason,omitempty"`
}

// AccountMetadata represents additional account metadata
type AccountMetadata struct {
	Email        string            `json:"email"`
	ContactName  string            `json:"contact_name,omitempty"`
	ContactPhone string            `json:"contact_phone,omitempty"`
	Notes        string            `json:"notes,omitempty"`
	Tags         []string          `json:"tags,omitempty"`
	Custom       map[string]string `json:"custom,omitempty"`
	UpdatedAt    string            `json:"updated_at"`
}

// AccountUsage represents current resource usage
type AccountUsage struct {
	DiskUsedMB    int       `json:"disk_used_mb"`
	BandwidthUsed int       `json:"bandwidth_used_gb"`
	DatabaseCount int       `json:"database_count"`
	DomainCount   int       `json:"domain_count"`
	EmailCount    int       `json:"email_count"`
	InodeCount    int       `json:"inode_count"`
	LastUpdated   time.Time `json:"last_updated"`
}

// Account represents the complete account state
type Account struct {
	Identity *AccountIdentity  `json:"identity"`
	Limits   *ResourceLimits   `json:"limits"`
	Status   *AccountStatus    `json:"status"`
	Metadata *AccountMetadata  `json:"metadata,omitempty"`
	Usage    *AccountUsage     `json:"usage,omitempty"`
}

// ApplyConfig represents the desired state to apply
type ApplyConfig struct {
	Identity *AccountIdentity
	Limits   *ResourceLimits
	Status   *AccountStatus
	Metadata *AccountMetadata
}

// AccountState constants
const (
	StateActive     = "active"
	StateSuspended  = "suspended"
	StateTerminated = "terminated"
	StatePending    = "pending"
)

// Plan presets
var PlanLimits = map[string]ResourceLimits{
	"starter": {
		DiskMB:        5120,  // 5GB
		CPUPercent:    50,
		RAMMB:         512,
		Databases:     3,
		Domains:       5,
		Subdomains:    10,
		EmailAccounts: 10,
		FTPAccounts:   3,
		Bandwidth:     50,
		Inodes:        100000,
	},
	"standard": {
		DiskMB:        10240, // 10GB
		CPUPercent:    100,
		RAMMB:         2048,
		Databases:     10,
		Domains:       20,
		Subdomains:    50,
		EmailAccounts: 50,
		FTPAccounts:   10,
		Bandwidth:     200,
		Inodes:        250000,
	},
	"premium": {
		DiskMB:        51200, // 50GB
		CPUPercent:    200,
		RAMMB:         4096,
		Databases:     50,
		Domains:       100,
		Subdomains:    200,
		EmailAccounts: 200,
		FTPAccounts:   50,
		Bandwidth:     1000,
		Inodes:        500000,
	},
	"enterprise": {
		DiskMB:        102400, // 100GB
		CPUPercent:    400,
		RAMMB:         8192,
		Databases:     -1, // Unlimited
		Domains:       -1,
		Subdomains:    -1,
		EmailAccounts: -1,
		FTPAccounts:   -1,
		Bandwidth:     -1,
		Inodes:        -1,
	},
}

// GetPlanLimits returns the resource limits for a plan
func GetPlanLimits(plan string) ResourceLimits {
	if limits, ok := PlanLimits[plan]; ok {
		return limits
	}
	return PlanLimits["starter"]
}
