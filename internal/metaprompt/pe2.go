package metaprompt

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/tmc/pe/internal/llm"
)

// PE2Optimizer implements the PE2 (Prompt Engineering a Prompt Engineer) method
// Based on 2024 research: "Prompt Engineering a Prompt Engineer"
// PE2 infuses three key components into meta-prompts:
// 1. Detailed descriptions
// 2. Context specification
// 3. Step-by-step reasoning template
type PE2Optimizer struct {
	llm llm.Provider
}

// PE2Config contains configuration for PE2 optimization
type PE2Config struct {
	ReasoningTemplate    string // "chain_of_thought", "step_by_step", "analytical"
	ContextSpecification string // "minimal", "standard", "detailed"
	DescriptionDepth     string // "basic", "standard", "comprehensive"
	MetaPromptStyle      string // "expert", "systematic", "creative"
	FeedbackIntegration  bool   // Use feedback from previous iterations
	ErrorCorrection      bool   // Active error detection and correction
}

// NewPE2Optimizer creates a new PE2 optimizer
func NewPE2Optimizer(llmProvider llm.Provider) *PE2Optimizer {
	return &PE2Optimizer{
		llm: llmProvider,
	}
}

// OptimizeWithPE2 runs PE2-style optimization
func (o *PE2Optimizer) OptimizeWithPE2(ctx context.Context, cfg Config) (*OptimizationResult, error) {
	startTime := time.Now()

	// Parse PE2-specific config from method parameters
	pe2Config := o.parsePE2Config(cfg)

	result := &OptimizationResult{
		OriginalPrompt: cfg.InitialPrompt,
		Iterations:     make([]IterationResult, 0, cfg.Iterations),
		CreatedAt:      startTime,
	}

	currentPrompt := cfg.InitialPrompt
	var previousFeedback []string

	for i := 0; i < cfg.Iterations; i++ {
		iterStart := time.Now()

		// Generate PE2 meta-prompt
		metaPrompt := o.generatePE2MetaPrompt(currentPrompt, pe2Config, previousFeedback)

		// Execute optimization iteration
		optimizedPrompt, feedback, score, err := o.executePE2Iteration(ctx, metaPrompt, cfg)
		if err != nil {
			return nil, fmt.Errorf("PE2 iteration %d failed: %w", i+1, err)
		}

		iteration := IterationResult{
			Iteration: i + 1,
			Prompt:    optimizedPrompt,
			Score:     score,
			Feedback:  feedback,
			Duration:  time.Since(iterStart),
		}

		result.Iterations = append(result.Iterations, iteration)

		// Update for next iteration
		currentPrompt = optimizedPrompt
		if pe2Config.FeedbackIntegration {
			previousFeedback = append(previousFeedback, feedback)
			// Keep only last 3 feedback items to avoid context overload
			if len(previousFeedback) > 3 {
				previousFeedback = previousFeedback[len(previousFeedback)-3:]
			}
		}
	}

	result.OptimizedPrompt = currentPrompt
	result.ImprovementScore = o.calculatePE2ImprovementScore(result)
	result.TotalDuration = time.Since(startTime)

	return result, nil
}

// generatePE2MetaPrompt creates the PE2-style meta-prompt
func (o *PE2Optimizer) generatePE2MetaPrompt(currentPrompt string, config PE2Config, previousFeedback []string) string {
	var metaPrompt strings.Builder

	// 1. Expert persona and detailed description
	metaPrompt.WriteString(o.generateExpertPersona(config.MetaPromptStyle))
	metaPrompt.WriteString("\n\n")

	// 2. Detailed task description
	metaPrompt.WriteString(o.generateDetailedDescription(config.DescriptionDepth))
	metaPrompt.WriteString("\n\n")

	// 3. Context specification
	metaPrompt.WriteString(o.generateContextSpecification(config.ContextSpecification))
	metaPrompt.WriteString("\n\n")

	// 4. Step-by-step reasoning template
	metaPrompt.WriteString(o.generateReasoningTemplate(config.ReasoningTemplate))
	metaPrompt.WriteString("\n\n")

	// 5. Current prompt to optimize
	metaPrompt.WriteString("CURRENT PROMPT TO OPTIMIZE:\n")
	metaPrompt.WriteString("---\n")
	metaPrompt.WriteString(currentPrompt)
	metaPrompt.WriteString("\n---\n\n")

	// 6. Previous feedback integration (if enabled)
	if config.FeedbackIntegration && len(previousFeedback) > 0 {
		metaPrompt.WriteString("PREVIOUS OPTIMIZATION FEEDBACK:\n")
		for i, feedback := range previousFeedback {
			metaPrompt.WriteString(fmt.Sprintf("Iteration %d: %s\n", i+1, feedback))
		}
		metaPrompt.WriteString("\n")
	}

	// 7. Error correction instructions (if enabled)
	if config.ErrorCorrection {
		metaPrompt.WriteString(o.generateErrorCorrectionInstructions())
		metaPrompt.WriteString("\n\n")
	}

	// 8. Output format specification
	metaPrompt.WriteString(o.generateOutputFormat())

	return metaPrompt.String()
}

// generateExpertPersona creates the expert persona based on style
func (o *PE2Optimizer) generateExpertPersona(style string) string {
	switch style {
	case "expert":
		return `You are a world-class prompt engineering expert with deep expertise in:
- Advanced prompting techniques (CoT, few-shot, zero-shot)
- LLM behavior analysis and optimization
- Cognitive psychology and instruction design
- Natural language processing and computational linguistics
- Systematic evaluation and quality assessment

Your expertise spans both theoretical foundations and practical application across diverse domains including reasoning, creative generation, code synthesis, and analytical tasks.`

	case "systematic":
		return `You are a systematic prompt optimization specialist who follows rigorous methodologies:
- Apply structured analysis frameworks to identify improvement opportunities
- Use evidence-based approaches for prompt refinement
- Implement systematic testing and validation procedures
- Follow established best practices from prompt engineering research
- Maintain consistency and reproducibility in optimization processes

Your approach is methodical, data-driven, and focuses on measurable improvements.`

	case "creative":
		return `You are an innovative prompt engineering researcher who:
- Explores novel approaches to prompt design and optimization
- Combines creative thinking with technical precision
- Develops breakthrough techniques for challenging tasks
- Adapts emerging research findings to practical applications
- Balances innovation with proven optimization principles

Your strength lies in finding creative solutions while maintaining technical rigor.`

	default:
		return `You are an expert prompt engineer specializing in systematic prompt optimization and improvement.`
	}
}

// generateDetailedDescription creates detailed task description
func (o *PE2Optimizer) generateDetailedDescription(depth string) string {
	switch depth {
	case "comprehensive":
		return `COMPREHENSIVE TASK DESCRIPTION:

Your primary objective is to analyze and systematically improve the given prompt to maximize its effectiveness, clarity, and reliability. This involves:

CORE OPTIMIZATION DIMENSIONS:
1. CLARITY: Ensure instructions are unambiguous and precisely communicate the intended task
2. SPECIFICITY: Define requirements, constraints, and success criteria with appropriate detail
3. COMPLETENESS: Include all necessary context, examples, and guidance for optimal performance
4. STRUCTURE: Organize information in a logical, easy-to-follow sequence
5. ROBUSTNESS: Design prompts that handle edge cases and prevent common failure modes
6. EFFICIENCY: Optimize for token usage while maintaining effectiveness

ADVANCED CONSIDERATIONS:
- Cognitive load optimization for the target LLM
- Integration of appropriate reasoning frameworks
- Consideration of potential misinterpretations or ambiguities
- Balance between specificity and generalizability
- Incorporation of quality control mechanisms

SUCCESS METRICS:
- Improved task completion accuracy
- Reduced ambiguity and misinterpretation
- Enhanced consistency across diverse inputs
- Better handling of edge cases and error conditions
- Optimal balance of comprehensiveness and conciseness`

	case "standard":
		return `TASK DESCRIPTION:

Analyze the provided prompt and create an improved version that:
1. Communicates the task more clearly and precisely
2. Provides appropriate context and constraints
3. Includes relevant examples or formatting guidance
4. Reduces potential for misinterpretation or errors
5. Optimizes for the intended use case and audience

Focus on practical improvements that will measurably enhance prompt effectiveness while maintaining the original intent and requirements.`

	case "basic":
		return `TASK: Analyze and improve the given prompt to make it clearer, more effective, and more reliable.`

	default:
		return `TASK: Optimize the given prompt for improved clarity and effectiveness.`
	}
}

// generateContextSpecification creates context specification guidelines
func (o *PE2Optimizer) generateContextSpecification(level string) string {
	switch level {
	case "detailed":
		return `CONTEXT SPECIFICATION REQUIREMENTS:

When optimizing the prompt, ensure comprehensive context coverage:

DOMAIN CONTEXT:
- Specify the subject matter domain and any specialized knowledge required
- Include relevant technical terminology and domain-specific conventions
- Define the scope and boundaries of the task within the domain

AUDIENCE CONTEXT:
- Identify the intended audience (expert, general, specific role)
- Adjust language complexity and explanation depth accordingly
- Consider background knowledge assumptions

TASK CONTEXT:
- Clarify the specific use case and application scenario
- Define success criteria and quality standards
- Specify any constraints or limitations

INTERACTION CONTEXT:
- Define the expected input/output format and structure
- Specify any multi-turn conversation requirements
- Clarify the role of the AI assistant in the interaction

TECHNICAL CONTEXT:
- Consider token limitations and efficiency requirements
- Account for model capabilities and limitations
- Optimize for the target LLM architecture and training`

	case "standard":
		return `CONTEXT REQUIREMENTS:
- Specify the domain and subject matter
- Define the intended audience and use case
- Clarify input/output format expectations
- Include relevant constraints and limitations`

	case "minimal":
		return `CONTEXT: Consider the domain, audience, and use case when optimizing.`

	default:
		return `CONTEXT: Provide appropriate context for the optimization task.`
	}
}

// generateReasoningTemplate creates step-by-step reasoning template
func (o *PE2Optimizer) generateReasoningTemplate(template string) string {
	switch template {
	case "chain_of_thought":
		return `STEP-BY-STEP REASONING TEMPLATE:

Follow this chain-of-thought process for systematic optimization:

STEP 1 - ANALYSIS:
Think through: What is the current prompt trying to achieve? What are its strengths and weaknesses?

STEP 2 - PROBLEM IDENTIFICATION:
Think through: What specific issues might cause confusion, errors, or suboptimal outputs?

STEP 3 - SOLUTION DESIGN:
Think through: How can each identified issue be addressed? What improvements would be most impactful?

STEP 4 - INTEGRATION:
Think through: How can improvements be integrated while maintaining coherence and flow?

STEP 5 - VALIDATION:
Think through: Does the optimized prompt address the original requirements? Are there any new issues?

Show your reasoning for each step before providing the final optimized prompt.`

	case "analytical":
		return `ANALYTICAL REASONING FRAMEWORK:

Use systematic analysis to guide optimization:

DECOMPOSITION:
- Break down the prompt into component parts
- Analyze each component's function and effectiveness
- Identify dependencies and relationships between parts

EVALUATION:
- Assess clarity, specificity, and completeness of each component
- Identify potential failure modes or edge cases
- Evaluate alignment with intended outcomes

SYNTHESIS:
- Design improved versions of each component
- Ensure components work together effectively
- Optimize overall structure and flow

VERIFICATION:
- Check that improvements address identified issues
- Ensure no new problems are introduced
- Validate against success criteria

Present your analysis and reasoning before the optimized prompt.`

	case "step_by_step":
		return `STEP-BY-STEP OPTIMIZATION PROCESS:

Follow these systematic steps:

1. READ and understand the current prompt thoroughly
2. IDENTIFY specific areas for improvement
3. PRIORITIZE improvements by potential impact
4. DESIGN solutions for each priority area
5. INTEGRATE solutions into a coherent whole
6. REVIEW the optimized prompt for completeness
7. VALIDATE against original requirements

Execute each step explicitly and show your work.`

	default:
		return `REASONING: Use step-by-step thinking to analyze and improve the prompt systematically.`
	}
}

// generateErrorCorrectionInstructions creates error correction guidelines
func (o *PE2Optimizer) generateErrorCorrectionInstructions() string {
	return `ERROR CORRECTION INSTRUCTIONS:

Actively identify and correct these common prompt engineering errors:

CLARITY ERRORS:
- Ambiguous pronouns or references
- Unclear task boundaries or scope
- Missing or incomplete instructions
- Inconsistent terminology

LOGICAL ERRORS:
- Contradictory requirements or constraints
- Circular reasoning or dependencies
- Invalid assumptions about capabilities
- Unrealistic expectations

STRUCTURAL ERRORS:
- Poor information organization
- Missing context or background
- Inappropriate level of detail
- Ineffective examples or demonstrations

ROBUSTNESS ERRORS:
- Vulnerability to edge cases
- Lack of error handling guidance
- Missing quality control mechanisms
- Inadequate constraint specification

Check for and explicitly address any errors you identify.`
}

// generateOutputFormat specifies the expected output format
func (o *PE2Optimizer) generateOutputFormat() string {
	return `OUTPUT FORMAT:

Provide your response in this exact structure:

ANALYSIS:
[Your systematic analysis of the current prompt, identifying specific strengths and areas for improvement]

REASONING:
[Your step-by-step reasoning process for the optimization, showing how you arrived at each improvement]

OPTIMIZED PROMPT:
[The complete optimized prompt, ready for use]

IMPROVEMENT SUMMARY:
[Brief summary of key improvements made and expected impact]

VALIDATION:
[Verification that the optimized prompt meets requirements and addresses identified issues]

Ensure each section is complete and clearly labeled.`
}

// executePE2Iteration executes a single PE2 optimization iteration
func (o *PE2Optimizer) executePE2Iteration(ctx context.Context, metaPrompt string, cfg Config) (string, string, float64, error) {
	options := llm.GenerateOptions{
		Temperature: &cfg.Temperature,
		MaxTokens:   &cfg.MaxTokens,
	}

	response, err := o.llm.Generate(ctx, metaPrompt, options)
	if err != nil {
		return "", "", 0.0, fmt.Errorf("failed to generate PE2 response: %w", err)
	}

	// Parse the structured PE2 response
	optimizedPrompt, feedback, score := o.parsePE2Response(response.Text)

	return optimizedPrompt, feedback, score, nil
}

// parsePE2Response extracts components from PE2 response
func (o *PE2Optimizer) parsePE2Response(response string) (string, string, float64) {
	sections := make(map[string]string)

	// Split response into sections
	lines := strings.Split(response, "\n")
	currentSection := ""
	var sectionContent strings.Builder

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Check if this is a section header
		if strings.HasSuffix(line, ":") &&
			(strings.Contains(line, "ANALYSIS") ||
				strings.Contains(line, "REASONING") ||
				strings.Contains(line, "OPTIMIZED PROMPT") ||
				strings.Contains(line, "IMPROVEMENT SUMMARY") ||
				strings.Contains(line, "VALIDATION")) {

			// Save previous section
			if currentSection != "" {
				sections[currentSection] = strings.TrimSpace(sectionContent.String())
			}

			// Start new section
			currentSection = strings.ToUpper(strings.TrimSuffix(line, ":"))
			sectionContent.Reset()
		} else if currentSection != "" {
			sectionContent.WriteString(line)
			sectionContent.WriteString("\n")
		}
	}

	// Save final section
	if currentSection != "" {
		sections[currentSection] = strings.TrimSpace(sectionContent.String())
	}

	// Extract optimized prompt
	optimizedPrompt := sections["OPTIMIZED PROMPT"]
	if optimizedPrompt == "" {
		// Fallback: try to find prompt in response
		optimizedPrompt = o.extractPromptFallback(response)
	}

	// Create feedback from analysis and reasoning
	feedback := ""
	if analysis := sections["ANALYSIS"]; analysis != "" {
		feedback += "Analysis: " + analysis + " "
	}
	if reasoning := sections["REASONING"]; reasoning != "" {
		feedback += "Reasoning: " + reasoning + " "
	}
	if summary := sections["IMPROVEMENT SUMMARY"]; summary != "" {
		feedback += "Improvements: " + summary
	}

	// Calculate score based on response quality
	score := o.calculatePE2Score(sections)

	return optimizedPrompt, feedback, score
}

// extractPromptFallback attempts to extract prompt when structured parsing fails
func (o *PE2Optimizer) extractPromptFallback(response string) string {
	// Look for common patterns that might contain the optimized prompt
	patterns := []string{
		"OPTIMIZED PROMPT:",
		"IMPROVED PROMPT:",
		"FINAL PROMPT:",
		"OPTIMIZED:",
		"IMPROVED:",
	}

	for _, pattern := range patterns {
		if idx := strings.Index(strings.ToUpper(response), pattern); idx != -1 {
			remaining := response[idx+len(pattern):]
			// Take content until next section or end
			lines := strings.Split(remaining, "\n")
			var promptLines []string
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if line == "" {
					continue
				}
				// Stop at next section header
				if strings.HasSuffix(line, ":") && len(line) < 50 {
					break
				}
				promptLines = append(promptLines, line)
			}
			if len(promptLines) > 0 {
				return strings.Join(promptLines, "\n")
			}
		}
	}

	// Last resort: return first substantial paragraph
	paragraphs := strings.Split(response, "\n\n")
	for _, para := range paragraphs {
		para = strings.TrimSpace(para)
		if len(para) > 100 { // Substantial content
			return para
		}
	}

	return response // Return full response as fallback
}

// calculatePE2Score calculates quality score based on response sections
func (o *PE2Optimizer) calculatePE2Score(sections map[string]string) float64 {
	score := 5.0 // Base score

	// Bonus for complete sections
	if sections["ANALYSIS"] != "" {
		score += 1.0
	}
	if sections["REASONING"] != "" {
		score += 1.0
	}
	if sections["OPTIMIZED PROMPT"] != "" {
		score += 1.5
	}
	if sections["IMPROVEMENT SUMMARY"] != "" {
		score += 0.5
	}
	if sections["VALIDATION"] != "" {
		score += 1.0
	}

	// Quality bonuses based on content length and structure
	for _, content := range sections {
		if len(content) > 200 { // Substantial content
			score += 0.2
		}
		if strings.Contains(content, "step") || strings.Contains(content, "analysis") {
			score += 0.1
		}
	}

	// Cap at 10.0
	if score > 10.0 {
		score = 10.0
	}

	return score
}

// calculatePE2ImprovementScore calculates overall improvement for PE2 results
func (o *PE2Optimizer) calculatePE2ImprovementScore(result *OptimizationResult) float64 {
	if len(result.Iterations) == 0 {
		return 0.0
	}

	// PE2 uses a sophisticated scoring approach
	var totalScore float64
	var maxScore float64

	for _, iter := range result.Iterations {
		totalScore += iter.Score
		if iter.Score > maxScore {
			maxScore = iter.Score
		}
	}

	avgScore := totalScore / float64(len(result.Iterations))

	// PE2 improvement score: weighted combination emphasizing final quality
	// and consistency across iterations
	finalScore := result.Iterations[len(result.Iterations)-1].Score

	// Weight: 50% final score, 30% max score, 20% average score
	improvementScore := (finalScore * 0.5) + (maxScore * 0.3) + (avgScore * 0.2)

	return improvementScore
}

// parsePE2Config extracts PE2-specific configuration
func (o *PE2Optimizer) parsePE2Config(cfg Config) PE2Config {
	// Set defaults
	pe2Config := PE2Config{
		ReasoningTemplate:    "chain_of_thought",
		ContextSpecification: "detailed",
		DescriptionDepth:     "comprehensive",
		MetaPromptStyle:      "expert",
		FeedbackIntegration:  true,
		ErrorCorrection:      true,
	}

	// TODO: Parse from cfg.Method parameters when advanced config is implemented
	// For now, use defaults which provide optimal PE2 behavior

	return pe2Config
}
