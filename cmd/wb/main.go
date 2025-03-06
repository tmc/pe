// Command wb (workbench) provides a command-line interface for working with
// prompt engineering protobuf files.
package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "wb",
	Short: "Prompt Engineering Workbench",
	Long: `WB (Workbench) is a tool for working with prompt engineering protobuf files.
It provides utilities for creating, viewing, analyzing, and manipulating prompts.`,
	SilenceUsage: true,
}

func init() {
	// Add subcommands
	rootCmd.AddCommand(createCmd)
	rootCmd.AddCommand(showCmd)
	rootCmd.AddCommand(analyzeCmd)
	rootCmd.AddCommand(validateCmd)
	rootCmd.AddCommand(convertCmd)
	rootCmd.AddCommand(evalCmd)
	rootCmd.AddCommand(formatCmd)
	rootCmd.AddCommand(mergeCmd)
	rootCmd.AddCommand(diffCmd)
	rootCmd.AddCommand(exportCmd)
	rootCmd.AddCommand(importCmd)
	rootCmd.AddCommand(testCmd)
	rootCmd.AddCommand(statCmd)
	rootCmd.AddCommand(searchCmd)
}