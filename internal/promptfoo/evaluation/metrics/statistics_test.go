package metrics

import (
	"math"
	"strings"
	"testing"
)

func TestStatisticalAnalyzerSummary(t *testing.T) {
	sa := NewStatisticalAnalyzer(0.95, 4)
	data := []float64{1, 2, 2, 3, 4, 100}
	summary, err := sa.CalculateStatisticalSummary(data)
	if err != nil {
		t.Fatal(err)
	}
	if summary.Count != len(data) {
		t.Fatalf("count = %d, want %d", summary.Count, len(data))
	}
	if summary.Median != 2.5 {
		t.Fatalf("median = %v, want 2.5", summary.Median)
	}
	if len(summary.Mode) != 1 || summary.Mode[0] != 2 {
		t.Fatalf("mode = %v, want [2]", summary.Mode)
	}
	if summary.Min != 1 || summary.Max != 100 || summary.Range != 99 {
		t.Fatalf("range fields = min %v max %v range %v", summary.Min, summary.Max, summary.Range)
	}
	if summary.Percentiles["50th"] != summary.Median {
		t.Fatalf("50th percentile = %v, median = %v", summary.Percentiles["50th"], summary.Median)
	}
	if summary.ConfidenceInterval.Level != 0.95 || !math.IsNaN(summary.ConfidenceInterval.MarginOfError) {
		t.Fatalf("confidence interval = %#v", summary.ConfidenceInterval)
	}
	if summary.Distribution.DistributionType != "unknown" || !math.IsNaN(summary.Distribution.Normality) {
		t.Fatalf("distribution = %#v", summary.Distribution)
	}
	if len(summary.Outliers) != 1 || summary.Outliers[0] != 100 {
		t.Fatalf("outliers = %v, want [100]", summary.Outliers)
	}
}

func TestStatisticalAnalyzerSummaryErrors(t *testing.T) {
	sa := NewStatisticalAnalyzer(2, -1)
	if _, err := sa.CalculateStatisticalSummary(nil); err == nil || !strings.Contains(err.Error(), "empty dataset") {
		t.Fatalf("empty summary error = %v", err)
	}
	if _, err := sa.CalculateStatisticalSummary([]float64{1, 2}); err == nil || !strings.Contains(err.Error(), "below minimum") {
		t.Fatalf("small summary error = %v", err)
	}
}

func TestStatisticalAnalyzerTTests(t *testing.T) {
	sa := NewStatisticalAnalyzer(0.95, 3)
	group1 := []float64{10, 11, 12, 13, 14, 15}
	group2 := []float64{1, 2, 3, 4, 5, 6}

	if result, err := sa.PerformTTest(group1, group2, false); err == nil || result != nil || !strings.Contains(err.Error(), "not yet implemented") {
		t.Fatalf("independent result = %#v err=%v", result, err)
	}

	if result, err := sa.PerformTTest(group1, []float64{2, 3, 3, 5, 6, 8}, true); err == nil || result != nil || !strings.Contains(err.Error(), "not yet implemented") {
		t.Fatalf("paired result = %#v err=%v", result, err)
	}

	if _, err := sa.PerformTTest(nil, group2, false); err == nil {
		t.Fatalf("empty groups did not error")
	}
	if _, err := sa.PerformTTest(group1, group2[:3], true); err == nil || !strings.Contains(err.Error(), "equal sample sizes") {
		t.Fatalf("paired size error = %v", err)
	}
}

func TestStatisticalAnalyzerABTest(t *testing.T) {
	sa := NewStatisticalAnalyzer(0.95, 3)
	if result, err := sa.PerformABTest(60, 100, 75, 100); err == nil || result != nil || !strings.Contains(err.Error(), "not yet implemented") {
		t.Fatalf("ab result = %#v err=%v", result, err)
	}

	if _, err := sa.PerformABTest(1, 0, 1, 2); err == nil || !strings.Contains(err.Error(), "invalid trial") {
		t.Fatalf("trial error = %v", err)
	}
	if _, err := sa.PerformABTest(3, 2, 1, 2); err == nil || !strings.Contains(err.Error(), "successes") {
		t.Fatalf("success error = %v", err)
	}
}

func TestStatisticalAnalyzerCompareGroups(t *testing.T) {
	sa := NewStatisticalAnalyzer(0.95, 3)
	result, err := sa.CompareGroups(
		[]float64{10, 11, 12, 13, 14, 15},
		[]float64{1, 2, 3, 4, 5, 6},
	)
	if err == nil || result != nil || !strings.Contains(err.Error(), "not yet implemented") {
		t.Fatalf("compare result = %#v err=%v", result, err)
	}
	if _, err := sa.CompareGroups(nil, []float64{1}); err == nil || !strings.Contains(err.Error(), "empty groups") {
		t.Fatalf("compare error = %v", err)
	}
}

func TestStatisticalAnalyzerDefaultsAndSmallHelpers(t *testing.T) {
	sa := NewStatisticalAnalyzer(-1, -1)
	if got := sa.calculatePercentile([]float64{7}, 50); got != 7 {
		t.Fatalf("single percentile = %v", got)
	}
	if got := sa.calculatePercentile([]float64{1, 2, 3}, -1); got != 0 {
		t.Fatalf("invalid percentile = %v", got)
	}
	if got := sa.interpretEffectSize(0.1); got != "negligible" {
		t.Fatalf("small effect = %q", got)
	}
	if got := sa.interpretEffectSize(0.3); got != "small" {
		t.Fatalf("small effect = %q", got)
	}
	if got := sa.interpretEffectSize(0.6); got != "medium" {
		t.Fatalf("medium effect = %q", got)
	}
	if got := sa.interpretEffectSize(0.9); got != "large" {
		t.Fatalf("large effect = %q", got)
	}
}
