package metaprompt

import (
	"testing"
)

func TestParseSemanticGradients(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantLen int
		wantErr bool
	}{
		{
			name: "valid json",
			input: `{
				"gradients": [
					{
						"component": "clarity",
						"direction": "improve",
						"magnitude": 0.8,
						"reasoning": "test",
						"confidence": 0.9
					}
				]
			}`,
			wantLen: 1,
			wantErr: false,
		},
		{
			name: "json with out of range values",
			input: `{
				"gradients": [
					{
						"component": "clarity",
						"direction": "improve",
						"magnitude": 1.5,
						"reasoning": "test",
						"confidence": -0.5
					}
				]
			}`,
			wantLen: 1,
			wantErr: false,
		},
		{
			name: "text format",
			input: `Component: clarity
Direction: improve
Magnitude: 0.8
Reasoning: test reason
Confidence: 0.9`,
			wantLen: 1,
			wantErr: false,
		},
		{
			name:    "empty text",
			input:   "",
			wantLen: 1, // Returns default gradient
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseSemanticGradients(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseSemanticGradients() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(result) != tt.wantLen {
				t.Errorf("parseSemanticGradients() returned %d gradients, want %d", len(result), tt.wantLen)
			}
		})
	}
}

func TestParseGradientsFromText(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantLen int
	}{
		{
			name: "single gradient",
			input: `Component: clarity
Direction: improve specificity
Magnitude: 0.8
Reasoning: needs more detail
Confidence: 0.9`,
			wantLen: 1,
		},
		{
			name: "multiple gradients",
			input: `Component: clarity
Direction: improve
Component: structure
Direction: reorganize`,
			wantLen: 2,
		},
		{
			name:    "no gradients",
			input:   "Some random text",
			wantLen: 1, // Returns default
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseGradientsFromText(tt.input)
			if err != nil {
				t.Errorf("parseGradientsFromText() error = %v", err)
				return
			}
			if len(result) != tt.wantLen {
				t.Errorf("parseGradientsFromText() returned %d gradients, want %d", len(result), tt.wantLen)
			}
		})
	}
}

func TestFormatGradients(t *testing.T) {
	tests := []struct {
		name      string
		gradients []SemanticGradient
		want      string
		check     func(string) bool
	}{
		{
			name:      "empty gradients",
			gradients: []SemanticGradient{},
			want:      "No gradients computed",
		},
		{
			name: "single gradient",
			gradients: []SemanticGradient{
				{
					Component:  "clarity",
					Direction:  "improve",
					Magnitude:  0.8,
					Confidence: 0.9,
				},
			},
			check: func(s string) bool {
				return len(s) > 0 && s != "No gradients computed"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatGradients(tt.gradients)
			if tt.want != "" && result != tt.want {
				t.Errorf("formatGradients() = %q, want %q", result, tt.want)
			}
			if tt.check != nil && !tt.check(result) {
				t.Errorf("formatGradients() check failed: %s", result)
			}
		})
	}
}

func TestFormatGradientsForApplication(t *testing.T) {
	tests := []struct {
		name      string
		gradients []SemanticGradient
		want      string
		check     func(string) bool
	}{
		{
			name:      "empty gradients",
			gradients: []SemanticGradient{},
			want:      "No gradients to apply",
		},
		{
			name: "with gradients",
			gradients: []SemanticGradient{
				{
					Component:  "clarity",
					Direction:  "improve",
					Magnitude:  0.8,
					Reasoning:  "needs improvement",
					Confidence: 0.9,
				},
			},
			check: func(s string) bool {
				return len(s) > 0 && s != "No gradients to apply"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatGradientsForApplication(tt.gradients)
			if tt.want != "" && result != tt.want {
				t.Errorf("formatGradientsForApplication() = %q, want %q", result, tt.want)
			}
			if tt.check != nil && !tt.check(result) {
				t.Errorf("formatGradientsForApplication() check failed")
			}
		})
	}
}

func TestParseScore(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    float64
		wantErr bool
	}{
		{
			name:    "direct float",
			input:   "0.85",
			want:    0.85,
			wantErr: false,
		},
		{
			name:    "fraction",
			input:   "8/10",
			want:    0.8,
			wantErr: false,
		},
		{
			name:    "percentage",
			input:   "85%",
			want:    0.85,
			wantErr: false,
		},
		{
			name:    "embedded decimal",
			input:   "The score is 0.75 points",
			want:    0.75,
			wantErr: false,
		},
		{
			name:    "no parseable value",
			input:   "no score here",
			want:    0.5, // default fallback
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseScore(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseScore() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if result != tt.want {
				t.Errorf("parseScore() = %v, want %v", result, tt.want)
			}
		})
	}
}

func TestNormalizeScore(t *testing.T) {
	tests := []struct {
		name  string
		input float64
		want  float64
	}{
		{name: "normal score", input: 0.75, want: 0.75},
		{name: "negative score", input: -0.5, want: 0.0},
		{name: "score out of 10", input: 8.0, want: 0.8},
		{name: "score out of 100", input: 85.0, want: 0.85},
		{name: "very high score", input: 150.0, want: 1.0},
		{name: "zero score", input: 0.0, want: 0.0},
		{name: "one score", input: 1.0, want: 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeScore(tt.input)
			if result != tt.want {
				t.Errorf("normalizeScore(%v) = %v, want %v", tt.input, result, tt.want)
			}
		})
	}
}

func TestParseSemanticAnalysis(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		check   func(*SemanticAnalysis) bool
	}{
		{
			name: "valid json",
			input: `{
				"components": [{"id": "c1", "type": "instruction"}],
				"flows": [{"source": "c1", "target": "c2", "strength": 0.8}],
				"coherence_score": 0.85,
				"ambiguities": ["some ambiguity"],
				"strengths": ["good structure"],
				"weaknesses": ["needs examples"]
			}`,
			wantErr: false,
			check: func(a *SemanticAnalysis) bool {
				return a.CoherenceScore == 0.85 && len(a.Components) == 1
			},
		},
		{
			name:    "no json",
			input:   "plain text without json",
			wantErr: false,
			check: func(a *SemanticAnalysis) bool {
				return a.CoherenceScore == 0.5 // default
			},
		},
		{
			name: "out of range coherence",
			input: `{
				"coherence_score": 1.5,
				"components": [],
				"flows": [],
				"ambiguities": [],
				"strengths": [],
				"weaknesses": []
			}`,
			wantErr: false,
			check: func(a *SemanticAnalysis) bool {
				return a.CoherenceScore == 1.0 // normalized
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseSemanticAnalysis(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseSemanticAnalysis() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.check != nil && !tt.check(result) {
				t.Errorf("parseSemanticAnalysis() check failed")
			}
		})
	}
}

func TestSemanticConfigStruct(t *testing.T) {
	config := SemanticConfig{
		Target:     "improve clarity",
		Iterations: 5,
		Verbose:    true,
		Provider:   "openai",
		Model:      "gpt-4",
	}

	if config.Iterations != 5 {
		t.Error("SemanticConfig.Iterations mismatch")
	}
}

func TestSemanticDescentConfigStruct(t *testing.T) {
	config := SemanticDescentConfig{
		Objective:            "maximize accuracy",
		LearningRate:         0.1,
		Iterations:           10,
		ConvergenceThreshold: 0.01,
		AdaptiveLearning:     true,
		Provider:             "openai",
		Model:                "gpt-4",
	}

	if config.LearningRate != 0.1 {
		t.Error("SemanticDescentConfig.LearningRate mismatch")
	}
}

func TestSemanticResultStruct(t *testing.T) {
	result := SemanticResult{
		InitialScore:    0.5,
		FinalScore:      0.85,
		OptimizedPrompt: "improved prompt",
		Iterations:      5,
		Converged:       true,
		SemanticGradients: []SemanticGradient{
			{Component: "clarity", Magnitude: 0.8},
		},
	}

	improvement := result.FinalScore - result.InitialScore
	if improvement < 0.3 {
		t.Errorf("Improvement = %v, expected > 0.3", improvement)
	}
}

func TestSemanticGradientStruct(t *testing.T) {
	gradient := SemanticGradient{
		Component:  "instruction clarity",
		Direction:  "more specific",
		Magnitude:  0.8,
		Reasoning:  "current instruction is vague",
		Confidence: 0.9,
	}

	if gradient.Magnitude > 1.0 || gradient.Magnitude < 0 {
		t.Error("Magnitude should be between 0 and 1")
	}
	if gradient.Confidence > 1.0 || gradient.Confidence < 0 {
		t.Error("Confidence should be between 0 and 1")
	}
}

func TestSemanticOptimizationStepStruct(t *testing.T) {
	step := SemanticOptimizationStep{
		Iteration:   1,
		Score:       0.75,
		Prompt:      "test prompt",
		Gradient:    "improve clarity",
		Improvement: 0.1,
	}

	if step.Iteration != 1 {
		t.Error("Step.Iteration mismatch")
	}
}

func TestSemanticNodeStruct(t *testing.T) {
	node := SemanticNode{
		ID:           "node_1",
		Type:         "component",
		Content:      "test content",
		Dependencies: []string{"node_0"},
		Properties:   map[string]interface{}{"key": "value"},
	}

	if node.Type != "component" {
		t.Error("SemanticNode.Type mismatch")
	}
}

func TestSemanticFlowStruct(t *testing.T) {
	flow := SemanticFlow{
		Source:      "node_1",
		Target:      "node_2",
		FlowType:    "data",
		Strength:    0.8,
		Information: "some info",
	}

	if flow.Strength < 0 || flow.Strength > 1 {
		t.Error("Flow.Strength should be between 0 and 1")
	}
}

func TestSemanticAnalysisStruct(t *testing.T) {
	analysis := SemanticAnalysis{
		Components:     []SemanticNode{{ID: "c1"}},
		Flows:          []SemanticFlow{{Source: "c1", Target: "c2"}},
		CoherenceScore: 0.85,
		Ambiguities:    []string{"ambiguity1"},
		Strengths:      []string{"strength1"},
		Weaknesses:     []string{"weakness1"},
	}

	if len(analysis.Components) != 1 {
		t.Error("Analysis.Components length mismatch")
	}
}
