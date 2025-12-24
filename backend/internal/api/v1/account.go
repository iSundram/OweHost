package v1

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/iSundram/OweHost/internal/dns"
	"github.com/iSundram/OweHost/internal/domain"
	"github.com/iSundram/OweHost/internal/user"
	"github.com/iSundram/OweHost/pkg/models"
	"github.com/iSundram/OweHost/pkg/utils"
)

// AccountHandler groups "Account Functions" style operations.
// It builds on top of user, domain, and dns services.
type AccountHandler struct {
	userService   *user.Service
	domainService *domain.Service
	dnsService    *dns.Service
}

// NewAccountHandler creates a new AccountHandler.
func NewAccountHandler(userSvc *user.Service, domainSvc *domain.Service, dnsSvc *dns.Service) *AccountHandler {
	return &AccountHandler{
		userService:   userSvc,
		domainService: domainSvc,
		dnsService:    dnsSvc,
	}
}

type accountCreateRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"`
	Domain   string `json:"domain,omitempty"`
}

// Create creates a new account (user), and optionally a primary domain + DNS zone.
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

	role := models.UserRole(req.Role)
	if role == "" {
		role = models.UserRoleUser
	}

	user, err := h.userService.Create(&models.UserCreateRequest{
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password,
		Role:     role,
	})
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, err.Error())
		return
	}

	var createdDomain *models.Domain
	if req.Domain != "" {
		domain, err := h.domainService.Create(user.ID, &models.DomainCreateRequest{
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
		"user":   user,
		"domain": createdDomain,
	})
}

// List returns all accounts.
func (h *AccountHandler) List(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}
	users := h.userService.List(nil)
	utils.WriteSuccess(w, users)
}

// UpdateStatus updates status (suspend/unsuspend/terminate).
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
	accountID := parts[len(parts)-2]
	action := parts[len(parts)-1]

	var err error
	switch action {
	case "suspend":
		err = h.userService.Suspend(accountID)
	case "unsuspend":
		status := models.UserStatusActive
		_, err = h.userService.Update(accountID, &models.UserUpdateRequest{Status: &status})
	case "terminate":
		err = h.userService.Terminate(accountID)
	default:
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Unsupported action")
		return
	}

	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, err.Error())
		return
	}

	user, _ := h.userService.Get(accountID)
	utils.WriteSuccess(w, user)
}
