package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tmc/pe/internal/prompt"
)

func TestGetCmd_FlagParsing(t *testing.T) {
	if getCmd.Flags().Lookup("json") == nil {
		t.Error("Expected --json flag to exist")
	}
	if getCmd.Flags().Lookup("variant") == nil {
		t.Error("Expected --variant flag to exist")
	}
	if getCmd.Flags().Lookup("keys") == nil {
		t.Error("Expected --keys flag to exist")
	}
}

func TestGetCmd_CommandStructure(t *testing.T) {
	if getCmd.Use != "get [prompt-file] [field]" {
		t.Errorf("Unexpected Use: %s", getCmd.Use)
	}

	if getCmd.Short == "" {
		t.Error("Expected Short description to be set")
	}

	if getCmd.Long == "" {
		t.Error("Expected Long description to be set")
	}
}

func TestRunGet_FileNotFound(t *testing.T) {
	err := runGet(getCmd, []string{"nonexistent-file.txt", "prompt"})
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
}

func TestRunGet_ValidFile(t *testing.T) {
	tmpDir := t.TempDir()
	promptFile := filepath.Join(tmpDir, "test.prompt")

	content := "Hello {{.name}}!"
	if err := os.WriteFile(promptFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	err := runGet(getCmd, []string{promptFile, "prompt"})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestRunGet_Variables(t *testing.T) {
	tmpDir := t.TempDir()
	promptFile := filepath.Join(tmpDir, "test.prompt")

	content := "Hello {{.name}}, you are {{.age}} years old!"
	if err := os.WriteFile(promptFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	err := runGet(getCmd, []string{promptFile, "variables"})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestRunGetFieldsAndModes(t *testing.T) {
	oldJSON, oldVariant, oldKeys := getJSON, getVariant, getKeys
	defer func() {
		getJSON, getVariant, getKeys = oldJSON, oldVariant, oldKeys
	}()

	promptFile := filepath.Join(t.TempDir(), "test.prompt")
	content := strings.Join([]string{
		"Hello {{.name}}",
		"-- system-prompt --",
		"System {{.role}}",
		"-- defaults --",
		"name: Ada",
		"-- examples --",
		"Example text",
		"-- notes --",
		"Custom section",
		"-- variant:brief --",
		"extend-system-prompt brief mode",
		"append-prompt final line",
		"prepend-prompt start line",
	}, "\n")
	if err := os.WriteFile(promptFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	for _, tt := range []struct {
		name    string
		field   string
		json    bool
		keys    bool
		variant string
	}{
		{"all", "all", false, false, ""},
		{"json all", "all", true, false, ""},
		{"system", "system-prompt", false, false, ""},
		{"defaults", "defaults", true, false, ""},
		{"examples", "examples", false, false, ""},
		{"variants", "variants", true, false, ""},
		{"section", "notes", false, false, ""},
		{"keys", "all", false, true, ""},
		{"variant", "prompt", false, false, "brief"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			getJSON = tt.json
			getKeys = tt.keys
			getVariant = tt.variant
			if err := runGet(getCmd, []string{promptFile, tt.field}); err != nil {
				t.Fatal(err)
			}
		})
	}

	getJSON, getKeys, getVariant = false, false, ""
	if err := runGet(getCmd, []string{promptFile, "missing"}); err == nil {
		t.Fatal("unknown field succeeded")
	}
}

func TestGetHelpers(t *testing.T) {
	p := &prompt.Prompt{
		Main:         "Hello {{.name}}",
		SystemPrompt: "System {{.role}}",
		Defaults:     map[string]string{"name": "Ada"},
		Sections: map[string]string{
			"notes":         "Use {{.tone}}",
			"variant:brief": "extend-system-prompt brief\nappend-prompt done\nprepend-prompt start\nset-flag ignored",
		},
		Shebang: "#!/usr/bin/env pe",
	}
	vars := extractVariables(p)
	for _, want := range []string{"name", "role", "tone"} {
		if !getTestContains(vars, want) {
			t.Fatalf("vars = %#v, want %q", vars, want)
		}
	}
	if variants := getVariants(p); len(variants) != 1 || variants[0] != "brief" {
		t.Fatalf("variants = %#v", variants)
	}
	applied := applyVariant(p, "brief")
	if !strings.HasPrefix(applied.Main, "start\n") || !strings.Contains(applied.Main, "\ndone") || !strings.Contains(applied.SystemPrompt, "brief") {
		t.Fatalf("applied = %#v", applied)
	}
	info := getPromptInfo(applied)
	if info.Prompt == "" || info.Sections["notes"] == "" || info.Sections["variant:brief"] != "" {
		t.Fatalf("info = %#v", info)
	}
	if err := listKeys(applied); err != nil {
		t.Fatal(err)
	}
}

func TestRunGetModule(t *testing.T) {
	oldwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	tmp := t.TempDir()
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(oldwd)

	if err := os.WriteFile("pe.mod", []byte("module example\nrequire (\n  a/b v1.0.0\n  // comment\n)\nrequire c/d v1.2.3\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := runGet(getCmd, nil); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove("pe.mod"); err != nil {
		t.Fatal(err)
	}
	if err := runGet(getCmd, nil); err == nil {
		t.Fatal("missing pe.mod succeeded")
	}
}

func getTestContains(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}
