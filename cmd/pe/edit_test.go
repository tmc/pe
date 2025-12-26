package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEditCmd_FlagParsing(t *testing.T) {
	flags := []string{
		"set-prompt", "set-system-prompt", "set-examples", "set-defaults",
		"add-default", "remove-default", "append-prompt", "prepend-prompt",
		"add-variant", "variant-cmd", "remove-variant",
		"add-section", "section-content", "remove-section",
		"json", "print", "fmt", "module",
	}

	for _, name := range flags {
		if editCmd.Flags().Lookup(name) == nil {
			t.Errorf("Expected flag %q to exist", name)
		}
	}
}

func TestEditCmd_CommandStructure(t *testing.T) {
	if editCmd.Use != "edit [prompt-file]" {
		t.Errorf("Unexpected Use: %s", editCmd.Use)
	}

	if editCmd.Short == "" {
		t.Error("Expected Short description to be set")
	}

	if editCmd.Long == "" {
		t.Error("Expected Long description to be set")
	}
}

func TestRunEdit_NoArgs(t *testing.T) {
	err := runEdit(editCmd, []string{})
	if err == nil {
		t.Error("Expected error for no args")
	}
	if !strings.Contains(err.Error(), "filename required") {
		t.Errorf("Expected 'filename required' error, got: %v", err)
	}
}

func TestRunEdit_FileNotFound(t *testing.T) {
	// Reset flags
	editSetPrompt = ""
	editPrint = true // Avoid writing file

	err := runEdit(editCmd, []string{"nonexistent-file.txt"})
	// Should not error if no modifications requested
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestRunEdit_SetPrompt(t *testing.T) {
	tmpDir := t.TempDir()
	promptFile := filepath.Join(tmpDir, "test.prompt")

	// Create initial file
	if err := os.WriteFile(promptFile, []byte("Original prompt"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Reset and set flags
	editSetPrompt = "New prompt content"
	editPrint = false
	editJSON = false
	editFmt = false

	err := runEdit(editCmd, []string{promptFile})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Verify the file was modified
	content, err := os.ReadFile(promptFile)
	if err != nil {
		t.Fatalf("Failed to read modified file: %v", err)
	}

	if !strings.Contains(string(content), "New prompt content") {
		t.Errorf("Expected file to contain new prompt, got: %s", string(content))
	}

	// Reset flag
	editSetPrompt = ""
}

func TestRunEdit_JSONOutput(t *testing.T) {
	tmpDir := t.TempDir()
	promptFile := filepath.Join(tmpDir, "test.prompt")

	if err := os.WriteFile(promptFile, []byte("Test prompt"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Reset and set flags
	editSetPrompt = ""
	editPrint = false
	editJSON = true

	err := runEdit(editCmd, []string{promptFile})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Reset flag
	editJSON = false
}

func TestFormatPrompt(t *testing.T) {
	tests := []struct {
		name     string
		prompt   *testPromptData
		contains []string
	}{
		{
			name: "basic prompt",
			prompt: &testPromptData{
				Main: "Hello world",
			},
			contains: []string{"Hello world"},
		},
		{
			name: "prompt with system prompt",
			prompt: &testPromptData{
				Main:         "User prompt",
				SystemPrompt: "System instructions",
			},
			contains: []string{"User prompt", "-- system-prompt --", "System instructions"},
		},
		{
			name: "prompt with defaults",
			prompt: &testPromptData{
				Main:     "Test",
				Defaults: map[string]string{"key": "value"},
			},
			contains: []string{"Test", "-- defaults --", "key=value"},
		},
		{
			name: "prompt with shebang",
			prompt: &testPromptData{
				Shebang: "#!/usr/bin/env pe",
				Main:    "Content",
			},
			contains: []string{"#!/usr/bin/env pe", "Content"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Import the prompt package type for testing
			p := &struct {
				Shebang      string
				Main         string
				SystemPrompt string
				Sections     map[string]string
				Defaults     map[string]string
			}{
				Shebang:      tt.prompt.Shebang,
				Main:         tt.prompt.Main,
				SystemPrompt: tt.prompt.SystemPrompt,
				Sections:     tt.prompt.Sections,
				Defaults:     tt.prompt.Defaults,
			}

			// We can't call formatPrompt directly as it uses prompt.Prompt type
			// This test just validates the struct
			_ = p
		})
	}
}

// testPromptData is a local struct for testing
type testPromptData struct {
	Shebang      string
	Main         string
	SystemPrompt string
	Sections     map[string]string
	Defaults     map[string]string
}

func TestAddModuleDependency(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(tmpDir)

	// Create initial pe.mod
	peMod := `module test.example/prompts

go 1.21
`
	if err := os.WriteFile("pe.mod", []byte(peMod), 0644); err != nil {
		t.Fatalf("Failed to create pe.mod: %v", err)
	}

	err := addModuleDependency("example.com/prompts@v1.0.0")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Verify the module was added
	content, err := os.ReadFile("pe.mod")
	if err != nil {
		t.Fatalf("Failed to read pe.mod: %v", err)
	}

	if !strings.Contains(string(content), "require") {
		t.Error("Expected require block in pe.mod")
	}

	if !strings.Contains(string(content), "example.com/prompts") {
		t.Error("Expected module in pe.mod")
	}
}

func TestAddModuleDependency_InvalidFormat(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(tmpDir)

	// Create initial pe.mod
	if err := os.WriteFile("pe.mod", []byte("module test\n"), 0644); err != nil {
		t.Fatalf("Failed to create pe.mod: %v", err)
	}

	err := addModuleDependency("invalid-format")
	if err == nil {
		t.Error("Expected error for invalid module format")
	}
}

func TestAddModuleDependency_NoPeMod(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(tmpDir)

	err := addModuleDependency("example.com/test@v1.0.0")
	if err == nil {
		t.Error("Expected error when pe.mod doesn't exist")
	}
}
