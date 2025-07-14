package distributed

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/tmc/pe/internal/promptfoo/evaluation/evaluator"
	"github.com/tmc/pe/internal/llm"
	"github.com/tmc/pe/internal/metaprompt"
	"github.com/tmc/pe/internal/promptfoo/evaluation/metrics"
	"github.com/tmc/pe/internal/promptfoo"
)

// Common task types for distributed execution

// EvalTask represents a distributed evaluation task
type EvalTask struct {
	taskID         string
	promptTemplate string
	variables      map[string]interface{}
	providerConfig map[string]interface{}
	assertions     []map[string]interface{}
	estimatedTime  time.Duration
	priority       int
}

func (t *EvalTask) ID() string { return t.taskID }
func (t *EvalTask) Execute(ctx context.Context) (interface{}, error) {
	// Convert to promptfoo config format
	config := promptfoo.Config{
		Prompts:   []string{t.promptTemplate},
		Providers: []string{"openai:gpt-3.5-turbo"}, // Default provider, can be overridden
		Tests: []promptfoo.TestCase{{
			Vars: t.variables,
		}},
	}

	// Override provider from config if specified
	if providerStr, ok := t.providerConfig["provider"].(string); ok {
		config.Providers = []string{providerStr}
	}

	// Add assertions
	if len(t.assertions) > 0 {
		asserts := make([]promptfoo.Assertion, len(t.assertions))
		for i, a := range t.assertions {
			if assertData, err := json.Marshal(a); err == nil {
				var assertion promptfoo.Assertion
				if err := json.Unmarshal(assertData, &assertion); err == nil {
					asserts[i] = assertion
				}
			}
		}
		config.Tests[0].Assert = asserts
	}

	// Run evaluation with reasonable defaults
	result, err := evaluator.Evaluate(config, 5*time.Minute, false, 1, false)
	if err != nil {
		return nil, fmt.Errorf("evaluation failed: %w", err)
	}

	return result, nil
}
func (t *EvalTask) Priority() int                    { return t.priority }
func (t *EvalTask) EstimatedDuration() time.Duration { return t.estimatedTime }
func (t *EvalTask) Type() string                     { return "eval" }

// OptimizeTask represents a distributed optimization task
type OptimizeTask struct {
	taskID         string
	initialPrompt  string
	objective      string
	method         string // pe2, textgrad, semantic, etc.
	iterations     int
	providerConfig map[string]interface{}
	estimatedTime  time.Duration
	priority       int
}

func (t *OptimizeTask) ID() string { return t.taskID }
func (t *OptimizeTask) Execute(ctx context.Context) (interface{}, error) {
	// Get provider from config
	providerStr := "openai:gpt-3.5-turbo" // default
	if ps, ok := t.providerConfig["provider"].(string); ok {
		providerStr = ps
	}

	// Parse provider string
	providerParts := strings.Split(providerStr, ":")
	backend := providerParts[0]

	// Get LLM provider
	provider, err := llm.GetProvider(backend)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider: %w", err)
	}

	// Create optimizer
	optimizer := metaprompt.NewOptimizer(provider)

	// Build config
	cfg := metaprompt.Config{
		InitialPrompt: t.initialPrompt,
		Objective:     t.objective,
		Method:        t.method,
		Iterations:    t.iterations,
		MaxIterations: t.iterations,
		Temperature:   0.7,
		MaxTokens:     2048,
	}

	// Run optimization
	result, err := optimizer.Optimize(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("optimization failed: %w", err)
	}

	return result, nil
}
func (t *OptimizeTask) Priority() int                    { return t.priority }
func (t *OptimizeTask) EstimatedDuration() time.Duration { return t.estimatedTime }
func (t *OptimizeTask) Type() string                     { return "optimize" }

// BenchmarkTask represents a distributed benchmark task
type BenchmarkTask struct {
	taskID         string
	promptTemplate string
	scenarios      []map[string]interface{}
	metrics        []string // latency, throughput, cost, etc.
	iterations     int
	providerConfig map[string]interface{}
	estimatedTime  time.Duration
	priority       int
}

func (t *BenchmarkTask) ID() string { return t.taskID }
func (t *BenchmarkTask) Execute(ctx context.Context) (interface{}, error) {
	// Get provider from config
	providerStr := "openai:gpt-3.5-turbo" // default
	if ps, ok := t.providerConfig["provider"].(string); ok {
		providerStr = ps
	}

	// Create benchmark results structure
	results := &BenchmarkTaskResult{
		TaskID:    t.taskID,
		Provider:  providerStr,
		Scenarios: make([]BenchmarkScenario, 0, len(t.scenarios)),
		StartTime: time.Now(),
	}

	// Run benchmark for each scenario
	for _, scenario := range t.scenarios {
		scenarioName := "default"
		if name, ok := scenario["name"].(string); ok {
			scenarioName = name
		}

		scenarioResult := BenchmarkScenario{
			Name:       scenarioName,
			Iterations: t.iterations,
			Metrics:    make(map[string]float64),
		}

		// Prepare variables from scenario
		vars := make(map[string]interface{})
		if v, ok := scenario["vars"].(map[string]interface{}); ok {
			vars = v
		}

		// Run iterations
		var totalLatency float64
		var totalTokens int32
		for i := 0; i < t.iterations; i++ {
			// Convert to promptfoo config for consistency
			config := promptfoo.Config{
				Prompts:   []string{t.promptTemplate},
				Providers: []string{providerStr},
				Tests: []promptfoo.TestCase{{
					Vars: vars,
				}},
			}

			start := time.Now()
			result, err := evaluator.Evaluate(config, 2*time.Minute, false, 1, false)
			if err != nil {
				return nil, fmt.Errorf("benchmark iteration %d failed: %w", i, err)
			}

			latency := time.Since(start).Seconds()
			totalLatency += latency

			// Extract token usage if available
			if len(result.Results.Results) > 0 {
				if usage := result.Results.Results[0].Response.TokenUsage; usage != nil {
					totalTokens += usage.Total
				}
			}
		}

		// Calculate metrics
		avgLatency := totalLatency / float64(t.iterations)
		scenarioResult.Metrics["avg_latency_seconds"] = avgLatency
		scenarioResult.Metrics["total_tokens"] = float64(totalTokens)
		if totalTokens > 0 {
			scenarioResult.Metrics["tokens_per_second"] = float64(totalTokens) / totalLatency
		}

		results.Scenarios = append(results.Scenarios, scenarioResult)
	}

	results.EndTime = time.Now()
	results.TotalDuration = results.EndTime.Sub(results.StartTime)

	return results, nil
}
func (t *BenchmarkTask) Priority() int                    { return t.priority }
func (t *BenchmarkTask) EstimatedDuration() time.Duration { return t.estimatedTime }
func (t *BenchmarkTask) Type() string                     { return "benchmark" }

// PassNTask represents a distributed pass@n evaluation task
type PassNTask struct {
	taskID         string
	promptTemplate string
	testCases      []map[string]interface{}
	n              int     // pass@n value
	samples        int     // number of samples to generate
	temperature    float64 // sampling temperature
	providerConfig map[string]interface{}
	estimatedTime  time.Duration
	priority       int
}

func (t *PassNTask) ID() string { return t.taskID }
func (t *PassNTask) Execute(ctx context.Context) (interface{}, error) {
	// Get provider from config
	providerStr := "openai:gpt-3.5-turbo" // default
	if ps, ok := t.providerConfig["provider"].(string); ok {
		providerStr = ps
	}

	// Parse provider string
	providerParts := strings.Split(providerStr, ":")
	backend := providerParts[0]

	// Get LLM provider
	provider, err := llm.GetProvider(backend)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider: %w", err)
	}

	// Create advanced metrics instance
	am := metrics.NewAdvancedMetrics(provider)

	// Generate samples if not enough provided
	var samples []string
	if t.samples > 0 {
		// Generate samples using the provider
		generator := metrics.NewPassNGenerator(provider, metrics.PassNConfig{
			Temperature: t.temperature,
			MaxTokens:   2048,
		})
		var err error
		samples, err = generator.GenerateSamples(ctx, t.promptTemplate, t.samples)
		if err != nil {
			return nil, fmt.Errorf("failed to generate samples: %w", err)
		}
	} else {
		// Use a single sample for testing
		var opts llm.GenerateOptions
		temp := t.temperature
		opts.Temperature = &temp
		resp, err := provider.Generate(ctx, t.promptTemplate, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to generate sample: %w", err)
		}
		samples = []string{resp.Text}
	}

	// Calculate pass@n with test cases
	result := am.CalculatePassAtNWithTests(ctx, t.n, samples, t.testCases)

	return result, nil
}
func (t *PassNTask) Priority() int                    { return t.priority }
func (t *PassNTask) EstimatedDuration() time.Duration { return t.estimatedTime }
func (t *PassNTask) Type() string                     { return "pass_n" }

// SemanticTask represents a distributed semantic backpropagation task
type SemanticTask struct {
	taskID         string
	promptTemplate string
	objective      string
	method         string // backprop, descent, gaso
	iterations     int
	learningRate   float64
	providerConfig map[string]interface{}
	estimatedTime  time.Duration
	priority       int
}

func (t *SemanticTask) ID() string { return t.taskID }
func (t *SemanticTask) Execute(ctx context.Context) (interface{}, error) {
	// Get provider from config
	providerStr := "openai:gpt-3.5-turbo" // default
	if ps, ok := t.providerConfig["provider"].(string); ok {
		providerStr = ps
	}

	// Parse provider string
	providerParts := strings.Split(providerStr, ":")
	backend := providerParts[0]

	// Get LLM provider
	provider, err := llm.GetProvider(backend)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider: %w", err)
	}

	// Create semantic optimizer
	semantic := metaprompt.NewSemanticOptimizer(provider)

	// Build config based on method
	var result interface{}
	switch t.method {
	case "backprop":
		cfg := metaprompt.SemanticConfig{
			Target:     t.objective,
			Iterations: t.iterations,
			Verbose:    false,
		}
		result, err = semantic.SemanticBackpropagation(ctx, t.promptTemplate, cfg)
	case "descent":
		cfg := metaprompt.SemanticDescentConfig{
			Objective:            t.objective,
			LearningRate:         t.learningRate,
			Iterations:           t.iterations,
			AdaptiveLearning:     true,
			ConvergenceThreshold: 0.001,
		}
		result, err = semantic.SemanticGradientDescent(ctx, t.promptTemplate, cfg)
	case "gaso":
		// For GASO, we need a GASOOptimizer
		gasoOptimizer := metaprompt.NewGASOOptimizer(provider)

		// Build system definition from providerConfig
		var systemDef *metaprompt.SystemDefinition
		if sd, ok := t.providerConfig["system_definition"]; ok {
			// Convert map to SystemDefinition
			sdBytes, _ := json.Marshal(sd)
			systemDef = &metaprompt.SystemDefinition{}
			json.Unmarshal(sdBytes, systemDef)
		} else {
			// Create a simple system definition from the prompt
			systemDef = &metaprompt.SystemDefinition{
				Name:        "Default System",
				Description: "Single-component system for prompt optimization",
				Components: []metaprompt.SystemComponent{{
					ID:      "main",
					Name:    "Main Prompt",
					Type:    "prompt",
					Content: t.promptTemplate,
				}},
				Objectives: []metaprompt.SystemObjective{{
					Name:        "primary",
					Description: t.objective,
					Weight:      1.0,
					Type:        "maximize",
				}},
			}
		}

		cfg := metaprompt.GASOConfig{
			Objective:        t.objective,
			Iterations:       t.iterations,
			MultiObjective:   true,
			OptimizationType: "pareto",
		}
		result, err = gasoOptimizer.OptimizeSystem(ctx, systemDef, cfg)
	default:
		return nil, fmt.Errorf("unknown semantic method: %s", t.method)
	}

	if err != nil {
		return nil, fmt.Errorf("semantic optimization failed: %w", err)
	}

	return result, nil
}
func (t *SemanticTask) Priority() int                    { return t.priority }
func (t *SemanticTask) EstimatedDuration() time.Duration { return t.estimatedTime }
func (t *SemanticTask) Type() string                     { return "semantic" }

// TaskFactory creates tasks from remote task payloads
type TaskFactory struct {
	// Dependencies are injected at runtime via the task Execute methods
	// This keeps the factory lightweight and stateless
}

// CreateTask creates a task from a remote task
func (f *TaskFactory) CreateTask(remoteTask *RemoteTask) (Task, error) {
	switch remoteTask.Type {
	case "*distributed.EvalTask":
		var task EvalTask
		if err := json.Unmarshal(remoteTask.Payload, &task); err != nil {
			return nil, fmt.Errorf("failed to unmarshal eval task: %w", err)
		}
		task.taskID = remoteTask.ID
		return &task, nil

	case "*distributed.OptimizeTask":
		var task OptimizeTask
		if err := json.Unmarshal(remoteTask.Payload, &task); err != nil {
			return nil, fmt.Errorf("failed to unmarshal optimize task: %w", err)
		}
		task.taskID = remoteTask.ID
		return &task, nil

	case "*distributed.BenchmarkTask":
		var task BenchmarkTask
		if err := json.Unmarshal(remoteTask.Payload, &task); err != nil {
			return nil, fmt.Errorf("failed to unmarshal benchmark task: %w", err)
		}
		task.taskID = remoteTask.ID
		return &task, nil

	case "*distributed.PassNTask":
		var task PassNTask
		if err := json.Unmarshal(remoteTask.Payload, &task); err != nil {
			return nil, fmt.Errorf("failed to unmarshal pass@n task: %w", err)
		}
		task.taskID = remoteTask.ID
		return &task, nil

	case "*distributed.SemanticTask":
		var task SemanticTask
		if err := json.Unmarshal(remoteTask.Payload, &task); err != nil {
			return nil, fmt.Errorf("failed to unmarshal semantic task: %w", err)
		}
		task.taskID = remoteTask.ID
		return &task, nil

	default:
		return nil, fmt.Errorf("unknown task type: %s", remoteTask.Type)
	}
}

// DistributedTaskResult contains the results of distributed task execution
type DistributedTaskResult struct {
	TaskID    string                 `json:"task_id"`
	Type      string                 `json:"type"`
	Success   bool                   `json:"success"`
	Error     string                 `json:"error,omitempty"`
	Metrics   map[string]interface{} `json:"metrics,omitempty"`
	Output    interface{}            `json:"output,omitempty"`
	Duration  time.Duration          `json:"duration"`
	Timestamp time.Time              `json:"timestamp"`
}

// MarshalJSON implements custom JSON marshaling for DistributedTaskResult
func (r DistributedTaskResult) MarshalJSON() ([]byte, error) {
	type Alias DistributedTaskResult
	return json.Marshal(&struct {
		Duration string `json:"duration"`
		*Alias
	}{
		Duration: r.Duration.String(),
		Alias:    (*Alias)(&r),
	})
}

// BenchmarkTaskResult represents the result of a benchmark task
type BenchmarkTaskResult struct {
	TaskID        string              `json:"task_id"`
	Provider      string              `json:"provider"`
	Scenarios     []BenchmarkScenario `json:"scenarios"`
	StartTime     time.Time           `json:"start_time"`
	EndTime       time.Time           `json:"end_time"`
	TotalDuration time.Duration       `json:"total_duration"`
}

// BenchmarkScenario represents a single benchmark scenario result
type BenchmarkScenario struct {
	Name       string             `json:"name"`
	Iterations int                `json:"iterations"`
	Metrics    map[string]float64 `json:"metrics"`
}
