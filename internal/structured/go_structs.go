package structured

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// parseFloat is a helper to parse string to float64
func parseFloat(s string) (float64, error) {
	return strconv.ParseFloat(s, 64)
}

// FromGoStruct creates a Schema from a Go struct using reflection
func FromGoStruct(v interface{}) (*Schema, error) {
	typ := reflect.TypeOf(v)
	
	// Handle pointer types
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	
	if typ.Kind() != reflect.Struct {
		return nil, fmt.Errorf("expected struct type, got %s", typ.Kind())
	}
	
	schema := &Schema{
		Name:        toSnakeCase(typ.Name()),
		Type:        "object",
		Properties:  make(map[string]*Property),
		Required:    []string{},
		Description: getStructTag(typ, "description"),
	}
	
	// Process struct fields
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		
		// Skip unexported fields
		if !field.IsExported() {
			continue
		}
		
		// Get field name from json tag or use field name
		jsonTag := field.Tag.Get("json")
		if jsonTag == "-" {
			continue
		}
		
		fieldName := field.Name
		if jsonTag != "" {
			parts := strings.Split(jsonTag, ",")
			if parts[0] != "" {
				fieldName = parts[0]
			}
			
			// Check for omitempty
			omitempty := false
			for _, part := range parts[1:] {
				if part == "omitempty" {
					omitempty = true
					break
				}
			}
			
			// If not omitempty, it's required
			if !omitempty {
				schema.Required = append(schema.Required, fieldName)
			}
		}
		
		// Create property from field type
		prop, err := createPropertyFromType(field.Type, field.Tag)
		if err != nil {
			return nil, fmt.Errorf("failed to process field %s: %w", field.Name, err)
		}
		
		schema.Properties[fieldName] = prop
	}
	
	return schema, nil
}

// createPropertyFromType creates a Property from a reflect.Type
func createPropertyFromType(typ reflect.Type, tag reflect.StructTag) (*Property, error) {
	prop := &Property{
		Description: tag.Get("description"),
		Format:      tag.Get("format"),
	}
	
	// Handle validation tags
	if minStr := tag.Get("min"); minStr != "" {
		if val, err := parseFloat(minStr); err == nil {
			prop.Minimum = &val
		}
	}
	
	if maxStr := tag.Get("max"); maxStr != "" {
		if val, err := parseFloat(maxStr); err == nil {
			prop.Maximum = &val
		}
	}
	
	if minLenStr := tag.Get("minLength"); minLenStr != "" {
		if val, err := parseInt(minLenStr); err == nil {
			prop.MinLength = &val
		}
	}
	
	if maxLenStr := tag.Get("maxLength"); maxLenStr != "" {
		if val, err := parseInt(maxLenStr); err == nil {
			prop.MaxLength = &val
		}
	}
	
	if pattern := tag.Get("pattern"); pattern != "" {
		prop.Pattern = pattern
	}
	
	if enumStr := tag.Get("enum"); enumStr != "" {
		// Parse enum values
		enumVals := strings.Split(enumStr, "|")
		prop.Enum = make([]interface{}, len(enumVals))
		for i, v := range enumVals {
			prop.Enum[i] = strings.TrimSpace(v)
		}
	}
	
	if defaultStr := tag.Get("default"); defaultStr != "" {
		prop.Default = parseDefaultValue(defaultStr, typ)
	}
	
	if exampleStr := tag.Get("example"); exampleStr != "" {
		prop.Examples = []interface{}{parseDefaultValue(exampleStr, typ)}
	}
	
	// Determine type from Go type
	switch typ.Kind() {
	case reflect.String:
		prop.Type = "string"
		
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		prop.Type = "integer"
		
	case reflect.Float32, reflect.Float64:
		prop.Type = "number"
		
	case reflect.Bool:
		prop.Type = "boolean"
		
	case reflect.Slice, reflect.Array:
		prop.Type = "array"
		itemProp, err := createPropertyFromType(typ.Elem(), reflect.StructTag(""))
		if err != nil {
			return nil, fmt.Errorf("failed to process array items: %w", err)
		}
		prop.Items = itemProp
		
	case reflect.Map:
		prop.Type = "object"
		// For maps, we can't determine the property schema
		
	case reflect.Struct:
		// Handle time.Time specially
		if typ.String() == "time.Time" {
			prop.Type = "string"
			prop.Format = "date-time"
		} else {
			// Nested struct
			prop.Type = "object"
			prop.Properties = make(map[string]*Property)
			prop.Required = []string{}
			
			for i := 0; i < typ.NumField(); i++ {
				field := typ.Field(i)
				if !field.IsExported() {
					continue
				}
				
				jsonTag := field.Tag.Get("json")
				if jsonTag == "-" {
					continue
				}
				
				fieldName := field.Name
				if jsonTag != "" {
					parts := strings.Split(jsonTag, ",")
					if parts[0] != "" {
						fieldName = parts[0]
					}
				}
				
				fieldProp, err := createPropertyFromType(field.Type, field.Tag)
				if err != nil {
					return nil, err
				}
				
				prop.Properties[fieldName] = fieldProp
			}
		}
		
	case reflect.Ptr:
		// Handle pointer types
		return createPropertyFromType(typ.Elem(), tag)
		
	case reflect.Interface:
		// Interface{} maps to any
		prop.Type = "object"
		
	default:
		return nil, fmt.Errorf("unsupported type: %s", typ.Kind())
	}
	
	return prop, nil
}

// StructuredPromptOptions provides options for generating prompts from structs
type StructuredPromptOptions struct {
	Format      OutputFormat
	Examples    []interface{}
	Instruction string
	Style       string // "simple", "detailed", "chain-of-thought"
}

// GeneratePromptFromStruct generates a prompt with structured output instructions from a Go struct
func GeneratePromptFromStruct(v interface{}, opts StructuredPromptOptions) (string, error) {
	schema, err := FromGoStruct(v)
	if err != nil {
		return "", fmt.Errorf("failed to generate schema from struct: %w", err)
	}
	
	builder := NewPromptBuilder()
	
	if opts.Format == "" {
		opts.Format = FormatJSON
	}
	
	// Add examples if the struct has example values
	if opts.Examples == nil && hasExampleValues(v) {
		opts.Examples = []interface{}{v}
	}
	
	// Build prompt based on style
	switch opts.Style {
	case "chain-of-thought":
		return builder.BuildChainOfThought(opts.Instruction, schema, opts.Format)
	case "detailed":
		return builder.BuildWithExamples(opts.Instruction, schema, opts.Format, opts.Examples)
	default:
		return builder.BuildPrompt(opts.Instruction, schema, opts.Format)
	}
}

// ValidateStructuredOutput validates output against a Go struct schema
func ValidateStructuredOutput(output string, expectedStruct interface{}) error {
	schema, err := FromGoStruct(expectedStruct)
	if err != nil {
		return fmt.Errorf("failed to generate schema from struct: %w", err)
	}
	
	// Try to parse as JSON first
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(output), &data); err != nil {
		return fmt.Errorf("failed to parse output as JSON: %w", err)
	}
	
	// Validate against schema
	return validateAgainstSchema(data, schema)
}

// MarshalStructuredOutput marshals a Go struct to the specified format
func MarshalStructuredOutput(v interface{}, format OutputFormat) (string, error) {
	formatter := NewFormatter()
	
	switch format {
	case FormatJSON:
		data, err := json.MarshalIndent(v, "", "  ")
		if err != nil {
			return "", err
		}
		return string(data), nil
		
	case FormatYAML:
		plugin := formatter.plugins[FormatYAML]
		if plugin == nil {
			return "", fmt.Errorf("YAML plugin not found")
		}
		
		// Convert to map first
		jsonData, err := json.Marshal(v)
		if err != nil {
			return "", err
		}
		
		var mapData map[string]interface{}
		if err := json.Unmarshal(jsonData, &mapData); err != nil {
			return "", err
		}
		
		schema, err := FromGoStruct(v)
		if err != nil {
			return "", err
		}
		
		return plugin.Format(schema)
		
	default:
		return "", fmt.Errorf("unsupported format: %s", format)
	}
}

// Helper functions

func getStructTag(typ reflect.Type, tagName string) string {
	// Look for struct-level tags in a special field
	if field, found := typ.FieldByName("_"); found {
		return field.Tag.Get(tagName)
	}
	return ""
}

func parseInt(s string) (int, error) {
	var val int
	_, err := fmt.Sscanf(s, "%d", &val)
	return val, err
}

func parseDefaultValue(s string, typ reflect.Type) interface{} {
	switch typ.Kind() {
	case reflect.String:
		return s
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if val, err := parseInt(s); err == nil {
			return val
		}
	case reflect.Float32, reflect.Float64:
		if val, err := parseFloat(s); err == nil {
			return val
		}
	case reflect.Bool:
		return s == "true" || s == "1"
	}
	return s
}

func hasExampleValues(v interface{}) bool {
	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	
	if val.Kind() != reflect.Struct {
		return false
	}
	
	// Check if any field has a non-zero value
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		if !field.IsZero() {
			return true
		}
	}
	
	return false
}

// Example structs with tags for documentation

// ExampleProduct demonstrates struct tags for product information
type ExampleProduct struct {
	Name        string   `json:"name" description:"Product name" minLength:"1" maxLength:"100" example:"iPhone 15"`
	Price       float64  `json:"price" description:"Price in USD" min:"0" max:"999999" example:"999.99"`
	Currency    string   `json:"currency" description:"Currency code" enum:"USD|EUR|GBP" default:"USD"`
	InStock     bool     `json:"in_stock" description:"Whether the product is in stock" example:"true"`
	Categories  []string `json:"categories" description:"Product categories"`
	Description string   `json:"description,omitempty" description:"Product description" maxLength:"500"`
}

// ExampleCodeOutput demonstrates struct for code generation
type ExampleCodeOutput struct {
	Language    string     `json:"language" description:"Programming language" enum:"python|javascript|go|rust"`
	Code        string     `json:"code" description:"The generated code" minLength:"1"`
	Explanation string     `json:"explanation" description:"Explanation of the code"`
	Complexity  string     `json:"complexity" description:"Time complexity" enum:"O(1)|O(log n)|O(n)|O(n log n)|O(n²)"`
	TestCases   []TestCase `json:"test_cases,omitempty" description:"Test cases for the code"`
}

// TestCase for code examples
type TestCase struct {
	Input    string `json:"input" description:"Test input"`
	Expected string `json:"expected" description:"Expected output"`
}

// ExampleAnalysis demonstrates struct for analysis output
type ExampleAnalysis struct {
	Topic       string         `json:"topic" description:"The main topic or subject"`
	Findings    []Finding      `json:"findings" description:"Key findings from the analysis"`
	Conclusion  string         `json:"conclusion" description:"Overall conclusion"`
	Confidence  float64        `json:"confidence" description:"Confidence level" min:"0" max:"1"`
	Metadata    map[string]any `json:"metadata,omitempty" description:"Additional metadata"`
}

// Finding represents a single finding in an analysis
type Finding struct {
	Point      string  `json:"point" description:"The finding or observation"`
	Evidence   string  `json:"evidence" description:"Supporting evidence"`
	Importance string  `json:"importance" description:"Importance level" enum:"low|medium|high"`
	Confidence float64 `json:"confidence" description:"Confidence in this finding" min:"0" max:"1"`
}