// Package account provides filesystem-based account state management
package account

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

// StateManager handles filesystem state for accounts
type StateManager struct {
	basePath string
	mu       sync.RWMutex
}

// NewStateManager creates a new account state manager
func NewStateManager() *StateManager {
	return &StateManager{basePath: BaseAccountPath}
}

// NewStateManagerWithPath creates a state manager with custom base path
func NewStateManagerWithPath(basePath string) *StateManager {
	return &StateManager{basePath: basePath}
}

// AccountPath returns the full path for an account directory
func (s *StateManager) AccountPath(accountID int) string {
	return filepath.Join(s.basePath, fmt.Sprintf("%s%d", AccountPrefix, accountID))
}

// Exists checks if an account exists
func (s *StateManager) Exists(accountID int) bool {
	path := s.AccountPath(accountID)
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

// ReadIdentity reads account.json
func (s *StateManager) ReadIdentity(accountID int) (*AccountIdentity, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	path := filepath.Join(s.AccountPath(accountID), "account.json")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("account %d not found", accountID)
		}
		return nil, fmt.Errorf("failed to read account.json: %w", err)
	}

	var identity AccountIdentity
	if err := json.Unmarshal(data, &identity); err != nil {
		return nil, fmt.Errorf("failed to parse account.json: %w", err)
	}

	return &identity, nil
}

// WriteIdentity writes account.json atomically
func (s *StateManager) WriteIdentity(accountID int, identity *AccountIdentity) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	path := filepath.Join(s.AccountPath(accountID), "account.json")
	return s.atomicWrite(path, identity)
}

// ReadLimits reads limits.json
func (s *StateManager) ReadLimits(accountID int) (*ResourceLimits, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	path := filepath.Join(s.AccountPath(accountID), "limits.json")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("limits.json not found for account %d", accountID)
		}
		return nil, fmt.Errorf("failed to read limits.json: %w", err)
	}

	var limits ResourceLimits
	if err := json.Unmarshal(data, &limits); err != nil {
		return nil, fmt.Errorf("failed to parse limits.json: %w", err)
	}

	return &limits, nil
}

// WriteLimits writes limits.json atomically
func (s *StateManager) WriteLimits(accountID int, limits *ResourceLimits) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	path := filepath.Join(s.AccountPath(accountID), "limits.json")
	return s.atomicWrite(path, limits)
}

// ReadStatus reads status.json
func (s *StateManager) ReadStatus(accountID int) (*AccountStatus, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	path := filepath.Join(s.AccountPath(accountID), "status.json")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Return default status if file doesn't exist
			return &AccountStatus{Suspended: false, Locked: false}, nil
		}
		return nil, fmt.Errorf("failed to read status.json: %w", err)
	}

	var status AccountStatus
	if err := json.Unmarshal(data, &status); err != nil {
		return nil, fmt.Errorf("failed to parse status.json: %w", err)
	}

	return &status, nil
}

// WriteStatus writes status.json atomically
func (s *StateManager) WriteStatus(accountID int, status *AccountStatus) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	path := filepath.Join(s.AccountPath(accountID), "status.json")
	return s.atomicWrite(path, status)
}

// ReadMetadata reads metadata.json
func (s *StateManager) ReadMetadata(accountID int) (*AccountMetadata, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	path := filepath.Join(s.AccountPath(accountID), "metadata.json")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // Metadata is optional
		}
		return nil, fmt.Errorf("failed to read metadata.json: %w", err)
	}

	var metadata AccountMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, fmt.Errorf("failed to parse metadata.json: %w", err)
	}

	return &metadata, nil
}

// WriteMetadata writes metadata.json atomically
func (s *StateManager) WriteMetadata(accountID int, metadata *AccountMetadata) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	path := filepath.Join(s.AccountPath(accountID), "metadata.json")
	return s.atomicWrite(path, metadata)
}

// ReadAccount reads the complete account state
func (s *StateManager) ReadAccount(accountID int) (*Account, error) {
	identity, err := s.ReadIdentity(accountID)
	if err != nil {
		return nil, err
	}

	limits, _ := s.ReadLimits(accountID)
	status, _ := s.ReadStatus(accountID)
	metadata, _ := s.ReadMetadata(accountID)

	return &Account{
		Identity: identity,
		Limits:   limits,
		Status:   status,
		Metadata: metadata,
	}, nil
}

// CreateAccountStructure creates the full account directory structure
func (s *StateManager) CreateAccountStructure(accountID int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	base := s.AccountPath(accountID)
	dirs := []string{
		base,
		filepath.Join(base, "home"),
		filepath.Join(base, "web"),
		filepath.Join(base, "mail"),
		filepath.Join(base, "databases"),
		filepath.Join(base, "databases", "mysql"),
		filepath.Join(base, "databases", "postgres"),
		filepath.Join(base, "dns"),
		filepath.Join(base, "ssl"),
		filepath.Join(base, "cron"),
		filepath.Join(base, "runtime"),
		filepath.Join(base, "backups"),
		filepath.Join(base, "logs"),
		filepath.Join(base, "tmp"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return nil
}

// DeleteAccountStructure removes the entire account directory
func (s *StateManager) DeleteAccountStructure(accountID int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	path := s.AccountPath(accountID)
	return os.RemoveAll(path)
}

// ListAccounts returns all account IDs by scanning the base directory
func (s *StateManager) ListAccounts() ([]int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entries, err := os.ReadDir(s.basePath)
	if err != nil {
		if os.IsNotExist(err) {
			return []int{}, nil
		}
		return nil, fmt.Errorf("failed to read accounts directory: %w", err)
	}

	var accounts []int
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		name := entry.Name()
		if !strings.HasPrefix(name, AccountPrefix) {
			continue
		}

		idStr := strings.TrimPrefix(name, AccountPrefix)
		id, err := strconv.Atoi(idStr)
		if err != nil {
			continue // Skip invalid directories
		}

		accounts = append(accounts, id)
	}

	return accounts, nil
}

// GetNextAccountID returns the next available account ID
func (s *StateManager) GetNextAccountID() (int, error) {
	accounts, err := s.ListAccounts()
	if err != nil {
		return 0, err
	}

	maxID := 10000 // Start from 10001
	for _, id := range accounts {
		if id > maxID {
			maxID = id
		}
	}

	return maxID + 1, nil
}

// atomicWrite writes data atomically using temp file + rename
func (s *StateManager) atomicWrite(path string, data interface{}) error {
	// Ensure parent directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	// Write to temp file first
	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write temp file: %w", err)
	}

	// Atomic rename
	if err := os.Rename(tmpPath, path); err != nil {
		os.Remove(tmpPath) // Clean up temp file
		return fmt.Errorf("failed to rename temp file: %w", err)
	}

	return nil
}

// readJSON is a helper to read and unmarshal JSON files
func (s *StateManager) readJSON(path string, v interface{}) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}
