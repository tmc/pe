// Package template provides utilities for working with templated prompts.
package template

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
	"text/template"
)

// ErrMissingVariable is returned when a required variable is missing.
type ErrMissingVariable struct {
	Variable string
}

// Error implements the error interface for ErrMissingVariable.
func (e ErrMissingVariable) Error() string {
	return fmt.Sprintf("missing required variable: %s", e.Variable)
}

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
	// Process file inclusions first
	processed, err := processFileInclusions(promptTemplate)
	if err != nil {
		return "", err
	}
	
	// Then process variables
	return ProcessPrompt(processed, vars)
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
			
			// Skip file inclusions
			if strings.HasPrefix(varName, "file ") {
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
func processFileInclusions(promptTemplate string) (string, error) {
	re := regexp.MustCompile(`{{file\s+"([^"]+)"}}`)
	
	processed := re.ReplaceAllStringFunc(promptTemplate, func(match string) string {
		fileMatch := re.FindStringSubmatch(match)
		if len(fileMatch) < 2 {
			return match // Keep original if no file path found
		}
		
		filePath := fileMatch[1]
		content, err := os.ReadFile(filePath)
		if err != nil {
			// Return an error placeholder that will be easy to identify
			return fmt.Sprintf("ERROR_READING_FILE:%s", filePath)
		}
		
		return string(content)
	})
	
	// Check for any file error placeholders
	if strings.Contains(processed, "ERROR_READING_FILE:") {
		re := regexp.MustCompile(`ERROR_READING_FILE:([^\s]+)`)
		match := re.FindStringSubmatch(processed)
		if len(match) >= 2 {
			return "", fmt.Errorf("error reading file: %s", match[1])
		}
	}
	
	return processed, nil
}