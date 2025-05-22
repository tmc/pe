package metaprompt

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/tmc/pe/internal/llm"
)

// SemanticOptimizer implements semantic backpropagation and gradient descent
// Based on 2025 KAUST/IDSIA research on semantic optimization
type SemanticOptimizer struct {
	llm llm.Provider
}

// NewSemanticOptimizer creates a new semantic optimizer
func NewSemanticOptimizer(llmProvider llm.Provider) *SemanticOptimizer {
	return &SemanticOptimizer{
		llm: llmProvider,
	}
}

// SemanticConfig configures semantic backpropagation
type SemanticConfig struct {
	Target     string
	Iterations int
	Verbose    bool
	Provider   string
	Model      string
}

// SemanticDescentConfig configures semantic gradient descent
type SemanticDescentConfig struct {
	Objective            string
	LearningRate         float64
	Iterations           int
	ConvergenceThreshold float64
	AdaptiveLearning     bool
	Provider             string
	Model                string
}

// SemanticResult represents the result of semantic optimization
type SemanticResult struct {
	InitialScore     float64           `json:"initial_score"`
	FinalScore       float64           `json:"final_score"`
	OptimizedPrompt  string            `json:"optimized_prompt"`
	Iterations       int               `json:"iterations"`
	Converged        bool              `json:"converged"`
	SemanticGradients []SemanticGradient `json:"semantic_gradients"`
	Duration         time.Duration     `json:"duration"`
	Trajectory       []SemanticOptimizationStep `json:"trajectory"`
}

// SemanticGradient represents a semantic gradient with directional information
type SemanticGradient struct {
	Component  string  `json:"component"`
	Direction  string  `json:"direction"`
	Magnitude  float64 `json:"magnitude"`
	Reasoning  string  `json:"reasoning"`
	Confidence float64 `json:"confidence"`
}

// SemanticOptimizationStep represents a single step in the semantic optimization trajectory
type SemanticOptimizationStep struct {
	Iteration   int     `json:"iteration"`
	Score       float64 `json:"score"`
	Prompt      string  `json:"prompt"`
	Gradient    string  `json:"gradient"`
	Improvement float64 `json:"improvement"`
}

// SemanticBackpropagation implements semantic backpropagation algorithm
func (so *SemanticOptimizer) SemanticBackpropagation(ctx context.Context, prompt string, config SemanticConfig) (*SemanticResult, error) {
	start := time.Now()
	
	result := &SemanticResult{
		OptimizedPrompt: prompt,
		Trajectory:      make([]SemanticOptimizationStep, 0, config.Iterations),
	}
	
	// Initial evaluation
	initialScore, err := so.evaluatePrompt(ctx, prompt, config.Target)
	if err != nil {
		return nil, fmt.Errorf("initial evaluation failed: %v", err)
	}
	result.InitialScore = initialScore
	result.FinalScore = initialScore
	
	currentPrompt := prompt
	currentScore := initialScore
	
	if config.Verbose {
		fmt.Printf("Initial score: %.4f\n", initialScore)
	}
	
	// Semantic backpropagation iterations
	for i := 0; i < config.Iterations; i++ {
		// Compute semantic gradients
		gradients, err := so.computeSemanticGradients(ctx, currentPrompt, config.Target, currentScore)
		if err != nil {
			return nil, fmt.Errorf("gradient computation failed at iteration %d: %v", i, err)
		}
		
		// Apply semantic gradients to optimize prompt
		optimizedPrompt, err := so.applySemanticGradients(ctx, currentPrompt, gradients)
		if err != nil {
			return nil, fmt.Errorf("gradient application failed at iteration %d: %v", i, err)
		}
		
		// Evaluate optimized prompt
		newScore, err := so.evaluatePrompt(ctx, optimizedPrompt, config.Target)
		if err != nil {
			return nil, fmt.Errorf("evaluation failed at iteration %d: %v", i, err)
		}
		
		// Record optimization step
		step := SemanticOptimizationStep{
			Iteration:   i + 1,
			Score:       newScore,
			Prompt:      optimizedPrompt,
			Gradient:    formatGradients(gradients),
			Improvement: newScore - currentScore,
		}
		result.Trajectory = append(result.Trajectory, step)
		
		if config.Verbose {
			fmt.Printf("Iteration %d: score %.4f (Δ%.4f)\n", i+1, newScore, newScore-currentScore)
		}
		
		// Update current state if improvement
		if newScore > currentScore {
			currentPrompt = optimizedPrompt
			currentScore = newScore
			result.FinalScore = newScore
			result.OptimizedPrompt = optimizedPrompt
		}
		
		// Store gradients from best iteration
		if i == 0 || newScore > result.FinalScore {
			result.SemanticGradients = gradients
		}
	}
	
	result.Iterations = config.Iterations
	result.Converged = result.FinalScore > result.InitialScore
	result.Duration = time.Since(start)
	
	return result, nil
}

// SemanticGradientDescent implements semantic gradient descent optimization
func (so *SemanticOptimizer) SemanticGradientDescent(ctx context.Context, prompt string, config SemanticDescentConfig) (*SemanticResult, error) {
	start := time.Now()
	
	result := &SemanticResult{
		OptimizedPrompt: prompt,
		Trajectory:      make([]SemanticOptimizationStep, 0, config.Iterations),
	}
	
	// Initial evaluation
	initialScore, err := so.evaluatePrompt(ctx, prompt, config.Objective)
	if err != nil {
		return nil, fmt.Errorf("initial evaluation failed: %v", err)
	}
	result.InitialScore = initialScore
	result.FinalScore = initialScore
	
	currentPrompt := prompt
	currentScore := initialScore
	learningRate := config.LearningRate
	
	// Gradient descent iterations
	for i := 0; i < config.Iterations; i++ {
		// Compute semantic gradients
		gradients, err := so.computeSemanticGradients(ctx, currentPrompt, config.Objective, currentScore)
		if err != nil {
			return nil, fmt.Errorf("gradient computation failed at iteration %d: %v", i, err)
		}
		
		// Apply gradients with learning rate
		optimizedPrompt, err := so.applyGradientDescent(ctx, currentPrompt, gradients, learningRate)
		if err != nil {
			return nil, fmt.Errorf("gradient descent failed at iteration %d: %v", i, err)
		}
		
		// Evaluate new prompt
		newScore, err := so.evaluatePrompt(ctx, optimizedPrompt, config.Objective)
		if err != nil {
			return nil, fmt.Errorf("evaluation failed at iteration %d: %v", i, err)
		}
		
		improvement := newScore - currentScore
		
		// Record step
		step := SemanticOptimizationStep{
			Iteration:   i + 1,
			Score:       newScore,
			Prompt:      optimizedPrompt,
			Gradient:    formatGradients(gradients),
			Improvement: improvement,
		}
		result.Trajectory = append(result.Trajectory, step)
		
		// Adaptive learning rate
		if config.AdaptiveLearning {
			if improvement > 0 {
				learningRate *= 1.1 // Increase learning rate on improvement
			} else {
				learningRate *= 0.9 // Decrease learning rate on decline
			}
			learningRate = math.Max(0.01, math.Min(1.0, learningRate))
		}
		
		// Update current state
		currentPrompt = optimizedPrompt
		currentScore = newScore
		result.FinalScore = newScore
		result.OptimizedPrompt = optimizedPrompt
		
		// Check convergence
		if math.Abs(improvement) < config.ConvergenceThreshold {
			result.Converged = true
			result.Iterations = i + 1
			break
		}
	}
	
	if !result.Converged {
		result.Iterations = config.Iterations
	}
	
	result.Duration = time.Since(start)
	return result, nil
}

// computeSemanticGradients computes semantic gradients using LLM feedback
func (so *SemanticOptimizer) computeSemanticGradients(ctx context.Context, prompt, objective string, currentScore float64) ([]SemanticGradient, error) {
	gradientPrompt := fmt.Sprintf(`You are a semantic gradient computer for prompt optimization. 

Analyze the following prompt and provide semantic gradients for improvement:

CURRENT PROMPT:
%s

OBJECTIVE: %s
CURRENT SCORE: %.4f

Please provide 3-5 semantic gradients, each consisting of:
1. Component: Which part of the prompt to modify
2. Direction: How to modify it (more specific, clearer, concise, etc.)
3. Magnitude: How important this change is (0.0-1.0)
4. Reasoning: Why this change would improve the prompt
5. Confidence: How confident you are in this gradient (0.0-1.0)

Format your response as JSON with the following structure:
{
  "gradients": [
    {
      "component": "instruction clarity",
      "direction": "make more specific and actionable",
      "magnitude": 0.8,
      "reasoning": "The current instruction is too vague...",
      "confidence": 0.9
    }
  ]
}`, prompt, objective, currentScore)

	response, err := so.llm.Generate(ctx, gradientPrompt, llm.GenerateOptions{
		Temperature: &[]float64{0.1}[0], // Low temperature for consistent analysis
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate semantic gradients: %v", err)
	}

	// Parse gradients from response
	gradients, err := parseSemanticGradients(response.Text)
	if err != nil {
		return nil, fmt.Errorf("failed to parse semantic gradients: %v", err)
	}

	return gradients, nil
}

// applySemanticGradients applies semantic gradients to optimize the prompt
func (so *SemanticOptimizer) applySemanticGradients(ctx context.Context, prompt string, gradients []SemanticGradient) (string, error) {
	optimizationPrompt := fmt.Sprintf(`Apply the following semantic gradients to optimize this prompt:

ORIGINAL PROMPT:
%s

SEMANTIC GRADIENTS:
%s

Please provide an optimized version of the prompt that incorporates these gradients.
Focus on the highest magnitude gradients first.

Return only the optimized prompt without additional explanation.`, 
		prompt, formatGradientsForApplication(gradients))

	response, err := so.llm.Generate(ctx, optimizationPrompt, llm.GenerateOptions{
		Temperature: &[]float64{0.3}[0],
	})
	if err != nil {
		return "", fmt.Errorf("failed to apply semantic gradients: %v", err)
	}

	return response.Text, nil
}

// applyGradientDescent applies gradients with learning rate scaling
func (so *SemanticOptimizer) applyGradientDescent(ctx context.Context, prompt string, gradients []SemanticGradient, learningRate float64) (string, error) {
	// Scale gradients by learning rate
	scaledGradients := make([]SemanticGradient, len(gradients))
	for i, grad := range gradients {
		scaledGradients[i] = grad
		scaledGradients[i].Magnitude *= learningRate
	}

	return so.applySemanticGradients(ctx, prompt, scaledGradients)
}

// evaluatePrompt evaluates a prompt against an objective
func (so *SemanticOptimizer) evaluatePrompt(ctx context.Context, prompt, objective string) (float64, error) {
	evaluationPrompt := fmt.Sprintf(`Evaluate the following prompt against the given objective:

PROMPT:
%s

OBJECTIVE: %s

Rate the prompt's effectiveness on a scale of 0.0 to 1.0, considering:
- Clarity and specificity
- Alignment with objective
- Potential for good results
- Structure and organization

Provide only a numeric score between 0.0 and 1.0.`, prompt, objective)

	response, err := so.llm.Generate(ctx, evaluationPrompt, llm.GenerateOptions{
		Temperature: &[]float64{0.0}[0], // Deterministic evaluation
	})
	if err != nil {
		return 0, fmt.Errorf("failed to evaluate prompt: %v", err)
	}

	// Parse score from response
	score, err := parseScore(response.Text)
	if err != nil {
		return 0, fmt.Errorf("failed to parse evaluation score: %v", err)
	}

	return score, nil
}

// Helper functions (implementation details)
func parseSemanticGradients(response string) ([]SemanticGradient, error) {
	// TODO: Implement JSON parsing of gradients
	// For now, return a placeholder
	return []SemanticGradient{
		{
			Component:  "clarity",
			Direction:  "more specific",
			Magnitude:  0.8,
			Reasoning:  "Improve specificity",
			Confidence: 0.9,
		},
	}, nil
}

func formatGradients(gradients []SemanticGradient) string {
	result := ""
	for i, grad := range gradients {
		result += fmt.Sprintf("%d. %s: %s (%.2f)\n", i+1, grad.Component, grad.Direction, grad.Magnitude)
	}
	return result
}

func formatGradientsForApplication(gradients []SemanticGradient) string {
	result := ""
	for i, grad := range gradients {
		result += fmt.Sprintf("%d. Component: %s\n   Direction: %s\n   Magnitude: %.2f\n   Reasoning: %s\n\n", 
			i+1, grad.Component, grad.Direction, grad.Magnitude, grad.Reasoning)
	}
	return result
}

func parseScore(response string) (float64, error) {
	// TODO: Implement robust score parsing
	// For now, return a placeholder
	return 0.75, nil
}