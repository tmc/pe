package main

import (
	"testing"
)

func TestViewCmd_FlagParsing(t *testing.T) {
	cmd := viewCmd()

	// Test flag existence
	if cmd.Flags().Lookup("file") == nil {
		t.Error("Expected --file flag to exist")
	}
	if cmd.Flags().Lookup("port") == nil {
		t.Error("Expected --port flag to exist")
	}
	if cmd.Flags().Lookup("yes") == nil {
		t.Error("Expected --yes flag to exist")
	}
}

func TestViewCmd_CommandStructure(t *testing.T) {
	cmd := viewCmd()

	if cmd.Use != "view [evalId]" {
		t.Errorf("Unexpected Use: %s", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("Expected Short description to be set")
	}

	if cmd.Long == "" {
		t.Error("Expected Long description to be set")
	}
}

func TestViewCmd_FlagDefaults(t *testing.T) {
	cmd := viewCmd()

	// Check port default
	port, err := cmd.Flags().GetInt("port")
	if err != nil {
		t.Errorf("Failed to get port flag: %v", err)
	}
	if port != 8080 {
		t.Errorf("Expected default port 8080, got %d", port)
	}
}
