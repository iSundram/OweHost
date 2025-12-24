package v1

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/iSundram/OweHost/internal/backup"
	"github.com/iSundram/OweHost/internal/user"
	"github.com/iSundram/OweHost/pkg/models"
)

type BackupHandler struct {
	backupService *backup.Service
	userService   *user.Service
}

func NewBackupHandler(backupService *backup.Service, userService *user.Service) *BackupHandler {
	return &BackupHandler{
		backupService: backupService,
		userService:   userService,
	}
}

// ListBackups lists all backups for the authenticated user
func (h *BackupHandler) ListBackups(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)

	backups := h.backupService.ListByUser(userID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(backups)
}

// GetBackup retrieves a specific backup
func (h *BackupHandler) GetBackup(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)
	parts := strings.Split(strings.TrimSuffix(r.URL.Path, "/"), "/")
	if len(parts) < 5 {
		http.Error(w, "Invalid backup ID", http.StatusBadRequest)
		return
	}
	backupID := parts[4]

	backup, err := h.backupService.Get(backupID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// Authorization check
	if backup.UserID != userID {
		role := r.Context().Value("role").(string)
		if role != "admin" {
			http.Error(w, "Unauthorized", http.StatusForbidden)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(backup)
}

// CreateBackup creates a new backup
func (h *BackupHandler) CreateBackup(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)

	var req models.BackupCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	backup, err := h.backupService.Create(userID, &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(backup)
}

// DeleteBackup deletes a backup
func (h *BackupHandler) DeleteBackup(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)
	parts := strings.Split(strings.TrimSuffix(r.URL.Path, "/"), "/")
	if len(parts) < 5 {
		http.Error(w, "Invalid backup ID", http.StatusBadRequest)
		return
	}
	backupID := parts[4]

	backup, err := h.backupService.Get(backupID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// Authorization check
	if backup.UserID != userID {
		role := r.Context().Value("role").(string)
		if role != "admin" {
			http.Error(w, "Unauthorized", http.StatusForbidden)
			return
		}
	}

	if err := h.backupService.Delete(backupID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// RestoreBackup restores a backup
func (h *BackupHandler) RestoreBackup(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)
	parts := strings.Split(strings.TrimSuffix(r.URL.Path, "/"), "/")
	if len(parts) < 6 {
		http.Error(w, "Invalid backup ID", http.StatusBadRequest)
		return
	}
	backupID := parts[4]

	backup, err := h.backupService.Get(backupID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// Authorization check
	if backup.UserID != userID {
		role := r.Context().Value("role").(string)
		if role != "admin" {
			http.Error(w, "Unauthorized", http.StatusForbidden)
			return
		}
	}

	type RestoreReq struct {
		TargetPath string `json:"target_path"`
	}
	var restoreReq RestoreReq
	json.NewDecoder(r.Body).Decode(&restoreReq)

	restoreStatus, err := h.backupService.Restore(userID, &models.RestoreRequest{
		BackupID: backupID,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(restoreStatus)
}

// DownloadBackup provides a download link for a backup
func (h *BackupHandler) DownloadBackup(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)
	parts := strings.Split(strings.TrimSuffix(r.URL.Path, "/"), "/")
	if len(parts) < 6 {
		http.Error(w, "Invalid backup ID", http.StatusBadRequest)
		return
	}
	backupID := parts[4]

	backup, err := h.backupService.Get(backupID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// Authorization check
	if backup.UserID != userID {
		role := r.Context().Value("role").(string)
		if role != "admin" {
			http.Error(w, "Unauthorized", http.StatusForbidden)
			return
		}
	}

	// Return the storage path as download URL placeholder
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"download_url": backup.StoragePath,
		"expires_at":   "3600",
	})
}

// GetBackupSchedule retrieves the backup schedule for a user
func (h *BackupHandler) GetBackupSchedule(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)

	schedules := h.backupService.ListSchedules(userID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(schedules)
}

// UpdateBackupSchedule updates the backup schedule
func (h *BackupHandler) UpdateBackupSchedule(w http.ResponseWriter, r *http.Request) {
	type UpdateReq struct {
		ScheduleID string `json:"schedule_id"`
		Enabled    bool   `json:"enabled"`
	}
	var req UpdateReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err := h.backupService.UpdateSchedule(req.ScheduleID, req.Enabled)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Schedule updated"})
}

// GetRestoreStatus retrieves the status of a restore operation
func (h *BackupHandler) GetRestoreStatus(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)
	parts := strings.Split(strings.TrimSuffix(r.URL.Path, "/"), "/")
	if len(parts) < 6 {
		http.Error(w, "Invalid restore ID", http.StatusBadRequest)
		return
	}
	restoreID := parts[5]

	status, err := h.backupService.GetRestoreStatus(restoreID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// Authorization check - verify backup belongs to user
	backup, err := h.backupService.Get(status.BackupID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	if backup.UserID != userID {
		role := r.Context().Value("role").(string)
		if role != "admin" {
			http.Error(w, "Unauthorized", http.StatusForbidden)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}
