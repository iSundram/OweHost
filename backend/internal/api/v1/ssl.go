package v1

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/iSundram/OweHost/internal/ssl"
	"github.com/iSundram/OweHost/internal/user"
	"github.com/iSundram/OweHost/pkg/models"
)

type SSLHandler struct {
	sslService  *ssl.Service
	userService *user.Service
}

func NewSSLHandler(sslService *ssl.Service, userService *user.Service) *SSLHandler {
	return &SSLHandler{
		sslService:  sslService,
		userService: userService,
	}
}

// ListCertificates lists all SSL certificates
func (h *SSLHandler) ListCertificates(w http.ResponseWriter, r *http.Request) {
	// Check expiring certificates (within 30 days)
	certificates := h.sslService.CheckExpiring(30)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(certificates)
}

// GetCertificate retrieves a specific SSL certificate
func (h *SSLHandler) GetCertificate(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.TrimSuffix(r.URL.Path, "/"), "/")
	if len(parts) < 6 {
		http.Error(w, "Invalid certificate ID", http.StatusBadRequest)
		return
	}
	certID := parts[5]

	certificate, err := h.sslService.Get(certID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(certificate)
}

// CreateCertificate creates a new SSL certificate (self-signed)
func (h *SSLHandler) CreateCertificate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		DomainID   string `json:"domain_id"`
		CommonName string `json:"common_name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	certificate, err := h.sslService.GenerateSelfSigned(req.DomainID, req.CommonName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(certificate)
}

// DeleteCertificate deletes an SSL certificate
func (h *SSLHandler) DeleteCertificate(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.TrimSuffix(r.URL.Path, "/"), "/")
	if len(parts) < 6 {
		http.Error(w, "Invalid certificate ID", http.StatusBadRequest)
		return
	}
	certID := parts[5]

	_, err := h.sslService.Get(certID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	if err := h.sslService.Delete(certID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// RenewCertificate renews an SSL certificate
func (h *SSLHandler) RenewCertificate(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.TrimSuffix(r.URL.Path, "/"), "/")
	if len(parts) < 7 {
		http.Error(w, "Invalid certificate ID", http.StatusBadRequest)
		return
	}
	certID := parts[5]

	// Enable auto-renew
	if err := h.sslService.EnableAutoRenew(certID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	certificate, _ := h.sslService.Get(certID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(certificate)
}

// RequestLetsEncrypt requests a Let's Encrypt certificate
func (h *SSLHandler) RequestLetsEncrypt(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)

	var req struct {
		DomainID string   `json:"domain_id"`
		Domains  []string `json:"domains"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	certificate, err := h.sslService.RequestLetsEncrypt(userID, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(certificate)
}

// InstallCertificate installs an uploaded certificate
func (h *SSLHandler) InstallCertificate(w http.ResponseWriter, r *http.Request) {
	var req models.CertificateUploadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	certificate, err := h.sslService.UploadCertificate(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(certificate)
}

// GenerateCSR generates a Certificate Signing Request
func (h *SSLHandler) GenerateCSR(w http.ResponseWriter, r *http.Request) {
	var req models.CSRCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	csr, err := h.sslService.GenerateCSR(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(csr)
}

// VerifyDomain verifies domain ownership for SSL
func (h *SSLHandler) VerifyDomain(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Domain         string `json:"domain"`
		VerifyMethod   string `json:"verify_method"` // dns, http, email
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Placeholder response - actual verification would be async
	result := map[string]interface{}{
		"domain":      req.Domain,
		"method":      req.VerifyMethod,
		"status":      "pending",
		"instructions": "Add the following DNS TXT record to verify domain ownership",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// IssueLetsEncrypt initiates Let's Encrypt certificate issuance
func (h *SSLHandler) IssueLetsEncrypt(w http.ResponseWriter, r *http.Request) {
	var req models.LetsEncryptRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.ChallengeType == "" {
		req.ChallengeType = "http-01"
	}

	order, err := h.sslService.IssueLetsEncrypt(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(order)
}

// ValidateChallenge validates an ACME challenge
func (h *SSLHandler) ValidateChallenge(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.TrimSuffix(r.URL.Path, "/"), "/")
	if len(parts) < 4 {
		http.Error(w, "Challenge ID required", http.StatusBadRequest)
		return
	}
	challengeID := parts[len(parts)-2]

	challenge, err := h.sslService.ValidateChallenge(challengeID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(challenge)
}

// FinalizeOrder finalizes an ACME order
func (h *SSLHandler) FinalizeOrder(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.TrimSuffix(r.URL.Path, "/"), "/")
	if len(parts) < 4 {
		http.Error(w, "Order ID required", http.StatusBadRequest)
		return
	}
	orderID := parts[len(parts)-2]

	cert, err := h.sslService.FinalizeOrder(orderID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cert)
}

// GetSSLSettings gets SSL settings for a domain
func (h *SSLHandler) GetSSLSettings(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.TrimSuffix(r.URL.Path, "/"), "/")
	if len(parts) < 5 {
		http.Error(w, "Domain ID required", http.StatusBadRequest)
		return
	}
	domainID := parts[len(parts)-2]

	settings := h.sslService.GetSSLSettings(domainID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(settings)
}

// UpdateSSLSettings updates SSL settings for a domain
func (h *SSLHandler) UpdateSSLSettings(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.TrimSuffix(r.URL.Path, "/"), "/")
	if len(parts) < 5 {
		http.Error(w, "Domain ID required", http.StatusBadRequest)
		return
	}
	domainID := parts[len(parts)-2]

	var req models.SSLSettingsUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	settings := h.sslService.UpdateSSLSettings(domainID, &req)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(settings)
}

// ListAllCertificates lists all certificates (admin)
func (h *SSLHandler) ListAllCertificates(w http.ResponseWriter, r *http.Request) {
	certificates := h.sslService.ListAll()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"certificates": certificates,
		"count":        len(certificates),
	})
}
