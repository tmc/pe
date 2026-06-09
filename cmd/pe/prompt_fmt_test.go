package main

import (
	"os"
	"strings"
	"testing"
)

func TestFormatPromptFileWriteDeniedByPolicy(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	if err := os.WriteFile("prompt.txt", []byte("Summarize this"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("pe.mod", []byte(promptFmtPolicyTestModule()), 0644); err != nil {
		t.Fatal(err)
	}

	err := formatPromptFile("prompt.txt", true, false, "anthropic", false)
	if err == nil || !strings.Contains(err.Error(), "tool write is denied") {
		t.Fatalf("formatPromptFile error = %v, want write policy denial", err)
	}
	data, err := os.ReadFile("prompt.txt")
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "Summarize this" {
		t.Fatalf("prompt content changed despite write policy: %q", data)
	}
}

func TestFormatPromptFileCheckAllowedByPolicy(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	if err := os.WriteFile("prompt.txt", []byte("Summarize this"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("pe.mod", []byte(promptFmtPolicyTestModule()), 0644); err != nil {
		t.Fatal(err)
	}

	err := formatPromptFile("prompt.txt", false, true, "anthropic", false)
	if err == nil || !strings.Contains(err.Error(), "needs formatting") {
		t.Fatalf("formatPromptFile check error = %v, want formatting issue", err)
	}
	data, err := os.ReadFile("prompt.txt")
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "Summarize this" {
		t.Fatalf("prompt content changed in check mode: %q", data)
	}
}

func promptFmtPolicyTestModule() string {
	return `module example.com/prompts

pe 1

capability {
    tools deny write
}
`
}
