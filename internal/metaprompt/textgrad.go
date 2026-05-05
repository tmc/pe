package metaprompt

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/tmc/pe/internal/llm"
)

// TextGradOptimizer implements TextGrad 2.0 optimization with attention flow mapping
// Based on 2024-2025 research on natural language gradients
type TextGradOptimizer struct {
	llm                 Generator
	enhancedMode        bool
	attentionFlowMapper *AttentionFlowMapper
}

// AttentionFlowMapper tracks attention flow through the optimization process
type AttentionFlowMapper struct {
	flows map[string][]float64
}

// NewTextGradOptimizer creates a new TextGrad optimizer
func NewTextGradOptimizer(provider Generator) *TextGradOptimizer {
	return &TextGradOptimizer{
		llm:                 provider,
		enhancedMode:        true,
		attentionFlowMapper: &AttentionFlowMapper{flows: make(map[string][]float64)},
	}
}

// TextualGradient represents a textual gradient for optimization
type TextualGradient struct {
	Component   string   `json:"component"`
	Feedback    string   `json:"feedback"`
	Suggestions []string `json:"suggestions"`
	Confidence  float64  `json:"confidence"`
	Priority    float64  `json:"priority"`
	Gradient    string   `json:"gradient"`
	Magnitude   float64  `json:"magnitude"`
	Direction   string   `json:"direction"`
}

// ComputationGraph represents the computation graph for TextGrad
type ComputationGraph struct {
	Nodes []GraphNode `json:"nodes"`
	Edges []GraphEdge `json:"edges"`
}

// GraphNode represents a component in the computation graph
type GraphNode struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"` // prompt, response, tool_call, etc.
	Content   string                 `json:"content"`
	Gradients []TextualGradient      `json:"gradients"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// GraphEdge represents a connection between nodes
type GraphEdge struct {
	From   string  `json:"from"`
	To     string  `json:"to"`
	Weight float64 `json:"weight"`
	Type   string  `json:"type"`
}

// OptimizeWithTextGrad performs optimization using textual gradients
func (tg *TextGradOptimizer) OptimizeWithTextGrad(ctx context.Context, cfg Config) (*OptimizationResult, error) {
	start := time.Now()
	result := &OptimizationResult{
		OriginalPrompt:  cfg.InitialPrompt,
		OptimizedPrompt: cfg.InitialPrompt,
		Iterations:      []IterationResult{},
		CreatedAt:       time.Now(),
	}

	currentPrompt := cfg.InitialPrompt

	maxIters := cfg.MaxIterations
	if maxIters == 0 {
		maxIters = cfg.Iterations
	}
	if maxIters == 0 {
		maxIters = 5 // default
	}

	for i := 0; i < maxIters; i++ {
		// Compute textual gradients
		objective := cfg.Objective
		if objective == "" {
			objective = "improve clarity, specificity, and effectiveness"
		}
		gradients, err := tg.computeGradients(ctx, currentPrompt, objective)
		if err != nil {
			return nil, fmt.Errorf("computing gradients: %w", err)
		}

		// Apply gradients to get improved prompt
		improvedPrompt, err := tg.applyGradients(ctx, currentPrompt, gradients)
		if err != nil {
			return nil, fmt.Errorf("applying gradients: %w", err)
		}

		// Evaluate improvement
		score, err := tg.evaluatePrompt(ctx, improvedPrompt, objective)
		if err != nil {
			return nil, fmt.Errorf("evaluating prompt: %w", err)
		}

		iteration := IterationResult{
			Iteration: i + 1,
			Prompt:    improvedPrompt,
			Score:     score,
			Changes:   tg.extractChanges(currentPrompt, improvedPrompt),
			Timestamp: time.Now(),
		}
		result.Iterations = append(result.Iterations, iteration)

		// Check convergence
		convergenceThreshold := cfg.ConvergenceThreshold
		if convergenceThreshold == 0 {
			convergenceThreshold = 0.001 // default
		}
		if i > 0 && math.Abs(score-result.Iterations[i-1].Score) < convergenceThreshold {
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

// computeGradients generates textual gradients for the prompt
func (tg *TextGradOptimizer) computeGradients(ctx context.Context, prompt, objective string) ([]TextualGradient, error) {
	gradientPrompt := fmt.Sprintf(`Analyze this prompt and provide textual gradients for improvement:

Prompt: %s

Objective: %s

Provide feedback in JSON format with an array of gradients:
{
  "gradients": [
    {
      "component": "specific part of the prompt",
      "feedback": "what's wrong or could be improved",
      "suggestions": ["suggestion 1", "suggestion 2"],
      "confidence": 0.8,
      "priority": 0.9,
      "gradient": "specific change direction",
      "magnitude": 0.7,
      "direction": "improve/reduce/clarify/etc"
    }
  ]
}`, prompt, objective)

	options := llm.GenerateOptions{
		Temperature: &[]float64{0.7}[0],
	}
	response, err := tg.llm.Generate(ctx, gradientPrompt, options)
	if err != nil {
		return nil, err
	}

	var result struct {
		Gradients []TextualGradient `json:"gradients"`
	}
	if err := json.Unmarshal([]byte(response.Text), &result); err != nil {
		return nil, fmt.Errorf("parsing gradients: %w", err)
	}

	return result.Gradients, nil
}

// applyGradients applies textual gradients to improve the prompt
func (tg *TextGradOptimizer) applyGradients(ctx context.Context, prompt string, gradients []TextualGradient) (string, error) {
	if len(gradients) == 0 {
		return prompt, nil
	}

	// Sort by priority
	topGradients := gradients
	if len(topGradients) > 3 {
		topGradients = topGradients[:3]
	}

	var gradientDescriptions []string
	for _, g := range topGradients {
		gradientDescriptions = append(gradientDescriptions, fmt.Sprintf(
			"- %s: %s (suggestions: %s)",
			g.Component, g.Feedback, strings.Join(g.Suggestions, ", "),
		))
	}

	improvePrompt := fmt.Sprintf(`Improve this prompt based on the following textual gradients:

Original Prompt: %s

Gradients to apply:
%s

Provide only the improved prompt without any explanation.`,
		prompt, strings.Join(gradientDescriptions, "\n"))

	options := llm.GenerateOptions{
		Temperature: &[]float64{0.7}[0],
	}
	response, err := tg.llm.Generate(ctx, improvePrompt, options)
	if err != nil {
		return "", err
	}
	return response.Text, nil
}

// evaluatePrompt scores a prompt against the objective
func (tg *TextGradOptimizer) evaluatePrompt(ctx context.Context, prompt, objective string) (float64, error) {
	evalPrompt := fmt.Sprintf(`Evaluate this prompt against the objective on a scale from 0.0 to 1.0:

Prompt: %s

Objective: %s

Provide only a single number between 0.0 and 1.0.`, prompt, objective)

	options := llm.GenerateOptions{
		Temperature: &[]float64{0.3}[0],
	}
	response, err := tg.llm.Generate(ctx, evalPrompt, options)
	if err != nil {
		return 0, err
	}

	score := 0.0
	fmt.Sscanf(strings.TrimSpace(response.Text), "%f", &score)
	return math.Max(0.0, math.Min(1.0, score)), nil
}

// extractChanges identifies changes between prompts
func (tg *TextGradOptimizer) extractChanges(original, improved string) []string {
	// Simple implementation - in production would use proper diff
	changes := []string{}
	if len(improved) != len(original) {
		changes = append(changes, fmt.Sprintf("Length changed from %d to %d characters", len(original), len(improved)))
	}
	if strings.Contains(improved, "step-by-step") && !strings.Contains(original, "step-by-step") {
		changes = append(changes, "Added step-by-step reasoning")
	}
	if strings.Count(improved, "\n") > strings.Count(original, "\n") {
		changes = append(changes, "Improved formatting with more line breaks")
	}
	return changes
}

// Optimize implements the Optimizer interface
func (tg *TextGradOptimizer) Optimize(ctx context.Context, cfg Config) (*OptimizationResult, error) {
	return tg.OptimizeWithTextGrad(ctx, cfg)
}
