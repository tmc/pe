package structured

import (
	"fmt"
	"strings"
	"text/template"
)

// PromptBuilder helps construct prompts with structured output requirements
type PromptBuilder struct {
	formatter     *Formatter
	templates     map[string]*template.Template
	defaultFormat OutputFormat
}

// NewPromptBuilder creates a new prompt builder
func NewPromptBuilder() *PromptBuilder {
	return &PromptBuilder{
		formatter:     NewFormatter(),
		templates:     make(map[string]*template.Template),
		defaultFormat: FormatJSON,
	}
}

// SetDefaultFormat sets the default output format
func (pb *PromptBuilder) SetDefaultFormat(format OutputFormat) {
	pb.defaultFormat = format
}

// RegisterTemplate registers a prompt template
func (pb *PromptBuilder) RegisterTemplate(name string, tmpl string) error {
	t, err := template.New(name).Parse(tmpl)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}
	pb.templates[name] = t
	return nil
}

// BuildPrompt constructs a prompt with structured output instructions
func (pb *PromptBuilder) BuildPrompt(instruction string, schema *Schema, format OutputFormat) (string, error) {
	if format == "" {
		format = pb.defaultFormat
	}

	formatInstructions, err := pb.formatter.GetPromptInstructions(schema, format)
	if err != nil {
		return "", err
	}

	var prompt strings.Builder

	// Main instruction
	prompt.WriteString(instruction)
	prompt.WriteString("\n\n")

	// Format-specific instructions
	prompt.WriteString(formatInstructions)
	prompt.WriteString("\n\n")

	// Additional constraints
	if len(schema.Required) > 0 {
		prompt.WriteString("Required fields: ")
		prompt.WriteString(strings.Join(schema.Required, ", "))
		prompt.WriteString("\n\n")
	}

	// Validation rules
	validationRules := pb.generateValidationRules(schema)
	if validationRules != "" {
		prompt.WriteString("Validation rules:\n")
		prompt.WriteString(validationRules)
		prompt.WriteString("\n")
	}

	return prompt.String(), nil
}

// BuildFromTemplate builds a prompt using a registered template
func (pb *PromptBuilder) BuildFromTemplate(templateName string, data interface{}, schema *Schema, format OutputFormat) (string, error) {
	tmpl, exists := pb.templates[templateName]
	if !exists {
		return "", fmt.Errorf("template not found: %s", templateName)
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return pb.BuildPrompt(buf.String(), schema, format)
}

// BuildWithExamples builds a prompt with examples
func (pb *PromptBuilder) BuildWithExamples(instruction string, schema *Schema, format OutputFormat, examples []interface{}) (string, error) {
	basePrompt, err := pb.BuildPrompt(instruction, schema, format)
	if err != nil {
		return "", err
	}

	if len(examples) == 0 {
		return basePrompt, nil
	}

	var prompt strings.Builder
	prompt.WriteString(basePrompt)
	prompt.WriteString("\nExamples:\n")

	for i, example := range examples {
		formatted, err := pb.formatExample(example, format)
		if err != nil {
			continue
		}
		prompt.WriteString(fmt.Sprintf("\nExample %d:\n%s\n", i+1, formatted))
	}

	return prompt.String(), nil
}

// BuildChainOfThought builds a prompt for chain-of-thought with structured output
func (pb *PromptBuilder) BuildChainOfThought(instruction string, schema *Schema, format OutputFormat) (string, error) {
	cotInstruction := fmt.Sprintf(`%s

Please think through this step-by-step:
1. Analyze the requirements
2. Plan your approach
3. Generate the response

Show your reasoning, then provide the final answer in the specified format.

<reasoning>
[Your step-by-step reasoning here]
</reasoning>

<answer>`, instruction)

	return pb.BuildPrompt(cotInstruction, schema, format)
}

// BuildFunctionCall builds a prompt for function/tool calling with structured arguments
func (pb *PromptBuilder) BuildFunctionCall(functionName string, description string, paramSchema *Schema) (string, error) {
	instruction := fmt.Sprintf(`You need to call the function "%s".

Function description: %s

Generate the function call arguments:`, functionName, description)

	return pb.BuildPrompt(instruction, paramSchema, FormatJSON)
}

// BuildMultiStep builds a prompt for multi-step structured output
func (pb *PromptBuilder) BuildMultiStep(steps []StructuredStep) (string, error) {
	var prompt strings.Builder

	prompt.WriteString("Complete the following steps in order:\n\n")

	for i, step := range steps {
		prompt.WriteString(fmt.Sprintf("Step %d: %s\n", i+1, step.Description))

		if step.Schema != nil {
			formatInstructions, err := pb.formatter.GetPromptInstructions(step.Schema, step.Format)
			if err != nil {
				return "", err
			}
			prompt.WriteString(fmt.Sprintf("Output format for step %d:\n%s\n\n", i+1, formatInstructions))
		}
	}

	prompt.WriteString("Provide your response for each step clearly labeled.")

	return prompt.String(), nil
}

// StructuredStep represents a step in multi-step structured output
type StructuredStep struct {
	Description string
	Schema      *Schema
	Format      OutputFormat
}

// BuildConditional builds a prompt with conditional structured output
func (pb *PromptBuilder) BuildConditional(instruction string, conditions []ConditionalOutput) (string, error) {
	var prompt strings.Builder

	prompt.WriteString(instruction)
	prompt.WriteString("\n\nBased on your analysis, provide output in one of the following formats:\n\n")

	for _, condition := range conditions {
		prompt.WriteString(fmt.Sprintf("If %s:\n", condition.Condition))

		formatInstructions, err := pb.formatter.GetPromptInstructions(condition.Schema, condition.Format)
		if err != nil {
			return "", err
		}

		prompt.WriteString(formatInstructions)
		prompt.WriteString("\n\n")
	}

	return prompt.String(), nil
}

// ConditionalOutput represents a conditional output format
type ConditionalOutput struct {
	Condition string
	Schema    *Schema
	Format    OutputFormat
}

// Helper methods

func (pb *PromptBuilder) generateValidationRules(schema *Schema) string {
	var rules []string

	for name, prop := range schema.Properties {
		if prop.MinLength != nil {
			rules = append(rules, fmt.Sprintf("- %s: minimum length %d", name, *prop.MinLength))
		}
		if prop.MaxLength != nil {
			rules = append(rules, fmt.Sprintf("- %s: maximum length %d", name, *prop.MaxLength))
		}
		if prop.Pattern != "" {
			rules = append(rules, fmt.Sprintf("- %s: must match pattern %s", name, prop.Pattern))
		}
		if len(prop.Enum) > 0 {
			enumStr := make([]string, len(prop.Enum))
			for i, v := range prop.Enum {
				enumStr[i] = fmt.Sprintf("%v", v)
			}
			rules = append(rules, fmt.Sprintf("- %s: must be one of [%s]", name, strings.Join(enumStr, ", ")))
		}
	}

	return strings.Join(rules, "\n")
}

func (pb *PromptBuilder) formatExample(example interface{}, format OutputFormat) (string, error) {
	// Convert example to appropriate format
	switch format {
	case FormatJSON:
		return formatAsJSON(example)
	case FormatYAML:
		return formatAsYAML(example)
	default:
		return fmt.Sprintf("%v", example), nil
	}
}

func formatAsJSON(data interface{}) (string, error) {
	// Implementation would use json.MarshalIndent
	return fmt.Sprintf("%v", data), nil
}

func formatAsYAML(data interface{}) (string, error) {
	// Implementation would use yaml.Marshal
	return fmt.Sprintf("%v", data), nil
}

// CommonSchemas provides pre-built schemas for common use cases
var CommonSchemas = struct {
	// Analysis represents a structured analysis output
	Analysis *Schema

	// CodeGeneration represents code generation output
	CodeGeneration *Schema

	// DataExtraction represents data extraction output
	DataExtraction *Schema

	// Classification represents classification output
	Classification *Schema

	// Summary represents summarization output
	Summary *Schema
}{
	Analysis: &Schema{
		Name:        "analysis",
		Description: "Structured analysis output",
		Type:        "object",
		Properties: map[string]*Property{
			"topic": {
				Type:        "string",
				Description: "The main topic or subject of analysis",
			},
			"findings": {
				Type:        "array",
				Description: "Key findings from the analysis",
				Items: &Property{
					Type: "object",
					Properties: map[string]*Property{
						"point": {
							Type:        "string",
							Description: "The finding or observation",
						},
						"evidence": {
							Type:        "string",
							Description: "Supporting evidence",
						},
						"confidence": {
							Type:        "number",
							Description: "Confidence level (0-1)",
							Minimum:     &[]float64{0}[0],
							Maximum:     &[]float64{1}[0],
						},
					},
				},
			},
			"conclusion": {
				Type:        "string",
				Description: "Overall conclusion",
			},
			"recommendations": {
				Type:        "array",
				Description: "Actionable recommendations",
				Items: &Property{
					Type: "string",
				},
			},
		},
		Required: []string{"topic", "findings", "conclusion"},
	},

	CodeGeneration: &Schema{
		Name:        "code_generation",
		Description: "Generated code output",
		Type:        "object",
		Properties: map[string]*Property{
			"language": {
				Type:        "string",
				Description: "Programming language",
			},
			"code": {
				Type:        "string",
				Description: "The generated code",
			},
			"explanation": {
				Type:        "string",
				Description: "Explanation of the code",
			},
			"dependencies": {
				Type:        "array",
				Description: "Required dependencies",
				Items: &Property{
					Type: "string",
				},
			},
			"usage_example": {
				Type:        "string",
				Description: "Example of how to use the code",
			},
		},
		Required: []string{"language", "code"},
	},

	DataExtraction: &Schema{
		Name:        "data_extraction",
		Description: "Extracted data output",
		Type:        "object",
		Properties: map[string]*Property{
			"source": {
				Type:        "string",
				Description: "Data source identifier",
			},
			"extracted_data": {
				Type:        "array",
				Description: "Extracted data items",
				Items: &Property{
					Type: "object",
					Properties: map[string]*Property{
						"field": {
							Type: "string",
						},
						"value": {
							Type: "string",
						},
						"confidence": {
							Type:    "number",
							Minimum: &[]float64{0}[0],
							Maximum: &[]float64{1}[0],
						},
					},
				},
			},
			"metadata": {
				Type:        "object",
				Description: "Additional metadata",
			},
		},
		Required: []string{"extracted_data"},
	},

	Classification: &Schema{
		Name:        "classification",
		Description: "Classification output",
		Type:        "object",
		Properties: map[string]*Property{
			"input": {
				Type:        "string",
				Description: "The input that was classified",
			},
			"category": {
				Type:        "string",
				Description: "Primary category",
			},
			"confidence": {
				Type:        "number",
				Description: "Classification confidence",
				Minimum:     &[]float64{0}[0],
				Maximum:     &[]float64{1}[0],
			},
			"alternative_categories": {
				Type:        "array",
				Description: "Alternative possible categories",
				Items: &Property{
					Type: "object",
					Properties: map[string]*Property{
						"category": {
							Type: "string",
						},
						"confidence": {
							Type: "number",
						},
					},
				},
			},
			"reasoning": {
				Type:        "string",
				Description: "Explanation for the classification",
			},
		},
		Required: []string{"category", "confidence"},
	},

	Summary: &Schema{
		Name:        "summary",
		Description: "Summarization output",
		Type:        "object",
		Properties: map[string]*Property{
			"title": {
				Type:        "string",
				Description: "Summary title",
			},
			"summary": {
				Type:        "string",
				Description: "Main summary text",
			},
			"key_points": {
				Type:        "array",
				Description: "Key points or highlights",
				Items: &Property{
					Type: "string",
				},
			},
			"length": {
				Type:        "integer",
				Description: "Summary length in words",
			},
			"compression_ratio": {
				Type:        "number",
				Description: "Ratio of summary to original length",
			},
		},
		Required: []string{"summary", "key_points"},
	},
}
