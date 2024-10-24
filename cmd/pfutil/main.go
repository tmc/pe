// The pfutil command performs operations on promptfoo configuration files.
// It is designed to integrate with Unix pipelines for validation and formatting tasks.
//
// Usage:
//
//	cat config.yaml | pfutil [command]
//
// The commands are:
//
//	vet     validate promptfoo configuration files
//	fmt     format promptfoo configuration files
//
// Examples:
//
//	cat config.yaml | pfutil vet
//	cat config.json | pfutil fmt --output yaml
//	pfutil fmt config.yaml --write
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
