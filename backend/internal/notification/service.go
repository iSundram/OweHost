// Package notification provides notification and event services for OweHost
package notification

import (
	"errors"
	"sync"
	"time"

	"github.com/iSundram/OweHost/pkg/models"
	"github.com/iSundram/OweHost/pkg/utils"
)

// Service provides notification functionality
type Service struct {
	webhooks      map[string]*models.Webhook
	deliveries    map[string][]*models.WebhookDelivery
	subscriptions map[string][]*models.EventSubscription
	events        chan *models.Event
	byUser        map[string][]*models.Webhook
	mu            sync.RWMutex
}

// NewService creates a new notification service
func NewService() *Service {
	svc := &Service{
		webhooks:      make(map[string]*models.Webhook),
		deliveries:    make(map[string][]*models.WebhookDelivery),
		subscriptions: make(map[string][]*models.EventSubscription),
		events:        make(chan *models.Event, 1000),
		byUser:        make(map[string][]*models.Webhook),
	}
	go svc.processEvents()
	return svc
}

// processEvents processes events from the channel
func (s *Service) processEvents() {
	for event := range s.events {
		s.handleEvent(event)
	}
}

// handleEvent handles a single event
func (s *Service) handleEvent(event *models.Event) {
	s.mu.RLock()
	subs := s.subscriptions[string(event.Type)]
	s.mu.RUnlock()

	for _, sub := range subs {
		// Execute handler (in production, would call actual handlers)
		_ = sub
	}

	// Trigger webhooks
	s.triggerWebhooks(event)
}

// Publish publishes an event
func (s *Service) Publish(eventType models.EventType, source, subject string, data map[string]interface{}) *models.Event {
	event := &models.Event{
		ID:        utils.GenerateID("evt"),
		Type:      eventType,
		Source:    source,
		Subject:   subject,
		Data:      data,
		Timestamp: time.Now(),
	}

	select {
	case s.events <- event:
	default:
		// Channel full, drop event (in production, would handle this better)
	}

	return event
}

// Subscribe subscribes to events
func (s *Service) Subscribe(eventType models.EventType, handler string, filter string) *models.EventSubscription {
	s.mu.Lock()
	defer s.mu.Unlock()

	sub := &models.EventSubscription{
		ID:        utils.GenerateID("sub"),
		EventType: eventType,
		Handler:   handler,
		Filter:    filter,
		CreatedAt: time.Now(),
	}

	key := string(eventType)
	s.subscriptions[key] = append(s.subscriptions[key], sub)

	return sub
}

// Unsubscribe removes an event subscription
func (s *Service) Unsubscribe(subID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for eventType, subs := range s.subscriptions {
		for i, sub := range subs {
			if sub.ID == subID {
				s.subscriptions[eventType] = append(subs[:i], subs[i+1:]...)
				return nil
			}
		}
	}

	return errors.New("subscription not found")
}

// CreateWebhook creates a webhook
func (s *Service) CreateWebhook(userID string, req *models.WebhookCreateRequest) (*models.Webhook, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	secret, _ := utils.GenerateSecureToken(32)

	retryCount := req.RetryCount
	if retryCount == 0 {
		retryCount = 3
	}

	retryDelay := req.RetryDelay
	if retryDelay == 0 {
		retryDelay = 60
	}

	webhook := &models.Webhook{
		ID:         utils.GenerateID("whk"),
		UserID:     userID,
		Name:       req.Name,
		URL:        req.URL,
		Secret:     secret,
		Events:     req.Events,
		Enabled:    true,
		RetryCount: retryCount,
		RetryDelay: retryDelay,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	s.webhooks[webhook.ID] = webhook
	s.byUser[userID] = append(s.byUser[userID], webhook)
	s.deliveries[webhook.ID] = make([]*models.WebhookDelivery, 0)

	return webhook, nil
}

// GetWebhook gets a webhook by ID
func (s *Service) GetWebhook(id string) (*models.Webhook, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	webhook, exists := s.webhooks[id]
	if !exists {
		return nil, errors.New("webhook not found")
	}
	return webhook, nil
}

// ListWebhooks lists webhooks for a user
func (s *Service) ListWebhooks(userID string) []*models.Webhook {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.byUser[userID]
}

// UpdateWebhook updates a webhook
func (s *Service) UpdateWebhook(id string, req *models.WebhookUpdateRequest) (*models.Webhook, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	webhook, exists := s.webhooks[id]
	if !exists {
		return nil, errors.New("webhook not found")
	}

	if req.Name != nil {
		webhook.Name = *req.Name
	}
	if req.URL != nil {
		webhook.URL = *req.URL
	}
	if req.Events != nil {
		webhook.Events = *req.Events
	}
	if req.Enabled != nil {
		webhook.Enabled = *req.Enabled
	}
	if req.RetryCount != nil {
		webhook.RetryCount = *req.RetryCount
	}
	if req.RetryDelay != nil {
		webhook.RetryDelay = *req.RetryDelay
	}

	webhook.UpdatedAt = time.Now()
	return webhook, nil
}

// DeleteWebhook deletes a webhook
func (s *Service) DeleteWebhook(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	webhook, exists := s.webhooks[id]
	if !exists {
		return errors.New("webhook not found")
	}

	// Remove from user's webhooks
	userWebhooks := s.byUser[webhook.UserID]
	for i, w := range userWebhooks {
		if w.ID == id {
			s.byUser[webhook.UserID] = append(userWebhooks[:i], userWebhooks[i+1:]...)
			break
		}
	}

	delete(s.webhooks, id)
	delete(s.deliveries, id)
	return nil
}

// triggerWebhooks triggers webhooks for an event
func (s *Service) triggerWebhooks(event *models.Event) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, webhook := range s.webhooks {
		if !webhook.Enabled {
			continue
		}

		// Check if webhook subscribes to this event
		for _, e := range webhook.Events {
			if e == string(event.Type) || e == "*" {
				go s.deliverWebhook(webhook.ID, event)
				break
			}
		}
	}
}

// deliverWebhook delivers a webhook
func (s *Service) deliverWebhook(webhookID string, event *models.Event) {
	s.mu.Lock()
	webhook := s.webhooks[webhookID]
	if webhook == nil {
		s.mu.Unlock()
		return
	}
	s.mu.Unlock()

	delivery := &models.WebhookDelivery{
		ID:          utils.GenerateID("del"),
		WebhookID:   webhookID,
		EventID:     event.ID,
		Payload:     "", // Would serialize event to JSON
		StatusCode:  0,
		Response:    "",
		Attempt:     1,
		Success:     false,
		DeliveredAt: time.Now(),
	}

	// In production, would make HTTP request here
	// For now, simulate success
	delivery.StatusCode = 200
	delivery.Success = true

	s.mu.Lock()
	s.deliveries[webhookID] = append(s.deliveries[webhookID], delivery)
	now := time.Now()
	webhook.LastCalledAt = &now

	// Keep only last 100 deliveries per webhook
	if len(s.deliveries[webhookID]) > 100 {
		s.deliveries[webhookID] = s.deliveries[webhookID][len(s.deliveries[webhookID])-100:]
	}
	s.mu.Unlock()
}

// GetDeliveries gets webhook deliveries
func (s *Service) GetDeliveries(webhookID string, limit int) []*models.WebhookDelivery {
	s.mu.RLock()
	defer s.mu.RUnlock()

	deliveries := s.deliveries[webhookID]
	if limit <= 0 || limit > len(deliveries) {
		limit = len(deliveries)
	}

	// Return most recent
	start := len(deliveries) - limit
	if start < 0 {
		start = 0
	}
	return deliveries[start:]
}

// RetryDelivery retries a failed delivery
func (s *Service) RetryDelivery(deliveryID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for webhookID, deliveries := range s.deliveries {
		for _, del := range deliveries {
			if del.ID == deliveryID {
				webhook := s.webhooks[webhookID]
				if webhook == nil {
					return errors.New("webhook not found")
				}

				if del.Attempt >= webhook.RetryCount {
					return errors.New("max retries exceeded")
				}

				// Create new delivery attempt
				newDel := &models.WebhookDelivery{
					ID:          utils.GenerateID("del"),
					WebhookID:   webhookID,
					EventID:     del.EventID,
					Payload:     del.Payload,
					StatusCode:  200, // Simulate success
					Response:    "",
					Attempt:     del.Attempt + 1,
					Success:     true,
					DeliveredAt: time.Now(),
				}

				s.deliveries[webhookID] = append(s.deliveries[webhookID], newDel)
				return nil
			}
		}
	}

	return errors.New("delivery not found")
}
