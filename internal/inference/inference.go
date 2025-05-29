// Package inference provides a generic API for calling LLM inference tools.
package inference

import (
	"context"
	"fmt"
	"io"
)

// Request represents a generic inference request.
type Request struct {
	// Prompt is the input text to send to the model
	Prompt string

	// Model specifies which model to use (e.g., "gpt-4", "claude-3")
	Model string

	// Temperature controls randomness (0.0 to 1.0)
	Temperature float32

	// MaxTokens limits the response length
	MaxTokens int

	// Stream indicates whether to stream the response
	Stream bool

	// SystemPrompt sets the system message (if supported)
	SystemPrompt string

	// Options for provider-specific settings
	Options map[string]interface{}
}

// Response represents the inference result.
type Response struct {
	// Content is the generated text
	Content string

	// Model that was actually used
	Model string

	// TokensUsed tracks token consumption
	TokensUsed TokenUsage

	// Metadata includes provider-specific information
	Metadata map[string]interface{}
}

// TokenUsage tracks token consumption.
type TokenUsage struct {
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
}

// StreamChunk represents a piece of a streaming response.
type StreamChunk struct {
	// Delta is the incremental text
	Delta string

	// Done indicates if this is the final chunk
	Done bool

	// Error if any occurred during streaming
	Error error
}

// Provider is the interface that inference providers must implement.
type Provider interface {
	// Name returns the provider identifier
	Name() string

	// Complete performs a non-streaming inference
	Complete(ctx context.Context, req Request) (*Response, error)

	// Stream performs a streaming inference
	Stream(ctx context.Context, req Request) (<-chan StreamChunk, error)

	// Models returns available models for this provider
	Models(ctx context.Context) ([]string, error)

	// Close cleans up any resources
	Close() error
}

// Client manages inference providers.
type Client struct {
	providers       map[string]Provider
	defaultProvider string
}

// NewClient creates a new inference client.
func NewClient() *Client {
	return &Client{
		providers: make(map[string]Provider),
	}
}

// Register adds a provider to the client.
func (c *Client) Register(name string, provider Provider) {
	c.providers[name] = provider
	if c.defaultProvider == "" {
		c.defaultProvider = name
	}
}

// SetDefault sets the default provider.
func (c *Client) SetDefault(name string) error {
	if _, ok := c.providers[name]; !ok {
		return fmt.Errorf("provider %q not found", name)
	}
	c.defaultProvider = name
	return nil
}

// Complete performs inference using the default provider.
func (c *Client) Complete(ctx context.Context, req Request) (*Response, error) {
	return c.CompleteWith(ctx, c.defaultProvider, req)
}

// CompleteWith performs inference using a specific provider.
func (c *Client) CompleteWith(ctx context.Context, provider string, req Request) (*Response, error) {
	p, ok := c.providers[provider]
	if !ok {
		return nil, fmt.Errorf("provider %q not found", provider)
	}
	return p.Complete(ctx, req)
}

// Stream performs streaming inference using the default provider.
func (c *Client) Stream(ctx context.Context, req Request) (<-chan StreamChunk, error) {
	return c.StreamWith(ctx, c.defaultProvider, req)
}

// StreamWith performs streaming inference using a specific provider.
func (c *Client) StreamWith(ctx context.Context, provider string, req Request) (<-chan StreamChunk, error) {
	p, ok := c.providers[provider]
	if !ok {
		return nil, fmt.Errorf("provider %q not found", provider)
	}
	return p.Stream(ctx, req)
}

// Models returns available models for a provider.
func (c *Client) Models(ctx context.Context, provider string) ([]string, error) {
	p, ok := c.providers[provider]
	if !ok {
		return nil, fmt.Errorf("provider %q not found", provider)
	}
	return p.Models(ctx)
}

// Providers returns a list of registered providers.
func (c *Client) Providers() []string {
	var names []string
	for name := range c.providers {
		names = append(names, name)
	}
	return names
}

// Close closes all providers.
func (c *Client) Close() error {
	var errs []error
	for _, p := range c.providers {
		if err := p.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("errors closing providers: %v", errs)
	}
	return nil
}

// StreamToWriter is a helper that writes streaming chunks to an io.Writer.
func StreamToWriter(ctx context.Context, chunks <-chan StreamChunk, w io.Writer) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case chunk, ok := <-chunks:
			if !ok {
				return nil
			}
			if chunk.Error != nil {
				return chunk.Error
			}
			if _, err := w.Write([]byte(chunk.Delta)); err != nil {
				return err
			}
			if chunk.Done {
				return nil
			}
		}
	}
}

// CollectStream collects all chunks into a single response.
func CollectStream(ctx context.Context, chunks <-chan StreamChunk) (string, error) {
	var content string
	for {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case chunk, ok := <-chunks:
			if !ok {
				return content, nil
			}
			if chunk.Error != nil {
				return "", chunk.Error
			}
			content += chunk.Delta
			if chunk.Done {
				return content, nil
			}
		}
	}
}