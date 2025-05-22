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

// OpenAIProvider implements the LLM provider interface for OpenAI API
type OpenAIProvider struct {
	client     *http.Client
	apiKey     string
	baseURL    string
	model      string
	maxRetries int
}

// OpenAIRequest represents the request structure for OpenAI API
type OpenAIRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature *float64  `json:"temperature,omitempty"`
	MaxTokens   *int      `json:"max_tokens,omitempty"`
	Stream      bool      `json:"stream,omitempty"`
}

// Message represents a chat message
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// OpenAIResponse represents the response structure from OpenAI API
type OpenAIResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
}

// Choice represents a response choice
type Choice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
}

// Usage represents token usage information
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// ErrorResponse represents an error response from OpenAI API
type ErrorResponse struct {
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    string `json:"code"`
	} `json:"error"`
}

// NewOpenAIProvider creates a new OpenAI provider instance
func NewOpenAIProvider(model string, options map[string]interface{}) (*OpenAIProvider, error) {
	apiKey := getStringOption(options, "apiKey", "")
	if apiKey == "" {
		// Try environment variable
		apiKey = getEnvVar("OPENAI_API_KEY")
		if apiKey == "" {
			return nil, fmt.Errorf("OpenAI API key not provided")
		}
	}

	baseURL := getStringOption(options, "baseURL", "https://api.openai.com/v1")
	maxRetries := getIntOption(options, "maxRetries", 3)

	client := &http.Client{
		Timeout: time.Minute,
	}

	return &OpenAIProvider{
		client:     client,
		apiKey:     apiKey,
		baseURL:    strings.TrimSuffix(baseURL, "/"),
		model:      model,
		maxRetries: maxRetries,
	}, nil
}

// Name returns the provider name
func (p *OpenAIProvider) Name() string {
	return "openai"
}

// Model returns the model being used
func (p *OpenAIProvider) Model() string {
	return p.model
}

// Generate sends a prompt to OpenAI and returns the response
func (p *OpenAIProvider) Generate(ctx context.Context, prompt string, options llm.GenerateOptions) (*llm.GenerateResponse, error) {
	startTime := time.Now()

	// Build request
	req := OpenAIRequest{
		Model: p.model,
		Messages: []Message{
			{Role: "user", Content: prompt},
		},
	}

	// Apply options
	if options.Temperature != nil {
		req.Temperature = options.Temperature
	}
	if options.MaxTokens != nil {
		req.MaxTokens = options.MaxTokens
	}

	// Marshal request
	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %v", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/chat/completions", bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)
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
		var errResp ErrorResponse
		if err := json.Unmarshal(respBody, &errResp); err == nil {
			return nil, fmt.Errorf("OpenAI API error: %s", errResp.Error.Message)
		}
		return nil, fmt.Errorf("OpenAI API error: status %d, body: %s", resp.StatusCode, string(respBody))
	}

	// Parse successful response
	var openaiResp OpenAIResponse
	if err := json.Unmarshal(respBody, &openaiResp); err != nil {
		return nil, fmt.Errorf("error parsing response: %v", err)
	}

	if len(openaiResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	latency := time.Since(startTime)
	
	// Calculate cost (approximate)
	cost := calculateOpenAICost(p.model, openaiResp.Usage.PromptTokens, openaiResp.Usage.CompletionTokens)

	return &llm.GenerateResponse{
		Text:             openaiResp.Choices[0].Message.Content,
		PromptTokens:     openaiResp.Usage.PromptTokens,
		CompletionTokens: openaiResp.Usage.CompletionTokens,
		TotalTokens:      openaiResp.Usage.TotalTokens,
		Latency:          latency,
		Cost:             cost,
		Model:            openaiResp.Model,
		FinishReason:     openaiResp.Choices[0].FinishReason,
	}, nil
}

// GenerateStream sends a prompt to OpenAI and returns a streaming response
func (p *OpenAIProvider) GenerateStream(ctx context.Context, prompt string, options llm.GenerateOptions) (<-chan *llm.StreamResponse, error) {
	// Build request with streaming enabled
	req := OpenAIRequest{
		Model:  p.model,
		Stream: true,
		Messages: []Message{
			{Role: "user", Content: prompt},
		},
	}

	// Apply options
	if options.Temperature != nil {
		req.Temperature = options.Temperature
	}
	if options.MaxTokens != nil {
		req.MaxTokens = options.MaxTokens
	}

	// Marshal request
	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %v", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/chat/completions", bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)
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
		return nil, fmt.Errorf("OpenAI API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	// Create response channel
	respChan := make(chan *llm.StreamResponse, 10)

	// Start streaming goroutine
	go p.handleStreamResponse(ctx, resp, respChan)

	return respChan, nil
}

// handleStreamResponse processes the streaming response from OpenAI
func (p *OpenAIProvider) handleStreamResponse(ctx context.Context, resp *http.Response, respChan chan<- *llm.StreamResponse) {
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
		if event.Data == "[DONE]" {
			respChan <- &llm.StreamResponse{
				Done:    true,
				Latency: time.Since(startTime),
			}
			return
		}

		// Parse the streaming response
		var streamResp struct {
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

		if err := json.Unmarshal([]byte(event.Data), &streamResp); err != nil {
			respChan <- &llm.StreamResponse{Error: fmt.Errorf("error parsing stream response: %v", err)}
			return
		}

		if len(streamResp.Choices) > 0 {
			choice := streamResp.Choices[0]
			respChan <- &llm.StreamResponse{
				Text:    choice.Delta.Content,
				Latency: time.Since(startTime),
			}
		}
	}

	if err := scanner.Err(); err != nil {
		respChan <- &llm.StreamResponse{Error: fmt.Errorf("error reading stream: %v", err)}
	}
}

// SupportsStreaming returns whether the provider supports streaming
func (p *OpenAIProvider) SupportsStreaming() bool {
	return true
}

// SupportsBatch returns whether the provider supports batch processing
func (p *OpenAIProvider) SupportsBatch() bool {
	return false // OpenAI doesn't have native batch support in the chat API
}

// calculateOpenAICost calculates the approximate cost for OpenAI API usage
func calculateOpenAICost(model string, promptTokens, completionTokens int) float64 {
	// Pricing as of 2024 (approximate, in USD per 1K tokens)
	costs := map[string]struct {
		input  float64
		output float64
	}{
		"gpt-4":                    {0.03, 0.06},
		"gpt-4-turbo":              {0.01, 0.03},
		"gpt-4-turbo-preview":      {0.01, 0.03},
		"gpt-4-0125-preview":       {0.01, 0.03},
		"gpt-4-1106-preview":       {0.01, 0.03},
		"gpt-3.5-turbo":            {0.0015, 0.002},
		"gpt-3.5-turbo-0125":       {0.0005, 0.0015},
		"gpt-3.5-turbo-instruct":   {0.0015, 0.002},
	}

	cost, exists := costs[model]
	if !exists {
		// Default to GPT-4 pricing for unknown models
		cost = costs["gpt-4"]
	}

	inputCost := (float64(promptTokens) / 1000.0) * cost.input
	outputCost := (float64(completionTokens) / 1000.0) * cost.output

	return inputCost + outputCost
}

// EvaluatePrompt implements the legacy Provider interface for backward compatibility
func (p *OpenAIProvider) EvaluatePrompt(ctx context.Context, prompt string, vars map[string]interface{}) (*promptfoo.ProviderResponse, error) {
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