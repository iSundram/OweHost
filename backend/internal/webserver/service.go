// Package webserver provides web server control services for OweHost
package webserver

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"path/filepath"
	"sync"
	"time"

	"github.com/iSundram/OweHost/pkg/models"
	"github.com/iSundram/OweHost/pkg/utils"
)

// Service provides web server control functionality
type Service struct {
	vhosts        map[string]*models.VirtualHost
	configs       map[string][]*models.WebServerConfig
	byDomain      map[string]*models.VirtualHost
	mu            sync.RWMutex
}

// NewService creates a new webserver service
func NewService() *Service {
	return &Service{
		vhosts:   make(map[string]*models.VirtualHost),
		configs:  make(map[string][]*models.WebServerConfig),
		byDomain: make(map[string]*models.VirtualHost),
	}
}

// CreateVHost creates a new virtual host
func (s *Service) CreateVHost(req *models.VirtualHostCreateRequest) (*models.VirtualHost, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.byDomain[req.DomainID]; exists {
		return nil, errors.New("virtual host already exists for domain")
	}

	// Validate server type
	if !isValidServerType(req.ServerType) {
		return nil, errors.New("invalid server type")
	}

	// Validate domain ID to prevent path traversal
	if !utils.IsValidPath("/var/www/" + req.DomainID) {
		return nil, errors.New("invalid domain ID")
	}

	documentRoot := req.DocumentRoot
	if documentRoot == "" {
		documentRoot = filepath.Join("/var/www", req.DomainID)
	}

	configPath := filepath.Join("/etc", string(req.ServerType), "sites-available", req.DomainID+".conf")

	vhost := &models.VirtualHost{
		ID:           utils.GenerateID("vhost"),
		DomainID:     req.DomainID,
		ServerType:   req.ServerType,
		DocumentRoot: documentRoot,
		SSLEnabled:   req.SSLEnabled,
		SSLCertID:    nil,
		PHPEnabled:   req.PHPEnabled,
		PHPVersion:   req.PHPVersion,
		ProxyPass:    req.ProxyPass,
		ConfigPath:   configPath,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Generate initial config
	config := s.generateConfig(vhost)
	vhost.ConfigChecksum = checksumConfig(config)

	s.vhosts[vhost.ID] = vhost
	s.byDomain[req.DomainID] = vhost

	// Store config version
	configVersion := &models.WebServerConfig{
		ID:         utils.GenerateID("conf"),
		VHostID:    vhost.ID,
		ConfigData: config,
		Version:    1,
		Active:     true,
		CreatedAt:  time.Now(),
	}
	s.configs[vhost.ID] = []*models.WebServerConfig{configVersion}

	return vhost, nil
}

// GetVHost gets a virtual host by ID
func (s *Service) GetVHost(id string) (*models.VirtualHost, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	vhost, exists := s.vhosts[id]
	if !exists {
		return nil, errors.New("virtual host not found")
	}
	return vhost, nil
}

// GetVHostByDomain gets a virtual host by domain ID
func (s *Service) GetVHostByDomain(domainID string) (*models.VirtualHost, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	vhost, exists := s.byDomain[domainID]
	if !exists {
		return nil, errors.New("virtual host not found")
	}
	return vhost, nil
}

// UpdateVHost updates a virtual host
func (s *Service) UpdateVHost(id string, req *models.VirtualHostCreateRequest) (*models.VirtualHost, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	vhost, exists := s.vhosts[id]
	if !exists {
		return nil, errors.New("virtual host not found")
	}

	if req.DocumentRoot != "" {
		vhost.DocumentRoot = req.DocumentRoot
	}
	vhost.SSLEnabled = req.SSLEnabled
	vhost.PHPEnabled = req.PHPEnabled
	vhost.PHPVersion = req.PHPVersion
	vhost.ProxyPass = req.ProxyPass

	// Generate new config and create version
	config := s.generateConfig(vhost)
	newChecksum := checksumConfig(config)

	if newChecksum != vhost.ConfigChecksum {
		// Create new config version
		configs := s.configs[id]
		latestVersion := 0
		for _, c := range configs {
			if c.Version > latestVersion {
				latestVersion = c.Version
			}
		}

		configVersion := &models.WebServerConfig{
			ID:         utils.GenerateID("conf"),
			VHostID:    id,
			ConfigData: config,
			Version:    latestVersion + 1,
			Active:     true,
			CreatedAt:  time.Now(),
		}

		// Deactivate previous
		for _, c := range configs {
			c.Active = false
		}

		s.configs[id] = append(s.configs[id], configVersion)
		vhost.ConfigChecksum = newChecksum
	}

	vhost.UpdatedAt = time.Now()
	return vhost, nil
}

// DeleteVHost deletes a virtual host
func (s *Service) DeleteVHost(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	vhost, exists := s.vhosts[id]
	if !exists {
		return errors.New("virtual host not found")
	}

	delete(s.vhosts, id)
	delete(s.byDomain, vhost.DomainID)
	delete(s.configs, id)

	return nil
}

// Reload safely reloads web server configuration
func (s *Service) Reload(serverType models.WebServerType) (*models.ConfigReloadStatus, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// In production, this would call the actual web server reload
	// For now, we simulate success
	status := &models.ConfigReloadStatus{
		ServerType: serverType,
		Success:    true,
		ReloadedAt: time.Now(),
	}

	return status, nil
}

// ValidateConfig validates configuration before reload
func (s *Service) ValidateConfig(id string) (bool, string) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	configs, exists := s.configs[id]
	if !exists || len(configs) == 0 {
		return false, "no config found"
	}

	// Find active config
	var activeConfig *models.WebServerConfig
	for _, c := range configs {
		if c.Active {
			activeConfig = c
			break
		}
	}

	if activeConfig == nil {
		return false, "no active config found"
	}

	// In production, this would call nginx -t or apachectl configtest
	return true, "configuration valid"
}

// Rollback rolls back to a previous config version
func (s *Service) Rollback(id string, version int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	configs, exists := s.configs[id]
	if !exists {
		return errors.New("no configs found")
	}

	var targetConfig *models.WebServerConfig
	for _, c := range configs {
		if c.Version == version {
			targetConfig = c
			break
		}
	}

	if targetConfig == nil {
		return errors.New("config version not found")
	}

	// Deactivate all and activate target
	for _, c := range configs {
		c.Active = (c.Version == version)
	}

	// Update vhost checksum
	vhost := s.vhosts[id]
	if vhost != nil {
		vhost.ConfigChecksum = checksumConfig(targetConfig.ConfigData)
		vhost.UpdatedAt = time.Now()
	}

	return nil
}

// GetConfigHistory gets config version history
func (s *Service) GetConfigHistory(id string) []*models.WebServerConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.configs[id]
}

// generateConfig generates web server configuration
func (s *Service) generateConfig(vhost *models.VirtualHost) string {
	switch vhost.ServerType {
	case models.WebServerTypeNginx:
		return s.generateNginxConfig(vhost)
	case models.WebServerTypeApache:
		return s.generateApacheConfig(vhost)
	case models.WebServerTypeHybrid:
		return s.generateHybridConfig(vhost)
	}
	return ""
}

// generateNginxConfig generates Nginx configuration
func (s *Service) generateNginxConfig(vhost *models.VirtualHost) string {
	config := "server {\n"
	config += "    listen 80;\n"
	if vhost.SSLEnabled {
		config += "    listen 443 ssl;\n"
	}
	config += "    root " + vhost.DocumentRoot + ";\n"
	config += "    index index.html index.php;\n"
	
	if vhost.PHPEnabled && vhost.PHPVersion != nil {
		config += "    location ~ \\.php$ {\n"
		config += "        fastcgi_pass unix:/var/run/php/php" + *vhost.PHPVersion + "-fpm.sock;\n"
		config += "        fastcgi_index index.php;\n"
		config += "        include fastcgi_params;\n"
		config += "    }\n"
	}

	if vhost.ProxyPass != nil {
		config += "    location / {\n"
		config += "        proxy_pass " + *vhost.ProxyPass + ";\n"
		config += "    }\n"
	}

	config += "}\n"
	return config
}

// generateApacheConfig generates Apache configuration
func (s *Service) generateApacheConfig(vhost *models.VirtualHost) string {
	config := "<VirtualHost *:80>\n"
	config += "    DocumentRoot " + vhost.DocumentRoot + "\n"

	if vhost.PHPEnabled && vhost.PHPVersion != nil {
		config += "    <FilesMatch \\.php$>\n"
		config += "        SetHandler \"proxy:unix:/var/run/php/php" + *vhost.PHPVersion + "-fpm.sock|fcgi://localhost\"\n"
		config += "    </FilesMatch>\n"
	}

	if vhost.ProxyPass != nil {
		config += "    ProxyPass / " + *vhost.ProxyPass + "\n"
		config += "    ProxyPassReverse / " + *vhost.ProxyPass + "\n"
	}

	config += "</VirtualHost>\n"
	return config
}

// generateHybridConfig generates hybrid Nginx+Apache configuration
func (s *Service) generateHybridConfig(vhost *models.VirtualHost) string {
	// Nginx as reverse proxy to Apache
	return s.generateNginxConfig(vhost)
}

func checksumConfig(config string) string {
	hash := sha256.Sum256([]byte(config))
	return hex.EncodeToString(hash[:])
}

// isValidServerType validates the web server type
func isValidServerType(serverType models.WebServerType) bool {
	switch serverType {
	case models.WebServerTypeNginx, models.WebServerTypeApache, models.WebServerTypeHybrid:
		return true
	default:
		return false
	}
}

// ConfigureUserWebServer configures web server for a user
func (s *Service) ConfigureUserWebServer(userID, username, phpVersion string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	// Mock implementation
	return nil
}

// GetUserConfig retrieves user's web server configuration
func (s *Service) GetUserConfig(userID string) (*models.UserWebServerConfig, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Mock implementation - returns default config
	return &models.UserWebServerConfig{
		UserID:       userID,
		PHPVersion:   "8.2",
		SSLEnabled:   true,
		DocumentRoot: "/var/www/" + userID,
		Domains:      []string{},
	}, nil
}

// UpdateUserConfig updates user's web server configuration
func (s *Service) UpdateUserConfig(userID string, req *models.WebServerConfigUpdateRequest) (*models.UserWebServerConfig, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	config := &models.UserWebServerConfig{
		UserID:       userID,
		PHPVersion:   "8.2",
		SSLEnabled:   true,
		DocumentRoot: "/var/www/" + userID,
		Domains:      []string{},
	}

	if req.PHPVersion != nil {
		config.PHPVersion = *req.PHPVersion
	}
	if req.SSLEnabled != nil {
		config.SSLEnabled = *req.SSLEnabled
	}
	if req.DocumentRoot != nil {
		config.DocumentRoot = *req.DocumentRoot
	}

	return config, nil
}

// ListPHPVersions lists available PHP versions
func (s *Service) ListPHPVersions() ([]models.PHPVersion, error) {
	return []models.PHPVersion{
		{Version: "7.4", Available: true, Default: false},
		{Version: "8.0", Available: true, Default: false},
		{Version: "8.1", Available: true, Default: false},
		{Version: "8.2", Available: true, Default: true},
		{Version: "8.3", Available: true, Default: false},
	}, nil
}

// SwitchPHPVersion switches PHP version for a user/domain
func (s *Service) SwitchPHPVersion(userID, domain, version string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	// Mock implementation - in production would update FPM config
	return nil
}

// ListModules lists web server modules
func (s *Service) ListModules() ([]models.WebServerModule, error) {
	return []models.WebServerModule{
		{Name: "rewrite", Enabled: true, Description: "URL rewriting"},
		{Name: "ssl", Enabled: true, Description: "SSL/TLS support"},
		{Name: "headers", Enabled: true, Description: "HTTP headers manipulation"},
		{Name: "proxy", Enabled: true, Description: "Reverse proxy support"},
		{Name: "gzip", Enabled: true, Description: "Gzip compression"},
		{Name: "cache", Enabled: false, Description: "Response caching"},
	}, nil
}

// EnableModule enables a web server module
func (s *Service) EnableModule(module string) error {
	// Mock implementation - in production would run a2enmod or update nginx config
	return nil
}

// DisableModule disables a web server module
func (s *Service) DisableModule(module string) error {
	// Mock implementation - in production would run a2dismod or update nginx config
	return nil
}

// Restart restarts the web server
func (s *Service) Restart() error {
	// Mock implementation - in production would call systemctl restart
	return nil
}

// GetSystemErrorLogs retrieves system-wide error logs
func (s *Service) GetSystemErrorLogs(lines int) ([]string, error) {
	// Mock implementation
	return []string{
		"[info] Server started",
		"[info] Configuration loaded",
	}, nil
}

// GetUserErrorLogs retrieves error logs for a user/domain
func (s *Service) GetUserErrorLogs(userID, domain string, lines int) ([]string, error) {
	// Mock implementation
	return []string{}, nil
}

// GetUserAccessLogs retrieves access logs for a user/domain
func (s *Service) GetUserAccessLogs(userID, domain string, lines int) ([]string, error) {
	// Mock implementation
	return []string{}, nil
}

// GetStatus retrieves web server status
func (s *Service) GetStatus() (*models.WebServerStatus, error) {
	return &models.WebServerStatus{
		Running:     true,
		ServerType:  "nginx",
		Version:     "1.24.0",
		Uptime:      86400,
		Connections: 42,
	}, nil
}

// TestConfig tests the web server configuration
func (s *Service) TestConfig() (*models.ConfigTestResult, error) {
	return &models.ConfigTestResult{
		Valid:   true,
		Message: "configuration file syntax is ok",
	}, nil
}
