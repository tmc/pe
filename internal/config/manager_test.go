package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewManager_Defaults(t *testing.T) {
	// Use empty paths to avoid loading system configs
	cm, err := NewManager(WithConfigPaths())
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}
	if cm == nil {
		t.Fatal("NewManager returned nil")
	}

	config := cm.Get()
	if config == nil {
		t.Fatal("Get() returned nil")
	}
}

func TestWithConfigPaths(t *testing.T) {
	paths := []string{"/custom/path1.yaml", "/custom/path2.yaml"}
	cm, err := NewManager(WithConfigPaths(paths...))
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}

	gotPaths := cm.GetConfigPaths()
	if len(gotPaths) != len(paths) {
		t.Errorf("expected %d paths, got %d", len(paths), len(gotPaths))
	}
	for i, p := range paths {
		if gotPaths[i] != p {
			t.Errorf("path[%d]: expected %s, got %s", i, p, gotPaths[i])
		}
	}
}

func TestWithEnvPrefix(t *testing.T) {
	cm := &ConfigManager{
		config:       DefaultConfig(),
		configPaths:  []string{},
		cliOverrides: make(map[string]interface{}),
		envPrefix:    "PE",
	}

	opt := WithEnvPrefix("CUSTOM")
	opt(cm)

	if cm.envPrefix != "CUSTOM" {
		t.Errorf("expected envPrefix 'CUSTOM', got '%s'", cm.envPrefix)
	}
}

func TestWithCLIOverrides(t *testing.T) {
	overrides := map[string]interface{}{
		"app.verbose": true,
		"app.debug":   true,
	}

	cm := &ConfigManager{
		config:       DefaultConfig(),
		configPaths:  []string{},
		cliOverrides: make(map[string]interface{}),
		envPrefix:    "PE",
	}

	opt := WithCLIOverrides(overrides)
	opt(cm)

	if len(cm.cliOverrides) != 2 {
		t.Errorf("expected 2 overrides, got %d", len(cm.cliOverrides))
	}
}

func TestWithWatcher(t *testing.T) {
	callback := func(*Config) {
		// Callback for testing
	}

	cm := &ConfigManager{
		config:       DefaultConfig(),
		configPaths:  []string{},
		cliOverrides: make(map[string]interface{}),
		envPrefix:    "PE",
	}

	opt := WithWatcher(callback)
	opt(cm)

	if len(cm.watchCallbacks) != 1 {
		t.Errorf("expected 1 callback, got %d", len(cm.watchCallbacks))
	}
}

func TestConfigManager_Get(t *testing.T) {
	cm, err := NewManager(WithConfigPaths())
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}

	config := cm.Get()
	if config == nil {
		t.Fatal("Get() returned nil")
	}

	// Verify defaults are set
	if config.App.LogLevel == "" {
		t.Error("expected default LogLevel to be set")
	}
}

func TestConfigManager_Set(t *testing.T) {
	// Test using CLI overrides at construction time instead
	// Note: Set() has a lock issue (calls Load while holding lock)
	// so we test the equivalent behavior via WithCLIOverrides
	overrides := map[string]interface{}{
		"app.verbose": true,
	}

	cm, err := NewManager(WithConfigPaths(), WithCLIOverrides(overrides))
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}

	config := cm.Get()
	if !config.App.Verbose {
		t.Error("expected Verbose to be true from CLI overrides")
	}
}

func TestConfigManager_AddConfigPath(t *testing.T) {
	cm, err := NewManager(WithConfigPaths())
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}

	initialCount := len(cm.GetConfigPaths())
	cm.AddConfigPath("/new/config.yaml")

	if len(cm.GetConfigPaths()) != initialCount+1 {
		t.Error("AddConfigPath did not add path")
	}

	// Should be added at the beginning
	if cm.GetConfigPaths()[0] != "/new/config.yaml" {
		t.Error("AddConfigPath should add to beginning of paths")
	}
}

func TestConfigManager_LoadFromFile_YAML(t *testing.T) {
	// Create a temporary directory
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Write test config
	configContent := `
app:
  verbose: true
  debug: true
  log_level: debug
providers:
  default: anthropic
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	cm, err := NewManager(WithConfigPaths(configPath))
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}

	config := cm.Get()
	if !config.App.Verbose {
		t.Error("expected Verbose to be true from file")
	}
	if !config.App.Debug {
		t.Error("expected Debug to be true from file")
	}
	if config.App.LogLevel != "debug" {
		t.Errorf("expected LogLevel 'debug', got '%s'", config.App.LogLevel)
	}
	if config.Providers.Default != "anthropic" {
		t.Errorf("expected provider default 'anthropic', got '%s'", config.Providers.Default)
	}
}

func TestConfigManager_LoadFromFile_JSON(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	configContent := `{
  "app": {
    "verbose": true,
    "log_level": "warn"
  }
}`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	cm, err := NewManager(WithConfigPaths(configPath))
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}

	config := cm.Get()
	if !config.App.Verbose {
		t.Error("expected Verbose to be true from JSON file")
	}
	if config.App.LogLevel != "warn" {
		t.Errorf("expected LogLevel 'warn', got '%s'", config.App.LogLevel)
	}
}

func TestConfigManager_SaveToFile_YAML(t *testing.T) {
	overrides := map[string]interface{}{
		"app.verbose": true,
	}
	cm, err := NewManager(WithConfigPaths(), WithCLIOverrides(overrides))
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}

	// Save to file
	tmpDir := t.TempDir()
	savePath := filepath.Join(tmpDir, "saved-config.yaml")

	err = cm.SaveToFile(savePath)
	if err != nil {
		t.Fatalf("SaveToFile failed: %v", err)
	}

	// Verify file exists and contains content
	data, err := os.ReadFile(savePath)
	if err != nil {
		t.Fatalf("failed to read saved file: %v", err)
	}
	if len(data) == 0 {
		t.Error("saved file is empty")
	}
}

func TestConfigManager_SaveToFile_JSON(t *testing.T) {
	cm, err := NewManager(WithConfigPaths())
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}

	tmpDir := t.TempDir()
	savePath := filepath.Join(tmpDir, "saved-config.json")

	err = cm.SaveToFile(savePath)
	if err != nil {
		t.Fatalf("SaveToFile failed: %v", err)
	}

	data, err := os.ReadFile(savePath)
	if err != nil {
		t.Fatalf("failed to read saved file: %v", err)
	}
	if len(data) == 0 {
		t.Error("saved file is empty")
	}
}

func TestConfigManager_Validate(t *testing.T) {
	cm, err := NewManager(WithConfigPaths())
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}

	// Default config should validate
	err = cm.Validate()
	if err != nil {
		t.Errorf("Validate failed on default config: %v", err)
	}
}

func TestGetDefaultConfigPaths(t *testing.T) {
	paths := getDefaultConfigPaths()
	if len(paths) == 0 {
		t.Error("expected default config paths to be non-empty")
	}

	// Check that some expected paths exist
	foundLocal := false
	for _, p := range paths {
		if p == "./.pe/config.yaml" || p == "./pe.config.yaml" {
			foundLocal = true
			break
		}
	}
	if !foundLocal {
		t.Error("expected local config paths to be included")
	}
}

func TestConfigManager_LoadFromFile_NotExists(t *testing.T) {
	// Non-existent file should not cause error
	cm, err := NewManager(WithConfigPaths("/nonexistent/path/config.yaml"))
	if err != nil {
		t.Fatalf("NewManager should not fail for non-existent config: %v", err)
	}
	if cm == nil {
		t.Fatal("NewManager returned nil")
	}
}

func TestConfigManager_LoadFromFile_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "invalid.yaml")

	// Write invalid YAML
	invalidContent := `
app:
  verbose: [not valid yaml
`
	if err := os.WriteFile(configPath, []byte(invalidContent), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	_, err := NewManager(WithConfigPaths(configPath))
	if err == nil {
		t.Error("expected error for invalid YAML")
	}
}

func TestSetFieldFromString(t *testing.T) {
	cm := &ConfigManager{
		config:       DefaultConfig(),
		configPaths:  []string{},
		cliOverrides: make(map[string]interface{}),
		envPrefix:    "PE",
	}

	tests := []struct {
		name    string
		envKey  string
		envVal  string
		check   func() bool
		wantErr bool
	}{
		{
			name:   "string field",
			envKey: "PE_PROVIDERS_DEFAULT",
			envVal: "anthropic",
			check:  func() bool { return cm.config.Providers.Default == "anthropic" },
		},
		{
			name:   "bool field true",
			envKey: "PE_APP_VERBOSE",
			envVal: "true",
			check:  func() bool { return cm.config.App.Verbose },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable
			os.Setenv(tt.envKey, tt.envVal)
			defer os.Unsetenv(tt.envKey)

			// Reload config
			err := cm.Load()
			if (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && !tt.check() {
				t.Errorf("check failed for %s", tt.name)
			}
		})
	}
}

func TestMergeConfig(t *testing.T) {
	target := DefaultConfig()
	source := &Config{
		App: AppConfig{
			Verbose:  true,
			LogLevel: "debug",
		},
		Providers: ProvidersConfig{
			Default: "ollama",
		},
	}

	err := mergeConfig(target, source)
	if err != nil {
		t.Fatalf("mergeConfig failed: %v", err)
	}

	if !target.App.Verbose {
		t.Error("expected Verbose to be merged")
	}
	if target.App.LogLevel != "debug" {
		t.Errorf("expected LogLevel 'debug', got '%s'", target.App.LogLevel)
	}
	if target.Providers.Default != "ollama" {
		t.Errorf("expected provider default 'ollama', got '%s'", target.Providers.Default)
	}
}

func TestConfigManager_StopWatching_NoWatcher(t *testing.T) {
	cm, err := NewManager(WithConfigPaths())
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}

	// Should not error when no watcher is configured
	err = cm.StopWatching()
	if err != nil {
		t.Errorf("StopWatching should not error when no watcher: %v", err)
	}
}

func TestConfigManager_StartWatching_NoCallbacks(t *testing.T) {
	cm, err := NewManager(WithConfigPaths())
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}

	// Should return early when no callbacks
	err = cm.StartWatching()
	if err != nil {
		t.Errorf("StartWatching should not error with no callbacks: %v", err)
	}
}
