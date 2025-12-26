package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"redis-web-manager/middleware"
	"redis-web-manager/model"
	"redis-web-manager/service"
)

// UserConnectionRequest represents a request to create/update a user connection
type UserConnectionRequest struct {
	Name     string `json:"name" binding:"required"`
	GroupID  string `json:"groupId"`
	Host     string `json:"host" binding:"required"`
	Port     int    `json:"port"`
	Password string `json:"password"`
	DB       int    `json:"db"`
	Timeout  int    `json:"timeout"`
}

// UserConnectionResponse represents a user connection in API responses
type UserConnectionResponse struct {
	ID        string `json:"id"`
	UserID    uint   `json:"userId"`
	GroupID   string `json:"groupId"`
	Name      string `json:"name"`
	Host      string `json:"host"`
	Port      int    `json:"port"`
	DB        int    `json:"db"`
	Timeout   int    `json:"timeout"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

// toResponse converts UserConnection to UserConnectionResponse (without password)
func toResponse(conn *model.UserConnection) UserConnectionResponse {
	return UserConnectionResponse{
		ID:        strconv.FormatUint(uint64(conn.ID), 10),
		UserID:    conn.UserID,
		GroupID:   conn.GroupID,
		Name:      conn.Name,
		Host:      conn.Host,
		Port:      conn.Port,
		DB:        conn.DB,
		Timeout:   conn.Timeout,
		CreatedAt: conn.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt: conn.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

// ListUserConnections returns all connections for the authenticated user
// GET /api/connections
// Requirements: 6.1, 6.2
func ListUserConnections(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, model.APIResponse{
			Success: false,
			Error:   "认证已过期",
		})
		return
	}

	connService := service.NewConnectionService()
	connections, err := connService.GetUserConnections(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.APIResponse{
			Success: false,
			Error:   "获取连接列表失败",
		})
		return
	}

	// Convert to response format (without passwords)
	responses := make([]UserConnectionResponse, len(connections))
	for i, conn := range connections {
		responses[i] = toResponse(&conn)
	}

	c.JSON(http.StatusOK, model.APIResponse{
		Success: true,
		Data:    responses,
	})
}

// CreateUserConnection creates a new connection for the authenticated user
// POST /api/connections
// Requirements: 6.1
func CreateUserConnection(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, model.APIResponse{
			Success: false,
			Error:   "认证已过期",
		})
		return
	}

	var req UserConnectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Error:   "参数错误: " + err.Error(),
		})
		return
	}

	// Set defaults
	if req.Port == 0 {
		req.Port = 6379
	}
	if req.Timeout == 0 {
		req.Timeout = 5
	}

	conn := &model.UserConnection{
		Name:        req.Name,
		GroupID:     req.GroupID,
		Host:        req.Host,
		Port:        req.Port,
		PasswordEnc: req.Password, // Will be encrypted by service
		DB:          req.DB,
		Timeout:     req.Timeout,
	}

	connService := service.NewConnectionService()
	if err := connService.CreateUserConnection(userID, conn); err != nil {
		c.JSON(http.StatusInternalServerError, model.APIResponse{
			Success: false,
			Error:   "创建连接失败",
		})
		return
	}

	c.JSON(http.StatusOK, model.APIResponse{
		Success: true,
		Data:    toResponse(conn),
	})
}

// GetUserConnection returns a specific connection for the authenticated user
// GET /api/connections/:id
// Requirements: 6.2, 6.3
func GetUserConnection(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, model.APIResponse{
			Success: false,
			Error:   "认证已过期",
		})
		return
	}

	connID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Error:   "无效的连接ID",
		})
		return
	}

	connService := service.NewConnectionService()
	conn, err := connService.GetUserConnectionByID(userID, uint(connID))
	if err != nil {
		if errors.Is(err, service.ErrConnectionNotFound) {
			c.JSON(http.StatusNotFound, model.APIResponse{
				Success: false,
				Error:   "连接不存在",
			})
			return
		}
		if errors.Is(err, service.ErrConnectionAccessDenied) {
			c.JSON(http.StatusForbidden, model.APIResponse{
				Success: false,
				Error:   "没有权限",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, model.APIResponse{
			Success: false,
			Error:   "获取连接失败",
		})
		return
	}

	c.JSON(http.StatusOK, model.APIResponse{
		Success: true,
		Data:    toResponse(conn),
	})
}

// UpdateUserConnection updates a connection for the authenticated user
// PUT /api/connections/:id
// Requirements: 6.2, 6.3
func UpdateUserConnection(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, model.APIResponse{
			Success: false,
			Error:   "认证已过期",
		})
		return
	}

	connID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Error:   "无效的连接ID",
		})
		return
	}

	var req UserConnectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Error:   "参数错误: " + err.Error(),
		})
		return
	}

	// Set defaults
	if req.Port == 0 {
		req.Port = 6379
	}
	if req.Timeout == 0 {
		req.Timeout = 5
	}

	updates := &model.UserConnection{
		Name:        req.Name,
		GroupID:     req.GroupID,
		Host:        req.Host,
		Port:        req.Port,
		PasswordEnc: req.Password, // Will be encrypted by service if not empty
		DB:          req.DB,
		Timeout:     req.Timeout,
	}

	connService := service.NewConnectionService()
	conn, err := connService.UpdateUserConnection(userID, uint(connID), updates)
	if err != nil {
		if errors.Is(err, service.ErrConnectionNotFound) {
			c.JSON(http.StatusNotFound, model.APIResponse{
				Success: false,
				Error:   "连接不存在",
			})
			return
		}
		if errors.Is(err, service.ErrConnectionAccessDenied) {
			c.JSON(http.StatusForbidden, model.APIResponse{
				Success: false,
				Error:   "没有权限",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, model.APIResponse{
			Success: false,
			Error:   "更新连接失败",
		})
		return
	}

	c.JSON(http.StatusOK, model.APIResponse{
		Success: true,
		Data:    toResponse(conn),
	})
}

// DeleteUserConnection deletes a connection for the authenticated user
// DELETE /api/connections/:id
// Requirements: 6.2, 6.3
func DeleteUserConnection(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, model.APIResponse{
			Success: false,
			Error:   "认证已过期",
		})
		return
	}

	connID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Error:   "无效的连接ID",
		})
		return
	}

	connService := service.NewConnectionService()
	err = connService.DeleteUserConnection(userID, uint(connID))
	if err != nil {
		if errors.Is(err, service.ErrConnectionNotFound) {
			c.JSON(http.StatusNotFound, model.APIResponse{
				Success: false,
				Error:   "连接不存在",
			})
			return
		}
		if errors.Is(err, service.ErrConnectionAccessDenied) {
			c.JSON(http.StatusForbidden, model.APIResponse{
				Success: false,
				Error:   "没有权限",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, model.APIResponse{
			Success: false,
			Error:   "删除连接失败",
		})
		return
	}

	c.JSON(http.StatusOK, model.APIResponse{
		Success: true,
	})
}

// TestUserConnection tests a connection for the authenticated user
// POST /api/connections/:id/test
// Requirements: 6.2, 6.3
func TestUserConnection(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, model.APIResponse{
			Success: false,
			Error:   "认证已过期",
		})
		return
	}

	connID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Error:   "无效的连接ID",
		})
		return
	}

	connService := service.NewConnectionService()
	conn, err := connService.GetUserConnectionByID(userID, uint(connID))
	if err != nil {
		if errors.Is(err, service.ErrConnectionNotFound) {
			c.JSON(http.StatusNotFound, model.APIResponse{
				Success: false,
				Error:   "连接不存在",
			})
			return
		}
		if errors.Is(err, service.ErrConnectionAccessDenied) {
			c.JSON(http.StatusForbidden, model.APIResponse{
				Success: false,
				Error:   "没有权限",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, model.APIResponse{
			Success: false,
			Error:   "获取连接失败",
		})
		return
	}

	// Decrypt password for testing
	password, err := connService.GetDecryptedPassword(userID, uint(connID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.APIResponse{
			Success: false,
			Error:   "解密密码失败",
		})
		return
	}

	// Create a ConnectionConfig for testing
	config := model.ConnectionConfig{
		Host:     conn.Host,
		Port:     conn.Port,
		Password: password,
	}

	if err := service.TestConnection(config); err != nil {
		c.JSON(http.StatusOK, model.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.APIResponse{
		Success: true,
		Data:    "连接测试成功",
	})
}


// TestConnectionConfig tests a connection configuration without saving
// POST /api/connections/test
// This allows testing a new connection before creating it
func TestConnectionConfig(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, model.APIResponse{
			Success: false,
			Error:   "认证已过期",
		})
		return
	}

	var req UserConnectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Error:   "参数错误: " + err.Error(),
		})
		return
	}

	// Set defaults
	if req.Port == 0 {
		req.Port = 6379
	}

	// Create a ConnectionConfig for testing
	config := model.ConnectionConfig{
		Host:     req.Host,
		Port:     req.Port,
		Password: req.Password,
	}

	if err := service.TestConnection(config); err != nil {
		c.JSON(http.StatusOK, model.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.APIResponse{
		Success: true,
		Data:    "连接测试成功",
	})
}


// ConnectUserConnection connects to a Redis server for the authenticated user
// POST /api/connections/:id/connect
func ConnectUserConnection(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, model.APIResponse{
			Success: false,
			Error:   "认证已过期",
		})
		return
	}

	connID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Error:   "无效的连接ID",
		})
		return
	}

	connService := service.NewConnectionService()
	conn, err := connService.GetUserConnectionByID(userID, uint(connID))
	if err != nil {
		if errors.Is(err, service.ErrConnectionNotFound) {
			c.JSON(http.StatusNotFound, model.APIResponse{
				Success: false,
				Error:   "连接不存在",
			})
			return
		}
		if errors.Is(err, service.ErrConnectionAccessDenied) {
			c.JSON(http.StatusForbidden, model.APIResponse{
				Success: false,
				Error:   "没有权限",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, model.APIResponse{
			Success: false,
			Error:   "获取连接失败: " + err.Error(),
		})
		return
	}

	// Decrypt password
	password, err := connService.GetDecryptedPassword(userID, uint(connID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.APIResponse{
			Success: false,
			Error:   "解密密码失败: " + err.Error(),
		})
		return
	}

	// Create connection config
	config := model.ConnectionConfig{
		ID:       c.Param("id"),
		Name:     conn.Name,
		Host:     conn.Host,
		Port:     conn.Port,
		Password: password,
	}

	// Connect to Redis
	if err := service.GetManager().Connect(config); err != nil {
		c.JSON(http.StatusOK, model.APIResponse{
			Success: false,
			Error:   "连接Redis失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.APIResponse{
		Success: true,
	})
}

// DisconnectUserConnection disconnects from a Redis server
// POST /api/connections/:id/disconnect
func DisconnectUserConnection(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, model.APIResponse{
			Success: false,
			Error:   "认证已过期",
		})
		return
	}

	connID := c.Param("id")
	service.GetManager().Disconnect(connID)

	c.JSON(http.StatusOK, model.APIResponse{
		Success: true,
	})
}

// GetUserConnectionDatabases returns database list for a connection
// GET /api/connections/:id/databases
func GetUserConnectionDatabases(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, model.APIResponse{
			Success: false,
			Error:   "认证已过期",
		})
		return
	}

	connID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Error:   "无效的连接ID",
		})
		return
	}

	// Verify ownership
	connService := service.NewConnectionService()
	_, err = connService.GetUserConnectionByID(userID, uint(connID))
	if err != nil {
		if errors.Is(err, service.ErrConnectionNotFound) {
			c.JSON(http.StatusNotFound, model.APIResponse{
				Success: false,
				Error:   "连接不存在",
			})
			return
		}
		if errors.Is(err, service.ErrConnectionAccessDenied) {
			c.JSON(http.StatusForbidden, model.APIResponse{
				Success: false,
				Error:   "没有权限",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, model.APIResponse{
			Success: false,
			Error:   "获取连接失败",
		})
		return
	}

	// Get databases from Redis service
	conn, err := service.GetManager().GetConnection(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.APIResponse{
			Success: false,
			Error:   "连接未打开",
		})
		return
	}

	databases, err := conn.GetDatabases(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.APIResponse{
		Success: true,
		Data:    databases,
	})
}

// UpdateUserDatabaseAlias updates database alias
// PUT /api/connections/:id/databases/:db
func UpdateUserDatabaseAlias(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, model.APIResponse{
			Success: false,
			Error:   "认证已过期",
		})
		return
	}

	connID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Error:   "无效的连接ID",
		})
		return
	}

	// Verify ownership
	connService := service.NewConnectionService()
	_, err = connService.GetUserConnectionByID(userID, uint(connID))
	if err != nil {
		if errors.Is(err, service.ErrConnectionNotFound) {
			c.JSON(http.StatusNotFound, model.APIResponse{
				Success: false,
				Error:   "连接不存在",
			})
			return
		}
		if errors.Is(err, service.ErrConnectionAccessDenied) {
			c.JSON(http.StatusForbidden, model.APIResponse{
				Success: false,
				Error:   "没有权限",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, model.APIResponse{
			Success: false,
			Error:   "获取连接失败",
		})
		return
	}

	var req struct {
		Alias string `json:"alias"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Error:   "参数错误",
		})
		return
	}

	// For now, just return success (alias storage can be implemented later)
	c.JSON(http.StatusOK, model.APIResponse{
		Success: true,
	})
}

// SelectUserDatabase selects a database
// POST /api/connections/:id/databases/:db/select
func SelectUserDatabase(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, model.APIResponse{
			Success: false,
			Error:   "认证已过期",
		})
		return
	}

	connID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Error:   "无效的连接ID",
		})
		return
	}

	dbIndex, err := strconv.Atoi(c.Param("db"))
	if err != nil {
		c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Error:   "无效的数据库索引",
		})
		return
	}

	// Verify ownership
	connService := service.NewConnectionService()
	_, err = connService.GetUserConnectionByID(userID, uint(connID))
	if err != nil {
		if errors.Is(err, service.ErrConnectionNotFound) {
			c.JSON(http.StatusNotFound, model.APIResponse{
				Success: false,
				Error:   "连接不存在",
			})
			return
		}
		if errors.Is(err, service.ErrConnectionAccessDenied) {
			c.JSON(http.StatusForbidden, model.APIResponse{
				Success: false,
				Error:   "没有权限",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, model.APIResponse{
			Success: false,
			Error:   "获取连接失败",
		})
		return
	}

	// Select database
	conn, err := service.GetManager().GetConnection(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.APIResponse{
			Success: false,
			Error:   "连接未打开",
		})
		return
	}

	if err := conn.SelectDB(c.Request.Context(), dbIndex); err != nil {
		c.JSON(http.StatusInternalServerError, model.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.APIResponse{
		Success: true,
	})
}
