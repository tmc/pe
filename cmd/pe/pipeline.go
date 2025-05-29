package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"
)

// askCmd returns a cobra.Command for the 'ask' subcommand.
//
// ask is a pipeline-friendly command for single prompt evaluation.
//
// Usage:
//
//	echo "What is the capital of France?" | pe ask --provider openai:gpt-4
//	pe ask --provider anthropic:claude-3-haiku "Explain quantum computing"
func askCmd() *cobra.Command {
	var provider string
	var temperature float64
	var maxTokens int
	var outputFormat string

	cmd := &cobra.Command{
		Use:   "ask [prompt]",
		Short: "Ask a single question to an LLM provider",
		Long: `Ask is a pipeline-friendly command for single prompt evaluation.
It can read prompts from stdin or arguments and outputs responses in various formats.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			var prompt string
			
			// Read from stdin if no args provided
			if len(args) == 0 {
				stdin, err := io.ReadAll(os.Stdin)
				if err != nil {
					return fmt.Errorf("error reading from stdin: %v", err)
				}
				prompt = strings.TrimSpace(string(stdin))
			} else {
				prompt = strings.Join(args, " ")
			}

			if prompt == "" {
				return fmt.Errorf("no prompt provided")
			}

			// Create a minimal config for evaluation
			config := map[string]interface{}{
				"prompts":   []string{prompt},
				"providers": []string{provider},
				"tests": []map[string]interface{}{
					{"vars": map[string]interface{}{}},
				},
			}

			if temperature != -1 {
				config["defaultTest"] = map[string]interface{}{
					"options": map[string]interface{}{
						"temperature": temperature,
						"max_tokens":  maxTokens,
					},
				}
			}

			// TODO: Implement direct provider call instead of full eval
			// For now, create temp config and run evaluation
			configData, err := yaml.Marshal(config)
			if err != nil {
				return fmt.Errorf("error creating config: %v", err)
			}

			// Write to temporary file
			tmpFile, err := os.CreateTemp("", "pe-ask-*.yaml")
			if err != nil {
				return fmt.Errorf("error creating temp file: %v", err)
			}
			defer os.Remove(tmpFile.Name())
			defer tmpFile.Close()

			if _, err := tmpFile.Write(configData); err != nil {
				return fmt.Errorf("error writing temp config: %v", err)
			}
			tmpFile.Close()

			// Run evaluation
			evalCmd := evalCmd()
			evalCmd.SetArgs([]string{tmpFile.Name(), "--output", "/dev/stdout", "--format", outputFormat})
			evalCmd.SetOut(cmd.OutOrStdout())
			evalCmd.SetErr(cmd.OutOrStderr())

			return evalCmd.Execute()
		},
	}

	cmd.Flags().StringVarP(&provider, "provider", "p", "openai:gpt-4", "LLM provider to use")
	cmd.Flags().Float64VarP(&temperature, "temperature", "t", -1, "Temperature for generation")
	cmd.Flags().IntVarP(&maxTokens, "max-tokens", "m", 0, "Maximum tokens to generate")
	cmd.Flags().StringVarP(&outputFormat, "format", "f", "json", "Output format: json, yaml, text")

	return cmd
}

// streamCmd returns a cobra.Command for the 'stream' subcommand.
//
// stream processes evaluation results as a stream for pipeline processing.
//
// Usage:
//
//	pe eval config.yaml | pe stream --select response,latency
//	pe eval config.yaml | pe stream --format ndjson
func streamCmd() *cobra.Command {
	var selectFields string
	var format string
	var filter string

	cmd := &cobra.Command{
		Use:   "stream",
		Short: "Process evaluation results as a stream",
		Long: `Stream processes evaluation results line by line for pipeline processing.
Supports field selection, filtering, and format conversion.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStream(cmd, selectFields, format, filter)
		},
	}

	cmd.Flags().StringVarP(&selectFields, "select", "s", "", "Comma-separated list of fields to output")
	cmd.Flags().StringVarP(&format, "format", "f", "json", "Output format: json, ndjson, csv, tsv")
	cmd.Flags().StringVarP(&filter, "filter", "", "", "Filter expression (e.g., 'success=true')")

	return cmd
}

// filterCmd returns a cobra.Command for the 'filter' subcommand.
//
// filter filters evaluation results based on conditions.
//
// Usage:
//
//	pe eval config.yaml | pe filter --success
//	pe eval config.yaml | pe filter --provider openai --min-score 0.8
func filterCmd() *cobra.Command {
	var success bool
	var failure bool
	var provider string
	var minScore float64
	var maxLatency int

	cmd := &cobra.Command{
		Use:   "filter",
		Short: "Filter evaluation results",
		Long: `Filter evaluation results based on various conditions like success/failure,
provider, score thresholds, and performance metrics.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runFilter(cmd, success, failure, provider, minScore, maxLatency)
		},
	}

	cmd.Flags().BoolVar(&success, "success", false, "Filter for successful results only")
	cmd.Flags().BoolVar(&failure, "failure", false, "Filter for failed results only")
	cmd.Flags().StringVar(&provider, "provider", "", "Filter by provider")
	cmd.Flags().Float64Var(&minScore, "min-score", -1, "Minimum score threshold")
	cmd.Flags().IntVar(&maxLatency, "max-latency", -1, "Maximum latency in ms")

	return cmd
}

// analyzeCmd returns a cobra.Command for the 'analyze' subcommand.
//
// analyze computes statistics and insights from evaluation results.
//
// Usage:
//
//	pe eval config.yaml | pe analyze --metric latency
//	pe eval config.yaml | pe analyze --group-by provider --metric score
func analyzeCmd() *cobra.Command {
	var metric string
	var groupBy string
	var percentiles string

	cmd := &cobra.Command{
		Use:   "analyze",
		Short: "Analyze evaluation results",
		Long: `Analyze computes statistics and insights from evaluation results.
Supports grouping, percentile calculations, and various metrics.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAnalyze(cmd, metric, groupBy, percentiles)
		},
	}

	cmd.Flags().StringVarP(&metric, "metric", "m", "score", "Metric to analyze: score, latency, tokens, cost")
	cmd.Flags().StringVarP(&groupBy, "group-by", "g", "", "Group results by field: provider, prompt, test")
	cmd.Flags().StringVarP(&percentiles, "percentiles", "p", "50,90,95,99", "Comma-separated percentiles to calculate")

	return cmd
}

// statsCmd returns a cobra.Command for the 'stats' subcommand.
//
// stats provides quick statistics from evaluation results.
//
// Usage:
//
//	pe eval config.yaml | pe stats
//	pe eval config.yaml | pe stats --format table
func statsCmd() *cobra.Command {
	var format string

	cmd := &cobra.Command{
		Use:   "stats",
		Short: "Show quick statistics",
		Long: `Stats provides quick statistics from evaluation results including
success rate, average scores, latency metrics, and cost summaries.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStats(cmd, format)
		},
	}

	cmd.Flags().StringVarP(&format, "format", "f", "table", "Output format: table, json, yaml")

	return cmd
}

// diffCmd returns a cobra.Command for the 'diff' subcommand.
//
// diff compares two evaluation results.
//
// Usage:
//
//	pe diff baseline.json current.json
//	pe eval config.yaml | pe diff baseline.json
func diffCmd() *cobra.Command {
	var metric string
	var threshold float64

	cmd := &cobra.Command{
		Use:   "diff [baseline_file] [current_file]",
		Short: "Compare evaluation results",
		Long: `Diff compares two evaluation results and shows differences in
scores, latency, costs, and other metrics. Can detect regressions.`,
		Args: cobra.MaximumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDiff(cmd, args, metric, threshold)
		},
	}

	cmd.Flags().StringVarP(&metric, "metric", "m", "score", "Primary metric for comparison")
	cmd.Flags().Float64VarP(&threshold, "threshold", "t", 0.05, "Regression threshold (0.05 = 5%)")

	return cmd
}

// interactiveCmd returns a cobra.Command for the 'interactive' subcommand.
//
// interactive starts an interactive REPL for prompt development.
//
// Usage:
//
//	pe interactive --provider openai:gpt-4
//	pe interactive --config config.yaml
func interactiveCmd() *cobra.Command {
	var provider string
	var configFile string
	var temperature float64

	cmd := &cobra.Command{
		Use:   "interactive",
		Short: "Start interactive REPL mode",
		Long: `Interactive starts a REPL (Read-Eval-Print Loop) for rapid prompt development.
Supports provider switching, configuration loading, and session management.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInteractive(cmd, provider, configFile, temperature)
		},
	}

	cmd.Flags().StringVarP(&provider, "provider", "p", "openai:gpt-4", "Default LLM provider")
	cmd.Flags().StringVarP(&configFile, "config", "c", "", "Load configuration file")
	cmd.Flags().Float64VarP(&temperature, "temperature", "t", 0.7, "Default temperature")

	return cmd
}

// Implementation functions for the pipeline commands

func runStream(cmd *cobra.Command, selectFields, format, filter string) error {
	// Read evaluation results from stdin
	decoder := json.NewDecoder(os.Stdin)
	
	// Parse select fields
	var fields []string
	if selectFields != "" {
		fields = strings.Split(selectFields, ",")
		for i, field := range fields {
			fields[i] = strings.TrimSpace(field)
		}
	}

	for {
		var result map[string]interface{}
		if err := decoder.Decode(&result); err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("error decoding JSON: %v", err)
		}

		// Apply filter if specified
		if filter != "" && !matchesFilter(result, filter) {
			continue
		}

		// Select fields if specified
		if len(fields) > 0 {
			filtered := make(map[string]interface{})
			for _, field := range fields {
				if value, exists := getNestedValue(result, field); exists {
					filtered[field] = value
				}
			}
			result = filtered
		}

		// Output in specified format
		switch format {
		case "json":
			output, _ := json.Marshal(result)
			fmt.Fprintln(cmd.OutOrStdout(), string(output))
		case "ndjson":
			output, _ := json.Marshal(result)
			fmt.Fprintln(cmd.OutOrStdout(), string(output))
		case "csv", "tsv":
			separator := ","
			if format == "tsv" {
				separator = "\t"
			}
			// First result - print headers
			if len(fields) == 0 {
				// Use all keys as fields
				fields = make([]string, 0, len(result))
				for k := range result {
					fields = append(fields, k)
				}
				sort.Strings(fields)
			}
			// Print headers on first row
			fmt.Fprintln(cmd.OutOrStdout(), strings.Join(fields, separator))
			
			// Print values
			values := make([]string, len(fields))
			for i, field := range fields {
				if v, exists := result[field]; exists {
					values[i] = fmt.Sprintf("%v", v)
				} else {
					values[i] = ""
				}
			}
			fmt.Fprintln(cmd.OutOrStdout(), strings.Join(values, separator))
		default:
			return fmt.Errorf("unsupported format: %s", format)
		}
	}

	return nil
}

func runFilter(cmd *cobra.Command, success, failure bool, provider string, minScore float64, maxLatency int) error {
	decoder := json.NewDecoder(os.Stdin)
	encoder := json.NewEncoder(cmd.OutOrStdout())

	for {
		var result map[string]interface{}
		if err := decoder.Decode(&result); err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("error decoding JSON: %v", err)
		}

		// Apply filters
		if success && !isSuccessful(result) {
			continue
		}
		if failure && isSuccessful(result) {
			continue
		}
		if provider != "" && getStringValue(result, "provider") != provider {
			continue
		}
		if minScore >= 0 && getFloatValue(result, "score") < minScore {
			continue
		}
		if maxLatency >= 0 && getIntValue(result, "latencyMs") > maxLatency {
			continue
		}

		encoder.Encode(result)
	}

	return nil
}

func runAnalyze(cmd *cobra.Command, metric, groupBy, percentiles string) error {
	// Read all results first
	var results []map[string]interface{}
	decoder := json.NewDecoder(os.Stdin)

	for {
		var result map[string]interface{}
		if err := decoder.Decode(&result); err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("error decoding JSON: %v", err)
		}
		results = append(results, result)
	}

	// Parse percentiles
	var pctls []float64
	for _, p := range strings.Split(percentiles, ",") {
		if pct, err := strconv.ParseFloat(strings.TrimSpace(p), 64); err == nil {
			pctls = append(pctls, pct)
		}
	}

	// Group results if specified
	groups := make(map[string][]float64)
	if groupBy != "" {
		for _, result := range results {
			group := getStringValue(result, groupBy)
			value := getFloatValue(result, metric)
			groups[group] = append(groups[group], value)
		}
	} else {
		var values []float64
		for _, result := range results {
			value := getFloatValue(result, metric)
			values = append(values, value)
		}
		groups["all"] = values
	}

	// Calculate statistics for each group
	analysis := make(map[string]interface{})
	for group, values := range groups {
		if len(values) == 0 {
			continue
		}

		sort.Float64s(values)
		stats := map[string]interface{}{
			"count": len(values),
			"min":   values[0],
			"max":   values[len(values)-1],
			"mean":  mean(values),
		}

		// Calculate percentiles
		for _, pct := range pctls {
			stats[fmt.Sprintf("p%g", pct)] = percentile(values, pct)
		}

		analysis[group] = stats
	}

	// Output analysis
	output, err := json.MarshalIndent(analysis, "", "  ")
	if err != nil {
		return fmt.Errorf("error formatting analysis: %v", err)
	}

	fmt.Fprintln(cmd.OutOrStdout(), string(output))
	return nil
}

func runStats(cmd *cobra.Command, format string) error {
	var results []map[string]interface{}
	decoder := json.NewDecoder(os.Stdin)

	// Collect all results
	for {
		var result map[string]interface{}
		if err := decoder.Decode(&result); err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("error decoding JSON: %v", err)
		}
		results = append(results, result)
	}

	if len(results) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No results to analyze")
		return nil
	}

	// Calculate basic statistics
	stats := map[string]interface{}{
		"total_results": len(results),
		"success_rate":  calculateSuccessRate(results),
		"avg_score":     calculateAverageScore(results),
		"avg_latency":   calculateAverageLatency(results),
		"total_cost":    calculateTotalCost(results),
	}

	// Output in specified format
	switch format {
	case "json":
		output, _ := json.MarshalIndent(stats, "", "  ")
		fmt.Fprintln(cmd.OutOrStdout(), string(output))
	case "yaml":
		output, _ := yaml.Marshal(stats)
		fmt.Fprintln(cmd.OutOrStdout(), string(output))
	case "table":
		fmt.Fprintf(cmd.OutOrStdout(), "Total Results:  %d\n", stats["total_results"])
		fmt.Fprintf(cmd.OutOrStdout(), "Success Rate:   %.1f%%\n", stats["success_rate"].(float64)*100)
		fmt.Fprintf(cmd.OutOrStdout(), "Average Score:  %.3f\n", stats["avg_score"])
		fmt.Fprintf(cmd.OutOrStdout(), "Average Latency: %.0fms\n", stats["avg_latency"])
		fmt.Fprintf(cmd.OutOrStdout(), "Total Cost:     $%.4f\n", stats["total_cost"])
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}

	return nil
}

func runDiff(cmd *cobra.Command, args []string, metric string, threshold float64) error {
	var baseline, current []map[string]interface{}
	
	// Read baseline file
	if len(args) > 0 {
		data, err := os.ReadFile(args[0])
		if err != nil {
			return fmt.Errorf("error reading baseline file: %v", err)
		}
		if err := json.Unmarshal(data, &baseline); err != nil {
			return fmt.Errorf("error parsing baseline file: %v", err)
		}
	} else {
		return fmt.Errorf("baseline file required")
	}
	
	// Read current results from file or stdin
	if len(args) > 1 {
		data, err := os.ReadFile(args[1])
		if err != nil {
			return fmt.Errorf("error reading current file: %v", err)
		}
		if err := json.Unmarshal(data, &current); err != nil {
			return fmt.Errorf("error parsing current file: %v", err)
		}
	} else {
		// Read from stdin
		decoder := json.NewDecoder(os.Stdin)
		for {
			var result map[string]interface{}
			if err := decoder.Decode(&result); err != nil {
				if err == io.EOF {
					break
				}
				return fmt.Errorf("error decoding JSON: %v", err)
			}
			current = append(current, result)
		}
	}
	
	// Calculate statistics for both sets
	baselineStats := calculateStats(baseline, metric)
	currentStats := calculateStats(current, metric)
	
	// Calculate differences
	diff := map[string]interface{}{
		"metric": metric,
		"baseline": map[string]interface{}{
			"count": len(baseline),
			"mean":  baselineStats["mean"],
			"min":   baselineStats["min"],
			"max":   baselineStats["max"],
		},
		"current": map[string]interface{}{
			"count": len(current),
			"mean":  currentStats["mean"],
			"min":   currentStats["min"],
			"max":   currentStats["max"],
		},
	}
	
	// Calculate change
	baselineMean := baselineStats["mean"].(float64)
	currentMean := currentStats["mean"].(float64)
	
	if baselineMean != 0 {
		changePercent := ((currentMean - baselineMean) / baselineMean) * 100
		diff["change_percent"] = changePercent
		diff["regression"] = changePercent < -threshold*100
		diff["improvement"] = changePercent > threshold*100
		
		// Pretty print
		fmt.Fprintf(cmd.OutOrStdout(), "=== Evaluation Diff Report ===\n")
		fmt.Fprintf(cmd.OutOrStdout(), "Metric: %s\n\n", metric)
		
		fmt.Fprintf(cmd.OutOrStdout(), "Baseline:\n")
		fmt.Fprintf(cmd.OutOrStdout(), "  Count: %d\n", len(baseline))
		fmt.Fprintf(cmd.OutOrStdout(), "  Mean:  %.3f\n", baselineMean)
		fmt.Fprintf(cmd.OutOrStdout(), "  Min:   %.3f\n", baselineStats["min"])
		fmt.Fprintf(cmd.OutOrStdout(), "  Max:   %.3f\n\n", baselineStats["max"])
		
		fmt.Fprintf(cmd.OutOrStdout(), "Current:\n")
		fmt.Fprintf(cmd.OutOrStdout(), "  Count: %d\n", len(current))
		fmt.Fprintf(cmd.OutOrStdout(), "  Mean:  %.3f\n", currentMean)
		fmt.Fprintf(cmd.OutOrStdout(), "  Min:   %.3f\n", currentStats["min"])
		fmt.Fprintf(cmd.OutOrStdout(), "  Max:   %.3f\n\n", currentStats["max"])
		
		fmt.Fprintf(cmd.OutOrStdout(), "Change: %.2f%%\n", changePercent)
		
		if changePercent < -threshold*100 {
			fmt.Fprintf(cmd.OutOrStdout(), "⚠️  REGRESSION DETECTED (threshold: %.1f%%)\n", threshold*100)
		} else if changePercent > threshold*100 {
			fmt.Fprintf(cmd.OutOrStdout(), "✅ IMPROVEMENT DETECTED (threshold: %.1f%%)\n", threshold*100)
		} else {
			fmt.Fprintf(cmd.OutOrStdout(), "No significant change\n")
		}
	}
	
	return nil
}

func calculateStats(results []map[string]interface{}, metric string) map[string]interface{} {
	var values []float64
	
	for _, result := range results {
		value := getFloatValue(result, metric)
		values = append(values, value)
	}
	
	if len(values) == 0 {
		return map[string]interface{}{
			"mean": 0.0,
			"min":  0.0,
			"max":  0.0,
		}
	}
	
	sort.Float64s(values)
	
	return map[string]interface{}{
		"mean": mean(values),
		"min":  values[0],
		"max":  values[len(values)-1],
	}
}

func runInteractive(cmd *cobra.Command, provider, configFile string, temperature float64) error {
	// Create and run enhanced REPL session
	session := NewREPLSession(cmd, provider, configFile, temperature)
	return session.Run()
}

func printInteractiveHelp(cmd *cobra.Command) {
	help := `
Interactive Commands:
  :quit, :q           Exit interactive mode
  :help, :h           Show this help
  :provider <name>    Switch to a different provider
  :temp <value>       Set temperature (0.0-2.0)
  
Examples:
  :provider anthropic:claude-3-haiku
  :temp 0.9
  What is the capital of France?
`
	fmt.Fprint(cmd.OutOrStdout(), help)
}

// Helper functions

func matchesFilter(result map[string]interface{}, filter string) bool {
	// Simple filter implementation
	// TODO: Implement more sophisticated filtering
	return true
}

func getNestedValue(data map[string]interface{}, key string) (interface{}, bool) {
	parts := strings.Split(key, ".")
	current := data
	
	for i, part := range parts {
		if i == len(parts)-1 {
			value, exists := current[part]
			return value, exists
		}
		
		if next, ok := current[part].(map[string]interface{}); ok {
			current = next
		} else {
			return nil, false
		}
	}
	
	return nil, false
}

func getStringValue(data map[string]interface{}, key string) string {
	if value, exists := getNestedValue(data, key); exists {
		if str, ok := value.(string); ok {
			return str
		}
	}
	return ""
}

func getFloatValue(data map[string]interface{}, key string) float64 {
	if value, exists := getNestedValue(data, key); exists {
		if f, ok := value.(float64); ok {
			return f
		}
		if i, ok := value.(int); ok {
			return float64(i)
		}
	}
	return 0
}

func getIntValue(data map[string]interface{}, key string) int {
	if value, exists := getNestedValue(data, key); exists {
		if i, ok := value.(int); ok {
			return i
		}
		if f, ok := value.(float64); ok {
			return int(f)
		}
	}
	return 0
}

func isSuccessful(result map[string]interface{}) bool {
	// Check if all assertions passed
	if success, exists := getNestedValue(result, "success"); exists {
		if b, ok := success.(bool); ok {
			return b
		}
	}
	return false
}

func calculateSuccessRate(results []map[string]interface{}) float64 {
	if len(results) == 0 {
		return 0
	}
	
	successful := 0
	for _, result := range results {
		if isSuccessful(result) {
			successful++
		}
	}
	
	return float64(successful) / float64(len(results))
}

func calculateAverageScore(results []map[string]interface{}) float64 {
	if len(results) == 0 {
		return 0
	}
	
	total := 0.0
	count := 0
	for _, result := range results {
		if score := getFloatValue(result, "score"); score > 0 {
			total += score
			count++
		}
	}
	
	if count == 0 {
		return 0
	}
	return total / float64(count)
}

func calculateAverageLatency(results []map[string]interface{}) float64 {
	if len(results) == 0 {
		return 0
	}
	
	total := 0.0
	count := 0
	for _, result := range results {
		if latency := getFloatValue(result, "latencyMs"); latency > 0 {
			total += latency
			count++
		}
	}
	
	if count == 0 {
		return 0
	}
	return total / float64(count)
}

func calculateTotalCost(results []map[string]interface{}) float64 {
	total := 0.0
	for _, result := range results {
		if cost := getFloatValue(result, "cost"); cost > 0 {
			total += cost
		}
	}
	return total
}

func mean(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	
	total := 0.0
	for _, v := range values {
		total += v
	}
	return total / float64(len(values))
}

