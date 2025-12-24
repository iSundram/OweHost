// Package database provides filesystem-based database state management
package database

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/iSundram/OweHost/internal/storage/account"
)

// DatabaseMeta represents database metadata stored in meta.json
type DatabaseMeta struct {
	Databases []DatabaseInfo `json:"databases"`
	UpdatedAt string         `json:"updated_at"`
}

// DatabaseInfo represents a single database
type DatabaseInfo struct {
	Name      string    `json:"name"`
	Type      string    `json:"type"`      // mysql, postgres, mariadb
	Charset   string    `json:"charset"`
	Collation string    `json:"collation"`
	SizeMB    int       `json:"size_mb"`
	Users     []DBUser  `json:"users"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// DBUser represents a database user
type DBUser struct {
	Username     string   `json:"username"`
	Host         string   `json:"host"`        // localhost, %, or specific IP
	Privileges   []string `json:"privileges"`
	RequireSSL   bool     `json:"require_ssl"`
	CreatedAt    time.Time `json:"created_at"`
}

// StateManager handles database state in the filesystem
type StateManager struct {
	mu sync.RWMutex
}

// NewStateManager creates a new database state manager
func NewStateManager() *StateManager {
	return &StateManager{}
}

// DatabasePath returns the path for database storage
func (s *StateManager) DatabasePath(accountID int) string {
	return filepath.Join(
		account.BaseAccountPath,
		fmt.Sprintf("%s%d", account.AccountPrefix, accountID),
		"databases",
	)
}

// MetaPath returns the path to meta.json
func (s *StateManager) MetaPath(accountID int) string {
	return filepath.Join(s.DatabasePath(accountID), "meta.json")
}

// ReadMeta reads the database metadata
func (s *StateManager) ReadMeta(accountID int) (*DatabaseMeta, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, err := os.ReadFile(s.MetaPath(accountID))
	if err != nil {
		if os.IsNotExist(err) {
			return &DatabaseMeta{Databases: []DatabaseInfo{}}, nil
		}
		return nil, err
	}

	var meta DatabaseMeta
	if err := json.Unmarshal(data, &meta); err != nil {
		return nil, err
	}

	return &meta, nil
}

// WriteMeta writes the database metadata
func (s *StateManager) WriteMeta(accountID int, meta *DatabaseMeta) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	meta.UpdatedAt = time.Now().Format(time.RFC3339)

	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return err
	}

	// Ensure directory exists
	if err := os.MkdirAll(s.DatabasePath(accountID), 0755); err != nil {
		return err
	}

	return os.WriteFile(s.MetaPath(accountID), data, 0644)
}

// AddDatabase adds a database to the account
func (s *StateManager) AddDatabase(accountID int, db DatabaseInfo) error {
	meta, err := s.ReadMeta(accountID)
	if err != nil {
		return err
	}

	// Check if database already exists
	for _, existing := range meta.Databases {
		if existing.Name == db.Name && existing.Type == db.Type {
			return fmt.Errorf("database %s already exists", db.Name)
		}
	}

	db.CreatedAt = time.Now()
	db.UpdatedAt = time.Now()
	meta.Databases = append(meta.Databases, db)

	// Create type-specific directory
	typeDir := filepath.Join(s.DatabasePath(accountID), db.Type, db.Name)
	if err := os.MkdirAll(typeDir, 0755); err != nil {
		return err
	}

	return s.WriteMeta(accountID, meta)
}

// RemoveDatabase removes a database from the account
func (s *StateManager) RemoveDatabase(accountID int, dbName, dbType string) error {
	meta, err := s.ReadMeta(accountID)
	if err != nil {
		return err
	}

	var updated []DatabaseInfo
	found := false
	for _, db := range meta.Databases {
		if db.Name == dbName && db.Type == dbType {
			found = true
			continue
		}
		updated = append(updated, db)
	}

	if !found {
		return fmt.Errorf("database %s not found", dbName)
	}

	meta.Databases = updated

	// Remove directory
	typeDir := filepath.Join(s.DatabasePath(accountID), dbType, dbName)
	os.RemoveAll(typeDir)

	return s.WriteMeta(accountID, meta)
}

// GetDatabase returns a specific database
func (s *StateManager) GetDatabase(accountID int, dbName, dbType string) (*DatabaseInfo, error) {
	meta, err := s.ReadMeta(accountID)
	if err != nil {
		return nil, err
	}

	for _, db := range meta.Databases {
		if db.Name == dbName && db.Type == dbType {
			return &db, nil
		}
	}

	return nil, fmt.Errorf("database %s not found", dbName)
}

// ListDatabases returns all databases for an account
func (s *StateManager) ListDatabases(accountID int) ([]DatabaseInfo, error) {
	meta, err := s.ReadMeta(accountID)
	if err != nil {
		return nil, err
	}
	return meta.Databases, nil
}

// AddUser adds a user to a database
func (s *StateManager) AddUser(accountID int, dbName, dbType string, user DBUser) error {
	meta, err := s.ReadMeta(accountID)
	if err != nil {
		return err
	}

	for i, db := range meta.Databases {
		if db.Name == dbName && db.Type == dbType {
			// Check if user already exists
			for _, existing := range db.Users {
				if existing.Username == user.Username && existing.Host == user.Host {
					return fmt.Errorf("user %s@%s already exists", user.Username, user.Host)
				}
			}

			user.CreatedAt = time.Now()
			meta.Databases[i].Users = append(meta.Databases[i].Users, user)
			meta.Databases[i].UpdatedAt = time.Now()
			return s.WriteMeta(accountID, meta)
		}
	}

	return fmt.Errorf("database %s not found", dbName)
}

// RemoveUser removes a user from a database
func (s *StateManager) RemoveUser(accountID int, dbName, dbType, username, host string) error {
	meta, err := s.ReadMeta(accountID)
	if err != nil {
		return err
	}

	for i, db := range meta.Databases {
		if db.Name == dbName && db.Type == dbType {
			var updated []DBUser
			found := false
			for _, u := range db.Users {
				if u.Username == username && u.Host == host {
					found = true
					continue
				}
				updated = append(updated, u)
			}

			if !found {
				return fmt.Errorf("user %s@%s not found", username, host)
			}

			meta.Databases[i].Users = updated
			meta.Databases[i].UpdatedAt = time.Now()
			return s.WriteMeta(accountID, meta)
		}
	}

	return fmt.Errorf("database %s not found", dbName)
}

// UpdateSize updates the size of a database
func (s *StateManager) UpdateSize(accountID int, dbName, dbType string, sizeMB int) error {
	meta, err := s.ReadMeta(accountID)
	if err != nil {
		return err
	}

	for i, db := range meta.Databases {
		if db.Name == dbName && db.Type == dbType {
			meta.Databases[i].SizeMB = sizeMB
			meta.Databases[i].UpdatedAt = time.Now()
			return s.WriteMeta(accountID, meta)
		}
	}

	return fmt.Errorf("database %s not found", dbName)
}

// CountDatabases returns the total count of databases for an account
func (s *StateManager) CountDatabases(accountID int) (int, error) {
	meta, err := s.ReadMeta(accountID)
	if err != nil {
		return 0, err
	}
	return len(meta.Databases), nil
}

// CountByType returns count of databases by type
func (s *StateManager) CountByType(accountID int, dbType string) (int, error) {
	meta, err := s.ReadMeta(accountID)
	if err != nil {
		return 0, err
	}

	count := 0
	for _, db := range meta.Databases {
		if db.Type == dbType {
			count++
		}
	}
	return count, nil
}

// TotalSize returns total database size in MB for an account
func (s *StateManager) TotalSize(accountID int) (int, error) {
	meta, err := s.ReadMeta(accountID)
	if err != nil {
		return 0, err
	}

	total := 0
	for _, db := range meta.Databases {
		total += db.SizeMB
	}
	return total, nil
}
