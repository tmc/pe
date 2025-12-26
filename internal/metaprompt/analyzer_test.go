package metaprompt

import (
	"testing"
)

func TestNewTextGradAnalyzer(t *testing.T) {
	provider := &mockProvider{}
	analyzer := NewTextGradAnalyzer(provider)

	if analyzer == nil {
		t.Fatal("NewTextGradAnalyzer returned nil")
	}
	if analyzer.llm == nil {
		t.Error("TextGradAnalyzer has nil llm provider")
	}
}

func TestComputeGradientStrength(t *testing.T) {
	tests := []struct {
		name    string
		flows   []AttentionFlow
		drifts  []SemanticDrift
		metrics CoherenceMetrics
		want    float64
		wantMin float64
		wantMax float64
	}{
		{
			name: "high quality",
			flows: []AttentionFlow{
				{Weight: 0.9},
				{Weight: 0.8},
			},
			drifts: []SemanticDrift{
				{DriftScore: 0.1},
			},
			metrics: CoherenceMetrics{
				LocalCoherence:  0.9,
				GlobalCoherence: 0.85,
				LogicalFlow:     0.9,
				Consistency:     0.88,
			},
			wantMin: 0.7,
			wantMax: 1.0,
		},
		{
			name:   "empty flows",
			flows:  []AttentionFlow{},
			drifts: []SemanticDrift{},
			metrics: CoherenceMetrics{
				LocalCoherence:  0.5,
				GlobalCoherence: 0.5,
				LogicalFlow:     0.5,
				Consistency:     0.5,
			},
			wantMin: 0.0,
			wantMax: 0.6,
		},
		{
			name: "high drift",
			flows: []AttentionFlow{
				{Weight: 0.5},
			},
			drifts: []SemanticDrift{
				{DriftScore: 0.8},
				{DriftScore: 0.9},
			},
			metrics: CoherenceMetrics{
				LocalCoherence:  0.3,
				GlobalCoherence: 0.3,
				LogicalFlow:     0.3,
				Consistency:     0.3,
			},
			wantMin: 0.0,
			wantMax: 0.4,
		},
	}

	provider := &mockProvider{}
	analyzer := NewTextGradAnalyzer(provider)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := analyzer.computeGradientStrength(tt.flows, tt.drifts, tt.metrics)
			if got < tt.wantMin || got > tt.wantMax {
				t.Errorf("computeGradientStrength() = %v, want between %v and %v", got, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestParseAttentionFlows(t *testing.T) {
	provider := &mockProvider{}
	analyzer := NewTextGradAnalyzer(provider)

	// Test fallback parsing
	result := analyzer.parseAttentionFlows("unparseable text")
	if len(result) != 1 {
		t.Errorf("parseAttentionFlows() returned %d flows, want 1", len(result))
	}
	if result[0].SourceToken != "prompt_instruction" {
		t.Error("parseAttentionFlows() fallback has wrong source token")
	}
}

func TestParseSemanticDrifts(t *testing.T) {
	provider := &mockProvider{}
	analyzer := NewTextGradAnalyzer(provider)

	result := analyzer.parseSemanticDrifts("unparseable text")
	if len(result) != 1 {
		t.Errorf("parseSemanticDrifts() returned %d drifts, want 1", len(result))
	}
	if result[0].DriftType != "semantic" {
		t.Error("parseSemanticDrifts() fallback has wrong drift type")
	}
}

func TestParseCoherenceMetrics(t *testing.T) {
	provider := &mockProvider{}
	analyzer := NewTextGradAnalyzer(provider)

	result := analyzer.parseCoherenceMetrics("unparseable text")
	if result.LocalCoherence != 0.7 {
		t.Errorf("parseCoherenceMetrics() LocalCoherence = %v, want 0.7", result.LocalCoherence)
	}
}

func TestParseOptimizationHints(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int
	}{
		{
			name: "numbered hints",
			input: `1. First hint
2. Second hint
3. Third hint
4. Fourth hint
5. Fifth hint`,
			want: 5,
		},
		{
			name:  "no hints",
			input: "Some random text without numbered items",
			want:  0,
		},
		{
			name: "partial hints",
			input: `1. Only hint
Some other text`,
			want: 1,
		},
	}

	provider := &mockProvider{}
	analyzer := NewTextGradAnalyzer(provider)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzer.parseOptimizationHints(tt.input)
			if len(result) != tt.want {
				t.Errorf("parseOptimizationHints() returned %d hints, want %d", len(result), tt.want)
			}
		})
	}
}

func TestSummarizeIssues(t *testing.T) {
	tests := []struct {
		name     string
		analysis *AnalysisResult
		want     string
		check    func(string) bool
	}{
		{
			name: "low gradient strength",
			analysis: &AnalysisResult{
				GradientStrength: 0.4,
				SemanticDrifts:   []SemanticDrift{},
				CoherenceMetrics: CoherenceMetrics{
					LocalCoherence:  0.8,
					GlobalCoherence: 0.8,
					LogicalFlow:     0.8,
					Consistency:     0.8,
				},
			},
			check: func(s string) bool {
				return s != "No significant issues identified"
			},
		},
		{
			name: "with semantic drifts",
			analysis: &AnalysisResult{
				GradientStrength: 0.8,
				SemanticDrifts:   []SemanticDrift{{}, {}},
				CoherenceMetrics: CoherenceMetrics{
					LocalCoherence:  0.8,
					GlobalCoherence: 0.8,
					LogicalFlow:     0.8,
					Consistency:     0.8,
				},
			},
			check: func(s string) bool {
				return s != "No significant issues identified"
			},
		},
		{
			name: "low coherence",
			analysis: &AnalysisResult{
				GradientStrength: 0.8,
				SemanticDrifts:   []SemanticDrift{},
				CoherenceMetrics: CoherenceMetrics{
					LocalCoherence:  0.5,
					GlobalCoherence: 0.5,
					LogicalFlow:     0.5,
					Consistency:     0.5,
				},
			},
			check: func(s string) bool {
				return s != "No significant issues identified"
			},
		},
		{
			name: "no issues",
			analysis: &AnalysisResult{
				GradientStrength: 0.8,
				SemanticDrifts:   []SemanticDrift{},
				CoherenceMetrics: CoherenceMetrics{
					LocalCoherence:  0.9,
					GlobalCoherence: 0.9,
					LogicalFlow:     0.9,
					Consistency:     0.9,
				},
			},
			want: "No significant issues identified",
		},
	}

	provider := &mockProvider{}
	analyzer := NewTextGradAnalyzer(provider)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzer.summarizeIssues(tt.analysis)
			if tt.want != "" && result != tt.want {
				t.Errorf("summarizeIssues() = %q, want %q", result, tt.want)
			}
			if tt.check != nil && !tt.check(result) {
				t.Errorf("summarizeIssues() check failed: %s", result)
			}
		})
	}
}

func TestAttentionFlowStruct(t *testing.T) {
	flow := AttentionFlow{
		SourceToken: "source",
		TargetToken: "target",
		Weight:      0.85,
		Metadata:    map[string]string{"type": "test"},
	}

	if flow.SourceToken != "source" {
		t.Error("AttentionFlow.SourceToken mismatch")
	}
	if flow.Weight != 0.85 {
		t.Error("AttentionFlow.Weight mismatch")
	}
}

func TestSemanticDriftStruct(t *testing.T) {
	drift := SemanticDrift{
		ConceptID:     "concept_1",
		OriginalValue: "original",
		CurrentValue:  "current",
		DriftScore:    0.5,
		DriftType:     "semantic",
	}

	if drift.DriftType != "semantic" {
		t.Error("SemanticDrift.DriftType mismatch")
	}
}

func TestCoherenceMetricsStruct(t *testing.T) {
	metrics := CoherenceMetrics{
		LocalCoherence:  0.8,
		GlobalCoherence: 0.75,
		LogicalFlow:     0.9,
		Consistency:     0.85,
	}

	avg := (metrics.LocalCoherence + metrics.GlobalCoherence +
		metrics.LogicalFlow + metrics.Consistency) / 4.0

	if avg < 0.8 || avg > 0.9 {
		t.Errorf("Average coherence = %v, expected ~0.825", avg)
	}
}

func TestAnalysisResultStruct(t *testing.T) {
	result := AnalysisResult{
		AttentionFlows: []AttentionFlow{
			{Weight: 0.8},
		},
		SemanticDrifts: []SemanticDrift{
			{DriftScore: 0.2},
		},
		CoherenceMetrics:  CoherenceMetrics{LocalCoherence: 0.9},
		GradientStrength:  0.85,
		OptimizationHints: []string{"hint1", "hint2"},
	}

	if len(result.AttentionFlows) != 1 {
		t.Error("AnalysisResult.AttentionFlows length mismatch")
	}
	if len(result.OptimizationHints) != 2 {
		t.Error("AnalysisResult.OptimizationHints length mismatch")
	}
}

func TestFloatPtrHelper(t *testing.T) {
	val := floatPtr(0.5)
	if val == nil {
		t.Fatal("floatPtr returned nil")
	}
	if *val != 0.5 {
		t.Errorf("floatPtr value = %v, want 0.5", *val)
	}
}

func TestIntPtrHelper(t *testing.T) {
	val := intPtr(100)
	if val == nil {
		t.Fatal("intPtr returned nil")
	}
	if *val != 100 {
		t.Errorf("intPtr value = %v, want 100", *val)
	}
}
