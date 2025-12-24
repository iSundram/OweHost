// Package licensing provides license and feature flag management for OweHost
package licensing

import (
	"errors"
	"sync"
	"time"

	"github.com/iSundram/OweHost/pkg/models"
	"github.com/iSundram/OweHost/pkg/utils"
)

// Service provides licensing functionality
type Service struct {
	license      *models.License
	featureFlags map[string]*models.FeatureFlag
	lastValidated time.Time
	mu           sync.RWMutex
}

// NewService creates a new licensing service
func NewService() *Service {
	svc := &Service{
		featureFlags: make(map[string]*models.FeatureFlag),
	}
	svc.initDefaultFeatureFlags()
	return svc
}

// initDefaultFeatureFlags initializes default feature flags
func (s *Service) initDefaultFeatureFlags() {
	flags := []*models.FeatureFlag{
		{
			ID:          utils.GenerateID("ff"),
			Name:        "multi_php",
			Description: "Multiple PHP versions support",
			Enabled:     true,
			Plans:       []string{"basic", "pro", "enterprise"},
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          utils.GenerateID("ff"),
			Name:        "nodejs_apps",
			Description: "Node.js application hosting",
			Enabled:     true,
			Plans:       []string{"pro", "enterprise"},
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          utils.GenerateID("ff"),
			Name:        "python_apps",
			Description: "Python application hosting",
			Enabled:     true,
			Plans:       []string{"pro", "enterprise"},
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          utils.GenerateID("ff"),
			Name:        "clustering",
			Description: "Multi-node clustering",
			Enabled:     true,
			Plans:       []string{"enterprise"},
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          utils.GenerateID("ff"),
			Name:        "reseller",
			Description: "Reseller functionality",
			Enabled:     true,
			Plans:       []string{"pro", "enterprise"},
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          utils.GenerateID("ff"),
			Name:        "api_access",
			Description: "API access",
			Enabled:     true,
			Plans:       []string{"basic", "pro", "enterprise"},
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}

	for _, flag := range flags {
		s.featureFlags[flag.Name] = flag
	}
}

// ValidateLicense validates a license
func (s *Service) ValidateLicense(req *models.LicenseValidationRequest) (*models.LicenseValidationResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Simulate license validation
	// In production, would call license server
	license := &models.License{
		ID:         utils.GenerateID("lic"),
		Key:        req.Key,
		Status:     models.LicenseStatusValid,
		ProductID:  req.ProductID,
		CustomerID: "customer-123",
		Plan:       "pro",
		Features: []string{
			"multi_php", "nodejs_apps", "python_apps",
			"reseller", "api_access",
		},
		MaxUsers:      100,
		MaxDomains:    500,
		MaxNodes:      5,
		ExpiresAt:     time.Now().AddDate(1, 0, 0),
		LastValidated: time.Now(),
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	s.license = license
	s.lastValidated = time.Now()

	return &models.LicenseValidationResponse{
		Valid:      true,
		Status:     license.Status,
		ExpiresAt:  license.ExpiresAt,
		Features:   license.Features,
		Plan:       license.Plan,
		MaxUsers:   license.MaxUsers,
		MaxDomains: license.MaxDomains,
		MaxNodes:   license.MaxNodes,
		Message:    "License validated successfully",
	}, nil
}

// GetLicense gets the current license
func (s *Service) GetLicense() (*models.License, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.license == nil {
		return nil, errors.New("no license registered")
	}
	return s.license, nil
}

// IsLicenseValid checks if the current license is valid
func (s *Service) IsLicenseValid() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.license == nil {
		return false
	}

	if s.license.Status != models.LicenseStatusValid && s.license.Status != models.LicenseStatusGrace {
		return false
	}

	if time.Now().After(s.license.ExpiresAt) {
		return false
	}

	return true
}

// CheckGracePeriod checks if license is in grace period
func (s *Service) CheckGracePeriod() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.license == nil {
		return false
	}

	return s.license.Status == models.LicenseStatusGrace
}

// GetPlan gets the current plan
func (s *Service) GetPlan() string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.license == nil {
		return "free"
	}
	return s.license.Plan
}

// GetFeatureFlag gets a feature flag by name
func (s *Service) GetFeatureFlag(name string) (*models.FeatureFlag, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	flag, exists := s.featureFlags[name]
	if !exists {
		return nil, errors.New("feature flag not found")
	}
	return flag, nil
}

// ListFeatureFlags lists all feature flags
func (s *Service) ListFeatureFlags() []*models.FeatureFlag {
	s.mu.RLock()
	defer s.mu.RUnlock()

	flags := make([]*models.FeatureFlag, 0, len(s.featureFlags))
	for _, flag := range s.featureFlags {
		flags = append(flags, flag)
	}
	return flags
}

// CheckFeature checks if a feature is enabled for the current plan
func (s *Service) CheckFeature(req *models.FeatureFlagCheckRequest) *models.FeatureFlagCheckResponse {
	s.mu.RLock()
	defer s.mu.RUnlock()

	flag, exists := s.featureFlags[req.Feature]
	if !exists {
		return &models.FeatureFlagCheckResponse{
			Enabled: false,
			Reason:  "Feature not found",
		}
	}

	if !flag.Enabled {
		return &models.FeatureFlagCheckResponse{
			Enabled: false,
			Reason:  "Feature disabled globally",
		}
	}

	// Check plan-based gating
	currentPlan := s.GetPlan()
	for _, plan := range flag.Plans {
		if plan == currentPlan {
			return &models.FeatureFlagCheckResponse{
				Enabled: true,
				Reason:  "Feature enabled for plan: " + currentPlan,
			}
		}
	}

	return &models.FeatureFlagCheckResponse{
		Enabled: false,
		Reason:  "Feature not available for plan: " + currentPlan,
	}
}

// SetFeatureFlag sets a feature flag
func (s *Service) SetFeatureFlag(name string, enabled bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	flag, exists := s.featureFlags[name]
	if !exists {
		return errors.New("feature flag not found")
	}

	flag.Enabled = enabled
	flag.UpdatedAt = time.Now()
	return nil
}

// CreateFeatureFlag creates a new feature flag
func (s *Service) CreateFeatureFlag(name, description string, plans []string) (*models.FeatureFlag, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.featureFlags[name]; exists {
		return nil, errors.New("feature flag already exists")
	}

	flag := &models.FeatureFlag{
		ID:          utils.GenerateID("ff"),
		Name:        name,
		Description: description,
		Enabled:     true,
		Plans:       plans,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	s.featureFlags[name] = flag
	return flag, nil
}

// DeleteFeatureFlag deletes a feature flag
func (s *Service) DeleteFeatureFlag(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.featureFlags[name]; !exists {
		return errors.New("feature flag not found")
	}

	delete(s.featureFlags, name)
	return nil
}

// CheckLimits checks if within license limits
func (s *Service) CheckLimits(userCount, domainCount, nodeCount int) (bool, string) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.license == nil {
		return false, "No license registered"
	}

	if userCount > s.license.MaxUsers {
		return false, "User limit exceeded"
	}

	if domainCount > s.license.MaxDomains {
		return false, "Domain limit exceeded"
	}

	if nodeCount > s.license.MaxNodes {
		return false, "Node limit exceeded"
	}

	return true, "Within limits"
}
