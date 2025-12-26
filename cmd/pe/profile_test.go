package main

import (
	"testing"
)

func TestProfileCmd_CommandStructure(t *testing.T) {
	cmd := profileCmd()

	if cmd.Use != "profile" {
		t.Errorf("Unexpected Use: %s", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("Expected Short description to be set")
	}

	// Verify subcommands exist
	subcommandNames := []string{"start", "stop", "status", "report", "trace", "metrics"}
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

func TestProfileStartCmd_FlagParsing(t *testing.T) {
	cmd := profileStartCmd()

	flags := []string{"type", "output-dir", "duration"}
	for _, name := range flags {
		if cmd.Flags().Lookup(name) == nil {
			t.Errorf("Expected flag %q to exist", name)
		}
	}
}

func TestProfileStartCmd_FlagDefaults(t *testing.T) {
	cmd := profileStartCmd()

	// Check output-dir default
	outputDir, err := cmd.Flags().GetString("output-dir")
	if err != nil {
		t.Errorf("Failed to get output-dir flag: %v", err)
	}
	if outputDir != "profiles" {
		t.Errorf("Expected default output-dir 'profiles', got %s", outputDir)
	}
}

func TestProfileStopCmd_FlagParsing(t *testing.T) {
	cmd := profileStopCmd()

	if cmd.Flags().Lookup("type") == nil {
		t.Error("Expected --type flag to exist")
	}
}

func TestProfileStatusCmd_CommandStructure(t *testing.T) {
	cmd := profileStatusCmd()

	if cmd.Use != "status" {
		t.Errorf("Unexpected Use: %s", cmd.Use)
	}
}

func TestProfileReportCmd_FlagParsing(t *testing.T) {
	cmd := profileReportCmd()

	flags := []string{"output", "format"}
	for _, name := range flags {
		if cmd.Flags().Lookup(name) == nil {
			t.Errorf("Expected flag %q to exist", name)
		}
	}
}

func TestProfileReportCmd_FlagDefaults(t *testing.T) {
	cmd := profileReportCmd()

	format, err := cmd.Flags().GetString("format")
	if err != nil {
		t.Errorf("Failed to get format flag: %v", err)
	}
	if format != "json" {
		t.Errorf("Expected default format 'json', got %s", format)
	}
}

func TestTraceCmd_CommandStructure(t *testing.T) {
	cmd := traceCmd()

	if cmd.Use != "trace" {
		t.Errorf("Unexpected Use: %s", cmd.Use)
	}

	// Verify subcommands
	subcommandNames := []string{"start", "stop", "report"}
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

func TestTraceStartCmd_FlagParsing(t *testing.T) {
	cmd := traceStartCmd()

	if cmd.Flags().Lookup("output") == nil {
		t.Error("Expected --output flag to exist")
	}
}

func TestTraceReportCmd_FlagParsing(t *testing.T) {
	cmd := traceReportCmd()

	flags := []string{"trace-file", "output", "format"}
	for _, name := range flags {
		if cmd.Flags().Lookup(name) == nil {
			t.Errorf("Expected flag %q to exist", name)
		}
	}
}

func TestProfileMetricsCmd_CommandStructure(t *testing.T) {
	cmd := profileMetricsCmd()

	if cmd.Use != "metrics" {
		t.Errorf("Unexpected Use: %s", cmd.Use)
	}

	// Verify subcommands
	subcommandNames := []string{"start", "stop", "report"}
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

func TestMetricsStartCmd_FlagParsing(t *testing.T) {
	cmd := metricsStartCmd()

	flags := []string{"output", "interval"}
	for _, name := range flags {
		if cmd.Flags().Lookup(name) == nil {
			t.Errorf("Expected flag %q to exist", name)
		}
	}
}

func TestMetricsStartCmd_FlagDefaults(t *testing.T) {
	cmd := metricsStartCmd()

	output, err := cmd.Flags().GetString("output")
	if err != nil {
		t.Errorf("Failed to get output flag: %v", err)
	}
	if output != "metrics.jsonl" {
		t.Errorf("Expected default output 'metrics.jsonl', got %s", output)
	}

	interval, err := cmd.Flags().GetString("interval")
	if err != nil {
		t.Errorf("Failed to get interval flag: %v", err)
	}
	if interval != "10s" {
		t.Errorf("Expected default interval '10s', got %s", interval)
	}
}

func TestMetricsReportCmd_FlagParsing(t *testing.T) {
	cmd := metricsReportCmd()

	flags := []string{"output", "format"}
	for _, name := range flags {
		if cmd.Flags().Lookup(name) == nil {
			t.Errorf("Expected flag %q to exist", name)
		}
	}
}
