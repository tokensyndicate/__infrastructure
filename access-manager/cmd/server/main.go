// cmd/server/main.go
package main

import (
	"flag"
	"fmt"
	"log"

	"access-manager/internal/config"
	"access-manager/internal/ports/http"
	"access-manager/internal/repository"
	"access-manager/internal/service"

	"github.com/go-redis/redis/v8"
	vault "github.com/hashicorp/vault/api"
)

func main() {
	configPath := flag.String("config", "", "path to config file")
	flag.Parse()

	// Load configuration
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize Vault client
	vaultConfig := vault.DefaultConfig()
	vaultConfig.Address = cfg.Vault.Address

	vaultClient, err := vault.NewClient(vaultConfig)
	if err != nil {
		log.Fatalf("Failed to create Vault client: %v", err)
	}
	vaultClient.SetToken(cfg.Vault.Token)

	// Initialize Redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Address,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	// Initialize repositories
	vaultRepo := repository.NewVaultRepository(vaultClient)
	redisRepo := repository.NewRedisRepository(redisClient)

	// Initialize service
	manager := service.NewAccessManager(vaultRepo, redisRepo)

	// Initialize and start HTTP server
	server := http.NewServer(manager)
	serverAddr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)

	log.Printf("Starting server on %s", serverAddr)
	if err := server.Start(serverAddr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
