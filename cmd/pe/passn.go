package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/tmc/pe/internal/llm"
	"github.com/tmc/pe/internal/metrics"
	"github.com/tmc/pe/internal/structured"
)

var passnCmd = &cobra.Command{
	Use:   "passn",
	Short: "Evaluate and store pass@n metrics for prompts",
	Long: `Evaluate pass@n metrics for code generation and other tasks.
	
Pass@n measures how often a model generates a correct solution within n attempts.
This is particularly useful for evaluating code generation, problem-solving, and
other tasks where multiple attempts might yield different results.`,
	RunE: runPassN,
}

var (
	passnN               int
	passnSamples         int
	passnProvider        string
	passnTemperature     float64
	passnTestFile        string
	passnOutputFormat    string
	passnSaveResults     bool
	passnDataDir         string
	passnTags            []string
	passnStrategy        string
	passnStructuredType  string
	passnSchema          string
)

func init() {
	passnCmd.Flags().IntVarP(&passnN, "n", "n", 1, "Number of attempts to consider for pass@n")
	passnCmd.Flags().IntVarP(&passnSamples, "samples", "s", 10, "Number of samples to generate")
	passnCmd.Flags().StringVarP(&passnProvider, "provider", "p", "openai", "LLM provider to use")
	passnCmd.Flags().Float64VarP(&passnTemperature, "temperature", "t", 0.8, "Sampling temperature")
	passnCmd.Flags().StringVarP(&passnTestFile, "tests", "f", "", "JSON file containing test cases")
	passnCmd.Flags().StringVarP(&passnOutputFormat, "output", "o", "json", "Output format (json, table, report)")
	passnCmd.Flags().BoolVarP(&passnSaveResults, "save", "", true, "Save results to storage")
	passnCmd.Flags().StringVarP(&passnDataDir, "data-dir", "d", "~/.pe/passn", "Directory for storing pass@n data")
	passnCmd.Flags().StringSliceVarP(&passnTags, "tags", "", []string{}, "Tags for categorizing the evaluation")
	passnCmd.Flags().StringVarP(&passnStrategy, "strategy", "", "random", "Sampling strategy (random, diverse, adaptive)")
	passnCmd.Flags().StringVarP(&passnStructuredType, "structured", "", "", "Use structured output (json, yaml, code, etc.)")
	passnCmd.Flags().StringVarP(&passnSchema, "schema", "", "", "Path to schema file for structured output")

	// Add subcommands
	passnCmd.AddCommand(passnReportCmd)
	passnCmd.AddCommand(passnSearchCmd)
	passnCmd.AddCommand(passnCompareCmd)
}

func runPassN(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("please provide a prompt or prompt file")
	}

	// Read prompt
	prompt := args[0]
	if _, err := os.Stat(prompt); err == nil {
		// It's a file
		data, err := os.ReadFile(prompt)
		if err != nil {
			return fmt.Errorf("failed to read prompt file: %w", err)
		}
		prompt = string(data)
	}

	// Get provider
	provider, err := llm.GetProvider(passnProvider)
	if err != nil {
		return fmt.Errorf("failed to get provider: %w", err)
	}

	// Setup structured output if requested
	if passnStructuredType != "" {
		prompt, err = setupStructuredPrompt(prompt)
		if err != nil {
			return fmt.Errorf("failed to setup structured prompt: %w", err)
		}
	}

	// Create pass@n generator
	config := metrics.PassNConfig{
		Temperature:      passnTemperature,
		TopP:             0.95,
		SamplingStrategy: passnStrategy,
	}
	generator := metrics.NewPassNGenerator(provider, config)

	// Generate samples
	ctx := context.Background()
	fmt.Printf("Generating %d samples for pass@%d evaluation...\n", passnSamples, passnN)
	
	samples, err := generator.GenerateSamples(ctx, prompt, passnSamples)
	if err != nil {
		return fmt.Errorf("failed to generate samples: %w", err)
	}

	// Load test cases if provided
	var testCases []map[string]interface{}
	if passnTestFile != "" {
		testCases, err = loadTestCases(passnTestFile)
		if err != nil {
			return fmt.Errorf("failed to load test cases: %w", err)
		}
	}

	// Calculate pass@n
	am := metrics.NewAdvancedMetrics(provider)
	var result *metrics.PassAtNResult

	if len(testCases) > 0 {
		result = am.CalculatePassAtNWithTests(ctx, passnN, samples, testCases)
	} else {
		// Default test function - checks if output is non-empty
		testFunc := func(output string) bool {
			return len(output) > 0 && !containsError(output)
		}
		result = am.CalculatePassAtN(passnN, samples, testFunc)
	}

	// Create evaluation record
	evaluation := metrics.PassNEvaluation{
		ID:         fmt.Sprintf("eval-%d", time.Now().Unix()),
		Timestamp:  time.Now(),
		N:          passnN,
		NumSamples: len(samples),
		NumPassed:  result.NumPassed,
		PassRate:   result.PassRate,
		Duration:   result.Duration,
	}

	// Add sample details
	for i, sample := range samples {
		passed := false
		if indices, ok := result.Details["passed_indices"].([]int); ok {
			for _, idx := range indices {
				if idx == i {
					passed = true
					break
				}
			}
		}

		evaluation.Samples = append(evaluation.Samples, metrics.PassNSample{
			Index:   i,
			Content: sample,
			Passed:  passed,
		})
	}

	// Save results if requested
	if passnSaveResults {
		dataDir := expandPath(passnDataDir)
		store, err := metrics.NewPassNStore(dataDir)
		if err != nil {
			return fmt.Errorf("failed to create store: %w", err)
		}

		promptID := generatePromptID(prompt)
		err = store.StoreEvaluation(promptID, prompt, passnProvider, evaluation)
		if err != nil {
			return fmt.Errorf("failed to store evaluation: %w", err)
		}
		fmt.Printf("Results saved with ID: %s\n", promptID)
	}

	// Output results
	return outputPassNResults(result, evaluation)
}

var passnReportCmd = &cobra.Command{
	Use:   "report [prompt-id]",
	Short: "Generate a report for pass@n evaluations",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dataDir := expandPath(passnDataDir)
		store, err := metrics.NewPassNStore(dataDir)
		if err != nil {
			return fmt.Errorf("failed to create store: %w", err)
		}

		report, err := store.GeneratePassNReport(args[0])
		if err != nil {
			return fmt.Errorf("failed to generate report: %w", err)
		}

		// Output report
		data, err := json.MarshalIndent(report, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
		return nil
	},
}

var passnSearchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search for pass@n evaluations by tags",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(passnTags) == 0 {
			return fmt.Errorf("please provide tags to search for")
		}

		dataDir := expandPath(passnDataDir)
		store, err := metrics.NewPassNStore(dataDir)
		if err != nil {
			return fmt.Errorf("failed to create store: %w", err)
		}

		results, err := store.SearchByTags(passnTags)
		if err != nil {
			return fmt.Errorf("search failed: %w", err)
		}

		// Output results
		for _, result := range results {
			fmt.Printf("ID: %s, Best Pass Rate: %.2f%%, Provider: %s\n",
				result.PromptID, result.BestPassRate*100, result.ProviderID)
		}
		return nil
	},
}

var passnCompareCmd = &cobra.Command{
	Use:   "compare [prompt-id1] [prompt-id2]",
	Short: "Compare pass@n results between two prompts",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		dataDir := expandPath(passnDataDir)
		store, err := metrics.NewPassNStore(dataDir)
		if err != nil {
			return fmt.Errorf("failed to create store: %w", err)
		}

		// Get both evaluations
		eval1, err := store.GetEvaluations(args[0])
		if err != nil {
			return fmt.Errorf("failed to get first evaluation: %w", err)
		}

		eval2, err := store.GetEvaluations(args[1])
		if err != nil {
			return fmt.Errorf("failed to get second evaluation: %w", err)
		}

		// Compare and output
		fmt.Printf("Comparison of %s vs %s\n\n", args[0], args[1])
		fmt.Printf("Prompt 1 - Best Pass Rate: %.2f%%, Avg: %.2f%%, Total Samples: %d\n",
			eval1.BestPassRate*100, eval1.AvgPassRate*100, eval1.TotalSamples)
		fmt.Printf("Prompt 2 - Best Pass Rate: %.2f%%, Avg: %.2f%%, Total Samples: %d\n",
			eval2.BestPassRate*100, eval2.AvgPassRate*100, eval2.TotalSamples)

		improvement := (eval2.BestPassRate - eval1.BestPassRate) / eval1.BestPassRate * 100
		fmt.Printf("\nImprovement: %.2f%%\n", improvement)

		return nil
	},
}

// Helper functions

func setupStructuredPrompt(basePrompt string) (string, error) {
	builder := structured.NewPromptBuilder()

	// Load schema if provided
	var schema *structured.Schema
	if passnSchema != "" {
		data, err := os.ReadFile(passnSchema)
		if err != nil {
			return "", fmt.Errorf("failed to read schema file: %w", err)
		}

		schema = &structured.Schema{}
		if err := json.Unmarshal(data, schema); err != nil {
			return "", fmt.Errorf("failed to parse schema: %w", err)
		}
	} else {
		// Use a default schema based on type
		switch passnStructuredType {
		case "code":
			schema = structured.CommonSchemas.CodeGeneration
		case "analysis":
			schema = structured.CommonSchemas.Analysis
		case "data":
			schema = structured.CommonSchemas.DataExtraction
		default:
			// Basic JSON schema
			schema = &structured.Schema{
				Name: "output",
				Type: "object",
				Properties: map[string]*structured.Property{
					"result": {
						Type:        "string",
						Description: "The generated result",
					},
				},
				Required: []string{"result"},
			}
		}
	}

	// Build structured prompt
	format := structured.OutputFormat(passnStructuredType)
	if format == "" {
		format = structured.FormatJSON
	}

	return builder.BuildPrompt(basePrompt, schema, format)
}

func loadTestCases(filename string) ([]map[string]interface{}, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var testCases []map[string]interface{}
	if err := json.Unmarshal(data, &testCases); err != nil {
		return nil, err
	}

	return testCases, nil
}

func outputPassNResults(result *metrics.PassAtNResult, evaluation metrics.PassNEvaluation) error {
	switch passnOutputFormat {
	case "json":
		data, err := json.MarshalIndent(map[string]interface{}{
			"pass_rate": result.PassRate,
			"n":         result.N,
			"samples":   result.NumSamples,
			"passed":    result.NumPassed,
			"duration":  result.Duration.String(),
			"details":   evaluation,
		}, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))

	case "table":
		fmt.Printf("Pass@%d Evaluation Results\n", result.N)
		fmt.Printf("========================\n")
		fmt.Printf("Pass Rate: %.2f%%\n", result.PassRate*100)
		fmt.Printf("Samples: %d\n", result.NumSamples)
		fmt.Printf("Passed: %d\n", result.NumPassed)
		fmt.Printf("Duration: %s\n", result.Duration)

	case "report":
		fmt.Printf("# Pass@%d Evaluation Report\n\n", result.N)
		fmt.Printf("## Summary\n")
		fmt.Printf("- **Pass Rate**: %.2f%%\n", result.PassRate*100)
		fmt.Printf("- **Total Samples**: %d\n", result.NumSamples)
		fmt.Printf("- **Successful Samples**: %d\n", result.NumPassed)
		fmt.Printf("- **Evaluation Time**: %s\n\n", result.Duration)

		fmt.Printf("## Sample Results\n")
		for i, sample := range evaluation.Samples[:min(5, len(evaluation.Samples))] {
			status := "❌"
			if sample.Passed {
				status = "✅"
			}
			fmt.Printf("\n### Sample %d %s\n", i+1, status)
			fmt.Printf("```\n%s\n```\n", truncateString(sample.Content, 200))
		}

	default:
		return fmt.Errorf("unsupported output format: %s", passnOutputFormat)
	}

	return nil
}

func containsError(output string) bool {
	errorIndicators := []string{"error", "exception", "failed", "invalid", "undefined"}
	lowerOutput := strings.ToLower(output)
	
	for _, indicator := range errorIndicators {
		if strings.Contains(lowerOutput, indicator) {
			return true
		}
	}
	return false
}

func generatePromptID(prompt string) string {
	// Simple ID generation - in production, use proper hashing
	return fmt.Sprintf("prompt-%d", time.Now().Unix())
}

func expandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, _ := os.UserHomeDir()
		path = filepath.Join(home, path[2:])
	}
	return path
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

