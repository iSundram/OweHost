// Package filesystem provides file system management for OweHost
package filesystem

import (
	"errors"
	"sync"
	"time"

	"github.com/iSundram/OweHost/pkg/models"
	"github.com/iSundram/OweHost/pkg/utils"
)

// Service provides file system functionality
type Service struct {
	files       map[string]*models.FileInfo
	mounts      map[string]*models.StorageMount
	mountsByUser map[string][]*models.StorageMount
	mu          sync.RWMutex
}

// NewService creates a new filesystem service
func NewService() *Service {
	return &Service{
		files:        make(map[string]*models.FileInfo),
		mounts:       make(map[string]*models.StorageMount),
		mountsByUser: make(map[string][]*models.StorageMount),
	}
}

// Read reads file info
func (s *Service) Read(userID, path string) (*models.FileInfo, error) {
	if !s.checkAccess(userID, path, "read") {
		return nil, errors.New("access denied")
	}

	// Simulate file info (in production, would use os.Stat)
	info := &models.FileInfo{
		Name:        utils.SanitizePath(path),
		Path:        path,
		Size:        0,
		IsDirectory: false,
		Permissions: "0644",
		Owner:       userID,
		Group:       userID,
		ModifiedAt:  time.Now(),
	}

	return info, nil
}

// ReadContent reads file content
func (s *Service) ReadContent(userID, path string) (*models.FileContent, error) {
	if !s.checkAccess(userID, path, "read") {
		return nil, errors.New("access denied")
	}

	// In production, would use os.ReadFile
	content := &models.FileContent{
		Path:     path,
		Content:  "",
		Encoding: "utf-8",
	}

	return content, nil
}

// Write writes file content
func (s *Service) Write(userID, path, content string) error {
	if !s.checkAccess(userID, path, "write") {
		return errors.New("access denied")
	}

	// In production, would use os.WriteFile
	return nil
}

// ListDirectory lists directory contents
func (s *Service) ListDirectory(userID, path string) ([]*models.FileInfo, error) {
	if !s.checkAccess(userID, path, "read") {
		return nil, errors.New("access denied")
	}

	// In production, would use os.ReadDir
	files := make([]*models.FileInfo, 0)
	return files, nil
}

// CreateDirectory creates a directory
func (s *Service) CreateDirectory(userID, path string) error {
	if !s.checkAccess(userID, path, "write") {
		return errors.New("access denied")
	}

	// In production, would use os.MkdirAll
	return nil
}

// Delete deletes a file or directory
func (s *Service) Delete(userID, path string) error {
	if !s.checkAccess(userID, path, "write") {
		return errors.New("access denied")
	}

	// In production, would use os.RemoveAll
	return nil
}

// Copy copies a file or directory
func (s *Service) Copy(userID, source, dest string) error {
	if !s.checkAccess(userID, source, "read") {
		return errors.New("access denied to source")
	}
	if !s.checkAccess(userID, dest, "write") {
		return errors.New("access denied to destination")
	}

	// In production, would use io.Copy
	return nil
}

// Move moves a file or directory
func (s *Service) Move(userID, source, dest string) error {
	if !s.checkAccess(userID, source, "write") {
		return errors.New("access denied to source")
	}
	if !s.checkAccess(userID, dest, "write") {
		return errors.New("access denied to destination")
	}

	// In production, would use os.Rename
	return nil
}

// Chmod changes file permissions
func (s *Service) Chmod(userID, path, mode string) error {
	if !s.checkAccess(userID, path, "write") {
		return errors.New("access denied")
	}

	// In production, would use os.Chmod
	return nil
}

// Chown changes file ownership
func (s *Service) Chown(userID, path, owner, group string) error {
	if !s.checkAccess(userID, path, "write") {
		return errors.New("access denied")
	}

	// In production, would use os.Chown
	return nil
}

// Compress creates an archive
func (s *Service) Compress(userID string, req *models.ArchiveRequest) error {
	for _, path := range req.SourcePaths {
		if !s.checkAccess(userID, path, "read") {
			return errors.New("access denied to source: " + path)
		}
	}
	if !s.checkAccess(userID, req.DestPath, "write") {
		return errors.New("access denied to destination")
	}

	// In production, would use archive/zip or archive/tar
	return nil
}

// Extract extracts an archive
func (s *Service) Extract(userID string, req *models.ExtractRequest) error {
	if !s.checkAccess(userID, req.SourcePath, "read") {
		return errors.New("access denied to source")
	}
	if !s.checkAccess(userID, req.DestPath, "write") {
		return errors.New("access denied to destination")
	}

	// In production, would use archive/zip or archive/tar
	return nil
}

// CreateMount creates a storage mount
func (s *Service) CreateMount(userID string, req *models.StorageMountCreateRequest) (*models.StorageMount, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	mount := &models.StorageMount{
		ID:         utils.GenerateID("mnt"),
		UserID:     userID,
		MountPoint: req.MountPoint,
		SourcePath: req.SourcePath,
		FSType:     req.FSType,
		Options:    req.Options,
		ReadOnly:   req.ReadOnly,
		CreatedAt:  time.Now(),
	}

	s.mounts[mount.ID] = mount
	s.mountsByUser[userID] = append(s.mountsByUser[userID], mount)

	return mount, nil
}

// GetMount gets a mount by ID
func (s *Service) GetMount(id string) (*models.StorageMount, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	mount, exists := s.mounts[id]
	if !exists {
		return nil, errors.New("mount not found")
	}
	return mount, nil
}

// ListMounts lists mounts for a user
func (s *Service) ListMounts(userID string) []*models.StorageMount {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.mountsByUser[userID]
}

// DeleteMount deletes a mount
func (s *Service) DeleteMount(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	mount, exists := s.mounts[id]
	if !exists {
		return errors.New("mount not found")
	}

	// Remove from user's mounts
	userMounts := s.mountsByUser[mount.UserID]
	for i, m := range userMounts {
		if m.ID == id {
			s.mountsByUser[mount.UserID] = append(userMounts[:i], userMounts[i+1:]...)
			break
		}
	}

	delete(s.mounts, id)
	return nil
}

// checkAccess checks if user has access to a path
func (s *Service) checkAccess(userID, path, action string) bool {
	sanitized := utils.SanitizePath(path)
	
	// Check for path traversal
	if !utils.IsValidPath(sanitized) {
		return false
	}

	// Check if path is within user's home directory
	userHome := "/home/" + userID
	if len(sanitized) < len(userHome) {
		return false
	}

	return true
}

// OperateFile performs file operations
func (s *Service) OperateFile(userID string, req *models.FileOperationRequest) error {
	switch req.Operation {
	case "copy":
		return s.Copy(userID, req.Source, req.Target)
	case "move":
		return s.Move(userID, req.Source, req.Target)
	case "delete":
		return s.Delete(userID, req.Source)
	case "rename":
		return s.Move(userID, req.Source, req.Target)
	case "chmod":
		return s.Chmod(userID, req.Source, req.Mode)
	case "chown":
		return s.Chown(userID, req.Source, req.Owner, req.Group)
	default:
		return errors.New("unknown operation")
	}
}

// InitializeUserFileSystem initializes the file system structure for a new user
func (s *Service) InitializeUserFileSystem(userID, homeDir string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	// In production, this would create directories: public_html, www, logs, mail, tmp, backups
	return nil
}
