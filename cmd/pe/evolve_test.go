package main

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestEvolveCmd_FlagParsing(t *testing.T) {
	cmd := evolveCmd()

	flags := []string{
		"provider", "generations", "population",
		"mutation-rate", "crossover-rate", "elitism-rate",
		"objectives", "output", "verbose",
	}

	for _, name := range flags {
		if cmd.Flags().Lookup(name) == nil {
			t.Errorf("Expected flag %q to exist", name)
		}
	}
}

func TestEvolveCmd_CommandStructure(t *testing.T) {
	cmd := evolveCmd()

	if cmd.Use != "evolve [prompt-file]" {
		t.Errorf("Unexpected Use: %s", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("Expected Short description to be set")
	}

	if cmd.Long == "" {
		t.Error("Expected Long description to be set")
	}
}

func TestEvolveCmd_FlagDefaults(t *testing.T) {
	cmd := evolveCmd()

	// Check generations default
	generations, err := cmd.Flags().GetInt("generations")
	if err != nil {
		t.Errorf("Failed to get generations flag: %v", err)
	}
	if generations != 20 {
		t.Errorf("Expected default generations 20, got %d", generations)
	}

	// Check population default
	population, err := cmd.Flags().GetInt("population")
	if err != nil {
		t.Errorf("Failed to get population flag: %v", err)
	}
	if population != 10 {
		t.Errorf("Expected default population 10, got %d", population)
	}

	// Check mutation rate default
	mutationRate, err := cmd.Flags().GetFloat64("mutation-rate")
	if err != nil {
		t.Errorf("Failed to get mutation-rate flag: %v", err)
	}
	if mutationRate != 0.3 {
		t.Errorf("Expected default mutation-rate 0.3, got %f", mutationRate)
	}

	// Check crossover rate default
	crossoverRate, err := cmd.Flags().GetFloat64("crossover-rate")
	if err != nil {
		t.Errorf("Failed to get crossover-rate flag: %v", err)
	}
	if crossoverRate != 0.7 {
		t.Errorf("Expected default crossover-rate 0.7, got %f", crossoverRate)
	}

	// Check elitism rate default
	elitismRate, err := cmd.Flags().GetFloat64("elitism-rate")
	if err != nil {
		t.Errorf("Failed to get elitism-rate flag: %v", err)
	}
	if elitismRate != 0.2 {
		t.Errorf("Expected default elitism-rate 0.2, got %f", elitismRate)
	}

	// Check objectives default
	objectives, err := cmd.Flags().GetString("objectives")
	if err != nil {
		t.Errorf("Failed to get objectives flag: %v", err)
	}
	if objectives != "effectiveness,clarity,conciseness" {
		t.Errorf("Expected default objectives 'effectiveness,clarity,conciseness', got %s", objectives)
	}
}

func TestRunEvolve_NoArgs(t *testing.T) {
	cmd := evolveCmd()
	err := runEvolve(cmd, []string{})
	if err == nil {
		t.Error("Expected error for no args")
	}
}

func TestRunEvolve_FileNotFound(t *testing.T) {
	cmd := evolveCmd()
	err := runEvolve(cmd, []string{"nonexistent-file.txt"})
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
}

func TestEvolveCmdOutputDeniedByWritePolicy(t *testing.T) {
	oldTestMode := os.Getenv("PE_TEST_MODE")
	os.Setenv("PE_TEST_MODE", "true")
	defer os.Setenv("PE_TEST_MODE", oldTestMode)

	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	if err := os.WriteFile("prompt.txt", []byte("Summarize clearly."), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("pe.mod", []byte(evolvePolicyTestModule()), 0644); err != nil {
		t.Fatal(err)
	}

	cmd := evolveCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{
		"prompt.txt",
		"--provider", "mock",
		"--generations", "1",
		"--population", "2",
		"--output", "evolution.json",
	})

	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "tool write is denied") {
		t.Fatalf("evolve error = %v, want write policy denial", err)
	}
	if _, err := os.Stat("evolution.json"); !os.IsNotExist(err) {
		t.Fatalf("evolve wrote output despite write policy: %v", err)
	}
}

func TestEvolveCmdStdoutAllowedByWritePolicy(t *testing.T) {
	oldTestMode := os.Getenv("PE_TEST_MODE")
	os.Setenv("PE_TEST_MODE", "true")
	defer os.Setenv("PE_TEST_MODE", oldTestMode)

	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	if err := os.WriteFile("prompt.txt", []byte("Summarize clearly."), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("pe.mod", []byte(evolvePolicyTestModule()), 0644); err != nil {
		t.Fatal(err)
	}

	cmd := evolveCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{
		"prompt.txt",
		"--provider", "mock",
		"--generations", "1",
		"--population", "2",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("evolve stdout: %v", err)
	}
	if _, err := os.Stat("evolution.json"); !os.IsNotExist(err) {
		t.Fatalf("evolve created unexpected output: %v", err)
	}
}

func evolvePolicyTestModule() string {
	return `module example.com/prompts

pe 1

capability {
    tools deny write
}
`
}
