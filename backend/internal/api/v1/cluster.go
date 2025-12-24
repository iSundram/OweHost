package v1

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/iSundram/OweHost/internal/cluster"
	"github.com/iSundram/OweHost/pkg/models"
	"github.com/iSundram/OweHost/pkg/utils"
)

// ClusterHandler handles cluster and node endpoints
type ClusterHandler struct {
	clusterService *cluster.Service
}

// NewClusterHandler creates a new cluster handler
func NewClusterHandler(clusterSvc *cluster.Service) *ClusterHandler {
	return &ClusterHandler{
		clusterService: clusterSvc,
	}
}

// ListNodes handles listing all nodes
func (h *ClusterHandler) ListNodes(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	nodes := h.clusterService.ListNodes()
	utils.WriteSuccess(w, nodes)
}

// GetNode handles getting a node by ID
func (h *ClusterHandler) GetNode(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Node ID required")
		return
	}
	nodeID := parts[len(parts)-1]

	node, err := h.clusterService.GetNode(nodeID)
	if err != nil {
		utils.WriteError(w, http.StatusNotFound, utils.ErrCodeNotFound, err.Error())
		return
	}

	utils.WriteSuccess(w, node)
}

// RegisterNode handles node registration
func (h *ClusterHandler) RegisterNode(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	var req models.NodeRegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Invalid request body")
		return
	}

	node, err := h.clusterService.RegisterNode(&req)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, err.Error())
		return
	}

	utils.WriteCreated(w, node)
}

// RemoveNode handles node removal
func (h *ClusterHandler) RemoveNode(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Node ID required")
		return
	}
	nodeID := parts[len(parts)-1]

	if err := h.clusterService.RemoveNode(nodeID); err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// UpdateNodeStatus handles updating a node's status
func (h *ClusterHandler) UpdateNodeStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut && r.Method != http.MethodPatch {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Node ID required")
		return
	}
	nodeID := parts[len(parts)-2]

	var req struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Invalid request body")
		return
	}

	var status models.NodeStatus
	switch req.Status {
	case "online":
		status = models.NodeStatusOnline
	case "offline":
		status = models.NodeStatusOffline
	case "maintenance":
		status = models.NodeStatusMaintenance
	default:
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Invalid status")
		return
	}

	if err := h.clusterService.UpdateNodeStatus(nodeID, status); err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, err.Error())
		return
	}

	node, _ := h.clusterService.GetNode(nodeID)
	utils.WriteSuccess(w, node)
}

// ProcessHeartbeat handles node heartbeat
func (h *ClusterHandler) ProcessHeartbeat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Node ID required")
		return
	}
	nodeID := parts[len(parts)-2]

	var req models.NodeHeartbeat
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Invalid request body")
		return
	}

	req.NodeID = nodeID
	req.Timestamp = time.Now()

	if err := h.clusterService.ProcessHeartbeat(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, err.Error())
		return
	}

	utils.WriteSuccess(w, map[string]string{"message": "Heartbeat received"})
}

// GetOnlineNodes handles listing online nodes
func (h *ClusterHandler) GetOnlineNodes(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	nodes := h.clusterService.GetOnlineNodes()
	utils.WriteSuccess(w, nodes)
}

// DiscoverCapabilities handles discovering node capabilities
func (h *ClusterHandler) DiscoverCapabilities(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	capabilities := h.clusterService.DiscoverCapabilities()
	utils.WriteSuccess(w, capabilities)
}

// PlaceResource handles resource placement
func (h *ClusterHandler) PlaceResource(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	var req models.PlacementRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Invalid request body")
		return
	}

	result, err := h.clusterService.PlaceResource(&req)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, err.Error())
		return
	}

	utils.WriteSuccess(w, result)
}

// CheckDeadNodes handles checking for dead nodes
func (h *ClusterHandler) CheckDeadNodes(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	var req struct {
		TimeoutSeconds int `json:"timeout_seconds"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		req.TimeoutSeconds = 60 // Default 60 seconds
	}

	timeout := time.Duration(req.TimeoutSeconds) * time.Second
	deadNodes := h.clusterService.CheckDeadNodes(timeout)

	utils.WriteSuccess(w, map[string]interface{}{
		"dead_nodes": deadNodes,
		"count":      len(deadNodes),
	})
}

// ListCloudProviders handles listing cloud providers
func (h *ClusterHandler) ListCloudProviders(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	providers := h.clusterService.ListCloudProviders()
	utils.WriteSuccess(w, providers)
}

// GetCloudProvider handles getting a cloud provider
func (h *ClusterHandler) GetCloudProvider(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Provider ID required")
		return
	}
	providerID := parts[len(parts)-1]

	provider, err := h.clusterService.GetCloudProvider(providerID)
	if err != nil {
		utils.WriteError(w, http.StatusNotFound, utils.ErrCodeNotFound, err.Error())
		return
	}

	utils.WriteSuccess(w, provider)
}

// RegisterCloudProvider handles cloud provider registration
func (h *ClusterHandler) RegisterCloudProvider(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	var req struct {
		Name        string                 `json:"name"`
		Type        string                 `json:"type"`
		Region      string                 `json:"region"`
		Credentials map[string]string      `json:"credentials"`
		Settings    map[string]interface{} `json:"settings"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Invalid request body")
		return
	}

	provider, err := h.clusterService.RegisterCloudProvider(req.Name, req.Type, req.Region, req.Credentials, req.Settings)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, err.Error())
		return
	}

	utils.WriteCreated(w, provider)
}

// DeleteCloudProvider handles cloud provider deletion
func (h *ClusterHandler) DeleteCloudProvider(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Provider ID required")
		return
	}
	providerID := parts[len(parts)-1]

	if err := h.clusterService.DeleteCloudProvider(providerID); err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ListVMInstances handles listing VM instances
func (h *ClusterHandler) ListVMInstances(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	providerID := r.URL.Query().Get("provider_id")
	instances := h.clusterService.ListVMInstances(providerID)
	utils.WriteSuccess(w, instances)
}

// VMLifecycle handles VM lifecycle actions
func (h *ClusterHandler) VMLifecycle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	var req struct {
		ProviderID string `json:"provider_id"`
		InstanceID string `json:"instance_id"`
		Action     string `json:"action"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Invalid request body")
		return
	}

	var action models.VMLifecycleAction
	switch req.Action {
	case "create":
		action = models.VMActionCreate
	case "start":
		action = models.VMActionStart
	case "stop":
		action = models.VMActionStop
	case "restart":
		action = models.VMActionRestart
	case "destroy":
		action = models.VMActionDestroy
	default:
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Invalid action")
		return
	}

	if err := h.clusterService.VMLifecycle(req.ProviderID, req.InstanceID, action); err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, err.Error())
		return
	}

	utils.WriteSuccess(w, map[string]string{"message": "Action completed"})
}
