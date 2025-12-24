package models

// ServiceStatus represents the status of a running service/daemon
type ServiceStatus struct {
	Name   string `json:"name"`
	Status string `json:"status"` // running|stopped|warning
	Uptime string `json:"uptime"`
	Load   string `json:"load,omitempty"`
	PID    int    `json:"pid,omitempty"`
}
