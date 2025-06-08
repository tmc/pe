package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"
)

// Vet command flags
var (
	vetProvider   string
	vetQuiet      bool
	vetVerbose    bool
	vetStopOnFail bool
)

// vetCmd returns a cobra.Command for the 'vet' subcommand.
//
// vet validates prompt files and runs their embedded evaluations.
// This is the default command for testing prompts.
//
// Usage:
//
//	pe vet [file...]  # Validate and test prompt files
//	pe vet            # Vet all .txt files in current directory
func vetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vet [file...]",
		Short: "Validate prompt files and run their evals",
		Long: `Vet examines prompt files and runs their embedded evaluations.

If a prompt file contains an -- evals -- section, vet will automatically
run those evaluations. Otherwise, it performs basic validation.

This is the default command for testing prompts, similar to 'go vet'.`,
		RunE: runVetPrompts,
	}

	cmd.Flags().StringVar(&vetProvider, "provider", "", "LLM provider to use for evals")
	cmd.Flags().BoolVarP(&vetQuiet, "quiet", "q", false, "Only show failures")
	cmd.Flags().BoolVarP(&vetVerbose, "verbose", "v", false, "Show detailed output")
	cmd.Flags().BoolVar(&vetStopOnFail, "stop-on-fail", false, "Stop on first failure")

	return cmd
}

// fmtCmd returns a cobra.Command for the 'fmt' subcommand.
//
// fmt formats promptfoo configuration files and can convert between YAML and JSON.
//
// Usage:
//
//	cat config.yaml | pe fmt [--output yaml|json]
//	pe fmt [file...] [--write] [--output yaml|json]
func fmtCmd() *cobra.Command {
	var writeFlag bool
	var outputFormat string

	cmd := &cobra.Command{
		Use:   "fmt [file...]",
		Short: "Format promptfoo configuration files",
		RunE:  runFmt,
	}

	cmd.Flags().BoolVarP(&writeFlag, "write", "w", false, "Write result to (source) file instead of stdout")
	cmd.Flags().StringVarP(&outputFormat, "output", "o", "", "Output format: 'yaml' or 'json' (default is input format)")

	return cmd
}

// initCmd returns a cobra.Command for the 'init' subcommand.
//
// init creates a new promptfoo configuration file with a basic template.
//
// Usage:
//
//	pe init [output_file]
func initCmd() *cobra.Command {
	var format string
	var force bool

	cmd := &cobra.Command{
		Use:   "init [output_file]",
		Short: "Create a new promptfoo configuration file",
		Long:  `Create a new promptfoo configuration file with a basic template.`,
		Args:  cobra.MaximumNArgs(1),
		RunE:  runInit,
	}

	cmd.Flags().StringVarP(&format, "format", "f", "yaml", "Output format: 'yaml' or 'json'")
	cmd.Flags().BoolVarP(&force, "force", "", false, "Overwrite existing file if it exists")

	return cmd
}

// watchCmd returns a cobra.Command for the 'watch' subcommand.
//
// watch monitors configuration files and prompts for changes and re-runs evaluations.
//
// Usage:
//
//	pe watch [config_file]
func watchCmd() *cobra.Command {
	var configFile string
	var outputFile string
	var include string

	cmd := &cobra.Command{
		Use:   "watch [config_file]",
		Short: "Watch files and re-run evaluations on changes",
		Long:  `Watch configuration files and prompt files for changes and automatically re-run evaluations.`,
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Check if config file provided as positional arg
			if len(args) > 0 {
				configFile = args[0]
			}

			// If no config file provided, check for default config files
			if configFile == "" {
				defaultConfigPaths := []string{"promptfooconfig.yaml", "pe-config.yaml", "config.yaml"}
				for _, path := range defaultConfigPaths {
					if _, err := os.Stat(path); err == nil {
						configFile = path
						fmt.Fprintf(cmd.OutOrStdout(), "Using default configuration file: %s\n", path)
						break
					}
				}

				if configFile == "" {
					return fmt.Errorf("no configuration file provided")
				}
			}

			return runWatch(cmd, configFile, outputFile, include)
		},
	}

	cmd.Flags().StringVarP(&configFile, "config", "c", "", "Path to configuration file")
	cmd.Flags().StringVarP(&outputFile, "output", "o", "", "Write results to file")
	cmd.Flags().StringVarP(&include, "include", "i", "*.yaml,*.yml,*.json,prompts/**/*", "Comma-separated list of glob patterns to watch")

	return cmd
}

func runVetConfig(cmd *cobra.Command, args []string) error {
	// Original vet for YAML config files
	for _, file := range args {
		data, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("error reading file %s: %v", file, err)
		}

		var config map[string]interface{}
		err = yaml.Unmarshal(data, &config)
		if err != nil {
			return fmt.Errorf("%s: Error: %v", file, err)
		}

		// Basic validation: check if 'prompts' field exists
		if _, ok := config["prompts"]; !ok {
			return fmt.Errorf("%s: Error: missing required field 'prompts'", file)
		}

		fmt.Fprintf(cmd.OutOrStdout(), "%s: OK\n", file)
	}
	return nil
}

func runVetPrompts(cmd *cobra.Command, args []string) error {
	// Simple implementation that calls eval-prompt for each file with evals
	if len(args) == 0 {
		// Default to all .txt files
		matches, err := filepath.Glob("*.txt")
		if err == nil && len(matches) > 0 {
			args = matches
		}
	}

	hasFailures := false
	for _, file := range args {
		// Check if it's a YAML/JSON config file
		if strings.HasSuffix(file, ".yaml") || strings.HasSuffix(file, ".yml") || strings.HasSuffix(file, ".json") {
			// Use old config vet
			if err := runVetConfig(cmd, []string{file}); err != nil {
				hasFailures = true
			}
			continue
		}

		// For prompt files, run eval-prompt if they have evals
		if !vetQuiet {
			fmt.Printf("=== VET %s\n", file)
		}

		// Use eval-prompt command to run the evals
		evalArgs := []string{file}
		if vetProvider != "" {
			evalArgs = append(evalArgs, "--provider", vetProvider)
		}

		evalCmd := evalPromptCmd
		evalCmd.SetArgs(evalArgs)

		if err := evalCmd.Execute(); err != nil {
			if !vetQuiet {
				fmt.Printf("FAIL: %v\n", err)
			}
			hasFailures = true
		} else if !vetQuiet {
			fmt.Printf("PASS\n")
		}
	}

	if hasFailures {
		return fmt.Errorf("some files failed validation")
	}
	return nil
}

func runFmt(cmd *cobra.Command, args []string) error {
	writeFlag, _ := cmd.Flags().GetBool("write")
	outputFormat, _ := cmd.Flags().GetString("output")

	for _, file := range args {
		data, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("error reading file %s: %v", file, err)
		}

		var config map[string]interface{}
		err = yaml.Unmarshal(data, &config)
		if err != nil {
			return fmt.Errorf("error parsing file %s: %v", file, err)
		}

		var output []byte
		if outputFormat == "json" || (outputFormat == "" && filepath.Ext(file) == ".json") {
			output, err = json.MarshalIndent(config, "", "  ")
		} else {
			output, err = yaml.Marshal(config)
		}
		if err != nil {
			return fmt.Errorf("error formatting file %s: %v", file, err)
		}

		if writeFlag {
			err = os.WriteFile(file, output, 0644)
			if err != nil {
				return fmt.Errorf("error writing file %s: %v", file, err)
			}
		} else {
			fmt.Fprint(cmd.OutOrStdout(), string(output))
		}
	}
	return nil
}

func runInit(cmd *cobra.Command, args []string) error {
	format, _ := cmd.Flags().GetString("format")
	force, _ := cmd.Flags().GetBool("force")

	// Determine output file
	outputFile := "pe-config.yaml"
	if len(args) > 0 {
		outputFile = args[0]
	}

	// If format not specified in filename, update extension
	if format == "json" && !strings.HasSuffix(outputFile, ".json") {
		outputFile = strings.TrimSuffix(outputFile, filepath.Ext(outputFile)) + ".json"
	} else if format == "yaml" && !strings.HasSuffix(outputFile, ".yaml") && !strings.HasSuffix(outputFile, ".yml") {
		outputFile = strings.TrimSuffix(outputFile, filepath.Ext(outputFile)) + ".yaml"
	}

	// Check if file exists and not force flag
	if _, err := os.Stat(outputFile); err == nil && !force {
		return fmt.Errorf("file %s already exists, use --force to overwrite", outputFile)
	}

	// Create template configuration
	templateConfig := map[string]interface{}{
		"prompts": []string{
			"What is the capital of {{country}}?",
			"Tell me about the capital city of {{country}}.",
		},
		"providers": []string{
			"openai:gpt-4",
			"anthropic:claude-3-haiku",
		},
		"tests": []map[string]interface{}{
			{
				"vars": map[string]string{
					"country": "France",
				},
				"assert": []map[string]string{
					{
						"type":  "contains",
						"value": "Paris",
					},
				},
			},
			{
				"vars": map[string]string{
					"country": "Japan",
				},
				"assert": []map[string]string{
					{
						"type":  "contains",
						"value": "Tokyo",
					},
				},
			},
		},
	}

	// Format the configuration
	var output []byte
	var err error
	if format == "json" {
		output, err = json.MarshalIndent(templateConfig, "", "  ")
	} else {
		output, err = yaml.Marshal(templateConfig)
	}
	if err != nil {
		return fmt.Errorf("error formatting configuration: %v", err)
	}

	// Write the configuration to file
	err = os.WriteFile(outputFile, output, 0644)
	if err != nil {
		return fmt.Errorf("error writing configuration to %s: %v", outputFile, err)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created new configuration file: %s\n", outputFile)
	fmt.Fprintf(cmd.OutOrStdout(), "\nTo run an evaluation:\n  pe eval %s\n", outputFile)
	return nil
}

func runWatch(cmd *cobra.Command, configFile, outputFile, includePatterns string) error {
	// Create a new file watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("error creating file watcher: %v", err)
	}
	defer watcher.Close()

	// Parse include patterns
	patterns := strings.Split(includePatterns, ",")
	for i, pattern := range patterns {
		patterns[i] = strings.TrimSpace(pattern)
	}

	// Get the directory containing the config file
	configDir := filepath.Dir(configFile)
	if configDir == "" {
		configDir = "."
	}

	// Add files to watch
	filesToWatch := make(map[string]bool)
	filesToWatch[configFile] = true

	// Add prompt files mentioned in the config
	config, err := os.ReadFile(configFile)
	if err == nil {
		var configData map[string]interface{}
		if err := yaml.Unmarshal(config, &configData); err == nil {
			// Add any prompt files found in the config
			if prompts, ok := configData["prompts"].([]interface{}); ok {
				for _, prompt := range prompts {
					if promptStr, ok := prompt.(string); ok {
						// If prompt looks like a file path, add it to watch list
						if strings.HasPrefix(promptStr, "./") || strings.HasPrefix(promptStr, "/") {
							if _, err := os.Stat(promptStr); err == nil {
								filesToWatch[promptStr] = true
							}
						}
					}
				}
			}
		}
	}

	// Find files matching the patterns to watch
	for _, pattern := range patterns {
		matches, _ := filepath.Glob(filepath.Join(configDir, pattern))
		for _, match := range matches {
			absPath, _ := filepath.Abs(match)
			filesToWatch[absPath] = true
		}
	}

	// Watch directories containing files we want to monitor
	dirsToWatch := make(map[string]bool)
	for file := range filesToWatch {
		dir := filepath.Dir(file)
		dirsToWatch[dir] = true
	}

	// Add directories to watcher
	for dir := range dirsToWatch {
		fmt.Fprintf(cmd.OutOrStdout(), "Watching directory: %s\n", dir)
		err = watcher.Add(dir)
		if err != nil {
			return fmt.Errorf("error watching directory %s: %v", dir, err)
		}
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Watching for changes to %s and related files...\n", configFile)
	fmt.Fprintf(cmd.OutOrStdout(), "Press Ctrl+C to stop\n\n")

	// Initial run
	runEval(cmd, configFile, outputFile)

	// Create a debounce timer to avoid running multiple evaluations in quick succession
	var timer *time.Timer
	debounceDelay := 500 * time.Millisecond

	// Process events
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return nil
			}

			// Check if this is a relevant file
			isWatched := false
			for file := range filesToWatch {
				if event.Name == file {
					isWatched = true
					break
				}
			}

			// For new files that match our patterns
			if !isWatched {
				for _, pattern := range patterns {
					matched, _ := filepath.Match(pattern, filepath.Base(event.Name))
					if matched {
						isWatched = true
						filesToWatch[event.Name] = true
						break
					}
				}
			}

			if isWatched && (event.Op&(fsnotify.Write|fsnotify.Create) != 0) {
				// Debounce events
				if timer != nil {
					timer.Stop()
				}
				timer = time.AfterFunc(debounceDelay, func() {
					fmt.Fprintf(cmd.OutOrStdout(), "\nFile changed: %s\n", event.Name)
					runEval(cmd, configFile, outputFile)
				})
			}

		case err, ok := <-watcher.Errors:
			if !ok {
				return nil
			}
			fmt.Fprintf(cmd.OutOrStderr(), "Error: %v\n", err)
		}
	}
}

// Helper function to run evaluation
func runEval(cmd *cobra.Command, configFile, outputFile string) {
	fmt.Fprintf(cmd.OutOrStdout(), "Running evaluation at %s...\n", time.Now().Format("15:04:05"))

	// Build arguments for eval command
	args := []string{configFile}
	if outputFile != "" {
		args = append(args, "-o", outputFile)
	}

	// Create and run the eval command
	evalCmd := evalCmd()
	evalCmd.SetOut(cmd.OutOrStdout())
	evalCmd.SetErr(cmd.OutOrStderr())
	evalCmd.SetArgs(args)

	if err := evalCmd.Execute(); err != nil {
		fmt.Fprintf(cmd.OutOrStderr(), "Error: %v\n", err)
	}
}

// convertCmd returns a cobra.Command for the 'convert' subcommand.
//
// convert transforms promptfoo configuration files between different formats (yaml, json, etc.)
//
// Usage:
//
//	pe convert input.yaml output.json [--output json|yaml]
func convertCmd() *cobra.Command {
	var outputFormat string

	cmd := &cobra.Command{
		Use:   "convert [input_file] [output_file]",
		Short: "Convert promptfoo configuration files between formats",
		Long:  `Convert transforms promptfoo configuration files between different formats (yaml, json, etc.)`,
		Args:  cobra.ExactArgs(2),
		RunE:  runConvert,
	}

	cmd.Flags().StringVarP(&outputFormat, "output", "o", "", "Output format: 'yaml' or 'json' (default is determined by output file extension)")

	return cmd
}

func runConvert(cmd *cobra.Command, args []string) error {
	inputFile := args[0]
	outputFile := args[1]
	outputFormat, _ := cmd.Flags().GetString("output")

	// If output format not specified, determine from output file extension
	if outputFormat == "" {
		ext := filepath.Ext(outputFile)
		if ext == ".json" {
			outputFormat = "json"
		} else if ext == ".yaml" || ext == ".yml" {
			outputFormat = "yaml"
		} else {
			return fmt.Errorf("cannot determine output format from extension '%s', please specify --output", ext)
		}
	}

	// Read input file
	data, err := os.ReadFile(inputFile)
	if err != nil {
		return fmt.Errorf("error reading input file %s: %v", inputFile, err)
	}

	// Parse the input
	var config map[string]interface{}
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return fmt.Errorf("error parsing input file %s: %v", inputFile, err)
	}

	// Convert to the output format
	var output []byte
	if outputFormat == "json" {
		output, err = json.MarshalIndent(config, "", "  ")
	} else if outputFormat == "yaml" {
		output, err = yaml.Marshal(config)
	} else {
		return fmt.Errorf("unsupported output format: %s", outputFormat)
	}

	if err != nil {
		return fmt.Errorf("error formatting data: %v", err)
	}

	// Write to output file
	err = os.WriteFile(outputFile, output, 0644)
	if err != nil {
		return fmt.Errorf("error writing output file %s: %v", outputFile, err)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Successfully converted %s to %s\n", inputFile, outputFile)
	return nil
}
