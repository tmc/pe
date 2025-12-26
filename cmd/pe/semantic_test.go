package main

import (
	"testing"
)

func TestSemanticCmd_CommandStructure(t *testing.T) {
	cmd := semanticCmd()

	if cmd.Use != "semantic" {
		t.Errorf("Unexpected Use: %s", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("Expected Short description to be set")
	}

	// Verify subcommands exist
	subcommandNames := []string{"backprop", "descent", "gaso", "flow", "gradients", "monitor", "analyze", "benchmark"}
	for _, name := range subcommandNames {
		found := false
		for _, c := range cmd.Commands() {
			if c.Name() == name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected subcommand %q to exist", name)
		}
	}
}

func TestSemanticBackpropCmd_FlagParsing(t *testing.T) {
	cmd := semanticBackpropCmd()

	flags := []string{"prompt", "prompt-file", "objective", "iterations", "provider", "model", "output", "format", "verbose"}
	for _, name := range flags {
		if cmd.Flags().Lookup(name) == nil {
			t.Errorf("Expected flag %q to exist", name)
		}
	}
}

func TestSemanticBackpropCmd_CommandStructure(t *testing.T) {
	cmd := semanticBackpropCmd()

	if cmd.Use != "backprop" {
		t.Errorf("Unexpected Use: %s", cmd.Use)
	}
}

func TestSemanticDescentCmd_FlagParsing(t *testing.T) {
	cmd := semanticDescentCmd()

	flags := []string{"prompt", "prompt-file", "objective", "learning-rate", "iterations", "convergence", "adaptive", "provider", "model", "output", "format"}
	for _, name := range flags {
		if cmd.Flags().Lookup(name) == nil {
			t.Errorf("Expected flag %q to exist", name)
		}
	}
}

func TestSemanticDescentCmd_FlagDefaults(t *testing.T) {
	cmd := semanticDescentCmd()

	// Check learning rate default
	learningRate, err := cmd.Flags().GetFloat64("learning-rate")
	if err != nil {
		t.Errorf("Failed to get learning-rate flag: %v", err)
	}
	if learningRate != 0.1 {
		t.Errorf("Expected default learning-rate 0.1, got %f", learningRate)
	}

	// Check iterations default
	iterations, err := cmd.Flags().GetInt("iterations")
	if err != nil {
		t.Errorf("Failed to get iterations flag: %v", err)
	}
	if iterations != 10 {
		t.Errorf("Expected default iterations 10, got %d", iterations)
	}

	// Check convergence default
	convergence, err := cmd.Flags().GetFloat64("convergence")
	if err != nil {
		t.Errorf("Failed to get convergence flag: %v", err)
	}
	if convergence != 0.001 {
		t.Errorf("Expected default convergence 0.001, got %f", convergence)
	}
}

func TestGASOCmd_FlagParsing(t *testing.T) {
	cmd := gasoCmd()

	flags := []string{"system", "objective", "iterations", "provider", "model", "output", "format", "graph", "type", "multi-objective"}
	for _, name := range flags {
		if cmd.Flags().Lookup(name) == nil {
			t.Errorf("Expected flag %q to exist", name)
		}
	}
}

func TestGASOCmd_CommandStructure(t *testing.T) {
	cmd := gasoCmd()

	if cmd.Use != "gaso" {
		t.Errorf("Unexpected Use: %s", cmd.Use)
	}
}

func TestSemanticFlowCmd_FlagParsing(t *testing.T) {
	cmd := semanticFlowCmd()

	flags := []string{"system", "output", "format"}
	for _, name := range flags {
		if cmd.Flags().Lookup(name) == nil {
			t.Errorf("Expected flag %q to exist", name)
		}
	}
}

func TestSemanticGradientsCmd_FlagParsing(t *testing.T) {
	cmd := semanticGradientsCmd()

	flags := []string{"prompt", "prompt-file", "visualize", "output"}
	for _, name := range flags {
		if cmd.Flags().Lookup(name) == nil {
			t.Errorf("Expected flag %q to exist", name)
		}
	}
}

func TestSemanticMonitorCmd_FlagParsing(t *testing.T) {
	cmd := semanticMonitorCmd()

	flags := []string{"baseline", "current"}
	for _, name := range flags {
		if cmd.Flags().Lookup(name) == nil {
			t.Errorf("Expected flag %q to exist", name)
		}
	}
}

func TestSemanticAnalyzeCmd_FlagParsing(t *testing.T) {
	cmd := semanticAnalyzeCmd()

	flags := []string{"system", "dependencies"}
	for _, name := range flags {
		if cmd.Flags().Lookup(name) == nil {
			t.Errorf("Expected flag %q to exist", name)
		}
	}
}

func TestSemanticBenchmarkCmd_FlagParsing(t *testing.T) {
	cmd := semanticBenchmarkCmd()

	flags := []string{"prompt", "prompt-file", "baselines"}
	for _, name := range flags {
		if cmd.Flags().Lookup(name) == nil {
			t.Errorf("Expected flag %q to exist", name)
		}
	}
}

func TestLoadPromptContent(t *testing.T) {
	tests := []struct {
		name       string
		prompt     string
		promptFile string
		wantErr    bool
	}{
		{
			name:    "direct prompt",
			prompt:  "test prompt",
			wantErr: false,
		},
		{
			name:       "file not found",
			promptFile: "nonexistent.txt",
			wantErr:    true,
		},
		{
			name:    "no input",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := loadPromptContent(tt.prompt, tt.promptFile)
			if tt.wantErr && err == nil {
				t.Error("Expected error")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestLoadSystemDefinition(t *testing.T) {
	tests := []struct {
		name       string
		systemFile string
		wantErr    bool
	}{
		{
			name:       "empty file path",
			systemFile: "",
			wantErr:    true,
		},
		{
			name:       "file not found",
			systemFile: "nonexistent.yaml",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := loadSystemDefinition(tt.systemFile)
			if tt.wantErr && err == nil {
				t.Error("Expected error")
			}
		})
	}
}
