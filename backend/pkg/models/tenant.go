package models

import "time"

// Tenant represents a tenant for multi-tenancy isolation
type Tenant struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Namespace string    `json:"namespace"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TenantCreateRequest represents a request to create a tenant
type TenantCreateRequest struct {
	Name      string `json:"name" validate:"required,min=2,max=64"`
	Namespace string `json:"namespace" validate:"required,min=2,max=32,alphanum"`
}
