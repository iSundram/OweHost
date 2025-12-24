// Package accountsvc provides account management using filesystem-based storage
package accountsvc

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/iSundram/OweHost/internal/storage/account"
	"github.com/iSundram/OweHost/internal/storage/events"
	"github.com/iSundram/OweHost/internal/storage/web"
)

// Service provides account management functionality using filesystem storage
type Service struct {
	accountState *account.StateManager
	accountApply *account.Applier
	webState     *web.StateManager
	webApply     *web.Applier
	events       *events.Emitter
	mu           sync.RWMutex
}

// NewService creates a new account service
func NewService() *Service {
	return &Service{
		accountState: account.NewStateManager(),
		accountApply: account.NewApplier(),
		webState:     web.NewStateManager(),
		webApply:     web.NewApplier(),
		events:       events.NewEmitter(),
	}
}

// CreateRequest represents a request to create an account
type CreateRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Plan     string `json:"plan"`
	Owner    string `json:"owner"` // Parent owner (admin, reseller-X)
}

// CreateResponse represents the response from creating an account
type CreateResponse struct {
	AccountID int                      `json:"account_id"`
	Identity  *account.AccountIdentity `json:"identity"`
	Limits    *account.ResourceLimits  `json:"limits"`
}

// Create creates a new account
func (s *Service) Create(ctx context.Context, req *CreateRequest, actor, actorType, actorIP string) (*CreateResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Get next account ID
	accountID, err := s.accountState.GetNextAccountID()
	if err != nil {
		return nil, fmt.Errorf("failed to get next account ID: %w", err)
	}

	// Create identity
	identity := &account.AccountIdentity{
		ID:        accountID,
		Name:      req.Username,
		UID:       10000 + accountID,
		GID:       10000 + accountID,
		Owner:     req.Owner,
		Plan:      req.Plan,
		Node:      "node-1", // Default node, would be selected by scheduler
		CreatedAt: time.Now().Format(time.RFC3339),
		State:     account.StateActive,
	}

	// Validate identity
	if err := account.ValidateIdentity(identity); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Get plan limits
	limits := account.GetPlanLimits(req.Plan)

	// Create metadata
	metadata := &account.AccountMetadata{
		Email:     req.Email,
		UpdatedAt: time.Now().Format(time.RFC3339),
	}

	// Apply to filesystem (source of truth)
	config := &account.ApplyConfig{
		Identity: identity,
		Limits:   &limits,
		Status:   &account.AccountStatus{},
		Metadata: metadata,
	}

	if err := s.accountApply.Apply(accountID, config); err != nil {
		return nil, fmt.Errorf("failed to apply account: %w", err)
	}

	// Emit event
	s.events.AccountCreated(accountID, req.Username, actor, actorType, actorIP)

	return &CreateResponse{
		AccountID: accountID,
		Identity:  identity,
		Limits:    &limits,
	}, nil
}

// Get retrieves an account by ID
func (s *Service) Get(ctx context.Context, accountID int) (*account.Account, error) {
	return s.accountState.ReadAccount(accountID)
}

// List returns all accounts
func (s *Service) List(ctx context.Context) ([]account.Account, error) {
	accountIDs, err := s.accountState.ListAccounts()
	if err != nil {
		return nil, err
	}

	var accounts []account.Account
	for _, id := range accountIDs {
		acc, err := s.accountState.ReadAccount(id)
		if err == nil && acc != nil {
			accounts = append(accounts, *acc)
		}
	}

	return accounts, nil
}

// Suspend suspends an account
func (s *Service) Suspend(ctx context.Context, accountID int, reason, actor, actorType string) error {
	if err := s.accountApply.Suspend(accountID, reason, actor); err != nil {
		return err
	}

	s.events.AccountSuspended(accountID, reason, actor, actorType)
	return nil
}

// Unsuspend unsuspends an account
func (s *Service) Unsuspend(ctx context.Context, accountID int, actor, actorType string) error {
	if err := s.accountApply.Unsuspend(accountID); err != nil {
		return err
	}

	s.events.AccountUnsuspended(accountID, actor, actorType)
	return nil
}

// Terminate terminates an account
func (s *Service) Terminate(ctx context.Context, accountID int, reason, actor, actorType string) error {
	if err := s.accountApply.Terminate(accountID, reason, actor); err != nil {
		return err
	}

	s.events.AccountTerminated(accountID, reason, actor, actorType)
	return nil
}

// Delete permanently deletes an account
func (s *Service) Delete(ctx context.Context, accountID int, actor, actorType string) error {
	// Get account info before deletion for event
	acc, _ := s.accountState.ReadAccount(accountID)
	accountName := ""
	if acc != nil && acc.Identity != nil {
		accountName = acc.Identity.Name
	}

	if err := s.accountApply.Delete(accountID); err != nil {
		return err
	}

	s.events.EmitSuccess(events.EventAccountDelete, events.EmitOptions{
		AccountID: accountID,
		Actor:     actor,
		ActorType: actorType,
		Data: map[string]interface{}{
			"account_name": accountName,
		},
	})

	return nil
}

// UpdateLimits updates account resource limits
func (s *Service) UpdateLimits(ctx context.Context, accountID int, limits *account.ResourceLimits, actor, actorType string) error {
	if err := account.ValidateLimits(limits); err != nil {
		return err
	}

	if err := s.accountState.WriteLimits(accountID, limits); err != nil {
		return err
	}

	s.events.EmitSuccess(events.EventAccountUpdate, events.EmitOptions{
		AccountID: accountID,
		Actor:     actor,
		ActorType: actorType,
		Data: map[string]interface{}{
			"action": "update_limits",
		},
	})

	return nil
}

// ChangePlan changes an account's plan
func (s *Service) ChangePlan(ctx context.Context, accountID int, newPlan, actor, actorType string) error {
	if !account.IsValidPlan(newPlan) {
		return fmt.Errorf("invalid plan: %s", newPlan)
	}

	// Read current identity
	identity, err := s.accountState.ReadIdentity(accountID)
	if err != nil {
		return err
	}

	oldPlan := identity.Plan
	identity.Plan = newPlan

	// Write updated identity
	if err := s.accountState.WriteIdentity(accountID, identity); err != nil {
		return err
	}

	// Update limits based on new plan
	limits := account.GetPlanLimits(newPlan)
	if err := s.accountState.WriteLimits(accountID, &limits); err != nil {
		return err
	}

	s.events.EmitSuccess(events.EventAccountUpdate, events.EmitOptions{
		AccountID: accountID,
		Actor:     actor,
		ActorType: actorType,
		Data: map[string]interface{}{
			"action":   "change_plan",
			"old_plan": oldPlan,
			"new_plan": newPlan,
		},
	})

	return nil
}

// AddDomain adds a domain to an account
func (s *Service) AddDomain(ctx context.Context, accountID int, site *web.SiteDescriptor, actor, actorType string) error {
	// Check domain limit
	limits, _ := s.accountState.ReadLimits(accountID)
	if limits != nil && limits.Domains > 0 {
		sites, _ := s.webState.ListSites(accountID)
		if len(sites) >= limits.Domains {
			return fmt.Errorf("domain limit exceeded: %d/%d", len(sites), limits.Domains)
		}
	}

	// Apply the site
	if err := s.webApply.ApplySite(accountID, site); err != nil {
		return err
	}

	s.events.DomainAdded(accountID, site.Domain, actor, actorType)
	return nil
}

// RemoveDomain removes a domain from an account
func (s *Service) RemoveDomain(ctx context.Context, accountID int, domain, actor, actorType string) error {
	if err := s.webApply.DeleteSite(accountID, domain); err != nil {
		return err
	}

	s.events.DomainRemoved(accountID, domain, actor, actorType)
	return nil
}

// ListDomains lists all domains for an account
func (s *Service) ListDomains(ctx context.Context, accountID int) ([]web.SiteDescriptor, error) {
	return s.webState.ListSites(accountID)
}

// GetUsage returns current resource usage for an account
func (s *Service) GetUsage(ctx context.Context, accountID int) (*account.AccountUsage, error) {
	sites, _ := s.webState.ListSites(accountID)

	return &account.AccountUsage{
		DomainCount: len(sites),
		LastUpdated: time.Now(),
	}, nil
}

// Exists checks if an account exists
func (s *Service) Exists(ctx context.Context, accountID int) bool {
	return s.accountState.Exists(accountID)
}

// GetByUsername finds an account by username
func (s *Service) GetByUsername(ctx context.Context, username string) (*account.Account, error) {
	accountIDs, err := s.accountState.ListAccounts()
	if err != nil {
		return nil, err
	}

	for _, id := range accountIDs {
		acc, err := s.accountState.ReadAccount(id)
		if err != nil {
			continue
		}
		if acc.Identity != nil && acc.Identity.Name == username {
			return acc, nil
		}
	}

	return nil, fmt.Errorf("account not found: %s", username)
}
