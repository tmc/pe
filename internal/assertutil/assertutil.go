// Package assertutil provides utilities for testing assertions against model outputs.
package assertutil

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

// AssertionType represents a type of assertion that can be performed.
type AssertionType string

const (
	// Contains asserts that output contains a specific string.
	Contains AssertionType = "contains"
	
	// NotContains asserts that output does not contain a specific string.
	NotContains AssertionType = "not-contains"
	
	// Equals asserts that output exactly equals a specific string.
	Equals AssertionType = "equals"
	
	// Regex asserts that output matches a regular expression.
	Regex AssertionType = "regex"
	
	// StartsWith asserts that output starts with a specific string.
	StartsWith AssertionType = "starts-with"
	
	// EndsWith asserts that output ends with a specific string.
	EndsWith AssertionType = "ends-with"
	
	// JSON asserts that output contains valid JSON that matches a JSON path.
	JSON AssertionType = "json"
	
	// Length asserts that output is of a specific length.
	Length AssertionType = "length"
)

// Assertion represents a single assertion to validate against model output.
type Assertion struct {
	Type  AssertionType `json:"type"`
	Value interface{}   `json:"value"`
	Path  string        `json:"path,omitempty"` // For JSON assertions
}

// AssertionResult represents the result of checking an assertion.
type AssertionResult struct {
	Assertion Assertion  `json:"assertion"`
	Success   bool       `json:"success"`
	Reason    string     `json:"reason,omitempty"`
	Expected  string     `json:"expected,omitempty"`
	Actual    string     `json:"actual,omitempty"`
}

// Check performs the assertion against the given output and returns a result.
func (a Assertion) Check(output string) AssertionResult {
	result := AssertionResult{
		Assertion: a,
		Actual:    output,
	}
	
	// Convert value to string if it's not already
	expectedStr, ok := toString(a.Value)
	if !ok && a.Type != Length { // Length can use numeric values
		result.Success = false
		result.Reason = fmt.Sprintf("assertion value must be a string, got %T", a.Value)
		return result
	}
	
	result.Expected = expectedStr
	
	// Check assertion based on type
	switch a.Type {
	case Contains:
		result.Success = strings.Contains(output, expectedStr)
		if !result.Success {
			result.Reason = fmt.Sprintf("output does not contain expected string: %s", expectedStr)
		}
		
	case NotContains:
		result.Success = !strings.Contains(output, expectedStr)
		if !result.Success {
			result.Reason = fmt.Sprintf("output contains string that should not be present: %s", expectedStr)
		}
		
	case Equals:
		result.Success = output == expectedStr
		if !result.Success {
			result.Reason = "output does not match expected string"
		}
		
	case Regex:
		re, err := regexp.Compile(expectedStr)
		if err != nil {
			result.Success = false
			result.Reason = fmt.Sprintf("invalid regex pattern: %s", err)
			return result
		}
		
		result.Success = re.MatchString(output)
		if !result.Success {
			result.Reason = fmt.Sprintf("output does not match regex pattern: %s", expectedStr)
		}
		
	case StartsWith:
		result.Success = strings.HasPrefix(output, expectedStr)
		if !result.Success {
			result.Reason = fmt.Sprintf("output does not start with: %s", expectedStr)
		}
		
	case EndsWith:
		result.Success = strings.HasSuffix(output, expectedStr)
		if !result.Success {
			result.Reason = fmt.Sprintf("output does not end with: %s", expectedStr)
		}
		
	case JSON:
		// Validate that output is valid JSON
		var outputJSON interface{}
		if err := json.Unmarshal([]byte(output), &outputJSON); err != nil {
			result.Success = false
			result.Reason = fmt.Sprintf("output is not valid JSON: %s", err)
			return result
		}
		
		// TODO: Implement JSON path checking
		result.Success = true
		result.Reason = "JSON assertion not fully implemented yet"
		
	case Length:
		// Length can be a number or comparison
		expectedLen, ok := a.Value.(float64)
		if !ok {
			// Try string comparison
			lenStr, ok := a.Value.(string)
			if !ok {
				result.Success = false
				result.Reason = fmt.Sprintf("length value must be a number or comparison string, got %T", a.Value)
				return result
			}
			
			result.Success, result.Reason = checkLengthComparison(lenStr, len(output))
		} else {
			// Exact length match
			result.Success = len(output) == int(expectedLen)
			if !result.Success {
				result.Reason = fmt.Sprintf("output length %d does not match expected length %d", len(output), int(expectedLen))
			}
		}
		
	default:
		result.Success = false
		result.Reason = fmt.Sprintf("unknown assertion type: %s", a.Type)
	}
	
	// Set a success reason if not already set
	if result.Success && result.Reason == "" {
		result.Reason = "Assertion passed"
	}
	
	return result
}

// CheckAll checks all assertions against the given output and returns the results.
func CheckAll(output string, assertions []Assertion) []AssertionResult {
	var results []AssertionResult
	
	for _, assertion := range assertions {
		results = append(results, assertion.Check(output))
	}
	
	return results
}

// AllPassed returns true if all assertion results are successful.
func AllPassed(results []AssertionResult) bool {
	for _, result := range results {
		if !result.Success {
			return false
		}
	}
	
	return true
}

// ParseFromMap parses an assertion from a generic map.
func ParseFromMap(m map[string]interface{}) (Assertion, error) {
	var assertion Assertion
	
	// Get assertion type
	typeStr, ok := m["type"].(string)
	if !ok {
		return assertion, fmt.Errorf("assertion missing 'type' field")
	}
	
	assertion.Type = AssertionType(typeStr)
	
	// Get value
	value, ok := m["value"]
	if !ok {
		return assertion, fmt.Errorf("assertion missing 'value' field")
	}
	
	assertion.Value = value
	
	// Get path for JSON assertions
	if path, ok := m["path"].(string); ok {
		assertion.Path = path
	}
	
	return assertion, nil
}

// ParseAll parses multiple assertions from a list of generic maps.
func ParseAll(assertions []interface{}) ([]Assertion, error) {
	var result []Assertion
	
	for i, a := range assertions {
		assertMap, ok := a.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("assertion %d is not a map", i)
		}
		
		assertion, err := ParseFromMap(assertMap)
		if err != nil {
			return nil, fmt.Errorf("invalid assertion %d: %w", i, err)
		}
		
		result = append(result, assertion)
	}
	
	return result, nil
}

// Helper functions

// toString attempts to convert a value to a string.
func toString(v interface{}) (string, bool) {
	switch val := v.(type) {
	case string:
		return val, true
	case []byte:
		return string(val), true
	case fmt.Stringer:
		return val.String(), true
	default:
		return fmt.Sprintf("%v", v), true
	}
}

// checkLengthComparison parses and checks a length comparison string (e.g., ">10", "<=20").
func checkLengthComparison(comparison string, actualLen int) (bool, string) {
	comparison = strings.TrimSpace(comparison)
	
	// Parse the comparison operator and expected length
	var op string
	var expectedLen int
	
	if strings.HasPrefix(comparison, ">=") {
		op = ">="
		fmt.Sscanf(comparison[2:], "%d", &expectedLen)
	} else if strings.HasPrefix(comparison, "<=") {
		op = "<="
		fmt.Sscanf(comparison[2:], "%d", &expectedLen)
	} else if strings.HasPrefix(comparison, ">") {
		op = ">"
		fmt.Sscanf(comparison[1:], "%d", &expectedLen)
	} else if strings.HasPrefix(comparison, "<") {
		op = "<"
		fmt.Sscanf(comparison[1:], "%d", &expectedLen)
	} else if strings.HasPrefix(comparison, "==") {
		op = "=="
		fmt.Sscanf(comparison[2:], "%d", &expectedLen)
	} else if strings.HasPrefix(comparison, "=") {
		op = "="
		fmt.Sscanf(comparison[1:], "%d", &expectedLen)
	} else {
		// Try to parse as a direct number
		if _, err := fmt.Sscanf(comparison, "%d", &expectedLen); err == nil {
			op = "="
		} else {
			return false, fmt.Sprintf("invalid length comparison: %s", comparison)
		}
	}
	
	// Check the comparison
	switch op {
	case ">=":
		if actualLen >= expectedLen {
			return true, ""
		}
		return false, fmt.Sprintf("output length %d is not >= %d", actualLen, expectedLen)
		
	case "<=":
		if actualLen <= expectedLen {
			return true, ""
		}
		return false, fmt.Sprintf("output length %d is not <= %d", actualLen, expectedLen)
		
	case ">":
		if actualLen > expectedLen {
			return true, ""
		}
		return false, fmt.Sprintf("output length %d is not > %d", actualLen, expectedLen)
		
	case "<":
		if actualLen < expectedLen {
			return true, ""
		}
		return false, fmt.Sprintf("output length %d is not < %d", actualLen, expectedLen)
		
	case "==", "=":
		if actualLen == expectedLen {
			return true, ""
		}
		return false, fmt.Sprintf("output length %d does not equal %d", actualLen, expectedLen)
		
	default:
		return false, fmt.Sprintf("unknown comparison operator: %s", op)
	}
}