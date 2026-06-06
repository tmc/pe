package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tmc/pe/internal/inference"
	"github.com/tmc/pe/internal/metaprompt"
)

func evolveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "evolve [prompt-file]",
		Short: "Optimize prompts using evolutionary algorithms (NSGA-II)",
		Long: `Perform evolutionary optimization on prompts using genetic algorithms.

This command supports multi-objective optimization with NSGA-II
(Non-dominated Sorting Genetic Algorithm II) for finding Pareto-optimal prompt
variants.

Features:
- Population-based prompt improvement
- Multi-objective optimization (e.g., accuracy vs. length)
- Automatic mutation and crossover operations
- Convergence detection and early stopping
- Pareto frontier extraction for trade-off analysis

Examples:
  # Basic evolutionary optimization
  pe evolve prompt.txt

  # Optimize with specific objectives
  pe evolve prompt.txt --objectives="accuracy,conciseness,clarity"

  # Run for more generations with larger population
  pe evolve prompt.txt --generations=50 --population=30

  # Save evolution history and analysis
  pe evolve prompt.txt --output=evolution-results.json

  # Use specific provider for evaluation
  pe evolve prompt.txt --provider="openai:gpt-4-turbo"`,
		RunE: runEvolve,
	}

	cmd.Flags().StringVarP(&evolveProvider, "provider", "p", "openai:gpt-4-turbo", "LLM provider for evaluation")
	cmd.Flags().IntVarP(&evolveGenerations, "generations", "g", 20, "Number of evolutionary generations")
	cmd.Flags().IntVarP(&evolvePopulation, "population", "n", 10, "Population size")
	cmd.Flags().Float64VarP(&evolveMutationRate, "mutation-rate", "m", 0.3, "Mutation probability (0.0-1.0)")
	cmd.Flags().Float64VarP(&evolveCrossoverRate, "crossover-rate", "c", 0.7, "Crossover probability (0.0-1.0)")
	cmd.Flags().Float64VarP(&evolveElitismRate, "elitism-rate", "e", 0.2, "Elitism rate (0.0-1.0)")
	cmd.Flags().StringVar(&evolveObjectives, "objectives", "effectiveness,clarity,conciseness", "Comma-separated optimization objectives")
	cmd.Flags().StringVarP(&evolveOutput, "output", "o", "", "Output file for results (JSON format)")
	cmd.Flags().BoolVarP(&evolveVerbose, "verbose", "v", false, "Show detailed evolution progress")

	return cmd
}

var (
	evolveProvider      string
	evolveGenerations   int
	evolvePopulation    int
	evolveMutationRate  float64
	evolveCrossoverRate float64
	evolveElitismRate   float64
	evolveObjectives    string
	evolveOutput        string
	evolveVerbose       bool
)

func runEvolve(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("prompt file required")
	}

	// Read base prompt
	promptFile := args[0]
	promptBytes, err := os.ReadFile(promptFile)
	if err != nil {
		return fmt.Errorf("failed to read prompt file: %w", err)
	}
	basePrompt := string(promptBytes)

	// Create provider through the inference migration path, adapting only at
	// the evolutionary optimizer boundary.
	inferenceProvider, err := inference.CreateProviderFromSpec(evolveProvider, nil)
	if err != nil {
		return fmt.Errorf("failed to create provider: %w", err)
	}
	provider, err := inference.AsLegacyProvider(inferenceProvider, "")
	if err != nil {
		return fmt.Errorf("failed to adapt provider: %w", err)
	}

	// Parse objectives
	objectives := strings.Split(evolveObjectives, ",")
	for i := range objectives {
		objectives[i] = strings.TrimSpace(objectives[i])
	}

	// Create evolution config
	config := metaprompt.EvolutionConfig{
		PopulationSize: evolvePopulation,
		Generations:    evolveGenerations,
		MutationRate:   evolveMutationRate,
		CrossoverRate:  evolveCrossoverRate,
		ElitismRate:    evolveElitismRate,
		Objectives:     objectives,
	}

	// Create optimizer
	optimizer := metaprompt.NewEvolutionaryOptimizer(provider, config)

	// Run evolution
	ctx := context.Background()

	if evolveVerbose {
		fmt.Printf("Starting evolutionary optimization...\n")
		fmt.Printf("Population size: %d\n", evolvePopulation)
		fmt.Printf("Generations: %d\n", evolveGenerations)
		fmt.Printf("Objectives: %v\n", objectives)
		fmt.Printf("Base prompt: %s\n\n", truncateString(basePrompt, 100))
	}

	result, err := optimizer.Evolve(ctx, basePrompt, config)
	if err != nil {
		return fmt.Errorf("evolution failed: %w", err)
	}

	// Display results
	fmt.Printf("\n=== Evolution Results ===\n")
	fmt.Printf("Best fitness: %.4f\n", result.BestIndividual.Fitness)
	fmt.Printf("Generations run: %d\n", len(result.EvolutionHistory))
	fmt.Printf("Final population diversity: %.4f\n", result.FinalPopulation.Diversity)

	fmt.Printf("\n=== Best Prompt ===\n%s\n", result.BestIndividual.Prompt)

	fmt.Printf("\n=== Objective Scores ===\n")
	for obj, score := range result.BestIndividual.Objectives {
		fmt.Printf("  %s: %.4f\n", obj, score)
	}

	// Show evolution history if verbose
	if evolveVerbose && len(result.BestIndividual.Mutations) > 0 {
		fmt.Printf("\n=== Mutation History ===\n")
		for _, mutation := range result.BestIndividual.Mutations {
			fmt.Printf("  Gen %d: %s - %s\n", mutation.Generation, mutation.Type, mutation.Description)
		}
	}

	// Show Pareto frontier
	if len(result.ParetoFrontier) > 1 {
		fmt.Printf("\n=== Pareto Frontier ===\n")
		fmt.Printf("Found %d non-dominated solutions:\n", len(result.ParetoFrontier))
		for i, individual := range result.ParetoFrontier {
			if i >= 5 && !evolveVerbose {
				fmt.Printf("  ... and %d more\n", len(result.ParetoFrontier)-5)
				break
			}
			fmt.Printf("  %d. Fitness: %.4f", i+1, individual.Fitness)
			for obj, score := range individual.Objectives {
				fmt.Printf(", %s: %.4f", obj, score)
			}
			fmt.Printf("\n")
		}
	}

	// Show convergence data
	if evolveVerbose && len(result.ConvergenceData) > 0 {
		fmt.Printf("\n=== Convergence History ===\n")
		for i, fitness := range result.ConvergenceData {
			if i%5 == 0 || i == len(result.ConvergenceData)-1 {
				fmt.Printf("  Generation %d: %.4f\n", i, fitness)
			}
		}
	}

	// Save results if output file specified
	if evolveOutput != "" {
		outputData, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal results: %w", err)
		}

		if err := os.WriteFile(evolveOutput, outputData, 0644); err != nil {
			return fmt.Errorf("failed to write output file: %w", err)
		}

		fmt.Printf("\nResults saved to: %s\n", evolveOutput)
	}

	return nil
}
