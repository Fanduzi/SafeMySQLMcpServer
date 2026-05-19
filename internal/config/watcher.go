// Package config provides file watching for hot reload.
// input: config file paths, fsnotify events, poll interval, SIGHUP signal
// output: OnChange callbacks when files change
// pos: config hot reload, monitors config files for changes
// note: supports dual-mode detection (fsnotify + polling) and manual SIGHUP reload
package config

import (
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// Watcher monitors configuration files for changes.
type Watcher struct {
	watcher      *fsnotify.Watcher
	configPath   string
	securityPath string
	callbacks    []func(*Config, *SecurityConfig)
	mu           sync.RWMutex
	done         chan struct{}

	pollInterval    time.Duration
	lastConfigMod   time.Time
	lastSecurityMod time.Time
}

// Option configures a Watcher.
type Option func(*Watcher)

// WithPollInterval enables polling-based config detection at the given interval.
// Use this in environments where inotify is unreliable (e.g., Docker Desktop macOS).
func WithPollInterval(d time.Duration) Option {
	return func(w *Watcher) {
		w.pollInterval = d
	}
}

// NewWatcher creates a new configuration file watcher.
func NewWatcher(configPath string, opts ...Option) (*Watcher, error) {
	fsWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

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

	for _, opt := range opts {
		opt(w)
	}

	if info, err := os.Stat(absPath); err == nil {
		w.lastConfigMod = info.ModTime()
	}

	configDir := filepath.Dir(absPath)
	if err := fsWatcher.Add(configDir); err != nil {
		_ = fsWatcher.Close()
		return nil, err
	}

	return w, nil
}

// SetSecurityPath sets the path to the security config file.
func (w *Watcher) SetSecurityPath(path string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	w.securityPath = absPath

	if info, err := os.Stat(absPath); err == nil {
		w.lastSecurityMod = info.ModTime()
	}

	securityDir := filepath.Dir(absPath)
	return w.watcher.Add(securityDir)
}

// OnChange registers a callback to be called when config changes.
func (w *Watcher) OnChange(callback func(*Config, *SecurityConfig)) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.callbacks = append(w.callbacks, callback)
}

// Start begins watching for configuration changes.
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

	if w.pollInterval > 0 {
		go w.pollLoop()
	}
}

// Reload forces an immediate configuration reload.
func (w *Watcher) Reload() {
	w.reload()
}

// Stop stops the watcher.
func (w *Watcher) Stop() {
	close(w.done)
	_ = w.watcher.Close()
}

func (w *Watcher) handleEvent(filename string) {
	absFilename, err := filepath.Abs(filename)
	if err != nil {
		return
	}

	if absFilename != w.configPath && absFilename != w.securityPath {
		return
	}

	log.Printf("Configuration file changed: %s", filename)
	w.reload()
}

func (w *Watcher) reload() {
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

	w.mu.Lock()
	if info, err := os.Stat(w.configPath); err == nil {
		w.lastConfigMod = info.ModTime()
	}
	if w.securityPath != "" {
		if info, err := os.Stat(w.securityPath); err == nil {
			w.lastSecurityMod = info.ModTime()
		}
	}
	callbacks := make([]func(*Config, *SecurityConfig), len(w.callbacks))
	copy(callbacks, w.callbacks)
	w.mu.Unlock()

	for _, cb := range callbacks {
		cb(cfg, securityCfg)
	}
}

func (w *Watcher) pollLoop() {
	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-w.done:
			return
		case <-ticker.C:
			w.checkMtime()
		}
	}
}

func (w *Watcher) checkMtime() {
	changed := false

	w.mu.RLock()
	if info, err := os.Stat(w.configPath); err == nil {
		if info.ModTime().After(w.lastConfigMod) {
			changed = true
		}
	}

	if w.securityPath != "" {
		if info, err := os.Stat(w.securityPath); err == nil {
			if info.ModTime().After(w.lastSecurityMod) {
				changed = true
			}
		}
	}
	w.mu.RUnlock()

	if changed {
		log.Printf("Config change detected by polling")
		w.reload()
	}
}

// ReloadableConfig holds configuration that can be hot-reloaded.
type ReloadableConfig struct {
	mu       sync.RWMutex
	config   *Config
	security *SecurityConfig
}

// NewReloadableConfig creates a new reloadable configuration.
func NewReloadableConfig(configPath string) (*ReloadableConfig, error) {
	cfg, err := Load(configPath)
	if err != nil {
		return nil, err
	}

	rc := &ReloadableConfig{
		config: cfg,
	}

	if cfg.Security.ConfigFile != "" {
		securityCfg, err := LoadSecurity(cfg.Security.ConfigFile)
		if err != nil {
			return nil, err
		}
		rc.security = securityCfg
	}

	return rc, nil
}

// Get returns the current configuration.
func (rc *ReloadableConfig) Get() *Config {
	rc.mu.RLock()
	defer rc.mu.RUnlock()
	return rc.config
}

// GetSecurity returns the current security configuration.
func (rc *ReloadableConfig) GetSecurity() *SecurityRules {
	rc.mu.RLock()
	defer rc.mu.RUnlock()
	if rc.security == nil {
		return nil
	}
	return &rc.security.Security
}

// Update updates the configuration.
func (rc *ReloadableConfig) Update(cfg *Config, security *SecurityConfig) {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	rc.config = cfg
	rc.security = security
}
