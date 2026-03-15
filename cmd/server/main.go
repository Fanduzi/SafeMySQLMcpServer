// Package main provides the entry point for SafeMySQLMcpServer.
// input: config.yaml file path (flag), JWT_SECRET env var
// output: starts HTTP server, handles graceful shutdown
// pos: application entry point, wires config watcher to server
// note: if this file changes, update header and cmd/server/README.md
package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fan/safe-mysql-mcp/internal/config"
	"github.com/fan/safe-mysql-mcp/internal/server"
)

func main() {
	// Parse command line arguments
	configPath := flag.String("config", "config/config.yaml", "Path to configuration file")
	flag.Parse()

	// Load configuration
	reloadableCfg, err := config.NewReloadableConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	cfg := reloadableCfg.Get()

	// Create server
	srv, err := server.New(reloadableCfg)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	// Setup config watcher for hot reload
	watcher, err := config.NewWatcher(*configPath)
	if err != nil {
		log.Printf("Warning: Failed to create config watcher: %v", err)
	} else {
		// Set security config path
		if cfg.Security.ConfigFile != "" {
			if err := watcher.SetSecurityPath(cfg.Security.ConfigFile); err != nil {
				log.Printf("Warning: Failed to watch security config: %v", err)
			}
		}

		// Register callback for config changes
		watcher.OnChange(func(newCfg *config.Config, security *config.SecurityConfig) {
			log.Println("Configuration changed, updating server...")
			srv.UpdateConfig(newCfg, security)
		})

		watcher.Start()
		defer watcher.Stop()
	}

	// Start server in goroutine
	go func() {
		if err := srv.Start(); err != nil {
			log.Printf("Server error: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}

	log.Println("Server stopped")
}
