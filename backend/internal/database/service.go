// Package database provides database management services for OweHost
package database

import (
	"errors"
	"sync"
	"time"

	"github.com/iSundram/OweHost/pkg/models"
	"github.com/iSundram/OweHost/pkg/utils"
)

// Service provides database management functionality
type Service struct {
	databases    map[string]*models.Database
	users        map[string]*models.DatabaseUser
	backups      map[string]*models.DatabaseBackup
	byUser       map[string][]*models.Database
	usersByDB    map[string][]*models.DatabaseUser
	mu           sync.RWMutex
}

// NewService creates a new database service
func NewService() *Service {
	return &Service{
		databases: make(map[string]*models.Database),
		users:     make(map[string]*models.DatabaseUser),
		backups:   make(map[string]*models.DatabaseBackup),
		byUser:    make(map[string][]*models.Database),
		usersByDB: make(map[string][]*models.DatabaseUser),
	}
}

// Create creates a new database
func (s *Service) Create(userID string, req *models.DatabaseCreateRequest) (*models.Database, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check for duplicate name for user
	for _, db := range s.byUser[userID] {
		if db.Name == req.Name {
			return nil, errors.New("database with this name already exists")
		}
	}

	charset := req.Charset
	if charset == "" {
		charset = "utf8mb4"
	}

	collation := req.Collation
	if collation == "" {
		collation = "utf8mb4_unicode_ci"
	}

	db := &models.Database{
		ID:        utils.GenerateID("db"),
		UserID:    userID,
		Name:      req.Name,
		Type:      req.Type,
		Charset:   charset,
		Collation: collation,
		SizeMB:    0,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	s.databases[db.ID] = db
	s.byUser[userID] = append(s.byUser[userID], db)
	s.usersByDB[db.ID] = make([]*models.DatabaseUser, 0)

	return db, nil
}

// Get gets a database by ID
func (s *Service) Get(id string) (*models.Database, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	db, exists := s.databases[id]
	if !exists {
		return nil, errors.New("database not found")
	}
	return db, nil
}

// ListByUser lists databases for a user
func (s *Service) ListByUser(userID string) []*models.Database {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.byUser[userID]
}

// ListAll lists all databases (for admin)
func (s *Service) ListAll() []*models.Database {
	s.mu.RLock()
	defer s.mu.RUnlock()

	databases := make([]*models.Database, 0, len(s.databases))
	for _, db := range s.databases {
		databases = append(databases, db)
	}
	return databases
}

// Delete deletes a database
func (s *Service) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	db, exists := s.databases[id]
	if !exists {
		return errors.New("database not found")
	}

	// Remove database users
	for _, user := range s.usersByDB[id] {
		delete(s.users, user.ID)
	}
	delete(s.usersByDB, id)

	// Remove from user's databases
	userDBs := s.byUser[db.UserID]
	for i, d := range userDBs {
		if d.ID == id {
			s.byUser[db.UserID] = append(userDBs[:i], userDBs[i+1:]...)
			break
		}
	}

	delete(s.databases, id)
	return nil
}

// CreateUser creates a database user
func (s *Service) CreateUser(dbID string, req *models.DatabaseUserCreateRequest) (*models.DatabaseUser, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, exists := s.databases[dbID]
	if !exists {
		return nil, errors.New("database not found")
	}

	// Check for duplicate username
	for _, u := range s.usersByDB[dbID] {
		if u.Username == req.Username {
			return nil, errors.New("username already exists for this database")
		}
	}

	passwordHash, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	remoteHost := req.RemoteHost
	if remoteHost == "" {
		remoteHost = "localhost"
	}

	user := &models.DatabaseUser{
		ID:           utils.GenerateID("dbu"),
		DatabaseID:   dbID,
		Username:     req.Username,
		PasswordHash: passwordHash,
		Privileges:   req.Privileges,
		RemoteHost:   remoteHost,
		TLSRequired:  req.TLSRequired,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	s.users[user.ID] = user
	s.usersByDB[dbID] = append(s.usersByDB[dbID], user)

	return user, nil
}

// GetUser gets a database user by ID
func (s *Service) GetUser(id string) (*models.DatabaseUser, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, exists := s.users[id]
	if !exists {
		return nil, errors.New("user not found")
	}
	return user, nil
}

// ListUsers lists users for a database
func (s *Service) ListUsers(dbID string) []*models.DatabaseUser {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.usersByDB[dbID]
}

// UpdateUserPrivileges updates database user privileges
func (s *Service) UpdateUserPrivileges(userID string, privileges []string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	user, exists := s.users[userID]
	if !exists {
		return errors.New("user not found")
	}

	user.Privileges = privileges
	user.UpdatedAt = time.Now()
	return nil
}

// DeleteUser deletes a database user
func (s *Service) DeleteUser(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	user, exists := s.users[id]
	if !exists {
		return errors.New("user not found")
	}

	// Remove from database's users
	dbUsers := s.usersByDB[user.DatabaseID]
	for i, u := range dbUsers {
		if u.ID == id {
			s.usersByDB[user.DatabaseID] = append(dbUsers[:i], dbUsers[i+1:]...)
			break
		}
	}

	delete(s.users, id)
	return nil
}

// SetRemoteAccess configures remote access for a database user
func (s *Service) SetRemoteAccess(userID, remoteHost string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	user, exists := s.users[userID]
	if !exists {
		return errors.New("user not found")
	}

	user.RemoteHost = remoteHost
	user.UpdatedAt = time.Now()
	return nil
}

// RequireTLS sets TLS requirement for a database user
func (s *Service) RequireTLS(userID string, required bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	user, exists := s.users[userID]
	if !exists {
		return errors.New("user not found")
	}

	user.TLSRequired = required
	user.UpdatedAt = time.Now()
	return nil
}

// CreateBackup creates a database backup
func (s *Service) CreateBackup(dbID string) (*models.DatabaseBackup, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	db, exists := s.databases[dbID]
	if !exists {
		return nil, errors.New("database not found")
	}

	backup := &models.DatabaseBackup{
		ID:         utils.GenerateID("dbbk"),
		DatabaseID: dbID,
		FilePath:   "/backups/db/" + dbID + "/" + time.Now().Format("20060102150405") + ".sql.gz",
		SizeMB:     db.SizeMB,
		Type:       "logical",
		Checksum:   "sha256:dummy",
		CreatedAt:  time.Now(),
	}

	s.backups[backup.ID] = backup
	return backup, nil
}

// DeleteAllByUser deletes all databases for a user
func (s *Service) DeleteAllByUser(userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	for id, db := range s.databases {
		if db.UserID == userID {
			delete(s.databases, id)
		}
	}
	delete(s.byUser, userID)
	return nil
}

// ListBackups lists backups for a database
func (s *Service) ListBackups(dbID string) []*models.DatabaseBackup {
	s.mu.RLock()
	defer s.mu.RUnlock()

	backups := make([]*models.DatabaseBackup, 0)
	for _, backup := range s.backups {
		if backup.DatabaseID == dbID {
			backups = append(backups, backup)
		}
	}
	return backups
}

// RestoreBackup restores a database from backup
func (s *Service) RestoreBackup(backupID string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	_, exists := s.backups[backupID]
	if !exists {
		return errors.New("backup not found")
	}

	// In production, this would run pg_restore or mysql import
	return nil
}
