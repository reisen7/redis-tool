package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	ws "github.com/gorilla/websocket"
	"redis-web-manager/middleware"
	"redis-web-manager/model"
	"redis-web-manager/service"
	"redis-web-manager/websocket"
)

const (
	// Time allowed to write a message to the peer
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer
	pongWait = 60 * time.Second

	// Send pings to peer with this period (must be less than pongWait)
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer
	maxMessageSize = 512
)

var upgrader = ws.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for development
	},
}

var (
	hub           *websocket.Hub
	pubsubManager *service.PubSubManager
)

// InitPubSubHub initializes the WebSocket hub and starts it
func InitPubSubHub() *websocket.Hub {
	if hub == nil {
		hub = websocket.NewHub()
		go hub.Run()
		pubsubManager = service.GetPubSubManager(hub)
	}
	return hub
}

// GetPubSubHub returns the WebSocket hub instance
func GetPubSubHub() *websocket.Hub {
	return hub
}

// HandlePubSubWebSocket handles WebSocket connections for Pub/Sub
// Requirements: 2.6, 2.7
func HandlePubSubWebSocket(c *gin.Context) {
	// Ensure hub is initialized
	if hub == nil {
		InitPubSubHub()
	}

	// Get user ID from context (set by JWT middleware)
	userID := middleware.GetUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "认证已过期",
		})
		return
	}

	// Get connection ID from query parameter
	connectionIDStr := c.Query("connectionId")
	if connectionIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "缺少连接ID",
		})
		return
	}

	connectionID, err := strconv.ParseUint(connectionIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "无效的连接ID",
		})
		return
	}

	// Verify user owns this connection
	connService := service.NewConnectionService()
	userConn, err := connService.GetUserConnectionByID(userID, uint(connectionID))
	if err != nil {
		if err == service.ErrConnectionNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "连接不存在",
			})
			return
		}
		if err == service.ErrConnectionAccessDenied {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   "没有权限",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}

	// Create client
	client := &websocket.Client{
		ID:           uuid.New().String(),
		UserID:       userID,
		ConnectionID: connectionIDStr,
		Hub:          hub,
		Conn:         conn,
		Send:         make(chan []byte, 256),
	}

	// Register client with hub
	hub.Register(client)

	// Ensure Redis connection is established
	if err := ensureRedisConnection(userConn); err != nil {
		client.SendError("无法连接到Redis服务器: "+err.Error(), "REDIS_CONNECTION_ERROR")
		hub.Unregister(client)
		conn.Close()
		return
	}

	// Send initial status
	subs := pubsubManager.GetSubscriptions(connectionIDStr)
	subChannels := make([]string, len(subs))
	for i, s := range subs {
		subChannels[i] = s.Channel
	}
	client.SendStatus(true, subChannels)

	// Start goroutines for reading and writing
	go writePump(client)
	go readPump(client)
}


// ensureRedisConnection ensures a Redis connection is established for the user connection
func ensureRedisConnection(userConn *model.UserConnection) error {
	manager := service.GetManager()
	connID := strconv.FormatUint(uint64(userConn.ID), 10)

	// Check if already connected
	_, err := manager.GetConnection(connID)
	if err == nil {
		return nil // Already connected
	}

	// Get decrypted password
	connService := service.NewConnectionService()
	password, err := connService.GetDecryptedPassword(userConn.UserID, userConn.ID)
	if err != nil {
		return err
	}

	// Create connection config
	config := model.ConnectionConfig{
		ID:       connID,
		Name:     userConn.Name,
		Host:     userConn.Host,
		Port:     userConn.Port,
		Password: password,
		Mode:     model.ModeStandalone,
	}

	return manager.Connect(config)
}

// readPump pumps messages from the WebSocket connection to the hub
func readPump(client *websocket.Client) {
	defer func() {
		hub.Unregister(client)
		client.Conn.Close()
	}()

	client.Conn.SetReadLimit(maxMessageSize)
	client.Conn.SetReadDeadline(time.Now().Add(pongWait))
	client.Conn.SetPongHandler(func(string) error {
		client.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := client.Conn.ReadMessage()
		if err != nil {
			if ws.IsUnexpectedCloseError(err, ws.CloseGoingAway, ws.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		handleClientMessage(client, message)
	}
}

// writePump pumps messages from the hub to the WebSocket connection
func writePump(client *websocket.Client) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		client.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-client.Send:
			client.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// Hub closed the channel
				client.Conn.WriteMessage(ws.CloseMessage, []byte{})
				return
			}

			w, err := client.Conn.NextWriter(ws.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued messages to the current WebSocket message
			n := len(client.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-client.Send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			client.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := client.Conn.WriteMessage(ws.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleClientMessage processes messages from WebSocket clients
func handleClientMessage(client *websocket.Client, message []byte) {
	var msg websocket.WSMessage
	if err := json.Unmarshal(message, &msg); err != nil {
		client.SendError("无效的消息格式", "INVALID_MESSAGE")
		return
	}

	switch msg.Type {
	case websocket.TypeSubscribe:
		handleSubscribe(client, msg.Payload)
	case websocket.TypeUnsubscribe:
		handleUnsubscribe(client, msg.Payload)
	default:
		client.SendError("未知的消息类型", "UNKNOWN_MESSAGE_TYPE")
	}
}

// handleSubscribe handles subscribe requests
func handleSubscribe(client *websocket.Client, payload interface{}) {
	data, err := json.Marshal(payload)
	if err != nil {
		client.SendError("无效的订阅请求", "INVALID_PAYLOAD")
		return
	}

	var subPayload websocket.SubscribePayload
	if err := json.Unmarshal(data, &subPayload); err != nil {
		client.SendError("无效的订阅请求", "INVALID_PAYLOAD")
		return
	}

	if subPayload.Channel == "" {
		client.SendError("频道名称不能为空", "EMPTY_CHANNEL")
		return
	}

	err = pubsubManager.Subscribe(client.ConnectionID, subPayload.Channel, subPayload.IsPattern)
	if err != nil {
		client.SendError("订阅失败: "+err.Error(), "SUBSCRIBE_ERROR")
		return
	}

	// Send confirmation
	client.SendMessage(websocket.TypeSubscribed, subPayload)
	log.Printf("Client %s subscribed to %s (pattern: %v)", client.ID, subPayload.Channel, subPayload.IsPattern)
}

// handleUnsubscribe handles unsubscribe requests
func handleUnsubscribe(client *websocket.Client, payload interface{}) {
	data, err := json.Marshal(payload)
	if err != nil {
		client.SendError("无效的取消订阅请求", "INVALID_PAYLOAD")
		return
	}

	var subPayload websocket.SubscribePayload
	if err := json.Unmarshal(data, &subPayload); err != nil {
		client.SendError("无效的取消订阅请求", "INVALID_PAYLOAD")
		return
	}

	if subPayload.Channel == "" {
		client.SendError("频道名称不能为空", "EMPTY_CHANNEL")
		return
	}

	err = pubsubManager.Unsubscribe(client.ConnectionID, subPayload.Channel, subPayload.IsPattern)
	if err != nil {
		client.SendError("取消订阅失败: "+err.Error(), "UNSUBSCRIBE_ERROR")
		return
	}

	// Send confirmation
	client.SendMessage(websocket.TypeUnsubscribed, subPayload)
	log.Printf("Client %s unsubscribed from %s (pattern: %v)", client.ID, subPayload.Channel, subPayload.IsPattern)
}

// PublishMessage handles HTTP POST requests to publish messages
// POST /api/pubsub/publish
// Requirements: 2.4
func PublishMessage(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "认证已过期",
		})
		return
	}

	var req struct {
		ConnectionID uint   `json:"connectionId" binding:"required"`
		Channel      string `json:"channel" binding:"required"`
		Message      string `json:"message" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "参数错误: " + err.Error(),
		})
		return
	}

	// Verify user owns this connection
	connService := service.NewConnectionService()
	userConn, err := connService.GetUserConnectionByID(userID, req.ConnectionID)
	if err != nil {
		if err == service.ErrConnectionNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "连接不存在",
			})
			return
		}
		if err == service.ErrConnectionAccessDenied {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   "没有权限",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// Ensure Redis connection
	if err := ensureRedisConnection(userConn); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "无法连接到Redis服务器: " + err.Error(),
		})
		return
	}

	// Publish message
	connectionIDStr := strconv.FormatUint(uint64(req.ConnectionID), 10)
	if pubsubManager == nil {
		InitPubSubHub()
	}

	err = pubsubManager.Publish(connectionIDStr, req.Channel, req.Message)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "发布消息失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"channel": req.Channel,
			"message": req.Message,
		},
	})
}
