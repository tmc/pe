package llm

import (
	"context"
	"time"

	"github.com/tmc/pe/internal/promptfoo"
)

// GenerateOptions contains options for text generation
type GenerateOptions struct {
	Temperature     *float64
	MaxTokens       *int
	TopP            *float64
	TopK            *int
	Stop            []string
	ProviderOptions map[string]interface{}
}

// GenerateResponse contains the response from a generation request
type GenerateResponse struct {
	Text             string
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
	Latency          time.Duration
	Cost             float64
	Model            string
	FinishReason     string
	Metadata         map[string]interface{}
}

// StreamResponse contains a streaming response chunk
type StreamResponse struct {
	Text    string
	Done    bool
	Latency time.Duration
	Error   error
}

// Provider is an interface for language model providers
type Provider interface {
	// Legacy method for backward compatibility
	EvaluatePrompt(ctx context.Context, prompt string, vars map[string]interface{}) (*promptfoo.ProviderResponse, error)

	// Enhanced methods for native providers
	Name() string
	Model() string
	Generate(ctx context.Context, prompt string, options GenerateOptions) (*GenerateResponse, error)
	SupportsStreaming() bool
	SupportsBatch() bool
}

// StreamingProvider extends Provider with streaming capabilities
type StreamingProvider interface {
	Provider
	GenerateStream(ctx context.Context, prompt string, options GenerateOptions) (<-chan *StreamResponse, error)
}

// BatchProvider extends Provider with batch processing capabilities
type BatchProvider interface {
	Provider
	GenerateBatch(ctx context.Context, prompts []string, options GenerateOptions) ([]*GenerateResponse, error)
}
