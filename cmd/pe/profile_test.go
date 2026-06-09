package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
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

func TestProfileStartDeniedByWritePolicy(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	writeProfileDenyWritePeMod(t)
	cmd := profileStartCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	err := runProfileStart(cmd, []string{"cpu"}, "profiles", "")
	if err == nil || !strings.Contains(err.Error(), "tool write is denied") {
		t.Fatalf("runProfileStart error = %v, want write policy denial", err)
	}
	if _, err := os.Stat("profiles"); !os.IsNotExist(err) {
		t.Fatalf("profiles stat error = %v, want not exist", err)
	}
}

func TestProfileReportOutputDeniedByWritePolicy(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	writeProfileDenyWritePeMod(t)
	cmd := profileReportCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	err := runProfileReport(cmd, "report.json", "json")
	if err == nil || !strings.Contains(err.Error(), "tool write is denied") {
		t.Fatalf("runProfileReport error = %v, want write policy denial", err)
	}
	if _, err := os.Stat("report.json"); !os.IsNotExist(err) {
		t.Fatalf("report.json stat error = %v, want not exist", err)
	}
}

func TestProfileReportStdoutAllowedByWritePolicy(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	writeProfileDenyWritePeMod(t)
	cmd := profileReportCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	if err := runProfileReport(cmd, "", "json"); err != nil {
		t.Fatalf("runProfileReport stdout: %v", err)
	}
	if !strings.Contains(out.String(), "timestamp") {
		t.Fatalf("stdout = %q, want report JSON", out.String())
	}
}

func TestTraceAndMetricsStartDeniedByWritePolicy(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	writeProfileDenyWritePeMod(t)
	cmd := profileCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	if err := runTraceStart(cmd, "traces.jsonl"); err == nil || !strings.Contains(err.Error(), "tool write is denied") {
		t.Fatalf("runTraceStart error = %v, want write policy denial", err)
	}
	if _, err := os.Stat("traces.jsonl"); !os.IsNotExist(err) {
		t.Fatalf("traces.jsonl stat error = %v, want not exist", err)
	}
	if err := runMetricsStart(cmd, "metrics.jsonl", "10s"); err == nil || !strings.Contains(err.Error(), "tool write is denied") {
		t.Fatalf("runMetricsStart error = %v, want write policy denial", err)
	}
	if _, err := os.Stat("metrics.jsonl"); !os.IsNotExist(err) {
		t.Fatalf("metrics.jsonl stat error = %v, want not exist", err)
	}
}

func TestTraceAndMetricsReportOutputDeniedByWritePolicy(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	writeProfileDenyWritePeMod(t)
	traceFile := filepath.Join(tmpDir, "trace-input.jsonl")
	if err := os.WriteFile(traceFile, nil, 0644); err != nil {
		t.Fatalf("writing trace input: %v", err)
	}
	cmd := profileCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	if err := runTraceReport(cmd, traceFile, "trace-report.json", "json"); err == nil || !strings.Contains(err.Error(), "tool write is denied") {
		t.Fatalf("runTraceReport error = %v, want write policy denial", err)
	}
	if _, err := os.Stat("trace-report.json"); !os.IsNotExist(err) {
		t.Fatalf("trace-report.json stat error = %v, want not exist", err)
	}
	if err := runMetricsReport(cmd, "metrics-report.json", "json"); err == nil || !strings.Contains(err.Error(), "tool write is denied") {
		t.Fatalf("runMetricsReport error = %v, want write policy denial", err)
	}
	if _, err := os.Stat("metrics-report.json"); !os.IsNotExist(err) {
		t.Fatalf("metrics-report.json stat error = %v, want not exist", err)
	}
}

func writeProfileDenyWritePeMod(t *testing.T) {
	t.Helper()
	if err := os.WriteFile("pe.mod", []byte(`module example.com/app

pe 1

capability {
    tools deny write
}
`), 0644); err != nil {
		t.Fatalf("writing pe.mod: %v", err)
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
