package metaprompt

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/tmc/pe/internal/llm"
)

// Config contains configuration for prompt optimization
type Config struct {
	InitialPrompt        string
	Iterations           int
	MaxIterations        int     // Maximum iterations for methods that use it
	Temperature          float64
	MaxTokens            int
	UseTextGrad          bool    // Enable TextGrad-style optimization
	Method               string  // "standard", "textgrad", "hybrid"
	Objective            string  // Optimization objective
	ConvergenceThreshold float64 // Threshold for convergence
}

// OptimizationResult contains the results of prompt optimization
type OptimizationResult struct {
	OriginalPrompt    string             `json:"original_prompt"`
	OptimizedPrompt   string             `json:"optimized_prompt"`
	Iterations        []IterationResult  `json:"iterations"`
	ImprovementScore  float64            `json:"improvement_score"`
	TotalDuration     time.Duration      `json:"total_duration"`
	CreatedAt         time.Time          `json:"created_at"`
}

// IterationResult contains the result of a single optimization iteration
type IterationResult struct {
	Iteration     int           `json:"iteration"`
	Prompt        string        `json:"prompt"`
	Score         float64       `json:"score"`
	Feedback      string        `json:"feedback"`
	Suggestions   []string      `json:"suggestions"`
	Duration      time.Duration `json:"duration"`
	Changes       []string      `json:"changes,omitempty"`
	Timestamp     time.Time     `json:"timestamp,omitempty"`
}

// Optimizer implements prompt optimization using metaprompting techniques
type Optimizer struct {
	llm      llm.Provider
	textGrad *TextGradOptimizer
	pe2      *PE2Optimizer
	apex     *APEXOptimizer
}

// NewOptimizer creates a new prompt optimizer
func NewOptimizer(llmProvider llm.Provider) *Optimizer {
	return &Optimizer{
		llm:      llmProvider,
		textGrad: NewTextGradOptimizer(llmProvider),
		pe2:      NewPE2Optimizer(llmProvider),
		apex:     NewAPEXOptimizer(llmProvider),
	}
}

// Optimize runs the prompt optimization process
func (o *Optimizer) Optimize(ctx context.Context, cfg Config) (*OptimizationResult, error) {
	// Choose optimization method
	switch cfg.Method {
	case "pe2":
		return o.pe2.OptimizeWithPE2(ctx, cfg)
	case "apex":
		return o.apex.OptimizeWithAPEX(ctx, cfg)
	case "textgrad":
		return o.textGrad.OptimizeWithTextGrad(ctx, cfg)
	case "hybrid":
		return o.optimizeHybrid(ctx, cfg)
	default:
		return o.optimizeStandard(ctx, cfg)
	}
}

// optimizeStandard runs the standard prompt optimization process
func (o *Optimizer) optimizeStandard(ctx context.Context, cfg Config) (*OptimizationResult, error) {
	startTime := time.Now()
	
	result := &OptimizationResult{
		OriginalPrompt: cfg.InitialPrompt,
		Iterations:     make([]IterationResult, 0, cfg.Iterations),
		CreatedAt:      startTime,
	}

	currentPrompt := cfg.InitialPrompt

	for i := 0; i < cfg.Iterations; i++ {
		iterStart := time.Now()
		
		// Generate analysis and suggestions
		_, suggestions, err := o.generateOptimizationSuggestions(ctx, currentPrompt, cfg)
		if err != nil {
			return nil, fmt.Errorf("iteration %d failed: %w", i+1, err)
		}

		// Select best suggestion and score it
		bestPrompt, score, feedback := o.selectBestImprovement(ctx, currentPrompt, suggestions, cfg)
		
		iteration := IterationResult{
			Iteration:   i + 1,
			Prompt:      bestPrompt,
			Score:       score,
			Feedback:    feedback,
			Suggestions: suggestions,
			Duration:    time.Since(iterStart),
		}
		
		result.Iterations = append(result.Iterations, iteration)
		
		// Update current prompt for next iteration
		currentPrompt = bestPrompt
	}

	result.OptimizedPrompt = currentPrompt
	result.ImprovementScore = o.calculateImprovementScore(result)
	result.TotalDuration = time.Since(startTime)

	return result, nil
}

// optimizeHybrid combines standard and TextGrad approaches
func (o *Optimizer) optimizeHybrid(ctx context.Context, cfg Config) (*OptimizationResult, error) {
	// Use standard optimization for first half of iterations
	standardCfg := cfg
	standardCfg.Iterations = cfg.Iterations / 2
	standardCfg.Method = "standard"
	
	standardResult, err := o.optimizeStandard(ctx, standardCfg)
	if err != nil {
		return nil, fmt.Errorf("standard optimization phase failed: %w", err)
	}

	// Use TextGrad for second half, starting with standard result
	textgradCfg := cfg
	textgradCfg.InitialPrompt = standardResult.OptimizedPrompt
	textgradCfg.Iterations = cfg.Iterations - standardCfg.Iterations
	textgradCfg.Method = "textgrad"
	
	textgradResult, err := o.textGrad.OptimizeWithTextGrad(ctx, textgradCfg)
	if err != nil {
		return nil, fmt.Errorf("textgrad optimization phase failed: %w", err)
	}

	// Combine results
	combinedResult := &OptimizationResult{
		OriginalPrompt:   standardResult.OriginalPrompt,
		OptimizedPrompt:  textgradResult.OptimizedPrompt,
		Iterations:       append(standardResult.Iterations, textgradResult.Iterations...),
		ImprovementScore: (standardResult.ImprovementScore + textgradResult.ImprovementScore) / 2,
		TotalDuration:    standardResult.TotalDuration + textgradResult.TotalDuration,
		CreatedAt:        standardResult.CreatedAt,
	}

	return combinedResult, nil
}

// generateOptimizationSuggestions creates suggestions for improving the prompt
func (o *Optimizer) generateOptimizationSuggestions(ctx context.Context, currentPrompt string, cfg Config) (string, []string, error) {
	metaPrompt := fmt.Sprintf(`You are an expert prompt engineer specializing in prompt optimization. 

Analyze this prompt and provide specific, actionable improvements:

CURRENT PROMPT:
---
%s
---

ANALYSIS FRAMEWORK:
1. Clarity: Is the instruction clear and unambiguous?
2. Specificity: Are requirements and constraints well-defined?
3. Context: Is sufficient context provided?
4. Structure: Is the prompt well-organized?
5. Error Prevention: Does it prevent common failure modes?

TASK:
1. First, provide a detailed analysis of the current prompt's strengths and weaknesses
2. Then suggest 3 improved versions that address the identified issues
3. Each suggestion should focus on a different improvement aspect

FORMAT YOUR RESPONSE AS:
ANALYSIS:
[Detailed analysis of current prompt]

SUGGESTION 1: [Focus area]
[Improved prompt version 1]

SUGGESTION 2: [Focus area]  
[Improved prompt version 2]

SUGGESTION 3: [Focus area]
[Improved prompt version 3]`, currentPrompt)

	options := llm.GenerateOptions{
		Temperature: &cfg.Temperature,
		MaxTokens:   &cfg.MaxTokens,
	}

	response, err := o.llm.Generate(ctx, metaPrompt, options)
	if err != nil {
		return "", nil, fmt.Errorf("failed to generate suggestions: %w", err)
	}

	analysis, suggestions := o.parseOptimizationResponse(response.Text)
	return analysis, suggestions, nil
}

// selectBestImprovement evaluates suggestions and selects the best one
func (o *Optimizer) selectBestImprovement(ctx context.Context, originalPrompt string, suggestions []string, cfg Config) (string, float64, string) {
	if len(suggestions) == 0 {
		return originalPrompt, 0.0, "No suggestions generated"
	}

	// Use LLM to evaluate and rank suggestions
	evaluationPrompt := fmt.Sprintf(`You are an expert prompt evaluator. Compare these prompt versions and select the best one.

ORIGINAL PROMPT:
%s

CANDIDATE IMPROVEMENTS:
%s

EVALUATION CRITERIA:
- Clarity and precision of instructions
- Completeness of context and constraints  
- Robustness against edge cases
- Likelihood of producing desired outputs

TASK:
1. Evaluate each candidate against the criteria
2. Select the best improvement
3. Provide a score from 1-10 for the selected prompt
4. Explain your reasoning

FORMAT:
SELECTED: [number of best candidate]
SCORE: [1-10]
REASONING: [explanation of why this is the best choice]`, originalPrompt, o.formatSuggestions(suggestions))

	options := llm.GenerateOptions{
		Temperature: &cfg.Temperature,
		MaxTokens:   &cfg.MaxTokens,
	}

	response, err := o.llm.Generate(ctx, evaluationPrompt, options)
	if err != nil {
		// Fallback: return first suggestion with default score
		return suggestions[0], 5.0, "Evaluation failed, using first suggestion"
	}

	selectedIndex, score, reasoning := o.parseEvaluationResponse(response.Text)
	
	// Validate selection
	if selectedIndex < 0 || selectedIndex >= len(suggestions) {
		selectedIndex = 0
	}

	return suggestions[selectedIndex], score, reasoning
}

// parseOptimizationResponse extracts analysis and suggestions from the response
func (o *Optimizer) parseOptimizationResponse(response string) (string, []string) {
	parts := strings.Split(response, "SUGGESTION")
	
	var analysis string
	var suggestions []string

	if len(parts) > 0 {
		analysisSection := parts[0]
		if idx := strings.Index(analysisSection, "ANALYSIS:"); idx != -1 {
			analysis = strings.TrimSpace(analysisSection[idx+9:])
		}
	}

	for i := 1; i < len(parts); i++ {
		suggestion := strings.TrimSpace(parts[i])
		// Extract the actual prompt text after the focus area line
		lines := strings.Split(suggestion, "\n")
		if len(lines) > 1 {
			suggestionText := strings.Join(lines[1:], "\n")
			suggestions = append(suggestions, strings.TrimSpace(suggestionText))
		}
	}

	return analysis, suggestions
}

// parseEvaluationResponse extracts the selected index, score, and reasoning
func (o *Optimizer) parseEvaluationResponse(response string) (int, float64, string) {
	lines := strings.Split(response, "\n")
	
	selectedIndex := 0
	score := 5.0
	reasoning := "Default evaluation"

	for _, line := range lines {
		line = strings.TrimSpace(line)
		
		if strings.HasPrefix(line, "SELECTED:") {
			selectedText := strings.TrimSpace(line[9:])
			// Extract number from text like "1", "candidate 1", etc.
			if len(selectedText) > 0 {
				if selectedText[0] >= '1' && selectedText[0] <= '9' {
					selectedIndex = int(selectedText[0] - '1') // Convert to 0-based index
				}
			}
		} else if strings.HasPrefix(line, "SCORE:") {
			scoreText := strings.TrimSpace(line[6:])
			if len(scoreText) > 0 && scoreText[0] >= '1' && scoreText[0] <= '9' {
				score = float64(scoreText[0] - '0')
			}
		} else if strings.HasPrefix(line, "REASONING:") {
			reasoning = strings.TrimSpace(line[10:])
		}
	}

	return selectedIndex, score, reasoning
}

// formatSuggestions formats suggestions for display in evaluation prompt
func (o *Optimizer) formatSuggestions(suggestions []string) string {
	var formatted strings.Builder
	for i, suggestion := range suggestions {
		formatted.WriteString(fmt.Sprintf("CANDIDATE %d:\n%s\n\n", i+1, suggestion))
	}
	return formatted.String()
}

// calculateImprovementScore calculates overall improvement score
func (o *Optimizer) calculateImprovementScore(result *OptimizationResult) float64 {
	if len(result.Iterations) == 0 {
		return 0.0
	}
	
	// Use the final iteration's score
	finalScore := result.Iterations[len(result.Iterations)-1].Score
	
	// Apply bonus for consistency across iterations
	var totalScore float64
	for _, iter := range result.Iterations {
		totalScore += iter.Score
	}
	avgScore := totalScore / float64(len(result.Iterations))
	
	// Weighted combination of final score and average
	return (finalScore * 0.7) + (avgScore * 0.3)
}

// SaveToFile saves the optimization result to a JSON file
func (r *OptimizationResult) SaveToFile(path string) error {
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal result: %w", err)
	}
	
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}
	
	return nil
}