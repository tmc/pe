package adapters

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/tmc/pe/internal/optimization"
)

// StandardEvaluator provides standard evaluation capabilities using a language model
type StandardEvaluator struct {
	provider optimization.LanguageModelProvider
}

// NewStandardEvaluator creates a new standard evaluator
func NewStandardEvaluator(provider optimization.LanguageModelProvider) optimization.Evaluator {
	return &StandardEvaluator{
		provider: provider,
	}
}

// EvaluatePrompt scores a prompt against an objective
func (e *StandardEvaluator) EvaluatePrompt(ctx context.Context, prompt string, objective string) (float64, error) {
	evalPrompt := fmt.Sprintf(`Evaluate this prompt against the given objective on a scale from 0.0 to 1.0:

PROMPT TO EVALUATE:
%s

OBJECTIVE:
%s

EVALUATION CRITERIA:
- Clarity and precision of instructions
- Alignment with the stated objective
- Completeness and specificity
- Likelihood of producing desired results
- Robustness against edge cases

Provide only a numeric score between 0.0 and 1.0, with no additional text.`, prompt, objective)

	options := optimization.GenerationOptions{
		Temperature: &[]float64{0.1}[0], // Low temperature for consistent scoring
		MaxTokens:   &[]int{10}[0],      // Short response expected
	}

	resp, err := e.provider.Generate(ctx, evalPrompt, options)
	if err != nil {
		return 0, fmt.Errorf("failed to evaluate prompt: %w", err)
	}

	// Parse score from response
	score, err := parseScore(resp.Text)
	if err != nil {
		return 0, fmt.Errorf("failed to parse evaluation score: %w", err)
	}

	// Ensure score is in valid range
	if score < 0.0 {
		score = 0.0
	} else if score > 1.0 {
		score = 1.0
	}

	return score, nil
}

// EvaluateComparative compares two prompts and returns which is better
func (e *StandardEvaluator) EvaluateComparative(ctx context.Context, prompt1, prompt2 string, objective string) (optimization.ComparisonResult, error) {
	evalPrompt := fmt.Sprintf(`Compare these two prompts against the given objective and determine which is better:

OBJECTIVE: %s

PROMPT A:
%s

PROMPT B:
%s

EVALUATION CRITERIA:
- Clarity and precision of instructions
- Alignment with the stated objective
- Completeness and specificity
- Likelihood of producing desired results
- Robustness against edge cases

Respond in this exact format:
WINNER: A|B|TIE
SCORE_A: [0.0-1.0]
SCORE_B: [0.0-1.0]
REASONING: [Brief explanation]`, objective, prompt1, prompt2)

	options := optimization.GenerationOptions{
		Temperature: &[]float64{0.2}[0],
		MaxTokens:   &[]int{200}[0],
	}

	resp, err := e.provider.Generate(ctx, evalPrompt, options)
	if err != nil {
		return optimization.ComparisonResult{}, fmt.Errorf("failed to compare prompts: %w", err)
	}

	// Parse comparison result
	return parseComparisonResult(resp.Text)
}

// EvaluateSystem evaluates a multi-component system
func (e *StandardEvaluator) EvaluateSystem(ctx context.Context, system optimization.SystemDefinition, objective string) (float64, error) {
	// Format system components for evaluation
	componentsText := formatSystemComponents(system.Components)
	dependenciesText := formatSystemDependencies(system.Dependencies)

	evalPrompt := fmt.Sprintf(`Evaluate this multi-component system against the given objective on a scale from 0.0 to 1.0:

SYSTEM: %s
OBJECTIVE: %s

COMPONENTS:
%s

DEPENDENCIES:
%s

EVALUATION CRITERIA:
- Component quality and coherence
- Inter-component synergy and integration
- Alignment with system objective
- Overall system effectiveness and robustness
- Scalability and maintainability

Provide only a numeric score between 0.0 and 1.0, with no additional text.`,
		system.Description, objective, componentsText, dependenciesText)

	options := optimization.GenerationOptions{
		Temperature: &[]float64{0.1}[0],
		MaxTokens:   &[]int{10}[0],
	}

	resp, err := e.provider.Generate(ctx, evalPrompt, options)
	if err != nil {
		return 0, fmt.Errorf("failed to evaluate system: %w", err)
	}

	// Parse score from response
	score, err := parseScore(resp.Text)
	if err != nil {
		return 0, fmt.Errorf("failed to parse system evaluation score: %w", err)
	}

	// Ensure score is in valid range
	if score < 0.0 {
		score = 0.0
	} else if score > 1.0 {
		score = 1.0
	}

	return score, nil
}

// Helper functions

// parseScore extracts a floating-point score from text
func parseScore(text string) (float64, error) {
	// Clean the text
	text = strings.TrimSpace(text)

	// Try to find a number in the text
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Try to parse as float directly
		if score, err := strconv.ParseFloat(line, 64); err == nil {
			return score, nil
		}

		// Try to extract number from line
		words := strings.Fields(line)
		for _, word := range words {
			if score, err := strconv.ParseFloat(word, 64); err == nil {
				return score, nil
			}
		}
	}

	return 0, fmt.Errorf("no valid score found in text: %s", text)
}

// parseComparisonResult parses a comparison result from evaluation text
func parseComparisonResult(text string) (optimization.ComparisonResult, error) {
	result := optimization.ComparisonResult{}

	lines := strings.Split(text, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "WINNER:") {
			winner := strings.TrimSpace(line[7:])
			switch strings.ToUpper(winner) {
			case "A":
				result.Winner = 1
			case "B":
				result.Winner = 2
			case "TIE":
				result.Winner = 0
			}
		} else if strings.HasPrefix(line, "SCORE_A:") {
			scoreText := strings.TrimSpace(line[8:])
			if score, err := strconv.ParseFloat(scoreText, 64); err == nil {
				result.Score1 = score
			}
		} else if strings.HasPrefix(line, "SCORE_B:") {
			scoreText := strings.TrimSpace(line[8:])
			if score, err := strconv.ParseFloat(scoreText, 64); err == nil {
				result.Score2 = score
			}
		} else if strings.HasPrefix(line, "REASONING:") {
			result.Reasoning = strings.TrimSpace(line[10:])
		}
	}

	// If no explicit winner was found, determine from scores
	if result.Winner == 0 && result.Score1 != result.Score2 {
		if result.Score1 > result.Score2 {
			result.Winner = 1
		} else {
			result.Winner = 2
		}
	}

	return result, nil
}

// formatSystemComponents formats system components for display
func formatSystemComponents(components []optimization.SystemComponent) string {
	var parts []string
	for _, comp := range components {
		parts = append(parts, fmt.Sprintf("- %s (%s): %s", comp.Name, comp.Type, comp.Content))
	}
	return strings.Join(parts, "\n")
}

// formatSystemDependencies formats system dependencies for display
func formatSystemDependencies(dependencies []optimization.ComponentDependency) string {
	var parts []string
	for _, dep := range dependencies {
		parts = append(parts, fmt.Sprintf("- %s -> %s (%s)", dep.From, dep.To, dep.Type))
	}
	return strings.Join(parts, "\n")
}
