package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/tmc/pe/internal/cgpt"
	"github.com/tmc/pe/internal/promptfoo"
)

// evaluateWithCGPT implements a CGPT-based evaluator that can be used directly
// when external evaluators aren't available.
func evaluateWithCGPT(config map[string]interface{}) (map[string]interface{}, error) {
	return evaluateWithCGPTOptions(config, false)
}

// evaluateWithCGPTDryRun implements a dry-run version of the CGPT evaluator
func evaluateWithCGPTDryRun(config map[string]interface{}) (map[string]interface{}, error) {
	// For dry run, just print commands for all providers specified in the config
	prompts, _ := config["prompts"].([]interface{})
	providers, _ := config["providers"].([]interface{})
	tests, _ := config["tests"].([]interface{})
	
	if len(prompts) == 0 || len(tests) == 0 || len(providers) == 0 {
		return nil, fmt.Errorf("configuration missing required fields")
	}
	
	// Add a minimal header explaining what these commands are
	fmt.Println("# Commands to run each prompt with each provider/model:")
	fmt.Println()
	
	// Group commands by provider for better organization
	for _, provider := range providers {
		providerStr, ok := provider.(string)
		if !ok {
			continue
		}
		
		// Parse provider to get backend and model
		parts := strings.Split(providerStr, ":")
		backend := parts[0]
		model := ""
		if len(parts) > 1 {
			model = parts[1]
		}
		
		// For each prompt and test case
		for _, prompt := range prompts {
			promptStr, ok := prompt.(string)
			if !ok {
				continue
			}
			
			for _, test := range tests {
				testMap, ok := test.(map[string]interface{})
				if !ok {
					continue
				}
				
				testVars, _ := testMap["vars"].(map[string]interface{})
				
				// Process template variables
				processedPrompt := promptStr
				for varName, varValue := range testVars {
					valStr := fmt.Sprintf("%v", varValue)
					placeholder := fmt.Sprintf("{{%s}}", varName)
					processedPrompt = strings.ReplaceAll(processedPrompt, placeholder, valStr)
				}
				
				// Format command for easy execution in shell scripts
				// Using %q handles shell escaping properly to avoid issues with quotes and special characters
				fmt.Printf("cgpt -b %s -m %s %q\n", backend, model, processedPrompt)
			}
		}
	}
	
	// Add a brief footer comment
	fmt.Println("\n# Save these commands to a file and run with 'bash filename' to execute them")
	
	// Don't return results in dry run mode - just exit after showing commands
	os.Exit(0)
	return nil, nil // This line is never reached but keeps the compiler happy
}

// evaluateWithCGPTOptions is the shared implementation for CGPT evaluator with options
func evaluateWithCGPTOptions(config map[string]interface{}, dryRun bool) (map[string]interface{}, error) {
	// Extract configuration
	prompts, _ := config["prompts"].([]interface{})
	providers, _ := config["providers"].([]interface{})
	tests, _ := config["tests"].([]interface{})
	
	if len(prompts) == 0 {
		return nil, fmt.Errorf("no prompts found in configuration")
	}
	
	if len(tests) == 0 {
		return nil, fmt.Errorf("no tests found in configuration")
	}
	
	// Create a map of providers from the config
	providerConfigs := make(map[string]*cgpt.ModelProvider)
	
	// If no providers specified, use the default
	if len(providers) == 0 {
		providerConfigs["default"] = cgpt.DefaultProvider()
	} else {
		// Process each provider from the config
		for _, p := range providers {
			providerStr, ok := p.(string)
			if !ok {
				continue
			}
			
			// Parse provider string into backend and model
			parts := strings.Split(providerStr, ":")
			backend := parts[0]
			model := ""
			if len(parts) > 1 {
				model = parts[1]
			}
			
			// Create provider configuration
			providerConfigs[providerStr] = &cgpt.ModelProvider{
				Backend:     backend,
				Model:       model,
				MaxTokens:   1024,
				Temperature: 0.7,
			}
		}
	}
	
	// Make sure we have at least one provider
	if len(providerConfigs) == 0 {
		// Fall back to default provider
		providerConfigs["default"] = cgpt.DefaultProvider()
	}
	
	// Create a unique evalId for this run matching promptfoo's format exactly
	evalId := fmt.Sprintf("eval-%s-%s", 
		randomString(3), 
		time.Now().Format("2006-01-02T15:04:05"))
	timestamp := time.Now().Format(time.RFC3339)
	
	// Prepare result structures
	var promptMetadata []map[string]interface{}
	var detailedResults []map[string]interface{}
	
	// We'll evaluate for each provider in the config
	for _, providerItem := range providers {
		providerStr, ok := providerItem.(string)
		if !ok {
			continue
		}
		
		// Get the provider config
		provider, ok := providerConfigs[providerStr]
		if !ok {
			fmt.Fprintf(os.Stderr, "Warning: Provider '%s' not configured, skipping\n", providerStr)
			continue
		}
		
		// Process each prompt for this provider
		for i, prompt := range prompts {
			promptStr, ok := prompt.(string)
			if !ok {
				continue
			}
			
			// Generate a unique ID for this prompt
			promptId := fmt.Sprintf("p%x", generateStableHash(promptStr))
			
			// Create metrics structure for this prompt
			promptMetrics := &promptfoo.PromptMetrics{
				Score:           1.0,
				TestPassCount:   int32(len(tests)),
				TestFailCount:   0,
				AssertPassCount: int32(len(tests)),
				AssertFailCount: 0,
				TokenUsage: &promptfoo.TokenUsage{
					Total:       0,
					Prompt:      0,
					Completion:  0,
					NumRequests: int32(len(tests)),
				},
				NamedScores:      make(map[string]float64),
				NamedScoresCount: make(map[string]int32),
			}
			
			// Initialize metrics in the format expected by the results
			promptMetadataItem := map[string]interface{}{
				"raw":      promptStr,
				"label":    promptStr,
				"id":       promptId,
				"provider": providerStr,
				"metrics": map[string]interface{}{
					"score":          len(tests),
					"testPassCount":  len(tests),
					"testFailCount":  0,
					"assertPassCount": len(tests),
					"assertFailCount": 0,
					"tokenUsage": map[string]interface{}{
						"total":       0,
						"prompt":      0,
						"completion":  0,
						"cached":      0,
						"numRequests": len(tests),
					},
					"namedScores":      map[string]interface{}{},
					"namedScoresCount": map[string]interface{}{},
					"cost":             0.0,
				},
			}
			promptMetadata = append(promptMetadata, promptMetadataItem)
			
			// Process each test for this prompt
			for j, test := range tests {
				testMap, ok := test.(map[string]interface{})
				if !ok {
					continue
				}
				
				testVars, _ := testMap["vars"].(map[string]interface{})
				assertions, _ := testMap["assert"].([]interface{})
				
				// Evaluate against the model with dry-run option if specified
				response, err := provider.EvaluatePromptWithOptions(promptStr, testVars, dryRun)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error evaluating prompt with provider %s: %v\n", providerStr, err)
					continue
				}
				
				// Check assertions
				success := true
				var componentResults []map[string]interface{}
				
				for _, assertion := range assertions {
					assertMap, ok := assertion.(map[string]interface{})
					if !ok {
						continue
					}
					
					// Get assertion parameters
					assertType, _ := assertMap["type"].(string)
					assertValue, _ := assertMap["value"]
					
					// Check assertion
					pass := checkAssertion(response.Output, assertType, assertValue)
					success = success && pass
					
					// Generate assertion result
					reason := "Assertion passed"
					if !pass {
						reason = fmt.Sprintf("Output does not %s: %v", assertType, assertValue)
					}
					
					componentResults = append(componentResults, map[string]interface{}{
						"pass":      pass,
						"score":     1.0,
						"reason":    reason,
						"assertion": assertMap,
					})
				}
				
				// Update token usage metrics
				if response.TokenUsage != nil {
					promptMetrics.TokenUsage.Total += response.TokenUsage.Total
					promptMetrics.TokenUsage.Prompt += response.TokenUsage.Prompt
					promptMetrics.TokenUsage.Completion += response.TokenUsage.Completion
					
					promptMetadataItem["metrics"].(map[string]interface{})["tokenUsage"].(map[string]interface{})["total"] = int(promptMetrics.TokenUsage.Total)
					promptMetadataItem["metrics"].(map[string]interface{})["tokenUsage"].(map[string]interface{})["prompt"] = int(promptMetrics.TokenUsage.Prompt)
					promptMetadataItem["metrics"].(map[string]interface{})["tokenUsage"].(map[string]interface{})["completion"] = int(promptMetrics.TokenUsage.Completion)
				}
				
				// Update cost metrics
				promptMetadataItem["metrics"].(map[string]interface{})["cost"] = promptMetadataItem["metrics"].(map[string]interface{})["cost"].(float64) + response.Cost
				
				// Generate a stable ID for this result based on inputs
				resultIdInput := fmt.Sprintf("%s-%v", promptStr, testVars)
				resultId := fmt.Sprintf("r%x", generateStableHash(resultIdInput))
				
				// Generate grading result
				gradingResult := map[string]interface{}{
					"pass":   success,
					"score":  1.0,
					"reason": "All assertions passed",
					"namedScores": map[string]interface{}{},
					"tokensUsed": map[string]interface{}{
						"total":      0,
						"prompt":     0,
						"completion": 0,
						"cached":     0,
					},
					"componentResults": componentResults,
				}
				
				if !success {
					gradingResult["reason"] = "Some assertions failed"
					promptMetrics.TestFailCount++
					promptMetadataItem["metrics"].(map[string]interface{})["testFailCount"] = int(promptMetrics.TestFailCount)
					promptMetrics.TestPassCount--
					promptMetadataItem["metrics"].(map[string]interface{})["testPassCount"] = int(promptMetrics.TestPassCount)
				}
				
				// Add formatted variables to the result
				processedPrompt := promptStr
				for varName, varValue := range testVars {
					valStr := fmt.Sprintf("%v", varValue)
					placeholder := fmt.Sprintf("{{%s}}", varName)
					processedPrompt = strings.ReplaceAll(processedPrompt, placeholder, valStr)
				}
				
				// Add the result
				detailedResults = append(detailedResults, map[string]interface{}{
					"id":        resultId,
					"promptId":  promptId,
					"promptIdx": i,
					"testIdx":   j,
					"prompt": map[string]interface{}{
						"raw":   processedPrompt,
						"label": promptStr,
					},
					"provider": map[string]interface{}{
						"id":    providerStr,
						"label": "",
					},
					"response": map[string]interface{}{
						"output": response.Output,
						"tokenUsage": map[string]interface{}{
							"total":      int(response.TokenUsage.Total),
							"prompt":     int(response.TokenUsage.Prompt),
							"completion": int(response.TokenUsage.Completion),
							"completionDetails": map[string]interface{}{
								"reasoning":          0,
								"acceptedPrediction": int(response.TokenUsage.Completion),
								"rejectedPrediction": 0,
							},
						},
						"cached": false,
						"cost":   response.Cost,
					},
					"latencyMs":     0,
					"cost":          response.Cost,
					"success":       success,
					"score":         1.0,
					"namedScores":   map[string]interface{}{},
					"vars":          testVars,
					"metadata":      map[string]interface{}{},
					"failureReason": 0,
					"testCase": map[string]interface{}{
						"vars":     testVars,
						"assert":   assertions,
						"options":  map[string]interface{}{},
						"metadata": map[string]interface{}{},
					},
					"gradingResult": gradingResult,
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
	
	// Build the full promptfoo-compatible result structure
	tokenUsage := map[string]interface{}{
		"cached":     0,
		"completion": 0,
		"prompt":     0,
		"total":      0,
		"numRequests": len(detailedResults),
		"completionDetails": map[string]interface{}{
			"reasoning":          0,
			"acceptedPrediction": 0,
			"rejectedPrediction": 0,
		},
	}
	
	// Sum up token usage across all results
	for _, result := range detailedResults {
		response, ok := result["response"].(map[string]interface{})
		if !ok {
			continue
		}
		
		usage, ok := response["tokenUsage"].(map[string]interface{})
		if !ok {
			continue
		}
		
		tokenUsage["total"] = tokenUsage["total"].(int) + usage["total"].(int)
		tokenUsage["prompt"] = tokenUsage["prompt"].(int) + usage["prompt"].(int)
		tokenUsage["completion"] = tokenUsage["completion"].(int) + usage["completion"].(int)
	}
	
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
				"tokenUsage": tokenUsage,
			},
		},
		"shareableUrl": nil,
	}, nil
}

// checkAssertion evaluates whether an assertion passes for a given output
func checkAssertion(output string, assertType string, assertValue interface{}) bool {
	switch assertType {
	case "contains":
		valueStr, ok := assertValue.(string)
		if !ok {
			return false
		}
		return strings.Contains(output, valueStr)
		
	case "icontains":
		valueStr, ok := assertValue.(string)
		if !ok {
			return false
		}
		return strings.Contains(
			strings.ToLower(output), 
			strings.ToLower(valueStr),
		)
		
	case "not-contains":
		valueStr, ok := assertValue.(string)
		if !ok {
			return false
		}
		return !strings.Contains(output, valueStr)
		
	case "equals":
		valueStr, ok := assertValue.(string)
		if !ok {
			return false
		}
		return output == valueStr
		
	case "iequals":
		valueStr, ok := assertValue.(string)
		if !ok {
			return false
		}
		return strings.EqualFold(output, valueStr)
		
	case "regex":
		valueStr, ok := assertValue.(string)
		if !ok {
			return false
		}
		re, err := regexp.Compile(valueStr)
		if err != nil {
			return false
		}
		return re.MatchString(output)
		
	case "starts-with":
		valueStr, ok := assertValue.(string)
		if !ok {
			return false
		}
		return strings.HasPrefix(output, valueStr)
		
	case "ends-with":
		valueStr, ok := assertValue.(string)
		if !ok {
			return false
		}
		return strings.HasSuffix(output, valueStr)
		
	default:
		// Unknown assertion type
		return false
	}
}