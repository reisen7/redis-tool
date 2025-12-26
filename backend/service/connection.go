package service

import (
	"errors"
	"fmt"

	"gorm.io/gorm"
	"redis-web-manager/model"
)

var (
	// ErrConnectionNotFound is returned when a connection is not found
	ErrConnectionNotFound = errors.New("connection not found")
	// ErrConnectionAccessDenied is returned when user tries to access another user's connection
	ErrConnectionAccessDenied = errors.New("access denied")
)

// ConnectionService handles user connection operations with user isolation
type ConnectionService struct {
	db *gorm.DB
}

// NewConnectionService creates a new ConnectionService instance
func NewConnectionService() *ConnectionService {
	return &ConnectionService{db: GetDB()}
}

// CreateUserConnection creates a new connection for a user
// The password is encrypted using AES-256 before storage
// Requirements: 6.1, 10.4
func (s *ConnectionService) CreateUserConnection(userID uint, conn *model.UserConnection) error {
	conn.UserID = userID

	// Encrypt password if provided
	if conn.PasswordEnc != "" {
		encrypted, err := Encrypt(conn.PasswordEnc)
		if err != nil {
			return fmt.Errorf("failed to encrypt password: %w", err)
		}
		conn.PasswordEnc = encrypted
	}

	return s.db.Create(conn).Error
}

// GetUserConnections returns all connections owned by a specific user
// Requirements: 6.2
func (s *ConnectionService) GetUserConnections(userID uint) ([]model.UserConnection, error) {
	var connections []model.UserConnection
	err := s.db.Where("user_id = ?", userID).Order("created_at DESC").Find(&connections).Error
	return connections, err
}

// GetUserConnectionByID returns a connection by ID, validating user ownership
// Returns ErrConnectionNotFound if connection doesn't exist
// Returns ErrConnectionAccessDenied if user doesn't own the connection
// Requirements: 6.2, 6.3, 6.5
func (s *ConnectionService) GetUserConnectionByID(userID uint, connectionID uint) (*model.UserConnection, error) {
	var conn model.UserConnection
	err := s.db.First(&conn, connectionID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrConnectionNotFound
		}
		return nil, err
	}

	// Validate ownership
	if conn.UserID != userID {
		return nil, ErrConnectionAccessDenied
	}

	return &conn, nil
}

// UpdateUserConnection updates a connection, validating user ownership
// Requirements: 6.2, 6.3, 6.5
func (s *ConnectionService) UpdateUserConnection(userID uint, connectionID uint, updates *model.UserConnection) (*model.UserConnection, error) {
	// First verify ownership
	existing, err := s.GetUserConnectionByID(userID, connectionID)
	if err != nil {
		return nil, err
	}

	// Update fields
	existing.Name = updates.Name
	existing.GroupID = updates.GroupID
	existing.Host = updates.Host
	existing.Port = updates.Port
	existing.DB = updates.DB
	existing.Timeout = updates.Timeout

	// Handle password update - only encrypt if a new password is provided
	if updates.PasswordEnc != "" {
		encrypted, err := Encrypt(updates.PasswordEnc)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt password: %w", err)
		}
		existing.PasswordEnc = encrypted
	}

	if err := s.db.Save(existing).Error; err != nil {
		return nil, err
	}

	return existing, nil
}

// DeleteUserConnection deletes a connection, validating user ownership
// Requirements: 6.2, 6.3, 6.5
func (s *ConnectionService) DeleteUserConnection(userID uint, connectionID uint) error {
	// First verify ownership
	_, err := s.GetUserConnectionByID(userID, connectionID)
	if err != nil {
		return err
	}

	return s.db.Delete(&model.UserConnection{}, connectionID).Error
}

// GetDecryptedPassword returns the decrypted password for a connection
// Validates user ownership before returning
// Requirements: 10.4
func (s *ConnectionService) GetDecryptedPassword(userID uint, connectionID uint) (string, error) {
	conn, err := s.GetUserConnectionByID(userID, connectionID)
	if err != nil {
		return "", err
	}

	if conn.PasswordEnc == "" {
		return "", nil
	}

	// Try to decrypt - if it fails, assume it's stored in plaintext (legacy)
	decrypted, err := Decrypt(conn.PasswordEnc)
	if err != nil {
		// If decryption fails, return the stored value as-is (might be plaintext)
		return conn.PasswordEnc, nil
	}
	return decrypted, nil
}

// DeleteAllUserConnections deletes all connections for a user
// Used when deleting a user account
// Requirements: 6.4
func (s *ConnectionService) DeleteAllUserConnections(userID uint) error {
	return s.db.Where("user_id = ?", userID).Delete(&model.UserConnection{}).Error
}

// ConnectionExists checks if a connection exists and belongs to the user
// Requirements: 6.5
func (s *ConnectionService) ConnectionExists(userID uint, connectionID uint) (bool, error) {
	var count int64
	err := s.db.Model(&model.UserConnection{}).
		Where("id = ? AND user_id = ?", connectionID, userID).
		Count(&count).Error
	return count > 0, err
}
