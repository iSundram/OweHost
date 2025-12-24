package v1

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/iSundram/OweHost/internal/api/middleware"
	"github.com/iSundram/OweHost/internal/dns"
	"github.com/iSundram/OweHost/internal/domain"
	"github.com/iSundram/OweHost/pkg/models"
	"github.com/iSundram/OweHost/pkg/utils"
)

// DNSHandler exposes DNS zone and record endpoints.
type DNSHandler struct {
	dnsService    *dns.Service
	domainService *domain.Service
}

// NewDNSHandler creates a DNS handler.
func NewDNSHandler(dnsSvc *dns.Service, domainSvc *domain.Service) *DNSHandler {
	return &DNSHandler{
		dnsService:    dnsSvc,
		domainService: domainSvc,
	}
}

// CreateZone creates a DNS zone for a domain.
func (h *DNSHandler) CreateZone(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	var req struct {
		DomainID string `json:"domain_id"`
		Name     string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Invalid request body")
		return
	}

	if req.DomainID == "" || req.Name == "" {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeValidation, "domain_id and name are required")
		return
	}

	// ownership check
	userID := middleware.GetUserID(r.Context())
	dom, err := h.domainService.Get(req.DomainID)
	if err != nil || dom.UserID != userID {
		utils.WriteError(w, http.StatusForbidden, utils.ErrCodeForbidden, "Access denied")
		return
	}

	zone, err := h.dnsService.CreateZone(req.DomainID, req.Name)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, err.Error())
		return
	}

	utils.WriteCreated(w, zone)
}

// ListZones lists zones for the current user (or all for admin).
func (h *DNSHandler) ListZones(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		utils.WriteError(w, http.StatusUnauthorized, utils.ErrCodeUnauthorized, "Not authenticated")
		return
	}

	// Admin: return all zones; others get zones for their domains
	allZones := h.dnsService.ListAllZones()
	filtered := make([]*models.DNSZone, 0)
	for _, z := range allZones {
		dom, err := h.domainService.Get(z.DomainID)
		if err != nil {
			continue
		}
		if dom.UserID == userID || dom.UserID == "" {
			filtered = append(filtered, z)
		}
	}
	utils.WriteSuccess(w, filtered)
}

// GetZone gets zone by ID.
func (h *DNSHandler) GetZone(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Zone ID required")
		return
	}
	zoneID := parts[len(parts)-1]

	zone, err := h.dnsService.GetZone(zoneID)
	if err != nil {
		utils.WriteError(w, http.StatusNotFound, utils.ErrCodeNotFound, err.Error())
		return
	}

	utils.WriteSuccess(w, zone)
}

// DeleteZone deletes zone.
func (h *DNSHandler) DeleteZone(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Zone ID required")
		return
	}
	zoneID := parts[len(parts)-1]

	if err := h.dnsService.DeleteZone(zoneID); err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// CreateRecord creates record in a zone.
func (h *DNSHandler) CreateRecord(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	parts := strings.Split(strings.TrimSuffix(r.URL.Path, "/"), "/")
	if len(parts) < 6 {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Zone ID required")
		return
	}
	zoneID := parts[len(parts)-2]

	var req models.DNSRecordCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Invalid request body")
		return
	}

	record, err := h.dnsService.CreateRecord(zoneID, &req)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, err.Error())
		return
	}

	utils.WriteCreated(w, record)
}

// ListRecords lists records for a zone.
func (h *DNSHandler) ListRecords(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	parts := strings.Split(strings.TrimSuffix(r.URL.Path, "/"), "/")
	if len(parts) < 6 {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Zone ID required")
		return
	}
	zoneID := parts[len(parts)-2]

	records := h.dnsService.ListRecords(zoneID)
	utils.WriteSuccess(w, records)
}

// UpdateRecord updates a record.
func (h *DNSHandler) UpdateRecord(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut && r.Method != http.MethodPatch {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	parts := strings.Split(strings.TrimSuffix(r.URL.Path, "/"), "/")
	if len(parts) < 5 {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Record ID required")
		return
	}
	recordID := parts[len(parts)-1]

	var req models.DNSRecordCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Invalid request body")
		return
	}

	record, err := h.dnsService.UpdateRecord(recordID, &req)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, err.Error())
		return
	}

	utils.WriteSuccess(w, record)
}

// DeleteRecord deletes a record.
func (h *DNSHandler) DeleteRecord(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	parts := strings.Split(strings.TrimSuffix(r.URL.Path, "/"), "/")
	if len(parts) < 5 {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Record ID required")
		return
	}
	recordID := parts[len(parts)-1]

	if err := h.dnsService.DeleteRecord(recordID); err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// EnableDNSSEC enables DNSSEC for a zone.
func (h *DNSHandler) EnableDNSSEC(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	parts := strings.Split(strings.TrimSuffix(r.URL.Path, "/"), "/")
	if len(parts) < 6 {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Zone ID required")
		return
	}
	zoneID := parts[len(parts)-2]

	key, err := h.dnsService.EnableDNSSEC(zoneID)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, err.Error())
		return
	}

	utils.WriteSuccess(w, key)
}

// SyncZone triggers a sync to provider (mocked).
func (h *DNSHandler) SyncZone(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	parts := strings.Split(strings.TrimSuffix(r.URL.Path, "/"), "/")
	if len(parts) < 6 {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Zone ID required")
		return
	}
	zoneID := parts[len(parts)-2]

	var req struct {
		Provider string `json:"provider"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Invalid request body")
		return
	}
	if req.Provider == "" {
		req.Provider = "default"
	}

	state, err := h.dnsService.SyncWithProvider(zoneID, req.Provider)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, err.Error())
		return
	}

	utils.WriteSuccess(w, state)
}
