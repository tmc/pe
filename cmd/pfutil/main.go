// The pfutil command performs operations on promptfoo configuration files.
// It is designed to integrate with Unix pipelines for validation and formatting tasks.
//
// Usage:
//
//	cat config.yaml | pfutil [command]
//
// The commands are:
//
//	vet      validate promptfoo configuration files
//	fmt      format promptfoo configuration files
//	convert  convert promptfoo configuration files between formats
//	run      execute promptfoo configuration files against LLM providers
//
// Examples:
//
//	cat config.yaml | pfutil vet
//	cat config.json | pfutil fmt --output yaml
//	pfutil fmt config.yaml --write
//	pfutil convert config.yaml config.json --output json
//	pfutil run config.yaml
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func main() {
	root := &cobra.Command{
		Use:   "pfutil",
		Short: "Utility for promptfoo configuration files",
		Long:  `pfutil performs operations on promptfoo configuration files.`,
	}

	root.AddCommand(vetCmd())
	root.AddCommand(fmtCmd())
	root.AddCommand(convertCmd())
	root.AddCommand(runCmd())

	if err := root.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// vetCmd returns a cobra.Command for the 'vet' subcommand.
//
// vet validates promptfoo configuration files.
//
// Usage:
//
//	cat config.yaml | pfutil vet
//	pfutil vet [file...]
func vetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "vet [file...]",
		Short: "Validate promptfoo configuration files",
		RunE:  runVet,
	}
}

// fmtCmd returns a cobra.Command for the 'fmt' subcommand.
//
// fmt formats promptfoo configuration files and can convert between YAML and JSON.
//
// Usage:
//
//	cat config.yaml | pfutil fmt [--output yaml|json]
//	pfutil fmt [file...] [--write] [--output yaml|json]
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

func runVet(cmd *cobra.Command, args []string) error {
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

// convertCmd returns a cobra.Command for the 'convert' subcommand.
//
// convert transforms promptfoo configuration files between different formats (yaml, json, etc.)
//
// Usage:
//
//	pfutil convert input.yaml output.json [--output json|yaml]
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
