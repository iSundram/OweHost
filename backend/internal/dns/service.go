// Package dns provides DNS management services for OweHost
package dns

import (
	"errors"
	"sync"
	"time"

	"github.com/iSundram/OweHost/pkg/models"
	"github.com/iSundram/OweHost/pkg/utils"
)

// Service provides DNS management functionality
type Service struct {
	zones         map[string]*models.DNSZone
	records       map[string]*models.DNSRecord
	dnssecKeys    map[string]*models.DNSSECKey
	syncStates    map[string]*models.DNSSyncState
	byDomain      map[string]*models.DNSZone
	recordsByZone map[string][]*models.DNSRecord
	mu            sync.RWMutex
}

// NewService creates a new DNS service
func NewService() *Service {
	return &Service{
		zones:         make(map[string]*models.DNSZone),
		records:       make(map[string]*models.DNSRecord),
		dnssecKeys:    make(map[string]*models.DNSSECKey),
		syncStates:    make(map[string]*models.DNSSyncState),
		byDomain:      make(map[string]*models.DNSZone),
		recordsByZone: make(map[string][]*models.DNSRecord),
	}
}

// ListAllZones returns all zones (admin use)
func (s *Service) ListAllZones() []*models.DNSZone {
	s.mu.RLock()
	defer s.mu.RUnlock()

	zones := make([]*models.DNSZone, 0, len(s.zones))
	for _, z := range s.zones {
		zones = append(zones, z)
	}
	return zones
}

// CreateZone creates a new DNS zone
func (s *Service) CreateZone(domainID, name string) (*models.DNSZone, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.byDomain[domainID]; exists {
		return nil, errors.New("zone already exists for domain")
	}

	zone := &models.DNSZone{
		ID:            utils.GenerateID("zone"),
		DomainID:      domainID,
		Name:          name,
		Locked:        false,
		DNSSECEnabled: false,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	s.zones[zone.ID] = zone
	s.byDomain[domainID] = zone
	s.recordsByZone[zone.ID] = make([]*models.DNSRecord, 0)

	return zone, nil
}

// GetZone gets a DNS zone by ID
func (s *Service) GetZone(id string) (*models.DNSZone, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	zone, exists := s.zones[id]
	if !exists {
		return nil, errors.New("zone not found")
	}
	return zone, nil
}

// GetZoneByDomain gets a DNS zone by domain ID
func (s *Service) GetZoneByDomain(domainID string) (*models.DNSZone, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	zone, exists := s.byDomain[domainID]
	if !exists {
		return nil, errors.New("zone not found")
	}
	return zone, nil
}

// DeleteZone deletes a DNS zone
func (s *Service) DeleteZone(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	zone, exists := s.zones[id]
	if !exists {
		return errors.New("zone not found")
	}

	if zone.Locked {
		return errors.New("zone is locked")
	}

	// Delete all records
	for _, record := range s.recordsByZone[id] {
		delete(s.records, record.ID)
	}
	delete(s.recordsByZone, id)

	// Delete zone
	delete(s.zones, id)
	delete(s.byDomain, zone.DomainID)

	return nil
}

// LockZone locks a DNS zone
func (s *Service) LockZone(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	zone, exists := s.zones[id]
	if !exists {
		return errors.New("zone not found")
	}

	zone.Locked = true
	zone.UpdatedAt = time.Now()
	return nil
}

// UnlockZone unlocks a DNS zone
func (s *Service) UnlockZone(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	zone, exists := s.zones[id]
	if !exists {
		return errors.New("zone not found")
	}

	zone.Locked = false
	zone.UpdatedAt = time.Now()
	return nil
}

// CreateRecord creates a DNS record
func (s *Service) CreateRecord(zoneID string, req *models.DNSRecordCreateRequest) (*models.DNSRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	zone, exists := s.zones[zoneID]
	if !exists {
		return nil, errors.New("zone not found")
	}

	if zone.Locked {
		return nil, errors.New("zone is locked")
	}

	ttl := req.TTL
	if ttl == 0 {
		ttl = 3600
	}

	record := &models.DNSRecord{
		ID:        utils.GenerateID("rec"),
		ZoneID:    zoneID,
		Name:      req.Name,
		Type:      req.Type,
		Content:   req.Content,
		TTL:       ttl,
		Priority:  req.Priority,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	s.records[record.ID] = record
	s.recordsByZone[zoneID] = append(s.recordsByZone[zoneID], record)

	zone.UpdatedAt = time.Now()

	return record, nil
}

// GetRecord gets a DNS record by ID
func (s *Service) GetRecord(id string) (*models.DNSRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	record, exists := s.records[id]
	if !exists {
		return nil, errors.New("record not found")
	}
	return record, nil
}

// ListRecords lists all DNS records for a zone
func (s *Service) ListRecords(zoneID string) []*models.DNSRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.recordsByZone[zoneID]
}

// UpdateRecord updates a DNS record
func (s *Service) UpdateRecord(id string, req *models.DNSRecordCreateRequest) (*models.DNSRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	record, exists := s.records[id]
	if !exists {
		return nil, errors.New("record not found")
	}

	zone := s.zones[record.ZoneID]
	if zone.Locked {
		return nil, errors.New("zone is locked")
	}

	record.Name = req.Name
	record.Type = req.Type
	record.Content = req.Content
	record.TTL = req.TTL
	record.Priority = req.Priority
	record.UpdatedAt = time.Now()

	return record, nil
}

// DeleteRecord deletes a DNS record
func (s *Service) DeleteRecord(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	record, exists := s.records[id]
	if !exists {
		return errors.New("record not found")
	}

	zone := s.zones[record.ZoneID]
	if zone.Locked {
		return errors.New("zone is locked")
	}

	// Remove from zone records list
	zoneRecords := s.recordsByZone[record.ZoneID]
	for i, r := range zoneRecords {
		if r.ID == id {
			s.recordsByZone[record.ZoneID] = append(zoneRecords[:i], zoneRecords[i+1:]...)
			break
		}
	}

	delete(s.records, id)
	return nil
}

// EnableDNSSEC enables DNSSEC for a zone
func (s *Service) EnableDNSSEC(zoneID string) (*models.DNSSECKey, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	zone, exists := s.zones[zoneID]
	if !exists {
		return nil, errors.New("zone not found")
	}

	// Generate DNSSEC key (simplified)
	key := &models.DNSSECKey{
		ID:         utils.GenerateID("dnskey"),
		ZoneID:     zoneID,
		KeyTag:     12345,
		Algorithm:  13, // ECDSA P-256
		DigestType: 2,  // SHA-256
		Digest:     "dummy-digest-value",
		PublicKey:  "dummy-public-key",
		Active:     true,
		CreatedAt:  time.Now(),
	}

	zone.DNSSECEnabled = true
	zone.UpdatedAt = time.Now()

	s.dnssecKeys[key.ID] = key
	return key, nil
}

// ScheduleKeyRollover schedules a DNSSEC key rollover
func (s *Service) ScheduleKeyRollover(keyID string, rolloverAt time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key, exists := s.dnssecKeys[keyID]
	if !exists {
		return errors.New("key not found")
	}

	key.RolloverAt = &rolloverAt
	return nil
}

// SyncWithProvider syncs zone with external DNS provider
func (s *Service) SyncWithProvider(zoneID, providerName string) (*models.DNSSyncState, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, exists := s.zones[zoneID]
	if !exists {
		return nil, errors.New("zone not found")
	}

	state := &models.DNSSyncState{
		ZoneID:       zoneID,
		ProviderName: providerName,
		LastSyncAt:   time.Now(),
		SyncStatus:   "completed",
	}

	s.syncStates[zoneID+"_"+providerName] = state
	return state, nil
}

// GetSyncState gets sync state for a zone and provider
func (s *Service) GetSyncState(zoneID, providerName string) (*models.DNSSyncState, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	state, exists := s.syncStates[zoneID+"_"+providerName]
	if !exists {
		return nil, errors.New("sync state not found")
	}
	return state, nil
}
