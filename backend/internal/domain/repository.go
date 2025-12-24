// Package domain provides domain repository implementation
package domain

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/iSundram/OweHost/pkg/database"
	"github.com/iSundram/OweHost/pkg/models"
)

// Repository handles domain data persistence
type Repository struct {
	*database.BaseRepository
}

// NewRepository creates a new domain repository
func NewRepository(db *database.DB) *Repository {
	return &Repository{
		BaseRepository: database.NewBaseRepository(db),
	}
}

// Create creates a new domain in the database
func (r *Repository) Create(ctx context.Context, domain *models.Domain) error {
	query := `
		INSERT INTO domains (
			id, user_id, name, type, status, document_root,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8
		)
	`

	_, err := r.Querier().ExecContext(ctx, query,
		domain.ID,
		domain.UserID,
		domain.Name,
		string(domain.Type),
		string(domain.Status),
		domain.DocumentRoot,
		domain.CreatedAt,
		domain.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create domain: %w", err)
	}

	return nil
}

// GetByID retrieves a domain by ID
func (r *Repository) GetByID(ctx context.Context, id string) (*models.Domain, error) {
	query := `
		SELECT id, user_id, name, type, status, document_root,
		       created_at, updated_at
		FROM domains
		WHERE id = $1
	`

	domain := &models.Domain{}

	err := r.Querier().QueryRowContext(ctx, query, id).Scan(
		&domain.ID,
		&domain.UserID,
		&domain.Name,
		&domain.Type,
		&domain.Status,
		&domain.DocumentRoot,
		&domain.CreatedAt,
		&domain.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("domain not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get domain: %w", err)
	}

	return domain, nil
}

// GetByName retrieves a domain by name
func (r *Repository) GetByName(ctx context.Context, name string) (*models.Domain, error) {
	query := `
		SELECT id, user_id, name, type, status, document_root,
		       created_at, updated_at
		FROM domains
		WHERE name = $1
	`

	domain := &models.Domain{}

	err := r.Querier().QueryRowContext(ctx, query, name).Scan(
		&domain.ID,
		&domain.UserID,
		&domain.Name,
		&domain.Type,
		&domain.Status,
		&domain.DocumentRoot,
		&domain.CreatedAt,
		&domain.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("domain not found: %s", name)
		}
		return nil, fmt.Errorf("failed to get domain: %w", err)
	}

	return domain, nil
}

// Update updates a domain in the database
func (r *Repository) Update(ctx context.Context, domain *models.Domain) error {
	query := `
		UPDATE domains SET
			name = $2,
			type = $3,
			status = $4,
			document_root = $5,
			updated_at = $6
		WHERE id = $1
	`

	result, err := r.Querier().ExecContext(ctx, query,
		domain.ID,
		domain.Name,
		string(domain.Type),
		string(domain.Status),
		domain.DocumentRoot,
		time.Now(),
	)

	if err != nil {
		return fmt.Errorf("failed to update domain: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("domain not found: %s", domain.ID)
	}

	return nil
}

// Delete deletes a domain from the database
func (r *Repository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM domains WHERE id = $1`

	result, err := r.Querier().ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete domain: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("domain not found: %s", id)
	}

	return nil
}

// ListByUserID retrieves domains by user ID
func (r *Repository) ListByUserID(ctx context.Context, userID string, pagination database.Pagination) ([]models.Domain, int64, error) {
	// Get total count
	countQuery := `SELECT COUNT(*) FROM domains WHERE user_id = $1`
	var total int64
	if err := r.Querier().QueryRowContext(ctx, countQuery, userID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count domains: %w", err)
	}

	// Get domains
	query := fmt.Sprintf(`
		SELECT id, user_id, name, type, status, document_root,
		       created_at, updated_at
		FROM domains
		WHERE user_id = $1
		ORDER BY %s %s
		LIMIT $2 OFFSET $3
	`, pagination.OrderBy, pagination.OrderDir)

	rows, err := r.Querier().QueryContext(ctx, query, userID, pagination.Limit(), pagination.Offset())
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list domains: %w", err)
	}
	defer rows.Close()

	return r.scanDomains(rows, total)
}

// List retrieves all domains with pagination
func (r *Repository) List(ctx context.Context, pagination database.Pagination) ([]models.Domain, int64, error) {
	// Get total count
	countQuery := `SELECT COUNT(*) FROM domains`
	var total int64
	if err := r.Querier().QueryRowContext(ctx, countQuery).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count domains: %w", err)
	}

	// Get domains
	query := fmt.Sprintf(`
		SELECT id, user_id, name, type, status, document_root,
		       created_at, updated_at
		FROM domains
		ORDER BY %s %s
		LIMIT $1 OFFSET $2
	`, pagination.OrderBy, pagination.OrderDir)

	rows, err := r.Querier().QueryContext(ctx, query, pagination.Limit(), pagination.Offset())
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list domains: %w", err)
	}
	defer rows.Close()

	return r.scanDomains(rows, total)
}

func (r *Repository) scanDomains(rows *sql.Rows, total int64) ([]models.Domain, int64, error) {
	domains := make([]models.Domain, 0)

	for rows.Next() {
		var domain models.Domain

		if err := rows.Scan(
			&domain.ID,
			&domain.UserID,
			&domain.Name,
			&domain.Type,
			&domain.Status,
			&domain.DocumentRoot,
			&domain.CreatedAt,
			&domain.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan domain: %w", err)
		}

		domains = append(domains, domain)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating domains: %w", err)
	}

	return domains, total, nil
}

// ExistsByName checks if a domain exists by name
func (r *Repository) ExistsByName(ctx context.Context, name string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM domains WHERE name = $1)`
	var exists bool
	err := r.Querier().QueryRowContext(ctx, query, name).Scan(&exists)
	return exists, err
}

// CountByUserID counts domains by user ID
func (r *Repository) CountByUserID(ctx context.Context, userID string) (int64, error) {
	query := `SELECT COUNT(*) FROM domains WHERE user_id = $1`
	var count int64
	err := r.Querier().QueryRowContext(ctx, query, userID).Scan(&count)
	return count, err
}

// CountByStatus counts domains by status
func (r *Repository) CountByStatus(ctx context.Context, status models.DomainStatus) (int64, error) {
	query := `SELECT COUNT(*) FROM domains WHERE status = $1`
	var count int64
	err := r.Querier().QueryRowContext(ctx, query, string(status)).Scan(&count)
	return count, err
}

// UpdateStatus updates a domain's status
func (r *Repository) UpdateStatus(ctx context.Context, id string, status models.DomainStatus) error {
	query := `UPDATE domains SET status = $2, updated_at = $3 WHERE id = $1`
	_, err := r.Querier().ExecContext(ctx, query, id, string(status), time.Now())
	return err
}

// CreateSubdomain creates a new subdomain
func (r *Repository) CreateSubdomain(ctx context.Context, subdomain *models.Subdomain) error {
	query := `
		INSERT INTO subdomains (
			id, domain_id, name, document_root,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6
		)
	`

	_, err := r.Querier().ExecContext(ctx, query,
		subdomain.ID,
		subdomain.DomainID,
		subdomain.Name,
		subdomain.DocumentRoot,
		subdomain.CreatedAt,
		subdomain.UpdatedAt,
	)

	return err
}

// DeleteSubdomain deletes a subdomain
func (r *Repository) DeleteSubdomain(ctx context.Context, id string) error {
	query := `DELETE FROM subdomains WHERE id = $1`
	_, err := r.Querier().ExecContext(ctx, query, id)
	return err
}

// ListSubdomains lists subdomains for a domain
func (r *Repository) ListSubdomains(ctx context.Context, domainID string) ([]models.Subdomain, error) {
	query := `
		SELECT id, domain_id, name, document_root, created_at, updated_at
		FROM subdomains
		WHERE domain_id = $1
		ORDER BY name
	`

	rows, err := r.Querier().QueryContext(ctx, query, domainID)
	if err != nil {
		return nil, fmt.Errorf("failed to list subdomains: %w", err)
	}
	defer rows.Close()

	subdomains := make([]models.Subdomain, 0)
	for rows.Next() {
		var subdomain models.Subdomain
		if err := rows.Scan(
			&subdomain.ID,
			&subdomain.DomainID,
			&subdomain.Name,
			&subdomain.DocumentRoot,
			&subdomain.CreatedAt,
			&subdomain.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan subdomain: %w", err)
		}
		subdomains = append(subdomains, subdomain)
	}

	return subdomains, rows.Err()
}
