package v1

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/iSundram/OweHost/internal/accountsvc"
	"github.com/iSundram/OweHost/internal/api/middleware"
	"github.com/iSundram/OweHost/internal/dns"
	"github.com/iSundram/OweHost/internal/domain"
	"github.com/iSundram/OweHost/internal/user"
	"github.com/iSundram/OweHost/pkg/models"
	"github.com/iSundram/OweHost/pkg/utils"
)

// AccountHandler groups "Account Functions" style operations.
// It builds on top of accountsvc, user, domain, and dns services.
type AccountHandler struct {
	accountService *accountsvc.Service
	userService    *user.Service
	domainService  *domain.Service
	dnsService     *dns.Service
}

// NewAccountHandler creates a new AccountHandler.
func NewAccountHandler(accountSvc *accountsvc.Service, userSvc *user.Service, domainSvc *domain.Service, dnsSvc *dns.Service) *AccountHandler {
	return &AccountHandler{
		accountService: accountSvc,
		userService:    userSvc,
		domainService:  domainSvc,
		dnsService:     dnsSvc,
	}
}

type accountCreateRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Plan     string `json:"plan"`     // Plan: starter, standard, premium, enterprise
	Owner    string `json:"owner"`    // Owner: admin, reseller-X, partner-X
	Domain   string `json:"domain,omitempty"`
}

// Create creates a new account using filesystem storage, and optionally a primary domain + DNS zone.
func (h *AccountHandler) Create(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	var req accountCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Invalid request body")
		return
	}

	// Validate required fields
	if req.Username == "" {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Username is required")
		return
	}
	if req.Email == "" {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Email is required")
		return
	}
	if req.Plan == "" {
		req.Plan = "starter" // Default plan
	}
	if req.Owner == "" {
		// Try to determine owner from authenticated user
		userID := middleware.GetUserID(r.Context())
		if userID != "" {
			user, err := h.userService.Get(userID)
			if err == nil {
				if user.Role == models.UserRoleAdmin {
					req.Owner = "admin"
				} else if user.Role == models.UserRoleReseller {
					req.Owner = "reseller-" + userID
				} else {
					req.Owner = "admin" // Default to admin
				}
			} else {
				req.Owner = "admin" // Default to admin
			}
		} else {
			req.Owner = "admin" // Default to admin
		}
	}

	// Get actor information from request context
	actorID := middleware.GetUserID(r.Context())
	actorRole := middleware.GetUserRole(r.Context())
	actorType := "user"
	if actorRole == models.UserRoleAdmin {
		actorType = "admin"
	} else if actorRole == models.UserRoleReseller {
		actorType = "reseller"
	}
	
	// Get actor username
	actor := "system"
	if actorID != "" {
		user, err := h.userService.Get(actorID)
		if err == nil {
			actor = user.Username
		}
	}

	// Get IP address
	actorIP := r.RemoteAddr
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		actorIP = forwarded
	}

	// Create account using accountsvc (filesystem-based)
	ctx := r.Context()
	accountResp, err := h.accountService.Create(ctx, &accountsvc.CreateRequest{
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password,
		Plan:     req.Plan,
		Owner:    req.Owner,
	}, actor, actorType, actorIP)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, err.Error())
		return
	}

	// Also create user record in user service for API compatibility
	// This maintains backward compatibility with existing user management
	userRecord, err := h.userService.Create(&models.UserCreateRequest{
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password,
		Role:     models.UserRoleUser, // Accounts created via this endpoint are users
	})
	if err != nil {
		// Log but don't fail - account is already created in filesystem
		// This is for backward compatibility only
	}

	var createdDomain *models.Domain
	if req.Domain != "" {
		// Use account ID from filesystem account
		userIDForDomain := userRecord.ID
		if userRecord == nil {
			userIDForDomain = utils.GenerateID("usr")
		}
		
		domain, err := h.domainService.Create(userIDForDomain, &models.DomainCreateRequest{
			Name: req.Domain,
			Type: models.DomainTypePrimary,
		})
		if err == nil {
			createdDomain = domain
			// auto-create zone
			_, _ = h.dnsService.CreateZone(domain.ID, domain.Name)
		}
	}

	utils.WriteCreated(w, map[string]interface{}{
		"account": accountResp,
		"user":    userRecord,
		"domain":  createdDomain,
	})
}

// List returns all accounts from filesystem storage.
func (h *AccountHandler) List(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}
	
	ctx := r.Context()
	accounts, err := h.accountService.List(ctx)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, utils.ErrCodeInternalError, err.Error())
		return
	}
	
	utils.WriteSuccess(w, accounts)
}

// UpdateStatus updates status (suspend/unsuspend/terminate) using filesystem storage.
func (h *AccountHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Account ID required")
		return
	}
	
	// Parse account ID (should be integer for filesystem accounts)
	accountIDStr := parts[len(parts)-2]
	var accountID int
	if _, err := fmt.Sscanf(accountIDStr, "%d", &accountID); err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Invalid account ID format")
		return
	}
	
	action := parts[len(parts)-1]

	// Get actor information
	actorID := middleware.GetUserID(r.Context())
	actorRole := middleware.GetUserRole(r.Context())
	actorType := "user"
	if actorRole == models.UserRoleAdmin {
		actorType = "admin"
	} else if actorRole == models.UserRoleReseller {
		actorType = "reseller"
	}
	
	actor := "system"
	if actorID != "" {
		user, err := h.userService.Get(actorID)
		if err == nil {
			actor = user.Username
		}
	}

	ctx := r.Context()
	var err error
	switch action {
	case "suspend":
		err = h.accountService.Suspend(ctx, accountID, "Suspended via API", actor, actorType)
	case "unsuspend":
		err = h.accountService.Unsuspend(ctx, accountID, actor, actorType)
	case "terminate":
		err = h.accountService.Terminate(ctx, accountID, "Terminated via API", actor, actorType)
	default:
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Unsupported action")
		return
	}

	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, err.Error())
		return
	}

	account, err := h.accountService.Get(ctx, accountID)
	if err != nil {
		utils.WriteError(w, http.StatusNotFound, utils.ErrCodeNotFound, "Account not found")
		return
	}

	utils.WriteSuccess(w, account)
}
