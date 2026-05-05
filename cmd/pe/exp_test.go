package main

import (
	"bytes"
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
			if cmdName == "compose" || cmdName == "optimize" {
				return
			}

			// Capture output
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)

			// Execute the command
			cmd.Run(cmd, []string{})

			output := buf.String()
			expectedOutput := "is an experimental prototype"
			if !strings.Contains(output, expectedOutput) {
				t.Errorf("Command %q output does not contain expected stub message.\nGot: %s\nExpected to contain: %s", cmdName, output, expectedOutput)
			}
		})
	}
}
