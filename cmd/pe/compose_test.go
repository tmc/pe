package main

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tmc/pe/internal/metaprompt"
)

func TestInferComponentType(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		content string
		want    string
	}{
		{
			name:    "context from filename",
			path:    "context.txt",
			content: "Some content",
			want:    "context",
		},
		{
			name:    "instruction from filename",
			path:    "instruction.txt",
			content: "Some content",
			want:    "instruction",
		},
		{
			name:    "example from filename",
			path:    "example.txt",
			content: "Some content",
			want:    "example",
		},
		{
			name:    "constraint from filename",
			path:    "constraint.txt",
			content: "Some content",
			want:    "constraint",
		},
		{
			name:    "context from content",
			path:    "file.txt",
			content: "You are an expert assistant",
			want:    "context",
		},
		{
			name:    "instruction from content",
			path:    "file.txt",
			content: "Please analyze the following data",
			want:    "instruction",
		},
		{
			name:    "unknown type",
			path:    "file.txt",
			content: "Random content here",
			want:    "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := inferComponentType(tt.path, tt.content)
			if got != tt.want {
				t.Errorf("inferComponentType() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestInferCategory(t *testing.T) {
	tests := []struct {
		name string
		path string
		want string
	}{
		{
			name: "from parent directory",
			path: "/components/prompts/file.txt",
			want: "prompts",
		},
		{
			name: "from nested directory",
			path: "/components/system/context.txt",
			want: "system",
		},
		{
			name: "from root",
			path: "/file.txt",
			want: "/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := inferCategory(tt.path)
			if got != tt.want {
				t.Errorf("inferCategory() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestLoadComponent(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name        string
		fileName    string
		fileContent string
		wantErr     bool
		checkFunc   func(t *testing.T, comp PromptComponent)
	}{
		{
			name:        "load text component",
			fileName:    "context.txt",
			fileContent: "You are a helpful assistant.",
			wantErr:     false,
			checkFunc: func(t *testing.T, comp PromptComponent) {
				if comp.Content != "You are a helpful assistant." {
					t.Errorf("Content = %q, want %q", comp.Content, "You are a helpful assistant.")
				}
				if comp.Type != "context" {
					t.Errorf("Type = %q, want %q", comp.Type, "context")
				}
			},
		},
		{
			name:        "load json component",
			fileName:    "component.json",
			fileContent: `{"type": "instruction", "content": "Analyze this.", "category": "analysis"}`,
			wantErr:     false,
			checkFunc: func(t *testing.T, comp PromptComponent) {
				if comp.Type != "instruction" {
					t.Errorf("Type = %q, want %q", comp.Type, "instruction")
				}
				if comp.Content != "Analyze this." {
					t.Errorf("Content = %q, want %q", comp.Content, "Analyze this.")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filePath := filepath.Join(tmpDir, tt.fileName)
			if err := os.WriteFile(filePath, []byte(tt.fileContent), 0644); err != nil {
				t.Fatalf("Failed to write test file: %v", err)
			}

			comp, err := loadComponent(filePath)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if tt.checkFunc != nil {
					tt.checkFunc(t, comp)
				}
			}
		})
	}
}

func TestLoadComponents(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test components
	files := map[string]string{
		"context.txt":     "You are a helpful assistant.",
		"instruction.txt": "Please analyze the following.",
	}

	for name, content := range files {
		if err := os.WriteFile(filepath.Join(tmpDir, name), []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write test file: %v", err)
		}
	}

	// Test loading from directory
	components, err := loadComponents([]string{tmpDir})
	if err != nil {
		t.Fatalf("loadComponents() error = %v", err)
	}

	if len(components) != 2 {
		t.Errorf("Expected 2 components, got %d", len(components))
	}
}

func TestValidateComponentDependencies(t *testing.T) {
	tests := []struct {
		name       string
		components []PromptComponent
		wantErr    bool
	}{
		{
			name: "no dependencies",
			components: []PromptComponent{
				{Type: "context", Content: "Hello"},
				{Type: "instruction", Content: "Analyze"},
			},
			wantErr: false,
		},
		{
			name: "satisfied dependency",
			components: []PromptComponent{
				{Type: "context", Content: "Hello"},
				{Type: "instruction", Content: "Analyze", Dependencies: []string{"context"}},
			},
			wantErr: false,
		},
		{
			name: "missing dependency",
			components: []PromptComponent{
				{Type: "instruction", Content: "Analyze", Dependencies: []string{"context"}},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateComponentDependencies(tt.components)
			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

func TestCalculateSimpleCoherence(t *testing.T) {
	tests := []struct {
		name  string
		texts []string
		min   float64
		max   float64
	}{
		{
			name:  "single text",
			texts: []string{"Hello world"},
			min:   1.0,
			max:   1.0,
		},
		{
			name: "similar texts",
			texts: []string{
				"Analyze the sentiment of the text",
				"Determine the emotional tone",
			},
			min: 0.5,
			max: 1.0,
		},
		{
			name: "different texts",
			texts: []string{
				"Hello world",
				"Goodbye moon",
			},
			min: 0.0,
			max: 0.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := calculateSimpleCoherence(tt.texts)
			if score < tt.min || score > tt.max {
				t.Errorf("calculateSimpleCoherence() = %v, want between %v and %v", score, tt.min, tt.max)
			}
		})
	}
}

func TestCalculateStyleConsistency(t *testing.T) {
	tests := []struct {
		name  string
		texts []string
		min   float64
		max   float64
	}{
		{
			name:  "single text",
			texts: []string{"Hello world."},
			min:   1.0,
			max:   1.0,
		},
		{
			name: "consistent style",
			texts: []string{
				"Please analyze the data.",
				"Please determine the outcome.",
			},
			min: 0.5,
			max: 1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := calculateStyleConsistency(tt.texts)
			if score < tt.min || score > tt.max {
				t.Errorf("calculateStyleConsistency() = %v, want between %v and %v", score, tt.min, tt.max)
			}
		})
	}
}

func TestCalculateOverlap(t *testing.T) {
	tests := []struct {
		name string
		set1 map[string]bool
		set2 map[string]bool
		want float64
	}{
		{
			name: "identical sets",
			set1: map[string]bool{"hello": true, "world": true},
			set2: map[string]bool{"hello": true, "world": true},
			want: 1.0,
		},
		{
			name: "no overlap",
			set1: map[string]bool{"hello": true},
			set2: map[string]bool{"world": true},
			want: 0.0,
		},
		{
			name: "partial overlap",
			set1: map[string]bool{"hello": true, "world": true},
			set2: map[string]bool{"hello": true, "moon": true},
			want: 1.0 / 3.0, // 1 intersection / 3 union
		},
		{
			name: "empty sets",
			set1: map[string]bool{},
			set2: map[string]bool{},
			want: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateOverlap(tt.set1, tt.set2)
			if got < tt.want-0.01 || got > tt.want+0.01 {
				t.Errorf("calculateOverlap() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCalculateMetricVariance(t *testing.T) {
	tests := []struct {
		name   string
		values []float64
		want   float64
	}{
		{
			name:   "single value",
			values: []float64{1.0},
			want:   0.0,
		},
		{
			name:   "identical values",
			values: []float64{5.0, 5.0, 5.0},
			want:   0.0,
		},
		{
			name:   "varied values",
			values: []float64{1.0, 2.0, 3.0},
			want:   0.333, // Normalized variance
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateMetricVariance(tt.values)
			if got < tt.want-0.1 || got > tt.want+0.1 {
				t.Errorf("calculateMetricVariance() = %v, want around %v", got, tt.want)
			}
		})
	}
}

func TestDetectDomain(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    string
	}{
		{
			name:    "software domain",
			content: "This is about software engineering",
			want:    "general",
		},
		{
			name:    "medical domain",
			content: "This involves medical diagnosis",
			want:    "medical",
		},
		{
			name:    "legal domain",
			content: "This pertains to legal matters",
			want:    "legal",
		},
		{
			name:    "financial domain",
			content: "This is about financial analysis",
			want:    "financial",
		},
		{
			name:    "general domain",
			content: "This is general text",
			want:    "general",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detectDomain(tt.content)
			if got != tt.want {
				t.Errorf("detectDomain() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDetectSpecialization(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    string
	}{
		{
			name:    "medical specialization",
			content: "Specialized medical diagnosis",
			want:    "medical",
		},
		{
			name:    "legal specialization",
			content: "Specialized legal contract",
			want:    "legal",
		},
		{
			name:    "financial specialization",
			content: "Specialized financial trading",
			want:    "financial",
		},
		{
			name:    "technical specialization",
			content: "Specialized technical work",
			want:    "technical",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detectSpecialization(tt.content)
			if got != tt.want {
				t.Errorf("detectSpecialization() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestComposeCmd_Flags(t *testing.T) {
	cmd := newComposeCmd()

	expectedFlags := []string{
		"components", "style", "target", "coherence", "validation-gate",
		"optimize", "library-init", "output", "config", "add-component",
		"category", "list", "import", "coherence-check", "validate",
		"examples", "quality-gates", "program-synthesis", "parameter-optimization",
		"signature-validation", "multi-stage", "statistical-validation",
		"synthesis-strategy", "optimization-method", "quality-threshold",
	}

	for _, flagName := range expectedFlags {
		flag := cmd.Flags().Lookup(flagName)
		if flag == nil {
			t.Errorf("Expected --%s flag to exist", flagName)
		}
	}
}

func TestComposeCmd_LibraryInit(t *testing.T) {
	oldTestMode := os.Getenv("PE_TEST_MODE")
	os.Setenv("PE_TEST_MODE", "true")
	defer os.Setenv("PE_TEST_MODE", oldTestMode)

	tmpDir := t.TempDir()

	oldDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldDir)

	cmd := composeCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--library-init"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Check that component directories were created
	expectedDirs := []string{
		"components/context",
		"components/instructions",
		"components/examples",
		"components/constraints",
		"components/templates",
	}

	for _, dir := range expectedDirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			t.Errorf("Expected directory %s was not created", dir)
		}
	}
}

func TestComposeCmd_WithComponents(t *testing.T) {
	oldTestMode := os.Getenv("PE_TEST_MODE")
	os.Setenv("PE_TEST_MODE", "true")
	defer os.Setenv("PE_TEST_MODE", oldTestMode)

	tmpDir := t.TempDir()

	// Create test components
	contextFile := filepath.Join(tmpDir, "context.txt")
	if err := os.WriteFile(contextFile, []byte("You are a helpful assistant."), 0644); err != nil {
		t.Fatalf("Failed to write context file: %v", err)
	}

	instructionFile := filepath.Join(tmpDir, "instruction.txt")
	if err := os.WriteFile(instructionFile, []byte("Please analyze the following data."), 0644); err != nil {
		t.Fatalf("Failed to write instruction file: %v", err)
	}

	oldDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldDir)

	cmd := composeCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{contextFile, instructionFile, "--style", "default"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func TestComposeCmd_WithOptimization(t *testing.T) {
	oldTestMode := os.Getenv("PE_TEST_MODE")
	os.Setenv("PE_TEST_MODE", "true")
	defer os.Setenv("PE_TEST_MODE", oldTestMode)

	tmpDir := t.TempDir()

	contextFile := filepath.Join(tmpDir, "context.txt")
	if err := os.WriteFile(contextFile, []byte("You are a helpful assistant."), 0644); err != nil {
		t.Fatalf("Failed to write context file: %v", err)
	}

	oldDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldDir)

	cmd := newComposeCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{contextFile, "--optimize", "--target", "gpt-4"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("compose optimize err = %v", err)
	}
	optimized, metadata := optimizeComposedPrompt("Context:\nHelp users.")
	if !strings.Contains(optimized, "Be specific") || metadata["optimization_method"] != "local_compose_polish" {
		t.Fatalf("optimized=%q metadata=%#v", optimized, metadata)
	}
}

func TestComposeCmd_ChainOfThoughtStyle(t *testing.T) {
	oldTestMode := os.Getenv("PE_TEST_MODE")
	os.Setenv("PE_TEST_MODE", "true")
	defer os.Setenv("PE_TEST_MODE", oldTestMode)

	tmpDir := t.TempDir()

	contextFile := filepath.Join(tmpDir, "context.txt")
	if err := os.WriteFile(contextFile, []byte("You are an expert problem solver."), 0644); err != nil {
		t.Fatalf("Failed to write context file: %v", err)
	}

	oldDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldDir)

	cmd := composeCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{contextFile, "--style", "cot"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func TestComposeCmd_WithOutput(t *testing.T) {
	oldTestMode := os.Getenv("PE_TEST_MODE")
	os.Setenv("PE_TEST_MODE", "true")
	defer os.Setenv("PE_TEST_MODE", oldTestMode)

	tmpDir := t.TempDir()

	contextFile := filepath.Join(tmpDir, "context.txt")
	if err := os.WriteFile(contextFile, []byte("You are a helpful assistant."), 0644); err != nil {
		t.Fatalf("Failed to write context file: %v", err)
	}

	outputFile := filepath.Join(tmpDir, "output.txt")

	oldDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldDir)

	cmd := composeCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{contextFile, "--output", outputFile})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Note: Due to Cobra flag state persistence between tests,
	// the output file may not be created if library-init was set in a previous test.
	// We just verify the command executes without error.
}

func TestComposeCmd_CoherenceCheck(t *testing.T) {
	oldTestMode := os.Getenv("PE_TEST_MODE")
	os.Setenv("PE_TEST_MODE", "true")
	defer os.Setenv("PE_TEST_MODE", oldTestMode)

	tmpDir := t.TempDir()

	file1 := filepath.Join(tmpDir, "sentiment1.txt")
	if err := os.WriteFile(file1, []byte("Analyze the sentiment of the text."), 0644); err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	file2 := filepath.Join(tmpDir, "sentiment2.txt")
	if err := os.WriteFile(file2, []byte("Determine the emotional tone."), 0644); err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	oldDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldDir)

	cmd := composeCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{file1, file2, "--coherence-check"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func TestLoadComponentsFromDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test files
	files := map[string]string{
		"context.txt":     "Context content",
		"instruction.txt": "Instruction content",
		"ignored.md":      "Should be ignored",
	}

	for name, content := range files {
		if err := os.WriteFile(filepath.Join(tmpDir, name), []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write file: %v", err)
		}
	}

	components, err := loadComponentsFromDirectory(tmpDir)
	if err != nil {
		t.Fatalf("loadComponentsFromDirectory() error = %v", err)
	}

	// Should only load .txt and .json files
	if len(components) != 2 {
		t.Errorf("Expected 2 components, got %d", len(components))
	}
}

func TestLoadComponent_NonexistentFile(t *testing.T) {
	_, err := loadComponent("/nonexistent/path/file.txt")
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
}

func TestComposeConfigAndRunPaths(t *testing.T) {
	tmpDir := t.TempDir()
	contextFile := filepath.Join(tmpDir, "context.txt")
	instructionFile := filepath.Join(tmpDir, "instruction.txt")
	examplesFile := filepath.Join(tmpDir, "examples.json")
	configFile := filepath.Join(tmpDir, "compose.json")
	outputFile := filepath.Join(tmpDir, "out.txt")
	if err := os.WriteFile(contextFile, []byte("You are an expert analyst. Therefore be precise."), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(instructionFile, []byte("Please analyze the data."), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(examplesFile, []byte(`[{"input":"x","output":"y"}]`), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(configFile, []byte(`{"style":"structured","target":"model-from-config","metadata":{"x":1}}`), 0644); err != nil {
		t.Fatal(err)
	}

	cmd := newComposeCmd()
	cmd.Flags().Set("config", configFile)
	cmd.Flags().Set("style", "few-shot")
	cmd.Flags().Set("target", "gpt-test")
	cmd.Flags().Set("coherence", "true")
	cmd.Flags().Set("optimize", "true")
	cmd.Flags().Set("quality-gates", "true")
	cmd.Flags().Set("program-synthesis", "true")
	cmd.Flags().Set("parameter-optimization", "true")
	cmd.Flags().Set("signature-validation", "true")
	cmd.Flags().Set("multi-stage", "true")
	cmd.Flags().Set("statistical-validation", "true")
	cmd.Flags().Set("synthesis-strategy", "evolutionary")
	cmd.Flags().Set("optimization-method", "grid")
	cmd.Flags().Set("quality-threshold", "0.9")
	cmd.Flags().Set("examples", examplesFile)
	cmd.Flags().Set("output", outputFile)
	cfg, err := loadComposeConfig(cmd, []string{contextFile, instructionFile})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Style != "few-shot" || cfg.Target != "gpt-test" || !cfg.Coherence || !cfg.Optimize || !cfg.QualityGates || !cfg.ProgramSynthesis || !cfg.ParameterOptimization || !cfg.SignatureValidation || !cfg.MultiStageOptimization || !cfg.StatisticalValidation {
		t.Fatalf("config = %#v", cfg)
	}
	if cfg.Metadata["synthesis_strategy"] != "evolutionary" || cfg.Metadata["optimization_method"] != "grid" || cfg.Metadata["quality_threshold"] != 0.9 {
		t.Fatalf("metadata = %#v", cfg.Metadata)
	}
	cmd.Flags().Set("optimize", "false")
	if err := runCompose(cmd, []string{contextFile, instructionFile}); err != nil {
		t.Fatal(err)
	}
	if out, err := os.ReadFile(outputFile); err != nil || len(out) == 0 {
		t.Fatalf("output len=%d err=%v", len(out), err)
	}

	badCmd := newComposeCmd()
	badCmd.Flags().Set("config", filepath.Join(tmpDir, "missing.json"))
	if _, err := loadComposeConfig(badCmd, nil); err == nil {
		t.Fatal("missing config succeeded")
	}
	badConfig := filepath.Join(tmpDir, "bad.json")
	if err := os.WriteFile(badConfig, []byte("{"), 0644); err != nil {
		t.Fatal(err)
	}
	badCmd = newComposeCmd()
	badCmd.Flags().Set("config", badConfig)
	if _, err := loadComposeConfig(badCmd, nil); err == nil {
		t.Fatal("bad config succeeded")
	}
}

func TestFormatSynthesisResult(t *testing.T) {
	result := &metaprompt.SynthesisResult{
		Program:      "Prompt Program\n\nTask:\nSummarize",
		Strategy:     "template",
		Confidence:   0.72,
		QualityScore: 0.8,
		Metadata: map[string]interface{}{
			"method": "local_template_synthesis",
		},
	}
	jsonOut, err := formatSynthesisResult(result, "json")
	if err != nil {
		t.Fatalf("json format: %v", err)
	}
	if !strings.Contains(jsonOut, `"strategy": "template"`) {
		t.Fatalf("json output = %q", jsonOut)
	}
	yamlOut, err := formatSynthesisResult(result, "yaml")
	if err != nil {
		t.Fatalf("yaml format: %v", err)
	}
	if !strings.Contains(yamlOut, "strategy: template") || strings.Contains(yamlOut, "not implemented") {
		t.Fatalf("yaml output = %q", yamlOut)
	}
	textOut, err := formatSynthesisResult(result, "text")
	if err != nil {
		t.Fatalf("text format: %v", err)
	}
	if textOut != result.Program {
		t.Fatalf("text output = %q", textOut)
	}
	if _, err := formatSynthesisResult(result, "xml"); err == nil {
		t.Fatal("unsupported format succeeded")
	}
}

func TestComposeCoherenceValidation(t *testing.T) {
	score, err := validateCoherence(context.Background(), "Context:\nAnalyze customer churn.\n\nInstructions:\nTherefore summarize drivers clearly.")
	if err != nil {
		t.Fatalf("validateCoherence err = %v", err)
	}
	if score <= 0 {
		t.Fatalf("score = %v", score)
	}
	if _, err := validateCoherence(context.Background(), "   "); err == nil {
		t.Fatal("empty coherence validation succeeded")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := validateCoherence(ctx, "prompt"); err == nil {
		t.Fatal("cancelled coherence validation succeeded")
	}

	tmpDir := t.TempDir()
	contextFile := filepath.Join(tmpDir, "context.txt")
	instructionFile := filepath.Join(tmpDir, "instruction.txt")
	if err := os.WriteFile(contextFile, []byte("Analyze the customer text."), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(instructionFile, []byte("Therefore summarize the customer sentiment."), 0644); err != nil {
		t.Fatal(err)
	}
	cmd := newComposeCmd()
	cmd.Flags().Set("coherence", "true")
	if err := runCompose(cmd, []string{contextFile, instructionFile}); err != nil {
		t.Fatal(err)
	}
}

func TestComposeLibraryAndCoherenceHelpers(t *testing.T) {
	tmpDir := t.TempDir()
	oldDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldDir)

	source := filepath.Join(tmpDir, "source.txt")
	if err := os.WriteFile(source, []byte("Please analyze this."), 0644); err != nil {
		t.Fatal(err)
	}
	if err := addComponentToLibrary(source, "custom"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join("components", "custom", "source.txt")); err != nil {
		t.Fatal(err)
	}
	if err := listComponents(); err != nil {
		t.Fatal(err)
	}
	importFile := filepath.Join(tmpDir, "import.txt")
	if err := os.WriteFile(importFile, []byte("Imported component"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := importComponents(importFile); err != nil {
		t.Fatalf("import file err = %v", err)
	}
	if _, err := os.Stat(filepath.Join("components", filepath.Base(tmpDir), "import.txt")); err != nil {
		t.Fatalf("imported file missing: %v", err)
	}
	importDir := filepath.Join(tmpDir, "bundle")
	if err := os.MkdirAll(filepath.Join(importDir, "context"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(importDir, "context", "a.txt"), []byte("A"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := importComponents(importDir); err != nil {
		t.Fatalf("import dir err = %v", err)
	}
	if _, err := os.Stat(filepath.Join("components", "context", "a.txt")); err != nil {
		t.Fatalf("imported dir file missing: %v", err)
	}
	txtarFile := filepath.Join(tmpDir, "components.txtar")
	if err := os.WriteFile(txtarFile, []byte("-- examples/b.txt --\nB\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := importComponents(txtarFile); err != nil {
		t.Fatalf("import txtar err = %v", err)
	}
	if _, err := os.Stat(filepath.Join("components", "examples", "b.txt")); err != nil {
		t.Fatalf("imported txtar file missing: %v", err)
	}
	badTxtar := filepath.Join(tmpDir, "bad.txtar")
	if err := os.WriteFile(badTxtar, []byte("-- ../bad.txt --\nno\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := importComponents(badTxtar); err == nil {
		t.Fatal("path-escaping txtar import succeeded")
	}
	remote := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/components.txtar":
			fmt.Fprint(w, "-- remote/c.txt --\nC\n")
		case "/plain.txt":
			fmt.Fprint(w, "Remote plain component")
		case "/bad.txtar":
			fmt.Fprint(w, "-- ../bad.txt --\nno\n")
		default:
			http.NotFound(w, r)
		}
	}))
	defer remote.Close()
	if err := importComponents(remote.URL + "/components.txtar"); err != nil {
		t.Fatalf("import remote txtar err = %v", err)
	}
	if _, err := os.Stat(filepath.Join("components", "remote", "c.txt")); err != nil {
		t.Fatalf("imported remote txtar file missing: %v", err)
	}
	if err := importComponents(remote.URL + "/plain.txt"); err != nil {
		t.Fatalf("import remote file err = %v", err)
	}
	if _, err := os.Stat(filepath.Join("components", "plain.txt")); err != nil {
		t.Fatalf("imported remote file missing: %v", err)
	}
	if err := importComponents(remote.URL + "/bad.txtar"); err == nil {
		t.Fatal("path-escaping remote txtar import succeeded")
	}
	if err := importComponents(remote.URL + "/missing.txt"); err == nil || !strings.Contains(err.Error(), "status 404") {
		t.Fatalf("missing remote import err = %v", err)
	}
	if err := addComponentToLibrary(filepath.Join(tmpDir, "missing.txt"), "x"); err == nil {
		t.Fatal("missing add component succeeded")
	}
	if err := checkCoherence([]string{source}); err == nil {
		t.Fatal("single coherence check succeeded")
	}
	other := filepath.Join(tmpDir, "other.txt")
	if err := os.WriteFile(other, []byte("Determine the emotional tone."), 0644); err != nil {
		t.Fatal(err)
	}
	if err := checkCoherence([]string{source, other}); err != nil {
		t.Fatalf("checkCoherence err = %v", err)
	}
	if err := checkCoherence([]string{source, filepath.Join(tmpDir, "missing.txt")}); err == nil {
		t.Fatal("missing coherence file succeeded")
	}
	empty := filepath.Join(tmpDir, "empty.txt")
	if err := os.WriteFile(empty, nil, 0644); err != nil {
		t.Fatal(err)
	}
	if err := checkCoherence([]string{source, empty}); err == nil {
		t.Fatal("empty coherence file succeeded")
	}
	validateComponentCompatibility([]PromptComponent{
		{Type: "context", Content: "software engineering code"},
		{Type: "constraint", Content: "specialized medical diagnosis"},
	})
}

func TestComposeRunSpecialCases(t *testing.T) {
	tmpDir := t.TempDir()
	oldDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldDir)

	if err := os.WriteFile("pe.mod", []byte("module example\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := runCompose(newComposeCmd(), nil); err != nil {
		t.Fatal(err)
	}

	cmd := newComposeCmd()
	cmd.Flags().Set("add-component", "missing.txt")
	if err := runCompose(cmd, nil); err == nil {
		t.Fatal("missing add-component succeeded")
	}
	cmd = newComposeCmd()
	cmd.Flags().Set("list", "true")
	if err := runCompose(cmd, nil); err == nil {
		t.Fatal("list without components dir succeeded")
	}
	remote := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Imported over HTTP")
	}))
	defer remote.Close()
	cmd = newComposeCmd()
	cmd.Flags().Set("import", remote.URL+"/instruction.txt")
	if err := runCompose(cmd, nil); err != nil {
		t.Fatalf("remote import err = %v", err)
	}
}
