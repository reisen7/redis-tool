package service

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"redis-web-manager/config"
	"redis-web-manager/model"
)

var db *gorm.DB

// InitDatabase initializes SQLite connection and creates tables
func InitDatabase() error {
	cfg := config.Get()
	dbPath := cfg.GetDBPath()

	// Ensure directory exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create database directory: %w", err)
	}

	// Connect to SQLite database
	var err error
	db, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// Auto migrate tables (including new V2 models)
	if err := db.AutoMigrate(
		&model.ConnectionGroup{},
		&model.ConnectionConfig{},
		&model.DatabaseAlias{},
		// V2 models
		&model.User{},
		&model.UserConnection{},
		&model.SystemSetting{},
		&model.RecoveryCode{},
	); err != nil {
		return fmt.Errorf("failed to migrate tables: %w", err)
	}

	// Initialize encryption key
	encKey, err := cfg.GetEncryptionKey()
	if err != nil {
		return fmt.Errorf("failed to get encryption key: %w", err)
	}
	if err := SetEncryptionKey(encKey); err != nil {
		return fmt.Errorf("failed to set encryption key: %w", err)
	}

	// Initialize default settings
	if err := initDefaultSettings(); err != nil {
		return fmt.Errorf("failed to initialize default settings: %w", err)
	}

	// Create default admin user if no users exist
	if err := createDefaultAdminUser(cfg); err != nil {
		return fmt.Errorf("failed to create default admin user: %w", err)
	}

	log.Printf("Database initialized successfully (path: %s)", dbPath)
	return nil
}

// GetDB returns the database instance
func GetDB() *gorm.DB {
	return db
}

// --- Connection Group Operations ---

// CreateGroup creates a new connection group for a user
func CreateGroup(userID uint, group *model.ConnectionGroup) error {
	group.UserID = userID
	return db.Create(group).Error
}

// GetAllGroups returns all connection groups (legacy - no user filter)
func GetAllGroups() ([]model.ConnectionGroup, error) {
	var groups []model.ConnectionGroup
	err := db.Order("sort_order ASC, created_at ASC").Find(&groups).Error
	return groups, err
}

// GetUserGroups returns all connection groups for a specific user
func GetUserGroups(userID uint) ([]model.ConnectionGroup, error) {
	var groups []model.ConnectionGroup
	err := db.Where("user_id = ?", userID).Order("sort_order ASC, created_at ASC").Find(&groups).Error
	return groups, err
}

// GetGroupByID returns a group by ID
func GetGroupByID(id string) (*model.ConnectionGroup, error) {
	var group model.ConnectionGroup
	err := db.First(&group, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &group, nil
}

// GetUserGroupByID returns a group by ID with user ownership validation
func GetUserGroupByID(userID uint, id string) (*model.ConnectionGroup, error) {
	var group model.ConnectionGroup
	err := db.Where("id = ? AND user_id = ?", id, userID).First(&group).Error
	if err != nil {
		return nil, err
	}
	return &group, nil
}

// UpdateGroup updates a connection group
func UpdateGroup(group *model.ConnectionGroup) error {
	return db.Save(group).Error
}

// UpdateUserGroup updates a connection group with user ownership validation
func UpdateUserGroup(userID uint, group *model.ConnectionGroup) error {
	// Verify ownership
	existing, err := GetUserGroupByID(userID, group.ID)
	if err != nil {
		return err
	}
	// Preserve user ID
	group.UserID = existing.UserID
	return db.Save(group).Error
}

// DeleteGroup deletes a connection group
func DeleteGroup(id string) error {
	// Set connections in this group to ungrouped
	db.Model(&model.ConnectionConfig{}).Where("group_id = ?", id).Update("group_id", "")
	return db.Delete(&model.ConnectionGroup{}, "id = ?", id).Error
}

// DeleteUserGroup deletes a connection group with user ownership validation
func DeleteUserGroup(userID uint, id string) error {
	// Verify ownership
	_, err := GetUserGroupByID(userID, id)
	if err != nil {
		return err
	}
	// Set user connections in this group to ungrouped
	db.Model(&model.UserConnection{}).Where("group_id = ? AND user_id = ?", id, userID).Update("group_id", "")
	return db.Delete(&model.ConnectionGroup{}, "id = ? AND user_id = ?", id, userID).Error
}

// --- Connection Operations ---

// CreateConnection creates a new connection
func CreateConnection(conn *model.ConnectionConfig) error {
	// Serialize JSON fields
	if err := serializeConnectionJSON(conn); err != nil {
		return err
	}
	return db.Create(conn).Error
}

// GetAllConnections returns all connections
func GetAllConnections() ([]model.ConnectionConfig, error) {
	var connections []model.ConnectionConfig
	err := db.Order("sort_order ASC, created_at ASC").Find(&connections).Error
	if err != nil {
		return nil, err
	}

	// Deserialize JSON fields
	for i := range connections {
		if err := deserializeConnectionJSON(&connections[i]); err != nil {
			log.Printf("Failed to deserialize connection %s: %v", connections[i].ID, err)
		}
	}

	return connections, nil
}

// GetConnectionByID returns a connection by ID
func GetConnectionByID(id string) (*model.ConnectionConfig, error) {
	var conn model.ConnectionConfig
	err := db.First(&conn, "id = ?", id).Error
	if err != nil {
		return nil, err
	}

	if err := deserializeConnectionJSON(&conn); err != nil {
		return nil, err
	}

	return &conn, nil
}

// UpdateConnection updates a connection
func UpdateConnection(conn *model.ConnectionConfig) error {
	if err := serializeConnectionJSON(conn); err != nil {
		return err
	}
	return db.Save(conn).Error
}

// DeleteConnection deletes a connection
func DeleteConnection(id string) error {
	// Delete associated aliases
	db.Delete(&model.DatabaseAlias{}, "connection_id = ?", id)
	return db.Delete(&model.ConnectionConfig{}, "id = ?", id).Error
}

// --- Database Alias Operations ---

// GetDatabaseAliases returns all aliases for a connection
func GetDatabaseAliases(connectionID string) (map[int]string, error) {
	var aliases []model.DatabaseAlias
	err := db.Where("connection_id = ?", connectionID).Find(&aliases).Error
	if err != nil {
		return nil, err
	}

	result := make(map[int]string)
	for _, alias := range aliases {
		result[alias.DBIndex] = alias.Alias
	}
	return result, nil
}

// SetDatabaseAlias sets or updates a database alias
func SetDatabaseAlias(connectionID string, dbIndex int, alias string) error {
	var existing model.DatabaseAlias
	err := db.Where("connection_id = ? AND db_index = ?", connectionID, dbIndex).First(&existing).Error

	if err == gorm.ErrRecordNotFound {
		// Create new
		return db.Create(&model.DatabaseAlias{
			ConnectionID: connectionID,
			DBIndex:      dbIndex,
			Alias:        alias,
		}).Error
	} else if err != nil {
		return err
	}

	// Update existing
	existing.Alias = alias
	return db.Save(&existing).Error
}

// --- Helper Functions ---

// initDefaultSettings initializes default system settings if they don't exist
func initDefaultSettings() error {
	defaults := model.DefaultSettings()
	for _, setting := range defaults {
		var existing model.SystemSetting
		err := db.Where("key = ?", setting.Key).First(&existing).Error
		if err == gorm.ErrRecordNotFound {
			if err := db.Create(&setting).Error; err != nil {
				return err
			}
			log.Printf("Created default setting: %s = %s", setting.Key, setting.Value)
		}
	}
	return nil
}

// createDefaultAdminUser creates the default admin user if no users exist
func createDefaultAdminUser(cfg *config.Config) error {
	var count int64
	if err := db.Model(&model.User{}).Count(&count).Error; err != nil {
		return err
	}

	if count > 0 {
		return nil // Users already exist
	}

	// Hash the default admin password
	hashedPassword, err := bcrypt.GenerateFromPassword(
		[]byte(cfg.Security.DefaultAdminPass),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return fmt.Errorf("failed to hash default admin password: %w", err)
	}

	adminUser := &model.User{
		Username:     cfg.Security.DefaultAdminUser,
		PasswordHash: string(hashedPassword),
		Role:         model.RoleAdmin,
		Enabled:      true,
	}

	if err := db.Create(adminUser).Error; err != nil {
		return err
	}

	log.Printf("Created default admin user: %s", adminUser.Username)
	return nil
}

func serializeConnectionJSON(conn *model.ConnectionConfig) error {
	// Serialize SentinelNodes
	if len(conn.SentinelNodes) > 0 {
		data, err := json.Marshal(conn.SentinelNodes)
		if err != nil {
			return err
		}
		conn.SentinelNodesJSON = string(data)
	} else {
		conn.SentinelNodesJSON = "[]"
	}

	// Serialize ClusterNodes
	if len(conn.ClusterNodes) > 0 {
		data, err := json.Marshal(conn.ClusterNodes)
		if err != nil {
			return err
		}
		conn.ClusterNodesJSON = string(data)
	} else {
		conn.ClusterNodesJSON = "[]"
	}

	// Serialize SSH
	sshData, err := json.Marshal(conn.SSH)
	if err != nil {
		return err
	}
	conn.SSHJSON = string(sshData)

	// Serialize TLS
	tlsData, err := json.Marshal(conn.TLS)
	if err != nil {
		return err
	}
	conn.TLSJSON = string(tlsData)

	return nil
}

func deserializeConnectionJSON(conn *model.ConnectionConfig) error {
	// Deserialize SentinelNodes
	if conn.SentinelNodesJSON != "" {
		if err := json.Unmarshal([]byte(conn.SentinelNodesJSON), &conn.SentinelNodes); err != nil {
			conn.SentinelNodes = []string{}
		}
	} else {
		conn.SentinelNodes = []string{}
	}

	// Deserialize ClusterNodes
	if conn.ClusterNodesJSON != "" {
		if err := json.Unmarshal([]byte(conn.ClusterNodesJSON), &conn.ClusterNodes); err != nil {
			conn.ClusterNodes = []string{}
		}
	} else {
		conn.ClusterNodes = []string{}
	}

	// Deserialize SSH
	if conn.SSHJSON != "" {
		if err := json.Unmarshal([]byte(conn.SSHJSON), &conn.SSH); err != nil {
			conn.SSH = model.SSHConfig{}
		}
	}

	// Deserialize TLS
	if conn.TLSJSON != "" {
		if err := json.Unmarshal([]byte(conn.TLSJSON), &conn.TLS); err != nil {
			conn.TLS = model.TLSConfig{}
		}
	}

	return nil
}
