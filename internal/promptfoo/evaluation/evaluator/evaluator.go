package evaluator

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/tmc/pe/internal/distributed"
	"github.com/tmc/pe/internal/llm"
	"github.com/tmc/pe/internal/promptfoo"
	"github.com/tmc/pe/internal/providers"
	"sigs.k8s.io/yaml"
)

type evalOutcome struct {
	result promptfoo.TestResult
	err    error
}

type responseCache struct {
	mu      sync.Mutex
	entries map[string]*promptfoo.ProviderResponse
}

func newResponseCache() *responseCache {
	return &responseCache{entries: make(map[string]*promptfoo.ProviderResponse)}
}

func (c *responseCache) get(key string) (*promptfoo.ProviderResponse, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	response, ok := c.entries[key]
	if !ok {
		return nil, false
	}
	return cloneProviderResponse(response), true
}

func (c *responseCache) put(key string, response *promptfoo.ProviderResponse) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries[key] = cloneProviderResponse(response)
}

func Evaluate(config promptfoo.Config, timeout time.Duration, dryRun bool, maxConcurrency int, showProgressBar bool) (promptfoo.EvaluationResult, error) {
	// Expand promptfoo scenarios into concrete tests so the rest of the runner
	// is scenario-agnostic. A config without scenarios is unchanged.
	if len(config.Scenarios) > 0 {
		config.Tests = config.ExpandScenarios()
	}
	// Merge defaultTest into every test (including scenario-expanded ones), so
	// baseline vars/asserts apply uniformly.
	if config.DefaultTest != nil {
		config.Tests = config.ApplyDefaults(config.Tests)
	}

	if showProgressBar {
		fmt.Printf("Running %d evaluations with up to %d threads...\n\n",
			len(config.Prompts)*len(config.Providers)*len(config.Tests), maxConcurrency)
	}

	if len(config.Prompts) == 0 || len(config.Providers) == 0 || len(config.Tests) == 0 {
		return promptfoo.EvaluationResult{}, fmt.Errorf("missing required config fields")
	}
	if maxConcurrency < 1 {
		maxConcurrency = 1
	}

	materializedProviders, err := providers.MaterializeProviders(config.Providers)
	if err != nil {
		return promptfoo.EvaluationResult{}, err
	}

	evalId := fmt.Sprintf("eval-%s-%s", generateRandomString(3), time.Now().Format("2006-01-02T15:04:05"))

	// Create prompt metadata for each prompt-provider combination
	promptMetadata := make([]promptfoo.PromptData, 0, len(config.Prompts)*len(materializedProviders))
	for _, prompt := range config.Prompts {
		for _, provider := range materializedProviders {
			if !provider.AppliesToPrompt(prompt) {
				continue
			}
			promptID := generatePromptID(prompt, provider.Spec.ID)
			promptMetadata = append(promptMetadata, promptfoo.PromptData{
				Raw:      prompt,
				Label:    prompt,
				ID:       promptID,
				Provider: provider.DisplayName(),
				Metrics: promptfoo.PromptMetrics{
					NamedScores:      make(map[string]float64),
					NamedScoresCount: make(map[string]int),
				},
			})
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cache := newResponseCache()
	var tasks []distributed.Task[evalOutcome]
	for _, prompt := range config.Prompts {
		for _, provider := range materializedProviders {
			if !provider.AppliesToPrompt(prompt) {
				continue
			}
			for k, test := range config.Tests {
				prompt := prompt
				provider := provider
				test := test
				tasks = append(tasks, distributed.Task[evalOutcome]{
					ID: fmt.Sprintf("%s/%s/%d", prompt, provider.Spec.ID, k),
					Run: func(ctx context.Context) (evalOutcome, error) {
						outcome, err := evaluateOne(ctx, prompt, provider, test, dryRun, cache)
						if err != nil && ctx.Err() != nil {
							return evalOutcome{}, err
						}
						outcome.err = err
						return outcome, nil
					},
				})
			}
		}
	}

	executionResults, executionErr := distributed.RunLocal(ctx, maxConcurrency, tasks)

	var detailedResults []promptfoo.TestResult
	var evalErrors []error
	var totalTokens, promptTokens, completionTokens, totalNumRequests int32
	passedTests, failedTests, errorTests := 0, 0, 0

	for _, executionResult := range executionResults {
		if executionResult.Err != nil {
			evalErrors = append(evalErrors, executionResult.Err)
			continue
		}
		if executionResult.ID == "" {
			continue
		}
		if executionResult.Value.err != nil {
			evalErrors = append(evalErrors, executionResult.Value.err)
			continue
		}
		result := executionResult.Value.result
		if result.Success {
			passedTests++
		} else {
			failedTests++
		}
		if result.Response.TokenUsage != nil {
			totalTokens += result.Response.TokenUsage.Total
			promptTokens += result.Response.TokenUsage.Prompt
			completionTokens += result.Response.TokenUsage.Completion
			if result.Response.TokenUsage.NumRequests > 0 {
				totalNumRequests += result.Response.TokenUsage.NumRequests
			} else {
				totalNumRequests++
			}
		}
		detailedResults = append(detailedResults, result)
	}
	if executionErr != nil {
		evalErrors = append(evalErrors, executionErr)
	}

	if len(evalErrors) > 0 {
		var errs []string
		for i := 0; i < min(5, len(evalErrors)); i++ {
			errs = append(errs, evalErrors[i].Error())
		}
		errorTests = len(evalErrors)
		if len(detailedResults) == 0 {
			return promptfoo.EvaluationResult{}, fmt.Errorf("evaluation errors: %s", strings.Join(errs, "; "))
		}
		fmt.Printf("Warning: some evaluations failed: %s\n", strings.Join(errs, "; "))
	}

	// Update prompt metrics
	for i := range promptMetadata {
		successes, failures, errors := 0, 0, 0
		totalLatency := int64(0)
		var promptTotal, promptPrompt, promptCompletion, numRequests int32
		totalCost := 0.0
		assertPassCount, assertFailCount := 0, 0
		namedScores := make(map[string]float64)
		namedScoresCount := make(map[string]int)
		resultCount := 0

		for _, result := range detailedResults {
			if result.PromptID == promptMetadata[i].ID {
				resultCount++
				if result.Success {
					successes++
				} else {
					failures++
				}
				totalLatency += result.LatencyMs
				if result.Response.TokenUsage != nil {
					promptTotal += result.Response.TokenUsage.Total
					promptPrompt += result.Response.TokenUsage.Prompt
					promptCompletion += result.Response.TokenUsage.Completion
					numRequests++
				}
				if result.Response.Cost > 0 {
					totalCost += result.Response.Cost
				}

				// Count assertions
				for _, componentResult := range result.GradingResult.ComponentResults {
					if componentResult.Pass {
						assertPassCount++
					} else {
						assertFailCount++
					}
				}

				// Aggregate named scores from labeled assertions.
				for name, score := range result.GradingResult.NamedScores {
					namedScores[name] += score
					namedScoresCount[name]++
				}
			}
		}

		// Compute derivedMetrics over the aggregated named scores for this
		// prompt. A bad expression is skipped, not fatal.
		if len(config.DerivedMetrics) > 0 {
			derived, derr := promptfoo.ComputeDerivedMetrics(config.DerivedMetrics, namedScores, resultCount)
			if derr != nil {
				fmt.Printf("Warning: %v\n", derr)
			}
			for name, v := range derived {
				namedScores[name] = v
				namedScoresCount[name]++
			}
		}

		// Create completion details (typically would be populated by the provider)
		completionDetails := &promptfoo.CompletionDetails{
			Reasoning:          0,
			AcceptedPrediction: 0,
			RejectedPrediction: 0,
		}

		promptMetadata[i].Metrics = promptfoo.PromptMetrics{
			Score:           float64(successes),
			TestPassCount:   successes,
			TestFailCount:   failures,
			TestErrorCount:  errors,
			AssertPassCount: assertPassCount,
			AssertFailCount: assertFailCount,
			TotalLatencyMs:  totalLatency,
			TokenUsage: promptfoo.TokenUsage{
				Total:       promptTotal,
				Prompt:      promptPrompt,
				Completion:  promptCompletion,
				Cached:      0,
				NumRequests: numRequests,
				Details:     completionDetails,
			},
			NamedScores:      namedScores,
			NamedScoresCount: namedScoresCount,
			Cost:             totalCost,
		}
	}

	// Create completion details for the global stats
	completionDetails := &promptfoo.CompletionDetails{
		Reasoning:          0,
		AcceptedPrediction: 0,
		RejectedPrediction: 0,
	}

	return promptfoo.EvaluationResult{
		EvalID: evalId,
		Results: promptfoo.ResultSet{
			Version:   3,
			Timestamp: time.Now().Format(time.RFC3339),
			Prompts:   promptMetadata,
			Results:   detailedResults,
			Stats: promptfoo.Stats{
				Successes: passedTests,
				Failures:  failedTests,
				Errors:    errorTests,
				TokenUsage: promptfoo.TokenUsage{
					Total:       totalTokens,
					Prompt:      promptTokens,
					Completion:  completionTokens,
					Cached:      0,
					NumRequests: totalNumRequests,
					Details:     completionDetails,
				},
			},
		},
		Config: config,
	}, nil
}

func evaluateOne(ctx context.Context, prompt string, provider *providers.MaterializedProvider, test promptfoo.TestCase, dryRun bool, cache *responseCache) (evalOutcome, error) {
	if err := ctx.Err(); err != nil {
		return evalOutcome{}, err
	}

	if provider.Delay > 0 && !dryRun {
		timer := time.NewTimer(provider.Delay)
		defer timer.Stop()
		select {
		case <-ctx.Done():
			return evalOutcome{}, ctx.Err()
		case <-timer.C:
		}
	}

	processedPrompt := promptfoo.ApplyVars(prompt, test.Vars)
	promptID := generatePromptID(prompt, provider.Spec.ID)

	evalVars := make(map[string]interface{})
	for k, v := range provider.Spec.Config {
		evalVars[k] = v
	}
	for k, v := range test.Vars {
		evalVars[k] = v
	}
	evalVars["provider"] = provider.Spec.ID

	startTime := time.Now()
	var response *promptfoo.ProviderResponse
	var err error
	if !dryRun {
		cacheKey := evaluationCacheKey(provider.Spec, processedPrompt, evalVars)
		if cache != nil {
			if cached, ok := cache.get(cacheKey); ok {
				cached.Cached = true
				if cached.TokenUsage != nil {
					cached.TokenUsage.Cached = cached.TokenUsage.Total
				}
				response = cached
			}
		}
		if response == nil {
			response, err = provider.Executor.EvaluatePrompt(ctx, processedPrompt, evalVars)
			if err == nil && cache != nil {
				cache.put(cacheKey, response)
			}
		}
	} else {
		response = &promptfoo.ProviderResponse{
			Output: "Dry run response",
			TokenUsage: &promptfoo.TokenUsage{
				Total:       10,
				Prompt:      5,
				Completion:  5,
				NumRequests: 1,
				Details: &promptfoo.CompletionDetails{
					Reasoning:          0,
					AcceptedPrediction: 0,
					RejectedPrediction: 0,
				},
			},
			Cost:      0.001,
			Cached:    false,
			LatencyMs: 1,
		}
	}
	latency := time.Since(startTime)
	if err != nil {
		return evalOutcome{}, fmt.Errorf("provider %s: %v", provider.Spec.ID, err)
	}

	var judgeProvider llm.Provider
	if !dryRun {
		judgeProvider = provider.Executor
	}
	// Make the rendered question and any context var available to model-graded
	// assertions (answer-relevance, context-*).
	assertMeta := map[string]interface{}{"input": processedPrompt}
	if c, ok := test.Vars["context"]; ok {
		assertMeta["context"] = fmt.Sprintf("%v", c)
	}
	success, grading := evaluateAssertionsWithMeta(ctx, response.Output, test.Assert, judgeProvider, assertMeta)
	resultID := generateResultID(prompt, provider.Spec.ID, test.Vars)
	latencyMs := latency.Milliseconds()
	if response.LatencyMs > 0 {
		latencyMs = response.LatencyMs
	}

	return evalOutcome{
		result: promptfoo.TestResult{
			ID:            resultID,
			PromptID:      promptID,
			Prompt:        map[string]string{"raw": processedPrompt, "label": prompt},
			Provider:      map[string]string{"id": provider.Spec.ID, "label": provider.DisplayName()},
			Response:      *response,
			Success:       success,
			Score:         ifThenElse(success, 1.0, 0.0),
			Vars:          test.Vars,
			GradingResult: grading,
			LatencyMs:     latencyMs,
		},
	}, nil
}

func evaluationCacheKey(provider promptfoo.ProviderConfig, prompt string, vars map[string]interface{}) string {
	data, _ := json.Marshal(struct {
		Provider promptfoo.ProviderConfig `json:"provider"`
		Prompt   string                   `json:"prompt"`
		Vars     map[string]interface{}   `json:"vars"`
	}{
		Provider: provider,
		Prompt:   prompt,
		Vars:     vars,
	})
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func cloneProviderResponse(response *promptfoo.ProviderResponse) *promptfoo.ProviderResponse {
	if response == nil {
		return nil
	}
	clone := *response
	if response.TokenUsage != nil {
		tokenUsage := *response.TokenUsage
		if response.TokenUsage.Details != nil {
			details := *response.TokenUsage.Details
			tokenUsage.Details = &details
		}
		clone.TokenUsage = &tokenUsage
	}
	if response.Metadata != nil {
		clone.Metadata = make(map[string]interface{}, len(response.Metadata))
		for k, v := range response.Metadata {
			clone.Metadata[k] = v
		}
	}
	return &clone
}

func evaluateAssertions(output string, asserts []promptfoo.Assertion) (bool, promptfoo.GradingResult) {
	return evaluateAssertionsWithProvider(context.Background(), output, asserts, nil)
}

func evaluateAssertionsWithProvider(ctx context.Context, output string, asserts []promptfoo.Assertion, judgeProvider llm.Provider) (bool, promptfoo.GradingResult) {
	return evaluateAssertionsWithMeta(ctx, output, asserts, judgeProvider, nil)
}

func evaluateAssertionsWithMeta(ctx context.Context, output string, asserts []promptfoo.Assertion, judgeProvider llm.Provider, meta map[string]interface{}) (bool, promptfoo.GradingResult) {
	success := true
	var componentResults []promptfoo.ComponentResult
	totalScore := 0.0
	assertPassCount, assertFailCount := 0, 0

	namedScores := make(map[string]float64)
	for _, assert := range asserts {
		pass, score, reason := evaluateAssertion(ctx, output, assert, judgeProvider, meta)
		totalScore += score

		if pass {
			assertPassCount++
		} else {
			assertFailCount++
			success = false
		}

		// Label this assertion's score as a named score when assert.metric is
		// set, so it can feed derivedMetrics. Scores for the same metric sum,
		// matching promptfoo's aggregation.
		if assert.Metric != "" {
			namedScores[assert.Metric] += score
		}

		componentResults = append(componentResults, promptfoo.ComponentResult{
			Pass:      pass,
			Score:     score,
			Reason:    reason,
			Assertion: assert,
		})
	}

	// Create empty token usage with appropriate structure
	tokenUsage := promptfoo.TokenUsage{
		Total:       0,
		Prompt:      0,
		Completion:  0,
		Cached:      0,
		NumRequests: 1,
		Details: &promptfoo.CompletionDetails{
			Reasoning:          0,
			AcceptedPrediction: 0,
			RejectedPrediction: 0,
		},
	}

	if len(namedScores) == 0 {
		namedScores = nil
	}
	return success, promptfoo.GradingResult{
		Pass:             success,
		Score:            totalScore,
		Reason:           ifThenElseString(success, "All assertions passed", "Some assertions failed"),
		ComponentResults: componentResults,
		TokensUsed:       tokenUsage,
		NamedScores:      namedScores,
	}
}

func evaluateAssertion(ctx context.Context, output string, assert promptfoo.Assertion, judgeProvider llm.Provider, meta map[string]interface{}) (bool, float64, string) {
	// Resolve promptfoo/pe id divergences (regex->matches, llm-rubric->llm-judge,
	// similar->similarity, not-* inversion) before dispatch so unmodified
	// promptfoo configs run against pe's evaluator.
	canonical, negate, ok := normalizeAssertionType(assert.Type)
	if ok {
		canonAssert := promptfooAssertion(assert)
		canonAssert.Type = canonical
		evaluator := NewAssertionEvaluator(judgeProvider)
		result, err := evaluator.EvaluateAssertion(ctx, canonAssert, output, meta)
		if err != nil {
			// Model-graded assertions without a judge provider are a hard
			// error; other evaluators degrade to the string-match fallback.
			if isModelGraded(canonical) {
				return false, 0, err.Error()
			}
		} else {
			pass, score := result.Passed, result.Score
			if negate {
				pass = !pass
				score = 1 - score
			}
			return pass, score, result.Message
		}
	}

	pass := checkAssertion(output, assert.Type, assert.Value)
	score := ifThenElse(pass, 1.0, 0.0)
	reason := ifThenElseString(pass, "Assertion passed", fmt.Sprintf("Expected output to %s %v", assert.Type, assert.Value))
	return pass, score, reason
}

func promptfooAssertion(assert promptfoo.Assertion) Assertion {
	var threshold *float64
	if assert.Threshold != 0 {
		threshold = &assert.Threshold
	}
	canonical, _, ok := normalizeAssertionType(assert.Type)
	if !ok {
		canonical = AssertionType(assert.Type)
	}
	return Assertion{
		Type:      canonical,
		Value:     assert.Value,
		Provider:  assert.Provider,
		Threshold: threshold,
		Config:    assert.Config,
	}
}

func checkAssertion(output string, assertType string, assertValue interface{}) bool {
	strValue := fmt.Sprintf("%v", assertValue)
	switch assertType {
	case "equals", "==":
		return output == strValue
	case "contains":
		return strings.Contains(output, strValue)
	case "not-contains", "!contains":
		return !strings.Contains(output, strValue)
	case "icontains":
		return strings.Contains(strings.ToLower(output), strings.ToLower(strValue))
	case "starts-with":
		return strings.HasPrefix(output, strValue)
	case "ends-with":
		return strings.HasSuffix(output, strValue)
	case "regex", "matches":
		matched, err := regexp.MatchString(strValue, output)
		return err == nil && matched
	default:
		return false
	}
}

func generatePromptID(prompt, provider string) string {
	return hashID(prompt, provider)
}

func generateResultID(prompt, provider string, vars map[string]interface{}) string {
	varJSON, err := json.Marshal(vars)
	if err != nil {
		varJSON = []byte(fmt.Sprintf("%#v", vars))
	}
	return hashID(prompt, provider, string(varJSON))
}

func hashID(parts ...string) string {
	h := sha256.New()
	for _, part := range parts {
		h.Write([]byte(part))
		h.Write([]byte{0})
	}
	return hex.EncodeToString(h.Sum(nil))
}

func ifThenElse(condition bool, trueVal, falseVal interface{}) float64 {
	if condition {
		return 1.0
	}
	return 0.0
}

// ifThenElseString returns one of two string values based on a condition
func ifThenElseString(condition bool, trueVal, falseVal string) string {
	if condition {
		return trueVal
	}
	return falseVal
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func FormatResults(results promptfoo.EvaluationResult, format string) ([]byte, error) {
	switch format {
	case "json":
		return json.MarshalIndent(results, "", "  ")
	case "yaml":
		return yaml.Marshal(results)
	case "csv":
		return formatResultsAsCSV(results), nil
	case "junit", "junit.xml", "xml":
		return formatResultsAsJUnit(results)
	case "table", "":
		return formatResultsAsTable(results), nil
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
}

// Helper function to generate a random string
func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[time.Now().UnixNano()%int64(len(charset))]
		time.Sleep(1 * time.Nanosecond) // Ensure different values
	}
	return string(result)
}

// Format results as CSV
func formatResultsAsCSV(results promptfoo.EvaluationResult) []byte {
	// Create a buffer for CSV output
	var buffer bytes.Buffer

	// Add CSV header
	buffer.WriteString("Test,Provider,Prompt,Success,Output\n")

	// Process each result row
	for _, result := range results.Results.Results {
		// Extract fields
		promptText := result.Prompt["raw"]
		providerText := result.Provider["id"]
		outputText := result.Response.Output
		success := result.Success

		// Format status
		status := "PASS"
		if !success {
			status = "FAIL"
		}

		// Escape fields for CSV
		promptText = escapeCSV(promptText)
		providerText = escapeCSV(providerText)
		outputText = escapeCSV(outputText)

		// Add the row
		buffer.WriteString(fmt.Sprintf("%s,%s,%s,%s,%s\n",
			result.ID, providerText, promptText, status, outputText))
	}

	return buffer.Bytes()
}

// Helper function to escape characters for CSV
func escapeCSV(s string) string {
	if strings.Contains(s, ",") || strings.Contains(s, "\"") || strings.Contains(s, "\n") {
		s = strings.ReplaceAll(s, "\"", "\"\"")
		s = "\"" + s + "\""
	}
	return s
}

// Format results as a table
func formatResultsAsTable(results promptfoo.EvaluationResult) []byte {
	var buffer bytes.Buffer

	// Add basic information
	buffer.WriteString(fmt.Sprintf("Evaluation: %s\n", results.EvalID))
	buffer.WriteString(fmt.Sprintf("Timestamp: %s\n\n", results.Results.Timestamp))

	// Add statistics
	stats := results.Results.Stats
	buffer.WriteString(fmt.Sprintf("Success: %d, Failures: %d, Total: %d\n",
		stats.Successes, stats.Failures, stats.Successes+stats.Failures))
	buffer.WriteString(fmt.Sprintf("Token Usage: %d (Prompt: %d, Completion: %d)\n\n",
		stats.TokenUsage.Total, stats.TokenUsage.Prompt, stats.TokenUsage.Completion))

	// Create header for results table
	buffer.WriteString("ID\tPrompt\tProvider\tSuccess\tScore\n")
	buffer.WriteString("--\t------\t--------\t-------\t-----\n")

	// Add each result
	for _, result := range results.Results.Results {
		promptLabel := result.Prompt["label"]
		providerID := result.Provider["id"]
		success := "✓"
		if !result.Success {
			success = "✗"
		}

		buffer.WriteString(fmt.Sprintf("%s\t%s\t%s\t%s\t%.2f\n",
			result.ID, promptLabel, providerID, success, result.Score))
	}

	return buffer.Bytes()
}
