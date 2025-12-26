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

// ============================================================================
// Request/Response Types
// ============================================================================

// CreateUserRequest represents the request to create a new user
type CreateUserRequest struct {
	Username string         `json:"username" binding:"required,min=3,max=50"`
	Password string         `json:"password" binding:"required,min=8"`
	Role     model.UserRole `json:"role" binding:"required,oneof=admin user"`
}

// UpdateUserRequest represents the request to update a user
type UpdateUserRequest struct {
	Enabled *bool          `json:"enabled"`
	Role    model.UserRole `json:"role"`
}

// ============================================================================
// Admin User Management Handlers (Task 8.2 - Requirements 7.1-7.6)
// ============================================================================

// ListUsers returns all users
// GET /api/admin/users
// Requirements: 7.1
func ListUsers(c *gin.Context) {
	users, err := service.GetAllUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.APIResponse{
			Success: false,
			Error:   "服务器内部错误",
		})
		return
	}

	c.JSON(http.StatusOK, model.APIResponse{
		Success: true,
		Data:    users,
	})
}

// CreateUserByAdmin creates a new user account
// POST /api/admin/users
// Requirements: 7.2
func CreateUserByAdmin(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Error:   "参数错误",
		})
		return
	}

	// Validate username
	if valid, msg := ValidateUsername(req.Username); !valid {
		c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Error:   msg,
		})
		return
	}

	// Validate password strength
	if valid, msg := ValidatePassword(req.Password); !valid {
		c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Error:   msg,
		})
		return
	}

	// Create user
	user, err := service.CreateUser(req.Username, req.Password, req.Role)
	if err != nil {
		if errors.Is(err, service.ErrUsernameExists) {
			c.JSON(http.StatusConflict, model.APIResponse{
				Success: false,
				Error:   "用户名已存在",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, model.APIResponse{
			Success: false,
			Error:   "服务器内部错误",
		})
		return
	}

	c.JSON(http.StatusOK, model.APIResponse{
		Success: true,
		Data:    user,
	})
}

// GetUser returns a specific user by ID
// GET /api/admin/users/:id
func GetUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Error:   "无效的用户ID",
		})
		return
	}

	user, err := service.GetUserByID(uint(id))
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, model.APIResponse{
				Success: false,
				Error:   "用户不存在",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, model.APIResponse{
			Success: false,
			Error:   "服务器内部错误",
		})
		return
	}

	c.JSON(http.StatusOK, model.APIResponse{
		Success: true,
		Data:    user,
	})
}

// UpdateUserByAdmin updates a user's information
// PUT /api/admin/users/:id
// Requirements: 7.3, 7.4
func UpdateUserByAdmin(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Error:   "无效的用户ID",
		})
		return
	}

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Error:   "参数错误",
		})
		return
	}

	// Get current user ID from context
	currentUserID := middleware.GetUserID(c)

	// Build updates map
	updates := make(map[string]interface{})
	if req.Enabled != nil {
		updates["enabled"] = *req.Enabled
	}
	if req.Role != "" {
		updates["role"] = req.Role
	}

	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Error:   "没有要更新的字段",
		})
		return
	}

	user, err := service.UpdateUser(uint(id), updates, currentUserID)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, model.APIResponse{
				Success: false,
				Error:   "用户不存在",
			})
			return
		}
		if errors.Is(err, service.ErrCannotDisableSelf) {
			c.JSON(http.StatusBadRequest, model.APIResponse{
				Success: false,
				Error:   "不能禁用自己的账户",
			})
			return
		}
		if errors.Is(err, service.ErrLastAdmin) {
			c.JSON(http.StatusBadRequest, model.APIResponse{
				Success: false,
				Error:   "不能移除最后一个管理员",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, model.APIResponse{
			Success: false,
			Error:   "服务器内部错误",
		})
		return
	}

	c.JSON(http.StatusOK, model.APIResponse{
		Success: true,
		Data:    user,
	})
}


// DeleteUserByAdmin deletes a user and all associated data
// DELETE /api/admin/users/:id
// Requirements: 7.6
func DeleteUserByAdmin(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Error:   "无效的用户ID",
		})
		return
	}

	// Get current user ID from context
	currentUserID := middleware.GetUserID(c)

	err = service.DeleteUser(uint(id), currentUserID)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, model.APIResponse{
				Success: false,
				Error:   "用户不存在",
			})
			return
		}
		if errors.Is(err, service.ErrCannotDeleteSelf) {
			c.JSON(http.StatusBadRequest, model.APIResponse{
				Success: false,
				Error:   "不能删除自己的账户",
			})
			return
		}
		if errors.Is(err, service.ErrLastAdmin) {
			c.JSON(http.StatusBadRequest, model.APIResponse{
				Success: false,
				Error:   "不能删除最后一个管理员",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, model.APIResponse{
			Success: false,
			Error:   "服务器内部错误",
		})
		return
	}

	c.JSON(http.StatusOK, model.APIResponse{
		Success: true,
		Data: gin.H{
			"message": "用户已删除",
		},
	})
}

// ResetUserPasswordByAdmin resets a user's password to a temporary one
// POST /api/admin/users/:id/reset-password
// Requirements: 7.5
func ResetUserPasswordByAdmin(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Error:   "无效的用户ID",
		})
		return
	}

	tempPassword, err := service.ResetUserPassword(uint(id))
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, model.APIResponse{
				Success: false,
				Error:   "用户不存在",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, model.APIResponse{
			Success: false,
			Error:   "服务器内部错误",
		})
		return
	}

	c.JSON(http.StatusOK, model.APIResponse{
		Success: true,
		Data: gin.H{
			"tempPassword": tempPassword,
		},
	})
}


// ============================================================================
// Admin Settings Handlers (Task 10.3 - Requirements 8.1-8.3)
// ============================================================================

// UpdateSettingRequest represents the request to update a system setting
type UpdateSettingRequest struct {
	Key   string `json:"key" binding:"required"`
	Value string `json:"value" binding:"required"`
}

// GetSettings returns all system settings
// GET /api/admin/settings
// Requirements: 8.2
func GetSettings(c *gin.Context) {
	settings, err := service.GetSettingsMap()
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.APIResponse{
			Success: false,
			Error:   "服务器内部错误",
		})
		return
	}

	c.JSON(http.StatusOK, model.APIResponse{
		Success: true,
		Data:    settings,
	})
}

// UpdateSetting updates a system setting
// PUT /api/admin/settings
// Requirements: 8.1, 8.3
func UpdateSetting(c *gin.Context) {
	var req UpdateSettingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Error:   "参数错误",
		})
		return
	}

	// Validate the setting key is a known setting
	validKeys := map[string]bool{
		model.SettingRegistrationEnabled: true,
	}

	if !validKeys[req.Key] {
		c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Error:   "无效的设置项",
		})
		return
	}

	// Update the setting
	if err := service.SetSetting(req.Key, req.Value); err != nil {
		c.JSON(http.StatusInternalServerError, model.APIResponse{
			Success: false,
			Error:   "服务器内部错误",
		})
		return
	}

	c.JSON(http.StatusOK, model.APIResponse{
		Success: true,
		Data: gin.H{
			"key":   req.Key,
			"value": req.Value,
		},
	})
}
