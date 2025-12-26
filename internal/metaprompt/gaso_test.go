package metaprompt

import (
	"math"
	"testing"
)

func TestNewGASOOptimizer(t *testing.T) {
	provider := &mockProvider{}
	optimizer := NewGASOOptimizer(provider)

	if optimizer == nil {
		t.Fatal("NewGASOOptimizer returned nil")
	}
	if optimizer.llm == nil {
		t.Error("GASOOptimizer has nil llm provider")
	}
}

func TestBuildComputationalGraph(t *testing.T) {
	provider := &mockProvider{}
	optimizer := NewGASOOptimizer(provider)

	system := &SystemDefinition{
		Name:        "test-system",
		Description: "A test system",
		Components: []SystemComponent{
			{ID: "comp1", Name: "Component 1", Type: "prompt", Content: "Content 1"},
			{ID: "comp2", Name: "Component 2", Type: "agent", Content: "Content 2"},
		},
		Dependencies: []ComponentDependency{
			{From: "comp1", To: "comp2", Type: "input", Weight: 0.8, Description: "comp1 provides input to comp2"},
		},
	}

	graph, err := optimizer.buildComputationalGraph(system)
	if err != nil {
		t.Fatalf("buildComputationalGraph error = %v", err)
	}

	if graph == nil {
		t.Fatal("buildComputationalGraph returned nil")
	}

	if len(graph.Nodes) != 2 {
		t.Errorf("expected 2 nodes, got %d", len(graph.Nodes))
	}

	if len(graph.Edges) != 1 {
		t.Errorf("expected 1 edge, got %d", len(graph.Edges))
	}

	// Check node data
	for _, node := range graph.Nodes {
		if node.ID == "" {
			t.Error("node has empty ID")
		}
		if node.Type == "" {
			t.Error("node has empty Type")
		}
		if node.Position == nil {
			t.Error("node has nil Position")
		}
	}

	// Check edge data
	for _, edge := range graph.Edges {
		if edge.From != "comp1" {
			t.Errorf("edge.From = %v, want comp1", edge.From)
		}
		if edge.To != "comp2" {
			t.Errorf("edge.To = %v, want comp2", edge.To)
		}
	}
}

func TestCalculateComponentPriority(t *testing.T) {
	provider := &mockProvider{}
	optimizer := NewGASOOptimizer(provider)

	tests := []struct {
		name      string
		component SystemComponent
		system    *SystemDefinition
		wantMin   float64
		wantMax   float64
	}{
		{
			name: "isolated component",
			component: SystemComponent{
				ID:   "isolated",
				Type: "prompt",
			},
			system: &SystemDefinition{
				Dependencies: []ComponentDependency{},
			},
			wantMin: 0.5,
			wantMax: 0.7,
		},
		{
			name: "agent component with dependencies",
			component: SystemComponent{
				ID:   "central",
				Type: "agent",
			},
			system: &SystemDefinition{
				Dependencies: []ComponentDependency{
					{From: "central", To: "other1", Weight: 0.8},
					{From: "central", To: "other2", Weight: 0.9},
					{From: "input", To: "central", Weight: 0.7},
				},
			},
			wantMin: 0.7,
			wantMax: 1.0,
		},
		{
			name: "workflow component",
			component: SystemComponent{
				ID:   "workflow",
				Type: "workflow",
			},
			system: &SystemDefinition{
				Dependencies: []ComponentDependency{},
			},
			wantMin: 0.6,
			wantMax: 0.8,
		},
		{
			name: "tool component",
			component: SystemComponent{
				ID:   "tool",
				Type: "tool",
			},
			system: &SystemDefinition{
				Dependencies: []ComponentDependency{},
			},
			wantMin: 0.5,
			wantMax: 0.7,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			priority := optimizer.calculateComponentPriority(tt.component, tt.system)
			if priority < tt.wantMin || priority > tt.wantMax {
				t.Errorf("calculateComponentPriority() = %v, want between %v and %v", priority, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestFormatDependencies(t *testing.T) {
	provider := &mockProvider{}
	optimizer := NewGASOOptimizer(provider)

	system := &SystemDefinition{
		Dependencies: []ComponentDependency{
			{From: "comp1", To: "comp2", Type: "input"},
			{From: "comp2", To: "comp3", Type: "data"},
			{From: "comp1", To: "comp3", Type: "control"},
		},
	}

	result := optimizer.formatDependencies("comp1", system)

	// comp1 appears in two dependencies
	if len(result) == 0 {
		t.Error("formatDependencies returned empty string")
	}
}

func TestFormatComponents(t *testing.T) {
	provider := &mockProvider{}
	optimizer := NewGASOOptimizer(provider)

	components := []SystemComponent{
		{Name: "Component 1", Type: "prompt", Content: "Content 1"},
		{Name: "Component 2", Type: "agent", Content: "Content 2"},
	}

	result := optimizer.formatComponents(components)

	if len(result) == 0 {
		t.Error("formatComponents returned empty string")
	}
}

func TestFormatAllDependencies(t *testing.T) {
	provider := &mockProvider{}
	optimizer := NewGASOOptimizer(provider)

	system := &SystemDefinition{
		Dependencies: []ComponentDependency{
			{From: "comp1", To: "comp2", Type: "input", Weight: 0.8},
		},
	}

	result := optimizer.formatAllDependencies(system)

	if len(result) == 0 {
		t.Error("formatAllDependencies returned empty string")
	}
}

func TestAnalyzeConvergence(t *testing.T) {
	provider := &mockProvider{}
	optimizer := NewGASOOptimizer(provider)

	tests := []struct {
		name    string
		history []GASOIteration
		wantConverged bool
	}{
		{
			name:    "empty history",
			history: []GASOIteration{},
			wantConverged: false,
		},
		{
			name: "single iteration",
			history: []GASOIteration{
				{Iteration: 1, Performance: 0.5, Improvement: 0.1},
			},
			wantConverged: false,
		},
		{
			name: "converging history",
			history: []GASOIteration{
				{Iteration: 1, Performance: 0.5, Improvement: 0.1},
				{Iteration: 2, Performance: 0.6, Improvement: 0.05},
				{Iteration: 3, Performance: 0.65, Improvement: 0.02},
				{Iteration: 4, Performance: 0.67, Improvement: 0.005},
				{Iteration: 5, Performance: 0.675, Improvement: 0.005},
				{Iteration: 6, Performance: 0.68, Improvement: 0.005},
				{Iteration: 7, Performance: 0.68, Improvement: 0.005},
				{Iteration: 8, Performance: 0.68, Improvement: 0.005},
			},
			wantConverged: true,
		},
		{
			name: "non-converging history",
			history: []GASOIteration{
				{Iteration: 1, Performance: 0.5, Improvement: 0.1},
				{Iteration: 2, Performance: 0.6, Improvement: 0.1},
				{Iteration: 3, Performance: 0.7, Improvement: 0.1},
			},
			wantConverged: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := optimizer.analyzeConvergence(tt.history)
			if result == nil {
				t.Fatal("analyzeConvergence returned nil")
			}
			if result.Converged != tt.wantConverged {
				t.Errorf("Converged = %v, want %v", result.Converged, tt.wantConverged)
			}
		})
	}
}

func TestCheckParetoEfficiency(t *testing.T) {
	provider := &mockProvider{}
	optimizer := NewGASOOptimizer(provider)

	tests := []struct {
		name       string
		scores     map[string]float64
		objectives []SystemObjective
		want       bool
	}{
		{
			name:       "empty objectives",
			scores:     map[string]float64{},
			objectives: []SystemObjective{},
			want:       true,
		},
		{
			name: "maximize objective met",
			scores: map[string]float64{
				"accuracy": 0.95,
			},
			objectives: []SystemObjective{
				{Name: "accuracy", Type: "maximize", Target: 0.9},
			},
			want: true,
		},
		{
			name: "maximize objective not met",
			scores: map[string]float64{
				"accuracy": 0.8,
			},
			objectives: []SystemObjective{
				{Name: "accuracy", Type: "maximize", Target: 0.9},
			},
			want: false,
		},
		{
			name: "minimize objective met",
			scores: map[string]float64{
				"latency": 0.1,
			},
			objectives: []SystemObjective{
				{Name: "latency", Type: "minimize", Target: 0.2},
			},
			want: true,
		},
		{
			name: "minimize objective not met",
			scores: map[string]float64{
				"latency": 0.3,
			},
			objectives: []SystemObjective{
				{Name: "latency", Type: "minimize", Target: 0.2},
			},
			want: false,
		},
		{
			name: "target objective met within tolerance",
			scores: map[string]float64{
				"precision": 0.85,
			},
			objectives: []SystemObjective{
				{Name: "precision", Type: "target", Target: 0.9},
			},
			want: true,
		},
		{
			name: "target objective not met outside tolerance",
			scores: map[string]float64{
				"precision": 0.5,
			},
			objectives: []SystemObjective{
				{Name: "precision", Type: "target", Target: 0.9},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := optimizer.checkParetoEfficiency(tt.scores, tt.objectives)
			if result != tt.want {
				t.Errorf("checkParetoEfficiency() = %v, want %v", result, tt.want)
			}
		})
	}
}

func TestParseComponentGradient(t *testing.T) {
	tests := []struct {
		name     string
		response string
		wantErr  bool
	}{
		{
			name: "valid JSON gradient",
			response: `{
				"component": "clarity",
				"direction": "improve",
				"magnitude": 0.8,
				"reasoning": "needs clarity",
				"confidence": 0.9
			}`,
			wantErr: false,
		},
		{
			name:     "no JSON in response",
			response: "This is just text without any JSON",
			wantErr:  true,
		},
		{
			name:     "invalid JSON",
			response: `{ invalid json }`,
			wantErr:  true,
		},
		{
			name: "magnitude out of range (normalized)",
			response: `{
				"component": "test",
				"direction": "improve",
				"magnitude": 1.5,
				"confidence": -0.5
			}`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gradient, err := parseComponentGradient(tt.response)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseComponentGradient() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				// Check that values are normalized
				if gradient.Magnitude < 0 || gradient.Magnitude > 1 {
					t.Errorf("magnitude = %v, should be in [0,1]", gradient.Magnitude)
				}
				if gradient.Confidence < 0 || gradient.Confidence > 1 {
					t.Errorf("confidence = %v, should be in [0,1]", gradient.Confidence)
				}
			}
		})
	}
}

func TestSystemGradientStruct(t *testing.T) {
	gradient := SystemGradient{
		ComponentID: "comp1",
		Gradient: SemanticGradient{
			Component:  "clarity",
			Direction:  "improve",
			Magnitude:  0.8,
			Confidence: 0.9,
		},
		Priority: 0.7,
	}

	if gradient.ComponentID != "comp1" {
		t.Error("SystemGradient.ComponentID mismatch")
	}
	if gradient.Priority != 0.7 {
		t.Error("SystemGradient.Priority mismatch")
	}
}

func TestGASOGraphNodeStruct(t *testing.T) {
	node := GASOGraphNode{
		ID:   "node1",
		Type: "prompt",
		Data: map[string]interface{}{
			"name": "Test Node",
		},
		Position: map[string]float64{
			"x": 100,
			"y": 200,
		},
	}

	if node.ID != "node1" {
		t.Error("GASOGraphNode.ID mismatch")
	}
	if node.Position["x"] != 100 {
		t.Error("GASOGraphNode.Position[x] mismatch")
	}
}

func TestGASOGraphEdgeStruct(t *testing.T) {
	edge := GASOGraphEdge{
		From:   "node1",
		To:     "node2",
		Type:   "input",
		Weight: 0.8,
		Data:   map[string]interface{}{"description": "test edge"},
	}

	if edge.From != "node1" {
		t.Error("GASOGraphEdge.From mismatch")
	}
	if edge.Weight != 0.8 {
		t.Error("GASOGraphEdge.Weight mismatch")
	}
}

func TestGASOStructs(t *testing.T) {
	// Test SystemDefinition
	system := SystemDefinition{
		Name:        "test",
		Description: "Test system",
		Components: []SystemComponent{
			{ID: "c1", Name: "Component 1", Type: "prompt"},
		},
		Dependencies: []ComponentDependency{
			{From: "c1", To: "c2", Type: "input"},
		},
		Objectives: []SystemObjective{
			{Name: "accuracy", Type: "maximize", Target: 0.9},
		},
		Constraints: []SystemConstraint{
			{Name: "latency", Type: "hard", Value: 1000},
		},
	}

	if system.Name != "test" {
		t.Error("SystemDefinition.Name mismatch")
	}
	if len(system.Components) != 1 {
		t.Error("SystemDefinition.Components length mismatch")
	}

	// Test GASOConfig
	config := GASOConfig{
		Objective:        "improve performance",
		Iterations:       10,
		OptimizationType: "pareto",
		MultiObjective:   true,
	}

	if config.Iterations != 10 {
		t.Error("GASOConfig.Iterations mismatch")
	}

	// Test GASOResult
	result := GASOResult{
		OverallPerformance:  0.85,
		ObjectiveScores:     map[string]float64{"accuracy": 0.9},
		ParetoEfficient:     true,
		OptimizationHistory: []GASOIteration{},
		Iterations:          10,
	}

	if math.Abs(result.OverallPerformance-0.85) > 0.001 {
		t.Error("GASOResult.OverallPerformance mismatch")
	}

	// Test GASOIteration
	iteration := GASOIteration{
		Iteration:   1,
		Performance: 0.8,
		Improvement: 0.1,
		Changes:     []ComponentChange{},
	}

	if iteration.Iteration != 1 {
		t.Error("GASOIteration.Iteration mismatch")
	}

	// Test ComponentChange
	change := ComponentChange{
		ComponentID: "c1",
		ChangeType:  "content_optimization",
		OldValue:    "old",
		NewValue:    "new",
		Impact:      0.5,
	}

	if change.ComponentID != "c1" {
		t.Error("ComponentChange.ComponentID mismatch")
	}

	// Test ConvergenceAnalysis
	convergence := ConvergenceAnalysis{
		Converged:       true,
		ConvergenceRate: 0.1,
		FinalGradient:   0.01,
		StabilityMetric: 0.95,
	}

	if !convergence.Converged {
		t.Error("ConvergenceAnalysis.Converged mismatch")
	}
}
