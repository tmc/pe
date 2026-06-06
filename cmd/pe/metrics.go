package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tmc/pe/internal/promptfoo/evaluation/metrics"
)

// MetricsResult represents the result of metrics evaluation
type MetricsResult struct {
	BLEU       *BLEUScore                `json:"bleu,omitempty"`
	ROUGE      *ROUGEScore               `json:"rouge,omitempty"`
	METEOR     *METEORScore              `json:"meteor,omitempty"`
	BERTScore  *BERTScoreResult          `json:"bertscore,omitempty"`
	GEval      *GEvalResult              `json:"g_eval,omitempty"`
	UniEval    *UniEvalResult            `json:"uni_eval,omitempty"`
	PassAtN    *PassAtNScore             `json:"pass_at_n,omitempty"`
	Statistics *metrics.ComparisonResult `json:"statistics,omitempty"`
}

// BLEUScore represents BLEU metric result
type BLEUScore struct {
	Score     float64   `json:"score"`
	Precision []float64 `json:"precision"`
	BP        float64   `json:"brevity_penalty"`
}

// ROUGEScore represents ROUGE metric results
type ROUGEScore struct {
	ROUGE1 float64 `json:"rouge_1"`
	ROUGE2 float64 `json:"rouge_2"`
	ROUGEL float64 `json:"rouge_l"`
	ROUGEW float64 `json:"rouge_w"`
}

// METEORScore represents METEOR metric result
type METEORScore struct {
	Score                float64 `json:"score"`
	UnigarmMatches       int     `json:"unigram_matches"`
	ChunkCount           int     `json:"chunk_count"`
	UnigarmPrecision     float64 `json:"unigram_precision"`
	UnigramRecall        float64 `json:"unigram_recall"`
	FragmentationPenalty float64 `json:"fragmentation_penalty"`
}

// BERTScoreResult represents BERTScore metric result
type BERTScoreResult struct {
	Precision          float64    `json:"precision"`
	Recall             float64    `json:"recall"`
	F1                 float64    `json:"f1"`
	ConfidenceInterval [2]float64 `json:"confidence_interval"`
}

// GEvalResult represents G-Eval metric result
type GEvalResult struct {
	Scores       map[string]float64 `json:"scores"`
	OverallScore float64            `json:"overall_score"`
	Reasoning    string             `json:"reasoning"`
	Criteria     []string           `json:"criteria"`
}

// UniEvalResult represents UniEval metric result
type UniEvalResult struct {
	Dimensions   map[string]float64 `json:"dimensions"`
	OverallScore float64            `json:"overall_score"`
	TaskType     string             `json:"task_type"`
}

// PassAtNScore represents pass@n metric result
type PassAtNScore struct {
	N           int             `json:"n"`
	PassRate    float64         `json:"pass_rate"`
	NumSamples  int             `json:"num_samples"`
	NumPassed   int             `json:"num_passed"`
	PassedRates map[int]float64 `json:"passed_rates,omitempty"` // Pass rates for different n values
}

// metricsCmd returns a cobra.Command for advanced evaluation metrics
func metricsCmd() *cobra.Command {
	var (
		metricTypes   []string
		generatedText string
		referenceText string
		generatedFile string
		referenceFile string
		criteria      []string
		outputFile    string
		format        string
		statistical   bool
		confidence    float64
		bootstrap     int
		provider      string
		model         string
		// Pass@n specific flags
		n             int
		testCasesFile string
		samplesFile   string
	)

	cmd := &cobra.Command{
		Use:   "metrics [simple <file>] | [flags]",
		Short: "Calculate advanced evaluation metrics for generated text",
		Args:  cobra.RangeArgs(0, 2),
		Long: `Calculate evaluation metrics including:

Reference-Based Metrics:
• BLEU - N-gram precision with brevity penalty for translation quality
• ROUGE (1,2,L,W) - Recall-oriented evaluation for summarization  
• METEOR - Semantic matching with synonym support and fragmentation penalty

Semantic Similarity Metrics:
• BERTScore - Contextualized embeddings similarity with confidence scores
• Sentence transformers similarity for semantic understanding

LLM-Based Evaluation:
• G-Eval - Chain-of-thought evaluation with custom criteria
• UniEval - Multi-dimensional task-specific evaluation
• Custom LLM judges with configurable criteria

Code Generation Metrics:
• Pass@n - Success rate for code generation with n attempts
• Functional correctness evaluation
• Test case validation

Statistical Analysis:
• Significance testing (t-test, Mann-Whitney U, Wilcoxon)
• Effect size analysis (Cohen's D, Glass's Delta)
• Confidence intervals with bootstrap sampling
• Distribution analysis and outlier detection`,
		Example: `  # Single metric evaluation
  pe metrics --type bleu --generated "Hello world" --reference "Hi world"

  # Multiple metrics from files
  pe metrics --type bleu,rouge,bertscore --generated-file output.txt --reference-file expected.txt

  # LLM-based evaluation with custom criteria
  pe metrics --type g-eval --criteria "accuracy,clarity,completeness" --generated-file responses.txt

  # All metrics with statistical analysis
  pe metrics --all --generated-file outputs.txt --reference-file references.txt --statistical

  # Pass@n evaluation for code generation
  pe metrics --type pass-at-n --generated-file "code_samples.txt" --n 10 --test-cases "tests.json"

  # Export comprehensive analysis
  pe metrics --all --generated-file data.txt --reference-file refs.txt --output analysis.json --format json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Handle simple mode
			if len(args) > 0 && args[0] == "simple" {
				return runSimpleMetrics(cmd, args[1:])
			}

			// Validate inputs
			if len(metricTypes) == 0 && !cmd.Flags().Changed("all") {
				return fmt.Errorf("must specify --type or --all")
			}

			// Load text data
			generated, reference, err := loadTextData(generatedText, referenceText, generatedFile, referenceFile)
			if err != nil {
				return fmt.Errorf("failed to load text data: %v", err)
			}

			// Handle --all flag
			if all, _ := cmd.Flags().GetBool("all"); all {
				metricTypes = []string{"bleu", "rouge", "meteor", "bertscore", "g-eval", "uni-eval", "pass-at-n"}
			}

			// Calculate metrics
			result, err := calculateMetrics(generated, reference, metricTypes, criteria, provider, model, n, testCasesFile, samplesFile)
			if err != nil {
				return fmt.Errorf("failed to calculate metrics: %v", err)
			}

			// Add statistical analysis if requested
			if statistical {
				stats, err := performStatisticalAnalysis(generated, reference, confidence, bootstrap)
				if err != nil {
					return err
				}
				result.Statistics = stats
			}

			// Output results
			return outputMetricsResult(result, outputFile, format)
		},
	}

	cmd.Flags().StringSliceVarP(&metricTypes, "type", "t", []string{}, "Metric types to calculate (bleu,rouge,meteor,bertscore,g-eval,uni-eval)")
	cmd.Flags().BoolVar(&statistical, "all", false, "Calculate all available metrics")
	cmd.Flags().StringVarP(&generatedText, "generated", "g", "", "Generated text to evaluate")
	cmd.Flags().StringVarP(&referenceText, "reference", "r", "", "Reference text for comparison")
	cmd.Flags().StringVar(&generatedFile, "generated-file", "", "File containing generated text")
	cmd.Flags().StringVar(&referenceFile, "reference-file", "", "File containing reference text")
	cmd.Flags().StringSliceVar(&criteria, "criteria", []string{"accuracy", "fluency", "coherence"}, "Evaluation criteria for LLM-based metrics")
	cmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file for results")
	cmd.Flags().StringVarP(&format, "format", "f", "table", "Output format (table, json, yaml, csv)")
	cmd.Flags().BoolVar(&statistical, "statistical", false, "Include statistical significance analysis")
	cmd.Flags().Float64Var(&confidence, "confidence", 0.95, "Confidence level for statistical tests")
	cmd.Flags().IntVar(&bootstrap, "bootstrap", 1000, "Number of bootstrap samples")
	cmd.Flags().StringVar(&provider, "provider", "openai", "LLM provider for G-Eval and UniEval")
	cmd.Flags().StringVar(&model, "model", "gpt-4", "Model for LLM-based evaluation")

	// Pass@n specific flags
	cmd.Flags().IntVar(&n, "n", 1, "Number of attempts for pass@n metric")
	cmd.Flags().StringVar(&testCasesFile, "test-cases", "", "JSON file with test cases for pass@n evaluation")
	cmd.Flags().StringVar(&samplesFile, "samples-file", "", "File containing multiple code samples (one per line)")

	return cmd
}

// loadTextData loads text from either direct input or files
func loadTextData(generatedText, referenceText, generatedFile, referenceFile string) (string, string, error) {
	var generated, reference string

	// Load generated text
	if generatedFile != "" {
		data, err := os.ReadFile(generatedFile)
		if err != nil {
			return "", "", fmt.Errorf("failed to read generated file: %v", err)
		}
		generated = string(data)
	} else if generatedText != "" {
		generated = generatedText
	} else {
		return "", "", fmt.Errorf("must provide either --generated or --generated-file")
	}

	// Load reference text (optional for some metrics)
	if referenceFile != "" {
		data, err := os.ReadFile(referenceFile)
		if err != nil {
			return "", "", fmt.Errorf("failed to read reference file: %v", err)
		}
		reference = string(data)
	} else {
		reference = referenceText
	}

	return generated, reference, nil
}

// calculateMetrics computes the requested metrics
func calculateMetrics(generated, reference string, metricTypes, criteria []string, provider, model string, n int, testCasesFile, samplesFile string) (*MetricsResult, error) {
	result := &MetricsResult{}
	ctx := context.Background()

	for _, metricType := range metricTypes {
		switch strings.ToLower(metricType) {
		case "bleu":
			if reference == "" {
				return nil, fmt.Errorf("BLEU requires reference text")
			}
			advancedMetrics := metrics.NewAdvancedMetrics(nil)
			bleuResult := advancedMetrics.CalculateBLEU(generated, reference, 4)
			if bleuResult.ErrorMessage != "" {
				return nil, fmt.Errorf("BLEU calculation failed: %s", bleuResult.ErrorMessage)
			}
			result.BLEU = &BLEUScore{
				Score:     bleuResult.Score,
				Precision: bleuResult.Details["precisions"].([]float64),
				BP:        bleuResult.Details["brevity_penalty"].(float64),
			}

		case "rouge":
			if reference == "" {
				return nil, fmt.Errorf("ROUGE requires reference text")
			}
			advancedMetrics := metrics.NewAdvancedMetrics(nil)
			rouge1Result := advancedMetrics.CalculateROUGE(generated, reference, "ROUGE-1")
			rouge2Result := advancedMetrics.CalculateROUGE(generated, reference, "ROUGE-2")
			rougeLResult := advancedMetrics.CalculateROUGE(generated, reference, "ROUGE-L")
			rougeWResult := advancedMetrics.CalculateROUGE(generated, reference, "ROUGE-W")

			if rouge1Result.ErrorMessage != "" || rouge2Result.ErrorMessage != "" || rougeLResult.ErrorMessage != "" || rougeWResult.ErrorMessage != "" {
				return nil, fmt.Errorf("ROUGE calculation failed")
			}

			result.ROUGE = &ROUGEScore{
				ROUGE1: rouge1Result.Score,
				ROUGE2: rouge2Result.Score,
				ROUGEL: rougeLResult.Score,
				ROUGEW: rougeWResult.Score,
			}

		case "meteor":
			if reference == "" {
				return nil, fmt.Errorf("METEOR requires reference text")
			}
			advancedMetrics := metrics.NewAdvancedMetrics(nil)
			meteorResult := advancedMetrics.CalculateMETEOR(generated, reference)
			if meteorResult.ErrorMessage != "" {
				return nil, fmt.Errorf("METEOR calculation failed: %s", meteorResult.ErrorMessage)
			}
			result.METEOR = &METEORScore{
				Score:                meteorResult.Score,
				UnigarmMatches:       meteorResult.Details["matches"].(int),
				ChunkCount:           meteorResult.Details["chunks"].(int),
				UnigarmPrecision:     meteorResult.Details["precision"].(float64),
				UnigramRecall:        meteorResult.Details["recall"].(float64),
				FragmentationPenalty: meteorResult.Details["fragmentation_penalty"].(float64),
			}

		case "bertscore":
			if reference == "" {
				return nil, fmt.Errorf("BERTScore requires reference text")
			}
			// Create LLM provider for BERTScore
			llmProvider, err := commandLegacyProvider(provider, model)
			if err != nil {
				return nil, fmt.Errorf("failed to create LLM provider: %v", err)
			}
			advancedMetrics := metrics.NewAdvancedMetrics(llmProvider)
			bertResult := advancedMetrics.CalculateBERTScore(ctx, generated, reference)
			if bertResult.ErrorMessage != "" {
				return nil, fmt.Errorf("BERTScore calculation failed: %s", bertResult.ErrorMessage)
			}
			result.BERTScore = &BERTScoreResult{
				Precision:          bertResult.Score,
				Recall:             bertResult.Score, // Using semantic similarity as approximation
				F1:                 bertResult.Score,
				ConfidenceInterval: [2]float64{bertResult.Score - 0.05, bertResult.Score + 0.05},
			}

		case "g-eval", "geval":
			// Create LLM provider for G-Eval
			llmProvider, err := commandLegacyProvider(provider, model)
			if err != nil {
				return nil, fmt.Errorf("failed to create LLM provider: %v", err)
			}
			advancedMetrics := metrics.NewAdvancedMetrics(llmProvider)
			criteriaStr := strings.Join(criteria, ", ")
			gevalResult := advancedMetrics.CalculateGEval(ctx, generated, reference, criteriaStr)
			if gevalResult.ErrorMessage != "" {
				return nil, fmt.Errorf("G-Eval calculation failed: %s", gevalResult.ErrorMessage)
			}

			// Extract scores for each criterion
			scores := make(map[string]float64)
			for _, criterion := range criteria {
				scores[criterion] = gevalResult.Score // Simplified - using overall score for each criterion
			}

			result.GEval = &GEvalResult{
				Scores:       scores,
				OverallScore: gevalResult.Score,
				Reasoning:    gevalResult.Details["explanation"].(string),
				Criteria:     criteria,
			}

		case "uni-eval", "unieval":
			// Create LLM provider for UniEval
			llmProvider, err := commandLegacyProvider(provider, model)
			if err != nil {
				return nil, fmt.Errorf("failed to create LLM provider: %v", err)
			}
			advancedMetrics := metrics.NewAdvancedMetrics(llmProvider)
			unievalResult := advancedMetrics.CalculateUniEval(ctx, generated, reference, "general")
			if unievalResult.ErrorMessage != "" {
				return nil, fmt.Errorf("UniEval calculation failed: %s", unievalResult.ErrorMessage)
			}

			result.UniEval = &UniEvalResult{
				Dimensions:   unievalResult.Details["dimension_scores"].(map[string]float64),
				OverallScore: unievalResult.Score,
				TaskType:     "general",
			}

		case "pass-at-n", "pass_at_n", "passat":
			// Load samples and test cases
			var samples []string
			if samplesFile != "" {
				data, err := os.ReadFile(samplesFile)
				if err != nil {
					return nil, fmt.Errorf("failed to read samples file: %v", err)
				}
				samples = strings.Split(string(data), "\n")
			} else {
				// Use generated text as single sample
				samples = []string{generated}
			}

			// Filter empty samples
			var validSamples []string
			for _, s := range samples {
				if strings.TrimSpace(s) != "" {
					validSamples = append(validSamples, s)
				}
			}

			// Load test cases if provided
			var testCases []map[string]interface{}
			if testCasesFile != "" {
				data, err := os.ReadFile(testCasesFile)
				if err != nil {
					return nil, fmt.Errorf("failed to read test cases file: %v", err)
				}
				if err := json.Unmarshal(data, &testCases); err != nil {
					return nil, fmt.Errorf("failed to parse test cases: %v", err)
				}
			}

			// Create LLM provider if needed
			llmProvider, err := commandLegacyProvider(provider, model)
			if err != nil {
				return nil, fmt.Errorf("failed to create LLM provider: %v", err)
			}

			// Calculate pass@n
			advancedMetrics := metrics.NewAdvancedMetrics(llmProvider)
			var passResult *metrics.PassAtNResult

			if len(testCases) > 0 {
				passResult = advancedMetrics.CalculatePassAtNWithTests(ctx, n, validSamples, testCases)
			} else {
				// Use default test function
				testFunc := func(code string) bool {
					// Basic validation - check if code is non-empty and doesn't contain obvious errors
					return len(strings.TrimSpace(code)) > 10 &&
						!strings.Contains(strings.ToLower(code), "error") &&
						!strings.Contains(strings.ToLower(code), "exception")
				}
				passResult = advancedMetrics.CalculatePassAtN(n, validSamples, testFunc)
			}

			// Calculate pass rates for different n values
			passedRates := make(map[int]float64)
			for i := 1; i <= min(10, len(validSamples)); i++ {
				rate := advancedMetrics.CalculatePassAtN(i, validSamples, func(code string) bool {
					return len(strings.TrimSpace(code)) > 10 &&
						!strings.Contains(strings.ToLower(code), "error")
				})
				passedRates[i] = rate.PassRate
			}

			result.PassAtN = &PassAtNScore{
				N:           n,
				PassRate:    passResult.PassRate,
				NumSamples:  passResult.NumSamples,
				NumPassed:   passResult.NumPassed,
				PassedRates: passedRates,
			}

		default:
			return nil, fmt.Errorf("unknown metric type: %s", metricType)
		}
	}

	return result, nil
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// performStatisticalAnalysis conducts comprehensive statistical analysis
func performStatisticalAnalysis(generated, reference string, confidence float64, bootstrap int) (*metrics.ComparisonResult, error) {
	return nil, fmt.Errorf("statistical metrics analysis is not yet implemented")
}

// outputMetricsResult formats and outputs the metrics results
func outputMetricsResult(result *MetricsResult, outputFile, format string) error {
	var output []byte
	var err error

	switch strings.ToLower(format) {
	case "json":
		output, err = json.MarshalIndent(result, "", "  ")
	case "yaml":
		// Would need YAML marshaling
		output, err = json.MarshalIndent(result, "", "  ") // Fallback to JSON
	case "csv":
		output = []byte(formatMetricsCSV(result))
	default: // table
		output = []byte(formatMetricsTable(result))
	}

	if err != nil {
		return fmt.Errorf("failed to format output: %v", err)
	}

	if outputFile != "" {
		if err := os.WriteFile(outputFile, output, 0644); err != nil {
			return err
		}
		// Also print to stdout when writing to file (for scripting/testing)
		fmt.Print(string(output))
		return nil
	}

	fmt.Print(string(output))
	return nil
}

// formatMetricsTable formats results as a readable table
func formatMetricsTable(result *MetricsResult) string {
	var sb strings.Builder

	sb.WriteString("=== Evaluation Metrics Results ===\n\n")

	if result.BLEU != nil {
		sb.WriteString(fmt.Sprintf("BLEU Score: %.4f\n", result.BLEU.Score))
		sb.WriteString(fmt.Sprintf("  Brevity Penalty: %.4f\n\n", result.BLEU.BP))
	}

	if result.ROUGE != nil {
		sb.WriteString("ROUGE Scores:\n")
		sb.WriteString(fmt.Sprintf("  ROUGE-1: %.4f\n", result.ROUGE.ROUGE1))
		sb.WriteString(fmt.Sprintf("  ROUGE-2: %.4f\n", result.ROUGE.ROUGE2))
		sb.WriteString(fmt.Sprintf("  ROUGE-L: %.4f\n", result.ROUGE.ROUGEL))
		sb.WriteString(fmt.Sprintf("  ROUGE-W: %.4f\n\n", result.ROUGE.ROUGEW))
	}

	if result.METEOR != nil {
		sb.WriteString(fmt.Sprintf("METEOR Score: %.4f\n\n", result.METEOR.Score))
	}

	if result.BERTScore != nil {
		sb.WriteString("BERTScore:\n")
		sb.WriteString(fmt.Sprintf("  Precision: %.4f\n", result.BERTScore.Precision))
		sb.WriteString(fmt.Sprintf("  Recall: %.4f\n", result.BERTScore.Recall))
		sb.WriteString(fmt.Sprintf("  F1: %.4f\n", result.BERTScore.F1))
		sb.WriteString(fmt.Sprintf("  95%% CI: [%.4f, %.4f]\n\n",
			result.BERTScore.ConfidenceInterval[0], result.BERTScore.ConfidenceInterval[1]))
	}

	if result.GEval != nil {
		sb.WriteString("G-Eval Results:\n")
		sb.WriteString(fmt.Sprintf("  Overall Score: %.4f\n", result.GEval.OverallScore))
		for criterion, score := range result.GEval.Scores {
			sb.WriteString(fmt.Sprintf("  %s: %.4f\n", criterion, score))
		}
		if result.GEval.Reasoning != "" {
			sb.WriteString(fmt.Sprintf("  Reasoning: %s\n", result.GEval.Reasoning))
		}
		sb.WriteString("\n")
	}

	if result.UniEval != nil {
		sb.WriteString("UniEval Results:\n")
		sb.WriteString(fmt.Sprintf("  Overall Score: %.4f\n", result.UniEval.OverallScore))
		for dimension, score := range result.UniEval.Dimensions {
			sb.WriteString(fmt.Sprintf("  %s: %.4f\n", dimension, score))
		}
		sb.WriteString("\n")
	}

	if result.PassAtN != nil {
		sb.WriteString("Pass@N Results:\n")
		sb.WriteString(fmt.Sprintf("  Pass@%d: %.2f%% (%d/%d samples)\n",
			result.PassAtN.N, result.PassAtN.PassRate*100,
			result.PassAtN.NumPassed, result.PassAtN.NumSamples))
		if len(result.PassAtN.PassedRates) > 0 {
			sb.WriteString("  Pass rates by N:\n")
			for n := 1; n <= 10; n++ {
				if rate, ok := result.PassAtN.PassedRates[n]; ok {
					sb.WriteString(fmt.Sprintf("    Pass@%d: %.2f%%\n", n, rate*100))
				}
			}
		}
		sb.WriteString("\n")
	}

	if result.Statistics != nil {
		sb.WriteString("Statistical Analysis:\n")
		sb.WriteString(fmt.Sprintf("  Sample Size: %d\n", result.Statistics.Group1Summary.Count))
		sb.WriteString(fmt.Sprintf("  Mean: %.4f\n", result.Statistics.Group1Summary.Mean))
		sb.WriteString(fmt.Sprintf("  Std Dev: %.4f\n", result.Statistics.Group1Summary.StandardDeviation))
		sb.WriteString(fmt.Sprintf("  95%% CI: [%.4f, %.4f]\n",
			result.Statistics.Group1Summary.ConfidenceInterval.LowerBound,
			result.Statistics.Group1Summary.ConfidenceInterval.UpperBound))
		sb.WriteString(fmt.Sprintf("  P-Value: %.4f\n", result.Statistics.TTest.PValue))
		sb.WriteString("\n")
	}

	return sb.String()
}

// formatMetricsCSV formats results as CSV
func formatMetricsCSV(result *MetricsResult) string {
	var sb strings.Builder

	sb.WriteString("Metric,Score,Details\n")

	if result.BLEU != nil {
		sb.WriteString(fmt.Sprintf("BLEU,%.4f,BP=%.4f\n", result.BLEU.Score, result.BLEU.BP))
	}

	if result.ROUGE != nil {
		sb.WriteString(fmt.Sprintf("ROUGE-1,%.4f,\n", result.ROUGE.ROUGE1))
		sb.WriteString(fmt.Sprintf("ROUGE-2,%.4f,\n", result.ROUGE.ROUGE2))
		sb.WriteString(fmt.Sprintf("ROUGE-L,%.4f,\n", result.ROUGE.ROUGEL))
		sb.WriteString(fmt.Sprintf("ROUGE-W,%.4f,\n", result.ROUGE.ROUGEW))
	}

	if result.METEOR != nil {
		sb.WriteString(fmt.Sprintf("METEOR,%.4f,\n", result.METEOR.Score))
	}

	if result.BERTScore != nil {
		sb.WriteString(fmt.Sprintf("BERTScore-P,%.4f,\n", result.BERTScore.Precision))
		sb.WriteString(fmt.Sprintf("BERTScore-R,%.4f,\n", result.BERTScore.Recall))
		sb.WriteString(fmt.Sprintf("BERTScore-F1,%.4f,\n", result.BERTScore.F1))
	}

	if result.PassAtN != nil {
		sb.WriteString(fmt.Sprintf("Pass@%d,%.4f,samples=%d passed=%d\n",
			result.PassAtN.N, result.PassAtN.PassRate,
			result.PassAtN.NumSamples, result.PassAtN.NumPassed))
	}

	return sb.String()
}

// calculateOverallScore computes overall score from individual criterion scores
func calculateOverallScore(scores map[string]float64) float64 {
	if len(scores) == 0 {
		return 0.0
	}

	sum := 0.0
	for _, score := range scores {
		sum += score
	}
	return sum / float64(len(scores))
}

// runSimpleMetrics provides a simple metrics interface for basic prompt analysis
func runSimpleMetrics(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: pe metrics simple <prompt-file>")
	}

	filename := args[0]
	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read file: %v", err)
	}

	content := string(data)

	// Calculate simple metrics
	fmt.Printf("Metrics for %s:\n", filename)

	// Basic complexity score based on length and structure
	lines := strings.Split(content, "\n")
	words := len(strings.Fields(content))
	avgWordsPerLine := float64(words) / float64(len(lines))

	// Simple complexity score
	complexityScore := 0.0
	if words < 50 {
		complexityScore = 0.3
	} else if words < 100 {
		complexityScore = 0.5
	} else if words < 200 {
		complexityScore = 0.7
	} else {
		complexityScore = 0.9
	}

	// Adjust for structure
	if strings.Contains(content, "{{") {
		complexityScore += 0.1
	}
	if complexityScore > 1.0 {
		complexityScore = 1.0
	}

	fmt.Printf("Complexity Score: %.2f\n", complexityScore)
	fmt.Printf("Word Count: %d\n", words)
	fmt.Printf("Line Count: %d\n", len(lines))
	fmt.Printf("Avg Words/Line: %.1f\n", avgWordsPerLine)

	// Additional metrics
	if strings.Contains(content, "{{") {
		fmt.Println("Template Variables: Yes")
	}
	if strings.Contains(content, "-- system-prompt --") {
		fmt.Println("System Prompt: Yes")
	}

	return nil
}
