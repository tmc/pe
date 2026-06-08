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
	if summary.ConfidenceInterval.Level != 0.95 || math.IsNaN(summary.ConfidenceInterval.MarginOfError) || summary.ConfidenceInterval.LowerBound >= summary.ConfidenceInterval.UpperBound {
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

	result, err := sa.PerformTTest(group1, group2, false)
	if err != nil {
		t.Fatalf("independent result = %#v err=%v", result, err)
	}
	if result.TestName != "welch_t_test" || result.Statistic <= 8 || result.PValue >= 0.001 || !result.IsSignificant {
		t.Fatalf("independent result = %#v", result)
	}
	if result.EffectSize <= 4 {
		t.Fatalf("effect size = %v, want large positive", result.EffectSize)
	}

	paired, err := sa.PerformTTest(group1, []float64{2, 3, 3, 5, 6, 8}, true)
	if err != nil {
		t.Fatalf("paired result = %#v err=%v", result, err)
	}
	if paired.Statistic <= 10 || paired.PValue >= 0.001 || !paired.IsSignificant {
		t.Fatalf("paired result = %#v", paired)
	}

	if _, err := sa.PerformTTest(nil, group2, false); err == nil {
		t.Fatalf("empty groups did not error")
	}
	if _, err := sa.PerformTTest([]float64{1}, group2, false); err == nil || !strings.Contains(err.Error(), "below minimum") {
		t.Fatalf("tiny group error = %v", err)
	}
	if _, err := sa.PerformTTest([]float64{math.NaN(), 1}, group2, false); err == nil || !strings.Contains(err.Error(), "non-finite") {
		t.Fatalf("nan error = %v", err)
	}
	if _, err := sa.PerformTTest([]float64{1, 1, 1}, []float64{2, 2, 2}, false); err == nil || !strings.Contains(err.Error(), "zero variance") {
		t.Fatalf("zero variance error = %v", err)
	}
	if _, err := sa.PerformTTest(group1, group2[:3], true); err == nil || !strings.Contains(err.Error(), "equal sample sizes") {
		t.Fatalf("paired size error = %v", err)
	}
}

func TestStatisticalAnalyzerABTest(t *testing.T) {
	sa := NewStatisticalAnalyzer(0.95, 3)
	result, err := sa.PerformABTest(60, 100, 75, 100)
	if err != nil {
		t.Fatalf("ab result = %#v err=%v", result, err)
	}
	if result.SampleSizeA != 100 || result.SampleSizeB != 100 {
		t.Fatalf("sample sizes = %#v", result)
	}
	if !near(result.ConversionRateA, 0.60, 1e-12) || !near(result.ConversionRateB, 0.75, 1e-12) {
		t.Fatalf("conversion rates = %#v", result)
	}
	if !near(result.RelativeImprovement, 0.25, 1e-12) {
		t.Fatalf("relative improvement = %v", result.RelativeImprovement)
	}
	if result.StatisticalTest.TestName != "two_proportion_z_test" || result.StatisticalTest.PValue <= 0 || result.StatisticalTest.PValue >= 0.05 {
		t.Fatalf("statistical test = %#v", result.StatisticalTest)
	}
	if result.ConfidenceInterval.UpperBound <= result.ConfidenceInterval.LowerBound {
		t.Fatalf("confidence interval = %#v", result.ConfidenceInterval)
	}

	if _, err := sa.PerformABTest(1, 0, 1, 2); err == nil || !strings.Contains(err.Error(), "invalid trial") {
		t.Fatalf("trial error = %v", err)
	}
	if _, err := sa.PerformABTest(3, 2, 1, 2); err == nil || !strings.Contains(err.Error(), "successes") {
		t.Fatalf("success error = %v", err)
	}
	if _, err := sa.PerformABTest(-1, 2, 1, 2); err == nil || !strings.Contains(err.Error(), "successes") {
		t.Fatalf("negative success error = %v", err)
	}
	if _, err := sa.PerformABTest(0, 10, 0, 10); err == nil || !strings.Contains(err.Error(), "zero standard error") {
		t.Fatalf("zero standard error = %v", err)
	}
}

func TestStatisticalAnalyzerCompareGroups(t *testing.T) {
	sa := NewStatisticalAnalyzer(0.95, 3)
	result, err := sa.CompareGroups(
		[]float64{10, 11, 12, 13, 14, 15},
		[]float64{1, 2, 3, 4, 5, 6},
	)
	if err != nil {
		t.Fatalf("compare result = %#v err=%v", result, err)
	}
	if result.Group1Summary.Count != 6 || result.Group2Summary.Count != 6 {
		t.Fatalf("summaries = %#v", result)
	}
	if !result.TTest.IsSignificant || result.MannWhitneyU.PValue >= 0.05 || result.KolmogorovSmirnov.Statistic != 1 {
		t.Fatalf("tests = %#v", result)
	}
	if result.EffectSize.Interpretation != "large" || result.Recommendation == "" {
		t.Fatalf("effect/recommendation = %#v %q", result.EffectSize, result.Recommendation)
	}
	tied, err := sa.CompareGroups([]float64{1, 2, 2, 3}, []float64{2, 2, 4, 5})
	if err != nil {
		t.Fatalf("tied compare: %v", err)
	}
	if tied.MannWhitneyU.Statistic <= 0 || tied.MannWhitneyU.PValue <= 0 {
		t.Fatalf("tied Mann-Whitney = %#v", tied.MannWhitneyU)
	}
	if _, err := sa.CompareGroups(nil, []float64{1}); err == nil || !strings.Contains(err.Error(), "empty groups") {
		t.Fatalf("compare error = %v", err)
	}
}

func near(got, want, tolerance float64) bool {
	return math.Abs(got-want) <= tolerance
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
