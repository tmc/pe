// Package optimizers contains optimization implementations that use the decoupled interfaces
package optimizers

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/tmc/pe/internal/optimization"
)

// TextGradOptimizer implements TextGrad 2.0 optimization with attention flow mapping
// Based on 2024-2025 research on natural language gradients
type TextGradOptimizer struct {
	provider     optimization.LanguageModelProvider
	evaluator    optimization.Evaluator
	enhancedMode bool
}

// NewTextGradOptimizer creates a new TextGrad optimizer
func NewTextGradOptimizer(provider optimization.LanguageModelProvider, evaluator optimization.Evaluator) optimization.GradientOptimizer {
	return &TextGradOptimizer{
		provider:     provider,
		evaluator:    evaluator,
		enhancedMode: true,
	}
}

// Name returns the optimizer identifier
func (o *TextGradOptimizer) Name() string {
	return "textgrad"
}

// SupportsObjective returns true if the optimizer supports the given objective type
func (o *TextGradOptimizer) SupportsObjective(objectiveType string) bool {
	// TextGrad supports most objective types
	supportedTypes := []string{"improve", "optimize", "refine", "enhance", "clarity", "effectiveness"}
	for _, supported := range supportedTypes {
		if strings.Contains(strings.ToLower(objectiveType), supported) {
			return true
		}
	}
	return true // Default to true for flexibility
}

// SupportsBatch returns true if the optimizer can optimize multiple prompts simultaneously
func (o *TextGradOptimizer) SupportsBatch() bool {
	return false // TextGrad focuses on single prompt optimization
}

// Optimize performs optimization using the given configuration
func (o *TextGradOptimizer) Optimize(ctx context.Context, config optimization.OptimizationConfig) (*optimization.OptimizationResult, error) {
	start := time.Now()

	result := &optimization.OptimizationResult{
		OriginalPrompt:  config.InitialPrompt,
		OptimizedPrompt: config.InitialPrompt,
		Iterations:      []optimization.IterationResult{},
		CreatedAt:       time.Now(),
		Method:          "textgrad",
	}

	currentPrompt := config.InitialPrompt

	maxIters := config.MaxIterations
	if maxIters == 0 {
		maxIters = config.Iterations
	}
	if maxIters == 0 {
		maxIters = 5 // default
	}

	objective := config.Objective
	if objective == "" {
		objective = "improve clarity, specificity, and effectiveness"
	}

	convergenceThreshold := config.ConvergenceThreshold
	if convergenceThreshold == 0 {
		convergenceThreshold = 0.001 // default
	}

	for i := 0; i < maxIters; i++ {
		iterStart := time.Now()

		// Compute textual gradients
		gradients, err := o.ComputeGradients(ctx, currentPrompt, objective)
		if err != nil {
			return nil, fmt.Errorf("computing gradients at iteration %d: %w", i+1, err)
		}

		// Apply gradients to get improved prompt
		improvedPrompt, err := o.ApplyGradients(ctx, currentPrompt, gradients)
		if err != nil {
			return nil, fmt.Errorf("applying gradients at iteration %d: %w", i+1, err)
		}

		// Evaluate improvement
		score, err := o.evaluator.EvaluatePrompt(ctx, improvedPrompt, objective)
		if err != nil {
			return nil, fmt.Errorf("evaluating prompt at iteration %d: %w", i+1, err)
		}

		iteration := optimization.IterationResult{
			Iteration: i + 1,
			Prompt:    improvedPrompt,
			Score:     score,
			Changes:   o.extractChanges(currentPrompt, improvedPrompt),
			Timestamp: time.Now(),
			Duration:  time.Since(iterStart),
			Gradients: gradients,
		}
		result.Iterations = append(result.Iterations, iteration)

		// Check convergence
		if i > 0 && math.Abs(score-result.Iterations[i-1].Score) < convergenceThreshold {
			result.Converged = true
			break
		}

		currentPrompt = improvedPrompt
	}

	result.OptimizedPrompt = currentPrompt
	if len(result.Iterations) > 0 {
		result.ImprovementScore = result.Iterations[len(result.Iterations)-1].Score
	}
	result.TotalDuration = time.Since(start)

	return result, nil
}

// ComputeGradients computes gradients for the given input
func (o *TextGradOptimizer) ComputeGradients(ctx context.Context, input string, objective string) ([]optimization.Gradient, error) {
	gradientPrompt := fmt.Sprintf(`Analyze this prompt and provide textual gradients for improvement:

PROMPT TO ANALYZE:
%s

OBJECTIVE:
%s

Provide feedback in JSON format with an array of gradients. Each gradient should identify a specific component of the prompt that can be improved, provide actionable feedback, and suggest concrete improvements.

Required JSON format:
{
  "gradients": [
    {
      "component": "specific part of the prompt to improve",
      "feedback": "what's wrong or could be improved",
      "suggestions": ["specific suggestion 1", "specific suggestion 2"],
      "confidence": 0.8,
      "priority": 0.9,
      "direction": "improve/clarify/specify/restructure/etc",
      "magnitude": 0.7
    }
  ]
}

Focus on:
1. Clarity and precision of instructions
2. Specificity and completeness
3. Structure and organization
4. Potential ambiguities or edge cases
5. Alignment with the objective`, input, objective)

	options := optimization.GenerationOptions{
		Temperature: &[]float64{0.2}[0], // Low temperature for consistent analysis
		MaxTokens:   &[]int{1000}[0],
	}

	response, err := o.provider.Generate(ctx, gradientPrompt, options)
	if err != nil {
		return nil, fmt.Errorf("failed to generate gradients: %w", err)
	}

	// Parse gradients from JSON response
	gradients, err := o.parseGradients(response.Text)
	if err != nil {
		// Fallback: create a basic gradient if parsing fails
		return []optimization.Gradient{{
			Component:   "overall structure",
			Feedback:    "Prompt needs optimization for better clarity and effectiveness",
			Suggestions: []string{"Add more specific instructions", "Include examples or context"},
			Confidence:  0.6,
			Priority:    0.7,
			Direction:   "improve",
			Magnitude:   0.5,
		}}, nil
	}

	return gradients, nil
}

// ApplyGradients applies gradients to optimize the input
func (o *TextGradOptimizer) ApplyGradients(ctx context.Context, input string, gradients []optimization.Gradient) (string, error) {
	if len(gradients) == 0 {
		return input, nil
	}

	// Sort gradients by priority (highest first)
	sortedGradients := make([]optimization.Gradient, len(gradients))
	copy(sortedGradients, gradients)

	// Simple bubble sort by priority
	for i := 0; i < len(sortedGradients)-1; i++ {
		for j := 0; j < len(sortedGradients)-i-1; j++ {
			if sortedGradients[j].Priority < sortedGradients[j+1].Priority {
				sortedGradients[j], sortedGradients[j+1] = sortedGradients[j+1], sortedGradients[j]
			}
		}
	}

	// Use top 3 gradients to avoid overwhelming the model
	topGradients := sortedGradients
	if len(topGradients) > 3 {
		topGradients = topGradients[:3]
	}

	// Format gradients for the improvement prompt
	var gradientDescriptions []string
	for _, g := range topGradients {
		suggestions := strings.Join(g.Suggestions, ", ")
		description := fmt.Sprintf("- %s: %s (suggestions: %s)", g.Component, g.Feedback, suggestions)
		gradientDescriptions = append(gradientDescriptions, description)
	}

	improvePrompt := fmt.Sprintf(`Improve this prompt by applying the following textual gradients:

ORIGINAL PROMPT:
%s

GRADIENTS TO APPLY:
%s

Instructions:
1. Apply the suggested improvements while maintaining the core intent
2. Ensure the result is a complete, coherent prompt
3. Integrate changes smoothly without making the prompt overly complex
4. Preserve any essential context or constraints from the original

Provide only the improved prompt without any explanation or commentary.`,
		input, strings.Join(gradientDescriptions, "\n"))

	options := optimization.GenerationOptions{
		Temperature: &[]float64{0.3}[0], // Moderate temperature for creativity while staying focused
		MaxTokens:   &[]int{500}[0],
	}

	response, err := o.provider.Generate(ctx, improvePrompt, options)
	if err != nil {
		return "", fmt.Errorf("failed to apply gradients: %w", err)
	}

	return strings.TrimSpace(response.Text), nil
}

// parseGradients parses gradients from LLM response JSON
func (o *TextGradOptimizer) parseGradients(responseText string) ([]optimization.Gradient, error) {
	// Find JSON in the response
	start := strings.Index(responseText, "{")
	end := strings.LastIndex(responseText, "}")
	if start == -1 || end == -1 || start > end {
		return nil, fmt.Errorf("no JSON found in response")
	}

	jsonStr := responseText[start : end+1]

	var result struct {
		Gradients []optimization.Gradient `json:"gradients"`
	}

	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil, fmt.Errorf("failed to parse gradients JSON: %w", err)
	}

	// Validate and normalize gradients
	for i := range result.Gradients {
		g := &result.Gradients[i]

		// Normalize magnitude and confidence to [0, 1]
		if g.Magnitude < 0 {
			g.Magnitude = 0
		} else if g.Magnitude > 1 {
			g.Magnitude = 1
		}

		if g.Confidence < 0 {
			g.Confidence = 0
		} else if g.Confidence > 1 {
			g.Confidence = 1
		}

		if g.Priority < 0 {
			g.Priority = 0
		} else if g.Priority > 1 {
			g.Priority = 1
		}

		// Ensure required fields have defaults
		if g.Component == "" {
			g.Component = "general improvement"
		}
		if g.Direction == "" {
			g.Direction = "improve"
		}
		if len(g.Suggestions) == 0 {
			g.Suggestions = []string{"enhance clarity and effectiveness"}
		}
	}

	return result.Gradients, nil
}

// extractChanges identifies changes between prompts
func (o *TextGradOptimizer) extractChanges(original, improved string) []string {
	changes := []string{}

	// Basic change detection
	if len(improved) != len(original) {
		if len(improved) > len(original) {
			changes = append(changes, fmt.Sprintf("Expanded content (+%d characters)", len(improved)-len(original)))
		} else {
			changes = append(changes, fmt.Sprintf("Condensed content (-%d characters)", len(original)-len(improved)))
		}
	}

	// Detect common improvements
	if strings.Contains(improved, "step-by-step") && !strings.Contains(original, "step-by-step") {
		changes = append(changes, "Added step-by-step instruction structure")
	}

	if strings.Contains(improved, "specific") && !strings.Contains(original, "specific") {
		changes = append(changes, "Enhanced specificity")
	}

	if strings.Count(improved, "\n") > strings.Count(original, "\n") {
		changes = append(changes, "Improved formatting and structure")
	}

	if strings.Contains(improved, "example") && !strings.Contains(original, "example") {
		changes = append(changes, "Added examples or context")
	}

	// If no specific changes detected, provide general change description
	if len(changes) == 0 {
		changes = append(changes, "Applied textual gradients for optimization")
	}

	return changes
}
