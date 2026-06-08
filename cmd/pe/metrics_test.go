package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMetricsCmd_FlagParsing(t *testing.T) {
	cmd := metricsCmd()

	flags := []string{
		"type", "all", "generated", "reference",
		"generated-file", "reference-file", "criteria",
		"output", "format", "statistical", "confidence",
		"bootstrap", "provider", "model", "n", "test-cases", "samples-file",
	}

	for _, name := range flags {
		if cmd.Flags().Lookup(name) == nil {
			t.Errorf("Expected flag %q to exist", name)
		}
	}
}

func TestMetricsCmd_CommandStructure(t *testing.T) {
	cmd := metricsCmd()

	if cmd.Use != "metrics [simple <file>] | [flags]" {
		t.Errorf("Unexpected Use: %s", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("Expected Short description to be set")
	}

	if cmd.Long == "" {
		t.Error("Expected Long description to be set")
	}
}

func TestMetricsCmd_FlagDefaults(t *testing.T) {
	cmd := metricsCmd()

	// Check format default
	format, err := cmd.Flags().GetString("format")
	if err != nil {
		t.Errorf("Failed to get format flag: %v", err)
	}
	if format != "table" {
		t.Errorf("Expected default format 'table', got %s", format)
	}

	// Check confidence default
	confidence, err := cmd.Flags().GetFloat64("confidence")
	if err != nil {
		t.Errorf("Failed to get confidence flag: %v", err)
	}
	if confidence != 0.95 {
		t.Errorf("Expected default confidence 0.95, got %f", confidence)
	}

	// Check bootstrap default
	bootstrap, err := cmd.Flags().GetInt("bootstrap")
	if err != nil {
		t.Errorf("Failed to get bootstrap flag: %v", err)
	}
	if bootstrap != 1000 {
		t.Errorf("Expected default bootstrap 1000, got %d", bootstrap)
	}

	// Check n default
	n, err := cmd.Flags().GetInt("n")
	if err != nil {
		t.Errorf("Failed to get n flag: %v", err)
	}
	if n != 1 {
		t.Errorf("Expected default n 1, got %d", n)
	}

	// Check provider default
	provider, err := cmd.Flags().GetString("provider")
	if err != nil {
		t.Errorf("Failed to get provider flag: %v", err)
	}
	if provider != "openai" {
		t.Errorf("Expected default provider 'openai', got %s", provider)
	}

	// Check model default
	model, err := cmd.Flags().GetString("model")
	if err != nil {
		t.Errorf("Failed to get model flag: %v", err)
	}
	if model != "gpt-4" {
		t.Errorf("Expected default model 'gpt-4', got %s", model)
	}
}

func TestLoadTextData(t *testing.T) {
	tests := []struct {
		name         string
		genText      string
		refText      string
		genFile      string
		refFile      string
		wantErr      bool
		errSubstring string
	}{
		{
			name:    "direct text input",
			genText: "generated text",
			refText: "reference text",
			wantErr: false,
		},
		{
			name:         "no generated text",
			genText:      "",
			refText:      "",
			wantErr:      true,
			errSubstring: "must provide",
		},
		{
			name:         "nonexistent generated file",
			genText:      "",
			genFile:      "nonexistent.txt",
			wantErr:      true,
			errSubstring: "failed to read generated file",
		},
		{
			name:         "nonexistent reference file",
			genText:      "text",
			refFile:      "nonexistent.txt",
			wantErr:      true,
			errSubstring: "failed to read reference file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen, ref, err := loadTextData(tt.genText, tt.refText, tt.genFile, tt.refFile)
			if tt.wantErr {
				if err == nil {
					t.Error("Expected error")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if gen != tt.genText {
					t.Errorf("Expected generated %q, got %q", tt.genText, gen)
				}
				if ref != tt.refText {
					t.Errorf("Expected reference %q, got %q", tt.refText, ref)
				}
			}
		})
	}
}

func TestMin(t *testing.T) {
	tests := []struct {
		a, b, want int
	}{
		{1, 2, 1},
		{2, 1, 1},
		{5, 5, 5},
		{0, 10, 0},
		{-1, 1, -1},
	}

	for _, tt := range tests {
		got := min(tt.a, tt.b)
		if got != tt.want {
			t.Errorf("min(%d, %d) = %d, want %d", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestCalculateOverallScore(t *testing.T) {
	tests := []struct {
		name   string
		scores map[string]float64
		want   float64
	}{
		{
			name:   "empty scores",
			scores: map[string]float64{},
			want:   0.0,
		},
		{
			name:   "single score",
			scores: map[string]float64{"accuracy": 0.8},
			want:   0.8,
		},
		{
			name:   "multiple scores",
			scores: map[string]float64{"accuracy": 0.8, "clarity": 0.6},
			want:   0.7,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateOverallScore(tt.scores)
			if got != tt.want {
				t.Errorf("calculateOverallScore() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMetricsResultStruct(t *testing.T) {
	result := MetricsResult{
		BLEU: &BLEUScore{
			Score:     0.75,
			Precision: []float64{0.8, 0.7, 0.6, 0.5},
			BP:        1.0,
		},
		ROUGE: &ROUGEScore{
			ROUGE1: 0.85,
			ROUGE2: 0.65,
			ROUGEL: 0.80,
			ROUGEW: 0.70,
		},
		METEOR: &METEORScore{
			Score:                0.72,
			UnigarmMatches:       50,
			ChunkCount:           5,
			UnigarmPrecision:     0.8,
			UnigramRecall:        0.7,
			FragmentationPenalty: 0.1,
		},
	}

	if result.BLEU.Score != 0.75 {
		t.Error("MetricsResult.BLEU.Score mismatch")
	}
	if result.ROUGE.ROUGE1 != 0.85 {
		t.Error("MetricsResult.ROUGE.ROUGE1 mismatch")
	}
	if result.METEOR.Score != 0.72 {
		t.Error("MetricsResult.METEOR.Score mismatch")
	}
}

func TestMetricsCalculationOutputAndSimpleMode(t *testing.T) {
	old := os.Getenv("PE_TEST_MODE")
	os.Setenv("PE_TEST_MODE", "true")
	defer os.Setenv("PE_TEST_MODE", old)

	generated := "the cat sat on the mat"
	reference := "the cat is on the mat"
	result, err := calculateMetrics(generated, reference, []string{"bleu", "rouge", "meteor"}, nil, "mock", "test-model", 1, "", "")
	if err != nil {
		t.Fatal(err)
	}
	if result.BLEU == nil || result.ROUGE == nil || result.METEOR == nil {
		t.Fatalf("result = %#v", result)
	}
	samplesFile := filepath.Join(t.TempDir(), "samples.txt")
	if err := os.WriteFile(samplesFile, []byte("working code sample\nerror sample\nanother working code sample\n"), 0644); err != nil {
		t.Fatal(err)
	}
	pass, err := calculateMetrics("working code sample", "", []string{"pass-at-n"}, nil, "mock", "test-model", 2, "", samplesFile)
	if err != nil {
		t.Fatal(err)
	}
	if pass.PassAtN == nil || pass.PassAtN.NumSamples != 3 {
		t.Fatalf("pass@n = %#v", pass.PassAtN)
	}
	testCasesFile := filepath.Join(t.TempDir(), "cases.json")
	if err := os.WriteFile(testCasesFile, []byte(`[{"input":"x","expected":"working"}]`), 0644); err != nil {
		t.Fatal(err)
	}
	pass, err = calculateMetrics("working code sample", "", []string{"pass_at_n"}, nil, "mock", "test-model", 1, testCasesFile, "")
	if err != nil || pass.PassAtN == nil {
		t.Fatalf("pass@n tests = %#v err=%v", pass, err)
	}
	for _, tt := range []struct {
		metrics []string
		want    string
	}{
		{[]string{"bleu"}, "BLEU requires reference"},
		{[]string{"rouge"}, "ROUGE requires reference"},
		{[]string{"meteor"}, "METEOR requires reference"},
		{[]string{"unknown"}, "unknown metric"},
	} {
		if _, err := calculateMetrics("generated", "", tt.metrics, nil, "mock", "test-model", 1, "", ""); err == nil || !strings.Contains(err.Error(), tt.want) {
			t.Fatalf("%v error = %v, want %q", tt.metrics, err, tt.want)
		}
	}
	if _, err := calculateMetrics("generated", "", []string{"pass-at-n"}, nil, "mock", "test-model", 1, filepath.Join(t.TempDir(), "missing.json"), ""); err == nil {
		t.Fatal("missing test cases succeeded")
	}
	badCases := filepath.Join(t.TempDir(), "bad.json")
	if err := os.WriteFile(badCases, []byte("{"), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := calculateMetrics("generated", "", []string{"pass-at-n"}, nil, "mock", "test-model", 1, badCases, ""); err == nil {
		t.Fatal("bad test cases succeeded")
	}
	if _, err := calculateMetrics("generated", "", []string{"pass-at-n"}, nil, "mock", "test-model", 1, "", filepath.Join(t.TempDir(), "missing.txt")); err == nil {
		t.Fatal("missing samples succeeded")
	}

	stats, err := performStatisticalAnalysis("1 2 3 4", "2 3 4 5", 0.95, 10)
	if err != nil {
		t.Fatalf("stats = %#v err=%v", stats, err)
	}
	if stats.Group1Summary.Count != 4 || stats.Group2Summary.Count != 4 {
		t.Fatalf("stats counts = %d/%d, want 4/4", stats.Group1Summary.Count, stats.Group2Summary.Count)
	}
	if _, err := performStatisticalAnalysis("1", "2 3", 0.95, 10); err == nil || !strings.Contains(err.Error(), "at least 2 samples") {
		t.Fatalf("insufficient samples error = %v", err)
	}
	if _, err := performStatisticalAnalysis("1 bad 3", "2 3 4", 0.95, 10); err == nil || !strings.Contains(err.Error(), "parse generated sample") {
		t.Fatalf("malformed samples error = %v", err)
	}
	if _, err := performStatisticalAnalysis(`[1,2,3]`, `[2,3,4]`, 0.95, 10); err != nil {
		t.Fatalf("json statistical samples: %v", err)
	}
	full := &MetricsResult{
		BLEU:       &BLEUScore{Score: 0.5, BP: 1},
		ROUGE:      &ROUGEScore{ROUGE1: 0.5, ROUGE2: 0.4, ROUGEL: 0.6, ROUGEW: 0.3},
		METEOR:     &METEORScore{Score: 0.7},
		BERTScore:  &BERTScoreResult{Precision: 0.8, Recall: 0.7, F1: 0.75, ConfidenceInterval: [2]float64{0.7, 0.8}},
		GEval:      &GEvalResult{Scores: map[string]float64{"accuracy": 0.8}, OverallScore: 0.8, Reasoning: "ok"},
		UniEval:    &UniEvalResult{Dimensions: map[string]float64{"coherence": 0.9}, OverallScore: 0.9},
		PassAtN:    &PassAtNScore{N: 2, PassRate: 0.5, NumSamples: 2, NumPassed: 1, PassedRates: map[int]float64{1: 0.5}},
		Statistics: stats,
		Caveat:     statisticalCaveat,
	}
	for _, format := range []string{"table", "json", "yaml", "csv"} {
		outFile := filepath.Join(t.TempDir(), "metrics."+format)
		if err := outputMetricsResult(full, outFile, format); err != nil {
			t.Fatalf("output %s: %v", format, err)
		}
		if data, err := os.ReadFile(outFile); err != nil || len(data) == 0 {
			t.Fatalf("output file %s len=%d err=%v", format, len(data), err)
		}
	}
	if table := formatMetricsTable(full); !strings.Contains(table, "BLEU Score") || !strings.Contains(table, "Pass@N") || !strings.Contains(table, statisticalCaveat) {
		t.Fatalf("table = %s", table)
	}
	if csv := formatMetricsCSV(full); !strings.Contains(csv, "Metric,Score") || !strings.Contains(csv, "Pass@2") || !strings.Contains(csv, "StatisticsPValue") {
		t.Fatalf("csv = %s", csv)
	}
	promptFile := filepath.Join(t.TempDir(), "prompt.txt")
	if err := os.WriteFile(promptFile, []byte("Hello {{name}}\n-- system-prompt --\nSystem"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := runSimpleMetrics(metricsCmd(), []string{promptFile}); err != nil {
		t.Fatal(err)
	}
	if err := runSimpleMetrics(metricsCmd(), nil); err == nil {
		t.Fatal("simple without file succeeded")
	}
	if err := runSimpleMetrics(metricsCmd(), []string{filepath.Join(t.TempDir(), "missing.txt")}); err == nil {
		t.Fatal("simple missing file succeeded")
	}
}

func TestBERTScoreResultStruct(t *testing.T) {
	result := BERTScoreResult{
		Precision:          0.85,
		Recall:             0.80,
		F1:                 0.825,
		ConfidenceInterval: [2]float64{0.80, 0.85},
	}

	if result.F1 != 0.825 {
		t.Error("BERTScoreResult.F1 mismatch")
	}
}

func TestGEvalResultStruct(t *testing.T) {
	result := GEvalResult{
		Scores: map[string]float64{
			"accuracy": 0.9,
			"clarity":  0.85,
		},
		OverallScore: 0.875,
		Reasoning:    "Good performance",
		Criteria:     []string{"accuracy", "clarity"},
	}

	if result.OverallScore != 0.875 {
		t.Error("GEvalResult.OverallScore mismatch")
	}
}

func TestUniEvalResultStruct(t *testing.T) {
	result := UniEvalResult{
		Dimensions: map[string]float64{
			"fluency":   0.9,
			"coherence": 0.85,
		},
		OverallScore: 0.875,
		TaskType:     "summarization",
	}

	if result.TaskType != "summarization" {
		t.Error("UniEvalResult.TaskType mismatch")
	}
}

func TestPassAtNScoreStruct(t *testing.T) {
	result := PassAtNScore{
		N:          10,
		PassRate:   0.7,
		NumSamples: 100,
		NumPassed:  70,
		PassedRates: map[int]float64{
			1:  0.3,
			5:  0.5,
			10: 0.7,
		},
	}

	if result.N != 10 {
		t.Error("PassAtNScore.N mismatch")
	}
	if result.NumPassed != 70 {
		t.Error("PassAtNScore.NumPassed mismatch")
	}
}

func TestFormatMetricsTable(t *testing.T) {
	result := &MetricsResult{
		BLEU: &BLEUScore{
			Score: 0.75,
			BP:    1.0,
		},
	}

	output := formatMetricsTable(result)
	if output == "" {
		t.Error("Expected non-empty table output")
	}
	if len(output) < 50 {
		t.Error("Expected substantial table output")
	}
}

func TestFormatMetricsCSV(t *testing.T) {
	result := &MetricsResult{
		BLEU: &BLEUScore{
			Score: 0.75,
			BP:    1.0,
		},
	}

	output := formatMetricsCSV(result)
	if output == "" {
		t.Error("Expected non-empty CSV output")
	}
	if len(output) < 20 {
		t.Error("Expected substantial CSV output")
	}
}
