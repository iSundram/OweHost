// Package runtime provides runtime and language management for OweHost
package runtime

import (
	"errors"
	"sync"
	"time"

	"github.com/iSundram/OweHost/pkg/models"
	"github.com/iSundram/OweHost/pkg/utils"
)

// Service provides runtime management functionality
type Service struct {
	phpPools     map[string]*models.PHPPool
	nodejsApps   map[string]*models.NodeJSApp
	pythonApps   map[string]*models.PythonApp
	versions     map[models.RuntimeType][]models.RuntimeVersion
	extensions   map[string][]models.PHPExtension
	mu           sync.RWMutex
}

// NewService creates a new runtime service
func NewService() *Service {
	svc := &Service{
		phpPools:   make(map[string]*models.PHPPool),
		nodejsApps: make(map[string]*models.NodeJSApp),
		pythonApps: make(map[string]*models.PythonApp),
		versions:   make(map[models.RuntimeType][]models.RuntimeVersion),
		extensions: make(map[string][]models.PHPExtension),
	}
	svc.initDefaultVersions()
	return svc
}

// initDefaultVersions initializes available runtime versions
func (s *Service) initDefaultVersions() {
	s.versions[models.RuntimeTypePHP] = []models.RuntimeVersion{
		{Type: models.RuntimeTypePHP, Version: "7.4", Path: "/usr/bin/php7.4", Default: false, Available: true},
		{Type: models.RuntimeTypePHP, Version: "8.0", Path: "/usr/bin/php8.0", Default: false, Available: true},
		{Type: models.RuntimeTypePHP, Version: "8.1", Path: "/usr/bin/php8.1", Default: false, Available: true},
		{Type: models.RuntimeTypePHP, Version: "8.2", Path: "/usr/bin/php8.2", Default: true, Available: true},
		{Type: models.RuntimeTypePHP, Version: "8.3", Path: "/usr/bin/php8.3", Default: false, Available: true},
	}

	s.versions[models.RuntimeTypeNodeJS] = []models.RuntimeVersion{
		{Type: models.RuntimeTypeNodeJS, Version: "16", Path: "/usr/bin/node16", Default: false, Available: true},
		{Type: models.RuntimeTypeNodeJS, Version: "18", Path: "/usr/bin/node18", Default: false, Available: true},
		{Type: models.RuntimeTypeNodeJS, Version: "20", Path: "/usr/bin/node20", Default: true, Available: true},
	}

	s.versions[models.RuntimeTypePython] = []models.RuntimeVersion{
		{Type: models.RuntimeTypePython, Version: "3.9", Path: "/usr/bin/python3.9", Default: false, Available: true},
		{Type: models.RuntimeTypePython, Version: "3.10", Path: "/usr/bin/python3.10", Default: false, Available: true},
		{Type: models.RuntimeTypePython, Version: "3.11", Path: "/usr/bin/python3.11", Default: true, Available: true},
		{Type: models.RuntimeTypePython, Version: "3.12", Path: "/usr/bin/python3.12", Default: false, Available: true},
	}

	s.versions[models.RuntimeTypeGo] = []models.RuntimeVersion{
		{Type: models.RuntimeTypeGo, Version: "1.21", Path: "/usr/local/go/bin/go", Default: true, Available: true},
	}

	s.versions[models.RuntimeTypeJava] = []models.RuntimeVersion{
		{Type: models.RuntimeTypeJava, Version: "11", Path: "/usr/lib/jvm/java-11/bin/java", Default: false, Available: true},
		{Type: models.RuntimeTypeJava, Version: "17", Path: "/usr/lib/jvm/java-17/bin/java", Default: true, Available: true},
		{Type: models.RuntimeTypeJava, Version: "21", Path: "/usr/lib/jvm/java-21/bin/java", Default: false, Available: true},
	}
}

// ListVersions lists available versions for a runtime type
func (s *Service) ListVersions(runtimeType models.RuntimeType) []models.RuntimeVersion {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.versions[runtimeType]
}

// CreatePHPPool creates a PHP-FPM pool for a user
func (s *Service) CreatePHPPool(userID string, req *models.PHPPoolCreateRequest) (*models.PHPPool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if user already has a pool
	for _, pool := range s.phpPools {
		if pool.UserID == userID && pool.Version == req.Version {
			return nil, errors.New("pool already exists for user with this version")
		}
	}

	pool := &models.PHPPool{
		ID:              utils.GenerateID("php"),
		UserID:          userID,
		Version:         req.Version,
		PoolName:        "user_" + userID,
		SocketPath:      "/var/run/php/php" + req.Version + "-fpm-" + userID + ".sock",
		MaxChildren:     req.MaxChildren,
		StartServers:    2,
		MinSpareServers: 1,
		MaxSpareServers: 3,
		Extensions:      req.Extensions,
		INIOverrides:    req.INIOverrides,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	if pool.MaxChildren == 0 {
		pool.MaxChildren = 5
	}

	s.phpPools[pool.ID] = pool
	return pool, nil
}

// GetPHPPool gets a PHP pool by ID
func (s *Service) GetPHPPool(id string) (*models.PHPPool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	pool, exists := s.phpPools[id]
	if !exists {
		return nil, errors.New("pool not found")
	}
	return pool, nil
}

// ListPHPPoolsByUser lists PHP pools for a user
func (s *Service) ListPHPPoolsByUser(userID string) []*models.PHPPool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	pools := make([]*models.PHPPool, 0)
	for _, pool := range s.phpPools {
		if pool.UserID == userID {
			pools = append(pools, pool)
		}
	}
	return pools
}

// UpdatePHPPool updates a PHP pool
func (s *Service) UpdatePHPPool(id string, req *models.PHPPoolCreateRequest) (*models.PHPPool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	pool, exists := s.phpPools[id]
	if !exists {
		return nil, errors.New("pool not found")
	}

	if req.MaxChildren > 0 {
		pool.MaxChildren = req.MaxChildren
	}
	if req.Extensions != nil {
		pool.Extensions = req.Extensions
	}
	if req.INIOverrides != nil {
		pool.INIOverrides = req.INIOverrides
	}

	pool.UpdatedAt = time.Now()
	return pool, nil
}

// DeletePHPPool deletes a PHP pool
func (s *Service) DeletePHPPool(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.phpPools[id]; !exists {
		return errors.New("pool not found")
	}

	delete(s.phpPools, id)
	return nil
}

// EnablePHPExtension enables a PHP extension for a pool
func (s *Service) EnablePHPExtension(poolID, extension string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	pool, exists := s.phpPools[poolID]
	if !exists {
		return errors.New("pool not found")
	}

	// Check if already enabled
	for _, ext := range pool.Extensions {
		if ext == extension {
			return nil
		}
	}

	pool.Extensions = append(pool.Extensions, extension)
	pool.UpdatedAt = time.Now()
	return nil
}

// DisablePHPExtension disables a PHP extension for a pool
func (s *Service) DisablePHPExtension(poolID, extension string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	pool, exists := s.phpPools[poolID]
	if !exists {
		return errors.New("pool not found")
	}

	for i, ext := range pool.Extensions {
		if ext == extension {
			pool.Extensions = append(pool.Extensions[:i], pool.Extensions[i+1:]...)
			pool.UpdatedAt = time.Now()
			return nil
		}
	}

	return errors.New("extension not enabled")
}

// CreateNodeJSApp creates a Node.js application
func (s *Service) CreateNodeJSApp(userID string, req *models.NodeJSAppCreateRequest) (*models.NodeJSApp, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Find available port
	port := s.findAvailablePort(3000)

	app := &models.NodeJSApp{
		ID:          utils.GenerateID("node"),
		UserID:      userID,
		DomainID:    req.DomainID,
		Name:        req.Name,
		Version:     req.Version,
		AppRoot:     req.AppRoot,
		StartFile:   req.StartFile,
		Port:        port,
		Environment: req.Environment,
		Running:     false,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	s.nodejsApps[app.ID] = app
	return app, nil
}

// GetNodeJSApp gets a Node.js app by ID
func (s *Service) GetNodeJSApp(id string) (*models.NodeJSApp, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	app, exists := s.nodejsApps[id]
	if !exists {
		return nil, errors.New("app not found")
	}
	return app, nil
}

// StartNodeJSApp starts a Node.js application
func (s *Service) StartNodeJSApp(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	app, exists := s.nodejsApps[id]
	if !exists {
		return errors.New("app not found")
	}

	app.Running = true
	pid := 12345 // Simulated PID
	app.PID = &pid
	app.UpdatedAt = time.Now()
	return nil
}

// StopNodeJSApp stops a Node.js application
func (s *Service) StopNodeJSApp(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	app, exists := s.nodejsApps[id]
	if !exists {
		return errors.New("app not found")
	}

	app.Running = false
	app.PID = nil
	app.UpdatedAt = time.Now()
	return nil
}

// DeleteNodeJSApp deletes a Node.js application
func (s *Service) DeleteNodeJSApp(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	app, exists := s.nodejsApps[id]
	if !exists {
		return errors.New("app not found")
	}

	if app.Running {
		return errors.New("stop app before deleting")
	}

	delete(s.nodejsApps, id)
	return nil
}

// CreatePythonApp creates a Python application
func (s *Service) CreatePythonApp(userID string, req *models.PythonAppCreateRequest) (*models.PythonApp, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	port := s.findAvailablePort(8000)

	app := &models.PythonApp{
		ID:          utils.GenerateID("py"),
		UserID:      userID,
		DomainID:    req.DomainID,
		Name:        req.Name,
		Version:     req.Version,
		VenvPath:    req.AppRoot + "/venv",
		AppRoot:     req.AppRoot,
		WSGIFile:    req.WSGIFile,
		Port:        port,
		Environment: req.Environment,
		Running:     false,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	s.pythonApps[app.ID] = app
	return app, nil
}

// GetPythonApp gets a Python app by ID
func (s *Service) GetPythonApp(id string) (*models.PythonApp, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	app, exists := s.pythonApps[id]
	if !exists {
		return nil, errors.New("app not found")
	}
	return app, nil
}

// ProvisionVirtualenv provisions a Python virtual environment
func (s *Service) ProvisionVirtualenv(appID string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	_, exists := s.pythonApps[appID]
	if !exists {
		return errors.New("app not found")
	}

	// In production, this would run python -m venv
	return nil
}

// DeletePythonApp deletes a Python application
func (s *Service) DeletePythonApp(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	app, exists := s.pythonApps[id]
	if !exists {
		return errors.New("app not found")
	}

	if app.Running {
		return errors.New("stop app before deleting")
	}

	delete(s.pythonApps, id)
	return nil
}

// findAvailablePort finds an available port starting from base
func (s *Service) findAvailablePort(base int) int {
	usedPorts := make(map[int]bool)

	for _, app := range s.nodejsApps {
		usedPorts[app.Port] = true
	}
	for _, app := range s.pythonApps {
		usedPorts[app.Port] = true
	}

	for port := base; port < 65535; port++ {
		if !usedPorts[port] {
			return port
		}
	}
	return base
}
