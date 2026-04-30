package providers

import (
	"context"
	"testing"
	"time"

	"github.com/tmc/pe/internal/llm"
	"github.com/tmc/pe/internal/promptfoo"
)

func TestMaterializeProvider_MergesEnvAndDelay(t *testing.T) {
	spec := promptfoo.ProviderConfig{
		ID:    "cli:test",
		Label: "cli label",
		Env: map[string]string{
			"FOO": "bar",
		},
		Prompts: []string{"keep me"},
		Delay:   "25ms",
		Config: map[string]interface{}{
			"command": `sh -c 'printf "%s %s" "$FOO" "$BAR"'`,
			"env": map[string]interface{}{
				"BAR": "baz",
			},
		},
	}

	provider, err := MaterializeProvider(spec)
	if err != nil {
		t.Fatalf("MaterializeProvider() failed: %v", err)
	}
	if provider.DisplayName() != "cli label" {
		t.Fatalf("DisplayName() = %q, want cli label", provider.DisplayName())
	}
	if !provider.AppliesToPrompt("keep me") || provider.AppliesToPrompt("skip me") {
		t.Fatalf("AppliesToPrompt() mismatch for %#v", provider.Spec.Prompts)
	}
	if provider.Delay != 25*time.Millisecond {
		t.Fatalf("Delay = %v, want 25ms", provider.Delay)
	}

	resp, err := provider.Executor.Generate(context.Background(), "ignored", llm.GenerateOptions{})
	if err != nil {
		t.Fatalf("Generate() failed: %v", err)
	}
	if resp.Text != "bar baz" {
		t.Fatalf("resp.Text = %q, want merged env output", resp.Text)
	}
}
