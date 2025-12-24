// Package appinstaller provides application installation services for OweHost
package appinstaller

import (
	"errors"
	"sync"
	"time"

	"github.com/iSundram/OweHost/pkg/models"
	"github.com/iSundram/OweHost/pkg/utils"
)

// Service provides application installation functionality
type Service struct {
	definitions   map[string]*models.AppDefinition
	installed     map[string]*models.InstalledApp
	byUser        map[string][]*models.InstalledApp
	byDomain      map[string][]*models.InstalledApp
	mu            sync.RWMutex
}

// NewService creates a new app installer service
func NewService() *Service {
	svc := &Service{
		definitions: make(map[string]*models.AppDefinition),
		installed:   make(map[string]*models.InstalledApp),
		byUser:      make(map[string][]*models.InstalledApp),
		byDomain:    make(map[string][]*models.InstalledApp),
	}
	svc.loadDefaultApps()
	return svc
}

// loadDefaultApps loads default application definitions
func (s *Service) loadDefaultApps() {
	apps := []*models.AppDefinition{
		// CMS
		{
			ID:           utils.GenerateID("app"),
			Name:         "WordPress",
			Slug:         "wordpress",
			Version:      "6.4",
			Category:     "cms",
			Description:  "Popular blogging and content management system",
			Requirements: []string{"php>=8.0", "mysql>=5.7"},
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		{
			ID:           utils.GenerateID("app"),
			Name:         "Joomla",
			Slug:         "joomla",
			Version:      "5.0",
			Category:     "cms",
			Description:  "Flexible content management system",
			Requirements: []string{"php>=8.1", "mysql>=5.7"},
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		{
			ID:           utils.GenerateID("app"),
			Name:         "Drupal",
			Slug:         "drupal",
			Version:      "10.2",
			Category:     "cms",
			Description:  "Enterprise content management system",
			Requirements: []string{"php>=8.1", "mysql>=5.7"},
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		{
			ID:           utils.GenerateID("app"),
			Name:         "Ghost",
			Slug:         "ghost",
			Version:      "5.74",
			Category:     "cms",
			Description:  "Professional publishing platform",
			Requirements: []string{"nodejs>=18"},
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		{
			ID:           utils.GenerateID("app"),
			Name:         "Grav",
			Slug:         "grav",
			Version:      "1.7",
			Category:     "cms",
			Description:  "Modern flat-file CMS",
			Requirements: []string{"php>=8.0"},
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		// E-commerce
		{
			ID:           utils.GenerateID("app"),
			Name:         "PrestaShop",
			Slug:         "prestashop",
			Version:      "8.1",
			Category:     "ecommerce",
			Description:  "E-commerce solution",
			Requirements: []string{"php>=8.0", "mysql>=5.7"},
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		{
			ID:           utils.GenerateID("app"),
			Name:         "WooCommerce",
			Slug:         "woocommerce",
			Version:      "8.4",
			Category:     "ecommerce",
			Description:  "WordPress e-commerce plugin",
			Requirements: []string{"php>=8.0", "mysql>=5.7", "wordpress>=6.0"},
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		{
			ID:           utils.GenerateID("app"),
			Name:         "Magento",
			Slug:         "magento",
			Version:      "2.4",
			Category:     "ecommerce",
			Description:  "Enterprise e-commerce platform",
			Requirements: []string{"php>=8.1", "mysql>=8.0", "elasticsearch>=7.0"},
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		{
			ID:           utils.GenerateID("app"),
			Name:         "OpenCart",
			Slug:         "opencart",
			Version:      "4.0",
			Category:     "ecommerce",
			Description:  "Open source shopping cart solution",
			Requirements: []string{"php>=8.0", "mysql>=5.7"},
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		// Forums
		{
			ID:           utils.GenerateID("app"),
			Name:         "phpBB",
			Slug:         "phpbb",
			Version:      "3.3",
			Category:     "forum",
			Description:  "Popular open source forum software",
			Requirements: []string{"php>=7.4", "mysql>=5.7"},
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		{
			ID:           utils.GenerateID("app"),
			Name:         "Discourse",
			Slug:         "discourse",
			Version:      "3.2",
			Category:     "forum",
			Description:  "Modern discussion platform",
			Requirements: []string{"ruby>=3.0", "postgresql>=13"},
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		{
			ID:           utils.GenerateID("app"),
			Name:         "Flarum",
			Slug:         "flarum",
			Version:      "1.8",
			Category:     "forum",
			Description:  "Simple and elegant forum software",
			Requirements: []string{"php>=8.0", "mysql>=5.7"},
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		// Cloud Storage
		{
			ID:           utils.GenerateID("app"),
			Name:         "Nextcloud",
			Slug:         "nextcloud",
			Version:      "28.0",
			Category:     "storage",
			Description:  "Self-hosted productivity platform",
			Requirements: []string{"php>=8.1", "mysql>=8.0"},
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		{
			ID:           utils.GenerateID("app"),
			Name:         "ownCloud",
			Slug:         "owncloud",
			Version:      "10.14",
			Category:     "storage",
			Description:  "File sync and share platform",
			Requirements: []string{"php>=7.4", "mysql>=5.7"},
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		// Frameworks
		{
			ID:           utils.GenerateID("app"),
			Name:         "Laravel",
			Slug:         "laravel",
			Version:      "10.0",
			Category:     "framework",
			Description:  "PHP web application framework",
			Requirements: []string{"php>=8.1", "composer"},
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		{
			ID:           utils.GenerateID("app"),
			Name:         "Symfony",
			Slug:         "symfony",
			Version:      "6.4",
			Category:     "framework",
			Description:  "PHP framework for web applications",
			Requirements: []string{"php>=8.1", "composer"},
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		{
			ID:           utils.GenerateID("app"),
			Name:         "CodeIgniter",
			Slug:         "codeigniter",
			Version:      "4.4",
			Category:     "framework",
			Description:  "Powerful PHP framework",
			Requirements: []string{"php>=8.0"},
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		// Wiki
		{
			ID:           utils.GenerateID("app"),
			Name:         "MediaWiki",
			Slug:         "mediawiki",
			Version:      "1.41",
			Category:     "wiki",
			Description:  "Wiki software powering Wikipedia",
			Requirements: []string{"php>=8.0", "mysql>=5.7"},
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		{
			ID:           utils.GenerateID("app"),
			Name:         "DokuWiki",
			Slug:         "dokuwiki",
			Version:      "2024-02",
			Category:     "wiki",
			Description:  "Simple to use wiki software",
			Requirements: []string{"php>=7.4"},
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		// Project Management
		{
			ID:           utils.GenerateID("app"),
			Name:         "GitLab",
			Slug:         "gitlab",
			Version:      "16.7",
			Category:     "devops",
			Description:  "Complete DevOps platform",
			Requirements: []string{"ruby>=3.0", "postgresql>=13"},
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		{
			ID:           utils.GenerateID("app"),
			Name:         "Gitea",
			Slug:         "gitea",
			Version:      "1.21",
			Category:     "devops",
			Description:  "Lightweight Git hosting solution",
			Requirements: []string{"go>=1.21"},
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		// Other
		{
			ID:           utils.GenerateID("app"),
			Name:         "phpMyAdmin",
			Slug:         "phpmyadmin",
			Version:      "5.2",
			Category:     "database",
			Description:  "MySQL database management tool",
			Requirements: []string{"php>=8.0", "mysql>=5.7"},
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		{
			ID:           utils.GenerateID("app"),
			Name:         "Adminer",
			Slug:         "adminer",
			Version:      "4.8",
			Category:     "database",
			Description:  "Database management in a single file",
			Requirements: []string{"php>=7.4"},
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		{
			ID:           utils.GenerateID("app"),
			Name:         "Matomo",
			Slug:         "matomo",
			Version:      "5.0",
			Category:     "analytics",
			Description:  "Open source web analytics",
			Requirements: []string{"php>=8.0", "mysql>=5.7"},
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
	}

	for _, app := range apps {
		s.definitions[app.ID] = app
	}
}

// ListApps lists available applications
func (s *Service) ListApps(category string) []*models.AppDefinition {
	s.mu.RLock()
	defer s.mu.RUnlock()

	apps := make([]*models.AppDefinition, 0)
	for _, app := range s.definitions {
		if category == "" || app.Category == category {
			apps = append(apps, app)
		}
	}
	return apps
}

// GetApp gets an application definition by ID
func (s *Service) GetApp(id string) (*models.AppDefinition, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	app, exists := s.definitions[id]
	if !exists {
		return nil, errors.New("application not found")
	}
	return app, nil
}

// GetAppBySlug gets an application definition by slug
func (s *Service) GetAppBySlug(slug string) (*models.AppDefinition, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, app := range s.definitions {
		if app.Slug == slug {
			return app, nil
		}
	}
	return nil, errors.New("application not found")
}

// Install installs an application
func (s *Service) Install(userID string, req *models.AppInstallRequest) (*models.InstalledApp, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	appDef, exists := s.definitions[req.AppID]
	if !exists {
		return nil, errors.New("application not found")
	}

	installPath := req.InstallPath
	if installPath == "" {
		installPath = "/home/" + userID + "/public_html/" + appDef.Slug
	}

	installed := &models.InstalledApp{
		ID:          utils.GenerateID("inst"),
		UserID:      userID,
		DomainID:    req.DomainID,
		AppID:       req.AppID,
		Version:     appDef.Version,
		InstallPath: installPath,
		Status:      models.AppInstallStatusPending,
		Settings:    req.Settings,
		InstalledAt: time.Now(),
		UpdatedAt:   time.Now(),
	}

	s.installed[installed.ID] = installed
	s.byUser[userID] = append(s.byUser[userID], installed)
	s.byDomain[req.DomainID] = append(s.byDomain[req.DomainID], installed)

	// Start installation (would be async in production)
	go s.performInstallation(installed.ID)

	return installed, nil
}

// performInstallation performs the actual installation
func (s *Service) performInstallation(installedID string) {
	s.mu.Lock()
	installed := s.installed[installedID]
	if installed == nil {
		s.mu.Unlock()
		return
	}
	installed.Status = models.AppInstallStatusInstalling
	s.mu.Unlock()

	// Simulate installation time
	time.Sleep(100 * time.Millisecond)

	s.mu.Lock()
	defer s.mu.Unlock()

	installed = s.installed[installedID]
	if installed == nil {
		return
	}
	installed.Status = models.AppInstallStatusCompleted
	installed.UpdatedAt = time.Now()
}

// GetInstalled gets an installed application by ID
func (s *Service) GetInstalled(id string) (*models.InstalledApp, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	installed, exists := s.installed[id]
	if !exists {
		return nil, errors.New("installed app not found")
	}
	return installed, nil
}

// ListInstalledByUser lists installed applications for a user
func (s *Service) ListInstalledByUser(userID string) []*models.InstalledApp {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.byUser[userID]
}

// ListInstalledByDomain lists installed applications for a domain
func (s *Service) ListInstalledByDomain(domainID string) []*models.InstalledApp {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.byDomain[domainID]
}

// Uninstall uninstalls an application
func (s *Service) Uninstall(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	installed, exists := s.installed[id]
	if !exists {
		return errors.New("installed app not found")
	}

	// Remove from user's apps
	userApps := s.byUser[installed.UserID]
	for i, app := range userApps {
		if app.ID == id {
			s.byUser[installed.UserID] = append(userApps[:i], userApps[i+1:]...)
			break
		}
	}

	// Remove from domain's apps
	domainApps := s.byDomain[installed.DomainID]
	for i, app := range domainApps {
		if app.ID == id {
			s.byDomain[installed.DomainID] = append(domainApps[:i], domainApps[i+1:]...)
			break
		}
	}

	delete(s.installed, id)
	return nil
}

// Update updates an installed application
func (s *Service) Update(id string, req *models.AppUpdateRequest) (*models.InstalledApp, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	installed, exists := s.installed[id]
	if !exists {
		return nil, errors.New("installed app not found")
	}

	appDef, exists := s.definitions[installed.AppID]
	if !exists {
		return nil, errors.New("application definition not found")
	}

	targetVersion := req.TargetVersion
	if targetVersion == "" {
		targetVersion = appDef.Version
	}

	if installed.Version == targetVersion {
		return nil, errors.New("already on target version")
	}

	installed.Status = models.AppInstallStatusUpdating
	installed.UpdatedAt = time.Now()

	// Perform update (would be async in production)
	go s.performUpdate(id, targetVersion)

	return installed, nil
}

// performUpdate performs the actual update
func (s *Service) performUpdate(installedID, targetVersion string) {
	time.Sleep(100 * time.Millisecond)

	s.mu.Lock()
	defer s.mu.Unlock()

	installed := s.installed[installedID]
	if installed == nil {
		return
	}

	installed.Version = targetVersion
	installed.Status = models.AppInstallStatusCompleted
	installed.UpdatedAt = time.Now()
}

// CheckForUpdates checks if updates are available for installed apps
func (s *Service) CheckForUpdates(userID string) []*models.InstalledApp {
	s.mu.RLock()
	defer s.mu.RUnlock()

	needsUpdate := make([]*models.InstalledApp, 0)

	for _, installed := range s.byUser[userID] {
		appDef := s.definitions[installed.AppID]
		if appDef != nil && appDef.Version != installed.Version {
			needsUpdate = append(needsUpdate, installed)
		}
	}

	return needsUpdate
}

// RegisterApp registers a new application definition
func (s *Service) RegisterApp(app *models.AppDefinition) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if app.ID == "" {
		app.ID = utils.GenerateID("app")
	}
	app.CreatedAt = time.Now()
	app.UpdatedAt = time.Now()

	s.definitions[app.ID] = app
	return nil
}

// ParseManifest parses an application manifest
func (s *Service) ParseManifest(manifestURL string) (*models.AppManifest, error) {
	// In production, would fetch and parse the manifest
	manifest := &models.AppManifest{
		Name:        "Custom App",
		Version:     "1.0.0",
		Description: "Custom application",
		Author:      "Unknown",
		Requirements: map[string]string{
			"php": ">=8.0",
		},
		Files:       []models.FileRequirement{},
		PostInstall: []string{},
	}
	return manifest, nil
}
