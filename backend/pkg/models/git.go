package models

import "time"

// GitRepository represents a Git repository
type GitRepository struct {
	ID            string     `json:"id"`
	UserID        string     `json:"user_id"`
	DomainID      *string    `json:"domain_id,omitempty"`
	Name          string     `json:"name"`
	Description   string     `json:"description"`
	Path          string     `json:"path"`
	RemoteURL     *string    `json:"remote_url,omitempty"`
	CloneURL      string     `json:"clone_url"`
	HTTPCloneURL  string     `json:"http_clone_url"`
	DefaultBranch string     `json:"default_branch"`
	IsPrivate     bool       `json:"is_private"`
	AutoDeploy    bool       `json:"auto_deploy"`
	DeployPath    *string    `json:"deploy_path,omitempty"`
	LastPullAt    *time.Time `json:"last_pull_at,omitempty"`
	LastDeployAt  *time.Time `json:"last_deploy_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// GitRepositoryCreateRequest represents a request to create a repository
type GitRepositoryCreateRequest struct {
	Name        string  `json:"name" validate:"required,alphanum"`
	Description string  `json:"description"`
	DomainID    *string `json:"domain_id"`
	AutoDeploy  bool    `json:"auto_deploy"`
	DeployPath  *string `json:"deploy_path"`
}

// GitRepositoryUpdateRequest represents a request to update a repository
type GitRepositoryUpdateRequest struct {
	Description   *string `json:"description"`
	AutoDeploy    *bool   `json:"auto_deploy"`
	DeployPath    *string `json:"deploy_path"`
	DefaultBranch *string `json:"default_branch"`
}

// GitCloneRequest represents a request to clone a remote repository
type GitCloneRequest struct {
	Name       string  `json:"name" validate:"required"`
	RemoteURL  string  `json:"remote_url" validate:"required,url"`
	Branch     string  `json:"branch"`
	DomainID   *string `json:"domain_id"`
	AutoDeploy bool    `json:"auto_deploy"`
	DeployPath *string `json:"deploy_path"`
}

// DeployKey represents a deploy key for a repository
type DeployKey struct {
	ID           string    `json:"id"`
	RepositoryID string    `json:"repository_id"`
	Title        string    `json:"title"`
	PublicKey    string    `json:"public_key"`
	Fingerprint  string    `json:"fingerprint"`
	ReadOnly     bool      `json:"read_only"`
	CreatedAt    time.Time `json:"created_at"`
}

// DeployKeyCreateRequest represents a request to create a deploy key
type DeployKeyCreateRequest struct {
	Title     string `json:"title" validate:"required"`
	PublicKey string `json:"public_key" validate:"required"`
	ReadOnly  bool   `json:"read_only"`
}

// GitWebhook represents a webhook for a repository
type GitWebhook struct {
	ID           string    `json:"id"`
	RepositoryID string    `json:"repository_id"`
	URL          string    `json:"url"`
	Events       []string  `json:"events"`
	Secret       string    `json:"-"`
	Enabled      bool      `json:"enabled"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// GitWebhookCreateRequest represents a request to create a webhook
type GitWebhookCreateRequest struct {
	URL    string   `json:"url" validate:"required,url"`
	Events []string `json:"events" validate:"required,min=1"`
	Secret string   `json:"secret"`
}

// DeploymentInfo represents information about a deployment
type DeploymentInfo struct {
	ID           string    `json:"id"`
	RepositoryID string    `json:"repository_id"`
	Branch       string    `json:"branch"`
	Commit       string    `json:"commit"`
	Status       string    `json:"status"`
	Message      string    `json:"message,omitempty"`
	DeployedAt   time.Time `json:"deployed_at"`
}

// GitCommit represents a Git commit
type GitCommit struct {
	Hash        string    `json:"hash"`
	ShortHash   string    `json:"short_hash"`
	Message     string    `json:"message"`
	Author      string    `json:"author"`
	AuthorEmail string    `json:"author_email"`
	Date        time.Time `json:"date"`
}
