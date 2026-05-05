package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tmc/pe/internal/promptfoo"
)

// statsCmd shows quick statistics from evaluation results
func statsCmd() *cobra.Command {
	var format string

	cmd := &cobra.Command{
		Use:   "stats [file]",
		Short: "Show quick statistics from evaluation results",
		Long: `Display statistics from evaluation results.

Input may be a promptfoo-style JSON result file, a JSON array of test results,
or JSONL test results. With no file argument, stats reads stdin.`,
		Example: `  pe stats results.json
  pe eval config.yaml -o results.json && pe stats --format json results.json
  cat results.jsonl | pe stats`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			in := cmd.InOrStdin()
			name := "stdin"
			if len(args) == 1 {
				f, err := os.Open(args[0])
				if err != nil {
					return fmt.Errorf("open %s: %w", args[0], err)
				}
				defer f.Close()
				in = f
				name = args[0]
			}

			summary, err := readResultSummary(in)
			if err != nil {
				return fmt.Errorf("read %s: %w", name, err)
			}
			return writeStats(cmd.OutOrStdout(), summary, format)
		},
	}
	cmd.Flags().StringVarP(&format, "format", "f", "text", "Output format: text or json")

	return cmd
}

// diffCmd compares two evaluation results
func diffCmd() *cobra.Command {
	var format string
	var failOnChange bool
	var failOnRegression bool
	var maxPassRateDrop float64
	var maxScoreDrop float64
	var maxLatencyIncrease float64
	var maxTokenIncrease int32
	var maxFailureIncrease int
	var maxErrorIncrease int

	cmd := &cobra.Command{
		Use:   "diff [baseline] [current]",
		Short: "Compare two evaluation results",
		Long: `Compare two evaluation result files and show deterministic metric deltas.

Use "-" as the current file to read the current result from stdin.`,
		Example: `  pe diff baseline.json current.json
  pe diff --format json baseline.json current.json
  pe eval config.yaml -o current.json && pe diff --fail-on-regression baseline.json current.json`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			base, err := readResultSummaryFile(args[0], cmd.InOrStdin())
			if err != nil {
				return fmt.Errorf("read baseline: %w", err)
			}
			current, err := readResultSummaryFile(args[1], cmd.InOrStdin())
			if err != nil {
				return fmt.Errorf("read current: %w", err)
			}
			diff := compareSummaries(base, current)
			gate := evaluateDiffGate(diff, diffGateConfig{
				FailOnChange:         failOnChange,
				FailOnRegression:     failOnRegression,
				MaxPassRateDrop:      maxPassRateDrop / 100,
				MaxScoreDrop:         maxScoreDrop,
				MaxLatencyIncreaseMs: maxLatencyIncrease,
				MaxTokenIncrease:     maxTokenIncrease,
				MaxFailureIncrease:   maxFailureIncrease,
				MaxErrorIncrease:     maxErrorIncrease,
			})
			if gate.Enabled {
				diff.Gate = &gate
			}
			if err := writeDiff(cmd.OutOrStdout(), diff, format); err != nil {
				return err
			}
			if gate.Enabled && !gate.Passed {
				cmd.SilenceUsage = true
				cmd.SilenceErrors = true
				return fmt.Errorf("diff gate failed: %s", strings.Join(gate.Reasons, "; "))
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&format, "format", "f", "text", "Output format: text or json")
	cmd.Flags().BoolVar(&failOnChange, "fail-on-change", false, "Exit non-zero if any compared metric changes")
	cmd.Flags().BoolVar(&failOnRegression, "fail-on-regression", false, "Exit non-zero if regression metrics exceed thresholds")
	cmd.Flags().Float64Var(&maxPassRateDrop, "max-pass-rate-drop", 0, "Allowed pass-rate drop in percentage points with --fail-on-regression")
	cmd.Flags().Float64Var(&maxScoreDrop, "max-score-drop", 0, "Allowed average score drop with --fail-on-regression")
	cmd.Flags().Float64Var(&maxLatencyIncrease, "max-latency-increase-ms", 0, "Allowed average latency increase in milliseconds with --fail-on-regression")
	cmd.Flags().Int32Var(&maxTokenIncrease, "max-token-increase", 0, "Allowed token total increase with --fail-on-regression")
	cmd.Flags().IntVar(&maxFailureIncrease, "max-failure-increase", 0, "Allowed failure count increase with --fail-on-regression")
	cmd.Flags().IntVar(&maxErrorIncrease, "max-error-increase", 0, "Allowed error count increase with --fail-on-regression")

	return cmd
}

type resultSummary struct {
	TotalTests       int                        `json:"totalTests"`
	Successes        int                        `json:"successes"`
	Failures         int                        `json:"failures"`
	Errors           int                        `json:"errors"`
	PassRate         float64                    `json:"passRate"`
	AverageScore     float64                    `json:"averageScore"`
	AverageLatencyMs float64                    `json:"averageLatencyMs"`
	TokenUsage       promptfoo.TokenUsage       `json:"tokenUsage"`
	Cost             float64                    `json:"cost"`
	Providers        map[string]providerSummary `json:"providers,omitempty"`
}

type providerSummary struct {
	TotalTests       int     `json:"totalTests"`
	Successes        int     `json:"successes"`
	Failures         int     `json:"failures"`
	Errors           int     `json:"errors"`
	PassRate         float64 `json:"passRate"`
	AverageScore     float64 `json:"averageScore"`
	AverageLatencyMs float64 `json:"averageLatencyMs"`
	TokenTotal       int32   `json:"tokenTotal"`
	Cost             float64 `json:"cost"`
}

type summaryBuilder struct {
	resultSummary
	scoreSum      float64
	scoreCount    int
	latencySum    int64
	latencyCount  int
	providerBuild map[string]*providerBuilder
}

type providerBuilder struct {
	providerSummary
	scoreSum     float64
	scoreCount   int
	latencySum   int64
	latencyCount int
}

type resultDiff struct {
	Baseline resultSummary `json:"baseline"`
	Current  resultSummary `json:"current"`
	Delta    diffDelta     `json:"delta"`
	Gate     *diffGate     `json:"gate,omitempty"`
}

type diffDelta struct {
	TotalTests       int     `json:"totalTests"`
	Successes        int     `json:"successes"`
	Failures         int     `json:"failures"`
	Errors           int     `json:"errors"`
	PassRate         float64 `json:"passRate"`
	AverageScore     float64 `json:"averageScore"`
	AverageLatencyMs float64 `json:"averageLatencyMs"`
	TokenTotal       int32   `json:"tokenTotal"`
	Cost             float64 `json:"cost"`
}

type diffGateConfig struct {
	FailOnChange         bool
	FailOnRegression     bool
	MaxPassRateDrop      float64
	MaxScoreDrop         float64
	MaxLatencyIncreaseMs float64
	MaxTokenIncrease     int32
	MaxFailureIncrease   int
	MaxErrorIncrease     int
}

type diffGate struct {
	Enabled bool     `json:"enabled"`
	Passed  bool     `json:"passed"`
	Reasons []string `json:"reasons,omitempty"`
}

func readResultSummaryFile(name string, stdin io.Reader) (resultSummary, error) {
	if name == "-" {
		return readResultSummary(stdin)
	}
	f, err := os.Open(name)
	if err != nil {
		return resultSummary{}, fmt.Errorf("open %s: %w", name, err)
	}
	defer f.Close()
	return readResultSummary(f)
}

func readResultSummary(r io.Reader) (resultSummary, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return resultSummary{}, err
	}
	data = bytes.TrimSpace(data)
	if len(data) == 0 {
		return resultSummary{}, fmt.Errorf("empty input")
	}

	var eval promptfoo.EvaluationResult
	if err := json.Unmarshal(data, &eval); err == nil {
		if len(eval.Results.Results) > 0 || eval.Results.Stats.Successes+eval.Results.Stats.Failures+eval.Results.Stats.Errors > 0 {
			return summarizeEvaluation(eval), nil
		}
	}

	var evalResults promptfoo.EvalResults
	if err := json.Unmarshal(data, &evalResults); err == nil {
		if len(evalResults.Results) > 0 || evalResults.Stats.Successes+evalResults.Stats.Failures+evalResults.Stats.Errors > 0 {
			return summarizeResultSet(evalResults.Results, evalResults.Stats), nil
		}
	}

	var results []promptfoo.TestResult
	if err := json.Unmarshal(data, &results); err == nil {
		return summarizeResultSet(results, promptfoo.Stats{}), nil
	}

	results, err = readJSONLResults(bytes.NewReader(data))
	if err != nil {
		return resultSummary{}, fmt.Errorf("parse JSON or JSONL results: %w", err)
	}
	return summarizeResultSet(results, promptfoo.Stats{}), nil
}

func readJSONLResults(r io.Reader) ([]promptfoo.TestResult, error) {
	var results []promptfoo.TestResult
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var result promptfoo.TestResult
		if err := json.Unmarshal([]byte(line), &result); err == nil {
			results = append(results, result)
			continue
		}
		var eval promptfoo.EvaluationResult
		if err := json.Unmarshal([]byte(line), &eval); err == nil && len(eval.Results.Results) > 0 {
			results = append(results, eval.Results.Results...)
			continue
		}
		return nil, fmt.Errorf("invalid JSONL result: %q", line)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("no results found")
	}
	return results, nil
}

func summarizeEvaluation(eval promptfoo.EvaluationResult) resultSummary {
	return summarizeResultSet(eval.Results.Results, eval.Results.Stats)
}

func summarizeResultSet(results []promptfoo.TestResult, stats promptfoo.Stats) resultSummary {
	b := summaryBuilder{
		resultSummary: resultSummary{
			Providers: make(map[string]providerSummary),
		},
		providerBuild: make(map[string]*providerBuilder),
	}

	for _, result := range results {
		b.addResult(result)
	}
	if len(results) == 0 {
		b.Successes = stats.Successes
		b.Failures = stats.Failures
		b.Errors = stats.Errors
		b.TotalTests = stats.Successes + stats.Failures + stats.Errors
		b.TokenUsage = stats.TokenUsage
	}
	return b.finish()
}

func (b *summaryBuilder) addResult(result promptfoo.TestResult) {
	b.TotalTests++
	if result.Success || result.GradingResult.Pass {
		b.Successes++
	} else {
		b.Failures++
	}
	score := result.Score
	if score == 0 && result.GradingResult.Score != 0 {
		score = result.GradingResult.Score
	}
	b.scoreSum += score
	b.scoreCount++
	latency := result.LatencyMs
	if latency == 0 {
		latency = result.Response.LatencyMs
	}
	if latency > 0 {
		b.latencySum += latency
		b.latencyCount++
	}
	if result.Response.TokenUsage != nil {
		addTokenUsage(&b.TokenUsage, *result.Response.TokenUsage)
	}
	addTokenUsage(&b.TokenUsage, result.GradingResult.TokensUsed)
	b.Cost += result.Response.Cost

	provider := resultProvider(result)
	if provider == "" {
		provider = "(unknown)"
	}
	pb := b.providerBuild[provider]
	if pb == nil {
		pb = &providerBuilder{}
		b.providerBuild[provider] = pb
	}
	pb.addResult(result)
}

func (b *summaryBuilder) finish() resultSummary {
	if b.TotalTests > 0 {
		b.PassRate = float64(b.Successes) / float64(b.TotalTests)
	}
	if b.scoreCount > 0 {
		b.AverageScore = b.scoreSum / float64(b.scoreCount)
	}
	if b.latencyCount > 0 {
		b.AverageLatencyMs = float64(b.latencySum) / float64(b.latencyCount)
	}

	providers := make([]string, 0, len(b.providerBuild))
	for provider := range b.providerBuild {
		providers = append(providers, provider)
	}
	sort.Strings(providers)
	for _, provider := range providers {
		b.Providers[provider] = b.providerBuild[provider].finish()
	}
	if len(b.Providers) == 0 {
		b.Providers = nil
	}
	return b.resultSummary
}

func (b *providerBuilder) addResult(result promptfoo.TestResult) {
	b.TotalTests++
	if result.Success || result.GradingResult.Pass {
		b.Successes++
	} else {
		b.Failures++
	}
	score := result.Score
	if score == 0 && result.GradingResult.Score != 0 {
		score = result.GradingResult.Score
	}
	b.scoreSum += score
	b.scoreCount++
	latency := result.LatencyMs
	if latency == 0 {
		latency = result.Response.LatencyMs
	}
	if latency > 0 {
		b.latencySum += latency
		b.latencyCount++
	}
	if result.Response.TokenUsage != nil {
		b.TokenTotal += result.Response.TokenUsage.Total
	}
	b.TokenTotal += result.GradingResult.TokensUsed.Total
	b.Cost += result.Response.Cost
}

func (b *providerBuilder) finish() providerSummary {
	if b.TotalTests > 0 {
		b.PassRate = float64(b.Successes) / float64(b.TotalTests)
	}
	if b.scoreCount > 0 {
		b.AverageScore = b.scoreSum / float64(b.scoreCount)
	}
	if b.latencyCount > 0 {
		b.AverageLatencyMs = float64(b.latencySum) / float64(b.latencyCount)
	}
	return b.providerSummary
}

func addTokenUsage(total *promptfoo.TokenUsage, usage promptfoo.TokenUsage) {
	total.Total += usage.Total
	total.Prompt += usage.Prompt
	total.Completion += usage.Completion
	total.Cached += usage.Cached
	total.NumRequests += usage.NumRequests
}

func resultProvider(result promptfoo.TestResult) string {
	for _, key := range []string{"label", "id", "name"} {
		if v := result.Provider[key]; v != "" {
			return v
		}
	}
	return ""
}

func writeStats(w io.Writer, summary resultSummary, format string) error {
	switch format {
	case "", "text":
		fmt.Fprintf(w, "Total tests: %d\n", summary.TotalTests)
		fmt.Fprintf(w, "Successes: %d\n", summary.Successes)
		fmt.Fprintf(w, "Failures: %d\n", summary.Failures)
		fmt.Fprintf(w, "Errors: %d\n", summary.Errors)
		fmt.Fprintf(w, "Pass rate: %.2f%%\n", summary.PassRate*100)
		fmt.Fprintf(w, "Average score: %.4f\n", summary.AverageScore)
		fmt.Fprintf(w, "Average latency: %.2f ms\n", summary.AverageLatencyMs)
		fmt.Fprintf(w, "Total tokens: %d\n", summary.TokenUsage.Total)
		fmt.Fprintf(w, "Cost: %.6f\n", summary.Cost)
		if len(summary.Providers) > 0 {
			fmt.Fprintln(w, "\nProviders:")
			names := make([]string, 0, len(summary.Providers))
			for name := range summary.Providers {
				names = append(names, name)
			}
			sort.Strings(names)
			for _, name := range names {
				p := summary.Providers[name]
				fmt.Fprintf(w, "  %s: tests=%d pass_rate=%.2f%% avg_score=%.4f avg_latency=%.2fms tokens=%d cost=%.6f\n",
					name, p.TotalTests, p.PassRate*100, p.AverageScore, p.AverageLatencyMs, p.TokenTotal, p.Cost)
			}
		}
		return nil
	case "json":
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(summary)
	default:
		return fmt.Errorf("unsupported format %q", format)
	}
}

func compareSummaries(base, current resultSummary) resultDiff {
	return resultDiff{
		Baseline: base,
		Current:  current,
		Delta: diffDelta{
			TotalTests:       current.TotalTests - base.TotalTests,
			Successes:        current.Successes - base.Successes,
			Failures:         current.Failures - base.Failures,
			Errors:           current.Errors - base.Errors,
			PassRate:         current.PassRate - base.PassRate,
			AverageScore:     current.AverageScore - base.AverageScore,
			AverageLatencyMs: current.AverageLatencyMs - base.AverageLatencyMs,
			TokenTotal:       current.TokenUsage.Total - base.TokenUsage.Total,
			Cost:             current.Cost - base.Cost,
		},
	}
}

func evaluateDiffGate(diff resultDiff, cfg diffGateConfig) diffGate {
	gate := diffGate{
		Enabled: cfg.FailOnChange || cfg.FailOnRegression,
		Passed:  true,
	}
	if !gate.Enabled {
		return gate
	}

	if cfg.FailOnChange {
		addChangeReason(&gate, "total tests", float64(diff.Delta.TotalTests))
		addChangeReason(&gate, "successes", float64(diff.Delta.Successes))
		addChangeReason(&gate, "failures", float64(diff.Delta.Failures))
		addChangeReason(&gate, "errors", float64(diff.Delta.Errors))
		addChangeReason(&gate, "pass rate", diff.Delta.PassRate*100)
		addChangeReason(&gate, "average score", diff.Delta.AverageScore)
		addChangeReason(&gate, "average latency", diff.Delta.AverageLatencyMs)
		addChangeReason(&gate, "total tokens", float64(diff.Delta.TokenTotal))
		addChangeReason(&gate, "cost", diff.Delta.Cost)
	}

	if cfg.FailOnRegression {
		if drop := -diff.Delta.PassRate; drop > cfg.MaxPassRateDrop {
			gate.fail(fmt.Sprintf("pass rate dropped %.2f percentage points", drop*100))
		}
		if drop := -diff.Delta.AverageScore; drop > cfg.MaxScoreDrop {
			gate.fail(fmt.Sprintf("average score dropped %.4f", drop))
		}
		if diff.Delta.AverageLatencyMs > cfg.MaxLatencyIncreaseMs {
			gate.fail(fmt.Sprintf("average latency increased %.2f ms", diff.Delta.AverageLatencyMs))
		}
		if diff.Delta.TokenTotal > cfg.MaxTokenIncrease {
			gate.fail(fmt.Sprintf("total tokens increased %+d", diff.Delta.TokenTotal))
		}
		if diff.Delta.Failures > cfg.MaxFailureIncrease {
			gate.fail(fmt.Sprintf("failures increased %+d", diff.Delta.Failures))
		}
		if diff.Delta.Errors > cfg.MaxErrorIncrease {
			gate.fail(fmt.Sprintf("errors increased %+d", diff.Delta.Errors))
		}
	}

	return gate
}

func addChangeReason(gate *diffGate, name string, delta float64) {
	if absFloat(delta) > 0.0000001 {
		gate.fail(fmt.Sprintf("%s changed", name))
	}
}

func (g *diffGate) fail(reason string) {
	g.Passed = false
	g.Reasons = append(g.Reasons, reason)
}

func writeDiff(w io.Writer, diff resultDiff, format string) error {
	switch format {
	case "", "text":
		fmt.Fprintln(w, "Evaluation diff:")
		fmt.Fprintf(w, "  Total tests: %s\n", signedInt(diff.Delta.TotalTests))
		fmt.Fprintf(w, "  Successes: %s\n", signedInt(diff.Delta.Successes))
		fmt.Fprintf(w, "  Failures: %s\n", signedInt(diff.Delta.Failures))
		fmt.Fprintf(w, "  Errors: %s\n", signedInt(diff.Delta.Errors))
		fmt.Fprintf(w, "  Pass rate: %+.2f%%\n", diff.Delta.PassRate*100)
		fmt.Fprintf(w, "  Average score: %+.4f\n", diff.Delta.AverageScore)
		fmt.Fprintf(w, "  Average latency: %+.2f ms\n", diff.Delta.AverageLatencyMs)
		fmt.Fprintf(w, "  Total tokens: %+d\n", diff.Delta.TokenTotal)
		fmt.Fprintf(w, "  Cost: %+.6f\n", diff.Delta.Cost)
		if diff.Gate != nil {
			status := "pass"
			if !diff.Gate.Passed {
				status = "fail"
			}
			fmt.Fprintf(w, "\nGate: %s\n", status)
			for _, reason := range diff.Gate.Reasons {
				fmt.Fprintf(w, "  - %s\n", reason)
			}
		}
		return nil
	case "json":
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(diff)
	default:
		return fmt.Errorf("unsupported format %q", format)
	}
}

func signedInt(n int) string {
	return fmt.Sprintf("%+d", n)
}

func absFloat(n float64) float64 {
	if n < 0 {
		return -n
	}
	return n
}

// interactiveCmd starts interactive REPL mode
func interactiveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "interactive",
		Short: "Start interactive REPL mode for prompt development",
		Long:  `Launch an interactive REPL (Read-Eval-Print Loop) for iterative prompt development.`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			provider, err := cmd.Flags().GetString("provider")
			if err != nil {
				return err
			}
			configFile, err := cmd.Flags().GetString("config")
			if err != nil {
				return err
			}
			temperature, err := cmd.Flags().GetFloat64("temperature")
			if err != nil {
				return err
			}
			if temperature < 0 || temperature > 2 {
				return fmt.Errorf("temperature must be between 0.0 and 2.0")
			}

			return NewREPLSession(cmd, provider, configFile, temperature).Run()
		},
	}

	cmd.Flags().String("provider", "openai:gpt-4", "LLM provider to use")
	cmd.Flags().String("config", "", "Path to configuration file")
	cmd.Flags().Float64("temperature", 0.7, "Generation temperature")

	return cmd
}
