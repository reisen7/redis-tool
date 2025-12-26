package websocket

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// MessageType represents the type of WebSocket message
type MessageType string

const (
	// Client -> Server message types
	TypeSubscribe   MessageType = "subscribe"
	TypeUnsubscribe MessageType = "unsubscribe"

	// Server -> Client message types
	TypeMessage MessageType = "message"
	TypeError   MessageType = "error"
	TypeStatus  MessageType = "status"
	TypeSubscribed   MessageType = "subscribed"
	TypeUnsubscribed MessageType = "unsubscribed"
)

// WSMessage represents a WebSocket message
type WSMessage struct {
	Type    MessageType `json:"type"`
	Payload interface{} `json:"payload"`
}

// SubscribePayload represents the payload for subscribe/unsubscribe messages
type SubscribePayload struct {
	Channel   string `json:"channel"`
	IsPattern bool   `json:"isPattern"`
}

// PubSubMessage represents a message received from Redis Pub/Sub
type PubSubMessage struct {
	Channel   string    `json:"channel"`
	Pattern   string    `json:"pattern,omitempty"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

// StatusPayload represents connection status information
type StatusPayload struct {
	Connected     bool     `json:"connected"`
	Subscriptions []string `json:"subscriptions"`
}

// ErrorPayload represents an error message
type ErrorPayload struct {
	Message string `json:"message"`
	Code    string `json:"code,omitempty"`
}

// Client represents a WebSocket client connection
type Client struct {
	ID           string
	UserID       uint
	ConnectionID string // Redis connection ID
	Hub          *Hub
	Conn         *websocket.Conn
	Send         chan []byte
	mu           sync.Mutex
	closed       bool
}

// Hub maintains the set of active clients and broadcasts messages to them
type Hub struct {
	// Registered clients by user ID and connection ID
	clients map[string]map[*Client]bool

	// Register requests from clients
	register chan *Client

	// Unregister requests from clients
	unregister chan *Client

	// Broadcast messages to specific connection
	broadcast chan *BroadcastMessage

	mu sync.RWMutex
}

// BroadcastMessage represents a message to broadcast to clients
type BroadcastMessage struct {
	ConnectionID string
	Message      []byte
}

// NewHub creates a new Hub instance
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[string]map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan *BroadcastMessage, 256),
	}
}


// Run starts the hub's main loop
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.registerClient(client)
		case client := <-h.unregister:
			h.unregisterClient(client)
		case msg := <-h.broadcast:
			h.broadcastToConnection(msg)
		}
	}
}

// registerClient adds a client to the hub
func (h *Hub) registerClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.clients[client.ConnectionID] == nil {
		h.clients[client.ConnectionID] = make(map[*Client]bool)
	}
	h.clients[client.ConnectionID][client] = true
	log.Printf("Client %s registered for connection %s", client.ID, client.ConnectionID)
}

// unregisterClient removes a client from the hub
func (h *Hub) unregisterClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if clients, ok := h.clients[client.ConnectionID]; ok {
		if _, ok := clients[client]; ok {
			delete(clients, client)
			client.Close()
			log.Printf("Client %s unregistered from connection %s", client.ID, client.ConnectionID)

			// Clean up empty connection maps
			if len(clients) == 0 {
				delete(h.clients, client.ConnectionID)
			}
		}
	}
}

// broadcastToConnection sends a message to all clients subscribed to a connection
func (h *Hub) broadcastToConnection(msg *BroadcastMessage) {
	h.mu.RLock()
	clients := h.clients[msg.ConnectionID]
	h.mu.RUnlock()

	for client := range clients {
		select {
		case client.Send <- msg.Message:
		default:
			// Client's send buffer is full, close the connection
			h.unregister <- client
		}
	}
}

// Register adds a client to the hub
func (h *Hub) Register(client *Client) {
	h.register <- client
}

// Unregister removes a client from the hub
func (h *Hub) Unregister(client *Client) {
	h.unregister <- client
}

// Broadcast sends a message to all clients for a specific connection
func (h *Hub) Broadcast(connectionID string, message []byte) {
	h.broadcast <- &BroadcastMessage{
		ConnectionID: connectionID,
		Message:      message,
	}
}

// BroadcastMessage sends a typed message to all clients for a specific connection
func (h *Hub) BroadcastWSMessage(connectionID string, msgType MessageType, payload interface{}) {
	msg := WSMessage{
		Type:    msgType,
		Payload: payload,
	}
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Error marshaling broadcast message: %v", err)
		return
	}
	h.Broadcast(connectionID, data)
}

// GetClientCount returns the number of clients for a connection
func (h *Hub) GetClientCount(connectionID string) int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients[connectionID])
}

// HasClients checks if there are any clients for a connection
func (h *Hub) HasClients(connectionID string) bool {
	return h.GetClientCount(connectionID) > 0
}

// Close closes the client connection
func (c *Client) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return
	}
	c.closed = true
	close(c.Send)
}

// IsClosed checks if the client is closed
func (c *Client) IsClosed() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.closed
}

// SendMessage sends a typed message to the client
func (c *Client) SendMessage(msgType MessageType, payload interface{}) error {
	msg := WSMessage{
		Type:    msgType,
		Payload: payload,
	}
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return nil
	}

	select {
	case c.Send <- data:
		return nil
	default:
		return nil // Buffer full, message dropped
	}
}

// SendError sends an error message to the client
func (c *Client) SendError(message string, code string) error {
	return c.SendMessage(TypeError, ErrorPayload{
		Message: message,
		Code:    code,
	})
}

// SendStatus sends a status message to the client
func (c *Client) SendStatus(connected bool, subscriptions []string) error {
	return c.SendMessage(TypeStatus, StatusPayload{
		Connected:     connected,
		Subscriptions: subscriptions,
	})
}
