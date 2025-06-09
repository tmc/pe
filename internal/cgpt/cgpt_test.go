package cgpt

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultProvider(t *testing.T) {
	provider := DefaultProvider()

	assert.NotNil(t, provider)
	assert.Equal(t, "gpt-4o", provider.Model)
	assert.Equal(t, 1024, provider.MaxTokens)
	assert.Equal(t, 0.2, provider.Temperature)
	assert.Equal(t, "openai", provider.Backend)
}

func TestModelProvider_ApplyConfigFromVars(t *testing.T) {
	tests := []struct {
		name     string
		provider *ModelProvider
		vars     map[string]interface{}
		expected *ModelProvider
	}{
		{
			name:     "apply model from vars",
			provider: DefaultProvider(),
			vars: map[string]interface{}{
				"model": "gpt-3.5-turbo",
			},
			expected: &ModelProvider{
				Model:       "gpt-3.5-turbo",
				MaxTokens:   1024,
				Temperature: 0.2,
				Backend:     "openai",
			},
		},
		{
			name:     "apply temperature from vars",
			provider: DefaultProvider(),
			vars: map[string]interface{}{
				"temperature": 0.8,
			},
			expected: &ModelProvider{
				Model:       "gpt-4o",
				MaxTokens:   1024,
				Temperature: 0.8,
				Backend:     "openai",
			},
		},
		{
			name:     "apply max_tokens from vars",
			provider: DefaultProvider(),
			vars: map[string]interface{}{
				"max_tokens": 2048,
			},
			expected: &ModelProvider{
				Model:       "gpt-4o",
				MaxTokens:   2048,
				Temperature: 0.2,
				Backend:     "openai",
			},
		},
		{
			name:     "apply backend from vars",
			provider: DefaultProvider(),
			vars: map[string]interface{}{
				"backend": "anthropic",
			},
			expected: &ModelProvider{
				Model:       "gpt-4o",
				MaxTokens:   1024,
				Temperature: 0.2,
				Backend:     "anthropic",
			},
		},
		{
			name:     "apply all configs",
			provider: DefaultProvider(),
			vars: map[string]interface{}{
				"model":       "claude-3",
				"temperature": 0.5,
				"max_tokens":  4096,
				"backend":     "anthropic",
			},
			expected: &ModelProvider{
				Model:       "claude-3",
				MaxTokens:   4096,
				Temperature: 0.5,
				Backend:     "anthropic",
			},
		},
		{
			name:     "ignore invalid types",
			provider: DefaultProvider(),
			vars: map[string]interface{}{
				"temperature": "not-a-float",
				"max_tokens":  "not-an-int",
			},
			expected: &ModelProvider{
				Model:       "gpt-4o",
				MaxTokens:   1024,
				Temperature: 0.2,
				Backend:     "openai",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.provider.ApplyConfigFromVars(tt.vars)
			assert.Equal(t, tt.expected, tt.provider)
		})
	}
}

func TestReplaceVariables(t *testing.T) {
	tests := []struct {
		name     string
		prompt   string
		vars     map[string]interface{}
		expected string
	}{
		{
			name:     "no variables",
			prompt:   "Hello world",
			vars:     map[string]interface{}{},
			expected: "Hello world",
		},
		{
			name:   "single variable",
			prompt: "Hello {{name}}",
			vars: map[string]interface{}{
				"name": "Alice",
			},
			expected: "Hello Alice",
		},
		{
			name:   "multiple variables",
			prompt: "{{greeting}} {{name}}, today is {{day}}",
			vars: map[string]interface{}{
				"greeting": "Hi",
				"name":     "Bob",
				"day":      "Monday",
			},
			expected: "Hi Bob, today is Monday",
		},
		{
			name:   "repeated variable",
			prompt: "{{name}} is {{name}}'s name",
			vars: map[string]interface{}{
				"name": "Charlie",
			},
			expected: "Charlie is Charlie's name",
		},
		{
			name:     "missing variable",
			prompt:   "Hello {{name}}",
			vars:     map[string]interface{}{},
			expected: "Hello {{name}}",
		},
		{
			name:   "numeric values",
			prompt: "Count: {{count}}, Price: {{price}}",
			vars: map[string]interface{}{
				"count": 42,
				"price": 9.99,
			},
			expected: "Count: 42, Price: 9.99",
		},
		{
			name:   "boolean values",
			prompt: "Active: {{active}}, Ready: {{ready}}",
			vars: map[string]interface{}{
				"active": true,
				"ready":  false,
			},
			expected: "Active: true, Ready: false",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := replaceVariables(tt.prompt, tt.vars)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		expected string
	}{
		{
			name:     "milliseconds",
			duration: 100 * time.Millisecond,
			expected: "100ms",
		},
		{
			name:     "seconds",
			duration: 2 * time.Second,
			expected: "2.00s",
		},
		{
			name:     "minutes",
			duration: 3 * time.Minute,
			expected: "3.00m",
		},
		{
			name:     "mixed",
			duration: 1*time.Minute + 30*time.Second + 500*time.Millisecond,
			expected: "1.51m",
		},
		{
			name:     "zero",
			duration: 0,
			expected: "0ms",
		},
		{
			name:     "microseconds",
			duration: 50 * time.Microsecond,
			expected: "0ms",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatDuration(tt.duration)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsKnownProvider(t *testing.T) {
	tests := []struct {
		name     string
		model    string
		expected bool
	}{
		{"openai model", "gpt-4", true},
		{"openai turbo", "gpt-3.5-turbo", true},
		{"anthropic claude", "claude-3", true},
		{"anthropic claude-2", "claude-2", true},
		{"google gemini", "gemini-pro", true},
		{"google bison", "text-bison", true},
		{"unknown model", "unknown-model", false},
		{"empty model", "", false},
		{"partial match", "gpt", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isKnownProvider(tt.model)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseProviderString(t *testing.T) {
	tests := []struct {
		name        string
		provider    string
		wantBackend string
		wantModel   string
		wantTemp    float64
		wantMaxTok  int
	}{
		{
			name:        "simple openai",
			provider:    "openai:gpt-4",
			wantBackend: "openai",
			wantModel:   "gpt-4",
			wantTemp:    0.0,
			wantMaxTok:  0,
		},
		{
			name:        "openai with temperature",
			provider:    "openai:gpt-3.5-turbo:temperature=0.8",
			wantBackend: "openai",
			wantModel:   "gpt-3.5-turbo",
			wantTemp:    0.8,
			wantMaxTok:  0,
		},
		{
			name:        "anthropic with max_tokens",
			provider:    "anthropic:claude-3:max_tokens=2048",
			wantBackend: "anthropic",
			wantModel:   "claude-3",
			wantTemp:    0.0,
			wantMaxTok:  2048,
		},
		{
			name:        "full config",
			provider:    "openai:gpt-4:temperature=0.5:max_tokens=1000",
			wantBackend: "openai",
			wantModel:   "gpt-4",
			wantTemp:    0.5,
			wantMaxTok:  1000,
		},
		{
			name:        "model only",
			provider:    "gpt-4",
			wantBackend: "gpt-4",
			wantModel:   "",
			wantTemp:    0.0,
			wantMaxTok:  0,
		},
		{
			name:        "invalid format",
			provider:    "invalid::format",
			wantBackend: "invalid",
			wantModel:   "",
			wantTemp:    0.0,
			wantMaxTok:  0,
		},
		{
			name:        "empty provider",
			provider:    "",
			wantBackend: "",
			wantModel:   "",
			wantTemp:    0.0,
			wantMaxTok:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			backend, model, temp, maxTok := parseProviderString(tt.provider)
			assert.Equal(t, tt.wantBackend, backend)
			assert.Equal(t, tt.wantModel, model)
			assert.Equal(t, tt.wantTemp, temp)
			assert.Equal(t, tt.wantMaxTok, maxTok)
		})
	}
}

// Test the dry run functionality
func TestEvaluatePromptDryRun(t *testing.T) {
	provider := DefaultProvider()

	prompt := "Test prompt"
	vars := map[string]interface{}{
		"name": "Test",
	}

	resp, err := provider.EvaluatePromptWithOptions(prompt, vars, true)

	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "Dry run - no actual execution", resp.Output)
	assert.NotNil(t, resp.TokenUsage)
	assert.True(t, resp.TokenUsage.Total > 0) // Just check it's positive
}

// Test error scenarios
func TestEvaluatePromptErrors(t *testing.T) {
	// This test would need mocking of exec.Command or would run in integration mode
	// For now, we'll skip the actual execution test
	t.Skip("Requires cgpt CLI or mocking")
}
