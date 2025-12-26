package structured

import (
	"reflect"
	"testing"
	"time"
)

func TestFromGoStruct(t *testing.T) {
	type SimpleStruct struct {
		Name   string `json:"name"`
		Age    int    `json:"age"`
		Active bool   `json:"active,omitempty"`
	}

	type NestedStruct struct {
		Title string `json:"title"`
		Inner struct {
			Value int `json:"value"`
		} `json:"inner"`
	}

	type TaggedStruct struct {
		Email       string  `json:"email" description:"User email" pattern:"^[a-z]+@[a-z]+\\.[a-z]+$"`
		Score       float64 `json:"score" min:"0" max:"100"`
		Status      string  `json:"status" enum:"active|inactive|pending"`
		Description string  `json:"description" minLength:"10" maxLength:"500"`
		Priority    int     `json:"priority" default:"5" example:"3"`
	}

	tests := []struct {
		name       string
		input      interface{}
		wantErr    bool
		checkFunc  func(*Schema) bool
		wantFields []string
	}{
		{
			name:    "simple struct",
			input:   SimpleStruct{},
			wantErr: false,
			checkFunc: func(s *Schema) bool {
				return s.Type == "object" && len(s.Properties) == 3
			},
			wantFields: []string{"name", "age", "active"},
		},
		{
			name:    "struct pointer",
			input:   &SimpleStruct{},
			wantErr: false,
			checkFunc: func(s *Schema) bool {
				return s.Type == "object"
			},
		},
		{
			name:    "nested struct",
			input:   NestedStruct{},
			wantErr: false,
			checkFunc: func(s *Schema) bool {
				inner, ok := s.Properties["inner"]
				return ok && inner.Type == "object"
			},
		},
		{
			name:    "tagged struct",
			input:   TaggedStruct{},
			wantErr: false,
			checkFunc: func(s *Schema) bool {
				email := s.Properties["email"]
				if email == nil || email.Description != "User email" {
					return false
				}
				score := s.Properties["score"]
				if score == nil || score.Minimum == nil || *score.Minimum != 0 {
					return false
				}
				status := s.Properties["status"]
				if status == nil || len(status.Enum) != 3 {
					return false
				}
				return true
			},
		},
		{
			name:    "non-struct",
			input:   "not a struct",
			wantErr: true,
		},
		{
			name:    "int value",
			input:   42,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema, err := FromGoStruct(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("FromGoStruct() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if tt.checkFunc != nil && !tt.checkFunc(schema) {
				t.Errorf("FromGoStruct() schema check failed")
			}
			for _, field := range tt.wantFields {
				if _, ok := schema.Properties[field]; !ok {
					t.Errorf("FromGoStruct() missing field %s", field)
				}
			}
		})
	}
}

func TestFromGoStructWithAllTypes(t *testing.T) {
	type AllTypes struct {
		StringField  string                 `json:"string_field"`
		IntField     int                    `json:"int_field"`
		Int8Field    int8                   `json:"int8_field"`
		Int16Field   int16                  `json:"int16_field"`
		Int32Field   int32                  `json:"int32_field"`
		Int64Field   int64                  `json:"int64_field"`
		UintField    uint                   `json:"uint_field"`
		Uint8Field   uint8                  `json:"uint8_field"`
		Uint16Field  uint16                 `json:"uint16_field"`
		Uint32Field  uint32                 `json:"uint32_field"`
		Uint64Field  uint64                 `json:"uint64_field"`
		Float32Field float32                `json:"float32_field"`
		Float64Field float64                `json:"float64_field"`
		BoolField    bool                   `json:"bool_field"`
		SliceField   []string               `json:"slice_field"`
		ArrayField   [3]int                 `json:"array_field"`
		MapField     map[string]interface{} `json:"map_field"`
		TimeField    time.Time              `json:"time_field"`
		PtrField     *string                `json:"ptr_field"`
		AnyField     interface{}            `json:"any_field"`
	}

	schema, err := FromGoStruct(AllTypes{})
	if err != nil {
		t.Fatalf("FromGoStruct() error = %v", err)
	}

	typeChecks := map[string]string{
		"string_field":  "string",
		"int_field":     "integer",
		"int8_field":    "integer",
		"int16_field":   "integer",
		"int32_field":   "integer",
		"int64_field":   "integer",
		"uint_field":    "integer",
		"float32_field": "number",
		"float64_field": "number",
		"bool_field":    "boolean",
		"slice_field":   "array",
		"array_field":   "array",
		"map_field":     "object",
		"time_field":    "string",
		"ptr_field":     "string",
		"any_field":     "object",
	}

	for field, expectedType := range typeChecks {
		prop, ok := schema.Properties[field]
		if !ok {
			t.Errorf("missing field %s", field)
			continue
		}
		if prop.Type != expectedType {
			t.Errorf("field %s: got type %s, want %s", field, prop.Type, expectedType)
		}
	}

	// Check time.Time has date-time format
	if timeField := schema.Properties["time_field"]; timeField.Format != "date-time" {
		t.Errorf("time_field format = %s, want date-time", timeField.Format)
	}
}

func TestFromGoStructWithUnexportedFields(t *testing.T) {
	type WithUnexported struct {
		Public  string `json:"public"`
		private string
	}

	schema, err := FromGoStruct(WithUnexported{})
	if err != nil {
		t.Fatalf("FromGoStruct() error = %v", err)
	}

	if len(schema.Properties) != 1 {
		t.Errorf("expected 1 property, got %d", len(schema.Properties))
	}
	if _, ok := schema.Properties["public"]; !ok {
		t.Error("expected 'public' field")
	}
}

func TestFromGoStructWithJSONIgnore(t *testing.T) {
	type WithIgnored struct {
		Included string `json:"included"`
		Ignored  string `json:"-"`
	}

	schema, err := FromGoStruct(WithIgnored{})
	if err != nil {
		t.Fatalf("FromGoStruct() error = %v", err)
	}

	if len(schema.Properties) != 1 {
		t.Errorf("expected 1 property, got %d", len(schema.Properties))
	}
	if _, ok := schema.Properties["ignored"]; ok {
		t.Error("ignored field should not be included")
	}
}

func TestFromGoStructRequired(t *testing.T) {
	type RequiredFields struct {
		Required   string `json:"required"`
		Optional   string `json:"optional,omitempty"`
		AlsoNeeded int    `json:"also_needed"`
	}

	schema, err := FromGoStruct(RequiredFields{})
	if err != nil {
		t.Fatalf("FromGoStruct() error = %v", err)
	}

	// Fields without omitempty should be required
	expectedRequired := map[string]bool{
		"required":    true,
		"also_needed": true,
	}

	for _, req := range schema.Required {
		if !expectedRequired[req] {
			t.Errorf("unexpected required field: %s", req)
		}
		delete(expectedRequired, req)
	}

	if len(expectedRequired) > 0 {
		for field := range expectedRequired {
			t.Errorf("missing required field: %s", field)
		}
	}
}

func TestCreatePropertyFromType(t *testing.T) {
	tests := []struct {
		name     string
		typ      reflect.Type
		tag      reflect.StructTag
		wantType string
		wantErr  bool
	}{
		{
			name:     "string",
			typ:      reflect.TypeOf(""),
			wantType: "string",
		},
		{
			name:     "int",
			typ:      reflect.TypeOf(0),
			wantType: "integer",
		},
		{
			name:     "float64",
			typ:      reflect.TypeOf(0.0),
			wantType: "number",
		},
		{
			name:     "bool",
			typ:      reflect.TypeOf(true),
			wantType: "boolean",
		},
		{
			name:     "slice",
			typ:      reflect.TypeOf([]string{}),
			wantType: "array",
		},
		{
			name:     "map",
			typ:      reflect.TypeOf(map[string]interface{}{}),
			wantType: "object",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prop, err := createPropertyFromType(tt.typ, tt.tag)
			if (err != nil) != tt.wantErr {
				t.Errorf("createPropertyFromType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && prop.Type != tt.wantType {
				t.Errorf("createPropertyFromType() type = %v, want %v", prop.Type, tt.wantType)
			}
		})
	}
}

func TestGeneratePromptFromStruct(t *testing.T) {
	type TestStruct struct {
		Name string `json:"name" description:"User name"`
		Age  int    `json:"age"`
	}

	tests := []struct {
		name    string
		input   interface{}
		opts    StructuredPromptOptions
		wantErr bool
		check   func(string) bool
	}{
		{
			name:  "default format",
			input: TestStruct{},
			opts: StructuredPromptOptions{
				Instruction: "Generate user info",
			},
			wantErr: false,
			check: func(s string) bool {
				return len(s) > 0
			},
		},
		{
			name:  "json format",
			input: TestStruct{},
			opts: StructuredPromptOptions{
				Instruction: "Generate user info",
				Format:      FormatJSON,
			},
			wantErr: false,
		},
		{
			name:  "chain-of-thought style",
			input: TestStruct{},
			opts: StructuredPromptOptions{
				Instruction: "Generate user info",
				Style:       "chain-of-thought",
			},
			wantErr: false,
			check: func(s string) bool {
				return true
			},
		},
		{
			name:  "detailed style",
			input: TestStruct{},
			opts: StructuredPromptOptions{
				Instruction: "Generate user info",
				Style:       "detailed",
			},
			wantErr: false,
		},
		{
			name:    "non-struct",
			input:   "not a struct",
			opts:    StructuredPromptOptions{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := GeneratePromptFromStruct(tt.input, tt.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("GeneratePromptFromStruct() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.check != nil && !tt.check(result) {
				t.Errorf("GeneratePromptFromStruct() check failed")
			}
		})
	}
}

func TestValidateStructuredOutput(t *testing.T) {
	type Expected struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	tests := []struct {
		name    string
		output  string
		expect  interface{}
		wantErr bool
	}{
		{
			name:    "valid output",
			output:  `{"name": "John", "age": 30}`,
			expect:  Expected{},
			wantErr: false,
		},
		{
			name:    "invalid json",
			output:  `{invalid}`,
			expect:  Expected{},
			wantErr: true,
		},
		{
			name:    "missing required",
			output:  `{"age": 30}`,
			expect:  Expected{},
			wantErr: true,
		},
		{
			name:    "non-struct expect",
			output:  `{"name": "John"}`,
			expect:  "not a struct",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateStructuredOutput(tt.output, tt.expect)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateStructuredOutput() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMarshalStructuredOutput(t *testing.T) {
	type TestData struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}

	tests := []struct {
		name    string
		input   interface{}
		format  OutputFormat
		wantErr bool
		check   func(string) bool
	}{
		{
			name:    "json format",
			input:   TestData{Name: "test", Value: 42},
			format:  FormatJSON,
			wantErr: false,
			check: func(s string) bool {
				return len(s) > 0
			},
		},
		{
			name:    "yaml format",
			input:   TestData{Name: "test", Value: 42},
			format:  FormatYAML,
			wantErr: false,
		},
		{
			name:    "unsupported format",
			input:   TestData{Name: "test", Value: 42},
			format:  FormatCSV,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := MarshalStructuredOutput(tt.input, tt.format)
			if (err != nil) != tt.wantErr {
				t.Errorf("MarshalStructuredOutput() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.check != nil && !tt.check(result) {
				t.Errorf("MarshalStructuredOutput() check failed")
			}
		})
	}
}

func TestHasExampleValues(t *testing.T) {
	type TestStruct struct {
		Name string
		Age  int
	}

	tests := []struct {
		name  string
		input interface{}
		want  bool
	}{
		{
			name:  "zero struct",
			input: TestStruct{},
			want:  false,
		},
		{
			name:  "non-zero struct",
			input: TestStruct{Name: "test"},
			want:  true,
		},
		{
			name:  "pointer to zero struct",
			input: &TestStruct{},
			want:  false,
		},
		{
			name:  "pointer to non-zero struct",
			input: &TestStruct{Age: 25},
			want:  true,
		},
		{
			name:  "non-struct",
			input: "not a struct",
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hasExampleValues(tt.input)
			if got != tt.want {
				t.Errorf("hasExampleValues() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseDefaultValue(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		typ     reflect.Type
		want    interface{}
		wantStr bool
	}{
		{
			name:    "string value",
			input:   "hello",
			typ:     reflect.TypeOf(""),
			want:    "hello",
			wantStr: true,
		},
		{
			name:  "int value",
			input: "42",
			typ:   reflect.TypeOf(0),
			want:  42,
		},
		{
			name:  "float value",
			input: "3.14",
			typ:   reflect.TypeOf(0.0),
			want:  3.14,
		},
		{
			name:  "bool true",
			input: "true",
			typ:   reflect.TypeOf(true),
			want:  true,
		},
		{
			name:  "bool 1",
			input: "1",
			typ:   reflect.TypeOf(true),
			want:  true,
		},
		{
			name:  "bool false",
			input: "false",
			typ:   reflect.TypeOf(true),
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseDefaultValue(tt.input, tt.typ)
			if got != tt.want {
				t.Errorf("parseDefaultValue() = %v, want %v", got, tt.want)
			}
		})
	}
}
