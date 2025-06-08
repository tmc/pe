package structured

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"sigs.k8s.io/yaml"
)

// OutputFormat represents different structured output formats
type OutputFormat string

const (
	FormatJSON       OutputFormat = "json"
	FormatYAML       OutputFormat = "yaml"
	FormatXML        OutputFormat = "xml"
	FormatTOML       OutputFormat = "toml"
	FormatMarkdown   OutputFormat = "markdown"
	FormatCSV        OutputFormat = "csv"
	FormatJSONSchema OutputFormat = "json-schema"
	FormatTypeScript OutputFormat = "typescript"
	FormatPydantic   OutputFormat = "pydantic"
	FormatProtobuf   OutputFormat = "protobuf"
	FormatCustom     OutputFormat = "custom"
)

// Schema represents a structured output schema
type Schema struct {
	ID          string                 `json:"id,omitempty" yaml:"id,omitempty"`
	Name        string                 `json:"name" yaml:"name"`
	Description string                 `json:"description,omitempty" yaml:"description,omitempty"`
	Type        string                 `json:"type" yaml:"type"`
	Properties  map[string]*Property   `json:"properties,omitempty" yaml:"properties,omitempty"`
	Required    []string               `json:"required,omitempty" yaml:"required,omitempty"`
	Examples    []interface{}          `json:"examples,omitempty" yaml:"examples,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty" yaml:"metadata,omitempty"`
	Version     string                 `json:"version,omitempty" yaml:"version,omitempty"`
}

// Property represents a property in a schema
type Property struct {
	Type        string               `json:"type" yaml:"type"`
	Description string               `json:"description,omitempty" yaml:"description,omitempty"`
	Format      string               `json:"format,omitempty" yaml:"format,omitempty"`
	Enum        []interface{}        `json:"enum,omitempty" yaml:"enum,omitempty"`
	Pattern     string               `json:"pattern,omitempty" yaml:"pattern,omitempty"`
	MinLength   *int                 `json:"minLength,omitempty" yaml:"minLength,omitempty"`
	MaxLength   *int                 `json:"maxLength,omitempty" yaml:"maxLength,omitempty"`
	Minimum     *float64             `json:"minimum,omitempty" yaml:"minimum,omitempty"`
	Maximum     *float64             `json:"maximum,omitempty" yaml:"maximum,omitempty"`
	Items       *Property            `json:"items,omitempty" yaml:"items,omitempty"`
	Properties  map[string]*Property `json:"properties,omitempty" yaml:"properties,omitempty"`
	Required    []string             `json:"required,omitempty" yaml:"required,omitempty"`
	Default     interface{}          `json:"default,omitempty" yaml:"default,omitempty"`
	Examples    []interface{}        `json:"examples,omitempty" yaml:"examples,omitempty"`
}

// Formatter handles conversion between different structured formats
type Formatter struct {
	plugins map[OutputFormat]FormatterPlugin
}

// FormatterPlugin defines the interface for format plugins
type FormatterPlugin interface {
	// Format converts a schema to the target format
	Format(schema *Schema) (string, error)

	// Parse parses content in the format to extract structured data
	Parse(content string) (map[string]interface{}, error)

	// Validate checks if content matches the expected format
	Validate(content string, schema *Schema) error

	// GetPromptInstructions returns instructions for generating this format
	GetPromptInstructions(schema *Schema) string
}

// NewFormatter creates a new formatter with built-in plugins
func NewFormatter() *Formatter {
	f := &Formatter{
		plugins: make(map[OutputFormat]FormatterPlugin),
	}

	// Register built-in plugins
	f.RegisterPlugin(FormatJSON, &JSONPlugin{})
	f.RegisterPlugin(FormatYAML, &YAMLPlugin{})
	f.RegisterPlugin(FormatMarkdown, &MarkdownPlugin{})
	f.RegisterPlugin(FormatJSONSchema, &JSONSchemaPlugin{})
	f.RegisterPlugin(FormatTypeScript, &TypeScriptPlugin{})
	f.RegisterPlugin(FormatPydantic, &PydanticPlugin{})

	return f
}

// RegisterPlugin registers a custom formatter plugin
func (f *Formatter) RegisterPlugin(format OutputFormat, plugin FormatterPlugin) {
	f.plugins[format] = plugin
}

// Format converts a schema to the specified format
func (f *Formatter) Format(schema *Schema, format OutputFormat) (string, error) {
	plugin, exists := f.plugins[format]
	if !exists {
		return "", fmt.Errorf("unsupported format: %s", format)
	}

	return plugin.Format(schema)
}

// GetPromptInstructions returns prompt instructions for a format
func (f *Formatter) GetPromptInstructions(schema *Schema, format OutputFormat) (string, error) {
	plugin, exists := f.plugins[format]
	if !exists {
		return "", fmt.Errorf("unsupported format: %s", format)
	}

	return plugin.GetPromptInstructions(schema), nil
}

// Validate validates content against a schema in the specified format
func (f *Formatter) Validate(content string, schema *Schema, format OutputFormat) error {
	plugin, exists := f.plugins[format]
	if !exists {
		return fmt.Errorf("unsupported format: %s", format)
	}

	return plugin.Validate(content, schema)
}

// JSONPlugin handles JSON formatting
type JSONPlugin struct{}

func (p *JSONPlugin) Format(schema *Schema) (string, error) {
	example := p.generateExample(schema)
	data, err := json.MarshalIndent(example, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (p *JSONPlugin) Parse(content string) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := json.Unmarshal([]byte(content), &result)
	return result, err
}

func (p *JSONPlugin) Validate(content string, schema *Schema) error {
	data, err := p.Parse(content)
	if err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}

	return validateAgainstSchema(data, schema)
}

func (p *JSONPlugin) GetPromptInstructions(schema *Schema) string {
	example, _ := p.Format(schema)
	return fmt.Sprintf(`Generate a JSON response with the following structure:

%s

Ensure all required fields are present and follow the exact format shown.`, example)
}

func (p *JSONPlugin) generateExample(schema *Schema) map[string]interface{} {
	if len(schema.Examples) > 0 {
		if example, ok := schema.Examples[0].(map[string]interface{}); ok {
			return example
		}
	}

	example := make(map[string]interface{})
	for name, prop := range schema.Properties {
		example[name] = generatePropertyExample(prop)
	}
	return example
}

// YAMLPlugin handles YAML formatting
type YAMLPlugin struct{}

func (p *YAMLPlugin) Format(schema *Schema) (string, error) {
	example := generateSchemaExample(schema)
	data, err := yaml.Marshal(example)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (p *YAMLPlugin) Parse(content string) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := yaml.Unmarshal([]byte(content), &result)
	return result, err
}

func (p *YAMLPlugin) Validate(content string, schema *Schema) error {
	data, err := p.Parse(content)
	if err != nil {
		return fmt.Errorf("invalid YAML: %w", err)
	}

	return validateAgainstSchema(data, schema)
}

func (p *YAMLPlugin) GetPromptInstructions(schema *Schema) string {
	example, _ := p.Format(schema)
	return fmt.Sprintf(`Generate a YAML response with the following structure:

%s

Use proper YAML syntax with correct indentation.`, example)
}

// MarkdownPlugin handles Markdown table formatting
type MarkdownPlugin struct{}

func (p *MarkdownPlugin) Format(schema *Schema) (string, error) {
	var sb strings.Builder

	// Generate header
	sb.WriteString("| Field | Type | Required | Description |\n")
	sb.WriteString("|-------|------|----------|-------------|\n")

	// Generate rows
	for name, prop := range schema.Properties {
		required := "No"
		for _, req := range schema.Required {
			if req == name {
				required = "Yes"
				break
			}
		}

		sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s |\n",
			name, prop.Type, required, prop.Description))
	}

	return sb.String(), nil
}

func (p *MarkdownPlugin) Parse(content string) (map[string]interface{}, error) {
	// Simple markdown table parser
	lines := strings.Split(content, "\n")
	if len(lines) < 3 {
		return nil, fmt.Errorf("invalid markdown table")
	}

	// Extract headers
	headers := strings.Split(lines[0], "|")
	headers = headers[1 : len(headers)-1] // Remove empty first and last
	for i := range headers {
		headers[i] = strings.TrimSpace(headers[i])
	}

	// Parse data rows
	result := make(map[string]interface{})
	for i := 2; i < len(lines); i++ {
		if lines[i] == "" {
			continue
		}

		cells := strings.Split(lines[i], "|")
		if len(cells) < 3 {
			continue
		}
		cells = cells[1 : len(cells)-1]

		if len(cells) >= 2 {
			field := strings.TrimSpace(cells[0])
			value := strings.TrimSpace(cells[1])
			result[field] = value
		}
	}

	return result, nil
}

func (p *MarkdownPlugin) Validate(content string, schema *Schema) error {
	_, err := p.Parse(content)
	return err
}

func (p *MarkdownPlugin) GetPromptInstructions(schema *Schema) string {
	example, _ := p.Format(schema)
	return fmt.Sprintf(`Generate a markdown table with the following structure:

%s

Fill in appropriate values for each field.`, example)
}

// JSONSchemaPlugin handles JSON Schema generation
type JSONSchemaPlugin struct{}

func (p *JSONSchemaPlugin) Format(schema *Schema) (string, error) {
	jsonSchema := map[string]interface{}{
		"$schema":     "http://json-schema.org/draft-07/schema#",
		"type":        "object",
		"title":       schema.Name,
		"description": schema.Description,
		"properties":  convertPropertiesToJSONSchema(schema.Properties),
		"required":    schema.Required,
	}

	data, err := json.MarshalIndent(jsonSchema, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (p *JSONSchemaPlugin) Parse(content string) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := json.Unmarshal([]byte(content), &result)
	return result, err
}

func (p *JSONSchemaPlugin) Validate(content string, schema *Schema) error {
	// JSON Schema validation would go here
	return nil
}

func (p *JSONSchemaPlugin) GetPromptInstructions(schema *Schema) string {
	return fmt.Sprintf(`Generate output that conforms to the following JSON Schema:

%s

Ensure the output validates against this schema.`, schema.Name)
}

// TypeScriptPlugin handles TypeScript interface generation
type TypeScriptPlugin struct{}

func (p *TypeScriptPlugin) Format(schema *Schema) (string, error) {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("interface %s {\n", toPascalCase(schema.Name)))

	for name, prop := range schema.Properties {
		required := true
		for _, req := range schema.Required {
			if req == name {
				required = false
				break
			}
		}

		optional := ""
		if !required {
			optional = "?"
		}

		tsType := goTypeToTypeScript(prop.Type)
		sb.WriteString(fmt.Sprintf("  %s%s: %s;\n", name, optional, tsType))
	}

	sb.WriteString("}")

	return sb.String(), nil
}

func (p *TypeScriptPlugin) Parse(content string) (map[string]interface{}, error) {
	// TypeScript parsing would require more complex logic
	return nil, fmt.Errorf("TypeScript parsing not implemented")
}

func (p *TypeScriptPlugin) Validate(content string, schema *Schema) error {
	return fmt.Errorf("TypeScript validation not implemented")
}

func (p *TypeScriptPlugin) GetPromptInstructions(schema *Schema) string {
	tsInterface, _ := p.Format(schema)
	return fmt.Sprintf(`Generate a TypeScript object that implements this interface:

%s

Use proper TypeScript syntax.`, tsInterface)
}

// PydanticPlugin handles Pydantic model generation
type PydanticPlugin struct{}

func (p *PydanticPlugin) Format(schema *Schema) (string, error) {
	var sb strings.Builder

	sb.WriteString("from pydantic import BaseModel, Field\n")
	sb.WriteString("from typing import Optional, List, Dict, Any\n\n")

	sb.WriteString(fmt.Sprintf("class %s(BaseModel):\n", toPascalCase(schema.Name)))

	if schema.Description != "" {
		sb.WriteString(fmt.Sprintf(`    """%s"""`+"\n", schema.Description))
	}

	for name, prop := range schema.Properties {
		required := false
		for _, req := range schema.Required {
			if req == name {
				required = true
				break
			}
		}

		pyType := goTypeToPython(prop.Type)
		if !required {
			pyType = fmt.Sprintf("Optional[%s]", pyType)
		}

		fieldDef := fmt.Sprintf("    %s: %s", toSnakeCase(name), pyType)

		if prop.Description != "" {
			fieldDef += fmt.Sprintf(` = Field(description="%s")`, prop.Description)
		} else if prop.Default != nil {
			fieldDef += fmt.Sprintf(" = %v", prop.Default)
		} else if !required {
			fieldDef += " = None"
		}

		sb.WriteString(fieldDef + "\n")
	}

	return sb.String(), nil
}

func (p *PydanticPlugin) Parse(content string) (map[string]interface{}, error) {
	// Python parsing would require executing Python code
	return nil, fmt.Errorf("Pydantic parsing not implemented")
}

func (p *PydanticPlugin) Validate(content string, schema *Schema) error {
	return fmt.Errorf("Pydantic validation not implemented")
}

func (p *PydanticPlugin) GetPromptInstructions(schema *Schema) string {
	model, _ := p.Format(schema)
	return fmt.Sprintf(`Generate a Python dictionary that matches this Pydantic model:

%s

Return only the dictionary data, not the model definition.`, model)
}

// Helper functions

func generateSchemaExample(schema *Schema) map[string]interface{} {
	if len(schema.Examples) > 0 {
		if example, ok := schema.Examples[0].(map[string]interface{}); ok {
			return example
		}
	}

	example := make(map[string]interface{})
	for name, prop := range schema.Properties {
		example[name] = generatePropertyExample(prop)
	}
	return example
}

func generatePropertyExample(prop *Property) interface{} {
	if prop.Default != nil {
		return prop.Default
	}

	if len(prop.Examples) > 0 {
		return prop.Examples[0]
	}

	switch prop.Type {
	case "string":
		if len(prop.Enum) > 0 {
			return prop.Enum[0]
		}
		return "example_string"
	case "number", "integer":
		return 42
	case "boolean":
		return true
	case "array":
		if prop.Items != nil {
			return []interface{}{generatePropertyExample(prop.Items)}
		}
		return []interface{}{}
	case "object":
		if prop.Properties != nil {
			obj := make(map[string]interface{})
			for name, subProp := range prop.Properties {
				obj[name] = generatePropertyExample(subProp)
			}
			return obj
		}
		return map[string]interface{}{}
	default:
		return nil
	}
}

func validateAgainstSchema(data map[string]interface{}, schema *Schema) error {
	// Check required fields
	for _, required := range schema.Required {
		if _, exists := data[required]; !exists {
			return fmt.Errorf("missing required field: %s", required)
		}
	}

	// Validate each property
	for name, value := range data {
		prop, exists := schema.Properties[name]
		if !exists {
			continue // Allow additional properties for now
		}

		if err := validateProperty(name, value, prop); err != nil {
			return err
		}
	}

	return nil
}

func validateProperty(name string, value interface{}, prop *Property) error {
	// Type checking
	valueType := reflect.TypeOf(value).Kind()

	switch prop.Type {
	case "string":
		if valueType != reflect.String {
			return fmt.Errorf("field %s: expected string, got %T", name, value)
		}

		str := value.(string)
		if prop.MinLength != nil && len(str) < *prop.MinLength {
			return fmt.Errorf("field %s: string too short (min %d)", name, *prop.MinLength)
		}
		if prop.MaxLength != nil && len(str) > *prop.MaxLength {
			return fmt.Errorf("field %s: string too long (max %d)", name, *prop.MaxLength)
		}
		if prop.Pattern != "" {
			if matched, _ := regexp.MatchString(prop.Pattern, str); !matched {
				return fmt.Errorf("field %s: does not match pattern %s", name, prop.Pattern)
			}
		}

	case "number", "integer":
		if valueType != reflect.Float64 && valueType != reflect.Int {
			return fmt.Errorf("field %s: expected number, got %T", name, value)
		}

	case "boolean":
		if valueType != reflect.Bool {
			return fmt.Errorf("field %s: expected boolean, got %T", name, value)
		}

	case "array":
		if valueType != reflect.Slice {
			return fmt.Errorf("field %s: expected array, got %T", name, value)
		}

	case "object":
		if valueType != reflect.Map {
			return fmt.Errorf("field %s: expected object, got %T", name, value)
		}
	}

	// Enum validation
	if len(prop.Enum) > 0 {
		found := false
		for _, enumVal := range prop.Enum {
			if reflect.DeepEqual(value, enumVal) {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("field %s: value not in enum", name)
		}
	}

	return nil
}

func convertPropertiesToJSONSchema(props map[string]*Property) map[string]interface{} {
	result := make(map[string]interface{})

	for name, prop := range props {
		jsonProp := map[string]interface{}{
			"type": prop.Type,
		}

		if prop.Description != "" {
			jsonProp["description"] = prop.Description
		}
		if prop.Format != "" {
			jsonProp["format"] = prop.Format
		}
		if len(prop.Enum) > 0 {
			jsonProp["enum"] = prop.Enum
		}
		if prop.Pattern != "" {
			jsonProp["pattern"] = prop.Pattern
		}
		if prop.MinLength != nil {
			jsonProp["minLength"] = *prop.MinLength
		}
		if prop.MaxLength != nil {
			jsonProp["maxLength"] = *prop.MaxLength
		}
		if prop.Minimum != nil {
			jsonProp["minimum"] = *prop.Minimum
		}
		if prop.Maximum != nil {
			jsonProp["maximum"] = *prop.Maximum
		}
		if prop.Items != nil {
			jsonProp["items"] = convertPropertyToJSONSchema(prop.Items)
		}
		if prop.Properties != nil {
			jsonProp["properties"] = convertPropertiesToJSONSchema(prop.Properties)
		}
		if len(prop.Required) > 0 {
			jsonProp["required"] = prop.Required
		}

		result[name] = jsonProp
	}

	return result
}

func convertPropertyToJSONSchema(prop *Property) map[string]interface{} {
	return map[string]interface{}{
		"type": prop.Type,
	}
}

func toPascalCase(s string) string {
	words := strings.Split(s, "_")
	for i := range words {
		words[i] = strings.Title(words[i])
	}
	return strings.Join(words, "")
}

func toSnakeCase(s string) string {
	var result []rune
	for i, r := range s {
		if i > 0 && 'A' <= r && r <= 'Z' {
			result = append(result, '_')
		}
		result = append(result, []rune(strings.ToLower(string(r)))...)
	}
	return string(result)
}

func goTypeToTypeScript(goType string) string {
	switch goType {
	case "string":
		return "string"
	case "integer", "number":
		return "number"
	case "boolean":
		return "boolean"
	case "array":
		return "any[]"
	case "object":
		return "Record<string, any>"
	default:
		return "any"
	}
}

func goTypeToPython(goType string) string {
	switch goType {
	case "string":
		return "str"
	case "integer":
		return "int"
	case "number":
		return "float"
	case "boolean":
		return "bool"
	case "array":
		return "List[Any]"
	case "object":
		return "Dict[str, Any]"
	default:
		return "Any"
	}
}
