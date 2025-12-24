// Package firewall provides firewall and security services for OweHost
package firewall

import (
	"errors"
	"sync"
	"time"

	"github.com/iSundram/OweHost/pkg/models"
	"github.com/iSundram/OweHost/pkg/utils"
)

// Service provides firewall functionality
type Service struct {
	rules        map[string]*models.FirewallRule
	chains       map[string]*models.FirewallChain
	rateLimits   map[string]*models.RateLimit
	events       []*models.IntrusionEvent
	byChain      map[string][]*models.FirewallRule
	mu           sync.RWMutex
}

// NewService creates a new firewall service
func NewService() *Service {
	svc := &Service{
		rules:      make(map[string]*models.FirewallRule),
		chains:     make(map[string]*models.FirewallChain),
		rateLimits: make(map[string]*models.RateLimit),
		events:     make([]*models.IntrusionEvent, 0),
		byChain:    make(map[string][]*models.FirewallRule),
	}
	svc.initDefaultChains()
	return svc
}

// initDefaultChains initializes default firewall chains
func (s *Service) initDefaultChains() {
	chains := []string{"INPUT", "OUTPUT", "FORWARD"}
	for _, name := range chains {
		s.chains[name] = &models.FirewallChain{
			Name:       name,
			Policy:     models.FirewallActionAllow,
			RulesCount: 0,
			CreatedAt:  time.Now(),
		}
		s.byChain[name] = make([]*models.FirewallRule, 0)
	}
}

// CreateRule creates a firewall rule
func (s *Service) CreateRule(userID *string, req *models.FirewallRuleCreateRequest) (*models.FirewallRule, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.chains[req.ChainName]; !exists {
		return nil, errors.New("chain not found")
	}

	rule := &models.FirewallRule{
		ID:          utils.GenerateID("fwr"),
		UserID:      userID,
		ChainName:   req.ChainName,
		Priority:    req.Priority,
		Action:      req.Action,
		Protocol:    req.Protocol,
		SourceIP:    req.SourceIP,
		DestIP:      req.DestIP,
		SourcePort:  req.SourcePort,
		DestPort:    req.DestPort,
		Description: req.Description,
		Enabled:     true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	s.rules[rule.ID] = rule
	s.byChain[req.ChainName] = append(s.byChain[req.ChainName], rule)
	s.chains[req.ChainName].RulesCount++

	// Sort by priority
	s.sortChainRules(req.ChainName)

	return rule, nil
}

// GetRule gets a firewall rule by ID
func (s *Service) GetRule(id string) (*models.FirewallRule, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rule, exists := s.rules[id]
	if !exists {
		return nil, errors.New("rule not found")
	}
	return rule, nil
}

// ListRules lists rules for a chain
func (s *Service) ListRules(chainName string) []*models.FirewallRule {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.byChain[chainName]
}

// UpdateRule updates a firewall rule
func (s *Service) UpdateRule(id string, enabled bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	rule, exists := s.rules[id]
	if !exists {
		return errors.New("rule not found")
	}

	rule.Enabled = enabled
	rule.UpdatedAt = time.Now()
	return nil
}

// DeleteRule deletes a firewall rule
func (s *Service) DeleteRule(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	rule, exists := s.rules[id]
	if !exists {
		return errors.New("rule not found")
	}

	// Remove from chain
	chainRules := s.byChain[rule.ChainName]
	for i, r := range chainRules {
		if r.ID == id {
			s.byChain[rule.ChainName] = append(chainRules[:i], chainRules[i+1:]...)
			break
		}
	}
	s.chains[rule.ChainName].RulesCount--

	delete(s.rules, id)
	return nil
}

// CreateChain creates a custom firewall chain
func (s *Service) CreateChain(name string, policy models.FirewallAction) (*models.FirewallChain, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.chains[name]; exists {
		return nil, errors.New("chain already exists")
	}

	chain := &models.FirewallChain{
		Name:       name,
		Policy:     policy,
		RulesCount: 0,
		CreatedAt:  time.Now(),
	}

	s.chains[name] = chain
	s.byChain[name] = make([]*models.FirewallRule, 0)

	return chain, nil
}

// GetChain gets a firewall chain
func (s *Service) GetChain(name string) (*models.FirewallChain, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	chain, exists := s.chains[name]
	if !exists {
		return nil, errors.New("chain not found")
	}
	return chain, nil
}

// ListChains lists all firewall chains
func (s *Service) ListChains() []*models.FirewallChain {
	s.mu.RLock()
	defer s.mu.RUnlock()

	chains := make([]*models.FirewallChain, 0, len(s.chains))
	for _, chain := range s.chains {
		chains = append(chains, chain)
	}
	return chains
}

// SetChainPolicy sets the default policy for a chain
func (s *Service) SetChainPolicy(name string, policy models.FirewallAction) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	chain, exists := s.chains[name]
	if !exists {
		return errors.New("chain not found")
	}

	chain.Policy = policy
	return nil
}

// DeleteChain deletes a custom firewall chain
func (s *Service) DeleteChain(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	chain, exists := s.chains[name]
	if !exists {
		return errors.New("chain not found")
	}

	// Don't allow deleting default chains
	if name == "INPUT" || name == "OUTPUT" || name == "FORWARD" {
		return errors.New("cannot delete default chain")
	}

	if chain.RulesCount > 0 {
		return errors.New("chain has rules, delete them first")
	}

	delete(s.chains, name)
	delete(s.byChain, name)
	return nil
}

// CreateRateLimit creates a rate limit
func (s *Service) CreateRateLimit(userID *string, req *models.RateLimitCreateRequest) (*models.RateLimit, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	rateLimit := &models.RateLimit{
		ID:                utils.GenerateID("rl"),
		UserID:            userID,
		Type:              req.Type,
		IPAddress:         req.IPAddress,
		RequestsPerSecond: req.RequestsPerSecond,
		BurstSize:         req.BurstSize,
		Enabled:           true,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	if rateLimit.BurstSize == 0 {
		rateLimit.BurstSize = req.RequestsPerSecond * 2
	}

	s.rateLimits[rateLimit.ID] = rateLimit
	return rateLimit, nil
}

// GetRateLimit gets a rate limit by ID
func (s *Service) GetRateLimit(id string) (*models.RateLimit, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rateLimit, exists := s.rateLimits[id]
	if !exists {
		return nil, errors.New("rate limit not found")
	}
	return rateLimit, nil
}

// ListRateLimits lists rate limits for a user
func (s *Service) ListRateLimits(userID *string) []*models.RateLimit {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rateLimits := make([]*models.RateLimit, 0)
	for _, rl := range s.rateLimits {
		if userID == nil || (rl.UserID != nil && *rl.UserID == *userID) {
			rateLimits = append(rateLimits, rl)
		}
	}
	return rateLimits
}

// UpdateRateLimit updates a rate limit
func (s *Service) UpdateRateLimit(id string, rps, burst int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	rateLimit, exists := s.rateLimits[id]
	if !exists {
		return errors.New("rate limit not found")
	}

	rateLimit.RequestsPerSecond = rps
	rateLimit.BurstSize = burst
	rateLimit.UpdatedAt = time.Now()
	return nil
}

// DeleteRateLimit deletes a rate limit
func (s *Service) DeleteRateLimit(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.rateLimits[id]; !exists {
		return errors.New("rate limit not found")
	}

	delete(s.rateLimits, id)
	return nil
}

// EmitIntrusionEvent emits an intrusion detection event
func (s *Service) EmitIntrusionEvent(eventType, severity, sourceIP, targetIP, description, rawData string) *models.IntrusionEvent {
	s.mu.Lock()
	defer s.mu.Unlock()

	event := &models.IntrusionEvent{
		ID:          utils.GenerateID("ids"),
		EventType:   eventType,
		Severity:    severity,
		SourceIP:    sourceIP,
		TargetIP:    targetIP,
		Description: description,
		RawData:     rawData,
		DetectedAt:  time.Now(),
	}

	s.events = append(s.events, event)

	// Keep only last 10000 events
	if len(s.events) > 10000 {
		s.events = s.events[len(s.events)-10000:]
	}

	return event
}

// ListIntrusionEvents lists intrusion events
func (s *Service) ListIntrusionEvents(limit int) []*models.IntrusionEvent {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if limit <= 0 || limit > len(s.events) {
		limit = len(s.events)
	}

	// Return most recent events
	start := len(s.events) - limit
	return s.events[start:]
}

// CheckRateLimit checks if a request should be rate limited
func (s *Service) CheckRateLimit(ip string, userID *string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, rl := range s.rateLimits {
		if !rl.Enabled {
			continue
		}

		// Check IP-based rate limit
		if rl.Type == "ip" && rl.IPAddress != nil && *rl.IPAddress == ip {
			// In production, would use a token bucket or sliding window
			return true
		}

		// Check user-based rate limit
		if rl.Type == "user" && userID != nil && rl.UserID != nil && *rl.UserID == *userID {
			return true
		}

		// Check global rate limit
		if rl.Type == "global" {
			return true
		}
	}

	return false
}

// sortChainRules sorts rules in a chain by priority
func (s *Service) sortChainRules(chainName string) {
	rules := s.byChain[chainName]
	// Bubble sort for simplicity (use sort.Slice in production)
	for i := 0; i < len(rules); i++ {
		for j := i + 1; j < len(rules); j++ {
			if rules[i].Priority > rules[j].Priority {
				rules[i], rules[j] = rules[j], rules[i]
			}
		}
	}
}
