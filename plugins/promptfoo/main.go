package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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

			config, err := readConfig(data)
			if err != nil {
				return fmt.Errorf("failed to parse promptfoo config: %w", err)
			}

			// Convert to PE format
			peConfig := convertFromPromptfoo(config)

			// Output based on format
			outputData, err := marshalConfig(peConfig, format)
			if err != nil {
				return err
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

			peConfig, err := readConfig(data)
			if err != nil {
				return fmt.Errorf("failed to parse PE config: %w", err)
			}

			// Convert to promptfoo format
			pfConfig := convertToPromptfoo(peConfig)

			// Output based on format
			outputData, err := marshalConfig(pfConfig, format)
			if err != nil {
				return err
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
				peConfig, err := readConfigWithFormat(data, resolvedFormat(inputFormat, inputFile))
				if err != nil {
					return fmt.Errorf("failed to parse input: %w", err)
				}

				// Convert to promptfoo
				pfConfig := convertToPromptfoo(peConfig)

				outputData, err = marshalConfig(pfConfig, resolvedFormat(outputFormat, outputFile))

			case "promptfoo-to-pe":
				pfConfig, err := readConfigWithFormat(data, resolvedFormat(inputFormat, inputFile))
				if err != nil {
					return fmt.Errorf("failed to parse input: %w", err)
				}

				// Convert to PE
				peConfig := convertFromPromptfoo(pfConfig)

				outputData, err = marshalConfig(peConfig, resolvedFormat(outputFormat, outputFile))

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

	cmd.Flags().StringVar(&inputFormat, "input-format", "", "Input format (yaml or json; defaults from file extension)")
	cmd.Flags().StringVar(&outputFormat, "output-format", "", "Output format (yaml or json; defaults from file extension)")
	cmd.Flags().StringVar(&direction, "direction", "promptfoo-to-pe", "Conversion direction (promptfoo-to-pe or pe-to-promptfoo)")

	return cmd
}

// Conversion functions
func convertFromPromptfoo(config map[string]interface{}) map[string]interface{} {
	out := map[string]interface{}{
		"version": "1.0",
	}

	copyIfPresent(out, config, "description")
	copyIfPresent(out, config, "prompts")

	if providers, ok := config["providers"]; ok {
		out["providers"] = convertProvidersFromPromptfoo(providers)
	}
	if tests, ok := config["tests"]; ok {
		out["tests"] = normalizeTests(tests)
	}
	if defaults, ok := config["defaultTest"]; ok {
		out["defaultTest"] = normalizeDefaultTest(defaults)
	}

	return out
}

func convertToPromptfoo(peConfig map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{})

	copyIfPresent(out, peConfig, "description")
	copyIfPresent(out, peConfig, "prompts")

	if providers, ok := peConfig["providers"]; ok {
		out["providers"] = convertProvidersToPromptfoo(providers)
	}
	if tests, ok := peConfig["tests"]; ok {
		out["tests"] = normalizeTests(tests)
	}
	if defaults, ok := peConfig["defaultTest"]; ok {
		out["defaultTest"] = normalizeDefaultTest(defaults)
	} else if defaults, ok := peConfig["default_test"]; ok {
		out["defaultTest"] = normalizeDefaultTest(defaults)
	}

	return out
}

func readConfig(data []byte) (map[string]interface{}, error) {
	return readConfigWithFormat(data, "")
}

func readConfigWithFormat(data []byte, format string) (map[string]interface{}, error) {
	var config map[string]interface{}

	switch format {
	case "", "yaml":
		if err := yaml.Unmarshal(data, &config); err == nil {
			return config, nil
		}
		if format == "yaml" {
			return nil, fmt.Errorf("failed to parse yaml config")
		}
	case "json":
		if err := json.Unmarshal(data, &config); err == nil {
			return config, nil
		}
		return nil, fmt.Errorf("failed to parse json config")
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}

	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}
	return config, nil
}

func marshalConfig(config map[string]interface{}, format string) ([]byte, error) {
	switch format {
	case "json":
		return json.MarshalIndent(config, "", "  ")
	case "", "yaml":
		return yaml.Marshal(config)
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
}

func resolvedFormat(format, filename string) string {
	if format != "" {
		return format
	}
	if strings.EqualFold(filepath.Ext(filename), ".json") {
		return "json"
	}
	return "yaml"
}

func copyIfPresent(dst, src map[string]interface{}, key string) {
	if value, ok := src[key]; ok {
		dst[key] = value
	}
}

func normalizeTests(value interface{}) []interface{} {
	items, ok := value.([]interface{})
	if !ok {
		return nil
	}

	tests := make([]interface{}, 0, len(items))
	for _, item := range items {
		test, ok := stringMap(item)
		if !ok {
			tests = append(tests, item)
			continue
		}

		out := make(map[string]interface{})
		for key, val := range test {
			switch key {
			case "vars", "variables":
			case "assert", "assertions":
			default:
				out[key] = val
			}
		}

		if vars, ok := firstPresent(test, "vars", "variables"); ok {
			out["vars"] = vars
		}
		if assert, ok := firstPresent(test, "assert", "assertions"); ok {
			out["assert"] = normalizeAssertions(assert)
		}

		tests = append(tests, out)
	}

	return tests
}

func normalizeDefaultTest(value interface{}) interface{} {
	defaults, ok := stringMap(value)
	if !ok {
		return value
	}

	out := make(map[string]interface{})
	for key, val := range defaults {
		if key == "assert" || key == "assertions" {
			continue
		}
		out[key] = val
	}

	if assert, ok := firstPresent(defaults, "assert", "assertions"); ok {
		out["assert"] = normalizeAssertions(assert)
	}

	return out
}

func normalizeAssertions(value interface{}) []interface{} {
	items, ok := value.([]interface{})
	if !ok {
		return nil
	}

	assertions := make([]interface{}, 0, len(items))
	for _, item := range items {
		assertion, ok := stringMap(item)
		if !ok {
			assertions = append(assertions, item)
			continue
		}

		out := make(map[string]interface{}, len(assertion))
		for key, val := range assertion {
			if key == "type" {
				if typeName, ok := val.(string); ok {
					out[key] = strings.ReplaceAll(typeName, "_", "-")
					continue
				}
			}
			out[key] = val
		}

		assertions = append(assertions, out)
	}

	return assertions
}

func convertProvidersFromPromptfoo(value interface{}) []interface{} {
	items, ok := value.([]interface{})
	if !ok {
		return nil
	}

	providers := make([]interface{}, 0, len(items))
	for _, item := range items {
		provider, ok := stringMap(item)
		if !ok {
			providers = append(providers, item)
			continue
		}

		out := make(map[string]interface{})
		if id, ok := provider["id"]; ok {
			out["id"] = id
		}

		if name, ok := provider["apiProvider"]; ok {
			out["name"] = name
		} else if name, ok := provider["type"]; ok {
			out["name"] = name
		}

		config := make(map[string]interface{})
		for key, val := range provider {
			switch key {
			case "id", "apiProvider", "type":
			default:
				config[key] = val
			}
		}
		if len(config) > 0 {
			out["config"] = config
		}

		providers = append(providers, out)
	}

	return providers
}

func convertProvidersToPromptfoo(value interface{}) []interface{} {
	items, ok := value.([]interface{})
	if !ok {
		return nil
	}

	providers := make([]interface{}, 0, len(items))
	for _, item := range items {
		provider, ok := stringMap(item)
		if !ok {
			providers = append(providers, item)
			continue
		}

		if _, ok := provider["apiProvider"]; ok {
			providers = append(providers, provider)
			continue
		}

		out := make(map[string]interface{})
		if id, ok := provider["id"]; ok {
			out["id"] = id
		}

		if apiProvider, ok := firstPresent(provider, "name", "type", "apiProvider"); ok {
			out["apiProvider"] = apiProvider
		}

		if config, ok := stringMap(provider["config"]); ok {
			for key, val := range config {
				out[key] = val
			}
		}
		for key, val := range provider {
			switch key {
			case "id", "name", "type", "apiProvider", "config":
			default:
				out[key] = val
			}
		}

		providers = append(providers, out)
	}

	return providers
}

func firstPresent(m map[string]interface{}, keys ...string) (interface{}, bool) {
	for _, key := range keys {
		if value, ok := m[key]; ok {
			return value, true
		}
	}
	return nil, false
}

func stringMap(value interface{}) (map[string]interface{}, bool) {
	m, ok := value.(map[string]interface{})
	return m, ok
}
