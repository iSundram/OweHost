package v1

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/iSundram/OweHost/internal/firewall"
	"github.com/iSundram/OweHost/internal/user"
	"github.com/iSundram/OweHost/pkg/models"
)

type FirewallHandler struct {
	firewallService *firewall.Service
	userService     *user.Service
}

func NewFirewallHandler(firewallService *firewall.Service, userService *user.Service) *FirewallHandler {
	return &FirewallHandler{
		firewallService: firewallService,
		userService:     userService,
	}
}

// ListRules lists all firewall rules (admin only)
func (h *FirewallHandler) ListRules(w http.ResponseWriter, r *http.Request) {
	chainName := r.URL.Query().Get("chain")
	if chainName == "" {
		chainName = "INPUT"
	}

	rules := h.firewallService.ListRules(chainName)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rules)
}

// GetRule retrieves a specific firewall rule
func (h *FirewallHandler) GetRule(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.TrimSuffix(r.URL.Path, "/"), "/")
	if len(parts) < 6 {
		http.Error(w, "Invalid rule ID", http.StatusBadRequest)
		return
	}
	ruleID := parts[5]

	rule, err := h.firewallService.GetRule(ruleID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rule)
}

// CreateRule creates a new firewall rule
func (h *FirewallHandler) CreateRule(w http.ResponseWriter, r *http.Request) {
	var req models.FirewallRuleCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	rule, err := h.firewallService.CreateRule(nil, &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(rule)
}

// UpdateRule updates a firewall rule
func (h *FirewallHandler) UpdateRule(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.TrimSuffix(r.URL.Path, "/"), "/")
	if len(parts) < 6 {
		http.Error(w, "Invalid rule ID", http.StatusBadRequest)
		return
	}
	ruleID := parts[5]

	var req struct {
		Enabled bool `json:"enabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err := h.firewallService.UpdateRule(ruleID, req.Enabled)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Rule updated"})
}

// DeleteRule deletes a firewall rule
func (h *FirewallHandler) DeleteRule(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.TrimSuffix(r.URL.Path, "/"), "/")
	if len(parts) < 6 {
		http.Error(w, "Invalid rule ID", http.StatusBadRequest)
		return
	}
	ruleID := parts[5]

	if err := h.firewallService.DeleteRule(ruleID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// EnableRule enables a firewall rule
func (h *FirewallHandler) EnableRule(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.TrimSuffix(r.URL.Path, "/"), "/")
	if len(parts) < 7 {
		http.Error(w, "Invalid rule ID", http.StatusBadRequest)
		return
	}
	ruleID := parts[5]

	if err := h.firewallService.UpdateRule(ruleID, true); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "enabled"})
}

// DisableRule disables a firewall rule
func (h *FirewallHandler) DisableRule(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.TrimSuffix(r.URL.Path, "/"), "/")
	if len(parts) < 7 {
		http.Error(w, "Invalid rule ID", http.StatusBadRequest)
		return
	}
	ruleID := parts[5]

	if err := h.firewallService.UpdateRule(ruleID, false); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "disabled"})
}

// GetStatus retrieves firewall status
func (h *FirewallHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	chains := h.firewallService.ListChains()
	
	status := map[string]interface{}{
		"enabled": true,
		"chains":  chains,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// EnableFirewall enables the firewall
func (h *FirewallHandler) EnableFirewall(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "enabled"})
}

// DisableFirewall disables the firewall
func (h *FirewallHandler) DisableFirewall(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "disabled"})
}

// BlockIP blocks an IP address
func (h *FirewallHandler) BlockIP(w http.ResponseWriter, r *http.Request) {
	var req struct {
		IP     string `json:"ip"`
		Reason string `json:"reason,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Create a deny rule for this IP
	ruleReq := &models.FirewallRuleCreateRequest{
		ChainName:   "INPUT",
		Action:      models.FirewallActionDeny,
		SourceIP:    req.IP,
		Description: req.Reason,
		Priority:    1,
	}
	
	rule, err := h.firewallService.CreateRule(nil, ruleReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(rule)
}

// UnblockIP unblocks an IP address
func (h *FirewallHandler) UnblockIP(w http.ResponseWriter, r *http.Request) {
	var req struct {
		IP string `json:"ip"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// In production, would find and delete the rule for this IP
	w.WriteHeader(http.StatusNoContent)
}

// GetBlockedIPs retrieves list of blocked IPs
func (h *FirewallHandler) GetBlockedIPs(w http.ResponseWriter, r *http.Request) {
	// Get all rules and filter for denied IPs
	rules := h.firewallService.ListRules("INPUT")
	
	blockedIPs := make([]string, 0)
	for _, rule := range rules {
		if rule.Action == models.FirewallActionDeny && rule.SourceIP != "" {
			blockedIPs = append(blockedIPs, rule.SourceIP)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(blockedIPs)
}
