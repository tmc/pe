package optimization

import (
	"testing"
	"time"
)

func TestGenerationOptions_Struct(t *testing.T) {
	temp := 0.7
	maxTokens := 1000
	topP := 0.9
	topK := 50

	opts := GenerationOptions{
		Temperature: &temp,
		MaxTokens:   &maxTokens,
		TopP:        &topP,
		TopK:        &topK,
		Stop:        []string{"END", "STOP"},
	}

	if *opts.Temperature != 0.7 {
		t.Error("Temperature mismatch")
	}
	if *opts.MaxTokens != 1000 {
		t.Error("MaxTokens mismatch")
	}
	if *opts.TopP != 0.9 {
		t.Error("TopP mismatch")
	}
	if *opts.TopK != 50 {
		t.Error("TopK mismatch")
	}
	if len(opts.Stop) != 2 {
		t.Error("Stop length mismatch")
	}
}

func TestGenerationResponse_Struct(t *testing.T) {
	resp := GenerationResponse{
		Text:             "Generated text",
		PromptTokens:     100,
		CompletionTokens: 50,
		TotalTokens:      150,
		Latency:          100 * time.Millisecond,
		Cost:             0.01,
		Model:            "gpt-4",
		FinishReason:     "stop",
	}

	if resp.Text != "Generated text" {
		t.Error("Text mismatch")
	}
	if resp.PromptTokens != 100 {
		t.Error("PromptTokens mismatch")
	}
	if resp.TotalTokens != 150 {
		t.Error("TotalTokens mismatch")
	}
	if resp.Model != "gpt-4" {
		t.Error("Model mismatch")
	}
}

func TestComparisonResult_Struct(t *testing.T) {
	result := ComparisonResult{
		Winner:    1,
		Score1:    0.85,
		Score2:    0.75,
		Reasoning: "First prompt is more concise",
	}

	if result.Winner != 1 {
		t.Error("Winner mismatch")
	}
	if result.Score1 != 0.85 {
		t.Error("Score1 mismatch")
	}
	if result.Score2 != 0.75 {
		t.Error("Score2 mismatch")
	}
}

func TestOptimizationConfig_Struct(t *testing.T) {
	config := OptimizationConfig{
		InitialPrompt:        "Summarize this text",
		Objective:            "clarity",
		Iterations:           5,
		MaxIterations:        10,
		ConvergenceThreshold: 0.01,
		Method:               "textgrad",
		OptimizationType:     "single",
		Temperature:          0.7,
		MaxTokens:            1000,
		MultiObjective:       true,
		Objectives: []ObjectiveSpec{
			{Name: "clarity", Weight: 0.5},
			{Name: "conciseness", Weight: 0.5},
		},
		Provider: "openai",
		Model:    "gpt-4",
	}

	if config.InitialPrompt != "Summarize this text" {
		t.Error("InitialPrompt mismatch")
	}
	if config.Method != "textgrad" {
		t.Error("Method mismatch")
	}
	if !config.MultiObjective {
		t.Error("MultiObjective should be true")
	}
	if len(config.Objectives) != 2 {
		t.Error("Objectives length mismatch")
	}
}

func TestObjectiveSpec_Struct(t *testing.T) {
	spec := ObjectiveSpec{
		Name:        "accuracy",
		Description: "Improve factual accuracy",
		Weight:      0.8,
		Target:      0.95,
		Type:        "maximize",
	}

	if spec.Name != "accuracy" {
		t.Error("Name mismatch")
	}
	if spec.Weight != 0.8 {
		t.Error("Weight mismatch")
	}
	if spec.Target != 0.95 {
		t.Error("Target mismatch")
	}
	if spec.Type != "maximize" {
		t.Error("Type mismatch")
	}
}

func TestOptimizationResult_Struct(t *testing.T) {
	result := OptimizationResult{
		OriginalPrompt:  "Original",
		OptimizedPrompt: "Optimized",
		Iterations: []IterationResult{
			{Iteration: 1, Score: 0.7},
			{Iteration: 2, Score: 0.8},
		},
		ImprovementScore: 0.15,
		TotalDuration:    5 * time.Second,
		CreatedAt:        time.Now(),
		Method:           "pe2",
		Converged:        true,
	}

	if result.OriginalPrompt != "Original" {
		t.Error("OriginalPrompt mismatch")
	}
	if result.OptimizedPrompt != "Optimized" {
		t.Error("OptimizedPrompt mismatch")
	}
	if len(result.Iterations) != 2 {
		t.Error("Iterations length mismatch")
	}
	if !result.Converged {
		t.Error("Converged should be true")
	}
}

func TestIterationResult_Struct(t *testing.T) {
	result := IterationResult{
		Iteration: 3,
		Prompt:    "Test prompt",
		Score:     0.85,
		Feedback:  "Good improvement",
		Suggestions: []string{
			"Try being more specific",
			"Add examples",
		},
		Duration:  200 * time.Millisecond,
		Changes:   []string{"Rewrote intro", "Added conclusion"},
		Timestamp: time.Now(),
		Gradients: []Gradient{
			{Component: "intro", Magnitude: 0.1},
		},
	}

	if result.Iteration != 3 {
		t.Error("Iteration mismatch")
	}
	if result.Score != 0.85 {
		t.Error("Score mismatch")
	}
	if len(result.Suggestions) != 2 {
		t.Error("Suggestions length mismatch")
	}
	if len(result.Gradients) != 1 {
		t.Error("Gradients length mismatch")
	}
}

func TestGradient_Struct(t *testing.T) {
	grad := Gradient{
		Component:   "introduction",
		Feedback:    "Too verbose",
		Suggestions: []string{"Shorten", "Focus on key points"},
		Confidence:  0.9,
		Priority:    0.8,
		Direction:   "improve",
		Magnitude:   0.15,
	}

	if grad.Component != "introduction" {
		t.Error("Component mismatch")
	}
	if grad.Confidence != 0.9 {
		t.Error("Confidence mismatch")
	}
	if grad.Magnitude != 0.15 {
		t.Error("Magnitude mismatch")
	}
}

func TestSystemDefinition_Struct(t *testing.T) {
	sys := SystemDefinition{
		Name:        "multi-agent",
		Description: "Multi-agent system",
		Components: []SystemComponent{
			{ID: "agent1", Name: "Parser", Type: "agent"},
			{ID: "agent2", Name: "Writer", Type: "agent"},
		},
		Dependencies: []ComponentDependency{
			{From: "agent1", To: "agent2", Type: "data"},
		},
		Objectives: []ObjectiveSpec{
			{Name: "quality", Weight: 1.0},
		},
		Constraints: []SystemConstraint{
			{Name: "max_tokens", Type: "hard", Value: 4000},
		},
	}

	if sys.Name != "multi-agent" {
		t.Error("Name mismatch")
	}
	if len(sys.Components) != 2 {
		t.Error("Components length mismatch")
	}
	if len(sys.Dependencies) != 1 {
		t.Error("Dependencies length mismatch")
	}
}

func TestSystemComponent_Struct(t *testing.T) {
	comp := SystemComponent{
		ID:      "comp1",
		Name:    "Analyzer",
		Type:    "tool",
		Content: "Analysis logic",
		Parameters: map[string]interface{}{
			"threshold": 0.5,
		},
		Performance: 0.88,
	}

	if comp.ID != "comp1" {
		t.Error("ID mismatch")
	}
	if comp.Type != "tool" {
		t.Error("Type mismatch")
	}
	if comp.Performance != 0.88 {
		t.Error("Performance mismatch")
	}
}

func TestComponentDependency_Struct(t *testing.T) {
	dep := ComponentDependency{
		From:        "source",
		To:          "target",
		Type:        "control",
		Weight:      0.7,
		Description: "Control flow",
	}

	if dep.From != "source" {
		t.Error("From mismatch")
	}
	if dep.To != "target" {
		t.Error("To mismatch")
	}
	if dep.Weight != 0.7 {
		t.Error("Weight mismatch")
	}
}

func TestSystemConstraint_Struct(t *testing.T) {
	constraint := SystemConstraint{
		Name:        "latency",
		Description: "Maximum latency",
		Type:        "soft",
		Value:       1000,
	}

	if constraint.Name != "latency" {
		t.Error("Name mismatch")
	}
	if constraint.Type != "soft" {
		t.Error("Type mismatch")
	}
}

func TestSystemGradient_Struct(t *testing.T) {
	sysGrad := SystemGradient{
		ComponentID: "comp1",
		Gradient: Gradient{
			Component: "intro",
			Magnitude: 0.2,
		},
		Priority: 0.9,
	}

	if sysGrad.ComponentID != "comp1" {
		t.Error("ComponentID mismatch")
	}
	if sysGrad.Priority != 0.9 {
		t.Error("Priority mismatch")
	}
}

func TestSystemOptimizationResult_Struct(t *testing.T) {
	result := SystemOptimizationResult{
		OptimizedComponents: []SystemComponent{
			{ID: "comp1", Performance: 0.9},
		},
		OverallPerformance: 0.88,
		ObjectiveScores: map[string]float64{
			"quality": 0.9,
			"speed":   0.85,
		},
		ParetoEfficient: true,
		OptimizationHistory: []SystemIterationResult{
			{Iteration: 1, Performance: 0.8},
		},
		Iterations: 5,
		Duration:   10 * time.Second,
		ConvergenceAnalysis: &ConvergenceAnalysis{
			Converged:       true,
			ConvergenceRate: 0.05,
		},
	}

	if len(result.OptimizedComponents) != 1 {
		t.Error("OptimizedComponents length mismatch")
	}
	if result.OverallPerformance != 0.88 {
		t.Error("OverallPerformance mismatch")
	}
	if !result.ParetoEfficient {
		t.Error("ParetoEfficient should be true")
	}
	if result.ConvergenceAnalysis == nil {
		t.Error("ConvergenceAnalysis should not be nil")
	}
}

func TestSystemIterationResult_Struct(t *testing.T) {
	result := SystemIterationResult{
		Iteration:   2,
		Performance: 0.82,
		Objectives: map[string]float64{
			"quality": 0.85,
		},
		Gradients: []SystemGradient{
			{ComponentID: "c1", Priority: 0.8},
		},
		Changes: []ComponentChange{
			{ComponentID: "c1", ChangeType: "modify"},
		},
		Improvement: 0.05,
	}

	if result.Iteration != 2 {
		t.Error("Iteration mismatch")
	}
	if result.Performance != 0.82 {
		t.Error("Performance mismatch")
	}
	if result.Improvement != 0.05 {
		t.Error("Improvement mismatch")
	}
}

func TestComponentChange_Struct(t *testing.T) {
	change := ComponentChange{
		ComponentID: "comp1",
		ChangeType:  "replace",
		OldValue:    "old",
		NewValue:    "new",
		Impact:      0.1,
	}

	if change.ComponentID != "comp1" {
		t.Error("ComponentID mismatch")
	}
	if change.ChangeType != "replace" {
		t.Error("ChangeType mismatch")
	}
	if change.Impact != 0.1 {
		t.Error("Impact mismatch")
	}
}

func TestConvergenceAnalysis_Struct(t *testing.T) {
	analysis := ConvergenceAnalysis{
		Converged:       true,
		ConvergenceRate: 0.02,
		FinalGradient:   0.001,
		StabilityMetric: 0.95,
	}

	if !analysis.Converged {
		t.Error("Converged should be true")
	}
	if analysis.ConvergenceRate != 0.02 {
		t.Error("ConvergenceRate mismatch")
	}
	if analysis.StabilityMetric != 0.95 {
		t.Error("StabilityMetric mismatch")
	}
}
