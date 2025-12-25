package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestModCmd_CommandStructure(t *testing.T) {
	// Test parent mod command
	if modCmd.Use != "mod" {
		t.Errorf("Unexpected Use: %s", modCmd.Use)
	}

	if modCmd.Short == "" {
		t.Error("Expected Short description to be set")
	}

	// Verify subcommands exist
	subcommands := []string{"init", "list", "get", "download", "tidy", "vendor", "search", "publish"}
	for _, name := range subcommands {
		found := false
		for _, cmd := range modCmd.Commands() {
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

func TestModInitCmd_Structure(t *testing.T) {
	if modInitCmd.Use != "init [module-name]" {
		t.Errorf("Unexpected Use: %s", modInitCmd.Use)
	}
}

func TestModInitCmd_Success(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	// Run mod init
	err := runModInit(modInitCmd, []string{"example.com/test-module"})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Verify pe.mod was created
	if _, err := os.Stat(filepath.Join(tmpDir, "pe.mod")); os.IsNotExist(err) {
		t.Error("Expected pe.mod to be created")
	}
}

func TestModInitCmd_AlreadyExists(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	// Create existing pe.mod
	err := os.WriteFile(filepath.Join(tmpDir, "pe.mod"), []byte("module existing\n"), 0644)
	if err != nil {
		t.Fatalf("Failed to create pe.mod: %v", err)
	}

	// Run mod init without force flag
	modForce = false
	err = runModInit(modInitCmd, []string{"example.com/new-module"})
	if err == nil {
		t.Error("Expected error when pe.mod already exists")
	}
}

func TestModInitCmd_ForceOverwrite(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	// Create existing pe.mod
	err := os.WriteFile(filepath.Join(tmpDir, "pe.mod"), []byte("module existing\n"), 0644)
	if err != nil {
		t.Fatalf("Failed to create pe.mod: %v", err)
	}

	// Run mod init with force flag
	modForce = true
	defer func() { modForce = false }()

	err = runModInit(modInitCmd, []string{"example.com/new-module"})
	if err != nil {
		t.Errorf("Unexpected error with force flag: %v", err)
	}
}

func TestModTidyCmd_NoPeMod(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	// Should error when no pe.mod exists
	err := runModTidy(modTidyCmd, nil)
	if err == nil {
		t.Error("Expected error when pe.mod doesn't exist")
	}
}

func TestModVendorCmd_NoPeMod(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	// Should error when no pe.mod exists
	err := runModVendor(modVendorCmd, nil)
	if err == nil {
		t.Error("Expected error when pe.mod doesn't exist")
	}
}

func TestModDownloadCmd_NoPeMod(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	// Should error when no pe.mod exists
	err := runModDownload(modDownloadCmd, nil)
	if err == nil {
		t.Error("Expected error when pe.mod doesn't exist")
	}
}

func TestModListCmd_Structure(t *testing.T) {
	if modListCmd.Use != "list" {
		t.Errorf("Unexpected Use: %s", modListCmd.Use)
	}
}

func TestModGetCmd_Structure(t *testing.T) {
	if modGetCmd.Use != "get [module]" {
		t.Errorf("Unexpected Use: %s", modGetCmd.Use)
	}
}

func TestModSearchCmd_Structure(t *testing.T) {
	if modSearchCmd.Use != "search [query]" {
		t.Errorf("Unexpected Use: %s", modSearchCmd.Use)
	}
}

func TestModPublishCmd_Structure(t *testing.T) {
	if modPublishCmd.Use != "publish" {
		t.Errorf("Unexpected Use: %s", modPublishCmd.Use)
	}
}
