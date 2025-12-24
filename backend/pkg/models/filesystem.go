package models

import "time"

// FileInfo represents information about a file
type FileInfo struct {
	Name        string    `json:"name"`
	Path        string    `json:"path"`
	Size        int64     `json:"size"`
	IsDirectory bool      `json:"is_directory"`
	Permissions string    `json:"permissions"`
	Owner       string    `json:"owner"`
	Group       string    `json:"group"`
	ModifiedAt  time.Time `json:"modified_at"`
}

// FileContent represents file content
type FileContent struct {
	Path    string `json:"path"`
	Content string `json:"content"`
	Encoding string `json:"encoding"`
}

// ArchiveRequest represents a request to create an archive
type ArchiveRequest struct {
	SourcePaths []string `json:"source_paths" validate:"required,min=1"`
	DestPath    string   `json:"dest_path" validate:"required"`
	Format      string   `json:"format" validate:"required,oneof=zip tar.gz tar"`
}

// ExtractRequest represents a request to extract an archive
type ExtractRequest struct {
	SourcePath string `json:"source_path" validate:"required"`
	DestPath   string `json:"dest_path" validate:"required"`
}

// StorageMount represents an external storage mount
type StorageMount struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	MountPoint  string    `json:"mount_point"`
	SourcePath  string    `json:"source_path"`
	FSType      string    `json:"fs_type"`
	Options     []string  `json:"options"`
	ReadOnly    bool      `json:"read_only"`
	CreatedAt   time.Time `json:"created_at"`
}

// StorageMountCreateRequest represents a request to create a storage mount
type StorageMountCreateRequest struct {
	MountPoint string   `json:"mount_point" validate:"required"`
	SourcePath string   `json:"source_path" validate:"required"`
	FSType     string   `json:"fs_type" validate:"required"`
	Options    []string `json:"options,omitempty"`
	ReadOnly   bool     `json:"read_only"`
}

// FileOperationRequest represents a file operation request
type FileOperationRequest struct {
	Operation string `json:"operation" validate:"required,oneof=copy move delete rename chmod chown"`
	Source    string `json:"source" validate:"required"`
	Target    string `json:"target,omitempty"`
	Mode      string `json:"mode,omitempty"`
	Owner     string `json:"owner,omitempty"`
	Group     string `json:"group,omitempty"`
}
