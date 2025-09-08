package optimization

import (
	"context"
	"fmt"
	"time"
)

// OptimizationEngine coordinates different optimization strategies
type OptimizationEngine struct {
	strategies map[string]OptimizationStrategy
	evaluator  Evaluator
}

// NewOptimizationEngine creates a new optimization engine
func NewOptimizationEngine(evaluator Evaluator) *OptimizationEngine {
	engine := &OptimizationEngine{
		strategies: make(map[string]OptimizationStrategy),
		evaluator:  evaluator,
	}

	// Register built-in strategies
	engine.RegisterStrategy("single", &SingleOptimizerStrategy{})
	engine.RegisterStrategy("hybrid", &HybridStrategy{})
	engine.RegisterStrategy("multi", &MultiOptimizerStrategy{})
	engine.RegisterStrategy("comparative", &ComparativeStrategy{})

	return engine
}

// RegisterStrategy registers an optimization strategy
func (e *OptimizationEngine) RegisterStrategy(name string, strategy OptimizationStrategy) {
	e.strategies[name] = strategy
}

// GetStrategy returns a strategy by name
func (e *OptimizationEngine) GetStrategy(name string) (OptimizationStrategy, error) {
	strategy, exists := e.strategies[name]
	if !exists {
		return nil, fmt.Errorf("optimization strategy '%s' not found", name)
	}
	return strategy, nil
}

// ListStrategies returns all available strategy names
func (e *OptimizationEngine) ListStrategies() []string {
	names := make([]string, 0, len(e.strategies))
	for name := range e.strategies {
		names = append(names, name)
	}
	return names
}

// Optimize performs optimization using the specified strategy
func (e *OptimizationEngine) Optimize(ctx context.Context, optimizer Optimizer, config OptimizationConfig) (*OptimizationResult, error) {
	strategyName := config.Method
	if strategyName == "" {
		strategyName = "single" // default strategy
	}

	strategy, err := e.GetStrategy(strategyName)
	if err != nil {
		return nil, fmt.Errorf("failed to get strategy: %w", err)
	}

	return strategy.Execute(ctx, optimizer, config)
}

// SingleOptimizerStrategy executes optimization with a single optimizer
type SingleOptimizerStrategy struct{}

// Name returns the strategy name
func (s *SingleOptimizerStrategy) Name() string {
	return "single"
}

// SupportedOptimizers returns the optimizers this strategy supports
func (s *SingleOptimizerStrategy) SupportedOptimizers() []string {
	return []string{"*"} // supports all optimizers
}

// Execute runs single optimizer strategy
func (s *SingleOptimizerStrategy) Execute(ctx context.Context, optimizer Optimizer, config OptimizationConfig) (*OptimizationResult, error) {
	// Simply delegate to the optimizer
	return optimizer.Optimize(ctx, config)
}

// HybridStrategy combines multiple optimization approaches
type HybridStrategy struct{}

// Name returns the strategy name
func (s *HybridStrategy) Name() string {
	return "hybrid"
}

// SupportedOptimizers returns the optimizers this strategy supports
func (s *HybridStrategy) SupportedOptimizers() []string {
	return []string{"textgrad", "pe2", "gaso"}
}

// Execute runs hybrid optimization strategy
func (s *HybridStrategy) Execute(ctx context.Context, optimizer Optimizer, config OptimizationConfig) (*OptimizationResult, error) {
	startTime := time.Now()

	// Phase 1: Use primary optimizer for initial optimization
	phase1Config := config
	phase1Config.Iterations = config.Iterations / 2
	if phase1Config.Iterations < 1 {
		phase1Config.Iterations = 1
	}

	phase1Result, err := optimizer.Optimize(ctx, phase1Config)
	if err != nil {
		return nil, fmt.Errorf("phase 1 optimization failed: %w", err)
	}

	// Phase 2: Apply refinement (this could use a different optimizer)
	phase2Config := config
	phase2Config.InitialPrompt = phase1Result.OptimizedPrompt
	phase2Config.Iterations = config.Iterations - phase1Config.Iterations
	if phase2Config.Iterations < 1 {
		phase2Config.Iterations = 1
	}

	phase2Result, err := optimizer.Optimize(ctx, phase2Config)
	if err != nil {
		// If phase 2 fails, return phase 1 result
		return phase1Result, nil
	}

	// Combine results
	combinedResult := &OptimizationResult{
		OriginalPrompt:   phase1Result.OriginalPrompt,
		OptimizedPrompt:  phase2Result.OptimizedPrompt,
		Iterations:       append(phase1Result.Iterations, phase2Result.Iterations...),
		ImprovementScore: phase2Result.ImprovementScore,
		TotalDuration:    time.Since(startTime),
		CreatedAt:        phase1Result.CreatedAt,
		Method:           "hybrid",
		Converged:        phase2Result.Converged,
	}

	return combinedResult, nil
}

// MultiOptimizerStrategy runs multiple optimizers and selects the best result
type MultiOptimizerStrategy struct{}

// Name returns the strategy name
func (s *MultiOptimizerStrategy) Name() string {
	return "multi"
}

// SupportedOptimizers returns the optimizers this strategy supports
func (s *MultiOptimizerStrategy) SupportedOptimizers() []string {
	return []string{"*"} // can work with any set of optimizers
}

// Execute runs multiple optimizers and returns the best result
func (s *MultiOptimizerStrategy) Execute(ctx context.Context, optimizer Optimizer, config OptimizationConfig) (*OptimizationResult, error) {
	// For now, this just delegates to the single optimizer
	// In a full implementation, this would run multiple different optimizers
	// and compare their results
	return optimizer.Optimize(ctx, config)
}

// ComparativeStrategy uses comparative evaluation to guide optimization
type ComparativeStrategy struct{}

// Name returns the strategy name
func (s *ComparativeStrategy) Name() string {
	return "comparative"
}

// SupportedOptimizers returns the optimizers this strategy supports
func (s *ComparativeStrategy) SupportedOptimizers() []string {
	return []string{"*"}
}

// Execute runs comparative optimization strategy
func (s *ComparativeStrategy) Execute(ctx context.Context, optimizer Optimizer, config OptimizationConfig) (*OptimizationResult, error) {
	startTime := time.Now()

	result := &OptimizationResult{
		OriginalPrompt: config.InitialPrompt,
		Iterations:     make([]IterationResult, 0, config.Iterations),
		CreatedAt:      startTime,
		Method:         "comparative",
	}

	currentPrompt := config.InitialPrompt

	// For comparative strategy, we generate multiple candidates each iteration
	// and use comparative evaluation to select the best one
	for i := 0; i < config.Iterations; i++ {
		iterStart := time.Now()

		// Generate candidates using the optimizer
		candidateConfig := config
		candidateConfig.InitialPrompt = currentPrompt
		candidateConfig.Iterations = 1

		candidateResult, err := optimizer.Optimize(ctx, candidateConfig)
		if err != nil {
			return nil, fmt.Errorf("iteration %d failed: %w", i+1, err)
		}

		// For now, just use the candidate result
		// In a full implementation, this would generate multiple candidates
		// and use comparative evaluation
		if len(candidateResult.Iterations) > 0 {
			iteration := candidateResult.Iterations[0]
			iteration.Iteration = i + 1
			iteration.Duration = time.Since(iterStart)

			result.Iterations = append(result.Iterations, iteration)
			currentPrompt = iteration.Prompt
		}
	}

	result.OptimizedPrompt = currentPrompt
	if len(result.Iterations) > 0 {
		result.ImprovementScore = result.Iterations[len(result.Iterations)-1].Score
	}
	result.TotalDuration = time.Since(startTime)

	return result, nil
}

// IterativeImprovementStrategy performs iterative improvements with convergence checking
type IterativeImprovementStrategy struct {
	evaluator Evaluator
}

// NewIterativeImprovementStrategy creates a new iterative improvement strategy
func NewIterativeImprovementStrategy(evaluator Evaluator) *IterativeImprovementStrategy {
	return &IterativeImprovementStrategy{
		evaluator: evaluator,
	}
}

// Name returns the strategy name
func (s *IterativeImprovementStrategy) Name() string {
	return "iterative"
}

// SupportedOptimizers returns the optimizers this strategy supports
func (s *IterativeImprovementStrategy) SupportedOptimizers() []string {
	return []string{"*"}
}

// Execute runs iterative improvement strategy with convergence checking
func (s *IterativeImprovementStrategy) Execute(ctx context.Context, optimizer Optimizer, config OptimizationConfig) (*OptimizationResult, error) {
	startTime := time.Now()

	result := &OptimizationResult{
		OriginalPrompt: config.InitialPrompt,
		Iterations:     make([]IterationResult, 0, config.MaxIterations),
		CreatedAt:      startTime,
		Method:         "iterative",
	}

	currentPrompt := config.InitialPrompt
	previousScore := 0.0

	maxIters := config.MaxIterations
	if maxIters == 0 {
		maxIters = config.Iterations
	}
	if maxIters == 0 {
		maxIters = 10
	}

	convergenceThreshold := config.ConvergenceThreshold
	if convergenceThreshold == 0 {
		convergenceThreshold = 0.001
	}

	for i := 0; i < maxIters; i++ {
		iterStart := time.Now()

		// Optimize current prompt
		iterConfig := config
		iterConfig.InitialPrompt = currentPrompt
		iterConfig.Iterations = 1

		iterResult, err := optimizer.Optimize(ctx, iterConfig)
		if err != nil {
			return nil, fmt.Errorf("iteration %d failed: %w", i+1, err)
		}

		if len(iterResult.Iterations) == 0 {
			break
		}

		iteration := iterResult.Iterations[0]
		iteration.Iteration = i + 1
		iteration.Duration = time.Since(iterStart)

		result.Iterations = append(result.Iterations, iteration)

		// Check for convergence
		if i > 0 {
			improvement := iteration.Score - previousScore
			if improvement < convergenceThreshold {
				result.Converged = true
				break
			}
		}

		currentPrompt = iteration.Prompt
		previousScore = iteration.Score
	}

	result.OptimizedPrompt = currentPrompt
	if len(result.Iterations) > 0 {
		result.ImprovementScore = result.Iterations[len(result.Iterations)-1].Score
	}
	result.TotalDuration = time.Since(startTime)

	return result, nil
}
