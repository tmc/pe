package metaprompt

import (
	"testing"
)

func TestNewTextGradOptimizer(t *testing.T) {
	provider := &mockProvider{}
	optimizer := NewTextGradOptimizer(provider)

	if optimizer == nil {
		t.Fatal("NewTextGradOptimizer returned nil")
	}
	if optimizer.llm == nil {
		t.Error("TextGradOptimizer has nil llm provider")
	}
	if optimizer.attentionFlowMapper == nil {
		t.Error("TextGradOptimizer has nil attentionFlowMapper")
	}
	if !optimizer.enhancedMode {
		t.Error("TextGradOptimizer should have enhancedMode enabled by default")
	}
}

func TestTextGradOptimizerExtractChanges(t *testing.T) {
	tests := []struct {
		name     string
		original string
		improved string
		wantLen  int
	}{
		{
			name:     "length change",
			original: "short prompt",
			improved: "a much longer and detailed prompt with more content",
			wantLen:  1,
		},
		{
			name:     "step-by-step added",
			original: "analyze this",
			improved: "analyze this step-by-step",
			wantLen:  2, // length change + step-by-step
		},
		{
			name:     "more line breaks",
			original: "single line",
			improved: "line 1\nline 2\nline 3",
			wantLen:  2, // length change + formatting
		},
		{
			name:     "no changes",
			original: "same",
			improved: "same",
			wantLen:  0,
		},
	}

	provider := &mockProvider{}
	optimizer := NewTextGradOptimizer(provider)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			changes := optimizer.extractChanges(tt.original, tt.improved)
			if len(changes) != tt.wantLen {
				t.Errorf("extractChanges() returned %d changes, want %d", len(changes), tt.wantLen)
			}
		})
	}
}

func TestTextualGradientStruct(t *testing.T) {
	gradient := TextualGradient{
		Component:   "instruction",
		Feedback:    "needs more detail",
		Suggestions: []string{"add examples", "clarify scope"},
		Confidence:  0.85,
		Priority:    0.9,
		Gradient:    "improve specificity",
		Magnitude:   0.7,
		Direction:   "improve",
	}

	if gradient.Confidence < 0 || gradient.Confidence > 1 {
		t.Error("Confidence should be between 0 and 1")
	}
	if len(gradient.Suggestions) != 2 {
		t.Error("Suggestions length mismatch")
	}
}

func TestComputationGraphStruct(t *testing.T) {
	graph := ComputationGraph{
		Nodes: []GraphNode{
			{ID: "node1", Type: "prompt", Content: "test"},
		},
		Edges: []GraphEdge{
			{From: "node1", To: "node2", Weight: 0.8, Type: "flow"},
		},
	}

	if len(graph.Nodes) != 1 {
		t.Error("Nodes length mismatch")
	}
	if len(graph.Edges) != 1 {
		t.Error("Edges length mismatch")
	}
}

func TestGraphNodeStruct(t *testing.T) {
	node := GraphNode{
		ID:        "test_node",
		Type:      "response",
		Content:   "test content",
		Gradients: []TextualGradient{{Component: "test"}},
		Metadata:  map[string]interface{}{"key": "value"},
	}

	if node.Type != "response" {
		t.Error("GraphNode.Type mismatch")
	}
}

func TestGraphEdgeStruct(t *testing.T) {
	edge := GraphEdge{
		From:   "source",
		To:     "target",
		Weight: 0.75,
		Type:   "connection",
	}

	if edge.Weight < 0 || edge.Weight > 1 {
		t.Error("Edge.Weight should be between 0 and 1")
	}
}

func TestAttentionFlowMapperStruct(t *testing.T) {
	mapper := &AttentionFlowMapper{
		flows: make(map[string][]float64),
	}

	mapper.flows["test"] = []float64{0.1, 0.2, 0.3}

	if len(mapper.flows["test"]) != 3 {
		t.Error("AttentionFlowMapper flows mismatch")
	}
}
