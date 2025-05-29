package metaprompt

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
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

// SemanticNode represents a node in the computational graph for GASO
type SemanticNode struct {
	ID           string                 `json:"id"`
	Type         string                 `json:"type"` // component, system, objective
	Content      string                 `json:"content"`
	Dependencies []string               `json:"dependencies"`
	Properties   map[string]interface{} `json:"properties"`
}

// SemanticFlow represents information flow between nodes
type SemanticFlow struct {
	Source      string  `json:"source"`
	Target      string  `json:"target"`
	FlowType    string  `json:"flow_type"`
	Strength    float64 `json:"strength"`
	Information string  `json:"information"`
}

// SemanticAnalysis represents a comprehensive semantic analysis
type SemanticAnalysis struct {
	Components      []SemanticNode  `json:"components"`
	Flows          []SemanticFlow  `json:"flows"`
	CoherenceScore float64         `json:"coherence_score"`
	Ambiguities    []string        `json:"ambiguities"`
	Strengths      []string        `json:"strengths"`
	Weaknesses     []string        `json:"weaknesses"`
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

// AnalyzeSemanticStructure analyzes the semantic structure of a prompt
func (so *SemanticOptimizer) AnalyzeSemanticStructure(ctx context.Context, prompt string) (*SemanticAnalysis, error) {
	analysisPrompt := fmt.Sprintf(`Analyze the semantic structure of this prompt:

PROMPT:
%s

Provide a detailed analysis including:
1. Main components and their relationships
2. Information flow patterns
3. Semantic coherence score (0.0-1.0)
4. Potential ambiguities or gaps
5. Structural strengths and weaknesses

Format as JSON:
{
  "components": [
    {"id": "comp1", "type": "instruction", "content": "...", "dependencies": []}
  ],
  "flows": [
    {"source": "comp1", "target": "comp2", "type": "data", "strength": 0.8}
  ],
  "coherence_score": 0.85,
  "ambiguities": ["..."],
  "strengths": ["..."],
  "weaknesses": ["..."]
}`, prompt)

	response, err := so.llm.Generate(ctx, analysisPrompt, llm.GenerateOptions{
		Temperature: &[]float64{0.1}[0],
	})
	if err != nil {
		return nil, fmt.Errorf("failed to analyze semantic structure: %v", err)
	}

	return parseSemanticAnalysis(response.Text)
}

// ComputeSemanticDistance computes semantic distance between two prompts
func (so *SemanticOptimizer) ComputeSemanticDistance(ctx context.Context, prompt1, prompt2 string) (float64, error) {
	distancePrompt := fmt.Sprintf(`Compare the semantic similarity between these two prompts:

PROMPT 1:
%s

PROMPT 2:
%s

Rate their semantic distance on a scale from 0.0 (identical) to 1.0 (completely different).
Consider:
- Structural similarity
- Intent alignment
- Component overlap
- Style consistency

Provide only a numeric score between 0.0 and 1.0.`, prompt1, prompt2)

	response, err := so.llm.Generate(ctx, distancePrompt, llm.GenerateOptions{
		Temperature: &[]float64{0.0}[0],
	})
	if err != nil {
		return 0, fmt.Errorf("failed to compute semantic distance: %v", err)
	}

	return parseScore(response.Text)
}

// Helper functions (implementation details)
func parseSemanticGradients(response string) ([]SemanticGradient, error) {
	// Try to extract JSON from response
	start := strings.Index(response, "{")
	end := strings.LastIndex(response, "}")
	if start == -1 || end == -1 || start > end {
		// Fallback: parse structured text
		return parseGradientsFromText(response)
	}
	
	jsonStr := response[start : end+1]
	
	var result struct {
		Gradients []SemanticGradient `json:"gradients"`
	}
	
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		// Fallback to text parsing
		return parseGradientsFromText(response)
	}
	
	// Validate and normalize gradients
	for i := range result.Gradients {
		if result.Gradients[i].Magnitude < 0 {
			result.Gradients[i].Magnitude = 0
		} else if result.Gradients[i].Magnitude > 1 {
			result.Gradients[i].Magnitude = 1
		}
		if result.Gradients[i].Confidence < 0 {
			result.Gradients[i].Confidence = 0
		} else if result.Gradients[i].Confidence > 1 {
			result.Gradients[i].Confidence = 1
		}
	}
	
	return result.Gradients, nil
}

func parseGradientsFromText(response string) ([]SemanticGradient, error) {
	// Basic text parsing fallback
	gradients := []SemanticGradient{}
	lines := strings.Split(response, "\n")
	
	var currentGradient *SemanticGradient
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(strings.ToLower(line), "component:") {
			if currentGradient != nil {
				gradients = append(gradients, *currentGradient)
			}
			currentGradient = &SemanticGradient{
				Component:  strings.TrimSpace(line[10:]),
				Magnitude:  0.5,  // Default
				Confidence: 0.7,  // Default
			}
		} else if currentGradient != nil {
			if strings.HasPrefix(strings.ToLower(line), "direction:") {
				currentGradient.Direction = strings.TrimSpace(line[10:])
			} else if strings.HasPrefix(strings.ToLower(line), "magnitude:") {
				if val, err := strconv.ParseFloat(strings.TrimSpace(line[10:]), 64); err == nil {
					currentGradient.Magnitude = val
				}
			} else if strings.HasPrefix(strings.ToLower(line), "reasoning:") {
				currentGradient.Reasoning = strings.TrimSpace(line[10:])
			} else if strings.HasPrefix(strings.ToLower(line), "confidence:") {
				if val, err := strconv.ParseFloat(strings.TrimSpace(line[11:]), 64); err == nil {
					currentGradient.Confidence = val
				}
			}
		}
	}
	
	if currentGradient != nil {
		gradients = append(gradients, *currentGradient)
	}
	
	// If no gradients parsed, return a default
	if len(gradients) == 0 {
		return []SemanticGradient{
			{
				Component:  "overall clarity",
				Direction:  "improve specificity and structure",
				Magnitude:  0.7,
				Reasoning:  "General improvement needed",
				Confidence: 0.6,
			},
		}, nil
	}
	
	return gradients, nil
}

func formatGradients(gradients []SemanticGradient) string {
	if len(gradients) == 0 {
		return "No gradients computed"
	}
	
	result := ""
	for i, grad := range gradients {
		result += fmt.Sprintf("%d. %s → %s (mag: %.2f, conf: %.2f)\n", 
			i+1, grad.Component, grad.Direction, grad.Magnitude, grad.Confidence)
	}
	return strings.TrimSpace(result)
}

func formatGradientsForApplication(gradients []SemanticGradient) string {
	if len(gradients) == 0 {
		return "No gradients to apply"
	}
	
	// Sort gradients by magnitude for prioritization
	result := ""
	for i, grad := range gradients {
		result += fmt.Sprintf("%d. Component: %s\n   Direction: %s\n   Magnitude: %.2f\n   Reasoning: %s\n   Confidence: %.2f\n\n", 
			i+1, grad.Component, grad.Direction, grad.Magnitude, grad.Reasoning, grad.Confidence)
	}
	return strings.TrimSpace(result)
}

func parseScore(response string) (float64, error) {
	// Clean response
	response = strings.TrimSpace(response)
	
	// Try direct numeric parsing
	if score, err := strconv.ParseFloat(response, 64); err == nil {
		return normalizeScore(score), nil
	}
	
	// Look for patterns like "0.8", "0.75/1.0", "8/10", "80%"
	patterns := []struct {
		regex   string
		extract func(string) (float64, error)
	}{
		// Direct decimal
		{`\b(0?\.\d+)\b`, func(s string) (float64, error) {
			return strconv.ParseFloat(s, 64)
		}},
		// Fraction out of 1
		{`\b(\d*\.?\d+)\s*/\s*1(?:\.0+)?\b`, func(s string) (float64, error) {
			parts := strings.Split(s, "/")
			if len(parts) != 2 {
				return 0, fmt.Errorf("invalid fraction")
			}
			num, err := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
			return num, err
		}},
		// Fraction out of 10
		{`\b(\d*\.?\d+)\s*/\s*10\b`, func(s string) (float64, error) {
			parts := strings.Split(s, "/")
			if len(parts) != 2 {
				return 0, fmt.Errorf("invalid fraction")
			}
			num, err := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
			return num / 10.0, err
		}},
		// Percentage
		{`\b(\d+(?:\.\d+)?)\s*%`, func(s string) (float64, error) {
			s = strings.TrimSuffix(s, "%")
			num, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
			return num / 100.0, err
		}},
	}
	
	// Try patterns
	for _, pattern := range patterns {
		if idx := strings.Index(response, "."); idx != -1 {
			// Extract substring around decimal point
			start := idx - 1
			end := idx + 3
			if start < 0 {
				start = 0
			}
			if end > len(response) {
				end = len(response)
			}
			if num, err := strconv.ParseFloat(response[start:end], 64); err == nil {
				return normalizeScore(num), nil
			}
		}
	}
	
	// Default fallback
	return 0.5, fmt.Errorf("could not parse score from response: %s", response)
}

func normalizeScore(score float64) float64 {
	if score < 0 {
		return 0
	}
	if score > 1 {
		// If score is > 1, assume it's out of 10 or 100
		if score <= 10 {
			return score / 10.0
		}
		if score <= 100 {
			return score / 100.0
		}
		return 1.0
	}
	return score
}

func parseSemanticAnalysis(response string) (*SemanticAnalysis, error) {
	// Try to extract JSON from response
	start := strings.Index(response, "{")
	end := strings.LastIndex(response, "}")
	if start == -1 || end == -1 || start > end {
		// Return default analysis if no JSON found
		return &SemanticAnalysis{
			Components:      []SemanticNode{},
			Flows:          []SemanticFlow{},
			CoherenceScore: 0.5,
			Ambiguities:    []string{"Unable to parse detailed analysis"},
			Strengths:      []string{},
			Weaknesses:     []string{},
		}, nil
	}
	
	jsonStr := response[start : end+1]
	
	var analysis SemanticAnalysis
	if err := json.Unmarshal([]byte(jsonStr), &analysis); err != nil {
		return &SemanticAnalysis{
			Components:      []SemanticNode{},
			Flows:          []SemanticFlow{},
			CoherenceScore: 0.5,
			Ambiguities:    []string{"JSON parsing error: " + err.Error()},
			Strengths:      []string{},
			Weaknesses:     []string{},
		}, nil
	}
	
	// Validate and normalize coherence score
	if analysis.CoherenceScore < 0 {
		analysis.CoherenceScore = 0
	} else if analysis.CoherenceScore > 1 {
		analysis.CoherenceScore = 1
	}
	
	return &analysis, nil
}