package providers

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"text/template"

	"github.com/kballard/go-shellquote"
	"github.com/tmc/pe/internal/llm"
	"github.com/tmc/pe/internal/promptfoo"
)

// GenericCLIProvider implements the Provider interface for any CLI tool
type GenericCLIProvider struct {
	commandTemplate string
	model           string
	env             map[string]string
}

// NewGenericCLIProvider creates a new generic CLI provider
func NewGenericCLIProvider(model string, options map[string]interface{}) (*GenericCLIProvider, error) {
	cmdTemplate := getStringOption(options, "command", "")
	if cmdTemplate == "" {
		return nil, fmt.Errorf("generic cli provider requires 'command' option")
	}

	env := make(map[string]string)
	if envMap, ok := options["env"].(map[string]interface{}); ok {
		for k, v := range envMap {
			env[k] = fmt.Sprintf("%v", v)
		}
	}

	return &GenericCLIProvider{
		commandTemplate: cmdTemplate,
		model:           model,
		env:             env,
	}, nil
}

// Name returns the provider name
func (p *GenericCLIProvider) Name() string {
	return "cli"
}

// Model returns the model name
func (p *GenericCLIProvider) Model() string {
	return p.model
}

type templateData struct {
	Model       string
	Prompt      string
	Temperature float64
	MaxTokens   int
}

// Generate generates a response using the configured CLI command
func (p *GenericCLIProvider) Generate(ctx context.Context, prompt string, options llm.GenerateOptions) (*llm.GenerateResponse, error) {
	// Prepare template data
	data := templateData{
		Model:       p.model,
		Prompt:      prompt,
		Temperature: 0.7,  // Default
		MaxTokens:   1000, // Default
	}

	if options.Temperature != nil {
		data.Temperature = *options.Temperature
	}
	if options.MaxTokens != nil {
		data.MaxTokens = *options.MaxTokens
	}

	// Parse and execute template
	tmpl, err := template.New("cli").Parse(p.commandTemplate)
	if err != nil {
		return nil, fmt.Errorf("failed to parse command template: %w", err)
	}

	var cmdStr bytes.Buffer
	if err := tmpl.Execute(&cmdStr, data); err != nil {
		return nil, fmt.Errorf("failed to execute command template: %w", err)
	}

	// Parse command string into executable and arguments
	// We use shellquote to handle quoted arguments correctly
	parts, err := shellquote.Split(cmdStr.String())
	if err != nil {
		return nil, fmt.Errorf("failed to split command string: %w", err)
	}
	if len(parts) == 0 {
		return nil, fmt.Errorf("empty command resulted from template")
	}

	executable := parts[0]
	args := parts[1:]

	// Look for executable
	if _, err := exec.LookPath(executable); err != nil {
		return nil, fmt.Errorf("executable not found: %s", executable)
	}

	cmd := exec.CommandContext(ctx, executable, args...)

	// Set environment variables
	cmd.Env = os.Environ()
	for k, v := range p.env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}

	var out bytes.Buffer
	var errBuf bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errBuf

	if err := cmd.Run(); err != nil {
		if errBuf.Len() > 0 {
			return nil, fmt.Errorf("cli error: %s", errBuf.String())
		}
		return nil, fmt.Errorf("cli execution failed: %w", err)
	}

	return &llm.GenerateResponse{
		Text:  strings.TrimSpace(out.String()),
		Model: p.model,
	}, nil
}

// EvaluatePrompt implements the legacy Provider interface method
func (p *GenericCLIProvider) EvaluatePrompt(ctx context.Context, prompt string, vars map[string]interface{}) (*promptfoo.ProviderResponse, error) {
	// Apply variables to prompt if needed
	finalPrompt := prompt
	if len(vars) > 0 {
		for key, value := range vars {
			placeholder := fmt.Sprintf("{{%s}}", key)
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

func (p *GenericCLIProvider) SupportsStreaming() bool {
	return false
}

func (p *GenericCLIProvider) SupportsBatch() bool {
	return false
}
