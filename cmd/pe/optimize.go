package main

import (
	"context"
	"fmt"
	"os"

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
		method        string
	)

	cmd := &cobra.Command{
		Use:   "optimize [prompt-file]",
		Short: "Optimize prompts using metaprompting techniques",
		Long: `Optimize prompts using advanced metaprompting techniques including:
- PE2: Prompt Engineering a Prompt Engineer (2024 breakthrough)
- APEX: Automated Prompt Engineering Xpert for long prompts
- TextGrad: Natural language gradients optimization
- Standard: Enhanced iterative refinement with LLM feedback
- Hybrid: Combining multiple state-of-the-art methods

The optimize command implements cutting-edge 2024-2025 research in
automated prompt engineering and systematic optimization.`,
		Example: `  # PE2: Meta-prompt optimization with reasoning templates
  pe optimize --prompt "Analyze sentiment" --method pe2 --iterations 5

  # APEX: Long prompt optimization with beam search
  pe optimize --prompt-file complex-system.txt --method apex --iterations 8

  # TextGrad: Natural language gradients
  pe optimize --prompt "Classify text" --method textgrad --iterations 6

  # Hybrid: Best of multiple methods
  pe optimize --prompt "Generate code" --method hybrid --iterations 6

  # Standard with specific provider
  pe optimize --prompt "Summarize" --provider anthropic --output results.json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Handle file argument
			if len(args) > 0 {
				content, err := os.ReadFile(args[0])
				if err != nil {
					return fmt.Errorf("failed to read prompt file: %w", err)
				}
				initialPrompt = string(content)
			}

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
				Method:        method,
			}

			// Create optimizer
			optimizer := metaprompt.NewOptimizer(llmProvider)

			fmt.Printf("Optimizing prompt with %d iterations using %s method...\n", iterations, method)

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

			// Always save the optimized prompt to a default file based on method
			defaultFilename := fmt.Sprintf("prompt_%s.txt", method)

			// Save just the optimized prompt text
			if err := os.WriteFile(defaultFilename, []byte(result.OptimizedPrompt), 0644); err != nil {
				return fmt.Errorf("failed to save optimized prompt: %w", err)
			}
			fmt.Printf("\nOptimized prompt saved to: %s\n", defaultFilename)

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
	cmd.Flags().StringVarP(&method, "method", "m", "standard", "Optimization method (standard, pe2, apex, textgrad, hybrid)")

	// Remove required flag since we can also accept file argument

	return cmd
}
