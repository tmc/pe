package testing

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"time"

	"github.com/tmc/pe/internal/llm"
)

// RegressionTest represents a regression test configuration
type RegressionTest struct {
	Name          string                 `yaml:"name"`
	Baseline      string                 `yaml:"baseline"`
	Tolerance     float64                `yaml:"tolerance"`
	Metrics       []string               `yaml:"metrics"`
	ThresholdType string                 `yaml:"threshold_type"` // "absolute" or "relative"
	Options       map[string]interface{} `yaml:"options"`
}

// RegressionResult contains the result of a regression test
type RegressionResult struct {
	Name          string                  `json:"name"`
	Passed        bool                    `json:"passed"`
	BaselineFile  string                  `json:"baseline_file"`
	Tolerance     float64                 `json:"tolerance"`
	Metrics       map[string]MetricResult `json:"metrics"`
	OverallChange float64                 `json:"overall_change"`
	Regressions   []RegressionDetection   `json:"regressions"`
	Improvements  []ImprovementDetection  `json:"improvements"`
	TotalDuration time.Duration           `json:"total_duration"`
}

// MetricResult contains comparison results for a specific metric
type MetricResult struct {
	Baseline      float64 `json:"baseline"`
	Current       float64 `json:"current"`
	Change        float64 `json:"change"`
	ChangePercent float64 `json:"change_percent"`
	Passed        bool    `json:"passed"`
}

// RegressionDetection represents a detected regression
type RegressionDetection struct {
	Metric        string  `json:"metric"`
	TestCase      string  `json:"test_case"`
	Baseline      float64 `json:"baseline"`
	Current       float64 `json:"current"`
	Change        float64 `json:"change"`
	ChangePercent float64 `json:"change_percent"`
	Severity      string  `json:"severity"`
}

// ImprovementDetection represents a detected improvement
type ImprovementDetection struct {
	Metric             string  `json:"metric"`
	TestCase           string  `json:"test_case"`
	Baseline           float64 `json:"baseline"`
	Current            float64 `json:"current"`
	Improvement        float64 `json:"improvement"`
	ImprovementPercent float64 `json:"improvement_percent"`
}

// BaselineData represents stored baseline evaluation results
type BaselineData struct {
	Timestamp time.Time                 `json:"timestamp"`
	Provider  string                    `json:"provider"`
	Model     string                    `json:"model"`
	TestCases map[string]TestCaseResult `json:"test_cases"`
	Summary   SummaryMetrics            `json:"summary"`
}

// TestCaseResult contains results for a single test case
type TestCaseResult struct {
	Prompt   string        `json:"prompt"`
	Response string        `json:"response"`
	Latency  time.Duration `json:"latency"`
	Tokens   TokenMetrics  `json:"tokens"`
	Cost     float64       `json:"cost"`
	Score    float64       `json:"score"`
	Success  bool          `json:"success"`
}

// TokenMetrics contains token usage information
type TokenMetrics struct {
	Prompt     int `json:"prompt"`
	Completion int `json:"completion"`
	Total      int `json:"total"`
}

// SummaryMetrics contains aggregated metrics
type SummaryMetrics struct {
	SuccessRate    float64 `json:"success_rate"`
	AverageLatency float64 `json:"average_latency"`
	AverageScore   float64 `json:"average_score"`
	TotalCost      float64 `json:"total_cost"`
	TotalTokens    int     `json:"total_tokens"`
	TestCount      int     `json:"test_count"`
}

// RegressionTester handles regression testing
type RegressionTester struct {
	provider llm.Provider
	options  llm.GenerateOptions
}

// NewRegressionTester creates a new regression tester
func NewRegressionTester(provider llm.Provider, options llm.GenerateOptions) *RegressionTester {
	return &RegressionTester{
		provider: provider,
		options:  options,
	}
}

// RunRegressionTest executes a regression test
func (rt *RegressionTester) RunRegressionTest(ctx context.Context, test RegressionTest, currentResults BaselineData) (*RegressionResult, error) {
	startTime := time.Now()

	// Load baseline data
	baselineData, err := rt.loadBaseline(test.Baseline)
	if err != nil {
		return nil, fmt.Errorf("failed to load baseline: %v", err)
	}

	result := &RegressionResult{
		Name:         test.Name,
		BaselineFile: test.Baseline,
		Tolerance:    test.Tolerance,
		Metrics:      make(map[string]MetricResult),
		Regressions:  []RegressionDetection{},
		Improvements: []ImprovementDetection{},
	}

	// Compare metrics
	for _, metric := range test.Metrics {
		metricResult := rt.compareMetric(metric, *baselineData, currentResults, test.Tolerance, test.ThresholdType)
		result.Metrics[metric] = metricResult

		if !metricResult.Passed {
			result.Passed = false
		}
	}

	// Detect regressions and improvements
	rt.detectChanges(*baselineData, currentResults, test.Tolerance, result)

	// Calculate overall change
	result.OverallChange = rt.calculateOverallChange(result.Metrics)
	result.TotalDuration = time.Since(startTime)

	// Determine if test passed overall
	result.Passed = len(result.Regressions) == 0

	return result, nil
}

// loadBaseline loads baseline data from a file
func (rt *RegressionTester) loadBaseline(filename string) (*BaselineData, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read baseline file: %v", err)
	}

	var baseline BaselineData
	if err := json.Unmarshal(data, &baseline); err != nil {
		return nil, fmt.Errorf("failed to parse baseline data: %v", err)
	}

	return &baseline, nil
}

// SaveBaseline saves current results as a new baseline
func (rt *RegressionTester) SaveBaseline(filename string, data BaselineData) error {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal baseline data: %v", err)
	}

	if err := os.WriteFile(filename, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write baseline file: %v", err)
	}

	return nil
}

// compareMetric compares a specific metric between baseline and current results
func (rt *RegressionTester) compareMetric(metric string, baseline, current BaselineData, tolerance float64, thresholdType string) MetricResult {
	baselineValue := rt.extractMetricValue(metric, baseline)
	currentValue := rt.extractMetricValue(metric, current)

	change := currentValue - baselineValue
	changePercent := 0.0
	if baselineValue != 0 {
		changePercent = (change / baselineValue) * 100
	}

	// Determine if change is within tolerance
	passed := true
	if thresholdType == "relative" {
		passed = math.Abs(changePercent) <= tolerance
	} else {
		passed = math.Abs(change) <= tolerance
	}

	return MetricResult{
		Baseline:      baselineValue,
		Current:       currentValue,
		Change:        change,
		ChangePercent: changePercent,
		Passed:        passed,
	}
}

// extractMetricValue extracts a metric value from baseline data
func (rt *RegressionTester) extractMetricValue(metric string, data BaselineData) float64 {
	switch metric {
	case "success_rate":
		return data.Summary.SuccessRate
	case "average_latency":
		return data.Summary.AverageLatency
	case "average_score":
		return data.Summary.AverageScore
	case "total_cost":
		return data.Summary.TotalCost
	case "total_tokens":
		return float64(data.Summary.TotalTokens)
	default:
		return 0.0
	}
}

// detectChanges detects regressions and improvements in individual test cases
func (rt *RegressionTester) detectChanges(baseline, current BaselineData, tolerance float64, result *RegressionResult) {
	for testID, currentCase := range current.TestCases {
		baselineCase, exists := baseline.TestCases[testID]
		if !exists {
			continue
		}

		// Check latency regression
		if rt.isRegression("latency", baselineCase.Latency.Seconds(), currentCase.Latency.Seconds(), tolerance) {
			result.Regressions = append(result.Regressions, RegressionDetection{
				Metric:        "latency",
				TestCase:      testID,
				Baseline:      baselineCase.Latency.Seconds(),
				Current:       currentCase.Latency.Seconds(),
				Change:        currentCase.Latency.Seconds() - baselineCase.Latency.Seconds(),
				ChangePercent: ((currentCase.Latency.Seconds() - baselineCase.Latency.Seconds()) / baselineCase.Latency.Seconds()) * 100,
				Severity:      rt.calculateSeverity(baselineCase.Latency.Seconds(), currentCase.Latency.Seconds()),
			})
		}

		// Check score regression (lower score is worse)
		if rt.isRegression("score", currentCase.Score, baselineCase.Score, tolerance) {
			result.Regressions = append(result.Regressions, RegressionDetection{
				Metric:        "score",
				TestCase:      testID,
				Baseline:      baselineCase.Score,
				Current:       currentCase.Score,
				Change:        currentCase.Score - baselineCase.Score,
				ChangePercent: ((currentCase.Score - baselineCase.Score) / baselineCase.Score) * 100,
				Severity:      rt.calculateSeverity(baselineCase.Score, currentCase.Score),
			})
		}

		// Check cost regression (higher cost is worse)
		if rt.isRegression("cost", baselineCase.Cost, currentCase.Cost, tolerance) {
			result.Regressions = append(result.Regressions, RegressionDetection{
				Metric:        "cost",
				TestCase:      testID,
				Baseline:      baselineCase.Cost,
				Current:       currentCase.Cost,
				Change:        currentCase.Cost - baselineCase.Cost,
				ChangePercent: ((currentCase.Cost - baselineCase.Cost) / baselineCase.Cost) * 100,
				Severity:      rt.calculateSeverity(baselineCase.Cost, currentCase.Cost),
			})
		}

		// Detect improvements
		rt.detectImprovements(testID, baselineCase, currentCase, tolerance, result)
	}
}

// isRegression determines if a change constitutes a regression
func (rt *RegressionTester) isRegression(metric string, baseline, current, tolerance float64) bool {
	change := current - baseline
	changePercent := 0.0
	if baseline != 0 {
		changePercent = (change / baseline) * 100
	}

	switch metric {
	case "latency", "cost":
		// Higher values are worse
		return changePercent > tolerance
	case "score":
		// Lower values are worse
		return changePercent < -tolerance
	default:
		return false
	}
}

// detectImprovements detects significant improvements
func (rt *RegressionTester) detectImprovements(testID string, baseline, current TestCaseResult, tolerance float64, result *RegressionResult) {
	// Latency improvement (lower is better)
	if baseline.Latency.Seconds() > 0 {
		improvement := (baseline.Latency.Seconds() - current.Latency.Seconds()) / baseline.Latency.Seconds() * 100
		if improvement > tolerance {
			result.Improvements = append(result.Improvements, ImprovementDetection{
				Metric:             "latency",
				TestCase:           testID,
				Baseline:           baseline.Latency.Seconds(),
				Current:            current.Latency.Seconds(),
				Improvement:        baseline.Latency.Seconds() - current.Latency.Seconds(),
				ImprovementPercent: improvement,
			})
		}
	}

	// Score improvement (higher is better)
	if baseline.Score > 0 {
		improvement := (current.Score - baseline.Score) / baseline.Score * 100
		if improvement > tolerance {
			result.Improvements = append(result.Improvements, ImprovementDetection{
				Metric:             "score",
				TestCase:           testID,
				Baseline:           baseline.Score,
				Current:            current.Score,
				Improvement:        current.Score - baseline.Score,
				ImprovementPercent: improvement,
			})
		}
	}

	// Cost improvement (lower is better)
	if baseline.Cost > 0 {
		improvement := (baseline.Cost - current.Cost) / baseline.Cost * 100
		if improvement > tolerance {
			result.Improvements = append(result.Improvements, ImprovementDetection{
				Metric:             "cost",
				TestCase:           testID,
				Baseline:           baseline.Cost,
				Current:            current.Cost,
				Improvement:        baseline.Cost - current.Cost,
				ImprovementPercent: improvement,
			})
		}
	}
}

// calculateSeverity determines the severity of a regression
func (rt *RegressionTester) calculateSeverity(baseline, current float64) string {
	if baseline == 0 {
		return "unknown"
	}

	changePercent := math.Abs((current - baseline) / baseline * 100)

	switch {
	case changePercent >= 50:
		return "critical"
	case changePercent >= 25:
		return "high"
	case changePercent >= 10:
		return "medium"
	default:
		return "low"
	}
}

// calculateOverallChange calculates the overall change across all metrics
func (rt *RegressionTester) calculateOverallChange(metrics map[string]MetricResult) float64 {
	if len(metrics) == 0 {
		return 0.0
	}

	totalChange := 0.0
	for _, metric := range metrics {
		totalChange += math.Abs(metric.ChangePercent)
	}

	return totalChange / float64(len(metrics))
}

// CreateBaselineFromEvaluation creates baseline data from evaluation results
func CreateBaselineFromEvaluation(provider, model string, testCases map[string]TestCaseResult) BaselineData {
	// Calculate summary metrics
	var totalLatency float64
	var totalScore float64
	var totalCost float64
	var totalTokens int
	var successCount int

	for _, testCase := range testCases {
		totalLatency += testCase.Latency.Seconds()
		totalScore += testCase.Score
		totalCost += testCase.Cost
		totalTokens += testCase.Tokens.Total
		if testCase.Success {
			successCount++
		}
	}

	testCount := len(testCases)
	summary := SummaryMetrics{
		SuccessRate:    float64(successCount) / float64(testCount),
		AverageLatency: totalLatency / float64(testCount),
		AverageScore:   totalScore / float64(testCount),
		TotalCost:      totalCost,
		TotalTokens:    totalTokens,
		TestCount:      testCount,
	}

	return BaselineData{
		Timestamp: time.Now(),
		Provider:  provider,
		Model:     model,
		TestCases: testCases,
		Summary:   summary,
	}
}
