// Package cgpt provides utilities for interacting with LLM providers
// through a command-line interface.
package cgpt

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/tmc/pe/internal/promptfoo"
)

// ModelProvider represents a model provider configuration
type ModelProvider struct {
	Model       string
	MaxTokens   int
	Temperature float64
	Backend     string
}

// DefaultProvider returns a default configured model provider
func DefaultProvider() *ModelProvider {
	return &ModelProvider{
		Model:       "gemini-2.0-flash",
		MaxTokens:   1024,
		Temperature: 0.2,
		Backend:     "googleai",
	}
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
	
	// Get provider string from vars
	if provider, ok := vars["provider"].(string); ok && provider \!= "" {
		// Extract backend from provider if specified (e.g., "anthropic:claude-3" -> "anthropic")
		parts := strings.Split(provider, ":")
		if len(parts) > 0 {
			p.Backend = parts[0]
		}
		
		// Extract model if specified
		if len(parts) > 1 {
			p.Model = parts[1]
		}
	}
	
	// Run cgpt command
	output, tokens, err := p.runCGPTCommand(processedPrompt, dryRun)
	if err \!= nil {
		return nil, err
	}
	
	// Calculate latency
	latency := time.Since(startTime)
	
	// Estimate costs (approximate)
	cost := estimateCost(tokens)
	
	// Create token usage details
	tokenUsage := &promptfoo.TokenUsage{
		Total:      int32(tokens.total),
		Prompt:     int32(tokens.prompt),
		Completion: int32(tokens.completion),
		Cached:     0,
		NumRequests: 1,
		Details:    &promptfoo.CompletionTokenDetails{
			Reasoning:          0,
			AcceptedPrediction: int32(tokens.completion),
			RejectedPrediction: 0,
		},
	}
	
	// Build and return response
	return &promptfoo.ProviderResponse{
		Output:     output,
		TokenUsage: tokenUsage,
		FinishTime: startTime.Add(latency),
		Cost:       cost,
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
	// Build the cgpt command with the appropriate parameters
	tempArg := fmt.Sprintf("%.1f", p.Temperature)
	maxTokensArg := fmt.Sprintf("%d", p.MaxTokens)
	
	// Build the cgpt command base
	args := []string{
		"-b", p.Backend,
		"-m", p.Model,
	}
	
	// Only add these flags if not in dry run mode
	if \!dryRun {
		args = append(args, "--temperature", tempArg, "--max-tokens", maxTokensArg)
	}
	
	// Add the prompt as the final argument
	args = append(args, prompt)
	
	cmd := exec.Command("cgpt", args...)
	
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
	
	if err := cmd.Run(); err \!= nil {
		// Return a formatted error message that includes the model information for better debugging
		errMsg := fmt.Sprintf("Backend: %s, Model: %s - Error: %v\n%s", 
			p.Backend, p.Model, err, stderr.String())
		return "", tokenCounts{}, fmt.Errorf("error running cgpt command: %s", errMsg)
	}
	
	// The output is not JSON, just plain text
	output := stdout.String()

	// Estimate token counts
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
// Using approximate pricing for API models
func estimateCost(tokens tokenCounts) float64 {
	// Approximate cost per 1K tokens
	const promptCostPer1K = 0.0005
	const completionCostPer1K = 0.0015
	
	promptCost := float64(tokens.prompt) * promptCostPer1K / 1000
	completionCost := float64(tokens.completion) * completionCostPer1K / 1000
	
	return promptCost + completionCost
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
