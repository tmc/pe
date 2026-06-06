package main

import (
	"strings"
	"testing"
)

func TestExpCommandsRegistry(t *testing.T) {
	// List of all expected subcommands under 'pe exp'
	expectedCmds := []string{
		"distributed",
		"attest",
		"cache",
		"compose", // Moved command
		"transform",
		"merge",
		"optimize",
		"consensus",
		"batch",
		"sweep",
		"schedule",
		"trace",
		"explain",
		"lint",
		"import",
		"export",
		"sync",
		"workflow",
		"hook",
		"report",
	}

	for _, cmdName := range expectedCmds {
		t.Run(cmdName, func(t *testing.T) {
			// 1. Verify Command Registration
			cmd, _, err := expCmd.Find([]string{cmdName})
			if err != nil {
				t.Fatalf("Failed to find subcommand %q: %v", cmdName, err)
			}
			if cmd.Name() != cmdName {
				t.Errorf("Expected command name %q, got %q", cmdName, cmd.Name())
			}

			// 2. Verify Execution (Stub Output)
			// Skip real commands with flags/args requirements.
			if cmdName == "compose" || cmdName == "distributed" || cmdName == "optimize" || cmdName == "consensus" {
				return
			}
			if cmd.RunE == nil {
				return
			}

			err = cmd.RunE(cmd, []string{})
			if err == nil || !strings.Contains(err.Error(), "not yet implemented") {
				t.Errorf("Command %q error = %v, want not implemented", cmdName, err)
			}
		})
	}
}
