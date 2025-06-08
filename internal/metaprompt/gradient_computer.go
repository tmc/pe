package metaprompt

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strings"

	"github.com/tmc/pe/internal/llm"
)

// GradientComputer computes optimization trajectories for prompt refinement
type GradientComputer struct {
	llm llm.Provider
}

// NewGradientComputer creates a new prompt gradient computer
func NewGradientComputer(llmProvider llm.Provider) *GradientComputer {
	return &GradientComputer{
		llm: llmProvider,
	}
}

// LossFunction represents semantic alignment loss calculation
type LossFunction struct {
	Type        string  `json:"type"`        // "semantic", "structural", "task_specific"
	Weight      float64 `json:"weight"`      // Importance weight
	Value       float64 `json:"value"`       // Current loss value
	Gradient    float64 `json:"gradient"`    // Loss gradient
	Description string  `json:"description"` // Human-readable description
}

// OptimizationStep represents a single optimization step
type OptimizationStep struct {
	StepSize     float64 `json:"step_size"`
	Direction    string  `json:"direction"`     // "increase", "decrease", "modify"
	Component    string  `json:"component"`     // Which part to modify
	Modification string  `json:"modification"`  // Specific change to make
	ExpectedGain float64 `json:"expected_gain"` // Predicted improvement
	Confidence   float64 `json:"confidence"`    // Confidence in this step
}

// GradientBuffer accumulates gradients over time
type GradientBuffer struct {
	Gradients    []TextualGradient `json:"gradients"`
	Capacity     int               `json:"capacity"`
	WindowSize   int               `json:"window_size"`
	Momentum     float64           `json:"momentum"`
	LearningRate float64           `json:"learning_rate"`
}

// ComputationResult contains gradient computation results
type ComputationResult struct {
	LossFunctions      []LossFunction     `json:"loss_functions"`
	OptimizationSteps  []OptimizationStep `json:"optimization_steps"`
	GradientBuffer     GradientBuffer     `json:"gradient_buffer"`
	ConvergenceMetrics ConvergenceMetrics `json:"convergence_metrics"`
	Recommendations    []string           `json:"recommendations"`
}

// ConvergenceMetrics tracks optimization progress
type ConvergenceMetrics struct {
	CurrentLoss    float64 `json:"current_loss"`
	LossReduction  float64 `json:"loss_reduction"`
	GradientNorm   float64 `json:"gradient_norm"`
	Convergence    float64 `json:"convergence"`     // 0.0 = not converged, 1.0 = fully converged
	Stability      float64 `json:"stability"`       // How stable the optimization is
	IterationsLeft int     `json:"iterations_left"` // Estimated iterations to convergence
}

// ComputeGradients calculates optimization trajectories for prompt refinement
func (gc *GradientComputer) ComputeGradients(ctx context.Context, currentPrompt, targetObjective string, history []IterationResult) (*ComputationResult, error) {
	result := &ComputationResult{}

	// Calculate loss functions
	lossFunctions, err := gc.calculateLossFunctions(ctx, currentPrompt, targetObjective, history)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate loss functions: %w", err)
	}
	result.LossFunctions = lossFunctions

	// Compute optimization steps
	steps, err := gc.computeOptimizationSteps(ctx, currentPrompt, lossFunctions)
	if err != nil {
		return nil, fmt.Errorf("failed to compute optimization steps: %w", err)
	}
	result.OptimizationSteps = steps

	// Update gradient buffer
	buffer := gc.updateGradientBuffer(history)
	result.GradientBuffer = buffer

	// Calculate convergence metrics
	metrics := gc.calculateConvergenceMetrics(lossFunctions, history, buffer)
	result.ConvergenceMetrics = metrics

	// Generate recommendations
	recommendations, err := gc.generateRecommendations(ctx, result)
	if err != nil {
		return nil, fmt.Errorf("failed to generate recommendations: %w", err)
	}
	result.Recommendations = recommendations

	return result, nil
}

// calculateLossFunctions computes semantic alignment loss
func (gc *GradientComputer) calculateLossFunctions(ctx context.Context, prompt, objective string, history []IterationResult) ([]LossFunction, error) {
	lossPrompt := fmt.Sprintf(`You are an expert in optimization theory and prompt engineering. Calculate loss functions for prompt optimization.

CURRENT PROMPT:
%s

TARGET OBJECTIVE:
%s

OPTIMIZATION HISTORY:
%s

TASK: Identify and quantify loss functions that measure how far the current prompt is from achieving the target objective.

Consider these loss types:
1. Semantic Loss - How well does the prompt convey the intended meaning?
2. Structural Loss - How well-organized and clear is the prompt structure?  
3. Task-Specific Loss - How well does it guide toward the specific objective?
4. Robustness Loss - How likely is it to work across different inputs?

FORMAT your response as JSON:
{
  "loss_functions": [
    {
      "type": "semantic",
      "weight": 0.4,
      "value": 0.3,
      "gradient": -0.15,
      "description": "Semantic alignment between prompt and objective"
    }
  ]
}`, prompt, objective, gc.formatHistory(history))

	options := llm.GenerateOptions{
		Temperature: floatPtr(0.2),
		MaxTokens:   intPtr(800),
	}

	resp, err := gc.llm.Generate(ctx, lossPrompt, options)
	if err != nil {
		return nil, err
	}

	var result struct {
		LossFunctions []LossFunction `json:"loss_functions"`
	}

	if err := json.Unmarshal([]byte(resp.Text), &result); err != nil {
		return gc.generateDefaultLossFunctions(), nil
	}

	return result.LossFunctions, nil
}

// computeOptimizationSteps determines the best optimization trajectory
func (gc *GradientComputer) computeOptimizationSteps(ctx context.Context, prompt string, lossFunctions []LossFunction) ([]OptimizationStep, error) {
	stepsPrompt := fmt.Sprintf(`You are an expert optimization engineer. Based on these loss functions, compute the optimal steps to improve the prompt.

CURRENT PROMPT:
%s

LOSS FUNCTIONS:
%s

TASK: Generate 3-5 optimization steps that will most effectively reduce the total loss. Each step should:
1. Target the highest-impact loss reductions
2. Have a clear modification strategy
3. Include expected improvement and confidence

FORMAT your response as JSON:
{
  "optimization_steps": [
    {
      "step_size": 0.3,
      "direction": "modify",
      "component": "instruction_clarity", 
      "modification": "Add specific output format requirements",
      "expected_gain": 0.15,
      "confidence": 0.8
    }
  ]
}`, prompt, gc.formatLossFunctions(lossFunctions))

	options := llm.GenerateOptions{
		Temperature: floatPtr(0.3),
		MaxTokens:   intPtr(1000),
	}

	resp, err := gc.llm.Generate(ctx, stepsPrompt, options)
	if err != nil {
		return nil, err
	}

	var result struct {
		OptimizationSteps []OptimizationStep `json:"optimization_steps"`
	}

	if err := json.Unmarshal([]byte(resp.Text), &result); err != nil {
		return gc.generateDefaultSteps(), nil
	}

	return result.OptimizationSteps, nil
}

// updateGradientBuffer accumulates gradients with momentum
func (gc *GradientComputer) updateGradientBuffer(history []IterationResult) GradientBuffer {
	buffer := GradientBuffer{
		Capacity:     10,
		WindowSize:   5,
		Momentum:     0.9,
		LearningRate: 0.1,
		Gradients:    []TextualGradient{},
	}

	// Extract gradients from recent history
	windowStart := max(0, len(history)-buffer.WindowSize)
	for i := windowStart; i < len(history); i++ {
		// Convert iteration feedback to textual gradients
		if history[i].Feedback != "" {
			gradient := TextualGradient{
				Component:   fmt.Sprintf("iteration_%d", history[i].Iteration),
				Gradient:    history[i].Feedback,
				Magnitude:   history[i].Score / 10.0, // Normalize score to 0-1
				Direction:   gc.determineDirection(history[i].Score),
				Suggestions: history[i].Suggestions,
			}
			buffer.Gradients = append(buffer.Gradients, gradient)
		}
	}

	// Keep only recent gradients within capacity
	if len(buffer.Gradients) > buffer.Capacity {
		buffer.Gradients = buffer.Gradients[len(buffer.Gradients)-buffer.Capacity:]
	}

	return buffer
}

// calculateConvergenceMetrics tracks optimization progress
func (gc *GradientComputer) calculateConvergenceMetrics(lossFunctions []LossFunction, history []IterationResult, buffer GradientBuffer) ConvergenceMetrics {
	// Calculate current total loss
	var currentLoss float64
	for _, lf := range lossFunctions {
		currentLoss += lf.Value * lf.Weight
	}

	// Calculate loss reduction from history
	var lossReduction float64
	if len(history) >= 2 {
		firstScore := history[0].Score
		lastScore := history[len(history)-1].Score
		lossReduction = (lastScore - firstScore) / 10.0 // Normalize
	}

	// Calculate gradient norm
	var gradientNorm float64
	for _, gradient := range buffer.Gradients {
		gradientNorm += gradient.Magnitude * gradient.Magnitude
	}
	gradientNorm = math.Sqrt(gradientNorm)

	// Estimate convergence (how close to optimal)
	convergence := 1.0 - currentLoss
	if convergence < 0 {
		convergence = 0
	}

	// Calculate stability (consistency of recent improvements)
	stability := gc.calculateStability(history)

	// Estimate iterations left
	iterationsLeft := gc.estimateIterationsToConvergence(currentLoss, gradientNorm, stability)

	return ConvergenceMetrics{
		CurrentLoss:    currentLoss,
		LossReduction:  lossReduction,
		GradientNorm:   gradientNorm,
		Convergence:    convergence,
		Stability:      stability,
		IterationsLeft: iterationsLeft,
	}
}

// generateRecommendations creates actionable optimization advice
func (gc *GradientComputer) generateRecommendations(ctx context.Context, result *ComputationResult) ([]string, error) {
	recPrompt := fmt.Sprintf(`Based on this gradient computation analysis, provide specific optimization recommendations.

CONVERGENCE STATUS:
- Current Loss: %.3f
- Convergence: %.3f
- Gradient Norm: %.3f
- Stability: %.3f

TOP OPTIMIZATION STEPS:
%s

TASK: Generate 3-5 actionable recommendations for the next optimization steps.

FORMAT:
1. [Specific recommendation 1]
2. [Specific recommendation 2]
3. [Specific recommendation 3]
4. [Specific recommendation 4]
5. [Specific recommendation 5]`,
		result.ConvergenceMetrics.CurrentLoss,
		result.ConvergenceMetrics.Convergence,
		result.ConvergenceMetrics.GradientNorm,
		result.ConvergenceMetrics.Stability,
		gc.formatOptimizationSteps(result.OptimizationSteps))

	options := llm.GenerateOptions{
		Temperature: floatPtr(0.3),
		MaxTokens:   intPtr(500),
	}

	resp, err := gc.llm.Generate(ctx, recPrompt, options)
	if err != nil {
		return nil, err
	}

	return gc.parseRecommendations(resp.Text), nil
}

// Helper functions
func (gc *GradientComputer) formatHistory(history []IterationResult) string {
	if len(history) == 0 {
		return "No optimization history available"
	}

	var formatted strings.Builder
	for i, iter := range history {
		if i >= 3 { // Limit to last 3 iterations
			break
		}
		formatted.WriteString(fmt.Sprintf("Iteration %d: Score %.1f - %s\n",
			iter.Iteration, iter.Score, truncateString(iter.Feedback, 100)))
	}
	return formatted.String()
}

func (gc *GradientComputer) formatLossFunctions(lossFunctions []LossFunction) string {
	var formatted strings.Builder
	for _, lf := range lossFunctions {
		formatted.WriteString(fmt.Sprintf("%s (Weight: %.2f, Value: %.3f, Gradient: %.3f): %s\n",
			lf.Type, lf.Weight, lf.Value, lf.Gradient, lf.Description))
	}
	return formatted.String()
}

func (gc *GradientComputer) formatOptimizationSteps(steps []OptimizationStep) string {
	var formatted strings.Builder
	for i, step := range steps {
		if i >= 3 { // Limit to top 3 steps
			break
		}
		formatted.WriteString(fmt.Sprintf("Step %d: %s (%s) - Expected gain: %.2f, Confidence: %.2f\n",
			i+1, step.Modification, step.Component, step.ExpectedGain, step.Confidence))
	}
	return formatted.String()
}

func (gc *GradientComputer) determineDirection(score float64) string {
	if score >= 7.0 {
		return "positive"
	} else if score <= 3.0 {
		return "negative"
	}
	return "neutral"
}

func (gc *GradientComputer) calculateStability(history []IterationResult) float64 {
	if len(history) < 3 {
		return 0.5 // Default for insufficient data
	}

	// Calculate variance in score improvements
	var improvements []float64
	for i := 1; i < len(history); i++ {
		improvement := history[i].Score - history[i-1].Score
		improvements = append(improvements, improvement)
	}

	// Calculate coefficient of variation (lower = more stable)
	mean := gc.calculateMean(improvements)
	variance := gc.calculateVariance(improvements, mean)
	stdDev := math.Sqrt(variance)

	if mean == 0 {
		return 0.5
	}

	cv := stdDev / math.Abs(mean)
	stability := 1.0 / (1.0 + cv) // Higher stability for lower coefficient of variation

	return stability
}

func (gc *GradientComputer) estimateIterationsToConvergence(currentLoss, gradientNorm, stability float64) int {
	// Simple heuristic based on current state
	if currentLoss < 0.1 && gradientNorm < 0.05 {
		return 1 // Almost converged
	}

	if stability > 0.8 && gradientNorm > 0.3 {
		return 2 // Good progress expected
	}

	if stability < 0.3 {
		return 6 // Unstable, may need more iterations
	}

	return 3 // Default estimate
}

func (gc *GradientComputer) calculateMean(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

func (gc *GradientComputer) calculateVariance(values []float64, mean float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range values {
		diff := v - mean
		sum += diff * diff
	}
	return sum / float64(len(values))
}

func (gc *GradientComputer) parseRecommendations(text string) []string {
	lines := strings.Split(text, "\n")
	var recommendations []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "1.") || strings.HasPrefix(line, "2.") ||
			strings.HasPrefix(line, "3.") || strings.HasPrefix(line, "4.") ||
			strings.HasPrefix(line, "5.") {
			rec := strings.TrimSpace(line[2:])
			if rec != "" {
				recommendations = append(recommendations, rec)
			}
		}
	}

	return recommendations
}

func (gc *GradientComputer) generateDefaultLossFunctions() []LossFunction {
	return []LossFunction{
		{
			Type:        "semantic",
			Weight:      0.4,
			Value:       0.5,
			Gradient:    -0.2,
			Description: "Semantic alignment with target objective",
		},
		{
			Type:        "structural",
			Weight:      0.3,
			Value:       0.3,
			Gradient:    -0.1,
			Description: "Prompt structure and clarity",
		},
		{
			Type:        "task_specific",
			Weight:      0.3,
			Value:       0.4,
			Gradient:    -0.15,
			Description: "Task-specific effectiveness",
		},
	}
}

func (gc *GradientComputer) generateDefaultSteps() []OptimizationStep {
	return []OptimizationStep{
		{
			StepSize:     0.3,
			Direction:    "modify",
			Component:    "instruction_clarity",
			Modification: "Improve instruction clarity and specificity",
			ExpectedGain: 0.2,
			Confidence:   0.7,
		},
	}
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
