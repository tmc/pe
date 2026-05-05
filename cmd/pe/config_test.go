package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tmc/pe/internal/config"
)

func TestConfigValue(t *testing.T) {
	cfg := config.DefaultConfig()
	got, ok, err := configValue(cfg, "providers.default")
	if err != nil {
		t.Fatalf("configValue: %v", err)
	}
	if !ok {
		t.Fatal("providers.default not found")
	}
	if got != "openai" {
		t.Fatalf("providers.default = %q, want openai", got)
	}
}

func TestConfigCmdValidate(t *testing.T) {
	cmd := configCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"validate"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("config validate: %v", err)
	}
	if strings.TrimSpace(out.String()) != "ok" {
		t.Fatalf("output = %q, want ok", out.String())
	}
}

func TestConfigCmdGet(t *testing.T) {
	cmd := configCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"get", "providers.default"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("config get: %v", err)
	}
	if strings.TrimSpace(out.String()) != "openai" {
		t.Fatalf("output = %q, want openai", out.String())
	}
}

func TestConfigCmdSet(t *testing.T) {
	path := filepath.Join(t.TempDir(), "pe.yaml")
	if err := os.WriteFile(path, []byte("providers:\n  default: openai\n"), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	cmd := configCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"set", "providers.default", "anthropic", "--file", path})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("config set: %v", err)
	}
	if strings.TrimSpace(out.String()) != "ok" {
		t.Fatalf("output = %q, want ok", out.String())
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	if !strings.Contains(string(data), "default: anthropic") {
		t.Fatalf("config file = %s, want anthropic", data)
	}
}

func TestConfigCmdMigrate(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "promptfooconfig.yaml")
	dst := filepath.Join(dir, ".pe", "config.yaml")
	if err := os.WriteFile(src, []byte("providers:\n  default: anthropic\n"), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cmd := configCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"migrate", src, "--out", dst})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("config migrate: %v", err)
	}
	if strings.TrimSpace(out.String()) != "ok" {
		t.Fatalf("output = %q, want ok", out.String())
	}
	data, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("read migrated config: %v", err)
	}
	if !strings.Contains(string(data), "default: anthropic") {
		t.Fatalf("config file = %s, want anthropic", data)
	}
}

func TestConfigCmdDocs(t *testing.T) {
	cmd := configCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"docs"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("config docs: %v", err)
	}
	if !strings.Contains(out.String(), "## `providers.openai.api_key`") {
		t.Fatalf("docs output missing openai api key: %s", out.String())
	}
}
