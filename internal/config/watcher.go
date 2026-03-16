// Package config provides file watching for hot reload.
// input: config file paths, fsnotify events
// output: OnChange callbacks when files change
// pos: config hot reload, monitors config files for changes
// note: if this file changes, update header and internal/config/README.md
package config

import (
	"log"
	"path/filepath"
	"sync"

	"github.com/fsnotify/fsnotify"
)

// Watcher monitors configuration files for changes
type Watcher struct {
	watcher      *fsnotify.Watcher
	configPath   string
	securityPath string
	callbacks    []func(*Config, *SecurityConfig)
	mu           sync.RWMutex
	done         chan struct{}
}

// NewWatcher creates a new configuration file watcher
func NewWatcher(configPath string) (*Watcher, error) {
	fsWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	// Get absolute path
	absPath, err := filepath.Abs(configPath)
	if err != nil {
		_ = fsWatcher.Close()
		return nil, err
	}

	w := &Watcher{
		watcher:    fsWatcher,
		configPath: absPath,
		done:       make(chan struct{}),
	}

	// Watch the directory containing the config file
	configDir := filepath.Dir(absPath)
	if err := fsWatcher.Add(configDir); err != nil {
		_ = fsWatcher.Close()
		return nil, err
	}

	return w, nil
}

// SetSecurityPath sets the path to the security config file
func (w *Watcher) SetSecurityPath(path string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	w.securityPath = absPath

	// Watch the directory containing the security config file
	securityDir := filepath.Dir(absPath)
	return w.watcher.Add(securityDir)
}

// OnChange registers a callback to be called when config changes
func (w *Watcher) OnChange(callback func(*Config, *SecurityConfig)) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.callbacks = append(w.callbacks, callback)
}

// Start begins watching for configuration changes
func (w *Watcher) Start() {
	go func() {
		for {
			select {
			case <-w.done:
				return
			case event, ok := <-w.watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create {
					w.handleEvent(event.Name)
				}
			case err, ok := <-w.watcher.Errors:
				if !ok {
					return
				}
				log.Printf("Config watcher error: %v", err)
			}
		}
	}()
}

// handleEvent handles a file system event
func (w *Watcher) handleEvent(filename string) {
	// Check if the changed file is one of our config files
	absFilename, err := filepath.Abs(filename)
	if err != nil {
		return
	}

	if absFilename != w.configPath && absFilename != w.securityPath {
		return
	}

	log.Printf("Configuration file changed: %s", filename)

	// Reload configuration
	cfg, err := Load(w.configPath)
	if err != nil {
		log.Printf("Failed to reload config: %v", err)
		return
	}

	var securityCfg *SecurityConfig
	if w.securityPath != "" {
		securityCfg, err = LoadSecurity(w.securityPath)
		if err != nil {
			log.Printf("Failed to reload security config: %v", err)
			return
		}
	}

	// Notify all callbacks
	w.mu.RLock()
	callbacks := make([]func(*Config, *SecurityConfig), len(w.callbacks))
	copy(callbacks, w.callbacks)
	w.mu.RUnlock()

	for _, cb := range callbacks {
		cb(cfg, securityCfg)
	}
}

// Stop stops the watcher
func (w *Watcher) Stop() {
	close(w.done)
	_ = w.watcher.Close()
}

// ReloadableConfig holds configuration that can be hot-reloaded
type ReloadableConfig struct {
	mu       sync.RWMutex
	config   *Config
	security *SecurityConfig
}

// NewReloadableConfig creates a new reloadable configuration
func NewReloadableConfig(configPath string) (*ReloadableConfig, error) {
	cfg, err := Load(configPath)
	if err != nil {
		return nil, err
	}

	rc := &ReloadableConfig{
		config: cfg,
	}

	// Load security config if specified
	if cfg.Security.ConfigFile != "" {
		securityCfg, err := LoadSecurity(cfg.Security.ConfigFile)
		if err != nil {
			return nil, err
		}
		rc.security = securityCfg
	}

	return rc, nil
}

// Get returns the current configuration
func (rc *ReloadableConfig) Get() *Config {
	rc.mu.RLock()
	defer rc.mu.RUnlock()
	return rc.config
}

// GetSecurity returns the current security configuration
func (rc *ReloadableConfig) GetSecurity() *SecurityRules {
	rc.mu.RLock()
	defer rc.mu.RUnlock()
	if rc.security == nil {
		return nil
	}
	return &rc.security.Security
}

// Update updates the configuration
func (rc *ReloadableConfig) Update(cfg *Config, security *SecurityConfig) {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	rc.config = cfg
	rc.security = security
}
