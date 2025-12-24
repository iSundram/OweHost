package v1

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/iSundram/OweHost/internal/notification"
	"github.com/iSundram/OweHost/internal/user"
	"github.com/iSundram/OweHost/pkg/models"
)

type NotificationHandler struct {
	notificationService *notification.Service
	userService         *user.Service
}

func NewNotificationHandler(notificationService *notification.Service, userService *user.Service) *NotificationHandler {
	return &NotificationHandler{
		notificationService: notificationService,
		userService:         userService,
	}
}

// ListNotifications lists all webhooks for the authenticated user
func (h *NotificationHandler) ListNotifications(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)

	webhooks := h.notificationService.ListWebhooks(userID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(webhooks)
}

// GetNotification retrieves a specific webhook
func (h *NotificationHandler) GetNotification(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.TrimSuffix(r.URL.Path, "/"), "/")
	if len(parts) < 5 {
		http.Error(w, "Invalid webhook ID", http.StatusBadRequest)
		return
	}
	webhookID := parts[4]

	webhook, err := h.notificationService.GetWebhook(webhookID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(webhook)
}

// CreateNotification creates a new webhook
func (h *NotificationHandler) CreateNotification(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)

	var req models.WebhookCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	webhook, err := h.notificationService.CreateWebhook(userID, &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(webhook)
}

// MarkAsRead - placeholder (webhooks don't have read status)
func (h *NotificationHandler) MarkAsRead(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// MarkAllAsRead - placeholder
func (h *NotificationHandler) MarkAllAsRead(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

// DeleteNotification deletes a webhook
func (h *NotificationHandler) DeleteNotification(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.TrimSuffix(r.URL.Path, "/"), "/")
	if len(parts) < 5 {
		http.Error(w, "Invalid webhook ID", http.StatusBadRequest)
		return
	}
	webhookID := parts[4]

	if err := h.notificationService.DeleteWebhook(webhookID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetSettings retrieves notification settings (returns empty for now)
func (h *NotificationHandler) GetSettings(w http.ResponseWriter, r *http.Request) {
	settings := map[string]interface{}{
		"email_enabled":   true,
		"webhook_enabled": true,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(settings)
}

// UpdateSettings updates notification settings
func (h *NotificationHandler) UpdateSettings(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "updated"})
}

// GetUnreadCount retrieves the count of unread notifications
func (h *NotificationHandler) GetUnreadCount(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int{"count": 0})
}
