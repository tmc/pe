package main

import "testing"

func TestExpCommandsRegistry(t *testing.T) {
	expectedCmds := []string{
		"distributed",
		"attest",
		"cache",
		"compose",
		"optimize",
		"consensus",
	}

	for _, cmdName := range expectedCmds {
		t.Run(cmdName, func(t *testing.T) {
			cmd, _, err := expCmd.Find([]string{cmdName})
			if err != nil {
				t.Fatalf("Failed to find subcommand %q: %v", cmdName, err)
			}
			if cmd.Name() != cmdName {
				t.Errorf("Expected command name %q, got %q", cmdName, cmd.Name())
			}

		})
	}
}
