package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

// statsCmd shows quick statistics from evaluation results
func statsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stats",
		Short: "Show quick statistics from evaluation results",
		Long:  `Display statistics from evaluation results including pass rates, latency, and token usage.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Implement stats command
			fmt.Println("Evaluation Statistics:")
			fmt.Println("  Total tests: 10")
			fmt.Println("  Pass rate: 90%")
			fmt.Println("  Average latency: 1.2s")
			fmt.Println("  Total tokens: 5432")
			return nil
		},
	}

	return cmd
}

// diffCmd compares two evaluation results
func diffCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "diff [file1] [file2]",
		Short: "Compare two evaluation results",
		Long:  `Compare two evaluation result files and show differences in performance metrics.`,
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Implement diff command
			fmt.Printf("Comparing %s vs %s\n", args[0], args[1])
			fmt.Println("  Pass rate: +5%")
			fmt.Println("  Latency: -0.3s")
			fmt.Println("  Token usage: +120")
			return nil
		},
	}

	return cmd
}

// interactiveCmd starts interactive REPL mode
func interactiveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "interactive",
		Short: "Start interactive REPL mode for prompt development",
		Long:  `Launch an interactive REPL (Read-Eval-Print Loop) for iterative prompt development.`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			provider, err := cmd.Flags().GetString("provider")
			if err != nil {
				return err
			}
			configFile, err := cmd.Flags().GetString("config")
			if err != nil {
				return err
			}
			temperature, err := cmd.Flags().GetFloat64("temperature")
			if err != nil {
				return err
			}
			if temperature < 0 || temperature > 2 {
				return fmt.Errorf("temperature must be between 0.0 and 2.0")
			}

			return NewREPLSession(cmd, provider, configFile, temperature).Run()
		},
	}

	cmd.Flags().String("provider", "openai:gpt-4", "LLM provider to use")
	cmd.Flags().String("config", "", "Path to configuration file")
	cmd.Flags().Float64("temperature", 0.7, "Generation temperature")

	return cmd
}
