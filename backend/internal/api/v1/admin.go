package v1

import (
	"net/http"

	"github.com/iSundram/OweHost/internal/database"
	"github.com/iSundram/OweHost/internal/domain"
	"github.com/iSundram/OweHost/internal/oscontrol"
	"github.com/iSundram/OweHost/internal/reseller"
	"github.com/iSundram/OweHost/internal/user"
	"github.com/iSundram/OweHost/pkg/utils"
)

// AdminHandler handles admin-specific endpoints
type AdminHandler struct {
	userService     *user.Service
	resellerService *reseller.Service
	domainService   *domain.Service
	databaseService *database.Service
	osService       *oscontrol.Service
}

// NewAdminHandler creates a new admin handler
func NewAdminHandler(
	userSvc *user.Service,
	resellerSvc *reseller.Service,
	domainSvc *domain.Service,
	dbSvc *database.Service,
	osSvc *oscontrol.Service,
) *AdminHandler {
	return &AdminHandler{
		userService:     userSvc,
		resellerService: resellerSvc,
		domainService:   domainSvc,
		databaseService: dbSvc,
		osService:       osSvc,
	}
}

// GetSystemStats returns aggregated system statistics
func (h *AdminHandler) GetSystemStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	// Get system metrics
	metrics, err := h.osService.GetSystemMetrics()
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, utils.ErrCodeInternalError, "Failed to get system metrics")
		return
	}

	// Get user stats
	users := h.userService.List(nil)
	activeUsers := 0
	suspendedUsers := 0
	adminCount := 0
	resellerCount := 0
	userCount := 0

	for _, u := range users {
		if u.Status == "active" {
			activeUsers++
		}
		if u.Status == "suspended" {
			suspendedUsers++
		}
		switch u.Role {
		case "admin":
			adminCount++
		case "reseller":
			resellerCount++
		case "user":
			userCount++
		}
	}

	// Get reseller stats
	resellers := h.resellerService.List()

	// Get domain stats
	domains := h.domainService.ListAll()
	activeDomains := 0
	for _, d := range domains {
		if d.Status == "active" {
			activeDomains++
		}
	}

	// Get database stats
	databases := h.databaseService.ListAll()
	var totalDbSize int64
	mysqlCount := 0
	postgresCount := 0
	for _, db := range databases {
		totalDbSize += db.SizeMB
		if db.Type == "mysql" {
			mysqlCount++
		} else if db.Type == "postgresql" {
			postgresCount++
		}
	}

	stats := map[string]interface{}{
		"users": map[string]interface{}{
			"total":     len(users),
			"active":    activeUsers,
			"suspended": suspendedUsers,
			"byRole": map[string]int{
				"admin":    adminCount,
				"reseller": resellerCount,
				"user":     userCount,
			},
		},
		"resellers": map[string]interface{}{
			"total":  len(resellers),
			"active": len(resellers),
		},
		"domains": map[string]interface{}{
			"total":   len(domains),
			"active":  activeDomains,
			"withSSL": 0, // TODO: Track SSL certificates separately
		},
		"databases": map[string]interface{}{
			"total":       len(databases),
			"totalSizeMB": totalDbSize,
			"byType": map[string]int{
				"mysql":      mysqlCount,
				"postgresql": postgresCount,
			},
		},
		"resources": map[string]interface{}{
			"cpu": map[string]interface{}{
				"usage":     metrics.CPU.Usage,
				"cores":     metrics.CPU.Cores,
				"usedCores": metrics.CPU.UsedCores,
			},
			"memory": map[string]interface{}{
				"usage": metrics.Memory.Usage,
				"total": metrics.Memory.Total,
				"used":  metrics.Memory.Used,
			},
			"disk": map[string]interface{}{
				"usage": metrics.Disk.Usage,
				"total": metrics.Disk.Total,
				"used":  metrics.Disk.Used,
			},
			"network": map[string]interface{}{
				"usage":     metrics.Network.Usage,
				"bandwidth": metrics.Network.Bandwidth,
			},
		},
	}

	utils.WriteSuccess(w, stats)
}

// GetServiceStatus returns status of core services/daemons
func (h *AdminHandler) GetServiceStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	statuses := h.osService.GetServiceStatuses()
	utils.WriteSuccess(w, statuses)
}
