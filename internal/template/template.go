// Package template provides utilities for working with templated prompts.
package template

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
)

// Template represents a structured template with variables and file inclusions.
type Template struct {
	// Content is the template content
	Content string
	// BaseDir is the base directory for resolving relative file paths
	BaseDir string
	// Variables is a map of variable names to their values
	Variables map[string]interface{}
	// FileProcessor handles file inclusions
	fileProcessor *FileProcessor
	// VariableProcessor handles variable substitution
	variableProcessor *VariableProcessor
	// MaxDepth controls recursion depth for conditional processing
	MaxDepth int
}

// ErrMissingVariable is returned when a required variable is missing.
type ErrMissingVariable struct {
	Variable string
}

// Error implements the error interface for ErrMissingVariable.
func (e ErrMissingVariable) Error() string {
	return fmt.Sprintf("missing required variable: %s", e.Variable)
}

// NewTemplate creates a new Template with the given content and variables.
func NewTemplate(content string, vars map[string]interface{}) *Template {
	workingDir, _ := os.Getwd()
	return &Template{
		Content:           content,
		BaseDir:           workingDir,
		Variables:         vars,
		fileProcessor:     NewFileProcessor(workingDir),
		variableProcessor: NewVariableProcessor(vars),
		MaxDepth:          10,
	}
}

// SetBaseDir sets the base directory for resolving relative file paths.
func (t *Template) SetBaseDir(baseDir string) {
	t.BaseDir = baseDir
	t.fileProcessor.BaseDir = baseDir
}

// Validate checks if the template is valid.
func (t *Template) Validate() []string {
	var errors []string
	
	// Check for missing variables
	missingVars := ValidateVariables(t.Content, t.Variables)
	for _, v := range missingVars {
		errors = append(errors, fmt.Sprintf("missing variable: %s", v))
	}
	
	// Check for missing files
	missingFiles := ValidateFileInclusions(t.Content, t.BaseDir)
	for _, f := range missingFiles {
		errors = append(errors, fmt.Sprintf("missing file: %s", f))
	}
	
	return errors
}

// Process processes the template with variables, file inclusions, and conditionals.
func (t *Template) Process() (string, error) {
	// Step 1: Process file inclusions
	contentWithFiles, err := t.fileProcessor.Process(t.Content)
	if err != nil {
		return "", fmt.Errorf("error processing file inclusions: %w", err)
	}
	
	// Step 2: Process conditionals
	contentWithConditionals, err := t.processConditionals(contentWithFiles, 0)
	if err != nil {
		return "", fmt.Errorf("error processing conditionals: %w", err)
	}
	
	// Step 3: Process variables
	result, err := t.variableProcessor.Process(contentWithConditionals)
	if err != nil {
		return "", fmt.Errorf("error processing variables: %w", err)
	}
	
	return result, nil
}

// processConditionals processes conditional blocks in the template.
// Supports #if var then ... #else ... #endif syntax.
func (t *Template) processConditionals(content string, depth int) (string, error) {
	if depth > t.MaxDepth {
		return "", fmt.Errorf("maximum conditional nesting depth reached (%d)", t.MaxDepth)
	}
	
	// Regex to match conditional blocks
	// Matches: #if condition then content [#else alternative] #endif
	re := regexp.MustCompile(`#if\s+([^#]+?)\s+then\s+((?:.|\n)*?)(?:#else\s+((?:.|\n)*?))?#endif`)
	
	// Process all conditional blocks
	result := re.ReplaceAllStringFunc(content, func(match string) string {
		parts := re.FindStringSubmatch(match)
		if len(parts) < 3 {
			return match // Keep original if pattern doesn't match correctly
		}
		
		condition := strings.TrimSpace(parts[1])
		thenContent := parts[2]
		elseContent := ""
		if len(parts) > 3 {
			elseContent = parts[3]
		}
		
		// Evaluate the condition
		conditionResult, err := EvaluateExpression(condition, t.Variables)
		if err != nil {
			// Return an error placeholder that will be easy to identify
			return fmt.Sprintf("ERROR_EVALUATING_CONDITION:%s", condition)
		}
		
		var selectedContent string
		if conditionResult {
			selectedContent = thenContent
		} else {
			selectedContent = elseContent
		}
		
		// Process nested conditionals
		processed, err := t.processConditionals(selectedContent, depth+1)
		if err != nil {
			return fmt.Sprintf("ERROR_PROCESSING_NESTED_CONDITIONAL:%s", err.Error())
		}
		
		return processed
	})
	
	// Check for any error placeholders
	if strings.Contains(result, "ERROR_EVALUATING_CONDITION:") {
		re := regexp.MustCompile(`ERROR_EVALUATING_CONDITION:([^\s]+)`)
		match := re.FindStringSubmatch(result)
		if len(match) >= 2 {
			return "", fmt.Errorf("error evaluating condition: %s", match[1])
		}
	}
	
	if strings.Contains(result, "ERROR_PROCESSING_NESTED_CONDITIONAL:") {
		re := regexp.MustCompile(`ERROR_PROCESSING_NESTED_CONDITIONAL:(.+)`)
		match := re.FindStringSubmatch(result)
		if len(match) >= 2 {
			return "", fmt.Errorf("error processing nested conditional: %s", match[1])
		}
	}
	
	return result, nil
}

// Legacy functions for backward compatibility

// ProcessPrompt processes a templated prompt with the given variables.
func ProcessPrompt(promptTemplate string, vars map[string]interface{}) (string, error) {
	// Check for simple {{var}} style templates first
	missingVars := checkSimpleVars(promptTemplate, vars)
	if len(missingVars) > 0 {
		return "", ErrMissingVariable{Variable: missingVars[0]}
	}
	
	// Replace simple {{var}} templates with values
	result := promptTemplate
	for varName, varValue := range vars {
		var strValue string
		
		// Convert value to string
		switch v := varValue.(type) {
		case string:
			strValue = v
		case []byte:
			strValue = string(v)
		case json.Marshaler:
			data, err := v.MarshalJSON()
			if err != nil {
				return "", fmt.Errorf("error marshaling JSON for variable %s: %w", varName, err)
			}
			strValue = string(data)
		default:
			strValue = fmt.Sprintf("%v", v)
		}
		
		// Replace {{varName}} with strValue
		placeholder := fmt.Sprintf("{{%s}}", varName)
		result = strings.ReplaceAll(result, placeholder, strValue)
	}
	
	// Try parsing as Go template for more complex templates
	if strings.Contains(result, "{{") {
		var buf bytes.Buffer
		tmpl, err := template.New("prompt").Parse(promptTemplate)
		if err != nil {
			return "", fmt.Errorf("error parsing template: %w", err)
		}
		
		if err := tmpl.Execute(&buf, vars); err != nil {
			return "", fmt.Errorf("error executing template: %w", err)
		}
		
		result = buf.String()
	}
	
	return result, nil
}

// ProcessPromptWithFiles processes a templated prompt with variables and file inclusions.
// Supports {{file "/path/to/file"}} syntax for file inclusion.
func ProcessPromptWithFiles(promptTemplate string, vars map[string]interface{}) (string, error) {
	// Using the new structured approach
	tmpl := NewTemplate(promptTemplate, vars)
	return tmpl.Process()
}

// ProcessMany processes multiple templated prompts with the same variables.
func ProcessMany(promptTemplates []string, vars map[string]interface{}) ([]string, error) {
	results := make([]string, 0, len(promptTemplates))
	
	for _, promptTemplate := range promptTemplates {
		processed, err := ProcessPrompt(promptTemplate, vars)
		if err != nil {
			return nil, fmt.Errorf("error processing prompt template: %w", err)
		}
		
		results = append(results, processed)
	}
	
	return results, nil
}

// ProcessManyWithFiles processes multiple templated prompts with variables and file inclusions.
func ProcessManyWithFiles(promptTemplates []string, vars map[string]interface{}) ([]string, error) {
	results := make([]string, 0, len(promptTemplates))
	
	for _, promptTemplate := range promptTemplates {
		processed, err := ProcessPromptWithFiles(promptTemplate, vars)
		if err != nil {
			return nil, fmt.Errorf("error processing prompt template with files: %w", err)
		}
		
		results = append(results, processed)
	}
	
	return results, nil
}

// Private helper functions

// checkSimpleVars checks for simple {{var}} variables in the template and returns a list of
// missing variables.
func checkSimpleVars(promptTemplate string, vars map[string]interface{}) []string {
	var missingVars []string
	
	// Find all {{varName}} patterns
	re := regexp.MustCompile(`{{([^{}]+)}}`)
	matches := re.FindAllStringSubmatch(promptTemplate, -1)
	
	for _, match := range matches {
		if len(match) > 1 {
			varName := strings.TrimSpace(match[1])
			
			// Skip file inclusions and conditionals
			if strings.HasPrefix(varName, "file ") || 
			   strings.HasPrefix(varName, "if ") ||
			   strings.HasPrefix(varName, "else") ||
			   strings.HasPrefix(varName, "endif") {
				continue
			}
			
			// Check if variable exists in vars map
			if _, exists := vars[varName]; !exists {
				missingVars = append(missingVars, varName)
			}
		}
	}
	
	return missingVars
}

// processFileInclusions processes file inclusion directives in the template.
// Supports {{file "/path/to/file"}} syntax.
// This is kept for backward compatibility; new code should use FileProcessor.
func processFileInclusions(promptTemplate string) (string, error) {
	baseDir, _ := os.Getwd()
	// We're ensuring path/filepath is used to avoid import error
	baseDir = filepath.Clean(baseDir)
	processor := NewFileProcessor(baseDir)
	return processor.Process(promptTemplate)
}