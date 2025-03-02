// Package evaluator provides interfaces and utilities for working with prompt evaluators.
package evaluator

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"
)

// Result represents a structured evaluation result.
type Result struct {
	Output     string      `json:"output"`
	Success    bool        `json:"success"`
	Metrics    *Metrics    `json:"metrics,omitempty"`
	Error      string      `json:"error,omitempty"`
	FinishTime time.Time   `json:"finishTime,omitempty"`
	Cost       float64     `json:"cost,omitempty"`
	Metadata   interface{} `json:"metadata,omitempty"`
}

// Metrics contains evaluation metrics.
type Metrics struct {
	TokenCount   int     `json:"tokenCount,omitempty"`
	PromptTokens int     `json:"promptTokens,omitempty"`
	Score        float64 `json:"score,omitempty"`
	LatencyMs    int     `json:"latencyMs,omitempty"`
}

// Provider is an interface for evaluating prompts.
type Provider interface {
	// Evaluate evaluates a prompt with the given variables and returns a result.
	Evaluate(ctx context.Context, prompt string, vars map[string]interface{}) (*Result, error)
	
	// DryRun returns a command that would be used to evaluate the prompt without executing it.
	DryRun(prompt string, vars map[string]interface{}) (string, error)
	
	// Name returns the name of the provider.
	Name() string
}

// ExternalProvider represents an external evaluator executable.
type ExternalProvider struct {
	Path      string
	Arguments []string
	Timeout   time.Duration
}

// NewExternalProvider creates a new external provider with the given path and arguments.
func NewExternalProvider(path string, args ...string) *ExternalProvider {
	return &ExternalProvider{
		Path:      path,
		Arguments: args,
		Timeout:   30 * time.Second,
	}
}

// Evaluate implements the Provider interface for external executables.
func (p *ExternalProvider) Evaluate(ctx context.Context, prompt string, vars map[string]interface{}) (*Result, error) {
	// Create command with context for timeout
	ctx, cancel := context.WithTimeout(ctx, p.Timeout)
	defer cancel()
	
	cmd := exec.CommandContext(ctx, p.Path, p.Arguments...)
	
	// Set up pipes for stdin/stdout
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdin pipe: %w", err)
	}
	
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}
	
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stderr pipe: %w", err)
	}
	
	// Start the command
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start command: %w", err)
	}
	
	// Write prompt to stdin
	if _, err := io.WriteString(stdin, prompt); err != nil {
		return nil, fmt.Errorf("failed to write to stdin: %w", err)
	}
	stdin.Close()
	
	// Read stdout
	output, err := io.ReadAll(stdout)
	if err != nil {
		return nil, fmt.Errorf("failed to read stdout: %w", err)
	}
	
	// Read stderr
	errOutput, err := io.ReadAll(stderr)
	if err != nil {
		return nil, fmt.Errorf("failed to read stderr: %w", err)
	}
	
	// Wait for command to finish
	err = cmd.Wait()
	
	// Check for timeout
	if ctx.Err() == context.DeadlineExceeded {
		return &Result{
			Success: false,
			Error:   "timeout exceeded",
		}, nil
	}
	
	// Create result
	result := &Result{
		Output:     string(output),
		Success:    err == nil,
		FinishTime: time.Now(),
	}
	
	if err != nil {
		result.Error = fmt.Sprintf("command failed: %v\n%s", err, errOutput)
	}
	
	return result, nil
}

// DryRun implements the Provider interface for external executables.
func (p *ExternalProvider) DryRun(prompt string, vars map[string]interface{}) (string, error) {
	// Format a command that would be executed
	args := append([]string{p.Path}, p.Arguments...)
	return fmt.Sprintf("%s | %s", prompt, exec.Command(args[0], args[1:]...).String()), nil
}

// Name returns the name of the external provider.
func (p *ExternalProvider) Name() string {
	return fmt.Sprintf("external(%s)", p.Path)
}

// FindExternalEvaluator searches for provider-specific or generic evaluator executables
// in the PATH and returns the first one found.
func FindExternalEvaluator(provider string) *ExternalProvider {
	// Try provider-specific evaluator
	if provider != "" {
		path, err := exec.LookPath(fmt.Sprintf("pe-eval-provider-%s", provider))
		if err == nil {
			return NewExternalProvider(path)
		}
	}
	
	// Try generic evaluator
	path, err := exec.LookPath("pe-eval")
	if err == nil {
		return NewExternalProvider(path)
	}
	
	// Try CGPT-based evaluator
	path, err = exec.LookPath("pe-eval-provider-cgpt")
	if err == nil {
		return NewExternalProvider(path)
	}
	
	// Try CGPT directly
	path, err = exec.LookPath("cgpt")
	if err == nil {
		return NewExternalProvider(path, "-b", provider)
	}
	
	return nil
}

// WriteResultToFile writes evaluation results to a file in the specified format.
func WriteResultToFile(result interface{}, filename, format string) error {
	var data []byte
	var err error
	
	// Format the result based on the requested format
	switch format {
	case "json":
		// Json formatting happens in the caller
	case "yaml":
		// Yaml formatting happens in the caller
	case "text":
		// Simple text format for human readability
		data = []byte(fmt.Sprintf("%v", result))
	default:
		return fmt.Errorf("unsupported output format: %s", format)
	}
	
	// Write to file or stdout
	if filename == "" || filename == "-" {
		_, err = os.Stdout.Write(data)
	} else {
		err = os.WriteFile(filename, data, 0644)
	}
	
	return err
}