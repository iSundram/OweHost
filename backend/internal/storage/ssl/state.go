// Package ssl provides filesystem-based SSL certificate state management
package ssl

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/iSundram/OweHost/internal/storage/account"
)

// CertificateMeta represents SSL certificate metadata
type CertificateMeta struct {
	Domain      string    `json:"domain"`
	Type        string    `json:"type"`       // letsencrypt, custom, self-signed
	Issuer      string    `json:"issuer"`
	Subject     string    `json:"subject"`
	SANs        []string  `json:"sans,omitempty"`
	ValidFrom   time.Time `json:"valid_from"`
	ValidUntil  time.Time `json:"valid_until"`
	AutoRenew   bool      `json:"auto_renew"`
	LastRenewed *time.Time `json:"last_renewed,omitempty"`
	LastCheck   *time.Time `json:"last_check,omitempty"`
	RenewalError string   `json:"renewal_error,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CSRRequest represents a Certificate Signing Request
type CSRRequest struct {
	Domain       string   `json:"domain"`
	Organization string   `json:"organization,omitempty"`
	Country      string   `json:"country,omitempty"`
	State        string   `json:"state,omitempty"`
	City         string   `json:"city,omitempty"`
	Email        string   `json:"email,omitempty"`
	SANs         []string `json:"sans,omitempty"`
}

// StateManager handles SSL certificate state
type StateManager struct {
	mu sync.RWMutex
}

// NewStateManager creates a new SSL state manager
func NewStateManager() *StateManager {
	return &StateManager{}
}

// SSLPath returns the SSL directory path for an account
func (s *StateManager) SSLPath(accountID int) string {
	return filepath.Join(
		account.BaseAccountPath,
		fmt.Sprintf("%s%d", account.AccountPrefix, accountID),
		"ssl",
	)
}

// DomainSSLPath returns the SSL directory for a specific domain
func (s *StateManager) DomainSSLPath(accountID int, domain string) string {
	return filepath.Join(s.SSLPath(accountID), domain)
}

// ReadMeta reads certificate metadata
func (s *StateManager) ReadMeta(accountID int, domain string) (*CertificateMeta, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	path := filepath.Join(s.DomainSSLPath(accountID, domain), "meta.json")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var meta CertificateMeta
	if err := json.Unmarshal(data, &meta); err != nil {
		return nil, err
	}

	return &meta, nil
}

// WriteMeta writes certificate metadata
func (s *StateManager) WriteMeta(accountID int, domain string, meta *CertificateMeta) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	meta.UpdatedAt = time.Now()

	path := s.DomainSSLPath(accountID, domain)
	if err := os.MkdirAll(path, 0700); err != nil {
		return err
	}

	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(path, "meta.json"), data, 0644)
}

// InstallCertificate installs certificate files
func (s *StateManager) InstallCertificate(accountID int, domain string, cert, key, chain []byte, meta *CertificateMeta) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	sslPath := s.DomainSSLPath(accountID, domain)
	if err := os.MkdirAll(sslPath, 0700); err != nil {
		return err
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

	// Write fullchain (cert + chain)
	fullchain := append(cert, chain...)
	if err := os.WriteFile(filepath.Join(sslPath, "fullchain.pem"), fullchain, 0644); err != nil {
		return fmt.Errorf("failed to write fullchain: %w", err)
	}

	// Write metadata
	if meta == nil {
		meta = &CertificateMeta{
			Domain:    domain,
			Type:      "custom",
			CreatedAt: time.Now(),
		}
	}
	meta.UpdatedAt = time.Now()

	metaData, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(sslPath, "meta.json"), metaData, 0644)
}

// DeleteCertificate removes a certificate
func (s *StateManager) DeleteCertificate(accountID int, domain string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return os.RemoveAll(s.DomainSSLPath(accountID, domain))
}

// HasCertificate checks if a valid certificate exists
func (s *StateManager) HasCertificate(accountID int, domain string) bool {
	sslPath := s.DomainSSLPath(accountID, domain)
	certPath := filepath.Join(sslPath, "cert.pem")
	keyPath := filepath.Join(sslPath, "key.pem")

	certInfo, certErr := os.Stat(certPath)
	keyInfo, keyErr := os.Stat(keyPath)

	return certErr == nil && keyErr == nil && !certInfo.IsDir() && !keyInfo.IsDir()
}

// ListCertificates lists all certificates for an account
func (s *StateManager) ListCertificates(accountID int) ([]CertificateMeta, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	sslPath := s.SSLPath(accountID)
	entries, err := os.ReadDir(sslPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []CertificateMeta{}, nil
		}
		return nil, err
	}

	var certs []CertificateMeta
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		metaPath := filepath.Join(sslPath, entry.Name(), "meta.json")
		data, err := os.ReadFile(metaPath)
		if err != nil {
			continue
		}

		var meta CertificateMeta
		if json.Unmarshal(data, &meta) == nil {
			certs = append(certs, meta)
		}
	}

	return certs, nil
}

// GetExpiringCertificates returns certificates expiring within days
func (s *StateManager) GetExpiringCertificates(accountID int, days int) ([]CertificateMeta, error) {
	certs, err := s.ListCertificates(accountID)
	if err != nil {
		return nil, err
	}

	cutoff := time.Now().AddDate(0, 0, days)
	var expiring []CertificateMeta
	for _, cert := range certs {
		if cert.ValidUntil.Before(cutoff) && cert.AutoRenew {
			expiring = append(expiring, cert)
		}
	}

	return expiring, nil
}

// GenerateSelfSigned generates a self-signed certificate
func (s *StateManager) GenerateSelfSigned(accountID int, domain string, validDays int) error {
	// Generate private key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return fmt.Errorf("failed to generate private key: %w", err)
	}

	// Create certificate template
	notBefore := time.Now()
	notAfter := notBefore.AddDate(0, 0, validDays)

	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return fmt.Errorf("failed to generate serial number: %w", err)
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName:   domain,
			Organization: []string{"OweHost Self-Signed"},
		},
		DNSNames:    []string{domain, "www." + domain},
		NotBefore:   notBefore,
		NotAfter:    notAfter,
		KeyUsage:    x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}

	// Create certificate
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return fmt.Errorf("failed to create certificate: %w", err)
	}

	// Encode to PEM
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)})

	// Install certificate
	meta := &CertificateMeta{
		Domain:     domain,
		Type:       "self-signed",
		Issuer:     "OweHost Self-Signed",
		Subject:    domain,
		SANs:       []string{domain, "www." + domain},
		ValidFrom:  notBefore,
		ValidUntil: notAfter,
		AutoRenew:  false,
		CreatedAt:  time.Now(),
	}

	return s.InstallCertificate(accountID, domain, certPEM, keyPEM, nil, meta)
}

// GenerateCSR generates a Certificate Signing Request
func (s *StateManager) GenerateCSR(accountID int, req CSRRequest) (csr, key []byte, err error) {
	// Generate private key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate private key: %w", err)
	}

	// Create CSR template
	subject := pkix.Name{
		CommonName: req.Domain,
	}
	if req.Organization != "" {
		subject.Organization = []string{req.Organization}
	}
	if req.Country != "" {
		subject.Country = []string{req.Country}
	}
	if req.State != "" {
		subject.Province = []string{req.State}
	}
	if req.City != "" {
		subject.Locality = []string{req.City}
	}

	template := x509.CertificateRequest{
		Subject:  subject,
		DNSNames: append([]string{req.Domain}, req.SANs...),
	}

	// Generate CSR
	csrDER, err := x509.CreateCertificateRequest(rand.Reader, &template, privateKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create CSR: %w", err)
	}

	// Encode to PEM
	csrPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE REQUEST", Bytes: csrDER})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)})

	// Store private key for later use
	sslPath := s.DomainSSLPath(accountID, req.Domain)
	os.MkdirAll(sslPath, 0700)
	os.WriteFile(filepath.Join(sslPath, "key.pem"), keyPEM, 0600)
	os.WriteFile(filepath.Join(sslPath, "csr.pem"), csrPEM, 0644)

	return csrPEM, keyPEM, nil
}

// GetCertificatePaths returns paths to certificate files
func (s *StateManager) GetCertificatePaths(accountID int, domain string) (cert, key, chain, fullchain string) {
	sslPath := s.DomainSSLPath(accountID, domain)
	return filepath.Join(sslPath, "cert.pem"),
		filepath.Join(sslPath, "key.pem"),
		filepath.Join(sslPath, "chain.pem"),
		filepath.Join(sslPath, "fullchain.pem")
}
