package models

import "time"

// DatabaseType represents the type of database
type DatabaseType string

const (
	DatabaseTypeMySQL      DatabaseType = "mysql"
	DatabaseTypePostgreSQL DatabaseType = "postgresql"
	DatabaseTypeMariaDB    DatabaseType = "mariadb"
	DatabaseTypeMongoDB    DatabaseType = "mongodb"
	DatabaseTypeRedis      DatabaseType = "redis"
	DatabaseTypeSQLite     DatabaseType = "sqlite"
)

// Database represents a database
type Database struct {
	ID           string       `json:"id"`
	UserID       string       `json:"user_id"`
	Name         string       `json:"name"`
	Type         DatabaseType `json:"type"`
	Charset      string       `json:"charset"`
	Collation    string       `json:"collation"`
	SizeMB       int64        `json:"size_mb"`
	CreatedAt    time.Time    `json:"created_at"`
	UpdatedAt    time.Time    `json:"updated_at"`
}

// DatabaseUser represents a database user
type DatabaseUser struct {
	ID           string    `json:"id"`
	DatabaseID   string    `json:"database_id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"`
	Privileges   []string  `json:"privileges"`
	RemoteHost   string    `json:"remote_host"`
	TLSRequired  bool      `json:"tls_required"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// DatabaseBackup represents a database backup
type DatabaseBackup struct {
	ID         string    `json:"id"`
	DatabaseID string    `json:"database_id"`
	FilePath   string    `json:"file_path"`
	SizeMB     int64     `json:"size_mb"`
	Type       string    `json:"type"`
	Checksum   string    `json:"checksum"`
	CreatedAt  time.Time `json:"created_at"`
}

// DatabaseCreateRequest represents a request to create a database
type DatabaseCreateRequest struct {
	Name      string       `json:"name" validate:"required,min=1,max=64"`
	Type      DatabaseType `json:"type" validate:"required,oneof=mysql postgresql mariadb mongodb redis sqlite"`
	Charset   string       `json:"charset,omitempty"`
	Collation string       `json:"collation,omitempty"`
}

// DatabaseUserCreateRequest represents a request to create a database user
type DatabaseUserCreateRequest struct {
	Username    string   `json:"username" validate:"required,min=1,max=32"`
	Password    string   `json:"password" validate:"required,min=8"`
	Privileges  []string `json:"privileges" validate:"required"`
	RemoteHost  string   `json:"remote_host,omitempty"`
	TLSRequired bool     `json:"tls_required"`
}
