package v1

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/iSundram/OweHost/internal/api/middleware"
	"github.com/iSundram/OweHost/internal/git"
	"github.com/iSundram/OweHost/pkg/models"
	"github.com/iSundram/OweHost/pkg/utils"
)

// GitHandler handles Git repository endpoints
type GitHandler struct {
	gitService *git.Service
}

// NewGitHandler creates a new git handler
func NewGitHandler(gitSvc *git.Service) *GitHandler {
	return &GitHandler{
		gitService: gitSvc,
	}
}

// CreateRepository handles repository creation
func (h *GitHandler) CreateRepository(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	userID := middleware.GetUserID(r.Context())

	var req models.GitRepositoryCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Invalid request body")
		return
	}

	repo, err := h.gitService.CreateRepository(userID, &req)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, err.Error())
		return
	}

	utils.WriteCreated(w, repo)
}

// Get handles getting a repository
func (h *GitHandler) Get(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Repository ID required")
		return
	}
	repoID := parts[len(parts)-1]

	repo, err := h.gitService.Get(repoID)
	if err != nil {
		utils.WriteError(w, http.StatusNotFound, utils.ErrCodeNotFound, err.Error())
		return
	}

	utils.WriteSuccess(w, repo)
}

// List handles listing repositories
func (h *GitHandler) List(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	userID := middleware.GetUserID(r.Context())
	domainID := r.URL.Query().Get("domain_id")

	var repos []*models.GitRepository
	if domainID != "" {
		repos = h.gitService.ListByDomain(domainID)
	} else {
		repos = h.gitService.ListByUser(userID)
	}

	utils.WriteSuccess(w, repos)
}

// Update handles repository update
func (h *GitHandler) Update(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut && r.Method != http.MethodPatch {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Repository ID required")
		return
	}
	repoID := parts[len(parts)-1]

	var req models.GitRepositoryUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Invalid request body")
		return
	}

	repo, err := h.gitService.Update(repoID, &req)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, err.Error())
		return
	}

	utils.WriteSuccess(w, repo)
}

// Delete handles repository deletion
func (h *GitHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Repository ID required")
		return
	}
	repoID := parts[len(parts)-1]

	if err := h.gitService.Delete(repoID); err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Clone handles cloning a remote repository
func (h *GitHandler) Clone(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	userID := middleware.GetUserID(r.Context())

	var req models.GitCloneRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Invalid request body")
		return
	}

	repo, err := h.gitService.Clone(userID, &req)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, err.Error())
		return
	}

	utils.WriteCreated(w, repo)
}

// Pull handles pulling latest changes
func (h *GitHandler) Pull(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Repository ID required")
		return
	}
	repoID := parts[len(parts)-2]

	var req struct {
		Branch string `json:"branch"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	if err := h.gitService.Pull(repoID, req.Branch); err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, err.Error())
		return
	}

	utils.WriteSuccess(w, map[string]string{"message": "Pull successful"})
}

// Deploy handles deploying a repository
func (h *GitHandler) Deploy(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Repository ID required")
		return
	}
	repoID := parts[len(parts)-2]

	deployment, err := h.gitService.Deploy(repoID)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, err.Error())
		return
	}

	utils.WriteSuccess(w, deployment)
}

// GetBranches handles getting branches
func (h *GitHandler) GetBranches(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Repository ID required")
		return
	}
	repoID := parts[len(parts)-2]

	branches, err := h.gitService.GetBranches(repoID)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, err.Error())
		return
	}

	utils.WriteSuccess(w, branches)
}

// GetCommits handles getting commits
func (h *GitHandler) GetCommits(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Repository ID required")
		return
	}
	repoID := parts[len(parts)-2]

	branch := r.URL.Query().Get("branch")
	limitStr := r.URL.Query().Get("limit")
	limit := 10
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	commits, err := h.gitService.GetCommits(repoID, branch, limit)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, err.Error())
		return
	}

	utils.WriteSuccess(w, commits)
}

// AddDeployKey handles adding a deploy key
func (h *GitHandler) AddDeployKey(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Repository ID required")
		return
	}
	repoID := parts[len(parts)-2]

	var req models.DeployKeyCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Invalid request body")
		return
	}

	key, err := h.gitService.AddDeployKey(repoID, &req)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, err.Error())
		return
	}

	utils.WriteCreated(w, key)
}

// ListDeployKeys handles listing deploy keys
func (h *GitHandler) ListDeployKeys(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Repository ID required")
		return
	}
	repoID := parts[len(parts)-2]

	keys := h.gitService.ListDeployKeys(repoID)
	utils.WriteSuccess(w, keys)
}

// RemoveDeployKey handles removing a deploy key
func (h *GitHandler) RemoveDeployKey(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Key ID required")
		return
	}
	keyID := parts[len(parts)-1]

	if err := h.gitService.RemoveDeployKey(keyID); err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// AddWebhook handles adding a webhook
func (h *GitHandler) AddWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Repository ID required")
		return
	}
	repoID := parts[len(parts)-2]

	var req models.GitWebhookCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Invalid request body")
		return
	}

	webhook, err := h.gitService.AddWebhook(repoID, &req)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, err.Error())
		return
	}

	utils.WriteCreated(w, webhook)
}

// ListWebhooks handles listing webhooks
func (h *GitHandler) ListWebhooks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Repository ID required")
		return
	}
	repoID := parts[len(parts)-2]

	webhooks := h.gitService.ListWebhooks(repoID)
	utils.WriteSuccess(w, webhooks)
}

// RemoveWebhook handles removing a webhook
func (h *GitHandler) RemoveWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Webhook ID required")
		return
	}
	webhookID := parts[len(parts)-1]

	if err := h.gitService.RemoveWebhook(webhookID); err != nil {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
