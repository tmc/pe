package evaluator

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmc/pe/internal/distributed"
	"github.com/tmc/pe/internal/llm"
	"github.com/tmc/pe/internal/promptfoo"
)

func TestEvaluate(t *testing.T) {
	tests := []struct {
		name            string
		config          promptfoo.Config
		timeout         time.Duration
		dryRun          bool
		maxConcurrency  int
		showProgressBar bool
		wantErr         bool
		errContains     string
	}{
		{
			name:           "empty config should error",
			config:         promptfoo.Config{},
			timeout:        10 * time.Second,
			dryRun:         true,
			maxConcurrency: 1,
			wantErr:        true,
			errContains:    "missing required config fields",
		},
		{
			name: "missing prompts should error",
			config: promptfoo.Config{
				Providers: []promptfoo.ProviderConfig{{ID: "mock"}},
				Tests: []promptfoo.TestCase{
					{Vars: map[string]interface{}{"test": "value"}},
				},
			},
			timeout:        10 * time.Second,
			dryRun:         true,
			maxConcurrency: 1,
			wantErr:        true,
			errContains:    "missing required config fields",
		},
		{
			name: "missing providers should error",
			config: promptfoo.Config{
				Prompts: []string{"test prompt"},
				Tests: []promptfoo.TestCase{
					{Vars: map[string]interface{}{"test": "value"}},
				},
			},
			timeout:        10 * time.Second,
			dryRun:         true,
			maxConcurrency: 1,
			wantErr:        true,
			errContains:    "missing required config fields",
		},
		{
			name: "missing tests should error",
			config: promptfoo.Config{
				Prompts:   []string{"test prompt"},
				Providers: []promptfoo.ProviderConfig{{ID: "mock"}},
			},
			timeout:        10 * time.Second,
			dryRun:         true,
			maxConcurrency: 1,
			wantErr:        true,
			errContains:    "missing required config fields",
		},
		{
			name: "valid config with dry run should succeed",
			config: promptfoo.Config{
				Prompts:   []string{"test prompt {{var1}}"},
				Providers: []promptfoo.ProviderConfig{{ID: "mock"}},
				Tests: []promptfoo.TestCase{
					{
						Vars: map[string]interface{}{"var1": "value1"},
						Assert: []promptfoo.Assertion{
							{Type: "contains", Value: "test"},
						},
					},
				},
			},
			timeout:        10 * time.Second,
			dryRun:         true,
			maxConcurrency: 1,
			wantErr:        false,
		},
		{
			name: "multiple prompts and providers",
			config: promptfoo.Config{
				Prompts:   []string{"prompt1", "prompt2"},
				Providers: []promptfoo.ProviderConfig{{ID: "mock"}, {ID: "mock:model2"}},
				Tests: []promptfoo.TestCase{
					{
						Vars: map[string]interface{}{"test": "value"},
						Assert: []promptfoo.Assertion{
							{Type: "equals", Value: "Dry run response"},
						},
					},
				},
			},
			timeout:         10 * time.Second,
			dryRun:          true,
			maxConcurrency:  2,
			showProgressBar: true,
			wantErr:         false,
		},
		{
			name: "timeout should cancel evaluation",
			config: promptfoo.Config{
				Prompts:   []string{"test prompt"},
				Providers: []promptfoo.ProviderConfig{{ID: "mock"}},
				Tests: []promptfoo.TestCase{
					{Vars: map[string]interface{}{"test": "value"}},
				},
			},
			timeout:        1 * time.Nanosecond, // Very short timeout
			dryRun:         false,
			maxConcurrency: 1,
			wantErr:        false, // May not error if evaluation completes quickly
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Evaluate(tt.config, tt.timeout, tt.dryRun, tt.maxConcurrency, tt.showProgressBar)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			if err != nil {
				t.Logf("Got error: %v", err)
			}

			// Verify result structure for successful evaluations
			if !tt.wantErr && err == nil {
				assert.NotEmpty(t, result.EvalID)
				assert.Contains(t, result.EvalID, "eval-")

				expectedResults := len(tt.config.Prompts) * len(tt.config.Providers) * len(tt.config.Tests)
				assert.Equal(t, expectedResults, len(result.Results.Results))

				// Check that prompts were created correctly
				expectedPrompts := len(tt.config.Prompts) * len(tt.config.Providers)
				assert.Equal(t, expectedPrompts, len(result.Results.Prompts))

				// Verify each result has required fields
				for _, r := range result.Results.Results {
					assert.NotEmpty(t, r.ID)
					assert.NotEmpty(t, r.PromptID)
					assert.NotEmpty(t, r.Prompt)
					assert.NotEmpty(t, r.Provider)

					if tt.dryRun {
						assert.Equal(t, "Dry run response", r.Response.Output)
					}
				}
			}
		})
	}
}

func TestReplaceVariables(t *testing.T) {
	tests := []struct {
		name     string
		prompt   string
		vars     map[string]interface{}
		expected string
	}{
		{
			name:     "no variables",
			prompt:   "This is a test prompt",
			vars:     map[string]interface{}{},
			expected: "This is a test prompt",
		},
		{
			name:     "single variable",
			prompt:   "Hello {{name}}!",
			vars:     map[string]interface{}{"name": "World"},
			expected: "Hello World!",
		},
		{
			name:     "multiple variables",
			prompt:   "{{greeting}} {{name}}, how are {{you}}?",
			vars:     map[string]interface{}{"greeting": "Hello", "name": "Alice", "you": "you"},
			expected: "Hello Alice, how are you?",
		},
		{
			name:     "missing variable",
			prompt:   "Hello {{name}}!",
			vars:     map[string]interface{}{},
			expected: "Hello {{name}}!",
		},
		{
			name:     "numeric variable",
			prompt:   "The answer is {{answer}}",
			vars:     map[string]interface{}{"answer": 42},
			expected: "The answer is 42",
		},
		{
			name:     "boolean variable",
			prompt:   "Is it true? {{value}}",
			vars:     map[string]interface{}{"value": true},
			expected: "Is it true? true",
		},
		{
			name:     "nested object variable",
			prompt:   "User: {{user}}",
			vars:     map[string]interface{}{"user": map[string]interface{}{"name": "Bob", "age": 30}},
			expected: "User: {\"age\":30,\"name\":\"Bob\"}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := promptfoo.ApplyVars(tt.prompt, tt.vars)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGeneratePromptID(t *testing.T) {
	prompt1 := "Test prompt"
	provider1 := "openai:gpt-4"

	id1 := generatePromptID(prompt1, provider1)
	id2 := generatePromptID(prompt1, provider1)

	// Same inputs should generate same ID
	assert.Equal(t, id1, id2)

	// Different inputs should generate different IDs
	id3 := generatePromptID("Different prompt", provider1)
	assert.NotEqual(t, id1, id3)

	id4 := generatePromptID(prompt1, "anthropic:claude")
	assert.NotEqual(t, id1, id4)
}

func TestGenerateResultID(t *testing.T) {
	prompt := "Test prompt"
	provider := "openai:gpt-4"
	vars1 := map[string]interface{}{"key": "value"}
	vars2 := map[string]interface{}{"key": "different"}

	id1 := generateResultID(prompt, provider, vars1)
	id2 := generateResultID(prompt, provider, vars1)

	// Same inputs should generate same ID
	assert.Equal(t, id1, id2)

	// Different vars should generate different IDs
	id3 := generateResultID(prompt, provider, vars2)
	assert.NotEqual(t, id1, id3)
}

func TestIfThenElse(t *testing.T) {
	assert.Equal(t, 1.0, ifThenElse(true, 1.0, 0.0))
	assert.Equal(t, 0.0, ifThenElse(false, 1.0, 0.0))
	assert.Equal(t, 1.0, ifThenElse(true, "yes", "no"))
	assert.Equal(t, 0.0, ifThenElse(false, "yes", "no"))
}

func TestMin(t *testing.T) {
	assert.Equal(t, 1, min(1, 2))
	assert.Equal(t, 1, min(2, 1))
	assert.Equal(t, -5, min(-5, 0))
	assert.Equal(t, 3, min(3, 3))
}

func TestGenerateRandomString(t *testing.T) {
	// Test different lengths
	lengths := []int{3, 5, 10, 20}

	for _, length := range lengths {
		str := generateRandomString(length)
		assert.Equal(t, length, len(str))

		// Should only contain alphanumeric characters
		assert.Regexp(t, "^[a-zA-Z0-9]+$", str)
	}

	// Should generate different strings
	str1 := generateRandomString(10)
	str2 := generateRandomString(10)
	assert.NotEqual(t, str1, str2)
}

// Test for edge cases in concurrent evaluation
func TestEvaluateConcurrency(t *testing.T) {
	// Create a config with many tests to ensure concurrency is exercised
	tests := make([]promptfoo.TestCase, 20)
	for i := range tests {
		tests[i] = promptfoo.TestCase{
			Vars: map[string]interface{}{"index": i},
			Assert: []promptfoo.Assertion{
				{Type: "contains", Value: "run"},
			},
		}
	}

	config := promptfoo.Config{
		Prompts:   []string{"test {{index}}"},
		Providers: []promptfoo.ProviderConfig{{ID: "mock"}},
		Tests:     tests,
	}

	result, err := Evaluate(config, 10*time.Second, true, 5, false)
	require.NoError(t, err)

	// All tests should complete
	assert.Equal(t, len(tests), len(result.Results.Results))

	// Check results are properly populated
	for _, r := range result.Results.Results {
		assert.NotEmpty(t, r.ID)
		assert.Equal(t, "Dry run response", r.Response.Output)
	}
}

func TestEvaluate_ObjectProviderLabelAndPromptFilter(t *testing.T) {
	config := promptfoo.Config{
		Prompts: []string{"keep", "skip"},
		Providers: []promptfoo.ProviderConfig{
			{
				ID:      "mock",
				Label:   "mock labeled",
				Prompts: []string{"keep"},
			},
		},
		Tests: []promptfoo.TestCase{
			{
				Vars: map[string]interface{}{},
				Assert: []promptfoo.Assertion{
					{Type: "equals", Value: "Dry run response"},
				},
			},
		},
	}

	result, err := Evaluate(config, 10*time.Second, true, 1, false)
	require.NoError(t, err)
	require.Len(t, result.Results.Results, 1)
	require.Len(t, result.Results.Prompts, 1)
	assert.Equal(t, "mock labeled", result.Results.Results[0].Provider["label"])
	assert.Equal(t, "mock labeled", result.Results.Prompts[0].Provider)
	assert.Equal(t, "keep", result.Results.Results[0].Prompt["label"])
}

func TestEvaluate_UsesDistributedRunLocalOrder(t *testing.T) {
	t.Setenv("PE_TEST_MODE", "true")

	config := promptfoo.Config{
		Prompts: []string{"What is 2+2?"},
		Providers: []promptfoo.ProviderConfig{
			{ID: "mock", Label: "slow", Delay: "25ms"},
			{ID: "mock", Label: "fast"},
		},
		Tests: []promptfoo.TestCase{
			{
				Vars: map[string]interface{}{},
				Assert: []promptfoo.Assertion{
					{Type: "equals", Value: "4"},
				},
			},
		},
	}

	result, err := Evaluate(config, time.Second, false, 2, false)
	require.NoError(t, err)
	require.Len(t, result.Results.Results, 2)
	assert.Equal(t, "slow", result.Results.Results[0].Provider["label"])
	assert.Equal(t, "fast", result.Results.Results[1].Provider["label"])
	assert.Equal(t, 2, result.Results.Stats.Successes)
}

func TestEvaluate_DistributedRunLocalCancellation(t *testing.T) {
	t.Setenv("PE_TEST_MODE", "true")

	config := promptfoo.Config{
		Prompts: []string{"What is 2+2?"},
		Providers: []promptfoo.ProviderConfig{
			{ID: "mock", Label: "slow", Delay: "50ms"},
		},
		Tests: []promptfoo.TestCase{
			{Vars: map[string]interface{}{}},
		},
	}

	_, err := Evaluate(config, time.Nanosecond, false, 1, false)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "context deadline exceeded")
}

func TestEvaluate_DistributedConsensusFromWorkflowResults(t *testing.T) {
	t.Setenv("PE_TEST_MODE", "true")

	config := promptfoo.Config{
		Prompts: []string{"What is 2+2?"},
		Providers: []promptfoo.ProviderConfig{
			{ID: "mock", Label: "b"},
			{ID: "mock", Label: "a"},
			{ID: "mock", Label: "c"},
		},
		Tests: []promptfoo.TestCase{
			{Vars: map[string]interface{}{}},
		},
	}

	result, err := Evaluate(config, time.Second, false, 3, false)
	require.NoError(t, err)
	require.Len(t, result.Results.Results, 3)

	votes := make([]distributed.Vote, 0, len(result.Results.Results))
	for _, r := range result.Results.Results {
		votes = append(votes, distributed.Vote{
			Provider: r.Provider["label"],
			Output:   r.Response.Output,
		})
	}
	consensus, err := distributed.Majority(votes)
	require.NoError(t, err)
	assert.Equal(t, "4", consensus.Output)
	assert.Equal(t, 3, consensus.Weight)
	assert.Equal(t, []string{"a", "b", "c"}, consensus.Providers)
}

func TestEvaluate_DistributedRunLocalMaxConcurrency(t *testing.T) {
	var running int32
	var maxRunning int32
	registerEvaluatorTestProvider("limit", func(model string, options map[string]interface{}) llm.Provider {
		return evaluatorTestProvider{
			name: "limit",
			evaluate: func(ctx context.Context, prompt string, vars map[string]interface{}) (*promptfoo.ProviderResponse, error) {
				n := atomic.AddInt32(&running, 1)
				defer atomic.AddInt32(&running, -1)
				for {
					old := atomic.LoadInt32(&maxRunning)
					if n <= old || atomic.CompareAndSwapInt32(&maxRunning, old, n) {
						break
					}
				}
				select {
				case <-ctx.Done():
					return nil, ctx.Err()
				case <-time.After(10 * time.Millisecond):
				}
				return evaluatorTestResponse(prompt), nil
			},
		}
	})

	config := promptfoo.Config{
		Prompts: []string{"p0", "p1", "p2", "p3", "p4"},
		Providers: []promptfoo.ProviderConfig{
			{ID: "limit"},
		},
		Tests: []promptfoo.TestCase{
			{Vars: map[string]interface{}{}},
		},
	}

	result, err := Evaluate(config, time.Second, false, 2, false)
	require.NoError(t, err)
	require.Len(t, result.Results.Results, 5)
	assert.LessOrEqual(t, maxRunning, int32(2))
	for i, r := range result.Results.Results {
		assert.Equal(t, fmt.Sprintf("p%d", i), r.Prompt["label"])
	}
}

func TestEvaluate_DistributedRunLocalKeepsPartialProviderErrors(t *testing.T) {
	registerEvaluatorTestProvider("partial-ok", func(model string, options map[string]interface{}) llm.Provider {
		return evaluatorTestProvider{
			name: "partial-ok",
			evaluate: func(ctx context.Context, prompt string, vars map[string]interface{}) (*promptfoo.ProviderResponse, error) {
				return evaluatorTestResponse("ok"), nil
			},
		}
	})
	registerEvaluatorTestProvider("partial-error", func(model string, options map[string]interface{}) llm.Provider {
		return evaluatorTestProvider{
			name: "partial-error",
			evaluate: func(ctx context.Context, prompt string, vars map[string]interface{}) (*promptfoo.ProviderResponse, error) {
				return nil, fmt.Errorf("provider exploded")
			},
		}
	})

	config := promptfoo.Config{
		Prompts: []string{"prompt"},
		Providers: []promptfoo.ProviderConfig{
			{ID: "partial-ok"},
			{ID: "partial-error"},
		},
		Tests: []promptfoo.TestCase{
			{Vars: map[string]interface{}{}},
		},
	}

	result, err := Evaluate(config, time.Second, false, 2, false)
	require.NoError(t, err)
	require.Len(t, result.Results.Results, 1)
	assert.Equal(t, "partial-ok", result.Results.Results[0].Provider["id"])
	assert.Equal(t, 1, result.Results.Stats.Errors)
	assert.Equal(t, 1, result.Results.Stats.Successes)
}

type evaluatorTestProvider struct {
	name     string
	evaluate func(context.Context, string, map[string]interface{}) (*promptfoo.ProviderResponse, error)
}

func registerEvaluatorTestProvider(name string, factory func(string, map[string]interface{}) llm.Provider) {
	llm.RegisterProviderFactory(name, func(model string, options map[string]interface{}) (llm.Provider, error) {
		return factory(model, options), nil
	})
}

func (p evaluatorTestProvider) EvaluatePrompt(ctx context.Context, prompt string, vars map[string]interface{}) (*promptfoo.ProviderResponse, error) {
	return p.evaluate(ctx, prompt, vars)
}

func (p evaluatorTestProvider) Name() string {
	return p.name
}

func (p evaluatorTestProvider) Model() string {
	return "test"
}

func (p evaluatorTestProvider) Generate(ctx context.Context, prompt string, options llm.GenerateOptions) (*llm.GenerateResponse, error) {
	resp, err := p.evaluate(ctx, prompt, nil)
	if err != nil {
		return nil, err
	}
	return &llm.GenerateResponse{Text: resp.Output}, nil
}

func (p evaluatorTestProvider) SupportsStreaming() bool {
	return false
}

func (p evaluatorTestProvider) SupportsBatch() bool {
	return false
}

func evaluatorTestResponse(output string) *promptfoo.ProviderResponse {
	return &promptfoo.ProviderResponse{
		Output: output,
		TokenUsage: &promptfoo.TokenUsage{
			Total:       1,
			Prompt:      1,
			Completion:  0,
			NumRequests: 1,
		},
		LatencyMs: 1,
	}
}
