package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/tmc/pe/internal/metaprompt"
	"github.com/tmc/pe/internal/providers"
)

// semanticCmd implements semantic backpropagation and gradient descent for GASO
func semanticCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "semantic",
		Short: "Semantic backpropagation and gradient descent for Graph-based Agentic System Optimization (GASO)",
		Long: `Implements cutting-edge semantic backpropagation and gradient descent techniques
based on 2025 research from KAUST and IDSIA for optimizing language-based agentic systems.

Semantic gradients generalize mathematical gradients by representing directional 
information in semantically interoperable forms, enabling optimization of complex
AI systems through natural language feedback.`,
	}

	cmd.AddCommand(semanticBackpropCmd())
	cmd.AddCommand(semanticDescentCmd())
	cmd.AddCommand(gasoCmd())

	return cmd
}

// semanticBackpropCmd implements semantic backpropagation
func semanticBackpropCmd() *cobra.Command {
	var (
		prompt      string
		promptFile  string
		target      string
		iterations  int
		provider    string
		model       string
		outputFile  string
		format      string
		verbose     bool
	)

	cmd := &cobra.Command{
		Use:   "backprop",
		Short: "Apply semantic backpropagation to optimize prompts",
		Long: `Semantic backpropagation generalizes reverse-mode automatic differentiation
by introducing semantic gradients that provide directional information for
improving system performance through natural language feedback.

This implementation follows the KAUST/IDSIA 2025 research on semantic
optimization for language-based agentic systems.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			// Load prompt content
			promptContent, err := loadPromptContent(prompt, promptFile)
			if err != nil {
				return fmt.Errorf("failed to load prompt: %v", err)
			}

			// Create LLM provider
			llmProvider, err := providers.CreateProvider(provider, map[string]interface{}{
				"model": model,
			})
			if err != nil {
				return fmt.Errorf("failed to create provider: %v", err)
			}

			// Initialize semantic optimizer
			optimizer := metaprompt.NewSemanticOptimizer(llmProvider)

			// Configure semantic backpropagation
			config := metaprompt.SemanticConfig{
				Target:      target,
				Iterations:  iterations,
				Verbose:     verbose,
				Provider:    provider,
				Model:       model,
			}

			fmt.Printf("Starting semantic backpropagation...\n")
			fmt.Printf("Target objective: %s\n", target)
			fmt.Printf("Iterations: %d\n", iterations)
			fmt.Printf("Provider: %s (%s)\n\n", provider, model)

			// Run semantic backpropagation
			result, err := optimizer.SemanticBackpropagation(ctx, promptContent, config)
			if err != nil {
				return fmt.Errorf("semantic backpropagation failed: %v", err)
			}

			// Output results
			return outputSemanticResult(result, outputFile, format)
		},
	}

	cmd.Flags().StringVarP(&prompt, "prompt", "p", "", "Prompt to optimize")
	cmd.Flags().StringVar(&promptFile, "prompt-file", "", "File containing prompt to optimize")
	cmd.Flags().StringVarP(&target, "target", "t", "", "Target objective for optimization")
	cmd.Flags().IntVarP(&iterations, "iterations", "i", 5, "Number of backpropagation iterations")
	cmd.Flags().StringVar(&provider, "provider", "openai", "LLM provider")
	cmd.Flags().StringVar(&model, "model", "gpt-4", "Model for semantic evaluation")
	cmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file for results")
	cmd.Flags().StringVarP(&format, "format", "f", "json", "Output format (json, yaml, table)")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")

	cmd.MarkFlagRequired("target")

	return cmd
}

// semanticDescentCmd implements semantic gradient descent
func semanticDescentCmd() *cobra.Command {
	var (
		prompt       string
		promptFile   string
		objective    string
		learningRate float64
		iterations   int
		provider     string
		model        string
		outputFile   string
		format       string
		convergence  float64
		adaptive     bool
	)

	cmd := &cobra.Command{
		Use:   "descent",
		Short: "Apply semantic gradient descent optimization",
		Long: `Semantic gradient descent enables optimization of language-based systems
using semantic gradients that capture directional improvement information
in natural language form.

This method extends traditional gradient descent to semantic domains,
allowing for effective optimization of complex AI system parameters.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			// Load prompt content
			promptContent, err := loadPromptContent(prompt, promptFile)
			if err != nil {
				return fmt.Errorf("failed to load prompt: %v", err)
			}

			// Create LLM provider
			llmProvider, err := providers.CreateProvider(provider, map[string]interface{}{
				"model": model,
			})
			if err != nil {
				return fmt.Errorf("failed to create provider: %v", err)
			}

			// Initialize semantic optimizer
			optimizer := metaprompt.NewSemanticOptimizer(llmProvider)

			// Configure semantic gradient descent
			config := metaprompt.SemanticDescentConfig{
				Objective:        objective,
				LearningRate:     learningRate,
				Iterations:       iterations,
				ConvergenceThreshold: convergence,
				AdaptiveLearning: adaptive,
				Provider:         provider,
				Model:           model,
			}

			fmt.Printf("Starting semantic gradient descent...\n")
			fmt.Printf("Objective: %s\n", objective)
			fmt.Printf("Learning rate: %.4f\n", learningRate)
			fmt.Printf("Iterations: %d\n", iterations)
			fmt.Printf("Convergence threshold: %.6f\n", convergence)
			if adaptive {
				fmt.Printf("Adaptive learning: ENABLED\n")
			}
			fmt.Printf("Provider: %s (%s)\n\n", provider, model)

			// Run semantic gradient descent
			result, err := optimizer.SemanticGradientDescent(ctx, promptContent, config)
			if err != nil {
				return fmt.Errorf("semantic gradient descent failed: %v", err)
			}

			// Output results
			return outputSemanticResult(result, outputFile, format)
		},
	}

	cmd.Flags().StringVarP(&prompt, "prompt", "p", "", "Prompt to optimize")
	cmd.Flags().StringVar(&promptFile, "prompt-file", "", "File containing prompt to optimize")
	cmd.Flags().StringVarP(&objective, "objective", "b", "", "Optimization objective")
	cmd.Flags().Float64VarP(&learningRate, "learning-rate", "r", 0.1, "Learning rate for gradient descent")
	cmd.Flags().IntVarP(&iterations, "iterations", "i", 10, "Maximum number of iterations")
	cmd.Flags().Float64Var(&convergence, "convergence", 0.001, "Convergence threshold")
	cmd.Flags().BoolVar(&adaptive, "adaptive", false, "Enable adaptive learning rate")
	cmd.Flags().StringVar(&provider, "provider", "openai", "LLM provider")
	cmd.Flags().StringVar(&model, "model", "gpt-4", "Model for semantic evaluation")
	cmd.Flags().StringVar(&outputFile, "output", "", "Output file for results")
	cmd.Flags().StringVarP(&format, "format", "f", "json", "Output format (json, yaml, table)")

	cmd.MarkFlagRequired("objective")

	return cmd
}

// gasoCmd implements Graph-based Agentic System Optimization
func gasoCmd() *cobra.Command {
	var (
		systemFile    string
		objective     string
		iterations    int
		provider      string
		model         string
		outputFile    string
		format        string
		graphFile     string
		optimization  string
		multiObjective bool
	)

	cmd := &cobra.Command{
		Use:   "gaso",
		Short: "Graph-based Agentic System Optimization",
		Long: `GASO (Graph-based Agentic System Optimization) optimizes complex
multi-component agentic systems using semantic gradients and computational
graphs to model dependencies between system components.

This implementation enables optimization of entire AI systems rather than
individual prompts, supporting multi-objective optimization and complex
dependency relationships.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			// Load system definition
			systemDef, err := loadSystemDefinition(systemFile)
			if err != nil {
				return fmt.Errorf("failed to load system definition: %v", err)
			}

			// Create LLM provider
			llmProvider, err := providers.CreateProvider(provider, map[string]interface{}{
				"model": model,
			})
			if err != nil {
				return fmt.Errorf("failed to create provider: %v", err)
			}

			// Initialize GASO optimizer
			optimizer := metaprompt.NewGASOOptimizer(llmProvider)

			// Configure GASO optimization
			config := metaprompt.GASOConfig{
				Objective:        objective,
				Iterations:       iterations,
				OptimizationType: optimization,
				MultiObjective:   multiObjective,
				GraphFile:        graphFile,
				Provider:         provider,
				Model:           model,
			}

			fmt.Printf("Starting GASO optimization...\n")
			fmt.Printf("System components: %d\n", len(systemDef.Components))
			fmt.Printf("Objective: %s\n", objective)
			fmt.Printf("Optimization type: %s\n", optimization)
			fmt.Printf("Iterations: %d\n", iterations)
			if multiObjective {
				fmt.Printf("Multi-objective optimization: ENABLED\n")
			}
			fmt.Printf("Provider: %s (%s)\n\n", provider, model)

			// Run GASO optimization
			result, err := optimizer.OptimizeSystem(ctx, systemDef, config)
			if err != nil {
				return fmt.Errorf("GASO optimization failed: %v", err)
			}

			// Output results
			return outputGASOResult(result, outputFile, format)
		},
	}

	cmd.Flags().StringVarP(&systemFile, "system", "s", "", "System definition file (JSON/YAML)")
	cmd.Flags().StringVarP(&objective, "objective", "b", "", "Optimization objective")
	cmd.Flags().IntVarP(&iterations, "iterations", "i", 20, "Number of optimization iterations")
	cmd.Flags().StringVar(&provider, "provider", "openai", "LLM provider")
	cmd.Flags().StringVar(&model, "model", "gpt-4", "Model for semantic evaluation")
	cmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file for results")
	cmd.Flags().StringVarP(&format, "format", "f", "json", "Output format (json, yaml, table)")
	cmd.Flags().StringVar(&graphFile, "graph", "", "Output computational graph visualization")
	cmd.Flags().StringVar(&optimization, "type", "pareto", "Optimization type (pareto, weighted, lexicographic)")
	cmd.Flags().BoolVar(&multiObjective, "multi-objective", false, "Enable multi-objective optimization")

	cmd.MarkFlagRequired("system")
	cmd.MarkFlagRequired("objective")

	return cmd
}

// Helper functions

func loadPromptContent(prompt, promptFile string) (string, error) {
	if promptFile != "" {
		data, err := os.ReadFile(promptFile)
		if err != nil {
			return "", fmt.Errorf("failed to read prompt file: %v", err)
		}
		return string(data), nil
	}
	if prompt != "" {
		return prompt, nil
	}
	return "", fmt.Errorf("must provide either --prompt or --prompt-file")
}

func loadSystemDefinition(systemFile string) (*metaprompt.SystemDefinition, error) {
	if systemFile == "" {
		return nil, fmt.Errorf("system file is required")
	}
	
	data, err := os.ReadFile(systemFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read system file: %v", err)
	}
	
	var systemDef metaprompt.SystemDefinition
	if err := json.Unmarshal(data, &systemDef); err != nil {
		return nil, fmt.Errorf("failed to parse system definition: %v", err)
	}
	
	return &systemDef, nil
}

func outputSemanticResult(result *metaprompt.SemanticResult, outputFile, format string) error {
	var output []byte
	var err error

	switch format {
	case "json":
		output, err = json.MarshalIndent(result, "", "  ")
	case "yaml":
		// YAML output would require a YAML library
		return fmt.Errorf("YAML output not yet implemented")
	case "table":
		return outputSemanticTable(result)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}

	if err != nil {
		return fmt.Errorf("failed to marshal result: %v", err)
	}

	if outputFile != "" {
		return os.WriteFile(outputFile, output, 0644)
	}

	fmt.Println(string(output))
	return nil
}

func outputGASOResult(result *metaprompt.GASOResult, outputFile, format string) error {
	var output []byte
	var err error

	switch format {
	case "json":
		output, err = json.MarshalIndent(result, "", "  ")
	case "yaml":
		return fmt.Errorf("YAML output not yet implemented")
	case "table":
		return outputGASOTable(result)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}

	if err != nil {
		return fmt.Errorf("failed to marshal result: %v", err)
	}

	if outputFile != "" {
		return os.WriteFile(outputFile, output, 0644)
	}

	fmt.Println(string(output))
	return nil
}

func outputSemanticTable(result *metaprompt.SemanticResult) error {
	// Table output implementation
	fmt.Printf("Semantic Optimization Results\n")
	fmt.Printf("=============================\n\n")
	fmt.Printf("Initial Score: %.4f\n", result.InitialScore)
	fmt.Printf("Final Score: %.4f\n", result.FinalScore)
	fmt.Printf("Improvement: %.4f\n", result.FinalScore-result.InitialScore)
	fmt.Printf("Iterations: %d\n", result.Iterations)
	fmt.Printf("Convergence: %t\n", result.Converged)
	fmt.Printf("\nOptimized Prompt:\n%s\n", result.OptimizedPrompt)
	return nil
}

func outputGASOTable(result *metaprompt.GASOResult) error {
	// Table output implementation
	fmt.Printf("GASO Optimization Results\n")
	fmt.Printf("========================\n\n")
	fmt.Printf("System Components: %d\n", len(result.OptimizedComponents))
	fmt.Printf("Optimization Iterations: %d\n", result.Iterations)
	fmt.Printf("Overall Performance: %.4f\n", result.OverallPerformance)
	fmt.Printf("Pareto Efficient: %t\n", result.ParetoEfficient)
	return nil
}