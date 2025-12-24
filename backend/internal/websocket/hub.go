// Package websocket provides WebSocket support for real-time updates in OweHost
package websocket

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// In production, validate origin properly
		return true
	},
}

// MessageType represents the type of WebSocket message
type MessageType string

const (
	MessageTypeNotification  MessageType = "notification"
	MessageTypeBackupStatus  MessageType = "backup_status"
	MessageTypeInstallStatus MessageType = "install_status"
	MessageTypeResourceUsage MessageType = "resource_usage"
	MessageTypeSystemAlert   MessageType = "system_alert"
	MessageTypePing          MessageType = "ping"
	MessageTypePong          MessageType = "pong"
)

// Message represents a WebSocket message
type Message struct {
	Type      MessageType            `json:"type"`
	Channel   string                 `json:"channel,omitempty"`
	Data      map[string]interface{} `json:"data,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

// Client represents a WebSocket client connection
type Client struct {
	ID       string
	UserID   string
	Conn     *websocket.Conn
	Send     chan *Message
	Hub      *Hub
	Channels map[string]bool
	mu       sync.RWMutex
}

// Hub maintains the set of active clients and broadcasts messages
type Hub struct {
	clients    map[*Client]bool
	byUser     map[string][]*Client
	broadcast  chan *Message
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
}

// NewHub creates a new WebSocket hub
func NewHub() *Hub {
	hub := &Hub{
		clients:    make(map[*Client]bool),
		byUser:     make(map[string][]*Client),
		broadcast:  make(chan *Message, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
	go hub.run()
	return hub
}

// run handles hub events
func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.byUser[client.UserID] = append(h.byUser[client.UserID], client)
			h.mu.Unlock()
			log.Printf("WebSocket: Client connected, user=%s, total=%d", client.UserID, len(h.clients))

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.Send)
				
				// Remove from byUser
				userClients := h.byUser[client.UserID]
				for i, c := range userClients {
					if c == client {
						h.byUser[client.UserID] = append(userClients[:i], userClients[i+1:]...)
						break
					}
				}
			}
			h.mu.Unlock()
			log.Printf("WebSocket: Client disconnected, user=%s, total=%d", client.UserID, len(h.clients))

		case message := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				// Check if client is subscribed to the channel
				if message.Channel == "" || client.IsSubscribed(message.Channel) {
					select {
					case client.Send <- message:
					default:
						// Client buffer full, will be cleaned up
					}
				}
			}
			h.mu.RUnlock()
		}
	}
}

// Broadcast sends a message to all connected clients
func (h *Hub) Broadcast(msg *Message) {
	msg.Timestamp = time.Now()
	h.broadcast <- msg
}

// SendToUser sends a message to all connections of a specific user
func (h *Hub) SendToUser(userID string, msg *Message) {
	h.mu.RLock()
	clients := h.byUser[userID]
	h.mu.RUnlock()

	msg.Timestamp = time.Now()
	for _, client := range clients {
		select {
		case client.Send <- msg:
		default:
			// Client buffer full
		}
	}
}

// SendToChannel sends a message to all clients subscribed to a channel
func (h *Hub) SendToChannel(channel string, msg *Message) {
	msg.Channel = channel
	msg.Timestamp = time.Now()
	h.broadcast <- msg
}

// GetConnectedUsers returns the number of connected users
func (h *Hub) GetConnectedUsers() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.byUser)
}

// GetTotalConnections returns the total number of connections
func (h *Hub) GetTotalConnections() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// IsSubscribed checks if client is subscribed to a channel
func (c *Client) IsSubscribed(channel string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Channels[channel]
}

// Subscribe subscribes the client to a channel
func (c *Client) Subscribe(channel string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Channels[channel] = true
}

// Unsubscribe unsubscribes the client from a channel
func (c *Client) Unsubscribe(channel string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.Channels, channel)
}

// readPump pumps messages from the WebSocket connection
func (c *Client) readPump() {
	defer func() {
		c.Hub.unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(512)
	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// Parse message
		var msg Message
		if err := json.Unmarshal(message, &msg); err != nil {
			continue
		}

		// Handle client messages
		switch msg.Type {
		case MessageTypePing:
			c.Send <- &Message{Type: MessageTypePong, Timestamp: time.Now()}
		case "subscribe":
			if channel, ok := msg.Data["channel"].(string); ok {
				c.Subscribe(channel)
			}
		case "unsubscribe":
			if channel, ok := msg.Data["channel"].(string); ok {
				c.Unsubscribe(channel)
			}
		}
	}
}

// writePump pumps messages to the WebSocket connection
func (c *Client) writePump() {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}

			data, _ := json.Marshal(message)
			w.Write(data)

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// Handler returns an HTTP handler for WebSocket connections
func (h *Hub) Handler(getUserID func(r *http.Request) string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := getUserID(r)
		if userID == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("WebSocket upgrade error: %v", err)
			return
		}

		client := &Client{
			ID:       generateClientID(),
			UserID:   userID,
			Hub:      h,
			Conn:     conn,
			Send:     make(chan *Message, 256),
			Channels: make(map[string]bool),
		}

		h.register <- client

		go client.writePump()
		go client.readPump()
	}
}

// generateClientID generates a unique client ID
func generateClientID() string {
	return time.Now().Format("20060102150405.000000000")
}

// NotifyBackupProgress sends backup progress update
func (h *Hub) NotifyBackupProgress(userID, backupID string, progress int, status string) {
	h.SendToUser(userID, &Message{
		Type: MessageTypeBackupStatus,
		Data: map[string]interface{}{
			"backup_id": backupID,
			"progress":  progress,
			"status":    status,
		},
	})
}

// NotifyInstallProgress sends app installation progress update
func (h *Hub) NotifyInstallProgress(userID, installID string, progress int, step string) {
	h.SendToUser(userID, &Message{
		Type: MessageTypeInstallStatus,
		Data: map[string]interface{}{
			"install_id": installID,
			"progress":   progress,
			"step":       step,
		},
	})
}

// NotifyResourceUsage sends resource usage update
func (h *Hub) NotifyResourceUsage(userID string, cpu, memory, disk float64) {
	h.SendToUser(userID, &Message{
		Type: MessageTypeResourceUsage,
		Data: map[string]interface{}{
			"cpu":    cpu,
			"memory": memory,
			"disk":   disk,
		},
	})
}

// NotifySystemAlert sends a system alert to admins
func (h *Hub) NotifySystemAlert(severity, title, message string) {
	h.SendToChannel("admin", &Message{
		Type:    MessageTypeSystemAlert,
		Channel: "admin",
		Data: map[string]interface{}{
			"severity": severity,
			"title":    title,
			"message":  message,
		},
	})
}
