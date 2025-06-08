package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tmc/pe/internal/structured"
	"sigs.k8s.io/yaml"
)

var structuredCmd = &cobra.Command{
	Use:   "structured",
	Short: "Work with structured output schemas and formatting",
	Long: `Generate, validate, and convert structured output schemas for prompts.
	
This command helps you define structured output requirements for your prompts
and convert between different schema formats like JSON Schema, TypeScript,
Pydantic, and more.`,
}

var (
	structuredFormat     string
	structuredOutput     string
	structuredValidate   bool
	structuredExample    bool
	structuredPrompt     bool
	structuredSchemaFile string
)

func init() {
	// Add subcommands
	structuredCmd.AddCommand(structuredGenerateCmd)
	structuredCmd.AddCommand(structuredConvertCmd)
	structuredCmd.AddCommand(structuredValidateCmd)
	structuredCmd.AddCommand(structuredPromptCmd)

	// Common flags
	structuredCmd.PersistentFlags().StringVarP(&structuredFormat, "format", "f", "json", "Output format")
	structuredCmd.PersistentFlags().StringVarP(&structuredOutput, "output", "o", "", "Output file (default: stdout)")
}

var structuredGenerateCmd = &cobra.Command{
	Use:   "generate [type]",
	Short: "Generate a structured output schema",
	Long: `Generate a structured output schema for common use cases.
	
Available types:
  - analysis: Structured analysis output
  - code: Code generation output  
  - data: Data extraction output
  - classification: Classification output
  - summary: Summarization output`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		schemaType := args[0]

		var schema *structured.Schema
		switch schemaType {
		case "analysis":
			schema = structured.CommonSchemas.Analysis
		case "code":
			schema = structured.CommonSchemas.CodeGeneration
		case "data":
			schema = structured.CommonSchemas.DataExtraction
		case "classification":
			schema = structured.CommonSchemas.Classification
		case "summary":
			schema = structured.CommonSchemas.Summary
		default:
			return fmt.Errorf("unknown schema type: %s", schemaType)
		}

		// Format the schema
		formatter := structured.NewFormatter()
		formatted, err := formatter.Format(schema, structured.OutputFormat(structuredFormat))
		if err != nil {
			return fmt.Errorf("failed to format schema: %w", err)
		}

		// Output
		if structuredOutput != "" {
			return os.WriteFile(structuredOutput, []byte(formatted), 0644)
		}
		fmt.Println(formatted)
		return nil
	},
}

var structuredConvertCmd = &cobra.Command{
	Use:   "convert [schema-file]",
	Short: "Convert a schema between formats",
	Long: `Convert a structured output schema between different formats.
	
Supported formats:
  - json: JSON example
  - yaml: YAML example
  - json-schema: JSON Schema
  - typescript: TypeScript interface
  - pydantic: Pydantic model
  - markdown: Markdown table`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Read schema file
		data, err := os.ReadFile(args[0])
		if err != nil {
			return fmt.Errorf("failed to read schema file: %w", err)
		}

		// Parse schema
		var schema structured.Schema
		if strings.HasSuffix(args[0], ".yaml") || strings.HasSuffix(args[0], ".yml") {
			err = yaml.Unmarshal(data, &schema)
		} else {
			err = json.Unmarshal(data, &schema)
		}
		if err != nil {
			return fmt.Errorf("failed to parse schema: %w", err)
		}

		// Convert to target format
		formatter := structured.NewFormatter()
		formatted, err := formatter.Format(&schema, structured.OutputFormat(structuredFormat))
		if err != nil {
			return fmt.Errorf("failed to format schema: %w", err)
		}

		// Output
		if structuredOutput != "" {
			return os.WriteFile(structuredOutput, []byte(formatted), 0644)
		}
		fmt.Println(formatted)
		return nil
	},
}

var structuredValidateCmd = &cobra.Command{
	Use:   "validate [data-file] [schema-file]",
	Short: "Validate data against a schema",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Read data file
		data, err := os.ReadFile(args[0])
		if err != nil {
			return fmt.Errorf("failed to read data file: %w", err)
		}

		// Read schema file
		schemaData, err := os.ReadFile(args[1])
		if err != nil {
			return fmt.Errorf("failed to read schema file: %w", err)
		}

		// Parse schema
		var schema structured.Schema
		if err := json.Unmarshal(schemaData, &schema); err != nil {
			return fmt.Errorf("failed to parse schema: %w", err)
		}

		// Get formatter and plugin
		formatter := structured.NewFormatter()

		// Try to parse and validate the data
		var parsedData map[string]interface{}
		if strings.HasSuffix(args[0], ".yaml") || strings.HasSuffix(args[0], ".yml") {
			err = yaml.Unmarshal(data, &parsedData)
		} else {
			err = json.Unmarshal(data, &parsedData)
		}
		if err != nil {
			return fmt.Errorf("failed to parse data: %w", err)
		}

		// Validate against schema
		if err := formatter.Validate(string(data), &schema, structured.FormatJSON); err != nil {
			return fmt.Errorf("validation failed: %w", err)
		}

		fmt.Println("✓ Data validates against schema")
		return nil
	},
}

var structuredPromptCmd = &cobra.Command{
	Use:   "prompt [instruction] [schema-file]",
	Short: "Generate a prompt with structured output instructions",
	Long: `Generate a prompt that includes instructions for structured output.
	
This command combines your instruction with schema-based formatting guidelines
to create a complete prompt that will produce structured output.`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		instruction := args[0]

		var schema *structured.Schema

		if len(args) > 1 {
			// Read schema file
			data, err := os.ReadFile(args[1])
			if err != nil {
				return fmt.Errorf("failed to read schema file: %w", err)
			}

			schema = &structured.Schema{}
			if err := json.Unmarshal(data, schema); err != nil {
				return fmt.Errorf("failed to parse schema: %w", err)
			}
		} else {
			// Use a default schema
			schema = &structured.Schema{
				Name: "output",
				Type: "object",
				Properties: map[string]*structured.Property{
					"response": {
						Type:        "string",
						Description: "The response to the instruction",
					},
				},
				Required: []string{"response"},
			}
		}

		// Build the prompt
		builder := structured.NewPromptBuilder()
		prompt, err := builder.BuildPrompt(instruction, schema, structured.OutputFormat(structuredFormat))
		if err != nil {
			return fmt.Errorf("failed to build prompt: %w", err)
		}

		// Output
		if structuredOutput != "" {
			return os.WriteFile(structuredOutput, []byte(prompt), 0644)
		}
		fmt.Println(prompt)
		return nil
	},
}

// Add a subcommand for managing custom formatter plugins
var structuredPluginCmd = &cobra.Command{
	Use:   "plugin",
	Short: "Manage structured output formatter plugins",
	Long: `Manage custom formatter plugins for structured output.
	
Plugins allow you to extend PE with custom output formats and validation logic.`,
}

func init() {
	structuredCmd.AddCommand(structuredPluginCmd)

	// Plugin subcommands
	structuredPluginCmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List available formatter plugins",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Available formatter plugins:")
			for _, format := range []structured.OutputFormat{
				structured.FormatJSON,
				structured.FormatYAML,
				structured.FormatMarkdown,
				structured.FormatJSONSchema,
				structured.FormatTypeScript,
				structured.FormatPydantic,
			} {
				fmt.Printf("  - %s\n", format)
			}

			return nil
		},
	})
}
