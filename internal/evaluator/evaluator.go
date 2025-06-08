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

	"github.com/tmc/pe/internal/llm"
	"github.com/tmc/pe/internal/promptfoo"
	"sigs.k8s.io/yaml"
)

func Evaluate(config promptfoo.Config, timeout time.Duration, dryRun bool, maxConcurrency int, showProgressBar bool) (promptfoo.EvaluationResult, error) {
	if showProgressBar {
		fmt.Printf("Running %d evaluations with up to %d threads...\n\n",
			len(config.Prompts)*len(config.Providers)*len(config.Tests), maxConcurrency)
	}

	if len(config.Prompts) == 0 || len(config.Providers) == 0 || len(config.Tests) == 0 {
		return promptfoo.EvaluationResult{}, fmt.Errorf("missing required config fields")
	}

	evalId := fmt.Sprintf("eval-%s-%s", generateRandomString(3), time.Now().Format("2006-01-02T15:04:05"))

	// Create prompt metadata for each prompt-provider combination
	promptMetadata := make([]promptfoo.PromptData, 0, len(config.Prompts)*len(config.Providers))
	for _, prompt := range config.Prompts {
		for _, provider := range config.Providers {
			promptID := generatePromptID(prompt, provider)
			promptMetadata = append(promptMetadata, promptfoo.PromptData{
				Raw:      prompt,
				Label:    prompt,
				ID:       promptID,
				Provider: provider,
				Metrics: promptfoo.PromptMetrics{
					NamedScores:      make(map[string]float64),
					NamedScoresCount: make(map[string]int),
				},
			})
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	resultsChan := make(chan promptfoo.TestResult, len(config.Prompts)*len(config.Providers)*len(config.Tests))
	errorsChan := make(chan error, len(config.Prompts)*len(config.Providers)*len(config.Tests))
	var wg sync.WaitGroup

	for _, prompt := range config.Prompts {
		for _, providerStr := range config.Providers {
			for k, test := range config.Tests {
				wg.Add(1)
				go func(prompt string, providerStr string, test promptfoo.TestCase, testIdx int) {
					defer wg.Done()
					select {
					case <-ctx.Done():
						errorsChan <- ctx.Err()
						return
					default:
					}

					providerParts := strings.Split(providerStr, ":")
					backend := providerParts[0]
					provider, err := llm.GetProvider(backend)
					if err != nil {
						errorsChan <- fmt.Errorf("provider %s: %v", providerStr, err)
						return
					}

					processedPrompt := replaceVariables(prompt, test.Vars)
					promptID := generatePromptID(prompt, providerStr)

					evalVars := make(map[string]interface{})
					for k, v := range test.Vars {
						evalVars[k] = v
					}
					evalVars["provider"] = providerStr

					startTime := time.Now()
					var response *promptfoo.ProviderResponse
					if !dryRun {
						response, err = provider.EvaluatePrompt(ctx, processedPrompt, evalVars)
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
							Cost:   0.001,
							Cached: false,
						}
					}
					latency := time.Since(startTime)
					if err != nil {
						errorsChan <- fmt.Errorf("provider %s: %v", providerStr, err)
						return
					}

					success, grading := evaluateAssertions(response.Output, test.Assert)
					resultID := generateResultID(prompt, providerStr, test.Vars)

					resultsChan <- promptfoo.TestResult{
						ID:            resultID,
						PromptID:      promptID,
						Prompt:        map[string]string{"raw": processedPrompt, "label": prompt},
						Provider:      map[string]string{"id": providerStr},
						Response:      *response,
						Success:       success,
						Score:         ifThenElse(success, 1.0, 0.0),
						Vars:          test.Vars,
						GradingResult: grading,
						LatencyMs:     latency.Milliseconds(),
					}
				}(prompt, providerStr, test, k)
			}
		}
	}

	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	var detailedResults []promptfoo.TestResult
	var totalTokens, promptTokens, completionTokens, totalNumRequests int32
	passedTests, failedTests, errorTests := 0, 0, 0

	for result := range resultsChan {
		detailedResults = append(detailedResults, result)
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
	}

	if len(errorsChan) > 0 {
		var errs []string
		for i := 0; i < min(5, len(errorsChan)); i++ {
			errs = append(errs, (<-errorsChan).Error())
		}
		errorTests = len(errorsChan)
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

		for _, result := range detailedResults {
			if result.PromptID == promptMetadata[i].ID {
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
			NamedScores:      make(map[string]float64),
			NamedScoresCount: make(map[string]int),
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

func evaluateAssertions(output string, asserts []promptfoo.Assertion) (bool, promptfoo.GradingResult) {
	success := true
	var componentResults []promptfoo.ComponentResult
	totalScore := 0.0
	assertPassCount, assertFailCount := 0, 0

	for _, assert := range asserts {
		pass := checkAssertion(output, assert.Type, assert.Value)
		score := ifThenElse(pass, 1.0, 0.0)
		totalScore += score

		if pass {
			assertPassCount++
		} else {
			assertFailCount++
			success = false
		}

		componentResults = append(componentResults, promptfoo.ComponentResult{
			Pass:      pass,
			Score:     score,
			Reason:    ifThenElseString(pass, "Assertion passed", fmt.Sprintf("Expected output to %s %v", assert.Type, assert.Value)),
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

	return success, promptfoo.GradingResult{
		Pass:             success,
		Score:            totalScore,
		Reason:           ifThenElseString(success, "All assertions passed", "Some assertions failed"),
		ComponentResults: componentResults,
		TokensUsed:       tokenUsage,
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
	hash := sha256.Sum256([]byte(prompt + provider))
	return hex.EncodeToString(hash[:])
}

func generateResultID(prompt, provider string, vars map[string]interface{}) string {
	varStr := fmt.Sprintf("%v", vars)
	hash := sha256.Sum256([]byte(prompt + provider + varStr))
	return hex.EncodeToString(hash[:])[:8]
}

func replaceVariables(prompt string, vars map[string]interface{}) string {
	result := prompt
	for key, value := range vars {
		var strValue string
		switch v := value.(type) {
		case string:
			strValue = v
		case float64:
			strValue = fmt.Sprintf("%g", v)
		case int:
			strValue = fmt.Sprintf("%d", v)
		case bool:
			strValue = fmt.Sprintf("%t", v)
		default:
			jsonValue, err := json.Marshal(v)
			if err == nil {
				strValue = string(jsonValue)
			} else {
				strValue = fmt.Sprintf("%v", v)
			}
		}
		result = strings.ReplaceAll(result, "{{"+key+"}}", strValue)
	}
	return result
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
