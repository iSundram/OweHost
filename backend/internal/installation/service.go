// Package installation provides installation and setup services for OweHost
package installation

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	"github.com/iSundram/OweHost/internal/user"
	"github.com/iSundram/OweHost/pkg/config"
	"github.com/iSundram/OweHost/pkg/models"
	"github.com/iSundram/OweHost/pkg/utils"
)

const (
	installationFlagFile = "/var/lib/owehost/installed"
	installationDBFile   = "/var/lib/owehost/installation.db"
)

// Service provides installation management functionality
type Service struct {
	config      *config.Config
	userService *user.Service
	db          *sql.DB
	mu          sync.RWMutex
}

// NewService creates a new installation service
func NewService(cfg *config.Config, userSvc *user.Service) *Service {
	return &Service{
		config:      cfg,
		userService: userSvc,
	}
}

// IsInstalled checks if the system has been installed
func (s *Service) IsInstalled() bool {
	// Check for installation flag file
	if _, err := os.Stat(installationFlagFile); err == nil {
		return true
	}
	return false
}

// GetSupportedEngines returns list of supported database engines
func (s *Service) GetSupportedEngines() []models.DatabaseEngine {
	engines := []models.DatabaseEngine{
		{
			Type:        models.DatabaseTypeMySQL,
			Name:        "MySQL",
			Description: "MySQL is the world's most popular open source database",
			DefaultPort: 3306,
			IsInstalled: true, // Simplified check
		},
		{
			Type:        models.DatabaseTypeMariaDB,
			Name:        "MariaDB",
			Description: "MariaDB is a community-developed fork of MySQL",
			DefaultPort: 3306,
			IsInstalled: true,
		},
		{
			Type:        models.DatabaseTypePostgreSQL,
			Name:        "PostgreSQL",
			Description: "PostgreSQL is a powerful, open source object-relational database",
			DefaultPort: 5432,
			IsInstalled: true,
		},
		{
			Type:        models.DatabaseTypeMongoDB,
			Name:        "MongoDB",
			Description: "MongoDB is a document-oriented NoSQL database",
			DefaultPort: 27017,
			IsInstalled: true,
		},
		{
			Type:        models.DatabaseTypeRedis,
			Name:        "Redis",
			Description: "Redis is an in-memory data structure store",
			DefaultPort: 6379,
			IsInstalled: true,
		},
		{
			Type:        models.DatabaseTypeSQLite,
			Name:        "SQLite",
			Description: "SQLite is a self-contained, serverless SQL database engine",
			DefaultPort: 0,
			IsInstalled: true,
		},
	}
	return engines
}

// Install performs the installation process
func (s *Service) Install(req *models.InstallationRequest) (*models.Installation, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if already installed
	if s.IsInstalled() {
		return nil, errors.New("system is already installed")
	}

	installation := &models.Installation{
		ID:             utils.GenerateID("inst"),
		Status:         models.InstallationStatusInProgress,
		DatabaseEngine: req.DatabaseEngine,
		DatabaseHost:   req.DatabaseHost,
		DatabasePort:   req.DatabasePort,
		DatabaseName:   req.DatabaseName,
		DatabaseUser:   req.DatabaseUser,
		AdminUsername:  req.AdminUsername,
		AdminEmail:     req.AdminEmail,
		TotalSteps:     5,
		CreatedAt:      time.Now(),
	}

	// Step 1: Test database connection
	installation.InstallationStep = 1
	if err := s.testDatabaseConnection(req); err != nil {
		installation.Status = models.InstallationStatusFailed
		installation.ErrorMessage = fmt.Sprintf("Database connection failed: %v", err)
		return installation, err
	}

	// Step 2: Initialize database
	installation.InstallationStep = 2
	if err := s.initializeDatabase(req); err != nil {
		installation.Status = models.InstallationStatusFailed
		installation.ErrorMessage = fmt.Sprintf("Database initialization failed: %v", err)
		return installation, err
	}

	// Step 3: Create admin user (username: admin, password: admin@123)
	installation.InstallationStep = 3
	if err := s.createAdminUser(req); err != nil {
		installation.Status = models.InstallationStatusFailed
		installation.ErrorMessage = fmt.Sprintf("Admin user creation failed: %v", err)
		return installation, err
	}

	// Step 4: Setup system directories
	installation.InstallationStep = 4
	if err := s.setupSystemDirectories(); err != nil {
		installation.Status = models.InstallationStatusFailed
		installation.ErrorMessage = fmt.Sprintf("Directory setup failed: %v", err)
		return installation, err
	}

	// Step 5: Mark as installed
	installation.InstallationStep = 5
	if err := s.markAsInstalled(req); err != nil {
		installation.Status = models.InstallationStatusFailed
		installation.ErrorMessage = fmt.Sprintf("Failed to mark as installed: %v", err)
		return installation, err
	}

	completedAt := time.Now()
	installation.Status = models.InstallationStatusCompleted
	installation.CompletedAt = &completedAt

	return installation, nil
}

// testDatabaseConnection tests the database connection
func (s *Service) testDatabaseConnection(req *models.InstallationRequest) error {
	var dsn string
	var driver string

	switch req.DatabaseEngine {
	case models.DatabaseTypeMySQL, models.DatabaseTypeMariaDB:
		driver = "mysql"
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true",
			req.DatabaseUser, req.DatabasePassword, req.DatabaseHost, req.DatabasePort, req.DatabaseName)
	case models.DatabaseTypePostgreSQL:
		driver = "postgres"
		dsn = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			req.DatabaseHost, req.DatabasePort, req.DatabaseUser, req.DatabasePassword, req.DatabaseName)
	case models.DatabaseTypeSQLite:
		driver = "sqlite3"
		dsn = installationDBFile
	default:
		return fmt.Errorf("unsupported database engine: %s", req.DatabaseEngine)
	}

	db, err := sql.Open(driver, dsn)
	if err != nil {
		return err
	}
	defer db.Close()

	return db.Ping()
}

// initializeDatabase creates database tables
func (s *Service) initializeDatabase(req *models.InstallationRequest) error {
	var dsn string
	var driver string

	switch req.DatabaseEngine {
	case models.DatabaseTypeMySQL, models.DatabaseTypeMariaDB:
		driver = "mysql"
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true",
			req.DatabaseUser, req.DatabasePassword, req.DatabaseHost, req.DatabasePort, req.DatabaseName)
	case models.DatabaseTypePostgreSQL:
		driver = "postgres"
		dsn = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			req.DatabaseHost, req.DatabasePort, req.DatabaseUser, req.DatabasePassword, req.DatabaseName)
	case models.DatabaseTypeSQLite:
		driver = "sqlite3"
		dsn = installationDBFile
	default:
		return fmt.Errorf("unsupported database engine for migration: %s", req.DatabaseEngine)
	}

	db, err := sql.Open(driver, dsn)
	if err != nil {
		return err
	}
	defer db.Close()

	s.db = db

	// Create users table
	if err := s.createUsersTable(); err != nil {
		return err
	}

	// Create domains table
	if err := s.createDomainsTable(); err != nil {
		return err
	}

	// Create databases table
	if err := s.createDatabasesTable(); err != nil {
		return err
	}

	return nil
}

// createUsersTable creates the users table
func (s *Service) createUsersTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS users (
		id VARCHAR(64) PRIMARY KEY,
		tenant_id VARCHAR(64) NOT NULL,
		username VARCHAR(64) UNIQUE NOT NULL,
		email VARCHAR(255) UNIQUE NOT NULL,
		password_hash VARCHAR(255) NOT NULL,
		status VARCHAR(32) NOT NULL,
		uid INT NOT NULL,
		gid INT NOT NULL,
		home_directory VARCHAR(255) NOT NULL,
		namespace VARCHAR(64) NOT NULL,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`
	_, err := s.db.Exec(query)
	return err
}

// createDomainsTable creates the domains table
func (s *Service) createDomainsTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS domains (
		id VARCHAR(64) PRIMARY KEY,
		user_id VARCHAR(64) NOT NULL,
		name VARCHAR(255) UNIQUE NOT NULL,
		type VARCHAR(32) NOT NULL,
		status VARCHAR(32) NOT NULL,
		document_root VARCHAR(255) NOT NULL,
		validated BOOLEAN NOT NULL DEFAULT FALSE,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	)`
	_, err := s.db.Exec(query)
	return err
}

// createDatabasesTable creates the databases table
func (s *Service) createDatabasesTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS databases (
		id VARCHAR(64) PRIMARY KEY,
		user_id VARCHAR(64) NOT NULL,
		name VARCHAR(64) NOT NULL,
		type VARCHAR(32) NOT NULL,
		charset VARCHAR(32),
		collation VARCHAR(64),
		size_mb BIGINT NOT NULL DEFAULT 0,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	)`
	_, err := s.db.Exec(query)
	return err
}

// createAdminUser creates the default admin user with username: admin, password: admin@123
func (s *Service) createAdminUser(req *models.InstallationRequest) error {
	// Default credentials: admin / admin@123
	passwordHash, err := utils.HashPassword("admin@123")
	if err != nil {
		return err
	}

	query := `
	INSERT INTO users (id, tenant_id, username, email, password_hash, status, uid, gid, home_directory, namespace, created_at, updated_at)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	userID := utils.GenerateID("usr")
	now := time.Now()

	_, err = s.db.Exec(query,
		userID,
		"default",
		"admin",
		req.AdminEmail,
		passwordHash,
		"active",
		1000,
		1000,
		"/home/admin",
		"admin",
		now,
		now,
	)

	return err
}

// setupSystemDirectories creates necessary system directories
func (s *Service) setupSystemDirectories() error {
	dirs := []string{
		"/var/lib/owehost",
		"/var/log/owehost",
		"/etc/owehost",
		"/opt/owehost/backups",
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil && !os.IsExist(err) {
			return err
		}
	}

	return nil
}

// markAsInstalled creates the installation flag file and saves config
func (s *Service) markAsInstalled(req *models.InstallationRequest) error {
	dir := "/var/lib/owehost"
	if err := os.MkdirAll(dir, 0755); err != nil && !os.IsExist(err) {
		return err
	}

	// Create installation flag
	file, err := os.Create(installationFlagFile)
	if err != nil {
		return err
	}
	defer file.Close()

	installData := fmt.Sprintf(`Installation completed at: %s
Database Engine: %s
Database Host: %s
Database Port: %d
Database Name: %s
Admin Username: admin
Admin Email: %s
`,
		time.Now().Format(time.RFC3339),
		req.DatabaseEngine,
		req.DatabaseHost,
		req.DatabasePort,
		req.DatabaseName,
		req.AdminEmail,
	)

	_, err = file.WriteString(installData)
	return err
}
