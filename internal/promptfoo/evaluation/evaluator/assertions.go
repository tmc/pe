package evaluator

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/tmc/pe/internal/llm"
	. "github.com/tmc/pe/internal/promptfoo/evaluation/metrics"
)

// AssertionType represents different types of assertions that can be made
type AssertionType string

const (
	// Basic assertions
	AssertionContains    AssertionType = "contains"
	AssertionEquals      AssertionType = "equals"
	AssertionMatches     AssertionType = "matches"
	AssertionLength      AssertionType = "length"
	AssertionNotContains AssertionType = "not-contains"

	// Quality-based assertions
	AssertionReadability AssertionType = "readability"
	AssertionSentiment   AssertionType = "sentiment"
	AssertionToxicity    AssertionType = "toxicity"
	AssertionCoherence   AssertionType = "coherence"
	AssertionFactuality  AssertionType = "factuality"

	// LLM-based assertions
	AssertionLLMJudge   AssertionType = "llm-judge"
	AssertionClassify   AssertionType = "classify"
	AssertionSimilarity AssertionType = "similarity"

	// Performance assertions
	AssertionLatency AssertionType = "latency"
	AssertionCost    AssertionType = "cost"
	AssertionTokens  AssertionType = "tokens"

	// Advanced assertions
	AssertionJSON             AssertionType = "json"
	AssertionSQL              AssertionType = "sql"
	AssertionCode             AssertionType = "code"
	AssertionStructure        AssertionType = "structure"
	AssertionPassAtN          AssertionType = "pass-at-n"
	AssertionStructuredOutput AssertionType = "structured-output"
)

// Assertion represents a test assertion with its configuration
type Assertion struct {
	Type      AssertionType          `json:"type"`
	Value     interface{}            `json:"value,omitempty"`
	Provider  string                 `json:"provider,omitempty"`
	Min       *float64               `json:"min,omitempty"`
	Max       *float64               `json:"max,omitempty"`
	Threshold *float64               `json:"threshold,omitempty"`
	Weight    float64                `json:"weight,omitempty"`
	Message   string                 `json:"message,omitempty"`
	Config    map[string]interface{} `json:"config,omitempty"`
}

// AssertionResult represents the result of evaluating an assertion
type AssertionResult struct {
	Type     AssertionType          `json:"type"`
	Passed   bool                   `json:"passed"`
	Score    float64                `json:"score"`
	Expected interface{}            `json:"expected,omitempty"`
	Actual   interface{}            `json:"actual,omitempty"`
	Message  string                 `json:"message"`
	Duration time.Duration          `json:"duration"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// AssertionEvaluator handles evaluation of different assertion types
type AssertionEvaluator struct {
	llm             llm.Provider
	resolveProvider func(string) (llm.Provider, error)
}

// NewAssertionEvaluator creates a new assertion evaluator
func NewAssertionEvaluator(llmProvider llm.Provider) *AssertionEvaluator {
	return &AssertionEvaluator{
		llm:             llmProvider,
		resolveProvider: resolveAssertionProvider,
	}
}

// EvaluateAssertion evaluates a single assertion against an output
func (ae *AssertionEvaluator) EvaluateAssertion(ctx context.Context, assertion Assertion, output string, metadata map[string]interface{}) (*AssertionResult, error) {
	start := time.Now()

	result := &AssertionResult{
		Type:     assertion.Type,
		Duration: time.Since(start),
		Metadata: make(map[string]interface{}),
	}
	var err error

	if assertion.Provider != "" && assertion.Type != AssertionLLMJudge {
		return nil, fmt.Errorf("assertion provider override is only supported for llm-judge assertions")
	}

	switch assertion.Type {
	case AssertionContains:
		result = ae.evaluateContains(assertion, output)
	case AssertionEquals:
		result = ae.evaluateEquals(assertion, output)
	case AssertionMatches:
		result = ae.evaluateMatches(assertion, output)
	case AssertionLength:
		result = ae.evaluateLength(assertion, output)
	case AssertionNotContains:
		result = ae.evaluateNotContains(assertion, output)
	case AssertionReadability:
		result = ae.evaluateReadability(assertion, output)
	case AssertionSentiment:
		result = ae.evaluateSentiment(ctx, assertion, output)
	case AssertionToxicity:
		result = ae.evaluateToxicity(ctx, assertion, output)
	case AssertionCoherence:
		result = ae.evaluateCoherence(ctx, assertion, output)
	case AssertionFactuality:
		result = ae.evaluateFactuality(ctx, assertion, output)
	case AssertionLLMJudge:
		result, err = ae.evaluateLLMJudge(ctx, assertion, output)
		if err != nil {
			return nil, err
		}
	case AssertionClassify:
		result = ae.evaluateClassify(ctx, assertion, output)
	case AssertionSimilarity:
		result = ae.evaluateSimilarity(ctx, assertion, output)
	case AssertionLatency:
		result = ae.evaluateLatency(assertion, metadata)
	case AssertionCost:
		result = ae.evaluateCost(assertion, metadata)
	case AssertionTokens:
		result = ae.evaluateTokens(assertion, metadata)
	case AssertionJSON:
		result = ae.evaluateJSON(assertion, output)
	case AssertionSQL:
		result = ae.evaluateSQL(assertion, output)
	case AssertionCode:
		result = ae.evaluateCode(ctx, assertion, output)
	case AssertionStructure:
		result = ae.evaluateStructure(assertion, output)
	case AssertionPassAtN:
		result = ae.evaluatePassAtN(ctx, assertion, output, metadata)
	case AssertionStructuredOutput:
		result = ae.evaluateStructuredOutput(assertion, output)
	default:
		return nil, fmt.Errorf("unsupported assertion type: %s", assertion.Type)
	}

	result.Duration = time.Since(start)
	return result, nil
}

// Basic assertion implementations

func (ae *AssertionEvaluator) evaluateContains(assertion Assertion, output string) *AssertionResult {
	expectedStr, ok := assertion.Value.(string)
	if !ok {
		return &AssertionResult{
			Type:    assertion.Type,
			Passed:  false,
			Score:   0.0,
			Message: "Assertion value must be a string",
		}
	}

	contains := strings.Contains(strings.ToLower(output), strings.ToLower(expectedStr))
	score := 0.0
	if contains {
		score = 1.0
	}

	return &AssertionResult{
		Type:     assertion.Type,
		Passed:   contains,
		Score:    score,
		Expected: expectedStr,
		Actual:   contains,
		Message:  fmt.Sprintf("Expected output to contain '%s'", expectedStr),
	}
}

func (ae *AssertionEvaluator) evaluateEquals(assertion Assertion, output string) *AssertionResult {
	expectedStr, ok := assertion.Value.(string)
	if !ok {
		return &AssertionResult{
			Type:    assertion.Type,
			Passed:  false,
			Score:   0.0,
			Message: "Assertion value must be a string",
		}
	}

	equals := strings.TrimSpace(output) == strings.TrimSpace(expectedStr)
	score := 0.0
	if equals {
		score = 1.0
	}

	return &AssertionResult{
		Type:     assertion.Type,
		Passed:   equals,
		Score:    score,
		Expected: expectedStr,
		Actual:   output,
		Message:  fmt.Sprintf("Expected output to equal '%s'", expectedStr),
	}
}

func (ae *AssertionEvaluator) evaluateMatches(assertion Assertion, output string) *AssertionResult {
	pattern, ok := assertion.Value.(string)
	if !ok {
		return &AssertionResult{
			Type:    assertion.Type,
			Passed:  false,
			Score:   0.0,
			Message: "Assertion value must be a regex pattern string",
		}
	}

	regex, err := regexp.Compile(pattern)
	if err != nil {
		return &AssertionResult{
			Type:    assertion.Type,
			Passed:  false,
			Score:   0.0,
			Message: fmt.Sprintf("Invalid regex pattern: %v", err),
		}
	}

	matches := regex.MatchString(output)
	score := 0.0
	if matches {
		score = 1.0
	}

	return &AssertionResult{
		Type:     assertion.Type,
		Passed:   matches,
		Score:    score,
		Expected: pattern,
		Actual:   matches,
		Message:  fmt.Sprintf("Expected output to match pattern '%s'", pattern),
	}
}

func (ae *AssertionEvaluator) evaluateLength(assertion Assertion, output string) *AssertionResult {
	length := len(output)

	var passed bool
	var message string

	if assertion.Min != nil && assertion.Max != nil {
		passed = length >= int(*assertion.Min) && length <= int(*assertion.Max)
		message = fmt.Sprintf("Expected length between %d and %d, got %d", int(*assertion.Min), int(*assertion.Max), length)
	} else if assertion.Min != nil {
		passed = length >= int(*assertion.Min)
		message = fmt.Sprintf("Expected minimum length %d, got %d", int(*assertion.Min), length)
	} else if assertion.Max != nil {
		passed = length <= int(*assertion.Max)
		message = fmt.Sprintf("Expected maximum length %d, got %d", int(*assertion.Max), length)
	} else {
		expectedLength, ok := assertion.Value.(float64)
		if !ok {
			return &AssertionResult{
				Type:    assertion.Type,
				Passed:  false,
				Score:   0.0,
				Message: "Length assertion requires min/max or exact value",
			}
		}
		passed = length == int(expectedLength)
		message = fmt.Sprintf("Expected exact length %d, got %d", int(expectedLength), length)
	}

	score := 0.0
	if passed {
		score = 1.0
	}

	return &AssertionResult{
		Type:    assertion.Type,
		Passed:  passed,
		Score:   score,
		Actual:  length,
		Message: message,
	}
}

func (ae *AssertionEvaluator) evaluateNotContains(assertion Assertion, output string) *AssertionResult {
	expectedStr, ok := assertion.Value.(string)
	if !ok {
		return &AssertionResult{
			Type:    assertion.Type,
			Passed:  false,
			Score:   0.0,
			Message: "Assertion value must be a string",
		}
	}

	contains := strings.Contains(strings.ToLower(output), strings.ToLower(expectedStr))
	passed := !contains
	score := 0.0
	if passed {
		score = 1.0
	}

	return &AssertionResult{
		Type:     assertion.Type,
		Passed:   passed,
		Score:    score,
		Expected: fmt.Sprintf("not containing '%s'", expectedStr),
		Actual:   contains,
		Message:  fmt.Sprintf("Expected output to not contain '%s'", expectedStr),
	}
}

// Quality-based assertion implementations

func (ae *AssertionEvaluator) evaluateReadability(assertion Assertion, output string) *AssertionResult {
	// Simple readability score based on sentence and word length
	sentences := strings.Split(output, ".")
	words := strings.Fields(output)

	if len(sentences) == 0 || len(words) == 0 {
		return &AssertionResult{
			Type:    assertion.Type,
			Passed:  false,
			Score:   0.0,
			Message: "Cannot calculate readability for empty text",
		}
	}

	avgWordsPerSentence := float64(len(words)) / float64(len(sentences))
	avgSyllablesPerWord := ae.estimateSyllables(output) / float64(len(words))

	// Simplified Flesch Reading Ease approximation
	readabilityScore := 206.835 - (1.015 * avgWordsPerSentence) - (84.6 * avgSyllablesPerWord)

	// Normalize to 0-1 scale
	normalizedScore := math.Max(0, math.Min(1, readabilityScore/100))

	var passed bool
	if assertion.Min != nil && assertion.Max != nil {
		passed = normalizedScore >= *assertion.Min && normalizedScore <= *assertion.Max
	} else if assertion.Threshold != nil {
		passed = normalizedScore >= *assertion.Threshold
	} else {
		passed = normalizedScore >= 0.5 // Default threshold
	}

	return &AssertionResult{
		Type:    assertion.Type,
		Passed:  passed,
		Score:   normalizedScore,
		Actual:  normalizedScore,
		Message: fmt.Sprintf("Readability score: %.2f", normalizedScore),
		Metadata: map[string]interface{}{
			"avg_words_per_sentence": avgWordsPerSentence,
			"avg_syllables_per_word": avgSyllablesPerWord,
		},
	}
}

func (ae *AssertionEvaluator) estimateSyllables(text string) float64 {
	words := strings.Fields(strings.ToLower(text))
	totalSyllables := 0

	for _, word := range words {
		syllables := 1 // Minimum one syllable per word
		vowels := regexp.MustCompile(`[aeiou]`)
		matches := vowels.FindAllString(word, -1)
		if len(matches) > 1 {
			syllables = len(matches)
		}
		// Adjust for silent 'e' at the end
		if strings.HasSuffix(word, "e") && len(word) > 1 {
			syllables--
		}
		if syllables < 1 {
			syllables = 1
		}
		totalSyllables += syllables
	}

	return float64(totalSyllables)
}

// LLM-based assertion implementations

func (ae *AssertionEvaluator) evaluateSentiment(ctx context.Context, assertion Assertion, output string) *AssertionResult {
	sentimentPrompt := fmt.Sprintf(`Analyze the sentiment of the following text and provide a score from -1 (very negative) to 1 (very positive), with 0 being neutral.

Text: %s

Provide only a numeric score between -1 and 1.`, output)

	response, err := ae.llm.Generate(ctx, sentimentPrompt, llm.GenerateOptions{})
	if err != nil {
		return &AssertionResult{
			Type:    assertion.Type,
			Passed:  false,
			Score:   0.0,
			Message: fmt.Sprintf("Failed to evaluate sentiment: %v", err),
		}
	}

	score, err := strconv.ParseFloat(strings.TrimSpace(response.Text), 64)
	if err != nil {
		score = 0.0 // Default to neutral if parsing fails
	}

	// Normalize to 0-1 scale
	normalizedScore := (score + 1) / 2

	var passed bool
	if assertion.Min != nil && assertion.Max != nil {
		passed = score >= *assertion.Min && score <= *assertion.Max
	} else if assertion.Threshold != nil {
		passed = score >= *assertion.Threshold
	} else {
		expectedSentiment, ok := assertion.Value.(string)
		if ok {
			switch strings.ToLower(expectedSentiment) {
			case "positive":
				passed = score > 0.1
			case "negative":
				passed = score < -0.1
			case "neutral":
				passed = score >= -0.1 && score <= 0.1
			default:
				passed = true
			}
		} else {
			passed = true
		}
	}

	return &AssertionResult{
		Type:    assertion.Type,
		Passed:  passed,
		Score:   normalizedScore,
		Actual:  score,
		Message: fmt.Sprintf("Sentiment score: %.2f", score),
	}
}

func (ae *AssertionEvaluator) evaluateLLMJudge(ctx context.Context, assertion Assertion, output string) (*AssertionResult, error) {
	criteria, ok := assertion.Value.(string)
	if !ok {
		return &AssertionResult{
			Type:    assertion.Type,
			Passed:  false,
			Score:   0.0,
			Message: "LLM judge assertion requires criteria as string value",
		}, nil
	}

	judgeProvider, err := ae.llmJudgeProvider(assertion)
	if err != nil {
		return nil, err
	}

	judgePrompt := fmt.Sprintf(`You are an expert evaluator. Evaluate the following output based on the given criteria.

CRITERIA: %s

OUTPUT TO EVALUATE:
%s

TASK:
1. Evaluate how well the output meets the criteria
2. Provide a score from 0 to 10
3. Provide brief reasoning

FORMAT:
SCORE: [0-10]
REASONING: [brief explanation]`, criteria, output)

	response, err := judgeProvider.Generate(ctx, judgePrompt, llm.GenerateOptions{})
	if err != nil {
		return &AssertionResult{
			Type:    assertion.Type,
			Passed:  false,
			Score:   0.0,
			Message: fmt.Sprintf("Failed to evaluate with LLM judge: %v", err),
		}, nil
	}

	score, reasoning := ae.parseLLMJudgeResponse(response.Text)
	normalizedScore := score / 10.0

	threshold := 0.7
	if assertion.Threshold != nil {
		threshold = *assertion.Threshold
	}

	passed := normalizedScore >= threshold

	return &AssertionResult{
		Type:    assertion.Type,
		Passed:  passed,
		Score:   normalizedScore,
		Actual:  score,
		Message: fmt.Sprintf("LLM Judge Score: %.1f/10 - %s", score, reasoning),
		Metadata: map[string]interface{}{
			"reasoning": reasoning,
			"criteria":  criteria,
		},
	}, nil
}

func (ae *AssertionEvaluator) llmJudgeProvider(assertion Assertion) (llm.Provider, error) {
	if assertion.Provider == "" {
		if ae.llm == nil {
			return nil, fmt.Errorf("llm judge assertion requires a provider")
		}
		return ae.llm, nil
	}

	provider := strings.TrimSpace(assertion.Provider)
	if provider == "" {
		return nil, fmt.Errorf("llm judge assertion provider is empty")
	}
	if ae.resolveProvider == nil {
		return nil, fmt.Errorf("assertion provider overrides are not supported by this evaluator")
	}

	judgeProvider, err := ae.resolveProvider(provider)
	if err != nil {
		return nil, fmt.Errorf("resolve assertion provider %q: %w", provider, err)
	}
	if judgeProvider == nil {
		return nil, fmt.Errorf("resolve assertion provider %q: provider is nil", provider)
	}
	return judgeProvider, nil
}

func resolveAssertionProvider(provider string) (llm.Provider, error) {
	return llm.GetProviderWithOptions(provider, nil)
}

func (ae *AssertionEvaluator) parseLLMJudgeResponse(response string) (float64, string) {
	lines := strings.Split(response, "\n")
	score := 5.0 // default
	reasoning := "No reasoning provided"

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "SCORE:") {
			scoreStr := strings.TrimSpace(line[6:])
			if s, err := strconv.ParseFloat(scoreStr, 64); err == nil {
				score = s
			}
		} else if strings.HasPrefix(line, "REASONING:") {
			reasoning = strings.TrimSpace(line[10:])
		}
	}

	return score, reasoning
}

// Performance assertion implementations

func (ae *AssertionEvaluator) evaluateLatency(assertion Assertion, metadata map[string]interface{}) *AssertionResult {
	latency, ok := metadata["latency"].(time.Duration)
	if !ok {
		return &AssertionResult{
			Type:    assertion.Type,
			Passed:  false,
			Score:   0.0,
			Message: "Latency metadata not found",
		}
	}

	latencyMs := float64(latency.Milliseconds())

	var passed bool
	var message string

	if assertion.Max != nil {
		maxMs := *assertion.Max * 1000 // Convert seconds to milliseconds
		passed = latencyMs <= maxMs
		message = fmt.Sprintf("Expected latency <= %.0fms, got %.0fms", maxMs, latencyMs)
	} else {
		passed = true
		message = fmt.Sprintf("Latency: %.0fms", latencyMs)
	}

	score := 0.0
	if passed {
		score = 1.0
	}

	return &AssertionResult{
		Type:    assertion.Type,
		Passed:  passed,
		Score:   score,
		Actual:  latencyMs,
		Message: message,
	}
}

func (ae *AssertionEvaluator) evaluateCost(assertion Assertion, metadata map[string]interface{}) *AssertionResult {
	cost, ok := metadata["cost"].(float64)
	if !ok {
		return &AssertionResult{
			Type:    assertion.Type,
			Passed:  false,
			Score:   0.0,
			Message: "Cost metadata not found",
		}
	}

	var passed bool
	var message string

	if assertion.Max != nil {
		passed = cost <= *assertion.Max
		message = fmt.Sprintf("Expected cost <= $%.4f, got $%.4f", *assertion.Max, cost)
	} else {
		passed = true
		message = fmt.Sprintf("Cost: $%.4f", cost)
	}

	score := 0.0
	if passed {
		score = 1.0
	}

	return &AssertionResult{
		Type:    assertion.Type,
		Passed:  passed,
		Score:   score,
		Actual:  cost,
		Message: message,
	}
}

func (ae *AssertionEvaluator) evaluateTokens(assertion Assertion, metadata map[string]interface{}) *AssertionResult {
	tokens, ok := metadata["tokens"].(int)
	if !ok {
		return &AssertionResult{
			Type:    assertion.Type,
			Passed:  false,
			Score:   0.0,
			Message: "Token count metadata not found",
		}
	}

	var passed bool
	var message string

	if assertion.Max != nil {
		passed = float64(tokens) <= *assertion.Max
		message = fmt.Sprintf("Expected tokens <= %.0f, got %d", *assertion.Max, tokens)
	} else if assertion.Min != nil {
		passed = float64(tokens) >= *assertion.Min
		message = fmt.Sprintf("Expected tokens >= %.0f, got %d", *assertion.Min, tokens)
	} else {
		passed = true
		message = fmt.Sprintf("Tokens: %d", tokens)
	}

	score := 0.0
	if passed {
		score = 1.0
	}

	return &AssertionResult{
		Type:    assertion.Type,
		Passed:  passed,
		Score:   score,
		Actual:  tokens,
		Message: message,
	}
}

// Advanced assertion implementations

func (ae *AssertionEvaluator) evaluateJSON(assertion Assertion, output string) *AssertionResult {
	var jsonData interface{}
	err := json.Unmarshal([]byte(output), &jsonData)

	if err != nil {
		return &AssertionResult{
			Type:    assertion.Type,
			Passed:  false,
			Score:   0.0,
			Message: fmt.Sprintf("Invalid JSON: %v", err),
		}
	}

	// If schema validation is required
	if schema, ok := assertion.Config["schema"]; ok {
		// Basic schema validation would go here
		_ = schema // placeholder
	}

	return &AssertionResult{
		Type:    assertion.Type,
		Passed:  true,
		Score:   1.0,
		Message: "Valid JSON",
		Metadata: map[string]interface{}{
			"parsed_json": jsonData,
		},
	}
}

func (ae *AssertionEvaluator) evaluateCode(ctx context.Context, assertion Assertion, output string) *AssertionResult {
	language, ok := assertion.Config["language"].(string)
	if !ok {
		language = "unknown"
	}

	codeAnalysisPrompt := fmt.Sprintf(`Analyze the following %s code for correctness, quality, and best practices.

CODE:
%s

EVALUATION CRITERIA:
1. Syntax correctness
2. Logic soundness  
3. Code quality and style
4. Error handling
5. Best practices adherence

Provide a score from 0-10 and brief feedback.

FORMAT:
SCORE: [0-10]
FEEDBACK: [brief analysis]`, language, output)

	response, err := ae.llm.Generate(ctx, codeAnalysisPrompt, llm.GenerateOptions{})
	if err != nil {
		return &AssertionResult{
			Type:    assertion.Type,
			Passed:  false,
			Score:   0.0,
			Message: fmt.Sprintf("Failed to evaluate code: %v", err),
		}
	}

	score, feedback := ae.parseLLMJudgeResponse(response.Text)
	normalizedScore := score / 10.0

	threshold := 0.7
	if assertion.Threshold != nil {
		threshold = *assertion.Threshold
	}

	passed := normalizedScore >= threshold

	return &AssertionResult{
		Type:    assertion.Type,
		Passed:  passed,
		Score:   normalizedScore,
		Actual:  score,
		Message: fmt.Sprintf("Code Quality Score: %.1f/10 - %s", score, feedback),
		Metadata: map[string]interface{}{
			"feedback": feedback,
			"language": language,
		},
	}
}

// Placeholder implementations for remaining assertion types

func (ae *AssertionEvaluator) evaluateToxicity(ctx context.Context, assertion Assertion, output string) *AssertionResult {
	// NOTE: Toxicity detection pending implementation
	// Will integrate with toxicity detection service in future release
	return &AssertionResult{
		Type:    assertion.Type,
		Passed:  false,
		Score:   0.0,
		Message: "Toxicity evaluation not yet implemented (coming in future release)",
	}
}

func (ae *AssertionEvaluator) evaluateCoherence(ctx context.Context, assertion Assertion, output string) *AssertionResult {
	// NOTE: Coherence evaluation pending implementation
	// Will use LLM-based coherence scoring in future release
	return &AssertionResult{
		Type:    assertion.Type,
		Passed:  false,
		Score:   0.0,
		Message: "Coherence evaluation not yet implemented (coming in future release)",
	}
}

func (ae *AssertionEvaluator) evaluateFactuality(ctx context.Context, assertion Assertion, output string) *AssertionResult {
	// NOTE: Factuality checking pending implementation
	// Will integrate with fact-checking service in future release
	return &AssertionResult{
		Type:    assertion.Type,
		Passed:  false,
		Score:   0.0,
		Message: "Factuality evaluation not yet implemented (coming in future release)",
	}
}

func (ae *AssertionEvaluator) evaluateClassify(ctx context.Context, assertion Assertion, output string) *AssertionResult {
	// NOTE: Classification pending implementation
	// Will use LLM-based classification in future release
	return &AssertionResult{
		Type:    assertion.Type,
		Passed:  false,
		Score:   0.0,
		Message: "Classification evaluation not yet implemented (coming in future release)",
	}
}

func (ae *AssertionEvaluator) evaluateSimilarity(ctx context.Context, assertion Assertion, output string) *AssertionResult {
	reference, ok := assertion.Value.(string)
	if !ok || strings.TrimSpace(reference) == "" {
		return &AssertionResult{
			Type:    assertion.Type,
			Passed:  false,
			Score:   0.0,
			Message: "Similarity assertion requires non-empty string value",
		}
	}

	score := lexicalSimilarity(reference, output)
	threshold := 0.8
	if assertion.Threshold != nil {
		threshold = *assertion.Threshold
	}
	passed := score >= threshold

	return &AssertionResult{
		Type:     assertion.Type,
		Passed:   passed,
		Score:    score,
		Expected: reference,
		Actual:   output,
		Message:  fmt.Sprintf("Lexical similarity score: %.2f", score),
		Metadata: map[string]interface{}{
			"method":    "token_jaccard",
			"threshold": threshold,
		},
	}
}

func (ae *AssertionEvaluator) evaluateSQL(assertion Assertion, output string) *AssertionResult {
	normalized := strings.TrimSpace(output)
	if normalized == "" {
		return &AssertionResult{
			Type:    assertion.Type,
			Passed:  false,
			Score:   0.0,
			Message: "SQL output is empty",
		}
	}
	if err := validateSQLShape(normalized); err != nil {
		return &AssertionResult{
			Type:    assertion.Type,
			Passed:  false,
			Score:   0.0,
			Actual:  output,
			Message: fmt.Sprintf("Invalid SQL shape: %v", err),
		}
	}

	return &AssertionResult{
		Type:    assertion.Type,
		Passed:  true,
		Score:   1.0,
		Actual:  output,
		Message: "SQL shape is valid",
		Metadata: map[string]interface{}{
			"method": "local_shape_check",
		},
	}
}

func (ae *AssertionEvaluator) evaluateStructure(assertion Assertion, output string) *AssertionResult {
	required := requiredStructureMarkers(assertion)
	if len(required) == 0 {
		return &AssertionResult{
			Type:    assertion.Type,
			Passed:  false,
			Score:   0.0,
			Message: "Structure assertion requires required markers in value or config.required",
		}
	}

	lowerOutput := strings.ToLower(output)
	missing := make([]string, 0)
	for _, marker := range required {
		if !strings.Contains(lowerOutput, strings.ToLower(marker)) {
			missing = append(missing, marker)
		}
	}
	score := 1.0
	if len(required) > 0 {
		score = float64(len(required)-len(missing)) / float64(len(required))
	}
	passed := len(missing) == 0

	return &AssertionResult{
		Type:     assertion.Type,
		Passed:   passed,
		Score:    score,
		Expected: required,
		Actual:   output,
		Message:  fmt.Sprintf("Structure markers matched: %d/%d", len(required)-len(missing), len(required)),
		Metadata: map[string]interface{}{
			"missing": missing,
			"method":  "required_marker_contains",
		},
	}
}

func (ae *AssertionEvaluator) evaluateStructuredOutput(assertion Assertion, output string) *AssertionResult {
	result := ae.evaluateJSON(assertion, output)
	result.Type = assertion.Type
	if result.Passed {
		result.Message = "Valid structured output"
	}
	return result
}

func lexicalSimilarity(reference, output string) float64 {
	refTokens := tokenSet(reference)
	outTokens := tokenSet(output)
	if len(refTokens) == 0 && len(outTokens) == 0 {
		return 1
	}
	if len(refTokens) == 0 || len(outTokens) == 0 {
		return 0
	}
	intersection := 0
	for token := range refTokens {
		if outTokens[token] {
			intersection++
		}
	}
	union := len(refTokens) + len(outTokens) - intersection
	if union == 0 {
		return 0
	}
	return float64(intersection) / float64(union)
}

func tokenSet(text string) map[string]bool {
	words := regexp.MustCompile(`[A-Za-z0-9_]+`).FindAllString(strings.ToLower(text), -1)
	tokens := make(map[string]bool, len(words))
	for _, word := range words {
		tokens[word] = true
	}
	return tokens
}

func validateSQLShape(query string) error {
	if strings.Contains(query, "\x00") {
		return fmt.Errorf("contains null byte")
	}
	fields := strings.Fields(strings.ToLower(strings.TrimSuffix(query, ";")))
	if len(fields) == 0 {
		return fmt.Errorf("empty query")
	}
	switch fields[0] {
	case "select", "with", "insert", "update", "delete", "create", "alter", "drop":
	default:
		return fmt.Errorf("unsupported starting keyword %q", fields[0])
	}
	if !balancedDelimiters(query, '(', ')') {
		return fmt.Errorf("unbalanced parentheses")
	}
	if strings.Count(query, "'")%2 != 0 {
		return fmt.Errorf("unbalanced single quotes")
	}
	if strings.Count(query, `"`)%2 != 0 {
		return fmt.Errorf("unbalanced double quotes")
	}
	return nil
}

func balancedDelimiters(text string, open, close rune) bool {
	depth := 0
	for _, r := range text {
		switch r {
		case open:
			depth++
		case close:
			depth--
			if depth < 0 {
				return false
			}
		}
	}
	return depth == 0
}

func requiredStructureMarkers(assertion Assertion) []string {
	var markers []string
	add := func(value string) {
		value = strings.TrimSpace(value)
		if value != "" {
			markers = append(markers, value)
		}
	}
	switch value := assertion.Value.(type) {
	case string:
		add(value)
	case []string:
		for _, item := range value {
			add(item)
		}
	case []interface{}:
		for _, item := range value {
			if s, ok := item.(string); ok {
				add(s)
			}
		}
	}
	if raw, ok := assertion.Config["required"]; ok {
		switch value := raw.(type) {
		case string:
			add(value)
		case []string:
			for _, item := range value {
				add(item)
			}
		case []interface{}:
			for _, item := range value {
				if s, ok := item.(string); ok {
					add(s)
				}
			}
		}
	}
	return markers
}

func (ae *AssertionEvaluator) evaluatePassAtN(ctx context.Context, assertion Assertion, output string, metadata map[string]interface{}) *AssertionResult {
	// Get configuration from assertion
	n := 1
	if nVal, ok := assertion.Config["n"].(float64); ok {
		n = int(nVal)
	}

	// Get samples - either from metadata or use the single output
	var samples []string
	if samplesVal, ok := metadata["samples"].([]string); ok {
		samples = samplesVal
	} else if samplesVal, ok := metadata["samples"].([]interface{}); ok {
		// Convert interface slice to string slice
		for _, s := range samplesVal {
			if str, ok := s.(string); ok {
				samples = append(samples, str)
			}
		}
	} else {
		// If no samples provided, use the single output
		samples = []string{output}
	}

	// Get test cases from config
	testCases, hasTestCases := assertion.Config["test_cases"].([]interface{})

	// Create advanced metrics instance
	am := NewAdvancedMetrics(ae.llm)

	var result *PassAtNResult
	if hasTestCases {
		// Convert test cases to proper format
		var testCasesMaps []map[string]interface{}
		for _, tc := range testCases {
			if tcMap, ok := tc.(map[string]interface{}); ok {
				testCasesMaps = append(testCasesMaps, tcMap)
			}
		}
		result = am.CalculatePassAtNWithTests(ctx, n, samples, testCasesMaps)
	} else {
		// Use custom test function if provided
		testFunc := func(code string) bool {
			// Default: check if code is non-empty and appears syntactically valid
			return len(strings.TrimSpace(code)) > 0 && !strings.Contains(code, "error")
		}

		// If a custom validation prompt is provided, use LLM-based validation
		if validationPrompt, ok := assertion.Config["validation_prompt"].(string); ok {
			testFunc = func(code string) bool {
				evalPrompt := fmt.Sprintf(validationPrompt, code)
				response, err := ae.llm.Generate(ctx, evalPrompt, llm.GenerateOptions{
					Temperature: &[]float64{0.0}[0],
				})
				if err != nil {
					return false
				}
				return strings.Contains(strings.ToUpper(response.Text), "YES") ||
					strings.Contains(strings.ToUpper(response.Text), "PASS") ||
					strings.Contains(strings.ToUpper(response.Text), "TRUE")
			}
		}

		result = am.CalculatePassAtN(n, samples, testFunc)
	}

	// Determine if assertion passes based on threshold
	threshold := 0.5
	if assertion.Threshold != nil {
		threshold = *assertion.Threshold
	}

	passed := result.PassRate >= threshold

	return &AssertionResult{
		Type:     assertion.Type,
		Passed:   passed,
		Score:    result.PassRate,
		Expected: fmt.Sprintf("pass@%d >= %.2f", n, threshold),
		Actual:   result.PassRate,
		Message:  fmt.Sprintf("Pass@%d rate: %.2f%% (%d/%d samples passed)", n, result.PassRate*100, result.NumPassed, result.NumSamples),
		Metadata: map[string]interface{}{
			"n":              n,
			"pass_rate":      result.PassRate,
			"num_samples":    result.NumSamples,
			"num_passed":     result.NumPassed,
			"threshold":      threshold,
			"passed_indices": result.Details["passed_indices"],
		},
	}
}
