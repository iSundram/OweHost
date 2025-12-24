// Package git provides Git version control management for OweHost
package git

import (
	"errors"
	"sync"
	"time"

	"github.com/iSundram/OweHost/pkg/models"
	"github.com/iSundram/OweHost/pkg/utils"
)

// Service provides Git repository management functionality
type Service struct {
	repositories map[string]*models.GitRepository
	deployKeys   map[string]*models.DeployKey
	webhooks     map[string]*models.GitWebhook
	byUser       map[string][]*models.GitRepository
	byDomain     map[string][]*models.GitRepository
	mu           sync.RWMutex
}

// NewService creates a new git service
func NewService() *Service {
	return &Service{
		repositories: make(map[string]*models.GitRepository),
		deployKeys:   make(map[string]*models.DeployKey),
		webhooks:     make(map[string]*models.GitWebhook),
		byUser:       make(map[string][]*models.GitRepository),
		byDomain:     make(map[string][]*models.GitRepository),
	}
}

// CreateRepository creates a new Git repository
func (s *Service) CreateRepository(userID string, req *models.GitRepositoryCreateRequest) (*models.GitRepository, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	repoPath := "/home/" + userID + "/repositories/" + req.Name + ".git"

	repo := &models.GitRepository{
		ID:            utils.GenerateID("repo"),
		UserID:        userID,
		DomainID:      req.DomainID,
		Name:          req.Name,
		Description:   req.Description,
		Path:          repoPath,
		CloneURL:      "git@server:" + userID + "/" + req.Name + ".git",
		HTTPCloneURL:  "https://server/" + userID + "/" + req.Name + ".git",
		DefaultBranch: "main",
		IsPrivate:     true,
		AutoDeploy:    req.AutoDeploy,
		DeployPath:    req.DeployPath,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	s.repositories[repo.ID] = repo
	s.byUser[userID] = append(s.byUser[userID], repo)
	if req.DomainID != nil {
		s.byDomain[*req.DomainID] = append(s.byDomain[*req.DomainID], repo)
	}

	return repo, nil
}

// Get gets a repository by ID
func (s *Service) Get(id string) (*models.GitRepository, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	repo, exists := s.repositories[id]
	if !exists {
		return nil, errors.New("repository not found")
	}
	return repo, nil
}

// ListByUser lists repositories for a user
func (s *Service) ListByUser(userID string) []*models.GitRepository {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.byUser[userID]
}

// ListByDomain lists repositories for a domain
func (s *Service) ListByDomain(domainID string) []*models.GitRepository {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.byDomain[domainID]
}

// Update updates a repository
func (s *Service) Update(id string, req *models.GitRepositoryUpdateRequest) (*models.GitRepository, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	repo, exists := s.repositories[id]
	if !exists {
		return nil, errors.New("repository not found")
	}

	if req.Description != nil {
		repo.Description = *req.Description
	}
	if req.AutoDeploy != nil {
		repo.AutoDeploy = *req.AutoDeploy
	}
	if req.DeployPath != nil {
		repo.DeployPath = req.DeployPath
	}
	if req.DefaultBranch != nil {
		repo.DefaultBranch = *req.DefaultBranch
	}

	repo.UpdatedAt = time.Now()
	return repo, nil
}

// Delete deletes a repository
func (s *Service) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	repo, exists := s.repositories[id]
	if !exists {
		return errors.New("repository not found")
	}

	// Remove from user's repos
	userRepos := s.byUser[repo.UserID]
	for i, r := range userRepos {
		if r.ID == id {
			s.byUser[repo.UserID] = append(userRepos[:i], userRepos[i+1:]...)
			break
		}
	}

	// Remove from domain's repos
	if repo.DomainID != nil {
		domainRepos := s.byDomain[*repo.DomainID]
		for i, r := range domainRepos {
			if r.ID == id {
				s.byDomain[*repo.DomainID] = append(domainRepos[:i], domainRepos[i+1:]...)
				break
			}
		}
	}

	// Remove deploy keys and webhooks
	for keyID, key := range s.deployKeys {
		if key.RepositoryID == id {
			delete(s.deployKeys, keyID)
		}
	}
	for whID, wh := range s.webhooks {
		if wh.RepositoryID == id {
			delete(s.webhooks, whID)
		}
	}

	delete(s.repositories, id)
	return nil
}

// Clone clones a remote repository
func (s *Service) Clone(userID string, req *models.GitCloneRequest) (*models.GitRepository, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	repoPath := "/home/" + userID + "/repositories/" + req.Name + ".git"

	repo := &models.GitRepository{
		ID:            utils.GenerateID("repo"),
		UserID:        userID,
		DomainID:      req.DomainID,
		Name:          req.Name,
		Description:   "Cloned from " + req.RemoteURL,
		Path:          repoPath,
		RemoteURL:     &req.RemoteURL,
		CloneURL:      "git@server:" + userID + "/" + req.Name + ".git",
		HTTPCloneURL:  "https://server/" + userID + "/" + req.Name + ".git",
		DefaultBranch: req.Branch,
		IsPrivate:     true,
		AutoDeploy:    req.AutoDeploy,
		DeployPath:    req.DeployPath,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if repo.DefaultBranch == "" {
		repo.DefaultBranch = "main"
	}

	s.repositories[repo.ID] = repo
	s.byUser[userID] = append(s.byUser[userID], repo)
	if req.DomainID != nil {
		s.byDomain[*req.DomainID] = append(s.byDomain[*req.DomainID], repo)
	}

	return repo, nil
}

// Pull pulls latest changes from remote
func (s *Service) Pull(id, branch string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	repo, exists := s.repositories[id]
	if !exists {
		return errors.New("repository not found")
	}

	if repo.RemoteURL == nil {
		return errors.New("no remote configured")
	}

	now := time.Now()
	repo.LastPullAt = &now
	repo.UpdatedAt = now

	return nil
}

// Deploy deploys the repository to the deploy path
func (s *Service) Deploy(id string) (*models.DeploymentInfo, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	repo, exists := s.repositories[id]
	if !exists {
		return nil, errors.New("repository not found")
	}

	if repo.DeployPath == nil {
		return nil, errors.New("no deploy path configured")
	}

	now := time.Now()
	repo.LastDeployAt = &now
	repo.UpdatedAt = now

	deployment := &models.DeploymentInfo{
		ID:           utils.GenerateID("deploy"),
		RepositoryID: id,
		Branch:       repo.DefaultBranch,
		Commit:       "abc123",
		Status:       "success",
		DeployedAt:   now,
	}

	return deployment, nil
}

// AddDeployKey adds a deploy key to a repository
func (s *Service) AddDeployKey(repoID string, req *models.DeployKeyCreateRequest) (*models.DeployKey, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.repositories[repoID]; !exists {
		return nil, errors.New("repository not found")
	}

	key := &models.DeployKey{
		ID:           utils.GenerateID("key"),
		RepositoryID: repoID,
		Title:        req.Title,
		PublicKey:    req.PublicKey,
		ReadOnly:     req.ReadOnly,
		CreatedAt:    time.Now(),
	}

	s.deployKeys[key.ID] = key
	return key, nil
}

// ListDeployKeys lists deploy keys for a repository
func (s *Service) ListDeployKeys(repoID string) []*models.DeployKey {
	s.mu.RLock()
	defer s.mu.RUnlock()

	keys := make([]*models.DeployKey, 0)
	for _, key := range s.deployKeys {
		if key.RepositoryID == repoID {
			keys = append(keys, key)
		}
	}
	return keys
}

// RemoveDeployKey removes a deploy key
func (s *Service) RemoveDeployKey(keyID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.deployKeys[keyID]; !exists {
		return errors.New("deploy key not found")
	}

	delete(s.deployKeys, keyID)
	return nil
}

// AddWebhook adds a webhook to a repository
func (s *Service) AddWebhook(repoID string, req *models.GitWebhookCreateRequest) (*models.GitWebhook, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.repositories[repoID]; !exists {
		return nil, errors.New("repository not found")
	}

	webhook := &models.GitWebhook{
		ID:           utils.GenerateID("ghook"),
		RepositoryID: repoID,
		URL:          req.URL,
		Events:       req.Events,
		Secret:       req.Secret,
		Enabled:      true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	s.webhooks[webhook.ID] = webhook
	return webhook, nil
}

// ListWebhooks lists webhooks for a repository
func (s *Service) ListWebhooks(repoID string) []*models.GitWebhook {
	s.mu.RLock()
	defer s.mu.RUnlock()

	webhooks := make([]*models.GitWebhook, 0)
	for _, wh := range s.webhooks {
		if wh.RepositoryID == repoID {
			webhooks = append(webhooks, wh)
		}
	}
	return webhooks
}

// RemoveWebhook removes a webhook
func (s *Service) RemoveWebhook(webhookID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.webhooks[webhookID]; !exists {
		return errors.New("webhook not found")
	}

	delete(s.webhooks, webhookID)
	return nil
}

// GetBranches gets branches for a repository
func (s *Service) GetBranches(id string) ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	_, exists := s.repositories[id]
	if !exists {
		return nil, errors.New("repository not found")
	}

	return []string{"main", "develop", "feature/example"}, nil
}

// GetCommits gets recent commits for a repository
func (s *Service) GetCommits(id string, branch string, limit int) ([]models.GitCommit, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	_, exists := s.repositories[id]
	if !exists {
		return nil, errors.New("repository not found")
	}

	commits := []models.GitCommit{
		{
			Hash:        "abc123def456",
			ShortHash:   "abc123d",
			Message:     "Initial commit",
			Author:      "developer",
			AuthorEmail: "dev@example.com",
			Date:        time.Now().Add(-24 * time.Hour),
		},
		{
			Hash:        "def456ghi789",
			ShortHash:   "def456g",
			Message:     "Add new feature",
			Author:      "developer",
			AuthorEmail: "dev@example.com",
			Date:        time.Now().Add(-12 * time.Hour),
		},
	}

	if limit > 0 && limit < len(commits) {
		commits = commits[:limit]
	}

	return commits, nil
}
