package structured

import (
	"strings"
	"testing"
)

func TestNewFormatter(t *testing.T) {
	f := NewFormatter()
	if f == nil {
		t.Fatal("NewFormatter returned nil")
	}
	if f.plugins == nil {
		t.Fatal("NewFormatter created formatter with nil plugins")
	}

	// Check built-in plugins are registered
	expectedFormats := []OutputFormat{
		FormatJSON, FormatYAML, FormatMarkdown,
		FormatJSONSchema, FormatTypeScript, FormatPydantic,
	}
	for _, format := range expectedFormats {
		if _, exists := f.plugins[format]; !exists {
			t.Errorf("missing built-in plugin for format: %s", format)
		}
	}
}

func TestRegisterPlugin(t *testing.T) {
	f := NewFormatter()
	custom := &JSONPlugin{} // reuse as custom plugin
	f.RegisterPlugin(FormatCustom, custom)

	if _, exists := f.plugins[FormatCustom]; !exists {
		t.Error("RegisterPlugin failed to register custom plugin")
	}
}

func TestFormatterFormat(t *testing.T) {
	tests := []struct {
		name    string
		schema  *Schema
		format  OutputFormat
		wantErr bool
		check   func(string) bool
	}{
		{
			name: "json format",
			schema: &Schema{
				Name: "test",
				Type: "object",
				Properties: map[string]*Property{
					"name": {Type: "string"},
					"age":  {Type: "integer"},
				},
			},
			format:  FormatJSON,
			wantErr: false,
			check: func(s string) bool {
				return strings.Contains(s, "name") && strings.Contains(s, "age")
			},
		},
		{
			name: "yaml format",
			schema: &Schema{
				Name: "test",
				Type: "object",
				Properties: map[string]*Property{
					"status": {Type: "string"},
				},
			},
			format:  FormatYAML,
			wantErr: false,
			check: func(s string) bool {
				return strings.Contains(s, "status")
			},
		},
		{
			name: "unsupported format",
			schema: &Schema{
				Name: "test",
				Type: "object",
			},
			format:  "unsupported",
			wantErr: true,
		},
	}

	f := NewFormatter()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := f.Format(tt.schema, tt.format)
			if (err != nil) != tt.wantErr {
				t.Errorf("Format() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.check != nil && !tt.check(result) {
				t.Errorf("Format() result check failed: %s", result)
			}
		})
	}
}

func TestFormatterValidate(t *testing.T) {
	tests := []struct {
		name    string
		content string
		schema  *Schema
		format  OutputFormat
		wantErr bool
	}{
		{
			name:    "valid json",
			content: `{"name": "test", "count": 42}`,
			schema: &Schema{
				Type: "object",
				Properties: map[string]*Property{
					"name":  {Type: "string"},
					"count": {Type: "number"},
				},
				Required: []string{"name"},
			},
			format:  FormatJSON,
			wantErr: false,
		},
		{
			name:    "missing required field",
			content: `{"count": 42}`,
			schema: &Schema{
				Type: "object",
				Properties: map[string]*Property{
					"name":  {Type: "string"},
					"count": {Type: "number"},
				},
				Required: []string{"name"},
			},
			format:  FormatJSON,
			wantErr: true,
		},
		{
			name:    "unsupported format",
			content: `test`,
			schema:  &Schema{Type: "object"},
			format:  "unsupported",
			wantErr: true,
		},
	}

	f := NewFormatter()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := f.Validate(tt.content, tt.schema, tt.format)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFormatterGetPromptInstructions(t *testing.T) {
	tests := []struct {
		name    string
		schema  *Schema
		format  OutputFormat
		wantErr bool
		check   func(string) bool
	}{
		{
			name: "json instructions",
			schema: &Schema{
				Name: "test",
				Type: "object",
				Properties: map[string]*Property{
					"field": {Type: "string"},
				},
			},
			format:  FormatJSON,
			wantErr: false,
			check: func(s string) bool {
				return strings.Contains(s, "JSON")
			},
		},
		{
			name: "yaml instructions",
			schema: &Schema{
				Name: "test",
				Type: "object",
			},
			format:  FormatYAML,
			wantErr: false,
			check: func(s string) bool {
				return strings.Contains(s, "YAML")
			},
		},
		{
			name:    "unsupported format",
			schema:  &Schema{Type: "object"},
			format:  "unsupported",
			wantErr: true,
		},
	}

	f := NewFormatter()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := f.GetPromptInstructions(tt.schema, tt.format)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetPromptInstructions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.check != nil && !tt.check(result) {
				t.Errorf("GetPromptInstructions() check failed: %s", result)
			}
		})
	}
}

func TestJSONPluginFormat(t *testing.T) {
	tests := []struct {
		name   string
		schema *Schema
		check  func(string) bool
	}{
		{
			name: "with properties",
			schema: &Schema{
				Name: "test",
				Properties: map[string]*Property{
					"title": {Type: "string", Default: "Hello"},
					"count": {Type: "integer"},
				},
			},
			check: func(s string) bool {
				return strings.Contains(s, "title") && strings.Contains(s, "Hello")
			},
		},
		{
			name: "with examples",
			schema: &Schema{
				Name: "test",
				Examples: []interface{}{
					map[string]interface{}{"name": "example"},
				},
			},
			check: func(s string) bool {
				return strings.Contains(s, "example")
			},
		},
	}

	p := &JSONPlugin{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := p.Format(tt.schema)
			if err != nil {
				t.Errorf("Format() error = %v", err)
				return
			}
			if !tt.check(result) {
				t.Errorf("Format() result check failed: %s", result)
			}
		})
	}
}

func TestJSONPluginParse(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantErr bool
		check   func(map[string]interface{}) bool
	}{
		{
			name:    "valid json",
			content: `{"key": "value", "num": 42}`,
			wantErr: false,
			check: func(m map[string]interface{}) bool {
				return m["key"] == "value"
			},
		},
		{
			name:    "invalid json",
			content: `{invalid}`,
			wantErr: true,
		},
	}

	p := &JSONPlugin{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := p.Parse(tt.content)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.check != nil && !tt.check(result) {
				t.Errorf("Parse() result check failed")
			}
		})
	}
}

func TestYAMLPluginFormat(t *testing.T) {
	tests := []struct {
		name   string
		schema *Schema
		check  func(string) bool
	}{
		{
			name: "with properties",
			schema: &Schema{
				Name: "test",
				Properties: map[string]*Property{
					"name": {Type: "string"},
				},
			},
			check: func(s string) bool {
				return strings.Contains(s, "name")
			},
		},
	}

	p := &YAMLPlugin{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := p.Format(tt.schema)
			if err != nil {
				t.Errorf("Format() error = %v", err)
				return
			}
			if !tt.check(result) {
				t.Errorf("Format() result check failed: %s", result)
			}
		})
	}
}

func TestYAMLPluginParse(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantErr bool
	}{
		{
			name:    "valid yaml",
			content: "name: test\nvalue: 42",
			wantErr: false,
		},
	}

	p := &YAMLPlugin{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := p.Parse(tt.content)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMarkdownPluginFormat(t *testing.T) {
	schema := &Schema{
		Name: "test",
		Properties: map[string]*Property{
			"field1": {Type: "string", Description: "First field"},
			"field2": {Type: "integer", Description: "Second field"},
		},
		Required: []string{"field1"},
	}

	p := &MarkdownPlugin{}
	result, err := p.Format(schema)
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}

	if !strings.Contains(result, "| Field |") {
		t.Error("Format() should contain markdown table header")
	}
	if !strings.Contains(result, "field1") {
		t.Error("Format() should contain field1")
	}
}

func TestMarkdownPluginParse(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantErr bool
	}{
		{
			name:    "valid table",
			content: "| Field | Value |\n|-------|-------|\n| name | test |",
			wantErr: false,
		},
		{
			name:    "invalid table",
			content: "not a table",
			wantErr: true,
		},
	}

	p := &MarkdownPlugin{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := p.Parse(tt.content)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTypeScriptPluginFormat(t *testing.T) {
	schema := &Schema{
		Name: "user_profile",
		Properties: map[string]*Property{
			"name":  {Type: "string"},
			"age":   {Type: "integer"},
			"email": {Type: "string"},
		},
		Required: []string{"name"},
	}

	p := &TypeScriptPlugin{}
	result, err := p.Format(schema)
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}

	if !strings.Contains(result, "interface") {
		t.Error("Format() should contain 'interface'")
	}
	if !strings.Contains(result, "string") {
		t.Error("Format() should contain 'string' type")
	}
	if !strings.Contains(result, "number") {
		t.Error("Format() should contain 'number' type")
	}
}

func TestPydanticPluginFormat(t *testing.T) {
	schema := &Schema{
		Name:        "product",
		Description: "A product model",
		Properties: map[string]*Property{
			"name":     {Type: "string", Description: "Product name"},
			"price":    {Type: "number"},
			"in_stock": {Type: "boolean"},
		},
		Required: []string{"name", "price"},
	}

	p := &PydanticPlugin{}
	result, err := p.Format(schema)
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}

	if !strings.Contains(result, "class") {
		t.Error("Format() should contain 'class'")
	}
	if !strings.Contains(result, "BaseModel") {
		t.Error("Format() should contain 'BaseModel'")
	}
	if !strings.Contains(result, "str") {
		t.Error("Format() should contain 'str' type")
	}
}

func TestTypeScriptAndPydanticValidateJSONData(t *testing.T) {
	schema := &Schema{
		Name: "test",
		Type: "object",
		Properties: map[string]*Property{
			"name": {Type: "string"},
			"age":  {Type: "integer"},
		},
		Required: []string{"name", "age"},
	}

	tests := []struct {
		name   string
		plugin FormatterPlugin
	}{
		{
			name:   "typescript",
			plugin: &TypeScriptPlugin{},
		},
		{
			name:   "pydantic",
			plugin: &PydanticPlugin{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name+"/valid", func(t *testing.T) {
			result, err := tt.plugin.Parse(`{"name":"Ada","age":37}`)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}
			if result["name"] != "Ada" {
				t.Fatalf("Parse() = %#v", result)
			}
			if err := tt.plugin.Validate(`{"name":"Ada","age":37}`, schema); err != nil {
				t.Fatalf("Validate() error = %v", err)
			}
		})

		t.Run(tt.name+"/schema-error", func(t *testing.T) {
			err := tt.plugin.Validate(`{"name":"Ada"}`, schema)
			if err == nil {
				t.Fatal("Validate() succeeded without required age")
			}
			if !strings.Contains(err.Error(), "missing required field: age") {
				t.Fatalf("Validate() error = %v, want missing age", err)
			}
		})

		t.Run(tt.name+"/source-rejected", func(t *testing.T) {
			err := tt.plugin.Validate(`const value = { name: "Ada", age: 37 }`, schema)
			if err == nil {
				t.Fatal("Validate() accepted source code")
			}
			if !strings.Contains(err.Error(), "expects JSON object data") {
				t.Fatalf("Validate() error = %v, want JSON object data", err)
			}
		})
	}
}

func TestJSONSchemaPluginFormat(t *testing.T) {
	schema := &Schema{
		Name:        "test",
		Description: "Test schema",
		Properties: map[string]*Property{
			"id":   {Type: "integer", Description: "Identifier"},
			"name": {Type: "string", MinLength: intPtr(1), MaxLength: intPtr(100)},
			"tags": {Type: "array", Items: &Property{Type: "string"}},
		},
		Required: []string{"id", "name"},
	}

	p := &JSONSchemaPlugin{}
	result, err := p.Format(schema)
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}

	if !strings.Contains(result, "$schema") {
		t.Error("Format() should contain '$schema'")
	}
	if !strings.Contains(result, "properties") {
		t.Error("Format() should contain 'properties'")
	}
	if !strings.Contains(result, "required") {
		t.Error("Format() should contain 'required'")
	}
}

func TestHelperFunctions(t *testing.T) {
	t.Run("toPascalCase", func(t *testing.T) {
		tests := []struct {
			input string
			want  string
		}{
			{"hello_world", "HelloWorld"},
			{"test", "Test"},
			{"my_test_case", "MyTestCase"},
		}
		for _, tt := range tests {
			got := toPascalCase(tt.input)
			if got != tt.want {
				t.Errorf("toPascalCase(%q) = %q, want %q", tt.input, got, tt.want)
			}
		}
	})

	t.Run("toSnakeCase", func(t *testing.T) {
		tests := []struct {
			input string
			want  string
		}{
			{"HelloWorld", "hello_world"},
			{"Test", "test"},
			{"myTestCase", "my_test_case"},
		}
		for _, tt := range tests {
			got := toSnakeCase(tt.input)
			if got != tt.want {
				t.Errorf("toSnakeCase(%q) = %q, want %q", tt.input, got, tt.want)
			}
		}
	})

	t.Run("goTypeToTypeScript", func(t *testing.T) {
		tests := []struct {
			input string
			want  string
		}{
			{"string", "string"},
			{"integer", "number"},
			{"number", "number"},
			{"boolean", "boolean"},
			{"array", "any[]"},
			{"object", "Record<string, any>"},
			{"unknown", "any"},
		}
		for _, tt := range tests {
			got := goTypeToTypeScript(tt.input)
			if got != tt.want {
				t.Errorf("goTypeToTypeScript(%q) = %q, want %q", tt.input, got, tt.want)
			}
		}
	})

	t.Run("goTypeToPython", func(t *testing.T) {
		tests := []struct {
			input string
			want  string
		}{
			{"string", "str"},
			{"integer", "int"},
			{"number", "float"},
			{"boolean", "bool"},
			{"array", "List[Any]"},
			{"object", "Dict[str, Any]"},
			{"unknown", "Any"},
		}
		for _, tt := range tests {
			got := goTypeToPython(tt.input)
			if got != tt.want {
				t.Errorf("goTypeToPython(%q) = %q, want %q", tt.input, got, tt.want)
			}
		}
	})
}

func TestGeneratePropertyExample(t *testing.T) {
	tests := []struct {
		name string
		prop *Property
		want interface{}
	}{
		{
			name: "string with default",
			prop: &Property{Type: "string", Default: "default_value"},
			want: "default_value",
		},
		{
			name: "string with example",
			prop: &Property{Type: "string", Examples: []interface{}{"example_value"}},
			want: "example_value",
		},
		{
			name: "string with enum",
			prop: &Property{Type: "string", Enum: []interface{}{"opt1", "opt2"}},
			want: "opt1",
		},
		{
			name: "integer",
			prop: &Property{Type: "integer"},
			want: 42,
		},
		{
			name: "number",
			prop: &Property{Type: "number"},
			want: 42,
		},
		{
			name: "boolean",
			prop: &Property{Type: "boolean"},
			want: true,
		},
		{
			name: "array with items",
			prop: &Property{
				Type:  "array",
				Items: &Property{Type: "string"},
			},
		},
		{
			name: "object with properties",
			prop: &Property{
				Type: "object",
				Properties: map[string]*Property{
					"nested": {Type: "string"},
				},
			},
		},
		{
			name: "plain string",
			prop: &Property{Type: "string"},
			want: "example_string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := generatePropertyExample(tt.prop)
			if tt.want != nil && got != tt.want {
				t.Errorf("generatePropertyExample() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidateProperty(t *testing.T) {
	tests := []struct {
		name    string
		value   interface{}
		prop    *Property
		wantErr bool
	}{
		{
			name:    "valid string",
			value:   "hello",
			prop:    &Property{Type: "string"},
			wantErr: false,
		},
		{
			name:    "string type mismatch",
			value:   42,
			prop:    &Property{Type: "string"},
			wantErr: true,
		},
		{
			name:    "string too short",
			value:   "hi",
			prop:    &Property{Type: "string", MinLength: intPtr(5)},
			wantErr: true,
		},
		{
			name:    "string too long",
			value:   "hello world",
			prop:    &Property{Type: "string", MaxLength: intPtr(5)},
			wantErr: true,
		},
		{
			name:    "string pattern match",
			value:   "test123",
			prop:    &Property{Type: "string", Pattern: "^test\\d+$"},
			wantErr: false,
		},
		{
			name:    "string pattern mismatch",
			value:   "hello",
			prop:    &Property{Type: "string", Pattern: "^test\\d+$"},
			wantErr: true,
		},
		{
			name:    "valid number",
			value:   3.14,
			prop:    &Property{Type: "number"},
			wantErr: false,
		},
		{
			name:    "valid integer",
			value:   42.0,
			prop:    &Property{Type: "integer"},
			wantErr: false,
		},
		{
			name:    "number type mismatch",
			value:   "not a number",
			prop:    &Property{Type: "number"},
			wantErr: true,
		},
		{
			name:    "valid boolean",
			value:   true,
			prop:    &Property{Type: "boolean"},
			wantErr: false,
		},
		{
			name:    "boolean type mismatch",
			value:   "true",
			prop:    &Property{Type: "boolean"},
			wantErr: true,
		},
		{
			name:    "valid array",
			value:   []interface{}{"a", "b"},
			prop:    &Property{Type: "array"},
			wantErr: false,
		},
		{
			name:    "array type mismatch",
			value:   "not an array",
			prop:    &Property{Type: "array"},
			wantErr: true,
		},
		{
			name:    "valid object",
			value:   map[string]interface{}{"key": "value"},
			prop:    &Property{Type: "object"},
			wantErr: false,
		},
		{
			name:    "object type mismatch",
			value:   "not an object",
			prop:    &Property{Type: "object"},
			wantErr: true,
		},
		{
			name:    "enum valid",
			value:   "option1",
			prop:    &Property{Type: "string", Enum: []interface{}{"option1", "option2"}},
			wantErr: false,
		},
		{
			name:    "enum invalid",
			value:   "option3",
			prop:    &Property{Type: "string", Enum: []interface{}{"option1", "option2"}},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateProperty("test_field", tt.value, tt.prop)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateProperty() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateAgainstSchema(t *testing.T) {
	tests := []struct {
		name    string
		data    map[string]interface{}
		schema  *Schema
		wantErr bool
	}{
		{
			name: "valid data",
			data: map[string]interface{}{
				"name": "test",
				"age":  30.0,
			},
			schema: &Schema{
				Properties: map[string]*Property{
					"name": {Type: "string"},
					"age":  {Type: "number"},
				},
				Required: []string{"name"},
			},
			wantErr: false,
		},
		{
			name: "missing required",
			data: map[string]interface{}{
				"age": 30.0,
			},
			schema: &Schema{
				Properties: map[string]*Property{
					"name": {Type: "string"},
					"age":  {Type: "number"},
				},
				Required: []string{"name"},
			},
			wantErr: true,
		},
		{
			name: "type mismatch",
			data: map[string]interface{}{
				"name": 123,
			},
			schema: &Schema{
				Properties: map[string]*Property{
					"name": {Type: "string"},
				},
			},
			wantErr: true,
		},
		{
			name: "extra properties allowed",
			data: map[string]interface{}{
				"name":  "test",
				"extra": "allowed",
			},
			schema: &Schema{
				Properties: map[string]*Property{
					"name": {Type: "string"},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateAgainstSchema(tt.data, tt.schema)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateAgainstSchema() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func intPtr(i int) *int {
	return &i
}
