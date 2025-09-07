// Package cgpt provides an inference provider using the cgpt CLI tool.
// Compatible with cgpt v0.4.4 and later versions.
// See https://github.com/tmc/cgpt for the latest version.
package cgpt

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
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
	// Path to cgpt binary (default: uses go tool)
	binaryPath string

	// Whether to use go tool instead of a binary
	useGoTool bool

	// Whether to use go run instead of go tool
	useGoRun bool
}

// New creates a new cgpt provider.
func New() *Provider {
	// Check if cgpt is available in PATH
	if path, err := exec.LookPath("cgpt"); err == nil {
		return &Provider{
			binaryPath: path,
			useGoTool:  false,
			useGoRun:   false,
		}
	}

	// Try to use go tool cgpt (Go 1.24+ with tool directive)
	// Test if go tool cgpt works
	if testCmd := exec.Command("go", "tool", "cgpt", "--version"); testCmd.Run() == nil {
		return &Provider{
			useGoTool: true,
			useGoRun:  false,
		}
	}

	// Fall back to go run
	return &Provider{
		useGoTool: false,
		useGoRun:  true,
	}
}

// NewWithBinary creates a cgpt provider using a specific binary path.
func NewWithBinary(path string) *Provider {
	return &Provider{
		binaryPath: path,
		useGoTool:  false,
		useGoRun:   false,
	}
}

// Name returns the provider name.
func (p *Provider) Name() string {
	return "cgpt"
}

// Complete performs a non-streaming inference using cgpt.
func (p *Provider) Complete(ctx context.Context, req inference.Request) (*inference.Response, error) {
	// Check for test mode
	if os.Getenv("PE_TEST_MODE") == "true" || os.Getenv("PE_MOCK_PROVIDER") == "true" {
		return p.mockResponse(req), nil
	}

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
	// Check for test mode
	if os.Getenv("PE_TEST_MODE") == "true" || os.Getenv("PE_MOCK_PROVIDER") == "true" {
		return p.mockStream(ctx, req), nil
	}

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
	// These are the commonly available ones based on latest cgpt v0.4.4
	return []string{
		"claude-sonnet-4-20250514", // Latest default model
		"gpt-4o",
		"gpt-4",
		"gpt-4-turbo",
		"gpt-3.5-turbo",
		"claude-3-7-sonnet-20250219",
		"claude-3-opus",
		"claude-3-sonnet",
		"claude-3-haiku",
		"gemini-2.0-flash",
		"gemini-pro",
	}, nil
}

// Close cleans up any resources.
func (p *Provider) Close() error {
	// Nothing to clean up for CLI-based provider
	return nil
}

// buildCommand creates the exec.Cmd for cgpt.
func (p *Provider) buildCommand(ctx context.Context, args ...string) *exec.Cmd {
	if p.useGoTool {
		// Use go tool cgpt (Go 1.24+ with tool directive)
		fullArgs := append([]string{"tool", "cgpt"}, args...)
		return exec.CommandContext(ctx, "go", fullArgs...)
	}
	if p.useGoRun {
		// Use go run as fallback
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
		args = append(args, "--model", req.Model)

		// Detect backend based on model name
		if strings.HasPrefix(req.Model, "gpt-") {
			args = append(args, "--backend", "openai")
		} else if strings.HasPrefix(req.Model, "claude-") {
			args = append(args, "--backend", "anthropic")
		} else if strings.HasPrefix(req.Model, "gemini-") {
			args = append(args, "--backend", "googleai")
		}
	}

	// Temperature (cgpt uses --temperature)
	if req.Temperature > 0 {
		args = append(args, "--temperature", fmt.Sprintf("%.2f", req.Temperature))
	}

	// Max tokens (cgpt uses --max-tokens)
	if req.MaxTokens > 0 {
		args = append(args, "--max-tokens", fmt.Sprintf("%d", req.MaxTokens))
	}

	// System prompt
	if req.SystemPrompt != "" {
		args = append(args, "--system-prompt", req.SystemPrompt)
	}

	// Prefill (cgpt uses --prefill for assistant message start)
	if req.Prefill != "" {
		args = append(args, "--prefill", req.Prefill)
	}

	// Stop sequences (cgpt uses --stop for stop sequences)
	for _, stop := range req.StopSequences {
		args = append(args, "--stop", stop)
	}

	// Input (use --input for direct string input)
	if req.Prompt != "" {
		args = append(args, "--input", req.Prompt)
	}

	// Additional options
	if req.Options != nil {
		// Handle any cgpt-specific options
		if v, ok := req.Options["json"].(bool); ok && v {
			args = append(args, "--json")
		}
		if v, ok := req.Options["verbose"].(bool); ok && v {
			args = append(args, "--verbose")
		}
		if v, ok := req.Options["debug"].(bool); ok && v {
			args = append(args, "--debug")
		}
		// Streaming is enabled by default in cgpt v0.4.4, but we can explicitly set it
		if v, ok := req.Options["stream"].(bool); ok {
			if v {
				args = append(args, "--stream")
			} else {
				args = append(args, "--stream=false")
			}
		}
		// Add completion timeout support
		if v, ok := req.Options["completion_timeout"].(string); ok && v != "" {
			args = append(args, "--completion-timeout", v)
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

// mockResponse generates a mock response for testing
func (p *Provider) mockResponse(req inference.Request) *inference.Response {
	// Default responses for common test prompts
	responses := map[string]string{
		"What is 2+2?":   "4",
		"'What is 2+2?'": "4",
		`"What is 2+2?"`: "4",
		"Explain what a pointer is in one sentence.": "A pointer is a variable that stores the memory address of another variable.",
		"Translate Hello to Spanish":                 "Hola",
		"Test prompt":                                "Test response",
		"Generate a random number":                   "42",
		"Tell me a story":                            "Once upon a time...",
		"Count to 5":                                 "1 2 3 4 5",
		"Current time?":                              "The current time is 3:00 PM",
		"You are a helpful assistant. Explain the concept of recursion.": "As a helpful assistant, I'll explain recursion: a function that calls itself to solve a problem by breaking it down into smaller instances.",
		"Write a Haskell function that calculates the factorial:":        "factorial :: Integer -> Integer\nfactorial 0 = 1\nfactorial n = n * factorial (n - 1)",
	}

	// Check if prompt ends with expected pattern for file reads
	content := "Mock response for: " + req.Prompt

	// Strip trailing newline for comparison
	promptCompare := strings.TrimSpace(req.Prompt)

	// Debug output for testing
	if os.Getenv("PE_DEBUG") == "true" {
		fmt.Fprintf(os.Stderr, "DEBUG: Mock received prompt: %q\n", promptCompare)
		fmt.Fprintf(os.Stderr, "DEBUG: Mock prefill: %q\n", req.Prefill)
		fmt.Fprintf(os.Stderr, "DEBUG: Mock stop sequences: %v\n", req.StopSequences)
	}

	// Handle prefill content by prepending it to the response
	var prefillPrefix string
	if req.Prefill != "" {
		prefillPrefix = req.Prefill
	}

	// Check for exact matches first
	if resp, ok := responses[promptCompare]; ok {
		content = resp
	} else if resp, ok := responses[req.Prompt]; ok {
		content = resp
	} else if strings.Contains(promptCompare, "Process this input") && strings.Contains(promptCompare, "{{") {
		content = "Processed input successfully"
	} else if strings.Contains(promptCompare, "helpful assistant") || strings.Contains(promptCompare, "recursion") {
		content = "As a helpful assistant, I'll explain recursion: a function that calls itself to solve a problem by breaking it down into smaller instances."
	} else if strings.Contains(req.Prompt, "{{") && strings.Contains(req.Prompt, "}}") {
		// Handle template processing - already done by pe run
		content = "Mock response for templated prompt"
	}

	// Apply prefill if provided
	if prefillPrefix != "" {
		content = prefillPrefix + content
	}

	// Apply stop sequences by truncating content at first occurrence
	for _, stop := range req.StopSequences {
		if idx := strings.Index(content, stop); idx != -1 {
			content = content[:idx]
			break
		}
	}

	// Handle JSON output request
	if req.Options != nil {
		if v, ok := req.Options["json"].(bool); ok && v {
			jsonResp := map[string]interface{}{
				"response": content,
				"model":    req.Model,
				"tokens":   5,
			}
			jsonBytes, _ := json.Marshal(jsonResp)
			content = string(jsonBytes)
		}
	}

	return &inference.Response{
		Content: content,
		Model:   req.Model,
		TokensUsed: inference.TokenUsage{
			TotalTokens: 5,
		},
		Metadata: map[string]interface{}{
			"provider": "cgpt",
			"mock":     true,
		},
	}
}

// mockStream generates a mock streaming response for testing
func (p *Provider) mockStream(ctx context.Context, req inference.Request) <-chan inference.StreamChunk {
	chunks := make(chan inference.StreamChunk)

	go func() {
		defer close(chunks)

		// Stream specific responses
		if strings.Contains(req.Prompt, "Count to 5") {
			for i := 1; i <= 5; i++ {
				select {
				case <-ctx.Done():
					chunks <- inference.StreamChunk{Error: ctx.Err()}
					return
				case chunks <- inference.StreamChunk{
					Delta: fmt.Sprintf("%d\n", i),
					Done:  false,
				}:
				}
			}
		} else {
			// Default streaming response
			resp := p.mockResponse(req)
			lines := strings.Split(resp.Content, "\n")
			for _, line := range lines {
				select {
				case <-ctx.Done():
					chunks <- inference.StreamChunk{Error: ctx.Err()}
					return
				case chunks <- inference.StreamChunk{
					Delta: line + "\n",
					Done:  false,
				}:
				}
			}
		}

		// Send done signal
		chunks <- inference.StreamChunk{Done: true}
	}()

	return chunks
}
