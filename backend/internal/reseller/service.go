// Package reseller provides reseller management services for OweHost
package reseller

import (
	"errors"
	"sync"
	"time"

	"github.com/iSundram/OweHost/pkg/models"
	"github.com/iSundram/OweHost/pkg/utils"
)

// Service provides reseller management functionality
type Service struct {
	resellers   map[string]*models.Reseller
	byUserID    map[string]*models.Reseller
	mu          sync.RWMutex
}

// NewService creates a new reseller service
func NewService() *Service {
	return &Service{
		resellers: make(map[string]*models.Reseller),
		byUserID:  make(map[string]*models.Reseller),
	}
}

// Create creates a new reseller
func (s *Service) Create(req *models.ResellerCreateRequest) (*models.Reseller, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if user is already a reseller
	if _, exists := s.byUserID[req.UserID]; exists {
		return nil, errors.New("user is already a reseller")
	}

	reseller := &models.Reseller{
		ID:               utils.GenerateID("res"),
		UserID:           req.UserID,
		ParentResellerID: req.ParentResellerID,
		Name:             req.Name,
		ResourcePool:     req.ResourcePool,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	s.resellers[reseller.ID] = reseller
	s.byUserID[reseller.UserID] = reseller

	return reseller, nil
}

// Get gets a reseller by ID
func (s *Service) Get(id string) (*models.Reseller, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	reseller, exists := s.resellers[id]
	if !exists {
		return nil, errors.New("reseller not found")
	}
	return reseller, nil
}

// GetByUserID gets a reseller by user ID
func (s *Service) GetByUserID(userID string) (*models.Reseller, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	reseller, exists := s.byUserID[userID]
	if !exists {
		return nil, errors.New("reseller not found")
	}
	return reseller, nil
}

// List lists all resellers
func (s *Service) List() []*models.Reseller {
	s.mu.RLock()
	defer s.mu.RUnlock()

	resellers := make([]*models.Reseller, 0, len(s.resellers))
	for _, r := range s.resellers {
		resellers = append(resellers, r)
	}
	return resellers
}

// Update updates a reseller
func (s *Service) Update(id string, req *models.ResellerUpdateRequest) (*models.Reseller, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	reseller, exists := s.resellers[id]
	if !exists {
		return nil, errors.New("reseller not found")
	}

	if req.Name != nil {
		reseller.Name = *req.Name
	}

	if req.ResourcePool != nil {
		reseller.ResourcePool = *req.ResourcePool
	}

	reseller.UpdatedAt = time.Now()
	return reseller, nil
}

// Delete deletes a reseller
func (s *Service) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	reseller, exists := s.resellers[id]
	if !exists {
		return errors.New("reseller not found")
	}

	// Check for child resellers
	for _, r := range s.resellers {
		if r.ParentResellerID != nil && *r.ParentResellerID == id {
			return errors.New("cannot delete reseller with child resellers")
		}
	}

	delete(s.resellers, id)
	delete(s.byUserID, reseller.UserID)
	return nil
}

// GetChildren gets child resellers
func (s *Service) GetChildren(id string) []*models.Reseller {
	s.mu.RLock()
	defer s.mu.RUnlock()

	children := make([]*models.Reseller, 0)
	for _, r := range s.resellers {
		if r.ParentResellerID != nil && *r.ParentResellerID == id {
			children = append(children, r)
		}
	}
	return children
}

// GetOwnershipTree builds the ownership tree for a reseller
func (s *Service) GetOwnershipTree(id string) (*models.OwnershipNode, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	reseller, exists := s.resellers[id]
	if !exists {
		return nil, errors.New("reseller not found")
	}

	node := &models.OwnershipNode{
		ID:       reseller.ID,
		Type:     "reseller",
		ParentID: reseller.ParentResellerID,
		Children: make([]models.OwnershipNode, 0),
	}

	// Recursively build children
	s.buildOwnershipTree(node)

	return node, nil
}

// buildOwnershipTree recursively builds the ownership tree
func (s *Service) buildOwnershipTree(node *models.OwnershipNode) {
	for _, r := range s.resellers {
		if r.ParentResellerID != nil && *r.ParentResellerID == node.ID {
			child := models.OwnershipNode{
				ID:       r.ID,
				Type:     "reseller",
				ParentID: &node.ID,
				Children: make([]models.OwnershipNode, 0),
			}
			s.buildOwnershipTree(&child)
			node.Children = append(node.Children, child)
		}
	}
}

// SuspendCascade suspends a reseller and all children
func (s *Service) SuspendCascade(id string) error {
	// This would cascade to all child resellers and their users
	// Implementation would depend on the user service
	return nil
}

// TerminateCascade terminates a reseller and all children
func (s *Service) TerminateCascade(id string) error {
	// This would cascade to all child resellers and their users
	// Implementation would depend on the user service
	return nil
}

// AggregateUsage aggregates resource usage for a reseller and children
func (s *Service) AggregateUsage(id string) (*models.ResourcePool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	reseller, exists := s.resellers[id]
	if !exists {
		return nil, errors.New("reseller not found")
	}

	// Start with own pool
	total := reseller.ResourcePool

	// Add children
	for _, r := range s.resellers {
		if r.ParentResellerID != nil && *r.ParentResellerID == id {
			childUsage, _ := s.AggregateUsage(r.ID)
			if childUsage != nil {
				total.MaxUsers += childUsage.MaxUsers
				total.MaxDomains += childUsage.MaxDomains
				total.MaxDiskMB += childUsage.MaxDiskMB
				total.MaxBandwidthMB += childUsage.MaxBandwidthMB
				total.MaxDatabases += childUsage.MaxDatabases
			}
		}
	}

	return &total, nil
}
