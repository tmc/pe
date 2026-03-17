package promptfoo

import (
	"testing"

	"sigs.k8s.io/yaml"
)

func TestProviderConfig_UnmarshalStringAndObject(t *testing.T) {
	input := []byte(`
providers:
  - openai:gpt-4
  - id: ollama:qwen3.5:4b
    config:
      raw: true
      seed: 1
      num_predict: 100
`)

	var cfg Config
	if err := yaml.Unmarshal(input, &cfg); err != nil {
		t.Fatalf("Unmarshal() failed: %v", err)
	}
	if len(cfg.Providers) != 2 {
		t.Fatalf("len(Providers) = %d, want 2", len(cfg.Providers))
	}
	if cfg.Providers[0].ID != "openai:gpt-4" {
		t.Fatalf("Providers[0].ID = %q, want openai:gpt-4", cfg.Providers[0].ID)
	}
	if cfg.Providers[0].Config != nil {
		t.Fatalf("Providers[0].Config = %#v, want nil", cfg.Providers[0].Config)
	}
	if cfg.Providers[1].ID != "ollama:qwen3.5:4b" {
		t.Fatalf("Providers[1].ID = %q, want ollama:qwen3.5:4b", cfg.Providers[1].ID)
	}
	if got, ok := cfg.Providers[1].Config["raw"].(bool); !ok || !got {
		t.Fatalf("Providers[1].Config[raw] = %#v, want true", cfg.Providers[1].Config["raw"])
	}
	if got, ok := cfg.Providers[1].Config["seed"].(float64); !ok || got != 1 {
		t.Fatalf("Providers[1].Config[seed] = %#v, want 1", cfg.Providers[1].Config["seed"])
	}
}

func TestProviderConfig_UnmarshalLegacyObject(t *testing.T) {
	input := []byte(`
providers:
  - name: mlx-go-lm:mlx-community/Qwen3-4B-4bit
    executable: mlx-lm-generate
    args:
      - --ignore-chat-template
      - --kv-size
      - "4096"
`)

	var cfg Config
	if err := yaml.Unmarshal(input, &cfg); err != nil {
		t.Fatalf("Unmarshal() failed: %v", err)
	}
	if len(cfg.Providers) != 1 {
		t.Fatalf("len(Providers) = %d, want 1", len(cfg.Providers))
	}
	provider := cfg.Providers[0]
	if provider.ID != "mlx-go-lm:mlx-community/Qwen3-4B-4bit" {
		t.Fatalf("provider.ID = %q", provider.ID)
	}
	args, ok := provider.Config["args"].([]interface{})
	if !ok || len(args) != 3 {
		t.Fatalf("provider.Config[args] = %#v, want 3 entries", provider.Config["args"])
	}
}
