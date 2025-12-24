package user_test

import (
	"testing"

	"github.com/iSundram/OweHost/internal/user"
	"github.com/iSundram/OweHost/pkg/config"
	"github.com/iSundram/OweHost/pkg/models"
)

func TestUserService_Create(t *testing.T) {
	cfg := config.Load()
	svc := user.NewService(cfg)

	req := &models.UserCreateRequest{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
	}

	u, err := svc.Create(req)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	if u.Username != req.Username {
		t.Errorf("Expected username %s, got %s", req.Username, u.Username)
	}

	if u.Email != req.Email {
		t.Errorf("Expected email %s, got %s", req.Email, u.Email)
	}

	if u.Status != models.UserStatusActive {
		t.Errorf("Expected status active, got %s", u.Status)
	}

	if u.ID == "" {
		t.Error("Expected non-empty ID")
	}
}

func TestUserService_Get(t *testing.T) {
	cfg := config.Load()
	svc := user.NewService(cfg)

	req := &models.UserCreateRequest{
		Username: "getuser",
		Email:    "getuser@example.com",
		Password: "password123",
	}

	created, err := svc.Create(req)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	u, err := svc.Get(created.ID)
	if err != nil {
		t.Fatalf("Failed to get user: %v", err)
	}

	if u.ID != created.ID {
		t.Errorf("Expected ID %s, got %s", created.ID, u.ID)
	}
}

func TestUserService_DuplicateEmail(t *testing.T) {
	cfg := config.Load()
	svc := user.NewService(cfg)

	req := &models.UserCreateRequest{
		Username: "user1",
		Email:    "duplicate@example.com",
		Password: "password123",
	}

	_, err := svc.Create(req)
	if err != nil {
		t.Fatalf("Failed to create first user: %v", err)
	}

	req2 := &models.UserCreateRequest{
		Username: "user2",
		Email:    "duplicate@example.com",
		Password: "password123",
	}

	_, err = svc.Create(req2)
	if err == nil {
		t.Error("Expected error for duplicate email")
	}
}

func TestUserService_DuplicateUsername(t *testing.T) {
	cfg := config.Load()
	svc := user.NewService(cfg)

	req := &models.UserCreateRequest{
		Username: "duplicateuser",
		Email:    "user1@example.com",
		Password: "password123",
	}

	_, err := svc.Create(req)
	if err != nil {
		t.Fatalf("Failed to create first user: %v", err)
	}

	req2 := &models.UserCreateRequest{
		Username: "duplicateuser",
		Email:    "user2@example.com",
		Password: "password123",
	}

	_, err = svc.Create(req2)
	if err == nil {
		t.Error("Expected error for duplicate username")
	}
}

func TestUserService_Suspend(t *testing.T) {
	cfg := config.Load()
	svc := user.NewService(cfg)

	req := &models.UserCreateRequest{
		Username: "suspenduser",
		Email:    "suspend@example.com",
		Password: "password123",
	}

	u, err := svc.Create(req)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	err = svc.Suspend(u.ID)
	if err != nil {
		t.Fatalf("Failed to suspend user: %v", err)
	}

	suspended, _ := svc.Get(u.ID)
	if suspended.Status != models.UserStatusSuspended {
		t.Errorf("Expected status suspended, got %s", suspended.Status)
	}
}

func TestUserService_ValidateCredentials(t *testing.T) {
	cfg := config.Load()
	svc := user.NewService(cfg)

	req := &models.UserCreateRequest{
		Username: "authuser",
		Email:    "auth@example.com",
		Password: "password123",
	}

	_, err := svc.Create(req)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Valid credentials
	u, err := svc.ValidateCredentials("authuser", "password123")
	if err != nil {
		t.Fatalf("Failed to validate credentials: %v", err)
	}
	if u.Username != "authuser" {
		t.Errorf("Expected username authuser, got %s", u.Username)
	}

	// Invalid password
	_, err = svc.ValidateCredentials("authuser", "wrongpassword")
	if err == nil {
		t.Error("Expected error for invalid password")
	}

	// Invalid username
	_, err = svc.ValidateCredentials("nonexistent", "password123")
	if err == nil {
		t.Error("Expected error for invalid username")
	}
}

func TestUserService_Delete(t *testing.T) {
	cfg := config.Load()
	svc := user.NewService(cfg)

	req := &models.UserCreateRequest{
		Username: "deleteuser",
		Email:    "delete@example.com",
		Password: "password123",
	}

	u, err := svc.Create(req)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	err = svc.Delete(u.ID)
	if err != nil {
		t.Fatalf("Failed to delete user: %v", err)
	}

	_, err = svc.Get(u.ID)
	if err == nil {
		t.Error("Expected error for deleted user")
	}
}
