package metaprompt

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/tmc/pe/internal/llm"
)

// MultiStageOptimizer orchestrates sequential prompt refinements
type MultiStageOptimizer struct {
	llm              llm.Provider
	textGradAnalyzer *TextGradAnalyzer
	gradientComputer *GradientComputer
}

// NewMultiStageOptimizer creates a new multi-stage optimizer
func NewMultiStageOptimizer(llmProvider llm.Provider) *MultiStageOptimizer {
	return &MultiStageOptimizer{
		llm:              llmProvider,
		textGradAnalyzer: NewTextGradAnalyzer(llmProvider),
		gradientComputer: NewGradientComputer(llmProvider),
	}
}

// OptimizationStage represents a single stage in the optimization process
type OptimizationStage struct {
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Method      string        `json:"method"`      // "analysis", "refinement", "validation", "polishing"
	Iterations  int           `json:"iterations"`
	Temperature float64       `json:"temperature"`
	Criteria    StageCriteria `json:"criteria"`
}

// StageCriteria defines success criteria for each stage
type StageCriteria struct {
	MinScore        float64 `json:"min_score"`         // Minimum score to pass this stage
	MaxIterations   int     `json:"max_iterations"`    // Maximum iterations allowed
	RequiredMetrics []string `json:"required_metrics"` // Metrics that must be satisfied
	GatingFunction  string  `json:"gating_function"`   // Function that determines if stage passes
}

// StageResult contains the result of a single optimization stage
type StageResult struct {
	Stage           OptimizationStage  `json:"stage"`
	InputPrompt     string             `json:"input_prompt"`
	OutputPrompt    string             `json:"output_prompt"`
	Iterations      []IterationResult  `json:"iterations"`
	StageScore      float64            `json:"stage_score"`
	PassedCriteria  bool               `json:"passed_criteria"`
	Duration        time.Duration      `json:"duration"`
	NextStageAdvice string             `json:"next_stage_advice"`
	Metrics         map[string]float64 `json:"metrics"`
}

// MultiStageResult contains the complete multi-stage optimization result
type MultiStageResult struct {
	OriginalPrompt  string        `json:"original_prompt"`
	FinalPrompt     string        `json:"final_prompt"`
	Stages          []StageResult `json:"stages"`
	OverallScore    float64       `json:"overall_score"`
	TotalDuration   time.Duration `json:"total_duration"`
	Success         bool          `json:"success"`
	FailureReason   string        `json:"failure_reason,omitempty"`
	Recommendations []string      `json:"recommendations"`
}

// OptimizeMultiStage performs progressive multi-stage optimization
func (mso *MultiStageOptimizer) OptimizeMultiStage(ctx context.Context, initialPrompt string, stages []OptimizationStage) (*MultiStageResult, error) {
	startTime := time.Now()
	
	result := &MultiStageResult{
		OriginalPrompt: initialPrompt,
		Stages:         make([]StageResult, 0, len(stages)),
		TotalDuration:  0,
		Success:        true,
	}

	currentPrompt := initialPrompt

	// Execute each stage sequentially
	for i, stage := range stages {
		stageResult, err := mso.executeStage(ctx, stage, currentPrompt, i)
		if err != nil {
			result.Success = false
			result.FailureReason = fmt.Sprintf("Stage %d (%s) failed: %v", i+1, stage.Name, err)
			break
		}

		result.Stages = append(result.Stages, *stageResult)
		
		// Check if stage passed its criteria
		if !stageResult.PassedCriteria {
			result.Success = false
			result.FailureReason = fmt.Sprintf("Stage %d (%s) failed to meet criteria", i+1, stage.Name)
			
			// Allow continuing to next stage if explicitly configured
			if stage.Criteria.GatingFunction != "allow_failure" {
				break
			}
		}

		// Update current prompt for next stage
		currentPrompt = stageResult.OutputPrompt
	}

	result.FinalPrompt = currentPrompt
	result.OverallScore = mso.calculateOverallScore(result.Stages)
	result.TotalDuration = time.Since(startTime)
	
	// Generate final recommendations
	recommendations, err := mso.generateFinalRecommendations(ctx, result)
	if err == nil {
		result.Recommendations = recommendations
	}

	return result, nil
}

// executeStage runs a single optimization stage
func (mso *MultiStageOptimizer) executeStage(ctx context.Context, stage OptimizationStage, inputPrompt string, stageIndex int) (*StageResult, error) {
	stageStart := time.Now()
	
	result := &StageResult{
		Stage:       stage,
		InputPrompt: inputPrompt,
		Iterations:  make([]IterationResult, 0),
		Metrics:     make(map[string]float64),
	}

	// Execute stage-specific optimization
	switch stage.Method {
	case "analysis":
		err := mso.executeAnalysisStage(ctx, stage, result)
		if err != nil {
			return nil, err
		}
	case "refinement":
		err := mso.executeRefinementStage(ctx, stage, result)
		if err != nil {
			return nil, err
		}
	case "validation":
		err := mso.executeValidationStage(ctx, stage, result)
		if err != nil {
			return nil, err
		}
	case "polishing":
		err := mso.executePolishingStage(ctx, stage, result)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unknown stage method: %s", stage.Method)
	}

	result.Duration = time.Since(stageStart)
	
	// Evaluate stage criteria
	result.PassedCriteria = mso.evaluateStageCriteria(stage.Criteria, result)
	
	// Generate advice for next stage
	advice, err := mso.generateNextStageAdvice(ctx, stage, result, stageIndex)
	if err == nil {
		result.NextStageAdvice = advice
	}

	return result, nil
}

// executeAnalysisStage performs deep analysis of the prompt
func (mso *MultiStageOptimizer) executeAnalysisStage(ctx context.Context, stage OptimizationStage, result *StageResult) error {
	// Use TextGrad analyzer for deep prompt analysis
	analysisResult, err := mso.textGradAnalyzer.AnalyzePromptGradients(ctx, result.InputPrompt, "")
	if err != nil {
		return fmt.Errorf("analysis failed: %w", err)
	}

	// Create analysis iteration
	iteration := IterationResult{
		Iteration:   1,
		Prompt:      result.InputPrompt,
		Score:       analysisResult.GradientStrength * 10, // Convert to 0-10 scale
		Feedback:    fmt.Sprintf("Analysis complete. Found %d optimization hints.", len(analysisResult.OptimizationHints)),
		Suggestions: analysisResult.OptimizationHints,
		Duration:    time.Millisecond * 100, // Placeholder
	}

	result.Iterations = append(result.Iterations, iteration)
	result.OutputPrompt = result.InputPrompt // Analysis doesn't change the prompt
	result.StageScore = iteration.Score

	// Store analysis metrics
	result.Metrics["gradient_strength"] = analysisResult.GradientStrength
	result.Metrics["coherence_avg"] = (analysisResult.CoherenceMetrics.LocalCoherence +
		analysisResult.CoherenceMetrics.GlobalCoherence +
		analysisResult.CoherenceMetrics.LogicalFlow +
		analysisResult.CoherenceMetrics.Consistency) / 4.0
	result.Metrics["semantic_drifts"] = float64(len(analysisResult.SemanticDrifts))

	return nil
}

// executeRefinementStage performs iterative prompt refinement
func (mso *MultiStageOptimizer) executeRefinementStage(ctx context.Context, stage OptimizationStage, result *StageResult) error {
	// Use standard optimization for refinement
	cfg := Config{
		InitialPrompt: result.InputPrompt,
		Iterations:    stage.Iterations,
		Temperature:   stage.Temperature,
		MaxTokens:     1000,
		Method:        "standard",
	}

	optimizer := NewOptimizer(mso.llm)
	optResult, err := optimizer.optimizeStandard(ctx, cfg)
	if err != nil {
		return fmt.Errorf("refinement failed: %w", err)
	}

	result.Iterations = optResult.Iterations
	result.OutputPrompt = optResult.OptimizedPrompt
	result.StageScore = optResult.ImprovementScore

	// Calculate refinement metrics
	if len(result.Iterations) > 0 {
		firstScore := result.Iterations[0].Score
		lastScore := result.Iterations[len(result.Iterations)-1].Score
		result.Metrics["score_improvement"] = lastScore - firstScore
		result.Metrics["iterations_used"] = float64(len(result.Iterations))
	}

	return nil
}

// executeValidationStage validates the optimized prompt
func (mso *MultiStageOptimizer) executeValidationStage(ctx context.Context, stage OptimizationStage, result *StageResult) error {
	// Generate test cases and validate prompt performance
	validationPrompt := fmt.Sprintf(`You are a prompt validation expert. Evaluate this optimized prompt across multiple test scenarios.

PROMPT TO VALIDATE:
%s

VALIDATION TASKS:
1. Generate 3 diverse test inputs for this prompt
2. Predict how well the prompt will handle each test case
3. Identify potential failure modes
4. Rate overall robustness on a scale of 1-10

FORMAT:
TEST CASES:
1. [Test case 1]
2. [Test case 2]  
3. [Test case 3]

PREDICTED PERFORMANCE:
1. [Rating 1-10]: [Explanation]
2. [Rating 1-10]: [Explanation]
3. [Rating 1-10]: [Explanation]

FAILURE MODES:
- [Potential failure 1]
- [Potential failure 2]

OVERALL ROBUSTNESS: [1-10]
RECOMMENDATIONS: [Validation recommendations]`, result.InputPrompt)

	options := llm.GenerateOptions{
		Temperature: &stage.Temperature,
		MaxTokens:   intPtr(1000),
	}

	resp, err := mso.llm.Generate(ctx, validationPrompt, options)
	if err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Parse validation score
	score := mso.parseValidationScore(resp.Text)
	
	iteration := IterationResult{
		Iteration:   1,
		Prompt:      result.InputPrompt,
		Score:       score,
		Feedback:    resp.Text,
		Suggestions: mso.parseValidationSuggestions(resp.Text),
		Duration:    time.Millisecond * 200,
	}

	result.Iterations = append(result.Iterations, iteration)
	result.OutputPrompt = result.InputPrompt // Validation doesn't change the prompt
	result.StageScore = score

	result.Metrics["robustness_score"] = score
	result.Metrics["test_cases_generated"] = 3.0

	return nil
}

// executePolishingStage performs final polishing of the prompt
func (mso *MultiStageOptimizer) executePolishingStage(ctx context.Context, stage OptimizationStage, result *StageResult) error {
	// Use TextGrad for final polishing
	cfg := Config{
		InitialPrompt: result.InputPrompt,
		Iterations:    stage.Iterations,
		Temperature:   stage.Temperature,
		MaxTokens:     1000,
		Method:        "textgrad",
	}

	textGradOpt := NewTextGradOptimizer(mso.llm)
	optResult, err := textGradOpt.OptimizeWithTextGrad(ctx, cfg)
	if err != nil {
		return fmt.Errorf("polishing failed: %w", err)
	}

	result.Iterations = optResult.Iterations
	result.OutputPrompt = optResult.OptimizedPrompt
	result.StageScore = optResult.ImprovementScore

	// Calculate polishing metrics
	result.Metrics["textgrad_score"] = optResult.ImprovementScore
	result.Metrics["polish_iterations"] = float64(len(result.Iterations))

	return nil
}

// evaluateStageCriteria checks if stage criteria are met
func (mso *MultiStageOptimizer) evaluateStageCriteria(criteria StageCriteria, result *StageResult) bool {
	// Check minimum score
	if result.StageScore < criteria.MinScore {
		return false
	}

	// Check maximum iterations
	if len(result.Iterations) > criteria.MaxIterations {
		return false
	}

	// Check required metrics
	for _, metricName := range criteria.RequiredMetrics {
		value, exists := result.Metrics[metricName]
		if !exists {
			return false
		}
		
		// Apply metric-specific thresholds
		switch metricName {
		case "gradient_strength":
			if value < 0.6 {
				return false
			}
		case "coherence_avg":
			if value < 0.7 {
				return false
			}
		case "robustness_score":
			if value < 7.0 {
				return false
			}
		}
	}

	return true
}

// calculateOverallScore computes the final optimization score
func (mso *MultiStageOptimizer) calculateOverallScore(stages []StageResult) float64 {
	if len(stages) == 0 {
		return 0.0
	}

	// Weighted average with higher weight for later stages
	var weightedSum, totalWeight float64
	for i, stage := range stages {
		weight := float64(i+1) / float64(len(stages)) // Later stages get more weight
		weightedSum += stage.StageScore * weight
		totalWeight += weight
	}

	return weightedSum / totalWeight
}

// generateNextStageAdvice provides guidance for the next stage
func (mso *MultiStageOptimizer) generateNextStageAdvice(ctx context.Context, currentStage OptimizationStage, result *StageResult, stageIndex int) (string, error) {
	advicePrompt := fmt.Sprintf(`Based on this optimization stage result, provide specific advice for the next stage.

CURRENT STAGE: %s (%s)
STAGE SCORE: %.2f
PASSED CRITERIA: %t

STAGE METRICS:
%s

TASK: Provide specific, actionable advice for optimizing the next stage. Focus on:
1. What worked well in this stage
2. What areas need improvement
3. Specific recommendations for the next stage
4. Potential risks to watch for

Keep advice concise and actionable.`, 
		currentStage.Name, 
		currentStage.Method,
		result.StageScore,
		result.PassedCriteria,
		mso.formatMetrics(result.Metrics))

	options := llm.GenerateOptions{
		Temperature: floatPtr(0.3),
		MaxTokens:   intPtr(300),
	}

	resp, err := mso.llm.Generate(ctx, advicePrompt, options)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(resp.Text), nil
}

// generateFinalRecommendations creates overall optimization recommendations
func (mso *MultiStageOptimizer) generateFinalRecommendations(ctx context.Context, result *MultiStageResult) ([]string, error) {
	recPrompt := fmt.Sprintf(`Based on this complete multi-stage optimization, provide final recommendations.

OPTIMIZATION SUMMARY:
- Original vs Final Score: %.2f improvement
- Stages Completed: %d
- Overall Success: %t
- Total Duration: %v

STAGE SUMMARY:
%s

TASK: Provide 3-5 final recommendations for:
1. How to use the optimized prompt effectively
2. Potential further improvements
3. Monitoring and maintenance advice

FORMAT:
1. [Recommendation 1]
2. [Recommendation 2]
3. [Recommendation 3]
4. [Recommendation 4]
5. [Recommendation 5]`,
		result.OverallScore,
		len(result.Stages),
		result.Success,
		result.TotalDuration,
		mso.formatStagesSummary(result.Stages))

	options := llm.GenerateOptions{
		Temperature: floatPtr(0.3),
		MaxTokens:   intPtr(500),
	}

	resp, err := mso.llm.Generate(ctx, recPrompt, options)
	if err != nil {
		return nil, err
	}

	return mso.parseRecommendations(resp.Text), nil
}

// Helper functions
func (mso *MultiStageOptimizer) parseValidationScore(text string) float64 {
	lines := strings.Split(text, "\n")
	for _, line := range lines {
		if strings.Contains(line, "OVERALL ROBUSTNESS:") {
			parts := strings.Split(line, ":")
			if len(parts) > 1 {
				scoreText := strings.TrimSpace(parts[1])
				if len(scoreText) > 0 && scoreText[0] >= '1' && scoreText[0] <= '9' {
					return float64(scoreText[0] - '0')
				}
			}
		}
	}
	return 5.0 // Default score
}

func (mso *MultiStageOptimizer) parseValidationSuggestions(text string) []string {
	lines := strings.Split(text, "\n")
	var suggestions []string
	inRecommendations := false
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "RECOMMENDATIONS:") {
			inRecommendations = true
			continue
		}
		if inRecommendations && strings.HasPrefix(line, "- ") {
			suggestion := strings.TrimSpace(line[2:])
			if suggestion != "" {
				suggestions = append(suggestions, suggestion)
			}
		}
	}
	
	return suggestions
}

func (mso *MultiStageOptimizer) formatMetrics(metrics map[string]float64) string {
	var formatted strings.Builder
	for key, value := range metrics {
		formatted.WriteString(fmt.Sprintf("- %s: %.3f\n", key, value))
	}
	return formatted.String()
}

func (mso *MultiStageOptimizer) formatStagesSummary(stages []StageResult) string {
	var formatted strings.Builder
	for i, stage := range stages {
		formatted.WriteString(fmt.Sprintf("Stage %d (%s): Score %.2f, Passed: %t\n",
			i+1, stage.Stage.Name, stage.StageScore, stage.PassedCriteria))
	}
	return formatted.String()
}

func (mso *MultiStageOptimizer) parseRecommendations(text string) []string {
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

// GetDefaultStages returns a sensible default stage configuration
func GetDefaultStages() []OptimizationStage {
	return []OptimizationStage{
		{
			Name:        "Analysis",
			Description: "Deep analysis of prompt structure and semantics",
			Method:      "analysis",
			Iterations:  1,
			Temperature: 0.2,
			Criteria: StageCriteria{
				MinScore:        6.0,
				MaxIterations:   1,
				RequiredMetrics: []string{"gradient_strength", "coherence_avg"},
				GatingFunction:  "require_pass",
			},
		},
		{
			Name:        "Refinement",
			Description: "Iterative prompt improvement",
			Method:      "refinement",
			Iterations:  3,
			Temperature: 0.3,
			Criteria: StageCriteria{
				MinScore:        7.0,
				MaxIterations:   5,
				RequiredMetrics: []string{"score_improvement"},
				GatingFunction:  "require_pass",
			},
		},
		{
			Name:        "Validation",
			Description: "Robustness testing and validation",
			Method:      "validation",
			Iterations:  1,
			Temperature: 0.2,
			Criteria: StageCriteria{
				MinScore:        7.5,
				MaxIterations:   1,
				RequiredMetrics: []string{"robustness_score"},
				GatingFunction:  "require_pass",
			},
		},
		{
			Name:        "Polishing",
			Description: "Final TextGrad-based polishing",
			Method:      "polishing",
			Iterations:  2,
			Temperature: 0.3,
			Criteria: StageCriteria{
				MinScore:        8.0,
				MaxIterations:   3,
				RequiredMetrics: []string{"textgrad_score"},
				GatingFunction:  "allow_failure", // Final polishing can fail
			},
		},
	}
}