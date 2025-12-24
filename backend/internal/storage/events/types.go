// Package events provides immutable event/audit system for OweHost
package events

import "time"

// EventType represents the type of event
type EventType string

// Event types
const (
	// Account events
	EventAccountCreate    EventType = "account.create"
	EventAccountUpdate    EventType = "account.update"
	EventAccountSuspend   EventType = "account.suspend"
	EventAccountUnsuspend EventType = "account.unsuspend"
	EventAccountTerminate EventType = "account.terminate"
	EventAccountDelete    EventType = "account.delete"

	// Domain events
	EventDomainAdd       EventType = "domain.add"
	EventDomainRemove    EventType = "domain.remove"
	EventDomainUpdate    EventType = "domain.update"
	EventSubdomainAdd    EventType = "subdomain.add"
	EventSubdomainRemove EventType = "subdomain.remove"

	// Database events
	EventDatabaseCreate  EventType = "database.create"
	EventDatabaseDelete  EventType = "database.delete"
	EventDatabaseBackup  EventType = "database.backup"
	EventDatabaseRestore EventType = "database.restore"

	// SSL events
	EventSSLInstall  EventType = "ssl.install"
	EventSSLRenew    EventType = "ssl.renew"
	EventSSLRemove   EventType = "ssl.remove"
	EventSSLExpiring EventType = "ssl.expiring"

	// Email events
	EventEmailAccountCreate EventType = "email.account.create"
	EventEmailAccountDelete EventType = "email.account.delete"
	EventEmailForwarderAdd  EventType = "email.forwarder.add"

	// FTP events
	EventFTPAccountCreate EventType = "ftp.account.create"
	EventFTPAccountDelete EventType = "ftp.account.delete"

	// Backup events
	EventBackupStart    EventType = "backup.start"
	EventBackupComplete EventType = "backup.complete"
	EventBackupFailed   EventType = "backup.failed"
	EventRestoreStart   EventType = "restore.start"
	EventRestoreComplete EventType = "restore.complete"

	// Security events
	EventLoginSuccess EventType = "security.login.success"
	EventLoginFailed  EventType = "security.login.failed"
	EventPasswordChange EventType = "security.password.change"
	EventTwoFactorEnable EventType = "security.2fa.enable"
	EventTwoFactorDisable EventType = "security.2fa.disable"
	EventAPIKeyCreate EventType = "security.apikey.create"
	EventAPIKeyRevoke EventType = "security.apikey.revoke"

	// System events
	EventConfigChange   EventType = "system.config.change"
	EventServiceRestart EventType = "system.service.restart"
	EventNodeJoin       EventType = "system.node.join"
	EventNodeLeave      EventType = "system.node.leave"

	// Cron events
	EventCronJobCreate  EventType = "cron.job.create"
	EventCronJobDelete  EventType = "cron.job.delete"
	EventCronJobExecute EventType = "cron.job.execute"
)

// EventResult represents the result of an event
type EventResult string

const (
	ResultSuccess EventResult = "success"
	ResultFailed  EventResult = "failed"
	ResultPending EventResult = "pending"
)

// Event represents an immutable audit event
type Event struct {
	ID        string                 `json:"id"`
	Type      EventType              `json:"type"`
	AccountID int                    `json:"account_id,omitempty"`
	Actor     string                 `json:"actor"`      // user ID, "system", or API key ID
	ActorType string                 `json:"actor_type"` // user, admin, reseller, system, api
	ActorIP   string                 `json:"actor_ip,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
	Result    EventResult            `json:"result"`
	Error     string                 `json:"error,omitempty"`
	Duration  int64                  `json:"duration_ms,omitempty"`
	RequestID string                 `json:"request_id,omitempty"`
	NodeID    string                 `json:"node_id,omitempty"`
}

// EventFilters for querying events
type EventFilters struct {
	AccountID   *int
	Type        *EventType
	Actor       *string
	ActorType   *string
	Result      *EventResult
	StartTime   *time.Time
	EndTime     *time.Time
	RequestID   *string
	Limit       int
	Offset      int
}

// EventStats represents event statistics
type EventStats struct {
	TotalEvents   int64            `json:"total_events"`
	ByType        map[string]int64 `json:"by_type"`
	ByResult      map[string]int64 `json:"by_result"`
	ByActorType   map[string]int64 `json:"by_actor_type"`
	LastEventTime *time.Time       `json:"last_event_time"`
}

// SecurityAlert represents a security-related alert
type SecurityAlert struct {
	ID          string    `json:"id"`
	Type        string    `json:"type"`
	Severity    string    `json:"severity"` // low, medium, high, critical
	AccountID   int       `json:"account_id,omitempty"`
	Description string    `json:"description"`
	Data        map[string]interface{} `json:"data"`
	Timestamp   time.Time `json:"timestamp"`
	Resolved    bool      `json:"resolved"`
	ResolvedAt  *time.Time `json:"resolved_at,omitempty"`
	ResolvedBy  *string    `json:"resolved_by,omitempty"`
}
