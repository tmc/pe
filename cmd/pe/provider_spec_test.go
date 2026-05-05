package main

import "testing"

func TestCommandProviderSpec(t *testing.T) {
	tests := []struct {
		name     string
		provider string
		model    string
		want     string
	}{
		{name: "provider only", provider: "mock", want: "mock"},
		{name: "provider and model", provider: "openai", model: "gpt-4o-mini", want: "openai:gpt-4o-mini"},
		{name: "provider spec", provider: "openai:gpt-4", model: "ignored", want: "openai:gpt-4"},
		{name: "empty provider", model: "gpt-4", want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := commandProviderSpec(tt.provider, tt.model); got != tt.want {
				t.Fatalf("commandProviderSpec(%q, %q) = %q, want %q", tt.provider, tt.model, got, tt.want)
			}
		})
	}
}
