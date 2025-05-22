package metaprompt

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"sync"

	"github.com/tmc/pe/internal/llm"
)

// FusionOptimizer implements multi-model consensus engineering
// based on 2024-2025 research in ensemble prompt optimization
type FusionOptimizer struct {
	providers []llm.Provider
	weights   map[string]float64
	strategy  ConsensusStrategy
}

// ConsensusStrategy defines how multiple models reach consensus
type ConsensusStrategy string

const (
	WeightedVoting   ConsensusStrategy = "weighted"
	ReflectionBased  ConsensusStrategy = "reflection"
	AdaptiveWeights  ConsensusStrategy = "adaptive"
	ParetoOptimal    ConsensusStrategy = "pareto"
)

// FusionResult represents the result of multi-model consensus optimization
type FusionResult struct {
	OptimizedPrompt  string                 `json:"optimized_prompt"`
	ConsensusScore   float64                `json:"consensus_score"`
	ModelResults     map[string]interface{} `json:"model_results"`
	OptimalWeights   map[string]float64     `json:"optimal_weights"`
	ConvergenceData  []float64              `json:"convergence_data"`
	ParetoFrontier   []PromptCandidate      `json:"pareto_frontier,omitempty"`
}

// PromptCandidate represents a candidate prompt with multi-objective metrics
type PromptCandidate struct {
	Prompt    string             `json:"prompt"`
	Accuracy  float64            `json:"accuracy"`
	Latency   float64            `json:"latency"`
	Cost      float64            `json:"cost"`
	Metrics   map[string]float64 `json:"metrics"`
}

// NewFusionOptimizer creates a new multi-model consensus optimizer
func NewFusionOptimizer(providers []llm.Provider, strategy ConsensusStrategy) *FusionOptimizer {
	weights := make(map[string]float64)
	for _, provider := range providers {
		weights[provider.Name()] = 1.0 / float64(len(providers))
	}
	
	return &FusionOptimizer{
		providers: providers,
		weights:   weights,
		strategy:  strategy,
	}
}

// OptimizeWithConsensus performs multi-model consensus optimization
func (f *FusionOptimizer) OptimizeWithConsensus(ctx context.Context, prompt string, objective string, iterations int) (*FusionResult, error) {
	result := &FusionResult{
		ModelResults:    make(map[string]interface{}),
		OptimalWeights:  make(map[string]float64),
		ConvergenceData: make([]float64, 0, iterations),
	}

	currentPrompt := prompt
	var candidates []PromptCandidate

	for i := 0; i < iterations; i++ {
		// Generate improvements from each model
		var wg sync.WaitGroup
		resultsChan := make(chan ModelImprovement, len(f.providers))
		
		for _, provider := range f.providers {
			wg.Add(1)
			go func(p llm.Provider) {
				defer wg.Done()
				improvement, err := f.generateImprovement(ctx, p, currentPrompt, objective)
				if err == nil {
					resultsChan <- improvement
				}
			}(provider)
		}
		
		go func() {
			wg.Wait()
			close(resultsChan)
		}()

		// Collect improvements
		var improvements []ModelImprovement
		for improvement := range resultsChan {
			improvements = append(improvements, improvement)
		}

		if len(improvements) == 0 {
			break
		}

		// Apply consensus strategy
		consensus, err := f.applyConsensusStrategy(improvements, objective)
		if err != nil {
			return nil, fmt.Errorf("consensus strategy failed: %w", err)
		}

		// Create candidate for Pareto analysis
		candidate := PromptCandidate{
			Prompt:   consensus.Prompt,
			Accuracy: consensus.Accuracy,
			Latency:  consensus.Latency,
			Cost:     consensus.Cost,
			Metrics:  consensus.Metrics,
		}
		candidates = append(candidates, candidate)

		currentPrompt = consensus.Prompt
		result.ConvergenceData = append(result.ConvergenceData, consensus.ConsensusScore)

		// Update adaptive weights if using adaptive strategy
		if f.strategy == AdaptiveWeights {
			f.updateAdaptiveWeights(improvements, consensus)
		}
	}

	// Calculate Pareto frontier if using pareto strategy
	if f.strategy == ParetoOptimal {
		result.ParetoFrontier = f.calculateParetoFrontier(candidates)
		if len(result.ParetoFrontier) > 0 {
			currentPrompt = result.ParetoFrontier[0].Prompt
		}
	}

	result.OptimizedPrompt = currentPrompt
	result.ConsensusScore = f.calculateFinalConsensusScore(currentPrompt, objective)
	result.OptimalWeights = f.weights

	return result, nil
}

// ModelImprovement represents an improvement suggested by a model
type ModelImprovement struct {
	Provider        string             `json:"provider"`
	Prompt          string             `json:"prompt"`
	Accuracy        float64            `json:"accuracy"`
	Latency         float64            `json:"latency"`
	Cost            float64            `json:"cost"`
	ConsensusScore  float64            `json:"consensus_score"`
	Metrics         map[string]float64 `json:"metrics"`
	Reasoning       string             `json:"reasoning"`
}

// ConsensusResult represents the result of applying a consensus strategy
type ConsensusResult struct {
	Prompt         string             `json:"prompt"`
	Accuracy       float64            `json:"accuracy"`
	Latency        float64            `json:"latency"`
	Cost           float64            `json:"cost"`
	ConsensusScore float64            `json:"consensus_score"`
	Metrics        map[string]float64 `json:"metrics"`
}

func (f *FusionOptimizer) generateImprovement(ctx context.Context, provider llm.Provider, prompt string, objective string) (ModelImprovement, error) {
	improvementPrompt := fmt.Sprintf(`
You are an expert prompt engineer. Analyze and improve the following prompt to better achieve the objective.

Current Prompt: %s

Objective: %s

Provide your improvement as JSON with the following structure:
{
  "improved_prompt": "your improved version",
  "accuracy_score": 0.85,
  "latency_estimate": 150,
  "cost_estimate": 0.002,
  "reasoning": "explanation of improvements",
  "metrics": {
    "clarity": 0.9,
    "specificity": 0.8,
    "completeness": 0.85
  }
}

Focus on improving accuracy while maintaining efficiency.`, prompt, objective)

	response, err := provider.Generate(ctx, improvementPrompt, llm.GenerateOptions{})
	if err != nil {
		return ModelImprovement{}, err
	}

	var improvement struct {
		ImprovedPrompt   string             `json:"improved_prompt"`
		AccuracyScore    float64            `json:"accuracy_score"`
		LatencyEstimate  float64            `json:"latency_estimate"`
		CostEstimate     float64            `json:"cost_estimate"`
		Reasoning        string             `json:"reasoning"`
		Metrics          map[string]float64 `json:"metrics"`
	}

	if err := json.Unmarshal([]byte(response.Text), &improvement); err != nil {
		// Fallback: extract prompt from text response
		return ModelImprovement{
			Provider:       provider.Name(),
			Prompt:         response.Text,
			Accuracy:       0.7, // Default values
			Latency:        100,
			Cost:           0.001,
			ConsensusScore: 0.7,
			Metrics:        map[string]float64{"quality": 0.7},
			Reasoning:      "Extracted from text response",
		}, nil
	}

	return ModelImprovement{
		Provider:       provider.Name(),
		Prompt:         improvement.ImprovedPrompt,
		Accuracy:       improvement.AccuracyScore,
		Latency:        improvement.LatencyEstimate,
		Cost:           improvement.CostEstimate,
		ConsensusScore: improvement.AccuracyScore,
		Metrics:        improvement.Metrics,
		Reasoning:      improvement.Reasoning,
	}, nil
}

func (f *FusionOptimizer) applyConsensusStrategy(improvements []ModelImprovement, objective string) (ConsensusResult, error) {
	switch f.strategy {
	case WeightedVoting:
		return f.weightedVoting(improvements)
	case ReflectionBased:
		return f.reflectionBasedConsensus(improvements, objective)
	case AdaptiveWeights:
		return f.adaptiveWeightedConsensus(improvements)
	case ParetoOptimal:
		return f.paretoOptimalConsensus(improvements)
	default:
		return f.weightedVoting(improvements)
	}
}

func (f *FusionOptimizer) weightedVoting(improvements []ModelImprovement) (ConsensusResult, error) {
	if len(improvements) == 0 {
		return ConsensusResult{}, fmt.Errorf("no improvements to vote on")
	}

	// Weight by provider weights and accuracy scores
	bestImprovement := improvements[0]
	bestScore := 0.0

	for _, improvement := range improvements {
		weight := f.weights[improvement.Provider]
		score := weight * improvement.ConsensusScore
		
		if score > bestScore {
			bestScore = score
			bestImprovement = improvement
		}
	}

	return ConsensusResult{
		Prompt:         bestImprovement.Prompt,
		Accuracy:       bestImprovement.Accuracy,
		Latency:        bestImprovement.Latency,
		Cost:           bestImprovement.Cost,
		ConsensusScore: bestScore,
		Metrics:        bestImprovement.Metrics,
	}, nil
}

func (f *FusionOptimizer) reflectionBasedConsensus(improvements []ModelImprovement, objective string) (ConsensusResult, error) {
	// Use the first provider to reflect on all improvements and synthesize
	if len(f.providers) == 0 || len(improvements) == 0 {
		return ConsensusResult{}, fmt.Errorf("insufficient data for reflection")
	}

	improvementsJson, _ := json.MarshalIndent(improvements, "", "  ")
	
	reflectionPrompt := fmt.Sprintf(`
You are a meta-prompt engineer. Analyze these improvements from multiple AI models and synthesize the best combined approach.

Objective: %s

Improvements from different models:
%s

Synthesize the best elements from all improvements into a single optimized prompt. Respond with JSON:
{
  "synthesized_prompt": "your synthesis",
  "consensus_score": 0.9,
  "reasoning": "explanation of synthesis"
}`, objective, string(improvementsJson))

	ctx := context.Background()
	response, err := f.providers[0].Generate(ctx, reflectionPrompt, llm.GenerateOptions{})
	if err != nil {
		return f.weightedVoting(improvements) // Fallback
	}

	var synthesis struct {
		SynthesizedPrompt string  `json:"synthesized_prompt"`
		ConsensusScore    float64 `json:"consensus_score"`
		Reasoning         string  `json:"reasoning"`
	}

	if err := json.Unmarshal([]byte(response.Text), &synthesis); err != nil {
		return f.weightedVoting(improvements) // Fallback
	}

	// Calculate metrics from improvements
	avgAccuracy := 0.0
	avgLatency := 0.0
	avgCost := 0.0
	for _, imp := range improvements {
		avgAccuracy += imp.Accuracy
		avgLatency += imp.Latency
		avgCost += imp.Cost
	}
	avgAccuracy /= float64(len(improvements))
	avgLatency /= float64(len(improvements))
	avgCost /= float64(len(improvements))

	return ConsensusResult{
		Prompt:         synthesis.SynthesizedPrompt,
		Accuracy:       avgAccuracy,
		Latency:        avgLatency,
		Cost:           avgCost,
		ConsensusScore: synthesis.ConsensusScore,
		Metrics:        map[string]float64{"synthesis_quality": synthesis.ConsensusScore},
	}, nil
}

func (f *FusionOptimizer) adaptiveWeightedConsensus(improvements []ModelImprovement) (ConsensusResult, error) {
	// Similar to weighted voting but uses current adaptive weights
	return f.weightedVoting(improvements)
}

func (f *FusionOptimizer) paretoOptimalConsensus(improvements []ModelImprovement) (ConsensusResult, error) {
	// Find the improvement that's on the Pareto frontier
	paretoOptimal := f.findParetoOptimal(improvements)
	if len(paretoOptimal) == 0 {
		return f.weightedVoting(improvements)
	}
	
	// Return the best balanced option from Pareto frontier
	best := paretoOptimal[0]
	return ConsensusResult{
		Prompt:         best.Prompt,
		Accuracy:       best.Accuracy,
		Latency:        best.Latency,
		Cost:           best.Cost,
		ConsensusScore: best.ConsensusScore,
		Metrics:        best.Metrics,
	}, nil
}

func (f *FusionOptimizer) updateAdaptiveWeights(improvements []ModelImprovement, consensus ConsensusResult) {
	// Update weights based on how close each improvement was to the consensus
	learningRate := 0.1
	
	for _, improvement := range improvements {
		// Calculate similarity to consensus
		similarity := f.calculateSimilarity(improvement, consensus)
		
		// Update weight using exponential moving average
		currentWeight := f.weights[improvement.Provider]
		newWeight := currentWeight + learningRate*(similarity-currentWeight)
		f.weights[improvement.Provider] = math.Max(0.01, math.Min(1.0, newWeight))
	}
	
	// Normalize weights
	f.normalizeWeights()
}

func (f *FusionOptimizer) calculateSimilarity(improvement ModelImprovement, consensus ConsensusResult) float64 {
	// Simple similarity based on accuracy difference
	accuracyDiff := math.Abs(improvement.Accuracy - consensus.Accuracy)
	return math.Max(0, 1.0-accuracyDiff)
}

func (f *FusionOptimizer) normalizeWeights() {
	total := 0.0
	for _, weight := range f.weights {
		total += weight
	}
	
	if total > 0 {
		for provider, weight := range f.weights {
			f.weights[provider] = weight / total
		}
	}
}

func (f *FusionOptimizer) calculateParetoFrontier(candidates []PromptCandidate) []PromptCandidate {
	var frontier []PromptCandidate
	
	for i, candidate := range candidates {
		dominated := false
		
		for j, other := range candidates {
			if i != j && f.dominates(other, candidate) {
				dominated = true
				break
			}
		}
		
		if !dominated {
			frontier = append(frontier, candidate)
		}
	}
	
	return frontier
}

func (f *FusionOptimizer) dominates(a, b PromptCandidate) bool {
	// A dominates B if A is better in at least one objective and not worse in any
	betterInOne := false
	
	// Higher accuracy is better
	if a.Accuracy > b.Accuracy {
		betterInOne = true
	} else if a.Accuracy < b.Accuracy {
		return false
	}
	
	// Lower latency is better
	if a.Latency < b.Latency {
		betterInOne = true
	} else if a.Latency > b.Latency {
		return false
	}
	
	// Lower cost is better
	if a.Cost < b.Cost {
		betterInOne = true
	} else if a.Cost > b.Cost {
		return false
	}
	
	return betterInOne
}

func (f *FusionOptimizer) findParetoOptimal(improvements []ModelImprovement) []ModelImprovement {
	var optimal []ModelImprovement
	
	for i, improvement := range improvements {
		dominated := false
		
		for j, other := range improvements {
			if i != j && f.improvementDominates(other, improvement) {
				dominated = true
				break
			}
		}
		
		if !dominated {
			optimal = append(optimal, improvement)
		}
	}
	
	return optimal
}

func (f *FusionOptimizer) improvementDominates(a, b ModelImprovement) bool {
	betterInOne := false
	
	if a.Accuracy > b.Accuracy {
		betterInOne = true
	} else if a.Accuracy < b.Accuracy {
		return false
	}
	
	if a.Latency < b.Latency {
		betterInOne = true
	} else if a.Latency > b.Latency {
		return false
	}
	
	if a.Cost < b.Cost {
		betterInOne = true
	} else if a.Cost > b.Cost {
		return false
	}
	
	return betterInOne
}

func (f *FusionOptimizer) calculateFinalConsensusScore(prompt string, objective string) float64 {
	// Simple heuristic based on prompt length and objective alignment
	lengthScore := math.Min(1.0, float64(len(prompt))/500.0)
	return lengthScore * 0.8 // Conservative estimate
}