package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/tmc/pe/internal/llm"
	"github.com/tmc/pe/internal/metaprompt"
	"github.com/tmc/pe/internal/providers"
)

// fusionCmd represents the fusion command
var fusionCmd = &cobra.Command{
	Use:   "fusion [prompt_file]",
	Short: "Multi-model consensus optimization using fusion",
	Long: `Fusion optimizes prompts using multi-model consensus engineering.

This command implements advanced consensus strategies across multiple LLM providers:
- Weighted voting based on model performance
- Reflection-based consensus through iterative refinement
- Adaptive weighting using reinforcement learning
- Pareto optimal solutions for multi-objective optimization

The fusion process:
1. Generates candidates from multiple models
2. Applies consensus strategy to combine results
3. Validates consensus through cross-model evaluation
4. Iteratively refines based on performance metrics

Example usage:
  pe fusion prompt.txt --providers openai:gpt-4,anthropic:claude-3 --strategy weighted
  pe fusion prompt.txt --iterations 5 --consensus-threshold 0.8
  pe fusion prompt.txt --multi-objective --pareto-optimal`,
	RunE: runFusion,
}

var (
	fusionProviders       []string
	fusionStrategy        string
	fusionIterations      int
	fusionConsensusThresh float64
	fusionMultiObjective  bool
	fusionParetoOptimal   bool
	fusionOutputFile      string
	fusionVerbose         bool
	fusionAdaptiveWeights bool
	fusionCrossValidation bool
)

func init() {
	fusionCmd.Flags().StringSliceVar(&fusionProviders, "providers", []string{}, "Comma-separated list of providers (e.g., openai:gpt-4,anthropic:claude-3)")
	fusionCmd.Flags().StringVar(&fusionStrategy, "strategy", "weighted", "Consensus strategy: weighted, reflection, adaptive, pareto")
	fusionCmd.Flags().IntVar(&fusionIterations, "iterations", 3, "Number of optimization iterations")
	fusionCmd.Flags().Float64Var(&fusionConsensusThresh, "consensus-threshold", 0.75, "Minimum consensus threshold for accepting results")
	fusionCmd.Flags().BoolVar(&fusionMultiObjective, "multi-objective", false, "Enable multi-objective optimization")
	fusionCmd.Flags().BoolVar(&fusionParetoOptimal, "pareto-optimal", false, "Find Pareto optimal solutions")
	fusionCmd.Flags().StringVarP(&fusionOutputFile, "output", "o", "", "Output file for optimized prompt")
	fusionCmd.Flags().BoolVarP(&fusionVerbose, "verbose", "v", false, "Enable verbose output")
	fusionCmd.Flags().BoolVar(&fusionAdaptiveWeights, "adaptive-weights", false, "Enable adaptive weight learning")
	fusionCmd.Flags().BoolVar(&fusionCrossValidation, "cross-validation", false, "Enable cross-validation between models")
}

func runFusion(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("prompt file required")
	}

	promptFile := args[0]

	// Read prompt content
	promptContent, err := os.ReadFile(promptFile)
	if err != nil {
		return fmt.Errorf("failed to read prompt file: %v", err)
	}

	// Validate providers
	if len(fusionProviders) < 2 {
		return fmt.Errorf("fusion requires at least 2 providers, got %d", len(fusionProviders))
	}

	// Create LLM providers
	llmProviders := make([]llm.Provider, 0, len(fusionProviders))
	for _, providerSpec := range fusionProviders {
		provider, err := providers.CreateProvider(providerSpec, map[string]interface{}{})
		if err != nil {
			return fmt.Errorf("failed to create provider %s: %v", providerSpec, err)
		}
		llmProviders = append(llmProviders, provider)
	}

	// Create fusion optimizer with strategy
	strategy := getConsensusStrategy(fusionStrategy)
	fusionOptimizer := metaprompt.NewFusionOptimizer(llmProviders, strategy)

	// Configure fusion parameters
	objective := "Optimize prompt for accuracy, clarity, and efficiency"

	// Run fusion optimization
	ctx := context.Background()
	startTime := time.Now()

	if fusionVerbose {
		fmt.Printf("Starting fusion optimization with %d providers...\n", len(llmProviders))
		fmt.Printf("Strategy: %s\n", fusionStrategy)
		fmt.Printf("Iterations: %d\n", fusionIterations)
		fmt.Printf("Consensus threshold: %.2f\n", fusionConsensusThresh)
	}

	result, err := fusionOptimizer.OptimizeWithConsensus(ctx, string(promptContent), objective, fusionIterations)
	if err != nil {
		return fmt.Errorf("fusion optimization failed: %v", err)
	}

	elapsed := time.Since(startTime)

	// Display results
	fmt.Printf("\n=== Fusion Optimization Complete ===\n")
	fmt.Printf("Duration: %s\n", elapsed)
	fmt.Printf("Iterations: %d\n", len(result.ConvergenceData))
	fmt.Printf("Final consensus score: %.3f\n", result.ConsensusScore)
	// Overall improvement calculation
	if len(result.ConvergenceData) > 0 {
		improvement := result.ConvergenceData[len(result.ConvergenceData)-1] - result.ConvergenceData[0]
		fmt.Printf("Overall improvement: %.2f%%\n", improvement*100)
	}

	if fusionVerbose {
		fmt.Printf("\nModel contributions:\n")
		for model, weight := range result.OptimalWeights {
			fmt.Printf("  %s: %.3f\n", model, weight)
		}

		fmt.Printf("\nConsensus metrics:\n")
		for metric, value := range result.ModelResults {
			fmt.Printf("  %s: %v\n", metric, value)
		}

		if len(result.ConvergenceData) > 0 {
			fmt.Printf("\nOptimization history:\n")
			for i, score := range result.ConvergenceData {
				fmt.Printf("  Iteration %d: consensus=%.3f\n", i+1, score)
			}
		}
	}

	// Display optimized prompt
	fmt.Printf("\n=== Optimized Prompt ===\n%s\n", result.OptimizedPrompt)

	// Save to file if requested
	if fusionOutputFile != "" {
		// Create output structure
		output := map[string]interface{}{
			"original_prompt":     string(promptContent),
			"optimized_prompt":    result.OptimizedPrompt,
			"consensus_score":     result.ConsensusScore,
			"overall_improvement": 0.0, // Calculate from convergence data
			"model_weights":       result.OptimalWeights,
			"model_results":       result.ModelResults,
			"optimization_time":   elapsed.String(),
			"config": map[string]interface{}{
				"providers":  fusionProviders,
				"strategy":   fusionStrategy,
				"iterations": fusionIterations,
			},
		}

		outputJSON, err := json.MarshalIndent(output, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal output: %v", err)
		}

		if err := os.WriteFile(fusionOutputFile, outputJSON, 0644); err != nil {
			return fmt.Errorf("failed to write output file: %v", err)
		}

		fmt.Printf("\nResults saved to: %s\n", fusionOutputFile)
	}

	// Display Pareto frontier if multi-objective optimization was used
	if fusionMultiObjective && len(result.ParetoFrontier) > 0 {
		fmt.Printf("\n=== Pareto Frontier ===\n")
		fmt.Printf("Found %d Pareto-optimal solutions\n", len(result.ParetoFrontier))
		for i, candidate := range result.ParetoFrontier {
			if i >= 3 && !fusionVerbose {
				fmt.Printf("  ... and %d more\n", len(result.ParetoFrontier)-3)
				break
			}
			fmt.Printf("  %d. Accuracy: %.3f, Latency: %.1fms, Cost: $%.4f\n",
				i+1, candidate.Accuracy, candidate.Latency, candidate.Cost)
		}
	}

	return nil
}

func getConsensusStrategy(strategy string) metaprompt.ConsensusStrategy {
	switch strategy {
	case "weighted":
		return metaprompt.WeightedVoting
	case "reflection":
		return metaprompt.ReflectionBased
	case "adaptive":
		return metaprompt.AdaptiveWeights
	case "pareto":
		return metaprompt.ParetoOptimal
	default:
		return metaprompt.WeightedVoting
	}
}
