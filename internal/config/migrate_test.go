package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMigrateFile(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "promptfooconfig.yaml")
	dst := filepath.Join(dir, ".pe", "config.yaml")
	if err := os.WriteFile(src, []byte("providers:\n  default: anthropic\n"), 0644); err != nil {
		t.Fatalf("write source: %v", err)
	}

	if err := MigrateFile(src, dst); err != nil {
		t.Fatalf("MigrateFile: %v", err)
	}

	data, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("read migrated config: %v", err)
	}
	if !strings.Contains(string(data), "default: anthropic") {
		t.Fatalf("migrated config = %s, want anthropic default", data)
	}
}

func TestMigrateFileRequiresPaths(t *testing.T) {
	if err := MigrateFile("", "out.yaml"); err == nil {
		t.Fatal("MigrateFile with empty source succeeded")
	}
	if err := MigrateFile("in.yaml", ""); err == nil {
		t.Fatal("MigrateFile with empty destination succeeded")
	}
}
