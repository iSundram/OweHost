// Package domain provides domain management services for OweHost
package domain

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"sync"
	"time"

	"github.com/iSundram/OweHost/pkg/models"
	"github.com/iSundram/OweHost/pkg/utils"
)

// Service provides domain management functionality
type Service struct {
	domains    map[string]*models.Domain
	subdomains map[string]*models.Subdomain
	redirects  map[string]*models.DomainRedirect
	errorPages map[string]*models.DomainErrorPage
	settings   map[string]*models.DomainSettings
	byName     map[string]*models.Domain
	byUser     map[string][]*models.Domain
	mu         sync.RWMutex
}

// NewService creates a new domain service
func NewService() *Service {
	return &Service{
		domains:    make(map[string]*models.Domain),
		subdomains: make(map[string]*models.Subdomain),
		redirects:  make(map[string]*models.DomainRedirect),
		errorPages: make(map[string]*models.DomainErrorPage),
		settings:   make(map[string]*models.DomainSettings),
		byName:     make(map[string]*models.Domain),
		byUser:     make(map[string][]*models.Domain),
	}
}

// Create creates a new domain
func (s *Service) Create(userID string, req *models.DomainCreateRequest) (*models.Domain, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check for duplicate domain name
	if _, exists := s.byName[req.Name]; exists {
		return nil, errors.New("domain already exists")
	}

	// Generate validation key
	validationKey, err := generateValidationKey()
	if err != nil {
		return nil, err
	}

	documentRoot := req.DocumentRoot
	if documentRoot == "" {
		documentRoot = "/home/" + userID + "/public_html/" + req.Name
	}

	domain := &models.Domain{
		ID:            utils.GenerateID("dom"),
		UserID:        userID,
		Name:          req.Name,
		Type:          req.Type,
		Status:        models.DomainStatusPending,
		DocumentRoot:  documentRoot,
		TargetDomain:  req.TargetDomain,
		Validated:     false,
		ValidationKey: validationKey,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	s.domains[domain.ID] = domain
	s.byName[domain.Name] = domain
	s.byUser[userID] = append(s.byUser[userID], domain)

	return domain, nil
}

// Get gets a domain by ID
func (s *Service) Get(id string) (*models.Domain, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	domain, exists := s.domains[id]
	if !exists {
		return nil, errors.New("domain not found")
	}
	return domain, nil
}

// GetByName gets a domain by name
func (s *Service) GetByName(name string) (*models.Domain, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	domain, exists := s.byName[name]
	if !exists {
		return nil, errors.New("domain not found")
	}
	return domain, nil
}

// ListByUser lists all domains for a user
func (s *Service) ListByUser(userID string) []*models.Domain {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.byUser[userID]
}

// ListAll lists all domains (for admin)
func (s *Service) ListAll() []*models.Domain {
	s.mu.RLock()
	defer s.mu.RUnlock()

	domains := make([]*models.Domain, 0, len(s.domains))
	for _, domain := range s.domains {
		domains = append(domains, domain)
	}
	return domains
}

// Validate validates domain ownership
func (s *Service) Validate(id, validationKey string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	domain, exists := s.domains[id]
	if !exists {
		return errors.New("domain not found")
	}

	if domain.ValidationKey != validationKey {
		return errors.New("invalid validation key")
	}

	domain.Validated = true
	domain.Status = models.DomainStatusActive
	domain.UpdatedAt = time.Now()
	return nil
}

// Delete deletes a domain
func (s *Service) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	domain, exists := s.domains[id]
	if !exists {
		return errors.New("domain not found")
	}

	// Remove from byName index
	delete(s.byName, domain.Name)

	// Remove from byUser index
	userDomains := s.byUser[domain.UserID]
	for i, d := range userDomains {
		if d.ID == id {
			s.byUser[domain.UserID] = append(userDomains[:i], userDomains[i+1:]...)
			break
		}
	}

	// Remove subdomains
	for subID, sub := range s.subdomains {
		if sub.DomainID == id {
			delete(s.subdomains, subID)
		}
	}

	delete(s.domains, id)
	return nil
}

// CreateSubdomain creates a subdomain
func (s *Service) CreateSubdomain(domainID string, req *models.SubdomainCreateRequest) (*models.Subdomain, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	domain, exists := s.domains[domainID]
	if !exists {
		return nil, errors.New("domain not found")
	}

	fullName := req.Name + "." + domain.Name

	// Check for duplicate
	for _, sub := range s.subdomains {
		if sub.FullName == fullName {
			return nil, errors.New("subdomain already exists")
		}
	}

	documentRoot := req.DocumentRoot
	if documentRoot == "" {
		documentRoot = domain.DocumentRoot + "/" + req.Name
	}

	subdomain := &models.Subdomain{
		ID:           utils.GenerateID("sub"),
		DomainID:     domainID,
		Name:         req.Name,
		FullName:     fullName,
		DocumentRoot: documentRoot,
		PathMapping:  req.PathMapping,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	s.subdomains[subdomain.ID] = subdomain
	return subdomain, nil
}

// GetSubdomain gets a subdomain by ID
func (s *Service) GetSubdomain(id string) (*models.Subdomain, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	subdomain, exists := s.subdomains[id]
	if !exists {
		return nil, errors.New("subdomain not found")
	}
	return subdomain, nil
}

// ListSubdomains lists all subdomains for a domain
func (s *Service) ListSubdomains(domainID string) []*models.Subdomain {
	s.mu.RLock()
	defer s.mu.RUnlock()

	subdomains := make([]*models.Subdomain, 0)
	for _, sub := range s.subdomains {
		if sub.DomainID == domainID {
			subdomains = append(subdomains, sub)
		}
	}
	return subdomains
}

// DeleteSubdomain deletes a subdomain
func (s *Service) DeleteSubdomain(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.subdomains[id]; !exists {
		return errors.New("subdomain not found")
	}

	delete(s.subdomains, id)
	return nil
}

// CheckOwnership verifies domain ownership enforcement
func (s *Service) CheckOwnership(userID, domainID string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	domain, exists := s.domains[domainID]
	if !exists {
		return false
	}
	return domain.UserID == userID
}

func generateValidationKey() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// CreateRedirect creates a URL redirect for a domain
func (s *Service) CreateRedirect(domainID string, req *models.DomainRedirectCreateRequest) (*models.DomainRedirect, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.domains[domainID]; !exists {
		return nil, errors.New("domain not found")
	}

	redirect := &models.DomainRedirect{
		ID:           utils.GenerateID("rdr"),
		DomainID:     domainID,
		SourcePath:   req.SourcePath,
		TargetURL:    req.TargetURL,
		Type:         req.Type,
		PreservePath: req.PreservePath,
		Enabled:      true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	s.redirects[redirect.ID] = redirect
	return redirect, nil
}

// GetRedirect gets a redirect by ID
func (s *Service) GetRedirect(id string) (*models.DomainRedirect, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	redirect, exists := s.redirects[id]
	if !exists {
		return nil, errors.New("redirect not found")
	}
	return redirect, nil
}

// ListRedirects lists all redirects for a domain
func (s *Service) ListRedirects(domainID string) []*models.DomainRedirect {
	s.mu.RLock()
	defer s.mu.RUnlock()

	redirects := make([]*models.DomainRedirect, 0)
	for _, r := range s.redirects {
		if r.DomainID == domainID {
			redirects = append(redirects, r)
		}
	}
	return redirects
}

// UpdateRedirect updates a redirect
func (s *Service) UpdateRedirect(id string, req *models.DomainRedirectCreateRequest) (*models.DomainRedirect, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	redirect, exists := s.redirects[id]
	if !exists {
		return nil, errors.New("redirect not found")
	}

	redirect.SourcePath = req.SourcePath
	redirect.TargetURL = req.TargetURL
	redirect.Type = req.Type
	redirect.PreservePath = req.PreservePath
	redirect.UpdatedAt = time.Now()

	return redirect, nil
}

// DeleteRedirect deletes a redirect
func (s *Service) DeleteRedirect(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.redirects[id]; !exists {
		return errors.New("redirect not found")
	}

	delete(s.redirects, id)
	return nil
}

// ToggleRedirect enables or disables a redirect
func (s *Service) ToggleRedirect(id string, enabled bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	redirect, exists := s.redirects[id]
	if !exists {
		return errors.New("redirect not found")
	}

	redirect.Enabled = enabled
	redirect.UpdatedAt = time.Now()
	return nil
}

// CreateErrorPage creates a custom error page for a domain
func (s *Service) CreateErrorPage(domainID string, req *models.DomainErrorPageCreateRequest) (*models.DomainErrorPage, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.domains[domainID]; !exists {
		return nil, errors.New("domain not found")
	}

	// Check if error page for this code already exists
	for _, ep := range s.errorPages {
		if ep.DomainID == domainID && ep.ErrorCode == req.ErrorCode {
			return nil, errors.New("error page already exists for this code")
		}
	}

	errorPage := &models.DomainErrorPage{
		ID:        utils.GenerateID("err"),
		DomainID:  domainID,
		ErrorCode: req.ErrorCode,
		PagePath:  req.PagePath,
		Content:   req.Content,
		Enabled:   true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	s.errorPages[errorPage.ID] = errorPage
	return errorPage, nil
}

// GetErrorPage gets an error page by ID
func (s *Service) GetErrorPage(id string) (*models.DomainErrorPage, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	errorPage, exists := s.errorPages[id]
	if !exists {
		return nil, errors.New("error page not found")
	}
	return errorPage, nil
}

// ListErrorPages lists all error pages for a domain
func (s *Service) ListErrorPages(domainID string) []*models.DomainErrorPage {
	s.mu.RLock()
	defer s.mu.RUnlock()

	errorPages := make([]*models.DomainErrorPage, 0)
	for _, ep := range s.errorPages {
		if ep.DomainID == domainID {
			errorPages = append(errorPages, ep)
		}
	}
	return errorPages
}

// UpdateErrorPage updates an error page
func (s *Service) UpdateErrorPage(id string, req *models.DomainErrorPageCreateRequest) (*models.DomainErrorPage, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	errorPage, exists := s.errorPages[id]
	if !exists {
		return nil, errors.New("error page not found")
	}

	errorPage.PagePath = req.PagePath
	errorPage.Content = req.Content
	errorPage.UpdatedAt = time.Now()

	return errorPage, nil
}

// DeleteErrorPage deletes an error page
func (s *Service) DeleteErrorPage(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.errorPages[id]; !exists {
		return errors.New("error page not found")
	}

	delete(s.errorPages, id)
	return nil
}

// ToggleErrorPage enables or disables an error page
func (s *Service) ToggleErrorPage(id string, enabled bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	errorPage, exists := s.errorPages[id]
	if !exists {
		return errors.New("error page not found")
	}

	errorPage.Enabled = enabled
	errorPage.UpdatedAt = time.Now()
	return nil
}

// GetSettings gets domain settings
func (s *Service) GetSettings(domainID string) (*models.DomainSettings, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if _, exists := s.domains[domainID]; !exists {
		return nil, errors.New("domain not found")
	}

	settings, exists := s.settings[domainID]
	if !exists {
		// Return default settings
		return &models.DomainSettings{
			DomainID:       domainID,
			ForceHTTPS:     false,
			WWWRedirect:    "none",
			HSTSEnabled:    false,
			HSTSMaxAge:     31536000,
			IndexFiles:     "index.html,index.php",
			DirectoryIndex: false,
		}, nil
	}
	return settings, nil
}

// UpdateSettings updates domain settings
func (s *Service) UpdateSettings(domainID string, req *models.DomainSettingsUpdateRequest) (*models.DomainSettings, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.domains[domainID]; !exists {
		return nil, errors.New("domain not found")
	}

	settings, exists := s.settings[domainID]
	if !exists {
		settings = &models.DomainSettings{
			DomainID:       domainID,
			ForceHTTPS:     false,
			WWWRedirect:    "none",
			HSTSEnabled:    false,
			HSTSMaxAge:     31536000,
			IndexFiles:     "index.html,index.php",
			DirectoryIndex: false,
		}
	}

	if req.ForceHTTPS != nil {
		settings.ForceHTTPS = *req.ForceHTTPS
	}
	if req.WWWRedirect != nil {
		settings.WWWRedirect = *req.WWWRedirect
	}
	if req.HSTSEnabled != nil {
		settings.HSTSEnabled = *req.HSTSEnabled
	}
	if req.HSTSMaxAge != nil {
		settings.HSTSMaxAge = *req.HSTSMaxAge
	}
	if req.IndexFiles != nil {
		settings.IndexFiles = *req.IndexFiles
	}
	if req.DirectoryIndex != nil {
		settings.DirectoryIndex = *req.DirectoryIndex
	}

	s.settings[domainID] = settings
	return settings, nil
}

// TransferDomain transfers a domain to another user
func (s *Service) TransferDomain(domainID, newUserID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	domain, exists := s.domains[domainID]
	if !exists {
		return errors.New("domain not found")
	}

	oldUserID := domain.UserID

	// Remove from old user's list
	userDomains := s.byUser[oldUserID]
	for i, d := range userDomains {
		if d.ID == domainID {
			s.byUser[oldUserID] = append(userDomains[:i], userDomains[i+1:]...)
			break
		}
	}

	// Add to new user's list
	domain.UserID = newUserID
	domain.UpdatedAt = time.Now()
	s.byUser[newUserID] = append(s.byUser[newUserID], domain)

	return nil
}
