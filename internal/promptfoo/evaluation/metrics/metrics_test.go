package metrics

import (
	"context"
	"math"
	"strings"
	"testing"

	"github.com/tmc/pe/internal/llm"
	"github.com/tmc/pe/internal/promptfoo"
)

// Mock LLM provider for testing
type mockLLMProvider struct {
	responses map[string]string
	calls     []string
}

func (m *mockLLMProvider) EvaluatePrompt(ctx context.Context, prompt string, vars map[string]interface{}) (*promptfoo.ProviderResponse, error) {
	m.calls = append(m.calls, prompt)
	resp := "0.85" // Default similarity score

	// Return mock scores based on prompt content
	if strings.Contains(prompt, "semantic similarity") {
		if strings.Contains(prompt, "exact match") {
			resp = "0.98"
		} else if strings.Contains(prompt, "different meaning") {
			resp = "0.45"
		}
	}

	if r, ok := m.responses[prompt]; ok {
		resp = r
	}

	return &promptfoo.ProviderResponse{
		Output: resp,
		TokenUsage: &promptfoo.TokenUsage{
			Total: int32(len(strings.Fields(resp))),
		},
	}, nil
}

func (m *mockLLMProvider) Name() string {
	return "mock"
}

func (m *mockLLMProvider) Model() string {
	return "mock-model"
}

func (m *mockLLMProvider) Generate(ctx context.Context, prompt string, options llm.GenerateOptions) (*llm.GenerateResponse, error) {
	m.calls = append(m.calls, prompt)
	resp := "0.85" // Default similarity score

	// Return mock scores based on prompt content
	if strings.Contains(prompt, "semantic similarity") {
		// Check for exact match case
		if strings.Contains(prompt, "REFERENCE TEXT:\nthe cat is on the mat") &&
			strings.Contains(prompt, "GENERATED TEXT:\nthe cat is on the mat") {
			// Exact match
			resp = "0.98"
		} else if strings.Contains(prompt, "the feline is resting on the carpet") {
			// Semantic similarity
			resp = "0.85"
		} else if strings.Contains(prompt, "the dog is under the table") {
			// Different meaning
			resp = "0.45"
		}
	} else if strings.Contains(prompt, "You are an expert evaluator") {
		// G-Eval responses
		if strings.Contains(prompt, "The capital of France is Paris") {
			resp = "SCORE: 5\nEXPLANATION: Accurate and relevant"
		} else {
			resp = "SCORE: 3\nEXPLANATION: Partially relevant"
		}
	} else if strings.Contains(prompt, "Evaluate the following generated text on the dimension") {
		// UniEval responses - return higher scores
		resp = "8.5"
	}

	if r, ok := m.responses[prompt]; ok {
		resp = r
	}

	return &llm.GenerateResponse{
		Text:        resp,
		TotalTokens: len(strings.Fields(resp)),
		Model:       "mock",
	}, nil
}

func (m *mockLLMProvider) SupportsStreaming() bool {
	return false
}

func (m *mockLLMProvider) SupportsBatch() bool {
	return false
}

func TestCalculateBLEU(t *testing.T) {
	tests := []struct {
		name      string
		candidate string
		reference string
		maxN      int
		want      float64
		epsilon   float64
	}{
		{
			name:      "exact match",
			candidate: "the cat is on the mat",
			reference: "the cat is on the mat",
			maxN:      4,
			want:      1.0,
			epsilon:   0.01,
		},
		{
			name:      "partial match",
			candidate: "the cat sat on the mat",
			reference: "the cat is on the mat",
			maxN:      4,
			want:      0.45,
			epsilon:   0.1,
		},
		{
			name:      "no match",
			candidate: "hello world",
			reference: "the cat is on the mat",
			maxN:      4,
			want:      0.0,
			epsilon:   0.01,
		},
		{
			name:      "different lengths",
			candidate: "the cat",
			reference: "the cat is on the mat",
			maxN:      4,
			want:      0.014,
			epsilon:   0.01,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateBLEU(tt.candidate, tt.reference, tt.maxN)
			if result == nil {
				t.Fatalf("CalculateBLEU() returned nil")
			}
			if math.Abs(result.Score-tt.want) > tt.epsilon {
				t.Errorf("CalculateBLEU() = %v, want %v (±%v)", result.Score, tt.want, tt.epsilon)
			}
		})
	}
}

func TestCalculateROUGE(t *testing.T) {
	tests := []struct {
		name      string
		candidate string
		reference string
		variant   string
		want      float64
		epsilon   float64
	}{
		{
			name:      "ROUGE-1 exact match",
			candidate: "the cat is on the mat",
			reference: "the cat is on the mat",
			variant:   "rouge-1",
			want:      1.0,
			epsilon:   0.01,
		},
		{
			name:      "ROUGE-1 partial match",
			candidate: "the cat sat on the mat",
			reference: "the dog is on the mat",
			variant:   "rouge-1",
			want:      0.667,
			epsilon:   0.1,
		},
		{
			name:      "ROUGE-2 bigrams",
			candidate: "the cat is on the mat",
			reference: "the cat is on the mat",
			variant:   "rouge-2",
			want:      1.0,
			epsilon:   0.01,
		},
		{
			name:      "ROUGE-L longest common subsequence",
			candidate: "the cat is sleeping on the mat",
			reference: "the cat is on the mat",
			variant:   "rouge-l",
			want:      0.85,
			epsilon:   0.1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateROUGE(tt.candidate, tt.reference, tt.variant)
			if result == nil {
				t.Fatalf("CalculateROUGE() returned nil")
			}
			if math.Abs(result.Score-tt.want) > tt.epsilon {
				t.Errorf("CalculateROUGE() = %v, want %v (±%v)", result.Score, tt.want, tt.epsilon)
			}
		})
	}
}

func TestCalculateMETEOR(t *testing.T) {
	tests := []struct {
		name      string
		candidate string
		reference string
		want      float64
		epsilon   float64
	}{
		{
			name:      "exact match",
			candidate: "the cat is on the mat",
			reference: "the cat is on the mat",
			want:      1.0,
			epsilon:   0.01,
		},
		{
			name:      "synonyms and stemming",
			candidate: "the feline is on the rug",
			reference: "the cat is on the mat",
			want:      0.7,
			epsilon:   0.15,
		},
		{
			name:      "word order matters",
			candidate: "on the mat is the cat",
			reference: "the cat is on the mat",
			want:      0.998,
			epsilon:   0.01,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateMETEOR(tt.candidate, tt.reference)
			if result == nil {
				t.Fatalf("CalculateMETEOR() returned nil")
			}
			if math.Abs(result.Score-tt.want) > tt.epsilon {
				t.Errorf("CalculateMETEOR() = %v, want %v (±%v)", result.Score, tt.want, tt.epsilon)
			}
		})
	}
}

func TestCalculateBERTScore(t *testing.T) {
	tests := []struct {
		name      string
		candidate string
		reference string
		wantMin   float64
		wantMax   float64
	}{
		{
			name:      "exact match",
			candidate: "the cat is on the mat",
			reference: "the cat is on the mat",
			wantMin:   0.85,
			wantMax:   1.0,
		},
		{
			name:      "semantic similarity",
			candidate: "the feline is resting on the carpet",
			reference: "the cat is on the mat",
			wantMin:   0.8,
			wantMax:   0.9,
		},
		{
			name:      "different meaning",
			candidate: "the dog is under the table",
			reference: "the cat is on the mat",
			wantMin:   0.3,
			wantMax:   0.6,
		},
	}

	// Mock provider for BERTScore
	provider := &mockLLMProvider{
		responses: map[string]string{},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateBERTScoreDetailed(tt.candidate, tt.reference, provider)
			if result == nil {
				t.Fatalf("CalculateBERTScoreDetailed() returned nil")
			}
			if result.Score < tt.wantMin || result.Score > tt.wantMax {
				t.Errorf("CalculateBERTScoreDetailed() = %v, want between %v and %v", result.Score, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestGEval(t *testing.T) {
	tests := []struct {
		name      string
		generated string
		reference string
		criteria  string
		wantMin   float64
		wantMax   float64
	}{
		{
			name:      "high quality match",
			generated: "The capital of France is Paris",
			reference: "Paris is the capital city of France",
			criteria:  "accuracy",
			wantMin:   0.8,
			wantMax:   1.0,
		},
		{
			name:      "poor quality match",
			generated: "France is a country",
			reference: "Paris is the capital city of France",
			criteria:  "relevance",
			wantMin:   0.2,
			wantMax:   0.5,
		},
	}

	provider := &mockLLMProvider{
		responses: map[string]string{},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateGEval(tt.generated, tt.reference, tt.criteria, provider)
			if result == nil {
				t.Fatalf("CalculateGEval() returned nil")
			}
			if result.Score < tt.wantMin || result.Score > tt.wantMax {
				t.Errorf("CalculateGEval() = %v, want between %v and %v", result.Score, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestUniEval(t *testing.T) {
	tests := []struct {
		name      string
		generated string
		reference string
		taskType  string
		wantMin   float64
		wantMax   float64
	}{
		{
			name:      "summarization task",
			generated: "This article discusses climate change impacts",
			reference: "The article examines the effects of climate change on ecosystems",
			taskType:  "summarization",
			wantMin:   0.6,
			wantMax:   0.9,
		},
		{
			name:      "dialogue task",
			generated: "Hello, how can I help you today?",
			reference: "Hi there! What can I assist you with?",
			taskType:  "dialogue",
			wantMin:   0.7,
			wantMax:   1.0,
		},
	}

	provider := &mockLLMProvider{
		responses: map[string]string{},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateUniEval(tt.generated, tt.reference, tt.taskType, provider)
			if result == nil {
				t.Fatalf("CalculateUniEval() returned nil")
			}
			if result.Score < tt.wantMin || result.Score > tt.wantMax {
				t.Errorf("CalculateUniEval() = %v, want between %v and %v", result.Score, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func BenchmarkBLEU(b *testing.B) {
	candidate := "The quick brown fox jumps over the lazy dog"
	reference := "A fast brown fox leaps over a lazy dog"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = CalculateBLEU(candidate, reference, 4)
	}
}

func BenchmarkROUGE(b *testing.B) {
	candidate := "The quick brown fox jumps over the lazy dog"
	reference := "A fast brown fox leaps over a lazy dog"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = CalculateROUGE(candidate, reference, "rouge-1")
	}
}
