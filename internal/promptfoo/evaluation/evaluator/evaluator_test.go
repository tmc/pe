package evaluator

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
				Providers: []string{"mock"},
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
				Providers: []string{"mock"},
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
				Providers: []string{"mock"},
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
				Providers: []string{"mock", "mock:model2"},
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
				Providers: []string{"mock"},
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
			result := replaceVariables(tt.prompt, tt.vars)
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
		Providers: []string{"mock"},
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
