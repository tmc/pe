// Package llm provides a unified interface for interacting with different AI providers.
package llm

import (
	"context"
	"io"
)

// Provider represents an AI service provider (e.g., OpenAI, Anthropic, Google).
type Provider interface {
	// Name returns the name of the provider.
	Name() string

	// Models returns the available models for this provider.
	Models() []Model

	// GetModel returns a specific model by name or an error if not found.
	GetModel(name string) (Model, error)
}

// Model represents a specific AI model within a provider.
type Model interface {
	// Name returns the name of the model.
	Name() string

	// Provider returns the provider this model belongs to.
	Provider() Provider

	// Complete processes a prompt and returns a completion.
	Complete(ctx context.Context, request CompletionRequest) (CompletionResponse, error)

	// CompleteStream processes a prompt and streams the completion.
	CompleteStream(ctx context.Context, request CompletionRequest) (CompletionStream, error)
}

// CompletionRequest encapsulates the parameters for a completion request.
type CompletionRequest struct {
	// Prompt is the input text to process.
	Prompt string

	// MaxTokens is the maximum number of tokens to generate.
	MaxTokens int

	// Temperature controls randomness in the output (0.0-1.0).
	Temperature float64

	// StopSequences are custom stop sequences to end generation.
	StopSequences []string

	// SystemPrompt is context or instructions for the model (if supported).
	SystemPrompt string

	// Additional provider-specific parameters.
	Parameters map[string]interface{}
}

// CompletionResponse represents the output from a completion request.
type CompletionResponse struct {
	// Text is the generated completion.
	Text string

	// FinishReason indicates why the model stopped (length, stop sequence, etc.).
	FinishReason string

	// UsageStats contains token usage information.
	UsageStats UsageStats

	// ModelInfo contains information about the model used for generation.
	ModelInfo ModelInfo
}

// UsageStats represents token usage information.
type UsageStats struct {
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
}

// ModelInfo contains information about the model used for generation.
type ModelInfo struct {
	Name      string
	Provider  string
	ModelType string
}

// CompletionStream is an interface for streaming completions.
type CompletionStream interface {
	// Next returns the next chunk of the completion.
	// Returns io.EOF when the stream is complete.
	Next(ctx context.Context) (CompletionChunk, error)

	// Close closes the stream and frees any resources.
	Close() error
}

// CompletionChunk represents a piece of text in a streaming completion.
type CompletionChunk struct {
	// Text is the text chunk.
	Text string

	// IsFinal indicates whether this is the final chunk.
	IsFinal bool

	// FinishReason is set if this is the final chunk and has a finish reason.
	FinishReason string
}

// Error types
var (
	ErrModelNotFound    = Error("model not found")
	ErrProviderNotFound = Error("provider not found")
	ErrRequestFailed    = Error("request failed")
	ErrInvalidRequest   = Error("invalid request")
)

// Error represents an error in the LLM package.
type Error string

func (e Error) Error() string {
	return string(e)
}