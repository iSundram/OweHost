// Package twofactor provides two-factor authentication for OweHost
package twofactor

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base32"
	"encoding/binary"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/iSundram/OweHost/pkg/models"
	"github.com/iSundram/OweHost/pkg/utils"
)

const (
	// TOTP configuration
	totpDigits   = 6
	totpPeriod   = 30 // seconds
	totpIssuer   = "OweHost"
	
	// Backup codes
	backupCodeCount  = 10
	backupCodeLength = 8
)

// Service provides 2FA functionality
type Service struct {
	configs      map[string]*models.TwoFactorConfig
	attempts     []*models.LoginAttempt
	pendingSetup map[string]string // userID -> secret (before verification)
	mu           sync.RWMutex
}

// NewService creates a new 2FA service
func NewService() *Service {
	return &Service{
		configs:      make(map[string]*models.TwoFactorConfig),
		attempts:     make([]*models.LoginAttempt, 0),
		pendingSetup: make(map[string]string),
	}
}

// SetupTOTP initiates TOTP setup for a user
func (s *Service) SetupTOTP(userID, username string) (*models.TOTPSetupResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if already enabled
	if cfg, exists := s.configs[userID]; exists && cfg.Enabled {
		return nil, errors.New("2FA already enabled")
	}

	// Generate secret
	secret := make([]byte, 20)
	if _, err := rand.Read(secret); err != nil {
		return nil, err
	}
	secretBase32 := base32.StdEncoding.EncodeToString(secret)

	// Store pending setup
	s.pendingSetup[userID] = secretBase32

	// Generate QR code URL (otpauth format)
	qrURL := fmt.Sprintf("otpauth://totp/%s:%s?secret=%s&issuer=%s&digits=%d&period=%d",
		totpIssuer, username, secretBase32, totpIssuer, totpDigits, totpPeriod)

	// Generate backup codes
	backupCodes := s.generateBackupCodes()

	return &models.TOTPSetupResponse{
		Secret:      secretBase32,
		QRCodeURL:   qrURL,
		BackupCodes: backupCodes,
	}, nil
}

// VerifySetup verifies TOTP setup and enables 2FA
func (s *Service) VerifySetup(userID, code string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	secret, exists := s.pendingSetup[userID]
	if !exists {
		return errors.New("no pending 2FA setup")
	}

	// Verify the code
	if !s.verifyTOTP(secret, code) {
		return errors.New("invalid verification code")
	}

	// Generate and hash backup codes
	backupCodes := s.generateBackupCodes()
	hashedCodes := make([]string, len(backupCodes))
	for i, code := range backupCodes {
		hash, _ := utils.HashPassword(code)
		hashedCodes[i] = hash
	}

	// Create config
	now := time.Now()
	config := &models.TwoFactorConfig{
		ID:          utils.GenerateID("2fa"),
		UserID:      userID,
		Enabled:     true,
		Type:        models.TwoFactorTypeTOTP,
		Secret:      secret,
		BackupCodes: hashedCodes,
		VerifiedAt:  &now,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	s.configs[userID] = config
	delete(s.pendingSetup, userID)

	return nil
}

// VerifyCode verifies a TOTP code or backup code
func (s *Service) VerifyCode(userID, code string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	config, exists := s.configs[userID]
	if !exists || !config.Enabled {
		return false, errors.New("2FA not enabled")
	}

	// Try TOTP first
	if s.verifyTOTP(config.Secret, code) {
		now := time.Now()
		config.LastUsedAt = &now
		return true, nil
	}

	// Try backup codes
	code = strings.ReplaceAll(code, "-", "")
	for i, hashedCode := range config.BackupCodes {
		if hashedCode != "" && utils.CheckPassword(code, hashedCode) {
			// Invalidate used backup code
			config.BackupCodes[i] = ""
			config.BackupCodesUsed++
			now := time.Now()
			config.LastUsedAt = &now
			return true, nil
		}
	}

	return false, errors.New("invalid code")
}

// IsEnabled checks if 2FA is enabled for a user
func (s *Service) IsEnabled(userID string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	config, exists := s.configs[userID]
	return exists && config.Enabled
}

// GetConfig gets 2FA configuration for a user
func (s *Service) GetConfig(userID string) (*models.TwoFactorConfig, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	config, exists := s.configs[userID]
	if !exists {
		return nil, errors.New("2FA not configured")
	}
	return config, nil
}

// Disable disables 2FA for a user
func (s *Service) Disable(userID, code string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	config, exists := s.configs[userID]
	if !exists || !config.Enabled {
		return errors.New("2FA not enabled")
	}

	// Verify code before disabling
	if !s.verifyTOTP(config.Secret, code) {
		return errors.New("invalid code")
	}

	delete(s.configs, userID)
	return nil
}

// RegenerateBackupCodes generates new backup codes
func (s *Service) RegenerateBackupCodes(userID, code string) ([]string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	config, exists := s.configs[userID]
	if !exists || !config.Enabled {
		return nil, errors.New("2FA not enabled")
	}

	// Verify code
	if !s.verifyTOTP(config.Secret, code) {
		return nil, errors.New("invalid code")
	}

	// Generate new backup codes
	backupCodes := s.generateBackupCodes()
	hashedCodes := make([]string, len(backupCodes))
	for i, code := range backupCodes {
		hash, _ := utils.HashPassword(code)
		hashedCodes[i] = hash
	}

	config.BackupCodes = hashedCodes
	config.BackupCodesUsed = 0
	config.UpdatedAt = time.Now()

	return backupCodes, nil
}

// GetBackupCodesRemaining returns how many backup codes are remaining
func (s *Service) GetBackupCodesRemaining(userID string) (int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	config, exists := s.configs[userID]
	if !exists {
		return 0, errors.New("2FA not configured")
	}

	remaining := 0
	for _, code := range config.BackupCodes {
		if code != "" {
			remaining++
		}
	}
	return remaining, nil
}

// RecordLoginAttempt records a login attempt
func (s *Service) RecordLoginAttempt(attempt *models.LoginAttempt) {
	s.mu.Lock()
	defer s.mu.Unlock()

	attempt.ID = utils.GenerateID("la")
	attempt.CreatedAt = time.Now()
	s.attempts = append(s.attempts, attempt)

	// Keep only last 1000 attempts
	if len(s.attempts) > 1000 {
		s.attempts = s.attempts[len(s.attempts)-1000:]
	}
}

// GetLoginAttempts gets recent login attempts for a user
func (s *Service) GetLoginAttempts(userID string, limit int) []*models.LoginAttempt {
	s.mu.RLock()
	defer s.mu.RUnlock()

	attempts := make([]*models.LoginAttempt, 0)
	for i := len(s.attempts) - 1; i >= 0 && len(attempts) < limit; i-- {
		if s.attempts[i].UserID == userID {
			attempts = append(attempts, s.attempts[i])
		}
	}
	return attempts
}

// GetFailedAttempts counts failed login attempts from an IP in the last duration
func (s *Service) GetFailedAttempts(ip string, duration time.Duration) int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	cutoff := time.Now().Add(-duration)
	count := 0
	for _, attempt := range s.attempts {
		if attempt.IPAddress == ip && !attempt.Success && attempt.CreatedAt.After(cutoff) {
			count++
		}
	}
	return count
}

// verifyTOTP verifies a TOTP code against a secret
func (s *Service) verifyTOTP(secret, code string) bool {
	// Allow 1 time step before and after for clock drift
	now := time.Now().Unix()
	for i := int64(-1); i <= 1; i++ {
		counter := (now / totpPeriod) + i
		if s.generateTOTP(secret, counter) == code {
			return true
		}
	}
	return false
}

// generateTOTP generates a TOTP code for a given counter
func (s *Service) generateTOTP(secret string, counter int64) string {
	// Decode secret
	key, err := base32.StdEncoding.DecodeString(strings.ToUpper(secret))
	if err != nil {
		return ""
	}

	// Counter to bytes
	counterBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(counterBytes, uint64(counter))

	// HMAC-SHA1
	h := hmac.New(sha1.New, key)
	h.Write(counterBytes)
	hash := h.Sum(nil)

	// Dynamic truncation
	offset := hash[len(hash)-1] & 0x0f
	code := binary.BigEndian.Uint32(hash[offset:offset+4]) & 0x7fffffff

	// Generate digits
	return fmt.Sprintf("%0*d", totpDigits, code%1000000)
}

// generateBackupCodes generates random backup codes
func (s *Service) generateBackupCodes() []string {
	codes := make([]string, backupCodeCount)
	for i := 0; i < backupCodeCount; i++ {
		code := make([]byte, backupCodeLength/2)
		rand.Read(code)
		codeStr := fmt.Sprintf("%X", code)
		// Format as XXXX-XXXX
		codes[i] = codeStr[:4] + "-" + codeStr[4:]
	}
	return codes
}

// DeleteByUser deletes 2FA configuration for a user
func (s *Service) DeleteByUser(userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.configs, userID)
	delete(s.pendingSetup, userID)
	return nil
}
