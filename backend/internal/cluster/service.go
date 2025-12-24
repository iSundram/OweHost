// Package cluster provides node, cluster, and cloud management for OweHost
package cluster

import (
	"errors"
	"sync"
	"time"

	"github.com/iSundram/OweHost/pkg/models"
	"github.com/iSundram/OweHost/pkg/utils"
)

// Service provides cluster functionality
type Service struct {
	nodes      map[string]*models.Node
	providers  map[string]*models.CloudProvider
	instances  map[string]*models.VMInstance
	mu         sync.RWMutex
}

// NewService creates a new cluster service
func NewService() *Service {
	return &Service{
		nodes:     make(map[string]*models.Node),
		providers: make(map[string]*models.CloudProvider),
		instances: make(map[string]*models.VMInstance),
	}
}

// RegisterNode registers a node in the cluster
func (s *Service) RegisterNode(req *models.NodeRegisterRequest) (*models.Node, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	node := &models.Node{
		ID:            utils.GenerateID("node"),
		Name:          req.Name,
		Hostname:      req.Hostname,
		IPAddress:     req.IPAddress,
		Status:        models.NodeStatusOnline,
		Version:       "1.0.0",
		Capabilities:  req.Capabilities,
		Resources:     req.Resources,
		Metadata:      req.Metadata,
		LastHeartbeat: time.Now(),
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	s.nodes[node.ID] = node
	return node, nil
}

// GetNode gets a node by ID
func (s *Service) GetNode(id string) (*models.Node, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	node, exists := s.nodes[id]
	if !exists {
		return nil, errors.New("node not found")
	}
	return node, nil
}

// ListNodes lists all nodes
func (s *Service) ListNodes() []*models.Node {
	s.mu.RLock()
	defer s.mu.RUnlock()

	nodes := make([]*models.Node, 0, len(s.nodes))
	for _, node := range s.nodes {
		nodes = append(nodes, node)
	}
	return nodes
}

// UpdateNodeStatus updates a node's status
func (s *Service) UpdateNodeStatus(id string, status models.NodeStatus) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	node, exists := s.nodes[id]
	if !exists {
		return errors.New("node not found")
	}

	node.Status = status
	node.UpdatedAt = time.Now()
	return nil
}

// ProcessHeartbeat processes a heartbeat from a node
func (s *Service) ProcessHeartbeat(heartbeat *models.NodeHeartbeat) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	node, exists := s.nodes[heartbeat.NodeID]
	if !exists {
		return errors.New("node not found")
	}

	node.LastHeartbeat = heartbeat.Timestamp
	node.Status = heartbeat.Status
	node.Resources = heartbeat.Resources
	node.Version = heartbeat.Version
	node.UpdatedAt = time.Now()

	return nil
}

// RemoveNode removes a node from the cluster
func (s *Service) RemoveNode(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.nodes[id]; !exists {
		return errors.New("node not found")
	}

	delete(s.nodes, id)
	return nil
}

// GetOnlineNodes gets all online nodes
func (s *Service) GetOnlineNodes() []*models.Node {
	s.mu.RLock()
	defer s.mu.RUnlock()

	nodes := make([]*models.Node, 0)
	for _, node := range s.nodes {
		if node.Status == models.NodeStatusOnline {
			nodes = append(nodes, node)
		}
	}
	return nodes
}

// DiscoverCapabilities discovers capabilities of all nodes
func (s *Service) DiscoverCapabilities() map[string][]string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	capabilities := make(map[string][]string)
	for id, node := range s.nodes {
		capabilities[id] = node.Capabilities
	}
	return capabilities
}

// PlaceResource decides which node should host a resource
func (s *Service) PlaceResource(req *models.PlacementRequest) (*models.PlacementResult, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var bestNode *models.Node
	bestScore := 0

	for _, node := range s.nodes {
		if node.Status != models.NodeStatusOnline {
			continue
		}

		score := s.calculatePlacementScore(node, req)
		if score > bestScore {
			bestScore = score
			bestNode = node
		}
	}

	if bestNode == nil {
		return nil, errors.New("no suitable node found")
	}

	return &models.PlacementResult{
		NodeID: bestNode.ID,
		Score:  bestScore,
		Reason: "Best available resources",
	}, nil
}

// calculatePlacementScore calculates a placement score for a node
func (s *Service) calculatePlacementScore(node *models.Node, req *models.PlacementRequest) int {
	score := 100

	// Check resource requirements
	if req.Requirements != nil {
		if memReq, ok := req.Requirements["memory_mb"]; ok {
			availMem := node.Resources.MemoryMB - int64(node.Resources.MemoryUsage*float64(node.Resources.MemoryMB))
			if availMem < memReq {
				return 0
			}
			score += int(availMem / 1024)
		}

		if cpuReq, ok := req.Requirements["cpu_cores"]; ok {
			if int64(node.Resources.CPUCores) < cpuReq {
				return 0
			}
		}
	}

	// Prefer nodes with lower usage
	score -= int(node.Resources.CPUUsage * 10)
	score -= int(node.Resources.MemoryUsage * 10)

	return score
}

// CheckDeadNodes checks for nodes that haven't sent heartbeats
func (s *Service) CheckDeadNodes(timeout time.Duration) []*models.Node {
	s.mu.Lock()
	defer s.mu.Unlock()

	deadNodes := make([]*models.Node, 0)
	threshold := time.Now().Add(-timeout)

	for _, node := range s.nodes {
		if node.Status == models.NodeStatusOnline && node.LastHeartbeat.Before(threshold) {
			node.Status = models.NodeStatusOffline
			node.UpdatedAt = time.Now()
			deadNodes = append(deadNodes, node)
		}
	}

	return deadNodes
}

// RegisterCloudProvider registers a cloud provider
func (s *Service) RegisterCloudProvider(name, providerType, region string, credentials map[string]string, settings map[string]interface{}) (*models.CloudProvider, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	provider := &models.CloudProvider{
		ID:          utils.GenerateID("cp"),
		Name:        name,
		Type:        providerType,
		Credentials: credentials,
		Region:      region,
		Enabled:     true,
		Settings:    settings,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	s.providers[provider.ID] = provider
	return provider, nil
}

// GetCloudProvider gets a cloud provider by ID
func (s *Service) GetCloudProvider(id string) (*models.CloudProvider, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	provider, exists := s.providers[id]
	if !exists {
		return nil, errors.New("provider not found")
	}
	return provider, nil
}

// ListCloudProviders lists all cloud providers
func (s *Service) ListCloudProviders() []*models.CloudProvider {
	s.mu.RLock()
	defer s.mu.RUnlock()

	providers := make([]*models.CloudProvider, 0, len(s.providers))
	for _, p := range s.providers {
		providers = append(providers, p)
	}
	return providers
}

// DeleteCloudProvider deletes a cloud provider
func (s *Service) DeleteCloudProvider(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.providers[id]; !exists {
		return errors.New("provider not found")
	}

	delete(s.providers, id)
	return nil
}

// VMLifecycle performs a VM lifecycle action
func (s *Service) VMLifecycle(providerID, instanceID string, action models.VMLifecycleAction) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, exists := s.providers[providerID]
	if !exists {
		return errors.New("provider not found")
	}

	switch action {
	case models.VMActionCreate:
		instance := &models.VMInstance{
			ID:           utils.GenerateID("vm"),
			ProviderID:   providerID,
			Name:         instanceID,
			ExternalID:   "ext-" + instanceID,
			Status:       "running",
			IPAddress:    "10.0.0.1",
			InstanceType: "standard",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}
		s.instances[instance.ID] = instance

	case models.VMActionStart, models.VMActionStop, models.VMActionRestart:
		for _, instance := range s.instances {
			if instance.ExternalID == instanceID || instance.ID == instanceID {
				switch action {
				case models.VMActionStart:
					instance.Status = "running"
				case models.VMActionStop:
					instance.Status = "stopped"
				case models.VMActionRestart:
					instance.Status = "running"
				}
				instance.UpdatedAt = time.Now()
				return nil
			}
		}
		return errors.New("instance not found")

	case models.VMActionDestroy:
		for id, instance := range s.instances {
			if instance.ExternalID == instanceID || instance.ID == instanceID {
				delete(s.instances, id)
				return nil
			}
		}
		return errors.New("instance not found")
	}

	return nil
}

// ListVMInstances lists VM instances
func (s *Service) ListVMInstances(providerID string) []*models.VMInstance {
	s.mu.RLock()
	defer s.mu.RUnlock()

	instances := make([]*models.VMInstance, 0)
	for _, instance := range s.instances {
		if providerID == "" || instance.ProviderID == providerID {
			instances = append(instances, instance)
		}
	}
	return instances
}
