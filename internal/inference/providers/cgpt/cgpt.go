// Package cgpt provides an inference provider using the cgpt CLI tool.
package cgpt

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/tmc/pe/internal/inference"
)

func init() {
	// Register cgpt provider
	inference.MustRegister("cgpt", func(config map[string]interface{}) (inference.Provider, error) {
		// Check if binary path is specified
		if config != nil {
			if path, ok := config["binary"].(string); ok && path != "" {
				return NewWithBinary(path), nil
			}
		}
		return New(), nil
	})
}

// Provider implements the inference.Provider interface using cgpt.
type Provider struct {
	// Path to cgpt binary (default: uses go run)
	binaryPath string
	
	// Whether to use go run instead of a binary
	useGoRun bool
}

// New creates a new cgpt provider.
func New() *Provider {
	// Check if cgpt is available in PATH
	if path, err := exec.LookPath("cgpt"); err == nil {
		return &Provider{
			binaryPath: path,
			useGoRun:   false,
		}
	}
	
	// Fall back to go run
	return &Provider{
		useGoRun: true,
	}
}

// NewWithBinary creates a cgpt provider using a specific binary path.
func NewWithBinary(path string) *Provider {
	return &Provider{
		binaryPath: path,
		useGoRun:   false,
	}
}

// Name returns the provider name.
func (p *Provider) Name() string {
	return "cgpt"
}

// Complete performs a non-streaming inference using cgpt.
func (p *Provider) Complete(ctx context.Context, req inference.Request) (*inference.Response, error) {
	args := p.buildArgs(req)
	
	cmd := p.buildCommand(ctx, args...)
	
	// Capture output
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	
	// Run the command
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("cgpt error: %w\nstderr: %s", err, stderr.String())
	}
	
	// Parse the response
	content := stdout.String()
	
	return &inference.Response{
		Content: strings.TrimSpace(content),
		Model:   req.Model,
		TokensUsed: inference.TokenUsage{
			// cgpt doesn't provide token counts by default
			TotalTokens: -1,
		},
		Metadata: map[string]interface{}{
			"provider": "cgpt",
		},
	}, nil
}

// Stream performs a streaming inference using cgpt.
func (p *Provider) Stream(ctx context.Context, req inference.Request) (<-chan inference.StreamChunk, error) {
	args := p.buildArgs(req)
	
	// cgpt streams by default when stdout is a terminal
	// We'll capture line by line
	
	cmd := p.buildCommand(ctx, args...)
	
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}
	
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stderr pipe: %w", err)
	}
	
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start cgpt: %w", err)
	}
	
	chunks := make(chan inference.StreamChunk)
	
	go func() {
		defer close(chunks)
		defer cmd.Wait()
		
		// Read stderr for any errors
		var stderrBuf bytes.Buffer
		go func() {
			scanner := bufio.NewScanner(stderr)
			for scanner.Scan() {
				stderrBuf.WriteString(scanner.Text() + "\n")
			}
		}()
		
		// Read stdout and send chunks
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			select {
			case <-ctx.Done():
				cmd.Process.Kill()
				chunks <- inference.StreamChunk{
					Error: ctx.Err(),
				}
				return
			case chunks <- inference.StreamChunk{
				Delta: scanner.Text() + "\n",
				Done:  false,
			}:
			}
		}
		
		if err := scanner.Err(); err != nil {
			chunks <- inference.StreamChunk{
				Error: fmt.Errorf("read error: %w", err),
			}
			return
		}
		
		// Check if cgpt had any errors
		if err := cmd.Wait(); err != nil {
			chunks <- inference.StreamChunk{
				Error: fmt.Errorf("cgpt error: %w\nstderr: %s", err, stderrBuf.String()),
			}
			return
		}
		
		// Send done signal
		chunks <- inference.StreamChunk{
			Done: true,
		}
	}()
	
	return chunks, nil
}

// Models returns available models for cgpt.
func (p *Provider) Models(ctx context.Context) ([]string, error) {
	// cgpt supports various models via -m flag
	// These are the commonly available ones
	return []string{
		"gpt-4",
		"gpt-4-turbo-preview", 
		"gpt-3.5-turbo",
		"claude-3-opus",
		"claude-3-sonnet",
		"claude-3-haiku",
	}, nil
}

// Close cleans up any resources.
func (p *Provider) Close() error {
	// Nothing to clean up for CLI-based provider
	return nil
}

// buildCommand creates the exec.Cmd for cgpt.
func (p *Provider) buildCommand(ctx context.Context, args ...string) *exec.Cmd {
	if p.useGoRun {
		// Use go run
		fullArgs := append([]string{"run", "github.com/tmc/cgpt/cmd/cgpt"}, args...)
		return exec.CommandContext(ctx, "go", fullArgs...)
	}
	
	// Use binary
	return exec.CommandContext(ctx, p.binaryPath, args...)
}

// buildArgs builds command line arguments for cgpt.
func (p *Provider) buildArgs(req inference.Request) []string {
	var args []string
	
	// Model selection and backend detection
	if req.Model != "" {
		args = append(args, "-m", req.Model)
		
		// Detect backend based on model name
		if strings.HasPrefix(req.Model, "gpt-") {
			args = append(args, "-b", "openai")
		} else if strings.HasPrefix(req.Model, "claude-") {
			args = append(args, "-b", "anthropic")
		}
	}
	
	// Temperature (cgpt uses -T for temperature)
	if req.Temperature > 0 {
		args = append(args, "-T", fmt.Sprintf("%.2f", req.Temperature))
	}
	
	// Max tokens (cgpt uses -t for max-tokens)
	if req.MaxTokens > 0 {
		args = append(args, "-t", fmt.Sprintf("%d", req.MaxTokens))
	}
	
	// System prompt
	if req.SystemPrompt != "" {
		args = append(args, "-s", req.SystemPrompt)
	}
	
	// Input (use -i for direct string input)
	if req.Prompt != "" {
		args = append(args, "-i", req.Prompt)
	}
	
	// Additional options
	if req.Options != nil {
		// Handle any cgpt-specific options
		if v, ok := req.Options["json"].(bool); ok && v {
			args = append(args, "--json")
		}
		if v, ok := req.Options["verbose"].(bool); ok && v {
			args = append(args, "-v")
		}
	}
	
	return args
}

// ParseJSONResponse attempts to parse a JSON response from cgpt.
// This is useful when using the --json flag.
func ParseJSONResponse(content string) (map[string]interface{}, error) {
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %w", err)
	}
	return result, nil
}