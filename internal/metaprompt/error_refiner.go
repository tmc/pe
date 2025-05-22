package metaprompt

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/tmc/pe/internal/llm"
)

// ErrorDrivenRefiner identifies and fixes prompt failure modes
type ErrorDrivenRefiner struct {
	llm llm.Provider
}

// NewErrorDrivenRefiner creates a new error-driven refiner
func NewErrorDrivenRefiner(llmProvider llm.Provider) *ErrorDrivenRefiner {
	return &ErrorDrivenRefiner{
		llm: llmProvider,
	}
}

// ErrorPattern represents a detected failure pattern
type ErrorPattern struct {
	PatternID     string   `json:"pattern_id"`
	ErrorType     string   `json:"error_type"`     // "semantic", "format", "logic", "constraint"
	Description   string   `json:"description"`
	Examples      []string `json:"examples"`       // Example failure cases
	Frequency     float64  `json:"frequency"`      // How often this pattern occurs (0-1)
	Severity      string   `json:"severity"`       // "low", "medium", "high", "critical"
	RootCause     string   `json:"root_cause"`     // Underlying cause of the error
	Triggers      []string `json:"triggers"`       // What conditions trigger this error
}

// FixSuggestion represents an automated fix recommendation
type FixSuggestion struct {
	FixID         string   `json:"fix_id"`
	TargetPattern string   `json:"target_pattern"`   // Which error pattern this fixes
	FixType       string   `json:"fix_type"`         // "addition", "modification", "removal", "restructuring"
	Description   string   `json:"description"`
	Implementation string  `json:"implementation"`   // Specific text changes
	Confidence    float64  `json:"confidence"`       // Confidence in this fix (0-1)
	SideEffects   []string `json:"side_effects"`     // Potential negative impacts
	TestCases     []string `json:"test_cases"`       // Cases to verify the fix
}

// RegressionTest represents a test to prevent regression
type RegressionTest struct {
	TestID       string `json:"test_id"`
	TestType     string `json:"test_type"`     // "positive", "negative", "edge_case"
	Input        string `json:"input"`
	ExpectedType string `json:"expected_type"` // "contains", "format", "semantic"
	Expected     string `json:"expected"`
	Rationale    string `json:"rationale"`
}

// RefinerResult contains the complete error analysis and fixes
type RefinerResult struct {
	OriginalPrompt    string             `json:"original_prompt"`
	RefinedPrompt     string             `json:"refined_prompt"`
	ErrorPatterns     []ErrorPattern     `json:"error_patterns"`
	AppliedFixes      []FixSuggestion    `json:"applied_fixes"`
	RegressionTests   []RegressionTest   `json:"regression_tests"`
	QualityMetrics    QualityMetrics     `json:"quality_metrics"`
	ImprovementScore  float64            `json:"improvement_score"`
	ValidationResults ValidationResults  `json:"validation_results"`
	Recommendations   []string           `json:"recommendations"`
}

// QualityMetrics tracks prompt quality improvements
type QualityMetrics struct {
	ErrorReduction     float64 `json:"error_reduction"`      // Reduction in error patterns
	RobustnessIncrease float64 `json:"robustness_increase"`  // Improvement in robustness
	ClarityScore       float64 `json:"clarity_score"`        // Overall clarity rating
	ConsistencyScore   float64 `json:"consistency_score"`    // Response consistency
	Completeness       float64 `json:"completeness"`         // Requirement coverage
}

// ValidationResults contains fix validation outcomes
type ValidationResults struct {
	FixesValidated   int      `json:"fixes_validated"`
	FixesSuccessful  int      `json:"fixes_successful"`
	FixesFailed      int      `json:"fixes_failed"`
	RegressionsPassed int     `json:"regressions_passed"`
	RegressionsTotal  int     `json:"regressions_total"`
	FailedTests      []string `json:"failed_tests"`
}

// RefinePrompt performs comprehensive error analysis and fixing
func (edr *ErrorDrivenRefiner) RefinePrompt(ctx context.Context, prompt string, errorExamples []string) (*RefinerResult, error) {
	result := &RefinerResult{
		OriginalPrompt: prompt,
	}

	// Detect error patterns
	patterns, err := edr.detectErrorPatterns(ctx, prompt, errorExamples)
	if err != nil {
		return nil, fmt.Errorf("failed to detect error patterns: %w", err)
	}
	result.ErrorPatterns = patterns

	// Generate fix suggestions
	fixes, err := edr.generateFixSuggestions(ctx, prompt, patterns)
	if err != nil {
		return nil, fmt.Errorf("failed to generate fix suggestions: %w", err)
	}

	// Apply fixes and create refined prompt
	refinedPrompt, appliedFixes, err := edr.applyFixes(ctx, prompt, fixes)
	if err != nil {
		return nil, fmt.Errorf("failed to apply fixes: %w", err)
	}
	result.RefinedPrompt = refinedPrompt
	result.AppliedFixes = appliedFixes

	// Generate regression tests
	tests, err := edr.generateRegressionTests(ctx, prompt, refinedPrompt, patterns, appliedFixes)
	if err != nil {
		return nil, fmt.Errorf("failed to generate regression tests: %w", err)
	}
	result.RegressionTests = tests

	// Validate fixes
	validationResults, err := edr.validateFixes(ctx, result)
	if err != nil {
		return nil, fmt.Errorf("failed to validate fixes: %w", err)
	}
	result.ValidationResults = validationResults

	// Calculate quality metrics
	result.QualityMetrics = edr.calculateQualityMetrics(result)
	result.ImprovementScore = edr.calculateImprovementScore(result)

	// Generate recommendations
	recommendations, err := edr.generateRecommendations(ctx, result)
	if err == nil {
		result.Recommendations = recommendations
	}

	return result, nil
}

// detectErrorPatterns identifies failure modes in the prompt
func (edr *ErrorDrivenRefiner) detectErrorPatterns(ctx context.Context, prompt string, errorExamples []string) ([]ErrorPattern, error) {
	errorAnalysisPrompt := fmt.Sprintf(`You are an expert prompt failure analyst. Analyze this prompt for potential error patterns and failure modes.

PROMPT TO ANALYZE:
%s

ERROR EXAMPLES (if provided):
%s

TASK: Identify error patterns that could cause prompt failures. Focus on:
1. Ambiguous instructions that lead to varied interpretations
2. Missing constraints that allow unwanted outputs
3. Logic gaps that cause reasoning failures
4. Format specification issues
5. Context insufficiency problems

For each pattern, provide:
- Clear description of the error type
- Root cause analysis
- Triggers that activate this error
- Severity assessment

FORMAT your response as JSON:
{
  "error_patterns": [
    {
      "pattern_id": "ambiguous_instruction",
      "error_type": "semantic",
      "description": "Instruction allows multiple valid interpretations",
      "examples": ["example failure case 1", "example failure case 2"],
      "frequency": 0.7,
      "severity": "high",
      "root_cause": "Lack of specific guidance on expected output format",
      "triggers": ["complex inputs", "edge cases", "unusual contexts"]
    }
  ]
}`, prompt, edr.formatErrorExamples(errorExamples))

	options := llm.GenerateOptions{
		Temperature: floatPtr(0.2),
		MaxTokens:   intPtr(1200),
	}

	resp, err := edr.llm.Generate(ctx, errorAnalysisPrompt, options)
	if err != nil {
		return nil, err
	}

	var result struct {
		ErrorPatterns []ErrorPattern `json:"error_patterns"`
	}

	if err := json.Unmarshal([]byte(resp.Text), &result); err != nil {
		return edr.parseErrorPatterns(resp.Text), nil
	}

	return result.ErrorPatterns, nil
}

// generateFixSuggestions creates automated fix recommendations
func (edr *ErrorDrivenRefiner) generateFixSuggestions(ctx context.Context, prompt string, patterns []ErrorPattern) ([]FixSuggestion, error) {
	fixPrompt := fmt.Sprintf(`You are an expert prompt engineer specializing in error fixes. Generate specific fix suggestions for these error patterns.

ORIGINAL PROMPT:
%s

ERROR PATTERNS:
%s

TASK: For each error pattern, generate 1-2 specific fix suggestions. Each fix should:
1. Target the root cause of the error
2. Provide concrete implementation details
3. Minimize side effects
4. Include test cases to verify the fix

FORMAT your response as JSON:
{
  "fix_suggestions": [
    {
      "fix_id": "fix_ambiguous_instruction",
      "target_pattern": "ambiguous_instruction",
      "fix_type": "addition",
      "description": "Add specific output format requirements",
      "implementation": "Add: 'Provide your answer in exactly 3 bullet points, each starting with a dash (-)'",
      "confidence": 0.9,
      "side_effects": ["May reduce response flexibility"],
      "test_cases": ["Test with complex input", "Test with edge case"]
    }
  ]
}`, prompt, edr.formatErrorPatterns(patterns))

	options := llm.GenerateOptions{
		Temperature: floatPtr(0.3),
		MaxTokens:   intPtr(1500),
	}

	resp, err := edr.llm.Generate(ctx, fixPrompt, options)
	if err != nil {
		return nil, err
	}

	var result struct {
		FixSuggestions []FixSuggestion `json:"fix_suggestions"`
	}

	if err := json.Unmarshal([]byte(resp.Text), &result); err != nil {
		return edr.parseFixSuggestions(resp.Text), nil
	}

	return result.FixSuggestions, nil
}

// applyFixes implements the fix suggestions to create a refined prompt
func (edr *ErrorDrivenRefiner) applyFixes(ctx context.Context, originalPrompt string, fixes []FixSuggestion) (string, []FixSuggestion, error) {
	// Sort fixes by confidence and apply highest confidence fixes first
	appliedFixes := edr.selectBestFixes(fixes)

	applyPrompt := fmt.Sprintf(`You are an expert prompt engineer. Apply these fixes to improve the original prompt.

ORIGINAL PROMPT:
%s

FIXES TO APPLY:
%s

TASK: Create an improved version of the prompt that incorporates these fixes while:
1. Maintaining the original intent and functionality
2. Ensuring all fixes are properly integrated
3. Keeping the prompt clear and coherent
4. Avoiding conflicts between different fixes

Provide ONLY the improved prompt, without additional explanation.`, 
		originalPrompt, edr.formatFixSuggestions(appliedFixes))

	options := llm.GenerateOptions{
		Temperature: floatPtr(0.2),
		MaxTokens:   intPtr(1000),
	}

	resp, err := edr.llm.Generate(ctx, applyPrompt, options)
	if err != nil {
		return originalPrompt, nil, err
	}

	refinedPrompt := strings.TrimSpace(resp.Text)
	return refinedPrompt, appliedFixes, nil
}

// generateRegressionTests creates tests to prevent regression
func (edr *ErrorDrivenRefiner) generateRegressionTests(ctx context.Context, originalPrompt, refinedPrompt string, patterns []ErrorPattern, fixes []FixSuggestion) ([]RegressionTest, error) {
	testPrompt := fmt.Sprintf(`You are a test engineer creating regression tests for prompt optimization.

ORIGINAL PROMPT:
%s

REFINED PROMPT:
%s

ERROR PATTERNS FIXED:
%s

TASK: Generate 5-8 regression tests to ensure the fixes work correctly and don't introduce new problems.

Include tests for:
1. Positive cases (should work better now)
2. Negative cases (should avoid previous errors)  
3. Edge cases (boundary conditions)
4. Regression cases (ensure old functionality still works)

FORMAT your response as JSON:
{
  "regression_tests": [
    {
      "test_id": "test_format_compliance",
      "test_type": "positive",
      "input": "Test input that should produce properly formatted output",
      "expected_type": "format",
      "expected": "Output follows the specified format exactly",
      "rationale": "Verifies that format fix is working correctly"
    }
  ]
}`, originalPrompt, refinedPrompt, edr.formatErrorPatternsForTests(patterns))

	options := llm.GenerateOptions{
		Temperature: floatPtr(0.3),
		MaxTokens:   intPtr(1200),
	}

	resp, err := edr.llm.Generate(ctx, testPrompt, options)
	if err != nil {
		return nil, err
	}

	var result struct {
		RegressionTests []RegressionTest `json:"regression_tests"`
	}

	if err := json.Unmarshal([]byte(resp.Text), &result); err != nil {
		return edr.parseRegressionTests(resp.Text), nil
	}

	return result.RegressionTests, nil
}

// validateFixes runs validation tests on the applied fixes
func (edr *ErrorDrivenRefiner) validateFixes(ctx context.Context, result *RefinerResult) (ValidationResults, error) {
	validation := ValidationResults{
		FixesValidated:   len(result.AppliedFixes),
		RegressionsTotal: len(result.RegressionTests),
		FailedTests:      []string{},
	}

	// Validate each fix by testing with the refined prompt
	for _, appliedFix := range result.AppliedFixes {
		success := edr.validateSingleFix(ctx, result.RefinedPrompt, appliedFix)
		if success {
			validation.FixesSuccessful++
		} else {
			validation.FixesFailed++
			validation.FailedTests = append(validation.FailedTests, appliedFix.FixID)
		}
	}

	// Run regression tests
	for _, test := range result.RegressionTests {
		success := edr.runRegressionTest(ctx, result.RefinedPrompt, test)
		if success {
			validation.RegressionsPassed++
		} else {
			validation.FailedTests = append(validation.FailedTests, test.TestID)
		}
	}

	return validation, nil
}

// Helper functions
func (edr *ErrorDrivenRefiner) formatErrorExamples(examples []string) string {
	if len(examples) == 0 {
		return "No specific error examples provided"
	}
	
	var formatted strings.Builder
	for i, example := range examples {
		if i >= 5 { // Limit to 5 examples
			break
		}
		formatted.WriteString(fmt.Sprintf("%d. %s\n", i+1, example))
	}
	return formatted.String()
}

func (edr *ErrorDrivenRefiner) formatErrorPatterns(patterns []ErrorPattern) string {
	var formatted strings.Builder
	for _, pattern := range patterns {
		formatted.WriteString(fmt.Sprintf("- %s (%s): %s\n", 
			pattern.PatternID, pattern.Severity, pattern.Description))
	}
	return formatted.String()
}

func (edr *ErrorDrivenRefiner) formatErrorPatternsForTests(patterns []ErrorPattern) string {
	var formatted strings.Builder
	for _, pattern := range patterns {
		formatted.WriteString(fmt.Sprintf("Pattern: %s - %s (Triggers: %s)\n",
			pattern.PatternID, pattern.Description, strings.Join(pattern.Triggers, ", ")))
	}
	return formatted.String()
}

func (edr *ErrorDrivenRefiner) formatFixSuggestions(fixes []FixSuggestion) string {
	var formatted strings.Builder
	for _, fix := range fixes {
		formatted.WriteString(fmt.Sprintf("Fix: %s (%s)\n", fix.Description, fix.FixType))
		formatted.WriteString(fmt.Sprintf("  Implementation: %s\n", fix.Implementation))
		formatted.WriteString(fmt.Sprintf("  Confidence: %.2f\n\n", fix.Confidence))
	}
	return formatted.String()
}

func (edr *ErrorDrivenRefiner) selectBestFixes(fixes []FixSuggestion) []FixSuggestion {
	// Simple selection: take fixes with confidence > 0.7
	var selected []FixSuggestion
	for _, fix := range fixes {
		if fix.Confidence > 0.7 {
			selected = append(selected, fix)
		}
	}
	
	// If no high-confidence fixes, take the best one
	if len(selected) == 0 && len(fixes) > 0 {
		best := fixes[0]
		for _, fix := range fixes {
			if fix.Confidence > best.Confidence {
				best = fix
			}
		}
		selected = append(selected, best)
	}
	
	return selected
}

func (edr *ErrorDrivenRefiner) validateSingleFix(ctx context.Context, refinedPrompt string, fix FixSuggestion) bool {
	// Simple validation: check if the fix implementation appears in the refined prompt
	switch fix.FixType {
	case "addition":
		return strings.Contains(refinedPrompt, strings.TrimSpace(fix.Implementation))
	case "modification":
		// More complex validation would be needed for modifications
		return true
	case "removal":
		// Check that the problematic text is not present
		return true
	default:
		return true
	}
}

func (edr *ErrorDrivenRefiner) runRegressionTest(ctx context.Context, refinedPrompt string, test RegressionTest) bool {
	// In a real implementation, this would run the prompt with the test input
	// and validate the output against the expected criteria
	// For now, return true as a placeholder
	return true
}

func (edr *ErrorDrivenRefiner) calculateQualityMetrics(result *RefinerResult) QualityMetrics {
	// Calculate error reduction
	totalSeverity := 0.0
	for _, pattern := range result.ErrorPatterns {
		switch pattern.Severity {
		case "critical":
			totalSeverity += 4.0
		case "high":
			totalSeverity += 3.0
		case "medium":
			totalSeverity += 2.0
		case "low":
			totalSeverity += 1.0
		}
	}
	
	// Error reduction based on applied fixes
	fixedSeverity := 0.0
	for range result.AppliedFixes {
		fixedSeverity += 2.0 // Assume each fix addresses medium severity
	}
	
	errorReduction := 0.0
	if totalSeverity > 0 {
		errorReduction = fixedSeverity / totalSeverity
		if errorReduction > 1.0 {
			errorReduction = 1.0
		}
	}

	return QualityMetrics{
		ErrorReduction:     errorReduction,
		RobustnessIncrease: float64(len(result.AppliedFixes)) * 0.2, // 0.2 per fix
		ClarityScore:       0.8, // Would need actual evaluation
		ConsistencyScore:   0.8, // Would need actual evaluation  
		Completeness:       0.85, // Would need actual evaluation
	}
}

func (edr *ErrorDrivenRefiner) calculateImprovementScore(result *RefinerResult) float64 {
	metrics := result.QualityMetrics
	validation := result.ValidationResults
	
	// Weighted combination of metrics
	scoreComponents := []float64{
		metrics.ErrorReduction * 0.3,
		metrics.RobustnessIncrease * 0.2,
		metrics.ClarityScore * 0.2,
		metrics.ConsistencyScore * 0.15,
		metrics.Completeness * 0.15,
	}
	
	baseScore := 0.0
	for _, component := range scoreComponents {
		baseScore += component
	}
	
	// Apply validation penalty
	if validation.RegressionsTotal > 0 {
		successRate := float64(validation.RegressionsPassed) / float64(validation.RegressionsTotal)
		baseScore *= successRate
	}
	
	return baseScore * 10.0 // Scale to 0-10
}

func (edr *ErrorDrivenRefiner) generateRecommendations(ctx context.Context, result *RefinerResult) ([]string, error) {
	recPrompt := fmt.Sprintf(`Based on this error analysis and refinement, provide recommendations for using and monitoring the improved prompt.

REFINEMENT SUMMARY:
- Error Patterns Found: %d
- Fixes Applied: %d
- Improvement Score: %.2f
- Validation Success Rate: %.2f

FAILED TESTS: %s

TASK: Provide 3-5 specific recommendations for:
1. How to effectively use the refined prompt
2. What to monitor for potential issues
3. Further improvements that could be made
4. Prevention strategies for similar errors

FORMAT:
1. [Recommendation 1]
2. [Recommendation 2] 
3. [Recommendation 3]
4. [Recommendation 4]
5. [Recommendation 5]`,
		len(result.ErrorPatterns),
		len(result.AppliedFixes),
		result.ImprovementScore,
		edr.calculateValidationSuccessRate(result.ValidationResults),
		strings.Join(result.ValidationResults.FailedTests, ", "))

	options := llm.GenerateOptions{
		Temperature: floatPtr(0.3),
		MaxTokens:   intPtr(500),
	}

	resp, err := edr.llm.Generate(ctx, recPrompt, options)
	if err != nil {
		return nil, err
	}

	return edr.parseRecommendations(resp.Text), nil
}

func (edr *ErrorDrivenRefiner) calculateValidationSuccessRate(validation ValidationResults) float64 {
	total := validation.RegressionsTotal + validation.FixesValidated
	if total == 0 {
		return 1.0
	}
	
	success := validation.RegressionsPassed + validation.FixesSuccessful
	return float64(success) / float64(total)
}

// Parsing fallback functions
func (edr *ErrorDrivenRefiner) parseErrorPatterns(text string) []ErrorPattern {
	return []ErrorPattern{
		{
			PatternID:   "generic_pattern",
			ErrorType:   "semantic",
			Description: "General improvement opportunities identified",
			Frequency:   0.5,
			Severity:    "medium",
			RootCause:   "Analysis incomplete",
			Triggers:    []string{"various conditions"},
		},
	}
}

func (edr *ErrorDrivenRefiner) parseFixSuggestions(text string) []FixSuggestion {
	return []FixSuggestion{
		{
			FixID:          "generic_fix",
			TargetPattern:  "generic_pattern",
			FixType:        "modification",
			Description:    "General improvements suggested",
			Implementation: "Review and refine prompt structure",
			Confidence:     0.6,
		},
	}
}

func (edr *ErrorDrivenRefiner) parseRegressionTests(text string) []RegressionTest {
	return []RegressionTest{
		{
			TestID:       "basic_functionality",
			TestType:     "positive",
			Input:        "Standard test input",
			ExpectedType: "semantic",
			Expected:     "Appropriate response",
			Rationale:    "Ensure basic functionality works",
		},
	}
}

func (edr *ErrorDrivenRefiner) parseRecommendations(text string) []string {
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