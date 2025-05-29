package metrics

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/tmc/pe/internal/llm"
)

// MetricType defines the type of metric
type MetricType string

const (
	MetricTypeBuiltin MetricType = "builtin"
	MetricTypePython  MetricType = "python"
	MetricTypeScript  MetricType = "script"
	MetricTypeLLM     MetricType = "llm-graded"
	MetricTypeRegex   MetricType = "regex"
	MetricTypePassAtN MetricType = "pass-at-n"
)

// MetricConfig defines a custom metric configuration
type MetricConfig struct {
	Name        string                 `yaml:"name" json:"name"`
	Type        MetricType             `yaml:"type" json:"type"`
	Script      string                 `yaml:"script" json:"script"`
	Judge       string                 `yaml:"judge" json:"judge"`
	Prompt      string                 `yaml:"prompt" json:"prompt"`
	Criteria    string                 `yaml:"criteria" json:"criteria"`
	Pattern     string                 `yaml:"pattern" json:"pattern"`
	Config      map[string]interface{} `yaml:"config" json:"config"`
	Weight      float64                `yaml:"weight" json:"weight"`
	Threshold   float64                `yaml:"threshold" json:"threshold"`
	Description string                 `yaml:"description" json:"description"`
}

// MetricResult represents the result of a metric evaluation
type MetricResult struct {
	Name        string                 `json:"name"`
	Score       float64                `json:"score"`
	Pass        bool                   `json:"pass"`
	Reason      string                 `json:"reason"`
	Details     map[string]interface{} `json:"details"`
	Latency     time.Duration          `json:"latency"`
	Error       string                 `json:"error,omitempty"`
}

// MetricEvaluator evaluates custom metrics
type MetricEvaluator struct {
	configs   []MetricConfig
	providers map[string]llm.Provider
}

// NewMetricEvaluator creates a new metric evaluator
func NewMetricEvaluator(configs []MetricConfig, providers map[string]llm.Provider) *MetricEvaluator {
	return &MetricEvaluator{
		configs:   configs,
		providers: providers,
	}
}

// EvaluateAll evaluates all configured metrics for a given prompt and response
func (m *MetricEvaluator) EvaluateAll(ctx context.Context, prompt, response string, metadata map[string]interface{}) ([]MetricResult, error) {
	var results []MetricResult
	
	for _, config := range m.configs {
		result, err := m.evaluateMetric(ctx, config, prompt, response, metadata)
		if err != nil {
			result = MetricResult{
				Name:  config.Name,
				Score: 0,
				Pass:  false,
				Error: err.Error(),
			}
		}
		results = append(results, result)
	}
	
	return results, nil
}

// evaluateMetric evaluates a single metric
func (m *MetricEvaluator) evaluateMetric(ctx context.Context, config MetricConfig, prompt, response string, metadata map[string]interface{}) (MetricResult, error) {
	switch config.Type {
	case MetricTypeBuiltin:
		return m.evaluateBuiltinMetric(config, prompt, response, metadata)
	case MetricTypePython:
		return m.evaluatePythonMetric(ctx, config, prompt, response, metadata)
	case MetricTypeScript:
		return m.evaluateScriptMetric(ctx, config, prompt, response, metadata)
	case MetricTypeLLM:
		return m.evaluateLLMMetric(ctx, config, prompt, response, metadata)
	case MetricTypeRegex:
		return m.evaluateRegexMetric(config, prompt, response)
	case MetricTypePassAtN:
		return m.evaluatePassAtNMetric(ctx, config, prompt, response, metadata)
	default:
		return MetricResult{}, fmt.Errorf("unsupported metric type: %s", config.Type)
	}
}

// evaluateBuiltinMetric evaluates built-in metrics
func (m *MetricEvaluator) evaluateBuiltinMetric(config MetricConfig, prompt, response string, metadata map[string]interface{}) (MetricResult, error) {
	start := time.Now()
	
	switch config.Name {
	case "length":
		return m.evaluateLengthMetric(config, response, time.Since(start))
	case "word_count":
		return m.evaluateWordCountMetric(config, response, time.Since(start))
	case "sentiment":
		return m.evaluateSentimentMetric(config, response, time.Since(start))
	case "readability":
		return m.evaluateReadabilityMetric(config, response, time.Since(start))
	case "toxicity":
		return m.evaluateToxicityMetric(config, response, time.Since(start))
	case "coherence":
		return m.evaluateCoherenceMetric(config, response, time.Since(start))
	case "relevance":
		return m.evaluateRelevanceMetric(config, prompt, response, time.Since(start))
	case "pass-at-n", "pass_at_n":
		// Delegate to the pass@n metric evaluator
		return m.evaluatePassAtNMetric(context.Background(), config, prompt, response, metadata)
	default:
		return MetricResult{}, fmt.Errorf("unknown builtin metric: %s", config.Name)
	}
}

// evaluateLengthMetric evaluates response length
func (m *MetricEvaluator) evaluateLengthMetric(config MetricConfig, response string, latency time.Duration) (MetricResult, error) {
	length := len(response)
	
	minLength := int(config.Config["min"].(float64))
	maxLength := int(config.Config["max"].(float64))
	
	pass := length >= minLength && length <= maxLength
	score := 1.0
	reason := fmt.Sprintf("Length: %d characters", length)
	
	if !pass {
		if length < minLength {
			score = float64(length) / float64(minLength)
			reason += fmt.Sprintf(" (below minimum %d)", minLength)
		} else {
			score = float64(maxLength) / float64(length)
			reason += fmt.Sprintf(" (above maximum %d)", maxLength)
		}
	}
	
	return MetricResult{
		Name:    config.Name,
		Score:   score,
		Pass:    pass,
		Reason:  reason,
		Latency: latency,
		Details: map[string]interface{}{
			"length":     length,
			"min_length": minLength,
			"max_length": maxLength,
		},
	}, nil
}

// evaluateWordCountMetric evaluates word count
func (m *MetricEvaluator) evaluateWordCountMetric(config MetricConfig, response string, latency time.Duration) (MetricResult, error) {
	words := strings.Fields(response)
	wordCount := len(words)
	
	minWords := int(config.Config["min"].(float64))
	maxWords := int(config.Config["max"].(float64))
	
	pass := wordCount >= minWords && wordCount <= maxWords
	score := 1.0
	reason := fmt.Sprintf("Word count: %d", wordCount)
	
	if !pass {
		if wordCount < minWords {
			score = float64(wordCount) / float64(minWords)
			reason += fmt.Sprintf(" (below minimum %d)", minWords)
		} else {
			score = float64(maxWords) / float64(wordCount)
			reason += fmt.Sprintf(" (above maximum %d)", maxWords)
		}
	}
	
	return MetricResult{
		Name:    config.Name,
		Score:   score,
		Pass:    pass,
		Reason:  reason,
		Latency: latency,
		Details: map[string]interface{}{
			"word_count": wordCount,
			"min_words":  minWords,
			"max_words":  maxWords,
		},
	}, nil
}

// evaluateSentimentMetric evaluates sentiment (simplified)
func (m *MetricEvaluator) evaluateSentimentMetric(config MetricConfig, response string, latency time.Duration) (MetricResult, error) {
	// Simplified sentiment analysis using keyword matching
	positiveWords := []string{"good", "great", "excellent", "amazing", "wonderful", "fantastic", "positive", "happy", "love"}
	negativeWords := []string{"bad", "terrible", "awful", "horrible", "negative", "sad", "hate", "wrong", "fail"}
	
	response = strings.ToLower(response)
	positiveCount := 0
	negativeCount := 0
	
	for _, word := range positiveWords {
		if strings.Contains(response, word) {
			positiveCount++
		}
	}
	
	for _, word := range negativeWords {
		if strings.Contains(response, word) {
			negativeCount++
		}
	}
	
	// Calculate sentiment score (-1 to 1)
	sentimentScore := 0.0
	if positiveCount+negativeCount > 0 {
		sentimentScore = float64(positiveCount-negativeCount) / float64(positiveCount+negativeCount)
	}
	
	// Check against threshold
	targetSentiment := config.Config["target"].(string)
	threshold := config.Threshold
	
	pass := false
	reason := fmt.Sprintf("Sentiment score: %.2f", sentimentScore)
	
	switch targetSentiment {
	case "positive":
		pass = sentimentScore >= threshold
	case "negative":
		pass = sentimentScore <= -threshold
	case "neutral":
		pass = math.Abs(sentimentScore) <= threshold
	}
	
	return MetricResult{
		Name:    config.Name,
		Score:   math.Abs(sentimentScore),
		Pass:    pass,
		Reason:  reason,
		Latency: latency,
		Details: map[string]interface{}{
			"sentiment_score":  sentimentScore,
			"positive_words":   positiveCount,
			"negative_words":   negativeCount,
			"target_sentiment": targetSentiment,
		},
	}, nil
}

// evaluateReadabilityMetric evaluates readability (Flesch Reading Ease approximation)
func (m *MetricEvaluator) evaluateReadabilityMetric(config MetricConfig, response string, latency time.Duration) (MetricResult, error) {
	sentences := strings.Split(response, ". ")
	words := strings.Fields(response)
	syllables := 0
	
	// Approximate syllable count
	for _, word := range words {
		syllables += countSyllables(word)
	}
	
	if len(sentences) == 0 || len(words) == 0 {
		return MetricResult{
			Name:    config.Name,
			Score:   0,
			Pass:    false,
			Reason:  "Insufficient text for readability analysis",
			Latency: latency,
		}, nil
	}
	
	// Flesch Reading Ease formula
	avgSentenceLength := float64(len(words)) / float64(len(sentences))
	avgSyllablesPerWord := float64(syllables) / float64(len(words))
	
	fleschScore := 206.835 - (1.015 * avgSentenceLength) - (84.6 * avgSyllablesPerWord)
	
	// Convert to grade level (approximate)
	gradeLevel := (0.39 * avgSentenceLength) + (11.8 * avgSyllablesPerWord) - 15.59
	
	minGrade := config.Config["min_grade_level"].(float64)
	maxGrade := config.Config["max_grade_level"].(float64)
	
	pass := gradeLevel >= minGrade && gradeLevel <= maxGrade
	score := 1.0
	
	if !pass {
		if gradeLevel < minGrade {
			score = gradeLevel / minGrade
		} else {
			score = maxGrade / gradeLevel
		}
	}
	
	return MetricResult{
		Name:    config.Name,
		Score:   score,
		Pass:    pass,
		Reason:  fmt.Sprintf("Grade level: %.1f (Flesch: %.1f)", gradeLevel, fleschScore),
		Latency: latency,
		Details: map[string]interface{}{
			"grade_level":             gradeLevel,
			"flesch_score":            fleschScore,
			"avg_sentence_length":     avgSentenceLength,
			"avg_syllables_per_word":  avgSyllablesPerWord,
			"min_grade":               minGrade,
			"max_grade":               maxGrade,
		},
	}, nil
}

// evaluateToxicityMetric evaluates toxicity (simplified)
func (m *MetricEvaluator) evaluateToxicityMetric(config MetricConfig, response string, latency time.Duration) (MetricResult, error) {
	toxicWords := []string{"hate", "stupid", "idiot", "kill", "die", "dumb", "moron", "loser", "pathetic"}
	
	response = strings.ToLower(response)
	toxicCount := 0
	
	for _, word := range toxicWords {
		if strings.Contains(response, word) {
			toxicCount++
		}
	}
	
	toxicityScore := float64(toxicCount) / float64(len(strings.Fields(response)))
	threshold := config.Threshold
	
	pass := toxicityScore <= threshold
	score := 1.0 - toxicityScore
	
	return MetricResult{
		Name:    config.Name,
		Score:   score,
		Pass:    pass,
		Reason:  fmt.Sprintf("Toxicity score: %.3f", toxicityScore),
		Latency: latency,
		Details: map[string]interface{}{
			"toxicity_score": toxicityScore,
			"toxic_words":    toxicCount,
			"threshold":      threshold,
		},
	}, nil
}

// evaluateCoherenceMetric evaluates coherence (simplified)
func (m *MetricEvaluator) evaluateCoherenceMetric(config MetricConfig, response string, latency time.Duration) (MetricResult, error) {
	sentences := strings.Split(response, ". ")
	if len(sentences) < 2 {
		return MetricResult{
			Name:    config.Name,
			Score:   1.0,
			Pass:    true,
			Reason:  "Single sentence response",
			Latency: latency,
		}, nil
	}
	
	// Simple coherence check based on transition words and repetition
	transitionWords := []string{"however", "therefore", "additionally", "furthermore", "moreover", "consequently", "thus", "hence"}
	
	transitionCount := 0
	for _, word := range transitionWords {
		if strings.Contains(strings.ToLower(response), word) {
			transitionCount++
		}
	}
	
	// Check for excessive repetition
	words := strings.Fields(strings.ToLower(response))
	wordCount := make(map[string]int)
	for _, word := range words {
		wordCount[word]++
	}
	
	repetitionScore := 0.0
	for _, count := range wordCount {
		if count > 1 {
			repetitionScore += float64(count-1) / float64(len(words))
		}
	}
	
	coherenceScore := (float64(transitionCount) / float64(len(sentences))) - repetitionScore
	if coherenceScore < 0 {
		coherenceScore = 0
	}
	if coherenceScore > 1 {
		coherenceScore = 1
	}
	
	pass := coherenceScore >= config.Threshold
	
	return MetricResult{
		Name:    config.Name,
		Score:   coherenceScore,
		Pass:    pass,
		Reason:  fmt.Sprintf("Coherence score: %.2f", coherenceScore),
		Latency: latency,
		Details: map[string]interface{}{
			"coherence_score":   coherenceScore,
			"transition_words":  transitionCount,
			"repetition_score":  repetitionScore,
			"sentences":         len(sentences),
		},
	}, nil
}

// evaluateRelevanceMetric evaluates relevance to prompt
func (m *MetricEvaluator) evaluateRelevanceMetric(config MetricConfig, prompt, response string, latency time.Duration) (MetricResult, error) {
	// Simple relevance based on keyword overlap
	promptWords := extractKeywords(strings.ToLower(prompt))
	responseWords := extractKeywords(strings.ToLower(response))
	
	overlap := 0
	for word := range promptWords {
		if responseWords[word] {
			overlap++
		}
	}
	
	relevanceScore := 0.0
	if len(promptWords) > 0 {
		relevanceScore = float64(overlap) / float64(len(promptWords))
	}
	
	pass := relevanceScore >= config.Threshold
	
	return MetricResult{
		Name:    config.Name,
		Score:   relevanceScore,
		Pass:    pass,
		Reason:  fmt.Sprintf("Relevance score: %.2f (%d/%d keywords)", relevanceScore, overlap, len(promptWords)),
		Latency: latency,
		Details: map[string]interface{}{
			"relevance_score": relevanceScore,
			"keyword_overlap": overlap,
			"prompt_keywords": len(promptWords),
		},
	}, nil
}

// evaluatePythonMetric evaluates using Python script
func (m *MetricEvaluator) evaluatePythonMetric(ctx context.Context, config MetricConfig, prompt, response string, metadata map[string]interface{}) (MetricResult, error) {
	start := time.Now()
	
	// Create input data for the script
	input := map[string]interface{}{
		"prompt":   prompt,
		"response": response,
		"metadata": metadata,
		"config":   config.Config,
	}
	
	inputJSON, err := json.Marshal(input)
	if err != nil {
		return MetricResult{}, fmt.Errorf("error marshaling input: %v", err)
	}
	
	// Execute Python script
	cmd := exec.CommandContext(ctx, "python3", "-c", config.Script)
	cmd.Stdin = strings.NewReader(string(inputJSON))
	
	output, err := cmd.Output()
	if err != nil {
		return MetricResult{}, fmt.Errorf("error executing Python script: %v", err)
	}
	
	// Parse result
	var result map[string]interface{}
	if err := json.Unmarshal(output, &result); err != nil {
		return MetricResult{}, fmt.Errorf("error parsing Python script output: %v", err)
	}
	
	return MetricResult{
		Name:    config.Name,
		Score:   result["score"].(float64),
		Pass:    result["pass"].(bool),
		Reason:  result["reason"].(string),
		Latency: time.Since(start),
		Details: result["details"].(map[string]interface{}),
	}, nil
}

// evaluateScriptMetric evaluates using external script
func (m *MetricEvaluator) evaluateScriptMetric(ctx context.Context, config MetricConfig, prompt, response string, metadata map[string]interface{}) (MetricResult, error) {
	start := time.Now()
	
	// Execute script with arguments
	cmd := exec.CommandContext(ctx, config.Script, prompt, response)
	output, err := cmd.Output()
	if err != nil {
		return MetricResult{}, fmt.Errorf("error executing script: %v", err)
	}
	
	// Parse result (expecting JSON output)
	var result map[string]interface{}
	if err := json.Unmarshal(output, &result); err != nil {
		return MetricResult{}, fmt.Errorf("error parsing script output: %v", err)
	}
	
	return MetricResult{
		Name:    config.Name,
		Score:   result["score"].(float64),
		Pass:    result["pass"].(bool),
		Reason:  result["reason"].(string),
		Latency: time.Since(start),
		Details: result["details"].(map[string]interface{}),
	}, nil
}

// evaluateLLMMetric evaluates using another LLM as judge
func (m *MetricEvaluator) evaluateLLMMetric(ctx context.Context, config MetricConfig, prompt, response string, metadata map[string]interface{}) (MetricResult, error) {
	start := time.Now()
	
	provider, exists := m.providers[config.Judge]
	if !exists {
		return MetricResult{}, fmt.Errorf("judge provider not found: %s", config.Judge)
	}
	
	// Create judge prompt
	judgePrompt := fmt.Sprintf(`%s

Original Prompt: %s
Response to Evaluate: %s

Please evaluate the response according to the criteria and provide a JSON response with:
- score: a number between 0 and 1
- pass: true/false based on whether the response meets the criteria
- reason: explanation of the evaluation

JSON Response:`, config.Prompt, prompt, response)
	
	// Evaluate with judge
	judgeResponse, err := provider.EvaluatePrompt(ctx, judgePrompt, map[string]interface{}{})
	if err != nil {
		return MetricResult{}, fmt.Errorf("error evaluating with judge: %v", err)
	}
	
	// Parse judge response
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(judgeResponse.Output), &result); err != nil {
		// Fallback: try to extract JSON from response
		re := regexp.MustCompile(`\{[^}]*\}`)
		jsonMatch := re.FindString(judgeResponse.Output)
		if jsonMatch == "" {
			return MetricResult{}, fmt.Errorf("could not parse judge response as JSON")
		}
		if err := json.Unmarshal([]byte(jsonMatch), &result); err != nil {
			return MetricResult{}, fmt.Errorf("error parsing judge response: %v", err)
		}
	}
	
	return MetricResult{
		Name:    config.Name,
		Score:   result["score"].(float64),
		Pass:    result["pass"].(bool),
		Reason:  result["reason"].(string),
		Latency: time.Since(start),
		Details: map[string]interface{}{
			"judge":          config.Judge,
			"judge_response": judgeResponse.Output,
			"judge_cost":     judgeResponse.Cost,
		},
	}, nil
}

// evaluateRegexMetric evaluates using regex pattern matching
func (m *MetricEvaluator) evaluateRegexMetric(config MetricConfig, prompt, response string) (MetricResult, error) {
	start := time.Now()
	pattern := config.Pattern
	re, err := regexp.Compile(pattern)
	if err != nil {
		return MetricResult{}, fmt.Errorf("invalid regex pattern: %v", err)
	}
	
	matches := re.FindAllString(response, -1)
	matchCount := len(matches)
	
	pass := matchCount > 0
	score := 0.0
	if pass {
		score = 1.0
	}
	
	// Check for expected count if specified
	if expectedCount, exists := config.Config["expected_count"]; exists {
		expected := int(expectedCount.(float64))
		pass = matchCount == expected
		if matchCount > 0 {
			score = math.Min(1.0, float64(matchCount)/float64(expected))
		}
	}
	
	return MetricResult{
		Name:    config.Name,
		Score:   score,
		Pass:    pass,
		Reason:  fmt.Sprintf("Regex matches: %d", matchCount),
		Latency: time.Since(start),
		Details: map[string]interface{}{
			"pattern":     pattern,
			"matches":     matches,
			"match_count": matchCount,
		},
	}, nil
}

// Helper functions

// countSyllables approximates syllable count for readability metrics
func countSyllables(word string) int {
	word = strings.ToLower(word)
	vowels := "aeiouy"
	syllableCount := 0
	previousWasVowel := false
	
	for _, char := range word {
		isVowel := strings.ContainsRune(vowels, char)
		if isVowel && !previousWasVowel {
			syllableCount++
		}
		previousWasVowel = isVowel
	}
	
	// Handle silent 'e'
	if strings.HasSuffix(word, "e") && syllableCount > 1 {
		syllableCount--
	}
	
	// Ensure at least one syllable
	if syllableCount == 0 {
		syllableCount = 1
	}
	
	return syllableCount
}

// extractKeywords extracts meaningful keywords from text
func extractKeywords(text string) map[string]bool {
	// Remove common stop words
	stopWords := map[string]bool{
		"the": true, "a": true, "an": true, "and": true, "or": true, "but": true,
		"in": true, "on": true, "at": true, "to": true, "for": true, "of": true,
		"with": true, "by": true, "is": true, "are": true, "was": true, "were": true,
		"be": true, "been": true, "have": true, "has": true, "had": true, "do": true,
		"does": true, "did": true, "will": true, "would": true, "could": true, "should": true,
		"can": true, "may": true, "might": true, "must": true, "shall": true,
		"this": true, "that": true, "these": true, "those": true, "i": true, "you": true,
		"he": true, "she": true, "it": true, "we": true, "they": true, "me": true,
		"him": true, "her": true, "us": true, "them": true, "my": true, "your": true,
		"his": true, "our": true, "their": true,
	}
	
	words := strings.Fields(text)
	keywords := make(map[string]bool)
	
	for _, word := range words {
		// Remove punctuation
		word = regexp.MustCompile(`[^\w]`).ReplaceAllString(word, "")
		word = strings.ToLower(word)
		
		// Skip if stop word or too short
		if len(word) < 3 || stopWords[word] {
			continue
		}
		
		keywords[word] = true
	}
	
	return keywords
}

// evaluatePassAtNMetric evaluates pass@n metric for code generation tasks
func (m *MetricEvaluator) evaluatePassAtNMetric(ctx context.Context, config MetricConfig, prompt, response string, metadata map[string]interface{}) (MetricResult, error) {
	_ = time.Now() // start
	
	// Get configuration parameters
	n := 1
	if nVal, ok := config.Config["n"].(float64); ok {
		n = int(nVal)
	}
	
	// Get test cases from config
	testCases, hasTestCases := config.Config["test_cases"].([]interface{})
	
	// Get samples - either from metadata or generate multiple samples
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
		// If no samples provided, use the single response
		samples = []string{response}
	}
	
	// Create advanced metrics instance
	am := NewAdvancedMetrics(m.providers[config.Config["provider"].(string)])
	
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
		if validationPrompt, ok := config.Config["validation_prompt"].(string); ok {
			testFunc = func(code string) bool {
				evalPrompt := fmt.Sprintf(validationPrompt, code)
				response, err := m.providers[config.Config["provider"].(string)].Generate(ctx, evalPrompt, llm.GenerateOptions{
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
	
	// Determine if metric passes based on threshold
	threshold := 0.5
	if config.Threshold > 0 {
		threshold = config.Threshold
	}
	
	pass := result.PassRate >= threshold
	
	return MetricResult{
		Name:    config.Name,
		Score:   result.PassRate,
		Pass:    pass,
		Reason:  fmt.Sprintf("Pass@%d rate: %.2f%% (%d/%d samples passed)", n, result.PassRate*100, result.NumPassed, result.NumSamples),
		Latency: result.Duration,
		Details: map[string]interface{}{
			"n":              n,
			"pass_rate":      result.PassRate,
			"num_samples":    result.NumSamples,
			"num_passed":     result.NumPassed,
			"threshold":      threshold,
			"passed_indices": result.Details["passed_indices"],
		},
	}, nil
}