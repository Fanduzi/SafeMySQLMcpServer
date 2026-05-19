// Package main provides the entry point for SafeMySQLMcpServer.
// input: config.yaml file path (flag), JWT_SECRET env var, CONFIG_POLL_INTERVAL env var
// output: starts HTTP server, handles graceful shutdown and config hot reload
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

	_ "github.com/pingcap/tidb/parser/test_driver"

	"github.com/fan/safe-mysql-mcp/internal/config"
	"github.com/fan/safe-mysql-mcp/internal/server"
)

func main() {
	configPath := flag.String("config", "config/config.yaml", "Path to configuration file")
	pollInterval := flag.Duration("poll-interval", 0, "Config file poll interval for Docker (0 = fsnotify only, e.g. 30s)")
	flag.Parse()

	if envPoll := os.Getenv("CONFIG_POLL_INTERVAL"); envPoll != "" {
		if d, err := time.ParseDuration(envPoll); err == nil {
			*pollInterval = d
		} else {
			log.Printf("Warning: invalid CONFIG_POLL_INTERVAL %q: %v", envPoll, err)
		}
	}

	reloadableCfg, err := config.NewReloadableConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	cfg := reloadableCfg.Get()

	srv, err := server.New(reloadableCfg)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	var watcher *config.Watcher
	if *pollInterval > 0 {
		watcher, err = config.NewWatcher(*configPath, config.WithPollInterval(*pollInterval))
	} else {
		watcher, err = config.NewWatcher(*configPath)
	}
	if err != nil {
		log.Printf("Warning: Failed to create config watcher: %v", err)
	} else {
		if cfg.Security.ConfigFile != "" {
			if err := watcher.SetSecurityPath(cfg.Security.ConfigFile); err != nil {
				log.Printf("Warning: Failed to watch security config: %v", err)
			}
		}

		watcher.OnChange(func(newCfg *config.Config, security *config.SecurityConfig) {
			log.Println("Configuration changed, updating server...")
			srv.UpdateConfig(newCfg, security)
		})

		watcher.Start()
		defer watcher.Stop()
	}

	go func() {
		if err := srv.Start(); err != nil {
			log.Printf("Server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	sighup := make(chan os.Signal, 1)
	signal.Notify(sighup, syscall.SIGHUP)

	for {
		select {
		case <-sighup:
			if watcher != nil {
				log.Println("Received SIGHUP, reloading configuration...")
				watcher.Reload()
			}
		case <-quit:
			goto shutdown
		}
	}

shutdown:
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}

	log.Println("Server stopped")
}
