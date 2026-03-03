package providers

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmc/pe/internal/llm"
)

func TestNewRegistry(t *testing.T) {
	r := NewRegistry()
	assert.NotNil(t, r)
	assert.NotNil(t, r.factories)

	// Should have built-in providers registered
	assert.True(t, r.Exists("openai"))
	assert.True(t, r.Exists("anthropic"))
	assert.True(t, r.Exists("mock"))
}

func TestRegistry_Register(t *testing.T) {
	r := NewRegistry()

	// Register a custom provider
	r.Register("custom", func(model string, options map[string]interface{}) (llm.Provider, error) {
		return NewMockProvider(model, options)
	})

	assert.True(t, r.Exists("custom"))
}

func TestRegistry_Create(t *testing.T) {
	tests := []struct {
		name         string
		providerSpec string
		options      map[string]interface{}
		wantErr      bool
		errContains  string
	}{
		{
			name:         "valid openai provider",
			providerSpec: "openai:gpt-4",
			options:      map[string]interface{}{"apiKey": "test-key"},
			wantErr:      false,
		},
		{
			name:         "valid anthropic provider",
			providerSpec: "anthropic:claude-3-haiku",
			options:      map[string]interface{}{"apiKey": "test-key"},
			wantErr:      false,
		},
		{
			name:         "valid mock provider",
			providerSpec: "mock:test-model",
			options:      map[string]interface{}{},
			wantErr:      true, // Mock provider requires PE_TEST_MODE
			errContains:  "mock provider only available in test mode",
		},
		{
			name:         "invalid provider format",
			providerSpec: "invalid-format",
			options:      map[string]interface{}{},
			wantErr:      true,
			errContains:  "invalid provider format",
		},
		{
			name:         "unknown provider",
			providerSpec: "unknown:model",
			options:      map[string]interface{}{},
			wantErr:      true,
			errContains:  "unknown provider",
		},
		{
			name:         "model normalization",
			providerSpec: "openai:gpt4",
			options:      map[string]interface{}{"apiKey": "test-key"},
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewRegistry()
			provider, err := r.Create(tt.providerSpec, tt.options)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, provider)
			}
		})
	}
}

func TestRegistry_List(t *testing.T) {
	r := NewRegistry()
	providers := r.List()

	assert.Contains(t, providers, "openai")
	assert.Contains(t, providers, "anthropic")
	assert.Contains(t, providers, "mock")
	assert.GreaterOrEqual(t, len(providers), 3)
}

func TestRegistry_GetSupportedModels(t *testing.T) {
	r := NewRegistry()
	models := r.GetSupportedModels()

	// Check OpenAI models
	openaiModels, exists := models["openai"]
	assert.True(t, exists)
	assert.Contains(t, openaiModels, "gpt-4")
	assert.Contains(t, openaiModels, "gpt-3.5-turbo")

	// Check Anthropic models
	anthropicModels, exists := models["anthropic"]
	assert.True(t, exists)
	assert.Contains(t, anthropicModels, "claude-3-opus-20240229")
	assert.Contains(t, anthropicModels, "claude-3-sonnet-20240229")
}

func TestRegistry_ValidateProvider(t *testing.T) {
	tests := []struct {
		name         string
		providerSpec string
		wantErr      bool
		errContains  string
	}{
		{
			name:         "valid openai provider",
			providerSpec: "openai:gpt-4",
			wantErr:      false,
		},
		{
			name:         "valid anthropic provider",
			providerSpec: "anthropic:claude-3-opus-20240229",
			wantErr:      false,
		},
		{
			name:         "invalid format",
			providerSpec: "invalid",
			wantErr:      true,
			errContains:  "invalid provider format",
		},
		{
			name:         "unknown provider",
			providerSpec: "unknown:model",
			wantErr:      true,
			errContains:  "unknown provider",
		},
		{
			name:         "unsupported model (should not error)",
			providerSpec: "openai:gpt-5-future",
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewRegistry()
			err := r.ValidateProvider(tt.providerSpec)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDefaultRegistry(t *testing.T) {
	// Set test mode for mock provider
	os.Setenv("PE_TEST_MODE", "true")
	defer os.Unsetenv("PE_TEST_MODE")

	// Test CreateProvider with mock
	provider, err := CreateProvider("mock:test", map[string]interface{}{})
	assert.NoError(t, err)
	assert.NotNil(t, provider)

	// Test RegisterProvider
	RegisterProvider("test-provider", func(model string, options map[string]interface{}) (llm.Provider, error) {
		return NewMockProvider(model, options)
	})

	// Test ListProviders
	providers := ListProviders()
	assert.Contains(t, providers, "test-provider")

	// Test ValidateProviderSpec
	err = ValidateProviderSpec("openai:gpt-4")
	assert.NoError(t, err)

	err = ValidateProviderSpec("invalid")
	assert.Error(t, err)

	// Test GetSupportedModelsMap
	models := GetSupportedModelsMap()
	assert.NotNil(t, models["openai"])
	assert.NotNil(t, models["anthropic"])
}

func TestRegistry_Concurrency(t *testing.T) {
	// Set test mode for mock provider
	os.Setenv("PE_TEST_MODE", "true")
	defer os.Unsetenv("PE_TEST_MODE")

	r := NewRegistry()

	// Test concurrent registration and creation
	done := make(chan bool)

	// Concurrent registrations
	go func() {
		for i := 0; i < 100; i++ {
			r.Register("provider1", func(model string, options map[string]interface{}) (llm.Provider, error) {
				return NewMockProvider(model, options)
			})
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 100; i++ {
			r.Register("provider2", func(model string, options map[string]interface{}) (llm.Provider, error) {
				return NewMockProvider(model, options)
			})
		}
		done <- true
	}()

	// Concurrent creations
	go func() {
		for i := 0; i < 100; i++ {
			_, _ = r.Create("mock:test", map[string]interface{}{})
		}
		done <- true
	}()

	// Wait for all goroutines
	for i := 0; i < 3; i++ {
		<-done
	}

	// Verify registry is still functional
	provider, err := r.Create("mock:test", map[string]interface{}{})
	require.NoError(t, err)
	require.NotNil(t, provider)
}
