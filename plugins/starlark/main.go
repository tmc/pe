package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/tmc/pe/internal/promptfoo/evaluation/evaluator"
	"github.com/tmc/pe/internal/promptfoo"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

// StarlarkEvaluator provides a Starlark-based evaluation engine for PE
type StarlarkEvaluator struct {
	configPath string
	config     *StarlarkConfig
	globals    starlark.StringDict
}

// StarlarkConfig represents configuration loaded from a Starlark file
type StarlarkConfig struct {
	Provider  string
	Model     string
	Prompts   []string
	Tests     []StarlarkTest
	Variables map[string]interface{}
}

// StarlarkTest represents a test case defined in Starlark
type StarlarkTest struct {
	Name        string
	Input       string
	Variables   map[string]interface{}
	Assertions  []StarlarkAssertion
	Description string
}

// StarlarkAssertion represents an assertion in a test
type StarlarkAssertion struct {
	Type    string
	Value   interface{}
	Message string
	Weight  float64
}

// NewStarlarkEvaluator creates a new Starlark evaluator instance
func NewStarlarkEvaluator(configPath string) (*StarlarkEvaluator, error) {
	return &StarlarkEvaluator{
		configPath: configPath,
		globals:    make(starlark.StringDict),
	}, nil
}

// Load loads and parses the Starlark configuration file
func (e *StarlarkEvaluator) Load() error {
	// Setup built-in functions
	e.setupBuiltins()

	// Load and execute the Starlark file
	thread := &starlark.Thread{
		Name: "starlark-evaluator",
		Print: func(_ *starlark.Thread, msg string) {
			fmt.Println(msg)
		},
	}

	globals, err := starlark.ExecFile(thread, e.configPath, nil, e.globals)
	if err != nil {
		return fmt.Errorf("failed to execute Starlark file: %w", err)
	}

	// Extract configuration from globals
	config, err := e.extractConfig(globals)
	if err != nil {
		return fmt.Errorf("failed to extract configuration: %w", err)
	}

	e.config = config
	return nil
}

// setupBuiltins adds built-in functions to the Starlark environment
func (e *StarlarkEvaluator) setupBuiltins() {
	// Assertion builder functions
	e.globals["equals"] = starlark.NewBuiltin("equals", func(_ *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		var value starlark.Value
		var message string
		if err := starlark.UnpackArgs("equals", args, kwargs, "value", &value, "message?", &message); err != nil {
			return nil, err
		}
		return makeAssertion("equals", value, message, 1.0), nil
	})

	e.globals["contains"] = starlark.NewBuiltin("contains", func(_ *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		var value starlark.Value
		var message string
		if err := starlark.UnpackArgs("contains", args, kwargs, "value", &value, "message?", &message); err != nil {
			return nil, err
		}
		return makeAssertion("contains", value, message, 1.0), nil
	})

	e.globals["regex"] = starlark.NewBuiltin("regex", func(_ *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		var pattern starlark.String
		var message string
		if err := starlark.UnpackArgs("regex", args, kwargs, "pattern", &pattern, "message?", &message); err != nil {
			return nil, err
		}
		return makeAssertion("regex", pattern, message, 1.0), nil
	})

	e.globals["icontains"] = starlark.NewBuiltin("icontains", func(_ *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		var value starlark.Value
		var message string
		if err := starlark.UnpackArgs("icontains", args, kwargs, "value", &value, "message?", &message); err != nil {
			return nil, err
		}
		return makeAssertion("icontains", value, message, 1.0), nil
	})

	e.globals["starts_with"] = starlark.NewBuiltin("starts_with", func(_ *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		var value starlark.Value
		var message string
		if err := starlark.UnpackArgs("starts_with", args, kwargs, "value", &value, "message?", &message); err != nil {
			return nil, err
		}
		return makeAssertion("starts-with", value, message, 1.0), nil
	})

	e.globals["ends_with"] = starlark.NewBuiltin("ends_with", func(_ *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		var value starlark.Value
		var message string
		if err := starlark.UnpackArgs("ends_with", args, kwargs, "value", &value, "message?", &message); err != nil {
			return nil, err
		}
		return makeAssertion("ends-with", value, message, 1.0), nil
	})

	e.globals["not_contains"] = starlark.NewBuiltin("not_contains", func(_ *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		var value starlark.Value
		var message string
		if err := starlark.UnpackArgs("not_contains", args, kwargs, "value", &value, "message?", &message); err != nil {
			return nil, err
		}
		return makeAssertion("not-contains", value, message, 1.0), nil
	})

	// Test definition function
	e.globals["test"] = starlark.NewBuiltin("test", func(_ *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		var name, input, description starlark.String
		var variables *starlark.Dict
		var assertions *starlark.List

		if err := starlark.UnpackArgs("test", args, kwargs,
			"name", &name,
			"input", &input,
			"assertions", &assertions,
			"variables?", &variables,
			"description?", &description); err != nil {
			return nil, err
		}

		return starlarkstruct.FromStringDict(starlark.String("test"), starlark.StringDict{
			"name":        name,
			"input":       input,
			"assertions":  assertions,
			"variables":   variables,
			"description": description,
		}), nil
	})

	// Configuration function
	e.globals["config"] = starlark.NewBuiltin("config", func(_ *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		var provider, model starlark.String
		var prompts *starlark.List
		var variables *starlark.Dict

		if err := starlark.UnpackArgs("config", args, kwargs,
			"provider", &provider,
			"model", &model,
			"prompts", &prompts,
			"variables?", &variables); err != nil {
			return nil, err
		}

		return starlarkstruct.FromStringDict(starlark.String("config"), starlark.StringDict{
			"provider":  provider,
			"model":     model,
			"prompts":   prompts,
			"variables": variables,
		}), nil
	})

	// Load function for including other Starlark files
	e.globals["load_tests"] = starlark.NewBuiltin("load_tests", func(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		var path starlark.String
		if err := starlark.UnpackArgs("load_tests", args, kwargs, "path", &path); err != nil {
			return nil, err
		}

		// Resolve path relative to current file
		basePath := filepath.Dir(e.configPath)
		fullPath := filepath.Join(basePath, string(path))

		// Execute the loaded file
		globals, err := starlark.ExecFile(thread, fullPath, nil, e.globals)
		if err != nil {
			return nil, fmt.Errorf("failed to load tests from %s: %w", fullPath, err)
		}

		// Extract tests from the loaded file
		if tests, ok := globals["tests"]; ok {
			return tests, nil
		}

		return starlark.None, nil
	})
}

// makeAssertion creates a Starlark assertion struct
func makeAssertion(assertType string, value starlark.Value, message string, weight float64) starlark.Value {
	return starlarkstruct.FromStringDict(starlark.String("assertion"), starlark.StringDict{
		"type":    starlark.String(assertType),
		"value":   value,
		"message": starlark.String(message),
		"weight":  starlark.Float(weight),
	})
}

// extractConfig extracts configuration from Starlark globals
func (e *StarlarkEvaluator) extractConfig(globals starlark.StringDict) (*StarlarkConfig, error) {
	config := &StarlarkConfig{
		Variables: make(map[string]interface{}),
	}

	// Extract configuration
	if cfg, ok := globals["configuration"]; ok {
		if cfgStruct, ok := cfg.(*starlarkstruct.Struct); ok {
			// Extract provider
			if provider, err := cfgStruct.Attr("provider"); err == nil {
				config.Provider = string(provider.(starlark.String))
			}

			// Extract model
			if model, err := cfgStruct.Attr("model"); err == nil {
				config.Model = string(model.(starlark.String))
			}

			// Extract prompts
			if prompts, err := cfgStruct.Attr("prompts"); err == nil {
				if promptList, ok := prompts.(*starlark.List); ok {
					for i := 0; i < promptList.Len(); i++ {
						prompt := promptList.Index(i)
						config.Prompts = append(config.Prompts, string(prompt.(starlark.String)))
					}
				}
			}

			// Extract variables
			if vars, err := cfgStruct.Attr("variables"); err == nil {
				if varDict, ok := vars.(*starlark.Dict); ok {
					config.Variables = starlarkDictToMap(varDict)
				}
			}
		}
	}

	// Extract tests
	if tests, ok := globals["tests"]; ok {
		if testList, ok := tests.(*starlark.List); ok {
			for i := 0; i < testList.Len(); i++ {
				test := testList.Index(i)
				if testStruct, ok := test.(*starlarkstruct.Struct); ok {
					starlarkTest, err := e.parseTest(testStruct)
					if err != nil {
						return nil, fmt.Errorf("failed to parse test %d: %w", i, err)
					}
					config.Tests = append(config.Tests, starlarkTest)
				}
			}
		}
	}

	return config, nil
}

// parseTest parses a Starlark test struct
func (e *StarlarkEvaluator) parseTest(test *starlarkstruct.Struct) (StarlarkTest, error) {
	result := StarlarkTest{
		Variables: make(map[string]interface{}),
	}

	// Extract name
	if name, err := test.Attr("name"); err == nil {
		result.Name = string(name.(starlark.String))
	}

	// Extract input
	if input, err := test.Attr("input"); err == nil {
		result.Input = string(input.(starlark.String))
	}

	// Extract description
	if desc, err := test.Attr("description"); err == nil && desc != starlark.None {
		result.Description = string(desc.(starlark.String))
	}

	// Extract variables
	if vars, err := test.Attr("variables"); err == nil && vars != starlark.None {
		if varDict, ok := vars.(*starlark.Dict); ok {
			result.Variables = starlarkDictToMap(varDict)
		}
	}

	// Extract assertions
	if assertions, err := test.Attr("assertions"); err == nil {
		if assertList, ok := assertions.(*starlark.List); ok {
			for i := 0; i < assertList.Len(); i++ {
				assertion := assertList.Index(i)
				if assertStruct, ok := assertion.(*starlarkstruct.Struct); ok {
					starlarkAssertion, err := e.parseAssertion(assertStruct)
					if err != nil {
						return result, fmt.Errorf("failed to parse assertion %d: %w", i, err)
					}
					result.Assertions = append(result.Assertions, starlarkAssertion)
				}
			}
		}
	}

	return result, nil
}

// parseAssertion parses a Starlark assertion struct
func (e *StarlarkEvaluator) parseAssertion(assertion *starlarkstruct.Struct) (StarlarkAssertion, error) {
	result := StarlarkAssertion{
		Weight: 1.0, // default weight
	}

	// Extract type
	if assertType, err := assertion.Attr("type"); err == nil {
		result.Type = string(assertType.(starlark.String))
	}

	// Extract value
	if value, err := assertion.Attr("value"); err == nil {
		result.Value = starlarkValueToGo(value)
	}

	// Extract message
	if message, err := assertion.Attr("message"); err == nil && message != starlark.None {
		result.Message = string(message.(starlark.String))
	}

	// Extract weight
	if weight, err := assertion.Attr("weight"); err == nil && weight != starlark.None {
		if w, ok := weight.(starlark.Float); ok {
			result.Weight = float64(w)
		}
	}

	return result, nil
}

// starlarkValueToGo converts a Starlark value to a Go value
func starlarkValueToGo(v starlark.Value) interface{} {
	switch v := v.(type) {
	case starlark.String:
		return string(v)
	case starlark.Int:
		i, _ := v.Int64()
		return i
	case starlark.Float:
		return float64(v)
	case starlark.Bool:
		return bool(v)
	case *starlark.List:
		result := []interface{}{}
		for i := 0; i < v.Len(); i++ {
			result = append(result, starlarkValueToGo(v.Index(i)))
		}
		return result
	case *starlark.Dict:
		return starlarkDictToMap(v)
	default:
		return v.String()
	}
}

// starlarkDictToMap converts a Starlark dict to a Go map
func starlarkDictToMap(d *starlark.Dict) map[string]interface{} {
	result := make(map[string]interface{})
	for _, item := range d.Items() {
		key := string(item[0].(starlark.String))
		result[key] = starlarkValueToGo(item[1])
	}
	return result
}

// Run executes the evaluation based on the loaded configuration
func (e *StarlarkEvaluator) Run(ctx context.Context) (*promptfoo.EvaluationResult, error) {
	if e.config == nil {
		return nil, fmt.Errorf("configuration not loaded")
	}

	// Convert to promptfoo config
	pfConfig := e.toPromptfooConfig()

	// Run evaluation using the standard evaluator
	result, err := evaluator.Evaluate(pfConfig, 30*time.Second, false, 5, true)
	if err != nil {
		return nil, fmt.Errorf("evaluation failed: %w", err)
	}

	return &result, nil
}

// toPromptfooConfig converts Starlark config to promptfoo config
func (e *StarlarkEvaluator) toPromptfooConfig() promptfoo.Config {
	config := promptfoo.Config{
		Providers: []string{fmt.Sprintf("%s:%s", e.config.Provider, e.config.Model)},
		Prompts:   e.config.Prompts,
	}

	// Convert tests
	for _, test := range e.config.Tests {
		pfTest := promptfoo.TestCase{
			Description: test.Description,
			Vars:        mergeVars(e.config.Variables, test.Variables),
		}

		// Set input as a variable if provided
		if test.Input != "" {
			if pfTest.Vars == nil {
				pfTest.Vars = make(map[string]interface{})
			}
			pfTest.Vars["input"] = test.Input
		}

		// Convert assertions
		for _, assertion := range test.Assertions {
			pfAssertion := promptfoo.Assertion{
				Type:  assertion.Type,
				Value: assertion.Value,
			}
			pfTest.Assert = append(pfTest.Assert, pfAssertion)
		}

		config.Tests = append(config.Tests, pfTest)
	}

	return config
}

// mergeVars merges global and test-specific variables
func mergeVars(global, local map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	// Copy global variables
	for k, v := range global {
		result[k] = v
	}

	// Override with local variables
	for k, v := range local {
		result[k] = v
	}

	return result
}

// OutputJSON outputs the evaluation results as JSON
func (e *StarlarkEvaluator) OutputJSON(result *promptfoo.EvaluationResult, outputPath string) error {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal results: %w", err)
	}

	if outputPath == "-" {
		fmt.Println(string(data))
	} else {
		if err := os.WriteFile(outputPath, data, 0644); err != nil {
			return fmt.Errorf("failed to write output file: %w", err)
		}
	}

	return nil
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <config.star> [output.json]\n", os.Args[0])
		os.Exit(1)
	}

	configPath := os.Args[1]
	outputPath := "-"
	if len(os.Args) > 2 {
		outputPath = os.Args[2]
	}

	// Create evaluator
	evaluator, err := NewStarlarkEvaluator(configPath)
	if err != nil {
		log.Fatalf("Failed to create evaluator: %v", err)
	}

	// Load configuration
	if err := evaluator.Load(); err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Run evaluation
	ctx := context.Background()
	result, err := evaluator.Run(ctx)
	if err != nil {
		log.Fatalf("Failed to run evaluation: %v", err)
	}

	// Output results
	if err := evaluator.OutputJSON(result, outputPath); err != nil {
		log.Fatalf("Failed to output results: %v", err)
	}

	// Print summary
	stats := result.Results.Stats
	fmt.Fprintf(os.Stderr, "\n✅ Evaluation complete: %d passed, %d failed\n",
		stats.Successes, stats.Failures)
}
