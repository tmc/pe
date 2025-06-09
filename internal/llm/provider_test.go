package llm

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmc/pe/internal/promptfoo"
)

// TestProvider is a test provider that can be configured with a name
type TestProvider struct {
	name  string
	model string
}

func (p *TestProvider) Name() string  { return p.name }
func (p *TestProvider) Model() string { return p.model }
func (p *TestProvider) Generate(ctx context.Context, prompt string, options GenerateOptions) (*GenerateResponse, error) {
	return &GenerateResponse{
		Text:        "Test response",
		TotalTokens: 10,
	}, nil
}
func (p *TestProvider) SupportsStreaming() bool { return false }
func (p *TestProvider) SupportsBatch() bool     { return false }
func (p *TestProvider) EvaluatePrompt(ctx context.Context, prompt string, vars map[string]interface{}) (*promptfoo.ProviderResponse, error) {
	return &promptfoo.ProviderResponse{
		Output:     "Test response",
		TokenUsage: &promptfoo.TokenUsage{Total: 10},
	}, nil
}

func TestGetProvider(t *testing.T) {
	// Register test providers
	RegisterProviderFactory("openai", func(providerSpec string, options map[string]interface{}) (Provider, error) {
		return &TestProvider{name: "openai", model: "gpt-4"}, nil
	})
	RegisterProviderFactory("anthropic", func(providerSpec string, options map[string]interface{}) (Provider, error) {
		return &TestProvider{name: "anthropic", model: "claude-3"}, nil
	})
	RegisterProviderFactory("mock", func(providerSpec string, options map[string]interface{}) (Provider, error) {
		return &MockProviderProxy{}, nil
	})
	// Save and restore environment
	origTestMode := os.Getenv("PE_TEST_MODE")
	origMockProvider := os.Getenv("PE_MOCK_PROVIDER")
	origOpenAIKey := os.Getenv("OPENAI_API_KEY")
	origAnthropicKey := os.Getenv("ANTHROPIC_API_KEY")
	origNativeProviders := os.Getenv("PE_USE_NATIVE_PROVIDERS")
	defer func() {
		os.Setenv("PE_TEST_MODE", origTestMode)
		os.Setenv("PE_MOCK_PROVIDER", origMockProvider)
		os.Setenv("OPENAI_API_KEY", origOpenAIKey)
		os.Setenv("ANTHROPIC_API_KEY", origAnthropicKey)
		os.Setenv("PE_USE_NATIVE_PROVIDERS", origNativeProviders)
	}()

	tests := []struct {
		name     string
		backend  string
		setup    func()
		wantErr  bool
		wantType string
	}{
		{
			name:    "mock provider in test mode",
			backend: "mock",
			setup: func() {
				os.Setenv("PE_TEST_MODE", "true")
			},
			wantErr:  false,
			wantType: "mock",
		},
		{
			name:    "mock provider with PE_MOCK_PROVIDER",
			backend: "mock",
			setup: func() {
				os.Setenv("PE_MOCK_PROVIDER", "true")
			},
			wantErr:  false,
			wantType: "mock",
		},
		{
			name:    "openai provider",
			backend: "openai",
			setup: func() {
				os.Setenv("OPENAI_API_KEY", "test-key")
			},
			wantErr:  false,
			wantType: "openai",
		},
		{
			name:    "anthropic provider",
			backend: "anthropic",
			setup: func() {
				os.Setenv("ANTHROPIC_API_KEY", "test-key")
			},
			wantErr:  false,
			wantType: "anthropic",
		},
		{
			name:    "cgpt provider fallback",
			backend: "cgpt",
			setup: func() {
				os.Unsetenv("PE_USE_NATIVE_PROVIDERS")
			},
			wantErr:  false,
			wantType: "cgpt",
		},
		{
			name:    "cgpt with native providers enabled and OpenAI key",
			backend: "cgpt",
			setup: func() {
				os.Setenv("PE_USE_NATIVE_PROVIDERS", "true")
				os.Setenv("OPENAI_API_KEY", "test-key")
			},
			wantErr:  false,
			wantType: "openai",
		},
		{
			name:    "cgpt with native providers enabled and Anthropic key",
			backend: "cgpt",
			setup: func() {
				os.Setenv("PE_USE_NATIVE_PROVIDERS", "true")
				os.Setenv("ANTHROPIC_API_KEY", "test-key")
				os.Unsetenv("OPENAI_API_KEY")
			},
			wantErr:  false,
			wantType: "anthropic",
		},
		{
			name:    "auto-detect with OpenAI key",
			backend: "",
			setup: func() {
				os.Setenv("OPENAI_API_KEY", "test-key")
				os.Unsetenv("ANTHROPIC_API_KEY")
			},
			wantErr:  false,
			wantType: "openai",
		},
		{
			name:    "auto-detect with Anthropic key",
			backend: "",
			setup: func() {
				os.Setenv("ANTHROPIC_API_KEY", "test-key")
				os.Unsetenv("OPENAI_API_KEY")
			},
			wantErr:  false,
			wantType: "anthropic",
		},
		{
			name:    "auto-detect with no keys",
			backend: "",
			setup: func() {
				os.Unsetenv("OPENAI_API_KEY")
				os.Unsetenv("ANTHROPIC_API_KEY")
			},
			wantErr: true,
		},
		{
			name:    "unsupported backend",
			backend: "unsupported",
			setup:   func() {},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment
			os.Unsetenv("PE_TEST_MODE")
			os.Unsetenv("PE_MOCK_PROVIDER")
			os.Unsetenv("OPENAI_API_KEY")
			os.Unsetenv("ANTHROPIC_API_KEY")
			os.Unsetenv("PE_USE_NATIVE_PROVIDERS")

			// Setup test environment
			tt.setup()

			// Get provider
			provider, err := GetProvider(tt.backend)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, provider)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, provider)
				assert.Equal(t, tt.wantType, provider.Name())
			}
		})
	}
}

func TestProviderIntegration(t *testing.T) {
	// Set up test mode
	os.Setenv("PE_TEST_MODE", "true")
	defer os.Unsetenv("PE_TEST_MODE")

	// Get mock provider
	provider, err := GetProvider("mock")
	require.NoError(t, err)
	require.NotNil(t, provider)

	ctx := context.Background()

	t.Run("Generate", func(t *testing.T) {
		temp := 0.7
		maxTokens := 100
		resp, err := provider.Generate(ctx, "Test prompt", GenerateOptions{
			Temperature: &temp,
			MaxTokens:   &maxTokens,
		})

		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.NotEmpty(t, resp.Text)
		assert.Greater(t, resp.TotalTokens, 0)
	})

	t.Run("EvaluatePrompt", func(t *testing.T) {
		vars := map[string]interface{}{
			"temperature": 0.5,
			"max_tokens":  50,
		}

		resp, err := provider.EvaluatePrompt(ctx, "Test prompt", vars)

		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.NotEmpty(t, resp.Output)
		assert.NotNil(t, resp.TokenUsage)
	})

	t.Run("Features", func(t *testing.T) {
		assert.Equal(t, "mock", provider.Name())
		assert.Equal(t, "mock-model", provider.Model())
		assert.False(t, provider.SupportsStreaming())
		assert.False(t, provider.SupportsBatch())
	})
}

func TestDetectDefaultProvider(t *testing.T) {
	// Save and restore environment
	origOpenAI := os.Getenv("OPENAI_API_KEY")
	origAnthropic := os.Getenv("ANTHROPIC_API_KEY")
	defer func() {
		os.Setenv("OPENAI_API_KEY", origOpenAI)
		os.Setenv("ANTHROPIC_API_KEY", origAnthropic)
	}()

	tests := []struct {
		name     string
		setup    func()
		expected string
	}{
		{
			name: "detect OpenAI",
			setup: func() {
				os.Setenv("OPENAI_API_KEY", "test-key")
				os.Unsetenv("ANTHROPIC_API_KEY")
			},
			expected: "openai:gpt-4",
		},
		{
			name: "detect Anthropic",
			setup: func() {
				os.Unsetenv("OPENAI_API_KEY")
				os.Setenv("ANTHROPIC_API_KEY", "test-key")
			},
			expected: "anthropic:claude-3-sonnet-20240229",
		},
		{
			name: "no provider detected",
			setup: func() {
				os.Unsetenv("OPENAI_API_KEY")
				os.Unsetenv("ANTHROPIC_API_KEY")
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			result := detectDefaultProvider()
			assert.Equal(t, tt.expected, result)
		})
	}
}
