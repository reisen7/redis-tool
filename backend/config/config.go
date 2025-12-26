package config

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"log"
	"os"

	"github.com/BurntSushi/toml"
)

// Config represents the application configuration
type Config struct {
	Server   ServerConfig   `toml:"server"`
	Database DatabaseConfig `toml:"database"`
	Security SecurityConfig `toml:"security"`
}

// ServerConfig represents server configuration
type ServerConfig struct {
	Port string `toml:"port"`
	Host string `toml:"host"`
}

// DatabaseConfig represents SQLite database configuration
type DatabaseConfig struct {
	Path string `toml:"path"` // SQLite database file path
}

// SecurityConfig represents security configuration
type SecurityConfig struct {
	EncryptionKey    string `toml:"encryption_key"`     // 32-byte key for AES-256 (hex or base64 encoded)
	JWTSecret        string `toml:"jwt_secret"`         // Secret key for JWT signing
	DefaultAdminUser string `toml:"default_admin_user"` // Default admin username
	DefaultAdminPass string `toml:"default_admin_pass"` // Default admin password
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Port: "8080",
			Host: "0.0.0.0",
		},
		Database: DatabaseConfig{
			Path: "./data/redis_manager.db",
		},
		Security: SecurityConfig{
			EncryptionKey:    "", // Must be set via config or env
			JWTSecret:        "", // Will be auto-generated if not set
			DefaultAdminUser: "admin",
			DefaultAdminPass: "Admin@123",
		},
	}
}

// Load loads configuration from file and environment variables
// Environment variables take precedence over config file
func Load(configPath string) *Config {
	cfg := DefaultConfig()

	// Try to load from config file
	if configPath != "" {
		if _, err := os.Stat(configPath); err == nil {
			if _, err := toml.DecodeFile(configPath, cfg); err != nil {
				log.Printf("Warning: Failed to parse config file %s: %v", configPath, err)
			} else {
				log.Printf("Loaded configuration from %s", configPath)
			}
		}
	} else {
		// Try default config paths
		defaultPaths := []string{"config.toml", "./config/config.toml", "/etc/redis-manager/config.toml"}
		for _, path := range defaultPaths {
			if _, err := os.Stat(path); err == nil {
				if _, err := toml.DecodeFile(path, cfg); err != nil {
					log.Printf("Warning: Failed to parse config file %s: %v", path, err)
				} else {
					log.Printf("Loaded configuration from %s", path)
					break
				}
			}
		}
	}

	// Override with environment variables
	cfg.loadFromEnv()

	return cfg
}

// loadFromEnv loads configuration from environment variables
func (c *Config) loadFromEnv() {
	// Server config
	if port := os.Getenv("SERVER_PORT"); port != "" {
		c.Server.Port = port
	}
	if port := os.Getenv("PORT"); port != "" {
		c.Server.Port = port
	}
	if host := os.Getenv("SERVER_HOST"); host != "" {
		c.Server.Host = host
	}

	// Database config
	if path := os.Getenv("DB_PATH"); path != "" {
		c.Database.Path = path
	}

	// Security config
	if key := os.Getenv("ENCRYPTION_KEY"); key != "" {
		c.Security.EncryptionKey = key
	}
	if secret := os.Getenv("JWT_SECRET"); secret != "" {
		c.Security.JWTSecret = secret
	}
	if user := os.Getenv("DEFAULT_ADMIN_USER"); user != "" {
		c.Security.DefaultAdminUser = user
	}
	if pass := os.Getenv("DEFAULT_ADMIN_PASS"); pass != "" {
		c.Security.DefaultAdminPass = pass
	}
}

// GetDBPath returns the SQLite database file path
func (c *Config) GetDBPath() string {
	return c.Database.Path
}

// GetEncryptionKey returns the encryption key as bytes
// If not set, generates a random 32-byte key
func (c *Config) GetEncryptionKey() ([]byte, error) {
	if c.Security.EncryptionKey == "" {
		// Generate a random key if not configured
		key := make([]byte, 32)
		if _, err := rand.Read(key); err != nil {
			return nil, err
		}
		return key, nil
	}

	// Try hex decoding first
	key, err := hex.DecodeString(c.Security.EncryptionKey)
	if err == nil && len(key) == 32 {
		return key, nil
	}

	// Try base64 decoding
	key, err = base64.StdEncoding.DecodeString(c.Security.EncryptionKey)
	if err == nil && len(key) == 32 {
		return key, nil
	}

	// Use as raw string (padded/truncated to 32 bytes)
	rawKey := []byte(c.Security.EncryptionKey)
	if len(rawKey) >= 32 {
		return rawKey[:32], nil
	}
	// Pad with zeros if too short
	paddedKey := make([]byte, 32)
	copy(paddedKey, rawKey)
	return paddedKey, nil
}

// GetJWTSecret returns the JWT signing secret as bytes
// If not set, generates a random 32-byte key
func (c *Config) GetJWTSecret() ([]byte, error) {
	if c.Security.JWTSecret == "" {
		// Generate a random key if not configured
		key := make([]byte, 32)
		if _, err := rand.Read(key); err != nil {
			return nil, err
		}
		log.Printf("Warning: JWT_SECRET not configured, using auto-generated secret (tokens will be invalidated on restart)")
		return key, nil
	}

	// Try hex decoding first
	key, err := hex.DecodeString(c.Security.JWTSecret)
	if err == nil && len(key) >= 32 {
		return key[:32], nil
	}

	// Try base64 decoding
	key, err = base64.StdEncoding.DecodeString(c.Security.JWTSecret)
	if err == nil && len(key) >= 32 {
		return key[:32], nil
	}

	// Use as raw string (padded/truncated to 32 bytes)
	rawKey := []byte(c.Security.JWTSecret)
	if len(rawKey) >= 32 {
		return rawKey[:32], nil
	}
	// Pad with zeros if too short
	paddedKey := make([]byte, 32)
	copy(paddedKey, rawKey)
	return paddedKey, nil
}

// Global config instance
var globalConfig *Config

// Get returns the global configuration
func Get() *Config {
	if globalConfig == nil {
		globalConfig = Load("")
	}
	return globalConfig
}

// Init initializes the global configuration
func Init(configPath string) *Config {
	globalConfig = Load(configPath)
	return globalConfig
}
