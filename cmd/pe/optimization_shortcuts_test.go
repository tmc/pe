package main

import "testing"

func TestOptimizationShortcutCommands(t *testing.T) {
	tests := []struct {
		name string
		cmd  string
	}{
		{name: "textgrad", cmd: "textgrad"},
		{name: "pe2", cmd: "pe2"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := optimizeMethodCmd(tt.cmd, "test")
			if got := cmd.Name(); got != tt.cmd {
				t.Fatalf("Name = %q, want %q", got, tt.cmd)
			}
			flag := cmd.Flags().Lookup("method")
			if flag == nil {
				t.Fatal("method flag not found")
			}
			if got := flag.Value.String(); got != tt.cmd {
				t.Fatalf("method = %q, want %q", got, tt.cmd)
			}
			if !flag.Hidden {
				t.Fatal("method flag is visible")
			}
		})
	}
}
