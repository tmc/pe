package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/tmc/pe/internal/llm"
	"github.com/tmc/pe/internal/promptfoo"
)

// AnthropicProvider implements the LLM provider interface for Anthropic Claude API
type AnthropicProvider struct {
	client     *http.Client
	apiKey     string
	baseURL    string
	model      string
	maxRetries int
}

// AnthropicRequest represents the request structure for Anthropic API
type AnthropicRequest struct {
	Model       string          `json:"model"`
	MaxTokens   int             `json:"max_tokens"`
	Messages    []ClaudeMessage `json:"messages"`
	Temperature *float64        `json:"temperature,omitempty"`
	Stream      bool            `json:"stream,omitempty"`
}

// ClaudeMessage represents a Claude message
type ClaudeMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// AnthropicResponse represents the response structure from Anthropic API
type AnthropicResponse struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Role    string `json:"role"`
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Model        string `json:"model"`
	StopReason   string `json:"stop_reason"`
	StopSequence string `json:"stop_sequence"`
	Usage        struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

// AnthropicErrorResponse represents an error response from Anthropic API
type AnthropicErrorResponse struct {
	Type  string `json:"type"`
	Error struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error"`
}

// NewAnthropicProvider creates a new Anthropic provider instance
func NewAnthropicProvider(model string, options map[string]interface{}) (*AnthropicProvider, error) {
	apiKey := getStringOption(options, "apiKey", "")
	if apiKey == "" {
		// Try environment variable
		apiKey = getEnvVar("ANTHROPIC_API_KEY")
		if apiKey == "" {
			return nil, fmt.Errorf("Anthropic API key not provided")
		}
	}

	baseURL := getStringOption(options, "baseURL", "https://api.anthropic.com")
	maxRetries := getIntOption(options, "maxRetries", 3)

	client := &http.Client{
		Timeout: 2 * time.Minute, // Claude can be slower
	}

	return &AnthropicProvider{
		client:     client,
		apiKey:     apiKey,
		baseURL:    strings.TrimSuffix(baseURL, "/"),
		model:      model,
		maxRetries: maxRetries,
	}, nil
}

// Name returns the provider name
func (p *AnthropicProvider) Name() string {
	return "anthropic"
}

// Model returns the model being used
func (p *AnthropicProvider) Model() string {
	return p.model
}

// Generate sends a prompt to Anthropic and returns the response
func (p *AnthropicProvider) Generate(ctx context.Context, prompt string, options llm.GenerateOptions) (*llm.GenerateResponse, error) {
	startTime := time.Now()

	// Build request
	maxTokens := 4096 // Default max tokens
	if options.MaxTokens != nil {
		maxTokens = *options.MaxTokens
	}

	req := AnthropicRequest{
		Model:     p.model,
		MaxTokens: maxTokens,
		Messages: []ClaudeMessage{
			{Role: "user", Content: prompt},
		},
	}

	// Apply options
	if options.Temperature != nil {
		req.Temperature = options.Temperature
	}

	// Marshal request
	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %v", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/v1/messages", bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", p.apiKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")
	httpReq.Header.Set("User-Agent", "pe-toolkit/1.0")

	// Execute request with retries
	var resp *http.Response
	var lastErr error

	for attempt := 0; attempt < p.maxRetries; attempt++ {
		resp, lastErr = p.client.Do(httpReq)
		if lastErr == nil && resp.StatusCode < 500 {
			break
		}

		if resp != nil {
			resp.Body.Close()
		}

		// Wait before retry (exponential backoff)
		if attempt < p.maxRetries-1 {
			waitTime := time.Duration(1<<uint(attempt)) * time.Second
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(waitTime):
			}
		}
	}

	if lastErr != nil {
		return nil, fmt.Errorf("error executing request: %v", lastErr)
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %v", err)
	}

	// Handle error responses
	if resp.StatusCode != http.StatusOK {
		var errResp AnthropicErrorResponse
		if err := json.Unmarshal(respBody, &errResp); err == nil {
			return nil, fmt.Errorf("Anthropic API error: %s", errResp.Error.Message)
		}
		return nil, fmt.Errorf("Anthropic API error: status %d, body: %s", resp.StatusCode, string(respBody))
	}

	// Parse successful response
	var anthropicResp AnthropicResponse
	if err := json.Unmarshal(respBody, &anthropicResp); err != nil {
		return nil, fmt.Errorf("error parsing response: %v", err)
	}

	if len(anthropicResp.Content) == 0 {
		return nil, fmt.Errorf("no content in response")
	}

	latency := time.Since(startTime)

	// Calculate cost (approximate)
	cost := calculateAnthropicCost(p.model, anthropicResp.Usage.InputTokens, anthropicResp.Usage.OutputTokens)

	return &llm.GenerateResponse{
		Text:             anthropicResp.Content[0].Text,
		PromptTokens:     anthropicResp.Usage.InputTokens,
		CompletionTokens: anthropicResp.Usage.OutputTokens,
		TotalTokens:      anthropicResp.Usage.InputTokens + anthropicResp.Usage.OutputTokens,
		Latency:          latency,
		Cost:             cost,
		Model:            anthropicResp.Model,
		FinishReason:     anthropicResp.StopReason,
	}, nil
}

// GenerateStream sends a prompt to Anthropic and returns a streaming response
func (p *AnthropicProvider) GenerateStream(ctx context.Context, prompt string, options llm.GenerateOptions) (<-chan *llm.StreamResponse, error) {
	// Build request with streaming enabled
	maxTokens := 4096 // Default max tokens
	if options.MaxTokens != nil {
		maxTokens = *options.MaxTokens
	}

	req := AnthropicRequest{
		Model:     p.model,
		MaxTokens: maxTokens,
		Stream:    true,
		Messages: []ClaudeMessage{
			{Role: "user", Content: prompt},
		},
	}

	// Apply options
	if options.Temperature != nil {
		req.Temperature = options.Temperature
	}

	// Marshal request
	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %v", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/v1/messages", bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", p.apiKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")
	httpReq.Header.Set("Accept", "text/event-stream")
	httpReq.Header.Set("User-Agent", "pe-toolkit/1.0")

	// Execute request
	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("error executing request: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Anthropic API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	// Create response channel
	respChan := make(chan *llm.StreamResponse, 10)

	// Start streaming goroutine
	go p.handleStreamResponse(ctx, resp, respChan)

	return respChan, nil
}

// handleStreamResponse processes the streaming response from Anthropic
func (p *AnthropicProvider) handleStreamResponse(ctx context.Context, resp *http.Response, respChan chan<- *llm.StreamResponse) {
	defer close(respChan)
	defer resp.Body.Close()

	startTime := time.Now()
	scanner := NewSSEScanner(resp.Body)

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			respChan <- &llm.StreamResponse{Error: ctx.Err()}
			return
		default:
		}

		event := scanner.Event()

		// Handle different event types
		switch event.Type {
		case "content_block_delta":
			var delta struct {
				Type  string `json:"type"`
				Index int    `json:"index"`
				Delta struct {
					Type string `json:"type"`
					Text string `json:"text"`
				} `json:"delta"`
			}

			if err := json.Unmarshal([]byte(event.Data), &delta); err != nil {
				respChan <- &llm.StreamResponse{Error: fmt.Errorf("error parsing delta: %v", err)}
				return
			}

			respChan <- &llm.StreamResponse{
				Text:    delta.Delta.Text,
				Latency: time.Since(startTime),
			}

		case "message_stop":
			respChan <- &llm.StreamResponse{
				Done:    true,
				Latency: time.Since(startTime),
			}
			return

		case "error":
			var errEvent struct {
				Error struct {
					Type    string `json:"type"`
					Message string `json:"message"`
				} `json:"error"`
			}

			if err := json.Unmarshal([]byte(event.Data), &errEvent); err == nil {
				respChan <- &llm.StreamResponse{Error: fmt.Errorf("Anthropic API error: %s", errEvent.Error.Message)}
			} else {
				respChan <- &llm.StreamResponse{Error: fmt.Errorf("unknown Anthropic API error")}
			}
			return
		}
	}

	if err := scanner.Err(); err != nil {
		respChan <- &llm.StreamResponse{Error: fmt.Errorf("error reading stream: %v", err)}
	}
}

// SupportsStreaming returns whether the provider supports streaming
func (p *AnthropicProvider) SupportsStreaming() bool {
	return true
}

// SupportsBatch returns whether the provider supports batch processing
func (p *AnthropicProvider) SupportsBatch() bool {
	return false
}

// calculateAnthropicCost calculates the approximate cost for Anthropic API usage
func calculateAnthropicCost(model string, inputTokens, outputTokens int) float64 {
	// Pricing as of 2024 (approximate, in USD per 1K tokens)
	costs := map[string]struct {
		input  float64
		output float64
	}{
		"claude-3-opus-20240229":   {0.015, 0.075},
		"claude-3-sonnet-20240229": {0.003, 0.015},
		"claude-3-haiku-20240307":  {0.00025, 0.00125},
		"claude-2.1":               {0.008, 0.024},
		"claude-2.0":               {0.008, 0.024},
		"claude-instant-1.2":       {0.00163, 0.00551},
	}

	cost, exists := costs[model]
	if !exists {
		// Default to Claude-3 Sonnet pricing for unknown models
		cost = costs["claude-3-sonnet-20240229"]
	}

	inputCost := (float64(inputTokens) / 1000.0) * cost.input
	outputCost := (float64(outputTokens) / 1000.0) * cost.output

	return inputCost + outputCost
}

// EvaluatePrompt implements the legacy Provider interface for backward compatibility
func (p *AnthropicProvider) EvaluatePrompt(ctx context.Context, prompt string, vars map[string]interface{}) (*promptfoo.ProviderResponse, error) {
	// Convert vars to GenerateOptions
	options := llm.GenerateOptions{}

	if temp, ok := vars["temperature"].(float64); ok {
		options.Temperature = &temp
	}
	if maxTokens, ok := vars["max_tokens"].(int); ok {
		options.MaxTokens = &maxTokens
	}

	// Generate response
	resp, err := p.Generate(ctx, prompt, options)
	if err != nil {
		return nil, err
	}

	// Convert to promptfoo response format
	return &promptfoo.ProviderResponse{
		Output: resp.Text,
		TokenUsage: &promptfoo.TokenUsage{
			Total:      int32(resp.TotalTokens),
			Prompt:     int32(resp.PromptTokens),
			Completion: int32(resp.CompletionTokens),
			Cached:     0,
		},
		Cost:   resp.Cost,
		Cached: false,
	}, nil
}
