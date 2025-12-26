package templates

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewTemplateLibrary(t *testing.T) {
	lib := NewTemplateLibrary("/tmp")
	if lib == nil {
		t.Fatal("NewTemplateLibrary returned nil")
	}
	if lib.templates == nil {
		t.Error("templates map should be initialized")
	}
	if lib.indexes == nil {
		t.Error("indexes map should be initialized")
	}
	if lib.basePath != "/tmp" {
		t.Errorf("expected basePath /tmp, got %s", lib.basePath)
	}
}

func TestTemplateLibrary_AddTemplate(t *testing.T) {
	lib := NewTemplateLibrary("")

	template := &Template{
		Name:        "test-template",
		Description: "A test template",
		Category:    "test",
		Tags:        []string{"testing", "example"},
	}

	err := lib.AddTemplate(template)
	if err != nil {
		t.Fatalf("AddTemplate failed: %v", err)
	}

	// Verify template is in library
	got, err := lib.GetTemplate("test-template")
	if err != nil {
		t.Fatalf("GetTemplate failed: %v", err)
	}
	if got.Name != template.Name {
		t.Errorf("expected name %s, got %s", template.Name, got.Name)
	}
}

func TestTemplateLibrary_AddTemplate_MissingName(t *testing.T) {
	lib := NewTemplateLibrary("")

	template := &Template{
		Description: "A test template without name",
	}

	err := lib.AddTemplate(template)
	if err == nil {
		t.Error("expected error for template without name")
	}
}

func TestTemplateLibrary_GetTemplate_NotFound(t *testing.T) {
	lib := NewTemplateLibrary("")

	_, err := lib.GetTemplate("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent template")
	}
}

func TestTemplateLibrary_ListTemplates(t *testing.T) {
	lib := NewTemplateLibrary("")

	// Add some templates
	lib.AddTemplate(&Template{Name: "template1"})
	lib.AddTemplate(&Template{Name: "template2"})

	templates := lib.ListTemplates()
	if len(templates) != 2 {
		t.Errorf("expected 2 templates, got %d", len(templates))
	}
}

func TestTemplateLibrary_Search(t *testing.T) {
	lib := NewTemplateLibrary("")

	lib.AddTemplate(&Template{
		Name:        "code-review",
		Description: "Review code for quality",
		Category:    "development",
		Tags:        []string{"code", "review"},
	})
	lib.AddTemplate(&Template{
		Name:        "summarize",
		Description: "Summarize text content",
		Category:    "text",
		Tags:        []string{"summary"},
	})

	// Search by name
	results := lib.Search("code")
	if len(results) != 1 {
		t.Errorf("expected 1 result for 'code', got %d", len(results))
	}

	// Search by description
	results = lib.Search("quality")
	if len(results) != 1 {
		t.Errorf("expected 1 result for 'quality', got %d", len(results))
	}

	// Search with no matches
	results = lib.Search("nonexistent")
	if len(results) != 0 {
		t.Errorf("expected 0 results for 'nonexistent', got %d", len(results))
	}
}

func TestTemplateLibrary_GetByCategory(t *testing.T) {
	lib := NewTemplateLibrary("")

	lib.AddTemplate(&Template{Name: "dev1", Category: "development"})
	lib.AddTemplate(&Template{Name: "dev2", Category: "development"})
	lib.AddTemplate(&Template{Name: "text1", Category: "text"})

	devTemplates := lib.GetByCategory("development")
	if len(devTemplates) != 2 {
		t.Errorf("expected 2 development templates, got %d", len(devTemplates))
	}
}

func TestTemplateLibrary_GetByTag(t *testing.T) {
	lib := NewTemplateLibrary("")

	lib.AddTemplate(&Template{Name: "template1", Tags: []string{"code", "review"}})
	lib.AddTemplate(&Template{Name: "template2", Tags: []string{"code", "generate"}})
	lib.AddTemplate(&Template{Name: "template3", Tags: []string{"text"}})

	codeTemplates := lib.GetByTag("code")
	if len(codeTemplates) != 2 {
		t.Errorf("expected 2 templates with 'code' tag, got %d", len(codeTemplates))
	}
}

func TestTemplateLibrary_GetCategories(t *testing.T) {
	lib := NewTemplateLibrary("")

	lib.AddTemplate(&Template{Name: "t1", Category: "development"})
	lib.AddTemplate(&Template{Name: "t2", Category: "text"})
	lib.AddTemplate(&Template{Name: "t3", Category: "development"}) // duplicate category

	categories := lib.GetCategories()
	if len(categories) != 2 {
		t.Errorf("expected 2 categories, got %d", len(categories))
	}
}

func TestTemplateLibrary_GetTags(t *testing.T) {
	lib := NewTemplateLibrary("")

	lib.AddTemplate(&Template{Name: "t1", Tags: []string{"code", "review"}})
	lib.AddTemplate(&Template{Name: "t2", Tags: []string{"text", "summary"}})

	tags := lib.GetTags()
	if len(tags) < 2 {
		t.Errorf("expected at least 2 tags, got %d", len(tags))
	}
}

func TestTemplateLibrary_ApplyTemplate(t *testing.T) {
	lib := NewTemplateLibrary("")

	lib.AddTemplate(&Template{
		Name:   "greeting",
		Prompt: "Hello, {{name}}! Welcome to {{place}}.",
		Variables: map[string]Variable{
			"name":  {Name: "name", Type: "string", Required: true},
			"place": {Name: "place", Type: "string", Required: true},
		},
	})

	result, err := lib.ApplyTemplate("greeting", map[string]interface{}{
		"name":  "World",
		"place": "PE",
	})
	if err != nil {
		t.Fatalf("ApplyTemplate failed: %v", err)
	}

	expected := "Hello, World! Welcome to PE."
	if result != expected {
		t.Errorf("expected '%s', got '%s'", expected, result)
	}
}

func TestTemplateLibrary_ApplyTemplate_MissingRequired(t *testing.T) {
	lib := NewTemplateLibrary("")

	lib.AddTemplate(&Template{
		Name:   "greeting",
		Prompt: "Hello, {{name}}!",
		Variables: map[string]Variable{
			"name": {Name: "name", Type: "string", Required: true},
		},
	})

	_, err := lib.ApplyTemplate("greeting", map[string]interface{}{})
	if err == nil {
		t.Error("expected error for missing required variable")
	}
}

func TestTemplateLibrary_ApplyTemplate_NotFound(t *testing.T) {
	lib := NewTemplateLibrary("")

	_, err := lib.ApplyTemplate("nonexistent", map[string]interface{}{})
	if err == nil {
		t.Error("expected error for nonexistent template")
	}
}

func TestTemplateLibrary_ExportTemplate(t *testing.T) {
	lib := NewTemplateLibrary("")

	lib.AddTemplate(&Template{
		Name:        "test",
		Description: "Test template",
	})

	// Export as JSON
	jsonData, err := lib.ExportTemplate("test", "json")
	if err != nil {
		t.Fatalf("ExportTemplate JSON failed: %v", err)
	}
	if len(jsonData) == 0 {
		t.Error("JSON export is empty")
	}

	// Export as YAML
	yamlData, err := lib.ExportTemplate("test", "yaml")
	if err != nil {
		t.Fatalf("ExportTemplate YAML failed: %v", err)
	}
	if len(yamlData) == 0 {
		t.Error("YAML export is empty")
	}

	// Export unsupported format
	_, err = lib.ExportTemplate("test", "xml")
	if err == nil {
		t.Error("expected error for unsupported format")
	}
}

func TestTemplateLibrary_ExportTemplate_NotFound(t *testing.T) {
	lib := NewTemplateLibrary("")

	_, err := lib.ExportTemplate("nonexistent", "json")
	if err == nil {
		t.Error("expected error for nonexistent template")
	}
}

func TestTemplateLibrary_SaveTemplate(t *testing.T) {
	lib := NewTemplateLibrary("")
	tmpDir := t.TempDir()

	template := &Template{
		Name:        "test",
		Description: "Test template",
	}

	filename := filepath.Join(tmpDir, "test.yaml")
	err := lib.SaveTemplate(template, filename)
	if err != nil {
		t.Fatalf("SaveTemplate failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		t.Error("template file was not created")
	}
}

func TestTemplateLibrary_LoadTemplateFromFile(t *testing.T) {
	lib := NewTemplateLibrary("")
	tmpDir := t.TempDir()

	// Create a template file
	templateYAML := `
name: test-template
description: A test template
category: test
prompt: "Hello {{name}}"
`
	filename := filepath.Join(tmpDir, "test.yaml")
	if err := os.WriteFile(filename, []byte(templateYAML), 0644); err != nil {
		t.Fatalf("failed to write template file: %v", err)
	}

	template, err := lib.LoadTemplateFromFile(filename)
	if err != nil {
		t.Fatalf("LoadTemplateFromFile failed: %v", err)
	}

	if template.Name != "test-template" {
		t.Errorf("expected name 'test-template', got '%s'", template.Name)
	}
	if template.Category != "test" {
		t.Errorf("expected category 'test', got '%s'", template.Category)
	}
}

func TestTemplateLibrary_LoadFromDirectory(t *testing.T) {
	lib := NewTemplateLibrary("")
	tmpDir := t.TempDir()

	// Create template files
	template1 := `
name: template1
description: First template
category: test
`
	template2 := `
name: template2
description: Second template
category: test
`
	os.WriteFile(filepath.Join(tmpDir, "t1.yaml"), []byte(template1), 0644)
	os.WriteFile(filepath.Join(tmpDir, "t2.yml"), []byte(template2), 0644)

	err := lib.LoadFromDirectory(tmpDir)
	if err != nil {
		t.Fatalf("LoadFromDirectory failed: %v", err)
	}

	templates := lib.ListTemplates()
	if len(templates) != 2 {
		t.Errorf("expected 2 templates, got %d", len(templates))
	}
}

func TestTemplateLibrary_LoadBuiltinTemplates(t *testing.T) {
	lib := NewTemplateLibrary("")

	err := lib.LoadBuiltinTemplates()
	if err != nil {
		t.Fatalf("LoadBuiltinTemplates failed: %v", err)
	}

	templates := lib.ListTemplates()
	if len(templates) == 0 {
		t.Error("expected some builtin templates to be loaded")
	}
}

func TestTemplateLibrary_CreateTemplate(t *testing.T) {
	lib := NewTemplateLibrary("")

	template, err := lib.CreateTemplate(nil)
	if err != nil {
		t.Fatalf("CreateTemplate failed: %v", err)
	}

	if template == nil {
		t.Fatal("CreateTemplate returned nil")
	}
	if template.Name != "new-template" {
		t.Errorf("expected name 'new-template', got '%s'", template.Name)
	}
}

func TestTemplateLibrary_validateVariable_String(t *testing.T) {
	lib := NewTemplateLibrary("")

	varDef := Variable{
		Type: "string",
		Validation: Validation{
			MinLength: 2,
			MaxLength: 10,
		},
	}

	// Valid string
	err := lib.validateVariable(varDef, "hello")
	if err != nil {
		t.Errorf("unexpected error for valid string: %v", err)
	}

	// Too short
	err = lib.validateVariable(varDef, "a")
	if err == nil {
		t.Error("expected error for string too short")
	}

	// Too long
	err = lib.validateVariable(varDef, "this is way too long")
	if err == nil {
		t.Error("expected error for string too long")
	}

	// Wrong type
	err = lib.validateVariable(varDef, 123)
	if err == nil {
		t.Error("expected error for wrong type")
	}
}

func TestTemplateLibrary_validateVariable_Number(t *testing.T) {
	lib := NewTemplateLibrary("")

	varDef := Variable{
		Type: "number",
		Validation: Validation{
			Min: 1,
			Max: 100,
		},
	}

	// Valid int
	err := lib.validateVariable(varDef, 50)
	if err != nil {
		t.Errorf("unexpected error for valid int: %v", err)
	}

	// Valid float
	err = lib.validateVariable(varDef, 50.5)
	if err != nil {
		t.Errorf("unexpected error for valid float: %v", err)
	}

	// Too small
	err = lib.validateVariable(varDef, 0)
	if err == nil {
		t.Error("expected error for number too small")
	}

	// Too large
	err = lib.validateVariable(varDef, 200)
	if err == nil {
		t.Error("expected error for number too large")
	}
}

func TestTemplateLibrary_validateVariable_Boolean(t *testing.T) {
	lib := NewTemplateLibrary("")

	varDef := Variable{
		Type: "boolean",
	}

	// Valid boolean
	err := lib.validateVariable(varDef, true)
	if err != nil {
		t.Errorf("unexpected error for valid boolean: %v", err)
	}

	// Wrong type
	err = lib.validateVariable(varDef, "true")
	if err == nil {
		t.Error("expected error for wrong type")
	}
}

func TestGetDefaultLibrary(t *testing.T) {
	lib := GetDefaultLibrary()
	if lib == nil {
		t.Error("GetDefaultLibrary returned nil")
	}
}

func TestInitializeDefaultLibrary(t *testing.T) {
	err := InitializeDefaultLibrary()
	if err != nil {
		t.Errorf("InitializeDefaultLibrary failed: %v", err)
	}
}

func TestGetBuiltinTemplates(t *testing.T) {
	templates := getBuiltinTemplates()
	if len(templates) == 0 {
		t.Error("expected some builtin templates")
	}

	// Verify each template has required fields
	for _, tmpl := range templates {
		if tmpl.Name == "" {
			t.Error("builtin template missing name")
		}
		if tmpl.Prompt == "" {
			t.Error("builtin template missing prompt")
		}
	}
}

func TestTemplate_Struct(t *testing.T) {
	tmpl := Template{
		Name:        "test",
		Description: "Test description",
		Category:    "test",
		Tags:        []string{"tag1", "tag2"},
		Author:      "tester",
		Version:     "1.0.0",
		Prompt:      "Hello {{name}}",
		Variables: map[string]Variable{
			"name": {Name: "name", Type: "string", Required: true},
		},
	}

	if tmpl.Name != "test" {
		t.Error("Name mismatch")
	}
	if len(tmpl.Tags) != 2 {
		t.Error("Tags length mismatch")
	}
	if len(tmpl.Variables) != 1 {
		t.Error("Variables length mismatch")
	}
}

func TestVariable_Struct(t *testing.T) {
	v := Variable{
		Name:        "test",
		Description: "Test variable",
		Type:        "string",
		Required:    true,
		Default:     "default",
		Examples:    []string{"ex1", "ex2"},
		Validation: Validation{
			MinLength: 1,
			MaxLength: 100,
		},
	}

	if v.Name != "test" {
		t.Error("Name mismatch")
	}
	if !v.Required {
		t.Error("Required should be true")
	}
	if v.Validation.MinLength != 1 {
		t.Error("MinLength mismatch")
	}
}

func TestExample_Struct(t *testing.T) {
	ex := Example{
		Name:        "test",
		Description: "Test example",
		Variables: map[string]interface{}{
			"key": "value",
		},
		Expected: "expected output",
	}

	if ex.Name != "test" {
		t.Error("Name mismatch")
	}
	if len(ex.Variables) != 1 {
		t.Error("Variables length mismatch")
	}
}
