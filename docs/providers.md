# LLM Provider Integration Guide

This document describes how to integrate new LLM providers into the Prompt Engineering toolkit. 

## Architecture Overview

The toolkit uses a provider abstraction layer that makes it easy to add new LLM providers. The core components involved are:

1. **Provider Interface**: Defined in `internal/llm/llm.go`
2. **Provider Registry**: Implemented in `internal/llm/providers/providers.go`
3. **Provider Implementations**: Located in `internal/llm/providers/` directory

## Provider Interface

All providers must implement the `Provider` interface:

```go
type Provider interface {
    // Name returns the provider's name
    Name() string
    
    // Models returns the list of available models
    Models() []Model
    
    // GetModel returns a specific model by name or an error if not found
    GetModel(name string) (Model, error)
    
    // Complete sends a completion request to the provider
    Complete(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error)
    
    // CompleteStream sends a streaming completion request to the provider
    CompleteStream(ctx context.Context, req *CompletionRequest) (CompletionStream, error)
}
```

## Adding a New Provider

To add a new LLM provider to the system:

1. Create a new file in `internal/llm/providers/` named after your provider (e.g., `newprovider.go`)

2. Implement the `Provider` interface. Here's a template:

```go
package providers

import (
    "context"
    
    "github.com/tmc/pe/internal/llm"
)

type NewProvider struct {
    // Provider-specific fields (API keys, etc.)
}

func NewNewProvider(apiKey string) *NewProvider {
    return &NewProvider{
        // Initialize provider-specific fields
    }
}

func (p *NewProvider) Name() string {
    return "newprovider"
}

func (p *NewProvider) Models() []llm.Model {
    return []llm.Model{
        // List of available models
    }
}

func (p *NewProvider) GetModel(name string) (llm.Model, error) {
    // Find the model by name
}

func (p *NewProvider) Complete(ctx context.Context, req *llm.CompletionRequest) (*llm.CompletionResponse, error) {
    // Implementation for completing a prompt
}

func (p *NewProvider) CompleteStream(ctx context.Context, req *llm.CompletionRequest) (llm.CompletionStream, error) {
    // Implementation for streaming completions
}
```

3. Register your provider in `internal/llm/providers/providers.go`:

```go
func init() {
    // Add your provider to the registry
    RegisterProvider("newprovider", func(apiKey string) (llm.Provider, error) {
        return NewNewProvider(apiKey), nil
    })
}
```

4. Update documentation and examples to include your new provider

## Testing

When adding a new provider, you should add tests to ensure it works correctly:

1. Create unit tests for your provider implementation:

```go
func TestNewProvider(t *testing.T) {
    provider := NewNewProvider("test-api-key")
    
    // Test provider functionality
    t.Run("Models", func(t *testing.T) {
        models := provider.Models()
        if len(models) == 0 {
            t.Errorf("expected at least one model")
        }
    })
    
    // Additional test cases...
}
```

2. Add an integration test if possible (these can be skipped if API keys aren't available)

3. Update the example configurations to include your provider

## Handling API Keys

Providers typically require API keys for authentication. The toolkit handles this in a few ways:

1. Environment variables (recommended):
   - Follow naming conventions: `NEWPROVIDER_API_KEY`
   
2. Configuration files:
   - Store API keys in configuration files (with appropriate permissions)
   - Never commit API keys to version control

## Example Usage

After integrating a new provider, users should be able to use it like this:

```yaml
# config.yaml
providers:
  - newprovider:model-name
```

And then evaluate prompts with:

```bash
NEWPROVIDER_API_KEY=your-api-key pe eval config.yaml
```