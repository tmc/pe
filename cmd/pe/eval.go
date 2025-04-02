package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/tmc/pe/internal/evaluator"
	"sigs.k8s.io/yaml"
)

// evalCmd returns a cobra.Command for the 'eval' subcommand.
func evalCmd() *cobra.Command {
	var configFile string
	var outputFile string
	var timeout string
	var dryRun bool
	var saveToDb bool
	var share bool
	var maxConcurrency int
	var noProgressBar bool

	cmd := &cobra.Command{
		Use:   "eval [config_file]",
		Short: "Evaluate prompt configurations",
		Long:  `Evaluate prompt configurations against LLM providers.

When used with --save-db flag, results will be saved to the promptfoo database
and can be viewed later using the 'pe view' command with the evaluation ID.`,
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Check if config file provided as positional arg
			if len(args) > 0 {
				configFile = args[0]
			}
			
			// If no config file provided, check for default config files
			if configFile == "" {
				defaultConfigPaths := []string{"promptfooconfig.yaml"}
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
			
			// Parse timeout
			parsedTimeout, err := time.ParseDuration(timeout)
			if err != nil {
				return fmt.Errorf("invalid timeout: %v", err)
			}

			// Read config file
			data, err := os.ReadFile(configFile)
			if err != nil {
				return fmt.Errorf("error reading config file: %v", err)
			}

			// Parse the config
			var config map[string]interface{}
			err = yaml.Unmarshal(data, &config)
			if err != nil {
				return fmt.Errorf("error parsing config file: %v", err)
			}
			
			// Run the evaluator
			results, err := evaluator.Evaluate(config, parsedTimeout, dryRun, maxConcurrency, !noProgressBar)
			if err != nil {
				return fmt.Errorf("evaluation error: %v", err)
			}
			
			// Format results as a table by default for display
			output, err := evaluator.FormatResults(results, "table")
			if err != nil {
				return fmt.Errorf("error formatting results: %v", err)
			}
			
			// Write to output file or stdout
			if outputFile != "" {
				// Infer output format from file extension
				outputFormat := ""
				if strings.HasSuffix(outputFile, ".json") {
					outputFormat = "json"
				} else if strings.HasSuffix(outputFile, ".yaml") || strings.HasSuffix(outputFile, ".yml") {
					outputFormat = "yaml"
				} else if strings.HasSuffix(outputFile, ".csv") {
					outputFormat = "csv"
				}
				
				// For JSON output, ensure we're getting just the JSON data
				if outputFormat == "json" {
					// Re-marshal to ensure clean JSON output and match promptfoo's structure order
					// Create an ordered map that preserves insertion order for JSON serialization
					// The order must match: evalId, results, config, [shareableUrl]
					
					// Use a struct with fields in the correct order to preserve serialization order
					type OrderedOutput struct {
						EvalId       interface{} `json:"evalId"`
						Results      interface{} `json:"results"`
						Config       interface{} `json:"config"`
						ShareableUrl string      `json:"shareableUrl,omitempty"`
					}
					
					evalId := results["evalId"]
					
					// Create output struct with fields in the desired order
					output := OrderedOutput{
						EvalId:  evalId,
						Results: results["results"],
						Config:  results["config"],
					}
					
					// Add shareableUrl if share flag is set
					if share {
						output.ShareableUrl = fmt.Sprintf("https://promptfoo.dev/eval/%s", evalId)
					}
					
					// Marshal using the standard json package
					jsonData, err := json.Marshal(output)
					
					// Format the JSON with indentation for readability
					var prettyBuf bytes.Buffer
					err = json.Indent(&prettyBuf, jsonData, "", "  ")
					if err != nil {
						return fmt.Errorf("error indenting JSON: %v", err)
					}
					jsonData = prettyBuf.Bytes()
					if err != nil {
						return fmt.Errorf("error formatting JSON: %v", err)
					}
					err = os.WriteFile(outputFile, jsonData, 0644)
				} else if outputFormat == "yaml" {
					// Convert results to YAML
					yamlData, err := yaml.Marshal(results)
					if err != nil {
						return fmt.Errorf("error formatting YAML: %v", err)
					}
					err = os.WriteFile(outputFile, yamlData, 0644)
				} else if outputFormat == "csv" {
					// Format results as CSV
					csvData, err := evaluator.FormatResults(results, "csv")
					if err != nil {
						return fmt.Errorf("error formatting CSV: %v", err)
					}
					err = os.WriteFile(outputFile, csvData, 0644)
				} else {
					// Default to table format
					err = os.WriteFile(outputFile, output, 0644)
				}
				
				if err != nil {
					return fmt.Errorf("error writing output file: %v", err)
				}
				fmt.Fprintf(cmd.OutOrStdout(), "Results written to %s\n", outputFile)
			} else {
				fmt.Fprint(cmd.OutOrStdout(), string(output))
			}
			
			// Save to a file in the .promptfoo directory if requested
			if saveToDb {
				evalId, _ := results["evalId"].(string)
				
				// Create the .promptfoo directory in the user's home directory
				homeDir, err := os.UserHomeDir()
				if err != nil {
					return fmt.Errorf("error getting user home directory: %v", err)
				}
				
				// Create the .promptfoo/evals directory
				promptfooDir := filepath.Join(homeDir, ".promptfoo", "evals")
				if err := os.MkdirAll(promptfooDir, 0755); err != nil {
					return fmt.Errorf("error creating promptfoo directory: %v", err)
				}
				
				// Create a file for this evaluation
				evalFile := filepath.Join(promptfooDir, evalId+".json")
				
				// Format results as JSON for the storage
				jsonOutput, err := evaluator.FormatResults(results, "json")
				if err != nil {
					return fmt.Errorf("error formatting results as JSON: %v", err)
				}
				
				// Write the JSON to the evaluation file
				if err := os.WriteFile(evalFile, jsonOutput, 0644); err != nil {
					return fmt.Errorf("error writing to evaluation file: %v", err)
				}
				
				fmt.Fprintf(cmd.OutOrStdout(), "Evaluation results saved to: %s\n", evalFile)
				fmt.Fprintf(cmd.OutOrStdout(), "\nTo view these results, run:\n")
				fmt.Fprintf(cmd.OutOrStdout(), "pe view -f %s\n", evalFile)
			}
			
			return nil
		},
	}

	cmd.Flags().StringVarP(&configFile, "config", "c", "", "Path to configuration file")
	cmd.Flags().StringVarP(&outputFile, "output", "o", "", "Write results to file (format inferred from extension: .json, .yaml, .csv)")
	cmd.Flags().StringVarP(&timeout, "timeout", "t", "30s", "Timeout for the entire test run")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show commands that would be executed without running them")
	cmd.Flags().BoolVar(&saveToDb, "save-db", false, "Save results to promptfoo database for viewing with 'pe view'")
	cmd.Flags().BoolVar(&share, "share", false, "Create a shareable URL of evaluation results")
	cmd.Flags().IntVarP(&maxConcurrency, "max-concurrency", "j", 4, "Maximum number of concurrent API calls")
	cmd.Flags().BoolVar(&noProgressBar, "no-progress-bar", false, "Do not show progress bar")

	return cmd
}