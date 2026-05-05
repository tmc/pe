package main

import (
	"path/filepath"
	"testing"
)

func TestResolveRelativeFileRejectsTraversal(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.star")
	tests := []string{
		"../tests.star",
		filepath.Join("..", "tests.star"),
		filepath.Join(t.TempDir(), "tests.star"),
	}
	for _, path := range tests {
		t.Run(path, func(t *testing.T) {
			if _, err := resolveRelativeFile(configPath, path); err == nil {
				t.Fatalf("resolveRelativeFile(%q) succeeded, want error", path)
			}
		})
	}
}
