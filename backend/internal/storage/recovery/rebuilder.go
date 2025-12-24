// Package recovery provides filesystem scanning and database rebuild functionality
package recovery

import (
	"context"
	"fmt"
	"time"

	"github.com/iSundram/OweHost/internal/storage/events"
)

// Rebuilder rebuilds database state from filesystem
type Rebuilder struct {
	scanner *Scanner
	emitter *events.Emitter
}

// NewRebuilder creates a new database rebuilder
func NewRebuilder() *Rebuilder {
	return &Rebuilder{
		scanner: NewScanner(),
		emitter: events.NewEmitter(),
	}
}

// RebuildResult represents the result of a rebuild operation
type RebuildResult struct {
	StartTime       time.Time           `json:"start_time"`
	EndTime         time.Time           `json:"end_time"`
	Duration        time.Duration       `json:"duration"`
	AccountsScanned int                 `json:"accounts_scanned"`
	AccountsUpdated int                 `json:"accounts_updated"`
	SitesFound      int                 `json:"sites_found"`
	SSLCertsFound   int                 `json:"ssl_certs_found"`
	DatabasesFound  int                 `json:"databases_found"`
	Errors          []RebuildError      `json:"errors,omitempty"`
	Warnings        []string            `json:"warnings,omitempty"`
}

// RebuildError represents an error during rebuild
type RebuildError struct {
	AccountID int    `json:"account_id,omitempty"`
	Resource  string `json:"resource"`
	Error     string `json:"error"`
}

// RebuildOptions configures the rebuild operation
type RebuildOptions struct {
	DryRun          bool   // If true, don't actually update database
	AccountID       *int   // If set, only rebuild this account
	SkipValidation  bool   // Skip validation checks
	ForceOverwrite  bool   // Overwrite existing database entries
	Verbose         bool   // Enable verbose logging
}

// DatabaseWriter interface for writing to database
type DatabaseWriter interface {
	UpsertAccount(ctx context.Context, scan *AccountScan) error
	UpsertDomain(ctx context.Context, accountID int, domain string, site interface{}) error
	UpsertSSLCert(ctx context.Context, accountID int, domain string, meta interface{}) error
	UpsertDatabase(ctx context.Context, accountID int, dbName, dbType string) error
	DeleteStaleRecords(ctx context.Context, accountID int, currentDomains []string) error
}

// Rebuild rebuilds the database from filesystem state
func (r *Rebuilder) Rebuild(ctx context.Context, opts RebuildOptions, writer DatabaseWriter) (*RebuildResult, error) {
	result := &RebuildResult{
		StartTime: time.Now(),
		Errors:    []RebuildError{},
		Warnings:  []string{},
	}

	// Emit rebuild start event
	r.emitter.EmitSuccess(events.EventConfigChange, events.EmitOptions{
		Actor:     "system",
		ActorType: "system",
		Data: map[string]interface{}{
			"action":  "rebuild_start",
			"dry_run": opts.DryRun,
		},
	})

	// Scan filesystem
	var scanResult *ScanResult
	var err error

	if opts.AccountID != nil {
		// Scan single account
		scan, err := r.scanner.ScanAccount(*opts.AccountID)
		if err != nil {
			return nil, fmt.Errorf("failed to scan account %d: %w", *opts.AccountID, err)
		}
		scanResult = &ScanResult{
			Accounts:   []AccountScan{*scan},
			TotalSites: len(scan.Sites),
			TotalSSL:   len(scan.SSLCerts),
			TotalDBs:   len(scan.Databases),
		}
	} else {
		// Scan all accounts
		scanResult, err = r.scanner.ScanAll()
		if err != nil {
			return nil, fmt.Errorf("failed to scan filesystem: %w", err)
		}
	}

	result.AccountsScanned = len(scanResult.Accounts)
	result.SitesFound = scanResult.TotalSites
	result.SSLCertsFound = scanResult.TotalSSL
	result.DatabasesFound = scanResult.TotalDBs

	// Process each account
	for _, scan := range scanResult.Accounts {
		if err := r.rebuildAccount(ctx, &scan, opts, writer, result); err != nil {
			result.Errors = append(result.Errors, RebuildError{
				AccountID: scan.ID,
				Resource:  "account",
				Error:     err.Error(),
			})
		} else {
			result.AccountsUpdated++
		}
	}

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)

	// Emit rebuild complete event
	r.emitter.EmitSuccess(events.EventConfigChange, events.EmitOptions{
		Actor:     "system",
		ActorType: "system",
		Data: map[string]interface{}{
			"action":           "rebuild_complete",
			"accounts_updated": result.AccountsUpdated,
			"duration_ms":      result.Duration.Milliseconds(),
			"errors":           len(result.Errors),
		},
	})

	return result, nil
}

// rebuildAccount rebuilds a single account
func (r *Rebuilder) rebuildAccount(ctx context.Context, scan *AccountScan, opts RebuildOptions, writer DatabaseWriter, result *RebuildResult) error {
	// Validate if not skipped
	if !opts.SkipValidation {
		issues := r.scanner.ValidateIntegrity(scan)
		for _, issue := range issues {
			result.Warnings = append(result.Warnings, fmt.Sprintf("account %d: %s", scan.ID, issue))
		}
	}

	if opts.DryRun {
		return nil
	}

	if writer == nil {
		return nil // No writer provided, just scanning
	}

	// Upsert account
	if err := writer.UpsertAccount(ctx, scan); err != nil {
		return fmt.Errorf("failed to upsert account: %w", err)
	}

	// Upsert domains/sites
	var currentDomains []string
	for _, site := range scan.Sites {
		currentDomains = append(currentDomains, site.Domain)
		if err := writer.UpsertDomain(ctx, scan.ID, site.Domain, site); err != nil {
			result.Errors = append(result.Errors, RebuildError{
				AccountID: scan.ID,
				Resource:  fmt.Sprintf("domain:%s", site.Domain),
				Error:     err.Error(),
			})
		}
	}

	// Upsert SSL certificates
	for _, cert := range scan.SSLCerts {
		if err := writer.UpsertSSLCert(ctx, scan.ID, cert.Domain, cert.Meta); err != nil {
			result.Errors = append(result.Errors, RebuildError{
				AccountID: scan.ID,
				Resource:  fmt.Sprintf("ssl:%s", cert.Domain),
				Error:     err.Error(),
			})
		}
	}

	// Upsert databases
	for _, db := range scan.Databases {
		if err := writer.UpsertDatabase(ctx, scan.ID, db.Name, db.Type); err != nil {
			result.Errors = append(result.Errors, RebuildError{
				AccountID: scan.ID,
				Resource:  fmt.Sprintf("database:%s", db.Name),
				Error:     err.Error(),
			})
		}
	}

	// Clean up stale records
	if err := writer.DeleteStaleRecords(ctx, scan.ID, currentDomains); err != nil {
		result.Warnings = append(result.Warnings, fmt.Sprintf("account %d: failed to delete stale records: %v", scan.ID, err))
	}

	return nil
}

// VerifyConsistency verifies consistency between filesystem and database
func (r *Rebuilder) VerifyConsistency(ctx context.Context, reader DatabaseReader) (*ConsistencyReport, error) {
	report := &ConsistencyReport{
		Timestamp:   time.Now(),
		Mismatches:  []ConsistencyMismatch{},
		MissingInDB: []MissingRecord{},
		MissingInFS: []MissingRecord{},
	}

	// Scan filesystem
	scanResult, err := r.scanner.ScanAll()
	if err != nil {
		return nil, err
	}

	for _, scan := range scanResult.Accounts {
		// Check if account exists in database
		dbAccount, err := reader.GetAccount(ctx, scan.ID)
		if err != nil {
			report.MissingInDB = append(report.MissingInDB, MissingRecord{
				Type:      "account",
				AccountID: scan.ID,
				Name:      scan.Identity.Name,
			})
			continue
		}

		// Compare account data
		if scan.Identity != nil && dbAccount != nil {
			if scan.Identity.State != dbAccount.State {
				report.Mismatches = append(report.Mismatches, ConsistencyMismatch{
					Type:      "account_state",
					AccountID: scan.ID,
					FSValue:   scan.Identity.State,
					DBValue:   dbAccount.State,
				})
			}
		}

		// Check domains
		for _, site := range scan.Sites {
			if !reader.DomainExists(ctx, scan.ID, site.Domain) {
				report.MissingInDB = append(report.MissingInDB, MissingRecord{
					Type:      "domain",
					AccountID: scan.ID,
					Name:      site.Domain,
				})
			}
		}
	}

	report.Consistent = len(report.Mismatches) == 0 && len(report.MissingInDB) == 0 && len(report.MissingInFS) == 0

	return report, nil
}

// DatabaseReader interface for reading from database
type DatabaseReader interface {
	GetAccount(ctx context.Context, accountID int) (*DatabaseAccount, error)
	DomainExists(ctx context.Context, accountID int, domain string) bool
	ListDomains(ctx context.Context, accountID int) ([]string, error)
}

// DatabaseAccount represents account data from database
type DatabaseAccount struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	State string `json:"state"`
}

// ConsistencyReport represents the result of a consistency check
type ConsistencyReport struct {
	Timestamp   time.Time             `json:"timestamp"`
	Consistent  bool                  `json:"consistent"`
	Mismatches  []ConsistencyMismatch `json:"mismatches,omitempty"`
	MissingInDB []MissingRecord       `json:"missing_in_db,omitempty"`
	MissingInFS []MissingRecord       `json:"missing_in_fs,omitempty"`
}

// ConsistencyMismatch represents a data mismatch between filesystem and database
type ConsistencyMismatch struct {
	Type      string `json:"type"`
	AccountID int    `json:"account_id"`
	Resource  string `json:"resource,omitempty"`
	FSValue   string `json:"fs_value"`
	DBValue   string `json:"db_value"`
}

// MissingRecord represents a record missing from filesystem or database
type MissingRecord struct {
	Type      string `json:"type"`
	AccountID int    `json:"account_id"`
	Name      string `json:"name"`
}

// RecoverFromBackup recovers account state from a backup
func (r *Rebuilder) RecoverFromBackup(ctx context.Context, backupPath string, accountID int) error {
	// This would restore files from backup and then rebuild
	// Implementation depends on backup format
	return fmt.Errorf("not implemented")
}
