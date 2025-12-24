package v1

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/iSundram/OweHost/internal/api/middleware"
	"github.com/iSundram/OweHost/internal/database"
	"github.com/iSundram/OweHost/internal/user"
	"github.com/iSundram/OweHost/pkg/models"
	"github.com/iSundram/OweHost/pkg/utils"
)

// DatabaseHandler handles database endpoints
type DatabaseHandler struct {
	databaseService *database.Service
	userService     *user.Service
}

// NewDatabaseHandler creates a new database handler
func NewDatabaseHandler(dbSvc *database.Service, userSvc *user.Service) *DatabaseHandler {
	return &DatabaseHandler{
		databaseService: dbSvc,
		userService:     userSvc,
	}
}

// Create handles database creation
func (h *DatabaseHandler) Create(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	userID := middleware.GetUserID(r.Context())

	var req models.DatabaseCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Invalid request body")
		return
	}

	db, err := h.databaseService.Create(userID, &req)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, err.Error())
		return
	}

	utils.WriteCreated(w, db)
}

// Get handles getting a database
func (h *DatabaseHandler) Get(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Database ID required")
		return
	}
	dbID := parts[len(parts)-1]

	db, err := h.databaseService.Get(dbID)
	if err != nil {
		utils.WriteError(w, http.StatusNotFound, utils.ErrCodeNotFound, err.Error())
		return
	}

	utils.WriteSuccess(w, db)
}

// List handles listing databases
func (h *DatabaseHandler) List(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	userID := middleware.GetUserID(r.Context())
	currentUser, err := h.userService.Get(userID)
	if err != nil {
		utils.WriteError(w, http.StatusUnauthorized, utils.ErrCodeUnauthorized, "User not found")
		return
	}

	var databases []*models.Database

	// Admin sees all databases
	if currentUser.Role == models.UserRoleAdmin {
		databases = h.databaseService.ListAll()
	} else {
		// Regular users and resellers see their own databases
		databases = h.databaseService.ListByUser(userID)
	}

	utils.WriteSuccess(w, databases)
}

// Delete handles deleting a database
func (h *DatabaseHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Database ID required")
		return
	}
	dbID := parts[len(parts)-1]

	if err := h.databaseService.Delete(dbID); err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// CreateUser handles database user creation
func (h *DatabaseHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Database ID required")
		return
	}
	dbID := parts[len(parts)-2]

	var req models.DatabaseUserCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Invalid request body")
		return
	}

	dbUser, err := h.databaseService.CreateUser(dbID, &req)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, err.Error())
		return
	}

	utils.WriteCreated(w, dbUser)
}

// ListUsers handles listing database users
func (h *DatabaseHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Database ID required")
		return
	}
	dbID := parts[len(parts)-2]

	users := h.databaseService.ListUsers(dbID)
	utils.WriteSuccess(w, users)
}

// DeleteUser handles deleting a database user
func (h *DatabaseHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "User ID required")
		return
	}
	userID := parts[len(parts)-1]

	if err := h.databaseService.DeleteUser(userID); err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// CreateBackup handles database backup creation
func (h *DatabaseHandler) CreateBackup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Database ID required")
		return
	}
	dbID := parts[len(parts)-2]

	backup, err := h.databaseService.CreateBackup(dbID)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, err.Error())
		return
	}

	utils.WriteCreated(w, backup)
}

// ListBackups handles listing database backups
func (h *DatabaseHandler) ListBackups(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Database ID required")
		return
	}
	dbID := parts[len(parts)-2]

	backups := h.databaseService.ListBackups(dbID)
	utils.WriteSuccess(w, backups)
}

// RestoreBackup handles database restore
func (h *DatabaseHandler) RestoreBackup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Backup ID required")
		return
	}
	backupID := parts[len(parts)-2]

	if err := h.databaseService.RestoreBackup(backupID); err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, err.Error())
		return
	}

	utils.WriteSuccess(w, map[string]string{"message": "Restore initiated"})
}
