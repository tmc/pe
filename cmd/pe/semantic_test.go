package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tmc/pe/internal/metaprompt"
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

	flags := []string{"prompt", "prompt-file", "visualize", "output", "format"}
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

	flags := []string{"system", "dependencies", "output", "format"}
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

func TestSemanticHelpersAndOutput(t *testing.T) {
	tmpDir := t.TempDir()
	promptFile := filepath.Join(tmpDir, "prompt.txt")
	if err := os.WriteFile(promptFile, []byte("file prompt"), 0644); err != nil {
		t.Fatal(err)
	}
	got, err := loadPromptContent("", promptFile)
	if err != nil || got != "file prompt" {
		t.Fatalf("prompt content = %q err=%v", got, err)
	}

	jsonSystem := filepath.Join(tmpDir, "system.json")
	if err := os.WriteFile(jsonSystem, []byte(`{"description":"sys","components":[{"id":"a","name":"A","type":"prompt","content":"Do A"}],"dependencies":[{"from":"a","to":"b","type":"data"}]}`), 0644); err != nil {
		t.Fatal(err)
	}
	system, err := loadSystemDefinition(jsonSystem)
	if err != nil || len(system.Components) != 1 || len(system.Dependencies) != 1 {
		t.Fatalf("json system = %#v err=%v", system, err)
	}
	badJSON := filepath.Join(tmpDir, "bad.json")
	if err := os.WriteFile(badJSON, []byte("{"), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := loadSystemDefinition(badJSON); err == nil {
		t.Fatal("bad json parsed")
	}
	yamlSystem := filepath.Join(tmpDir, "system.yaml")
	if err := os.WriteFile(yamlSystem, []byte(`components:
  analyzer:
    type: prompt
    prompt: "Analyze"
  validator:
    type: checker
    prompt: "Validate"
`), 0644); err != nil {
		t.Fatal(err)
	}
	system, err = loadSystemDefinition(yamlSystem)
	if err != nil || len(system.Components) != 2 {
		t.Fatalf("yaml system = %#v err=%v", system, err)
	}
	if system.Components[0].ID != "analyzer" || system.Components[1].ID != "validator" {
		t.Fatalf("yaml component ids = %#v, want analyzer and validator", system.Components)
	}

	inputYAML := filepath.Join(tmpDir, "input-system.yaml")
	if err := os.WriteFile(inputYAML, []byte(`components:
  input:
    type: interface
    prompt: "Receive input"
  analyzer:
    type: processor
    prompt: "Analyze"
    inputs: [input]
`), 0644); err != nil {
		t.Fatal(err)
	}
	system, err = loadSystemDefinition(inputYAML)
	if err != nil || len(system.Components) != 2 || len(system.Dependencies) != 1 {
		t.Fatalf("input yaml system = %#v err=%v", system, err)
	}
	if system.Dependencies[0].From != "input" || system.Dependencies[0].To != "analyzer" {
		t.Fatalf("input dependency = %#v, want input -> analyzer", system.Dependencies[0])
	}

	semantic := &metaprompt.SemanticResult{InitialScore: 0.1, FinalScore: 0.9, OptimizedPrompt: "better", Iterations: 2, Converged: true}
	outFile := filepath.Join(tmpDir, "semantic.json")
	if err := outputSemanticResult(semantic, outFile, "json"); err != nil {
		t.Fatal(err)
	}
	if data, err := os.ReadFile(outFile); err != nil || !strings.Contains(string(data), "better") {
		t.Fatalf("semantic output = %q err=%v", data, err)
	}
	if err := outputSemanticResult(semantic, "", "table"); err != nil {
		t.Fatal(err)
	}
	semanticYAML := filepath.Join(tmpDir, "semantic.yaml")
	if err := outputSemanticResult(semantic, semanticYAML, "yaml"); err != nil {
		t.Fatal(err)
	}
	if data, err := os.ReadFile(semanticYAML); err != nil || !strings.Contains(string(data), "optimized_prompt: better") {
		t.Fatalf("semantic yaml = %q err=%v", data, err)
	}
	if err := outputSemanticResult(semantic, "", "bad"); err == nil {
		t.Fatal("semantic bad format succeeded")
	}

	gaso := &metaprompt.GASOResult{OptimizedComponents: []metaprompt.SystemComponent{{Name: "a"}}, Iterations: 1, OverallPerformance: 0.8, ParetoEfficient: true}
	gasoFile := filepath.Join(tmpDir, "gaso.json")
	if err := outputGASOResult(gaso, gasoFile, "json"); err != nil {
		t.Fatal(err)
	}
	if data, err := os.ReadFile(gasoFile); err != nil || !strings.Contains(string(data), "overall_performance") {
		t.Fatalf("gaso output = %q err=%v", data, err)
	}
	if err := outputGASOResult(gaso, "", "table"); err != nil {
		t.Fatal(err)
	}
	gasoYAML := filepath.Join(tmpDir, "gaso.yaml")
	if err := outputGASOResult(gaso, gasoYAML, "yaml"); err != nil {
		t.Fatal(err)
	}
	if data, err := os.ReadFile(gasoYAML); err != nil || !strings.Contains(string(data), "overall_performance: 0.8") {
		t.Fatalf("gaso yaml = %q err=%v", data, err)
	}
	if err := outputGASOResult(gaso, "", "bad"); err == nil {
		t.Fatal("gaso bad format succeeded")
	}
}

func TestSemanticLocalSubcommands(t *testing.T) {
	tmpDir := t.TempDir()
	oldDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldDir)

	systemFile := filepath.Join(tmpDir, "system.json")
	if err := os.WriteFile(systemFile, []byte(`{"components":[{"id":"input","name":"input","type":"interface","content":"in"},{"id":"analyzer","name":"analyzer","type":"prompt","content":"analyze"},{"id":"output","name":"output","type":"interface","content":"out"}],"dependencies":[{"from":"input","to":"analyzer","type":"data"},{"from":"analyzer","to":"output","type":"control"}]}`), 0644); err != nil {
		t.Fatal(err)
	}
	promptFile := filepath.Join(tmpDir, "prompt.txt")
	if err := os.WriteFile(promptFile, []byte("Analyze this prompt"), 0644); err != nil {
		t.Fatal(err)
	}
	baselineFile := filepath.Join(tmpDir, "baseline.txt")
	currentFile := filepath.Join(tmpDir, "current.txt")
	if err := os.WriteFile(baselineFile, []byte("old prompt"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(currentFile, []byte("new prompt"), 0644); err != nil {
		t.Fatal(err)
	}

	flow := semanticFlowCmd()
	flow.Flags().Set("system", systemFile)
	flowOut := filepath.Join(tmpDir, "flow.json")
	flow.Flags().Set("output", flowOut)
	if err := flow.RunE(flow, nil); err != nil {
		t.Fatalf("flow error = %v", err)
	}
	var flowResult semanticFlowResult
	flowData, err := os.ReadFile(flowOut)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(flowData, &flowResult); err != nil {
		t.Fatal(err)
	}
	if flowResult.NodeCount != 3 || flowResult.EdgeCount != 2 || len(flowResult.Sources) != 1 || flowResult.Sources[0] != "input" {
		t.Fatalf("flow result = %#v", flowResult)
	}
	htmlOut, err := formatSemanticFlow(&flowResult, "html")
	if err != nil {
		t.Fatalf("html flow: %v", err)
	}
	if !strings.Contains(htmlOut, "<svg") || !strings.Contains(htmlOut, "input -> analyzer") {
		t.Fatalf("html flow = %q", htmlOut)
	}
	gradients := semanticGradientsCmd()
	gradients.Flags().Set("prompt-file", promptFile)
	gradients.Flags().Set("visualize", "true")
	gradientOut := filepath.Join(tmpDir, "gradients.html")
	gradients.Flags().Set("output", gradientOut)
	if err := gradients.RunE(gradients, nil); err != nil {
		t.Fatalf("gradients error = %v", err)
	}
	gradientData, err := os.ReadFile(gradientOut)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(gradientData), "<h1>Semantic Gradients</h1>") {
		t.Fatalf("gradient output = %q", gradientData)
	}
	monitor := semanticMonitorCmd()
	monitor.Flags().Set("baseline", baselineFile)
	monitor.Flags().Set("current", currentFile)
	if err := monitor.RunE(monitor, nil); err == nil || !strings.Contains(err.Error(), "semantic drift monitoring is not yet implemented") {
		t.Fatalf("monitor error = %v", err)
	}
	analyze := semanticAnalyzeCmd()
	analyze.Flags().Set("system", systemFile)
	analyze.Flags().Set("dependencies", "true")
	analyzeOut := filepath.Join(tmpDir, "analysis.yaml")
	analyze.Flags().Set("output", analyzeOut)
	analyze.Flags().Set("format", "yaml")
	if err := analyze.RunE(analyze, nil); err != nil {
		t.Fatalf("analyze error = %v", err)
	}
	analyzeData, err := os.ReadFile(analyzeOut)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(analyzeData), "edge_count: 2") || !strings.Contains(string(analyzeData), "dependencies:") {
		t.Fatalf("analysis output = %q", analyzeData)
	}
	benchmark := semanticBenchmarkCmd()
	benchmark.Flags().Set("prompt-file", promptFile)
	benchmark.Flags().Set("baselines", "gpt4, claude, missing")
	if err := benchmark.RunE(benchmark, nil); err == nil || !strings.Contains(err.Error(), "semantic optimization benchmarking is not yet implemented") {
		t.Fatalf("benchmark error = %v", err)
	}

	badFlow := semanticFlowCmd()
	badFlow.Flags().Set("system", filepath.Join(tmpDir, "missing.json"))
	if err := badFlow.RunE(badFlow, nil); err == nil {
		t.Fatal("bad flow succeeded")
	}
	badGradients := semanticGradientsCmd()
	if err := badGradients.RunE(badGradients, nil); err == nil {
		t.Fatal("bad gradients succeeded")
	}
	badMonitor := semanticMonitorCmd()
	badMonitor.Flags().Set("baseline", filepath.Join(tmpDir, "missing.txt"))
	badMonitor.Flags().Set("current", currentFile)
	if err := badMonitor.RunE(badMonitor, nil); err == nil {
		t.Fatal("bad monitor succeeded")
	}
	badAnalyze := semanticAnalyzeCmd()
	badAnalyze.Flags().Set("system", filepath.Join(tmpDir, "missing.json"))
	if err := badAnalyze.RunE(badAnalyze, nil); err == nil {
		t.Fatal("bad analyze succeeded")
	}
	badBenchmark := semanticBenchmarkCmd()
	if err := badBenchmark.RunE(badBenchmark, nil); err == nil {
		t.Fatal("bad benchmark succeeded")
	}
}

func TestAnalyzeSemanticDependenciesDetectsCycles(t *testing.T) {
	report, err := analyzeSemanticDependencies(&metaprompt.SystemDefinition{
		Components: []metaprompt.SystemComponent{
			{ID: "a", Name: "a"},
			{ID: "b", Name: "b"},
		},
		Dependencies: []metaprompt.ComponentDependency{
			{From: "a", To: "b"},
			{From: "b", To: "a"},
		},
	}, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Cycles) != 1 || strings.Join(report.Cycles[0], " -> ") != "a -> b -> a" {
		t.Fatalf("cycles = %#v", report.Cycles)
	}
	if len(report.Dependencies) != 2 {
		t.Fatalf("dependencies = %#v", report.Dependencies)
	}
	text, err := formatSemanticDependencies(report, "text")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(text, "Cycle: a -> b -> a") {
		t.Fatalf("text = %q", text)
	}
}

func TestAnalyzeLocalSemanticGradients(t *testing.T) {
	report := analyzeLocalSemanticGradients("Summarize the issue clearly.")
	if report.TokenCount == 0 || len(report.Gradients) == 0 {
		t.Fatalf("report = %#v", report)
	}
	if report.Gradients[0].Component != "specificity" {
		t.Fatalf("gradients = %#v", report.Gradients)
	}
	text, err := formatSemanticGradients(report, "text")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(text, "Semantic Gradients") {
		t.Fatalf("text = %q", text)
	}
	if _, err := formatSemanticGradients(report, "bad"); err == nil {
		t.Fatal("bad format succeeded")
	}
}
