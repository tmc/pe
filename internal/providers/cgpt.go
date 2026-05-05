package providers

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/tmc/pe/internal/llm"
	"github.com/tmc/pe/internal/promptfoo"
	"github.com/tmc/pe/internal/security"
)

// CGPTProvider implements the Provider interface using the cgpt CLI tool
type CGPTProvider struct {
	executable string
}

// NewCGPTProvider creates a new cgpt provider
func NewCGPTProvider(options map[string]interface{}) (*CGPTProvider, error) {
	executable := getStringOption(options, "executable", "cgpt")

	// Check if cgpt is available
	if _, err := exec.LookPath(executable); err != nil {
		return nil, fmt.Errorf("cgpt executable not found: %w", err)
	}

	return &CGPTProvider{
		executable: executable,
	}, nil
}

// Name returns the provider name
func (p *CGPTProvider) Name() string {
	return "cgpt"
}

// Model returns the model name
func (p *CGPTProvider) Model() string {
	return "default"
}

// Generate generates a response using cgpt
func (p *CGPTProvider) Generate(ctx context.Context, prompt string, options llm.GenerateOptions) (*llm.GenerateResponse, error) {
	cmd := exec.CommandContext(ctx, p.executable)
	cmd.Stdin = strings.NewReader(prompt)

	var out bytes.Buffer
	var errBuf bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errBuf

	if err := cmd.Run(); err != nil {
		if errBuf.Len() > 0 {
			return nil, fmt.Errorf("cgpt error: %s", security.RedactSecrets(errBuf.String()))
		}
		return nil, fmt.Errorf("cgpt execution failed: %w", err)
	}

	return &llm.GenerateResponse{
		Text:  out.String(),
		Model: "cgpt",
	}, nil
}

// EvaluatePrompt implements the legacy Provider interface method
func (p *CGPTProvider) EvaluatePrompt(ctx context.Context, prompt string, vars map[string]interface{}) (*promptfoo.ProviderResponse, error) {
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
	}, nil
}

// SupportsStreaming returns whether the provider supports streaming
func (p *CGPTProvider) SupportsStreaming() bool {
	return false
}

// SupportsBatch returns whether the provider supports batch processing
func (p *CGPTProvider) SupportsBatch() bool {
	return false
}
