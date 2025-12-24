// Package database provides database migration functionality
package database

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"time"
)

// Migration represents a database migration
type Migration struct {
	Version     int
	Name        string
	Up          func(*sql.Tx) error
	Down        func(*sql.Tx) error
}

// Migrator handles database migrations
type Migrator struct {
	db         *DB
	migrations []Migration
}

// NewMigrator creates a new migrator
func NewMigrator(db *DB) *Migrator {
	return &Migrator{
		db:         db,
		migrations: make([]Migration, 0),
	}
}

// Register registers a migration
func (m *Migrator) Register(migration Migration) {
	m.migrations = append(m.migrations, migration)
}

// RegisterAll registers multiple migrations
func (m *Migrator) RegisterAll(migrations []Migration) {
	m.migrations = append(m.migrations, migrations...)
}

// ensureMigrationsTable creates the migrations tracking table if it doesn't exist
func (m *Migrator) ensureMigrationsTable(ctx context.Context) error {
	var query string
	switch m.db.config.Driver {
	case "postgres":
		query = `
			CREATE TABLE IF NOT EXISTS schema_migrations (
				version INTEGER PRIMARY KEY,
				name VARCHAR(255) NOT NULL,
				applied_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
			)
		`
	case "mysql":
		query = `
			CREATE TABLE IF NOT EXISTS schema_migrations (
				version INT PRIMARY KEY,
				name VARCHAR(255) NOT NULL,
				applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			)
		`
	default:
		return fmt.Errorf("unsupported driver: %s", m.db.config.Driver)
	}

	_, err := m.db.ExecContext(ctx, query)
	return err
}

// getAppliedMigrations returns a map of applied migration versions
func (m *Migrator) getAppliedMigrations(ctx context.Context) (map[int]bool, error) {
	rows, err := m.db.QueryContext(ctx, "SELECT version FROM schema_migrations")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	applied := make(map[int]bool)
	for rows.Next() {
		var version int
		if err := rows.Scan(&version); err != nil {
			return nil, err
		}
		applied[version] = true
	}

	return applied, rows.Err()
}

// Up runs all pending migrations
func (m *Migrator) Up(ctx context.Context) error {
	if err := m.ensureMigrationsTable(ctx); err != nil {
		return fmt.Errorf("failed to ensure migrations table: %w", err)
	}

	applied, err := m.getAppliedMigrations(ctx)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	// Sort migrations by version
	sort.Slice(m.migrations, func(i, j int) bool {
		return m.migrations[i].Version < m.migrations[j].Version
	})

	for _, migration := range m.migrations {
		if applied[migration.Version] {
			continue
		}

		if err := m.runMigration(ctx, migration, true); err != nil {
			return fmt.Errorf("failed to run migration %d (%s): %w", migration.Version, migration.Name, err)
		}

		fmt.Printf("Applied migration %d: %s\n", migration.Version, migration.Name)
	}

	return nil
}

// Down rolls back the last migration
func (m *Migrator) Down(ctx context.Context) error {
	if err := m.ensureMigrationsTable(ctx); err != nil {
		return fmt.Errorf("failed to ensure migrations table: %w", err)
	}

	applied, err := m.getAppliedMigrations(ctx)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	// Sort migrations by version descending
	sort.Slice(m.migrations, func(i, j int) bool {
		return m.migrations[i].Version > m.migrations[j].Version
	})

	for _, migration := range m.migrations {
		if !applied[migration.Version] {
			continue
		}

		if err := m.runMigration(ctx, migration, false); err != nil {
			return fmt.Errorf("failed to rollback migration %d (%s): %w", migration.Version, migration.Name, err)
		}

		fmt.Printf("Rolled back migration %d: %s\n", migration.Version, migration.Name)
		return nil // Only rollback one migration
	}

	return nil
}

// DownTo rolls back to a specific version
func (m *Migrator) DownTo(ctx context.Context, targetVersion int) error {
	if err := m.ensureMigrationsTable(ctx); err != nil {
		return fmt.Errorf("failed to ensure migrations table: %w", err)
	}

	applied, err := m.getAppliedMigrations(ctx)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	// Sort migrations by version descending
	sort.Slice(m.migrations, func(i, j int) bool {
		return m.migrations[i].Version > m.migrations[j].Version
	})

	for _, migration := range m.migrations {
		if migration.Version <= targetVersion {
			break
		}

		if !applied[migration.Version] {
			continue
		}

		if err := m.runMigration(ctx, migration, false); err != nil {
			return fmt.Errorf("failed to rollback migration %d (%s): %w", migration.Version, migration.Name, err)
		}

		fmt.Printf("Rolled back migration %d: %s\n", migration.Version, migration.Name)
	}

	return nil
}

// runMigration executes a single migration
func (m *Migrator) runMigration(ctx context.Context, migration Migration, up bool) error {
	return m.db.Transaction(ctx, func(tx *sql.Tx) error {
		if up {
			if migration.Up == nil {
				return fmt.Errorf("migration %d has no Up function", migration.Version)
			}
			if err := migration.Up(tx); err != nil {
				return err
			}
			_, err := tx.ExecContext(ctx, 
				"INSERT INTO schema_migrations (version, name) VALUES ($1, $2)",
				migration.Version, migration.Name,
			)
			return err
		}

		if migration.Down == nil {
			return fmt.Errorf("migration %d has no Down function", migration.Version)
		}
		if err := migration.Down(tx); err != nil {
			return err
		}
		_, err := tx.ExecContext(ctx,
			"DELETE FROM schema_migrations WHERE version = $1",
			migration.Version,
		)
		return err
	})
}

// Status returns the current migration status
func (m *Migrator) Status(ctx context.Context) ([]MigrationStatus, error) {
	if err := m.ensureMigrationsTable(ctx); err != nil {
		return nil, fmt.Errorf("failed to ensure migrations table: %w", err)
	}

	applied, err := m.getAppliedMigrations(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get applied migrations: %w", err)
	}

	// Sort migrations by version
	sort.Slice(m.migrations, func(i, j int) bool {
		return m.migrations[i].Version < m.migrations[j].Version
	})

	statuses := make([]MigrationStatus, len(m.migrations))
	for i, migration := range m.migrations {
		statuses[i] = MigrationStatus{
			Version: migration.Version,
			Name:    migration.Name,
			Applied: applied[migration.Version],
		}
	}

	return statuses, nil
}

// MigrationStatus represents the status of a migration
type MigrationStatus struct {
	Version   int
	Name      string
	Applied   bool
	AppliedAt time.Time
}
