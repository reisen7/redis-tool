package handler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"redis-web-manager/middleware"
	"redis-web-manager/model"
	"redis-web-manager/service"
)

// --- Group Handlers ---

// ListGroups returns all connection groups for the authenticated user
func ListGroups(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, model.APIResponse{
			Success: false,
			Error:   "认证已过期",
		})
		return
	}

	groups, err := service.GetUserGroups(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.APIResponse{
		Success: true,
		Data:    groups,
	})
}

// CreateGroup creates a new connection group for the authenticated user
func CreateGroup(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, model.APIResponse{
			Success: false,
			Error:   "认证已过期",
		})
		return
	}

	var group model.ConnectionGroup
	if err := c.ShouldBindJSON(&group); err != nil {
		c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	group.ID = uuid.New().String()
	group.CreatedAt = time.Now()
	group.UpdatedAt = time.Now()

	if err := service.CreateGroup(userID, &group); err != nil {
		c.JSON(http.StatusInternalServerError, model.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.APIResponse{
		Success: true,
		Data:    group,
	})
}

// UpdateGroup updates a connection group for the authenticated user
func UpdateGroup(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, model.APIResponse{
			Success: false,
			Error:   "认证已过期",
		})
		return
	}

	id := c.Param("id")

	var group model.ConnectionGroup
	if err := c.ShouldBindJSON(&group); err != nil {
		c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	existing, err := service.GetUserGroupByID(userID, id)
	if err != nil {
		c.JSON(http.StatusNotFound, model.APIResponse{
			Success: false,
			Error:   "分组不存在或无权限",
		})
		return
	}

	existing.Name = group.Name
	existing.Icon = group.Icon
	existing.Color = group.Color
	existing.SortOrder = group.SortOrder
	existing.UpdatedAt = time.Now()

	if err := service.UpdateUserGroup(userID, existing); err != nil {
		c.JSON(http.StatusInternalServerError, model.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.APIResponse{
		Success: true,
		Data:    existing,
	})
}

// DeleteGroup deletes a connection group for the authenticated user
func DeleteGroup(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, model.APIResponse{
			Success: false,
			Error:   "认证已过期",
		})
		return
	}

	id := c.Param("id")

	if err := service.DeleteUserGroup(userID, id); err != nil {
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

// --- Connection Handlers ---

// ListConnections returns all saved connections
func ListConnections(c *gin.Context) {
	connections, err := service.GetAllConnections()
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	// Don't expose passwords in list
	for i := range connections {
		connections[i].Password = ""
		connections[i].SSH.Password = ""
		connections[i].SSH.PrivateKey = ""
		connections[i].SSH.PrivateKeyPassphrase = ""
		connections[i].TLS.ClientKey = ""
	}

	c.JSON(http.StatusOK, model.APIResponse{
		Success: true,
		Data:    connections,
	})
}

// CreateConnection creates a new connection configuration
func CreateConnection(c *gin.Context) {
	var config model.ConnectionConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	config.ID = uuid.New().String()
	config.CreatedAt = time.Now()
	config.UpdatedAt = time.Now()

	if err := service.CreateConnection(&config); err != nil {
		c.JSON(http.StatusInternalServerError, model.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	// Return without sensitive data
	configCopy := config
	configCopy.Password = ""
	configCopy.SSH.Password = ""
	configCopy.SSH.PrivateKey = ""

	c.JSON(http.StatusOK, model.APIResponse{
		Success: true,
		Data:    configCopy,
	})
}

// UpdateConnection updates an existing connection configuration
func UpdateConnection(c *gin.Context) {
	id := c.Param("id")

	var config model.ConnectionConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	existing, err := service.GetConnectionByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, model.APIResponse{
			Success: false,
			Error:   "connection not found",
		})
		return
	}

	// Preserve passwords if not provided
	if config.Password == "" {
		config.Password = existing.Password
	}
	if config.SSH.Password == "" {
		config.SSH.Password = existing.SSH.Password
	}
	if config.SSH.PrivateKey == "" {
		config.SSH.PrivateKey = existing.SSH.PrivateKey
	}
	if config.SSH.PrivateKeyPassphrase == "" {
		config.SSH.PrivateKeyPassphrase = existing.SSH.PrivateKeyPassphrase
	}
	if config.TLS.ClientKey == "" {
		config.TLS.ClientKey = existing.TLS.ClientKey
	}

	config.ID = id
	config.CreatedAt = existing.CreatedAt
	config.UpdatedAt = time.Now()

	if err := service.UpdateConnection(&config); err != nil {
		c.JSON(http.StatusInternalServerError, model.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.APIResponse{
		Success: true,
		Data:    config,
	})
}

// DeleteConnection deletes a connection configuration
func DeleteConnection(c *gin.Context) {
	id := c.Param("id")

	if err := service.DeleteConnection(id); err != nil {
		c.JSON(http.StatusInternalServerError, model.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	// Also disconnect if connected
	service.GetManager().Disconnect(id)

	c.JSON(http.StatusOK, model.APIResponse{
		Success: true,
	})
}

// TestConnection tests a connection without saving
func TestConnection(c *gin.Context) {
	var config model.ConnectionConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	// If testing existing connection, get stored credentials
	if config.ID != "" {
		existing, err := service.GetConnectionByID(config.ID)
		if err == nil {
			if config.Password == "" {
				config.Password = existing.Password
			}
			if config.SSH.Password == "" {
				config.SSH.Password = existing.SSH.Password
			}
			if config.SSH.PrivateKey == "" {
				config.SSH.PrivateKey = existing.SSH.PrivateKey
			}
		}
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
	})
}

// Connect establishes a connection
func Connect(c *gin.Context) {
	id := c.Param("id")

	config, err := service.GetConnectionByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, model.APIResponse{
			Success: false,
			Error:   "connection not found",
		})
		return
	}

	if err := service.GetManager().Connect(*config); err != nil {
		c.JSON(http.StatusOK, model.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.APIResponse{
		Success: true,
	})
}

// Disconnect closes a connection
func Disconnect(c *gin.Context) {
	id := c.Param("id")
	service.GetManager().Disconnect(id)

	c.JSON(http.StatusOK, model.APIResponse{
		Success: true,
	})
}

// GetDatabases returns database information
func GetDatabases(c *gin.Context) {
	id := c.Param("id")

	conn, err := service.GetManager().GetConnection(id)
	if err != nil {
		c.JSON(http.StatusNotFound, model.APIResponse{
			Success: false,
			Error:   err.Error(),
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

	// Load aliases from database
	aliases, _ := service.GetDatabaseAliases(id)
	for i := range databases {
		if alias, ok := aliases[databases[i].Index]; ok {
			databases[i].Alias = alias
		}
	}

	c.JSON(http.StatusOK, model.APIResponse{
		Success: true,
		Data:    databases,
	})
}

// UpdateDatabaseAlias updates database alias
func UpdateDatabaseAlias(c *gin.Context) {
	id := c.Param("id")
	dbStr := c.Param("db")

	var db int
	if _, err := fmt.Sscanf(dbStr, "%d", &db); err != nil {
		c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Error:   "invalid database index",
		})
		return
	}

	var req model.DatabaseAliasRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	// Save to MySQL
	if err := service.SetDatabaseAlias(id, db, req.Alias); err != nil {
		c.JSON(http.StatusInternalServerError, model.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	// Also update in-memory connection
	conn, err := service.GetManager().GetConnection(id)
	if err == nil {
		conn.SetDatabaseAlias(db, req.Alias)
	}

	c.JSON(http.StatusOK, model.APIResponse{
		Success: true,
	})
}

// SelectDatabase selects a database
func SelectDatabase(c *gin.Context) {
	id := c.Param("id")
	dbStr := c.Param("db")

	var db int
	if _, err := fmt.Sscanf(dbStr, "%d", &db); err != nil {
		c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Error:   "invalid database index",
		})
		return
	}

	conn, err := service.GetManager().GetConnection(id)
	if err != nil {
		c.JSON(http.StatusNotFound, model.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	if err := conn.SelectDB(c.Request.Context(), db); err != nil {
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

// GetConnectionConfig returns full connection config (for internal use)
func GetConnectionConfig(id string) (model.ConnectionConfig, bool) {
	config, err := service.GetConnectionByID(id)
	if err != nil {
		return model.ConnectionConfig{}, false
	}
	return *config, true
}
