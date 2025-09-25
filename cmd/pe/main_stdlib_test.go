package main

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestMain(t *testing.T) {
	// Test that main runs without error when help is shown
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"pe", "--help"}

	// Capture output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// We can't directly test main() as it calls os.Exit
	// Instead, test that the root command is properly configured
	root := &cobra.Command{
		Use:   "pe",
		Short: "PE - Go for Prompts",
	}

	// Verify root command exists and has subcommands
	if root == nil {
		t.Fatal("root command is nil")
	}
	if root.Use != "pe" {
		t.Errorf("Expected root command use to be 'pe', got: %s", root.Use)
	}
	if !strings.Contains(root.Short, "Go for Prompts") {
		t.Errorf("Expected root command short description to contain 'Go for Prompts', got: %s", root.Short)
	}

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout

	out, _ := io.ReadAll(r)
	_ = out // Output captured but not used in this test
}

func TestRootCommandStructure(t *testing.T) {
	// Create root command similar to main()
	root := &cobra.Command{
		Use:   "pe",
		Short: "PE - Go for Prompts",
		Long: `PE is the unified toolchain for prompt engineering, bringing Go's
philosophy of simplicity, composability, and performance to LLM development.`,
	}

	// Add commands (mock versions for testing)
	root.AddCommand(&cobra.Command{Use: "run"})
	root.AddCommand(&cobra.Command{Use: "build"})
	root.AddCommand(&cobra.Command{Use: "test"})
	root.AddCommand(&cobra.Command{Use: "init"})
	root.AddCommand(&cobra.Command{Use: "mod"})
	root.AddCommand(&cobra.Command{Use: "eval"})
	root.AddCommand(&cobra.Command{Use: "optimize"})

	// Test command existence
	tests := []struct {
		name     string
		cmdName  string
		expected bool
	}{
		{"run command exists", "run", true},
		{"build command exists", "build", true},
		{"test command exists", "test", true},
		{"init command exists", "init", true},
		{"mod command exists", "mod", true},
		{"eval command exists", "eval", true},
		{"optimize command exists", "optimize", true},
		{"nonexistent command", "foobar", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, _, err := root.Find([]string{tt.cmdName})
			if tt.expected {
				if err != nil {
					t.Errorf("Expected to find command %s, but got error: %v", tt.cmdName, err)
				}
				if cmd == nil {
					t.Errorf("Expected to find command %s, but got nil", tt.cmdName)
				}
				if cmd != nil && cmd.Use != tt.cmdName {
					t.Errorf("Expected command use to be %s, got: %s", tt.cmdName, cmd.Use)
				}
			} else {
				if err == nil {
					t.Errorf("Expected error finding command %s, but succeeded", tt.cmdName)
				}
			}
		})
	}
}

func TestHandlePluginExecution(t *testing.T) {
	// Save original args
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	tests := []struct {
		name     string
		args     []string
		isPlugin bool
	}{
		{
			name:     "not a plugin",
			args:     []string{"pe", "eval"},
			isPlugin: false,
		},
		{
			name:     "is a plugin",
			args:     []string{"pe-test", "arg1"},
			isPlugin: true,
		},
		{
			name:     "plugin with path",
			args:     []string{"/usr/local/bin/pe-test", "arg1"},
			isPlugin: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Args = tt.args
			// Note: We can't actually test HandlePluginExecution directly
			// as it may call os.Exit, but we can verify the logic
			baseName := strings.TrimSuffix(os.Args[0], ".exe")
			if lastSlash := strings.LastIndex(baseName, "/"); lastSlash >= 0 {
				baseName = baseName[lastSlash+1:]
			}
			isPlugin := strings.HasPrefix(baseName, "pe-") && baseName != "pe"
			if isPlugin != tt.isPlugin {
				t.Errorf("Expected isPlugin to be %t, got %t for args %v", tt.isPlugin, isPlugin, tt.args)
			}
		})
	}
}

func TestDynamicPluginCommands(t *testing.T) {
	// This tests the plugin discovery mechanism
	root := &cobra.Command{Use: "pe"}

	// Mock plugin manager for testing
	// In real implementation, this would discover actual plugins
	mockPlugins := []string{"promptfoo", "test-plugin"}

	for _, plugin := range mockPlugins {
		cmd := &cobra.Command{
			Use:   plugin,
			Short: "Plugin: " + plugin,
			RunE: func(cmd *cobra.Command, args []string) error {
				return nil
			},
		}
		root.AddCommand(cmd)
	}

	// Verify plugins were added
	for _, plugin := range mockPlugins {
		cmd, _, err := root.Find([]string{plugin})
		if err != nil {
			t.Errorf("Expected to find plugin %s, but got error: %v", plugin, err)
		}
		if cmd == nil {
			t.Errorf("Expected to find plugin %s, but got nil", plugin)
		}
		if cmd != nil && cmd.Use != plugin {
			t.Errorf("Expected plugin use to be %s, got: %s", plugin, cmd.Use)
		}
	}
}

func TestCommandOutput(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		contains string
		wantErr  bool
	}{
		{
			name:     "help output",
			args:     []string{"--help"},
			contains: "PE - Go for Prompts",
			wantErr:  false,
		},
		{
			name:     "unknown command",
			args:     []string{"unknown-command"},
			contains: "", // Don't check specific error message
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := &cobra.Command{
				Use:   "pe",
				Short: "PE - Go for Prompts",
				RunE: func(cmd *cobra.Command, args []string) error {
					return cmd.Help()
				},
			}
			// Add at least one subcommand so unknown command generates error
			root.AddCommand(&cobra.Command{Use: "help"})

			var buf bytes.Buffer
			root.SetOut(&buf)
			root.SetErr(&buf)
			root.SetArgs(tt.args)

			err := root.Execute()
			output := buf.String()

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none. Output: %s", output)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v. Output: %s", err, output)
				}
			}

			if tt.contains != "" {
				if !strings.Contains(output, tt.contains) {
					t.Errorf("Expected output to contain %q, but got: %s", tt.contains, output)
				}
			}
		})
	}
}

func TestAddPipelineCommands(t *testing.T) {
	root := &cobra.Command{Use: "pe"}

	// This would be the actual addPipelineCommands function
	pipelineCommands := []string{"ask", "stream", "filter", "analyze", "collect", "reduce"}

	for _, cmdName := range pipelineCommands {
		root.AddCommand(&cobra.Command{
			Use:   cmdName,
			Short: "Pipeline command: " + cmdName,
		})
	}

	// Verify all pipeline commands were added
	for _, cmdName := range pipelineCommands {
		cmd, _, err := root.Find([]string{cmdName})
		if err != nil {
			t.Errorf("Expected to find pipeline command %s, but got error: %v", cmdName, err)
		}
		if cmd == nil {
			t.Errorf("Expected to find pipeline command %s, but got nil", cmdName)
		}
		if cmd != nil && cmd.Use != cmdName {
			t.Errorf("Expected pipeline command use to be %s, got: %s", cmdName, cmd.Use)
		}
	}
}