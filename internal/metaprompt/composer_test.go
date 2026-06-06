package metaprompt

import (
	"context"
	"strings"
	"testing"
)

func TestNewPromptComposer(t *testing.T) {
	composer := NewPromptComposer()
	if composer == nil {
		t.Fatal("NewPromptComposer returned nil")
	}
	if composer.signatureRegistry == nil {
		t.Error("PromptComposer has nil signatureRegistry")
	}
	if composer.programSynthesizer == nil {
		t.Error("PromptComposer has nil programSynthesizer")
	}
	if composer.qualityGates == nil {
		t.Error("PromptComposer has nil qualityGates")
	}
}

func TestNewPromptComposerWithLLM(t *testing.T) {
	provider := &mockProvider{}
	composer := NewPromptComposerWithLLM(provider)
	if composer == nil {
		t.Fatal("NewPromptComposerWithLLM returned nil")
	}
	if composer.llmProvider == nil {
		t.Error("composer.llmProvider is nil")
	}
}

func TestNewSignatureRegistry(t *testing.T) {
	registry := NewSignatureRegistry()
	if registry == nil {
		t.Fatal("NewSignatureRegistry returned nil")
	}
	if registry.signatures == nil {
		t.Error("SignatureRegistry has nil signatures")
	}
}

func TestSignatureRegistryRegister(t *testing.T) {
	registry := NewSignatureRegistry()

	signature := &ComponentSignature{
		Name: "test_component",
	}

	err := registry.RegisterSignature(signature)
	if err != nil {
		t.Errorf("RegisterSignature() error = %v", err)
	}

	// Test empty name
	err = registry.RegisterSignature(&ComponentSignature{})
	if err == nil {
		t.Error("RegisterSignature() should error on empty name")
	}
}

func TestSignatureRegistryValidateComponent(t *testing.T) {
	registry := NewSignatureRegistry()

	// Component without signature should pass
	component := PromptComponent{
		Content: "test content",
	}

	err := registry.ValidateComponent(component)
	if err != nil {
		t.Errorf("ValidateComponent() error = %v", err)
	}
}

func TestNewProgramSynthesizer(t *testing.T) {
	synthesizer := NewProgramSynthesizer()
	if synthesizer == nil {
		t.Fatal("NewProgramSynthesizer returned nil")
	}
	if len(synthesizer.synthesisStrategies) == 0 {
		t.Error("ProgramSynthesizer has no strategies")
	}
}

func TestNewQualityGateManager(t *testing.T) {
	manager := NewQualityGateManager()
	if manager == nil {
		t.Fatal("NewQualityGateManager returned nil")
	}
	if len(manager.gates) == 0 {
		t.Error("QualityGateManager has no gates")
	}
}

func TestCalculateCoherence(t *testing.T) {
	tests := []struct {
		name    string
		result  string
		wantMin float64
		wantMax float64
	}{
		{
			name: "coherent text",
			result: `First, we need to understand the problem.
Then, we analyze the data.
Finally, we make a conclusion.
Therefore, the result is clear.`,
			wantMin: 0.3,
			wantMax: 1.0,
		},
		{
			name:    "empty text",
			result:  "",
			wantMin: 0.0,
			wantMax: 1.0,
		},
	}

	manager := NewQualityGateManager()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := manager.calculateCoherence(tt.result)
			if score < tt.wantMin || score > tt.wantMax {
				t.Errorf("calculateCoherence() = %v, want between %v and %v", score, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestCalculateClarity(t *testing.T) {
	tests := []struct {
		name    string
		result  string
		wantMin float64
		wantMax float64
	}{
		{
			name:    "clear text",
			result:  "This is a clear and simple sentence. It has good structure.",
			wantMin: 0.3,
			wantMax: 1.0,
		},
		{
			name:    "complex text",
			result:  "Notwithstanding the aforementioned deliberations, the quintessential manifestation of the aforementioned phenomena necessitates a comprehensive reevaluation.",
			wantMin: 0.0,
			wantMax: 0.7,
		},
	}

	manager := NewQualityGateManager()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := manager.calculateClarity(tt.result)
			if score < tt.wantMin || score > tt.wantMax {
				t.Errorf("calculateClarity() = %v, want between %v and %v", score, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestCalculateCompleteness(t *testing.T) {
	manager := NewQualityGateManager()

	result := &ComposeResult{
		ComposedPrompt: "This is a complete result with all required information.",
	}

	score := manager.calculateCompleteness(result)
	if score < 0 || score > 1 {
		t.Errorf("calculateCompleteness() = %v, should be between 0 and 1", score)
	}
}

func TestNewParameterOptimizerEngine(t *testing.T) {
	engine := NewParameterOptimizerEngine()
	if engine == nil {
		t.Fatal("NewParameterOptimizerEngine returned nil")
	}
}

func TestPromptComponentStruct(t *testing.T) {
	component := PromptComponent{
		Content:  "content",
		Metadata: map[string]interface{}{"meta": "data"},
	}

	if component.Content != "content" {
		t.Error("PromptComponent.Content mismatch")
	}
}

func TestComposeResultStruct(t *testing.T) {
	result := ComposeResult{
		ComposedPrompt: "test prompt",
		Components:     []interface{}{},
		Style:          "default",
		CoherenceScore: 0.85,
		ValidationPass: true,
		Metadata:       map[string]interface{}{},
	}

	if result.CoherenceScore != 0.85 {
		t.Error("ComposeResult.CoherenceScore mismatch")
	}
}

func TestProgramSpecStruct(t *testing.T) {
	spec := ProgramSpec{
		Task:        "test task",
		Constraints: []string{"constraint1"},
		Style:       "default",
	}

	if spec.Task != "test task" {
		t.Error("ProgramSpec.Task mismatch")
	}
}

func TestSynthesisResultStruct(t *testing.T) {
	result := SynthesisResult{
		Program:      "test program",
		Components:   []PromptComponent{},
		Strategy:     "template",
		Confidence:   0.9,
		QualityScore: 0.85,
		Iterations:   5,
	}

	if result.Confidence != 0.9 {
		t.Error("SynthesisResult.Confidence mismatch")
	}
}

func TestComposerHelperFunctions(t *testing.T) {
	intVal := composerIntPtr(42)
	if intVal == nil || *intVal != 42 {
		t.Error("composerIntPtr failed")
	}

	floatVal := composerFloatPtr(0.5)
	if floatVal == nil || *floatVal != 0.5 {
		t.Error("composerFloatPtr failed")
	}
}

func TestComposerCompose(t *testing.T) {
	provider := &mockProvider{}
	composer := NewPromptComposerWithLLM(provider)

	tests := []struct {
		name       string
		components []interface{}
		style      string
		wantErr    bool
	}{
		{
			name: "basic composition",
			components: []interface{}{
				PromptComponent{Content: "First component"},
				PromptComponent{Content: "Second component"},
			},
			style:   "default",
			wantErr: false,
		},
		{
			name:       "empty components",
			components: []interface{}{},
			style:      "default",
			wantErr:    false,
		},
		{
			name: "single component",
			components: []interface{}{
				PromptComponent{Content: "Only component"},
			},
			style:   "minimal",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			result, err := composer.Compose(ctx, tt.components, tt.style)

			if (err != nil) != tt.wantErr {
				t.Errorf("Compose() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && result == nil {
				t.Error("Compose() returned nil result")
			}
		})
	}
}

func TestDefaultStyleHandler(t *testing.T) {
	handler := &DefaultStyleHandler{}

	components := []interface{}{
		PromptComponent{Content: "First component"},
		PromptComponent{Content: "Second component"},
	}

	ctx := context.Background()
	result, err := handler.Compose(ctx, components, nil)

	if err != nil {
		t.Fatalf("DefaultStyleHandler.Compose() error = %v", err)
	}

	if result == nil {
		t.Error("DefaultStyleHandler.Compose() returned nil result")
	}
}

func TestChainOfThoughtHandler(t *testing.T) {
	handler := &ChainOfThoughtHandler{}

	components := []interface{}{
		PromptComponent{Content: "Step 1"},
		PromptComponent{Content: "Step 2"},
	}

	ctx := context.Background()
	result, err := handler.Compose(ctx, components, nil)

	if err != nil {
		t.Fatalf("ChainOfThoughtHandler.Compose() error = %v", err)
	}

	if result == nil {
		t.Error("ChainOfThoughtHandler.Compose() returned nil result")
	}
}

func TestFewShotHandler(t *testing.T) {
	handler := &FewShotHandler{}

	components := []interface{}{
		PromptComponent{Content: "Example 1"},
		PromptComponent{Content: "Example 2"},
	}

	ctx := context.Background()
	result, err := handler.Compose(ctx, components, nil)

	if err != nil {
		t.Fatalf("FewShotHandler.Compose() error = %v", err)
	}

	if result == nil {
		t.Error("FewShotHandler.Compose() returned nil result")
	}
}

func TestStructuredHandler(t *testing.T) {
	handler := &StructuredHandler{}

	components := []interface{}{
		PromptComponent{Content: "Section 1"},
		PromptComponent{Content: "Section 2"},
	}

	ctx := context.Background()
	result, err := handler.Compose(ctx, components, nil)

	if err != nil {
		t.Fatalf("StructuredHandler.Compose() error = %v", err)
	}

	if result == nil {
		t.Error("StructuredHandler.Compose() returned nil result")
	}
}

func TestConversationalHandler(t *testing.T) {
	handler := &ConversationalHandler{}

	components := []interface{}{
		PromptComponent{Content: "Message 1"},
		PromptComponent{Content: "Message 2"},
	}

	ctx := context.Background()
	result, err := handler.Compose(ctx, components, nil)

	if err != nil {
		t.Fatalf("ConversationalHandler.Compose() error = %v", err)
	}

	if result == nil {
		t.Error("ConversationalHandler.Compose() returned nil result")
	}
}

func TestPromptComposerEnhancedPaths(t *testing.T) {
	composer := NewPromptComposerWithLLM(&mockProvider{})
	components := []interface{}{
		PromptComponent{Type: "context", Content: "Use repo context."},
		PromptComponent{Type: "instruction", Content: "Explain the change.", Dependencies: []string{"context"}},
		PromptComponent{Type: "example", Content: "Input: x Output: y"},
		PromptComponent{Type: "constraint", Content: "Be concise."},
	}

	result, err := composer.Compose(context.Background(), components, map[string]interface{}{
		"Style":                 "cot",
		"QualityGates":          true,
		"ParameterOptimization": true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Style != "cot" || !strings.Contains(result.ComposedPrompt, "Let's think step by step") {
		t.Fatalf("compose result = %#v", result)
	}
	if _, ok := result.Metadata["quality_evaluation"]; !ok {
		t.Fatalf("metadata missing quality evaluation: %#v", result.Metadata)
	}
	if _, ok := result.Metadata["optimized_parameters"]; !ok {
		t.Fatalf("metadata missing optimized parameters: %#v", result.Metadata)
	}

	synthesized, err := composer.Compose(context.Background(), components, &EnhancedComposeConfig{
		ComposeConfig:    ComposeConfig{Style: "structured"},
		ProgramSynthesis: true,
	})
	if err == nil || synthesized != nil || !strings.Contains(err.Error(), "not yet implemented") {
		t.Fatalf("synthesized = %#v err=%v", synthesized, err)
	}
}

func TestPromptComposerValidationAndConfigEdges(t *testing.T) {
	composer := NewPromptComposer()
	if _, err := composer.Compose(context.Background(), nil, map[string]interface{}{"Style": "missing"}); err == nil {
		t.Fatal("unsupported style succeeded")
	}
	components := []interface{}{PromptComponent{
		Type:      "instruction",
		Content:   "",
		Signature: &ComponentSignature{Name: "instruction", InputSchema: Schema{Type: "string"}},
	}}
	if _, err := composer.Compose(context.Background(), components, nil); err == nil {
		t.Fatal("invalid signed component succeeded")
	}
	cfg := composer.extractConfig(&ComposeConfig{Style: "dspy", ValidationGate: true})
	if cfg.Style != "dspy" || !cfg.QualityGates {
		t.Fatalf("config = %#v", cfg)
	}
	spec := composer.createProgramSpec([]interface{}{
		PromptComponent{Type: "instruction", Content: "Do the task", Dependencies: []string{"a", "b"}},
	}, &EnhancedComposeConfig{ComposeConfig: ComposeConfig{Style: "cot"}, ParameterOptimization: true})
	if spec.Task != "Do the task" || spec.Style != "cot" || !spec.Quality.RequireOptimization || len(spec.Constraints) != 1 {
		t.Fatalf("spec = %#v", spec)
	}
}

func TestComposerStyleHandlersUseTypedComponents(t *testing.T) {
	components := []interface{}{
		PromptComponent{Type: "context", Content: "Context text."},
		PromptComponent{Type: "instruction", Content: "Analyze data"},
		PromptComponent{Type: "example", Content: "Example text."},
		PromptComponent{Type: "constraint", Content: "No fluff."},
		"ignored",
	}
	tests := []struct {
		name    string
		handler StyleHandler
		want    string
	}{
		{"default", &DefaultStyleHandler{}, "Instructions:"},
		{"cot", &ChainOfThoughtHandler{}, "Constraints:"},
		{"few-shot", &FewShotHandler{}, "Here are some examples:"},
		{"structured", &StructuredHandler{}, "## Output Format"},
		{"conversational", &ConversationalHandler{}, "Thanks for your help!"},
		{"dspy", &DSPyHandler{}, "## Reasoning"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.handler.Compose(context.Background(), components, nil)
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(result.ComposedPrompt, tt.want) {
				t.Fatalf("prompt = %q", result.ComposedPrompt)
			}
		})
	}
}

func TestComposerSynthesisQualityAndOptimization(t *testing.T) {
	ps := NewProgramSynthesizer()
	spec := ProgramSpec{Task: "task"}
	if got := ps.selectStrategy(spec).GetName(); got != "template" {
		t.Fatalf("template strategy = %q", got)
	}
	spec.Quality.RequireOptimization = true
	if got := ps.selectStrategy(spec).GetName(); got != "evolutionary" {
		t.Fatalf("evolutionary strategy = %q", got)
	}
	spec.Examples = make([]SignatureExample, 6)
	if got := ps.selectStrategy(spec).GetName(); got != "neural" {
		t.Fatalf("neural strategy = %q", got)
	}
	result, err := ps.SynthesizePrompt(context.Background(), spec)
	if err == nil || result != nil || !strings.Contains(err.Error(), "not yet implemented") {
		t.Fatalf("synthesis = %#v err=%v", result, err)
	}

	qgm := NewQualityGateManager()
	for _, gate := range []QualityGate{
		{Metric: "coherence", Threshold: 0.1},
		{Metric: "clarity", Threshold: 0.1},
		{Metric: "completeness", Threshold: 0.5},
		{Metric: "unknown", Threshold: 2},
	} {
		passed, score := qgm.evaluateGate(gate, &ComposeResult{
			ComposedPrompt: "Please analyze this. Therefore explain it.",
			Components: []interface{}{
				PromptComponent{Type: "context"},
				PromptComponent{Type: "instruction"},
				PromptComponent{Type: "example"},
			},
		})
		if !passed || score <= 0 {
			t.Fatalf("gate %#v passed=%v score=%v", gate, passed, score)
		}
	}
	if got := calculateVariance(nil); got != 0 {
		t.Fatalf("empty variance = %v", got)
	}
}
