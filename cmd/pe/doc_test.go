package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDocCmd_FlagParsing(t *testing.T) {
	cmd := docCmd()

	// Test flag existence
	allFlag := cmd.Flags().Lookup("all")
	if allFlag == nil {
		t.Error("Expected --all flag to exist")
	}

	shortFlag := cmd.Flags().Lookup("short")
	if shortFlag == nil {
		t.Error("Expected --short flag to exist")
	}

	examplesFlag := cmd.Flags().Lookup("examples")
	if examplesFlag == nil {
		t.Error("Expected --examples flag to exist")
	}

	// Test flag types
	if allFlag != nil && allFlag.Value.Type() != "bool" {
		t.Errorf("Expected --all flag to be bool, got %s", allFlag.Value.Type())
	}

	if shortFlag != nil && shortFlag.Value.Type() != "bool" {
		t.Errorf("Expected --short flag to be bool, got %s", shortFlag.Value.Type())
	}
}

func TestDocCmd_CommandStructure(t *testing.T) {
	cmd := docCmd()

	if cmd.Use != "doc [prompt] [variable]" {
		t.Errorf("Unexpected Use: %s", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("Expected Short description to be set")
	}

	if cmd.Long == "" {
		t.Error("Expected Long description to be set")
	}
}

func TestDocCmd_NonexistentPrompt(t *testing.T) {
	// Create a temp directory
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	// Test with a nonexistent prompt
	err := runDoc(docCmd(), []string{"nonexistent"})
	if err == nil {
		t.Error("Expected error for nonexistent prompt")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("Expected 'not found' error, got: %v", err)
	}
}

func TestDocCmd_InvalidPath(t *testing.T) {
	// Test with invalid path format
	err := runDoc(docCmd(), []string{"a.b.c"})
	if err == nil {
		t.Error("Expected error for invalid path")
	}
	if !strings.Contains(err.Error(), "invalid") {
		t.Errorf("Expected 'invalid' error, got: %v", err)
	}
}

func TestDocCmd_ShowVariableDoc_NonexistentVariable(t *testing.T) {
	// Create a temp directory with a prompt file
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	// Create a test prompt file
	promptContent := `Hello {{.name}}`
	err := os.WriteFile(filepath.Join(tmpDir, "greeting.prompt"), []byte(promptContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test prompt: %v", err)
	}

	// Test with nonexistent variable
	err = showVariableDoc("greeting", "nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent variable")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("Expected 'not found' error, got: %v", err)
	}
}

func TestDocCmd_ShowPromptDoc_Success(t *testing.T) {
	// Create a temp directory with a prompt file
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	// Create a test prompt file
	promptContent := `Analyze the {{.text}} carefully.

-- summary --
Analysis prompt.
`
	err := os.WriteFile(filepath.Join(tmpDir, "analyzer.prompt"), []byte(promptContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test prompt: %v", err)
	}

	// Should not error for existing prompt
	err = showPromptDoc("analyzer")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestDocCmd_ListDocumentedPrompts_NoPrompts(t *testing.T) {
	// Create a temp directory with no prompt files
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	// Should not error when no prompts found
	err := listDocumentedPrompts()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestDocCmd_ListDocumentedPrompts_WithPrompts(t *testing.T) {
	// Create a temp directory with prompt files
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	// Create test prompt files
	for _, name := range []string{"alpha.prompt", "beta.prompt"} {
		err := os.WriteFile(filepath.Join(tmpDir, name), []byte("Test prompt"), 0644)
		if err != nil {
			t.Fatalf("Failed to create test prompt: %v", err)
		}
	}

	// Should not error when prompts exist
	err := listDocumentedPrompts()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestIndentText(t *testing.T) {
	tests := []struct {
		name   string
		text   string
		prefix string
		want   string
	}{
		{
			name:   "single line",
			text:   "hello",
			prefix: "  ",
			want:   "  hello",
		},
		{
			name:   "multiple lines",
			text:   "line1\nline2\nline3",
			prefix: "    ",
			want:   "    line1\n    line2\n    line3",
		},
		{
			name:   "empty lines preserved",
			text:   "line1\n\nline3",
			prefix: "> ",
			want:   "> line1\n\n> line3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := indentText(tt.text, tt.prefix)
			if got != tt.want {
				t.Errorf("indentText() = %q, want %q", got, tt.want)
			}
		})
	}
}
