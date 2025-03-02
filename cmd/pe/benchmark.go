package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cobra"
	"github.com/tmc/pe/internal/cgpt"
	"sigs.k8s.io/yaml"
)

// BenchmarkResult holds the result data for a single prompt+provider benchmark
type BenchmarkResult struct {
	Prompt      string  `json:"prompt"`
	Provider    string  `json:"provider"`
	Iteration   int     `json:"iteration"`
	LatencyMs   float64 `json:"latencyMs"`
	TokensTotal int32   `json:"tokensTotal"`
	TokensInput int32   `json:"tokensInput"`
	TokensOutput int32  `json:"tokensOutput"`
	Cost        float64 `json:"cost"`
}

// BenchmarkSummary contains aggregate statistics for a benchmark run
type BenchmarkSummary struct {
	Prompt         string  `json:"prompt"`
	Provider       string  `json:"provider"`
	AvgLatencyMs   float64 `json:"avgLatencyMs"`
	MinLatencyMs   float64 `json:"minLatencyMs"`
	MaxLatencyMs   float64 `json:"maxLatencyMs"`
	P50LatencyMs   float64 `json:"p50LatencyMs"`
	P90LatencyMs   float64 `json:"p90LatencyMs"`
	P95LatencyMs   float64 `json:"p95LatencyMs"`
	P99LatencyMs   float64 `json:"p99LatencyMs"`
	AvgTokensTotal float64 `json:"avgTokensTotal"`
	AvgTokensInput float64 `json:"avgTokensInput"`
	AvgTokensOutput float64 `json:"avgTokensOutput"`
	TotalCost      float64 `json:"totalCost"`
}

// benchmarkCmd returns a cobra.Command for the 'benchmark' subcommand.
//
// benchmark compares performance metrics of multiple prompts and providers.
//
// Usage:
//
//	pe benchmark [config_file] [--iterations 10] [--output results.json] [--format json|yaml|csv|text]
func benchmarkCmd() *cobra.Command {
	var configFile string
	var outputFile string
	var outputFormat string
	var iterations int
	var concurrency int

	cmd := &cobra.Command{
		Use:   "benchmark [config_file]",
		Short: "Benchmark performance metrics of prompts and providers",
		Long:  `Compare performance metrics of multiple prompts and providers, measuring response time, token usage, and costs.`,
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

			return runBenchmark(cmd, configFile, outputFile, outputFormat, iterations, concurrency)
		},
	}

	cmd.Flags().StringVarP(&configFile, "config", "c", "", "Path to configuration file")
	cmd.Flags().StringVarP(&outputFile, "output", "o", "", "Write results to file")
	cmd.Flags().StringVarP(&outputFormat, "format", "f", "json", "Output format: 'json', 'yaml', 'csv', or 'text'")
	cmd.Flags().IntVarP(&iterations, "iterations", "i", 3, "Number of times to run each prompt")
	cmd.Flags().IntVarP(&concurrency, "concurrency", "n", 1, "Number of concurrent benchmark runs")

	return cmd
}

func runBenchmark(cmd *cobra.Command, configFile, outputFile, outputFormat string, iterations, concurrency int) error {
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

	// Extract prompts and providers
	prompts, _ := config["prompts"].([]interface{})
	providers, _ := config["providers"].([]interface{})
	tests, _ := config["tests"].([]interface{})

	if len(prompts) == 0 {
		return fmt.Errorf("no prompts found in configuration")
	}

	if len(providers) == 0 {
		return fmt.Errorf("no providers found in configuration")
	}

	// Set up concurrency control
	if concurrency < 1 {
		concurrency = 1
	}
	
	if concurrency > 10 {
		// Warn about high concurrency values which might trigger rate limits
		fmt.Fprintf(cmd.OutOrStderr(), "Warning: High concurrency (%d) might trigger provider rate limits\n", concurrency)
	}

	var semaphore = make(chan struct{}, concurrency)
	var wg sync.WaitGroup
	var mu sync.Mutex
	var allResults []BenchmarkResult

	// Run benchmarks
	fmt.Fprintf(cmd.OutOrStdout(), "Starting benchmark with %d prompts x %d providers x %d iterations...\n", 
		len(prompts), len(providers), iterations)

	startTime := time.Now()

	for i, prompt := range prompts {
		promptStr, ok := prompt.(string)
		if !ok {
			continue
		}

		for _, provider := range providers {
			providerStr, ok := provider.(string)
			if !ok {
				continue
			}

			// For each prompt and provider, run the specified number of iterations
			for iter := 1; iter <= iterations; iter++ {
				wg.Add(1)
				
				// Use closure to capture loop variables
				go func(promptIdx int, prompt string, providerName string, iteration int) {
					defer wg.Done()
					
					// Acquire semaphore slot (blocking if we've reached max concurrency)
					semaphore <- struct{}{}
					defer func() { <-semaphore }()

					// Create a model provider for this run
					modelProvider := cgpt.DefaultProvider()
					
					// If provider explicitly specified in format like "googleai:gemini-2.0-flash"
					if strings.Contains(providerName, ":") {
						parts := strings.Split(providerName, ":")
						modelProvider.Backend = parts[0]
						if len(parts) > 1 {
							modelProvider.Model = parts[1]
						}
					}

					// Initialize vars from test cases if available
					vars := make(map[string]interface{})
					if len(tests) > 0 {
						// Just use the first test's vars for benchmarking
						if testMap, ok := tests[0].(map[string]interface{}); ok {
							if testVars, ok := testMap["vars"].(map[string]interface{}); ok {
								for k, v := range testVars {
									vars[k] = v
								}
							}
						}
					}
					
					// Add provider information to vars
					vars["provider"] = providerName
					
					// Measure execution time 
					runStart := time.Now()
					
					// Execute the prompt
					response, err := modelProvider.EvaluatePrompt(prompt, vars)
					
					executionTimeMs := float64(time.Since(runStart).Milliseconds())
					
					// Record results
					result := BenchmarkResult{
						Prompt:      prompt,
						Provider:    providerName,
						Iteration:   iteration,
						LatencyMs:   executionTimeMs,
					}
					
					if err == nil && response != nil {
						if response.TokenUsage != nil {
							result.TokensTotal = response.TokenUsage.Total
							result.TokensInput = response.TokenUsage.Prompt
							result.TokensOutput = response.TokenUsage.Completion
						}
						result.Cost = response.Cost
					} else {
						fmt.Fprintf(cmd.OutOrStderr(), "Error with prompt %d, provider %s, iteration %d: %v\n", 
							promptIdx+1, providerName, iteration, err)
					}
					
					// Thread-safe append to results
					mu.Lock()
					allResults = append(allResults, result)
					mu.Unlock()
					
					// Print progress indicator
					fmt.Fprintf(cmd.OutOrStdout(), ".")
				}(i, promptStr, providerStr, iter)
			}
		}
	}

	// Wait for all benchmark runs to complete
	wg.Wait()
	
	// Calculate total time
	totalDuration := time.Since(startTime)
	fmt.Fprintf(cmd.OutOrStdout(), "\nBenchmark completed in %v\n", totalDuration)

	// Generate summaries
	summaries := generateBenchmarkSummaries(allResults)

	// Format and output results
	output, err := formatBenchmarkResults(allResults, summaries, outputFormat)
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

// formatBenchmarkResults formats the benchmark results according to the specified format
func formatBenchmarkResults(results []BenchmarkResult, summaries []BenchmarkSummary, format string) ([]byte, error) {
	switch format {
	case "json":
		// Create a structured output with both detailed results and summaries
		output := map[string]interface{}{
			"timestamp": time.Now().Format(time.RFC3339),
			"results":   results,
			"summaries": summaries,
		}
		return json.MarshalIndent(output, "", "  ")
		
	case "yaml":
		// Create a structured output with both detailed results and summaries
		output := map[string]interface{}{
			"timestamp": time.Now().Format(time.RFC3339),
			"results":   results,
			"summaries": summaries,
		}
		return yaml.Marshal(output)
		
	case "csv":
		// Create CSV output (summaries only for conciseness)
		var buf strings.Builder
		w := csv.NewWriter(&buf)
		
		// Write header
		header := []string{
			"Prompt", "Provider", "AvgLatencyMs", "MinLatencyMs", "MaxLatencyMs", 
			"P50LatencyMs", "P90LatencyMs", "P95LatencyMs", "P99LatencyMs", 
			"AvgTokensTotal", "AvgTokensInput", "AvgTokensOutput", "TotalCost",
		}
		if err := w.Write(header); err != nil {
			return nil, err
		}
		
		// Write data rows
		for _, s := range summaries {
			row := []string{
				s.Prompt,
				s.Provider,
				fmt.Sprintf("%.2f", s.AvgLatencyMs),
				fmt.Sprintf("%.2f", s.MinLatencyMs),
				fmt.Sprintf("%.2f", s.MaxLatencyMs),
				fmt.Sprintf("%.2f", s.P50LatencyMs),
				fmt.Sprintf("%.2f", s.P90LatencyMs),
				fmt.Sprintf("%.2f", s.P95LatencyMs),
				fmt.Sprintf("%.2f", s.P99LatencyMs),
				fmt.Sprintf("%.2f", s.AvgTokensTotal),
				fmt.Sprintf("%.2f", s.AvgTokensInput),
				fmt.Sprintf("%.2f", s.AvgTokensOutput),
				fmt.Sprintf("%.6f", s.TotalCost),
			}
			if err := w.Write(row); err != nil {
				return nil, err
			}
		}
		
		w.Flush()
		return []byte(buf.String()), nil
		
	case "text":
		// Human-readable text output
		var buf strings.Builder
		
		buf.WriteString("Benchmark Results Summary\n")
		buf.WriteString("========================\n\n")
		
		// Group by prompt
		promptGroups := make(map[string][]BenchmarkSummary)
		for _, s := range summaries {
			promptGroups[s.Prompt] = append(promptGroups[s.Prompt], s)
		}
		
		for prompt, group := range promptGroups {
			// Trim long prompts for display
			displayPrompt := prompt
			if len(displayPrompt) > 50 {
				displayPrompt = displayPrompt[:47] + "..."
			}
			
			buf.WriteString(fmt.Sprintf("Prompt: %s\n", displayPrompt))
			buf.WriteString(strings.Repeat("-", 60) + "\n")
			
			// Table header
			buf.WriteString(fmt.Sprintf("%-20s %-10s %-10s %-10s %-10s %-10s\n", 
				"Provider", "Avg Latency", "Min", "Max", "Tokens", "Cost"))
			buf.WriteString(strings.Repeat("-", 60) + "\n")
			
			// Sort providers for consistent output
			sort.Slice(group, func(i, j int) bool {
				return group[i].Provider < group[j].Provider
			})
			
			// Table rows
			for _, s := range group {
				buf.WriteString(fmt.Sprintf("%-20s %-10.2f %-10.2f %-10.2f %-10.2f $%-9.6f\n", 
					s.Provider, s.AvgLatencyMs, s.MinLatencyMs, s.MaxLatencyMs, 
					s.AvgTokensTotal, s.TotalCost))
			}
			
			buf.WriteString("\n\n")
		}
		
		return []byte(buf.String()), nil
		
	default:
		return nil, fmt.Errorf("unsupported output format: %s", format)
	}
}

// generateBenchmarkSummaries creates summary statistics from the benchmark results
func generateBenchmarkSummaries(results []BenchmarkResult) []BenchmarkSummary {
	// Group results by prompt and provider
	groups := make(map[string][]BenchmarkResult)
	for _, r := range results {
		key := r.Prompt + "|" + r.Provider
		groups[key] = append(groups[key], r)
	}
	
	var summaries []BenchmarkSummary
	
	for key, group := range groups {
		parts := strings.Split(key, "|")
		prompt := parts[0]
		provider := parts[1]
		
		// Extract latencies for percentile calculations
		latencies := make([]float64, len(group))
		for i, r := range group {
			latencies[i] = r.LatencyMs
		}
		sort.Float64s(latencies)
		
		// Calculate statistics
		var totalLatency, totalTokens, totalInputTokens, totalOutputTokens, totalCost float64
		var minLatency, maxLatency float64
		
		if len(group) > 0 {
			minLatency = group[0].LatencyMs
			maxLatency = group[0].LatencyMs
		}
		
		for _, r := range group {
			totalLatency += r.LatencyMs
			totalTokens += float64(r.TokensTotal)
			totalInputTokens += float64(r.TokensInput)
			totalOutputTokens += float64(r.TokensOutput)
			totalCost += r.Cost
			
			if r.LatencyMs < minLatency {
				minLatency = r.LatencyMs
			}
			if r.LatencyMs > maxLatency {
				maxLatency = r.LatencyMs
			}
		}
		
		// Calculate averages
		count := float64(len(group))
		avgLatency := totalLatency / count
		avgTokens := totalTokens / count
		avgInputTokens := totalInputTokens / count
		avgOutputTokens := totalOutputTokens / count
		
		// Calculate percentiles
		p50 := percentile(latencies, 50)
		p90 := percentile(latencies, 90)
		p95 := percentile(latencies, 95)
		p99 := percentile(latencies, 99)
		
		// Create summary
		summary := BenchmarkSummary{
			Prompt:         prompt,
			Provider:       provider,
			AvgLatencyMs:   avgLatency,
			MinLatencyMs:   minLatency,
			MaxLatencyMs:   maxLatency,
			P50LatencyMs:   p50,
			P90LatencyMs:   p90,
			P95LatencyMs:   p95,
			P99LatencyMs:   p99,
			AvgTokensTotal: avgTokens,
			AvgTokensInput: avgInputTokens,
			AvgTokensOutput: avgOutputTokens,
			TotalCost:      totalCost,
		}
		
		summaries = append(summaries, summary)
	}
	
	return summaries
}

// percentile calculates the specified percentile from sorted data
func percentile(sortedData []float64, p float64) float64 {
	if len(sortedData) == 0 {
		return 0
	}
	
	if len(sortedData) == 1 {
		return sortedData[0]
	}
	
	// Calculate the position
	position := (p / 100.0) * float64(len(sortedData)-1)
	
	// Get the integer and fractional parts
	positionInt, positionFrac := int(position), position-float64(int(position))
	
	// If it's an exact position
	if positionFrac == 0 {
		return sortedData[positionInt]
	}
	
	// Interpolate between the two nearest values
	lower := sortedData[positionInt]
	upper := sortedData[positionInt+1]
	return lower + (upper-lower)*positionFrac
}