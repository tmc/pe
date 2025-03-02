// Package assertutil provides utilities for testing assertions against model outputs.
package assertutil

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/xeipuuv/gojsonschema"
)

// Assert provides assertion functionality for validating LLM responses.
type Assert struct {
	// Input is the response text to be validated.
	Input string
	// Results collects all assertion results.
	Results []AssertionResult
}

// New creates a new assertion validator for the given input.
func New(input string) *Assert {
	return &Assert{
		Input: input,
	}
}

// Contains asserts that the input contains the specified substring.
func (a *Assert) Contains(substr string) bool {
	result := AssertionResult{
		Assertion: Assertion{
			Type:  Contains,
			Value: substr,
		},
		Actual:   a.Input,
		Expected: substr,
	}

	result.Success = strings.Contains(a.Input, substr)
	if !result.Success {
		result.Reason = fmt.Sprintf("input does not contain expected string: %s", substr)
	} else {
		result.Reason = "Assertion passed"
	}

	a.Results = append(a.Results, result)
	return result.Success
}

// NotContains asserts that the input does not contain the specified substring.
func (a *Assert) NotContains(substr string) bool {
	result := AssertionResult{
		Assertion: Assertion{
			Type:  NotContains,
			Value: substr,
		},
		Actual:   a.Input,
		Expected: substr,
	}

	result.Success = !strings.Contains(a.Input, substr)
	if !result.Success {
		result.Reason = fmt.Sprintf("input contains string that should not be present: %s", substr)
	} else {
		result.Reason = "Assertion passed"
	}

	a.Results = append(a.Results, result)
	return result.Success
}

// Equals asserts that the input exactly equals the expected string.
func (a *Assert) Equals(expected string) bool {
	result := AssertionResult{
		Assertion: Assertion{
			Type:  Equals,
			Value: expected,
		},
		Actual:   a.Input,
		Expected: expected,
	}

	result.Success = a.Input == expected
	if !result.Success {
		result.Reason = "input does not match expected string"
	} else {
		result.Reason = "Assertion passed"
	}

	a.Results = append(a.Results, result)
	return result.Success
}

// StartsWith asserts that the input starts with the expected string.
func (a *Assert) StartsWith(prefix string) bool {
	result := AssertionResult{
		Assertion: Assertion{
			Type:  StartsWith,
			Value: prefix,
		},
		Actual:   a.Input,
		Expected: prefix,
	}

	result.Success = strings.HasPrefix(a.Input, prefix)
	if !result.Success {
		result.Reason = fmt.Sprintf("input does not start with: %s", prefix)
	} else {
		result.Reason = "Assertion passed"
	}

	a.Results = append(a.Results, result)
	return result.Success
}

// EndsWith asserts that the input ends with the expected string.
func (a *Assert) EndsWith(suffix string) bool {
	result := AssertionResult{
		Assertion: Assertion{
			Type:  EndsWith,
			Value: suffix,
		},
		Actual:   a.Input,
		Expected: suffix,
	}

	result.Success = strings.HasSuffix(a.Input, suffix)
	if !result.Success {
		result.Reason = fmt.Sprintf("input does not end with: %s", suffix)
	} else {
		result.Reason = "Assertion passed"
	}

	a.Results = append(a.Results, result)
	return result.Success
}

// MatchesRegex asserts that the input matches the given regular expression pattern.
func (a *Assert) MatchesRegex(pattern string) bool {
	result := AssertionResult{
		Assertion: Assertion{
			Type:  Regex,
			Value: pattern,
		},
		Actual:   a.Input,
		Expected: pattern,
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		result.Success = false
		result.Reason = fmt.Sprintf("invalid regex pattern: %s", err)
		a.Results = append(a.Results, result)
		return false
	}

	result.Success = re.MatchString(a.Input)
	if !result.Success {
		result.Reason = fmt.Sprintf("input does not match regex pattern: %s", pattern)
	} else {
		result.Reason = "Assertion passed"
	}

	a.Results = append(a.Results, result)
	return result.Success
}

// LengthEquals asserts that the input length equals the expected length.
func (a *Assert) LengthEquals(length int) bool {
	result := AssertionResult{
		Assertion: Assertion{
			Type:  Length,
			Value: float64(length),
		},
		Actual:   a.Input,
		Expected: fmt.Sprintf("%d", length),
	}

	result.Success = len(a.Input) == length
	if !result.Success {
		result.Reason = fmt.Sprintf("input length %d does not equal %d", len(a.Input), length)
	} else {
		result.Reason = "Assertion passed"
	}

	a.Results = append(a.Results, result)
	return result.Success
}

// LengthGreaterThan asserts that the input length is greater than the specified length.
func (a *Assert) LengthGreaterThan(length int) bool {
	result := AssertionResult{
		Assertion: Assertion{
			Type:  Length,
			Value: fmt.Sprintf(">%d", length),
		},
		Actual:   a.Input,
		Expected: fmt.Sprintf(">%d", length),
	}

	result.Success = len(a.Input) > length
	if !result.Success {
		result.Reason = fmt.Sprintf("input length %d is not > %d", len(a.Input), length)
	} else {
		result.Reason = "Assertion passed"
	}

	a.Results = append(a.Results, result)
	return result.Success
}

// LengthLessThan asserts that the input length is less than the specified length.
func (a *Assert) LengthLessThan(length int) bool {
	result := AssertionResult{
		Assertion: Assertion{
			Type:  Length,
			Value: fmt.Sprintf("<%d", length),
		},
		Actual:   a.Input,
		Expected: fmt.Sprintf("<%d", length),
	}

	result.Success = len(a.Input) < length
	if !result.Success {
		result.Reason = fmt.Sprintf("input length %d is not < %d", len(a.Input), length)
	} else {
		result.Reason = "Assertion passed"
	}

	a.Results = append(a.Results, result)
	return result.Success
}

// LengthInRange asserts that the input length is between the specified min and max lengths (inclusive).
func (a *Assert) LengthInRange(min, max int) bool {
	result := AssertionResult{
		Assertion: Assertion{
			Type:  Length,
			Value: fmt.Sprintf(">=%d,<=%d", min, max),
		},
		Actual:   a.Input,
		Expected: fmt.Sprintf("between %d and %d", min, max),
	}

	actualLen := len(a.Input)
	result.Success = actualLen >= min && actualLen <= max
	if !result.Success {
		result.Reason = fmt.Sprintf("input length %d is not between %d and %d", actualLen, min, max)
	} else {
		result.Reason = "Assertion passed"
	}

	a.Results = append(a.Results, result)
	return result.Success
}

// IsValidJSON asserts that the input is valid JSON.
func (a *Assert) IsValidJSON() bool {
	result := AssertionResult{
		Assertion: Assertion{
			Type:  JSON,
			Value: "valid",
		},
		Actual:   a.Input,
		Expected: "valid JSON",
	}

	var js interface{}
	err := json.Unmarshal([]byte(a.Input), &js)
	result.Success = err == nil
	if !result.Success {
		result.Reason = fmt.Sprintf("input is not valid JSON: %s", err)
	} else {
		result.Reason = "Assertion passed"
	}

	a.Results = append(a.Results, result)
	return result.Success
}

// MatchesJSONSchema asserts that the input is valid JSON and matches the provided JSON schema.
func (a *Assert) MatchesJSONSchema(schema string) bool {
	result := AssertionResult{
		Assertion: Assertion{
			Type:  JSON,
			Value: schema,
		},
		Actual:   a.Input,
		Expected: "JSON matching schema",
	}

	// First check if the input is valid JSON
	var js interface{}
	err := json.Unmarshal([]byte(a.Input), &js)
	if err != nil {
		result.Success = false
		result.Reason = fmt.Sprintf("input is not valid JSON: %s", err)
		a.Results = append(a.Results, result)
		return false
	}

	// Validate against the schema
	schemaLoader := gojsonschema.NewStringLoader(schema)
	documentLoader := gojsonschema.NewStringLoader(a.Input)

	validationResult, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		result.Success = false
		result.Reason = fmt.Sprintf("schema validation error: %s", err)
		a.Results = append(a.Results, result)
		return false
	}

	result.Success = validationResult.Valid()
	if !result.Success {
		var errMsgs []string
		for _, err := range validationResult.Errors() {
			errMsgs = append(errMsgs, err.String())
		}
		result.Reason = fmt.Sprintf("JSON schema validation failed: %s", strings.Join(errMsgs, "; "))
	} else {
		result.Reason = "Assertion passed"
	}

	a.Results = append(a.Results, result)
	return result.Success
}

// Custom performs a custom assertion using the provided function.
func (a *Assert) Custom(name string, fn func(string) (bool, string)) bool {
	result := AssertionResult{
		Assertion: Assertion{
			Type:  AssertionType("custom:" + name),
			Value: name,
		},
		Actual:   a.Input,
		Expected: "custom assertion",
	}

	success, reason := fn(a.Input)
	result.Success = success
	if !success {
		if reason == "" {
			reason = "custom assertion failed"
		}
		result.Reason = reason
	} else {
		if reason == "" {
			reason = "Assertion passed"
		}
		result.Reason = reason
	}

	a.Results = append(a.Results, result)
	return result.Success
}

// AllPassed returns true if all assertions passed.
func (a *Assert) AllPassed() bool {
	for _, result := range a.Results {
		if !result.Success {
			return false
		}
	}
	return true
}

// GetResults returns all assertion results.
func (a *Assert) GetResults() []AssertionResult {
	return a.Results
}

// GetFailures returns only failed assertion results.
func (a *Assert) GetFailures() []AssertionResult {
	var failures []AssertionResult
	for _, result := range a.Results {
		if !result.Success {
			failures = append(failures, result)
		}
	}
	return failures
}