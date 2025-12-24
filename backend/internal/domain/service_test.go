package domain_test

import (
	"testing"

	"github.com/iSundram/OweHost/internal/domain"
	"github.com/iSundram/OweHost/pkg/models"
)

func TestDomainService_Create(t *testing.T) {
	svc := domain.NewService()

	req := &models.DomainCreateRequest{
		Name: "example.com",
		Type: models.DomainTypePrimary,
	}

	d, err := svc.Create("user-123", req)
	if err != nil {
		t.Fatalf("Failed to create domain: %v", err)
	}

	if d.Name != req.Name {
		t.Errorf("Expected name %s, got %s", req.Name, d.Name)
	}

	if d.UserID != "user-123" {
		t.Errorf("Expected user_id user-123, got %s", d.UserID)
	}

	if d.Status != models.DomainStatusPending {
		t.Errorf("Expected status pending, got %s", d.Status)
	}
}

func TestDomainService_DuplicateDomain(t *testing.T) {
	svc := domain.NewService()

	req := &models.DomainCreateRequest{
		Name: "duplicate.com",
		Type: models.DomainTypePrimary,
	}

	_, err := svc.Create("user-123", req)
	if err != nil {
		t.Fatalf("Failed to create first domain: %v", err)
	}

	_, err = svc.Create("user-456", req)
	if err == nil {
		t.Error("Expected error for duplicate domain")
	}
}

func TestDomainService_ListByUser(t *testing.T) {
	svc := domain.NewService()

	req1 := &models.DomainCreateRequest{
		Name: "domain1.com",
		Type: models.DomainTypePrimary,
	}
	req2 := &models.DomainCreateRequest{
		Name: "domain2.com",
		Type: models.DomainTypeAddon,
	}

	_, _ = svc.Create("user-list", req1)
	_, _ = svc.Create("user-list", req2)
	_, _ = svc.Create("user-other", &models.DomainCreateRequest{Name: "other.com", Type: models.DomainTypePrimary})

	domains := svc.ListByUser("user-list")
	if len(domains) != 2 {
		t.Errorf("Expected 2 domains, got %d", len(domains))
	}
}

func TestDomainService_CreateSubdomain(t *testing.T) {
	svc := domain.NewService()

	domainReq := &models.DomainCreateRequest{
		Name: "subdomain-test.com",
		Type: models.DomainTypePrimary,
	}

	d, err := svc.Create("user-123", domainReq)
	if err != nil {
		t.Fatalf("Failed to create domain: %v", err)
	}

	subReq := &models.SubdomainCreateRequest{
		Name: "blog",
	}

	sub, err := svc.CreateSubdomain(d.ID, subReq)
	if err != nil {
		t.Fatalf("Failed to create subdomain: %v", err)
	}

	if sub.FullName != "blog.subdomain-test.com" {
		t.Errorf("Expected full name blog.subdomain-test.com, got %s", sub.FullName)
	}
}

func TestDomainService_CheckOwnership(t *testing.T) {
	svc := domain.NewService()

	req := &models.DomainCreateRequest{
		Name: "ownership.com",
		Type: models.DomainTypePrimary,
	}

	d, err := svc.Create("owner-user", req)
	if err != nil {
		t.Fatalf("Failed to create domain: %v", err)
	}

	if !svc.CheckOwnership("owner-user", d.ID) {
		t.Error("Expected ownership check to pass for owner")
	}

	if svc.CheckOwnership("other-user", d.ID) {
		t.Error("Expected ownership check to fail for non-owner")
	}
}
