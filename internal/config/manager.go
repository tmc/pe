package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"gopkg.in/yaml.v3"
)

// ConfigManager manages configuration with hierarchical lookup:
// 1. CLI flags (highest priority)
// 2. Environment variables
// 3. Configuration files
// 4. Default values (lowest priority)
type ConfigManager struct {
	config         *Config
	configPaths    []string
	cliOverrides   map[string]interface{}
	envPrefix      string
	watcher        *fsnotify.Watcher
	watchCallbacks []func(*Config)
	mu             sync.RWMutex
}

// ConfigOption configures the ConfigManager
type ConfigOption func(*ConfigManager)

// WithConfigPaths sets the configuration file search paths
func WithConfigPaths(paths ...string) ConfigOption {
	return func(cm *ConfigManager) {
		cm.configPaths = paths
	}
}

// WithCLIOverrides sets CLI flag overrides
func WithCLIOverrides(overrides map[string]interface{}) ConfigOption {
	return func(cm *ConfigManager) {
		cm.cliOverrides = overrides
	}
}

// WithEnvPrefix sets the environment variable prefix
func WithEnvPrefix(prefix string) ConfigOption {
	return func(cm *ConfigManager) {
		cm.envPrefix = prefix
	}
}

// WithWatcher enables file watching for hot reload
func WithWatcher(callback func(*Config)) ConfigOption {
	return func(cm *ConfigManager) {
		if cm.watchCallbacks == nil {
			cm.watchCallbacks = make([]func(*Config), 0)
		}
		cm.watchCallbacks = append(cm.watchCallbacks, callback)
	}
}

// NewManager creates a new configuration manager
func NewManager(options ...ConfigOption) (*ConfigManager, error) {
	cm := &ConfigManager{
		config:       DefaultConfig(),
		configPaths:  getDefaultConfigPaths(),
		cliOverrides: make(map[string]interface{}),
		envPrefix:    "PE",
	}

	// Apply options
	for _, opt := range options {
		opt(cm)
	}

	// Load configuration
	if err := cm.Load(); err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	return cm, nil
}

// getDefaultConfigPaths returns the default configuration file search paths
func getDefaultConfigPaths() []string {
	paths := []string{}

	// Current directory configurations
	paths = append(paths,
		"./.pe/config.yaml",
		"./.pe/config.yml",
		"./pe.config.yaml",
		"./pe.config.yml",
		"./promptfooconfig.yaml",
		"./promptfooconfig.yml",
	)

	// User home directory configurations
	if home, err := os.UserHomeDir(); err == nil {
		paths = append(paths,
			filepath.Join(home, ".pe", "config.yaml"),
			filepath.Join(home, ".pe", "config.yml"),
			filepath.Join(home, ".perc"),
		)
	}

	// System-wide configurations
	paths = append(paths,
		"/etc/pe/config.yaml",
		"/etc/pe/config.yml",
	)

	return paths
}

// Load loads configuration from all sources in hierarchical order
func (cm *ConfigManager) Load() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Start with default configuration
	config := DefaultConfig()

	// Load from configuration files
	for _, path := range cm.configPaths {
		if err := cm.loadFromFile(path, config); err != nil {
			// Don't fail if config file doesn't exist
			if !os.IsNotExist(err) {
				return fmt.Errorf("failed to load config from %s: %w", path, err)
			}
		}
	}

	// Apply environment variables
	if err := cm.loadFromEnv(config); err != nil {
		return fmt.Errorf("failed to load config from environment: %w", err)
	}

	// Apply CLI overrides
	if err := cm.applyCLIOverrides(config); err != nil {
		return fmt.Errorf("failed to apply CLI overrides: %w", err)
	}

	cm.config = config
	return nil
}

// loadFromFile loads configuration from a single file
func (cm *ConfigManager) loadFromFile(path string, config *Config) error {
	// Expand user home directory
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		path = filepath.Join(home, path[2:])
	}

	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	// Determine file format by extension
	ext := strings.ToLower(filepath.Ext(path))

	var fileConfig Config
	switch ext {
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(data, &fileConfig); err != nil {
			return fmt.Errorf("failed to parse YAML: %w", err)
		}
	case ".json":
		if err := json.Unmarshal(data, &fileConfig); err != nil {
			return fmt.Errorf("failed to parse JSON: %w", err)
		}
	default:
		// Assume YAML by default
		if err := yaml.Unmarshal(data, &fileConfig); err != nil {
			return fmt.Errorf("failed to parse as YAML: %w", err)
		}
	}

	// Merge file config into main config
	if err := mergeConfig(config, &fileConfig); err != nil {
		return fmt.Errorf("failed to merge config: %w", err)
	}

	return nil
}

// loadFromEnv loads configuration from environment variables
func (cm *ConfigManager) loadFromEnv(config *Config) error {
	return cm.loadEnvForStruct(reflect.ValueOf(config).Elem(), "")
}

// loadEnvForStruct recursively loads environment variables for a struct
func (cm *ConfigManager) loadEnvForStruct(v reflect.Value, prefix string) error {
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		// Skip unexported fields
		if !field.CanSet() {
			continue
		}

		// Build environment variable name
		envKey := strings.ToUpper(prefix + fieldType.Name)
		if prefix == "" {
			envKey = cm.envPrefix + "_" + envKey
		} else {
			envKey = cm.envPrefix + "_" + envKey
		}

		// Check for explicit env tag
		if envTag := fieldType.Tag.Get("env"); envTag != "" {
			envKey = envTag
		}

		// Handle nested structs
		if field.Kind() == reflect.Struct {
			newPrefix := prefix + fieldType.Name + "_"
			if prefix == "" {
				newPrefix = fieldType.Name + "_"
			}
			if err := cm.loadEnvForStruct(field, newPrefix); err != nil {
				return err
			}
			continue
		}

		// Get environment variable value
		envValue := os.Getenv(envKey)
		if envValue == "" {
			continue
		}

		// Set the value based on field type
		if err := cm.setFieldFromString(field, envValue); err != nil {
			return fmt.Errorf("failed to set field %s from env %s: %w", fieldType.Name, envKey, err)
		}
	}

	return nil
}

// setFieldFromString sets a reflect.Value from a string representation
func (cm *ConfigManager) setFieldFromString(field reflect.Value, value string) error {
	switch field.Kind() {
	case reflect.String:
		field.SetString(value)

	case reflect.Bool:
		b, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		field.SetBool(b)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		// Handle time.Duration specially
		if field.Type() == reflect.TypeOf(time.Duration(0)) {
			d, err := time.ParseDuration(value)
			if err != nil {
				return err
			}
			field.SetInt(int64(d))
		} else {
			i, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return err
			}
			field.SetInt(i)
		}

	case reflect.Float32, reflect.Float64:
		f, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		field.SetFloat(f)

	case reflect.Slice:
		// Handle string slices from comma-separated values
		if field.Type().Elem().Kind() == reflect.String {
			values := strings.Split(value, ",")
			slice := reflect.MakeSlice(field.Type(), len(values), len(values))
			for i, v := range values {
				slice.Index(i).SetString(strings.TrimSpace(v))
			}
			field.Set(slice)
		}

	case reflect.Map:
		// Handle maps from JSON-encoded values
		if field.Type().Key().Kind() == reflect.String {
			m := reflect.MakeMap(field.Type())
			if err := json.Unmarshal([]byte(value), m.Addr().Interface()); err != nil {
				return err
			}
			field.Set(m)
		}

	default:
		return fmt.Errorf("unsupported field type: %s", field.Kind())
	}

	return nil
}

// applyCLIOverrides applies CLI flag overrides to the configuration
func (cm *ConfigManager) applyCLIOverrides(config *Config) error {
	for key, value := range cm.cliOverrides {
		if err := cm.setConfigValue(config, key, value); err != nil {
			return fmt.Errorf("failed to set CLI override %s: %w", key, err)
		}
	}
	return nil
}

// setConfigValue sets a configuration value using dot notation (e.g., "providers.openai.api_key")
func (cm *ConfigManager) setConfigValue(config *Config, key string, value interface{}) error {
	keys := strings.Split(key, ".")
	v := reflect.ValueOf(config).Elem()

	// Navigate to the target field
	for i, k := range keys[:len(keys)-1] {
		v = cm.getFieldByName(v, k)
		if !v.IsValid() {
			return fmt.Errorf("invalid config path at %s", strings.Join(keys[:i+1], "."))
		}
	}

	// Set the final field
	finalField := cm.getFieldByName(v, keys[len(keys)-1])
	if !finalField.IsValid() || !finalField.CanSet() {
		return fmt.Errorf("cannot set config field %s", key)
	}

	// Convert value to appropriate type
	if err := cm.setReflectValue(finalField, value); err != nil {
		return fmt.Errorf("failed to convert value for %s: %w", key, err)
	}

	return nil
}

// getFieldByName gets a struct field by name, handling both direct names and YAML/JSON tags
func (cm *ConfigManager) getFieldByName(v reflect.Value, name string) reflect.Value {
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		// Check direct name match
		if strings.EqualFold(fieldType.Name, name) {
			return field
		}

		// Check YAML tag
		if yamlTag := fieldType.Tag.Get("yaml"); yamlTag != "" {
			yamlName := strings.Split(yamlTag, ",")[0]
			if yamlName == name {
				return field
			}
		}

		// Check JSON tag
		if jsonTag := fieldType.Tag.Get("json"); jsonTag != "" {
			jsonName := strings.Split(jsonTag, ",")[0]
			if jsonName == name {
				return field
			}
		}
	}

	return reflect.Value{}
}

// setReflectValue sets a reflect.Value from an interface{}
func (cm *ConfigManager) setReflectValue(field reflect.Value, value interface{}) error {
	valueType := reflect.TypeOf(value)
	fieldType := field.Type()

	// Direct assignment if types match
	if valueType.AssignableTo(fieldType) {
		field.Set(reflect.ValueOf(value))
		return nil
	}

	// Convert string values
	if valueType.Kind() == reflect.String {
		return cm.setFieldFromString(field, value.(string))
	}

	// Convert numeric values
	if valueType.ConvertibleTo(fieldType) {
		field.Set(reflect.ValueOf(value).Convert(fieldType))
		return nil
	}

	return fmt.Errorf("cannot convert %T to %s", value, fieldType)
}

// mergeConfig merges source configuration into target configuration
func mergeConfig(target, source *Config) error {
	return mergeStruct(reflect.ValueOf(target).Elem(), reflect.ValueOf(source).Elem())
}

// mergeStruct recursively merges two structs
func mergeStruct(target, source reflect.Value) error {
	targetType := target.Type()

	for i := 0; i < target.NumField(); i++ {
		targetField := target.Field(i)
		sourceField := source.Field(i)
		fieldType := targetType.Field(i)

		// Skip unexported fields
		if !targetField.CanSet() {
			continue
		}

		// Handle different field types
		switch targetField.Kind() {
		case reflect.Struct:
			// Handle time.Duration as a special case
			if targetField.Type() == reflect.TypeOf(time.Duration(0)) {
				if !sourceField.IsZero() {
					targetField.Set(sourceField)
				}
			} else {
				if err := mergeStruct(targetField, sourceField); err != nil {
					return fmt.Errorf("failed to merge field %s: %w", fieldType.Name, err)
				}
			}

		case reflect.Map:
			if !sourceField.IsNil() && sourceField.Len() > 0 {
				if targetField.IsNil() {
					targetField.Set(reflect.MakeMap(targetField.Type()))
				}
				// Merge map entries
				for _, key := range sourceField.MapKeys() {
					targetField.SetMapIndex(key, sourceField.MapIndex(key))
				}
			}

		case reflect.Slice:
			if !sourceField.IsNil() && sourceField.Len() > 0 {
				targetField.Set(sourceField)
			}

		default:
			// For primitive types, use source value if it's not zero
			if !sourceField.IsZero() {
				targetField.Set(sourceField)
			}
		}
	}

	return nil
}

// Get returns the current configuration
func (cm *ConfigManager) Get() *Config {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	// Return a copy to prevent external modifications
	config := *cm.config
	return &config
}

// Set updates a configuration value and triggers a reload
func (cm *ConfigManager) Set(key string, value interface{}) error {
	cm.mu.Lock()
	if cm.cliOverrides == nil {
		cm.cliOverrides = make(map[string]interface{})
	}
	cm.cliOverrides[key] = value
	cm.mu.Unlock()
	return cm.Load()
}

// StartWatching starts watching configuration files for changes
func (cm *ConfigManager) StartWatching() error {
	if len(cm.watchCallbacks) == 0 {
		return nil // No watchers configured
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create file watcher: %w", err)
	}

	cm.watcher = watcher

	// Watch all existing config files
	for _, path := range cm.configPaths {
		if strings.HasPrefix(path, "~/") {
			home, err := os.UserHomeDir()
			if err != nil {
				continue
			}
			path = filepath.Join(home, path[2:])
		}

		if _, err := os.Stat(path); err == nil {
			if err := watcher.Add(path); err != nil {
				continue // Skip files we can't watch
			}
		}
	}

	// Start the watcher goroutine
	go cm.watchLoop()

	return nil
}

// watchLoop processes file system events
func (cm *ConfigManager) watchLoop() {
	for {
		select {
		case event, ok := <-cm.watcher.Events:
			if !ok {
				return
			}

			// Only reload on write events
			if event.Op&fsnotify.Write == fsnotify.Write {
				if err := cm.Load(); err != nil {
					// Log error but continue watching
					continue
				}

				// Notify all callbacks
				config := cm.Get()
				for _, callback := range cm.watchCallbacks {
					callback(config)
				}
			}

		case err, ok := <-cm.watcher.Errors:
			if !ok {
				return
			}
			// Log error but continue watching
			_ = err
		}
	}
}

// StopWatching stops watching configuration files
func (cm *ConfigManager) StopWatching() error {
	if cm.watcher != nil {
		return cm.watcher.Close()
	}
	return nil
}

// SaveToFile saves the current configuration to a file
func (cm *ConfigManager) SaveToFile(path string) error {
	cm.mu.RLock()
	config := cm.config
	cm.mu.RUnlock()

	// Determine format from file extension
	ext := strings.ToLower(filepath.Ext(path))

	var data []byte
	var err error

	switch ext {
	case ".yaml", ".yml":
		data, err = yaml.Marshal(config)
	case ".json":
		data, err = json.MarshalIndent(config, "", "  ")
	default:
		// Default to YAML
		data, err = yaml.Marshal(config)
	}

	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// Validate validates the current configuration
func (cm *ConfigManager) Validate() error {
	config := cm.Get()
	return ValidateConfig(config)
}

// GetConfigPaths returns the list of configuration file paths being searched
func (cm *ConfigManager) GetConfigPaths() []string {
	return cm.configPaths
}

// AddConfigPath adds a configuration file path to the search list
func (cm *ConfigManager) AddConfigPath(path string) {
	cm.configPaths = append([]string{path}, cm.configPaths...)
}
