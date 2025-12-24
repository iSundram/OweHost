// Package migrations contains database migrations
package migrations

import (
	"database/sql"

	"github.com/iSundram/OweHost/pkg/database"
)

// All returns all migrations
func All() []database.Migration {
	return []database.Migration{
		Migration001CreateUsersTable(),
		Migration002CreateDomainsTable(),
		Migration003CreateDatabasesTable(),
	}
}

// Migration001CreateUsersTable creates the users table
func Migration001CreateUsersTable() database.Migration {
	return database.Migration{
		Version: 1,
		Name:    "create_users_table",
		Up: func(tx *sql.Tx) error {
			_, err := tx.Exec(`
				CREATE TABLE IF NOT EXISTS users (
					id VARCHAR(36) PRIMARY KEY,
					username VARCHAR(50) UNIQUE NOT NULL,
					email VARCHAR(255) UNIQUE NOT NULL,
					password_hash VARCHAR(255) NOT NULL,
					first_name VARCHAR(100),
					last_name VARCHAR(100),
					role VARCHAR(20) NOT NULL DEFAULT 'user',
					status VARCHAR(20) NOT NULL DEFAULT 'active',
					reseller_id VARCHAR(36),
					package_id VARCHAR(36),
					api_key VARCHAR(64) UNIQUE,
					two_factor_enabled BOOLEAN DEFAULT FALSE,
					two_factor_secret VARCHAR(255),
					last_login TIMESTAMP WITH TIME ZONE,
					login_count INTEGER DEFAULT 0,
					failed_login_count INTEGER DEFAULT 0,
					locked_until TIMESTAMP WITH TIME ZONE,
					created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
					updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
					suspended_at TIMESTAMP WITH TIME ZONE,
					suspended_reason TEXT,
					terminated_at TIMESTAMP WITH TIME ZONE,
					CONSTRAINT users_role_check CHECK (role IN ('admin', 'reseller', 'user')),
					CONSTRAINT users_status_check CHECK (status IN ('active', 'suspended', 'terminated', 'pending'))
				);

				CREATE INDEX idx_users_username ON users(username);
				CREATE INDEX idx_users_email ON users(email);
				CREATE INDEX idx_users_role ON users(role);
				CREATE INDEX idx_users_status ON users(status);
				CREATE INDEX idx_users_reseller_id ON users(reseller_id);
				CREATE INDEX idx_users_api_key ON users(api_key);
			`)
			return err
		},
		Down: func(tx *sql.Tx) error {
			_, err := tx.Exec(`DROP TABLE IF EXISTS users CASCADE`)
			return err
		},
	}
}

// Migration002CreateDomainsTable creates the domains table
func Migration002CreateDomainsTable() database.Migration {
	return database.Migration{
		Version: 2,
		Name:    "create_domains_table",
		Up: func(tx *sql.Tx) error {
			_, err := tx.Exec(`
				CREATE TABLE IF NOT EXISTS domains (
					id VARCHAR(36) PRIMARY KEY,
					user_id VARCHAR(36) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
					name VARCHAR(255) UNIQUE NOT NULL,
					type VARCHAR(20) NOT NULL DEFAULT 'primary',
					status VARCHAR(20) NOT NULL DEFAULT 'active',
					document_root VARCHAR(500),
					php_version VARCHAR(10) DEFAULT '8.2',
					ssl_enabled BOOLEAN DEFAULT FALSE,
					ssl_certificate_id VARCHAR(36),
					dns_zone_id VARCHAR(36),
					created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
					updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
					expires_at TIMESTAMP WITH TIME ZONE,
					CONSTRAINT domains_type_check CHECK (type IN ('primary', 'addon', 'subdomain', 'parked', 'alias')),
					CONSTRAINT domains_status_check CHECK (status IN ('active', 'suspended', 'pending', 'expired'))
				);

				CREATE INDEX idx_domains_user_id ON domains(user_id);
				CREATE INDEX idx_domains_name ON domains(name);
				CREATE INDEX idx_domains_type ON domains(type);
				CREATE INDEX idx_domains_status ON domains(status);

				CREATE TABLE IF NOT EXISTS subdomains (
					id VARCHAR(36) PRIMARY KEY,
					domain_id VARCHAR(36) NOT NULL REFERENCES domains(id) ON DELETE CASCADE,
					user_id VARCHAR(36) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
					name VARCHAR(100) NOT NULL,
					document_root VARCHAR(500),
					php_version VARCHAR(10) DEFAULT '8.2',
					ssl_enabled BOOLEAN DEFAULT FALSE,
					created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
					updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
					UNIQUE(domain_id, name)
				);

				CREATE INDEX idx_subdomains_domain_id ON subdomains(domain_id);
				CREATE INDEX idx_subdomains_user_id ON subdomains(user_id);

				CREATE TABLE IF NOT EXISTS domain_redirects (
					id VARCHAR(36) PRIMARY KEY,
					domain_id VARCHAR(36) NOT NULL REFERENCES domains(id) ON DELETE CASCADE,
					source_path VARCHAR(500) NOT NULL,
					target_url VARCHAR(1000) NOT NULL,
					redirect_type INTEGER NOT NULL DEFAULT 301,
					is_wildcard BOOLEAN DEFAULT FALSE,
					created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
				);

				CREATE INDEX idx_domain_redirects_domain_id ON domain_redirects(domain_id);

				CREATE TABLE IF NOT EXISTS domain_error_pages (
					id VARCHAR(36) PRIMARY KEY,
					domain_id VARCHAR(36) NOT NULL REFERENCES domains(id) ON DELETE CASCADE,
					error_code INTEGER NOT NULL,
					content TEXT NOT NULL,
					is_file BOOLEAN DEFAULT FALSE,
					file_path VARCHAR(500),
					created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
					updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
					UNIQUE(domain_id, error_code)
				);

				CREATE INDEX idx_domain_error_pages_domain_id ON domain_error_pages(domain_id);
			`)
			return err
		},
		Down: func(tx *sql.Tx) error {
			_, err := tx.Exec(`
				DROP TABLE IF EXISTS domain_error_pages CASCADE;
				DROP TABLE IF EXISTS domain_redirects CASCADE;
				DROP TABLE IF EXISTS subdomains CASCADE;
				DROP TABLE IF EXISTS domains CASCADE;
			`)
			return err
		},
	}
}

// Migration003CreateDatabasesTable creates the databases table
func Migration003CreateDatabasesTable() database.Migration {
	return database.Migration{
		Version: 3,
		Name:    "create_databases_table",
		Up: func(tx *sql.Tx) error {
			_, err := tx.Exec(`
				CREATE TABLE IF NOT EXISTS databases (
					id VARCHAR(36) PRIMARY KEY,
					user_id VARCHAR(36) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
					name VARCHAR(64) UNIQUE NOT NULL,
					type VARCHAR(20) NOT NULL DEFAULT 'mysql',
					status VARCHAR(20) NOT NULL DEFAULT 'active',
					size_bytes BIGINT DEFAULT 0,
					charset VARCHAR(50) DEFAULT 'utf8mb4',
					collation VARCHAR(50) DEFAULT 'utf8mb4_unicode_ci',
					created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
					updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
					CONSTRAINT databases_type_check CHECK (type IN ('mysql', 'postgresql', 'mariadb')),
					CONSTRAINT databases_status_check CHECK (status IN ('active', 'suspended', 'creating', 'deleting'))
				);

				CREATE INDEX idx_databases_user_id ON databases(user_id);
				CREATE INDEX idx_databases_name ON databases(name);
				CREATE INDEX idx_databases_type ON databases(type);

				CREATE TABLE IF NOT EXISTS database_users (
					id VARCHAR(36) PRIMARY KEY,
					database_id VARCHAR(36) NOT NULL REFERENCES databases(id) ON DELETE CASCADE,
					user_id VARCHAR(36) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
					username VARCHAR(64) NOT NULL,
					password_hash VARCHAR(255),
					privileges TEXT[] DEFAULT ARRAY['ALL'],
					remote_access BOOLEAN DEFAULT FALSE,
					remote_hosts TEXT[] DEFAULT ARRAY['localhost'],
					created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
					updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
					UNIQUE(database_id, username)
				);

				CREATE INDEX idx_database_users_database_id ON database_users(database_id);
				CREATE INDEX idx_database_users_user_id ON database_users(user_id);

				CREATE TABLE IF NOT EXISTS database_backups (
					id VARCHAR(36) PRIMARY KEY,
					database_id VARCHAR(36) NOT NULL REFERENCES databases(id) ON DELETE CASCADE,
					filename VARCHAR(255) NOT NULL,
					size_bytes BIGINT DEFAULT 0,
					type VARCHAR(20) NOT NULL DEFAULT 'manual',
					status VARCHAR(20) NOT NULL DEFAULT 'pending',
					compression VARCHAR(20) DEFAULT 'gzip',
					storage_path VARCHAR(500),
					created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
					completed_at TIMESTAMP WITH TIME ZONE,
					expires_at TIMESTAMP WITH TIME ZONE,
					CONSTRAINT database_backups_type_check CHECK (type IN ('manual', 'scheduled', 'pre_upgrade')),
					CONSTRAINT database_backups_status_check CHECK (status IN ('pending', 'in_progress', 'completed', 'failed'))
				);

				CREATE INDEX idx_database_backups_database_id ON database_backups(database_id);
				CREATE INDEX idx_database_backups_status ON database_backups(status);
			`)
			return err
		},
		Down: func(tx *sql.Tx) error {
			_, err := tx.Exec(`
				DROP TABLE IF EXISTS database_backups CASCADE;
				DROP TABLE IF EXISTS database_users CASCADE;
				DROP TABLE IF EXISTS databases CASCADE;
			`)
			return err
		},
	}
}
