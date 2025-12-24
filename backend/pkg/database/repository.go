// Package database provides repository interfaces and base implementations
package database

import (
	"context"
	"database/sql"
)

// Repository is the base interface for all repositories
type Repository interface {
	// DB returns the underlying database connection
	DB() *DB
	// Tx returns the current transaction if any
	Tx() *sql.Tx
	// WithTx returns a new repository instance with the given transaction
	WithTx(tx *sql.Tx) Repository
}

// BaseRepository provides common repository functionality
type BaseRepository struct {
	db *DB
	tx *sql.Tx
}

// NewBaseRepository creates a new base repository
func NewBaseRepository(db *DB) *BaseRepository {
	return &BaseRepository{db: db}
}

// DB returns the underlying database connection
func (r *BaseRepository) DB() *DB {
	return r.db
}

// Tx returns the current transaction
func (r *BaseRepository) Tx() *sql.Tx {
	return r.tx
}

// Querier returns the appropriate querier (tx or db)
func (r *BaseRepository) Querier() Querier {
	if r.tx != nil {
		return r.tx
	}
	return r.db.DB
}

// Querier interface for both *sql.DB and *sql.Tx
type Querier interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

// Pagination holds pagination parameters
type Pagination struct {
	Page     int
	PerPage  int
	OrderBy  string
	OrderDir string // "ASC" or "DESC"
}

// DefaultPagination returns default pagination settings
func DefaultPagination() Pagination {
	return Pagination{
		Page:     1,
		PerPage:  20,
		OrderBy:  "created_at",
		OrderDir: "DESC",
	}
}

// Offset calculates the offset for SQL queries
func (p Pagination) Offset() int {
	if p.Page < 1 {
		p.Page = 1
	}
	return (p.Page - 1) * p.PerPage
}

// Limit returns the limit for SQL queries
func (p Pagination) Limit() int {
	if p.PerPage < 1 {
		return 20
	}
	if p.PerPage > 100 {
		return 100
	}
	return p.PerPage
}

// PaginatedResult holds paginated query results
type PaginatedResult[T any] struct {
	Items      []T   `json:"items"`
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	PerPage    int   `json:"per_page"`
	TotalPages int   `json:"total_pages"`
}

// NewPaginatedResult creates a new paginated result
func NewPaginatedResult[T any](items []T, total int64, pagination Pagination) PaginatedResult[T] {
	totalPages := int(total) / pagination.Limit()
	if int(total)%pagination.Limit() > 0 {
		totalPages++
	}

	return PaginatedResult[T]{
		Items:      items,
		Total:      total,
		Page:       pagination.Page,
		PerPage:    pagination.Limit(),
		TotalPages: totalPages,
	}
}

// Filter represents query filters
type Filter struct {
	Field    string
	Operator string // "=", "!=", ">", "<", ">=", "<=", "LIKE", "IN", "IS NULL", "IS NOT NULL"
	Value    interface{}
}

// QueryBuilder helps build SQL queries
type QueryBuilder struct {
	table      string
	columns    []string
	filters    []Filter
	orderBy    string
	orderDir   string
	limit      int
	offset     int
	args       []interface{}
	argCounter int
}

// NewQueryBuilder creates a new query builder
func NewQueryBuilder(table string) *QueryBuilder {
	return &QueryBuilder{
		table:      table,
		columns:    []string{"*"},
		filters:    make([]Filter, 0),
		argCounter: 1,
	}
}

// Select sets the columns to select
func (qb *QueryBuilder) Select(columns ...string) *QueryBuilder {
	qb.columns = columns
	return qb
}

// Where adds a filter condition
func (qb *QueryBuilder) Where(field, operator string, value interface{}) *QueryBuilder {
	qb.filters = append(qb.filters, Filter{
		Field:    field,
		Operator: operator,
		Value:    value,
	})
	return qb
}

// OrderBy sets the order by clause
func (qb *QueryBuilder) OrderBy(field, direction string) *QueryBuilder {
	qb.orderBy = field
	qb.orderDir = direction
	return qb
}

// Limit sets the limit
func (qb *QueryBuilder) Limit(limit int) *QueryBuilder {
	qb.limit = limit
	return qb
}

// Offset sets the offset
func (qb *QueryBuilder) Offset(offset int) *QueryBuilder {
	qb.offset = offset
	return qb
}

// Paginate sets pagination
func (qb *QueryBuilder) Paginate(p Pagination) *QueryBuilder {
	qb.limit = p.Limit()
	qb.offset = p.Offset()
	if p.OrderBy != "" {
		qb.orderBy = p.OrderBy
		qb.orderDir = p.OrderDir
	}
	return qb
}
