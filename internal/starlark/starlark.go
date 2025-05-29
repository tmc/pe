package starlark

import (
	"fmt"
	"regexp"
	"strings"

	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

// Interpreter wraps a Starlark interpreter for PE assertions
type Interpreter struct {
	thread  *starlark.Thread
	globals starlark.StringDict
}

// NewInterpreter creates a new Starlark interpreter with PE built-ins
func NewInterpreter() *Interpreter {
	thread := &starlark.Thread{
		Name: "pe-starlark",
		Print: func(_ *starlark.Thread, msg string) {
			fmt.Println(msg)
		},
	}

	// Initialize globals with PE built-ins
	globals := starlark.StringDict{
		"struct": starlark.NewBuiltin("struct", starlarkstruct.Make),
	}

	// Add PE-specific built-in functions
	addBuiltins(globals)

	return &Interpreter{
		thread:  thread,
		globals: globals,
	}
}

// addBuiltins adds PE-specific built-in functions to the globals
func addBuiltins(globals starlark.StringDict) {
	// String assertion functions
	globals["contains"] = starlark.NewBuiltin("contains", contains)
	globals["equals"] = starlark.NewBuiltin("equals", equals)
	globals["min_length"] = starlark.NewBuiltin("min_length", minLength)
	globals["max_length"] = starlark.NewBuiltin("max_length", maxLength)
	globals["regex_match"] = starlark.NewBuiltin("regex_match", regexMatch)
	
	// Numeric assertion functions
	globals["greater_than"] = starlark.NewBuiltin("greater_than", greaterThan)
	globals["less_than"] = starlark.NewBuiltin("less_than", lessThan)
	globals["between"] = starlark.NewBuiltin("between", between)
	
	// List/collection functions
	globals["has_length"] = starlark.NewBuiltin("has_length", hasLength)
	globals["is_subset"] = starlark.NewBuiltin("is_subset", isSubset)
	
	// Type checking functions
	globals["is_string"] = starlark.NewBuiltin("is_string", isString)
	globals["is_int"] = starlark.NewBuiltin("is_int", isInt)
	globals["is_float"] = starlark.NewBuiltin("is_float", isFloat)
	globals["is_list"] = starlark.NewBuiltin("is_list", isList)
	globals["is_dict"] = starlark.NewBuiltin("is_dict", isDict)
}

// Eval evaluates a Starlark expression with the given variables
func (i *Interpreter) Eval(expr string, vars map[string]interface{}) (bool, error) {
	// Convert Go variables to Starlark values
	starlarkVars := make(starlark.StringDict)
	for k, v := range vars {
		sv, err := goToStarlark(v)
		if err != nil {
			return false, fmt.Errorf("failed to convert variable %s: %w", k, err)
		}
		starlarkVars[k] = sv
	}

	// Merge variables with globals
	env := make(starlark.StringDict)
	for k, v := range i.globals {
		env[k] = v
	}
	for k, v := range starlarkVars {
		env[k] = v
	}

	// Evaluate the expression
	result, err := starlark.Eval(i.thread, "<assertion>", expr, env)
	if err != nil {
		return false, fmt.Errorf("evaluation error: %w", err)
	}

	// Convert result to boolean
	switch v := result.(type) {
	case starlark.Bool:
		return bool(v), nil
	case starlark.NoneType:
		return false, nil
	default:
		// Truthy evaluation: non-zero numbers, non-empty strings/collections are true
		return result.Truth(), nil
	}
}

// goToStarlark converts a Go value to a Starlark value
func goToStarlark(v interface{}) (starlark.Value, error) {
	switch val := v.(type) {
	case nil:
		return starlark.None, nil
	case bool:
		return starlark.Bool(val), nil
	case int:
		return starlark.MakeInt(val), nil
	case int64:
		return starlark.MakeInt64(val), nil
	case float64:
		return starlark.Float(val), nil
	case string:
		return starlark.String(val), nil
	case []interface{}:
		list := make([]starlark.Value, len(val))
		for i, item := range val {
			sv, err := goToStarlark(item)
			if err != nil {
				return nil, err
			}
			list[i] = sv
		}
		return starlark.NewList(list), nil
	case map[string]interface{}:
		dict := starlark.NewDict(len(val))
		for k, v := range val {
			kv, err := goToStarlark(k)
			if err != nil {
				return nil, err
			}
			vv, err := goToStarlark(v)
			if err != nil {
				return nil, err
			}
			dict.SetKey(kv, vv)
		}
		return dict, nil
	default:
		return nil, fmt.Errorf("unsupported type: %T", v)
	}
}

// Built-in function implementations

func contains(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var haystack, needle starlark.String
	if err := starlark.UnpackArgs(fn.Name(), args, kwargs, "haystack", &haystack, "needle", &needle); err != nil {
		return nil, err
	}
	return starlark.Bool(strings.Contains(string(haystack), string(needle))), nil
}

func equals(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("%s: want 2 arguments, got %d", fn.Name(), len(args))
	}
	return starlark.Bool(starlark.Equal(args[0], args[1])), nil
}

func minLength(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var str starlark.String
	var minLen starlark.Int
	if err := starlark.UnpackArgs(fn.Name(), args, kwargs, "str", &str, "min_len", &minLen); err != nil {
		return nil, err
	}
	
	min, ok := minLen.Int64()
	if !ok {
		return nil, fmt.Errorf("%s: min_len too large", fn.Name())
	}
	
	return starlark.Bool(len(string(str)) >= int(min)), nil
}

func maxLength(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var str starlark.String
	var maxLen starlark.Int
	if err := starlark.UnpackArgs(fn.Name(), args, kwargs, "str", &str, "max_len", &maxLen); err != nil {
		return nil, err
	}
	
	max, ok := maxLen.Int64()
	if !ok {
		return nil, fmt.Errorf("%s: max_len too large", fn.Name())
	}
	
	return starlark.Bool(len(string(str)) <= int(max)), nil
}

func regexMatch(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var str, pattern starlark.String
	if err := starlark.UnpackArgs(fn.Name(), args, kwargs, "str", &str, "pattern", &pattern); err != nil {
		return nil, err
	}
	
	re, err := regexp.Compile(string(pattern))
	if err != nil {
		return nil, fmt.Errorf("%s: invalid regex pattern: %w", fn.Name(), err)
	}
	
	return starlark.Bool(re.MatchString(string(str))), nil
}

func greaterThan(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("%s: want 2 arguments, got %d", fn.Name(), len(args))
	}
	
	// Use Starlark's comparison
	result, err := starlark.Compare(starlark.GT, args[0], args[1])
	if err != nil {
		return nil, err
	}
	return starlark.Bool(result), nil
}

func lessThan(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("%s: want 2 arguments, got %d", fn.Name(), len(args))
	}
	
	result, err := starlark.Compare(starlark.LT, args[0], args[1])
	if err != nil {
		return nil, err
	}
	return starlark.Bool(result), nil
}

func between(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if len(args) != 3 {
		return nil, fmt.Errorf("%s: want 3 arguments (value, min, max), got %d", fn.Name(), len(args))
	}
	
	// Check value >= min
	geMin, err := starlark.Compare(starlark.GE, args[0], args[1])
	if err != nil {
		return nil, err
	}
	
	// Check value <= max
	leMax, err := starlark.Compare(starlark.LE, args[0], args[2])
	if err != nil {
		return nil, err
	}
	
	return starlark.Bool(geMin && leMax), nil
}

func hasLength(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("%s: want 2 arguments (collection, length), got %d", fn.Name(), len(args))
	}
	
	collection, ok := args[0].(starlark.Indexable)
	if !ok {
		return nil, fmt.Errorf("%s: first argument must be a collection", fn.Name())
	}
	
	expectedLen, err := starlark.AsInt32(args[1])
	if err != nil {
		return nil, fmt.Errorf("%s: second argument must be an integer: %w", fn.Name(), err)
	}
	
	return starlark.Bool(collection.Len() == expectedLen), nil
}

func isSubset(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("%s: want 2 arguments (subset, superset), got %d", fn.Name(), len(args))
	}
	
	subset, ok := args[0].(*starlark.List)
	if !ok {
		return nil, fmt.Errorf("%s: first argument must be a list", fn.Name())
	}
	
	superset, ok := args[1].(*starlark.List)
	if !ok {
		return nil, fmt.Errorf("%s: second argument must be a list", fn.Name())
	}
	
	// Check if all elements in subset are in superset
	for i := 0; i < subset.Len(); i++ {
		found := false
		subElem := subset.Index(i)
		for j := 0; j < superset.Len(); j++ {
			if starlark.Equal(subElem, superset.Index(j)) {
				found = true
				break
			}
		}
		if !found {
			return starlark.Bool(false), nil
		}
	}
	
	return starlark.Bool(true), nil
}

// Type checking functions

func isString(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("%s: want 1 argument, got %d", fn.Name(), len(args))
	}
	_, ok := args[0].(starlark.String)
	return starlark.Bool(ok), nil
}

func isInt(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("%s: want 1 argument, got %d", fn.Name(), len(args))
	}
	_, ok := args[0].(starlark.Int)
	return starlark.Bool(ok), nil
}

func isFloat(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("%s: want 1 argument, got %d", fn.Name(), len(args))
	}
	_, ok := args[0].(starlark.Float)
	return starlark.Bool(ok), nil
}

func isList(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("%s: want 1 argument, got %d", fn.Name(), len(args))
	}
	_, ok := args[0].(*starlark.List)
	return starlark.Bool(ok), nil
}

func isDict(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("%s: want 1 argument, got %d", fn.Name(), len(args))
	}
	_, ok := args[0].(*starlark.Dict)
	return starlark.Bool(ok), nil
}

// SetGlobal adds a global variable to the interpreter
func (i *Interpreter) SetGlobal(name string, value interface{}) error {
	sv, err := goToStarlark(value)
	if err != nil {
		return fmt.Errorf("failed to convert value: %w", err)
	}
	i.globals[name] = sv
	return nil
}

// AddBuiltin adds a custom built-in function to the interpreter
func (i *Interpreter) AddBuiltin(name string, fn func(*starlark.Thread, *starlark.Builtin, starlark.Tuple, []starlark.Tuple) (starlark.Value, error)) {
	i.globals[name] = starlark.NewBuiltin(name, fn)
}