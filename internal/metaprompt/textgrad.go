package metaprompt

import (
	"context"
	"fmt"
	"strings"

	"github.com/tmc/pe/internal/llm"
)

// TextGradOptimizer implements TextGrad-style optimization using textual gradients
type TextGradOptimizer struct {
	llm llm.Provider
}

// NewTextGradOptimizer creates a new TextGrad-style optimizer
func NewTextGradOptimizer(llmProvider llm.Provider) *TextGradOptimizer {
	return &TextGradOptimizer{
		llm: llmProvider,
	}
}

// TextualGradient represents feedback in natural language form
type TextualGradient struct {
	Component   string  `json:"component"`    // Which part of the prompt
	Gradient    string  `json:"gradient"`     // Natural language feedback
	Magnitude   float64 `json:"magnitude"`    // How strong the gradient is
	Direction   string  `json:"direction"`    // Positive or negative feedback
	Suggestions []string `json:"suggestions"` // Specific improvement suggestions
}

// ComputationGraph represents the prompt optimization computation graph
type ComputationGraph struct {
	Nodes []GraphNode `json:"nodes"`
	Edges []GraphEdge `json:"edges"`
}

// GraphNode represents a component in the computation graph
type GraphNode struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"` // prompt, response, tool_call, etc.
	Content     string                 `json:"content"`
	Gradients   []TextualGradient      `json:"gradients"`
	Metadata    map[string]interface{} `json:"metadata"`
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
	result := &OptimizationResult{
		OriginalPrompt: cfg.InitialPrompt,
		Iterations:     make([]IterationResult, 0, cfg.Iterations),
	}

	currentPrompt := cfg.InitialPrompt
	graph := tg.buildComputationGraph(currentPrompt)

	for i := 0; i < cfg.Iterations; i++ {
		// Compute textual gradients for each component
		gradients, err := tg.computeTextualGradients(ctx, graph, cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to compute gradients at iteration %d: %w", i+1, err)
		}

		// Apply gradients to improve the prompt
		improvedPrompt, score, feedback := tg.applyGradients(ctx, currentPrompt, gradients, cfg)

		iteration := IterationResult{
			Iteration: i + 1,
			Prompt:    improvedPrompt,
			Score:     score,
			Feedback:  feedback,
			Suggestions: tg.extractSuggestions(gradients),
		}

		result.Iterations = append(result.Iterations, iteration)
		currentPrompt = improvedPrompt

		// Update computation graph for next iteration
		graph = tg.buildComputationGraph(currentPrompt)
	}

	result.OptimizedPrompt = currentPrompt
	result.ImprovementScore = tg.calculateFinalScore(result)

	return result, nil
}

// computeTextualGradients generates natural language feedback for each component
func (tg *TextGradOptimizer) computeTextualGradients(ctx context.Context, graph ComputationGraph, cfg Config) ([]TextualGradient, error) {
	var gradients []TextualGradient

	for _, node := range graph.Nodes {
		if node.Type == "prompt" {
			gradient, err := tg.generateGradientForComponent(ctx, node, cfg)
			if err != nil {
				return nil, fmt.Errorf("failed to generate gradient for node %s: %w", node.ID, err)
			}
			gradients = append(gradients, gradient)
		}
	}

	return gradients, nil
}

// generateGradientForComponent creates a textual gradient for a specific component
func (tg *TextGradOptimizer) generateGradientForComponent(ctx context.Context, node GraphNode, cfg Config) (TextualGradient, error) {
	gradientPrompt := fmt.Sprintf(`You are an expert prompt optimization system that provides textual gradients for improvement.

COMPONENT TO ANALYZE:
Type: %s
Content: %s

TASK: Provide detailed textual gradient feedback to improve this component.

Your feedback should be:
1. SPECIFIC: Point to exact issues and improvements
2. ACTIONABLE: Provide concrete steps for improvement  
3. DIRECTIONAL: Indicate whether to increase/decrease certain aspects
4. MAGNITUDE-AWARE: Indicate how critical each improvement is

FORMAT YOUR RESPONSE AS:
GRADIENT: [Natural language description of what needs to change]
MAGNITUDE: [High/Medium/Low - how critical this change is]
DIRECTION: [Positive/Negative - whether to add or remove elements]
SUGGESTIONS:
- [Specific suggestion 1]
- [Specific suggestion 2]
- [Specific suggestion 3]

ANALYSIS:
Focus on these aspects:
- Clarity and precision of instructions
- Context sufficiency
- Constraint specification
- Error prevention
- Output quality predictability`, node.Type, node.Content)

	options := llm.GenerateOptions{
		Temperature: &cfg.Temperature,
		MaxTokens:   &cfg.MaxTokens,
	}

	response, err := tg.llm.Generate(ctx, gradientPrompt, options)
	if err != nil {
		return TextualGradient{}, fmt.Errorf("failed to generate gradient: %w", err)
	}

	return tg.parseGradientResponse(response.Text, node.ID), nil
}

// parseGradientResponse extracts gradient information from LLM response
func (tg *TextGradOptimizer) parseGradientResponse(response, componentID string) TextualGradient {
	lines := strings.Split(response, "\n")
	
	gradient := TextualGradient{
		Component: componentID,
		Magnitude: 0.5, // default
		Direction: "positive",
		Suggestions: []string{},
	}

	var inSuggestions bool
	for _, line := range lines {
		line = strings.TrimSpace(line)
		
		if strings.HasPrefix(line, "GRADIENT:") {
			gradient.Gradient = strings.TrimSpace(line[9:])
		} else if strings.HasPrefix(line, "MAGNITUDE:") {
			magnitudeStr := strings.ToLower(strings.TrimSpace(line[10:]))
			switch magnitudeStr {
			case "high":
				gradient.Magnitude = 0.9
			case "medium":
				gradient.Magnitude = 0.6
			case "low":
				gradient.Magnitude = 0.3
			}
		} else if strings.HasPrefix(line, "DIRECTION:") {
			gradient.Direction = strings.ToLower(strings.TrimSpace(line[10:]))
		} else if strings.HasPrefix(line, "SUGGESTIONS:") {
			inSuggestions = true
		} else if inSuggestions && strings.HasPrefix(line, "- ") {
			suggestion := strings.TrimSpace(line[2:])
			if suggestion != "" {
				gradient.Suggestions = append(gradient.Suggestions, suggestion)
			}
		}
	}

	return gradient
}

// applyGradients uses gradients to improve the prompt
func (tg *TextGradOptimizer) applyGradients(ctx context.Context, currentPrompt string, gradients []TextualGradient, cfg Config) (string, float64, string) {
	if len(gradients) == 0 {
		return currentPrompt, 5.0, "No gradients to apply"
	}

	// Combine all gradient feedback
	var feedback strings.Builder
	var allSuggestions []string

	feedback.WriteString("TEXTUAL GRADIENTS ANALYSIS:\n")
	for i, grad := range gradients {
		feedback.WriteString(fmt.Sprintf("\nGradient %d (Magnitude: %.1f, Direction: %s):\n", i+1, grad.Magnitude, grad.Direction))
		feedback.WriteString(fmt.Sprintf("Feedback: %s\n", grad.Gradient))
		allSuggestions = append(allSuggestions, grad.Suggestions...)
	}

	// Apply gradients to create improved prompt
	improvementPrompt := fmt.Sprintf(`You are an expert prompt engineer applying textual gradients for optimization.

CURRENT PROMPT:
%s

TEXTUAL GRADIENTS TO APPLY:
%s

IMPROVEMENT TASK:
Using the textual gradients above, create an improved version of the prompt that:
1. Addresses all the gradient feedback
2. Maintains the original intent and functionality
3. Incorporates the specific suggestions where applicable
4. Results in clearer, more effective instructions

Provide ONLY the improved prompt, without additional explanation.`, currentPrompt, feedback.String())

	options := llm.GenerateOptions{
		Temperature: &cfg.Temperature,
		MaxTokens:   &cfg.MaxTokens,
	}

	response, err := tg.llm.Generate(ctx, improvementPrompt, options)
	if err != nil {
		return currentPrompt, 0.0, "Failed to apply gradients"
	}

	improvedPrompt := strings.TrimSpace(response.Text)
	
	// Score the improvement
	score := tg.scoreImprovement(ctx, currentPrompt, improvedPrompt, gradients, cfg)
	
	return improvedPrompt, score, feedback.String()
}

// scoreImprovement evaluates how well the gradients were applied
func (tg *TextGradOptimizer) scoreImprovement(ctx context.Context, originalPrompt, improvedPrompt string, gradients []TextualGradient, cfg Config) float64 {
	scoringPrompt := fmt.Sprintf(`You are an expert prompt evaluator. Score how well the improved prompt addresses the textual gradients.

ORIGINAL PROMPT:
%s

IMPROVED PROMPT:
%s

GRADIENTS THAT WERE APPLIED:
%s

EVALUATION TASK:
Rate the improvement on a scale of 1-10 based on:
1. How well the gradients were addressed (40%%)
2. Overall prompt quality improvement (30%%)
3. Preservation of original intent (20%%)
4. Practical effectiveness (10%%)

Provide ONLY a single number from 1-10.`, originalPrompt, improvedPrompt, tg.formatGradientsForScoring(gradients))

	options := llm.GenerateOptions{
		Temperature: &cfg.Temperature,
		MaxTokens:   &cfg.MaxTokens,
	}

	response, err := tg.llm.Generate(ctx, scoringPrompt, options)
	if err != nil {
		return 5.0 // default score
	}

	// Extract numeric score
	scoreText := strings.TrimSpace(response.Text)
	if len(scoreText) > 0 && scoreText[0] >= '1' && scoreText[0] <= '9' {
		return float64(scoreText[0] - '0')
	}
	
	return 5.0 // default if parsing fails
}

// buildComputationGraph creates a computation graph for the prompt
func (tg *TextGradOptimizer) buildComputationGraph(prompt string) ComputationGraph {
	// Simple graph with just the prompt node for now
	// This can be extended to include multi-step reasoning, tool calls, etc.
	graph := ComputationGraph{
		Nodes: []GraphNode{
			{
				ID:      "prompt_root",
				Type:    "prompt",
				Content: prompt,
				Metadata: map[string]interface{}{
					"is_root": true,
				},
			},
		},
		Edges: []GraphEdge{},
	}

	return graph
}

// extractSuggestions extracts all suggestions from gradients
func (tg *TextGradOptimizer) extractSuggestions(gradients []TextualGradient) []string {
	var allSuggestions []string
	for _, grad := range gradients {
		allSuggestions = append(allSuggestions, grad.Suggestions...)
	}
	return allSuggestions
}

// formatGradientsForScoring formats gradients for the scoring prompt
func (tg *TextGradOptimizer) formatGradientsForScoring(gradients []TextualGradient) string {
	var formatted strings.Builder
	for i, grad := range gradients {
		formatted.WriteString(fmt.Sprintf("Gradient %d: %s (Magnitude: %.1f)\n", i+1, grad.Gradient, grad.Magnitude))
	}
	return formatted.String()
}

// calculateFinalScore calculates the final improvement score
func (tg *TextGradOptimizer) calculateFinalScore(result *OptimizationResult) float64 {
	if len(result.Iterations) == 0 {
		return 0.0
	}

	// Weight more recent iterations higher
	var weightedSum, totalWeight float64
	for i, iter := range result.Iterations {
		weight := float64(i+1) / float64(len(result.Iterations)) // Later iterations get higher weight
		weightedSum += iter.Score * weight
		totalWeight += weight
	}

	return weightedSum / totalWeight
}