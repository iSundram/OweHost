package v1

import (
	"net/http"
	"strings"

	"github.com/iSundram/OweHost/internal/packages"
	"github.com/iSundram/OweHost/pkg/utils"
)

// PackageHandler handles package/plan management
type PackageHandler struct {
	packageService *packages.Service
}

// NewPackageHandler creates a new package handler
func NewPackageHandler(packageSvc *packages.Service) *PackageHandler {
	return &PackageHandler{
		packageService: packageSvc,
	}
}

// List returns all available packages
func (h *PackageHandler) List(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	pkgs, err := h.packageService.List()
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, utils.ErrCodeInternalError, err.Error())
		return
	}

	utils.WriteSuccess(w, pkgs)
}

// Get returns a specific package by name
func (h *PackageHandler) Get(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteError(w, http.StatusMethodNotAllowed, utils.ErrCodeBadRequest, "Method not allowed")
		return
	}

	parts := strings.Split(strings.TrimSuffix(r.URL.Path, "/"), "/")
	if len(parts) < 5 {
		utils.WriteError(w, http.StatusBadRequest, utils.ErrCodeBadRequest, "Package name required")
		return
	}

	packageName := parts[len(parts)-1]

	pkg, err := h.packageService.Get(packageName)
	if err != nil {
		utils.WriteError(w, http.StatusNotFound, utils.ErrCodeNotFound, err.Error())
		return
	}

	utils.WriteSuccess(w, pkg)
}
