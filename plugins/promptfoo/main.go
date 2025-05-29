package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/tmc/pe/internal/promptfoo"
	"gopkg.in/yaml.v3"
)

const version = "0.1.0"

func main() {
	// Handle plugin info request
	if len(os.Args) > 1 && os.Args[1] == "--pe-plugin-info" {
		info := map[string]interface{}{
			"description": "Promptfoo compatibility layer for PE",
			"version":     version,
			"commands": []map[string]string{
				{
					"name":        "import",
					"description": "Import promptfoo configuration",
					"usage":       "pe promptfoo import <config.yaml>",
				},
				{
					"name":        "export", 
					"description": "Export to promptfoo format",
					"usage":       "pe promptfoo export <config.yaml>",
				},
				{
					"name":        "convert",
					"description": "Convert between formats",
					"usage":       "pe promptfoo convert <input> <output>",
				},
			},
		}
		
		data, _ := json.Marshal(info)
		fmt.Println(string(data))
		os.Exit(0)
	}

	// Run as normal command
	rootCmd := &cobra.Command{
		Use:   "pe-promptfoo",
		Short: "Promptfoo compatibility plugin for PE",
		Long:  `Import, export, and convert between PE and Promptfoo formats.`,
	}

	rootCmd.AddCommand(
		importCmd(),
		exportCmd(),
		convertCmd(),
	)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func importCmd() *cobra.Command {
	var (
		output string
		format string
	)

	cmd := &cobra.Command{
		Use:   "import <config-file>",
		Short: "Import a Promptfoo configuration file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			inputFile := args[0]

			// Read the promptfoo config
			data, err := os.ReadFile(inputFile)
			if err != nil {
				return fmt.Errorf("failed to read input file: %w", err)
			}

			var config promptfoo.Config
			if err := yaml.Unmarshal(data, &config); err != nil {
				return fmt.Errorf("failed to parse promptfoo config: %w", err)
			}

			// Convert to PE format
			peConfig := convertFromPromptfoo(&config)

			// Output based on format
			var outputData []byte
			switch format {
			case "json":
				outputData, err = json.MarshalIndent(peConfig, "", "  ")
			case "yaml":
				outputData, err = yaml.Marshal(peConfig)
			default:
				return fmt.Errorf("unsupported format: %s", format)
			}

			if err != nil {
				return fmt.Errorf("failed to marshal output: %w", err)
			}

			if output == "-" || output == "" {
				fmt.Print(string(outputData))
			} else {
				if err := os.WriteFile(output, outputData, 0644); err != nil {
					return fmt.Errorf("failed to write output file: %w", err)
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&output, "output", "o", "-", "Output file (- for stdout)")
	cmd.Flags().StringVarP(&format, "format", "f", "yaml", "Output format (yaml or json)")

	return cmd
}

func exportCmd() *cobra.Command {
	var (
		output string
		format string
	)

	cmd := &cobra.Command{
		Use:   "export <config-file>",
		Short: "Export a PE configuration to Promptfoo format",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			inputFile := args[0]

			// Read the PE config
			data, err := os.ReadFile(inputFile)
			if err != nil {
				return fmt.Errorf("failed to read input file: %w", err)
			}

			// Parse based on extension
			var peConfig map[string]interface{}
			if err := yaml.Unmarshal(data, &peConfig); err != nil {
				// Try JSON
				if err := json.Unmarshal(data, &peConfig); err != nil {
					return fmt.Errorf("failed to parse PE config: %w", err)
				}
			}

			// Convert to promptfoo format
			pfConfig := convertToPromptfoo(peConfig)

			// Output based on format
			var outputData []byte
			switch format {
			case "json":
				outputData, err = json.MarshalIndent(pfConfig, "", "  ")
			case "yaml":
				outputData, err = yaml.Marshal(pfConfig)
			default:
				return fmt.Errorf("unsupported format: %s", format)
			}

			if err != nil {
				return fmt.Errorf("failed to marshal output: %w", err)
			}

			if output == "-" || output == "" {
				fmt.Print(string(outputData))
			} else {
				if err := os.WriteFile(output, outputData, 0644); err != nil {
					return fmt.Errorf("failed to write output file: %w", err)
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&output, "output", "o", "-", "Output file (- for stdout)")
	cmd.Flags().StringVarP(&format, "format", "f", "yaml", "Output format (yaml or json)")

	return cmd
}

func convertCmd() *cobra.Command {
	var (
		inputFormat  string
		outputFormat string
		direction    string
	)

	cmd := &cobra.Command{
		Use:   "convert <input-file> <output-file>",
		Short: "Convert between PE and Promptfoo formats",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			inputFile := args[0]
			outputFile := args[1]

			// Read input
			data, err := os.ReadFile(inputFile)
			if err != nil {
				return fmt.Errorf("failed to read input file: %w", err)
			}

			var outputData []byte

			switch direction {
			case "pe-to-promptfoo":
				// Parse PE config
				var peConfig map[string]interface{}
				if inputFormat == "json" {
					err = json.Unmarshal(data, &peConfig)
				} else {
					err = yaml.Unmarshal(data, &peConfig)
				}
				if err != nil {
					return fmt.Errorf("failed to parse input: %w", err)
				}

				// Convert to promptfoo
				pfConfig := convertToPromptfoo(peConfig)

				// Marshal output
				if outputFormat == "json" {
					outputData, err = json.MarshalIndent(pfConfig, "", "  ")
				} else {
					outputData, err = yaml.Marshal(pfConfig)
				}

			case "promptfoo-to-pe":
				// Parse promptfoo config
				var pfConfig promptfoo.Config
				if inputFormat == "json" {
					err = json.Unmarshal(data, &pfConfig)
				} else {
					err = yaml.Unmarshal(data, &pfConfig)
				}
				if err != nil {
					return fmt.Errorf("failed to parse input: %w", err)
				}

				// Convert to PE
				peConfig := convertFromPromptfoo(&pfConfig)

				// Marshal output
				if outputFormat == "json" {
					outputData, err = json.MarshalIndent(peConfig, "", "  ")
				} else {
					outputData, err = yaml.Marshal(peConfig)
				}

			default:
				return fmt.Errorf("unknown direction: %s", direction)
			}

			if err != nil {
				return fmt.Errorf("failed to marshal output: %w", err)
			}

			// Write output
			if outputFile == "-" {
				fmt.Print(string(outputData))
			} else {
				if err := os.WriteFile(outputFile, outputData, 0644); err != nil {
					return fmt.Errorf("failed to write output file: %w", err)
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&inputFormat, "input-format", "yaml", "Input format (yaml or json)")
	cmd.Flags().StringVar(&outputFormat, "output-format", "yaml", "Output format (yaml or json)")
	cmd.Flags().StringVar(&direction, "direction", "promptfoo-to-pe", "Conversion direction (promptfoo-to-pe or pe-to-promptfoo)")

	return cmd
}

// Conversion functions
func convertFromPromptfoo(config *promptfoo.Config) map[string]interface{} {
	// TODO: Implement full conversion logic
	// For now, return a basic structure
	return map[string]interface{}{
		"version": "1.0",
		"prompts": config.Prompts,
		"providers": config.Providers,
		"tests": config.Tests,
		"defaultTest": config.DefaultTest,
		"description": config.Description,
	}
}

func convertToPromptfoo(peConfig map[string]interface{}) *promptfoo.Config {
	// TODO: Implement full conversion logic
	// For now, return a basic structure
	config := &promptfoo.Config{}
	
	// Convert prompts if present
	if prompts, ok := peConfig["prompts"]; ok {
		// Handle different prompt formats
		switch v := prompts.(type) {
		case []interface{}:
			for _, p := range v {
				if str, ok := p.(string); ok {
					config.Prompts = append(config.Prompts, str)
				}
			}
		case []string:
			config.Prompts = v
		}
	}
	
	// Convert providers if present
	if providers, ok := peConfig["providers"]; ok {
		switch v := providers.(type) {
		case []interface{}:
			for _, p := range v {
				if str, ok := p.(string); ok {
					config.Providers = append(config.Providers, str)
				}
			}
		case []string:
			config.Providers = v
		}
	}
	
	// Convert tests if present - simplified for now
	if tests, ok := peConfig["tests"]; ok {
		switch v := tests.(type) {
		case []interface{}:
			for _, t := range v {
				if testMap, ok := t.(map[string]interface{}); ok {
					testCase := promptfoo.TestCase{}
					if vars, ok := testMap["vars"].(map[string]interface{}); ok {
						testCase.Vars = vars
					}
					config.Tests = append(config.Tests, testCase)
				}
			}
		}
	}
	
	return config
}