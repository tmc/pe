// Package openai provides a native OpenAI API implementation for the inference provider interface.
package openai

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
)

func init() {
	// Register OpenAI provider
	inference.MustRegister("openai", Factory)
}

// Factory creates a new OpenAI provider from configuration.
func Factory(config map[string]interface{}) (inference.Provider, error) {
	apiKey := ""
	if config != nil {
		if k, ok := config["api_key"].(string); ok {
			apiKey = k
		}
	}
	if apiKey == "" {
		apiKey = os.Getenv("OPENAI_API_KEY")
	}
	if apiKey == "" {
		return nil, fmt.Errorf("OpenAI API key not provided")
	}

	baseURL := "https://api.openai.com/v1"
	if config != nil {
		if u, ok := config["base_url"].(string); ok {
			baseURL = u
		}
	}

	return New(apiKey, baseURL), nil
}

// Provider implements the inference.Provider interface for OpenAI.
type Provider struct {
	apiKey  string
	baseURL string
	client  *http.Client
}

// New creates a new OpenAI provider.
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
	return "openai"
}

// chatRequest represents the OpenAI chat completion request.
type chatRequest struct {
	Model       string        `json:"model"`
	Messages    []chatMessage `json:"messages"`
	Temperature float32       `json:"temperature,omitempty"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
	Stream      bool          `json:"stream,omitempty"`
}

// chatMessage represents a message in the chat.
type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// chatResponse represents the OpenAI chat completion response.
type chatResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index        int         `json:"index"`
		Message      chatMessage `json:"message"`
		FinishReason string      `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// streamChunk represents a streaming response chunk.
type streamChunk struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index int `json:"index"`
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
		FinishReason *string `json:"finish_reason"`
	} `json:"choices"`
}

// Complete performs a non-streaming inference using OpenAI.
func (p *Provider) Complete(ctx context.Context, req inference.Request) (*inference.Response, error) {
	// Build messages
	messages := []chatMessage{}
	if req.SystemPrompt != "" {
		messages = append(messages, chatMessage{
			Role:    "system",
			Content: req.SystemPrompt,
		})
	}
	messages = append(messages, chatMessage{
		Role:    "user",
		Content: req.Prompt,
	})

	// Build request
	chatReq := chatRequest{
		Model:       req.Model,
		Messages:    messages,
		Temperature: req.Temperature,
		MaxTokens:   req.MaxTokens,
		Stream:      false,
	}

	// Default model if not specified
	if chatReq.Model == "" {
		chatReq.Model = "gpt-3.5-turbo"
	}

	// Marshal request
	body, err := json.Marshal(chatReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)

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
		return nil, fmt.Errorf("OpenAI API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	// Parse response
	var chatResp chatResponse
	if err := json.Unmarshal(respBody, &chatResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if len(chatResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	return &inference.Response{
		Content: chatResp.Choices[0].Message.Content,
		Model:   chatResp.Model,
		TokensUsed: inference.TokenUsage{
			PromptTokens:     chatResp.Usage.PromptTokens,
			CompletionTokens: chatResp.Usage.CompletionTokens,
			TotalTokens:      chatResp.Usage.TotalTokens,
		},
		Metadata: map[string]interface{}{
			"provider":      "openai",
			"finish_reason": chatResp.Choices[0].FinishReason,
		},
	}, nil
}

// Stream performs a streaming inference using OpenAI.
func (p *Provider) Stream(ctx context.Context, req inference.Request) (<-chan inference.StreamChunk, error) {
	// Build messages
	messages := []chatMessage{}
	if req.SystemPrompt != "" {
		messages = append(messages, chatMessage{
			Role:    "system",
			Content: req.SystemPrompt,
		})
	}
	messages = append(messages, chatMessage{
		Role:    "user",
		Content: req.Prompt,
	})

	// Build request
	chatReq := chatRequest{
		Model:       req.Model,
		Messages:    messages,
		Temperature: req.Temperature,
		MaxTokens:   req.MaxTokens,
		Stream:      true,
	}

	// Default model if not specified
	if chatReq.Model == "" {
		chatReq.Model = "gpt-3.5-turbo"
	}

	// Marshal request
	body, err := json.Marshal(chatReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)
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
		return nil, fmt.Errorf("OpenAI API error (status %d): %s", resp.StatusCode, string(body))
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

			// Parse SSE data
			if strings.HasPrefix(line, "data: ") {
				data := strings.TrimPrefix(line, "data: ")

				// Check for end of stream
				if data == "[DONE]" {
					chunks <- inference.StreamChunk{Done: true}
					return
				}

				// Parse JSON chunk
				var chunk streamChunk
				if err := json.Unmarshal([]byte(data), &chunk); err != nil {
					chunks <- inference.StreamChunk{Error: fmt.Errorf("failed to parse chunk: %w", err)}
					return
				}

				// Extract content
				if len(chunk.Choices) > 0 && chunk.Choices[0].Delta.Content != "" {
					select {
					case <-ctx.Done():
						chunks <- inference.StreamChunk{Error: ctx.Err()}
						return
					case chunks <- inference.StreamChunk{Delta: chunk.Choices[0].Delta.Content}:
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

// Models returns available models for OpenAI.
func (p *Provider) Models(ctx context.Context) ([]string, error) {
	return []string{
		"gpt-4-turbo-preview",
		"gpt-4",
		"gpt-3.5-turbo",
		"gpt-3.5-turbo-16k",
	}, nil
}

// Close cleans up any resources.
func (p *Provider) Close() error {
	// Nothing to clean up
	return nil
}
