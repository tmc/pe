// Package plugin provides the plugin system for PE
package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

// Plugin represents a PE plugin
type Plugin struct {
	Name        string
	Path        string
	Description string
	Version     string
	Commands    []Command
}

// Command represents a plugin command
type Command struct {
	Name        string
	Description string
	Usage       string
	Flags       []Flag
}

// Flag represents a command flag
type Flag struct {
	Name        string
	Shorthand   string
	Description string
	Type        string // string, bool, int, float
	Default     interface{}
}

// Manager manages PE plugins
type Manager struct {
	plugins map[string]*Plugin
	mu      sync.RWMutex
}

// NewManager creates a new plugin manager
func NewManager() *Manager {
	return &Manager{
		plugins: make(map[string]*Plugin),
	}
}

// Discover finds all available plugins
func (m *Manager) Discover() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Look for plugins in PATH
	paths := strings.Split(os.Getenv("PATH"), string(os.PathListSeparator))
	
	for _, dir := range paths {
		if err := m.discoverInDir(dir); err != nil {
			// Log but don't fail on individual directory errors
			continue
		}
	}

	// Also look in PE_PLUGIN_PATH if set
	if pluginPath := os.Getenv("PE_PLUGIN_PATH"); pluginPath != "" {
		for _, dir := range strings.Split(pluginPath, string(os.PathListSeparator)) {
			if err := m.discoverInDir(dir); err != nil {
				continue
			}
		}
	}

	return nil
}

// discoverInDir finds plugins in a specific directory
func (m *Manager) discoverInDir(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	prefix := "pe-"
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		// Check for pe-* executables
		if strings.HasPrefix(name, prefix) {
			// Remove .exe suffix on Windows
			if runtime.GOOS == "windows" && strings.HasSuffix(name, ".exe") {
				name = strings.TrimSuffix(name, ".exe")
			}

			pluginName := strings.TrimPrefix(name, prefix)
			pluginPath := filepath.Join(dir, entry.Name())

			// Check if it's executable
			info, err := os.Stat(pluginPath)
			if err != nil {
				continue
			}

			if info.Mode()&0111 == 0 {
				continue // Not executable
			}

			// Try to get plugin info
			if plugin, err := m.loadPlugin(pluginName, pluginPath); err == nil {
				m.plugins[pluginName] = plugin
			}
		}
	}

	return nil
}

// loadPlugin loads information about a plugin
func (m *Manager) loadPlugin(name, path string) (*Plugin, error) {
	// Call the plugin with --pe-plugin-info to get metadata
	cmd := exec.Command(path, "--pe-plugin-info")
	output, err := cmd.Output()
	if err != nil {
		// Fallback to basic info if plugin doesn't support --pe-plugin-info
		return &Plugin{
			Name:        name,
			Path:        path,
			Description: fmt.Sprintf("Plugin: %s", name),
		}, nil
	}

	var plugin Plugin
	if err := json.Unmarshal(output, &plugin); err != nil {
		return nil, fmt.Errorf("failed to parse plugin info: %w", err)
	}

	plugin.Name = name
	plugin.Path = path
	return &plugin, nil
}

// Get returns a plugin by name
func (m *Manager) Get(name string) (*Plugin, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	plugin, ok := m.plugins[name]
	return plugin, ok
}

// List returns all discovered plugins
func (m *Manager) List() []*Plugin {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var plugins []*Plugin
	for _, p := range m.plugins {
		plugins = append(plugins, p)
	}
	return plugins
}

// Execute runs a plugin command
func (m *Manager) Execute(ctx context.Context, pluginName string, args []string) error {
	plugin, ok := m.Get(pluginName)
	if !ok {
		return fmt.Errorf("plugin %q not found", pluginName)
	}

	cmd := exec.CommandContext(ctx, plugin.Path, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// PluginInfo is used for plugin metadata exchange
type PluginInfo struct {
	Description string    `json:"description"`
	Version     string    `json:"version"`
	Commands    []Command `json:"commands"`
}