package optimizers

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/tmc/pe/internal/optimization"
)

// GASOOptimizer implements Graph-based Agentic System Optimization
// Based on 2025 KAUST/IDSIA research on semantic optimization for multi-component systems
type GASOOptimizer struct {
	provider  optimization.LanguageModelProvider
	evaluator optimization.Evaluator
}

// NewGASOOptimizer creates a new GASO optimizer
func NewGASOOptimizer(provider optimization.LanguageModelProvider, evaluator optimization.Evaluator) optimization.SystemOptimizer {
	return &GASOOptimizer{
		provider:  provider,
		evaluator: evaluator,
	}
}

// Name returns the optimizer identifier
func (o *GASOOptimizer) Name() string {
	return "gaso"
}

// SupportsObjective returns true if the optimizer supports the given objective type
func (o *GASOOptimizer) SupportsObjective(objectiveType string) bool {
	// GASO supports system-level objectives
	supportedTypes := []string{"system", "multi", "component", "integration", "synergy", "efficiency"}
	for _, supported := range supportedTypes {
		if strings.Contains(strings.ToLower(objectiveType), supported) {
			return true
		}
	}
	return true // Default to true for flexibility
}

// SupportsBatch returns true if the optimizer can optimize multiple prompts simultaneously
func (o *GASOOptimizer) SupportsBatch() bool {
	return true // GASO naturally handles multiple components
}

// Optimize performs single-prompt optimization (delegates to system optimization)
func (o *GASOOptimizer) Optimize(ctx context.Context, config optimization.OptimizationConfig) (*optimization.OptimizationResult, error) {
	// Convert single prompt to a simple system
	system := optimization.SystemDefinition{
		Name:        "Single Component System",
		Description: "System with a single prompt component for optimization",
		Components: []optimization.SystemComponent{
			{
				ID:          "main",
				Name:        "Main Prompt",
				Type:        "prompt",
				Content:     config.InitialPrompt,
				Parameters:  make(map[string]interface{}),
				Performance: 0.0,
			},
		},
		Dependencies: []optimization.ComponentDependency{},
		Objectives:   []optimization.ObjectiveSpec{},
		Constraints:  []optimization.SystemConstraint{},
	}

	// Use system optimization
	systemResult, err := o.OptimizeSystem(ctx, system, config)
	if err != nil {
		return nil, fmt.Errorf("system optimization failed: %w", err)
	}

	// Convert system result back to optimization result
	result := &optimization.OptimizationResult{
		OriginalPrompt:   config.InitialPrompt,
		ImprovementScore: systemResult.OverallPerformance,
		TotalDuration:    systemResult.Duration,
		CreatedAt:        time.Now(),
		Method:           "gaso",
		Converged:        systemResult.ConvergenceAnalysis != nil && systemResult.ConvergenceAnalysis.Converged,
	}

	// Extract optimized prompt from first component
	if len(systemResult.OptimizedComponents) > 0 {
		result.OptimizedPrompt = systemResult.OptimizedComponents[0].Content
	} else {
		result.OptimizedPrompt = config.InitialPrompt
	}

	// Convert system iterations to standard iterations
	for _, sysIter := range systemResult.OptimizationHistory {
		var gradients []optimization.Gradient
		for _, sysGrad := range sysIter.Gradients {
			gradients = append(gradients, sysGrad.Gradient)
		}

		var changes []string
		for _, change := range sysIter.Changes {
			changes = append(changes, fmt.Sprintf("%s: %s -> %s", change.ChangeType, change.OldValue[:min(50, len(change.OldValue))], change.NewValue[:min(50, len(change.NewValue))]))
		}

		iteration := optimization.IterationResult{
			Iteration: sysIter.Iteration,
			Prompt:    result.OptimizedPrompt, // Use final optimized prompt
			Score:     sysIter.Performance,
			Changes:   changes,
			Timestamp: time.Now(),
			Gradients: gradients,
		}
		result.Iterations = append(result.Iterations, iteration)
	}

	return result, nil
}

// OptimizeSystem optimizes a multi-component system
func (o *GASOOptimizer) OptimizeSystem(ctx context.Context, system optimization.SystemDefinition, config optimization.OptimizationConfig) (*optimization.SystemOptimizationResult, error) {
	start := time.Now()

	result := &optimization.SystemOptimizationResult{
		OptimizedComponents: make([]optimization.SystemComponent, len(system.Components)),
		ObjectiveScores:     make(map[string]float64),
		OptimizationHistory: make([]optimization.SystemIterationResult, 0, config.Iterations),
	}

	// Copy initial components
	copy(result.OptimizedComponents, system.Components)

	// Initial system evaluation
	initialPerformance, err := o.evaluator.EvaluateSystem(ctx, system, config.Objective)
	if err != nil {
		return nil, fmt.Errorf("initial system evaluation failed: %w", err)
	}
	result.OverallPerformance = initialPerformance

	// GASO optimization iterations
	currentSystem := system
	maxIters := config.MaxIterations
	if maxIters == 0 {
		maxIters = config.Iterations
	}
	if maxIters == 0 {
		maxIters = 5
	}

	for i := 0; i < maxIters; i++ {
		// Compute semantic gradients for each component
		systemGradients, err := o.ComputeSystemGradients(ctx, currentSystem, config.Objective)
		if err != nil {
			return nil, fmt.Errorf("gradient computation failed at iteration %d: %w", i+1, err)
		}

		// Apply multi-component optimization
		optimizedSystem, changes, err := o.applySystemGradients(ctx, &currentSystem, systemGradients)
		if err != nil {
			return nil, fmt.Errorf("system optimization failed at iteration %d: %w", i+1, err)
		}

		// Evaluate optimized system
		newPerformance, err := o.evaluator.EvaluateSystem(ctx, *optimizedSystem, config.Objective)
		if err != nil {
			return nil, fmt.Errorf("system evaluation failed at iteration %d: %w", i+1, err)
		}

		// Record iteration
		iteration := optimization.SystemIterationResult{
			Iteration:   i + 1,
			Performance: newPerformance,
			Objectives:  make(map[string]float64),
			Gradients:   systemGradients,
			Changes:     changes,
			Improvement: newPerformance - result.OverallPerformance,
		}
		result.OptimizationHistory = append(result.OptimizationHistory, iteration)

		// Update best result if improvement
		if newPerformance > result.OverallPerformance {
			result.OverallPerformance = newPerformance
			result.OptimizedComponents = optimizedSystem.Components
		}

		currentSystem = *optimizedSystem

		// Check convergence
		if config.ConvergenceThreshold > 0 && iteration.Improvement < config.ConvergenceThreshold {
			break
		}
	}

	// Analyze convergence
	result.ConvergenceAnalysis = o.analyzeConvergence(result.OptimizationHistory)
	result.Iterations = len(result.OptimizationHistory)
	result.Duration = time.Since(start)

	return result, nil
}

// ComputeSystemGradients computes gradients for system components
func (o *GASOOptimizer) ComputeSystemGradients(ctx context.Context, system optimization.SystemDefinition, objective string) ([]optimization.SystemGradient, error) {
	gradients := make([]optimization.SystemGradient, 0, len(system.Components))

	for _, component := range system.Components {
		// Compute gradients for this component considering system context
		gradient, err := o.computeComponentGradient(ctx, component, system, objective)
		if err != nil {
			return nil, fmt.Errorf("failed to compute gradient for component %s: %w", component.ID, err)
		}

		// Calculate priority based on component's role in system
		priority := o.calculateComponentPriority(component, system)

		systemGrad := optimization.SystemGradient{
			ComponentID: component.ID,
			Gradient:    gradient,
			Priority:    priority,
		}
		gradients = append(gradients, systemGrad)
	}

	return gradients, nil
}

// computeComponentGradient computes semantic gradient for a specific component
func (o *GASOOptimizer) computeComponentGradient(ctx context.Context, component optimization.SystemComponent, system optimization.SystemDefinition, objective string) (optimization.Gradient, error) {
	gradientPrompt := fmt.Sprintf(`You are optimizing a component within a multi-component agentic system.

SYSTEM CONTEXT:
System: %s
Objective: %s

COMPONENT TO OPTIMIZE:
ID: %s
Name: %s
Type: %s
Content: %s

DEPENDENCIES:
%s

Provide a semantic gradient for optimizing this component within the system context.
Consider how changes to this component will affect dependent components and overall system performance.

Response format (JSON):
{
  "component": "specific aspect to modify",
  "feedback": "what needs improvement",
  "suggestions": ["suggestion 1", "suggestion 2"],
  "confidence": 0.9,
  "priority": 0.8,
  "direction": "how to modify it",
  "magnitude": 0.7
}`, system.Description, objective, component.ID, component.Name, component.Type, component.Content, o.formatDependencies(component.ID, system))

	options := optimization.GenerationOptions{
		Temperature: &[]float64{0.2}[0],
		MaxTokens:   &[]int{500}[0],
	}

	response, err := o.provider.Generate(ctx, gradientPrompt, options)
	if err != nil {
		return optimization.Gradient{}, fmt.Errorf("failed to generate component gradient: %w", err)
	}

	// Parse gradient from response
	gradient, err := o.parseComponentGradient(response.Text)
	if err != nil {
		// Fallback gradient
		return optimization.Gradient{
			Component:   "content optimization",
			Feedback:    "Component needs optimization for system coherence",
			Suggestions: []string{"improve clarity and effectiveness", "enhance system integration"},
			Direction:   "improve",
			Magnitude:   0.7,
			Confidence:  0.6,
			Priority:    0.7,
		}, nil
	}

	return gradient, nil
}

// applySystemGradients applies gradients to optimize the entire system
func (o *GASOOptimizer) applySystemGradients(ctx context.Context, system *optimization.SystemDefinition, gradients []optimization.SystemGradient) (*optimization.SystemDefinition, []optimization.ComponentChange, error) {
	optimizedSystem := *system // Copy system
	changes := make([]optimization.ComponentChange, 0)

	// Sort gradients by priority (highest first)
	sort.Slice(gradients, func(i, j int) bool {
		return gradients[i].Priority > gradients[j].Priority
	})

	// Apply gradients to components
	for _, sysGrad := range gradients {
		// Find component
		for i, component := range optimizedSystem.Components {
			if component.ID == sysGrad.ComponentID {
				// Apply gradient to component
				optimizedContent, err := o.applyComponentGradient(ctx, component, sysGrad.Gradient)
				if err != nil {
					return nil, nil, fmt.Errorf("failed to apply gradient to component %s: %w", component.ID, err)
				}

				// Record change
				change := optimization.ComponentChange{
					ComponentID: component.ID,
					ChangeType:  "content_optimization",
					OldValue:    component.Content,
					NewValue:    optimizedContent,
					Impact:      sysGrad.Priority,
				}
				changes = append(changes, change)

				// Update component
				optimizedSystem.Components[i].Content = optimizedContent
				break
			}
		}
	}

	return &optimizedSystem, changes, nil
}

// applyComponentGradient applies a semantic gradient to a specific component
func (o *GASOOptimizer) applyComponentGradient(ctx context.Context, component optimization.SystemComponent, gradient optimization.Gradient) (string, error) {
	optimizationPrompt := fmt.Sprintf(`Optimize this system component using the provided semantic gradient:

COMPONENT:
Type: %s
Current Content: %s

SEMANTIC GRADIENT:
Component: %s
Direction: %s
Feedback: %s
Suggestions: %s

Apply this gradient to optimize the component content.
Return only the optimized content without additional explanation.`,
		component.Type, component.Content, gradient.Component, gradient.Direction, gradient.Feedback, strings.Join(gradient.Suggestions, ", "))

	options := optimization.GenerationOptions{
		Temperature: &[]float64{0.3}[0],
		MaxTokens:   &[]int{500}[0],
	}

	response, err := o.provider.Generate(ctx, optimizationPrompt, options)
	if err != nil {
		return "", fmt.Errorf("failed to apply component gradient: %w", err)
	}

	return strings.TrimSpace(response.Text), nil
}

// Helper functions

func (o *GASOOptimizer) calculateComponentPriority(component optimization.SystemComponent, system optimization.SystemDefinition) float64 {
	priority := 0.5 // Base priority

	// Count dependencies
	outgoingCount := 0
	incomingCount := 0
	totalWeight := 0.0

	for _, dep := range system.Dependencies {
		if dep.From == component.ID {
			outgoingCount++
			totalWeight += dep.Weight
		}
		if dep.To == component.ID {
			incomingCount++
			totalWeight += dep.Weight
		}
	}

	// Components that affect many others have higher priority
	priority += float64(outgoingCount) * 0.1

	// Components that are central to the system have higher priority
	priority += float64(incomingCount) * 0.05

	// Weight-based priority adjustment
	priority += totalWeight * 0.1

	// Component type-based priority
	switch component.Type {
	case "agent":
		priority += 0.2
	case "workflow":
		priority += 0.15
	case "prompt":
		priority += 0.1
	case "tool":
		priority += 0.05
	}

	// Normalize to [0, 1]
	return math.Min(1.0, math.Max(0.0, priority))
}

func (o *GASOOptimizer) formatDependencies(componentID string, system optimization.SystemDefinition) string {
	result := ""
	for _, dep := range system.Dependencies {
		if dep.From == componentID || dep.To == componentID {
			result += fmt.Sprintf("- %s -> %s (%s)\n", dep.From, dep.To, dep.Type)
		}
	}
	if result == "" {
		result = "No direct dependencies"
	}
	return result
}

func (o *GASOOptimizer) analyzeConvergence(history []optimization.SystemIterationResult) *optimization.ConvergenceAnalysis {
	if len(history) < 2 {
		return &optimization.ConvergenceAnalysis{
			Converged:       false,
			ConvergenceRate: 0.0,
			FinalGradient:   1.0,
			StabilityMetric: 0.0,
		}
	}

	// Calculate convergence metrics
	lastN := 5
	if len(history) < lastN {
		lastN = len(history)
	}

	// Check if improvements are decreasing (convergence indicator)
	improvements := make([]float64, 0, lastN)
	for i := len(history) - lastN; i < len(history); i++ {
		improvements = append(improvements, math.Abs(history[i].Improvement))
	}

	// Calculate average improvement over last N iterations
	avgImprovement := 0.0
	for _, imp := range improvements {
		avgImprovement += imp
	}
	avgImprovement /= float64(len(improvements))

	// Calculate stability (variance of improvements)
	variance := 0.0
	for _, imp := range improvements {
		variance += math.Pow(imp-avgImprovement, 2)
	}
	variance /= float64(len(improvements))
	stability := 1.0 - math.Min(1.0, math.Sqrt(variance))

	// Determine convergence
	converged := avgImprovement < 0.01 && stability > 0.8

	// Calculate convergence rate
	convergenceRate := 0.0
	if len(history) > 1 {
		initialPerf := history[0].Performance
		finalPerf := history[len(history)-1].Performance
		if finalPerf > initialPerf {
			convergenceRate = (finalPerf - initialPerf) / float64(len(history))
		}
	}

	return &optimization.ConvergenceAnalysis{
		Converged:       converged,
		ConvergenceRate: convergenceRate,
		FinalGradient:   avgImprovement,
		StabilityMetric: stability,
	}
}

func (o *GASOOptimizer) parseComponentGradient(response string) (optimization.Gradient, error) {
	// Find JSON in the response
	start := strings.Index(response, "{")
	end := strings.LastIndex(response, "}")
	if start == -1 || end == -1 || start > end {
		return optimization.Gradient{}, fmt.Errorf("no JSON found in response")
	}

	jsonStr := response[start : end+1]

	var gradient optimization.Gradient
	if err := json.Unmarshal([]byte(jsonStr), &gradient); err != nil {
		return optimization.Gradient{}, fmt.Errorf("failed to parse gradient JSON: %w", err)
	}

	// Validate and normalize
	if gradient.Magnitude < 0 {
		gradient.Magnitude = 0
	} else if gradient.Magnitude > 1 {
		gradient.Magnitude = 1
	}
	if gradient.Confidence < 0 {
		gradient.Confidence = 0
	} else if gradient.Confidence > 1 {
		gradient.Confidence = 1
	}
	if gradient.Priority < 0 {
		gradient.Priority = 0
	} else if gradient.Priority > 1 {
		gradient.Priority = 1
	}

	return gradient, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
