package structured

import (
	"strings"
	"testing"
)

func TestNewPromptBuilder(t *testing.T) {
	pb := NewPromptBuilder()
	if pb == nil {
		t.Fatal("NewPromptBuilder returned nil")
	}
	if pb.formatter == nil {
		t.Error("PromptBuilder has nil formatter")
	}
	if pb.templates == nil {
		t.Error("PromptBuilder has nil templates")
	}
	if pb.defaultFormat != FormatJSON {
		t.Errorf("PromptBuilder defaultFormat = %v, want %v", pb.defaultFormat, FormatJSON)
	}
}

func TestPromptBuilderSetDefaultFormat(t *testing.T) {
	pb := NewPromptBuilder()
	pb.SetDefaultFormat(FormatYAML)
	if pb.defaultFormat != FormatYAML {
		t.Errorf("SetDefaultFormat() = %v, want %v", pb.defaultFormat, FormatYAML)
	}
}

func TestPromptBuilderRegisterTemplate(t *testing.T) {
	tests := []struct {
		name     string
		tmplName string
		tmpl     string
		wantErr  bool
	}{
		{
			name:     "valid template",
			tmplName: "test",
			tmpl:     "Hello {{.Name}}!",
			wantErr:  false,
		},
		{
			name:     "invalid template",
			tmplName: "bad",
			tmpl:     "{{.Unclosed",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pb := NewPromptBuilder()
			err := pb.RegisterTemplate(tt.tmplName, tt.tmpl)
			if (err != nil) != tt.wantErr {
				t.Errorf("RegisterTemplate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				if _, exists := pb.templates[tt.tmplName]; !exists {
					t.Error("Template was not registered")
				}
			}
		})
	}
}

func TestPromptBuilderBuildPrompt(t *testing.T) {
	tests := []struct {
		name        string
		instruction string
		schema      *Schema
		format      OutputFormat
		wantErr     bool
		check       func(string) bool
	}{
		{
			name:        "basic prompt",
			instruction: "Generate a user profile",
			schema: &Schema{
				Name: "user",
				Properties: map[string]*Property{
					"name": {Type: "string", Description: "User name"},
					"age":  {Type: "integer"},
				},
				Required: []string{"name"},
			},
			format:  FormatJSON,
			wantErr: false,
			check: func(s string) bool {
				return strings.Contains(s, "Generate a user profile") &&
					strings.Contains(s, "Required fields:")
			},
		},
		{
			name:        "with validation rules",
			instruction: "Create data",
			schema: &Schema{
				Name: "data",
				Properties: map[string]*Property{
					"email": {Type: "string", Pattern: "^[a-z]+@[a-z]+$"},
					"score": {Type: "number", Enum: []interface{}{"low", "medium", "high"}},
				},
			},
			format:  FormatJSON,
			wantErr: false,
			check: func(s string) bool {
				return strings.Contains(s, "Validation rules")
			},
		},
		{
			name:        "default format",
			instruction: "Test",
			schema:      &Schema{Name: "test"},
			format:      "",
			wantErr:     false,
		},
		{
			name:        "unsupported format",
			instruction: "Test",
			schema:      &Schema{Name: "test"},
			format:      "unsupported",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pb := NewPromptBuilder()
			result, err := pb.BuildPrompt(tt.instruction, tt.schema, tt.format)
			if (err != nil) != tt.wantErr {
				t.Errorf("BuildPrompt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.check != nil && !tt.check(result) {
				t.Errorf("BuildPrompt() check failed: %s", result)
			}
		})
	}
}

func TestPromptBuilderBuildFromTemplate(t *testing.T) {
	tests := []struct {
		name         string
		setupTmpl    string
		tmplName     string
		data         interface{}
		schema       *Schema
		format       OutputFormat
		wantErr      bool
		wantContains string
	}{
		{
			name:      "valid template",
			setupTmpl: "Task: {{.Task}}",
			tmplName:  "task_tmpl",
			data:      map[string]string{"Task": "Summarize"},
			schema:    &Schema{Name: "test"},
			format:    FormatJSON,
			wantErr:   false,
		},
		{
			name:      "template not found",
			setupTmpl: "",
			tmplName:  "nonexistent",
			data:      nil,
			schema:    &Schema{Name: "test"},
			format:    FormatJSON,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pb := NewPromptBuilder()
			if tt.setupTmpl != "" {
				pb.RegisterTemplate(tt.tmplName, tt.setupTmpl)
			}

			result, err := pb.BuildFromTemplate(tt.tmplName, tt.data, tt.schema, tt.format)
			if (err != nil) != tt.wantErr {
				t.Errorf("BuildFromTemplate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.wantContains != "" {
				if !strings.Contains(result, tt.wantContains) {
					t.Errorf("BuildFromTemplate() missing expected content: %s", tt.wantContains)
				}
			}
		})
	}
}

func TestPromptBuilderBuildWithExamples(t *testing.T) {
	tests := []struct {
		name        string
		instruction string
		schema      *Schema
		format      OutputFormat
		examples    []interface{}
		wantErr     bool
		check       func(string) bool
	}{
		{
			name:        "with examples",
			instruction: "Generate data",
			schema:      &Schema{Name: "test"},
			format:      FormatJSON,
			examples: []interface{}{
				map[string]interface{}{"name": "example1"},
				map[string]interface{}{"name": "example2"},
			},
			wantErr: false,
			check: func(s string) bool {
				return strings.Contains(s, "Example")
			},
		},
		{
			name:        "no examples",
			instruction: "Generate data",
			schema:      &Schema{Name: "test"},
			format:      FormatJSON,
			examples:    nil,
			wantErr:     false,
		},
		{
			name:        "unsupported format",
			instruction: "Test",
			schema:      &Schema{Name: "test"},
			format:      "unsupported",
			examples:    nil,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pb := NewPromptBuilder()
			result, err := pb.BuildWithExamples(tt.instruction, tt.schema, tt.format, tt.examples)
			if (err != nil) != tt.wantErr {
				t.Errorf("BuildWithExamples() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.check != nil && !tt.check(result) {
				t.Errorf("BuildWithExamples() check failed")
			}
		})
	}
}

func TestPromptBuilderBuildChainOfThought(t *testing.T) {
	tests := []struct {
		name        string
		instruction string
		schema      *Schema
		format      OutputFormat
		wantErr     bool
		check       func(string) bool
	}{
		{
			name:        "chain of thought",
			instruction: "Solve this problem",
			schema:      &Schema{Name: "solution"},
			format:      FormatJSON,
			wantErr:     false,
			check: func(s string) bool {
				return strings.Contains(s, "step-by-step") &&
					strings.Contains(s, "<reasoning>") &&
					strings.Contains(s, "<answer>")
			},
		},
		{
			name:        "unsupported format",
			instruction: "Test",
			schema:      &Schema{Name: "test"},
			format:      "unsupported",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pb := NewPromptBuilder()
			result, err := pb.BuildChainOfThought(tt.instruction, tt.schema, tt.format)
			if (err != nil) != tt.wantErr {
				t.Errorf("BuildChainOfThought() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.check != nil && !tt.check(result) {
				t.Errorf("BuildChainOfThought() check failed: %s", result)
			}
		})
	}
}

func TestPromptBuilderBuildFunctionCall(t *testing.T) {
	tests := []struct {
		name        string
		funcName    string
		description string
		paramSchema *Schema
		wantErr     bool
		check       func(string) bool
	}{
		{
			name:        "function call prompt",
			funcName:    "searchDatabase",
			description: "Search for records in the database",
			paramSchema: &Schema{
				Name: "params",
				Properties: map[string]*Property{
					"query": {Type: "string"},
					"limit": {Type: "integer"},
				},
			},
			wantErr: false,
			check: func(s string) bool {
				return strings.Contains(s, "searchDatabase") &&
					strings.Contains(s, "Search for records")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pb := NewPromptBuilder()
			result, err := pb.BuildFunctionCall(tt.funcName, tt.description, tt.paramSchema)
			if (err != nil) != tt.wantErr {
				t.Errorf("BuildFunctionCall() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.check != nil && !tt.check(result) {
				t.Errorf("BuildFunctionCall() check failed: %s", result)
			}
		})
	}
}

func TestPromptBuilderBuildMultiStep(t *testing.T) {
	tests := []struct {
		name    string
		steps   []StructuredStep
		wantErr bool
		check   func(string) bool
	}{
		{
			name: "multi step prompt",
			steps: []StructuredStep{
				{
					Description: "Analyze the input",
					Schema:      &Schema{Name: "analysis"},
					Format:      FormatJSON,
				},
				{
					Description: "Generate output",
					Schema:      &Schema{Name: "output"},
					Format:      FormatJSON,
				},
			},
			wantErr: false,
			check: func(s string) bool {
				return strings.Contains(s, "Step 1") &&
					strings.Contains(s, "Step 2") &&
					strings.Contains(s, "Analyze the input") &&
					strings.Contains(s, "Generate output")
			},
		},
		{
			name:    "empty steps",
			steps:   []StructuredStep{},
			wantErr: false,
			check: func(s string) bool {
				return strings.Contains(s, "Complete the following steps")
			},
		},
		{
			name: "step without schema",
			steps: []StructuredStep{
				{
					Description: "Do something",
					Schema:      nil,
					Format:      FormatJSON,
				},
			},
			wantErr: false,
		},
		{
			name: "unsupported format in step",
			steps: []StructuredStep{
				{
					Description: "Test",
					Schema:      &Schema{Name: "test"},
					Format:      "unsupported",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pb := NewPromptBuilder()
			result, err := pb.BuildMultiStep(tt.steps)
			if (err != nil) != tt.wantErr {
				t.Errorf("BuildMultiStep() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.check != nil && !tt.check(result) {
				t.Errorf("BuildMultiStep() check failed: %s", result)
			}
		})
	}
}

func TestPromptBuilderBuildConditional(t *testing.T) {
	tests := []struct {
		name        string
		instruction string
		conditions  []ConditionalOutput
		wantErr     bool
		check       func(string) bool
	}{
		{
			name:        "conditional output",
			instruction: "Analyze the data",
			conditions: []ConditionalOutput{
				{
					Condition: "the data is numerical",
					Schema:    &Schema{Name: "numerical_output"},
					Format:    FormatJSON,
				},
				{
					Condition: "the data is textual",
					Schema:    &Schema{Name: "textual_output"},
					Format:    FormatJSON,
				},
			},
			wantErr: false,
			check: func(s string) bool {
				return strings.Contains(s, "Analyze the data") &&
					strings.Contains(s, "If the data is numerical") &&
					strings.Contains(s, "If the data is textual")
			},
		},
		{
			name:        "empty conditions",
			instruction: "Process input",
			conditions:  []ConditionalOutput{},
			wantErr:     false,
		},
		{
			name:        "unsupported format",
			instruction: "Test",
			conditions: []ConditionalOutput{
				{
					Condition: "test",
					Schema:    &Schema{Name: "test"},
					Format:    "unsupported",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pb := NewPromptBuilder()
			result, err := pb.BuildConditional(tt.instruction, tt.conditions)
			if (err != nil) != tt.wantErr {
				t.Errorf("BuildConditional() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.check != nil && !tt.check(result) {
				t.Errorf("BuildConditional() check failed: %s", result)
			}
		})
	}
}

func TestGenerateValidationRules(t *testing.T) {
	tests := []struct {
		name   string
		schema *Schema
		check  func(string) bool
	}{
		{
			name: "with min/max length",
			schema: &Schema{
				Properties: map[string]*Property{
					"name": {
						Type:      "string",
						MinLength: intPtr(1),
						MaxLength: intPtr(100),
					},
				},
			},
			check: func(s string) bool {
				return strings.Contains(s, "minimum length") &&
					strings.Contains(s, "maximum length")
			},
		},
		{
			name: "with pattern",
			schema: &Schema{
				Properties: map[string]*Property{
					"email": {
						Type:    "string",
						Pattern: "^[a-z]+@[a-z]+$",
					},
				},
			},
			check: func(s string) bool {
				return strings.Contains(s, "must match pattern")
			},
		},
		{
			name: "with enum",
			schema: &Schema{
				Properties: map[string]*Property{
					"status": {
						Type: "string",
						Enum: []interface{}{"active", "inactive"},
					},
				},
			},
			check: func(s string) bool {
				return strings.Contains(s, "must be one of")
			},
		},
		{
			name: "no rules",
			schema: &Schema{
				Properties: map[string]*Property{
					"name": {Type: "string"},
				},
			},
			check: func(s string) bool {
				return s == ""
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pb := NewPromptBuilder()
			result := pb.generateValidationRules(tt.schema)
			if !tt.check(result) {
				t.Errorf("generateValidationRules() check failed: %s", result)
			}
		})
	}
}

func TestFormatExample(t *testing.T) {
	tests := []struct {
		name    string
		example interface{}
		format  OutputFormat
		wantErr bool
	}{
		{
			name:    "json format",
			example: map[string]interface{}{"key": "value"},
			format:  FormatJSON,
			wantErr: false,
		},
		{
			name:    "yaml format",
			example: map[string]interface{}{"key": "value"},
			format:  FormatYAML,
			wantErr: false,
		},
		{
			name:    "other format",
			example: "test",
			format:  FormatMarkdown,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pb := NewPromptBuilder()
			_, err := pb.formatExample(tt.example, tt.format)
			if (err != nil) != tt.wantErr {
				t.Errorf("formatExample() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCommonSchemas(t *testing.T) {
	schemas := []struct {
		name   string
		schema *Schema
	}{
		{"Analysis", CommonSchemas.Analysis},
		{"CodeGeneration", CommonSchemas.CodeGeneration},
		{"DataExtraction", CommonSchemas.DataExtraction},
		{"Classification", CommonSchemas.Classification},
		{"Summary", CommonSchemas.Summary},
	}

	for _, s := range schemas {
		t.Run(s.name, func(t *testing.T) {
			if s.schema == nil {
				t.Error("schema is nil")
				return
			}
			if s.schema.Name == "" {
				t.Error("schema has no name")
			}
			if s.schema.Type != "object" {
				t.Errorf("schema type = %s, want object", s.schema.Type)
			}
			if len(s.schema.Properties) == 0 {
				t.Error("schema has no properties")
			}
			if len(s.schema.Required) == 0 {
				t.Error("schema has no required fields")
			}
		})
	}
}
