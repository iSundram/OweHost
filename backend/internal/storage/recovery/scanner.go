// Package recovery provides filesystem scanning and database rebuild functionality
package recovery

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/iSundram/OweHost/internal/storage/account"
	"github.com/iSundram/OweHost/internal/storage/web"
)

// Scanner scans the filesystem for account and resource state
type Scanner struct {
	accountState *account.StateManager
	webState     *web.StateManager
}

// NewScanner creates a new filesystem scanner
func NewScanner() *Scanner {
	return &Scanner{
		accountState: account.NewStateManager(),
		webState:     web.NewStateManager(),
	}
}

// ScanResult represents the complete result of scanning the filesystem
type ScanResult struct {
	Accounts   []AccountScan   `json:"accounts"`
	TotalSites int             `json:"total_sites"`
	TotalSSL   int             `json:"total_ssl"`
	TotalDBs   int             `json:"total_databases"`
	Errors     []ScanError     `json:"errors,omitempty"`
}

// AccountScan represents a scanned account with all its resources
type AccountScan struct {
	ID          int                      `json:"id"`
	Path        string                   `json:"path"`
	Identity    *account.AccountIdentity `json:"identity,omitempty"`
	Limits      *account.ResourceLimits  `json:"limits,omitempty"`
	Status      *account.AccountStatus   `json:"status,omitempty"`
	Sites       []web.SiteDescriptor     `json:"sites,omitempty"`
	SSLCerts    []SSLCertScan            `json:"ssl_certs,omitempty"`
	Databases   []DatabaseScan           `json:"databases,omitempty"`
	CronJobs    []CronJobScan            `json:"cron_jobs,omitempty"`
	HasErrors   bool                     `json:"has_errors"`
	ScanErrors  []string                 `json:"scan_errors,omitempty"`
}

// ScanError represents an error encountered during scanning
type ScanError struct {
	AccountID int    `json:"account_id,omitempty"`
	Resource  string `json:"resource"`
	Path      string `json:"path"`
	Error     string `json:"error"`
}

// SSLCertScan represents a scanned SSL certificate
type SSLCertScan struct {
	Domain    string        `json:"domain"`
	Meta      *web.SSLMeta  `json:"meta,omitempty"`
	HasCert   bool          `json:"has_cert"`
	HasKey    bool          `json:"has_key"`
	HasChain  bool          `json:"has_chain"`
}

// DatabaseScan represents a scanned database
type DatabaseScan struct {
	Name   string `json:"name"`
	Type   string `json:"type"` // mysql, postgres
	Path   string `json:"path"`
}

// CronJobScan represents a scanned cron job
type CronJobScan struct {
	ID       string `json:"id"`
	Schedule string `json:"schedule"`
	Command  string `json:"command"`
	Enabled  bool   `json:"enabled"`
}

// ScanAll performs a complete scan of all accounts
func (s *Scanner) ScanAll() (*ScanResult, error) {
	result := &ScanResult{
		Accounts: []AccountScan{},
		Errors:   []ScanError{},
	}

	// Get all account IDs
	accountIDs, err := s.accountState.ListAccounts()
	if err != nil {
		return nil, fmt.Errorf("failed to list accounts: %w", err)
	}

	for _, accountID := range accountIDs {
		scan, err := s.ScanAccount(accountID)
		if err != nil {
			result.Errors = append(result.Errors, ScanError{
				AccountID: accountID,
				Resource:  "account",
				Error:     err.Error(),
			})
			continue
		}

		result.Accounts = append(result.Accounts, *scan)
		result.TotalSites += len(scan.Sites)
		result.TotalSSL += len(scan.SSLCerts)
		result.TotalDBs += len(scan.Databases)
	}

	return result, nil
}

// ScanAccount scans a single account
func (s *Scanner) ScanAccount(accountID int) (*AccountScan, error) {
	scan := &AccountScan{
		ID:   accountID,
		Path: s.accountState.AccountPath(accountID),
	}

	// Read identity
	identity, err := s.accountState.ReadIdentity(accountID)
	if err != nil {
		scan.HasErrors = true
		scan.ScanErrors = append(scan.ScanErrors, fmt.Sprintf("identity: %v", err))
	} else {
		scan.Identity = identity
	}

	// Read limits
	limits, err := s.accountState.ReadLimits(accountID)
	if err == nil {
		scan.Limits = limits
	}

	// Read status
	status, err := s.accountState.ReadStatus(accountID)
	if err == nil {
		scan.Status = status
	}

	// Scan sites
	sites, err := s.webState.ListSites(accountID)
	if err == nil {
		scan.Sites = sites
	}

	// Scan SSL certificates
	scan.SSLCerts = s.scanSSLCerts(accountID)

	// Scan databases
	scan.Databases = s.scanDatabases(accountID)

	// Scan cron jobs
	scan.CronJobs = s.scanCronJobs(accountID)

	return scan, nil
}

// scanSSLCerts scans SSL certificates for an account
func (s *Scanner) scanSSLCerts(accountID int) []SSLCertScan {
	sslPath := filepath.Join(
		account.BaseAccountPath,
		fmt.Sprintf("%s%d", account.AccountPrefix, accountID),
		"ssl",
	)

	entries, err := os.ReadDir(sslPath)
	if err != nil {
		return nil
	}

	var certs []SSLCertScan
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		domain := entry.Name()
		domainPath := filepath.Join(sslPath, domain)

		cert := SSLCertScan{
			Domain:   domain,
			HasCert:  fileExists(filepath.Join(domainPath, "cert.pem")),
			HasKey:   fileExists(filepath.Join(domainPath, "key.pem")),
			HasChain: fileExists(filepath.Join(domainPath, "chain.pem")),
		}

		// Read metadata
		metaPath := filepath.Join(domainPath, "meta.json")
		if data, err := os.ReadFile(metaPath); err == nil {
			var meta web.SSLMeta
			if json.Unmarshal(data, &meta) == nil {
				cert.Meta = &meta
			}
		}

		certs = append(certs, cert)
	}

	return certs
}

// scanDatabases scans databases for an account
func (s *Scanner) scanDatabases(accountID int) []DatabaseScan {
	dbPath := filepath.Join(
		account.BaseAccountPath,
		fmt.Sprintf("%s%d", account.AccountPrefix, accountID),
		"databases",
	)

	var databases []DatabaseScan

	// Scan MySQL databases
	mysqlPath := filepath.Join(dbPath, "mysql")
	if entries, err := os.ReadDir(mysqlPath); err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				databases = append(databases, DatabaseScan{
					Name: entry.Name(),
					Type: "mysql",
					Path: filepath.Join(mysqlPath, entry.Name()),
				})
			}
		}
	}

	// Scan PostgreSQL databases
	pgPath := filepath.Join(dbPath, "postgres")
	if entries, err := os.ReadDir(pgPath); err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				databases = append(databases, DatabaseScan{
					Name: entry.Name(),
					Type: "postgres",
					Path: filepath.Join(pgPath, entry.Name()),
				})
			}
		}
	}

	return databases
}

// scanCronJobs scans cron jobs for an account
func (s *Scanner) scanCronJobs(accountID int) []CronJobScan {
	cronPath := filepath.Join(
		account.BaseAccountPath,
		fmt.Sprintf("%s%d", account.AccountPrefix, accountID),
		"cron",
	)

	entries, err := os.ReadDir(cronPath)
	if err != nil {
		return nil
	}

	var jobs []CronJobScan
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		data, err := os.ReadFile(filepath.Join(cronPath, entry.Name()))
		if err != nil {
			continue
		}

		var job CronJobScan
		if json.Unmarshal(data, &job) == nil {
			jobs = append(jobs, job)
		}
	}

	return jobs
}

// ValidateIntegrity validates the integrity of scanned data
func (s *Scanner) ValidateIntegrity(scan *AccountScan) []string {
	var issues []string

	// Check identity
	if scan.Identity == nil {
		issues = append(issues, "missing account.json")
	} else {
		if err := account.ValidateIdentity(scan.Identity); err != nil {
			issues = append(issues, fmt.Sprintf("invalid identity: %v", err))
		}
	}

	// Check limits
	if scan.Limits == nil {
		issues = append(issues, "missing limits.json")
	}

	// Check sites
	for _, site := range scan.Sites {
		if err := web.ValidateSite(&site); err != nil {
			issues = append(issues, fmt.Sprintf("invalid site %s: %v", site.Domain, err))
		}
	}

	// Check SSL certs
	for _, cert := range scan.SSLCerts {
		if !cert.HasCert || !cert.HasKey {
			issues = append(issues, fmt.Sprintf("incomplete SSL for %s", cert.Domain))
		}
	}

	return issues
}

// fileExists checks if a file exists
func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

// ParseAccountID extracts account ID from a directory name
func ParseAccountID(dirName string) (int, error) {
	if !strings.HasPrefix(dirName, account.AccountPrefix) {
		return 0, fmt.Errorf("invalid account directory: %s", dirName)
	}

	idStr := strings.TrimPrefix(dirName, account.AccountPrefix)
	return strconv.Atoi(idStr)
}
