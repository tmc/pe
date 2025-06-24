// Package ollama provides a native Ollama API implementation for the inference provider interface.
package ollama

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
	// Register Ollama provider
	inference.MustRegister("ollama", Factory)
}

// Factory creates a new Ollama provider from configuration.
func Factory(config map[string]interface{}) (inference.Provider, error) {
	baseURL := "http://localhost:11434"
	if config != nil {
		if u, ok := config["base_url"].(string); ok {
			baseURL = u
		}
	}
	if envURL := os.Getenv("OLLAMA_HOST"); envURL != "" {
		baseURL = envURL
	}

	return New(baseURL), nil
}

// Provider implements the inference.Provider interface for Ollama.
type Provider struct {
	baseURL string
	client  *http.Client
}

// New creates a new Ollama provider.
func New(baseURL string) *Provider {
	return &Provider{
		baseURL: strings.TrimSuffix(baseURL, "/"),
		client: &http.Client{
			Timeout: 5 * time.Minute,
		},
	}
}

// Name returns the provider name.
func (p *Provider) Name() string {
	return "ollama"
}

// generateRequest represents the Ollama generate request.
type generateRequest struct {
	Model     string            `json:"model"`
	Prompt    string            `json:"prompt"`
	System    string            `json:"system,omitempty"`
	Stream    bool              `json:"stream,omitempty"`
	Options   map[string]interface{} `json:"options,omitempty"`
	Context   []int             `json:"context,omitempty"`
	KeepAlive string            `json:"keep_alive,omitempty"`
}

// generateResponse represents the Ollama generate response.
type generateResponse struct {
	Model     string `json:"model"`
	CreatedAt string `json:"created_at"`
	Response  string `json:"response"`
	Done      bool   `json:"done"`
	Context   []int  `json:"context,omitempty"`
	
	// Token usage information (available when done=true)
	TotalDuration      int64 `json:"total_duration,omitempty"`
	LoadDuration       int64 `json:"load_duration,omitempty"`
	PromptEvalCount    int   `json:"prompt_eval_count,omitempty"`
	PromptEvalDuration int64 `json:"prompt_eval_duration,omitempty"`
	EvalCount          int   `json:"eval_count,omitempty"`
	EvalDuration       int64 `json:"eval_duration,omitempty"`
}

// modelsResponse represents the Ollama models list response.
type modelsResponse struct {
	Models []struct {
		Name         string `json:"name"`
		ModifiedAt   string `json:"modified_at"`
		Size         int64  `json:"size"`
		Digest       string `json:"digest"`
		Details      struct {
			Format            string   `json:"format"`
			Family            string   `json:"family"`
			Families          []string `json:"families"`
			ParameterSize     string   `json:"parameter_size"`
			QuantizationLevel string   `json:"quantization_level"`
		} `json:"details"`
	} `json:"models"`
}

// Complete performs a non-streaming inference.
func (p *Provider) Complete(ctx context.Context, req inference.Request) (*inference.Response, error) {
	model := req.Model
	if model == "" {
		model = "llama2" // Default model
	}

	// Build options from request
	options := make(map[string]interface{})
	if req.Temperature > 0 {
		options["temperature"] = req.Temperature
	}
	if req.MaxTokens > 0 {
		options["num_predict"] = req.MaxTokens
	}
	if len(req.StopSequences) > 0 {
		options["stop"] = req.StopSequences
	}
	// Merge with any provider-specific options
	for k, v := range req.Options {
		options[k] = v
	}

	reqBody := generateRequest{
		Model:   model,
		Prompt:  req.Prompt,
		System:  req.SystemPrompt,
		Stream:  false,
		Options: options,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/api/generate", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ollama API error %d: %s", resp.StatusCode, string(body))
	}

	var genResp generateResponse
	if err := json.NewDecoder(resp.Body).Decode(&genResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &inference.Response{
		Content: genResp.Response,
		Model:   genResp.Model,
		TokensUsed: inference.TokenUsage{
			PromptTokens:     genResp.PromptEvalCount,
			CompletionTokens: genResp.EvalCount,
			TotalTokens:      genResp.PromptEvalCount + genResp.EvalCount,
		},
		Metadata: map[string]interface{}{
			"total_duration":      genResp.TotalDuration,
			"load_duration":       genResp.LoadDuration,
			"prompt_eval_duration": genResp.PromptEvalDuration,
			"eval_duration":       genResp.EvalDuration,
		},
	}, nil
}

// Stream performs a streaming inference.
func (p *Provider) Stream(ctx context.Context, req inference.Request) (<-chan inference.StreamChunk, error) {
	model := req.Model
	if model == "" {
		model = "llama2" // Default model
	}

	// Build options from request
	options := make(map[string]interface{})
	if req.Temperature > 0 {
		options["temperature"] = req.Temperature
	}
	if req.MaxTokens > 0 {
		options["num_predict"] = req.MaxTokens
	}
	if len(req.StopSequences) > 0 {
		options["stop"] = req.StopSequences
	}
	// Merge with any provider-specific options
	for k, v := range req.Options {
		options[k] = v
	}

	reqBody := generateRequest{
		Model:   model,
		Prompt:  req.Prompt,
		System:  req.SystemPrompt,
		Stream:  true,
		Options: options,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		ch := make(chan inference.StreamChunk, 1)
		ch <- inference.StreamChunk{Error: fmt.Errorf("failed to marshal request: %w", err)}
		close(ch)
		return ch, nil
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/api/generate", bytes.NewBuffer(jsonData))
	if err != nil {
		ch := make(chan inference.StreamChunk, 1)
		ch <- inference.StreamChunk{Error: fmt.Errorf("failed to create request: %w", err)}
		close(ch)
		return ch, nil
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(httpReq)
	if err != nil {
		ch := make(chan inference.StreamChunk, 1)
		ch <- inference.StreamChunk{Error: fmt.Errorf("failed to make request: %w", err)}
		close(ch)
		return ch, nil
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		ch := make(chan inference.StreamChunk, 1)
		ch <- inference.StreamChunk{Error: fmt.Errorf("ollama API error %d: %s", resp.StatusCode, string(body))}
		close(ch)
		return ch, nil
	}

	ch := make(chan inference.StreamChunk)
	go func() {
		defer close(ch)
		defer resp.Body.Close()

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			var genResp generateResponse
			if err := json.Unmarshal(scanner.Bytes(), &genResp); err != nil {
				ch <- inference.StreamChunk{Error: fmt.Errorf("failed to decode streaming response: %w", err)}
				return
			}

			ch <- inference.StreamChunk{
				Delta: genResp.Response,
				Done:  genResp.Done,
			}

			if genResp.Done {
				break
			}
		}

		if err := scanner.Err(); err != nil {
			ch <- inference.StreamChunk{Error: fmt.Errorf("error reading stream: %w", err)}
		}
	}()

	return ch, nil
}

// Models returns available models for this provider.
func (p *Provider) Models(ctx context.Context) ([]string, error) {
	httpReq, err := http.NewRequestWithContext(ctx, "GET", p.baseURL+"/api/tags", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ollama API error %d: %s", resp.StatusCode, string(body))
	}

	var modelsResp modelsResponse
	if err := json.NewDecoder(resp.Body).Decode(&modelsResp); err != nil {
		return nil, fmt.Errorf("failed to decode models response: %w", err)
	}

	models := make([]string, len(modelsResp.Models))
	for i, model := range modelsResp.Models {
		models[i] = model.Name
	}

	return models, nil
}

// Close cleans up any resources.
func (p *Provider) Close() error {
	// Ollama provider doesn't maintain persistent connections
	return nil
}