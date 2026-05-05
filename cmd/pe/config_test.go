package main

import (
	"bytes"
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
