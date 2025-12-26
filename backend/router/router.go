package router

import (
	"github.com/gin-gonic/gin"
	"redis-web-manager/handler"
	"redis-web-manager/middleware"
)

// Setup configures and returns the Gin router
func Setup() *gin.Engine {
	r := gin.Default()

	// Apply CORS middleware
	r.Use(middleware.CORS())

	// API routes
	api := r.Group("/api")
	{
		// Authentication routes (public - no JWT required)
		auth := api.Group("/auth")
		{
			auth.POST("/login", handler.Login)
			auth.POST("/register", handler.Register)
			auth.POST("/verify-2fa", handler.Verify2FA)
			auth.POST("/logout", handler.Logout)
			auth.GET("/me", middleware.JWTAuth(), handler.GetCurrentUser)
		}

		// 2FA routes (protected - JWT required)
		twoFA := api.Group("/2fa")
		twoFA.Use(middleware.JWTAuth())
		{
			twoFA.POST("/setup", handler.Setup2FA)
			twoFA.POST("/verify", handler.Activate2FA)
			twoFA.POST("/disable", handler.Disable2FA)
		}

		// Legacy routes (keep for backward compatibility)
		api.POST("/ping", handler.Ping)
		api.POST("/execute", handler.Execute)
		api.POST("/info", handler.GetInfo)

		// Admin routes (protected - JWT + Admin role required)
		admin := api.Group("/admin")
		admin.Use(middleware.JWTAuth(), middleware.AdminOnly())
		{
			// User management
			admin.GET("/users", handler.ListUsers)
			admin.POST("/users", handler.CreateUserByAdmin)
			admin.GET("/users/:id", handler.GetUser)
			admin.PUT("/users/:id", handler.UpdateUserByAdmin)
			admin.DELETE("/users/:id", handler.DeleteUserByAdmin)
			admin.POST("/users/:id/reset-password", handler.ResetUserPasswordByAdmin)

			// Settings management (Task 10.3 - Requirements 8.1, 8.2, 8.3)
			admin.GET("/settings", handler.GetSettings)
			admin.PUT("/settings", handler.UpdateSetting)
		}

		// Connection groups (authenticated - with user isolation)
		groups := api.Group("/groups")
		groups.Use(middleware.JWTAuth())
		{
			groups.GET("", handler.ListGroups)
			groups.POST("", handler.CreateGroup)
			groups.PUT("/:id", handler.UpdateGroup)
			groups.DELETE("/:id", handler.DeleteGroup)
		}

		// User connections (authenticated - with user isolation)
		// These routes require JWT authentication and enforce user data isolation
		// Requirements: 6.1, 6.2, 6.3
		connections := api.Group("/connections")
		connections.Use(middleware.JWTAuth())
		{
			connections.GET("", handler.ListUserConnections)
			connections.POST("", handler.CreateUserConnection)
			connections.POST("/test", handler.TestConnectionConfig) // Test connection without saving
			connections.GET("/:id", handler.GetUserConnection)
			connections.PUT("/:id", handler.UpdateUserConnection)
			connections.DELETE("/:id", handler.DeleteUserConnection)
			connections.POST("/:id/test", handler.TestUserConnection)
			connections.POST("/:id/connect", handler.ConnectUserConnection)
			connections.POST("/:id/disconnect", handler.DisconnectUserConnection)
			connections.GET("/:id/databases", handler.GetUserConnectionDatabases)
			connections.PUT("/:id/databases/:db", handler.UpdateUserDatabaseAlias)
			connections.POST("/:id/databases/:db/select", handler.SelectUserDatabase)

			// Key management (with user ownership validation)
			connections.GET("/:id/keys", handler.ListKeys)
			connections.POST("/:id/keys/scan", handler.ScanKeys)
			connections.POST("/:id/keys", handler.CreateKey)
			connections.DELETE("/:id/keys", handler.DeleteKeys)
			connections.GET("/:id/keys/:key", handler.GetKey)
			connections.PUT("/:id/keys/:key", handler.UpdateKey)
			connections.GET("/:id/keys/:key/ttl", handler.GetKeyTTL)
			connections.PUT("/:id/keys/:key/ttl", handler.SetKeyTTL)
			connections.DELETE("/:id/keys/:key/ttl", handler.RemoveKeyTTL)
			connections.POST("/:id/keys/:key/rename", handler.RenameKey)

			// List operations
			connections.POST("/:id/keys/:key/list/push", handler.ListPush)
			connections.PUT("/:id/keys/:key/list/:index", handler.ListSet)
			connections.DELETE("/:id/keys/:key/list/:index", handler.ListRemove)

			// Hash operations
			connections.PUT("/:id/keys/:key/hash/:field", handler.HashSet)
			connections.DELETE("/:id/keys/:key/hash", handler.HashDelete)

			// Set operations
			connections.POST("/:id/keys/:key/set", handler.SetAdd)
			connections.DELETE("/:id/keys/:key/set", handler.SetRemove)

			// ZSet operations
			connections.POST("/:id/keys/:key/zset", handler.ZSetAdd)
			connections.DELETE("/:id/keys/:key/zset", handler.ZSetRemove)

			// Command execution
			connections.POST("/:id/command", handler.ExecuteCommand)
		}

		// Pub/Sub routes (authenticated - with user isolation)
		// Requirements: 2.2, 2.4, 2.5, 2.6, 2.7
		pubsub := api.Group("/pubsub")
		pubsub.Use(middleware.JWTAuth())
		{
			pubsub.GET("/ws", handler.HandlePubSubWebSocket)
			pubsub.POST("/publish", handler.PublishMessage)
		}

		// Legacy connection management (no auth - for backward compatibility)
		legacyConnections := api.Group("/legacy/connections")
		{
			legacyConnections.GET("", handler.ListConnections)
			legacyConnections.POST("", handler.CreateConnection)
			legacyConnections.POST("/test", handler.TestConnection)
			legacyConnections.PUT("/:id", handler.UpdateConnection)
			legacyConnections.DELETE("/:id", handler.DeleteConnection)
			legacyConnections.POST("/:id/connect", handler.Connect)
			legacyConnections.POST("/:id/disconnect", handler.Disconnect)

			// Database management
			legacyConnections.GET("/:id/databases", handler.GetDatabases)
			legacyConnections.PUT("/:id/databases/:db", handler.UpdateDatabaseAlias)
			legacyConnections.POST("/:id/databases/:db/select", handler.SelectDatabase)

			// Key management
			legacyConnections.GET("/:id/keys", handler.ListKeys)
			legacyConnections.POST("/:id/keys/scan", handler.ScanKeys)
			legacyConnections.POST("/:id/keys", handler.CreateKey)
			legacyConnections.DELETE("/:id/keys", handler.DeleteKeys)
			legacyConnections.GET("/:id/keys/:key", handler.GetKey)
			legacyConnections.PUT("/:id/keys/:key", handler.UpdateKey)
			legacyConnections.GET("/:id/keys/:key/ttl", handler.GetKeyTTL)
			legacyConnections.PUT("/:id/keys/:key/ttl", handler.SetKeyTTL)
			legacyConnections.DELETE("/:id/keys/:key/ttl", handler.RemoveKeyTTL)
			legacyConnections.POST("/:id/keys/:key/rename", handler.RenameKey)

			// List operations
			legacyConnections.POST("/:id/keys/:key/list/push", handler.ListPush)
			legacyConnections.PUT("/:id/keys/:key/list/:index", handler.ListSet)
			legacyConnections.DELETE("/:id/keys/:key/list/:index", handler.ListRemove)

			// Hash operations
			legacyConnections.PUT("/:id/keys/:key/hash/:field", handler.HashSet)
			legacyConnections.DELETE("/:id/keys/:key/hash", handler.HashDelete)

			// Set operations
			legacyConnections.POST("/:id/keys/:key/set", handler.SetAdd)
			legacyConnections.DELETE("/:id/keys/:key/set", handler.SetRemove)

			// ZSet operations
			legacyConnections.POST("/:id/keys/:key/zset", handler.ZSetAdd)
			legacyConnections.DELETE("/:id/keys/:key/zset", handler.ZSetRemove)

			// Command execution
			legacyConnections.POST("/:id/command", handler.ExecuteCommand)
		}
	}

	return r
}
