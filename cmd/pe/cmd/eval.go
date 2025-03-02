package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/tmc/pe/internal/config"
	"github.com/tmc/pe/internal/evaluator"
)

// NewEvalCommand returns a command for evaluating prompt configurations.
func NewEvalCommand() *cobra.Command {
	var configFile string
	var outputFile string
	var outputFormat string
	var timeout string
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "eval [config_file]",
		Short: "Evaluate prompt configurations against LLM providers",
		Long:  `Evaluate prompt configurations against language model providers using definitions from a YAML or JSON file.`,
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Check if config file provided as positional arg
			if len(args) > 0 {
				configFile = args[0]
			}

			// If no config file provided, check flag
			if configFile == "" {
				configFile, _ = cmd.Flags().GetString("config")
				if configFile == "" {
					return fmt.Errorf("no configuration file provided")
				}
			}

			// Parse timeout
			timeoutDuration, err := time.ParseDuration(timeout)
			if err != nil {
				return fmt.Errorf("invalid timeout: %v", err)
			}

			// Load config
			cfg, err := config.LoadConfig(configFile)
			if err != nil {
				return err
			}

			// Validate config
			if err := cfg.Validate(); err != nil {
				return fmt.Errorf("invalid configuration: %v", err)
			}

			// Perform evaluation
			var result interface{}

			if dryRun {
				// Generate commands for all providers and prompts
				result, err = DryRunEval(cfg)
			} else {
				// Look for external evaluators first
				extEval := findEvaluator(cfg.Providers)
				if extEval != nil {
					fmt.Fprintf(cmd.OutOrStderr(), "Using external evaluator: %s\n", extEval.Name())
					result, err = runExternalEvaluation(extEval, cfg, timeoutDuration)
				} else {
					// Fall back to built-in implementation
					fmt.Fprintf(cmd.OutOrStderr(), "Using built-in evaluation\n")
					result, err = runBuiltinEvaluation(cfg, timeoutDuration)
				}
			}

			if err != nil {
				return err
			}

			// Format and output results
			if outputFile != "" {
				if err := evaluator.WriteResultToFile(result, outputFile, outputFormat); err != nil {
					return fmt.Errorf("error writing results: %v", err)
				}
			}

			return nil
		},
	}

	// Add flags
	cmd.Flags().StringVarP(&configFile, "config", "c", "", "Path to configuration file")
	cmd.Flags().StringVarP(&outputFile, "output", "o", "", "Write results to file")
	cmd.Flags().StringVarP(&outputFormat, "format", "f", "json", "Output format: 'json', 'yaml', or 'text'")
	cmd.Flags().StringVarP(&timeout, "timeout", "t", "30s", "Timeout for the entire test run")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show commands that would be executed without running them")

	return cmd
}

// DryRunEval performs a dry run evaluation, generating commands for all providers and prompts.
func DryRunEval(cfg *config.Config) (interface{}, error) {
	// For dry run, we'll use the specialized CGPT evaluator
	// that outputs shell commands without running them
	return RunCGPTDryRun(cfg.AsMap())
}

// findEvaluator looks for an appropriate evaluator for the providers.
func findEvaluator(providers []string) *evaluator.ExternalProvider {
	// Try each provider in order
	for _, provider := range providers {
		parts := strings.Split(provider, ":")
		providerName := parts[0]
		
		eval := evaluator.FindExternalEvaluator(providerName)
		if eval != nil {
			return eval
		}
	}
	
	// Try a generic evaluator as a fallback
	return evaluator.FindExternalEvaluator("")
}

// runExternalEvaluation runs an evaluation using an external evaluator.
func runExternalEvaluation(eval *evaluator.ExternalProvider, cfg *config.Config, timeout time.Duration) (interface{}, error) {
	// Set timeout on the evaluator
	eval.Timeout = timeout
	
	// Convert config to JSON for the external evaluator
	configMap := cfg.AsMap()
	
	// Run evaluation
	ctx := context.Background()
	result, err := eval.Evaluate(ctx, "", configMap)
	if err != nil {
		return nil, fmt.Errorf("external evaluation failed: %v", err)
	}
	
	return result, nil
}

// runBuiltinEvaluation runs an evaluation using the built-in CGPT evaluator.
func runBuiltinEvaluation(cfg *config.Config, timeout time.Duration) (interface{}, error) {
	// Use the CGPT evaluator with the provided configuration
	return EvaluateWithCGPT(cfg.AsMap())
}