package v1

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/iSundram/OweHost/internal/api/middleware"
	"github.com/iSundram/OweHost/internal/domain"
	"github.com/iSundram/OweHost/internal/user"
	"github.com/iSundram/OweHost/pkg/models"
	"github.com/iSundram/OweHost/pkg/utils"
)

// DomainHandler handles domain endpoints
type DomainHandler struct {
	domainService *domain.Service
	userService   *user.Service
}

// NewDomainHandler creates a new domain handler
func NewDomainHandler(domainSvc *domain.Service, userSvc *user.Service) *DomainHandler {
	return &DomainHandler{
		domainService: domainSvc,
		userService:   userSvc,
	}
}

// Create handles domain creation
func (h *DomainHandler) Create(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	userID := middleware.GetUserID(r.Context())

	var req models.DomainCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Invalid request body")
		return
	}

	if !utils.IsValidDomain(req.Name) {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeValidation, "Invalid domain name")
		return
	}

	domain, err := h.domainService.Create(userID, &req)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, err.Error())
		return
	}

	utils.WriteCreated(w, domain)
}

// Get handles getting a domain
func (h *DomainHandler) Get(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Domain ID required")
		return
	}
	domainID := parts[len(parts)-1]

	domain, err := h.domainService.Get(domainID)
	if err != nil {
		utils.WriteError(w, http.StatusNotFound, utils.ErrCodeNotFound, err.Error())
		return
	}

	// Check ownership
	userID := middleware.GetUserID(r.Context())
	if !h.domainService.CheckOwnership(userID, domainID) {
		utils.WriteError(w, http.StatusForbidden, utils.ErrCodeForbidden, "Access denied")
		return
	}

	utils.WriteSuccess(w, domain)
}

// List handles listing domains
func (h *DomainHandler) List(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	userID := middleware.GetUserID(r.Context())
	currentUser, err := h.userService.Get(userID)
	if err != nil {
		utils.WriteError(w, http.StatusUnauthorized, utils.ErrCodeUnauthorized, "User not found")
		return
	}

	var domains []*models.Domain

	// Admin sees all domains
	if currentUser.Role == models.UserRoleAdmin {
		domains = h.domainService.ListAll()
	} else {
		// Regular users and resellers see their own domains
		domains = h.domainService.ListByUser(userID)
	}

	utils.WriteSuccess(w, domains)
}

// Delete handles deleting a domain
func (h *DomainHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Domain ID required")
		return
	}
	domainID := parts[len(parts)-1]

	userID := middleware.GetUserID(r.Context())
	if !h.domainService.CheckOwnership(userID, domainID) {
		utils.WriteError(w, http.StatusForbidden, utils.ErrCodeForbidden, "Access denied")
		return
	}

	if err := h.domainService.Delete(domainID); err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Validate handles domain validation
func (h *DomainHandler) Validate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Domain ID required")
		return
	}
	domainID := parts[len(parts)-2]

	var req struct {
		ValidationKey string `json:"validation_key"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Invalid request body")
		return
	}

	if err := h.domainService.Validate(domainID, req.ValidationKey); err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, err.Error())
		return
	}

	utils.WriteSuccess(w, map[string]string{"message": "Domain validated"})
}

// CreateSubdomain handles subdomain creation
func (h *DomainHandler) CreateSubdomain(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Domain ID required")
		return
	}
	domainID := parts[len(parts)-2]

	userID := middleware.GetUserID(r.Context())
	if !h.domainService.CheckOwnership(userID, domainID) {
		utils.WriteError(w, http.StatusForbidden, utils.ErrCodeForbidden, "Access denied")
		return
	}

	var req models.SubdomainCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Invalid request body")
		return
	}

	subdomain, err := h.domainService.CreateSubdomain(domainID, &req)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, err.Error())
		return
	}

	utils.WriteCreated(w, subdomain)
}

// ListSubdomains handles listing subdomains
func (h *DomainHandler) ListSubdomains(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Domain ID required")
		return
	}
	domainID := parts[len(parts)-2]

	userID := middleware.GetUserID(r.Context())
	if !h.domainService.CheckOwnership(userID, domainID) {
		utils.WriteError(w, http.StatusForbidden, utils.ErrCodeForbidden, "Access denied")
		return
	}

	subdomains := h.domainService.ListSubdomains(domainID)
	utils.WriteSuccess(w, subdomains)
}

// DeleteSubdomain handles subdomain deletion
func (h *DomainHandler) DeleteSubdomain(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Subdomain ID required")
		return
	}
	subdomainID := parts[len(parts)-1]

	if err := h.domainService.DeleteSubdomain(subdomainID); err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// CreateRedirect handles redirect creation
func (h *DomainHandler) CreateRedirect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Domain ID required")
		return
	}
	domainID := parts[len(parts)-2]

	userID := middleware.GetUserID(r.Context())
	if !h.domainService.CheckOwnership(userID, domainID) {
		utils.WriteError(w, http.StatusForbidden, utils.ErrCodeForbidden, "Access denied")
		return
	}

	var req models.DomainRedirectCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Invalid request body")
		return
	}

	redirect, err := h.domainService.CreateRedirect(domainID, &req)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, err.Error())
		return
	}

	utils.WriteCreated(w, redirect)
}

// ListRedirects handles listing redirects for a domain
func (h *DomainHandler) ListRedirects(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Domain ID required")
		return
	}
	domainID := parts[len(parts)-2]

	userID := middleware.GetUserID(r.Context())
	if !h.domainService.CheckOwnership(userID, domainID) {
		utils.WriteError(w, http.StatusForbidden, utils.ErrCodeForbidden, "Access denied")
		return
	}

	redirects := h.domainService.ListRedirects(domainID)
	utils.WriteSuccess(w, redirects)
}

// UpdateRedirect handles redirect update
func (h *DomainHandler) UpdateRedirect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut && r.Method != http.MethodPatch {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Redirect ID required")
		return
	}
	redirectID := parts[len(parts)-1]

	var req models.DomainRedirectCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Invalid request body")
		return
	}

	redirect, err := h.domainService.UpdateRedirect(redirectID, &req)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, err.Error())
		return
	}

	utils.WriteSuccess(w, redirect)
}

// DeleteRedirect handles redirect deletion
func (h *DomainHandler) DeleteRedirect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Redirect ID required")
		return
	}
	redirectID := parts[len(parts)-1]

	if err := h.domainService.DeleteRedirect(redirectID); err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ToggleRedirect handles enabling/disabling a redirect
func (h *DomainHandler) ToggleRedirect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Redirect ID required")
		return
	}
	redirectID := parts[len(parts)-2]

	var req struct {
		Enabled bool `json:"enabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Invalid request body")
		return
	}

	if err := h.domainService.ToggleRedirect(redirectID, req.Enabled); err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, err.Error())
		return
	}

	utils.WriteSuccess(w, map[string]string{"message": "Redirect toggled"})
}

// CreateErrorPage handles error page creation
func (h *DomainHandler) CreateErrorPage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Domain ID required")
		return
	}
	domainID := parts[len(parts)-2]

	userID := middleware.GetUserID(r.Context())
	if !h.domainService.CheckOwnership(userID, domainID) {
		utils.WriteError(w, http.StatusForbidden, utils.ErrCodeForbidden, "Access denied")
		return
	}

	var req models.DomainErrorPageCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Invalid request body")
		return
	}

	errorPage, err := h.domainService.CreateErrorPage(domainID, &req)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, err.Error())
		return
	}

	utils.WriteCreated(w, errorPage)
}

// ListErrorPages handles listing error pages for a domain
func (h *DomainHandler) ListErrorPages(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Domain ID required")
		return
	}
	domainID := parts[len(parts)-2]

	userID := middleware.GetUserID(r.Context())
	if !h.domainService.CheckOwnership(userID, domainID) {
		utils.WriteError(w, http.StatusForbidden, utils.ErrCodeForbidden, "Access denied")
		return
	}

	errorPages := h.domainService.ListErrorPages(domainID)
	utils.WriteSuccess(w, errorPages)
}

// UpdateErrorPage handles error page update
func (h *DomainHandler) UpdateErrorPage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut && r.Method != http.MethodPatch {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Error page ID required")
		return
	}
	errorPageID := parts[len(parts)-1]

	var req models.DomainErrorPageCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Invalid request body")
		return
	}

	errorPage, err := h.domainService.UpdateErrorPage(errorPageID, &req)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, err.Error())
		return
	}

	utils.WriteSuccess(w, errorPage)
}

// DeleteErrorPage handles error page deletion
func (h *DomainHandler) DeleteErrorPage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Error page ID required")
		return
	}
	errorPageID := parts[len(parts)-1]

	if err := h.domainService.DeleteErrorPage(errorPageID); err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetSettings handles getting domain settings
func (h *DomainHandler) GetSettings(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Domain ID required")
		return
	}
	domainID := parts[len(parts)-2]

	userID := middleware.GetUserID(r.Context())
	if !h.domainService.CheckOwnership(userID, domainID) {
		utils.WriteError(w, http.StatusForbidden, utils.ErrCodeForbidden, "Access denied")
		return
	}

	settings, err := h.domainService.GetSettings(domainID)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, err.Error())
		return
	}

	utils.WriteSuccess(w, settings)
}

// UpdateSettings handles updating domain settings
func (h *DomainHandler) UpdateSettings(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut && r.Method != http.MethodPatch {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Domain ID required")
		return
	}
	domainID := parts[len(parts)-2]

	userID := middleware.GetUserID(r.Context())
	if !h.domainService.CheckOwnership(userID, domainID) {
		utils.WriteError(w, http.StatusForbidden, utils.ErrCodeForbidden, "Access denied")
		return
	}

	var req models.DomainSettingsUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Invalid request body")
		return
	}

	settings, err := h.domainService.UpdateSettings(domainID, &req)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, err.Error())
		return
	}

	utils.WriteSuccess(w, settings)
}

// TransferDomain handles domain transfer to another user
func (h *DomainHandler) TransferDomain(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Domain ID required")
		return
	}
	domainID := parts[len(parts)-2]

	var req struct {
		NewUserID string `json:"new_user_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Invalid request body")
		return
	}

	if err := h.domainService.TransferDomain(domainID, req.NewUserID); err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, err.Error())
		return
	}

	utils.WriteSuccess(w, map[string]string{"message": "Domain transferred"})
}
