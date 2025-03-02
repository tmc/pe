// Package template provides utilities for working with templated prompts.
package template

import (
	"fmt"
	"regexp"
	"strings"
)

// VariableProcessor handles variable substitution in templates.
type VariableProcessor struct {
	// Maps variable names to their values
	Variables map[string]interface{}
	// Escape character for variables (default: \)
	EscapeChar string
}

// NewVariableProcessor creates a new VariableProcessor with the given variables.
func NewVariableProcessor(vars map[string]interface{}) *VariableProcessor {
	return &VariableProcessor{
		Variables:  vars,
		EscapeChar: "\\",
	}
}

// Process substitutes variables in the given text.
func (p *VariableProcessor) Process(text string) (string, error) {
	// Handle escaped variables - temporarily replace them
	re := regexp.MustCompile(fmt.Sprintf("%s{{([^{}]+)}}", regexp.QuoteMeta(p.EscapeChar)))
	escapedVars := make(map[string]string)
	
	count := 0
	escapedText := re.ReplaceAllStringFunc(text, func(match string) string {
		placeholder := fmt.Sprintf("__ESCAPED_VAR_%d__", count)
		escapedVars[placeholder] = match
		count++
		return placeholder
	})
	
	// Process regular variables
	result, err := ProcessPrompt(escapedText, p.Variables)
	if err != nil {
		return "", err
	}
	
	// Restore escaped variables, but remove the escape character
	for placeholder, escapedVar := range escapedVars {
		result = strings.ReplaceAll(result, placeholder, strings.TrimPrefix(escapedVar, p.EscapeChar))
	}
	
	return result, nil
}

// ExtractVariables extracts all variable names from a template.
func ExtractVariables(text string) []string {
	re := regexp.MustCompile(`{{([^{}]+)}}`)
	matches := re.FindAllStringSubmatch(text, -1)
	
	var variables []string
	seen := make(map[string]bool)
	
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
			
			if !seen[varName] {
				variables = append(variables, varName)
				seen[varName] = true
			}
		}
	}
	
	return variables
}

// ValidateVariables checks if all required variables are present.
func ValidateVariables(text string, vars map[string]interface{}) []string {
	required := ExtractVariables(text)
	var missing []string
	
	for _, varName := range required {
		if _, exists := vars[varName]; !exists {
			missing = append(missing, varName)
		}
	}
	
	return missing
}

// EvaluateExpression evaluates a simple expression for conditional processing.
// Supports basic comparison operations: ==, !=, >, <, >=, <=
func EvaluateExpression(expr string, vars map[string]interface{}) (bool, error) {
	// Check for equality comparison
	if strings.Contains(expr, "==") {
		parts := strings.SplitN(expr, "==", 2)
		if len(parts) != 2 {
			return false, fmt.Errorf("invalid equality expression: %s", expr)
		}
		
		leftVar := strings.TrimSpace(parts[0])
		rightValue := strings.TrimSpace(parts[1])
		
		// Get the left variable's value
		leftValue, exists := vars[leftVar]
		if !exists {
			return false, nil // If variable doesn't exist, condition is false
		}
		
		// Handle string literals in quotes
		if strings.HasPrefix(rightValue, "\"") && strings.HasSuffix(rightValue, "\"") {
			rightValue = strings.Trim(rightValue, "\"")
			return fmt.Sprintf("%v", leftValue) == rightValue, nil
		}
		
		// Check if right side is another variable
		if rightVar, exists := vars[rightValue]; exists {
			return fmt.Sprintf("%v", leftValue) == fmt.Sprintf("%v", rightVar), nil
		}
		
		// Direct comparison
		return fmt.Sprintf("%v", leftValue) == rightValue, nil
	}
	
	// Check for inequality comparison
	if strings.Contains(expr, "!=") {
		parts := strings.SplitN(expr, "!=", 2)
		if len(parts) != 2 {
			return false, fmt.Errorf("invalid inequality expression: %s", expr)
		}
		
		leftVar := strings.TrimSpace(parts[0])
		rightValue := strings.TrimSpace(parts[1])
		
		// Get the left variable's value
		leftValue, exists := vars[leftVar]
		if !exists {
			return true, nil // If variable doesn't exist, inequality is true
		}
		
		// Handle string literals in quotes
		if strings.HasPrefix(rightValue, "\"") && strings.HasSuffix(rightValue, "\"") {
			rightValue = strings.Trim(rightValue, "\"")
			return fmt.Sprintf("%v", leftValue) != rightValue, nil
		}
		
		// Check if right side is another variable
		if rightVar, exists := vars[rightValue]; exists {
			return fmt.Sprintf("%v", leftValue) != fmt.Sprintf("%v", rightVar), nil
		}
		
		// Direct comparison
		return fmt.Sprintf("%v", leftValue) != rightValue, nil
	}
	
	// Check for existence of a variable (truth check)
	varName := strings.TrimSpace(expr)
	if value, exists := vars[varName]; exists {
		// Check if the value is truthy (not empty, not zero, etc.)
		switch v := value.(type) {
		case bool:
			return v, nil
		case string:
			return v != "", nil
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
			return fmt.Sprintf("%v", v) != "0", nil
		default:
			return true, nil // Any other non-nil value is considered true
		}
	}
	
	return false, nil
}