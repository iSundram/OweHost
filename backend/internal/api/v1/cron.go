package v1

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/iSundram/OweHost/internal/cron"
	"github.com/iSundram/OweHost/internal/user"
	"github.com/iSundram/OweHost/pkg/models"
)

type CronHandler struct {
	cronService *cron.Service
	userService *user.Service
}

func NewCronHandler(cronService *cron.Service, userService *user.Service) *CronHandler {
	return &CronHandler{
		cronService: cronService,
		userService: userService,
	}
}

// ListCronJobs lists all cron jobs for the authenticated user
func (h *CronHandler) ListCronJobs(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)

	jobs := h.cronService.ListByUser(userID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(jobs)
}

// GetCronJob retrieves a specific cron job
func (h *CronHandler) GetCronJob(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)
	parts := strings.Split(strings.TrimSuffix(r.URL.Path, "/"), "/")
	if len(parts) < 6 {
		http.Error(w, "Invalid cron job ID", http.StatusBadRequest)
		return
	}
	jobID := parts[5]

	job, err := h.cronService.Get(jobID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// Authorization check
	if job.UserID != userID {
		role := r.Context().Value("role").(string)
		if role != "admin" {
			http.Error(w, "Unauthorized", http.StatusForbidden)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(job)
}

// CreateCronJob creates a new cron job
func (h *CronHandler) CreateCronJob(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)

	var req models.CronJobCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	job, err := h.cronService.Create(userID, &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(job)
}

// UpdateCronJob updates a cron job
func (h *CronHandler) UpdateCronJob(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)
	parts := strings.Split(strings.TrimSuffix(r.URL.Path, "/"), "/")
	if len(parts) < 6 {
		http.Error(w, "Invalid cron job ID", http.StatusBadRequest)
		return
	}
	jobID := parts[5]

	job, err := h.cronService.Get(jobID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// Authorization check
	if job.UserID != userID {
		role := r.Context().Value("role").(string)
		if role != "admin" {
			http.Error(w, "Unauthorized", http.StatusForbidden)
			return
		}
	}

	var req models.CronJobUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	updated, err := h.cronService.Update(jobID, &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updated)
}

// DeleteCronJob deletes a cron job
func (h *CronHandler) DeleteCronJob(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)
	parts := strings.Split(strings.TrimSuffix(r.URL.Path, "/"), "/")
	if len(parts) < 6 {
		http.Error(w, "Invalid cron job ID", http.StatusBadRequest)
		return
	}
	jobID := parts[5]

	job, err := h.cronService.Get(jobID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// Authorization check
	if job.UserID != userID {
		role := r.Context().Value("role").(string)
		if role != "admin" {
			http.Error(w, "Unauthorized", http.StatusForbidden)
			return
		}
	}

	if err := h.cronService.Delete(jobID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// EnableCronJob enables a cron job
func (h *CronHandler) EnableCronJob(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)
	parts := strings.Split(strings.TrimSuffix(r.URL.Path, "/"), "/")
	if len(parts) < 7 {
		http.Error(w, "Invalid cron job ID", http.StatusBadRequest)
		return
	}
	jobID := parts[5]

	job, err := h.cronService.Get(jobID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// Authorization check
	if job.UserID != userID {
		role := r.Context().Value("role").(string)
		if role != "admin" {
			http.Error(w, "Unauthorized", http.StatusForbidden)
			return
		}
	}

	if err := h.cronService.Resume(jobID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "enabled"})
}

// DisableCronJob disables a cron job
func (h *CronHandler) DisableCronJob(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)
	parts := strings.Split(strings.TrimSuffix(r.URL.Path, "/"), "/")
	if len(parts) < 7 {
		http.Error(w, "Invalid cron job ID", http.StatusBadRequest)
		return
	}
	jobID := parts[5]

	job, err := h.cronService.Get(jobID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// Authorization check
	if job.UserID != userID {
		role := r.Context().Value("role").(string)
		if role != "admin" {
			http.Error(w, "Unauthorized", http.StatusForbidden)
			return
		}
	}

	if err := h.cronService.Pause(jobID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "disabled"})
}

// GetCronJobExecutions retrieves execution history for a cron job
func (h *CronHandler) GetCronJobExecutions(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)
	parts := strings.Split(strings.TrimSuffix(r.URL.Path, "/"), "/")
	if len(parts) < 7 {
		http.Error(w, "Invalid cron job ID", http.StatusBadRequest)
		return
	}
	jobID := parts[5]

	job, err := h.cronService.Get(jobID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// Authorization check
	if job.UserID != userID {
		role := r.Context().Value("role").(string)
		if role != "admin" {
			http.Error(w, "Unauthorized", http.StatusForbidden)
			return
		}
	}

	executions := h.cronService.GetExecutions(jobID, 50)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(executions)
}

// ValidateCronExpression validates a cron expression
func (h *CronHandler) ValidateCronExpression(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Expression string `json:"expression"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Basic validation - check if expression has 5 parts
	parts := strings.Fields(req.Expression)
	isValid := len(parts) == 5 || len(parts) == 6
	
	response := map[string]interface{}{
		"valid":      isValid,
		"expression": req.Expression,
	}
	
	if !isValid {
		response["error"] = "Cron expression must have 5 or 6 fields"
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
