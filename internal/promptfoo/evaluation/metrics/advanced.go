package metrics

import (
	"context"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/tmc/pe/internal/llm"
)

// Package-level convenience functions for backward compatibility
func CalculateBLEU(generated, reference string, maxN int) *AdvancedMetricResult {
	am := &AdvancedMetrics{}
	return am.CalculateBLEU(generated, reference, maxN)
}

func CalculateROUGE(generated, reference string, rougeType string) *AdvancedMetricResult {
	am := &AdvancedMetrics{}
	return am.CalculateROUGE(generated, reference, rougeType)
}

func CalculateROUGEDetailed(generated, reference string) *AdvancedMetricResult {
	am := &AdvancedMetrics{}
	return am.CalculateROUGE(generated, reference, "L")
}

func CalculateMETEOR(generated, reference string) *AdvancedMetricResult {
	am := &AdvancedMetrics{}
	return am.CalculateMETEOR(generated, reference)
}

func CalculateBERTScoreDetailed(generated, reference string, llmProvider llm.Provider) *AdvancedMetricResult {
	am := NewAdvancedMetrics(llmProvider)
	return am.CalculateBERTScore(context.Background(), generated, reference)
}

func CalculateGEval(generated, reference, criteria string, llmProvider llm.Provider) *AdvancedMetricResult {
	am := NewAdvancedMetrics(llmProvider)
	return am.CalculateGEval(context.Background(), generated, reference, criteria)
}

func CalculateUniEval(generated, reference, taskType string, llmProvider llm.Provider) *AdvancedMetricResult {
	am := NewAdvancedMetrics(llmProvider)
	return am.CalculateUniEval(context.Background(), generated, reference, taskType)
}

func CalculatePassAtNDetailed(n int, samples []string, testFunc func(string) bool) *PassAtNResult {
	am := &AdvancedMetrics{}
	return am.CalculatePassAtN(n, samples, testFunc)
}

// AdvancedMetrics provides state-of-the-art evaluation metrics for prompt engineering
type AdvancedMetrics struct {
	llm llm.Provider
}

// NewAdvancedMetrics creates a new advanced metrics evaluator
func NewAdvancedMetrics(llmProvider llm.Provider) *AdvancedMetrics {
	return &AdvancedMetrics{
		llm: llmProvider,
	}
}

// AdvancedMetricResult represents the result of an advanced metric evaluation
type AdvancedMetricResult struct {
	MetricName   string                 `json:"metric_name"`
	Score        float64                `json:"score"`
	Confidence   float64                `json:"confidence,omitempty"`
	Details      map[string]interface{} `json:"details,omitempty"`
	Duration     time.Duration          `json:"duration"`
	ErrorMessage string                 `json:"error_message,omitempty"`
}

// BLEU Score Implementation
// Bilingual Evaluation Understudy - measures n-gram precision between generated and reference text
func (am *AdvancedMetrics) CalculateBLEU(generated, reference string, maxN int) *AdvancedMetricResult {
	start := time.Now()

	if maxN <= 0 {
		maxN = 4
	}

	generatedTokens := tokenize(generated)
	referenceTokens := tokenize(reference)

	if len(generatedTokens) == 0 {
		return &AdvancedMetricResult{
			MetricName:   "BLEU",
			Score:        0.0,
			Duration:     time.Since(start),
			ErrorMessage: "Generated text is empty",
		}
	}

	// Calculate n-gram precisions with smoothing
	precisions := make([]float64, maxN)
	smoothing := 1.0 // Add-one smoothing

	for n := 1; n <= maxN; n++ {
		precision := am.calculateNGramPrecision(generatedTokens, referenceTokens, n)
		// Apply smoothing for higher order n-grams
		if n > 1 && precision == 0 {
			// Add smoothing only if we have enough tokens
			if len(generatedTokens) >= n {
				precision = smoothing / float64(len(generatedTokens)-n+1)
			}
		}
		precisions[n-1] = precision
	}

	// Geometric mean of precisions
	geometricMean := 1.0
	for i, p := range precisions {
		// Use modified precision for geometric mean
		// This prevents zero scores when higher n-grams don't match
		weight := 1.0 / float64(maxN)
		if p == 0 && i > 0 {
			// For higher order n-grams that don't match, use a small value
			p = 0.01
		}
		geometricMean *= math.Pow(p, weight)
	}

	// Brevity penalty
	bp := am.calculateBrevityPenalty(len(generatedTokens), len(referenceTokens))
	bleuScore := bp * geometricMean

	return &AdvancedMetricResult{
		MetricName: "BLEU",
		Score:      bleuScore,
		Duration:   time.Since(start),
		Details: map[string]interface{}{
			"precisions":       precisions,
			"brevity_penalty":  bp,
			"geometric_mean":   geometricMean,
			"generated_length": len(generatedTokens),
			"reference_length": len(referenceTokens),
		},
	}
}

// ROUGE Score Implementation
// Recall-Oriented Understudy for Gisting Evaluation - measures n-gram recall
func (am *AdvancedMetrics) CalculateROUGE(generated, reference string, rougeType string) *AdvancedMetricResult {
	start := time.Now()

	generatedTokens := tokenize(generated)
	referenceTokens := tokenize(reference)

	var score float64
	var details map[string]interface{}

	switch strings.ToUpper(rougeType) {
	case "ROUGE-1", "1":
		score, details = am.calculateROUGE1(generatedTokens, referenceTokens)
	case "ROUGE-2", "2":
		score, details = am.calculateROUGE2(generatedTokens, referenceTokens)
	case "ROUGE-L", "L":
		score, details = am.calculateROUGEL(generatedTokens, referenceTokens)
	case "ROUGE-W", "W":
		score, details = am.calculateROUGEW(generatedTokens, referenceTokens)
	default:
		return &AdvancedMetricResult{
			MetricName:   fmt.Sprintf("ROUGE-%s", rougeType),
			Score:        0.0,
			Duration:     time.Since(start),
			ErrorMessage: fmt.Sprintf("Unsupported ROUGE type: %s", rougeType),
		}
	}

	return &AdvancedMetricResult{
		MetricName: fmt.Sprintf("ROUGE-%s", strings.ToUpper(rougeType)),
		Score:      score,
		Duration:   time.Since(start),
		Details:    details,
	}
}

// METEOR Score Implementation
// Metric for Evaluation of Translation with Explicit Ordering
func (am *AdvancedMetrics) CalculateMETEOR(generated, reference string) *AdvancedMetricResult {
	start := time.Now()

	generatedTokens := tokenize(generated)
	referenceTokens := tokenize(reference)

	if len(generatedTokens) == 0 || len(referenceTokens) == 0 {
		return &AdvancedMetricResult{
			MetricName:   "METEOR",
			Score:        0.0,
			Duration:     time.Since(start),
			ErrorMessage: "Empty input text",
		}
	}

	// Find matches (exact and synonym matches)
	matches := am.findMatches(generatedTokens, referenceTokens)

	// Calculate precision and recall
	precision := float64(matches) / float64(len(generatedTokens))
	recall := float64(matches) / float64(len(referenceTokens))

	// F-mean (harmonic mean weighted towards recall)
	fMean := 0.0
	if precision+recall > 0 {
		fMean = (10 * precision * recall) / (recall + 9*precision)
	}

	// Fragmentation penalty (simplified version)
	chunks := am.countChunks(generatedTokens, referenceTokens)
	fragPenalty := 0.0
	if matches > 0 {
		fragPenalty = 0.5 * math.Pow(float64(chunks)/float64(matches), 3)
	}

	meteorScore := fMean * (1 - fragPenalty)

	return &AdvancedMetricResult{
		MetricName: "METEOR",
		Score:      meteorScore,
		Duration:   time.Since(start),
		Details: map[string]interface{}{
			"precision":             precision,
			"recall":                recall,
			"f_mean":                fMean,
			"matches":               matches,
			"chunks":                chunks,
			"fragmentation_penalty": fragPenalty,
		},
	}
}

// BERTScore Implementation (Simplified)
// Uses semantic similarity rather than exact token matching
func (am *AdvancedMetrics) CalculateBERTScore(ctx context.Context, generated, reference string) *AdvancedMetricResult {
	start := time.Now()

	// Since we don't have BERT embeddings, we'll use LLM-based semantic similarity
	similarityPrompt := fmt.Sprintf(`Compare the semantic similarity between these two texts on a scale of 0.0 to 1.0.

REFERENCE TEXT:
%s

GENERATED TEXT:  
%s

Consider:
1. Semantic meaning and content similarity
2. Factual consistency
3. Key concept overlap
4. Overall information preservation

Provide only a numeric score between 0.0 and 1.0.`, reference, generated)

	response, err := am.llm.Generate(ctx, similarityPrompt, llm.GenerateOptions{
		Temperature: &[]float64{0.1}[0], // Low temperature for consistent scoring
	})

	if err != nil {
		return &AdvancedMetricResult{
			MetricName:   "BERTScore",
			Score:        0.0,
			Duration:     time.Since(start),
			ErrorMessage: fmt.Sprintf("Failed to calculate semantic similarity: %v", err),
		}
	}

	// Parse the numeric score
	scoreText := strings.TrimSpace(response.Text)
	score := parseFloatFromText(scoreText, 0.0)

	// Calculate token-level alignments as additional detail
	generatedTokens := tokenize(generated)
	referenceTokens := tokenize(reference)
	tokenAlignments := am.calculateTokenAlignments(generatedTokens, referenceTokens)

	return &AdvancedMetricResult{
		MetricName: "BERTScore",
		Score:      score,
		Confidence: 0.8, // LLM-based scoring has inherent uncertainty
		Duration:   time.Since(start),
		Details: map[string]interface{}{
			"semantic_similarity": score,
			"token_alignments":    tokenAlignments,
			"generated_tokens":    len(generatedTokens),
			"reference_tokens":    len(referenceTokens),
			"evaluation_method":   "LLM-based semantic similarity",
		},
	}
}

// G-Eval Implementation
// LLM-based evaluation framework with better human alignment
func (am *AdvancedMetrics) CalculateGEval(ctx context.Context, generated, reference, criteria string) *AdvancedMetricResult {
	start := time.Now()

	// Create G-Eval prompt with chain-of-thought reasoning
	gEvalPrompt := fmt.Sprintf(`You are an expert evaluator. Evaluate the GENERATED TEXT against the REFERENCE TEXT using the given CRITERIA.

CRITERIA: %s

REFERENCE TEXT:
%s

GENERATED TEXT:
%s

EVALUATION STEPS:
1. Analyze the generated text for the specific criteria
2. Compare key aspects with the reference text
3. Consider both strengths and weaknesses
4. Assign a score from 1-5 where:
   - 5: Excellent - Fully meets criteria, high quality
   - 4: Good - Mostly meets criteria, minor issues
   - 3: Fair - Partially meets criteria, some issues
   - 2: Poor - Minimally meets criteria, significant issues  
   - 1: Very Poor - Fails to meet criteria

REASONING PROCESS:
[Think step by step about how the generated text performs on the criteria]

SCORE: [1-5]
EXPLANATION: [Brief explanation of the score]`, criteria, reference, generated)

	response, err := am.llm.Generate(ctx, gEvalPrompt, llm.GenerateOptions{
		Temperature: &[]float64{0.3}[0], // Moderate temperature for balanced evaluation
	})

	if err != nil {
		return &AdvancedMetricResult{
			MetricName:   "G-Eval",
			Score:        0.0,
			Duration:     time.Since(start),
			ErrorMessage: fmt.Sprintf("Failed to calculate G-Eval score: %v", err),
		}
	}

	// Parse the structured response
	score, explanation := am.parseGEvalResponse(response.Text)
	normalizedScore := (score - 1) / 4 // Convert 1-5 scale to 0-1 scale

	return &AdvancedMetricResult{
		MetricName: "G-Eval",
		Score:      normalizedScore,
		Confidence: 0.9, // G-Eval typically has good human alignment
		Duration:   time.Since(start),
		Details: map[string]interface{}{
			"raw_score":         score,
			"criteria":          criteria,
			"explanation":       explanation,
			"evaluation_method": "G-Eval with chain-of-thought",
			"full_response":     response.Text,
		},
	}
}

// UniEval Implementation (Unified evaluation for text generation)
func (am *AdvancedMetrics) CalculateUniEval(ctx context.Context, generated, reference, task string) *AdvancedMetricResult {
	start := time.Now()

	// Task-specific evaluation dimensions
	dimensions := am.getUniEvalDimensions(task)

	var totalScore float64
	var dimensionScores = make(map[string]float64)

	for _, dimension := range dimensions {
		dimensionPrompt := fmt.Sprintf(`Evaluate the following generated text on the dimension of %s.

TASK: %s
DIMENSION: %s

REFERENCE:
%s

GENERATED:
%s

Rate from 1-10 how well the generated text performs on this dimension.
Provide only a numeric score.`, dimension, task, dimension, reference, generated)

		response, err := am.llm.Generate(ctx, dimensionPrompt, llm.GenerateOptions{
			Temperature: &[]float64{0.2}[0],
		})

		if err != nil {
			continue // Skip failed dimensions
		}

		score := parseFloatFromText(response.Text, 5.0) / 10.0 // Normalize to 0-1
		dimensionScores[dimension] = score
		totalScore += score
	}

	if len(dimensionScores) == 0 {
		return &AdvancedMetricResult{
			MetricName:   "UniEval",
			Score:        0.0,
			Duration:     time.Since(start),
			ErrorMessage: "Failed to evaluate any dimensions",
		}
	}

	avgScore := totalScore / float64(len(dimensionScores))

	return &AdvancedMetricResult{
		MetricName: "UniEval",
		Score:      avgScore,
		Duration:   time.Since(start),
		Details: map[string]interface{}{
			"task":             task,
			"dimensions":       dimensions,
			"dimension_scores": dimensionScores,
			"num_dimensions":   len(dimensionScores),
		},
	}
}

// Helper methods

func (am *AdvancedMetrics) calculateNGramPrecision(generated, reference []string, n int) float64 {
	if len(generated) < n {
		return 0.0
	}

	generatedNGrams := extractNGrams(generated, n)
	referenceNGrams := extractNGrams(reference, n)

	totalGenerated := sum(generatedNGrams)
	if totalGenerated == 0 {
		return 0.0
	}

	matches := 0
	for ngram := range generatedNGrams {
		if count, exists := referenceNGrams[ngram]; exists {
			matches += min(generatedNGrams[ngram], count)
		}
	}

	return float64(matches) / float64(totalGenerated)
}

func (am *AdvancedMetrics) calculateBrevityPenalty(genLen, refLen int) float64 {
	if genLen >= refLen {
		return 1.0
	}
	return math.Exp(1.0 - float64(refLen)/float64(genLen))
}

func (am *AdvancedMetrics) calculateROUGE1(generated, reference []string) (float64, map[string]interface{}) {
	if len(reference) == 0 {
		return 0.0, map[string]interface{}{"error": "empty reference"}
	}

	generatedSet := make(map[string]bool)
	for _, token := range generated {
		generatedSet[token] = true
	}

	matches := 0
	for _, token := range reference {
		if generatedSet[token] {
			matches++
		}
	}

	recall := float64(matches) / float64(len(reference))
	precision := 0.0
	if len(generated) > 0 {
		precision = float64(matches) / float64(len(generated))
	}

	f1 := 0.0
	if precision+recall > 0 {
		f1 = 2 * precision * recall / (precision + recall)
	}

	return f1, map[string]interface{}{
		"precision": precision,
		"recall":    recall,
		"matches":   matches,
	}
}

func (am *AdvancedMetrics) calculateROUGE2(generated, reference []string) (float64, map[string]interface{}) {
	genBigrams := extractNGrams(generated, 2)
	refBigrams := extractNGrams(reference, 2)

	if len(refBigrams) == 0 {
		return 0.0, map[string]interface{}{"error": "no reference bigrams"}
	}

	matches := 0
	for bigram := range refBigrams {
		if genBigrams[bigram] > 0 {
			matches += min(genBigrams[bigram], refBigrams[bigram])
		}
	}

	recall := float64(matches) / float64(sum(refBigrams))
	precision := 0.0
	if sum(genBigrams) > 0 {
		precision = float64(matches) / float64(sum(genBigrams))
	}

	f1 := 0.0
	if precision+recall > 0 {
		f1 = 2 * precision * recall / (precision + recall)
	}

	return f1, map[string]interface{}{
		"precision":   precision,
		"recall":      recall,
		"matches":     matches,
		"ref_bigrams": len(refBigrams),
		"gen_bigrams": len(genBigrams),
	}
}

func (am *AdvancedMetrics) calculateROUGEL(generated, reference []string) (float64, map[string]interface{}) {
	// Longest Common Subsequence
	lcs := am.longestCommonSubsequence(generated, reference)

	if len(reference) == 0 || len(generated) == 0 {
		return 0.0, map[string]interface{}{"error": "empty input"}
	}

	recall := float64(lcs) / float64(len(reference))
	precision := float64(lcs) / float64(len(generated))

	f1 := 0.0
	if precision+recall > 0 {
		f1 = 2 * precision * recall / (precision + recall)
	}

	return f1, map[string]interface{}{
		"precision": precision,
		"recall":    recall,
		"lcs":       lcs,
	}
}

func (am *AdvancedMetrics) calculateROUGEW(generated, reference []string) (float64, map[string]interface{}) {
	// Weighted Longest Common Subsequence (simplified implementation)
	wlcs := am.weightedLCS(generated, reference)

	if len(reference) == 0 || len(generated) == 0 {
		return 0.0, map[string]interface{}{"error": "empty input"}
	}

	recall := wlcs / float64(len(reference)*len(reference))
	precision := wlcs / float64(len(generated)*len(generated))

	f1 := 0.0
	if precision+recall > 0 {
		f1 = 2 * precision * recall / (precision + recall)
	}

	return f1, map[string]interface{}{
		"precision": precision,
		"recall":    recall,
		"wlcs":      wlcs,
	}
}

func (am *AdvancedMetrics) longestCommonSubsequence(seq1, seq2 []string) int {
	m, n := len(seq1), len(seq2)
	dp := make([][]int, m+1)
	for i := range dp {
		dp[i] = make([]int, n+1)
	}

	for i := 1; i <= m; i++ {
		for j := 1; j <= n; j++ {
			if seq1[i-1] == seq2[j-1] {
				dp[i][j] = dp[i-1][j-1] + 1
			} else {
				dp[i][j] = max(dp[i-1][j], dp[i][j-1])
			}
		}
	}

	return dp[m][n]
}

func (am *AdvancedMetrics) weightedLCS(seq1, seq2 []string) float64 {
	// Simplified weighted LCS - weights based on consecutive matches
	m, n := len(seq1), len(seq2)
	if m == 0 || n == 0 {
		return 0
	}

	dp := make([][]float64, m+1)
	for i := range dp {
		dp[i] = make([]float64, n+1)
	}

	for i := 1; i <= m; i++ {
		for j := 1; j <= n; j++ {
			if seq1[i-1] == seq2[j-1] {
				// Weight increases for consecutive matches
				weight := 1.0
				if i > 1 && j > 1 && seq1[i-2] == seq2[j-2] {
					weight = 2.0
				}
				dp[i][j] = dp[i-1][j-1] + weight
			} else {
				dp[i][j] = math.Max(dp[i-1][j], dp[i][j-1])
			}
		}
	}

	return dp[m][n]
}

func (am *AdvancedMetrics) findMatches(generated, reference []string) int {
	referenceSet := make(map[string]bool)
	for _, token := range reference {
		referenceSet[token] = true
	}

	matches := 0
	for _, token := range generated {
		if referenceSet[token] {
			matches++
		}
	}

	return matches
}

func (am *AdvancedMetrics) countChunks(generated, reference []string) int {
	// Count matching chunks - consecutive sequences of matching tokens
	if len(generated) == 0 || len(reference) == 0 {
		return 0
	}

	// Create a map of reference tokens for quick lookup
	refMap := make(map[string]bool)
	for _, token := range reference {
		refMap[token] = true
	}

	// Count chunks of consecutive matching tokens
	chunks := 0
	inChunk := false

	for _, token := range generated {
		if refMap[token] {
			if !inChunk {
				chunks++
				inChunk = true
			}
		} else {
			inChunk = false
		}
	}

	// Ensure at least 1 chunk if there are any matches
	if chunks == 0 && am.findMatches(generated, reference) > 0 {
		chunks = 1
	}

	return chunks
}

func (am *AdvancedMetrics) calculateTokenAlignments(generated, reference []string) map[string]interface{} {
	// Simple token-level alignment calculation
	alignments := make(map[string]int)

	for _, genToken := range generated {
		for _, refToken := range reference {
			if genToken == refToken {
				alignments[genToken]++
			}
		}
	}

	return map[string]interface{}{
		"exact_matches":     len(alignments),
		"alignment_details": alignments,
	}
}

func (am *AdvancedMetrics) parseGEvalResponse(response string) (float64, string) {
	lines := strings.Split(response, "\n")
	score := 3.0 // default
	explanation := "No explanation provided"

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "SCORE:") {
			scoreText := strings.TrimSpace(line[6:])
			score = parseFloatFromText(scoreText, 3.0)
		} else if strings.HasPrefix(line, "EXPLANATION:") {
			explanation = strings.TrimSpace(line[12:])
		}
	}

	return score, explanation
}

func (am *AdvancedMetrics) getUniEvalDimensions(task string) []string {
	switch strings.ToLower(task) {
	case "summarization":
		return []string{"relevance", "consistency", "fluency", "coherence"}
	case "dialogue":
		return []string{"naturalness", "coherence", "engagingness", "groundedness"}
	case "translation":
		return []string{"fluency", "adequacy", "consistency"}
	case "data2text":
		return []string{"naturalness", "informativeness", "relevance"}
	default:
		return []string{"relevance", "coherence", "fluency", "factuality"}
	}
}

// Utility functions

func tokenize(text string) []string {
	// Simple tokenization - split on whitespace and punctuation
	re := regexp.MustCompile(`\w+`)
	tokens := re.FindAllString(strings.ToLower(text), -1)
	return tokens
}

func extractNGrams(tokens []string, n int) map[string]int {
	ngrams := make(map[string]int)

	for i := 0; i <= len(tokens)-n; i++ {
		ngram := strings.Join(tokens[i:i+n], " ")
		ngrams[ngram]++
	}

	return ngrams
}

func parseFloatFromText(text string, defaultValue float64) float64 {
	// Extract first number from text
	re := regexp.MustCompile(`\d+\.?\d*`)
	match := re.FindString(text)
	if match == "" {
		return defaultValue
	}

	if val, err := parseFloat(match); err == nil {
		return val
	}

	return defaultValue
}

func parseFloat(s string) (float64, error) {
	// Simple float parsing
	var result float64
	var err error

	if strings.Contains(s, ".") {
		result, err = strconv.ParseFloat(s, 64)
	} else {
		if i, parseErr := strconv.Atoi(s); parseErr == nil {
			result = float64(i)
		} else {
			err = parseErr
		}
	}

	return result, err
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func sum(m map[string]int) int {
	total := 0
	for _, v := range m {
		total += v
	}
	return total
}

// PassAtNResult represents the result of a pass@n metric evaluation
type PassAtNResult struct {
	N          int                    `json:"n"`
	PassRate   float64                `json:"pass_rate"`
	NumSamples int                    `json:"num_samples"`
	NumPassed  int                    `json:"num_passed"`
	Details    map[string]interface{} `json:"details,omitempty"`
	Duration   time.Duration          `json:"duration"`
}

// CalculatePassAtN evaluates the pass@n metric for code generation tasks
// n: number of attempts to consider
// samples: generated code samples
// testFunc: function that tests if a code sample passes (returns true if passes)
func (am *AdvancedMetrics) CalculatePassAtN(n int, samples []string, testFunc func(string) bool) *PassAtNResult {
	start := time.Now()

	if n <= 0 {
		n = 1
	}

	numSamples := len(samples)
	if numSamples == 0 {
		return &PassAtNResult{
			N:          n,
			PassRate:   0.0,
			NumSamples: 0,
			NumPassed:  0,
			Duration:   time.Since(start),
		}
	}

	// Test each sample
	numPassed := 0
	passedIndices := []int{}
	for i, sample := range samples {
		if testFunc(sample) {
			numPassed++
			passedIndices = append(passedIndices, i)
		}
	}

	// Calculate pass@n using the standard formula
	passRate := am.calculatePassAtNRate(numSamples, numPassed, n)

	return &PassAtNResult{
		N:          n,
		PassRate:   passRate,
		NumSamples: numSamples,
		NumPassed:  numPassed,
		Details: map[string]interface{}{
			"passed_indices": passedIndices,
			"formula":        "1 - C(numSamples-numPassed, n) / C(numSamples, n)",
		},
		Duration: time.Since(start),
	}
}

// CalculatePassAtNWithTests evaluates pass@n metric using test cases
// n: number of attempts to consider
// samples: generated code samples
// testCases: array of test inputs and expected outputs
func (am *AdvancedMetrics) CalculatePassAtNWithTests(ctx context.Context, n int, samples []string, testCases []map[string]interface{}) *PassAtNResult {
	_ = time.Now() // start

	testFunc := func(code string) bool {
		// Test the code against all test cases
		for _, testCase := range testCases {
			input, _ := testCase["input"].(string)
			expected, _ := testCase["expected"].(string)

			// Use LLM to evaluate if the code produces the expected output
			evalPrompt := fmt.Sprintf(`Execute the following code with the given input and determine if it produces the expected output.

CODE:
%s

INPUT:
%s

EXPECTED OUTPUT:
%s

Does the code produce the expected output? Reply with only "YES" or "NO".`, code, input, expected)

			response, err := am.llm.Generate(ctx, evalPrompt, llm.GenerateOptions{
				Temperature: &[]float64{0.0}[0],
			})

			if err != nil || !strings.Contains(strings.ToUpper(response.Text), "YES") {
				return false
			}
		}
		return true
	}

	result := am.CalculatePassAtN(n, samples, testFunc)
	result.Details["test_cases"] = len(testCases)
	return result
}

// calculatePassAtNRate computes the pass@n rate using the standard formula
// Formula: pass@n = 1 - C(numSamples-numPassed, n) / C(numSamples, n)
func (am *AdvancedMetrics) calculatePassAtNRate(numSamples, numPassed, n int) float64 {
	if n > numSamples {
		n = numSamples
	}

	if numPassed == numSamples {
		return 1.0
	}

	if numPassed == 0 {
		return 0.0
	}

	// Calculate using the standard pass@n formula
	// This avoids the bias from simply taking the top n samples
	numerator := binomialCoeff(numSamples-numPassed, n)
	denominator := binomialCoeff(numSamples, n)

	if denominator == 0 {
		return 0.0
	}

	return 1.0 - (float64(numerator) / float64(denominator))
}

// binomialCoeff calculates the binomial coefficient C(n, k)
func binomialCoeff(n, k int) int64 {
	if k > n || k < 0 {
		return 0
	}

	if k == 0 || k == n {
		return 1
	}

	// Optimize by using the smaller value
	if k > n-k {
		k = n - k
	}

	result := int64(1)
	for i := 0; i < k; i++ {
		result = result * int64(n-i) / int64(i+1)
	}

	return result
}
