package metaprompt

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/tmc/pe/internal/llm"
)

// GASOOptimizer implements Graph-based Agentic System Optimization
// Based on 2025 KAUST/IDSIA research on semantic optimization for multi-component systems
type GASOOptimizer struct {
	llm llm.Provider
}

// NewGASOOptimizer creates a new GASO optimizer
func NewGASOOptimizer(llmProvider llm.Provider) *GASOOptimizer {
	return &GASOOptimizer{
		llm: llmProvider,
	}
}

// SystemDefinition defines a multi-component agentic system
type SystemDefinition struct {
	Name         string                 `json:"name"`
	Description  string                 `json:"description"`
	Components   []SystemComponent      `json:"components"`
	Dependencies []ComponentDependency  `json:"dependencies"`
	Objectives   []SystemObjective      `json:"objectives"`
	Constraints  []SystemConstraint     `json:"constraints"`
}

// SystemComponent represents a single component in the agentic system
type SystemComponent struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Type        string                 `json:"type"` // "prompt", "tool", "agent", "workflow"
	Content     string                 `json:"content"`
	Parameters  map[string]interface{} `json:"parameters"`
	Performance float64               `json:"performance"`
}

// ComponentDependency represents a dependency relationship between components
type ComponentDependency struct {
	From         string  `json:"from"`
	To           string  `json:"to"`
	Type         string  `json:"type"` // "input", "control", "data"
	Weight       float64 `json:"weight"`
	Description  string  `json:"description"`
}

// SystemObjective represents an optimization objective for the system
type SystemObjective struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Weight      float64 `json:"weight"`
	Target      float64 `json:"target"`
	Type        string  `json:"type"` // "maximize", "minimize", "target"
}

// SystemConstraint represents a constraint on the system optimization
type SystemConstraint struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Type        string      `json:"type"` // "hard", "soft"
	Value       interface{} `json:"value"`
}

// GASOConfig configures GASO optimization
type GASOConfig struct {
	Objective        string
	Iterations       int
	OptimizationType string // "pareto", "weighted", "lexicographic"
	MultiObjective   bool
	GraphFile        string
	Provider         string
	Model            string
}

// GASOResult represents the result of GASO optimization
type GASOResult struct {
	OptimizedComponents  []SystemComponent      `json:"optimized_components"`
	OverallPerformance   float64               `json:"overall_performance"`
	ObjectiveScores      map[string]float64    `json:"objective_scores"`
	ParetoEfficient      bool                  `json:"pareto_efficient"`
	ComputationalGraph   *GASOComputationalGraph   `json:"computational_graph"`
	OptimizationHistory  []GASOIteration       `json:"optimization_history"`
	Iterations          int                   `json:"iterations"`
	Duration            time.Duration         `json:"duration"`
	ConvergenceAnalysis *ConvergenceAnalysis  `json:"convergence_analysis"`
}

// GASOComputationalGraph represents the computational graph of the system
type GASOComputationalGraph struct {
	Nodes []GASOGraphNode `json:"nodes"`
	Edges []GASOGraphEdge `json:"edges"`
}

// GASOGraphNode represents a node in the computational graph
type GASOGraphNode struct {
	ID       string                 `json:"id"`
	Type     string                 `json:"type"`
	Data     map[string]interface{} `json:"data"`
	Position map[string]float64     `json:"position"`
}

// GASOGraphEdge represents an edge in the computational graph
type GASOGraphEdge struct {
	From   string                 `json:"from"`
	To     string                 `json:"to"`
	Type   string                 `json:"type"`
	Weight float64               `json:"weight"`
	Data   map[string]interface{} `json:"data"`
}

// GASOIteration represents a single GASO optimization iteration
type GASOIteration struct {
	Iteration    int                   `json:"iteration"`
	Performance  float64              `json:"performance"`
	Objectives   map[string]float64   `json:"objectives"`
	Gradients    []SystemGradient     `json:"gradients"`
	Changes      []ComponentChange     `json:"changes"`
	Improvement  float64              `json:"improvement"`
}

// SystemGradient represents a semantic gradient for a system component
type SystemGradient struct {
	ComponentID string  `json:"component_id"`
	Gradient    SemanticGradient `json:"gradient"`
	Priority    float64 `json:"priority"`
}

// ComponentChange represents a change made to a system component
type ComponentChange struct {
	ComponentID  string `json:"component_id"`
	ChangeType   string `json:"change_type"`
	OldValue     string `json:"old_value"`
	NewValue     string `json:"new_value"`
	Impact       float64 `json:"impact"`
}

// ConvergenceAnalysis analyzes the convergence properties of the optimization
type ConvergenceAnalysis struct {
	Converged        bool    `json:"converged"`
	ConvergenceRate  float64 `json:"convergence_rate"`
	FinalGradient    float64 `json:"final_gradient"`
	StabilityMetric  float64 `json:"stability_metric"`
}

// OptimizeSystem optimizes a multi-component agentic system using GASO
func (gaso *GASOOptimizer) OptimizeSystem(ctx context.Context, system *SystemDefinition, config GASOConfig) (*GASOResult, error) {
	start := time.Now()
	
	result := &GASOResult{
		OptimizedComponents: make([]SystemComponent, len(system.Components)),
		ObjectiveScores:     make(map[string]float64),
		OptimizationHistory: make([]GASOIteration, 0, config.Iterations),
	}
	
	// Copy initial components
	copy(result.OptimizedComponents, system.Components)
	
	// Build computational graph
	graph, err := gaso.buildComputationalGraph(system)
	if err != nil {
		return nil, fmt.Errorf("failed to build computational graph: %v", err)
	}
	result.ComputationalGraph = graph
	
	// Initial system evaluation
	initialPerformance, err := gaso.evaluateSystem(ctx, system, config.Objective)
	if err != nil {
		return nil, fmt.Errorf("initial system evaluation failed: %v", err)
	}
	result.OverallPerformance = initialPerformance
	
	fmt.Printf("Initial system performance: %.4f\n", initialPerformance)
	
	// GASO optimization iterations
	currentSystem := system
	for i := 0; i < config.Iterations; i++ {
		// Compute semantic gradients for each component
		systemGradients, err := gaso.computeSystemGradients(ctx, currentSystem, config.Objective)
		if err != nil {
			return nil, fmt.Errorf("gradient computation failed at iteration %d: %v", i, err)
		}
		
		// Apply multi-component optimization
		optimizedSystem, changes, err := gaso.applySystemGradients(ctx, currentSystem, systemGradients, config)
		if err != nil {
			return nil, fmt.Errorf("system optimization failed at iteration %d: %v", i, err)
		}
		
		// Evaluate optimized system
		newPerformance, err := gaso.evaluateSystem(ctx, optimizedSystem, config.Objective)
		if err != nil {
			return nil, fmt.Errorf("system evaluation failed at iteration %d: %v", i, err)
		}
		
		// Evaluate individual objectives if multi-objective
		objectiveScores := make(map[string]float64)
		if config.MultiObjective {
			for _, obj := range system.Objectives {
				score, err := gaso.evaluateObjective(ctx, optimizedSystem, obj)
				if err != nil {
					return nil, fmt.Errorf("objective evaluation failed: %v", err)
				}
				objectiveScores[obj.Name] = score
			}
		}
		
		// Record iteration
		iteration := GASOIteration{
			Iteration:   i + 1,
			Performance: newPerformance,
			Objectives:  objectiveScores,
			Gradients:   systemGradients,
			Changes:     changes,
			Improvement: newPerformance - result.OverallPerformance,
		}
		result.OptimizationHistory = append(result.OptimizationHistory, iteration)
		
		fmt.Printf("Iteration %d: performance %.4f (Δ%.4f)\n", i+1, newPerformance, iteration.Improvement)
		
		// Update best result if improvement
		if newPerformance > result.OverallPerformance {
			result.OverallPerformance = newPerformance
			result.OptimizedComponents = optimizedSystem.Components
			result.ObjectiveScores = objectiveScores
		}
		
		currentSystem = optimizedSystem
	}
	
	// Analyze convergence
	result.ConvergenceAnalysis = gaso.analyzeConvergence(result.OptimizationHistory)
	result.Iterations = config.Iterations
	result.Duration = time.Since(start)
	
	// Check Pareto efficiency for multi-objective optimization
	if config.MultiObjective {
		result.ParetoEfficient = gaso.checkParetoEfficiency(result.ObjectiveScores, system.Objectives)
	}
	
	return result, nil
}

// buildComputationalGraph builds a computational graph representation of the system
func (gaso *GASOOptimizer) buildComputationalGraph(system *SystemDefinition) (*GASOComputationalGraph, error) {
	graph := &GASOComputationalGraph{
		Nodes: make([]GASOGraphNode, 0, len(system.Components)),
		Edges: make([]GASOGraphEdge, 0, len(system.Dependencies)),
	}
	
	// Create nodes for each component
	for i, component := range system.Components {
		node := GASOGraphNode{
			ID:   component.ID,
			Type: component.Type,
			Data: map[string]interface{}{
				"name":        component.Name,
				"content":     component.Content,
				"parameters":  component.Parameters,
				"performance": component.Performance,
			},
			Position: map[string]float64{
				"x": float64(i * 100),
				"y": float64(i * 50),
			},
		}
		graph.Nodes = append(graph.Nodes, node)
	}
	
	// Create edges for dependencies
	for _, dep := range system.Dependencies {
		edge := GASOGraphEdge{
			From:   dep.From,
			To:     dep.To,
			Type:   dep.Type,
			Weight: dep.Weight,
			Data: map[string]interface{}{
				"description": dep.Description,
			},
		}
		graph.Edges = append(graph.Edges, edge)
	}
	
	return graph, nil
}

// computeSystemGradients computes semantic gradients for all system components
func (gaso *GASOOptimizer) computeSystemGradients(ctx context.Context, system *SystemDefinition, objective string) ([]SystemGradient, error) {
	gradients := make([]SystemGradient, 0, len(system.Components))
	
	for _, component := range system.Components {
		// Compute gradients for this component considering system context
		gradient, err := gaso.computeComponentGradient(ctx, component, system, objective)
		if err != nil {
			return nil, fmt.Errorf("failed to compute gradient for component %s: %v", component.ID, err)
		}
		
		// Calculate priority based on component's role in system
		priority := gaso.calculateComponentPriority(component, system)
		
		systemGrad := SystemGradient{
			ComponentID: component.ID,
			Gradient:    gradient,
			Priority:    priority,
		}
		gradients = append(gradients, systemGrad)
	}
	
	return gradients, nil
}

// computeComponentGradient computes semantic gradient for a specific component
func (gaso *GASOOptimizer) computeComponentGradient(ctx context.Context, component SystemComponent, system *SystemDefinition, objective string) (SemanticGradient, error) {
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

Response format:
{
  "component": "specific aspect to modify",
  "direction": "how to modify it",
  "magnitude": 0.8,
  "reasoning": "why this change improves system performance",
  "confidence": 0.9
}`, system.Description, objective, component.ID, component.Name, component.Type, component.Content, gaso.formatDependencies(component.ID, system))

	response, err := gaso.llm.Generate(ctx, gradientPrompt, llm.GenerateOptions{
		Temperature: &[]float64{0.2}[0],
	})
	if err != nil {
		return SemanticGradient{}, fmt.Errorf("failed to generate component gradient: %v", err)
	}

	// Parse gradient from response
	gradient, err := parseComponentGradient(response.Text)
	if err != nil {
		// Fallback gradient
		return SemanticGradient{
			Component:  "content optimization",
			Direction:  "improve clarity and effectiveness",
			Magnitude:  0.7,
			Reasoning:  "Component needs optimization for system coherence",
			Confidence: 0.8,
		}, nil
	}
	
	return gradient, nil
}

// applySystemGradients applies gradients to optimize the entire system
func (gaso *GASOOptimizer) applySystemGradients(ctx context.Context, system *SystemDefinition, gradients []SystemGradient, config GASOConfig) (*SystemDefinition, []ComponentChange, error) {
	optimizedSystem := *system // Copy system
	changes := make([]ComponentChange, 0)
	
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
				optimizedContent, err := gaso.applyComponentGradient(ctx, component, sysGrad.Gradient)
				if err != nil {
					return nil, nil, fmt.Errorf("failed to apply gradient to component %s: %v", component.ID, err)
				}
				
				// Record change
				change := ComponentChange{
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
func (gaso *GASOOptimizer) applyComponentGradient(ctx context.Context, component SystemComponent, gradient SemanticGradient) (string, error) {
	optimizationPrompt := fmt.Sprintf(`Optimize this system component using the provided semantic gradient:

COMPONENT:
Type: %s
Current Content: %s

SEMANTIC GRADIENT:
Component: %s
Direction: %s
Reasoning: %s

Apply this gradient to optimize the component content.
Return only the optimized content without additional explanation.`, 
		component.Type, component.Content, gradient.Component, gradient.Direction, gradient.Reasoning)

	response, err := gaso.llm.Generate(ctx, optimizationPrompt, llm.GenerateOptions{
		Temperature: &[]float64{0.3}[0],
	})
	if err != nil {
		return "", fmt.Errorf("failed to apply component gradient: %v", err)
	}

	return response.Text, nil
}

// evaluateSystem evaluates the overall performance of the system
func (gaso *GASOOptimizer) evaluateSystem(ctx context.Context, system *SystemDefinition, objective string) (float64, error) {
	evaluationPrompt := fmt.Sprintf(`Evaluate the overall performance of this multi-component agentic system:

SYSTEM: %s
OBJECTIVE: %s

COMPONENTS:
%s

DEPENDENCIES:
%s

Rate the system's overall effectiveness on a scale of 0.0 to 1.0, considering:
- Component quality and coherence
- Inter-component synergy
- Alignment with objective
- System efficiency and robustness

Provide only a numeric score between 0.0 and 1.0.`, 
		system.Description, objective, gaso.formatComponents(system.Components), gaso.formatAllDependencies(system))

	response, err := gaso.llm.Generate(ctx, evaluationPrompt, llm.GenerateOptions{
		Temperature: &[]float64{0.0}[0],
	})
	if err != nil {
		return 0, fmt.Errorf("failed to evaluate system: %v", err)
	}

	// Parse score from response
	score, err := parseScore(response.Text)
	if err != nil {
		return 0.5, nil // Default fallback
	}
	
	return score, nil
}

// evaluateObjective evaluates a specific system objective
func (gaso *GASOOptimizer) evaluateObjective(ctx context.Context, system *SystemDefinition, objective SystemObjective) (float64, error) {
	evaluationPrompt := fmt.Sprintf(`Evaluate this multi-component system against a specific objective:

SYSTEM: %s

OBJECTIVE: %s
Type: %s (Target: %.2f)

COMPONENTS:
%s

Rate how well the system meets this objective on a scale of 0.0 to 1.0.
Provide only a numeric score.`, 
		system.Description, objective.Description, objective.Type, objective.Target,
		gaso.formatComponents(system.Components))

	response, err := gaso.llm.Generate(ctx, evaluationPrompt, llm.GenerateOptions{
		Temperature: &[]float64{0.0}[0],
	})
	if err != nil {
		return 0, fmt.Errorf("failed to evaluate objective %s: %v", objective.Name, err)
	}

	score, err := parseScore(response.Text)
	if err != nil {
		return 0.5, nil // Default fallback
	}
	
	return score, nil
}

// Helper functions
func (gaso *GASOOptimizer) calculateComponentPriority(component SystemComponent, system *SystemDefinition) float64 {
	// Calculate priority based on component dependencies and system structure
	priority := 0.5 // Base priority
	
	// Increase priority for components with many outgoing dependencies
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

func (gaso *GASOOptimizer) formatDependencies(componentID string, system *SystemDefinition) string {
	result := ""
	for _, dep := range system.Dependencies {
		if dep.From == componentID || dep.To == componentID {
			result += fmt.Sprintf("- %s -> %s (%s)\n", dep.From, dep.To, dep.Type)
		}
	}
	return result
}

func (gaso *GASOOptimizer) formatComponents(components []SystemComponent) string {
	result := ""
	for _, comp := range components {
		result += fmt.Sprintf("- %s (%s): %s\n", comp.Name, comp.Type, comp.Content)
	}
	return result
}

func (gaso *GASOOptimizer) formatAllDependencies(system *SystemDefinition) string {
	result := ""
	for _, dep := range system.Dependencies {
		result += fmt.Sprintf("- %s -> %s (%s, weight: %.2f)\n", dep.From, dep.To, dep.Type, dep.Weight)
	}
	return result
}

func (gaso *GASOOptimizer) analyzeConvergence(history []GASOIteration) *ConvergenceAnalysis {
	if len(history) < 2 {
		return &ConvergenceAnalysis{
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
	
	// Calculate convergence rate (how fast we're approaching optimum)
	convergenceRate := 0.0
	if len(history) > 1 {
		initialPerf := history[0].Performance
		finalPerf := history[len(history)-1].Performance
		if finalPerf > initialPerf {
			convergenceRate = (finalPerf - initialPerf) / float64(len(history))
		}
	}
	
	return &ConvergenceAnalysis{
		Converged:       converged,
		ConvergenceRate: convergenceRate,
		FinalGradient:   avgImprovement,
		StabilityMetric: stability,
	}
}

func (gaso *GASOOptimizer) checkParetoEfficiency(scores map[string]float64, objectives []SystemObjective) bool {
	// A solution is Pareto efficient if no other solution dominates it
	// For multi-objective optimization, check if all objectives meet their targets
	
	paretoEfficient := true
	
	for _, obj := range objectives {
		score, exists := scores[obj.Name]
		if !exists {
			continue
		}
		
		switch obj.Type {
		case "maximize":
			if score < obj.Target {
				paretoEfficient = false
			}
		case "minimize":
			if score > obj.Target {
				paretoEfficient = false
			}
		case "target":
			tolerance := 0.1 // 10% tolerance
			if math.Abs(score-obj.Target) > tolerance {
				paretoEfficient = false
			}
		}
	}
	
	return paretoEfficient
}

// parseComponentGradient parses a semantic gradient from LLM response
func parseComponentGradient(response string) (SemanticGradient, error) {
	// Try to extract JSON from response
	start := strings.Index(response, "{")
	end := strings.LastIndex(response, "}")
	if start == -1 || end == -1 || start > end {
		return SemanticGradient{}, fmt.Errorf("no JSON found in response")
	}
	
	jsonStr := response[start : end+1]
	
	var gradient SemanticGradient
	if err := json.Unmarshal([]byte(jsonStr), &gradient); err != nil {
		return SemanticGradient{}, fmt.Errorf("failed to parse gradient JSON: %v", err)
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
	
	return gradient, nil
}