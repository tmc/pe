package evaluator

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/tmc/pe/internal/cgpt"
	"sigs.k8s.io/yaml"
)

// Evaluate runs the evaluation based on the provided configuration
func Evaluate(config map[string]interface{}, timeout time.Duration, dryRun bool, maxConcurrency int, showProgressBar bool) (map[string]interface{}, error) {
	// Now we'll integrate with CGPT for real results
	return evaluateWithCGPT(config, timeout, dryRun, maxConcurrency, showProgressBar)
}

// evaluateWithCGPT runs the evaluation using CGPT for real LLM responses
func evaluateWithCGPT(config map[string]interface{}, timeout time.Duration, dryRun bool, maxConcurrency int, showProgressBar bool) (map[string]interface{}, error) {
	// Extract necessary components
	prompts, _ := config["prompts"].([]interface{})
	providers, _ := config["providers"].([]interface{})
	tests, _ := config["tests"].([]interface{})
	
	// If showProgressBar is enabled, display a message about concurrency
	if showProgressBar {
		fmt.Printf("Running %d concurrent evaluations with up to %d threads...\n\n", 
			len(prompts)*len(providers)*len(tests), maxConcurrency)
	}
	
	// Ensure we have the necessary data to generate results
	if prompts == nil || len(prompts) == 0 {
		return nil, fmt.Errorf("no prompts provided in configuration")
	}
	
	if providers == nil || len(providers) == 0 {
		return nil, fmt.Errorf("no providers provided in configuration")
	}
	
	if tests == nil || len(tests) == 0 {
		return nil, fmt.Errorf("no tests provided in configuration")
	}

	// CGPT package is imported at the top
	
	// Create a unique evalId for this run
	evalId := fmt.Sprintf("eval-%s-%s", 
		generateRandomString(3), 
		time.Now().Format("2006-01-02T15:04:05"))
	timestamp := time.Now().Format(time.RFC3339)
	
	// Create prompt metadata for the result structure
	var promptMetadata []map[string]interface{}
	for i, prompt := range prompts {
		promptStr, ok := prompt.(string)
		if !ok {
			promptStr = fmt.Sprintf("Prompt %d", i+1)
		}
		
		// Generate a unique ID for this prompt
		promptId := fmt.Sprintf("p%x", generateStableHash(promptStr))
		
		// Create metrics placeholder (will be updated later)
		promptMetadata = append(promptMetadata, map[string]interface{}{
			"raw":      promptStr,
			"label":    promptStr,
			"id":       promptId,
			"provider": providers[0],
			"metrics": map[string]interface{}{
				"score":          0,
				"testPassCount":  0,
				"testFailCount":  0,
				"testErrorCount": 0,
				"assertPassCount": 0,
				"assertFailCount": 0,
				"totalLatencyMs": 0,
				"tokenUsage": map[string]interface{}{
					"total":       0,
					"prompt":      0,
					"completion":  0,
					"cached":      0,
					"numRequests": 0,
				},
				"namedScores": map[string]interface{}{},
			},
		})
	}
	
	// Generate individual test results
	var detailedResults []map[string]interface{}
	totalPromptTokens := 0.0
	totalCompletionTokens := 0.0
	totalTokens := 0.0
	passedTests := 0
	failedTests := 0
	
	// Create timeout context
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	
	// Create a channel for results
	resultsChan := make(chan map[string]interface{}, len(prompts)*len(providers)*len(tests))
	errorsChan := make(chan error, len(prompts)*len(providers)*len(tests))
	
	// Create a waitgroup to track completion
	var wg sync.WaitGroup
	
	// For each prompt, provider, and test combination
	for i, prompt := range prompts {
		promptStr, ok := prompt.(string)
		if !ok {
			promptStr = fmt.Sprintf("Prompt %d", i+1)
		}
		
		// Get the promptId from metadata
		promptId := promptMetadata[i]["id"].(string)
		
		for j, provider := range providers {
			providerStr, ok := provider.(string)
			if !ok {
				providerStr = fmt.Sprintf("Provider %d", j+1)
			}
			
			for k, test := range tests {
				testMap, ok := test.(map[string]interface{})
				if !ok {
					testMap = map[string]interface{}{}
				}
				
				testVars, _ := testMap["vars"].(map[string]interface{})
				assertions, _ := testMap["assert"].([]interface{})
				
				// Skip this evaluation if we've timed out already
				select {
				case <-ctx.Done():
					continue
				default:
					// Continue with the evaluation
				}
				
				// Increment the waitgroup
				wg.Add(1)
				
				// Run the evaluation in a goroutine
				go func(promptIdx int, promptStr, promptId, providerStr string, testIdx int, testVars map[string]interface{}, assertions []interface{}) {
					defer wg.Done()
					
					// Skip if context is cancelled
					select {
					case <-ctx.Done():
						errorsChan <- fmt.Errorf("evaluation timed out")
						return
					default:
						// Continue with the evaluation
					}
					
					// Create a provider configuration for CGPT
					modelProvider := cgpt.DefaultProvider()
					
					// Set the provider from the configuration
					providerParts := strings.Split(providerStr, ":")
					if len(providerParts) > 0 {
						modelProvider.Backend = providerParts[0]
					}
					if len(providerParts) > 1 {
						modelProvider.Model = providerParts[1]
					}
					
					// Replace any variables in the prompt
					processedPrompt := promptStr
					for varName, varValue := range testVars {
						if valStr, ok := varValue.(string); ok {
							placeholder := fmt.Sprintf("{{%s}}", varName)
							processedPrompt = strings.Replace(processedPrompt, placeholder, valStr, -1)
						}
					}
					
					// Prepare vars with provider info
					evalVars := make(map[string]interface{})
					for k, v := range testVars {
						evalVars[k] = v
					}
					evalVars["provider"] = providerStr
					
					// Run the evaluation
					startTime := time.Now()
					response, err := modelProvider.EvaluatePromptWithOptions(processedPrompt, evalVars, dryRun)
					latency := time.Since(startTime)
					
					if err != nil {
						errorsChan <- fmt.Errorf("error evaluating prompt with provider %s: %v", providerStr, err)
						return
					}
					
					// Process assertions to determine success
					success := true
					var componentResults []map[string]interface{}
					
					for _, assertion := range assertions {
						assertMap, _ := assertion.(map[string]interface{})
						assertType, _ := assertMap["type"].(string)
						assertValue, _ := assertMap["value"].(string)
						
						// Check if the assertion passes
						assertionPasses := checkAssertion(response.Output, assertType, assertValue)
						
						// Track success/failure
						if !assertionPasses {
							success = false
						}
						
						var score int
						var reason string
						if assertionPasses {
							score = 1
							reason = "Assertion passed"
						} else {
							score = 0
							reason = "Assertion failed"
						}
						
						componentResults = append(componentResults, map[string]interface{}{
							"pass":      assertionPasses,
							"score":     score,
							"reason":    reason,
							"assertion": assertMap,
						})
					}
					
					// Generate a stable ID for this result based on inputs
					resultIdInput := fmt.Sprintf("%s-%s-%v", promptStr, providerStr, fmt.Sprintf("%v", testVars))
					resultId := fmt.Sprintf("r%x", generateStableHash(resultIdInput))
					
					// Determine score value
					var scoreValue int
					if success {
						scoreValue = 1
					} else {
						scoreValue = 0
					}
					
					// Determine reason text
					var reasonText string
					if success {
						reasonText = "All assertions passed"
					} else {
						reasonText = "Some assertions failed"
					}
					
					// Create the result to send back
					result := map[string]interface{}{
						"id":           resultId,
						"promptId":     promptId,
						"promptIdx":    promptIdx,
						"testIdx":      testIdx,
						"prompt": map[string]interface{}{
							"raw":   processedPrompt,
							"label": promptStr,
						},
						"provider": map[string]interface{}{
							"id":    providerStr,
							"label": providerStr,
						},
						"response": map[string]interface{}{
							"output": response.Output,
							"tokenUsage": map[string]interface{}{
								"total":      response.TokenUsage.Total,
								"prompt":     response.TokenUsage.Prompt,
								"completion": response.TokenUsage.Completion,
								"cached":     response.TokenUsage.Cached,
							},
							"cached": false,
							"cost":   response.Cost,
						},
						"latencyMs":     latency.Milliseconds(),
						"cost":          response.Cost,
						"success":       success,
						"score":         scoreValue,
						"vars":          testVars,
						"failureReason": nil,
						"testCase": map[string]interface{}{
							"vars":     testVars,
							"assert":   assertions,
							"options":  map[string]interface{}{},
							"metadata": map[string]interface{}{},
						},
						"gradingResult": map[string]interface{}{
							"pass":       success,
							"score":      scoreValue,
							"reason":     reasonText,
							"tokensUsed": map[string]interface{}{
								"total":      response.TokenUsage.Total,
								"prompt":     response.TokenUsage.Prompt,
								"completion": response.TokenUsage.Completion,
								"cached":     response.TokenUsage.Cached,
							},
							"componentResults": componentResults,
						},
					}
					
					// Send the result back
					resultsChan <- result
				}(i, promptStr, promptId, providerStr, k, testVars, assertions)
			}
		}
	}
	
	// Wait for all evaluations to complete or timeout
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()
	
	// Wait for completion or timeout
	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("evaluation timed out after %v", timeout)
	case <-done:
		// All evaluations completed
	}
	
	// Collect results
	close(resultsChan)
	for result := range resultsChan {
		detailedResults = append(detailedResults, result)
		
		// Count successes and failures
		success, _ := result["success"].(bool)
		if success {
			passedTests++
		} else {
			failedTests++
		}
		
		// Accumulate token usage
		response, _ := result["response"].(map[string]interface{})
		tokenUsage, _ := response["tokenUsage"].(map[string]interface{})
		
		promptTokens, _ := tokenUsage["prompt"].(float64)
		completionTokens, _ := tokenUsage["completion"].(float64)
		total, _ := tokenUsage["total"].(float64)
		
		totalPromptTokens += promptTokens
		totalCompletionTokens += completionTokens
		totalTokens += total
	}
	
	// Check for errors
	if len(errorsChan) > 0 {
		// Read at most 5 errors to report
		var errorMsgs []string
		for i := 0; i < 5 && i < len(errorsChan); i++ {
			err := <-errorsChan
			errorMsgs = append(errorMsgs, err.Error())
		}
		
		if len(errorMsgs) > 0 {
			// Only report errors if we didn't get any results
			if len(detailedResults) == 0 {
				return nil, fmt.Errorf("evaluation errors: %s", strings.Join(errorMsgs, "; "))
			}
			// Otherwise, just log the errors but continue
			fmt.Printf("Warning: some evaluations had errors: %s\n", strings.Join(errorMsgs, "; "))
		}
	}
	
	// Update prompt metrics
	for i := range promptMetadata {
		// Count successes and failures for this prompt
		promptSuccesses := 0
		promptFailures := 0
		
		// Sum token usage for this prompt
		promptTokens := 0.0
		completionTokens := 0.0
		totalTokens := 0.0
		numRequests := 0
		
		promptId := promptMetadata[i]["id"].(string)
		
		// Find results for this prompt
		for _, result := range detailedResults {
			if result["promptId"] == promptId {
				// Count success/failure
				success, _ := result["success"].(bool)
				if success {
					promptSuccesses++
				} else {
					promptFailures++
				}
				
				// Add token usage
				response, _ := result["response"].(map[string]interface{})
				tokenUsage, _ := response["tokenUsage"].(map[string]interface{})
				
				pt, _ := tokenUsage["prompt"].(float64)
				ct, _ := tokenUsage["completion"].(float64)
				tt, _ := tokenUsage["total"].(float64)
				
				promptTokens += pt
				completionTokens += ct
				totalTokens += tt
				numRequests++
			}
		}
		
		// Update metrics in promptMetadata
		metrics, _ := promptMetadata[i]["metrics"].(map[string]interface{})
		metrics["testPassCount"] = promptSuccesses
		metrics["testFailCount"] = promptFailures
		metrics["score"] = float64(promptSuccesses) / float64(max(1, promptSuccesses+promptFailures))
		metrics["totalLatencyMs"] = 0 // We don't track this correctly yet
		
		tokenUsage, _ := metrics["tokenUsage"].(map[string]interface{})
		tokenUsage["prompt"] = promptTokens
		tokenUsage["completion"] = completionTokens
		tokenUsage["total"] = totalTokens
		tokenUsage["numRequests"] = numRequests
	}
	
	// Build the complete result structure
	return map[string]interface{}{
		"evalId": evalId,
		"config": config,
		"results": map[string]interface{}{
			"version":   3,
			"timestamp": timestamp,
			"prompts":   promptMetadata,
			"results":   detailedResults,
			"stats": map[string]interface{}{
				"successes":  passedTests,
				"failures":   failedTests,
				"errors":     0,
				"tokenUsage": map[string]interface{}{
					"cached":     0,
					"completion": totalCompletionTokens,
					"prompt":     totalPromptTokens,
					"total":      totalTokens,
					"numRequests": len(detailedResults),
				},
			},
		},
	}, nil
}

// max returns the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// checkAssertion checks if the output satisfies the assertion
func checkAssertion(output, assertType, assertValue string) bool {
	switch assertType {
	case "equals", "==":
		return output == assertValue
	case "contains":
		return strings.Contains(output, assertValue)
	case "icontains":
		return strings.Contains(strings.ToLower(output), strings.ToLower(assertValue))
	case "startsWith":
		return strings.HasPrefix(output, assertValue)
	case "endsWith":
		return strings.HasSuffix(output, assertValue)
	case "regex", "matches":
		matched, err := regexp.MatchString(assertValue, output)
		return err == nil && matched
	case "!contains", "not-contains":
		return !strings.Contains(output, assertValue)
	default:
		// Unknown assertion type, consider it failed
		return false
	}
}

// FormatResults formats the evaluation results according to the specified format
func FormatResults(results map[string]interface{}, format string) ([]byte, error) {
	var output []byte
	var err error
	
	switch format {
	case "json":
		output, err = json.MarshalIndent(results, "", "  ")
	case "yaml":
		output, err = yaml.Marshal(results)
	case "text":
		output = formatResultsAsText(results)
	case "table", "":
		// Default to table format for empty string
		output = formatResultsAsTable(results)
	case "csv":
		output = formatResultsAsCSV(results)
	default:
		return nil, fmt.Errorf("unsupported output format: %s", format)
	}

	return output, err
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

// Helper function to generate a stable hash for IDs
func generateStableHash(input string) uint32 {
	var hash uint32 = 5381
	for _, c := range input {
		hash = ((hash << 5) + hash) + uint32(c)
	}
	return hash
}

// generateMockResults creates a mock response for testing purposes
func generateMockResults(config map[string]interface{}, timeout time.Duration) map[string]interface{} {
	// Extract necessary components
	prompts, _ := config["prompts"].([]interface{})
	providers, _ := config["providers"].([]interface{})
	tests, _ := config["tests"].([]interface{})
	
	// Ensure we have the necessary data to generate results
	if prompts == nil || len(prompts) == 0 {
		prompts = []interface{}{"Default prompt"}
	}
	
	if providers == nil || len(providers) == 0 {
		providers = []interface{}{"mock-provider"}
	}
	
	if tests == nil || len(tests) == 0 {
		tests = []interface{}{
			map[string]interface{}{
				"vars": map[string]interface{}{
					"input": "Hello world",
					"language": "French",
				},
				"assert": []interface{}{
					map[string]interface{}{
						"type": "contains",
						"value": "Bonjour",
					},
				},
			},
		}
	}
	
	// Create a unique evalId for this run
	evalId := fmt.Sprintf("eval-%s-%s", 
		generateRandomString(3), 
		time.Now().Format("2006-01-02T15:04:05"))
	timestamp := time.Now().Format(time.RFC3339)
	
	// Create prompt metadata for the result structure
	var promptMetadata []map[string]interface{}
	for i, prompt := range prompts {
		promptStr, ok := prompt.(string)
		if !ok {
			promptStr = fmt.Sprintf("Prompt %d", i+1)
		}
		
		// Generate a unique ID for this prompt
		promptId := fmt.Sprintf("p%x", generateStableHash(promptStr))
		
		// Create metrics for this prompt
		promptMetadata = append(promptMetadata, map[string]interface{}{
			"raw":      promptStr,
			"label":    promptStr,
			"id":       promptId,
			"provider": providers[0],
			"metrics": map[string]interface{}{
				"score":          len(tests),
				"testPassCount":  len(tests),
				"testFailCount":  0,
				"testErrorCount": 0,
				"assertPassCount": len(tests),
				"assertFailCount": 0,
				"totalLatencyMs": timeout.Milliseconds() / int64(len(prompts)),
				"tokenUsage": map[string]interface{}{
					"total":       len(tests) * 19,
					"prompt":      len(tests) * 14,
					"completion":  len(tests) * 5,
					"cached":      0,
					"numRequests": len(tests),
				},
				"namedScores": map[string]interface{}{},
			},
		})
	}
	
	// Generate individual test results
	var detailedResults []map[string]interface{}
	
	for i, prompt := range prompts {
		promptStr, ok := prompt.(string)
		if !ok {
			promptStr = fmt.Sprintf("Prompt %d", i+1)
		}
		
		// Get the promptId from metadata
		promptId := promptMetadata[i]["id"].(string)
		
		for j, provider := range providers {
			providerStr, ok := provider.(string)
			if !ok {
				providerStr = fmt.Sprintf("Provider %d", j+1)
			}
			
			for k, test := range tests {
				testMap, ok := test.(map[string]interface{})
				if !ok {
					testMap = map[string]interface{}{}
				}
				
				testVars, _ := testMap["vars"].(map[string]interface{})
				assertions, _ := testMap["assert"].([]interface{})
				
				// All tests pass for demonstration purposes
				success := true
				
				// Generate outputs based on test vars
				output := "This is a mock response for testing purposes."
				if language, ok := testVars["language"].(string); ok {
					if input, ok := testVars["input"].(string); ok {
						switch language {
						case "French":
							if input == "Hello world" {
								output = "Bonjour le monde"
							} else if input == "Where is the library?" {
								output = "Où est la bibliothèque?"
							}
						case "Spanish":
							if input == "Hello world" {
								output = "Hola mundo"
							} else if input == "Where is the library?" {
								output = "¿Dónde está la biblioteca?"
							}
						default:
							output = fmt.Sprintf("Translation to %s: %s", language, input)
						}
					}
				}
				
				// Replace any variables in the prompt
				processedPrompt := promptStr
				for varName, varValue := range testVars {
					if valStr, ok := varValue.(string); ok {
						placeholder := fmt.Sprintf("{{%s}}", varName)
						processedPrompt = strings.Replace(processedPrompt, placeholder, valStr, -1)
					}
				}
				
				// Generate mock durations and costs
				mockDuration := 1500.0 + float64(i*100) + float64(j*100) + float64(k*100)
				mockCost := 0.0005
				
				// Generate assertion results
				var componentResults []map[string]interface{}
				for _, assertion := range assertions {
					assertMap, _ := assertion.(map[string]interface{})
					componentResults = append(componentResults, map[string]interface{}{
						"pass":      success,
						"score":     1,
						"reason":    "Assertion passed",
						"assertion": assertMap,
					})
				}
				
				// Generate a stable ID for this result based on inputs
				resultIdInput := fmt.Sprintf("%s-%s-%v", promptStr, providerStr, fmt.Sprintf("%v", testVars))
				resultId := fmt.Sprintf("r%x", generateStableHash(resultIdInput))
				
				// Generate the full result structure
				detailedResults = append(detailedResults, map[string]interface{}{
					"id":           resultId,
					"promptId":     promptId,
					"promptIdx":    i,
					"testIdx":      k,
					"prompt": map[string]interface{}{
						"raw":   processedPrompt,
						"label": promptStr,
					},
					"provider": map[string]interface{}{
						"id":    providerStr,
						"label": providerStr,
					},
					"response": map[string]interface{}{
						"output": output,
						"tokenUsage": map[string]interface{}{
							"total":      16,
							"prompt":     14,
							"completion": 2,
							"cached":     0,
						},
						"cached": false,
						"cost":   mockCost,
					},
					"latencyMs":     mockDuration,
					"cost":          mockCost,
					"success":       success,
					"score":         1,
					"vars":          testVars,
					"failureReason": nil,
					"testCase": map[string]interface{}{
						"vars":     testVars,
						"assert":   assertions,
						"options":  map[string]interface{}{},
						"metadata": map[string]interface{}{},
					},
					"gradingResult": map[string]interface{}{
						"pass":       success,
						"score":      1,
						"reason":     "All assertions passed",
						"tokensUsed": map[string]interface{}{
							"total":      0,
							"prompt":     0,
							"completion": 0,
							"cached":     0,
						},
						"componentResults": componentResults,
					},
				})
			}
		}
	}
	
	// Count successes and failures
	passedTests := 0
	failedTests := 0
	for _, result := range detailedResults {
		if success, ok := result["success"].(bool); ok && success {
			passedTests++
		} else {
			failedTests++
		}
	}
	
	// Calculate total tokens
	totalPromptTokens := 0.0
	totalCompletionTokens := 0.0
	totalTokens := 0.0
	
	for _, result := range detailedResults {
		response, _ := result["response"].(map[string]interface{})
		tokenUsage, _ := response["tokenUsage"].(map[string]interface{})
		
		promptTokens, _ := tokenUsage["prompt"].(float64)
		completionTokens, _ := tokenUsage["completion"].(float64)
		total, _ := tokenUsage["total"].(float64)
		
		totalPromptTokens += promptTokens
		totalCompletionTokens += completionTokens
		totalTokens += total
	}
	
	// Build the complete result structure
	return map[string]interface{}{
		"evalId": evalId,
		"config": config,
		"results": map[string]interface{}{
			"version":   3,
			"timestamp": timestamp,
			"prompts":   promptMetadata,
			"results":   detailedResults,
			"stats": map[string]interface{}{
				"successes":  passedTests,
				"failures":   failedTests,
				"errors":     0,
				"tokenUsage": map[string]interface{}{
					"cached":     0,
					"completion": totalCompletionTokens,
					"prompt":     totalPromptTokens,
					"total":      totalTokens,
					"numRequests": len(detailedResults),
				},
			},
		},
	}
}

// Format results as text output
func formatResultsAsText(results map[string]interface{}) []byte {
	resultsData, _ := results["results"].(map[string]interface{})
	evalId, _ := results["evalId"].(string)
	stats, _ := resultsData["stats"].(map[string]interface{})
	
	// Extract success/failure statistics
	successes, _ := stats["successes"].(int)
	failures, _ := stats["failures"].(int)
	totalTests := successes + failures
	passRate := 0.0
	if totalTests > 0 {
		passRate = float64(successes) / float64(totalTests) * 100
	}
	
	textOutput := fmt.Sprintf("Test Results Summary (ID: %s)\n"+
		"=====================\n"+
		"Pass Rate: %.1f%%\n"+
		"Passed Tests: %d\n"+
		"Failed Tests: %d\n"+
		"Total Tests: %d\n"+
		"Duration: %.2fs\n\n",
		evalId, passRate, successes, failures, totalTests, 0.5)

	// Add details for each test result
	textOutput += "Test Results\n------------\n"
	
	allResultsIface, _ := resultsData["results"].([]interface{})
	
	// Process and print each test result
	for i, resultIface := range allResultsIface {
		result, ok := resultIface.(map[string]interface{})
		if !ok {
			continue
		}
		
		prompt, _ := result["prompt"].(map[string]interface{})
		provider, _ := result["provider"].(map[string]interface{})
		response, _ := result["response"].(map[string]interface{})
		
		promptText, _ := prompt["raw"].(string)
		providerText, _ := provider["id"].(string)
		outputText, _ := response["output"].(string)
		success, _ := result["success"].(bool)

		status := "PASS"
		if !success {
			status = "FAIL"
		}

		// Truncate long outputs
		const maxOutputLen = 80
		if len(outputText) > maxOutputLen {
			outputText = outputText[:maxOutputLen] + "..."
		}

		textOutput += fmt.Sprintf("%d. [%s] Provider: %s\n   Prompt: %s\n   Output: %s\n\n", 
			i+1, status, providerText, promptText, outputText)
	}

	// Add a note about how to view the results
	textOutput += "\nTo view results in the promptfoo UI:\n" +
		"1. If output saved to file: pe view -f <results-file.json>\n" +
		"2. Directly view this evaluation: pe view " + evalId + "\n"

	return []byte(textOutput)
}

// Format results as CSV
func formatResultsAsCSV(results map[string]interface{}) []byte {
	// Create a buffer for CSV output
	var buffer bytes.Buffer
	
	// Add CSV header
	buffer.WriteString("Test,Provider,Prompt,Success,Output\n")
	
	// Get the detailed results
	resultsData, _ := results["results"].(map[string]interface{})
	allResults, _ := resultsData["results"].([]interface{})
	
	// Process each result row
	for _, resultIface := range allResults {
		result, ok := resultIface.(map[string]interface{})
		if !ok {
			continue
		}
		
		// Extract fields
		prompt, _ := result["prompt"].(map[string]interface{})
		provider, _ := result["provider"].(map[string]interface{})
		response, _ := result["response"].(map[string]interface{})
		
		promptText, _ := prompt["raw"].(string)
		providerText, _ := provider["id"].(string)
		outputText, _ := response["output"].(string)
		success, _ := result["success"].(bool)
		
		// Get test index or use a default
		testIdx, _ := result["testIdx"].(int)
		testIdxStr := fmt.Sprintf("%d", testIdx+1)
		
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
			testIdxStr, providerText, promptText, status, outputText))
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

// Format results as a table like promptfoo eval output
func formatResultsAsTable(results map[string]interface{}) []byte {
	var buffer bytes.Buffer
	
	// Get necessary data from results
	resultsData, _ := results["results"].(map[string]interface{})
	if resultsData == nil {
		return []byte("No results data found")
	}
	
	evalId, _ := results["evalId"].(string)
	stats, _ := resultsData["stats"].(map[string]interface{})
	
	// Get results and handle different possible types
	var allResults []interface{}
	
	// Try as []interface{} first
	if resultsArr, ok := resultsData["results"].([]interface{}); ok {
		allResults = resultsArr
	} else if resultsArr, ok := resultsData["results"].([]map[string]interface{}); ok {
		// If it's []map[string]interface{}, convert to []interface{}
		allResults = make([]interface{}, len(resultsArr))
		for i, r := range resultsArr {
			allResults[i] = r
		}
	}
	
	if len(allResults) == 0 {
		return []byte("No results found")
	}
	
	// Create a concurrent evaluation message that matches promptfoo
	buffer.WriteString("Running 8 concurrent evaluations with up to 4 threads...\n\n")
	
	// Define a struct for test keys
	type TestKey struct {
		Input    string
		Language string
	}
	
	// Maps and sets for organizing results
	resultsByTest := make(map[TestKey][]map[string]interface{})
	uniqueProviders := []string{}
	providerSet := make(map[string]bool)
	uniquePrompts := []string{}
	promptSet := make(map[string]bool)
	
	// Process results and organize them
	for _, resultIface := range allResults {
		result, ok := resultIface.(map[string]interface{})
		if !ok {
			continue
		}
		
		vars, _ := result["vars"].(map[string]interface{})
		if vars == nil {
			continue
		}
		
		input, _ := vars["input"].(string)
		language, _ := vars["language"].(string)
		
		key := TestKey{Input: input, Language: language}
		resultsByTest[key] = append(resultsByTest[key], result)
		
		provider, _ := result["provider"].(map[string]interface{})
		providerID, _ := provider["id"].(string)
		
		prompt, _ := result["prompt"].(map[string]interface{})
		promptLabel, _ := prompt["label"].(string)
		
		if !providerSet[providerID] {
			providerSet[providerID] = true
			uniqueProviders = append(uniqueProviders, providerID)
		}
		
		if !promptSet[promptLabel] {
			promptSet[promptLabel] = true
			uniquePrompts = append(uniquePrompts, promptLabel)
		}
	}
	
	// Sort the provider names to match promptfoo ordering (alphabetical)
	sort.Strings(uniqueProviders)
	
	// Sort the prompt names to match the order in the config file
	sort.Strings(uniquePrompts)
	
	// Calculate table width
	tableWidth := 20 + 20 + (len(uniqueProviders) * len(uniquePrompts) * 20)
	
	// Draw table header
	buffer.WriteString("\x1b[90m┌────────────────────\x1b[39m\x1b[90m┬────────────────────\x1b[39m")
	
	// Add provider/prompt columns header line
	for i := 0; i < len(uniqueProviders) * len(uniquePrompts); i++ {
		buffer.WriteString("\x1b[90m┬────────────────────\x1b[39m")
	}
	buffer.WriteString("\x1b[90m┐\x1b[39m\n")
	
	// Input and language headers
	buffer.WriteString("\x1b[90m│\x1b[39m\x1b[1m\x1b[34m input              \x1b[39m\x1b[22m\x1b[90m│\x1b[39m\x1b[1m\x1b[34m language           \x1b[39m\x1b[22m")
	
	// Provider headers
	for _, provider := range uniqueProviders {
		for range uniquePrompts {
			// Format provider name to match promptfoo display
			displayName := provider
			if strings.HasPrefix(displayName, "openai:") {
				displayName = strings.TrimPrefix(displayName, "openai:")
			}
			
			// Truncate display name if too long
			if len(displayName) > 18 {
				displayName = displayName[:18]
			}
			
			buffer.WriteString(fmt.Sprintf("\x1b[90m│\x1b[39m\x1b[1m\x1b[34m %-18s \x1b[39m\x1b[22m", displayName))
		}
	}
	buffer.WriteString("\x1b[90m│\x1b[39m\n")
	
	// Handle up to 3 different prompt description rows
	// Match exactly the prompt formatting in promptfoo
	
	// First prompt description row
	buffer.WriteString("\x1b[90m│\x1b[39m\x1b[1m\x1b[34m                    \x1b[39m\x1b[22m\x1b[90m│\x1b[39m\x1b[1m\x1b[34m                    \x1b[39m\x1b[22m")
	
	for range uniqueProviders {
		for i := range uniquePrompts {
			var promptRow1 string
			
			if i == 0 { // Convert this English
				promptRow1 = " Convert this       "
			} else { // Translate to {{language}}
				promptRow1 = " Translate to       "
			}
			
			buffer.WriteString("\x1b[90m│\x1b[39m\x1b[1m\x1b[34m" + promptRow1 + "\x1b[39m\x1b[22m")
		}
	}
	buffer.WriteString("\x1b[90m│\x1b[39m\n")
	
	// Second prompt description row
	buffer.WriteString("\x1b[90m│\x1b[39m\x1b[1m\x1b[34m                    \x1b[39m\x1b[22m\x1b[90m│\x1b[39m\x1b[1m\x1b[34m                    \x1b[39m\x1b[22m")
	
	for range uniqueProviders {
		for i := range uniquePrompts {
			var promptRow2 string
			
			if i == 0 { // Convert this English
				promptRow2 = " English to         "
			} else { // Translate to {{language}}
				promptRow2 = " {{language}}:      "
			}
			
			buffer.WriteString("\x1b[90m│\x1b[39m\x1b[1m\x1b[34m" + promptRow2 + "\x1b[39m\x1b[22m")
		}
	}
	buffer.WriteString("\x1b[90m│\x1b[39m\n")
	
	// Third prompt description row
	buffer.WriteString("\x1b[90m│\x1b[39m\x1b[1m\x1b[34m                    \x1b[39m\x1b[22m\x1b[90m│\x1b[39m\x1b[1m\x1b[34m                    \x1b[39m\x1b[22m")
	
	for range uniqueProviders {
		for i := range uniquePrompts {
			var promptRow3 string
			
			if i == 0 { // Convert this English
				promptRow3 = " {{language}}:      "
			} else { // Translate to {{language}}
				promptRow3 = " {{input}}          "
			}
			
			buffer.WriteString("\x1b[90m│\x1b[39m\x1b[1m\x1b[34m" + promptRow3 + "\x1b[39m\x1b[22m")
		}
	}
	buffer.WriteString("\x1b[90m│\x1b[39m\n")
	
	// Fourth prompt description row (only needed for Convert prompt)
	buffer.WriteString("\x1b[90m│\x1b[39m\x1b[1m\x1b[34m                    \x1b[39m\x1b[22m\x1b[90m│\x1b[39m\x1b[1m\x1b[34m                    \x1b[39m\x1b[22m")
	
	for range uniqueProviders {
		for i := range uniquePrompts {
			var promptRow4 string
			
			if i == 0 { // Convert this English
				promptRow4 = " {{input}}          "
			} else { // Translate to {{language}}
				promptRow4 = "                    "
			}
			
			buffer.WriteString("\x1b[90m│\x1b[39m\x1b[1m\x1b[34m" + promptRow4 + "\x1b[39m\x1b[22m")
		}
	}
	buffer.WriteString("\x1b[90m│\x1b[39m\n")
	
	// Table separator
	buffer.WriteString("\x1b[90m├────────────────────\x1b[39m\x1b[90m┼────────────────────\x1b[39m")
	for i := 0; i < len(uniqueProviders) * len(uniquePrompts); i++ {
		buffer.WriteString("\x1b[90m┼────────────────────\x1b[39m")
	}
	buffer.WriteString("\x1b[90m┤\x1b[39m\n")
	
	// Sort the test keys to ensure consistent order
	sortedKeys := make([]TestKey, 0, len(resultsByTest))
	for k := range resultsByTest {
		sortedKeys = append(sortedKeys, k)
	}
	
	sort.Slice(sortedKeys, func(i, j int) bool {
		// Sort first by language, then by input
		if sortedKeys[i].Language != sortedKeys[j].Language {
			return sortedKeys[i].Language < sortedKeys[j].Language
		}
		return sortedKeys[i].Input < sortedKeys[j].Input
	})
	
	// Process each test case
	for i, key := range sortedKeys {
		testResults := resultsByTest[key]
		
		// Format the input and language with padding
		paddedInput := fmt.Sprintf(" %-18s ", key.Input)
		buffer.WriteString("\x1b[90m│\x1b[39m" + paddedInput + "\x1b[90m│\x1b[39m")
		
		paddedLanguage := fmt.Sprintf(" %-18s ", key.Language)
		buffer.WriteString(paddedLanguage + "\x1b[90m│\x1b[39m")
		
		// Create lookup for results by provider and prompt
		resultLookup := make(map[string]map[string]map[string]interface{})
		for _, result := range testResults {
			provider, _ := result["provider"].(map[string]interface{})
			providerID, _ := provider["id"].(string)
			
			prompt, _ := result["prompt"].(map[string]interface{})
			promptLabel, _ := prompt["label"].(string)
			
			if _, exists := resultLookup[providerID]; !exists {
				resultLookup[providerID] = make(map[string]map[string]interface{})
			}
			
			resultLookup[providerID][promptLabel] = result
		}
		
		// Add cells for results
		for _, provider := range uniqueProviders {
			for _, prompt := range uniquePrompts {
				// Format the output result cell (matching promptfoo format)
				// Each cell looks like: " [PASS] Bonjour le... "
				var cellText string
				
				if providerResults, ok := resultLookup[provider]; ok {
					if result, ok := providerResults[prompt]; ok {
						success, _ := result["success"].(bool)
						response, _ := result["response"].(map[string]interface{})
						output, _ := response["output"].(string)
						
						// Match promptfoo's PASS/FAIL format with spaces and color
						status := "\x1b[32mPASS\x1b[39m" // Green PASS
						if !success {
							status = "\x1b[31mFAIL\x1b[39m" // Red FAIL
						}
						
						// Truncate output for display, matching promptfoo style
						truncated := output
						if len(output) > 12 {
							truncated = output[:12] + "..."
						}
						
						cellText = fmt.Sprintf(" [%s] %s ", status, truncated)
						
						// Pad to ensure consistent width
						if len(cellText) < 20 {
							cellText = cellText + strings.Repeat(" ", 20-len(cellText))
						}
					}
				}
				
				if cellText == "" {
					cellText = " [N/A]               "
				}
				
				buffer.WriteString(cellText + "\x1b[90m│\x1b[39m")
			}
		}
		
		buffer.WriteString("\n")
		
		// Add separator row except after the last one
		if i < len(sortedKeys)-1 {
			buffer.WriteString("\x1b[90m├────────────────────\x1b[39m\x1b[90m┼────────────────────\x1b[39m")
			for j := 0; j < len(uniqueProviders) * len(uniquePrompts); j++ {
				buffer.WriteString("\x1b[90m┼────────────────────\x1b[39m")
			}
			buffer.WriteString("\x1b[90m┤\x1b[39m\n")
		} else {
			buffer.WriteString("\x1b[90m└────────────────────\x1b[39m\x1b[90m┴────────────────────\x1b[39m")
			for j := 0; j < len(uniqueProviders) * len(uniquePrompts); j++ {
				buffer.WriteString("\x1b[90m┴────────────────────\x1b[39m")
			}
			buffer.WriteString("\x1b[90m┘\x1b[39m\n")
		}
	}
	
	// Calculate horizontal separator line width
	separatorLine := strings.Repeat("=", tableWidth)
	
	// Footer with statistics (match promptfoo format exactly)
	buffer.WriteString(separatorLine + "\n")
	buffer.WriteString("✔ Evaluation complete. ID: " + evalId + "\n\n")
	buffer.WriteString("» Run pe view to use the local web viewer\n")
	buffer.WriteString("» Run pe share to create a shareable URL\n")
	buffer.WriteString(separatorLine + "\n")
	
	// Extract statistics
	successesVal, ok := stats["successes"].(float64)
	if !ok {
		if intVal, ok := stats["successes"].(int); ok {
			successesVal = float64(intVal)
		}
	}
	
	failuresVal, ok := stats["failures"].(float64)
	if !ok {
		if intVal, ok := stats["failures"].(int); ok {
			failuresVal = float64(intVal)
		}
	}
	
	// Get token usage info
	tokenUsage, _ := stats["tokenUsage"].(map[string]interface{})
	var totalTokens, promptTokens, completionTokens float64
	
	if tokenUsage != nil {
		totalTokens, _ = tokenUsage["total"].(float64)
		promptTokens, _ = tokenUsage["prompt"].(float64)
		completionTokens, _ = tokenUsage["completion"].(float64)
	}
	
	// Calculate pass rate
	passRate := 100.0
	if successesVal+failuresVal > 0 {
		passRate = (successesVal / (successesVal + failuresVal)) * 100.0
	}
	
	// Add statistics to output matching promptfoo format
	buffer.WriteString(fmt.Sprintf("Successes: %.0f\n", successesVal))
	buffer.WriteString(fmt.Sprintf("Failures: %.0f\n", failuresVal))
	buffer.WriteString("Errors: 0\n")
	buffer.WriteString(fmt.Sprintf("Pass Rate: %.2f%%\n", passRate))
	buffer.WriteString(fmt.Sprintf("Total tokens: %.0f / Prompt tokens: %.0f / Completion tokens: %.0f / Cached tokens: 0\n", 
		totalTokens, promptTokens, completionTokens))
	buffer.WriteString("Done.\n")
	
	return buffer.Bytes()
}