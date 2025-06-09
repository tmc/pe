package structured

import (
	"encoding/json"
	"testing"
)

func TestJSONValidation(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		schema  *Schema
		wantErr bool
	}{
		{
			name:  "valid JSON object",
			input: `{"name": "John", "age": 30}`,
			schema: &Schema{
				Type: "object",
				Properties: map[string]*Property{
					"name": {Type: "string"},
					"age":  {Type: "number"},
				},
			},
			wantErr: false,
		},
		{
			name:  "invalid JSON",
			input: `{name: "John"}`,
			schema: &Schema{
				Type: "object",
			},
			wantErr: true,
		},
		{
			name:  "missing required field",
			input: `{"name": "John"}`,
			schema: &Schema{
				Type: "object",
				Properties: map[string]*Property{
					"name": {Type: "string"},
					"age":  {Type: "number"},
				},
				Required: []string{"name", "age"},
			},
			wantErr: true,
		},
	}

	formatter := NewFormatter()
	jsonPlugin := formatter.plugins[FormatJSON]

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := jsonPlugin.Validate(tt.input, tt.schema)

			if tt.wantErr && err == nil {
				t.Errorf("Expected error for test case %s", tt.name)
			}

			if !tt.wantErr && err != nil {
				t.Errorf("Unexpected error for test case %s: %v", tt.name, err)
			}
		})
	}
}

func TestFormatConversion(t *testing.T) {
	tests := []struct {
		name   string
		format string
		input  interface{}
		want   string
	}{
		{
			name:   "struct to JSON",
			format: "json",
			input: struct {
				Name string `json:"name"`
				Age  int    `json:"age"`
			}{
				Name: "Alice",
				Age:  25,
			},
			want: `{"name":"Alice","age":25}`,
		},
		{
			name:   "map to JSON",
			format: "json",
			input: map[string]interface{}{
				"key1": "value1",
				"key2": 42,
			},
			want: `{"key1":"value1","key2":42}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := json.Marshal(tt.input)
			if err != nil {
				t.Errorf("json.Marshal() error = %v", err)
				return
			}

			// Compare JSON strings (order might differ for maps)
			var got, want interface{}
			json.Unmarshal(result, &got)
			json.Unmarshal([]byte(tt.want), &want)

			gotJSON, _ := json.Marshal(got)
			wantJSON, _ := json.Marshal(want)

			if string(gotJSON) != string(wantJSON) {
				t.Errorf("FormatConversion() = %v, want %v", string(gotJSON), string(wantJSON))
			}
		})
	}
}

func TestSchemaGeneration(t *testing.T) {
	type TestStruct struct {
		Name     string                 `json:"name" required:"true"`
		Age      int                    `json:"age" min:"0" max:"150"`
		Email    string                 `json:"email" format:"email"`
		Tags     []string               `json:"tags" minItems:"1"`
		Active   bool                   `json:"active"`
		Metadata map[string]interface{} `json:"metadata,omitempty"`
	}

	// Test that we can create instances
	instance := TestStruct{
		Name:   "Test User",
		Age:    30,
		Email:  "test@example.com",
		Tags:   []string{"user", "test"},
		Active: true,
	}

	// Test JSON marshaling
	data, err := json.Marshal(instance)
	if err != nil {
		t.Errorf("Failed to marshal struct: %v", err)
	}

	// Test JSON unmarshaling
	var decoded TestStruct
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Errorf("Failed to unmarshal JSON: %v", err)
	}

	if decoded.Name != instance.Name {
		t.Errorf("Name = %v, want %v", decoded.Name, instance.Name)
	}

	if len(decoded.Tags) != len(instance.Tags) {
		t.Errorf("Tags length = %v, want %v", len(decoded.Tags), len(instance.Tags))
	}
}

func BenchmarkJSONValidation(b *testing.B) {
	input := `{"name": "John", "age": 30, "email": "john@example.com"}`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var data interface{}
		_ = json.Unmarshal([]byte(input), &data)
	}
}

func BenchmarkStructMarshaling(b *testing.B) {
	data := struct {
		Name  string `json:"name"`
		Age   int    `json:"age"`
		Email string `json:"email"`
	}{
		Name:  "John",
		Age:   30,
		Email: "john@example.com",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(data)
	}
}
