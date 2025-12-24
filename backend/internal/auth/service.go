// Package auth provides authentication services for OweHost
package auth

import (
	"errors"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/iSundram/OweHost/pkg/config"
	"github.com/iSundram/OweHost/pkg/models"
	"github.com/iSundram/OweHost/pkg/utils"
)

// Service provides authentication functionality
type Service struct {
	config   *config.Config
	users    map[string]*models.User
	sessions map[string]*models.Session
	apiKeys  map[string]*models.APIKey
	mu       sync.RWMutex
}

// Claims represents JWT claims
type Claims struct {
	UserID   string `json:"user_id"`
	TenantID string `json:"tenant_id"`
	jwt.RegisteredClaims
}

// NewService creates a new auth service
func NewService(cfg *config.Config) *Service {
	return &Service{
		config:   cfg,
		users:    make(map[string]*models.User),
		sessions: make(map[string]*models.Session),
		apiKeys:  make(map[string]*models.APIKey),
	}
}

// GenerateTokens generates access and refresh tokens for a user
func (s *Service) GenerateTokens(user *models.User) (*models.LoginResponse, error) {
	// Create access token
	claims := &Claims{
		UserID:   user.ID,
		TenantID: user.TenantID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.config.Auth.JWTExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "owehost",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	accessToken, err := token.SignedString([]byte(s.config.Auth.JWTSecret))
	if err != nil {
		return nil, err
	}

	// Generate refresh token
	refreshToken, err := utils.GenerateRefreshToken()
	if err != nil {
		return nil, err
	}

	// Store session
	session := &models.Session{
		ID:             utils.GenerateID("sess"),
		UserID:         user.ID,
		RefreshToken:   refreshToken,
		CreatedAt:      time.Now(),
		ExpiresAt:      time.Now().Add(s.config.Auth.RefreshTokenExpiry),
		LastAccessedAt: time.Now(),
	}

	s.mu.Lock()
	s.sessions[refreshToken] = session
	s.mu.Unlock()

	return &models.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int(s.config.Auth.JWTExpiry.Seconds()),
		TokenType:    "Bearer",
	}, nil
}

// ValidateToken validates a JWT token and returns the claims
func (s *Service) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(s.config.Auth.JWTSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// RefreshTokens refreshes tokens using a refresh token
func (s *Service) RefreshTokens(refreshToken string) (*models.LoginResponse, error) {
	s.mu.RLock()
	session, exists := s.sessions[refreshToken]
	s.mu.RUnlock()

	if !exists {
		return nil, errors.New("invalid refresh token")
	}

	if time.Now().After(session.ExpiresAt) {
		s.mu.Lock()
		delete(s.sessions, refreshToken)
		s.mu.Unlock()
		return nil, errors.New("refresh token expired")
	}

	// Get user
	s.mu.RLock()
	user, exists := s.users[session.UserID]
	s.mu.RUnlock()

	if !exists {
		return nil, errors.New("user not found")
	}

	// Invalidate old refresh token
	s.mu.Lock()
	delete(s.sessions, refreshToken)
	s.mu.Unlock()

	// Generate new tokens
	return s.GenerateTokens(user)
}

// InvalidateSession invalidates a session
func (s *Service) InvalidateSession(refreshToken string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.sessions[refreshToken]; !exists {
		return errors.New("session not found")
	}

	delete(s.sessions, refreshToken)
	return nil
}

// CreateAPIKey creates a new API key
func (s *Service) CreateAPIKey(userID, name string, scopes []string, expiresAt *time.Time, ipBindings []string) (*models.APIKey, string, error) {
	key, hash := utils.GenerateAPIKey(s.config.Auth.APIKeyPrefix)

	apiKey := &models.APIKey{
		ID:         utils.GenerateID("key"),
		UserID:     userID,
		Name:       name,
		KeyHash:    hash,
		Prefix:     key[:12], // First 12 chars as prefix for identification
		Scopes:     scopes,
		IPBindings: ipBindings,
		ExpiresAt:  expiresAt,
		CreatedAt:  time.Now(),
	}

	s.mu.Lock()
	s.apiKeys[hash] = apiKey
	s.mu.Unlock()

	return apiKey, key, nil
}

// ValidateAPIKey validates an API key
func (s *Service) ValidateAPIKey(key string) (*models.APIKey, error) {
	hash := utils.HashAPIKey(key)

	s.mu.RLock()
	apiKey, exists := s.apiKeys[hash]
	s.mu.RUnlock()

	if !exists {
		return nil, errors.New("invalid API key")
	}

	if apiKey.ExpiresAt != nil && time.Now().After(*apiKey.ExpiresAt) {
		return nil, errors.New("API key expired")
	}

	return apiKey, nil
}

// RegisterUser registers a user for authentication
func (s *Service) RegisterUser(user *models.User) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.users[user.ID] = user
}

// GetUserByID gets a user by ID
func (s *Service) GetUserByID(id string) (*models.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, exists := s.users[id]
	if !exists {
		return nil, errors.New("user not found")
	}
	return user, nil
}
