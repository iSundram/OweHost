// Package feature provides feature flag management services
package feature

import (
	"fmt"
	"strings"
)

// Service provides feature flag management functionality
type Service struct {
	// In a real implementation, this would have a repository/store dependency
	// For now, we'll use in-memory storage
	features map[string]*Feature
}

// NewService creates a new feature service
func NewService() *Service {
	s := &Service{
		features: make(map[string]*Feature),
	}
	
	// Initialize default features
	s.initializeDefaultFeatures()
	
	return s
}

// Feature represents a feature flag
type Feature struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Description string `json:"description"`
	Enabled     bool   `json:"enabled"`
	Category    string `json:"category"`
}

// FeatureCategory represents a category of features
type FeatureCategory struct {
	Name        string    `json:"name"`
	DisplayName string    `json:"display_name"`
	Features    []Feature `json:"features"`
}

// List returns all features grouped by category
func (s *Service) List() ([]FeatureCategory, error) {
	categories := []FeatureCategory{
		{
			Name:        "web",
			DisplayName: "Web Hosting",
			Features:    s.getFeaturesByCategory("web"),
		},
		{
			Name:        "databases",
			DisplayName: "Databases",
			Features:    s.getFeaturesByCategory("databases"),
		},
		{
			Name:        "runtimes",
			DisplayName: "Runtime Environments",
			Features:    s.getFeaturesByCategory("runtimes"),
		},
		{
			Name:        "services",
			DisplayName: "Additional Services",
			Features:    s.getFeaturesByCategory("services"),
		},
		{
			Name:        "advanced",
			DisplayName: "Advanced Features",
			Features:    s.getFeaturesByCategory("advanced"),
		},
	}

	return categories, nil
}

// Get returns a specific feature by name
func (s *Service) Get(name string) (*Feature, error) {
	feature, exists := s.features[name]
	if !exists {
		return nil, fmt.Errorf("feature not found: %s", name)
	}
	return feature, nil
}

// Update enables/disables a feature
func (s *Service) Update(name string, enabled bool) (*Feature, error) {
	feature, exists := s.features[name]
	if !exists {
		return nil, fmt.Errorf("feature not found: %s", name)
	}

	feature.Enabled = enabled
	s.features[name] = feature

	return feature, nil
}

// IsEnabled checks if a feature is enabled
func (s *Service) IsEnabled(name string) bool {
	feature, exists := s.features[name]
	if !exists {
		return false
	}
	return feature.Enabled
}

// GetFeaturesByCategory returns all features in a category
func (s *Service) GetFeaturesByCategory(category string) []Feature {
	var features []Feature
	for _, feature := range s.features {
		if feature.Category == category {
			features = append(features, *feature)
		}
	}
	return features
}

// initializeDefaultFeatures initializes default feature flags
func (s *Service) initializeDefaultFeatures() {
	defaultFeatures := []*Feature{
		// Web Hosting
		{Name: "domains", DisplayName: "Domain Management", Description: "Add and manage domains", Enabled: true, Category: "web"},
		{Name: "subdomains", DisplayName: "Subdomain Management", Description: "Create and manage subdomains", Enabled: true, Category: "web"},
		{Name: "ssl", DisplayName: "SSL Certificates", Description: "Manage SSL certificates", Enabled: true, Category: "web"},
		{Name: "php", DisplayName: "PHP Support", Description: "Multiple PHP versions", Enabled: true, Category: "web"},

		// Databases
		{Name: "mysql", DisplayName: "MySQL", Description: "MySQL database support", Enabled: true, Category: "databases"},
		{Name: "postgresql", DisplayName: "PostgreSQL", Description: "PostgreSQL database support", Enabled: true, Category: "databases"},
		{Name: "mongodb", DisplayName: "MongoDB", Description: "MongoDB database support", Enabled: false, Category: "databases"},

		// Runtimes
		{Name: "nodejs", DisplayName: "Node.js", Description: "Node.js runtime support", Enabled: true, Category: "runtimes"},
		{Name: "python", DisplayName: "Python", Description: "Python runtime support", Enabled: true, Category: "runtimes"},
		{Name: "ruby", DisplayName: "Ruby", Description: "Ruby runtime support", Enabled: false, Category: "runtimes"},
		{Name: "java", DisplayName: "Java", Description: "Java runtime support", Enabled: false, Category: "runtimes"},

		// Services
		{Name: "email", DisplayName: "Email Accounts", Description: "Create email accounts", Enabled: true, Category: "services"},
		{Name: "ftp", DisplayName: "FTP Accounts", Description: "FTP access", Enabled: true, Category: "services"},
		{Name: "ssh", DisplayName: "SSH Access", Description: "SSH terminal access", Enabled: true, Category: "services"},
		{Name: "git", DisplayName: "Git Integration", Description: "Git repository management", Enabled: true, Category: "services"},
		{Name: "backups", DisplayName: "Backups", Description: "Automated backups", Enabled: true, Category: "services"},
		{Name: "cron", DisplayName: "Cron Jobs", Description: "Scheduled tasks", Enabled: true, Category: "services"},

		// Advanced
		{Name: "redis", DisplayName: "Redis Cache", Description: "Redis caching support", Enabled: true, Category: "advanced"},
		{Name: "cdn", DisplayName: "CDN Integration", Description: "Content delivery network", Enabled: false, Category: "advanced"},
		{Name: "load_balancer", DisplayName: "Load Balancer", Description: "Load balancing", Enabled: false, Category: "advanced"},
		{Name: "monitoring", DisplayName: "Monitoring", Description: "Advanced monitoring", Enabled: true, Category: "advanced"},
	}

	for _, feature := range defaultFeatures {
		s.features[feature.Name] = feature
	}
}

// getFeaturesByCategory returns features for a category
func (s *Service) getFeaturesByCategory(category string) []Feature {
	var features []Feature
	for _, feature := range s.features {
		if feature.Category == category {
			features = append(features, *feature)
		}
	}
	return features
}

// GetAllFeatures returns all features as a flat list
func (s *Service) GetAllFeatures() []Feature {
	var features []Feature
	for _, feature := range s.features {
		features = append(features, *feature)
	}
	return features
}

// Enable enables a feature
func (s *Service) Enable(name string) error {
	_, err := s.Update(name, true)
	return err
}

// Disable disables a feature
func (s *Service) Disable(name string) error {
	_, err := s.Update(name, false)
	return err
}

// GetCategoryDisplayName returns display name for a category
func (s *Service) GetCategoryDisplayName(category string) string {
	displayNames := map[string]string{
		"web":       "Web Hosting",
		"databases": "Databases",
		"runtimes":  "Runtime Environments",
		"services":  "Additional Services",
		"advanced":  "Advanced Features",
	}
	if name, ok := displayNames[category]; ok {
		return name
	}
	return strings.Title(category)
}
