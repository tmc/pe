package metaprompt

import (
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
