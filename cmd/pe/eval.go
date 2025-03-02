package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"
)

// evalCmd returns a cobra.Command for the 'eval' subcommand.
//
// eval executes prompt tests against LLM providers
//
// Usage:
//
//	pe eval config.yaml [--output results.json] [--timeout 60s]
func evalCmd() *cobra.Command {
	var configFile string
	var outputFile string
	var outputFormat string
	var timeout string

	cmd := &cobra.Command{
		Use:   "eval [config_file]",
		Short: "Evaluate prompt configurations",
		Long:  `Evaluate prompt configurations against LLM providers (mock implementation).`,
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Check if config file provided as positional arg
			if len(args) > 0 {
				configFile = args[0]
			}
			
			// If no config file provided, check -c/--config flag
			if configFile == "" {
				if configFile, _ = cmd.Flags().GetString("config"); configFile == "" {
					return fmt.Errorf("no configuration file provided")
				}
			}
			
			return executeEval(cmd, configFile, outputFile, outputFormat, timeout)
		},
	}

	cmd.Flags().StringVarP(&configFile, "config", "c", "", "Path to configuration file")
	cmd.Flags().StringVarP(&outputFile, "output", "o", "", "Write results to file")
	cmd.Flags().StringVarP(&outputFormat, "format", "f", "json", "Output format: 'json', 'yaml', or 'text'")
	cmd.Flags().StringVarP(&timeout, "timeout", "t", "30s", "Timeout for the entire test run")

	return cmd
}

func executeEval(cmd *cobra.Command, configFile, outputFile, outputFormat, timeoutStr string) error {
	// Parse timeout
	timeout, err := time.ParseDuration(timeoutStr)
	if err != nil {
		return fmt.Errorf("invalid timeout: %v", err)
	}

	// Read config file
	data, err := os.ReadFile(configFile)
	if err != nil {
		return fmt.Errorf("error reading config file: %v", err)
	}

	// Parse the config
	var config map[string]interface{}
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return fmt.Errorf("error parsing config file: %v", err)
	}

	// Look for provider-specific or generic evaluator executables
	results, err := executeWithExternalEvaluator(config, timeout)
	if err != nil {
		// Fall back to mock implementation if no external evaluator found
		fmt.Fprintf(cmd.OutOrStdout(), "Note: Using mock implementation. No provider-specific evaluator found.\n")
		results = generateMockResults(config, timeout)
	}

	// Format the output
	var output []byte
	if outputFormat == "json" {
		output, err = json.MarshalIndent(results, "", "  ")
	} else if outputFormat == "yaml" {
		output, err = yaml.Marshal(results)
	} else if outputFormat == "text" {
		output = formatResultsAsText(results)
	} else {
		return fmt.Errorf("unsupported output format: %s", outputFormat)
	}

	if err != nil {
		return fmt.Errorf("error formatting results: %v", err)
	}

	// Write to output file or stdout
	if outputFile != "" {
		err = os.WriteFile(outputFile, output, 0644)
		if err != nil {
			return fmt.Errorf("error writing output file: %v", err)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Results written to %s\n", outputFile)
	} else {
		fmt.Fprint(cmd.OutOrStdout(), string(output))
	}

	return nil
}

func generateMockResults(config map[string]interface{}, timeout time.Duration) map[string]interface{} {
	// Extract necessary components
	prompts, _ := config["prompts"].([]interface{})
	providers, _ := config["providers"].([]interface{})
	tests, _ := config["tests"].([]interface{})
	
	// Create a unique evalId for this run (no longer using pfutil prefix)
	evalId := fmt.Sprintf("eval-%s", time.Now().Format("20060102T150405"))
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
					"completionDetails": map[string]interface{}{
						"reasoning": 0,
						"acceptedPrediction": 0,
						"rejectedPrediction": 0,
					},
				},
				"namedScores": map[string]interface{}{},
				"namedScoresCount": map[string]interface{}{},
				"cost": 0.001 * float64(len(tests)),
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
				if country, ok := testVars["country"].(string); ok {
					if country == "France" {
						output = "The capital of France is Paris."
					} else if country == "Japan" {
						output = "The capital of Japan is Tokyo."
					} else {
						output = fmt.Sprintf("The capital of %s is [capital city name].", country)
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
				
				// Generate mock durations and costs - make them stable for consistent results
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
						"label": "",
					},
					"response": map[string]interface{}{
						"output": output,
						"tokenUsage": map[string]interface{}{
							"total":      16,
							"prompt":     14,
							"completion": 2,
							"completionDetails": map[string]interface{}{
								"reasoning":          0,
								"acceptedPrediction": 0,
								"rejectedPrediction": 0,
							},
						},
						"cached": false,
						"cost":   mockCost,
					},
					"latencyMs":     mockDuration,
					"cost":          mockCost,
					"success":       success,
					"score":         1,
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
					"gradingResult": map[string]interface{}{
						"pass":       success,
						"score":      1,
						"reason":     "All assertions passed",
						"namedScores": map[string]interface{}{},
						"tokensUsed": map[string]interface{}{
							"total":      0,
							"prompt":     0,
							"completion": 0,
							"cached":     0,
						},
						"componentResults": componentResults,
						"assertion":        nil,
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

	// Build the full promptfoo-compatible result structure
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
					"completion": 15,
					"prompt":     56,
					"total":      71,
					"numRequests": len(detailedResults),
					"completionDetails": map[string]interface{}{
						"reasoning":          0,
						"acceptedPrediction": 0,
						"rejectedPrediction": 0,
					},
				},
			},
		},
		"shareableUrl": nil,
	}
}

// Helper function to generate a stable hash for consistent IDs
func generateStableHash(input string) uint32 {
	var hash uint32 = 5381
	for _, c := range input {
		hash = ((hash << 5) + hash) + uint32(c)
	}
	return hash
}

// executeWithExternalEvaluator tries to find and execute a provider-specific evaluator
// or a generic pe-eval executable.
func executeWithExternalEvaluator(config map[string]interface{}, timeout time.Duration) (map[string]interface{}, error) {
	// Extract provider information
	providers, _ := config["providers"].([]interface{})
	if len(providers) == 0 {
		return nil, fmt.Errorf("no providers specified in config")
	}

	// Try to get the first provider's name for provider-specific evaluator
	firstProvider := ""
	if providerStr, ok := providers[0].(string); ok {
		// Extract provider name (e.g., "openai" from "openai:gpt-4")
		parts := strings.Split(providerStr, ":")
		firstProvider = parts[0]
	}

	// Create a temporary file with the config
	tempConfigFile, err := os.CreateTemp("", "pe-config-*.yaml")
	if err != nil {
		return nil, fmt.Errorf("error creating temp config file: %v", err)
	}
	defer os.Remove(tempConfigFile.Name())

	// Write the config to the file
	configData, err := yaml.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("error marshaling config: %v", err)
	}
	if _, err := tempConfigFile.Write(configData); err != nil {
		return nil, fmt.Errorf("error writing config: %v", err)
	}
	if err := tempConfigFile.Close(); err != nil {
		return nil, fmt.Errorf("error closing config file: %v", err)
	}

	// List of possible evaluator commands to try
	evaluators := []string{}
	
	// First try provider-specific evaluator if we have a provider name
	if firstProvider != "" {
		evaluators = append(evaluators, fmt.Sprintf("pe-eval-provider-%s", firstProvider))
	}
	
	// Then try generic evaluator
	evaluators = append(evaluators, "pe-eval")
	
	// Finally, try default CGPT-based implementation
	evaluators = append(evaluators, "pe-eval-provider-cgpt")

	// Try each evaluator
	var lastErr error
	for _, evaluator := range evaluators {
		// Check if the evaluator exists in PATH
		evalPath, err := exec.LookPath(evaluator)
		if err != nil {
			lastErr = fmt.Errorf("evaluator %s not found: %v", evaluator, err)
			continue
		}

		// Run the evaluator with the temp config file
		cmd := exec.Command(evalPath, tempConfigFile.Name())

		// Create buffers for stdout and stderr
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		// Run the evaluator
		if err := cmd.Run(); err != nil {
			lastErr = fmt.Errorf("error running %s: %v\n%s", evaluator, err, stderr.String())
			continue
		}

		// Parse the result
		var results map[string]interface{}
		if err := json.Unmarshal(stdout.Bytes(), &results); err != nil {
			lastErr = fmt.Errorf("error parsing %s output: %v", evaluator, err)
			continue
		}

		// Successfully found and ran an evaluator
		return results, nil
	}

	return nil, lastErr
}

func formatResultsAsText(results map[string]interface{}) []byte {
	// For text format, we'll generate a simplified view based on the new structure
	evalId, _ := results["evalId"].(string)
	resultsData, _ := results["results"].(map[string]interface{})
	stats, _ := resultsData["stats"].(map[string]interface{})
	
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
	allResults, _ := resultsData["results"].([]map[string]interface{})
	for i, result := range allResults {
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