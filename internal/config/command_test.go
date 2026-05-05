package config

import (
	"testing"
	"time"
)

func TestCommandConfigValidation(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Commands["run"] = CommandConfig{
		Provider: "openai",
		Timeout:  10 * time.Second,
		Options:  map[string]interface{}{"temperature": 0.2},
	}
	if err := ValidateConfig(cfg); err != nil {
		t.Fatalf("ValidateConfig: %v", err)
	}
}

func TestCommandConfigRejectsNegativeTimeout(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Commands["run"] = CommandConfig{Timeout: -time.Second}
	if err := ValidateConfig(cfg); err == nil {
		t.Fatal("ValidateConfig accepted negative command timeout")
	}
}

func TestCommandConfigSetOverride(t *testing.T) {
	manager, err := NewManager(WithConfigPaths())
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}
	if err := manager.Set("commands.run.provider", "anthropic"); err != nil {
		t.Fatalf("Set command provider: %v", err)
	}
	cfg := manager.Get()
	if cfg.Commands["run"].Provider != "anthropic" {
		t.Fatalf("commands.run.provider = %q, want anthropic", cfg.Commands["run"].Provider)
	}
}
