// Package provisioning provides account provisioning orchestration for OweHost
package provisioning

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/iSundram/OweHost/internal/backup"
	"github.com/iSundram/OweHost/internal/database"
	"github.com/iSundram/OweHost/internal/dns"
	"github.com/iSundram/OweHost/internal/domain"
	"github.com/iSundram/OweHost/internal/filesystem"
	"github.com/iSundram/OweHost/internal/oscontrol"
	"github.com/iSundram/OweHost/internal/resource"
	"github.com/iSundram/OweHost/internal/ssl"
	"github.com/iSundram/OweHost/internal/user"
	"github.com/iSundram/OweHost/internal/webserver"
	"github.com/iSundram/OweHost/pkg/models"
	"github.com/iSundram/OweHost/pkg/utils"
)

// Service orchestrates complete account provisioning
type Service struct {
	userService       *user.Service
	domainService     *domain.Service
	databaseService   *database.Service
	filesystemService *filesystem.Service
	dnsService        *dns.Service
	sslService        *ssl.Service
	resourceService   *resource.Service
	oscontrolService  *oscontrol.Service
	webserverService  *webserver.Service
	backupService     *backup.Service

	provisionings map[string]*ProvisioningStatus
	mu            sync.RWMutex
}

// ProvisioningStatus tracks the status of account provisioning
type ProvisioningStatus struct {
	ID          string             `json:"id"`
	UserID      string             `json:"user_id"`
	Status      string             `json:"status"` // pending, in_progress, completed, failed, rolling_back, rolled_back
	Progress    int                `json:"progress"`
	Steps       []ProvisioningStep `json:"steps"`
	Error       string             `json:"error,omitempty"`
	StartedAt   time.Time          `json:"started_at"`
	CompletedAt *time.Time         `json:"completed_at,omitempty"`
}

// ProvisioningStep represents a single provisioning step
type ProvisioningStep struct {
	Name        string     `json:"name"`
	Status      string     `json:"status"` // pending, running, completed, failed, rolled_back
	StartedAt   *time.Time `json:"started_at,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	Error       string     `json:"error,omitempty"`
}

// AccountProvisionRequest contains all data needed to provision an account
type AccountProvisionRequest struct {
	Username       string          `json:"username"`
	Email          string          `json:"email"`
	Password       string          `json:"password"`
	Role           models.UserRole `json:"role"`
	Domain         string          `json:"domain,omitempty"`
	Package        interface{}     `json:"package,omitempty"`
	CreateDatabase bool            `json:"create_database"`
	EnableSSL      bool            `json:"enable_ssl"`
	SetupBackup    bool            `json:"setup_backup"`
	InstallApps    []string        `json:"install_apps,omitempty"`
	PHPVersion     string          `json:"php_version,omitempty"`
	DatabaseType   string          `json:"database_type,omitempty"`
}

// AccountProvisionResult contains the result of provisioning
type AccountProvisionResult struct {
	Status    *ProvisioningStatus `json:"status"`
	User      *models.User        `json:"user"`
	Domain    *models.Domain      `json:"domain,omitempty"`
	Database  *models.Database    `json:"database,omitempty"`
	SSLCert   *models.Certificate `json:"ssl_cert,omitempty"`
	HomeDir   string              `json:"home_dir"`
	SystemUID int                 `json:"system_uid"`
	SystemGID int                 `json:"system_gid"`
}

// NewService creates a new provisioning service
func NewService(
	userSvc *user.Service,
	domainSvc *domain.Service,
	dbSvc *database.Service,
	fsSvc *filesystem.Service,
	dnsSvc *dns.Service,
	sslSvc *ssl.Service,
	resourceSvc *resource.Service,
	osSvc *oscontrol.Service,
	webserverSvc *webserver.Service,
	backupSvc *backup.Service,
) *Service {
	return &Service{
		userService:       userSvc,
		domainService:     domainSvc,
		databaseService:   dbSvc,
		filesystemService: fsSvc,
		dnsService:        dnsSvc,
		sslService:        sslSvc,
		resourceService:   resourceSvc,
		oscontrolService:  osSvc,
		webserverService:  webserverSvc,
		backupService:     backupSvc,
		provisionings:     make(map[string]*ProvisioningStatus),
	}
}

// ProvisionAccount provisions a complete account with all requested features
func (s *Service) ProvisionAccount(req *AccountProvisionRequest) (*AccountProvisionResult, error) {
	provisioningID := utils.GenerateID("prov")
	status := &ProvisioningStatus{
		ID:        provisioningID,
		Status:    "in_progress",
		Progress:  0,
		StartedAt: time.Now(),
		Steps: []ProvisioningStep{
			{Name: "Create User Account", Status: "pending"},
			{Name: "Create System User", Status: "pending"},
			{Name: "Setup Home Directory", Status: "pending"},
			{Name: "Initialize File System", Status: "pending"},
			{Name: "Allocate Resources", Status: "pending"},
			{Name: "Configure Web Server", Status: "pending"},
		},
	}

	if req.Domain != "" {
		status.Steps = append(status.Steps, ProvisioningStep{Name: "Create Domain", Status: "pending"})
		status.Steps = append(status.Steps, ProvisioningStep{Name: "Setup DNS Zone", Status: "pending"})
	}
	if req.CreateDatabase {
		status.Steps = append(status.Steps, ProvisioningStep{Name: "Create Database", Status: "pending"})
	}
	if req.EnableSSL && req.Domain != "" {
		status.Steps = append(status.Steps, ProvisioningStep{Name: "Setup SSL Certificate", Status: "pending"})
	}
	if req.SetupBackup {
		status.Steps = append(status.Steps, ProvisioningStep{Name: "Configure Backup Schedule", Status: "pending"})
	}

	s.mu.Lock()
	s.provisionings[provisioningID] = status
	s.mu.Unlock()

	result := &AccountProvisionResult{Status: status}

	// Step 1: Create user account
	if err := s.executeStep(status, 0, func() error {
		user, err := s.userService.Create(&models.UserCreateRequest{
			Username: req.Username,
			Email:    req.Email,
			Password: req.Password,
			Role:     req.Role,
		})
		if err != nil {
			return err
		}
		result.User = user
		status.UserID = user.ID
		result.HomeDir = user.HomeDirectory
		result.SystemUID = user.UID
		result.SystemGID = user.GID
		return nil
	}); err != nil {
		s.failProvisioning(status, err)
		return result, err
	}

	// Step 2: Create system user
	if err := s.executeStep(status, 1, func() error {
		return s.oscontrolService.CreateSystemUser(result.User.Username, result.User.UID, result.User.GID)
	}); err != nil {
		s.rollback(status, result)
		return result, err
	}

	// Step 3: Setup home directory
	if err := s.executeStep(status, 2, func() error {
		return s.oscontrolService.CreateHomeDirectory(result.User.HomeDirectory, result.User.UID, result.User.GID)
	}); err != nil {
		s.rollback(status, result)
		return result, err
	}

	// Step 4: Initialize file system
	if err := s.executeStep(status, 3, func() error {
		return s.filesystemService.InitializeUserFileSystem(result.User.ID, result.User.HomeDirectory)
	}); err != nil {
		s.rollback(status, result)
		return result, err
	}

	// Step 5: Allocate resources
	if err := s.executeStep(status, 4, func() error {
		if req.Package != nil {
			return s.resourceService.AllocatePackage(result.User.ID, req.Package)
		}
		return s.resourceService.AllocateDefault(result.User.ID)
	}); err != nil {
		s.rollback(status, result)
		return result, err
	}

	// Step 6: Configure web server
	if err := s.executeStep(status, 5, func() error {
		phpVersion := req.PHPVersion
		if phpVersion == "" {
			phpVersion = "8.2"
		}
		return s.webserverService.ConfigureUserWebServer(result.User.ID, result.User.Username, phpVersion)
	}); err != nil {
		s.rollback(status, result)
		return result, err
	}

	currentStep := 6

	// Optional steps
	if req.Domain != "" {
		if err := s.executeStep(status, currentStep, func() error {
			domain, err := s.domainService.Create(result.User.ID, &models.DomainCreateRequest{
				Name:         req.Domain,
				Type:         models.DomainTypePrimary,
				DocumentRoot: result.User.HomeDirectory + "/public_html",
			})
			if err != nil {
				return err
			}
			result.Domain = domain
			return nil
		}); err != nil {
			s.rollback(status, result)
			return result, err
		}
		currentStep++

		_ = s.executeStep(status, currentStep, func() error {
			_, err := s.dnsService.CreateZone(result.Domain.ID, result.Domain.Name)
			return err
		})
		currentStep++
	}

	if req.CreateDatabase {
		_ = s.executeStep(status, currentStep, func() error {
			dbType := req.DatabaseType
			if dbType == "" {
				dbType = "mysql"
			}
			db, err := s.databaseService.Create(result.User.ID, &models.DatabaseCreateRequest{
				Name:    result.User.Username + "_db",
				Type:    models.DatabaseType(dbType),
				Charset: "utf8mb4",
			})
			if err != nil {
				return err
			}
			result.Database = db
			return nil
		})
		currentStep++
	}

	if req.EnableSSL && req.Domain != "" {
		_ = s.executeStep(status, currentStep, func() error {
			cert, err := s.sslService.RequestLetsEncrypt(result.User.ID, nil)
			if err != nil {
				return err
			}
			result.SSLCert = cert
			return nil
		})
		currentStep++
	}

	if req.SetupBackup {
		_ = s.executeStep(status, currentStep, func() error {
			// Mock - just simulate backup setup
			return nil
		})
	}

	status.Status = "completed"
	status.Progress = 100
	now := time.Now()
	status.CompletedAt = &now

	return result, nil
}

func (s *Service) executeStep(status *ProvisioningStatus, stepIndex int, fn func() error) error {
	now := time.Now()
	status.Steps[stepIndex].Status = "running"
	status.Steps[stepIndex].StartedAt = &now

	err := fn()

	completedAt := time.Now()
	status.Steps[stepIndex].CompletedAt = &completedAt

	if err != nil {
		status.Steps[stepIndex].Status = "failed"
		status.Steps[stepIndex].Error = err.Error()
		return err
	}

	status.Steps[stepIndex].Status = "completed"
	status.Progress = ((stepIndex + 1) * 100) / len(status.Steps)
	return nil
}

func (s *Service) failProvisioning(status *ProvisioningStatus, err error) {
	status.Status = "failed"
	status.Error = err.Error()
	now := time.Now()
	status.CompletedAt = &now
}

func (s *Service) rollback(status *ProvisioningStatus, result *AccountProvisionResult) {
	status.Status = "rolling_back"
	if result.User != nil {
		_ = s.oscontrolService.DeleteSystemUser(result.User.Username)
		_ = s.oscontrolService.DeleteHomeDirectory(result.User.HomeDirectory)
		_ = s.userService.Delete(result.User.ID)
	}
	status.Status = "rolled_back"
	now := time.Now()
	status.CompletedAt = &now
}

func (s *Service) GetProvisioningStatus(provisioningID string) (*ProvisioningStatus, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	status, exists := s.provisionings[provisioningID]
	if !exists {
		return nil, errors.New("provisioning not found")
	}
	return status, nil
}

func (s *Service) DeprovisionAccount(userID string) error {
	user, err := s.userService.Get(userID)
	if err != nil {
		return err
	}

	steps := []struct {
		name string
		fn   func() error
	}{
		{"Delete backups", func() error { return s.backupService.DeleteAllByUser(userID) }},
		{"Delete databases", func() error { return s.databaseService.DeleteAllByUser(userID) }},
		{"Delete domains", func() error {
			// Mock implementation - domain service doesn't have DeleteAllByUser yet
			return nil
		}},
		{"Delete SSL certificates", func() error { return s.sslService.DeleteAllByUser(userID) }},
		{"Delete home directory", func() error { return s.oscontrolService.DeleteHomeDirectory(user.HomeDirectory) }},
		{"Delete system user", func() error { return s.oscontrolService.DeleteSystemUser(user.Username) }},
		{"Release resources", func() error { return s.resourceService.ReleaseAll(userID) }},
		{"Delete user account", func() error { return s.userService.Delete(userID) }},
	}

	var lastError error
	for _, step := range steps {
		if err := step.fn(); err != nil {
			lastError = fmt.Errorf("%s failed: %w", step.name, err)
		}
	}
	return lastError
}
