package templates

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"sigs.k8s.io/yaml"
)

// Template represents a prompt template
type Template struct {
	Name         string                 `yaml:"name" json:"name"`
	Description  string                 `yaml:"description" json:"description"`
	Category     string                 `yaml:"category" json:"category"`
	Tags         []string               `yaml:"tags" json:"tags"`
	Author       string                 `yaml:"author" json:"author"`
	Version      string                 `yaml:"version" json:"version"`
	License      string                 `yaml:"license" json:"license"`
	CreatedAt    time.Time              `yaml:"created_at" json:"created_at"`
	UpdatedAt    time.Time              `yaml:"updated_at" json:"updated_at"`
	Prompt       string                 `yaml:"prompt" json:"prompt"`
	Variables    map[string]Variable    `yaml:"variables" json:"variables"`
	Examples     []Example              `yaml:"examples" json:"examples"`
	Providers    []string               `yaml:"providers" json:"providers"`
	Options      map[string]interface{} `yaml:"options" json:"options"`
	Requirements []string               `yaml:"requirements" json:"requirements"`
	UsageCount   int                    `yaml:"usage_count" json:"usage_count"`
	Rating       float64                `yaml:"rating" json:"rating"`
}

// Variable describes a template variable
type Variable struct {
	Name        string      `yaml:"name" json:"name"`
	Description string      `yaml:"description" json:"description"`
	Type        string      `yaml:"type" json:"type"` // string, number, boolean, array, object
	Required    bool        `yaml:"required" json:"required"`
	Default     interface{} `yaml:"default" json:"default"`
	Examples    []string    `yaml:"examples" json:"examples"`
	Validation  Validation  `yaml:"validation" json:"validation"`
}

// Validation contains validation rules for variables
type Validation struct {
	MinLength int      `yaml:"min_length" json:"min_length"`
	MaxLength int      `yaml:"max_length" json:"max_length"`
	Pattern   string   `yaml:"pattern" json:"pattern"`
	Options   []string `yaml:"options" json:"options"`
	Min       float64  `yaml:"min" json:"min"`
	Max       float64  `yaml:"max" json:"max"`
}

// Example provides usage examples for the template
type Example struct {
	Name        string                 `yaml:"name" json:"name"`
	Description string                 `yaml:"description" json:"description"`
	Variables   map[string]interface{} `yaml:"variables" json:"variables"`
	Expected    string                 `yaml:"expected" json:"expected"`
}

// TemplateLibrary manages a collection of templates
type TemplateLibrary struct {
	templates map[string]*Template
	indexes   map[string][]string // category/tag -> template names
	basePath  string
}

// NewTemplateLibrary creates a new template library
func NewTemplateLibrary(basePath string) *TemplateLibrary {
	return &TemplateLibrary{
		templates: make(map[string]*Template),
		indexes:   make(map[string][]string),
		basePath:  basePath,
	}
}

// LoadBuiltinTemplates loads built-in templates
func (tl *TemplateLibrary) LoadBuiltinTemplates() error {
	builtinTemplates := getBuiltinTemplates()

	for _, template := range builtinTemplates {
		if err := tl.AddTemplate(template); err != nil {
			return fmt.Errorf("failed to load builtin template %s: %v", template.Name, err)
		}
	}

	return nil
}

// LoadFromDirectory loads templates from a directory
func (tl *TemplateLibrary) LoadFromDirectory(dir string) error {
	return filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		// Only process .yaml and .yml files
		ext := strings.ToLower(filepath.Ext(path))
		if ext != ".yaml" && ext != ".yml" {
			return nil
		}

		template, err := tl.LoadTemplateFromFile(path)
		if err != nil {
			return fmt.Errorf("failed to load template from %s: %v", path, err)
		}

		return tl.AddTemplate(template)
	})
}

// LoadTemplateFromFile loads a template from a file
func (tl *TemplateLibrary) LoadTemplateFromFile(filename string) (*Template, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var template Template
	if err := yaml.Unmarshal(data, &template); err != nil {
		return nil, err
	}

	// Set defaults
	if template.CreatedAt.IsZero() {
		template.CreatedAt = time.Now()
	}
	if template.UpdatedAt.IsZero() {
		template.UpdatedAt = time.Now()
	}
	if template.Version == "" {
		template.Version = "1.0.0"
	}

	return &template, nil
}

// AddTemplate adds a template to the library
func (tl *TemplateLibrary) AddTemplate(template *Template) error {
	if template.Name == "" {
		return fmt.Errorf("template name is required")
	}

	tl.templates[template.Name] = template

	// Update indexes
	if template.Category != "" {
		tl.addToIndex("category:"+template.Category, template.Name)
	}

	for _, tag := range template.Tags {
		tl.addToIndex("tag:"+tag, template.Name)
	}

	return nil
}

// GetTemplate retrieves a template by name
func (tl *TemplateLibrary) GetTemplate(name string) (*Template, error) {
	template, exists := tl.templates[name]
	if !exists {
		return nil, fmt.Errorf("template not found: %s", name)
	}

	return template, nil
}

// ListTemplates returns all templates
func (tl *TemplateLibrary) ListTemplates() []*Template {
	templates := make([]*Template, 0, len(tl.templates))
	for _, template := range tl.templates {
		templates = append(templates, template)
	}
	return templates
}

// Search searches for templates based on query
func (tl *TemplateLibrary) Search(query string) []*Template {
	var results []*Template

	query = strings.ToLower(query)

	for _, template := range tl.templates {
		if tl.matchesQuery(template, query) {
			results = append(results, template)
		}
	}

	return results
}

// GetByCategory returns templates in a specific category
func (tl *TemplateLibrary) GetByCategory(category string) []*Template {
	names := tl.indexes["category:"+category]
	templates := make([]*Template, 0, len(names))

	for _, name := range names {
		if template, exists := tl.templates[name]; exists {
			templates = append(templates, template)
		}
	}

	return templates
}

// GetByTag returns templates with a specific tag
func (tl *TemplateLibrary) GetByTag(tag string) []*Template {
	names := tl.indexes["tag:"+tag]
	templates := make([]*Template, 0, len(names))

	for _, name := range names {
		if template, exists := tl.templates[name]; exists {
			templates = append(templates, template)
		}
	}

	return templates
}

// GetCategories returns all available categories
func (tl *TemplateLibrary) GetCategories() []string {
	var categories []string
	for key := range tl.indexes {
		if strings.HasPrefix(key, "category:") {
			categories = append(categories, strings.TrimPrefix(key, "category:"))
		}
	}
	return categories
}

// GetTags returns all available tags
func (tl *TemplateLibrary) GetTags() []string {
	var tags []string
	for key := range tl.indexes {
		if strings.HasPrefix(key, "tag:") {
			tags = append(tags, strings.TrimPrefix(key, "tag:"))
		}
	}
	return tags
}

// ApplyTemplate applies a template with given variables
func (tl *TemplateLibrary) ApplyTemplate(name string, variables map[string]interface{}) (string, error) {
	template, err := tl.GetTemplate(name)
	if err != nil {
		return "", err
	}

	// Validate variables
	if err := tl.validateVariables(template, variables); err != nil {
		return "", err
	}

	// Apply template substitution
	result := template.Prompt
	for varName, value := range variables {
		placeholder := fmt.Sprintf("{{%s}}", varName)
		result = strings.ReplaceAll(result, placeholder, fmt.Sprintf("%v", value))
	}

	// Update usage count
	template.UsageCount++

	return result, nil
}

// SaveTemplate saves a template to file
func (tl *TemplateLibrary) SaveTemplate(template *Template, filename string) error {
	template.UpdatedAt = time.Now()

	data, err := yaml.Marshal(template)
	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, 0644)
}

// ExportTemplate exports a template to various formats
func (tl *TemplateLibrary) ExportTemplate(name, format string) ([]byte, error) {
	template, err := tl.GetTemplate(name)
	if err != nil {
		return nil, err
	}

	switch strings.ToLower(format) {
	case "json":
		return json.MarshalIndent(template, "", "  ")
	case "yaml", "yml":
		return yaml.Marshal(template)
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
}

// CreateTemplate creates a new template interactively
func (tl *TemplateLibrary) CreateTemplate(ctx context.Context) (*Template, error) {
	// This would be implemented with interactive prompts
	// For now, return a basic template structure
	return &Template{
		Name:        "new-template",
		Description: "New template",
		Category:    "general",
		Tags:        []string{"custom"},
		Version:     "1.0.0",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Prompt:      "{{input}}",
		Variables: map[string]Variable{
			"input": {
				Name:        "input",
				Description: "Input text",
				Type:        "string",
				Required:    true,
			},
		},
	}, nil
}

// Helper methods

func (tl *TemplateLibrary) addToIndex(key, value string) {
	if _, exists := tl.indexes[key]; !exists {
		tl.indexes[key] = []string{}
	}
	tl.indexes[key] = append(tl.indexes[key], value)
}

func (tl *TemplateLibrary) matchesQuery(template *Template, query string) bool {
	// Search in name, description, category, and tags
	fields := []string{
		strings.ToLower(template.Name),
		strings.ToLower(template.Description),
		strings.ToLower(template.Category),
		strings.ToLower(strings.Join(template.Tags, " ")),
	}

	for _, field := range fields {
		if strings.Contains(field, query) {
			return true
		}
	}

	return false
}

func (tl *TemplateLibrary) validateVariables(template *Template, variables map[string]interface{}) error {
	// Check required variables
	for varName, varDef := range template.Variables {
		if varDef.Required {
			if _, exists := variables[varName]; !exists {
				return fmt.Errorf("required variable missing: %s", varName)
			}
		}
	}

	// Validate variable types and constraints
	for varName, value := range variables {
		if varDef, exists := template.Variables[varName]; exists {
			if err := tl.validateVariable(varDef, value); err != nil {
				return fmt.Errorf("validation failed for variable %s: %v", varName, err)
			}
		}
	}

	return nil
}

func (tl *TemplateLibrary) validateVariable(varDef Variable, value interface{}) error {
	// Type validation
	switch varDef.Type {
	case "string":
		if _, ok := value.(string); !ok {
			return fmt.Errorf("expected string, got %T", value)
		}

		str := value.(string)
		if varDef.Validation.MinLength > 0 && len(str) < varDef.Validation.MinLength {
			return fmt.Errorf("string too short (min: %d)", varDef.Validation.MinLength)
		}
		if varDef.Validation.MaxLength > 0 && len(str) > varDef.Validation.MaxLength {
			return fmt.Errorf("string too long (max: %d)", varDef.Validation.MaxLength)
		}

	case "number":
		var num float64
		switch v := value.(type) {
		case int:
			num = float64(v)
		case float64:
			num = v
		default:
			return fmt.Errorf("expected number, got %T", value)
		}

		if varDef.Validation.Min != 0 && num < varDef.Validation.Min {
			return fmt.Errorf("number too small (min: %f)", varDef.Validation.Min)
		}
		if varDef.Validation.Max != 0 && num > varDef.Validation.Max {
			return fmt.Errorf("number too large (max: %f)", varDef.Validation.Max)
		}

	case "boolean":
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("expected boolean, got %T", value)
		}
	}

	return nil
}

// Global template library instance
var defaultLibrary = NewTemplateLibrary("")

// GetDefaultLibrary returns the default template library
func GetDefaultLibrary() *TemplateLibrary {
	return defaultLibrary
}

// InitializeDefaultLibrary initializes the default library with builtin templates
func InitializeDefaultLibrary() error {
	return defaultLibrary.LoadBuiltinTemplates()
}
