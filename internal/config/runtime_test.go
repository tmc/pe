package config

import (
	"testing"
	"time"
)

func TestConfigCommand(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Commands["run"] = CommandConfig{Provider: "mock", Timeout: time.Second}

	got, ok := cfg.Command("run")
	if !ok {
		t.Fatal("Command(run) ok = false")
	}
	if got.Provider != "mock" || got.Timeout != time.Second {
		t.Fatalf("Command(run) = %+v", got)
	}
}

func TestProviderOptions(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Providers.OpenAI.APIKey = "key"
	cfg.Providers.OpenAI.BaseURL = "https://example.test/v1"
	cfg.Providers.OpenAI.DefaultModel = "test-model"
	cfg.Providers.OpenAI.Timeout = 5 * time.Second
	cfg.Providers.OpenAI.MaxRetries = 7

	got := cfg.ProviderOptions("openai")
	tests := map[string]interface{}{
		"api_key":     "key",
		"base_url":    "https://example.test/v1",
		"model":       "test-model",
		"timeout":     5 * time.Second,
		"max_retries": 7,
	}
	for key, want := range tests {
		if got[key] != want {
			t.Fatalf("ProviderOptions(openai)[%s] = %v, want %v", key, got[key], want)
		}
	}
}
