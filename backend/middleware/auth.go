package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"redis-web-manager/model"
	"redis-web-manager/service"
)

// Context keys for user information
const (
	// ContextKeyUserID is the key for user ID in gin context
	ContextKeyUserID = "userID"
	// ContextKeyUsername is the key for username in gin context
	ContextKeyUsername = "username"
	// ContextKeyUserRole is the key for user role in gin context
	ContextKeyUserRole = "userRole"
)

// JWTAuth returns a middleware that validates JWT tokens from Authorization header or query parameter
// Returns 401 Unauthorized for invalid or missing tokens
// Requirements: 3.7, 9.5
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		var tokenString string

		// First try to get token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			// Check Bearer prefix
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
				tokenString = parts[1]
			}
		}

		// If no token in header, try query parameter (for WebSocket connections)
		if tokenString == "" {
			tokenString = c.Query("token")
		}

		if tokenString == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "认证已过期",
			})
			return
		}

		// Validate token
		claims, err := service.ValidateToken(tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "认证已过期",
			})
			return
		}

		// Set user context
		c.Set(ContextKeyUserID, claims.UserID)
		c.Set(ContextKeyUsername, claims.Username)
		c.Set(ContextKeyUserRole, claims.Role)

		c.Next()
	}
}

// AdminOnly returns a middleware that checks if the user has admin role
// Returns 403 Forbidden for non-admin users
// Must be used after JWTAuth middleware
// Requirements: 7.7
func AdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user role from context (set by JWTAuth middleware)
		roleValue, exists := c.Get(ContextKeyUserRole)
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   "没有权限",
			})
			return
		}

		role, ok := roleValue.(model.UserRole)
		if !ok || role != model.RoleAdmin {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   "没有权限",
			})
			return
		}

		c.Next()
	}
}

// GetUserID extracts user ID from gin context
// Returns 0 if not found
func GetUserID(c *gin.Context) uint {
	if userID, exists := c.Get(ContextKeyUserID); exists {
		if id, ok := userID.(uint); ok {
			return id
		}
	}
	return 0
}

// GetUsername extracts username from gin context
// Returns empty string if not found
func GetUsername(c *gin.Context) string {
	if username, exists := c.Get(ContextKeyUsername); exists {
		if name, ok := username.(string); ok {
			return name
		}
	}
	return ""
}

// GetUserRole extracts user role from gin context
// Returns empty string if not found
func GetUserRole(c *gin.Context) model.UserRole {
	if role, exists := c.Get(ContextKeyUserRole); exists {
		if r, ok := role.(model.UserRole); ok {
			return r
		}
	}
	return ""
}
