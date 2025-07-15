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
//	test        run advanced testing (property-based, regression)
//	template    manage prompt templates (list, search, apply)
//	profile     profiling and observability tools (CPU, memory, tracing)
//	ask         ask a single question to an LLM provider (pipeline-friendly)
//	stream      process evaluation results as a stream
//	filter      filter evaluation results based on conditions
//	analyze     analyze evaluation results with statistics
//	stats       show quick statistics from evaluation results
//	diff        compare two evaluation results
//	interactive start interactive REPL mode for prompt development
//	optimize    optimize prompts using metaprompting techniques
//	semantic    advanced semantic gradient descent optimization
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
//	
//	# Pipeline-friendly commands for Unix composability:
//	echo "What is AI?" | pe ask --provider openai:gpt-4
//	pe eval config.yaml | pe filter --success | pe stats
//	pe eval config.yaml | pe stream --select response,latency | pe analyze --metric latency
//	pe interactive --provider anthropic:claude-3-haiku
package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	
	// Import providers to register native provider factories
	_ "github.com/tmc/pe/internal/providers"
)

func main() {
	// Handle plugin execution first
	HandlePluginExecution()

	root := &cobra.Command{
		Use:   "pe",
		Short: "PE - Go for Prompts",
		Long: `PE is the unified toolchain for prompt engineering, bringing Go's 
philosophy of simplicity, composability, and performance to LLM development.

Just as Go revolutionized systems programming with its elegant toolchain, 
PE revolutionizes prompt engineering with a comprehensive set of tools that 
work together seamlessly.`,
	}

	// Core commands (like go toolchain)
	root.AddCommand(runCmd())      // go run for prompts
	root.AddCommand(buildCmd)      // go build for prompts
	root.AddCommand(testCmd())     // go test for prompts
	root.AddCommand(docCmd())      // go doc for prompts
	root.AddCommand(peInitCmd())   // pe init for repository
	root.AddCommand(modCmd)        // go mod for prompt modules
	root.AddCommand(pushCmd)       // push modules to registry
	root.AddCommand(editCmd)       // go mod edit for prompts
	root.AddCommand(getCmd)        // pe get for extracting prompt info
	root.AddCommand(evalPromptCmd) // pe eval-prompt for running evals from prompt files
	root.AddCommand(workCmd)       // go work for prompts
	root.AddCommand(attestCmd)     // cryptographic attestations
	root.AddCommand(catCmd())      // pe cat for inspecting prompt files
	
	// Existing commands
	root.AddCommand(evalCmd())
	root.AddCommand(viewCmd())
	root.AddCommand(vetCmd())
	root.AddCommand(promptFmtCmd()) // Use the prompt formatting command
	root.AddCommand(convertCmd())
	root.AddCommand(benchmarkCmd())
	root.AddCommand(watchCmd())
	root.AddCommand(templateCmd())
	root.AddCommand(profileCmd())
	
	// Pipeline-friendly commands for Unix composability
	addPipelineCommands(root)
	root.AddCommand(statsCmd())
	root.AddCommand(diffCmd())
	root.AddCommand(interactiveCmd())
	root.AddCommand(extractCmd())
	root.AddCommand(optimizeCmd())
	root.AddCommand(evolveCmd())
	root.AddCommand(fusionCmd)
	root.AddCommand(composeCmd)
	root.AddCommand(synthesizeCmd)
	root.AddCommand(playgroundCmd())
	root.AddCommand(metricsCmd())
	root.AddCommand(semanticCmd())
	root.AddCommand(securityCmd())

	// Plugin command
	root.AddCommand(pluginCmd())

	// Distributed system commands
	root.AddCommand(cacheCmd)
	root.AddCommand(distributedCmd())

	// Discover and add plugin commands dynamically
	dynamicPluginCommands(root)

	if err := root.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
