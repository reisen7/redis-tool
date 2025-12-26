package main

import (
	"flag"
	"log"

	"redis-web-manager/config"
	"redis-web-manager/router"
	"redis-web-manager/service"
)

func main() {
	// Parse command line flags
	configPath := flag.String("config", "", "Path to configuration file")
	flag.Parse()

	// Initialize configuration
	cfg := config.Init(*configPath)

	// Initialize JWT secret
	jwtSecret, err := cfg.GetJWTSecret()
	if err != nil {
		log.Fatalf("Failed to get JWT secret: %v", err)
	}
	service.SetJWTSecret(jwtSecret)

	// Initialize encryption key for connection passwords
	encryptionKey, err := cfg.GetEncryptionKey()
	if err != nil {
		log.Fatalf("Failed to get encryption key: %v", err)
	}
	if err := service.SetEncryptionKey(encryptionKey); err != nil {
		log.Fatalf("Failed to set encryption key: %v", err)
	}

	// Initialize database
	if err := service.InitDatabase(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	r := router.Setup()

	addr := cfg.Server.Host + ":" + cfg.Server.Port
	log.Printf("Starting server on %s", addr)
	r.Run(addr)
}
