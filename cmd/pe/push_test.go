package main

import (
	"path/filepath"
	"testing"
)

func TestModuleFilePathRejectsTraversal(t *testing.T) {
	moduleDir := t.TempDir()
	tests := []string{
		"../secret.prompt",
		filepath.Join("..", "secret.prompt"),
		filepath.Join(t.TempDir(), "secret.prompt"),
	}
	for _, name := range tests {
		t.Run(name, func(t *testing.T) {
			if _, err := moduleFilePath(moduleDir, name); err == nil {
				t.Fatalf("moduleFilePath(%q) succeeded, want error", name)
			}
		})
	}
}
