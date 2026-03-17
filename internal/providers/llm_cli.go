package providers

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/tmc/pe/internal/llm"
	"github.com/tmc/pe/internal/promptfoo"
)

// LLMCLIProvider implements the Provider interface using the llm CLI tool
type LLMCLIProvider struct {
	executable string
	model      string
}

// NewLLMCLIProvider creates a new llm CLI provider
func NewLLMCLIProvider(model string, options map[string]interface{}) (*LLMCLIProvider, error) {
	executable := getStringOption(options, "executable", "llm")

	// Check if llm is available
	if _, err := exec.LookPath(executable); err != nil {
		return nil, fmt.Errorf("llm executable not found: %w", err)
	}

	return &LLMCLIProvider{
		executable: executable,
		model:      model,
	}, nil
}

// Name returns the provider name
func (p *LLMCLIProvider) Name() string {
	return "llm"
}

// Model returns the model name
func (p *LLMCLIProvider) Model() string {
	return p.model
}

// Generate generates a response using the llm CLI
func (p *LLMCLIProvider) Generate(ctx context.Context, prompt string, options llm.GenerateOptions) (*llm.GenerateResponse, error) {
	args := []string{}

	// Model selection
	if p.model != "" && p.model != "default" {
		args = append(args, "-m", p.model)
	}

	// System prompt - GenerateOptions does not have a System field currently
	// If we need system prompt support, we might need to extend GenerateOptions or handle it in the prompt string

	// Options
	if options.Temperature != nil {
		// llm supports specific options via -o
		args = append(args, "-o", "temperature", fmt.Sprintf("%f", *options.Temperature))
	}

	if options.MaxTokens != nil {
		args = append(args, "-o", "max_tokens", fmt.Sprintf("%d", *options.MaxTokens))
	}

	// Add prompt as the last argument
	args = append(args, prompt)

	cmd := exec.CommandContext(ctx, p.executable, args...)

	var out bytes.Buffer
	var errBuf bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errBuf

	start := time.Now()
	if err := cmd.Run(); err != nil {
		if errBuf.Len() > 0 {
			return nil, fmt.Errorf("llm error: %s", errBuf.String())
		}
		return nil, fmt.Errorf("llm execution failed: %w", err)
	}
	return parseCLIResponse(strings.TrimSpace(out.String()), p.model, time.Since(start)), nil
}

// EvaluatePrompt implements the legacy Provider interface method
func (p *LLMCLIProvider) EvaluatePrompt(ctx context.Context, prompt string, vars map[string]interface{}) (*promptfoo.ProviderResponse, error) {
	// Apply variables to prompt if needed
	finalPrompt := prompt
	if len(vars) > 0 {
		// Simple variable substitution
		for key, value := range vars {
			placeholder := fmt.Sprintf("{{%s}}", key)
			finalPrompt = strings.ReplaceAll(finalPrompt, placeholder, fmt.Sprintf("%v", value))
			placeholder = fmt.Sprintf("{{.%s}}", key)
			finalPrompt = strings.ReplaceAll(finalPrompt, placeholder, fmt.Sprintf("%v", value))
		}
	}

	result, err := p.Generate(ctx, finalPrompt, llm.GenerateOptions{})
	if err != nil {
		return nil, err
	}

	return &promptfoo.ProviderResponse{
		Output: result.Text,
		TokenUsage: &promptfoo.TokenUsage{
			Total:      int32(result.TotalTokens),
			Prompt:     int32(result.PromptTokens),
			Completion: int32(result.CompletionTokens),
		},
		Cost:      result.Cost,
		LatencyMs: result.Latency.Milliseconds(),
		Metadata:  result.Metadata,
	}, nil
}

// SupportsStreaming returns whether the provider supports streaming
// The llm CLI supports streaming, but wrapping it is complex, enabling later if needed
func (p *LLMCLIProvider) SupportsStreaming() bool {
	return false
}

// SupportsBatch returns whether the provider supports batch processing
func (p *LLMCLIProvider) SupportsBatch() bool {
	return false
}
