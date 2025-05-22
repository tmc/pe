package main

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tmc/pe/internal/llm"
	"github.com/tmc/pe/internal/metaprompt"
)

// optimizeCmd returns a cobra.Command for prompt optimization using metaprompting techniques
func optimizeCmd() *cobra.Command {
	var (
		initialPrompt string
		iterations    int
		provider      string
		model         string
		outputFile    string
		temperature   float64
		maxTokens     int
	)

	cmd := &cobra.Command{
		Use:   "optimize",
		Short: "Optimize prompts using metaprompting techniques",
		Long: `Optimize prompts using advanced metaprompting techniques including:
- Iterative refinement with LLM feedback
- DSPy-style structured prompt generation  
- Reflection and critique mechanisms
- Automated prompt improvement

The optimize command uses a meta-LLM to analyze and improve your prompts,
applying the latest research in prompt optimization and engineering.`,
		Example: `  # Optimize a basic prompt
  pe optimize --prompt "Summarize this text" --iterations 3

  # Use specific provider and model
  pe optimize --prompt "Classify sentiment" --provider anthropic --model claude-3-sonnet

  # Save optimization results
  pe optimize --prompt "Generate code" --output optimized.json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if initialPrompt == "" {
				return fmt.Errorf("initial prompt is required")
			}

			// Create LLM provider
			llmProvider, err := llm.GetProvider(provider)
			if err != nil {
				return fmt.Errorf("failed to create provider: %w", err)
			}

			// Configure optimization
			cfg := metaprompt.Config{
				InitialPrompt: initialPrompt,
				Iterations:    iterations,
				Temperature:   temperature,
				MaxTokens:     maxTokens,
			}

			// Create optimizer
			optimizer := metaprompt.NewOptimizer(llmProvider)

			fmt.Printf("Optimizing prompt with %d iterations...\n", iterations)
			
			// Run optimization
			result, err := optimizer.Optimize(context.Background(), cfg)
			if err != nil {
				return fmt.Errorf("optimization failed: %w", err)
			}

			// Display results
			fmt.Printf("\n=== Optimization Results ===\n")
			fmt.Printf("Original Prompt:\n%s\n\n", result.OriginalPrompt)
			fmt.Printf("Optimized Prompt:\n%s\n\n", result.OptimizedPrompt)
			fmt.Printf("Improvement Score: %.2f\n", result.ImprovementScore)
			
			// Show iteration details
			fmt.Printf("\n=== Iteration Details ===\n")
			for i, iter := range result.Iterations {
				fmt.Printf("Iteration %d Score: %.2f\n", i+1, iter.Score)
				if len(iter.Feedback) > 0 {
					fmt.Printf("  Feedback: %s\n", iter.Feedback)
				}
			}

			// Save to file if requested
			if outputFile != "" {
				if err := result.SaveToFile(outputFile); err != nil {
					return fmt.Errorf("failed to save results: %w", err)
				}
				fmt.Printf("\nResults saved to: %s\n", outputFile)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&initialPrompt, "prompt", "p", "", "Initial prompt to optimize (required)")
	cmd.Flags().IntVarP(&iterations, "iterations", "i", 3, "Number of optimization iterations")
	cmd.Flags().StringVar(&provider, "provider", "openai", "LLM provider (openai, anthropic, etc.)")
	cmd.Flags().StringVar(&model, "model", "", "Specific model to use")
	cmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file for optimization results (JSON)")
	cmd.Flags().Float64Var(&temperature, "temperature", 0.3, "Temperature for generation")
	cmd.Flags().IntVar(&maxTokens, "max-tokens", 1000, "Maximum tokens per generation")

	cmd.MarkFlagRequired("prompt")

	return cmd
}