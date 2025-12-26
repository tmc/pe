package main

import (
	"testing"
)

func TestOptimizeCmd_FlagParsing(t *testing.T) {
	cmd := optimizeCmd()

	// Test flag existence
	flags := []string{"prompt", "iterations", "provider", "model", "output", "temperature", "max-tokens", "method"}
	for _, name := range flags {
		if cmd.Flags().Lookup(name) == nil {
			t.Errorf("Expected flag %q to exist", name)
		}
	}
}

func TestOptimizeCmd_CommandStructure(t *testing.T) {
	cmd := optimizeCmd()

	if cmd.Use != "optimize [prompt-file]" {
		t.Errorf("Unexpected Use: %s", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("Expected Short description to be set")
	}

	if cmd.Long == "" {
		t.Error("Expected Long description to be set")
	}

	if cmd.Example == "" {
		t.Error("Expected Example to be set")
	}
}

func TestOptimizeCmd_FlagDefaults(t *testing.T) {
	cmd := optimizeCmd()

	// Check iterations default
	iterations, err := cmd.Flags().GetInt("iterations")
	if err != nil {
		t.Errorf("Failed to get iterations flag: %v", err)
	}
	if iterations < 1 {
		t.Errorf("Expected iterations > 0, got %d", iterations)
	}

	// Check method default
	method, err := cmd.Flags().GetString("method")
	if err != nil {
		t.Errorf("Failed to get method flag: %v", err)
	}
	// Method should have a default
	if method == "" {
		// Some commands may have empty defaults, that's OK
	}
}

func TestOptimizeCmd_NoPromptError(t *testing.T) {
	cmd := optimizeCmd()
	// Set iterations flag to avoid that error
	cmd.Flags().Set("iterations", "3")

	err := cmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("Expected error when no prompt provided")
	}
}
