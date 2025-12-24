package v1

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/iSundram/OweHost/internal/filesystem"
	"github.com/iSundram/OweHost/internal/user"
	"github.com/iSundram/OweHost/pkg/models"
)

type FileSystemHandler struct {
	fsService   *filesystem.Service
	userService *user.Service
}

func NewFileSystemHandler(fsService *filesystem.Service, userService *user.Service) *FileSystemHandler {
	return &FileSystemHandler{
		fsService:   fsService,
		userService: userService,
	}
}

// ListFiles lists files and directories
func (h *FileSystemHandler) ListFiles(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)
	path := r.URL.Query().Get("path")
	if path == "" {
		path = "/"
	}

	files, err := h.fsService.ListDirectory(userID, path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(files)
}

// GetFile retrieves file information
func (h *FileSystemHandler) GetFile(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)
	path := r.URL.Query().Get("path")
	if path == "" {
		http.Error(w, "Path is required", http.StatusBadRequest)
		return
	}

	file, err := h.fsService.Read(userID, path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(file)
}

// CreateFile creates a new file or directory
func (h *FileSystemHandler) CreateFile(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)

	var req struct {
		Path        string `json:"path"`
		Content     string `json:"content"`
		IsDirectory bool   `json:"is_directory"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	var err error
	if req.IsDirectory {
		err = h.fsService.CreateDirectory(userID, req.Path)
	} else {
		err = h.fsService.Write(userID, req.Path, req.Content)
	}
	
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Created successfully"})
}

// UpdateFile updates file content or metadata
func (h *FileSystemHandler) UpdateFile(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)

	var req struct {
		Path    string `json:"path"`
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err := h.fsService.Write(userID, req.Path, req.Content)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Updated successfully"})
}

// DeleteFile deletes a file or directory
func (h *FileSystemHandler) DeleteFile(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)
	path := r.URL.Query().Get("path")
	if path == "" {
		http.Error(w, "Path is required", http.StatusBadRequest)
		return
	}

	if err := h.fsService.Delete(userID, path); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// DownloadFile downloads a file
func (h *FileSystemHandler) DownloadFile(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)
	path := r.URL.Query().Get("path")
	if path == "" {
		http.Error(w, "Path is required", http.StatusBadRequest)
		return
	}

	content, err := h.fsService.ReadContent(userID, path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// Get file name from path
	parts := strings.Split(path, "/")
	fileName := parts[len(parts)-1]

	w.Header().Set("Content-Disposition", "attachment; filename=\""+fileName+"\"")
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write([]byte(content.Content))
}

// UploadFile uploads a file
func (h *FileSystemHandler) UploadFile(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)
	
	if err := r.ParseMultipartForm(100 << 20); err != nil { // 100 MB max
		http.Error(w, "File too large", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "No file provided", http.StatusBadRequest)
		return
	}
	defer file.Close()

	path := r.FormValue("path")
	if path == "" {
		path = "/"
	}

	// Read file content
	content, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Failed to read file", http.StatusInternalServerError)
		return
	}

	fullPath := path + "/" + header.Filename
	if err := h.fsService.Write(userID, fullPath, string(content)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"path": fullPath, "message": "Uploaded successfully"})
}

// CopyFile copies a file or directory
func (h *FileSystemHandler) CopyFile(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)

	var req struct {
		Source      string `json:"source"`
		Destination string `json:"destination"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err := h.fsService.Copy(userID, req.Source, req.Destination)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Copied successfully"})
}

// MoveFile moves or renames a file or directory
func (h *FileSystemHandler) MoveFile(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)

	var req struct {
		Source      string `json:"source"`
		Destination string `json:"destination"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err := h.fsService.Move(userID, req.Source, req.Destination)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Moved successfully"})
}

// GetPermissions retrieves file permissions
func (h *FileSystemHandler) GetPermissions(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)
	path := r.URL.Query().Get("path")
	if path == "" {
		http.Error(w, "Path is required", http.StatusBadRequest)
		return
	}

	fileInfo, err := h.fsService.Read(userID, path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"path":        fileInfo.Path,
		"permissions": fileInfo.Permissions,
		"owner":       fileInfo.Owner,
		"group":       fileInfo.Group,
	})
}

// UpdatePermissions updates file permissions
func (h *FileSystemHandler) UpdatePermissions(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)

	var req struct {
		Path  string `json:"path"`
		Mode  string `json:"mode"`
		Owner string `json:"owner"`
		Group string `json:"group"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Mode != "" {
		if err := h.fsService.Chmod(userID, req.Path, req.Mode); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	
	if req.Owner != "" || req.Group != "" {
		if err := h.fsService.Chown(userID, req.Path, req.Owner, req.Group); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "updated"})
}

// CompressFiles compresses files or directories
func (h *FileSystemHandler) CompressFiles(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)

	var req models.ArchiveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err := h.fsService.Compress(userID, &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Archive created"})
}

// ExtractArchive extracts an archive file
func (h *FileSystemHandler) ExtractArchive(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)

	var req models.ExtractRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err := h.fsService.Extract(userID, &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Extracted successfully"})
}

// SearchFiles searches for files
func (h *FileSystemHandler) SearchFiles(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)
	
	path := r.URL.Query().Get("path")
	if path == "" {
		path = "/"
	}

	// Return list of files in directory as search result
	files, err := h.fsService.ListDirectory(userID, path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(files)
}

// GetDiskUsage retrieves disk usage information
func (h *FileSystemHandler) GetDiskUsage(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)
	
	// Get mounts for user to show disk usage
	mounts := h.fsService.ListMounts(userID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"user_id": userID,
		"mounts":  mounts,
	})
}
