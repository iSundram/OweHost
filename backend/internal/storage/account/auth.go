package account

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"os"
	"path/filepath"
	"time"
)

// AuthData represents account authentication information
type AuthData struct {
	PasswordHash string  `json:"password_hash"`
	Salt         string  `json:"salt"`
	LastLogin    *string `json:"last_login,omitempty"`
	LoginCount   int     `json:"login_count"`
	CreatedAt    string  `json:"created_at"`
	UpdatedAt    string  `json:"updated_at"`
}

// SetPassword sets the password for an account
func (s *StateManager) SetPassword(accountID int, password string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Generate salt
	salt := make([]byte, 16)
	rand.Read(salt)
	saltHex := hex.EncodeToString(salt)

	// Hash password with bcrypt
	hash, err := bcrypt.GenerateFromPassword([]byte(password+saltHex), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	authData := &AuthData{
		PasswordHash: string(hash),
		Salt:         saltHex,
		LoginCount:   0,
		CreatedAt:    time.Now().Format(time.RFC3339),
		UpdatedAt:    time.Now().Format(time.RFC3339),
	}

	path := filepath.Join(s.AccountPath(accountID), "auth.json")
	return s.atomicWrite(path, authData)
}

// VerifyPassword verifies a password for an account
func (s *StateManager) VerifyPassword(accountID int, password string) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	path := filepath.Join(s.AccountPath(accountID), "auth.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return false, fmt.Errorf("auth data not found: %w", err)
	}

	var authData AuthData
	if err := json.Unmarshal(data, &authData); err != nil {
		return false, fmt.Errorf("failed to parse auth data: %w", err)
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(authData.PasswordHash), []byte(password+authData.Salt))
	if err != nil {
		return false, nil
	}

	// Update login stats
	go s.updateLoginStats(accountID)
	return true, nil
}

// updateLoginStats updates login statistics
func (s *StateManager) updateLoginStats(accountID int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	path := filepath.Join(s.AccountPath(accountID), "auth.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}

	var authData AuthData
	if err := json.Unmarshal(data, &authData); err != nil {
		return
	}

	now := time.Now().Format(time.RFC3339)
	authData.LastLogin = &now
	authData.LoginCount++
	authData.UpdatedAt = now

	s.atomicWrite(path, &authData)
}

// AuthenticateAccount authenticates an account by username/password
func (s *StateManager) AuthenticateAccount(username, password string) (*AccountIdentity, error) {
	accounts, err := s.ListAccounts()
	if err != nil {
		return nil, err
	}

	for _, id := range accounts {
		identity, err := s.ReadIdentity(id)
		if err != nil {
			continue
		}

		if identity.Name == username {
			valid, err := s.VerifyPassword(id, password)
			if err != nil {
				return nil, err
			}
			if valid {
				return identity, nil
			}
			return nil, fmt.Errorf("invalid password")
		}
	}

	return nil, fmt.Errorf("account not found")
}
