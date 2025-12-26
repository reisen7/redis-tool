package service

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"redis-web-manager/model"
	"redis-web-manager/websocket"
)

// PubSubManager manages Redis Pub/Sub subscriptions for WebSocket clients
type PubSubManager struct {
	hub           *websocket.Hub
	subscriptions map[string]*ConnectionSubscription // connectionID -> subscription
	mu            sync.RWMutex
}

// ConnectionSubscription manages subscriptions for a single Redis connection
type ConnectionSubscription struct {
	connectionID string
	client       redis.UniversalClient
	pubsub       *redis.PubSub
	channels     map[string]bool // channel/pattern -> isPattern
	mu           sync.RWMutex
	ctx          context.Context
	cancel       context.CancelFunc
	hub          *websocket.Hub
}

// Subscription represents a single subscription
type Subscription struct {
	Channel   string `json:"channel"`
	IsPattern bool   `json:"isPattern"`
}

var (
	pubsubManager     *PubSubManager
	pubsubManagerOnce sync.Once
)

// GetPubSubManager returns the singleton PubSubManager instance
func GetPubSubManager(hub *websocket.Hub) *PubSubManager {
	pubsubManagerOnce.Do(func() {
		pubsubManager = &PubSubManager{
			hub:           hub,
			subscriptions: make(map[string]*ConnectionSubscription),
		}
	})
	return pubsubManager
}

// Subscribe subscribes to a channel or pattern for a connection
// Requirements: 2.2
func (m *PubSubManager) Subscribe(connectionID string, channel string, isPattern bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Get or create connection subscription
	connSub, exists := m.subscriptions[connectionID]
	if !exists {
		// Get Redis connection
		conn, err := GetManager().GetConnection(connectionID)
		if err != nil {
			return fmt.Errorf("connection not found: %w", err)
		}

		ctx, cancel := context.WithCancel(context.Background())
		connSub = &ConnectionSubscription{
			connectionID: connectionID,
			client:       conn.Client,
			channels:     make(map[string]bool),
			ctx:          ctx,
			cancel:       cancel,
			hub:          m.hub,
		}
		m.subscriptions[connectionID] = connSub
	}

	return connSub.Subscribe(channel, isPattern)
}


// Unsubscribe unsubscribes from a channel or pattern
// Requirements: 2.5
func (m *PubSubManager) Unsubscribe(connectionID string, channel string, isPattern bool) error {
	m.mu.RLock()
	connSub, exists := m.subscriptions[connectionID]
	m.mu.RUnlock()

	if !exists {
		return nil // No subscriptions for this connection
	}

	return connSub.Unsubscribe(channel, isPattern)
}

// GetSubscriptions returns all subscriptions for a connection
func (m *PubSubManager) GetSubscriptions(connectionID string) []Subscription {
	m.mu.RLock()
	connSub, exists := m.subscriptions[connectionID]
	m.mu.RUnlock()

	if !exists {
		return []Subscription{}
	}

	return connSub.GetSubscriptions()
}

// CleanupConnection removes all subscriptions for a connection
func (m *PubSubManager) CleanupConnection(connectionID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if connSub, exists := m.subscriptions[connectionID]; exists {
		connSub.Close()
		delete(m.subscriptions, connectionID)
	}
}

// Publish publishes a message to a channel
// Requirements: 2.4
func (m *PubSubManager) Publish(connectionID string, channel string, message string) error {
	conn, err := GetManager().GetConnection(connectionID)
	if err != nil {
		return fmt.Errorf("connection not found: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return conn.Client.Publish(ctx, channel, message).Err()
}

// Subscribe subscribes to a channel or pattern
func (cs *ConnectionSubscription) Subscribe(channel string, isPattern bool) error {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	// Check if already subscribed
	key := cs.subscriptionKey(channel, isPattern)
	if cs.channels[key] {
		return nil // Already subscribed
	}

	// Create pubsub if not exists
	if cs.pubsub == nil {
		cs.pubsub = cs.client.Subscribe(cs.ctx)
		go cs.receiveMessages()
	}

	var err error
	if isPattern {
		err = cs.pubsub.PSubscribe(cs.ctx, channel)
	} else {
		err = cs.pubsub.Subscribe(cs.ctx, channel)
	}

	if err != nil {
		return fmt.Errorf("subscribe failed: %w", err)
	}

	cs.channels[key] = isPattern
	log.Printf("Subscribed to %s (pattern: %v) for connection %s", channel, isPattern, cs.connectionID)

	return nil
}

// Unsubscribe unsubscribes from a channel or pattern
func (cs *ConnectionSubscription) Unsubscribe(channel string, isPattern bool) error {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	key := cs.subscriptionKey(channel, isPattern)
	if !cs.channels[key] {
		return nil // Not subscribed
	}

	if cs.pubsub == nil {
		return nil
	}

	var err error
	if isPattern {
		err = cs.pubsub.PUnsubscribe(cs.ctx, channel)
	} else {
		err = cs.pubsub.Unsubscribe(cs.ctx, channel)
	}

	if err != nil {
		return fmt.Errorf("unsubscribe failed: %w", err)
	}

	delete(cs.channels, key)
	log.Printf("Unsubscribed from %s (pattern: %v) for connection %s", channel, isPattern, cs.connectionID)

	return nil
}

// GetSubscriptions returns all active subscriptions
func (cs *ConnectionSubscription) GetSubscriptions() []Subscription {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	subs := make([]Subscription, 0, len(cs.channels))
	for key, isPattern := range cs.channels {
		channel := cs.channelFromKey(key, isPattern)
		subs = append(subs, Subscription{
			Channel:   channel,
			IsPattern: isPattern,
		})
	}
	return subs
}

// Close closes the subscription and releases resources
func (cs *ConnectionSubscription) Close() {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	if cs.cancel != nil {
		cs.cancel()
	}

	if cs.pubsub != nil {
		cs.pubsub.Close()
		cs.pubsub = nil
	}

	cs.channels = make(map[string]bool)
	log.Printf("Closed all subscriptions for connection %s", cs.connectionID)
}

// receiveMessages listens for messages from Redis and forwards them to WebSocket clients
func (cs *ConnectionSubscription) receiveMessages() {
	ch := cs.pubsub.Channel()

	for {
		select {
		case <-cs.ctx.Done():
			return
		case msg, ok := <-ch:
			if !ok {
				return
			}
			cs.handleMessage(msg)
		}
	}
}

// handleMessage processes a received Redis message and broadcasts it
func (cs *ConnectionSubscription) handleMessage(msg *redis.Message) {
	pubsubMsg := websocket.PubSubMessage{
		Channel:   msg.Channel,
		Pattern:   msg.Pattern,
		Message:   msg.Payload,
		Timestamp: time.Now(),
	}

	cs.hub.BroadcastWSMessage(cs.connectionID, websocket.TypeMessage, pubsubMsg)
	log.Printf("Broadcasted message from channel %s to connection %s", msg.Channel, cs.connectionID)
}

// subscriptionKey generates a unique key for a subscription
func (cs *ConnectionSubscription) subscriptionKey(channel string, isPattern bool) string {
	if isPattern {
		return "p:" + channel
	}
	return "c:" + channel
}

// channelFromKey extracts the channel name from a subscription key
func (cs *ConnectionSubscription) channelFromKey(key string, isPattern bool) string {
	if isPattern {
		return key[2:] // Remove "p:" prefix
	}
	return key[2:] // Remove "c:" prefix
}

// PublishRequest represents a publish request
type PublishRequest struct {
	ConnectionID string `json:"connectionId"`
	Channel      string `json:"channel"`
	Message      string `json:"message"`
}

// TestUserConnectionForPubSub tests if a user connection can be used for pub/sub
func TestUserConnectionForPubSub(userID uint, connectionID uint) (*model.UserConnection, error) {
	connService := NewConnectionService()
	return connService.GetUserConnectionByID(userID, connectionID)
}
