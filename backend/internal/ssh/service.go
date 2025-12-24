// Package ssh provides SSH key and access management for OweHost
package ssh

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/iSundram/OweHost/pkg/models"
	"github.com/iSundram/OweHost/pkg/utils"
	"golang.org/x/crypto/ssh"
)

// Service provides SSH management functionality
type Service struct {
	keys       map[string]*models.SSHKey
	access     map[string]*models.SSHAccess
	byUser     map[string][]*models.SSHKey
	sessions   map[string]*models.SSHSession
	mu         sync.RWMutex
}

// NewService creates a new SSH service
func NewService() *Service {
	return &Service{
		keys:     make(map[string]*models.SSHKey),
		access:   make(map[string]*models.SSHAccess),
		byUser:   make(map[string][]*models.SSHKey),
		sessions: make(map[string]*models.SSHSession),
	}
}

// AddKey adds an SSH public key for a user
func (s *Service) AddKey(userID string, req *models.SSHKeyCreateRequest) (*models.SSHKey, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Parse and validate the public key
	pubKey, comment, _, _, err := ssh.ParseAuthorizedKey([]byte(req.PublicKey))
	if err != nil {
		return nil, fmt.Errorf("invalid public key: %v", err)
	}

	// Calculate fingerprint
	fingerprint := ssh.FingerprintSHA256(pubKey)

	// Check for duplicate fingerprint
	for _, key := range s.byUser[userID] {
		if key.Fingerprint == fingerprint {
			return nil, errors.New("key already exists")
		}
	}

	// Determine key type and size
	keyType := pubKey.Type()
	var bitSize int
	switch keyType {
	case "ssh-rsa":
		// Extract bit size from RSA key (approximate)
		bitSize = len(pubKey.Marshal()) * 8 / 2
	case "ssh-ed25519":
		bitSize = 256
	case "ecdsa-sha2-nistp256":
		bitSize = 256
	case "ecdsa-sha2-nistp384":
		bitSize = 384
	case "ecdsa-sha2-nistp521":
		bitSize = 521
	}

	sshKey := &models.SSHKey{
		ID:          utils.GenerateID("sshk"),
		UserID:      userID,
		Name:        req.Name,
		PublicKey:   req.PublicKey,
		Fingerprint: fingerprint,
		KeyType:     keyType,
		BitSize:     bitSize,
		Comment:     comment,
		CreatedAt:   time.Now(),
	}

	s.keys[sshKey.ID] = sshKey
	s.byUser[userID] = append(s.byUser[userID], sshKey)

	return sshKey, nil
}

// GenerateKey generates a new SSH key pair
func (s *Service) GenerateKey(userID string, req *models.SSHKeyGenerateRequest) (*models.SSHKey, *models.SSHKeyPair, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var pubKey ssh.PublicKey
	var privateKeyPEM string

	switch req.KeyType {
	case "ed25519":
		// Generate Ed25519 key pair
		pubKeyRaw, privKeyRaw, err := ed25519.GenerateKey(rand.Reader)
		if err != nil {
			return nil, nil, err
		}

		// Create SSH public key
		sshPubKey, err := ssh.NewPublicKey(pubKeyRaw)
		if err != nil {
			return nil, nil, err
		}
		pubKey = sshPubKey

		// Encode private key to PEM (simplified - in production use proper encoding)
		block := &pem.Block{
			Type:  "OPENSSH PRIVATE KEY",
			Bytes: privKeyRaw,
		}
		privateKeyPEM = string(pem.EncodeToMemory(block))

	default:
		return nil, nil, errors.New("unsupported key type, use 'ed25519'")
	}

	// Calculate fingerprint
	fingerprint := ssh.FingerprintSHA256(pubKey)

	// Generate authorized_keys format
	pubKeyStr := string(ssh.MarshalAuthorizedKey(pubKey))
	pubKeyStr = strings.TrimSpace(pubKeyStr) + " " + req.Name

	sshKey := &models.SSHKey{
		ID:          utils.GenerateID("sshk"),
		UserID:      userID,
		Name:        req.Name,
		PublicKey:   pubKeyStr,
		Fingerprint: fingerprint,
		KeyType:     pubKey.Type(),
		BitSize:     256,
		CreatedAt:   time.Now(),
	}

	s.keys[sshKey.ID] = sshKey
	s.byUser[userID] = append(s.byUser[userID], sshKey)

	keyPair := &models.SSHKeyPair{
		PublicKey:   pubKeyStr,
		PrivateKey:  privateKeyPEM,
		Fingerprint: fingerprint,
	}

	return sshKey, keyPair, nil
}

// GetKey gets an SSH key by ID
func (s *Service) GetKey(id string) (*models.SSHKey, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	key, exists := s.keys[id]
	if !exists {
		return nil, errors.New("SSH key not found")
	}
	return key, nil
}

// ListKeys lists SSH keys for a user
func (s *Service) ListKeys(userID string) []*models.SSHKey {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.byUser[userID]
}

// DeleteKey deletes an SSH key
func (s *Service) DeleteKey(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key, exists := s.keys[id]
	if !exists {
		return errors.New("SSH key not found")
	}

	// Remove from user's keys
	userKeys := s.byUser[key.UserID]
	for i, k := range userKeys {
		if k.ID == id {
			s.byUser[key.UserID] = append(userKeys[:i], userKeys[i+1:]...)
			break
		}
	}

	delete(s.keys, id)
	return nil
}

// GetAccess gets SSH access configuration for a user
func (s *Service) GetAccess(userID string) (*models.SSHAccess, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	access, exists := s.access[userID]
	if !exists {
		// Return default access configuration
		return &models.SSHAccess{
			UserID:  userID,
			Enabled: false,
			Shell:   "/bin/bash",
			Jailed:  true,
		}, nil
	}
	return access, nil
}

// UpdateAccess updates SSH access configuration
func (s *Service) UpdateAccess(userID string, req *models.SSHAccessUpdateRequest) (*models.SSHAccess, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	access, exists := s.access[userID]
	if !exists {
		access = &models.SSHAccess{
			ID:      utils.GenerateID("ssha"),
			UserID:  userID,
			Enabled: false,
			Shell:   "/bin/bash",
			Jailed:  true,
		}
		s.access[userID] = access
	}

	if req.Enabled != nil {
		access.Enabled = *req.Enabled
	}
	if req.Shell != nil {
		// Validate shell
		validShells := []string{"/bin/bash", "/bin/sh", "/usr/bin/jailshell", "/usr/bin/nologin"}
		valid := false
		for _, shell := range validShells {
			if *req.Shell == shell {
				valid = true
				break
			}
		}
		if !valid {
			return nil, errors.New("invalid shell")
		}
		access.Shell = *req.Shell
	}
	if req.Jailed != nil {
		access.Jailed = *req.Jailed
	}
	if req.IPWhitelist != nil {
		access.IPWhitelist = *req.IPWhitelist
	}

	access.UpdatedAt = time.Now()
	return access, nil
}

// EnableAccess enables SSH access for a user
func (s *Service) EnableAccess(userID string) error {
	enabled := true
	_, err := s.UpdateAccess(userID, &models.SSHAccessUpdateRequest{Enabled: &enabled})
	return err
}

// DisableAccess disables SSH access for a user
func (s *Service) DisableAccess(userID string) error {
	enabled := false
	_, err := s.UpdateAccess(userID, &models.SSHAccessUpdateRequest{Enabled: &enabled})
	return err
}

// ValidateAccess validates if a user can SSH with the given key
func (s *Service) ValidateAccess(userID, fingerprint, remoteIP string) (*models.SSHAccess, *models.SSHKey, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	access, exists := s.access[userID]
	if !exists || !access.Enabled {
		return nil, nil, errors.New("SSH access not enabled")
	}

	// Check IP whitelist
	if len(access.IPWhitelist) > 0 {
		allowed := false
		for _, ip := range access.IPWhitelist {
			if ip == remoteIP {
				allowed = true
				break
			}
		}
		if !allowed {
			return nil, nil, errors.New("IP not allowed")
		}
	}

	// Find matching key
	var matchingKey *models.SSHKey
	for _, key := range s.byUser[userID] {
		if key.Fingerprint == fingerprint {
			matchingKey = key
			break
		}
	}

	if matchingKey == nil {
		return nil, nil, errors.New("key not found")
	}

	return access, matchingKey, nil
}

// RecordKeyUse records the use of an SSH key
func (s *Service) RecordKeyUse(keyID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key, exists := s.keys[keyID]
	if !exists {
		return errors.New("SSH key not found")
	}

	now := time.Now()
	key.LastUsedAt = &now
	return nil
}

// RecordLogin records an SSH login
func (s *Service) RecordLogin(userID, remoteIP string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	access, exists := s.access[userID]
	if !exists {
		return errors.New("access config not found")
	}

	now := time.Now()
	access.LastLoginAt = &now
	access.LastLoginIP = remoteIP
	return nil
}

// GetActiveSessions returns active SSH sessions
func (s *Service) GetActiveSessions() []*models.SSHSession {
	s.mu.RLock()
	defer s.mu.RUnlock()

	sessions := make([]*models.SSHSession, 0, len(s.sessions))
	for _, session := range s.sessions {
		sessions = append(sessions, session)
	}
	return sessions
}

// DeleteAllByUser deletes all SSH keys and access for a user
func (s *Service) DeleteAllByUser(userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, key := range s.byUser[userID] {
		delete(s.keys, key.ID)
	}
	delete(s.byUser, userID)
	delete(s.access, userID)
	return nil
}

// generateFingerprint generates SHA256 fingerprint for a key (helper)
func generateFingerprint(pubKeyBytes []byte) string {
	hash := sha256.Sum256(pubKeyBytes)
	return "SHA256:" + base64.StdEncoding.EncodeToString(hash[:])
}
