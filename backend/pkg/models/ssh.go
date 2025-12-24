// Package models provides SSH-related data models for OweHost
package models

import "time"

// SSHKey represents an SSH public key
type SSHKey struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	Name        string    `json:"name"`
	PublicKey   string    `json:"public_key"`
	Fingerprint string    `json:"fingerprint"`
	KeyType     string    `json:"key_type"` // rsa, ed25519, ecdsa
	BitSize     int       `json:"bit_size,omitempty"`
	Comment     string    `json:"comment,omitempty"`
	LastUsedAt  *time.Time `json:"last_used_at,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

// SSHKeyCreateRequest represents a request to add an SSH key
type SSHKeyCreateRequest struct {
	Name      string `json:"name"`
	PublicKey string `json:"public_key"`
}

// SSHKeyGenerateRequest represents a request to generate a new SSH key pair
type SSHKeyGenerateRequest struct {
	Name       string `json:"name"`
	KeyType    string `json:"key_type"`    // rsa, ed25519
	BitSize    int    `json:"bit_size"`    // for RSA: 2048, 4096
	Passphrase string `json:"passphrase"`  // optional
}

// SSHKeyPair represents a generated SSH key pair
type SSHKeyPair struct {
	PublicKey  string `json:"public_key"`
	PrivateKey string `json:"private_key"` // only returned on generation
	Fingerprint string `json:"fingerprint"`
}

// SSHAccess represents SSH access configuration for a user
type SSHAccess struct {
	ID            string    `json:"id"`
	UserID        string    `json:"user_id"`
	Enabled       bool      `json:"enabled"`
	Shell         string    `json:"shell"`         // /bin/bash, /bin/sh, /usr/bin/jailshell
	Jailed        bool      `json:"jailed"`        // chroot jail
	IPWhitelist   []string  `json:"ip_whitelist,omitempty"`
	LastLoginAt   *time.Time `json:"last_login_at,omitempty"`
	LastLoginIP   string    `json:"last_login_ip,omitempty"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// SSHAccessUpdateRequest represents a request to update SSH access
type SSHAccessUpdateRequest struct {
	Enabled     *bool     `json:"enabled,omitempty"`
	Shell       *string   `json:"shell,omitempty"`
	Jailed      *bool     `json:"jailed,omitempty"`
	IPWhitelist *[]string `json:"ip_whitelist,omitempty"`
}

// SSHSession represents an active SSH session
type SSHSession struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Username  string    `json:"username"`
	RemoteIP  string    `json:"remote_ip"`
	PTY       string    `json:"pty"`
	StartedAt time.Time `json:"started_at"`
	Command   string    `json:"command,omitempty"` // if running a specific command
}
