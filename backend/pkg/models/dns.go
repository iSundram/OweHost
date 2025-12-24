package models

import "time"

// DNSRecordType represents the type of DNS record
type DNSRecordType string

const (
	DNSRecordTypeA     DNSRecordType = "A"
	DNSRecordTypeAAAA  DNSRecordType = "AAAA"
	DNSRecordTypeCNAME DNSRecordType = "CNAME"
	DNSRecordTypeMX    DNSRecordType = "MX"
	DNSRecordTypeTXT   DNSRecordType = "TXT"
	DNSRecordTypeSRV   DNSRecordType = "SRV"
	DNSRecordTypeNS    DNSRecordType = "NS"
	DNSRecordTypeSOA   DNSRecordType = "SOA"
)

// DNSZone represents a DNS zone
type DNSZone struct {
	ID           string    `json:"id"`
	DomainID     string    `json:"domain_id"`
	Name         string    `json:"name"`
	Locked       bool      `json:"locked"`
	DNSSECEnabled bool     `json:"dnssec_enabled"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// DNSRecord represents a DNS record
type DNSRecord struct {
	ID        string        `json:"id"`
	ZoneID    string        `json:"zone_id"`
	Name      string        `json:"name"`
	Type      DNSRecordType `json:"type"`
	Content   string        `json:"content"`
	TTL       int           `json:"ttl"`
	Priority  *int          `json:"priority,omitempty"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`
}

// DNSSECKey represents a DNSSEC key
type DNSSECKey struct {
	ID         string    `json:"id"`
	ZoneID     string    `json:"zone_id"`
	KeyTag     int       `json:"key_tag"`
	Algorithm  int       `json:"algorithm"`
	DigestType int       `json:"digest_type"`
	Digest     string    `json:"digest"`
	PublicKey  string    `json:"public_key"`
	Active     bool      `json:"active"`
	CreatedAt  time.Time `json:"created_at"`
	RolloverAt *time.Time `json:"rollover_at,omitempty"`
}

// DNSRecordCreateRequest represents a request to create a DNS record
type DNSRecordCreateRequest struct {
	Name     string        `json:"name" validate:"required"`
	Type     DNSRecordType `json:"type" validate:"required,oneof=A AAAA CNAME MX TXT SRV NS"`
	Content  string        `json:"content" validate:"required"`
	TTL      int           `json:"ttl" validate:"min=60,max=86400"`
	Priority *int          `json:"priority,omitempty"`
}

// DNSSyncState represents the sync state with external DNS providers
type DNSSyncState struct {
	ZoneID       string    `json:"zone_id"`
	ProviderName string    `json:"provider_name"`
	LastSyncAt   time.Time `json:"last_sync_at"`
	SyncStatus   string    `json:"sync_status"`
	ErrorMessage *string   `json:"error_message,omitempty"`
}
