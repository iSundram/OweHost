// Package user provides user repository implementation
package user

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/iSundram/OweHost/pkg/database"
	"github.com/iSundram/OweHost/pkg/models"
)

// Repository handles user data persistence
type Repository struct {
	*database.BaseRepository
}

// NewRepository creates a new user repository
func NewRepository(db *database.DB) *Repository {
	return &Repository{
		BaseRepository: database.NewBaseRepository(db),
	}
}

// WithTx returns a new repository instance with the given transaction
func (r *Repository) WithTx(tx *sql.Tx) *Repository {
	return &Repository{
		BaseRepository: &database.BaseRepository{},
	}
}

// Create creates a new user in the database
func (r *Repository) Create(ctx context.Context, user *models.User) error {
	query := `
		INSERT INTO users (
			id, username, email, password_hash, first_name, last_name,
			role, status, reseller_id, package_id, api_key,
			two_factor_enabled, two_factor_secret,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6,
			$7, $8, $9, $10, $11,
			$12, $13,
			$14, $15
		)
	`

	_, err := r.Querier().ExecContext(ctx, query,
		user.ID,
		user.Username,
		user.Email,
		user.PasswordHash,
		sql.NullString{}, // first_name
		sql.NullString{}, // last_name
		string(user.Role),
		string(user.Status),
		sql.NullString{}, // reseller_id
		sql.NullString{}, // package_id
		sql.NullString{}, // api_key
		false,            // two_factor_enabled
		sql.NullString{}, // two_factor_secret
		user.CreatedAt,
		user.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// GetByID retrieves a user by ID
func (r *Repository) GetByID(ctx context.Context, id string) (*models.User, error) {
	query := `
		SELECT id, username, email, password_hash, role, status,
		       created_at, updated_at, last_login
		FROM users
		WHERE id = $1
	`

	user := &models.User{}
	var lastLogin sql.NullTime

	err := r.Querier().QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.Status,
		&user.CreatedAt,
		&user.UpdatedAt,
		&lastLogin,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("user not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if lastLogin.Valid {
		user.LastLoginAt = &lastLogin.Time
	}

	return user, nil
}

// GetByEmail retrieves a user by email
func (r *Repository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `
		SELECT id, username, email, password_hash, role, status,
		       created_at, updated_at, last_login
		FROM users
		WHERE email = $1
	`

	user := &models.User{}
	var lastLogin sql.NullTime

	err := r.Querier().QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.Status,
		&user.CreatedAt,
		&user.UpdatedAt,
		&lastLogin,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("user not found: %s", email)
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if lastLogin.Valid {
		user.LastLoginAt = &lastLogin.Time
	}

	return user, nil
}

// GetByUsername retrieves a user by username
func (r *Repository) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	query := `
		SELECT id, username, email, password_hash, role, status,
		       created_at, updated_at, last_login
		FROM users
		WHERE username = $1
	`

	user := &models.User{}
	var lastLogin sql.NullTime

	err := r.Querier().QueryRowContext(ctx, query, username).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.Status,
		&user.CreatedAt,
		&user.UpdatedAt,
		&lastLogin,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("user not found: %s", username)
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if lastLogin.Valid {
		user.LastLoginAt = &lastLogin.Time
	}

	return user, nil
}

// Update updates a user in the database
func (r *Repository) Update(ctx context.Context, user *models.User) error {
	query := `
		UPDATE users SET
			username = $2,
			email = $3,
			password_hash = $4,
			role = $5,
			status = $6,
			updated_at = $7
		WHERE id = $1
	`

	result, err := r.Querier().ExecContext(ctx, query,
		user.ID,
		user.Username,
		user.Email,
		user.PasswordHash,
		string(user.Role),
		string(user.Status),
		time.Now(),
	)

	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("user not found: %s", user.ID)
	}

	return nil
}

// Delete deletes a user from the database
func (r *Repository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM users WHERE id = $1`

	result, err := r.Querier().ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("user not found: %s", id)
	}

	return nil
}

// List retrieves users with pagination
func (r *Repository) List(ctx context.Context, pagination database.Pagination) ([]models.User, int64, error) {
	// Get total count
	countQuery := `SELECT COUNT(*) FROM users`
	var total int64
	if err := r.Querier().QueryRowContext(ctx, countQuery).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	// Get users
	query := fmt.Sprintf(`
		SELECT id, username, email, password_hash, role, status,
		       created_at, updated_at, last_login
		FROM users
		ORDER BY %s %s
		LIMIT $1 OFFSET $2
	`, pagination.OrderBy, pagination.OrderDir)

	rows, err := r.Querier().QueryContext(ctx, query, pagination.Limit(), pagination.Offset())
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}
	defer rows.Close()

	users := make([]models.User, 0)
	for rows.Next() {
		var user models.User
		var lastLogin sql.NullTime

		if err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.Email,
			&user.PasswordHash,
			&user.Role,
			&user.Status,
			&user.CreatedAt,
			&user.UpdatedAt,
			&lastLogin,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan user: %w", err)
		}

		if lastLogin.Valid {
			user.LastLoginAt = &lastLogin.Time
		}

		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating users: %w", err)
	}

	return users, total, nil
}

// ListByResellerID retrieves users by reseller ID
func (r *Repository) ListByResellerID(ctx context.Context, resellerID string, pagination database.Pagination) ([]models.User, int64, error) {
	// Get total count
	countQuery := `SELECT COUNT(*) FROM users WHERE reseller_id = $1`
	var total int64
	if err := r.Querier().QueryRowContext(ctx, countQuery, resellerID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	// Get users
	query := fmt.Sprintf(`
		SELECT id, username, email, password_hash, role, status,
		       created_at, updated_at, last_login
		FROM users
		WHERE reseller_id = $1
		ORDER BY %s %s
		LIMIT $2 OFFSET $3
	`, pagination.OrderBy, pagination.OrderDir)

	rows, err := r.Querier().QueryContext(ctx, query, resellerID, pagination.Limit(), pagination.Offset())
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}
	defer rows.Close()

	users := make([]models.User, 0)
	for rows.Next() {
		var user models.User
		var lastLogin sql.NullTime

		if err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.Email,
			&user.PasswordHash,
			&user.Role,
			&user.Status,
			&user.CreatedAt,
			&user.UpdatedAt,
			&lastLogin,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan user: %w", err)
		}

		if lastLogin.Valid {
			user.LastLoginAt = &lastLogin.Time
		}

		users = append(users, user)
	}

	return users, total, nil
}

// UpdateLastLogin updates the last login timestamp
func (r *Repository) UpdateLastLogin(ctx context.Context, id string) error {
	query := `
		UPDATE users SET
			last_login = $2,
			login_count = login_count + 1,
			failed_login_count = 0
		WHERE id = $1
	`

	_, err := r.Querier().ExecContext(ctx, query, id, time.Now())
	return err
}

// IncrementFailedLogin increments the failed login count
func (r *Repository) IncrementFailedLogin(ctx context.Context, id string) error {
	query := `UPDATE users SET failed_login_count = failed_login_count + 1 WHERE id = $1`
	_, err := r.Querier().ExecContext(ctx, query, id)
	return err
}

// SuspendUser suspends a user account
func (r *Repository) SuspendUser(ctx context.Context, id, reason string) error {
	query := `
		UPDATE users SET
			status = 'suspended',
			suspended_at = $2,
			suspended_reason = $3,
			updated_at = $2
		WHERE id = $1
	`

	_, err := r.Querier().ExecContext(ctx, query, id, time.Now(), reason)
	return err
}

// UnsuspendUser unsuspends a user account
func (r *Repository) UnsuspendUser(ctx context.Context, id string) error {
	query := `
		UPDATE users SET
			status = 'active',
			suspended_at = NULL,
			suspended_reason = NULL,
			updated_at = $2
		WHERE id = $1
	`

	_, err := r.Querier().ExecContext(ctx, query, id, time.Now())
	return err
}

// TerminateUser terminates a user account
func (r *Repository) TerminateUser(ctx context.Context, id string) error {
	query := `
		UPDATE users SET
			status = 'terminated',
			terminated_at = $2,
			updated_at = $2
		WHERE id = $1
	`

	_, err := r.Querier().ExecContext(ctx, query, id, time.Now())
	return err
}

// ExistsByEmail checks if a user exists by email
func (r *Repository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`
	var exists bool
	err := r.Querier().QueryRowContext(ctx, query, email).Scan(&exists)
	return exists, err
}

// ExistsByUsername checks if a user exists by username
func (r *Repository) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE username = $1)`
	var exists bool
	err := r.Querier().QueryRowContext(ctx, query, username).Scan(&exists)
	return exists, err
}

// CountByRole counts users by role
func (r *Repository) CountByRole(ctx context.Context, role models.UserRole) (int64, error) {
	query := `SELECT COUNT(*) FROM users WHERE role = $1`
	var count int64
	err := r.Querier().QueryRowContext(ctx, query, string(role)).Scan(&count)
	return count, err
}

// CountByStatus counts users by status
func (r *Repository) CountByStatus(ctx context.Context, status models.UserStatus) (int64, error) {
	query := `SELECT COUNT(*) FROM users WHERE status = $1`
	var count int64
	err := r.Querier().QueryRowContext(ctx, query, string(status)).Scan(&count)
	return count, err
}
