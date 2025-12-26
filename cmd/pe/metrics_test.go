package main

import (
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
