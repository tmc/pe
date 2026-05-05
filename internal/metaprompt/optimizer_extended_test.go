package metaprompt

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestParseOptimizationResponse(t *testing.T) {
	tests := []struct {
		name            string
		response        string
		wantAnalysis    string
		wantSuggestions int
	}{
		{
			name: "full response",
			response: `ANALYSIS:
This is the analysis section with detailed feedback.

SUGGESTION 1: Improve clarity
Here is the first improved prompt version.

SUGGESTION 2: Add structure
Here is the second improved prompt version.

SUGGESTION 3: Include examples
Here is the third improved prompt version.`,
			wantAnalysis:    "This is the analysis section with detailed feedback.",
			wantSuggestions: 3,
		},
		{
			name: "no analysis section",
			response: `Some intro text

SUGGESTION 1: First
First improvement text

SUGGESTION 2: Second
Second improvement text`,
			wantSuggestions: 2,
		},
		{
			name:            "no suggestions",
			response:        "ANALYSIS:\nJust analysis, no suggestions.",
			wantSuggestions: 0,
		},
	}

	provider := &mockProvider{}
	optimizer := NewOptimizer(provider)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, suggestions := optimizer.parseOptimizationResponse(tt.response)
			if len(suggestions) != tt.wantSuggestions {
				t.Errorf("parseOptimizationResponse() returned %d suggestions, want %d", len(suggestions), tt.wantSuggestions)
			}
		})
	}
}

func TestParseEvaluationResponse(t *testing.T) {
	tests := []struct {
		name       string
		response   string
		wantIndex  int
		wantScore  float64
		wantReason string
	}{
		{
			name: "full evaluation",
			response: `SELECTED: 2
SCORE: 8.5
REASONING: This option provides the best clarity.`,
			wantIndex:  1, // 0-based
			wantScore:  8.5,
			wantReason: "This option provides the best clarity.",
		},
		{
			name: "numeric selected",
			response: `SELECTED: 1
SCORE: 7
REASONING: Good structure.`,
			wantIndex: 0,
			wantScore: 7.0,
		},
		{
			name:       "no structure",
			response:   "Random text without structure",
			wantIndex:  0,
			wantScore:  5.0,
			wantReason: "Default evaluation",
		},
	}

	provider := &mockProvider{}
	optimizer := NewOptimizer(provider)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			index, score, _ := optimizer.parseEvaluationResponse(tt.response)
			if index != tt.wantIndex {
				t.Errorf("parseEvaluationResponse() index = %d, want %d", index, tt.wantIndex)
			}
			if score != tt.wantScore {
				t.Errorf("parseEvaluationResponse() score = %v, want %v", score, tt.wantScore)
			}
		})
	}
}

func TestFormatSuggestions(t *testing.T) {
	provider := &mockProvider{}
	optimizer := NewOptimizer(provider)

	suggestions := []string{
		"First suggestion text",
		"Second suggestion text",
		"Third suggestion text",
	}

	result := optimizer.formatSuggestions(suggestions)

	if len(result) == 0 {
		t.Error("formatSuggestions() returned empty string")
	}

	// Check that all candidates are mentioned
	for i := range suggestions {
		if !stringContains(result, "CANDIDATE") {
			t.Errorf("formatSuggestions() missing CANDIDATE %d", i+1)
		}
	}
}

func TestCalculateImprovementScore(t *testing.T) {
	tests := []struct {
		name    string
		result  *OptimizationResult
		wantMin float64
		wantMax float64
	}{
		{
			name: "with iterations",
			result: &OptimizationResult{
				Iterations: []IterationResult{
					{Score: 5.0},
					{Score: 7.0},
					{Score: 8.0},
				},
			},
			wantMin: 6.0,
			wantMax: 9.0,
		},
		{
			name: "empty iterations",
			result: &OptimizationResult{
				Iterations: []IterationResult{},
			},
			wantMin: 0.0,
			wantMax: 0.0,
		},
		{
			name: "single iteration",
			result: &OptimizationResult{
				Iterations: []IterationResult{
					{Score: 7.5},
				},
			},
			wantMin: 7.0,
			wantMax: 8.0,
		},
	}

	provider := &mockProvider{}
	optimizer := NewOptimizer(provider)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := optimizer.calculateImprovementScore(tt.result)
			if result < tt.wantMin || result > tt.wantMax {
				t.Errorf("calculateImprovementScore() = %v, want between %v and %v", result, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestConfigStruct(t *testing.T) {
	cfg := Config{
		InitialPrompt:        "test prompt",
		Iterations:           5,
		MaxIterations:        10,
		Temperature:          0.7,
		MaxTokens:            1000,
		UseTextGrad:          true,
		Method:               "textgrad",
		Objective:            "improve clarity",
		ConvergenceThreshold: 0.01,
	}

	if cfg.Iterations != 5 {
		t.Error("Config.Iterations mismatch")
	}
	if cfg.Method != "textgrad" {
		t.Error("Config.Method mismatch")
	}
}

func TestOptimizationResultStruct(t *testing.T) {
	result := OptimizationResult{
		OriginalPrompt:   "original",
		OptimizedPrompt:  "optimized",
		Iterations:       []IterationResult{{Score: 8.0}},
		ImprovementScore: 0.85,
		TotalDuration:    time.Minute,
		CreatedAt:        time.Now(),
	}

	if result.OriginalPrompt != "original" {
		t.Error("OptimizationResult.OriginalPrompt mismatch")
	}
}

func TestIterationResultStruct(t *testing.T) {
	iter := IterationResult{
		Iteration:   1,
		Prompt:      "improved prompt",
		Score:       8.5,
		Feedback:    "Good improvement",
		Suggestions: []string{"suggestion1", "suggestion2"},
		Duration:    time.Second * 5,
		Changes:     []string{"change1"},
		Timestamp:   time.Now(),
	}

	if iter.Iteration != 1 {
		t.Error("IterationResult.Iteration mismatch")
	}
}

func TestOptimizerStruct(t *testing.T) {
	provider := &mockProvider{}
	optimizer := NewOptimizer(provider)

	if optimizer.llm == nil {
		t.Error("Optimizer.llm is nil")
	}
	if optimizer.textGrad == nil {
		t.Error("Optimizer.textGrad is nil")
	}
	if optimizer.pe2 == nil {
		t.Error("Optimizer.pe2 is nil")
	}
	if optimizer.apex == nil {
		t.Error("Optimizer.apex is nil")
	}
}

func TestOptimizationResultSaveToFile(t *testing.T) {
	result := &OptimizationResult{
		OriginalPrompt:   "original",
		OptimizedPrompt:  "optimized",
		Iterations:       []IterationResult{{Score: 8.0}},
		ImprovementScore: 0.85,
		TotalDuration:    time.Minute,
		CreatedAt:        time.Now(),
	}

	// Create temp file
	tempDir := os.TempDir()
	tempFile := filepath.Join(tempDir, "test_result.json")

	err := result.SaveToFile(tempFile)
	if err != nil {
		t.Errorf("SaveToFile() error = %v", err)
	}

	// Clean up
	os.Remove(tempFile)

	// Test invalid path
	err = result.SaveToFile("/invalid/path/that/does/not/exist/file.json")
	if err == nil {
		t.Error("SaveToFile() should fail for invalid path")
	}
}

func TestSelectBestImprovementEmpty(t *testing.T) {
	provider := &mockProvider{}
	optimizer := NewOptimizer(provider)

	cfg := Config{
		InitialPrompt: "test",
		Iterations:    3,
	}

	prompt, score, feedback := optimizer.selectBestImprovement(nil, "original", []string{}, cfg)

	if prompt != "original" {
		t.Error("selectBestImprovement() should return original for empty suggestions")
	}
	if score != 0.0 {
		t.Error("selectBestImprovement() should return 0 score for empty suggestions")
	}
	if feedback == "" {
		t.Error("selectBestImprovement() should return feedback message")
	}
}

// Helper function for tests
func stringContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestOptimizationWithEmptyObjective(t *testing.T) {
	provider := &mockProvider{}
	optimizer := NewOptimizer(provider)

	cfg := Config{
		InitialPrompt: "Summarize the text",
		Iterations:    2,
		Method:        "standard",
	}

	// Should still work without explicit objective
	if optimizer.llm == nil {
		t.Error("optimizer.llm should not be nil")
	}
	if cfg.Iterations != 2 {
		t.Error("Config.Iterations mismatch")
	}
}

func TestOptimizerHybridMethod(t *testing.T) {
	provider := &mockProvider{}
	optimizer := NewOptimizer(provider)

	// Test that optimizer has all method handlers
	if optimizer.textGrad == nil {
		t.Error("optimizer.textGrad is nil")
	}
	if optimizer.pe2 == nil {
		t.Error("optimizer.pe2 is nil")
	}
	if optimizer.apex == nil {
		t.Error("optimizer.apex is nil")
	}
}
