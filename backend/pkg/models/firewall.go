package models

import "time"

// FirewallAction represents the action of a firewall rule
type FirewallAction string

const (
	FirewallActionAllow  FirewallAction = "allow"
	FirewallActionDeny   FirewallAction = "deny"
	FirewallActionReject FirewallAction = "reject"
)

// FirewallProtocol represents the protocol of a firewall rule
type FirewallProtocol string

const (
	FirewallProtocolTCP  FirewallProtocol = "tcp"
	FirewallProtocolUDP  FirewallProtocol = "udp"
	FirewallProtocolICMP FirewallProtocol = "icmp"
	FirewallProtocolAny  FirewallProtocol = "any"
)

// FirewallRule represents a firewall rule
type FirewallRule struct {
	ID          string           `json:"id"`
	UserID      *string          `json:"user_id,omitempty"`
	ChainName   string           `json:"chain_name"`
	Priority    int              `json:"priority"`
	Action      FirewallAction   `json:"action"`
	Protocol    FirewallProtocol `json:"protocol"`
	SourceIP    string           `json:"source_ip"`
	DestIP      string           `json:"dest_ip"`
	SourcePort  *int             `json:"source_port,omitempty"`
	DestPort    *int             `json:"dest_port,omitempty"`
	Description string           `json:"description"`
	Enabled     bool             `json:"enabled"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
}

// FirewallChain represents a firewall chain
type FirewallChain struct {
	Name        string    `json:"name"`
	Policy      FirewallAction `json:"policy"`
	RulesCount  int       `json:"rules_count"`
	CreatedAt   time.Time `json:"created_at"`
}

// RateLimit represents a rate limit configuration
type RateLimit struct {
	ID            string    `json:"id"`
	UserID        *string   `json:"user_id,omitempty"`
	Type          string    `json:"type"`
	IPAddress     *string   `json:"ip_address,omitempty"`
	RequestsPerSecond int   `json:"requests_per_second"`
	BurstSize     int       `json:"burst_size"`
	Enabled       bool      `json:"enabled"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// IntrusionEvent represents an intrusion detection event
type IntrusionEvent struct {
	ID          string    `json:"id"`
	EventType   string    `json:"event_type"`
	Severity    string    `json:"severity"`
	SourceIP    string    `json:"source_ip"`
	TargetIP    string    `json:"target_ip"`
	Description string    `json:"description"`
	RawData     string    `json:"raw_data"`
	DetectedAt  time.Time `json:"detected_at"`
}

// FirewallRuleCreateRequest represents a request to create a firewall rule
type FirewallRuleCreateRequest struct {
	ChainName   string           `json:"chain_name" validate:"required"`
	Priority    int              `json:"priority"`
	Action      FirewallAction   `json:"action" validate:"required,oneof=allow deny reject"`
	Protocol    FirewallProtocol `json:"protocol" validate:"required,oneof=tcp udp icmp any"`
	SourceIP    string           `json:"source_ip" validate:"required"`
	DestIP      string           `json:"dest_ip" validate:"required"`
	SourcePort  *int             `json:"source_port,omitempty"`
	DestPort    *int             `json:"dest_port,omitempty"`
	Description string           `json:"description,omitempty"`
}

// RateLimitCreateRequest represents a request to create a rate limit
type RateLimitCreateRequest struct {
	Type              string  `json:"type" validate:"required,oneof=ip user global"`
	IPAddress         *string `json:"ip_address,omitempty"`
	RequestsPerSecond int     `json:"requests_per_second" validate:"required,min=1"`
	BurstSize         int     `json:"burst_size" validate:"min=1"`
}
