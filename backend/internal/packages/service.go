// Package packages provides package/plan management services
package packages

import (
	"fmt"

	"github.com/iSundram/OweHost/internal/storage/account"
)

// Service provides package/plan management functionality
type Service struct {
	// No dependencies needed - uses account package constants
}

// NewService creates a new package service
func NewService() *Service {
	return &Service{}
}

// Package represents an account package/plan
type Package struct {
	Name        string                  `json:"name"`
	DisplayName string                  `json:"display_name"`
	Description string                  `json:"description"`
	Limits      account.ResourceLimits  `json:"limits"`
	Features    map[string]bool         `json:"features"`
	Price       *PackagePrice           `json:"price,omitempty"`
}

// PackagePrice represents pricing information
type PackagePrice struct {
	Monthly  float64 `json:"monthly"`
	Yearly   float64 `json:"yearly"`
	Currency string  `json:"currency"`
}

// List returns all available packages
func (s *Service) List() ([]Package, error) {
	packages := []Package{
		{
			Name:        "starter",
			DisplayName: "Starter",
			Description: "Perfect for small websites and personal projects",
			Limits:      account.PlanLimits["starter"],
			Features:    s.getPackageFeatures("starter"),
		},
		{
			Name:        "standard",
			DisplayName: "Standard",
			Description: "Ideal for growing businesses and websites",
			Limits:      account.PlanLimits["standard"],
			Features:    s.getPackageFeatures("standard"),
		},
		{
			Name:        "premium",
			DisplayName: "Premium",
			Description: "For high-traffic websites and applications",
			Limits:      account.PlanLimits["premium"],
			Features:    s.getPackageFeatures("premium"),
		},
		{
			Name:        "enterprise",
			DisplayName: "Enterprise",
			Description: "Unlimited resources for enterprise needs",
			Limits:      account.PlanLimits["enterprise"],
			Features:    s.getPackageFeatures("enterprise"),
		},
	}

	return packages, nil
}

// Get returns a specific package by name
func (s *Service) Get(name string) (*Package, error) {
	if !account.IsValidPlan(name) {
		return nil, fmt.Errorf("package not found: %s", name)
	}

	limits := account.GetPlanLimits(name)

	displayNames := map[string]string{
		"starter":    "Starter",
		"standard":   "Standard",
		"premium":    "Premium",
		"enterprise": "Enterprise",
	}

	descriptions := map[string]string{
		"starter":    "Perfect for small websites and personal projects",
		"standard":   "Ideal for growing businesses and websites",
		"premium":    "For high-traffic websites and applications",
		"enterprise": "Unlimited resources for enterprise needs",
	}

	pkg := &Package{
		Name:        name,
		DisplayName: displayNames[name],
		Description: descriptions[name],
		Limits:      limits,
		Features:    s.getPackageFeatures(name),
	}

	return pkg, nil
}

// getPackageFeatures returns features for a package
func (s *Service) getPackageFeatures(packageName string) map[string]bool {
	baseFeatures := map[string]bool{
		"domains":      true,
		"databases":    true,
		"email":        true,
		"ftp":          true,
		"ssl":          true,
		"backups":      true,
		"cron":         true,
		"file_manager": true,
	}

	switch packageName {
	case "standard":
		baseFeatures["git"] = true
		baseFeatures["nodejs"] = true
	case "premium":
		baseFeatures["git"] = true
		baseFeatures["nodejs"] = true
		baseFeatures["python"] = true
		baseFeatures["redis"] = true
	case "enterprise":
		baseFeatures["git"] = true
		baseFeatures["nodejs"] = true
		baseFeatures["python"] = true
		baseFeatures["redis"] = true
		baseFeatures["load_balancer"] = true
		baseFeatures["cdn"] = true
		baseFeatures["monitoring"] = true
	}

	return baseFeatures
}

// ValidatePackage validates if a package name is valid
func (s *Service) ValidatePackage(name string) bool {
	return account.IsValidPlan(name)
}

// GetPackageLimits returns resource limits for a package
func (s *Service) GetPackageLimits(name string) (account.ResourceLimits, error) {
	if !account.IsValidPlan(name) {
		return account.ResourceLimits{}, fmt.Errorf("invalid package: %s", name)
	}
	return account.GetPlanLimits(name), nil
}
