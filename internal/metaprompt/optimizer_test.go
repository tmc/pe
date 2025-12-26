package metaprompt

import (
	"context"
	"strings"
	"testing"

	"github.com/tmc/pe/internal/llm"
	"github.com/tmc/pe/internal/promptfoo"
)

type mockProvider struct {
	responses map[string]string
	calls     []string
}

func (m *mockProvider) EvaluatePrompt(ctx context.Context, prompt string, vars map[string]interface{}) (*promptfoo.ProviderResponse, error) {
	m.calls = append(m.calls, prompt)
	resp := "Default response"
	if r, ok := m.responses[prompt]; ok {
		resp = r
	} else if strings.Contains(prompt, "improve") || strings.Contains(prompt, "optimize") {
		resp = "Improved prompt: Be more specific and clear."
	} else if strings.Contains(prompt, "evaluate") {
		resp = "Score: 0.85\nReasoning: Good clarity but could be more specific."
	}

	return &promptfoo.ProviderResponse{
		Output: resp,
		TokenUsage: &promptfoo.TokenUsage{
			Total:      int32(len(strings.Fields(resp))),
			Prompt:     0,
			Completion: int32(len(strings.Fields(resp))),
		},
	}, nil
}

func (m *mockProvider) Name() string {
	return "mock"
}

func (m *mockProvider) Model() string {
	return "mock-model"
}

func (m *mockProvider) Generate(ctx context.Context, prompt string, options llm.GenerateOptions) (*llm.GenerateResponse, error) {
	m.calls = append(m.calls, prompt)
	resp := "Default response"
	if r, ok := m.responses[prompt]; ok {
		resp = r
	} else if strings.Contains(prompt, "gradients") && strings.Contains(prompt, "JSON") {
		// Return valid JSON for gradient computation
		resp = `{
			"gradients": [
				{
					"component": "instruction clarity",
					"feedback": "needs more specific detail",
					"suggestions": ["add examples", "clarify scope"],
					"confidence": 0.8,
					"priority": 0.9,
					"gradient": "improve specificity",
					"magnitude": 0.7,
					"direction": "improve"
				}
			]
		}`
	} else if strings.Contains(prompt, "improve") && strings.Contains(prompt, "SUGGESTION") {
		// Return suggestions for standard optimization
		resp = `ANALYSIS: The prompt could be more specific and clear.

SUGGESTION 1: Focus on clarity
Be more specific and clear about the task requirements and expected output format.

SUGGESTION 2: Focus on structure  
Structure the prompt with clear sections for context, task, and output specifications.

SUGGESTION 3: Focus on examples
Add examples to illustrate the expected behavior and output format.`
	} else if strings.Contains(prompt, "IMPROVED PROMPT:") || strings.Contains(prompt, "ENHANCED PROMPT:") ||
		strings.Contains(prompt, "RESTRUCTURED PROMPT:") || strings.Contains(prompt, "CLARIFIED PROMPT:") ||
		strings.Contains(prompt, "FORMATTED PROMPT:") || strings.Contains(prompt, "CONSTRAINED PROMPT:") ||
		strings.Contains(prompt, "SPECIFIC PROMPT:") || strings.Contains(prompt, "OPTIMIZED PROMPT:") {
		// Return improved prompt for APEX mutations
		resp = "IMPROVED PROMPT: Be more specific and clear about the classification task. Provide detailed instructions for edge cases and include examples of expected inputs and outputs."
	} else if strings.Contains(prompt, "improve") || strings.Contains(prompt, "optimize") {
		resp = "Improved prompt: Be more specific and clear."
	} else if strings.Contains(prompt, "SELECTED:") {
		// Return evaluation for standard optimization
		resp = `SELECTED: 1
SCORE: 8.5
REASONING: This prompt provides better clarity and structure.`
	} else if strings.Contains(prompt, "evaluate") || strings.Contains(prompt, "scale from 0.0 to 1.0") {
		resp = "0.85"
	}

	return &llm.GenerateResponse{
		Text:         resp,
		TotalTokens:  len(strings.Fields(resp)),
		Model:        "mock",
		FinishReason: "stop",
	}, nil
}

func (m *mockProvider) SupportsStreaming() bool {
	return false
}

func (m *mockProvider) SupportsBatch() bool {
	return false
}

func TestNewOptimizer(t *testing.T) {
	tests := []struct {
		name     string
		provider llm.Provider
	}{
		{
			name:     "with mock provider",
			provider: &mockProvider{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opt := NewOptimizer(tt.provider)
			if opt == nil {
				t.Errorf("NewOptimizer() returned nil")
			}
			if opt.llm == nil {
				t.Errorf("NewOptimizer() created optimizer with nil llm provider")
			}
		})
	}
}

func TestOptimize(t *testing.T) {
	tests := []struct {
		name          string
		config        Config
		mockResponses map[string]string
		wantImproved  bool
		wantError     bool
	}{
		{
			name: "basic optimization",
			config: Config{
				InitialPrompt: "Summarize this text",
				Objective:     "Make the prompt clearer",
				Iterations:    3,
				Method:        "standard",
			},
			wantImproved: true,
		},
		{
			name: "textgrad optimization",
			config: Config{
				InitialPrompt:        "Extract key points",
				Objective:            "Improve accuracy",
				Iterations:           2,
				Method:               "textgrad",
				UseTextGrad:          true,
				ConvergenceThreshold: 0.01,
			},
			mockResponses: map[string]string{
				"evaluate": "Score: 0.95\nReasoning: Excellent clarity",
			},
			wantImproved: true,
		},
		{
			name: "pe2 optimization",
			config: Config{
				InitialPrompt: "Analyze sentiment",
				Objective:     "Improve speed and accuracy",
				Iterations:    2,
				Method:        "pe2",
			},
			wantImproved: true,
		},
		{
			name: "apex optimization",
			config: Config{
				InitialPrompt: "Classify the input",
				Objective:     "Maximize classification accuracy",
				Iterations:    2,
				Method:        "apex",
			},
			wantImproved: false, // APEX uses random mutation probability, so may not always improve
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := &mockProvider{
				responses: tt.mockResponses,
				calls:     []string{},
			}
			opt := NewOptimizer(provider)

			ctx := context.Background()
			result, err := opt.Optimize(ctx, tt.config)

			if (err != nil) != tt.wantError {
				t.Errorf("Optimize() error = %v, wantError %v", err, tt.wantError)
			}

			if result != nil {
				if tt.wantImproved && result.OptimizedPrompt == tt.config.InitialPrompt {
					t.Errorf("Prompt was not improved: got %v", result.OptimizedPrompt)
				}

				if len(result.Iterations) == 0 {
					t.Errorf("No iterations recorded")
				}
			}

			// APEX may not make LLM calls due to random mutation probability
			if tt.config.Method != "apex" && len(provider.calls) == 0 {
				t.Errorf("No LLM calls were made")
			}
		})
	}
}

func TestOptimizationMethods(t *testing.T) {
	methods := []string{"standard", "textgrad", "pe2", "apex", "hybrid"}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			config := Config{
				InitialPrompt: "Test prompt",
				Objective:     "Improve clarity",
				Iterations:    2,
				Method:        method,
			}
			provider := &mockProvider{
				responses: map[string]string{},
				calls:     []string{},
			}
			opt := NewOptimizer(provider)

			ctx := context.Background()
			result, err := opt.Optimize(ctx, config)

			if err != nil {
				t.Errorf("Optimize() error = %v", err)
			}

			if result == nil {
				t.Errorf("Optimize() returned nil result")
			}
		})
	}
}

func TestIterationResults(t *testing.T) {
	config := Config{
		InitialPrompt: "Explain quantum computing",
		Objective:     "Make it accessible to beginners",
		Iterations:    3,
		Method:        "standard",
	}

	provider := &mockProvider{
		responses: map[string]string{},
		calls:     []string{},
	}
	opt := NewOptimizer(provider)

	ctx := context.Background()
	result, err := opt.Optimize(ctx, config)

	if err != nil {
		t.Errorf("Optimize() error = %v", err)
	}

	if result == nil {
		t.Errorf("Optimize() returned nil result")
		return
	}

	// Check iteration results
	if len(result.Iterations) != config.Iterations {
		t.Errorf("Expected %d iterations, got %d", config.Iterations, len(result.Iterations))
	}

	// Verify each iteration has required fields
	for i, iter := range result.Iterations {
		if iter.Iteration != i+1 {
			t.Errorf("Iteration %d has wrong iteration number: %d", i, iter.Iteration)
		}
		if iter.Prompt == "" {
			t.Errorf("Iteration %d has empty prompt", i)
		}
		if iter.Score == 0 {
			t.Errorf("Iteration %d has zero score", i)
		}
	}
}

func BenchmarkOptimize(b *testing.B) {
	config := Config{
		InitialPrompt: "Summarize this document",
		Objective:     "Improve clarity",
		Iterations:    3,
		Method:        "standard",
	}
	provider := &mockProvider{
		responses: map[string]string{},
		calls:     []string{},
	}
	opt := NewOptimizer(provider)

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = opt.Optimize(ctx, config)
	}
}
