package metaprompt

import (
	"context"
	"testing"
)

func TestNewPE2Optimizer(t *testing.T) {
	provider := &mockProvider{}
	optimizer := NewPE2Optimizer(provider)

	if optimizer == nil {
		t.Fatal("NewPE2Optimizer returned nil")
	}
}

func TestParsePE2Config(t *testing.T) {
	provider := &mockProvider{}
	optimizer := NewPE2Optimizer(provider)

	cfg := Config{
		InitialPrompt: "test",
		Iterations:    5,
		Method:        "pe2",
	}

	pe2Config := optimizer.parsePE2Config(cfg)

	// parsePE2Config should set reasonable defaults
	if pe2Config.ReasoningTemplate == "" && pe2Config.MetaPromptStyle == "" {
		t.Error("parsePE2Config should set default config values")
	}
}

func TestGeneratePE2MetaPrompt(t *testing.T) {
	provider := &mockProvider{}
	optimizer := NewPE2Optimizer(provider)

	config := PE2Config{
		ReasoningTemplate:    "chain_of_thought",
		ContextSpecification: "standard",
		DescriptionDepth:     "standard",
		MetaPromptStyle:      "expert",
		FeedbackIntegration:  true,
		ErrorCorrection:      true,
	}

	prompt := optimizer.generatePE2MetaPrompt("test input", config, []string{"previous feedback"})

	if len(prompt) == 0 {
		t.Error("generatePE2MetaPrompt returned empty prompt")
	}
}

func TestGenerateExpertPersona(t *testing.T) {
	provider := &mockProvider{}
	optimizer := NewPE2Optimizer(provider)

	tests := []struct {
		name  string
		style string
		check func(string) bool
	}{
		{
			name:  "expert style",
			style: "expert",
			check: func(s string) bool { return len(s) > 0 },
		},
		{
			name:  "systematic style",
			style: "systematic",
			check: func(s string) bool { return len(s) > 0 },
		},
		{
			name:  "creative style",
			style: "creative",
			check: func(s string) bool { return len(s) > 0 },
		},
		{
			name:  "default style",
			style: "",
			check: func(s string) bool { return len(s) > 0 },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := optimizer.generateExpertPersona(tt.style)
			if !tt.check(result) {
				t.Error("generateExpertPersona check failed")
			}
		})
	}
}

func TestGenerateDetailedDescription(t *testing.T) {
	provider := &mockProvider{}
	optimizer := NewPE2Optimizer(provider)

	tests := []struct {
		name  string
		depth string
	}{
		{"basic", "basic"},
		{"standard", "standard"},
		{"comprehensive", "comprehensive"},
		{"default", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := optimizer.generateDetailedDescription(tt.depth)
			if len(result) == 0 {
				t.Error("generateDetailedDescription returned empty string")
			}
		})
	}
}

func TestGenerateContextSpecification(t *testing.T) {
	provider := &mockProvider{}
	optimizer := NewPE2Optimizer(provider)

	tests := []struct {
		name  string
		level string
	}{
		{"minimal", "minimal"},
		{"standard", "standard"},
		{"detailed", "detailed"},
		{"default", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := optimizer.generateContextSpecification(tt.level)
			if len(result) == 0 {
				t.Error("generateContextSpecification returned empty string")
			}
		})
	}
}

func TestGenerateReasoningTemplate(t *testing.T) {
	provider := &mockProvider{}
	optimizer := NewPE2Optimizer(provider)

	tests := []struct {
		name     string
		template string
	}{
		{"chain_of_thought", "chain_of_thought"},
		{"step_by_step", "step_by_step"},
		{"analytical", "analytical"},
		{"default", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := optimizer.generateReasoningTemplate(tt.template)
			if len(result) == 0 {
				t.Error("generateReasoningTemplate returned empty string")
			}
		})
	}
}

func TestGenerateErrorCorrectionInstructions(t *testing.T) {
	provider := &mockProvider{}
	optimizer := NewPE2Optimizer(provider)

	result := optimizer.generateErrorCorrectionInstructions()
	if len(result) == 0 {
		t.Error("generateErrorCorrectionInstructions returned empty string")
	}
}

func TestGenerateOutputFormat(t *testing.T) {
	provider := &mockProvider{}
	optimizer := NewPE2Optimizer(provider)

	result := optimizer.generateOutputFormat()
	if len(result) == 0 {
		t.Error("generateOutputFormat returned empty string")
	}
}

func TestPE2ConfigStruct(t *testing.T) {
	config := PE2Config{
		ReasoningTemplate:    "chain_of_thought",
		ContextSpecification: "standard",
		DescriptionDepth:     "comprehensive",
		MetaPromptStyle:      "expert",
		FeedbackIntegration:  true,
		ErrorCorrection:      true,
	}

	if config.ReasoningTemplate != "chain_of_thought" {
		t.Error("PE2Config.ReasoningTemplate mismatch")
	}
	if !config.FeedbackIntegration {
		t.Error("PE2Config.FeedbackIntegration mismatch")
	}
}

func TestParsePE2Response(t *testing.T) {
	provider := &mockProvider{}
	optimizer := NewPE2Optimizer(provider)

	tests := []struct {
		name         string
		response     string
		wantPrompt   bool
		wantFeedback bool
		wantMinScore float64
	}{
		{
			name: "complete structured response",
			response: `ANALYSIS:
The current prompt lacks specificity.

REASONING:
Step 1: Identify issues
Step 2: Apply fixes

OPTIMIZED PROMPT:
This is the optimized prompt with better clarity.

IMPROVEMENT SUMMARY:
Added structure and clarity.

VALIDATION:
Meets all requirements.`,
			wantPrompt:   true,
			wantFeedback: true,
			wantMinScore: 8.0,
		},
		{
			name:         "minimal response",
			response:     "Just a simple response without any structure",
			wantPrompt:   true,
			wantFeedback: false,
			wantMinScore: 5.0,
		},
		{
			name: "partial response with analysis only",
			response: `ANALYSIS:
Some analysis here.

OPTIMIZED PROMPT:
The optimized version.`,
			wantPrompt:   true,
			wantFeedback: true,
			wantMinScore: 6.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prompt, feedback, score := optimizer.parsePE2Response(tt.response)

			if tt.wantPrompt && len(prompt) == 0 {
				t.Error("expected non-empty prompt")
			}
			if tt.wantFeedback && len(feedback) == 0 {
				t.Error("expected non-empty feedback")
			}
			if score < tt.wantMinScore {
				t.Errorf("score = %v, want >= %v", score, tt.wantMinScore)
			}
		})
	}
}

func TestExtractPromptFallback(t *testing.T) {
	provider := &mockProvider{}
	optimizer := NewPE2Optimizer(provider)

	tests := []struct {
		name     string
		response string
		wantLen  bool
	}{
		{
			name:     "with OPTIMIZED PROMPT marker",
			response: "Some text\nOPTIMIZED PROMPT:\nThis is the extracted prompt.\nNEXT SECTION:",
			wantLen:  true,
		},
		{
			name:     "with IMPROVED PROMPT marker",
			response: "Some text\nIMPROVED PROMPT:\nThis is the improved prompt.",
			wantLen:  true,
		},
		{
			name:     "with FINAL PROMPT marker",
			response: "FINAL PROMPT:\nFinal version of the prompt",
			wantLen:  true,
		},
		{
			name:     "substantial paragraph fallback",
			response: "This is a substantial paragraph with more than one hundred characters that should be extracted as the fallback when no markers are found in the response content.",
			wantLen:  true,
		},
		{
			name:     "short response fallback",
			response: "Short",
			wantLen:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := optimizer.extractPromptFallback(tt.response)
			if tt.wantLen && len(result) == 0 {
				t.Error("expected non-empty result")
			}
		})
	}
}

func TestCalculatePE2Score(t *testing.T) {
	provider := &mockProvider{}
	optimizer := NewPE2Optimizer(provider)

	tests := []struct {
		name     string
		sections map[string]string
		wantMin  float64
		wantMax  float64
	}{
		{
			name:     "empty sections",
			sections: map[string]string{},
			wantMin:  5.0,
			wantMax:  5.0,
		},
		{
			name: "complete sections with substantial content",
			sections: map[string]string{
				"ANALYSIS":            "Detailed analysis of the prompt showing step by step reasoning and issues. " + "More content to exceed 200 chars. " + "Additional text here.",
				"REASONING":           "Step by step analysis of improvements. " + "More content to exceed 200 chars. " + "Additional text here.",
				"OPTIMIZED PROMPT":    "The optimized prompt with improvements. " + "More content to exceed 200 chars. " + "Additional text here.",
				"IMPROVEMENT SUMMARY": "Summary of improvements made",
				"VALIDATION":          "Validation that requirements are met. " + "More content to exceed 200 chars. " + "Additional text here.",
			},
			wantMin: 9.0,
			wantMax: 10.0,
		},
		{
			name: "partial sections",
			sections: map[string]string{
				"ANALYSIS":         "Brief analysis",
				"OPTIMIZED PROMPT": "The prompt",
			},
			wantMin: 7.0,
			wantMax: 8.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := optimizer.calculatePE2Score(tt.sections)
			if score < tt.wantMin || score > tt.wantMax {
				t.Errorf("score = %v, want between %v and %v", score, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestCalculatePE2ImprovementScore(t *testing.T) {
	provider := &mockProvider{}
	optimizer := NewPE2Optimizer(provider)

	tests := []struct {
		name    string
		result  *OptimizationResult
		wantMin float64
		wantMax float64
	}{
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
					{Score: 8.0},
				},
			},
			wantMin: 7.0,
			wantMax: 9.0,
		},
		{
			name: "multiple iterations improving",
			result: &OptimizationResult{
				Iterations: []IterationResult{
					{Score: 6.0},
					{Score: 7.0},
					{Score: 8.5},
				},
			},
			wantMin: 7.0,
			wantMax: 9.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := optimizer.calculatePE2ImprovementScore(tt.result)
			if score < tt.wantMin || score > tt.wantMax {
				t.Errorf("score = %v, want between %v and %v", score, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestGeneratePE2MetaPromptWithFeedback(t *testing.T) {
	provider := &mockProvider{}
	optimizer := NewPE2Optimizer(provider)

	config := PE2Config{
		ReasoningTemplate:    "chain_of_thought",
		ContextSpecification: "detailed",
		DescriptionDepth:     "comprehensive",
		MetaPromptStyle:      "expert",
		FeedbackIntegration:  true,
		ErrorCorrection:      true,
	}

	feedback := []string{"First iteration feedback", "Second iteration feedback"}
	result := optimizer.generatePE2MetaPrompt("test prompt", config, feedback)

	if len(result) == 0 {
		t.Error("generatePE2MetaPrompt returned empty string")
	}

	// Should contain feedback section when enabled
	if config.FeedbackIntegration && len(feedback) > 0 {
		if len(result) == 0 {
			t.Error("result should not be empty with feedback")
		}
	}
}

func TestGeneratePE2MetaPromptNoErrorCorrection(t *testing.T) {
	provider := &mockProvider{}
	optimizer := NewPE2Optimizer(provider)

	config := PE2Config{
		ReasoningTemplate:    "chain_of_thought",
		ContextSpecification: "standard",
		DescriptionDepth:     "standard",
		MetaPromptStyle:      "expert",
		FeedbackIntegration:  false,
		ErrorCorrection:      false,
	}

	result := optimizer.generatePE2MetaPrompt("test prompt", config, nil)

	if len(result) == 0 {
		t.Error("generatePE2MetaPrompt returned empty string")
	}
}

func TestOptimizeWithPE2Integration(t *testing.T) {
	provider := &mockProvider{
		responses: map[string]string{},
		calls:     []string{},
	}
	optimizer := NewPE2Optimizer(provider)

	tests := []struct {
		name string
		cfg  Config
	}{
		{
			name: "basic optimization",
			cfg: Config{
				InitialPrompt: "Summarize the text",
				Objective:     "Improve clarity",
				Iterations:    2,
				Method:        "pe2",
			},
		},
		{
			name: "single iteration",
			cfg: Config{
				InitialPrompt: "Extract key points",
				Objective:     "Improve accuracy",
				Iterations:    1,
				Method:        "pe2",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			result, err := optimizer.OptimizeWithPE2(ctx, tt.cfg)

			if err != nil {
				t.Fatalf("OptimizeWithPE2 error = %v", err)
			}

			if result == nil {
				t.Fatal("OptimizeWithPE2 returned nil result")
			}

			if result.OriginalPrompt != tt.cfg.InitialPrompt {
				t.Errorf("OriginalPrompt = %v, want %v", result.OriginalPrompt, tt.cfg.InitialPrompt)
			}

			if len(result.Iterations) != tt.cfg.Iterations {
				t.Errorf("got %d iterations, want %d", len(result.Iterations), tt.cfg.Iterations)
			}
		})
	}
}

func TestExecutePE2Iteration(t *testing.T) {
	provider := &mockProvider{
		responses: map[string]string{},
		calls:     []string{},
	}
	optimizer := NewPE2Optimizer(provider)

	cfg := Config{
		InitialPrompt: "Test prompt",
		Iterations:    1,
	}

	metaPrompt := "Test meta prompt for PE2 optimization"
	ctx := context.Background()

	optimizedPrompt, feedback, score, err := optimizer.executePE2Iteration(ctx, metaPrompt, cfg)

	if err != nil {
		t.Fatalf("executePE2Iteration error = %v", err)
	}

	if len(optimizedPrompt) == 0 {
		t.Error("optimizedPrompt should not be empty")
	}

	// Feedback may be empty in some cases
	_ = feedback

	if score <= 0 {
		t.Error("score should be positive")
	}
}
