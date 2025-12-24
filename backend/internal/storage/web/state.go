// Package web provides filesystem-based web/site state management
package web

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/iSundram/OweHost/internal/storage/account"
)

// StateManager handles filesystem state for web sites
type StateManager struct {
	mu sync.RWMutex
}

// NewStateManager creates a new web state manager
func NewStateManager() *StateManager {
	return &StateManager{}
}

// SitePath returns the path for a site directory
func (s *StateManager) SitePath(accountID int, domain string) string {
	return filepath.Join(
		account.BaseAccountPath,
		fmt.Sprintf("%s%d", account.AccountPrefix, accountID),
		"web",
		domain,
	)
}

// SSLPath returns the path for SSL certificates
func (s *StateManager) SSLPath(accountID int, domain string) string {
	return filepath.Join(
		account.BaseAccountPath,
		fmt.Sprintf("%s%d", account.AccountPrefix, accountID),
		"ssl",
		domain,
	)
}

// Exists checks if a site exists
func (s *StateManager) Exists(accountID int, domain string) bool {
	path := s.SitePath(accountID, domain)
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

// ReadSite reads site.json for a domain
func (s *StateManager) ReadSite(accountID int, domain string) (*SiteDescriptor, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	path := filepath.Join(s.SitePath(accountID, domain), "site.json")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("site %s not found for account %d", domain, accountID)
		}
		return nil, fmt.Errorf("failed to read site.json: %w", err)
	}

	var site SiteDescriptor
	if err := json.Unmarshal(data, &site); err != nil {
		return nil, fmt.Errorf("failed to parse site.json: %w", err)
	}

	return &site, nil
}

// WriteSite writes site.json and creates the site structure
func (s *StateManager) WriteSite(accountID int, site *SiteDescriptor) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	basePath := s.SitePath(accountID, site.Domain)

	// Determine document root path
	docRoot := site.DocumentRoot
	if docRoot == "" {
		docRoot = "public"
	}

	// Create directory structure
	dirs := []string{
		basePath,
		filepath.Join(basePath, docRoot),
		filepath.Join(basePath, "logs"),
		filepath.Join(basePath, "tmp"),
		filepath.Join(basePath, "cache"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Update timestamp
	site.UpdatedAt = time.Now().Format(time.RFC3339)
	if site.CreatedAt == "" {
		site.CreatedAt = site.UpdatedAt
	}

	// Write site.json atomically
	return s.atomicWrite(filepath.Join(basePath, "site.json"), site)
}

// DeleteSite removes a site directory
func (s *StateManager) DeleteSite(accountID int, domain string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	path := s.SitePath(accountID, domain)
	return os.RemoveAll(path)
}

// ListSites lists all sites for an account
func (s *StateManager) ListSites(accountID int) ([]SiteDescriptor, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	webPath := filepath.Join(
		account.BaseAccountPath,
		fmt.Sprintf("%s%d", account.AccountPrefix, accountID),
		"web",
	)

	entries, err := os.ReadDir(webPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []SiteDescriptor{}, nil
		}
		return nil, fmt.Errorf("failed to read web directory: %w", err)
	}

	var sites []SiteDescriptor
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		site, err := s.ReadSite(accountID, entry.Name())
		if err == nil {
			sites = append(sites, *site)
		}
	}

	return sites, nil
}

// ReadSSLMeta reads SSL metadata for a domain
func (s *StateManager) ReadSSLMeta(accountID int, domain string) (*SSLMeta, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	path := filepath.Join(s.SSLPath(accountID, domain), "meta.json")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // No SSL configured
		}
		return nil, fmt.Errorf("failed to read SSL meta: %w", err)
	}

	var meta SSLMeta
	if err := json.Unmarshal(data, &meta); err != nil {
		return nil, fmt.Errorf("failed to parse SSL meta: %w", err)
	}

	return &meta, nil
}

// WriteSSLCertificate writes SSL certificate files
func (s *StateManager) WriteSSLCertificate(accountID int, domain string, cert, key, chain []byte, meta *SSLMeta) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	sslPath := s.SSLPath(accountID, domain)

	// Create directory
	if err := os.MkdirAll(sslPath, 0700); err != nil {
		return fmt.Errorf("failed to create SSL directory: %w", err)
	}

	// Write certificate
	if err := os.WriteFile(filepath.Join(sslPath, "cert.pem"), cert, 0644); err != nil {
		return fmt.Errorf("failed to write certificate: %w", err)
	}

	// Write private key (restricted permissions)
	if err := os.WriteFile(filepath.Join(sslPath, "key.pem"), key, 0600); err != nil {
		return fmt.Errorf("failed to write private key: %w", err)
	}

	// Write chain if provided
	if len(chain) > 0 {
		if err := os.WriteFile(filepath.Join(sslPath, "chain.pem"), chain, 0644); err != nil {
			return fmt.Errorf("failed to write chain: %w", err)
		}
	}

	// Write metadata
	if meta != nil {
		if err := s.atomicWrite(filepath.Join(sslPath, "meta.json"), meta); err != nil {
			return fmt.Errorf("failed to write SSL meta: %w", err)
		}
	}

	return nil
}

// DeleteSSLCertificate removes SSL certificate files
func (s *StateManager) DeleteSSLCertificate(accountID int, domain string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return os.RemoveAll(s.SSLPath(accountID, domain))
}

// GetSSLCertificatePaths returns paths to SSL certificate files
func (s *StateManager) GetSSLCertificatePaths(accountID int, domain string) (cert, key, chain string) {
	sslPath := s.SSLPath(accountID, domain)
	return filepath.Join(sslPath, "cert.pem"),
		filepath.Join(sslPath, "key.pem"),
		filepath.Join(sslPath, "chain.pem")
}

// CreateDefaultIndex creates a default index.html for a new site
func (s *StateManager) CreateDefaultIndex(accountID int, domain, docRoot string) error {
	sitePath := s.SitePath(accountID, domain)
	indexPath := filepath.Join(sitePath, docRoot, "index.html")

	// Check if file already exists
	if _, err := os.Stat(indexPath); err == nil {
		return nil // Already exists
	}

	defaultHTML := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Welcome to %s</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            display: flex;
            justify-content: center;
            align-items: center;
            min-height: 100vh;
            margin: 0;
            background: linear-gradient(135deg, #0F0E47 0%%, #272757 100%%);
            color: #fff;
        }
        .container {
            text-align: center;
            padding: 40px;
        }
        h1 {
            font-size: 2.5rem;
            margin-bottom: 1rem;
        }
        p {
            font-size: 1.2rem;
            color: #8686AC;
        }
        .logo {
            font-size: 4rem;
            margin-bottom: 1rem;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="logo">🚀</div>
        <h1>Welcome to %s</h1>
        <p>Your website is ready. Upload your files to get started.</p>
        <p style="margin-top: 2rem; font-size: 0.9rem;">Powered by OweHost</p>
    </div>
</body>
</html>`, domain, domain)

	return os.WriteFile(indexPath, []byte(defaultHTML), 0644)
}

// atomicWrite writes data atomically using temp file + rename
func (s *StateManager) atomicWrite(path string, data interface{}) error {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write temp file: %w", err)
	}

	if err := os.Rename(tmpPath, path); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("failed to rename temp file: %w", err)
	}

	return nil
}
