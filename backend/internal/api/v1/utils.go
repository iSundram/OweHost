// Package v1 provides shared utilities for API handlers
package v1

import (
	"strings"
)

// extractIDFromPath extracts a resource ID from the URL path
// e.g., extractIDFromPath("/api/v1/ftp/accounts/ftp_123", "ftp/accounts") returns "ftp_123"
func extractIDFromPath(path, resourcePrefix string) string {
	// Clean the path
	path = strings.TrimSuffix(path, "/")
	
	// Find the resource prefix in the path
	idx := strings.Index(path, resourcePrefix)
	if idx == -1 {
		return ""
	}
	
	// Get everything after the prefix
	remainder := path[idx+len(resourcePrefix):]
	remainder = strings.TrimPrefix(remainder, "/")
	
	// Take the first segment (the ID)
	parts := strings.Split(remainder, "/")
	if len(parts) > 0 {
		return parts[0]
	}
	
	return ""
}
