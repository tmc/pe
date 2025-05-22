package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"

	"github.com/tmc/pe/internal/llm"
	"github.com/tmc/pe/internal/testing"
)

// testCmd returns a cobra.Command for the 'test' subcommand.
//
// test runs advanced testing including property-based and regression tests.
//
// Usage:
//
//	pe test config.yaml --type property
//	pe test config.yaml --type regression --baseline baseline.json
func testCmd() *cobra.Command {
	var testType string
	var baseline string
	var saveBaseline string
	var iterations int
	var tolerance float64
	var outputFile string
	var verbose bool

	cmd := &cobra.Command{
		Use:   "test [config_file]",
		Short: "Run advanced testing (property-based, regression)",
		Long: `Test runs advanced testing scenarios including property-based testing
for robustness validation and regression testing for performance monitoring.

Property-based testing generates random inputs and validates that certain
properties hold across all inputs, helping discover edge cases and ensure
consistent behavior.

Regression testing compares current results against a baseline to detect
performance regressions or improvements in metrics like latency, cost,
and quality scores.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			configFile := "pe-config.yaml"
			if len(args) > 0 {
				configFile = args[0]
			}

			return runAdvancedTest(cmd, configFile, testType, baseline, saveBaseline, 
				iterations, tolerance, outputFile, verbose)
		},
	}

	cmd.Flags().StringVarP(&testType, "type", "t", "property", "Test type: property, regression, or both")
	cmd.Flags().StringVarP(&baseline, "baseline", "b", "", "Baseline file for regression testing")
	cmd.Flags().StringVarP(&saveBaseline, "save-baseline", "", "", "Save current results as baseline")
	cmd.Flags().IntVarP(&iterations, "iterations", "i", 50, "Number of iterations for property testing")
	cmd.Flags().Float64VarP(&tolerance, "tolerance", "", 5.0, "Tolerance percentage for regression detection")
	cmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file for test results")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")

	return cmd
}

// AdvancedTestConfig represents the configuration for advanced testing
type AdvancedTestConfig struct {
	PropertyTests   []testing.PropertyTest   `yaml:"property_tests"`
	RegressionTests []testing.RegressionTest `yaml:"regression_tests"`
	Providers       []string                 `yaml:"providers"`
	Prompts         []string                 `yaml:"prompts"`
	Options         map[string]interface{}   `yaml:"options"`
}

// TestResults contains all test results
type TestResults struct {
	Timestamp       time.Time                        `json:"timestamp"`
	Config          string                           `json:"config"`
	TestType        string                           `json:"test_type"`
	PropertyResults []testing.PropertyTestResult     `json:"property_results,omitempty"`
	RegressionResults []testing.RegressionResult     `json:"regression_results,omitempty"`
	Summary         TestSummary                      `json:"summary"`
	Duration        time.Duration                    `json:"duration"`
}

// TestSummary provides a high-level summary of test results
type TestSummary struct {
	TotalTests       int     `json:"total_tests"`
	PassedTests      int     `json:"passed_tests"`
	FailedTests      int     `json:"failed_tests"`
	SuccessRate      float64 `json:"success_rate"`
	TotalRegressions int     `json:"total_regressions"`
	TotalImprovements int    `json:"total_improvements"`
}

func runAdvancedTest(cmd *cobra.Command, configFile, testType, baseline, saveBaseline string, 
	iterations int, tolerance float64, outputFile string, verbose bool) error {
	
	startTime := time.Now()
	
	// Load test configuration
	config, err := loadTestConfig(configFile)
	if err != nil {
		return fmt.Errorf("failed to load config: %v", err)
	}
	
	results := &TestResults{
		Timestamp: startTime,
		Config:    configFile,
		TestType:  testType,
		Summary:   TestSummary{},
	}
	
	ctx := context.Background()
	
	// Run property-based tests
	if testType == "property" || testType == "both" {
		if verbose {
			fmt.Fprintf(cmd.OutOrStdout(), "Running property-based tests...\n")
		}
		
		propertyResults, err := runPropertyTests(ctx, config, iterations, verbose, cmd)
		if err != nil {
			return fmt.Errorf("property tests failed: %v", err)
		}
		
		results.PropertyResults = propertyResults
		for _, result := range propertyResults {
			results.Summary.TotalTests++
			if result.Passed {
				results.Summary.PassedTests++
			} else {
				results.Summary.FailedTests++
			}
		}
	}
	
	// Run regression tests
	if testType == "regression" || testType == "both" {
		if baseline == "" && testType == "regression" {
			return fmt.Errorf("baseline file required for regression testing")
		}
		
		if verbose && baseline != "" {
			fmt.Fprintf(cmd.OutOrStdout(), "Running regression tests...\n")
		}
		
		regressionResults, err := runRegressionTests(ctx, config, baseline, tolerance, verbose, cmd)
		if err != nil {
			return fmt.Errorf("regression tests failed: %v", err)
		}
		
		results.RegressionResults = regressionResults
		for _, result := range regressionResults {
			results.Summary.TotalTests++
			if result.Passed {
				results.Summary.PassedTests++
			} else {
				results.Summary.FailedTests++
			}
			results.Summary.TotalRegressions += len(result.Regressions)
			results.Summary.TotalImprovements += len(result.Improvements)
		}
	}
	
	// Calculate summary
	if results.Summary.TotalTests > 0 {
		results.Summary.SuccessRate = float64(results.Summary.PassedTests) / float64(results.Summary.TotalTests) * 100
	}
	results.Duration = time.Since(startTime)
	
	// Save baseline if requested
	if saveBaseline != "" {
		if err := saveCurrentAsBaseline(config, saveBaseline, cmd); err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "Warning: failed to save baseline: %v\n", err)
		} else if verbose {
			fmt.Fprintf(cmd.OutOrStdout(), "Baseline saved to: %s\n", saveBaseline)
		}
	}
	
	// Output results
	if err := outputTestResults(results, outputFile, cmd); err != nil {
		return fmt.Errorf("failed to output results: %v", err)
	}
	
	// Print summary
	printTestSummary(results, cmd)
	
	// Exit with error code if tests failed
	if results.Summary.FailedTests > 0 {
		os.Exit(1)
	}
	
	return nil
}

func loadTestConfig(filename string) (*AdvancedTestConfig, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	
	var config AdvancedTestConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}
	
	// Set defaults
	if len(config.Providers) == 0 {
		config.Providers = []string{"openai:gpt-3.5-turbo"}
	}
	
	return &config, nil
}

func runPropertyTests(ctx context.Context, config *AdvancedTestConfig, iterations int, verbose bool, cmd *cobra.Command) ([]testing.PropertyTestResult, error) {
	var results []testing.PropertyTestResult
	
	for _, providerSpec := range config.Providers {
		provider, err := llm.GetProvider(providerSpec)
		if err != nil {
			return nil, fmt.Errorf("failed to create provider %s: %v", providerSpec, err)
		}
		
		options := llm.GenerateOptions{}
		if temp, ok := config.Options["temperature"].(float64); ok {
			options.Temperature = &temp
		}
		
		tester := testing.NewPropertyTester(provider, options)
		
		for _, test := range config.PropertyTests {
			if verbose {
				fmt.Fprintf(cmd.OutOrStdout(), "  Running property test: %s with %s\n", test.Name, providerSpec)
			}
			
			test.Iterations = iterations // Override with command line value
			result, err := tester.RunPropertyTest(ctx, test)
			if err != nil {
				return nil, fmt.Errorf("property test %s failed: %v", test.Name, err)
			}
			
			results = append(results, *result)
			
			if verbose {
				if result.Passed {
					fmt.Fprintf(cmd.OutOrStdout(), "    ✓ PASSED (%d/%d iterations)\n", 
						result.Iterations-len(result.Failures), result.Iterations)
				} else {
					fmt.Fprintf(cmd.OutOrStdout(), "    ✗ FAILED (%d failures)\n", len(result.Failures))
				}
			}
		}
	}
	
	return results, nil
}

func runRegressionTests(ctx context.Context, config *AdvancedTestConfig, baseline string, tolerance float64, verbose bool, cmd *cobra.Command) ([]testing.RegressionResult, error) {
	var results []testing.RegressionResult
	
	for _, providerSpec := range config.Providers {
		provider, err := llm.GetProvider(providerSpec)
		if err != nil {
			return nil, fmt.Errorf("failed to create provider %s: %v", providerSpec, err)
		}
		
		options := llm.GenerateOptions{}
		if temp, ok := config.Options["temperature"].(float64); ok {
			options.Temperature = &temp
		}
		
		tester := testing.NewRegressionTester(provider, options)
		
		// Generate current results by running evaluation
		currentResults, err := generateCurrentResults(ctx, provider, config, options)
		if err != nil {
			return nil, fmt.Errorf("failed to generate current results: %v", err)
		}
		
		for _, test := range config.RegressionTests {
			if verbose {
				fmt.Fprintf(cmd.OutOrStdout(), "  Running regression test: %s\n", test.Name)
			}
			
			test.Tolerance = tolerance // Override with command line value
			if test.Baseline == "" {
				test.Baseline = baseline
			}
			
			result, err := tester.RunRegressionTest(ctx, test, currentResults)
			if err != nil {
				return nil, fmt.Errorf("regression test %s failed: %v", test.Name, err)
			}
			
			results = append(results, *result)
			
			if verbose {
				if result.Passed {
					fmt.Fprintf(cmd.OutOrStdout(), "    ✓ PASSED (no regressions detected)\n")
				} else {
					fmt.Fprintf(cmd.OutOrStdout(), "    ✗ FAILED (%d regressions detected)\n", len(result.Regressions))
				}
			}
		}
	}
	
	return results, nil
}

func generateCurrentResults(ctx context.Context, provider llm.Provider, config *AdvancedTestConfig, options llm.GenerateOptions) (testing.BaselineData, error) {
	testCases := make(map[string]testing.TestCaseResult)
	
	for i, prompt := range config.Prompts {
		testID := fmt.Sprintf("test_%d", i)
		
		startTime := time.Now()
		response, err := provider.Generate(ctx, prompt, options)
		latency := time.Since(startTime)
		
		success := err == nil
		score := 1.0
		if !success {
			score = 0.0
		}
		
		testCase := testing.TestCaseResult{
			Prompt:   prompt,
			Response: "",
			Latency:  latency,
			Tokens: testing.TokenMetrics{
				Prompt:     0,
				Completion: 0,
				Total:      0,
			},
			Cost:    0.0,
			Score:   score,
			Success: success,
		}
		
		if response != nil {
			testCase.Response = response.Text
			testCase.Tokens.Prompt = response.PromptTokens
			testCase.Tokens.Completion = response.CompletionTokens
			testCase.Tokens.Total = response.TotalTokens
			testCase.Cost = response.Cost
		}
		
		testCases[testID] = testCase
	}
	
	return testing.CreateBaselineFromEvaluation(provider.Name(), provider.Model(), testCases), nil
}

func saveCurrentAsBaseline(config *AdvancedTestConfig, filename string, cmd *cobra.Command) error {
	ctx := context.Background()
	
	// Use first provider for baseline generation
	if len(config.Providers) == 0 {
		return fmt.Errorf("no providers configured")
	}
	
	provider, err := llm.GetProvider(config.Providers[0])
	if err != nil {
		return err
	}
	
	options := llm.GenerateOptions{}
	if temp, ok := config.Options["temperature"].(float64); ok {
		options.Temperature = &temp
	}
	
	baseline, err := generateCurrentResults(ctx, provider, config, options)
	if err != nil {
		return err
	}
	
	tester := testing.NewRegressionTester(provider, options)
	return tester.SaveBaseline(filename, baseline)
}

func outputTestResults(results *TestResults, outputFile string, cmd *cobra.Command) error {
	jsonData, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return err
	}
	
	if outputFile != "" {
		return os.WriteFile(outputFile, jsonData, 0644)
	}
	
	fmt.Fprintln(cmd.OutOrStdout(), string(jsonData))
	return nil
}

func printTestSummary(results *TestResults, cmd *cobra.Command) {
	fmt.Fprintf(cmd.OutOrStdout(), "\n")
	fmt.Fprintf(cmd.OutOrStdout(), "=== Test Summary ===\n")
	fmt.Fprintf(cmd.OutOrStdout(), "Total Tests:     %d\n", results.Summary.TotalTests)
	fmt.Fprintf(cmd.OutOrStdout(), "Passed:          %d\n", results.Summary.PassedTests)
	fmt.Fprintf(cmd.OutOrStdout(), "Failed:          %d\n", results.Summary.FailedTests)
	fmt.Fprintf(cmd.OutOrStdout(), "Success Rate:    %.1f%%\n", results.Summary.SuccessRate)
	
	if results.Summary.TotalRegressions > 0 {
		fmt.Fprintf(cmd.OutOrStdout(), "Regressions:     %d\n", results.Summary.TotalRegressions)
	}
	
	if results.Summary.TotalImprovements > 0 {
		fmt.Fprintf(cmd.OutOrStdout(), "Improvements:    %d\n", results.Summary.TotalImprovements)
	}
	
	fmt.Fprintf(cmd.OutOrStdout(), "Duration:        %v\n", results.Duration)
	
	if results.Summary.FailedTests > 0 {
		fmt.Fprintf(cmd.OutOrStdout(), "\n⚠️  Some tests failed. Check the detailed results above.\n")
	} else {
		fmt.Fprintf(cmd.OutOrStdout(), "\n✅ All tests passed!\n")
	}
}