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
	"github.com/tmc/pe/internal/promptfoo"
	"github.com/tmc/pe/internal/providers"
	"sigs.k8s.io/yaml"
)

// BenchmarkResult holds the result data for a single prompt+provider benchmark
type BenchmarkResult struct {
	Prompt         string                 `json:"prompt"`
	Provider       string                 `json:"provider"`
	ProviderID     string                 `json:"providerId,omitempty"`
	Iteration      int                    `json:"iteration"`
	LatencyMs      float64                `json:"latencyMs"`
	TokensTotal    int32                  `json:"tokensTotal"`
	TokensInput    int32                  `json:"tokensInput"`
	TokensOutput   int32                  `json:"tokensOutput"`
	Cost           float64                `json:"cost"`
	RuntimeMetrics map[string]interface{} `json:"runtimeMetrics,omitempty"`
}

// BenchmarkSummary contains aggregate statistics for a benchmark run
type BenchmarkSummary struct {
	Prompt          string  `json:"prompt"`
	Provider        string  `json:"provider"`
	ProviderID      string  `json:"providerId,omitempty"`
	AvgLatencyMs    float64 `json:"avgLatencyMs"`
	MinLatencyMs    float64 `json:"minLatencyMs"`
	MaxLatencyMs    float64 `json:"maxLatencyMs"`
	P50LatencyMs    float64 `json:"p50LatencyMs"`
	P90LatencyMs    float64 `json:"p90LatencyMs"`
	P95LatencyMs    float64 `json:"p95LatencyMs"`
	P99LatencyMs    float64 `json:"p99LatencyMs"`
	AvgTokensTotal  float64 `json:"avgTokensTotal"`
	AvgTokensInput  float64 `json:"avgTokensInput"`
	AvgTokensOutput float64 `json:"avgTokensOutput"`
	TotalCost       float64 `json:"totalCost"`
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
	var goBenchFormat bool

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

			return runBenchmark(cmd, configFile, outputFile, outputFormat, iterations, concurrency, goBenchFormat)
		},
	}

	cmd.Flags().StringVarP(&configFile, "config", "c", "", "Path to configuration file")
	cmd.Flags().StringVarP(&outputFile, "output", "o", "", "Write results to file")
	cmd.Flags().StringVarP(&outputFormat, "format", "f", "json", "Output format: 'json', 'yaml', 'csv', or 'text'")
	cmd.Flags().IntVarP(&iterations, "iterations", "i", 3, "Number of times to run each prompt")
	cmd.Flags().IntVarP(&concurrency, "concurrency", "n", 1, "Number of concurrent benchmark runs")
	cmd.Flags().BoolVar(&goBenchFormat, "go-bench", false, "Output in Go benchmark format (compatible with golang.org/x/perf tools)")

	return cmd
}

func runBenchmark(cmd *cobra.Command, configFile, outputFile, outputFormat string, iterations, concurrency int, goBenchFormat bool) error {
	// Read config file
	data, err := os.ReadFile(configFile)
	if err != nil {
		return fmt.Errorf("error reading config file: %v", err)
	}

	// Parse the config using the same provider model as eval.
	var config promptfoo.Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return fmt.Errorf("error parsing config file: %v", err)
	}

	if len(config.Prompts) == 0 {
		return fmt.Errorf("no prompts found in configuration")
	}

	if len(config.Providers) == 0 {
		return fmt.Errorf("no providers found in configuration")
	}

	materializedProviders, err := providers.MaterializeProviders(config.Providers)
	if err != nil {
		return err
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
		len(config.Prompts), len(materializedProviders), iterations)

	startTime := time.Now()

	for i, promptStr := range config.Prompts {
		for _, provider := range materializedProviders {
			if !provider.AppliesToPrompt(promptStr) {
				continue
			}
			// For each prompt and provider, run the specified number of iterations
			for iter := 1; iter <= iterations; iter++ {
				wg.Add(1)

				// Use closure to capture loop variables
				go func(promptIdx int, prompt string, provider *providers.MaterializedProvider, iteration int) {
					defer wg.Done()

					// Acquire semaphore slot (blocking if we've reached max concurrency)
					semaphore <- struct{}{}
					defer func() { <-semaphore }()

					if provider.Delay > 0 {
						timer := time.NewTimer(provider.Delay)
						defer timer.Stop()
						select {
						case <-cmd.Context().Done():
							return
						case <-timer.C:
						}
					}

					// Initialize vars from provider defaults first, then the first test case if available.
					vars := make(map[string]interface{})
					for k, v := range provider.Spec.Config {
						vars[k] = v
					}
					if len(config.Tests) > 0 {
						for k, v := range config.Tests[0].Vars {
							vars[k] = v
						}
					}
					vars["provider"] = provider.Spec.ID

					processedPrompt := promptfoo.ApplyVars(prompt, vars)

					// Measure execution time
					runStart := time.Now()

					// Execute the prompt
					response, err := provider.Executor.EvaluatePrompt(cmd.Context(), processedPrompt, vars)

					executionTimeMs := float64(time.Since(runStart).Milliseconds())

					// Record results
					result := BenchmarkResult{
						Prompt:     prompt,
						Provider:   provider.DisplayName(),
						ProviderID: provider.Spec.ID,
						Iteration:  iteration,
						LatencyMs:  executionTimeMs,
					}

					if err == nil && response != nil {
						if response.LatencyMs > 0 {
							result.LatencyMs = float64(response.LatencyMs)
						}
						if response.TokenUsage != nil {
							result.TokensTotal = response.TokenUsage.Total
							result.TokensInput = response.TokenUsage.Prompt
							result.TokensOutput = response.TokenUsage.Completion
						}
						result.Cost = response.Cost
						result.RuntimeMetrics = response.Metadata
					} else {
						fmt.Fprintf(cmd.OutOrStderr(), "Error with prompt %d, provider %s, iteration %d: %v\n",
							promptIdx+1, provider.Spec.ID, iteration, err)
					}

					// Thread-safe append to results
					mu.Lock()
					allResults = append(allResults, result)
					mu.Unlock()

					// Print progress indicator
					fmt.Fprintf(cmd.OutOrStdout(), ".")
				}(i, promptStr, provider, iter)
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

	// Output in Go benchmark format if requested
	if goBenchFormat {
		goBenchOutput := formatAsGoBenchmarks(summaries)
		if goBenchOutput != "" {
			fmt.Fprint(cmd.OutOrStdout(), goBenchOutput)
		}
		return nil // Skip regular output when using Go benchmark format
	}

	// Write to output file or stdout
	if outputFile != "" {
		if err := enforceRuntimeToolPolicy("write"); err != nil {
			return err
		}
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
		key := r.Prompt + "|" + r.ProviderID + "|" + r.Provider
		groups[key] = append(groups[key], r)
	}

	var summaries []BenchmarkSummary

	for key, group := range groups {
		parts := strings.SplitN(key, "|", 3)
		prompt := parts[0]
		providerID := parts[1]
		provider := parts[2]

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
			Prompt:          prompt,
			Provider:        provider,
			ProviderID:      providerID,
			AvgLatencyMs:    avgLatency,
			MinLatencyMs:    minLatency,
			MaxLatencyMs:    maxLatency,
			P50LatencyMs:    p50,
			P90LatencyMs:    p90,
			P95LatencyMs:    p95,
			P99LatencyMs:    p99,
			AvgTokensTotal:  avgTokens,
			AvgTokensInput:  avgInputTokens,
			AvgTokensOutput: avgOutputTokens,
			TotalCost:       totalCost,
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

// formatAsGoBenchmarks converts PE benchmark summaries to Go benchmark format
// Compatible with golang.org/x/perf/cmd/benchstat and other Go perf tools
func formatAsGoBenchmarks(summaries []BenchmarkSummary) string {
	var buf strings.Builder

	for _, s := range summaries {
		// Create benchmark name following Go conventions
		// Replace spaces and special chars with valid identifier chars
		promptName := strings.ReplaceAll(s.Prompt, " ", "")
		promptName = strings.ReplaceAll(promptName, "?", "")
		promptName = strings.ReplaceAll(promptName, "!", "")
		promptName = strings.ReplaceAll(promptName, ".", "")
		promptName = strings.ReplaceAll(promptName, ",", "")
		if promptName == "" {
			promptName = "Prompt"
		}

		providerName := strings.ReplaceAll(s.Provider, ":", "_")
		providerName = strings.ReplaceAll(providerName, "-", "_")
		providerName = strings.ReplaceAll(providerName, " ", "_")

		benchmarkName := fmt.Sprintf("Benchmark%s_%s", promptName, providerName)

		// Calculate iterations (assume 1 for now, could be made configurable)
		iterations := 1

		// Convert latency from ms to ns for Go benchmark format
		latencyNs := s.AvgLatencyMs * 1_000_000

		// Calculate tokens per second (throughput)
		tokensPerSecond := (s.AvgTokensOutput / s.AvgLatencyMs) * 1000

		// Format: BenchmarkName iterations ns/op [other metrics]
		buf.WriteString(fmt.Sprintf("%s\t%d\t%.0f ns/op\t%.2f tokens/s\t%.0f tokens/op\t$%.6f/op\n",
			benchmarkName,
			iterations,
			latencyNs,
			tokensPerSecond,
			s.AvgTokensTotal,
			s.TotalCost,
		))
	}

	return buf.String()
}
