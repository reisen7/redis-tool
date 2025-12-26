package service

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"

	"gorm.io/gorm"
	"redis-web-manager/model"
)

// User service errors
var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUsernameExists    = errors.New("username already exists")
	ErrCannotDeleteSelf  = errors.New("cannot delete your own account")
	ErrCannotDisableSelf = errors.New("cannot disable your own account")
	ErrLastAdmin         = errors.New("cannot remove the last admin user")
)

// ============================================================================
// User CRUD Operations (Task 8.1 - Requirements 7.2, 7.3, 7.4, 7.5, 7.6)
// ============================================================================

// CreateUser creates a new user account
// Requirements: 7.2
func CreateUser(username, password string, role model.UserRole) (*model.User, error) {
	db := GetDB()

	// Check if username already exists
	var existing model.User
	if err := db.Where("username = ?", username).First(&existing).Error; err == nil {
		return nil, ErrUsernameExists
	}

	// Hash password
	hashedPassword, err := HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &model.User{
		Username:     username,
		PasswordHash: hashedPassword,
		Role:         role,
		Enabled:      true,
	}

	if err := db.Create(user).Error; err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// GetUserByID retrieves a user by ID
func GetUserByID(id uint) (*model.User, error) {
	db := GetDB()
	var user model.User
	if err := db.First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return &user, nil
}

// GetUserByUsername retrieves a user by username
func GetUserByUsername(username string) (*model.User, error) {
	db := GetDB()
	var user model.User
	if err := db.Where("username = ?", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return &user, nil
}

// GetAllUsers retrieves all users
// Requirements: 7.1
func GetAllUsers() ([]model.User, error) {
	db := GetDB()
	var users []model.User
	if err := db.Order("created_at ASC").Find(&users).Error; err != nil {
		return nil, fmt.Errorf("failed to get users: %w", err)
	}
	return users, nil
}


// UpdateUser updates user information
// Requirements: 7.3, 7.4
func UpdateUser(id uint, updates map[string]interface{}, currentUserID uint) (*model.User, error) {
	db := GetDB()

	// Get existing user
	var user model.User
	if err := db.First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Check if trying to disable self
	if enabled, ok := updates["enabled"]; ok {
		if enabledBool, ok := enabled.(bool); ok && !enabledBool && id == currentUserID {
			return nil, ErrCannotDisableSelf
		}
	}

	// Check if trying to demote the last admin
	if role, ok := updates["role"]; ok {
		if roleStr, ok := role.(model.UserRole); ok && roleStr != model.RoleAdmin && user.Role == model.RoleAdmin {
			// Count remaining admins
			var adminCount int64
			if err := db.Model(&model.User{}).Where("role = ? AND id != ?", model.RoleAdmin, id).Count(&adminCount).Error; err != nil {
				return nil, fmt.Errorf("failed to count admins: %w", err)
			}
			if adminCount == 0 {
				return nil, ErrLastAdmin
			}
		}
	}

	// Apply updates
	if err := db.Model(&user).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	// Reload user to get updated values
	if err := db.First(&user, id).Error; err != nil {
		return nil, fmt.Errorf("failed to reload user: %w", err)
	}

	return &user, nil
}

// DeleteUser deletes a user and all associated data
// Requirements: 7.6
func DeleteUser(id uint, currentUserID uint) error {
	db := GetDB()

	// Cannot delete self
	if id == currentUserID {
		return ErrCannotDeleteSelf
	}

	// Get user to check if exists and role
	var user model.User
	if err := db.First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUserNotFound
		}
		return fmt.Errorf("failed to get user: %w", err)
	}

	// Check if this is the last admin
	if user.Role == model.RoleAdmin {
		var adminCount int64
		if err := db.Model(&model.User{}).Where("role = ? AND id != ?", model.RoleAdmin, id).Count(&adminCount).Error; err != nil {
			return fmt.Errorf("failed to count admins: %w", err)
		}
		if adminCount == 0 {
			return ErrLastAdmin
		}
	}

	// Start transaction to delete user and associated data
	tx := db.Begin()

	// Delete user's connections
	if err := tx.Where("user_id = ?", id).Delete(&model.UserConnection{}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete user connections: %w", err)
	}

	// Delete user's recovery codes
	if err := tx.Where("user_id = ?", id).Delete(&model.RecoveryCode{}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete recovery codes: %w", err)
	}

	// Delete user
	if err := tx.Delete(&user).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete user: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// ResetUserPassword generates a temporary password for a user
// Requirements: 7.5
func ResetUserPassword(id uint) (string, error) {
	db := GetDB()

	// Get user
	var user model.User
	if err := db.First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", ErrUserNotFound
		}
		return "", fmt.Errorf("failed to get user: %w", err)
	}

	// Generate temporary password (12 characters)
	tempPassword, err := generateTempPassword(12)
	if err != nil {
		return "", fmt.Errorf("failed to generate temp password: %w", err)
	}

	// Hash the temporary password
	hashedPassword, err := HashPassword(tempPassword)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	// Update user with new password and set must_change_pwd flag
	if err := db.Model(&user).Updates(map[string]interface{}{
		"password_hash":  hashedPassword,
		"must_change_pwd": true,
	}).Error; err != nil {
		return "", fmt.Errorf("failed to update password: %w", err)
	}

	return tempPassword, nil
}

// EnableUser enables a user account
// Requirements: 7.4
func EnableUser(id uint) (*model.User, error) {
	return UpdateUser(id, map[string]interface{}{"enabled": true}, 0)
}

// DisableUser disables a user account
// Requirements: 7.3
func DisableUser(id uint, currentUserID uint) (*model.User, error) {
	return UpdateUser(id, map[string]interface{}{"enabled": false}, currentUserID)
}

// generateTempPassword generates a random temporary password
func generateTempPassword(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	// Use URL-safe base64 encoding and take first 'length' characters
	encoded := base64.URLEncoding.EncodeToString(bytes)
	if len(encoded) > length {
		encoded = encoded[:length]
	}
	return encoded, nil
}
