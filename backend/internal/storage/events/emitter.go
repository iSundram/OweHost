// Package events provides immutable event/audit system for OweHost
package events

import (
	"time"
)

// Emitter provides a convenient API for emitting events
type Emitter struct {
	store  *Store
	nodeID string
}

// NewEmitter creates a new event emitter
func NewEmitter() *Emitter {
	return &Emitter{
		store:  NewStore(),
		nodeID: "node-1", // Default, should be configured
	}
}

// NewEmitterWithStore creates an emitter with a custom store
func NewEmitterWithStore(store *Store, nodeID string) *Emitter {
	return &Emitter{
		store:  store,
		nodeID: nodeID,
	}
}

// EmitOptions contains options for emitting an event
type EmitOptions struct {
	AccountID int
	Actor     string
	ActorType string
	ActorIP   string
	RequestID string
	Data      map[string]interface{}
}

// Emit emits an event with the given options
func (e *Emitter) Emit(eventType EventType, result EventResult, opts EmitOptions) error {
	event := &Event{
		Type:      eventType,
		AccountID: opts.AccountID,
		Actor:     opts.Actor,
		ActorType: opts.ActorType,
		ActorIP:   opts.ActorIP,
		Timestamp: time.Now(),
		Data:      opts.Data,
		Result:    result,
		RequestID: opts.RequestID,
		NodeID:    e.nodeID,
	}

	return e.store.Save(event)
}

// EmitSuccess emits a successful event
func (e *Emitter) EmitSuccess(eventType EventType, opts EmitOptions) error {
	return e.Emit(eventType, ResultSuccess, opts)
}

// EmitFailed emits a failed event
func (e *Emitter) EmitFailed(eventType EventType, errorMsg string, opts EmitOptions) error {
	event := &Event{
		Type:      eventType,
		AccountID: opts.AccountID,
		Actor:     opts.Actor,
		ActorType: opts.ActorType,
		ActorIP:   opts.ActorIP,
		Timestamp: time.Now(),
		Data:      opts.Data,
		Result:    ResultFailed,
		Error:     errorMsg,
		RequestID: opts.RequestID,
		NodeID:    e.nodeID,
	}

	return e.store.Save(event)
}

// Account event helpers

// AccountCreated emits an account creation event
func (e *Emitter) AccountCreated(accountID int, accountName, actor, actorType, actorIP string) error {
	return e.EmitSuccess(EventAccountCreate, EmitOptions{
		AccountID: accountID,
		Actor:     actor,
		ActorType: actorType,
		ActorIP:   actorIP,
		Data: map[string]interface{}{
			"account_name": accountName,
		},
	})
}

// AccountSuspended emits an account suspension event
func (e *Emitter) AccountSuspended(accountID int, reason, actor, actorType string) error {
	return e.EmitSuccess(EventAccountSuspend, EmitOptions{
		AccountID: accountID,
		Actor:     actor,
		ActorType: actorType,
		Data: map[string]interface{}{
			"reason": reason,
		},
	})
}

// AccountUnsuspended emits an account unsuspension event
func (e *Emitter) AccountUnsuspended(accountID int, actor, actorType string) error {
	return e.EmitSuccess(EventAccountUnsuspend, EmitOptions{
		AccountID: accountID,
		Actor:     actor,
		ActorType: actorType,
	})
}

// AccountTerminated emits an account termination event
func (e *Emitter) AccountTerminated(accountID int, reason, actor, actorType string) error {
	return e.EmitSuccess(EventAccountTerminate, EmitOptions{
		AccountID: accountID,
		Actor:     actor,
		ActorType: actorType,
		Data: map[string]interface{}{
			"reason": reason,
		},
	})
}

// Domain event helpers

// DomainAdded emits a domain addition event
func (e *Emitter) DomainAdded(accountID int, domain, actor, actorType string) error {
	return e.EmitSuccess(EventDomainAdd, EmitOptions{
		AccountID: accountID,
		Actor:     actor,
		ActorType: actorType,
		Data: map[string]interface{}{
			"domain": domain,
		},
	})
}

// DomainRemoved emits a domain removal event
func (e *Emitter) DomainRemoved(accountID int, domain, actor, actorType string) error {
	return e.EmitSuccess(EventDomainRemove, EmitOptions{
		AccountID: accountID,
		Actor:     actor,
		ActorType: actorType,
		Data: map[string]interface{}{
			"domain": domain,
		},
	})
}

// Database event helpers

// DatabaseCreated emits a database creation event
func (e *Emitter) DatabaseCreated(accountID int, dbName, dbType, actor, actorType string) error {
	return e.EmitSuccess(EventDatabaseCreate, EmitOptions{
		AccountID: accountID,
		Actor:     actor,
		ActorType: actorType,
		Data: map[string]interface{}{
			"database_name": dbName,
			"database_type": dbType,
		},
	})
}

// DatabaseDeleted emits a database deletion event
func (e *Emitter) DatabaseDeleted(accountID int, dbName, actor, actorType string) error {
	return e.EmitSuccess(EventDatabaseDelete, EmitOptions{
		AccountID: accountID,
		Actor:     actor,
		ActorType: actorType,
		Data: map[string]interface{}{
			"database_name": dbName,
		},
	})
}

// SSL event helpers

// SSLInstalled emits an SSL installation event
func (e *Emitter) SSLInstalled(accountID int, domain, sslType, actor, actorType string) error {
	return e.EmitSuccess(EventSSLInstall, EmitOptions{
		AccountID: accountID,
		Actor:     actor,
		ActorType: actorType,
		Data: map[string]interface{}{
			"domain":   domain,
			"ssl_type": sslType,
		},
	})
}

// SSLRenewed emits an SSL renewal event
func (e *Emitter) SSLRenewed(accountID int, domain, actor, actorType string) error {
	return e.EmitSuccess(EventSSLRenew, EmitOptions{
		AccountID: accountID,
		Actor:     actor,
		ActorType: actorType,
		Data: map[string]interface{}{
			"domain": domain,
		},
	})
}

// Security event helpers

// LoginSuccess emits a successful login event
func (e *Emitter) LoginSuccess(accountID int, username, actorIP string) error {
	return e.EmitSuccess(EventLoginSuccess, EmitOptions{
		AccountID: accountID,
		Actor:     username,
		ActorType: "user",
		ActorIP:   actorIP,
		Data: map[string]interface{}{
			"username": username,
		},
	})
}

// LoginFailed emits a failed login event
func (e *Emitter) LoginFailed(username, actorIP, reason string) error {
	return e.EmitFailed(EventLoginFailed, reason, EmitOptions{
		Actor:     username,
		ActorType: "user",
		ActorIP:   actorIP,
		Data: map[string]interface{}{
			"username": username,
			"reason":   reason,
		},
	})
}

// PasswordChanged emits a password change event
func (e *Emitter) PasswordChanged(accountID int, actor, actorType, actorIP string) error {
	return e.EmitSuccess(EventPasswordChange, EmitOptions{
		AccountID: accountID,
		Actor:     actor,
		ActorType: actorType,
		ActorIP:   actorIP,
	})
}

// Backup event helpers

// BackupStarted emits a backup start event
func (e *Emitter) BackupStarted(accountID int, backupType, actor, actorType string) error {
	return e.EmitSuccess(EventBackupStart, EmitOptions{
		AccountID: accountID,
		Actor:     actor,
		ActorType: actorType,
		Data: map[string]interface{}{
			"backup_type": backupType,
		},
	})
}

// BackupCompleted emits a backup completion event
func (e *Emitter) BackupCompleted(accountID int, backupID string, sizeMB int64, durationSec int, actor, actorType string) error {
	return e.EmitSuccess(EventBackupComplete, EmitOptions{
		AccountID: accountID,
		Actor:     actor,
		ActorType: actorType,
		Data: map[string]interface{}{
			"backup_id":    backupID,
			"size_mb":      sizeMB,
			"duration_sec": durationSec,
		},
	})
}

// BackupFailed emits a backup failure event
func (e *Emitter) BackupFailed(accountID int, errorMsg, actor, actorType string) error {
	return e.EmitFailed(EventBackupFailed, errorMsg, EmitOptions{
		AccountID: accountID,
		Actor:     actor,
		ActorType: actorType,
	})
}

// System event helpers

// ConfigChanged emits a configuration change event
func (e *Emitter) ConfigChanged(component, key, oldValue, newValue, actor, actorType string) error {
	return e.EmitSuccess(EventConfigChange, EmitOptions{
		Actor:     actor,
		ActorType: actorType,
		Data: map[string]interface{}{
			"component": component,
			"key":       key,
			"old_value": oldValue,
			"new_value": newValue,
		},
	})
}

// ServiceRestarted emits a service restart event
func (e *Emitter) ServiceRestarted(serviceName, actor, actorType string) error {
	return e.EmitSuccess(EventServiceRestart, EmitOptions{
		Actor:     actor,
		ActorType: actorType,
		Data: map[string]interface{}{
			"service": serviceName,
		},
	})
}

// Alert helpers

// CreateSecurityAlert creates and saves a security alert
func (e *Emitter) CreateSecurityAlert(alertType, severity, description string, accountID int, data map[string]interface{}) error {
	alert := &SecurityAlert{
		Type:        alertType,
		Severity:    severity,
		AccountID:   accountID,
		Description: description,
		Data:        data,
		Timestamp:   time.Now(),
		Resolved:    false,
	}

	return e.store.SaveAlert(alert)
}
