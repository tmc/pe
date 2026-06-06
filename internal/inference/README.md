# PE Inference API

The inference package provides a generic API for calling LLM inference tools. It supports multiple providers through a plugin-like architecture.

## Features

- **Provider Abstraction**: Generic interface for different LLM providers
- **Streaming Support**: Both blocking and streaming inference modes
- **Provider Registry**: Automatic provider registration and discovery
- **Simple API**: Clean, Go-idiomatic interface

## Usage

### Basic Example

```go
import (
    "github.com/tmc/pe/internal/inference"
    _ "github.com/tmc/pe/internal/inference/providers/cgpt" // Auto-registers
)

// Create client with all registered providers
client, err := inference.DefaultClient()

// Simple completion
resp, err := client.Complete(ctx, inference.Request{
    Prompt:      "What is 2+2?",
    Model:       "gpt-3.5-turbo",
    Temperature: 0,
})
fmt.Println(resp.Content)

// Streaming
chunks, err := client.Stream(ctx, inference.Request{
    Prompt: "Write a story...",
    Stream: true,
})
for chunk := range chunks {
    fmt.Print(chunk.Delta)
}
```

### Creating a Provider

Implement the `Provider` interface:

```go
type Provider interface {
    Name() string
    Complete(ctx context.Context, req Request) (*Response, error)
    Stream(ctx context.Context, req Request) (<-chan StreamChunk, error)
    Models(ctx context.Context) ([]string, error)
    Close() error
}
```

Register it in an init function:

```go
func init() {
    inference.Register("myprovider", func(config map[string]interface{}) (inference.Provider, error) {
        return NewMyProvider(config), nil
    })
}
```

## Supported Providers

### cgpt

Uses the `cgpt` CLI tool (`github.com/tmc/cgpt`).

```go
// Automatic registration
import _ "github.com/tmc/pe/internal/inference/providers/cgpt"

// Or manual with custom binary
provider := cgpt.NewWithBinary("/path/to/cgpt")
```

Configuration:
- `binary`: Path to cgpt binary (optional, defaults to `go run`)

### Adding More Providers

To add a new provider:

1. Create a new package in `providers/`
2. Implement the `Provider` interface
3. Register in an `init()` function
4. Import the package to auto-register

Example providers to add:
- `ollama`: Local Ollama models
- `openai`: Direct OpenAI API
- `anthropic`: Direct Anthropic API
- `replicate`: Replicate.com models
- `huggingface`: HuggingFace inference

## CLI Integration

The inference API is used by PE's `run` command:

```bash
# Use default provider (cgpt)
pe run "What is 2+2?"

# Use specific provider
pe run "Explain AI" --provider ollama --model llama2

# With options
pe run prompt.txt --temperature 0.8 --max-tokens 100
```

## Design Decisions

1. **Provider Registry**: Allows providers to self-register, making it easy to add new ones
2. **Streaming First**: Designed with streaming as a primary use case
3. **Simple Request/Response**: Common structure works across providers
4. **Tool-Based Providers**: Many providers wrap CLI tools for simplicity
5. **Pluggable**: Easy to add new providers without changing core code

## Possible Enhancements

These are design ideas, not current API guarantees:

- Caching layer for identical requests
- Retry logic with backoff
- Request/response interceptors
- Metrics and observability hooks
- Load balancing across providers
- Cost tracking per provider
- Prompt template support
- Response validation
