package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetCmd_FlagParsing(t *testing.T) {
	if getCmd.Flags().Lookup("json") == nil {
		t.Error("Expected --json flag to exist")
	}
	if getCmd.Flags().Lookup("variant") == nil {
		t.Error("Expected --variant flag to exist")
	}
	if getCmd.Flags().Lookup("keys") == nil {
		t.Error("Expected --keys flag to exist")
	}
}

func TestGetCmd_CommandStructure(t *testing.T) {
	if getCmd.Use != "get [prompt-file] [field]" {
		t.Errorf("Unexpected Use: %s", getCmd.Use)
	}

	if getCmd.Short == "" {
		t.Error("Expected Short description to be set")
	}

	if getCmd.Long == "" {
		t.Error("Expected Long description to be set")
	}
}

func TestRunGet_FileNotFound(t *testing.T) {
	err := runGet(getCmd, []string{"nonexistent-file.txt", "prompt"})
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
}

func TestRunGet_ValidFile(t *testing.T) {
	tmpDir := t.TempDir()
	promptFile := filepath.Join(tmpDir, "test.prompt")

	content := "Hello {{.name}}!"
	if err := os.WriteFile(promptFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	err := runGet(getCmd, []string{promptFile, "prompt"})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestRunGet_Variables(t *testing.T) {
	tmpDir := t.TempDir()
	promptFile := filepath.Join(tmpDir, "test.prompt")

	content := "Hello {{.name}}, you are {{.age}} years old!"
	if err := os.WriteFile(promptFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	err := runGet(getCmd, []string{promptFile, "variables"})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}
