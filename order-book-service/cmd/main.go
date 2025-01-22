package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"order-book-service/internal/config"
	"order-book-service/internal/service"

	"github.com/go-redis/redis/v8"
)

func main() {
    cfg, err := config.Load()
    if err != nil {
        log.Fatalf("Failed to load config: %v", err)
    }

    // Create context that listens for the interrupt signal
    ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
    defer stop()

    // Initialize Redis if Access Manager is enabled
    var redisClient *redis.Client
    if cfg.AccessManager.Enabled {
        redisClient = redis.NewClient(&redis.Options{
            Addr:     cfg.Redis.Address,
            Password: cfg.Redis.Password,
            DB:       cfg.Redis.DB,
        })
        defer redisClient.Close()
    }

    // Initialize repository
    repo, err := initRepository(cfg)
    if err != nil {
        log.Fatalf("Failed to initialize repository: %v", err)
    }
    defer repo.Close()

    // Initialize settings API
    pairsAPI := service.NewPairsAPI(repo)
    server := &http.Server{
        Addr:    fmt.Sprintf(":%d", cfg.API.Port),
        Handler: pairsAPI,
    }

    // Start API server
    go func() {
        log.Printf("Starting settings API server on port %d", cfg.API.Port)
        if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Printf("Settings API server error: %v", err)
        }
    }()

    // Initialize and start service manager
    manager := service.NewManager(*cfg, repo, redisClient)
    if err := manager.Start(ctx); err != nil {
        log.Fatalf("Failed to start service: %v", err)
    }

    <-ctx.Done()
    log.Println("Shutting down gracefully...")

    // Shutdown settings API server
    shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.API.ShutdownTimeout)
    defer cancel()

    if err := server.Shutdown(shutdownCtx); err != nil {
        log.Printf("Settings API server shutdown error: %v", err)
    }
}
