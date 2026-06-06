package metrics

import (
	"fmt"
	"math"
	"sort"
)

// StatisticalAnalyzer provides advanced statistical analysis for prompt engineering evaluations
type StatisticalAnalyzer struct {
	confidenceLevel float64
	minSampleSize   int
}

// NewStatisticalAnalyzer creates a new statistical analyzer
func NewStatisticalAnalyzer(confidenceLevel float64, minSampleSize int) *StatisticalAnalyzer {
	if confidenceLevel <= 0 || confidenceLevel >= 1 {
		confidenceLevel = 0.95 // Default 95% confidence
	}
	if minSampleSize <= 0 {
		minSampleSize = 10 // Default minimum sample size
	}

	return &StatisticalAnalyzer{
		confidenceLevel: confidenceLevel,
		minSampleSize:   minSampleSize,
	}
}

// StatisticalSummary contains comprehensive statistical metrics
type StatisticalSummary struct {
	Count              int                `json:"count"`
	Mean               float64            `json:"mean"`
	Median             float64            `json:"median"`
	Mode               []float64          `json:"mode"`
	StandardDeviation  float64            `json:"standard_deviation"`
	Variance           float64            `json:"variance"`
	Skewness           float64            `json:"skewness"`
	Kurtosis           float64            `json:"kurtosis"`
	Min                float64            `json:"min"`
	Max                float64            `json:"max"`
	Range              float64            `json:"range"`
	Q1                 float64            `json:"q1"`
	Q3                 float64            `json:"q3"`
	IQR                float64            `json:"iqr"`
	Percentiles        map[string]float64 `json:"percentiles"`
	ConfidenceInterval ConfidenceInterval `json:"confidence_interval"`
	Distribution       DistributionInfo   `json:"distribution"`
	Outliers           []float64          `json:"outliers"`
}

// ConfidenceInterval represents a statistical confidence interval
type ConfidenceInterval struct {
	Level         float64 `json:"level"`
	LowerBound    float64 `json:"lower_bound"`
	UpperBound    float64 `json:"upper_bound"`
	MarginOfError float64 `json:"margin_of_error"`
}

// DistributionInfo contains information about data distribution
type DistributionInfo struct {
	IsNormal         bool    `json:"is_normal"`
	Normality        float64 `json:"normality_pvalue"`
	ShapiroWilk      float64 `json:"shapiro_wilk_pvalue"`
	DistributionType string  `json:"distribution_type"`
}

// HypothesisTestResult contains results of hypothesis testing
type HypothesisTestResult struct {
	TestName       string  `json:"test_name"`
	Statistic      float64 `json:"statistic"`
	PValue         float64 `json:"p_value"`
	CriticalValue  float64 `json:"critical_value"`
	IsSignificant  bool    `json:"is_significant"`
	EffectSize     float64 `json:"effect_size"`
	PowerAnalysis  float64 `json:"power_analysis"`
	Interpretation string  `json:"interpretation"`
}

// ComparisonResult contains results of comparing two groups
type ComparisonResult struct {
	Group1Summary     StatisticalSummary   `json:"group1_summary"`
	Group2Summary     StatisticalSummary   `json:"group2_summary"`
	TTest             HypothesisTestResult `json:"t_test"`
	MannWhitneyU      HypothesisTestResult `json:"mann_whitney_u"`
	KolmogorovSmirnov HypothesisTestResult `json:"kolmogorov_smirnov"`
	EffectSize        EffectSizeAnalysis   `json:"effect_size"`
	Recommendation    string               `json:"recommendation"`
}

// EffectSizeAnalysis contains effect size calculations
type EffectSizeAnalysis struct {
	CohensD        float64 `json:"cohens_d"`
	GlassesD       float64 `json:"glasses_d"`
	HedgesG        float64 `json:"hedges_g"`
	CliffsDelta    float64 `json:"cliffs_delta"`
	Interpretation string  `json:"interpretation"`
}

// ABTestResult contains A/B test analysis results
type ABTestResult struct {
	SampleSizeA         int                  `json:"sample_size_a"`
	SampleSizeB         int                  `json:"sample_size_b"`
	ConversionRateA     float64              `json:"conversion_rate_a"`
	ConversionRateB     float64              `json:"conversion_rate_b"`
	RelativeImprovement float64              `json:"relative_improvement"`
	StatisticalTest     HypothesisTestResult `json:"statistical_test"`
	ConfidenceInterval  ConfidenceInterval   `json:"confidence_interval"`
	MinimumDetectable   float64              `json:"minimum_detectable_effect"`
	PowerAnalysis       PowerAnalysis        `json:"power_analysis"`
	Recommendation      string               `json:"recommendation"`
}

// PowerAnalysis contains statistical power analysis
type PowerAnalysis struct {
	CurrentPower       float64 `json:"current_power"`
	RequiredSampleSize int     `json:"required_sample_size"`
	DetectedEffectSize float64 `json:"detected_effect_size"`
	Recommendation     string  `json:"recommendation"`
}

// CalculateStatisticalSummary computes comprehensive statistical summary
func (sa *StatisticalAnalyzer) CalculateStatisticalSummary(data []float64) (*StatisticalSummary, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty dataset")
	}

	if len(data) < sa.minSampleSize {
		return nil, fmt.Errorf("sample size %d is below minimum %d", len(data), sa.minSampleSize)
	}

	sortedData := make([]float64, len(data))
	copy(sortedData, data)
	sort.Float64s(sortedData)

	summary := &StatisticalSummary{
		Count: len(data),
	}

	// Basic statistics
	summary.Mean = sa.calculateMean(data)
	summary.Median = sa.calculateMedian(sortedData)
	summary.Mode = sa.calculateMode(data)
	summary.StandardDeviation = sa.calculateStandardDeviation(data, summary.Mean)
	summary.Variance = summary.StandardDeviation * summary.StandardDeviation
	summary.Min = sortedData[0]
	summary.Max = sortedData[len(sortedData)-1]
	summary.Range = summary.Max - summary.Min

	// Quartiles and IQR
	summary.Q1 = sa.calculatePercentile(sortedData, 25)
	summary.Q3 = sa.calculatePercentile(sortedData, 75)
	summary.IQR = summary.Q3 - summary.Q1

	// Percentiles
	summary.Percentiles = map[string]float64{
		"5th":  sa.calculatePercentile(sortedData, 5),
		"10th": sa.calculatePercentile(sortedData, 10),
		"25th": summary.Q1,
		"50th": summary.Median,
		"75th": summary.Q3,
		"90th": sa.calculatePercentile(sortedData, 90),
		"95th": sa.calculatePercentile(sortedData, 95),
		"99th": sa.calculatePercentile(sortedData, 99),
	}

	// Shape statistics
	summary.Skewness = sa.calculateSkewness(data, summary.Mean, summary.StandardDeviation)
	summary.Kurtosis = sa.calculateKurtosis(data, summary.Mean, summary.StandardDeviation)

	// Confidence interval
	summary.ConfidenceInterval = sa.calculateConfidenceInterval(data, summary.Mean, summary.StandardDeviation, sa.confidenceLevel)

	// Distribution analysis
	summary.Distribution = sa.analyzeDistribution(data)

	// Outlier detection
	summary.Outliers = sa.detectOutliers(sortedData, summary.Q1, summary.Q3, summary.IQR)

	return summary, nil
}

// PerformTTest performs Student's t-test
func (sa *StatisticalAnalyzer) PerformTTest(group1, group2 []float64, pairedTest bool) (*HypothesisTestResult, error) {
	if len(group1) == 0 || len(group2) == 0 {
		return nil, fmt.Errorf("empty groups")
	}

	if pairedTest && len(group1) != len(group2) {
		return nil, fmt.Errorf("paired test requires equal sample sizes")
	}

	return nil, fmt.Errorf("t-test statistical analysis is not yet implemented")
}

// PerformABTest performs comprehensive A/B test analysis
func (sa *StatisticalAnalyzer) PerformABTest(successesA, trialsA, successesB, trialsB int) (*ABTestResult, error) {
	if trialsA <= 0 || trialsB <= 0 {
		return nil, fmt.Errorf("invalid trial counts")
	}

	if successesA > trialsA || successesB > trialsB {
		return nil, fmt.Errorf("successes cannot exceed trials")
	}

	return nil, fmt.Errorf("A/B test statistical analysis is not yet implemented")
}

// CompareGroups performs comprehensive comparison between two groups
func (sa *StatisticalAnalyzer) CompareGroups(group1, group2 []float64) (*ComparisonResult, error) {
	if len(group1) == 0 || len(group2) == 0 {
		return nil, fmt.Errorf("empty groups")
	}

	// Calculate summaries for both groups
	if _, err := sa.CalculateStatisticalSummary(group1); err != nil {
		return nil, fmt.Errorf("failed to calculate summary for group 1: %w", err)
	}

	if _, err := sa.CalculateStatisticalSummary(group2); err != nil {
		return nil, fmt.Errorf("failed to calculate summary for group 2: %w", err)
	}

	return nil, fmt.Errorf("group comparison statistical tests are not yet implemented")
}

// Helper methods for statistical calculations

func (sa *StatisticalAnalyzer) calculateMean(data []float64) float64 {
	sum := 0.0
	for _, v := range data {
		sum += v
	}
	return sum / float64(len(data))
}

func (sa *StatisticalAnalyzer) calculateMedian(sortedData []float64) float64 {
	n := len(sortedData)
	if n%2 == 0 {
		return (sortedData[n/2-1] + sortedData[n/2]) / 2
	}
	return sortedData[n/2]
}

func (sa *StatisticalAnalyzer) calculateMode(data []float64) []float64 {
	frequency := make(map[float64]int)
	for _, v := range data {
		frequency[v]++
	}

	maxFreq := 0
	for _, freq := range frequency {
		if freq > maxFreq {
			maxFreq = freq
		}
	}

	var modes []float64
	for value, freq := range frequency {
		if freq == maxFreq && maxFreq > 1 {
			modes = append(modes, value)
		}
	}

	sort.Float64s(modes)
	return modes
}

func (sa *StatisticalAnalyzer) calculateStandardDeviation(data []float64, mean float64) float64 {
	sum := 0.0
	for _, v := range data {
		diff := v - mean
		sum += diff * diff
	}
	variance := sum / float64(len(data)-1) // Sample standard deviation
	return math.Sqrt(variance)
}

func (sa *StatisticalAnalyzer) calculatePercentile(sortedData []float64, percentile float64) float64 {
	if percentile < 0 || percentile > 100 {
		return 0
	}

	n := len(sortedData)
	if n == 1 {
		return sortedData[0]
	}

	index := (percentile / 100.0) * float64(n-1)
	lower := int(math.Floor(index))
	upper := int(math.Ceil(index))

	if lower == upper {
		return sortedData[lower]
	}

	weight := index - float64(lower)
	return sortedData[lower]*(1-weight) + sortedData[upper]*weight
}

func (sa *StatisticalAnalyzer) calculateSkewness(data []float64, mean, stdDev float64) float64 {
	n := float64(len(data))
	sum := 0.0

	for _, v := range data {
		standardized := (v - mean) / stdDev
		sum += standardized * standardized * standardized
	}

	return (n / ((n - 1) * (n - 2))) * sum
}

func (sa *StatisticalAnalyzer) calculateKurtosis(data []float64, mean, stdDev float64) float64 {
	n := float64(len(data))
	sum := 0.0

	for _, v := range data {
		standardized := (v - mean) / stdDev
		sum += standardized * standardized * standardized * standardized
	}

	kurtosis := (n*(n+1)/((n-1)*(n-2)*(n-3)))*sum - (3*(n-1)*(n-1))/((n-2)*(n-3))
	return kurtosis
}

func (sa *StatisticalAnalyzer) calculateConfidenceInterval(data []float64, mean, stdDev, level float64) ConfidenceInterval {
	return ConfidenceInterval{
		Level:         level,
		LowerBound:    math.NaN(),
		UpperBound:    math.NaN(),
		MarginOfError: math.NaN(),
	}
}

func (sa *StatisticalAnalyzer) analyzeDistribution(data []float64) DistributionInfo {
	return DistributionInfo{
		IsNormal:         false,
		Normality:        math.NaN(),
		ShapiroWilk:      math.NaN(),
		DistributionType: "unknown",
	}
}

func (sa *StatisticalAnalyzer) detectOutliers(sortedData []float64, q1, q3, iqr float64) []float64 {
	lowerFence := q1 - 1.5*iqr
	upperFence := q3 + 1.5*iqr

	var outliers []float64
	for _, v := range sortedData {
		if v < lowerFence || v > upperFence {
			outliers = append(outliers, v)
		}
	}

	return outliers
}

func (sa *StatisticalAnalyzer) calculateCohensD(group1, group2 []float64) float64 {
	mean1 := sa.calculateMean(group1)
	mean2 := sa.calculateMean(group2)
	std1 := sa.calculateStandardDeviation(group1, mean1)
	std2 := sa.calculateStandardDeviation(group2, mean2)

	n1, n2 := float64(len(group1)), float64(len(group2))

	// Pooled standard deviation
	pooledStd := math.Sqrt(((n1-1)*std1*std1 + (n2-1)*std2*std2) / (n1 + n2 - 2))

	return (mean1 - mean2) / pooledStd
}

func (sa *StatisticalAnalyzer) calculateEffectSizes(group1, group2 []float64) EffectSizeAnalysis {
	cohensD := sa.calculateCohensD(group1, group2)

	mean1 := sa.calculateMean(group1)
	mean2 := sa.calculateMean(group2)
	_ = sa.calculateStandardDeviation(group1, mean1) // std1 - currently unused
	std2 := sa.calculateStandardDeviation(group2, mean2)

	// Glass's Delta (uses control group standard deviation)
	glassesD := (mean1 - mean2) / std2

	// Hedges' g (bias-corrected Cohen's d)
	n1, n2 := float64(len(group1)), float64(len(group2))
	correction := 1 - (3 / (4*(n1+n2) - 9))
	hedgesG := cohensD * correction

	// Cliff's Delta (non-parametric effect size)
	cliffsDelta := sa.calculateCliffsDelta(group1, group2)

	interpretation := sa.interpretEffectSize(math.Abs(cohensD))

	return EffectSizeAnalysis{
		CohensD:        cohensD,
		GlassesD:       glassesD,
		HedgesG:        hedgesG,
		CliffsDelta:    cliffsDelta,
		Interpretation: interpretation,
	}
}

func (sa *StatisticalAnalyzer) calculateCliffsDelta(group1, group2 []float64) float64 {
	greater := 0
	less := 0

	for _, x := range group1 {
		for _, y := range group2 {
			if x > y {
				greater++
			} else if x < y {
				less++
			}
		}
	}

	total := len(group1) * len(group2)
	return float64(greater-less) / float64(total)
}

// Interpretation methods

func (sa *StatisticalAnalyzer) interpretTTestResult(pValue, effectSize float64, isSignificant bool) string {
	if !isSignificant {
		return "No statistically significant difference detected"
	}

	effectMagnitude := sa.interpretEffectSize(math.Abs(effectSize))
	return fmt.Sprintf("Statistically significant difference detected (p=%.4f) with %s effect size", pValue, effectMagnitude)
}

func (sa *StatisticalAnalyzer) interpretEffectSize(absEffectSize float64) string {
	if absEffectSize < 0.2 {
		return "negligible"
	} else if absEffectSize < 0.5 {
		return "small"
	} else if absEffectSize < 0.8 {
		return "medium"
	} else {
		return "large"
	}
}

func (sa *StatisticalAnalyzer) interpretZTestResult(pValue float64, isSignificant bool, difference float64) string {
	if !isSignificant {
		return "No statistically significant difference in proportions"
	}

	direction := "increase"
	if difference < 0 {
		direction = "decrease"
	}

	return fmt.Sprintf("Statistically significant %s in conversion rate (p=%.4f)", direction, pValue)
}

// Additional helper methods for A/B testing

func (sa *StatisticalAnalyzer) confidenceIntervalProportions(pA, pB float64, nA, nB int) ConfidenceInterval {
	diff := pB - pA
	se := math.Sqrt(pA*(1-pA)/float64(nA) + pB*(1-pB)/float64(nB))
	z := 1.96 // 95% confidence

	margin := z * se

	return ConfidenceInterval{
		Level:         0.95,
		LowerBound:    diff - margin,
		UpperBound:    diff + margin,
		MarginOfError: margin,
	}
}

func (sa *StatisticalAnalyzer) generateABTestRecommendation(test *HypothesisTestResult, relativeImprovement, power float64) string {
	if !test.IsSignificant {
		return "No statistically significant difference detected. Consider running the test longer or increasing sample size."
	}

	if power < 0.8 {
		return fmt.Sprintf("Significant result detected (%.2f%% relative improvement), but statistical power is low (%.2f). Consider increasing sample size for more reliable results.", relativeImprovement*100, power)
	}

	return fmt.Sprintf("Statistically significant improvement of %.2f%% detected with adequate statistical power (%.2f).", relativeImprovement*100, power)
}

func (sa *StatisticalAnalyzer) generateComparisonRecommendation(tTest *HypothesisTestResult, mannWhitney HypothesisTestResult, effectSize EffectSizeAnalysis) string {
	if !tTest.IsSignificant {
		return "No statistically significant difference detected between groups."
	}

	return fmt.Sprintf("Statistically significant difference detected with %s effect size. Consider practical significance of the %s effect.", effectSize.Interpretation, effectSize.Interpretation)
}
