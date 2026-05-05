// Package anthropic provides a native Anthropic Claude API implementation for the inference provider interface.
package anthropic

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/tmc/pe/internal/inference"
	"github.com/tmc/pe/internal/security"
)

func init() {
	// Register Anthropic provider
	inference.MustRegister("anthropic", Factory)
}

// Factory creates a new Anthropic provider from configuration.
func Factory(config map[string]interface{}) (inference.Provider, error) {
	apiKey := ""
	if config != nil {
		if k, ok := config["api_key"].(string); ok {
			apiKey = k
		}
	}
	if apiKey == "" {
		apiKey = os.Getenv("ANTHROPIC_API_KEY")
	}
	if apiKey == "" {
		return nil, fmt.Errorf("Anthropic API key not provided")
	}

	baseURL := "https://api.anthropic.com"
	if config != nil {
		if u, ok := config["base_url"].(string); ok {
			baseURL = u
		}
	}

	return New(apiKey, baseURL), nil
}

// Provider implements the inference.Provider interface for Anthropic.
type Provider struct {
	apiKey  string
	baseURL string
	client  *http.Client
}

// New creates a new Anthropic provider.
func New(apiKey, baseURL string) *Provider {
	return &Provider{
		apiKey:  apiKey,
		baseURL: strings.TrimSuffix(baseURL, "/"),
		client: &http.Client{
			Timeout: 2 * time.Minute,
		},
	}
}

// Name returns the provider name.
func (p *Provider) Name() string {
	return "anthropic"
}

// messageRequest represents the Anthropic messages API request.
type messageRequest struct {
	Model       string    `json:"model"`
	Messages    []message `json:"messages"`
	MaxTokens   int       `json:"max_tokens"`
	Temperature float32   `json:"temperature,omitempty"`
	System      string    `json:"system,omitempty"`
	Stream      bool      `json:"stream,omitempty"`
}

// message represents a message in the conversation.
type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// messageResponse represents the Anthropic messages API response.
type messageResponse struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Role    string `json:"role"`
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Model      string `json:"model"`
	StopReason string `json:"stop_reason"`
	Usage      struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

// streamEvent represents a streaming event from Anthropic.
type streamEvent struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data,omitempty"`
}

// contentBlockDelta represents a content delta in streaming.
type contentBlockDelta struct {
	Type  string `json:"type"`
	Index int    `json:"index"`
	Delta struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"delta"`
}

// Complete performs a non-streaming inference using Anthropic.
func (p *Provider) Complete(ctx context.Context, req inference.Request) (*inference.Response, error) {
	// Build request
	msgReq := messageRequest{
		Model: req.Model,
		Messages: []message{
			{Role: "user", Content: req.Prompt},
		},
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
		System:      req.SystemPrompt,
		Stream:      false,
	}

	// Default model if not specified
	if msgReq.Model == "" {
		msgReq.Model = "claude-3-sonnet-20240229"
	}

	// Default max tokens for Anthropic
	if msgReq.MaxTokens == 0 {
		msgReq.MaxTokens = 4096
	}

	// Marshal request
	body, err := json.Marshal(msgReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/v1/messages", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", p.apiKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	// Execute request
	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Anthropic API error (status %d): %s", resp.StatusCode, security.RedactSecrets(string(respBody)))
	}

	// Parse response
	var msgResp messageResponse
	if err := json.Unmarshal(respBody, &msgResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Extract text content
	var content string
	for _, c := range msgResp.Content {
		if c.Type == "text" {
			content += c.Text
		}
	}

	return &inference.Response{
		Content: content,
		Model:   msgResp.Model,
		TokensUsed: inference.TokenUsage{
			PromptTokens:     msgResp.Usage.InputTokens,
			CompletionTokens: msgResp.Usage.OutputTokens,
			TotalTokens:      msgResp.Usage.InputTokens + msgResp.Usage.OutputTokens,
		},
		Metadata: map[string]interface{}{
			"provider":    "anthropic",
			"stop_reason": msgResp.StopReason,
		},
	}, nil
}

// Stream performs a streaming inference using Anthropic.
func (p *Provider) Stream(ctx context.Context, req inference.Request) (<-chan inference.StreamChunk, error) {
	// Build request
	msgReq := messageRequest{
		Model: req.Model,
		Messages: []message{
			{Role: "user", Content: req.Prompt},
		},
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
		System:      req.SystemPrompt,
		Stream:      true,
	}

	// Default model if not specified
	if msgReq.Model == "" {
		msgReq.Model = "claude-3-sonnet-20240229"
	}

	// Default max tokens for Anthropic
	if msgReq.MaxTokens == 0 {
		msgReq.MaxTokens = 4096
	}

	// Marshal request
	body, err := json.Marshal(msgReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/v1/messages", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", p.apiKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")
	httpReq.Header.Set("Accept", "text/event-stream")

	// Execute request
	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	// Check status
	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Anthropic API error (status %d): %s", resp.StatusCode, security.RedactSecrets(string(body)))
	}

	// Create channel for streaming
	chunks := make(chan inference.StreamChunk)

	// Start goroutine to read SSE stream
	go func() {
		defer close(chunks)
		defer resp.Body.Close()

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := scanner.Text()

			// Skip empty lines
			if line == "" {
				continue
			}

			// Parse SSE event
			if strings.HasPrefix(line, "event: ") {
				eventType := strings.TrimPrefix(line, "event: ")

				// Read data line
				if scanner.Scan() {
					dataLine := scanner.Text()
					if strings.HasPrefix(dataLine, "data: ") {
						data := strings.TrimPrefix(dataLine, "data: ")

						switch eventType {
						case "content_block_delta":
							var delta contentBlockDelta
							if err := json.Unmarshal([]byte(data), &delta); err != nil {
								chunks <- inference.StreamChunk{Error: fmt.Errorf("failed to parse delta: %w", err)}
								return
							}

							if delta.Delta.Type == "text_delta" {
								select {
								case <-ctx.Done():
									chunks <- inference.StreamChunk{Error: ctx.Err()}
									return
								case chunks <- inference.StreamChunk{Delta: delta.Delta.Text}:
								}
							}

						case "message_stop":
							chunks <- inference.StreamChunk{Done: true}
							return

						case "error":
							var errData struct {
								Error struct {
									Type    string `json:"type"`
									Message string `json:"message"`
								} `json:"error"`
							}
							if err := json.Unmarshal([]byte(data), &errData); err == nil {
								chunks <- inference.StreamChunk{Error: fmt.Errorf("Anthropic error: %s", errData.Error.Message)}
							} else {
								chunks <- inference.StreamChunk{Error: fmt.Errorf("unknown Anthropic error")}
							}
							return
						}
					}
				}
			}
		}

		if err := scanner.Err(); err != nil {
			chunks <- inference.StreamChunk{Error: fmt.Errorf("stream read error: %w", err)}
		}
	}()

	return chunks, nil
}

// Models returns available models for Anthropic.
func (p *Provider) Models(ctx context.Context) ([]string, error) {
	return []string{
		"claude-3-opus-20240229",
		"claude-3-sonnet-20240229",
		"claude-3-haiku-20240307",
		"claude-2.1",
		"claude-2.0",
		"claude-instant-1.2",
	}, nil
}

// Close cleans up any resources.
func (p *Provider) Close() error {
	// Nothing to clean up
	return nil
}
