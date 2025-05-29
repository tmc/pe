// Package starlark provides Starlark-based evaluation capabilities for PE.
// This is an experimental extension that allows using Starlark scripts
// for dynamic test generation, custom rubrics, and complex evaluation logic.
package starlark

import (
	"fmt"
	"regexp"
	"strings"

	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

// Evaluator provides Starlark-based evaluation capabilities
type Evaluator struct {
	globals starlark.StringDict
}

// NewEvaluator creates a new Starlark evaluator with PE built-ins
func NewEvaluator() *Evaluator {
	return &Evaluator{
		globals: makeBuiltins(),
	}
}

// EvalFile executes a Starlark file and returns the result
func (e *Evaluator) EvalFile(filename string, src []byte) (starlark.StringDict, error) {
	thread := &starlark.Thread{Name: "pe-starlark"}
	
	// Execute the Starlark code
	globals, err := starlark.ExecFile(thread, filename, src, e.globals)
	if err != nil {
		return nil, fmt.Errorf("starlark execution error: %w", err)
	}
	
	return globals, nil
}

// EvaluateTest runs a test function from Starlark and returns the result
func (e *Evaluator) EvaluateTest(globals starlark.StringDict, testName string, response string) (*TestResult, error) {
	testFunc, ok := globals[testName]
	if !ok {
		return nil, fmt.Errorf("test function %q not found", testName)
	}
	
	thread := &starlark.Thread{Name: "pe-test"}
	
	// Call the test function with the response
	result, err := starlark.Call(thread, testFunc, starlark.Tuple{starlark.String(response)}, nil)
	if err != nil {
		return nil, fmt.Errorf("test execution error: %w", err)
	}
	
	// Convert result to TestResult
	return convertToTestResult(result)
}

// TestResult represents the result of a Starlark test evaluation
type TestResult struct {
	Pass    bool              `json:"pass"`
	Score   float64           `json:"score,omitempty"`
	Reason  string            `json:"reason,omitempty"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// convertToTestResult converts a Starlark value to TestResult
func convertToTestResult(val starlark.Value) (*TestResult, error) {
	switch v := val.(type) {
	case starlark.Bool:
		return &TestResult{Pass: bool(v)}, nil
	case *starlark.Dict:
		result := &TestResult{Details: make(map[string]interface{})}
		
		// Extract standard fields
		if pass, found, _ := v.Get(starlark.String("pass")); found {
			if passVal, ok := pass.(starlark.Bool); ok {
				result.Pass = bool(passVal)
			}
		}
		
		if score, found, _ := v.Get(starlark.String("score")); found {
			if scoreVal, ok := score.(starlark.Float); ok {
				result.Score = float64(scoreVal)
			}
		}
		
		if reason, found, _ := v.Get(starlark.String("reason")); found {
			if reasonVal, ok := reason.(starlark.String); ok {
				result.Reason = string(reasonVal)
			}
		}
		
		// Add all other fields to details
		for _, item := range v.Items() {
			key := item[0].(starlark.String)
			keyStr := string(key)
			if keyStr != "pass" && keyStr != "score" && keyStr != "reason" {
				result.Details[keyStr] = convertStarlarkValue(item[1])
			}
		}
		
		return result, nil
	default:
		return nil, fmt.Errorf("unsupported result type: %T", val)
	}
}

// convertStarlarkValue converts a Starlark value to Go interface{}
func convertStarlarkValue(val starlark.Value) interface{} {
	switch v := val.(type) {
	case starlark.String:
		return string(v)
	case starlark.Int:
		if i, ok := v.Int64(); ok {
			return i
		}
		return v.String()
	case starlark.Float:
		return float64(v)
	case starlark.Bool:
		return bool(v)
	case *starlark.List:
		result := make([]interface{}, v.Len())
		for i := 0; i < v.Len(); i++ {
			result[i] = convertStarlarkValue(v.Index(i))
		}
		return result
	case *starlark.Dict:
		result := make(map[string]interface{})
		for _, item := range v.Items() {
			key := string(item[0].(starlark.String))
			result[key] = convertStarlarkValue(item[1])
		}
		return result
	default:
		return v.String()
	}
}

// makeBuiltins creates the PE-specific built-in functions for Starlark
func makeBuiltins() starlark.StringDict {
	return starlark.StringDict{
		"contains":    starlark.NewBuiltin("contains", builtinContains),
		"min_length":  starlark.NewBuiltin("min_length", builtinMinLength),
		"max_length":  starlark.NewBuiltin("max_length", builtinMaxLength),
		"equals":      starlark.NewBuiltin("equals", builtinEquals),
		"regex_match": starlark.NewBuiltin("regex_match", builtinRegexMatch),
		"word_count":  starlark.NewBuiltin("word_count", builtinWordCount),
		"struct":      starlark.NewBuiltin("struct", starlarkstruct.Make),
	}
}

// Built-in function implementations
func builtinContains(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var text, substr starlark.String
	if err := starlark.UnpackArgs("contains", args, kwargs, "text", &text, "substr", &substr); err != nil {
		return nil, err
	}
	
	result := strings.Contains(string(text), string(substr))
	return starlark.Bool(result), nil
}

func builtinMinLength(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var text starlark.String
	var minLen starlark.Int
	if err := starlark.UnpackArgs("min_length", args, kwargs, "text", &text, "min", &minLen); err != nil {
		return nil, err
	}
	
	minLenInt, _ := minLen.Int64()
	result := len(string(text)) >= int(minLenInt)
	return starlark.Bool(result), nil
}

func builtinMaxLength(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var text starlark.String
	var maxLen starlark.Int
	if err := starlark.UnpackArgs("max_length", args, kwargs, "text", &text, "max", &maxLen); err != nil {
		return nil, err
	}
	
	maxLenInt, _ := maxLen.Int64()
	result := len(string(text)) <= int(maxLenInt)
	return starlark.Bool(result), nil
}

func builtinEquals(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var a, b starlark.String
	if err := starlark.UnpackArgs("equals", args, kwargs, "a", &a, "b", &b); err != nil {
		return nil, err
	}
	
	result := string(a) == string(b)
	return starlark.Bool(result), nil
}

func builtinRegexMatch(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var text, pattern starlark.String
	if err := starlark.UnpackArgs("regex_match", args, kwargs, "text", &text, "pattern", &pattern); err != nil {
		return nil, err
	}
	
	// Compile and match the regex
	regex, err := regexp.Compile(string(pattern))
	if err != nil {
		return nil, fmt.Errorf("invalid regex pattern: %w", err)
	}
	
	result := regex.MatchString(string(text))
	return starlark.Bool(result), nil
}

func builtinWordCount(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var text starlark.String
	if err := starlark.UnpackArgs("word_count", args, kwargs, "text", &text); err != nil {
		return nil, err
	}
	
	words := strings.Fields(string(text))
	return starlark.MakeInt(len(words)), nil
}