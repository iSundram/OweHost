// Package cluster provides filesystem-based cluster/node state management
package cluster

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const (
	// ClusterBasePath is the base path for cluster state
	ClusterBasePath = "/opt/owehost/cluster"
	// NodesPath is the path for node configurations
	NodesPath = "/opt/owehost/cluster/nodes"
)

// NodeConfig represents a cluster node configuration
type NodeConfig struct {
	ID          string    `json:"id"`
	IP          string    `json:"ip"`
	Hostname    string    `json:"hostname"`
	Roles       []string  `json:"roles"`       // web, data, mail, dns, backup
	Region      string    `json:"region,omitempty"`
	Datacenter  string    `json:"datacenter,omitempty"`
	Status      string    `json:"status"`      // online, offline, maintenance
	LastSeen    time.Time `json:"last_seen"`
	JoinedAt    time.Time `json:"joined_at"`
	Version     string    `json:"version"`     // OweHost version
	Resources   *NodeResources `json:"resources,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
}

// NodeResources represents available resources on a node
type NodeResources struct {
	CPUCores      int   `json:"cpu_cores"`
	RAMMB         int   `json:"ram_mb"`
	DiskGB        int   `json:"disk_gb"`
	UsedCPU       int   `json:"used_cpu_percent"`
	UsedRAM       int   `json:"used_ram_percent"`
	UsedDisk      int   `json:"used_disk_percent"`
	AccountCount  int   `json:"account_count"`
	DomainCount   int   `json:"domain_count"`
}

// NodeStatus constants
const (
	NodeStatusOnline      = "online"
	NodeStatusOffline     = "offline"
	NodeStatusMaintenance = "maintenance"
	NodeStatusDraining    = "draining"
)

// NodeRole constants
const (
	RoleWeb    = "web"
	RoleData   = "data"
	RoleMail   = "mail"
	RoleDNS    = "dns"
	RoleBackup = "backup"
	RoleAll    = "all"
)

// StateManager handles cluster state
type StateManager struct {
	basePath  string
	nodesPath string
	mu        sync.RWMutex
}

// NewStateManager creates a new cluster state manager
func NewStateManager() *StateManager {
	s := &StateManager{
		basePath:  ClusterBasePath,
		nodesPath: NodesPath,
	}
	s.ensureDirectories()
	return s
}

// ensureDirectories creates required directories
func (s *StateManager) ensureDirectories() {
	os.MkdirAll(s.nodesPath, 0755)
}

// GetNode retrieves a node by ID
func (s *StateManager) GetNode(nodeID string) (*NodeConfig, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	path := filepath.Join(s.nodesPath, nodeID+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("node not found: %s", nodeID)
		}
		return nil, err
	}

	var node NodeConfig
	if err := json.Unmarshal(data, &node); err != nil {
		return nil, err
	}

	return &node, nil
}

// SaveNode saves a node configuration
func (s *StateManager) SaveNode(node *NodeConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	path := filepath.Join(s.nodesPath, node.ID+".json")
	data, err := json.MarshalIndent(node, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// DeleteNode removes a node
func (s *StateManager) DeleteNode(nodeID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	path := filepath.Join(s.nodesPath, nodeID+".json")
	return os.Remove(path)
}

// ListNodes returns all nodes
func (s *StateManager) ListNodes() ([]NodeConfig, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entries, err := os.ReadDir(s.nodesPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []NodeConfig{}, nil
		}
		return nil, err
	}

	var nodes []NodeConfig
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		data, err := os.ReadFile(filepath.Join(s.nodesPath, entry.Name()))
		if err != nil {
			continue
		}

		var node NodeConfig
		if json.Unmarshal(data, &node) == nil {
			nodes = append(nodes, node)
		}
	}

	return nodes, nil
}

// ListOnlineNodes returns all online nodes
func (s *StateManager) ListOnlineNodes() ([]NodeConfig, error) {
	nodes, err := s.ListNodes()
	if err != nil {
		return nil, err
	}

	var online []NodeConfig
	for _, node := range nodes {
		if node.Status == NodeStatusOnline {
			online = append(online, node)
		}
	}

	return online, nil
}

// ListNodesByRole returns nodes with a specific role
func (s *StateManager) ListNodesByRole(role string) ([]NodeConfig, error) {
	nodes, err := s.ListNodes()
	if err != nil {
		return nil, err
	}

	var filtered []NodeConfig
	for _, node := range nodes {
		for _, r := range node.Roles {
			if r == role || r == RoleAll {
				filtered = append(filtered, node)
				break
			}
		}
	}

	return filtered, nil
}

// UpdateNodeStatus updates a node's status
func (s *StateManager) UpdateNodeStatus(nodeID, status string) error {
	node, err := s.GetNode(nodeID)
	if err != nil {
		return err
	}

	node.Status = status
	node.LastSeen = time.Now()
	return s.SaveNode(node)
}

// UpdateNodeResources updates a node's resource information
func (s *StateManager) UpdateNodeResources(nodeID string, resources *NodeResources) error {
	node, err := s.GetNode(nodeID)
	if err != nil {
		return err
	}

	node.Resources = resources
	node.LastSeen = time.Now()
	return s.SaveNode(node)
}

// ProcessHeartbeat processes a heartbeat from a node
func (s *StateManager) ProcessHeartbeat(nodeID string, resources *NodeResources) error {
	node, err := s.GetNode(nodeID)
	if err != nil {
		return err
	}

	node.Status = NodeStatusOnline
	node.LastSeen = time.Now()
	if resources != nil {
		node.Resources = resources
	}

	return s.SaveNode(node)
}

// GetDeadNodes returns nodes that haven't sent heartbeat within timeout
func (s *StateManager) GetDeadNodes(timeout time.Duration) ([]NodeConfig, error) {
	nodes, err := s.ListNodes()
	if err != nil {
		return nil, err
	}

	cutoff := time.Now().Add(-timeout)
	var dead []NodeConfig
	for _, node := range nodes {
		if node.Status == NodeStatusOnline && node.LastSeen.Before(cutoff) {
			dead = append(dead, node)
		}
	}

	return dead, nil
}

// MarkDeadNodes marks nodes as offline if they haven't sent heartbeat
func (s *StateManager) MarkDeadNodes(timeout time.Duration) ([]string, error) {
	dead, err := s.GetDeadNodes(timeout)
	if err != nil {
		return nil, err
	}

	var marked []string
	for _, node := range dead {
		if err := s.UpdateNodeStatus(node.ID, NodeStatusOffline); err == nil {
			marked = append(marked, node.ID)
		}
	}

	return marked, nil
}

// GetBestNodeForPlacement returns the best node for placing a new account
func (s *StateManager) GetBestNodeForPlacement(role string) (*NodeConfig, error) {
	nodes, err := s.ListNodesByRole(role)
	if err != nil {
		return nil, err
	}

	var best *NodeConfig
	var bestScore int = -1

	for i := range nodes {
		node := &nodes[i]
		if node.Status != NodeStatusOnline {
			continue
		}

		if node.Resources == nil {
			continue
		}

		// Simple scoring: prefer nodes with more free resources
		score := 100 - node.Resources.UsedCPU
		score += 100 - node.Resources.UsedRAM
		score += 100 - node.Resources.UsedDisk

		if score > bestScore {
			bestScore = score
			best = node
		}
	}

	if best == nil {
		return nil, fmt.Errorf("no suitable node found for role: %s", role)
	}

	return best, nil
}

// AccountPlacement represents where an account is placed
type AccountPlacement struct {
	AccountID int       `json:"account_id"`
	NodeID    string    `json:"node_id"`
	PlacedAt  time.Time `json:"placed_at"`
}

// GetAccountPlacement returns the node where an account is placed
func (s *StateManager) GetAccountPlacement(accountID int) (*AccountPlacement, error) {
	path := filepath.Join(s.basePath, "placements", fmt.Sprintf("a-%d.json", accountID))
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var placement AccountPlacement
	if err := json.Unmarshal(data, &placement); err != nil {
		return nil, err
	}

	return &placement, nil
}

// SetAccountPlacement sets the node placement for an account
func (s *StateManager) SetAccountPlacement(accountID int, nodeID string) error {
	os.MkdirAll(filepath.Join(s.basePath, "placements"), 0755)

	placement := AccountPlacement{
		AccountID: accountID,
		NodeID:    nodeID,
		PlacedAt:  time.Now(),
	}

	path := filepath.Join(s.basePath, "placements", fmt.Sprintf("a-%d.json", accountID))
	data, err := json.MarshalIndent(placement, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}
