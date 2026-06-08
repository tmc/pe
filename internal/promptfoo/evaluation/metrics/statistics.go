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

// PerformTTest performs a Welch or paired t-test with a normal approximation
// for the two-sided p-value.
func (sa *StatisticalAnalyzer) PerformTTest(group1, group2 []float64, pairedTest bool) (*HypothesisTestResult, error) {
	if err := validateSamples(group1, group2); err != nil {
		return nil, err
	}

	if pairedTest && len(group1) != len(group2) {
		return nil, fmt.Errorf("paired test requires equal sample sizes")
	}

	var statistic, effectSize float64
	if pairedTest {
		diffs := make([]float64, len(group1))
		for i := range group1 {
			diffs[i] = group1[i] - group2[i]
		}
		mean := sa.calculateMean(diffs)
		std := sa.calculateStandardDeviation(diffs, mean)
		if std == 0 {
			return nil, fmt.Errorf("paired differences have zero variance")
		}
		statistic = mean / (std / math.Sqrt(float64(len(diffs))))
		effectSize = mean / std
	} else {
		mean1 := sa.calculateMean(group1)
		mean2 := sa.calculateMean(group2)
		std1 := sa.calculateStandardDeviation(group1, mean1)
		std2 := sa.calculateStandardDeviation(group2, mean2)
		if std1 == 0 && std2 == 0 {
			return nil, fmt.Errorf("groups have zero variance")
		}
		se := math.Sqrt(std1*std1/float64(len(group1)) + std2*std2/float64(len(group2)))
		if se == 0 {
			return nil, fmt.Errorf("groups have zero standard error")
		}
		statistic = (mean1 - mean2) / se
		effectSize = sa.calculateCohensD(group1, group2)
	}

	pValue := twoSidedNormalP(statistic)
	isSignificant := pValue < alphaForConfidence(sa.confidenceLevel)

	return &HypothesisTestResult{
		TestName:       "welch_t_test",
		Statistic:      statistic,
		PValue:         pValue,
		CriticalValue:  normalCriticalValue(sa.confidenceLevel),
		IsSignificant:  isSignificant,
		EffectSize:     effectSize,
		PowerAnalysis:  approximatePower(math.Abs(effectSize), len(group1)+len(group2), sa.confidenceLevel),
		Interpretation: sa.interpretTTestResult(pValue, effectSize, isSignificant),
	}, nil
}

// PerformABTest performs two-proportion A/B test analysis with a normal
// approximation for the two-sided p-value.
func (sa *StatisticalAnalyzer) PerformABTest(successesA, trialsA, successesB, trialsB int) (*ABTestResult, error) {
	if trialsA <= 0 || trialsB <= 0 {
		return nil, fmt.Errorf("invalid trial counts")
	}

	if successesA < 0 || successesB < 0 || successesA > trialsA || successesB > trialsB {
		return nil, fmt.Errorf("successes cannot exceed trials")
	}

	pA := float64(successesA) / float64(trialsA)
	pB := float64(successesB) / float64(trialsB)
	pooled := float64(successesA+successesB) / float64(trialsA+trialsB)
	se := math.Sqrt(pooled * (1 - pooled) * (1/float64(trialsA) + 1/float64(trialsB)))
	if se == 0 {
		return nil, fmt.Errorf("proportions have zero standard error")
	}

	diff := pB - pA
	z := diff / se
	pValue := twoSidedNormalP(z)
	isSignificant := pValue < alphaForConfidence(sa.confidenceLevel)
	effectSize := cohensH(pA, pB)
	power := approximatePower(math.Abs(effectSize), trialsA+trialsB, sa.confidenceLevel)
	relativeImprovement := math.Inf(1)
	if pA != 0 {
		relativeImprovement = diff / pA
	}

	test := HypothesisTestResult{
		TestName:       "two_proportion_z_test",
		Statistic:      z,
		PValue:         pValue,
		CriticalValue:  normalCriticalValue(sa.confidenceLevel),
		IsSignificant:  isSignificant,
		EffectSize:     effectSize,
		PowerAnalysis:  power,
		Interpretation: sa.interpretZTestResult(pValue, isSignificant, diff),
	}

	requiredSampleSize := 0
	if effectSize != 0 {
		zAlpha := normalCriticalValue(sa.confidenceLevel)
		zPower := 0.84 // 80% power.
		requiredSampleSize = int(math.Ceil(2 * math.Pow((zAlpha+zPower)/math.Abs(effectSize), 2)))
	}

	return &ABTestResult{
		SampleSizeA:         trialsA,
		SampleSizeB:         trialsB,
		ConversionRateA:     pA,
		ConversionRateB:     pB,
		RelativeImprovement: relativeImprovement,
		StatisticalTest:     test,
		ConfidenceInterval:  sa.confidenceIntervalProportions(pA, pB, trialsA, trialsB),
		MinimumDetectable:   minimumDetectableEffect(trialsA, trialsB, sa.confidenceLevel),
		PowerAnalysis: PowerAnalysis{
			CurrentPower:       power,
			RequiredSampleSize: requiredSampleSize,
			DetectedEffectSize: effectSize,
			Recommendation:     powerRecommendation(power),
		},
		Recommendation: sa.generateABTestRecommendation(&test, relativeImprovement, power),
	}, nil
}

// CompareGroups performs comprehensive comparison between two groups
func (sa *StatisticalAnalyzer) CompareGroups(group1, group2 []float64) (*ComparisonResult, error) {
	if err := validateSamples(group1, group2); err != nil {
		return nil, err
	}

	group1Summary, err := sa.CalculateStatisticalSummary(group1)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate summary for group 1: %w", err)
	}

	group2Summary, err := sa.CalculateStatisticalSummary(group2)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate summary for group 2: %w", err)
	}

	tTest, err := sa.PerformTTest(group1, group2, false)
	if err != nil {
		return nil, fmt.Errorf("t-test: %w", err)
	}
	mannWhitney := sa.performMannWhitneyU(group1, group2)
	ks := sa.performKolmogorovSmirnov(group1, group2)
	effectSize := sa.calculateEffectSizes(group1, group2)

	return &ComparisonResult{
		Group1Summary:     *group1Summary,
		Group2Summary:     *group2Summary,
		TTest:             *tTest,
		MannWhitneyU:      mannWhitney,
		KolmogorovSmirnov: ks,
		EffectSize:        effectSize,
		Recommendation:    sa.generateComparisonRecommendation(tTest, mannWhitney, effectSize),
	}, nil
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
	if len(data) < 2 {
		return math.NaN()
	}
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
	if len(data) == 0 || math.IsNaN(stdDev) || math.IsInf(stdDev, 0) {
		return ConfidenceInterval{
			Level:         level,
			LowerBound:    math.NaN(),
			UpperBound:    math.NaN(),
			MarginOfError: math.NaN(),
		}
	}
	margin := normalCriticalValue(level) * stdDev / math.Sqrt(float64(len(data)))
	return ConfidenceInterval{
		Level:         level,
		LowerBound:    mean - margin,
		UpperBound:    mean + margin,
		MarginOfError: margin,
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
	if pooledStd == 0 {
		return 0
	}

	return (mean1 - mean2) / pooledStd
}

func (sa *StatisticalAnalyzer) calculateEffectSizes(group1, group2 []float64) EffectSizeAnalysis {
	cohensD := sa.calculateCohensD(group1, group2)

	mean1 := sa.calculateMean(group1)
	mean2 := sa.calculateMean(group2)
	_ = sa.calculateStandardDeviation(group1, mean1) // std1 - currently unused
	std2 := sa.calculateStandardDeviation(group2, mean2)

	// Glass's Delta (uses control group standard deviation)
	glassesD := 0.0
	if std2 != 0 {
		glassesD = (mean1 - mean2) / std2
	}

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
	if total == 0 {
		return 0
	}
	return float64(greater-less) / float64(total)
}

func (sa *StatisticalAnalyzer) performMannWhitneyU(group1, group2 []float64) HypothesisTestResult {
	u1, u2 := mannWhitneyU(group1, group2)
	u := math.Min(u1, u2)
	n1 := float64(len(group1))
	n2 := float64(len(group2))
	mean := n1 * n2 / 2
	std := math.Sqrt(n1 * n2 * (n1 + n2 + 1) / 12)
	z := 0.0
	pValue := 1.0
	if std != 0 {
		z = (u - mean) / std
		pValue = twoSidedNormalP(z)
	}
	isSignificant := pValue < alphaForConfidence(sa.confidenceLevel)
	effectSize := sa.calculateCliffsDelta(group1, group2)
	return HypothesisTestResult{
		TestName:       "mann_whitney_u",
		Statistic:      u,
		PValue:         pValue,
		CriticalValue:  normalCriticalValue(sa.confidenceLevel),
		IsSignificant:  isSignificant,
		EffectSize:     effectSize,
		PowerAnalysis:  approximatePower(math.Abs(effectSize), len(group1)+len(group2), sa.confidenceLevel),
		Interpretation: sa.interpretTTestResult(pValue, effectSize, isSignificant),
	}
}

func (sa *StatisticalAnalyzer) performKolmogorovSmirnov(group1, group2 []float64) HypothesisTestResult {
	statistic := kolmogorovSmirnovD(group1, group2)
	n1 := float64(len(group1))
	n2 := float64(len(group2))
	effectiveN := n1 * n2 / (n1 + n2)
	pValue := math.Min(1, 2*math.Exp(-2*effectiveN*statistic*statistic))
	isSignificant := pValue < alphaForConfidence(sa.confidenceLevel)
	return HypothesisTestResult{
		TestName:       "kolmogorov_smirnov",
		Statistic:      statistic,
		PValue:         pValue,
		CriticalValue:  1.36 * math.Sqrt((n1+n2)/(n1*n2)),
		IsSignificant:  isSignificant,
		EffectSize:     statistic,
		PowerAnalysis:  approximatePower(statistic, len(group1)+len(group2), sa.confidenceLevel),
		Interpretation: sa.interpretTTestResult(pValue, statistic, isSignificant),
	}
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

func validateSamples(groups ...[]float64) error {
	for _, group := range groups {
		if len(group) == 0 {
			return fmt.Errorf("empty groups")
		}
		if len(group) < 2 {
			return fmt.Errorf("sample size %d is below minimum 2", len(group))
		}
		for _, v := range group {
			if math.IsNaN(v) || math.IsInf(v, 0) {
				return fmt.Errorf("sample contains non-finite value")
			}
		}
	}
	return nil
}

func alphaForConfidence(confidence float64) float64 {
	if confidence <= 0 || confidence >= 1 {
		confidence = 0.95
	}
	return 1 - confidence
}

func normalCriticalValue(confidence float64) float64 {
	switch {
	case confidence >= 0.999:
		return 3.291
	case confidence >= 0.99:
		return 2.576
	case confidence >= 0.95:
		return 1.96
	case confidence >= 0.90:
		return 1.645
	default:
		return 1.96
	}
}

func twoSidedNormalP(statistic float64) float64 {
	return math.Erfc(math.Abs(statistic) / math.Sqrt2)
}

func approximatePower(effectSize float64, sampleSize int, confidence float64) float64 {
	if effectSize <= 0 || sampleSize <= 0 {
		return 0
	}
	zAlpha := normalCriticalValue(confidence)
	z := effectSize*math.Sqrt(float64(sampleSize)/2) - zAlpha
	power := 0.5 * math.Erfc(-z/math.Sqrt2)
	if power < 0 {
		return 0
	}
	if power > 1 {
		return 1
	}
	return power
}

func cohensH(pA, pB float64) float64 {
	return 2 * (math.Asin(math.Sqrt(pB)) - math.Asin(math.Sqrt(pA)))
}

func minimumDetectableEffect(trialsA, trialsB int, confidence float64) float64 {
	n := math.Min(float64(trialsA), float64(trialsB))
	if n <= 0 {
		return math.NaN()
	}
	zAlpha := normalCriticalValue(confidence)
	zPower := 0.84 // 80% power.
	return (zAlpha + zPower) * math.Sqrt(2/n)
}

func powerRecommendation(power float64) string {
	if power >= 0.8 {
		return "Current sample size has adequate approximate power."
	}
	return "Current sample size has low approximate power; collect more samples before relying on the result."
}

type rankedValue struct {
	value float64
	group int
	rank  float64
}

func mannWhitneyU(group1, group2 []float64) (float64, float64) {
	ranked := make([]rankedValue, 0, len(group1)+len(group2))
	for _, v := range group1 {
		ranked = append(ranked, rankedValue{value: v, group: 1})
	}
	for _, v := range group2 {
		ranked = append(ranked, rankedValue{value: v, group: 2})
	}
	sort.Slice(ranked, func(i, j int) bool {
		return ranked[i].value < ranked[j].value
	})
	for i := 0; i < len(ranked); {
		j := i + 1
		for j < len(ranked) && ranked[j].value == ranked[i].value {
			j++
		}
		rank := (float64(i+1) + float64(j)) / 2
		for k := i; k < j; k++ {
			ranked[k].rank = rank
		}
		i = j
	}

	rank1 := 0.0
	rank2 := 0.0
	for _, v := range ranked {
		if v.group == 1 {
			rank1 += v.rank
		} else {
			rank2 += v.rank
		}
	}
	n1 := float64(len(group1))
	n2 := float64(len(group2))
	u1 := rank1 - n1*(n1+1)/2
	u2 := rank2 - n2*(n2+1)/2
	return u1, u2
}

func kolmogorovSmirnovD(group1, group2 []float64) float64 {
	a := append([]float64(nil), group1...)
	b := append([]float64(nil), group2...)
	sort.Float64s(a)
	sort.Float64s(b)

	i := 0
	j := 0
	maxDiff := 0.0
	for i < len(a) || j < len(b) {
		var x float64
		switch {
		case j >= len(b):
			x = a[i]
		case i >= len(a):
			x = b[j]
		case a[i] <= b[j]:
			x = a[i]
		default:
			x = b[j]
		}
		for i < len(a) && a[i] <= x {
			i++
		}
		for j < len(b) && b[j] <= x {
			j++
		}
		diff := math.Abs(float64(i)/float64(len(a)) - float64(j)/float64(len(b)))
		if diff > maxDiff {
			maxDiff = diff
		}
	}
	return maxDiff
}
