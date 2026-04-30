package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
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
		Use:   "promptfoo",
		Short: "Promptfoo compatibility plugin for PE",
		Long:  `Import, export, and convert Promptfoo-compatible configuration files.`,
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

			var config map[string]interface{}
			if err := yaml.Unmarshal(data, &config); err != nil {
				return fmt.Errorf("failed to parse promptfoo config: %w", err)
			}

			// Basic conversion - just pass through for now
			// In a full implementation, this would transform the structure
			peConfig := map[string]interface{}{
				"version":     "1.0",
				"prompts":     config["prompts"],
				"providers":   config["providers"],
				"tests":       config["tests"],
				"description": "Imported from Promptfoo",
			}

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
				fmt.Printf("✓ Imported to %s\n", output)
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

			// Basic conversion - just pass through for now
			pfConfig := map[string]interface{}{
				"prompts":   peConfig["prompts"],
				"providers": peConfig["providers"],
				"tests":     peConfig["tests"],
			}

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
				fmt.Printf("✓ Exported to %s\n", output)
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
	)

	cmd := &cobra.Command{
		Use:   "convert <input-file> <output-file>",
		Short: "Convert between YAML and JSON formats",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			inputFile := args[0]
			outputFile := args[1]

			// Read input
			data, err := os.ReadFile(inputFile)
			if err != nil {
				return fmt.Errorf("failed to read input file: %w", err)
			}

			var content interface{}

			// Parse input
			if inputFormat == "json" || (inputFormat == "" && hasJSONExt(inputFile)) {
				err = json.Unmarshal(data, &content)
			} else {
				err = yaml.Unmarshal(data, &content)
			}
			if err != nil {
				return fmt.Errorf("failed to parse input: %w", err)
			}

			// Marshal output
			var outputData []byte
			if outputFormat == "json" || (outputFormat == "" && hasJSONExt(outputFile)) {
				outputData, err = json.MarshalIndent(content, "", "  ")
			} else {
				outputData, err = yaml.Marshal(content)
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
				fmt.Printf("✓ Converted to %s\n", outputFile)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&inputFormat, "input-format", "", "Input format (yaml or json, auto-detected if not specified)")
	cmd.Flags().StringVar(&outputFormat, "output-format", "", "Output format (yaml or json, auto-detected if not specified)")

	return cmd
}

func hasJSONExt(filename string) bool {
	return len(filename) > 5 && filename[len(filename)-5:] == ".json"
}
