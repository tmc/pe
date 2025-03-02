package template

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTemplateBasicVariableSubstitution(t *testing.T) {
	tmpl := NewTemplate("Hello, {{name}}!", map[string]interface{}{
		"name": "World",
	})
	
	result, err := tmpl.Process()
	if err != nil {
		t.Fatalf("Failed to process template: %v", err)
	}
	
	expected := "Hello, World!"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestTemplateNestedVariables(t *testing.T) {
	tmpl := NewTemplate("{{greeting}}, {{name}}!", map[string]interface{}{
		"greeting": "Hello",
		"name":     "World",
	})
	
	result, err := tmpl.Process()
	if err != nil {
		t.Fatalf("Failed to process template: %v", err)
	}
	
	expected := "Hello, World!"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestTemplateEscapedVariables(t *testing.T) {
	tmpl := NewTemplate("Literal {{name}} and \\{{name}}", map[string]interface{}{
		"name": "World",
	})
	
	result, err := tmpl.Process()
	if err != nil {
		t.Fatalf("Failed to process template: %v", err)
	}
	
	expected := "Literal World and {{name}}"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestTemplateConditionalLogic(t *testing.T) {
	tests := []struct {
		name     string
		template string
		vars     map[string]interface{}
		expected string
	}{
		{
			name:     "Simple if-then condition (true)",
			template: "#if showGreeting then Hello, {{name}}! #endif",
			vars: map[string]interface{}{
				"showGreeting": true,
				"name":         "World",
			},
			expected: "Hello, World!",
		},
		{
			name:     "Simple if-then condition (false)",
			template: "#if showGreeting then Hello, {{name}}! #endif",
			vars: map[string]interface{}{
				"showGreeting": false,
				"name":         "World",
			},
			expected: "",
		},
		{
			name:     "If-then-else condition (true)",
			template: "#if showGreeting then Hello, {{name}}! #else Goodbye, {{name}}! #endif",
			vars: map[string]interface{}{
				"showGreeting": true,
				"name":         "World",
			},
			expected: "Hello, World!",
		},
		{
			name:     "If-then-else condition (false)",
			template: "#if showGreeting then Hello, {{name}}! #else Goodbye, {{name}}! #endif",
			vars: map[string]interface{}{
				"showGreeting": false,
				"name":         "World",
			},
			expected: "Goodbye, World!",
		},
		{
			name:     "Nested conditions (true->true)",
			template: "#if outer then Outer #if inner then Inner #endif #endif",
			vars: map[string]interface{}{
				"outer": true,
				"inner": true,
			},
			expected: "Outer Inner",
		},
		{
			name:     "Nested conditions (true->false)",
			template: "#if outer then Outer #if inner then Inner #endif #endif",
			vars: map[string]interface{}{
				"outer": true,
				"inner": false,
			},
			expected: "Outer",
		},
		{
			name:     "Equality condition (true)",
			template: "#if name == \"World\" then Hello, World! #endif",
			vars: map[string]interface{}{
				"name": "World",
			},
			expected: "Hello, World!",
		},
		{
			name:     "Equality condition (false)",
			template: "#if name == \"Universe\" then Hello, Universe! #endif",
			vars: map[string]interface{}{
				"name": "World",
			},
			expected: "",
		},
		{
			name:     "Inequality condition (true)",
			template: "#if name != \"Universe\" then Hello, {{name}}! #endif",
			vars: map[string]interface{}{
				"name": "World",
			},
			expected: "Hello, World!",
		},
		{
			name:     "Inequality condition (false)",
			template: "#if name != \"World\" then Hello, Universe! #endif",
			vars: map[string]interface{}{
				"name": "World",
			},
			expected: "",
		},
	}
	
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tmpl := NewTemplate(tc.template, tc.vars)
			result, err := tmpl.Process()
			if err != nil {
				t.Fatalf("Failed to process template: %v", err)
			}
			
			if result != tc.expected {
				t.Errorf("Expected %q, got %q", tc.expected, result)
			}
		})
	}
}

func TestTemplateFileInclusion(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "template-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	// Create test files
	greetingFile := filepath.Join(tempDir, "greeting.txt")
	if err := os.WriteFile(greetingFile, []byte("Hello"), 0644); err != nil {
		t.Fatalf("Failed to write greeting file: %v", err)
	}
	
	nameFile := filepath.Join(tempDir, "name.txt")
	if err := os.WriteFile(nameFile, []byte("World"), 0644); err != nil {
		t.Fatalf("Failed to write name file: %v", err)
	}
	
	nestedFile := filepath.Join(tempDir, "nested.txt")
	nestedContent := "#include \"" + filepath.Base(greetingFile) + "\", #include \"" + filepath.Base(nameFile) + "\"!"
	if err := os.WriteFile(nestedFile, []byte(nestedContent), 0644); err != nil {
		t.Fatalf("Failed to write nested file: %v", err)
	}
	
	// Test simple file inclusion with legacy syntax
	tmpl := NewTemplate("{{file \""+greetingFile+"\"}}, {{name}}!", map[string]interface{}{
		"name": "World",
	})
	tmpl.SetBaseDir(tempDir)
	
	result, err := tmpl.Process()
	if err != nil {
		t.Fatalf("Failed to process template with legacy file inclusion: %v", err)
	}
	
	expected := "Hello, World!"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
	
	// Test simple file inclusion with new syntax
	tmpl = NewTemplate("#include \""+filepath.Base(greetingFile)+"\", {{name}}!", map[string]interface{}{
		"name": "World",
	})
	tmpl.SetBaseDir(tempDir)
	
	result, err = tmpl.Process()
	if err != nil {
		t.Fatalf("Failed to process template with new file inclusion: %v", err)
	}
	
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
	
	// Test nested file inclusion
	tmpl = NewTemplate("#include \""+filepath.Base(nestedFile)+"\"", nil)
	tmpl.SetBaseDir(tempDir)
	
	result, err = tmpl.Process()
	if err != nil {
		t.Fatalf("Failed to process template with nested file inclusion: %v", err)
	}
	
	expected = "Hello, World!"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestTemplateValidation(t *testing.T) {
	// Test missing variables
	tmpl := NewTemplate("Hello, {{name}}! Today is {{day}}.", map[string]interface{}{
		"name": "World",
	})
	
	errors := tmpl.Validate()
	if len(errors) != 1 {
		t.Errorf("Expected 1 validation error, got %d", len(errors))
	} else if errors[0] != "missing variable: day" {
		t.Errorf("Unexpected error message: %s", errors[0])
	}
	
	// Test missing files
	tmpl = NewTemplate("#include \"non_existent_file.txt\"", nil)
	errors = tmpl.Validate()
	if len(errors) != 1 {
		t.Errorf("Expected 1 validation error, got %d", len(errors))
	} else if errors[0] != "missing file: non_existent_file.txt" {
		t.Errorf("Unexpected error message: %s", errors[0])
	}
}

func TestExpressionEvaluation(t *testing.T) {
	vars := map[string]interface{}{
		"stringVar":   "test",
		"emptyString": "",
		"intVar":      42,
		"zeroInt":     0,
		"boolVar":     true,
		"falseBool":   false,
	}
	
	tests := []struct {
		expr     string
		expected bool
	}{
		// Existence checks
		{"stringVar", true},
		{"nonExistentVar", false},
		{"emptyString", false},
		{"intVar", true},
		{"zeroInt", false},
		{"boolVar", true},
		{"falseBool", false},
		
		// Equality checks
		{"stringVar == \"test\"", true},
		{"stringVar == \"other\"", false},
		{"intVar == 42", true},
		{"intVar == 0", false},
		
		// Inequality checks
		{"stringVar != \"other\"", true},
		{"stringVar != \"test\"", false},
		{"intVar != 0", true},
		{"intVar != 42", false},
		
		// Variable comparisons
		{"stringVar == stringVar", true},
		{"intVar == zeroInt", false},
	}
	
	for _, tc := range tests {
		result, err := EvaluateExpression(tc.expr, vars)
		if err != nil {
			t.Errorf("Error evaluating %q: %v", tc.expr, err)
			continue
		}
		
		if result != tc.expected {
			t.Errorf("Expression %q: expected %v, got %v", tc.expr, tc.expected, result)
		}
	}
}

// Test backwards compatibility
func TestLegacyFunctions(t *testing.T) {
	// Test ProcessPrompt
	result, err := ProcessPrompt("Hello, {{name}}!", map[string]interface{}{
		"name": "World",
	})
	if err != nil {
		t.Fatalf("ProcessPrompt failed: %v", err)
	}
	if result != "Hello, World!" {
		t.Errorf("Expected 'Hello, World!', got %q", result)
	}
	
	// Test ProcessMany
	results, err := ProcessMany([]string{
		"Hello, {{name}}!",
		"Goodbye, {{name}}!",
	}, map[string]interface{}{
		"name": "World",
	})
	if err != nil {
		t.Fatalf("ProcessMany failed: %v", err)
	}
	if len(results) != 2 || results[0] != "Hello, World!" || results[1] != "Goodbye, World!" {
		t.Errorf("ProcessMany returned unexpected results: %v", results)
	}
}