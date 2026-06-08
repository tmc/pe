package plugin

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestNewManager(t *testing.T) {
	m := NewManager()
	if m == nil {
		t.Fatal("NewManager returned nil")
	}
	if m.plugins == nil {
		t.Error("plugins map should be initialized")
	}
}

func TestManager_Get_NotFound(t *testing.T) {
	m := NewManager()

	_, ok := m.Get("nonexistent")
	if ok {
		t.Error("expected Get to return false for nonexistent plugin")
	}
}

func TestManager_List_Empty(t *testing.T) {
	m := NewManager()

	plugins := m.List()
	if len(plugins) != 0 {
		t.Errorf("expected 0 plugins, got %d", len(plugins))
	}
}

func TestManager_Discover(t *testing.T) {
	m := NewManager()

	// Discover shouldn't error even if no plugins found
	err := m.Discover()
	if err != nil {
		t.Fatalf("Discover failed: %v", err)
	}
}

func TestManager_Discover_CustomPluginPath(t *testing.T) {
	m := NewManager()
	tmpDir := t.TempDir()

	// Set PE_PLUGIN_PATH
	oldPath := os.Getenv("PE_PLUGIN_PATH")
	os.Setenv("PE_PLUGIN_PATH", tmpDir)
	defer os.Setenv("PE_PLUGIN_PATH", oldPath)

	err := m.Discover()
	if err != nil {
		t.Fatalf("Discover failed with custom plugin path: %v", err)
	}
}

func TestManager_discoverInDir_NonExistent(t *testing.T) {
	m := NewManager()

	err := m.discoverInDir("/nonexistent/directory")
	if err == nil {
		t.Error("expected error for nonexistent directory")
	}
}

func TestManager_discoverInDir_EmptyDir(t *testing.T) {
	m := NewManager()
	tmpDir := t.TempDir()

	err := m.discoverInDir(tmpDir)
	if err != nil {
		t.Fatalf("discoverInDir failed on empty dir: %v", err)
	}

	if len(m.plugins) != 0 {
		t.Errorf("expected 0 plugins, got %d", len(m.plugins))
	}
}

func TestManager_discoverInDir_WithExecutable(t *testing.T) {
	m := NewManager()
	tmpDir := t.TempDir()

	// Create a mock plugin executable
	pluginPath := filepath.Join(tmpDir, "pe-test-plugin")

	// Create a simple executable script
	script := "#!/bin/sh\necho '{\"description\":\"test\",\"version\":\"1.0.0\"}'"
	if err := os.WriteFile(pluginPath, []byte(script), 0755); err != nil {
		t.Fatalf("failed to create mock plugin: %v", err)
	}

	err := m.discoverInDir(tmpDir)
	if err != nil {
		t.Fatalf("discoverInDir failed: %v", err)
	}

	// Should have discovered the plugin
	plugin, ok := m.Get("test-plugin")
	if !ok {
		t.Error("expected to discover test-plugin")
	} else if plugin.Name != "test-plugin" {
		t.Errorf("expected name 'test-plugin', got '%s'", plugin.Name)
	}
}

func TestManager_DiscoverIgnoresPATH(t *testing.T) {
	m := NewManager()
	tmpDir := t.TempDir()
	marker := filepath.Join(tmpDir, "executed")
	pluginPath := filepath.Join(tmpDir, "pe-path-plugin")
	script := "#!/bin/sh\ntouch " + marker + "\n"
	if err := os.WriteFile(pluginPath, []byte(script), 0755); err != nil {
		t.Fatalf("failed to create mock plugin: %v", err)
	}

	t.Setenv("PATH", tmpDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("PE_PLUGIN_PATH", "")

	if err := m.Discover(); err != nil {
		t.Fatalf("Discover failed: %v", err)
	}
	if _, ok := m.Get("path-plugin"); ok {
		t.Fatal("discovered plugin from PATH")
	}
	if _, err := os.Stat(marker); !os.IsNotExist(err) {
		t.Fatal("executed plugin from PATH during discovery")
	}
}

func TestManager_discoverInDir_DoesNotExecuteMetadata(t *testing.T) {
	m := NewManager()
	tmpDir := t.TempDir()
	marker := filepath.Join(tmpDir, "executed")
	pluginPath := filepath.Join(tmpDir, "pe-test-plugin")
	script := "#!/bin/sh\ntouch " + marker + "\necho '{\"description\":\"test\",\"version\":\"1.0.0\"}'\n"
	if err := os.WriteFile(pluginPath, []byte(script), 0755); err != nil {
		t.Fatalf("failed to create mock plugin: %v", err)
	}

	if err := m.discoverInDir(tmpDir); err != nil {
		t.Fatalf("discoverInDir failed: %v", err)
	}
	plugin, ok := m.Get("test-plugin")
	if !ok {
		t.Fatal("expected to discover test-plugin")
	}
	if !strings.Contains(plugin.Description, "test-plugin") {
		t.Fatalf("Description = %q, want fallback", plugin.Description)
	}
	if _, err := os.Stat(marker); !os.IsNotExist(err) {
		t.Fatal("executed plugin during metadata discovery")
	}
}

func TestManager_Execute_NotFound(t *testing.T) {
	m := NewManager()

	err := m.Execute(context.Background(), "nonexistent", []string{})
	if err == nil {
		t.Error("expected error for nonexistent plugin")
	}
}

func TestManager_Execute_PassesArgsWithoutShell(t *testing.T) {
	m := NewManager()
	tmpDir := t.TempDir()
	argsFile := filepath.Join(tmpDir, "args")
	marker := filepath.Join(tmpDir, "marker")
	pluginPath := filepath.Join(tmpDir, "pe-argv")
	script := "#!/bin/sh\nprintf '%s\\n' \"$@\" > " + shellQuote(argsFile) + "\n"
	if err := os.WriteFile(pluginPath, []byte(script), 0755); err != nil {
		t.Fatalf("failed to create plugin: %v", err)
	}

	m.plugins["argv"] = &Plugin{Name: "argv", Path: pluginPath}
	arg := "hello $(touch " + marker + ")"
	if err := m.Execute(context.Background(), "argv", []string{arg, "two words"}); err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	data, err := os.ReadFile(argsFile)
	if err != nil {
		t.Fatalf("ReadFile(args): %v", err)
	}
	if got, want := string(data), arg+"\ntwo words\n"; got != want {
		t.Fatalf("args = %q, want %q", got, want)
	}
	if _, err := os.Stat(marker); !os.IsNotExist(err) {
		t.Fatal("plugin argument was interpreted by a shell before execution")
	}
}

func TestManager_Execute_ContextCancellation(t *testing.T) {
	m := NewManager()
	tmpDir := t.TempDir()
	pluginPath := filepath.Join(tmpDir, "pe-sleep")
	script := "#!/bin/sh\nsleep 1\n"
	if err := os.WriteFile(pluginPath, []byte(script), 0755); err != nil {
		t.Fatalf("failed to create plugin: %v", err)
	}
	m.plugins["sleep"] = &Plugin{Name: "sleep", Path: pluginPath}

	ctx, cancel := context.WithTimeout(context.Background(), time.Nanosecond)
	defer cancel()
	err := m.Execute(ctx, "sleep", nil)
	if err == nil {
		t.Fatal("Execute succeeded after context cancellation")
	}
}

func TestPlugin_Struct(t *testing.T) {
	p := Plugin{
		Name:        "test",
		Path:        "/usr/local/bin/pe-test",
		Description: "Test plugin",
		Version:     "1.0.0",
		Commands: []Command{
			{
				Name:        "run",
				Description: "Run the plugin",
				Usage:       "pe test run [args]",
			},
		},
	}

	if p.Name != "test" {
		t.Error("Name mismatch")
	}
	if p.Version != "1.0.0" {
		t.Error("Version mismatch")
	}
	if len(p.Commands) != 1 {
		t.Error("Commands length mismatch")
	}
}

func TestCommand_Struct(t *testing.T) {
	cmd := Command{
		Name:        "test",
		Description: "Test command",
		Usage:       "pe plugin test",
		Flags: []Flag{
			{
				Name:        "verbose",
				Shorthand:   "v",
				Description: "Enable verbose output",
				Type:        "bool",
				Default:     false,
			},
		},
	}

	if cmd.Name != "test" {
		t.Error("Name mismatch")
	}
	if len(cmd.Flags) != 1 {
		t.Error("Flags length mismatch")
	}
}

func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'"
}

func TestFlag_Struct(t *testing.T) {
	f := Flag{
		Name:        "output",
		Shorthand:   "o",
		Description: "Output file",
		Type:        "string",
		Default:     "output.txt",
	}

	if f.Name != "output" {
		t.Error("Name mismatch")
	}
	if f.Shorthand != "o" {
		t.Error("Shorthand mismatch")
	}
	if f.Type != "string" {
		t.Error("Type mismatch")
	}
}

func TestPluginInfo_Struct(t *testing.T) {
	info := PluginInfo{
		Description: "Test plugin",
		Version:     "1.0.0",
		Commands: []Command{
			{Name: "run", Description: "Run command"},
		},
	}

	if info.Description != "Test plugin" {
		t.Error("Description mismatch")
	}
	if info.Version != "1.0.0" {
		t.Error("Version mismatch")
	}
	if len(info.Commands) != 1 {
		t.Error("Commands length mismatch")
	}
}

func TestManager_loadPlugin_Fallback(t *testing.T) {
	m := NewManager()
	tmpDir := t.TempDir()

	// Create a simple executable that doesn't support --pe-plugin-info
	pluginPath := filepath.Join(tmpDir, "pe-simple")
	script := "#!/bin/sh\nexit 1"
	if err := os.WriteFile(pluginPath, []byte(script), 0755); err != nil {
		t.Fatalf("failed to create mock plugin: %v", err)
	}

	plugin, err := m.loadPlugin("simple", pluginPath)
	if err != nil {
		t.Fatalf("loadPlugin failed: %v", err)
	}

	// Should use fallback info
	if plugin.Name != "simple" {
		t.Errorf("expected name 'simple', got '%s'", plugin.Name)
	}
	if plugin.Path != pluginPath {
		t.Error("Path mismatch")
	}
}

func TestManager_discoverInDir_SkipsDirectories(t *testing.T) {
	m := NewManager()
	tmpDir := t.TempDir()

	// Create a subdirectory that looks like a plugin
	subDir := filepath.Join(tmpDir, "pe-subdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("failed to create subdirectory: %v", err)
	}

	err := m.discoverInDir(tmpDir)
	if err != nil {
		t.Fatalf("discoverInDir failed: %v", err)
	}

	// Should not have discovered the directory as a plugin
	_, ok := m.Get("subdir")
	if ok {
		t.Error("should not discover directories as plugins")
	}
}

func TestManager_discoverInDir_SkipsNonExecutable(t *testing.T) {
	m := NewManager()
	tmpDir := t.TempDir()

	// Create a non-executable file that looks like a plugin
	pluginPath := filepath.Join(tmpDir, "pe-noexec")
	if err := os.WriteFile(pluginPath, []byte("not executable"), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	err := m.discoverInDir(tmpDir)
	if err != nil {
		t.Fatalf("discoverInDir failed: %v", err)
	}

	// Should not have discovered the non-executable file
	_, ok := m.Get("noexec")
	if ok {
		t.Error("should not discover non-executable files as plugins")
	}
}
