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
		Use:     "pe",
		Short:   "PE - safe prompting toolchain",
		Version: versionLine(),
		Long: `PE is a safe prompting toolchain.

Its primary artifact is executable, templated, composable text. Plain text is
valid by default; files can add inputs, metadata, safety policy, and placement
rules when they need stronger contracts.`,
	}
	root.SetVersionTemplate("{{.Version}}\n")

	registerRootCommands(root)
	dynamicPluginCommands(root)
	applyRootMetadata(root)

	if err := root.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
