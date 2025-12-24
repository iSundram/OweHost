// Package ssl provides SSL certificate management for OweHost
package ssl

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"math/big"
	"sync"
	"time"

	"github.com/iSundram/OweHost/pkg/models"
	"github.com/iSundram/OweHost/pkg/utils"
)

// Service provides SSL certificate functionality
type Service struct {
	certificates  map[string]*models.Certificate
	csrs          map[string]*models.CSR
	renewals      map[string]*models.RenewalSchedule
	byDomain      map[string]*models.Certificate
	mu            sync.RWMutex
}

// NewService creates a new SSL service
func NewService() *Service {
	return &Service{
		certificates: make(map[string]*models.Certificate),
		csrs:         make(map[string]*models.CSR),
		renewals:     make(map[string]*models.RenewalSchedule),
		byDomain:     make(map[string]*models.Certificate),
	}
}

// GenerateCSR generates a Certificate Signing Request
func (s *Service) GenerateCSR(req *models.CSRCreateRequest) (*models.CSR, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Generate private key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}

	// Create CSR template
	template := x509.CertificateRequest{
		Subject: pkix.Name{
			CommonName:   req.CommonName,
			Organization: []string{req.Organization},
			Country:      []string{req.Country},
			Province:     []string{req.State},
			Locality:     []string{req.City},
		},
		DNSNames: append([]string{req.CommonName}, req.SANs...),
	}

	// Generate CSR
	csrBytes, err := x509.CreateCertificateRequest(rand.Reader, &template, privateKey)
	if err != nil {
		return nil, err
	}

	csrPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE REQUEST",
		Bytes: csrBytes,
	})

	keyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	csr := &models.CSR{
		ID:           utils.GenerateID("csr"),
		DomainID:     req.DomainID,
		CommonName:   req.CommonName,
		Organization: req.Organization,
		Country:      req.Country,
		State:        req.State,
		City:         req.City,
		CSRData:      string(csrPEM),
		PrivateKey:   string(keyPEM),
		CreatedAt:    time.Now(),
	}

	s.csrs[csr.ID] = csr
	return csr, nil
}

// GetCSR gets a CSR by ID
func (s *Service) GetCSR(id string) (*models.CSR, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	csr, exists := s.csrs[id]
	if !exists {
		return nil, errors.New("CSR not found")
	}
	return csr, nil
}

// UploadCertificate uploads a custom certificate
func (s *Service) UploadCertificate(req *models.CertificateUploadRequest) (*models.Certificate, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Parse certificate to extract info
	block, _ := pem.Decode([]byte(req.Certificate))
	if block == nil {
		return nil, errors.New("invalid certificate format")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, err
	}

	certificate := &models.Certificate{
		ID:           utils.GenerateID("cert"),
		DomainID:     req.DomainID,
		Type:         models.CertificateTypeCustom,
		Status:       models.CertificateStatusActive,
		CommonName:   cert.Subject.CommonName,
		SANs:         cert.DNSNames,
		Issuer:       cert.Issuer.CommonName,
		SerialNumber: cert.SerialNumber.String(),
		CertPath:     "/etc/ssl/certs/" + req.DomainID + ".crt",
		KeyPath:      "/etc/ssl/private/" + req.DomainID + ".key",
		IssuedAt:     cert.NotBefore,
		ExpiresAt:    cert.NotAfter,
		AutoRenew:    false,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if req.Chain != "" {
		chainPath := "/etc/ssl/certs/" + req.DomainID + ".chain.crt"
		certificate.ChainPath = &chainPath
	}

	s.certificates[certificate.ID] = certificate
	s.byDomain[req.DomainID] = certificate

	return certificate, nil
}

// GenerateSelfSigned generates a self-signed certificate
func (s *Service) GenerateSelfSigned(domainID, commonName string) (*models.Certificate, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Generate private key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}

	// Create certificate template
	template := x509.Certificate{
		SerialNumber: big.NewInt(time.Now().UnixNano()),
		Subject: pkix.Name{
			CommonName: commonName,
		},
		DNSNames:              []string{commonName},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(1, 0, 0), // 1 year
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	// Self-sign
	certBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return nil, err
	}

	certificate := &models.Certificate{
		ID:           utils.GenerateID("cert"),
		DomainID:     domainID,
		Type:         models.CertificateTypeSelfSigned,
		Status:       models.CertificateStatusActive,
		CommonName:   commonName,
		SANs:         []string{commonName},
		Issuer:       "Self-Signed",
		SerialNumber: template.SerialNumber.String(),
		CertPath:     "/etc/ssl/certs/" + domainID + ".crt",
		KeyPath:      "/etc/ssl/private/" + domainID + ".key",
		IssuedAt:     template.NotBefore,
		ExpiresAt:    template.NotAfter,
		AutoRenew:    false,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Store cert and key (in production, would write to files)
	_ = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certBytes})
	_ = pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)})

	s.certificates[certificate.ID] = certificate
	s.byDomain[domainID] = certificate

	return certificate, nil
}

// Get gets a certificate by ID
func (s *Service) Get(id string) (*models.Certificate, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	cert, exists := s.certificates[id]
	if !exists {
		return nil, errors.New("certificate not found")
	}
	return cert, nil
}

// GetByDomain gets a certificate by domain ID
func (s *Service) GetByDomain(domainID string) (*models.Certificate, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	cert, exists := s.byDomain[domainID]
	if !exists {
		return nil, errors.New("certificate not found")
	}
	return cert, nil
}

// Delete deletes a certificate
func (s *Service) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	cert, exists := s.certificates[id]
	if !exists {
		return errors.New("certificate not found")
	}

	delete(s.certificates, id)
	delete(s.byDomain, cert.DomainID)
	delete(s.renewals, id)

	return nil
}

// EnableAutoRenew enables auto-renewal for a certificate
func (s *Service) EnableAutoRenew(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	cert, exists := s.certificates[id]
	if !exists {
		return errors.New("certificate not found")
	}

	if cert.Type != models.CertificateTypeLetsEncrypt {
		return errors.New("auto-renewal only supported for Let's Encrypt certificates")
	}

	cert.AutoRenew = true
	cert.UpdatedAt = time.Now()

	// Schedule renewal
	renewAt := cert.ExpiresAt.AddDate(0, 0, -30) // 30 days before expiry
	s.renewals[id] = &models.RenewalSchedule{
		CertificateID:  id,
		ScheduledAt:    renewAt,
		MaxRetries:     3,
		CurrentRetries: 0,
	}

	return nil
}

// DisableAutoRenew disables auto-renewal for a certificate
func (s *Service) DisableAutoRenew(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	cert, exists := s.certificates[id]
	if !exists {
		return errors.New("certificate not found")
	}

	cert.AutoRenew = false
	cert.UpdatedAt = time.Now()
	delete(s.renewals, id)

	return nil
}

// GetRenewalSchedule gets the renewal schedule for a certificate
func (s *Service) GetRenewalSchedule(id string) (*models.RenewalSchedule, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	schedule, exists := s.renewals[id]
	if !exists {
		return nil, errors.New("no renewal scheduled")
	}
	return schedule, nil
}

// ProcessRenewals processes pending certificate renewals
func (s *Service) ProcessRenewals() []string {
	s.mu.Lock()
	defer s.mu.Unlock()

	renewed := make([]string, 0)
	now := time.Now()

	for certID, schedule := range s.renewals {
		if schedule.ScheduledAt.Before(now) {
			cert := s.certificates[certID]
			if cert != nil && cert.AutoRenew {
				// In production, this would request a new cert from Let's Encrypt
				renewed = append(renewed, certID)
				
				now := time.Now()
				schedule.LastAttemptAt = &now
				
				// Update certificate
				cert.LastRenewalAt = &now
				cert.RenewalAttempts++
				cert.UpdatedAt = now
			}
		}
	}

	return renewed
}

// CheckExpiring returns certificates expiring within the specified days
func (s *Service) CheckExpiring(days int) []*models.Certificate {
	s.mu.RLock()
	defer s.mu.RUnlock()

	expiring := make([]*models.Certificate, 0)
	threshold := time.Now().AddDate(0, 0, days)

	for _, cert := range s.certificates {
		if cert.ExpiresAt.Before(threshold) && cert.Status == models.CertificateStatusActive {
			expiring = append(expiring, cert)
		}
	}

	return expiring
}

// UpdateStatus updates certificate status
func (s *Service) UpdateStatus(id string, status models.CertificateStatus) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	cert, exists := s.certificates[id]
	if !exists {
		return errors.New("certificate not found")
	}

	cert.Status = status
	cert.UpdatedAt = time.Now()
	return nil
}

// RequestLetsEncrypt requests a Let's Encrypt certificate
func (s *Service) RequestLetsEncrypt(userID string, req interface{}) (*models.Certificate, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	cert := &models.Certificate{
		ID:           utils.GenerateID("cert"),
		DomainID:     userID, // Using userID temporarily
		Type:         models.CertificateTypeLetsEncrypt,
		Status:       models.CertificateStatusActive,
		CommonName:   "example.com",
		AutoRenew:    true,
		IssuedAt:     time.Now(),
		ExpiresAt:    time.Now().AddDate(0, 3, 0),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	
	s.certificates[cert.ID] = cert
	return cert, nil
}

// DeleteAllByUser deletes all certificates for a user
func (s *Service) DeleteAllByUser(userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	// Note: In production, we'd need to track which domains belong to which user
	// For now, we'll just delete all certificates (mock implementation)
	for id := range s.certificates {
		delete(s.certificates, id)
	}
	return nil
}

// IssueLetsEncrypt initiates a Let's Encrypt certificate issuance
func (s *Service) IssueLetsEncrypt(req *models.LetsEncryptRequest) (*models.ACMEOrder, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Create ACME order
	order := &models.ACMEOrder{
		ID:        utils.GenerateID("order"),
		DomainID:  req.DomainID,
		Domains:   req.Domains,
		Status:    "pending",
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Create challenges for each domain
	challenges := make([]models.ACMEChallenge, 0)
	for _, domain := range req.Domains {
		challenge := models.ACMEChallenge{
			ID:        utils.GenerateID("chal"),
			DomainID:  req.DomainID,
			Domain:    domain,
			Type:      req.ChallengeType,
			Token:     generateRandomToken(),
			KeyAuth:   generateRandomToken() + "." + generateRandomToken(),
			Status:    "pending",
			ExpiresAt: time.Now().Add(24 * time.Hour),
			CreatedAt: time.Now(),
		}
		challenges = append(challenges, challenge)
	}
	order.Challenges = challenges

	return order, nil
}

// ValidateChallenge validates an ACME challenge
func (s *Service) ValidateChallenge(challengeID string) (*models.ACMEChallenge, error) {
	// In production, this would verify the challenge with the ACME server
	// For now, we simulate a successful validation
	now := time.Now()
	challenge := &models.ACMEChallenge{
		ID:          challengeID,
		Status:      "valid",
		ValidatedAt: &now,
	}
	return challenge, nil
}

// FinalizeOrder finalizes an ACME order and issues the certificate
func (s *Service) FinalizeOrder(orderID string) (*models.Certificate, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// In production, this would finalize the order with the ACME server
	// and download the certificate
	cert := &models.Certificate{
		ID:        utils.GenerateID("cert"),
		Type:      models.CertificateTypeLetsEncrypt,
		Status:    models.CertificateStatusActive,
		Issuer:    "Let's Encrypt",
		IssuedAt:  time.Now(),
		ExpiresAt: time.Now().AddDate(0, 3, 0), // 90 days
		AutoRenew: true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	s.certificates[cert.ID] = cert
	return cert, nil
}

// GetSSLSettings gets SSL settings for a domain
func (s *Service) GetSSLSettings(domainID string) *models.SSLSettings {
	// In production, this would be stored in a database
	// Return default settings for now
	return &models.SSLSettings{
		DomainID:        domainID,
		ForceHTTPS:      false,
		HSTSEnabled:     false,
		HSTSMaxAge:      31536000,
		HSTSIncludeSubs: false,
		HSTSPreload:     false,
		MinTLSVersion:   "TLS1.2",
		CipherSuites:    []string{"TLS_AES_128_GCM_SHA256", "TLS_AES_256_GCM_SHA384"},
		OCSPStapling:    true,
	}
}

// UpdateSSLSettings updates SSL settings for a domain
func (s *Service) UpdateSSLSettings(domainID string, req *models.SSLSettingsUpdateRequest) *models.SSLSettings {
	settings := s.GetSSLSettings(domainID)

	if req.ForceHTTPS != nil {
		settings.ForceHTTPS = *req.ForceHTTPS
	}
	if req.HSTSEnabled != nil {
		settings.HSTSEnabled = *req.HSTSEnabled
	}
	if req.HSTSMaxAge != nil {
		settings.HSTSMaxAge = *req.HSTSMaxAge
	}
	if req.HSTSIncludeSubs != nil {
		settings.HSTSIncludeSubs = *req.HSTSIncludeSubs
	}
	if req.HSTSPreload != nil {
		settings.HSTSPreload = *req.HSTSPreload
	}
	if req.MinTLSVersion != nil {
		settings.MinTLSVersion = *req.MinTLSVersion
	}
	if req.CipherSuites != nil {
		settings.CipherSuites = *req.CipherSuites
	}
	if req.OCSPStapling != nil {
		settings.OCSPStapling = *req.OCSPStapling
	}

	return settings
}

// ListAll returns all certificates
func (s *Service) ListAll() []*models.Certificate {
	s.mu.RLock()
	defer s.mu.RUnlock()

	certs := make([]*models.Certificate, 0, len(s.certificates))
	for _, cert := range s.certificates {
		certs = append(certs, cert)
	}
	return certs
}

// generateRandomToken generates a random token for ACME challenges
func generateRandomToken() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// hex encoding helper
var hex = struct {
	EncodeToString func([]byte) string
}{
	EncodeToString: func(b []byte) string {
		const hextable = "0123456789abcdef"
		dst := make([]byte, len(b)*2)
		for i, v := range b {
			dst[i*2] = hextable[v>>4]
			dst[i*2+1] = hextable[v&0x0f]
		}
		return string(dst)
	},
}
