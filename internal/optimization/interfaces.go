// Package optimization provides interfaces and implementations for optimization algorithms
// that are decoupled from specific provider implementations.
package optimization

import (
	"context"
	"time"
)

// LanguageModelProvider represents a provider for language model operations
// This interface abstracts away specific LLM providers
type LanguageModelProvider interface {
	// Name returns the provider identifier
	Name() string

	// Model returns the model being used
	Model() string

	// Generate generates text given a prompt and options
	Generate(ctx context.Context, prompt string, options GenerationOptions) (*GenerationResponse, error)

	// SupportsStreaming indicates if the provider supports streaming
	SupportsStreaming() bool

	// SupportsBatch indicates if the provider supports batch processing
	SupportsBatch() bool
}

// GenerationOptions contains options for text generation
type GenerationOptions struct {
	Temperature *float64
	MaxTokens   *int
	TopP        *float64
	TopK        *int
	Stop        []string
}

// GenerationResponse contains the response from a generation request
type GenerationResponse struct {
	Text             string
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
	Latency          time.Duration
	Cost             float64
	Model            string
	FinishReason     string
}

// Evaluator evaluates the quality of prompts or optimization outcomes
type Evaluator interface {
	// EvaluatePrompt scores a prompt against an objective
	EvaluatePrompt(ctx context.Context, prompt string, objective string) (float64, error)

	// EvaluateComparative compares two prompts and returns which is better
	EvaluateComparative(ctx context.Context, prompt1, prompt2 string, objective string) (ComparisonResult, error)

	// EvaluateSystem evaluates a multi-component system
	EvaluateSystem(ctx context.Context, system SystemDefinition, objective string) (float64, error)
}

// ComparisonResult represents the result of comparing two prompts
type ComparisonResult struct {
	Winner    int     // 1 for first prompt, 2 for second, 0 for tie
	Score1    float64 // Score of first prompt
	Score2    float64 // Score of second prompt
	Reasoning string  // Explanation of the comparison
}

// Optimizer defines the interface for optimization algorithms
type Optimizer interface {
	// Name returns the optimizer identifier
	Name() string

	// Optimize performs optimization using the given configuration
	Optimize(ctx context.Context, config OptimizationConfig) (*OptimizationResult, error)

	// SupportsObjective returns true if the optimizer supports the given objective type
	SupportsObjective(objectiveType string) bool

	// SupportsBatch returns true if the optimizer can optimize multiple prompts simultaneously
	SupportsBatch() bool
}

// GradientOptimizer extends Optimizer with gradient-based optimization capabilities
type GradientOptimizer interface {
	Optimizer

	// ComputeGradients computes gradients for the given input
	ComputeGradients(ctx context.Context, input string, objective string) ([]Gradient, error)

	// ApplyGradients applies gradients to optimize the input
	ApplyGradients(ctx context.Context, input string, gradients []Gradient) (string, error)
}

// SystemOptimizer extends Optimizer for multi-component system optimization
type SystemOptimizer interface {
	Optimizer

	// OptimizeSystem optimizes a multi-component system
	OptimizeSystem(ctx context.Context, system SystemDefinition, config OptimizationConfig) (*SystemOptimizationResult, error)

	// ComputeSystemGradients computes gradients for system components
	ComputeSystemGradients(ctx context.Context, system SystemDefinition, objective string) ([]SystemGradient, error)
}

// OptimizationConfig contains configuration for optimization
type OptimizationConfig struct {
	// Basic configuration
	InitialPrompt        string
	Objective            string
	Iterations           int
	MaxIterations        int
	ConvergenceThreshold float64

	// Algorithm-specific settings
	Method           string // "standard", "textgrad", "pe2", "gaso", etc.
	OptimizationType string // "single", "multi", "pareto", "weighted"
	Temperature      float64
	MaxTokens        int

	// Multi-objective settings
	MultiObjective bool
	Objectives     []ObjectiveSpec

	// Provider settings
	Provider string
	Model    string
	Options  map[string]interface{}
}

// ObjectiveSpec specifies an optimization objective
type ObjectiveSpec struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Weight      float64 `json:"weight"`
	Target      float64 `json:"target"`
	Type        string  `json:"type"` // "maximize", "minimize", "target"
}

// OptimizationResult contains the results of optimization
type OptimizationResult struct {
	OriginalPrompt   string            `json:"original_prompt"`
	OptimizedPrompt  string            `json:"optimized_prompt"`
	Iterations       []IterationResult `json:"iterations"`
	ImprovementScore float64           `json:"improvement_score"`
	TotalDuration    time.Duration     `json:"total_duration"`
	CreatedAt        time.Time         `json:"created_at"`
	Method           string            `json:"method"`
	Converged        bool              `json:"converged"`
}

// IterationResult contains the result of a single optimization iteration
type IterationResult struct {
	Iteration   int           `json:"iteration"`
	Prompt      string        `json:"prompt"`
	Score       float64       `json:"score"`
	Feedback    string        `json:"feedback"`
	Suggestions []string      `json:"suggestions"`
	Duration    time.Duration `json:"duration"`
	Changes     []string      `json:"changes,omitempty"`
	Timestamp   time.Time     `json:"timestamp,omitempty"`
	Gradients   []Gradient    `json:"gradients,omitempty"`
}

// Gradient represents a gradient for optimization
type Gradient struct {
	Component   string   `json:"component"`
	Feedback    string   `json:"feedback"`
	Suggestions []string `json:"suggestions"`
	Confidence  float64  `json:"confidence"`
	Priority    float64  `json:"priority"`
	Direction   string   `json:"direction"`
	Magnitude   float64  `json:"magnitude"`
}

// SystemDefinition defines a multi-component system
type SystemDefinition struct {
	Name         string                `json:"name"`
	Description  string                `json:"description"`
	Components   []SystemComponent     `json:"components"`
	Dependencies []ComponentDependency `json:"dependencies"`
	Objectives   []ObjectiveSpec       `json:"objectives"`
	Constraints  []SystemConstraint    `json:"constraints"`
}

// SystemComponent represents a single component in a system
type SystemComponent struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Type        string                 `json:"type"` // "prompt", "tool", "agent", "workflow"
	Content     string                 `json:"content"`
	Parameters  map[string]interface{} `json:"parameters"`
	Performance float64                `json:"performance"`
}

// ComponentDependency represents a dependency between components
type ComponentDependency struct {
	From        string  `json:"from"`
	To          string  `json:"to"`
	Type        string  `json:"type"` // "input", "control", "data"
	Weight      float64 `json:"weight"`
	Description string  `json:"description"`
}

// SystemConstraint represents a constraint on system optimization
type SystemConstraint struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Type        string      `json:"type"` // "hard", "soft"
	Value       interface{} `json:"value"`
}

// SystemGradient represents a gradient for a system component
type SystemGradient struct {
	ComponentID string   `json:"component_id"`
	Gradient    Gradient `json:"gradient"`
	Priority    float64  `json:"priority"`
}

// SystemOptimizationResult contains results of system optimization
type SystemOptimizationResult struct {
	OptimizedComponents []SystemComponent       `json:"optimized_components"`
	OverallPerformance  float64                 `json:"overall_performance"`
	ObjectiveScores     map[string]float64      `json:"objective_scores"`
	ParetoEfficient     bool                    `json:"pareto_efficient"`
	OptimizationHistory []SystemIterationResult `json:"optimization_history"`
	Iterations          int                     `json:"iterations"`
	Duration            time.Duration           `json:"duration"`
	ConvergenceAnalysis *ConvergenceAnalysis    `json:"convergence_analysis"`
}

// SystemIterationResult represents a system optimization iteration
type SystemIterationResult struct {
	Iteration   int                `json:"iteration"`
	Performance float64            `json:"performance"`
	Objectives  map[string]float64 `json:"objectives"`
	Gradients   []SystemGradient   `json:"gradients"`
	Changes     []ComponentChange  `json:"changes"`
	Improvement float64            `json:"improvement"`
}

// ComponentChange represents a change to a system component
type ComponentChange struct {
	ComponentID string  `json:"component_id"`
	ChangeType  string  `json:"change_type"`
	OldValue    string  `json:"old_value"`
	NewValue    string  `json:"new_value"`
	Impact      float64 `json:"impact"`
}

// ConvergenceAnalysis analyzes convergence properties
type ConvergenceAnalysis struct {
	Converged       bool    `json:"converged"`
	ConvergenceRate float64 `json:"convergence_rate"`
	FinalGradient   float64 `json:"final_gradient"`
	StabilityMetric float64 `json:"stability_metric"`
}

// OptimizationStrategy defines how optimization should be performed
type OptimizationStrategy interface {
	// Execute runs the optimization strategy
	Execute(ctx context.Context, optimizer Optimizer, config OptimizationConfig) (*OptimizationResult, error)

	// Name returns the strategy identifier
	Name() string

	// SupportedOptimizers returns the list of optimizers this strategy can work with
	SupportedOptimizers() []string
}
