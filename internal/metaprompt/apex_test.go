package metaprompt

import (
	"testing"
)

func TestNewAPEXOptimizer(t *testing.T) {
	provider := &mockProvider{}
	optimizer := NewAPEXOptimizer(provider)

	if optimizer == nil {
		t.Fatal("NewAPEXOptimizer returned nil")
	}
	if optimizer.llm == nil {
		t.Error("APEXOptimizer has nil llm provider")
	}
}

func TestParseAPEXConfig(t *testing.T) {
	provider := &mockProvider{}
	optimizer := NewAPEXOptimizer(provider)

	cfg := Config{
		InitialPrompt: "test",
		Iterations:    5,
	}

	apexConfig := optimizer.parseAPEXConfig(cfg)

	if apexConfig.BeamWidth != 5 {
		t.Errorf("BeamWidth = %d, want 5", apexConfig.BeamWidth)
	}
	if len(apexConfig.MutationOperators) == 0 {
		t.Error("MutationOperators should not be empty")
	}
	if !apexConfig.UseSearchHistory {
		t.Error("UseSearchHistory should be true by default")
	}
	if !apexConfig.GreedySelection {
		t.Error("GreedySelection should be true by default")
	}
}

func TestSelectTopCandidates(t *testing.T) {
	tests := []struct {
		name       string
		candidates []*APEXCandidate
		beamWidth  int
		wantLen    int
		wantTop    float64
	}{
		{
			name: "select top 2 from 4",
			candidates: []*APEXCandidate{
				{Prompt: "a", Score: 5.0},
				{Prompt: "b", Score: 8.0},
				{Prompt: "c", Score: 3.0},
				{Prompt: "d", Score: 7.0},
			},
			beamWidth: 2,
			wantLen:   2,
			wantTop:   8.0,
		},
		{
			name: "beam width larger than candidates",
			candidates: []*APEXCandidate{
				{Prompt: "a", Score: 5.0},
				{Prompt: "b", Score: 8.0},
			},
			beamWidth: 5,
			wantLen:   2,
			wantTop:   8.0,
		},
		{
			name: "single candidate",
			candidates: []*APEXCandidate{
				{Prompt: "a", Score: 5.0},
			},
			beamWidth: 3,
			wantLen:   1,
			wantTop:   5.0,
		},
	}

	provider := &mockProvider{}
	optimizer := NewAPEXOptimizer(provider)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := optimizer.selectTopCandidates(tt.candidates, tt.beamWidth)
			if len(result) != tt.wantLen {
				t.Errorf("selectTopCandidates() returned %d candidates, want %d", len(result), tt.wantLen)
			}
			if len(result) > 0 && result[0].Score != tt.wantTop {
				t.Errorf("top candidate score = %v, want %v", result[0].Score, tt.wantTop)
			}
		})
	}
}

func TestTruncatePrompt(t *testing.T) {
	tests := []struct {
		name      string
		prompt    string
		maxLength int
		wantMax   int
	}{
		{
			name:      "no truncation needed",
			prompt:    "short prompt",
			maxLength: 100,
			wantMax:   12,
		},
		{
			name:      "truncate at sentence",
			prompt:    "First sentence. Second sentence. Third sentence.",
			maxLength: 30,
			wantMax:   30,
		},
		{
			name:      "truncate at word",
			prompt:    "oneword",
			maxLength: 3,
			wantMax:   3,
		},
	}

	provider := &mockProvider{}
	optimizer := NewAPEXOptimizer(provider)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := optimizer.truncatePrompt(tt.prompt, tt.maxLength)
			if len(result) > tt.wantMax {
				t.Errorf("truncatePrompt() returned %d chars, want <= %d", len(result), tt.wantMax)
			}
		})
	}
}

func TestExtractMutatedPrompt(t *testing.T) {
	tests := []struct {
		name     string
		response string
		original string
		wantLen  int
	}{
		{
			name:     "with IMPROVED PROMPT pattern",
			response: "Some intro text\n\nIMPROVED PROMPT:\nThis is the improved prompt that should be extracted and is longer than fifty characters.",
			original: "original",
			wantLen:  50, // Should be at least 50 chars
		},
		{
			name:     "with ENHANCED PROMPT pattern",
			response: "ENHANCED PROMPT:\nThis is the enhanced version of the prompt with more than fifty characters in it.",
			original: "original",
			wantLen:  50,
		},
		{
			name:     "no pattern found, use longest paragraph",
			response: "Short para.\n\nThis is a much longer paragraph that contains more than one hundred characters and should be selected as the result.",
			original: "original",
			wantLen:  50,
		},
		{
			name:     "fallback to original",
			response: "short",
			original: "original prompt",
			wantLen:  8,
		},
	}

	provider := &mockProvider{}
	optimizer := NewAPEXOptimizer(provider)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := optimizer.extractMutatedPrompt(tt.response, tt.original)
			if len(result) < tt.wantLen {
				t.Errorf("extractMutatedPrompt() returned %d chars, want >= %d", len(result), tt.wantLen)
			}
		})
	}
}

func TestGenerateAPEXFeedback(t *testing.T) {
	tests := []struct {
		name      string
		candidate *APEXCandidate
		history   *APEXSearchHistory
		wantCheck func(string) bool
	}{
		{
			name: "with mutations and history",
			candidate: &APEXCandidate{
				Score:       8.5,
				MutationLog: []string{"rephrase_section", "add_examples"},
			},
			history: &APEXSearchHistory{
				SuccessfulMutations: map[string]float64{
					"rephrase_section": 0.5,
					"add_examples":     0.3,
				},
			},
			wantCheck: func(s string) bool {
				return len(s) > 0
			},
		},
		{
			name: "no mutations",
			candidate: &APEXCandidate{
				Score:       7.0,
				MutationLog: []string{},
			},
			history: &APEXSearchHistory{
				SuccessfulMutations: map[string]float64{},
			},
			wantCheck: func(s string) bool {
				return len(s) > 0
			},
		},
	}

	provider := &mockProvider{}
	optimizer := NewAPEXOptimizer(provider)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := optimizer.generateAPEXFeedback(tt.candidate, tt.history)
			if !tt.wantCheck(result) {
				t.Errorf("generateAPEXFeedback() check failed: %s", result)
			}
		})
	}
}

func TestCalculateAPEXImprovementScore(t *testing.T) {
	tests := []struct {
		name    string
		result  *OptimizationResult
		wantMin float64
		wantMax float64
	}{
		{
			name: "consistent improvement",
			result: &OptimizationResult{
				Iterations: []IterationResult{
					{Score: 5.0},
					{Score: 6.0},
					{Score: 7.0},
				},
			},
			wantMin: 7.0,
			wantMax: 8.0,
		},
		{
			name: "no iterations",
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
					{Score: 8.0},
				},
			},
			wantMin: 8.0,
			wantMax: 8.0,
		},
	}

	provider := &mockProvider{}
	optimizer := NewAPEXOptimizer(provider)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := optimizer.calculateAPEXImprovementScore(tt.result)
			if result < tt.wantMin || result > tt.wantMax {
				t.Errorf("calculateAPEXImprovementScore() = %v, want between %v and %v", result, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestUpdateSearchHistory(t *testing.T) {
	provider := &mockProvider{}
	optimizer := NewAPEXOptimizer(provider)

	history := &APEXSearchHistory{
		SuccessfulMutations: make(map[string]float64),
		FailedMutations:     make(map[string]int),
		BestPrompts:         []*APEXCandidate{},
	}

	oldBeam := []*APEXCandidate{
		{Prompt: "original", Score: 5.0},
	}

	newCandidates := []*APEXCandidate{
		{
			Prompt:      "improved",
			Score:       7.0,
			MutationLog: []string{"rephrase_section"},
		},
		{
			Prompt:      "worse",
			Score:       4.0,
			MutationLog: []string{"add_examples"},
		},
	}

	optimizer.updateSearchHistory(history, oldBeam, newCandidates)

	// Check successful mutation was recorded
	if _, exists := history.SuccessfulMutations["rephrase_section"]; !exists {
		t.Error("successful mutation should be recorded")
	}

	// Check failed mutation was recorded
	if history.FailedMutations["add_examples"] != 1 {
		t.Error("failed mutation should be recorded")
	}

	// Check best prompts were updated
	if len(history.BestPrompts) == 0 {
		t.Error("best prompts should be updated")
	}
}

func TestAPEXConfigStruct(t *testing.T) {
	config := APEXConfig{
		BeamWidth:               5,
		MutationOperators:       []string{"op1", "op2"},
		UseSearchHistory:        true,
		GreedySelection:         true,
		LengthOptimization:      true,
		MaxPromptLength:         2000,
		MinImprovementThreshold: 0.05,
		MutationProbability:     0.3,
	}

	if config.BeamWidth != 5 {
		t.Error("BeamWidth mismatch")
	}
	if len(config.MutationOperators) != 2 {
		t.Error("MutationOperators length mismatch")
	}
}

func TestAPEXCandidateStruct(t *testing.T) {
	parent := &APEXCandidate{
		Prompt:     "parent",
		Score:      5.0,
		Generation: 0,
	}

	candidate := APEXCandidate{
		Prompt:      "child",
		Score:       7.0,
		Generation:  1,
		MutationLog: []string{"mutation1"},
		Parent:      parent,
	}

	if candidate.Parent != parent {
		t.Error("Parent reference mismatch")
	}
}

func TestAPEXSearchHistoryStruct(t *testing.T) {
	history := APEXSearchHistory{
		SuccessfulMutations: map[string]float64{"m1": 0.5},
		FailedMutations:     map[string]int{"m2": 3},
		BestPrompts:         []*APEXCandidate{{Score: 8.0}},
	}

	if len(history.SuccessfulMutations) != 1 {
		t.Error("SuccessfulMutations length mismatch")
	}
}

func TestGenerateMutationPrompts(t *testing.T) {
	provider := &mockProvider{}
	optimizer := NewAPEXOptimizer(provider)

	testPrompt := "Test prompt for mutation"

	tests := []struct {
		name     string
		generate func() string
		wantLen  int
	}{
		{"rephrase", func() string { return optimizer.generateRephraseMutation(testPrompt) }, 50},
		{"example", func() string { return optimizer.generateExampleMutation(testPrompt) }, 50},
		{"restructure", func() string { return optimizer.generateRestructureMutation(testPrompt) }, 50},
		{"clarify", func() string { return optimizer.generateClarificationMutation(testPrompt) }, 50},
		{"constraint", func() string { return optimizer.generateConstraintMutation(testPrompt) }, 50},
		{"length", func() string { return optimizer.generateLengthOptimizationMutation(testPrompt, 500) }, 50},
		{"specificity", func() string { return optimizer.generateSpecificityMutation(testPrompt) }, 50},
		{"formatting", func() string { return optimizer.generateFormattingMutation(testPrompt) }, 50},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.generate()
			if len(result) < tt.wantLen {
				t.Errorf("generate mutation prompt too short: %d chars", len(result))
			}
		})
	}
}
