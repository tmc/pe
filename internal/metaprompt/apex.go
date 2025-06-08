package metaprompt

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"sort"
	"strings"
	"time"

	"github.com/tmc/pe/internal/llm"
)

// APEXOptimizer implements APEX (Automated Prompt Engineering Xpert)
// Based on 2024 research: "Automatic Engineering of Long Prompts"
// APEX uses greedy algorithms with beam-search for efficiency and
// leverages search history to significantly enhance LLM-based mutation
type APEXOptimizer struct {
	llm llm.Provider
}

// APEXConfig contains configuration for APEX optimization
type APEXConfig struct {
	BeamWidth               int      // Beam search width (3-10 recommended)
	MutationOperators       []string // Operators for prompt mutation
	UseSearchHistory        bool     // Use search history for better mutations
	GreedySelection         bool     // Use greedy selection for efficiency
	LengthOptimization      bool     // Optimize for prompt length
	MaxPromptLength         int      // Maximum optimized prompt length
	MinImprovementThreshold float64  // Minimum improvement to continue
	MutationProbability     float64  // Probability of applying each mutation
}

// APEXCandidate represents a candidate prompt in the beam search
type APEXCandidate struct {
	Prompt      string
	Score       float64
	Generation  int
	MutationLog []string // Track mutations applied
	Parent      *APEXCandidate
}

// APEXSearchHistory tracks successful mutations for learning
type APEXSearchHistory struct {
	SuccessfulMutations map[string]float64 // mutation -> avg improvement
	FailedMutations     map[string]int     // mutation -> failure count
	BestPrompts         []*APEXCandidate   // Keep track of best prompts
}

// NewAPEXOptimizer creates a new APEX optimizer
func NewAPEXOptimizer(llmProvider llm.Provider) *APEXOptimizer {
	return &APEXOptimizer{
		llm: llmProvider,
	}
}

// OptimizeWithAPEX runs APEX-style long prompt optimization
func (o *APEXOptimizer) OptimizeWithAPEX(ctx context.Context, cfg Config) (*OptimizationResult, error) {
	startTime := time.Now()

	// Parse APEX-specific config
	apexConfig := o.parseAPEXConfig(cfg)

	result := &OptimizationResult{
		OriginalPrompt: cfg.InitialPrompt,
		Iterations:     make([]IterationResult, 0, cfg.Iterations),
		CreatedAt:      startTime,
	}

	// Initialize search history
	history := &APEXSearchHistory{
		SuccessfulMutations: make(map[string]float64),
		FailedMutations:     make(map[string]int),
		BestPrompts:         make([]*APEXCandidate, 0),
	}

	// Initialize beam with original prompt
	initialCandidate := &APEXCandidate{
		Prompt:      cfg.InitialPrompt,
		Score:       0.0, // Will be evaluated
		Generation:  0,
		MutationLog: []string{},
		Parent:      nil,
	}

	// Evaluate initial prompt
	initialScore, err := o.evaluatePrompt(ctx, cfg.InitialPrompt, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate initial prompt: %w", err)
	}
	initialCandidate.Score = initialScore

	beam := []*APEXCandidate{initialCandidate}

	// APEX beam search optimization
	for generation := 0; generation < cfg.Iterations; generation++ {
		iterStart := time.Now()

		// Generate new candidates through mutation
		newCandidates, err := o.generateAPEXCandidates(ctx, beam, apexConfig, history, cfg)
		if err != nil {
			return nil, fmt.Errorf("APEX generation %d failed: %w", generation+1, err)
		}

		// Evaluate all candidates
		for _, candidate := range newCandidates {
			score, evalErr := o.evaluatePrompt(ctx, candidate.Prompt, cfg)
			if evalErr != nil {
				candidate.Score = 0.0 // Penalize failed evaluations
			} else {
				candidate.Score = score
			}
			candidate.Generation = generation + 1
		}

		// Update search history with results
		o.updateSearchHistory(history, beam, newCandidates)

		// Select top candidates for next beam (greedy selection)
		allCandidates := append(beam, newCandidates...)
		beam = o.selectTopCandidates(allCandidates, apexConfig.BeamWidth)

		// Record best candidate for this iteration
		bestCandidate := beam[0]
		iteration := IterationResult{
			Iteration: generation + 1,
			Prompt:    bestCandidate.Prompt,
			Score:     bestCandidate.Score,
			Feedback:  o.generateAPEXFeedback(bestCandidate, history),
			Duration:  time.Since(iterStart),
		}

		result.Iterations = append(result.Iterations, iteration)

		// Early termination if improvement is too small
		if generation > 0 {
			prevScore := result.Iterations[generation-1].Score
			improvement := bestCandidate.Score - prevScore
			if improvement < apexConfig.MinImprovementThreshold {
				break
			}
		}
	}

	// Set final optimized prompt
	if len(beam) > 0 {
		result.OptimizedPrompt = beam[0].Prompt
	} else {
		result.OptimizedPrompt = cfg.InitialPrompt
	}

	result.ImprovementScore = o.calculateAPEXImprovementScore(result)
	result.TotalDuration = time.Since(startTime)

	return result, nil
}

// generateAPEXCandidates generates new candidate prompts through mutation
func (o *APEXOptimizer) generateAPEXCandidates(ctx context.Context, beam []*APEXCandidate, config APEXConfig, history *APEXSearchHistory, cfg Config) ([]*APEXCandidate, error) {
	var newCandidates []*APEXCandidate

	// Generate candidates from each beam member
	for _, parent := range beam {
		for _, mutationOp := range config.MutationOperators {
			// Apply mutation probability
			if rand.Float64() > config.MutationProbability {
				continue
			}

			// Skip mutations that have failed too often
			if config.UseSearchHistory {
				if failures, exists := history.FailedMutations[mutationOp]; exists && failures > 3 {
					continue
				}
			}

			// Apply mutation
			mutatedPrompt, err := o.applyMutation(ctx, parent.Prompt, mutationOp, config, cfg)
			if err != nil {
				continue // Skip failed mutations
			}

			// Create new candidate
			candidate := &APEXCandidate{
				Prompt:      mutatedPrompt,
				Score:       0.0, // Will be evaluated later
				Generation:  parent.Generation + 1,
				MutationLog: append(parent.MutationLog, mutationOp),
				Parent:      parent,
			}

			newCandidates = append(newCandidates, candidate)
		}
	}

	return newCandidates, nil
}

// applyMutation applies a specific mutation operator to a prompt
func (o *APEXOptimizer) applyMutation(ctx context.Context, prompt, mutation string, config APEXConfig, cfg Config) (string, error) {
	var mutationPrompt string

	switch mutation {
	case "rephrase_section":
		mutationPrompt = o.generateRephraseMutation(prompt)
	case "add_examples":
		mutationPrompt = o.generateExampleMutation(prompt)
	case "restructure_flow":
		mutationPrompt = o.generateRestructureMutation(prompt)
	case "clarify_instructions":
		mutationPrompt = o.generateClarificationMutation(prompt)
	case "add_constraints":
		mutationPrompt = o.generateConstraintMutation(prompt)
	case "optimize_length":
		mutationPrompt = o.generateLengthOptimizationMutation(prompt, config.MaxPromptLength)
	case "enhance_specificity":
		mutationPrompt = o.generateSpecificityMutation(prompt)
	case "improve_formatting":
		mutationPrompt = o.generateFormattingMutation(prompt)
	default:
		return "", fmt.Errorf("unknown mutation operator: %s", mutation)
	}

	options := llm.GenerateOptions{
		Temperature: &cfg.Temperature,
		MaxTokens:   &cfg.MaxTokens,
	}

	response, err := o.llm.Generate(ctx, mutationPrompt, options)
	if err != nil {
		return "", fmt.Errorf("mutation failed: %w", err)
	}

	// Extract mutated prompt from response
	mutatedPrompt := o.extractMutatedPrompt(response.Text, prompt)

	// Apply length constraints if needed
	if config.LengthOptimization && len(mutatedPrompt) > config.MaxPromptLength {
		mutatedPrompt = o.truncatePrompt(mutatedPrompt, config.MaxPromptLength)
	}

	return mutatedPrompt, nil
}

// generateRephraseMutation creates a mutation for rephrasing sections
func (o *APEXOptimizer) generateRephraseMutation(prompt string) string {
	return fmt.Sprintf(`You are an expert prompt engineer. Improve this prompt by rephrasing sections for better clarity and effectiveness.

CURRENT PROMPT:
%s

TASK: Rephrase key sections to improve clarity, remove ambiguity, and enhance effectiveness. Maintain the original intent and requirements.

IMPROVED PROMPT:`, prompt)
}

// generateExampleMutation creates a mutation for adding examples
func (o *APEXOptimizer) generateExampleMutation(prompt string) string {
	return fmt.Sprintf(`You are an expert prompt engineer. Enhance this prompt by adding relevant examples that demonstrate the desired behavior.

CURRENT PROMPT:
%s

TASK: Add 1-2 high-quality examples that clearly demonstrate the expected input/output format and quality. Examples should be diverse and instructive.

ENHANCED PROMPT:`, prompt)
}

// generateRestructureMutation creates a mutation for restructuring flow
func (o *APEXOptimizer) generateRestructureMutation(prompt string) string {
	return fmt.Sprintf(`You are an expert prompt engineer. Restructure this prompt to improve information flow and logical organization.

CURRENT PROMPT:
%s

TASK: Reorganize the prompt for better logical flow. Ensure:
1. Context comes before instructions
2. Instructions are in logical order
3. Examples follow instructions
4. Constraints are clearly stated
5. Output format is specified at the end

RESTRUCTURED PROMPT:`, prompt)
}

// generateClarificationMutation creates a mutation for clarifying instructions
func (o *APEXOptimizer) generateClarificationMutation(prompt string) string {
	return fmt.Sprintf(`You are an expert prompt engineer. Clarify ambiguous instructions in this prompt to prevent misinterpretation.

CURRENT PROMPT:
%s

TASK: Identify and clarify any ambiguous instructions, vague requirements, or unclear expectations. Make instructions more specific and actionable.

CLARIFIED PROMPT:`, prompt)
}

// generateConstraintMutation creates a mutation for adding constraints
func (o *APEXOptimizer) generateConstraintMutation(prompt string) string {
	return fmt.Sprintf(`You are an expert prompt engineer. Add appropriate constraints to this prompt to improve output quality and consistency.

CURRENT PROMPT:
%s

TASK: Add relevant constraints such as:
- Output format requirements
- Length limitations
- Quality standards
- Content restrictions
- Style guidelines

CONSTRAINED PROMPT:`, prompt)
}

// generateLengthOptimizationMutation creates a mutation for optimizing length
func (o *APEXOptimizer) generateLengthOptimizationMutation(prompt string, maxLength int) string {
	return fmt.Sprintf(`You are an expert prompt engineer. Optimize this prompt for length while maintaining effectiveness.

CURRENT PROMPT:
%s

TASK: Reduce prompt length to approximately %d characters while preserving:
- Core instructions and requirements
- Essential context
- Key examples (if any)
- Important constraints

Focus on removing redundancy and unnecessary verbosity.

OPTIMIZED PROMPT:`, prompt, maxLength)
}

// generateSpecificityMutation creates a mutation for enhancing specificity
func (o *APEXOptimizer) generateSpecificityMutation(prompt string) string {
	return fmt.Sprintf(`You are an expert prompt engineer. Make this prompt more specific and precise.

CURRENT PROMPT:
%s

TASK: Enhance specificity by:
- Defining vague terms
- Adding precise requirements
- Specifying expected behaviors
- Clarifying success criteria
- Adding measurable standards

SPECIFIC PROMPT:`, prompt)
}

// generateFormattingMutation creates a mutation for improving formatting
func (o *APEXOptimizer) generateFormattingMutation(prompt string) string {
	return fmt.Sprintf(`You are an expert prompt engineer. Improve the formatting and structure of this prompt for better readability.

CURRENT PROMPT:
%s

TASK: Improve formatting with:
- Clear sections and headers
- Bullet points or numbered lists where appropriate
- Proper spacing and organization
- Highlighting of key information
- Logical visual hierarchy

FORMATTED PROMPT:`, prompt)
}

// extractMutatedPrompt extracts the mutated prompt from the LLM response
func (o *APEXOptimizer) extractMutatedPrompt(response, originalPrompt string) string {
	// Look for common patterns in responses
	patterns := []string{
		"IMPROVED PROMPT:",
		"ENHANCED PROMPT:",
		"RESTRUCTURED PROMPT:",
		"CLARIFIED PROMPT:",
		"CONSTRAINED PROMPT:",
		"OPTIMIZED PROMPT:",
		"SPECIFIC PROMPT:",
		"FORMATTED PROMPT:",
	}

	for _, pattern := range patterns {
		if idx := strings.Index(strings.ToUpper(response), pattern); idx != -1 {
			extracted := strings.TrimSpace(response[idx+len(pattern):])
			if len(extracted) > 50 { // Ensure substantial content
				return extracted
			}
		}
	}

	// Fallback: look for the longest substantial paragraph
	paragraphs := strings.Split(response, "\n\n")
	longest := ""
	for _, para := range paragraphs {
		para = strings.TrimSpace(para)
		if len(para) > len(longest) && len(para) > 100 {
			longest = para
		}
	}

	if longest != "" {
		return longest
	}

	// Last resort: return original if extraction fails
	return originalPrompt
}

// truncatePrompt intelligently truncates a prompt to fit length constraints
func (o *APEXOptimizer) truncatePrompt(prompt string, maxLength int) string {
	if len(prompt) <= maxLength {
		return prompt
	}

	// Try to truncate at sentence boundaries
	sentences := strings.Split(prompt, ". ")
	result := ""

	for i, sentence := range sentences {
		candidate := result + sentence
		if i < len(sentences)-1 {
			candidate += ". "
		}

		if len(candidate) > maxLength {
			break
		}
		result = candidate
	}

	// If no sentences fit, truncate at word boundaries
	if result == "" {
		words := strings.Split(prompt, " ")
		for _, word := range words {
			candidate := result + " " + word
			if len(candidate) > maxLength {
				break
			}
			result = candidate
		}
	}

	return strings.TrimSpace(result)
}

// evaluatePrompt evaluates a prompt's quality (simplified for now)
func (o *APEXOptimizer) evaluatePrompt(ctx context.Context, prompt string, cfg Config) (float64, error) {
	// This is a simplified evaluation - in practice, this would use
	// more sophisticated metrics based on the specific use case

	// Base score on prompt characteristics
	score := 5.0

	// Length-based scoring
	promptLength := len(prompt)
	if promptLength > 100 && promptLength < 2000 {
		score += 1.0
	}

	// Structure-based scoring
	if strings.Contains(prompt, "TASK:") || strings.Contains(prompt, "INSTRUCTIONS:") {
		score += 0.5
	}
	if strings.Contains(prompt, "EXAMPLE:") || strings.Contains(prompt, "FORMAT:") {
		score += 0.5
	}

	// Clarity indicators
	sentences := strings.Split(prompt, ".")
	if len(sentences) > 3 && len(sentences) < 20 {
		score += 0.5
	}

	// Add some randomness to simulate real evaluation
	noise := (rand.Float64() - 0.5) * 0.5
	score += noise

	return math.Max(0.0, math.Min(10.0, score)), nil
}

// updateSearchHistory updates the search history with mutation results
func (o *APEXOptimizer) updateSearchHistory(history *APEXSearchHistory, oldBeam, newCandidates []*APEXCandidate) {
	// Find the best score from the old beam
	var bestOldScore float64
	for _, candidate := range oldBeam {
		if candidate.Score > bestOldScore {
			bestOldScore = candidate.Score
		}
	}

	// Evaluate each new candidate's improvement
	for _, candidate := range newCandidates {
		if len(candidate.MutationLog) > 0 {
			lastMutation := candidate.MutationLog[len(candidate.MutationLog)-1]
			improvement := candidate.Score - bestOldScore

			if improvement > 0 {
				// Track successful mutation
				if current, exists := history.SuccessfulMutations[lastMutation]; exists {
					history.SuccessfulMutations[lastMutation] = (current + improvement) / 2
				} else {
					history.SuccessfulMutations[lastMutation] = improvement
				}
			} else {
				// Track failed mutation
				history.FailedMutations[lastMutation]++
			}
		}
	}

	// Update best prompts
	allCandidates := append(oldBeam, newCandidates...)
	sort.Slice(allCandidates, func(i, j int) bool {
		return allCandidates[i].Score > allCandidates[j].Score
	})

	// Keep top 5 best prompts in history
	maxKeep := 5
	if len(allCandidates) < maxKeep {
		maxKeep = len(allCandidates)
	}
	history.BestPrompts = allCandidates[:maxKeep]
}

// selectTopCandidates selects the top candidates for the next beam
func (o *APEXOptimizer) selectTopCandidates(candidates []*APEXCandidate, beamWidth int) []*APEXCandidate {
	// Sort by score (descending)
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].Score > candidates[j].Score
	})

	// Select top candidates up to beam width
	if len(candidates) <= beamWidth {
		return candidates
	}

	return candidates[:beamWidth]
}

// generateAPEXFeedback generates feedback for APEX iterations
func (o *APEXOptimizer) generateAPEXFeedback(candidate *APEXCandidate, history *APEXSearchHistory) string {
	feedback := fmt.Sprintf("Score: %.2f", candidate.Score)

	if len(candidate.MutationLog) > 0 {
		feedback += fmt.Sprintf(" | Mutations: %s", strings.Join(candidate.MutationLog, " → "))
	}

	if len(history.SuccessfulMutations) > 0 {
		// Find best performing mutation
		bestMutation := ""
		bestScore := 0.0
		for mutation, score := range history.SuccessfulMutations {
			if score > bestScore {
				bestScore = score
				bestMutation = mutation
			}
		}
		if bestMutation != "" {
			feedback += fmt.Sprintf(" | Best mutation: %s (%.2f)", bestMutation, bestScore)
		}
	}

	return feedback
}

// calculateAPEXImprovementScore calculates overall improvement for APEX results
func (o *APEXOptimizer) calculateAPEXImprovementScore(result *OptimizationResult) float64 {
	if len(result.Iterations) == 0 {
		return 0.0
	}

	// APEX improvement is based on the final best score achieved
	finalScore := result.Iterations[len(result.Iterations)-1].Score

	// Apply bonus for consistent improvement across iterations
	consistencyBonus := 0.0
	for i := 1; i < len(result.Iterations); i++ {
		if result.Iterations[i].Score >= result.Iterations[i-1].Score {
			consistencyBonus += 0.1
		}
	}

	return finalScore + consistencyBonus
}

// parseAPEXConfig extracts APEX-specific configuration
func (o *APEXOptimizer) parseAPEXConfig(cfg Config) APEXConfig {
	// Set defaults optimized for APEX performance
	apexConfig := APEXConfig{
		BeamWidth: 5,
		MutationOperators: []string{
			"rephrase_section",
			"add_examples",
			"restructure_flow",
			"clarify_instructions",
			"add_constraints",
			"optimize_length",
			"enhance_specificity",
			"improve_formatting",
		},
		UseSearchHistory:        true,
		GreedySelection:         true,
		LengthOptimization:      true,
		MaxPromptLength:         2000,
		MinImprovementThreshold: 0.05,
		MutationProbability:     0.3,
	}

	// TODO: Parse from cfg.Method parameters when advanced config is implemented
	// For now, use defaults which provide optimal APEX behavior

	return apexConfig
}
