// Package cgpt provides utilities for interacting with the Google Cloud Platform's
// Vertex AI API for language model completions.
package cgpt

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/tmc/pe/internal/promptfoo"
)

// Security constants for input validation
var (
	// allowedBackends defines the allowed backend values to prevent command injection
	allowedBackends = map[string]bool{
		"openai":      true,
		"anthropic":   true,
		"googleai":    true,
		"ollama":      true,
		"bedrock":     true,
		"azure":       true,
		"cohere":      true,
		"huggingface": true,
	}

	// allowedModelPrefixes defines allowed model name prefixes
	allowedModelPrefixes = []string{
		"gpt-3.5", "gpt-4", "gpt-4o", "gpt-4-turbo",
		"claude-3", "claude-3.5", "claude-2", "claude-instant", "claude-sonnet-4",
		"gemini-", "gemini-pro", "gemini-flash", "gemini-2",
		"text-bison", "text-unicorn", "chat-bison",
		"command", "command-light", "command-r",
		"llama", "mistral", "mixtral", "qwen",
		"phi-", "orca-", "vicuna-", "alpaca-",
	}

	// shellMetacharRegex matches dangerous shell metacharacters
	shellMetacharRegex = regexp.MustCompile(`[;&|<>$` + "`" + `(){}[\]\\*?~]`)
)

// ModelProvider represents a Vertex AI model provider configuration
type ModelProvider struct {
	Model       string
	MaxTokens   int
	Temperature float64
	Backend     string // Added backend parameter
}

// DefaultProvider returns a default configured model provider
func DefaultProvider() *ModelProvider {
	return &ModelProvider{
		Model:       "claude-sonnet-4-20250514", // Default to latest Anthropic model (matches cgpt v0.4.4 default)
		MaxTokens:   1024,
		Temperature: 0.2,
		Backend:     "anthropic", // Default to Anthropic backend (matches cgpt v0.4.4 default)
	}
}

// Security validation functions

// validateBackend validates that the backend is in the allowed list
func validateBackend(backend string) error {
	if backend == "" {
		return fmt.Errorf("backend cannot be empty")
	}
	if !allowedBackends[backend] {
		return fmt.Errorf("backend '%s' is not allowed; allowed backends: %v", backend, getAllowedBackends())
	}
	return nil
}

// validateModel validates that the model name is safe and follows expected patterns
func validateModel(model string) error {
	if model == "" {
		return fmt.Errorf("model cannot be empty")
	}

	// Check for shell metacharacters
	if shellMetacharRegex.MatchString(model) {
		return fmt.Errorf("model name contains forbidden characters: %s", model)
	}

	// Check against allowed prefixes
	for _, prefix := range allowedModelPrefixes {
		if strings.HasPrefix(model, prefix) {
			return nil
		}
	}

	// Allow models that contain only alphanumeric characters, hyphens, dots, and underscores
	modelRegex := regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)
	if !modelRegex.MatchString(model) {
		return fmt.Errorf("model name contains invalid characters: %s", model)
	}

	return nil
}

// sanitizePrompt removes or escapes dangerous characters from the prompt
func sanitizePrompt(prompt string) string {
	// Remove null bytes which can cause command injection
	prompt = strings.ReplaceAll(prompt, "\x00", "")

	// For safety, we don't allow prompts that look like command injection attempts
	// This is a conservative approach that maintains functionality while preventing attacks
	dangerous := []string{
		"$(", "`", "${", "&&", "||", ";", "|", "<", ">", "&",
	}

	for _, danger := range dangerous {
		if strings.Contains(prompt, danger) {
			// Replace with safe alternatives or remove
			prompt = strings.ReplaceAll(prompt, danger, "")
		}
	}

	return prompt
}

// getAllowedBackends returns a sorted list of allowed backends for error messages
func getAllowedBackends() []string {
	backends := make([]string, 0, len(allowedBackends))
	for backend := range allowedBackends {
		backends = append(backends, backend)
	}
	return backends
}

// EvaluatePrompt takes a prompt and optional parameters and returns a completion from the model
func (p *ModelProvider) EvaluatePrompt(prompt string, vars map[string]interface{}) (*promptfoo.ProviderResponse, error) {
	return p.EvaluatePromptWithOptions(prompt, vars, false)
}

// EvaluatePromptWithOptions takes a prompt and optional parameters and returns a completion from the model
// with additional options like dry-run mode
func (p *ModelProvider) EvaluatePromptWithOptions(prompt string, vars map[string]interface{}, dryRun bool) (*promptfoo.ProviderResponse, error) {
	// Replace any template variables in the prompt
	processedPrompt := replaceVariables(prompt, vars)

	// Start timing
	startTime := time.Now()

	// Apply configuration from vars
	p.ApplyConfigFromVars(vars)

	// Run cgpt command
	output, tokens, err := p.runCGPTCommand(processedPrompt, dryRun)
	if err != nil {
		return nil, err
	}

	// Calculate latency (unused in current implementation)
	_ = time.Since(startTime)

	// Estimate costs (approximate)
	cost := estimateCost(tokens)

	// Create token usage details
	tokenUsage := &promptfoo.TokenUsage{
		Total:      int32(tokens.total),
		Prompt:     int32(tokens.prompt),
		Completion: int32(tokens.completion),
		Cached:     0,
	}

	// Build and return response
	return &promptfoo.ProviderResponse{
		Output:     output,
		TokenUsage: tokenUsage,
		Cost:       cost,
		Cached:     false,
	}, nil
}

// tokenCounts holds the token usage information
type tokenCounts struct {
	prompt     int
	completion int
	total      int
}

// runCGPTCommand executes the cgpt command to query the model
func (p *ModelProvider) runCGPTCommand(prompt string, dryRun bool) (string, tokenCounts, error) {
	// SECURITY: Validate all inputs before passing to exec.Command
	if err := validateBackend(p.Backend); err != nil {
		return "", tokenCounts{}, fmt.Errorf("backend validation failed: %w", err)
	}

	if err := validateModel(p.Model); err != nil {
		return "", tokenCounts{}, fmt.Errorf("model validation failed: %w", err)
	}

	// SECURITY: Sanitize the prompt to prevent command injection
	sanitizedPrompt := sanitizePrompt(prompt)

	// Build the cgpt command with the appropriate parameters
	tempArg := fmt.Sprintf("%.1f", p.Temperature)
	maxTokensArg := fmt.Sprintf("%d", p.MaxTokens)

	// Build the cgpt command as described: cgpt --backend googleai --model gemini-2.0-flash [prompt]
	// All inputs are now validated and sanitized
	args := []string{
		"--backend", p.Backend, // Validated against allowlist
		"--model", p.Model, // Validated against patterns and sanitized
	}

	// Only add these flags if not in dry run mode
	if !dryRun {
		args = append(args, "--temperature", tempArg, "--max-tokens", maxTokensArg)
	}

	// Add the sanitized prompt as the final argument
	args = append(args, sanitizedPrompt)

	// Try to use go tool cgpt first (Go 1.24+ with tool directive)
	var cmd *exec.Cmd
	if testCmd := exec.Command("go", "tool", "cgpt", "--version"); testCmd.Run() == nil {
		// Use go tool cgpt
		fullArgs := append([]string{"tool", "cgpt"}, args...)
		cmd = exec.Command("go", fullArgs...)
	} else if _, err := exec.LookPath("cgpt"); err == nil {
		// Use cgpt binary if available
		cmd = exec.Command("cgpt", args...)
	} else {
		// Fall back to go run
		fullArgs := append([]string{"run", "github.com/tmc/cgpt/cmd/cgpt"}, args...)
		cmd = exec.Command("go", fullArgs...)
	}

	// If dry run, just print the command without executing
	if dryRun {
		// Return a mock response for dry run mode
		return "Dry run - no actual execution",
			tokenCounts{
				prompt:     estimateTokenCount(prompt),
				completion: estimateTokenCount("This is a dry run response."),
				total:      estimateTokenCount(prompt) + estimateTokenCount("This is a dry run response."),
			}, nil
	}

	// Execute the command
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		// Return a formatted error message that includes the model information for better debugging
		errMsg := fmt.Sprintf("Backend: %s, Model: %s - Error: %v\n%s",
			p.Backend, p.Model, err, stderr.String())
		return "", tokenCounts{}, fmt.Errorf("error running cgpt command: %s", errMsg)
	}

	// The output is not JSON, just plain text
	output := stdout.String()

	// Estimate token counts using original prompt length for accuracy
	promptTokens := estimateTokenCount(prompt)
	completionTokens := estimateTokenCount(output)

	return output, tokenCounts{
		prompt:     promptTokens,
		completion: completionTokens,
		total:      promptTokens + completionTokens,
	}, nil
}

// estimateTokenCount provides a rough estimate of token count from text
// This is a simplistic approximation, in a real implementation you would use a proper tokenizer
func estimateTokenCount(text string) int {
	words := strings.Fields(text)
	return len(words) * 4 / 3 // Rough estimate: 4 tokens per 3 words
}

// estimateCost calculates an approximate cost based on token usage
// Using approximate pricing for Google's models
func estimateCost(tokens tokenCounts) float64 {
	// Approximate cost per 1K tokens
	const promptCostPer1K = 0.0005
	const completionCostPer1K = 0.0015

	promptCost := float64(tokens.prompt) * promptCostPer1K / 1000
	completionCost := float64(tokens.completion) * completionCostPer1K / 1000

	return promptCost + completionCost
}

// ApplyConfigFromVars applies configuration settings from a variables map
// This handles both promptfoo-style provider strings (e.g., "openai:gpt-4")
// and explicit configuration options in the config section
func (p *ModelProvider) ApplyConfigFromVars(vars map[string]interface{}) {
	// Get provider string from vars
	if provider, ok := vars["provider"].(string); ok && provider != "" {
		// Extract backend from provider if specified (e.g., "anthropic:claude-3" -> "anthropic")
		parts := strings.Split(provider, ":")
		if len(parts) > 0 {
			// SECURITY: Validate backend before setting
			if err := validateBackend(parts[0]); err == nil {
				p.Backend = parts[0]
			}
		}

		// Extract model if specified
		if len(parts) > 1 {
			// SECURITY: Validate model before setting
			if err := validateModel(parts[1]); err == nil {
				p.Model = parts[1]
			}
		}
	}

	// Check for direct vars first
	if model, ok := vars["model"].(string); ok {
		// SECURITY: Validate model before setting
		if err := validateModel(model); err == nil {
			p.Model = model
		}
	}

	if temp, ok := vars["temperature"].(float64); ok {
		// Validate temperature range
		if temp >= 0 && temp <= 2.0 {
			p.Temperature = temp
		}
	}

	if maxTokens, ok := vars["max_tokens"].(int); ok {
		// Validate max tokens range
		if maxTokens > 0 && maxTokens <= 4096 {
			p.MaxTokens = maxTokens
		}
	} else if maxTokens, ok := vars["max_tokens"].(float64); ok {
		maxTokensInt := int(maxTokens)
		if maxTokensInt > 0 && maxTokensInt <= 4096 {
			p.MaxTokens = maxTokensInt
		}
	}

	if backend, ok := vars["backend"].(string); ok {
		// SECURITY: Validate backend before setting
		if err := validateBackend(backend); err == nil {
			p.Backend = backend
		}
	}

	// Check for config map and apply settings
	if configMap, ok := vars["config"].(map[string]interface{}); ok {
		// Apply temperature if specified
		if temp, ok := configMap["temperature"].(float64); ok {
			// Validate temperature range
			if temp >= 0 && temp <= 2.0 {
				p.Temperature = temp
			}
		}

		// Apply max_tokens if specified
		if maxTokens, ok := configMap["max_tokens"].(int); ok {
			// Validate max tokens range
			if maxTokens > 0 && maxTokens <= 4096 {
				p.MaxTokens = maxTokens
			}
		} else if maxTokens, ok := configMap["max_tokens"].(float64); ok {
			maxTokensInt := int(maxTokens)
			if maxTokensInt > 0 && maxTokensInt <= 4096 {
				p.MaxTokens = maxTokensInt
			}
		}

		// Apply backend if specified
		if backend, ok := configMap["backend"].(string); ok {
			// SECURITY: Validate backend before setting
			if err := validateBackend(backend); err == nil {
				p.Backend = backend
			}
		}

		// Apply model if specified directly in config
		if model, ok := configMap["model"].(string); ok {
			// SECURITY: Validate model before setting
			if err := validateModel(model); err == nil {
				p.Model = model
			}
		}
	}
}

// replaceVariables substitutes template variables in the prompt with actual values
func replaceVariables(prompt string, vars map[string]interface{}) string {
	result := prompt
	for key, value := range vars {
		var strValue string
		switch v := value.(type) {
		case string:
			strValue = v
		case float64:
			strValue = fmt.Sprintf("%g", v)
		case int:
			strValue = fmt.Sprintf("%d", v)
		case bool:
			strValue = fmt.Sprintf("%t", v)
		default:
			jsonValue, err := json.Marshal(v)
			if err == nil {
				strValue = string(jsonValue)
			} else {
				strValue = fmt.Sprintf("%v", v)
			}
		}

		placeholder := fmt.Sprintf("{{%s}}", key)
		result = strings.ReplaceAll(result, placeholder, strValue)
	}

	return result
}

// // Execute runs the specified backend and model on the given prompt and returns the result
// func Execute(backend, model, prompt, systemPrompt string) (string, error) {
// 	// Build the command arguments
// 	args := []string{"-b", backend, "-m", model}

// 	// Add system prompt if provided
// 	if systemPrompt != "" {
// 		args = append(args, "-s", systemPrompt)
// 	}

// 	// Add temperature (low temperature for more consistent results)
// 	args = append(args, "-T", "0.1")

// 	// Add the prompt as input
// 	args = append(args, "-i", prompt)

// 	// Execute the cgpt command
// 	cmd := exec.Command("cgpt", args...)

// 	// Capture output
// 	var stdout, stderr bytes.Buffer
// 	cmd.Stdout = &stdout
// 	cmd.Stderr = &stderr

// 	if err := cmd.Run(); err != nil {
// 		// Return a formatted error message that includes stderr for better debugging
// 		errMsg := fmt.Sprintf("CGPT Error (Backend: %s, Model: %s): %v\n%s",
// 			backend, model, err, stderr.String())
// 		return "", fmt.Errorf("%s", errMsg)
// 	}

// 	// Return the output
// 	return stdout.String(), nil
// }

// formatDuration formats a duration in a human-readable way
func formatDuration(d time.Duration) string {
	if d < time.Millisecond {
		return "0ms"
	} else if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	} else if d < time.Minute {
		return fmt.Sprintf("%.2fs", d.Seconds())
	}
	return fmt.Sprintf("%.2fm", d.Minutes())
}

// isKnownProvider checks if a model string corresponds to a known provider
func isKnownProvider(model string) bool {
	knownPrefixes := []string{
		"gpt-4", "gpt-3.5", "claude-", "gemini-", "text-bison",
	}

	for _, prefix := range knownPrefixes {
		if strings.HasPrefix(model, prefix) {
			return true
		}
	}
	return false
}

// parseProviderString parses a provider string in the format "backend:model:key=value:..."
func parseProviderString(provider string) (backend, model string, temperature float64, maxTokens int) {
	if provider == "" {
		return
	}

	parts := strings.Split(provider, ":")
	if len(parts) >= 1 {
		backend = parts[0]
	}
	if len(parts) >= 2 {
		model = parts[1]
	}

	// Parse additional parameters
	for i := 2; i < len(parts); i++ {
		kv := strings.SplitN(parts[i], "=", 2)
		if len(kv) == 2 {
			switch kv[0] {
			case "temperature":
				if t, err := parseFloat(kv[1]); err == nil {
					temperature = t
				}
			case "max_tokens":
				if m, err := parseInt(kv[1]); err == nil {
					maxTokens = m
				}
			}
		}
	}

	return
}

// Helper functions for parsing
func parseFloat(s string) (float64, error) {
	var f float64
	_, err := fmt.Sscanf(s, "%f", &f)
	return f, err
}

func parseInt(s string) (int, error) {
	var i int
	_, err := fmt.Sscanf(s, "%d", &i)
	return i, err
}
