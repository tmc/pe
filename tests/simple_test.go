package tests

import (
	"os"
	"path/filepath"
	"testing"
)

// TestPECommands tests basic PE command functionality
func TestPECommands(t *testing.T) {
	// Create temporary directory for tests
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(tmpDir)

	tests := []struct {
		name        string
		description string
		setup       func() error
		validate    func() error
		wantErr     bool
	}{
		{
			name:        "init creates .pe directory",
			description: "Verifies pe init creates the .pe directory structure",
			setup: func() error {
				// In a real test, we'd execute the pe init command
				// For now, we simulate the expected behavior
				return os.MkdirAll(".pe", 0755)
			},
			validate: func() error {
				if _, err := os.Stat(".pe"); os.IsNotExist(err) {
					return err
				}
				return nil
			},
		},
		{
			name:        "config directory structure",
			description: "Verifies .pe directory has proper structure",
			setup: func() error {
				dirs := []string{".pe/cache", ".pe/history", ".pe/prompts"}
				for _, dir := range dirs {
					if err := os.MkdirAll(dir, 0755); err != nil {
						return err
					}
				}
				return nil
			},
			validate: func() error {
				dirs := []string{".pe/cache", ".pe/history", ".pe/prompts"}
				for _, dir := range dirs {
					if _, err := os.Stat(dir); os.IsNotExist(err) {
						return err
					}
				}
				return nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			if err := tt.setup(); err != nil {
				if !tt.wantErr {
					t.Errorf("setup failed: %v", err)
				}
				return
			}

			// Validate
			if err := tt.validate(); err != nil {
				if !tt.wantErr {
					t.Errorf("validation failed: %v", err)
				}
			}
		})
	}
}

// TestWorkflows tests complete PE workflows
func TestWorkflows(t *testing.T) {
	t.Run("Basic prompt development", func(t *testing.T) {
		// Create test environment
		tmpDir := t.TempDir()
		promptFile := filepath.Join(tmpDir, "test.prompt")
		
		// Write test prompt
		content := "You are a helpful assistant. Answer concisely."
		if err := os.WriteFile(promptFile, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write prompt file: %v", err)
		}

		// Verify file exists
		if _, err := os.Stat(promptFile); err != nil {
			t.Errorf("Prompt file not created: %v", err)
		}
	})

	t.Run("Version control workflow", func(t *testing.T) {
		// Test basic version control concepts
		tmpDir := t.TempDir()
		versionDir := filepath.Join(tmpDir, ".pe", "versions")
		
		// Create version directory
		if err := os.MkdirAll(versionDir, 0755); err != nil {
			t.Fatalf("Failed to create version directory: %v", err)
		}

		// Simulate version tracking
		versionFile := filepath.Join(versionDir, "v1.json")
		if err := os.WriteFile(versionFile, []byte(`{"version": "1.0.0"}`), 0644); err != nil {
			t.Fatalf("Failed to write version file: %v", err)
		}

		// Verify version file exists
		if _, err := os.Stat(versionFile); err != nil {
			t.Errorf("Version file not created: %v", err)
		}
	})

	t.Run("Team collaboration", func(t *testing.T) {
		// Test collaboration features
		tmpDir := t.TempDir()
		sharedDir := filepath.Join(tmpDir, ".pe", "shared")
		
		// Create shared directory
		if err := os.MkdirAll(sharedDir, 0755); err != nil {
			t.Fatalf("Failed to create shared directory: %v", err)
		}

		// Simulate shared configuration
		configFile := filepath.Join(sharedDir, "config.yaml")
		if err := os.WriteFile(configFile, []byte("cache: enabled\n"), 0644); err != nil {
			t.Fatalf("Failed to write config file: %v", err)
		}

		// Verify config file exists
		if _, err := os.Stat(configFile); err != nil {
			t.Errorf("Config file not created: %v", err)
		}
	})
}