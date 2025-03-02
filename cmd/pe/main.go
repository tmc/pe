// The pe command is a toolkit for prompt engineering tasks.
// It provides commands for evaluating, viewing, validating, and formatting prompt configurations.
//
// Usage:
//
//	pe [command]
//
// The commands are:
//
//	eval        evaluate prompt configurations against LLM providers
//	view        view evaluation results in browser UI
//	vet         validate promptfoo configuration files
//	fmt         format promptfoo configuration files
//	convert     convert promptfoo configuration files between formats
//	benchmark   compare performance metrics of prompts and providers
//
// Examples:
//
//	pe eval test-config.yaml
//	pe eval -c test-config.yaml -o results.json
//	pe view
//	pe vet config.yaml
//	pe fmt config.yaml --output yaml
//	pe convert config.yaml config.json --output json
//	pe benchmark benchmark-config.yaml --iterations 5 --concurrency 2 --format text
package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	root := &cobra.Command{
		Use:   "pe",
		Short: "Toolkit for prompt engineering",
		Long:  `pe is a collection of tools for working with prompt engineering concepts, files, and tools.`,
	}

	root.AddCommand(evalCmd())
	root.AddCommand(viewCmd())
	root.AddCommand(vetCmd())
	root.AddCommand(fmtCmd())
	root.AddCommand(convertCmd())
	root.AddCommand(benchmarkCmd())
	root.AddCommand(templateTestCmd())

	if err := root.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}