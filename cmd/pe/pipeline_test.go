package main

import (
	"testing"
)

func TestAskCmd_FlagParsing(t *testing.T) {
	cmd := askCmd()

	// Test flag existence
	if cmd.Flags().Lookup("template") == nil {
		t.Error("Expected --template flag to exist")
	}
	if cmd.Flags().Lookup("provider") == nil {
		t.Error("Expected --provider flag to exist")
	}
	if cmd.Flags().Lookup("parallel") == nil {
		t.Error("Expected --parallel flag to exist")
	}
}

func TestAskCmd_CommandStructure(t *testing.T) {
	cmd := askCmd()

	if cmd.Use != "ask [prompt]" {
		t.Errorf("Unexpected Use: %s", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("Expected Short description to be set")
	}
}

func TestCollectCmd_CommandStructure(t *testing.T) {
	cmd := collectCmd()

	if cmd.Use != "collect" {
		t.Errorf("Unexpected Use: %s", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("Expected Short description to be set")
	}
}

func TestReduceCmd_CommandStructure(t *testing.T) {
	cmd := reduceCmd()

	if cmd.Use != "reduce" {
		t.Errorf("Unexpected Use: %s", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("Expected Short description to be set")
	}
}

func TestFilterCmd_CommandStructure(t *testing.T) {
	cmd := filterCmd()

	if cmd.Use != "filter" {
		t.Errorf("Unexpected Use: %s", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("Expected Short description to be set")
	}
}
