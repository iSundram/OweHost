// Package plugin provides plugin management for OweHost
package plugin

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"sync"
	"time"

	"github.com/iSundram/OweHost/pkg/models"
	"github.com/iSundram/OweHost/pkg/utils"
)

// Service provides plugin functionality
type Service struct {
	plugins    map[string]*models.Plugin
	hooks      map[string][]*models.PluginHook
	scopes     map[string]*models.PluginAPIScope
	mu         sync.RWMutex
}

// NewService creates a new plugin service
func NewService() *Service {
	svc := &Service{
		plugins: make(map[string]*models.Plugin),
		hooks:   make(map[string][]*models.PluginHook),
		scopes:  make(map[string]*models.PluginAPIScope),
	}
	svc.initDefaultScopes()
	return svc
}

// initDefaultScopes initializes default API scopes
func (s *Service) initDefaultScopes() {
	scopes := []*models.PluginAPIScope{
		{Name: "users:read", Description: "Read user information", Permissions: []string{"users:read"}},
		{Name: "users:write", Description: "Modify users", Permissions: []string{"users:write", "users:read"}},
		{Name: "domains:read", Description: "Read domain information", Permissions: []string{"domains:read"}},
		{Name: "domains:write", Description: "Modify domains", Permissions: []string{"domains:write", "domains:read"}},
		{Name: "databases:read", Description: "Read database information", Permissions: []string{"databases:read"}},
		{Name: "databases:write", Description: "Modify databases", Permissions: []string{"databases:write", "databases:read"}},
		{Name: "files:read", Description: "Read files", Permissions: []string{"files:read"}},
		{Name: "files:write", Description: "Modify files", Permissions: []string{"files:write", "files:read"}},
		{Name: "system:read", Description: "Read system information", Permissions: []string{"system:read"}},
		{Name: "system:admin", Description: "System administration", Permissions: []string{"system:*"}},
	}

	for _, scope := range scopes {
		s.scopes[scope.Name] = scope
	}
}

// Install installs a plugin
func (s *Service) Install(req *models.PluginInstallRequest) (*models.Plugin, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Verify signature (simplified)
	if !s.verifySignature(req.PackageURL, req.Signature) {
		return nil, errors.New("invalid plugin signature")
	}

	plugin := &models.Plugin{
		ID:          utils.GenerateID("plg"),
		Name:        "Plugin", // Would be parsed from manifest
		Slug:        "plugin-" + req.Signature[:8],
		Version:     "1.0.0",
		Author:      "Unknown",
		Description: "Plugin description",
		Status:      models.PluginStatusInactive,
		Signature:   req.Signature,
		Verified:    true,
		APIScopes:   []string{},
		ConfigSchema: make(map[string]interface{}),
		Config:      make(map[string]interface{}),
		InstalledAt: time.Now(),
		UpdatedAt:   time.Now(),
	}

	s.plugins[plugin.ID] = plugin
	s.hooks[plugin.ID] = make([]*models.PluginHook, 0)

	return plugin, nil
}

// Get gets a plugin by ID
func (s *Service) Get(id string) (*models.Plugin, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	plugin, exists := s.plugins[id]
	if !exists {
		return nil, errors.New("plugin not found")
	}
	return plugin, nil
}

// List lists all plugins
func (s *Service) List() []*models.Plugin {
	s.mu.RLock()
	defer s.mu.RUnlock()

	plugins := make([]*models.Plugin, 0, len(s.plugins))
	for _, p := range s.plugins {
		plugins = append(plugins, p)
	}
	return plugins
}

// Activate activates a plugin
func (s *Service) Activate(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	plugin, exists := s.plugins[id]
	if !exists {
		return errors.New("plugin not found")
	}

	if !plugin.Verified {
		return errors.New("plugin signature not verified")
	}

	plugin.Status = models.PluginStatusActive
	plugin.UpdatedAt = time.Now()
	return nil
}

// Deactivate deactivates a plugin
func (s *Service) Deactivate(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	plugin, exists := s.plugins[id]
	if !exists {
		return errors.New("plugin not found")
	}

	plugin.Status = models.PluginStatusInactive
	plugin.UpdatedAt = time.Now()
	return nil
}

// Uninstall uninstalls a plugin
func (s *Service) Uninstall(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	plugin, exists := s.plugins[id]
	if !exists {
		return errors.New("plugin not found")
	}

	if plugin.Status == models.PluginStatusActive {
		return errors.New("deactivate plugin before uninstalling")
	}

	delete(s.plugins, id)
	delete(s.hooks, id)
	return nil
}

// Configure configures a plugin
func (s *Service) Configure(id string, req *models.PluginConfigRequest) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	plugin, exists := s.plugins[id]
	if !exists {
		return errors.New("plugin not found")
	}

	plugin.Config = req.Config
	plugin.UpdatedAt = time.Now()
	return nil
}

// RegisterHook registers a plugin hook
func (s *Service) RegisterHook(pluginID, event, handler string, priority int) (*models.PluginHook, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	plugin, exists := s.plugins[pluginID]
	if !exists {
		return nil, errors.New("plugin not found")
	}

	if plugin.Status != models.PluginStatusActive {
		return nil, errors.New("plugin not active")
	}

	hook := &models.PluginHook{
		ID:        utils.GenerateID("hook"),
		PluginID:  pluginID,
		Event:     event,
		Handler:   handler,
		Priority:  priority,
		CreatedAt: time.Now(),
	}

	s.hooks[pluginID] = append(s.hooks[pluginID], hook)
	return hook, nil
}

// GetHooks gets hooks for a plugin
func (s *Service) GetHooks(pluginID string) []*models.PluginHook {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.hooks[pluginID]
}

// GetHooksForEvent gets all hooks for an event
func (s *Service) GetHooksForEvent(event string) []*models.PluginHook {
	s.mu.RLock()
	defer s.mu.RUnlock()

	hooks := make([]*models.PluginHook, 0)
	for pluginID, pluginHooks := range s.hooks {
		plugin := s.plugins[pluginID]
		if plugin == nil || plugin.Status != models.PluginStatusActive {
			continue
		}

		for _, hook := range pluginHooks {
			if hook.Event == event {
				hooks = append(hooks, hook)
			}
		}
	}

	// Sort by priority
	for i := 0; i < len(hooks); i++ {
		for j := i + 1; j < len(hooks); j++ {
			if hooks[i].Priority > hooks[j].Priority {
				hooks[i], hooks[j] = hooks[j], hooks[i]
			}
		}
	}

	return hooks
}

// UnregisterHook unregisters a plugin hook
func (s *Service) UnregisterHook(hookID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for pluginID, pluginHooks := range s.hooks {
		for i, hook := range pluginHooks {
			if hook.ID == hookID {
				s.hooks[pluginID] = append(pluginHooks[:i], pluginHooks[i+1:]...)
				return nil
			}
		}
	}

	return errors.New("hook not found")
}

// ExecuteHooks executes hooks for an event
func (s *Service) ExecuteHooks(event string, data map[string]interface{}) []error {
	hooks := s.GetHooksForEvent(event)
	errs := make([]error, 0)

	for _, hook := range hooks {
		err := s.executeHook(hook, data)
		if err != nil {
			errs = append(errs, err)
		}
	}

	return errs
}

// executeHook executes a single hook
func (s *Service) executeHook(hook *models.PluginHook, data map[string]interface{}) error {
	// In production, would execute the hook handler in a sandbox
	_ = hook
	_ = data
	return nil
}

// GetScopes lists available API scopes
func (s *Service) GetScopes() []*models.PluginAPIScope {
	s.mu.RLock()
	defer s.mu.RUnlock()

	scopes := make([]*models.PluginAPIScope, 0, len(s.scopes))
	for _, scope := range s.scopes {
		scopes = append(scopes, scope)
	}
	return scopes
}

// CheckScope checks if a plugin has access to a scope
func (s *Service) CheckScope(pluginID, scope string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	plugin, exists := s.plugins[pluginID]
	if !exists || plugin.Status != models.PluginStatusActive {
		return false
	}

	for _, s := range plugin.APIScopes {
		if s == scope || s == "*" {
			return true
		}
	}
	return false
}

// GrantScope grants an API scope to a plugin
func (s *Service) GrantScope(pluginID, scope string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	plugin, exists := s.plugins[pluginID]
	if !exists {
		return errors.New("plugin not found")
	}

	if _, exists := s.scopes[scope]; !exists {
		return errors.New("scope not found")
	}

	// Check if already granted
	for _, s := range plugin.APIScopes {
		if s == scope {
			return nil
		}
	}

	plugin.APIScopes = append(plugin.APIScopes, scope)
	plugin.UpdatedAt = time.Now()
	return nil
}

// RevokeScope revokes an API scope from a plugin
func (s *Service) RevokeScope(pluginID, scope string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	plugin, exists := s.plugins[pluginID]
	if !exists {
		return errors.New("plugin not found")
	}

	for i, sc := range plugin.APIScopes {
		if sc == scope {
			plugin.APIScopes = append(plugin.APIScopes[:i], plugin.APIScopes[i+1:]...)
			plugin.UpdatedAt = time.Now()
			return nil
		}
	}

	return errors.New("scope not granted")
}

// verifySignature verifies plugin signature
func (s *Service) verifySignature(packageURL, signature string) bool {
	// Simplified verification - in production, would use proper crypto
	hash := sha256.Sum256([]byte(packageURL))
	expected := hex.EncodeToString(hash[:])
	return len(signature) > 0 && expected != ""
}
