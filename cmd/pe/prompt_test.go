package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPromptCmd_CommandStructure(t *testing.T) {
	if promptCmd.Use != "prompt" {
		t.Errorf("Unexpected Use: %s", promptCmd.Use)
	}

	if promptCmd.Short == "" {
		t.Error("Expected Short description to be set")
	}

	// Verify subcommands exist
	subcommandNames := []string{"init", "edit", "info", "tidy", "help"}
	for _, name := range subcommandNames {
		found := false
		for _, cmd := range promptCmd.Commands() {
			if cmd.Name() == name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected subcommand %q to exist", name)
		}
	}
}

func TestPromptInitCmd_DefaultFilename(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	cmd := promptInitCmd()
	err := cmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Should create prompt.prompt
	content, err := os.ReadFile(filepath.Join(tmpDir, "prompt.prompt"))
	if err != nil {
		t.Fatalf("Expected prompt.prompt to be created: %v", err)
	}

	// Should contain shebang
	if !strings.HasPrefix(string(content), "#!/usr/bin/env pe run") {
		t.Error("Expected shebang in prompt file")
	}
}

func TestPromptInitCmd_CustomFilename(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	cmd := promptInitCmd()
	err := cmd.RunE(cmd, []string{"custom"})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Should create custom.prompt (adds extension)
	if _, err := os.Stat(filepath.Join(tmpDir, "custom.prompt")); os.IsNotExist(err) {
		t.Error("Expected custom.prompt to be created")
	}
}

func TestPromptInitCmd_WithExtension(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	cmd := promptInitCmd()
	err := cmd.RunE(cmd, []string{"test.prompt"})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Should create test.prompt
	if _, err := os.Stat(filepath.Join(tmpDir, "test.prompt")); os.IsNotExist(err) {
		t.Error("Expected test.prompt to be created")
	}
}

func TestPromptInitCmd_FileExists(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	// Create existing file
	err := os.WriteFile(filepath.Join(tmpDir, "existing.prompt"), []byte("existing"), 0644)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	cmd := promptInitCmd()
	err = cmd.RunE(cmd, []string{"existing"})
	if err == nil {
		t.Error("Expected error when file exists")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("Expected 'already exists' error, got: %v", err)
	}
}

func TestPromptInitCmd_Flags(t *testing.T) {
	cmd := promptInitCmd()

	// Check flag existence
	flags := []string{"provider", "system", "with-defaults", "with-tests", "with-variants", "force"}
	for _, name := range flags {
		if cmd.Flags().Lookup(name) == nil {
			t.Errorf("Expected flag %q to exist", name)
		}
	}
}

func TestPromptInfoCmd_Structure(t *testing.T) {
	cmd := promptInfoCmd()
	if cmd.Use != "info [file]" {
		t.Errorf("Unexpected Use: %s", cmd.Use)
	}
}

func TestPromptInfoCmd_FileNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	cmd := promptInfoCmd()
	err := cmd.RunE(cmd, []string{"nonexistent.prompt"})
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
}

func TestPromptInfoCmd_Success(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	// Create a test prompt
	content := `#!/usr/bin/env pe run
Hello {{.name}}!

-- defaults --
name = World
`
	err := os.WriteFile(filepath.Join(tmpDir, "test.prompt"), []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	cmd := promptInfoCmd()
	err = cmd.RunE(cmd, []string{"test.prompt"})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestPromptTidyCmd_Structure(t *testing.T) {
	cmd := promptTidyCmd()
	if cmd.Use != "tidy [file]" {
		t.Errorf("Unexpected Use: %s", cmd.Use)
	}
}

func TestPromptEditCmd_Structure(t *testing.T) {
	cmd := promptEditCmd()
	if cmd.Use != "edit [file]" {
		t.Errorf("Unexpected Use: %s", cmd.Use)
	}
}

func TestPromptHelpCmd_Structure(t *testing.T) {
	cmd := promptHelpCmd()
	if !strings.HasPrefix(cmd.Use, "help") {
		t.Errorf("Unexpected Use: %s", cmd.Use)
	}
}
