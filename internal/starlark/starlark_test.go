package starlark

import (
	"fmt"
	"testing"

	"go.starlark.net/starlark"
)

func TestNewInterpreter(t *testing.T) {
	interp := NewInterpreter()
	if interp == nil {
		t.Fatal("NewInterpreter returned nil")
	}
	if interp.thread == nil {
		t.Error("thread is nil")
	}
	if interp.globals == nil {
		t.Error("globals is nil")
	}

	// Verify built-ins are registered
	builtins := []string{
		"struct",
		"contains", "equals", "min_length", "max_length", "regex_match",
		"greater_than", "less_than", "between",
		"has_length", "is_subset",
		"is_string", "is_int", "is_float", "is_list", "is_dict",
	}

	for _, name := range builtins {
		if _, ok := interp.globals[name]; !ok {
			t.Errorf("built-in %q not registered", name)
		}
	}
}

func TestEval(t *testing.T) {
	tests := []struct {
		name    string
		expr    string
		vars    map[string]interface{}
		want    bool
		wantErr bool
	}{
		{
			name: "simple boolean true",
			expr: "True",
			vars: nil,
			want: true,
		},
		{
			name: "simple boolean false",
			expr: "False",
			vars: nil,
			want: false,
		},
		{
			name: "variable equality",
			expr: "x == 42",
			vars: map[string]interface{}{"x": 42},
			want: true,
		},
		{
			name: "string comparison",
			expr: "name == 'test'",
			vars: map[string]interface{}{"name": "test"},
			want: true,
		},
		{
			name: "contains function",
			expr: "contains('hello world', 'world')",
			vars: nil,
			want: true,
		},
		{
			name: "equals function",
			expr: "equals(x, 42)",
			vars: map[string]interface{}{"x": 42},
			want: true,
		},
		{
			name: "greater_than function",
			expr: "greater_than(x, 10)",
			vars: map[string]interface{}{"x": 20},
			want: true,
		},
		{
			name: "less_than function",
			expr: "less_than(x, 100)",
			vars: map[string]interface{}{"x": 50},
			want: true,
		},
		{
			name: "between function",
			expr: "between(x, 10, 100)",
			vars: map[string]interface{}{"x": 50},
			want: true,
		},
		{
			name: "min_length function",
			expr: "min_length('hello', 3)",
			vars: nil,
			want: true,
		},
		{
			name: "max_length function",
			expr: "max_length('hi', 5)",
			vars: nil,
			want: true,
		},
		{
			name: "regex_match function",
			expr: "regex_match('test123', r'^test\\d+$')",
			vars: nil,
			want: true,
		},
		{
			name: "is_string function",
			expr: "is_string(x)",
			vars: map[string]interface{}{"x": "hello"},
			want: true,
		},
		{
			name: "is_int function",
			expr: "is_int(x)",
			vars: map[string]interface{}{"x": 42},
			want: true,
		},
		{
			name: "is_list function",
			expr: "is_list(x)",
			vars: map[string]interface{}{"x": []interface{}{1, 2, 3}},
			want: true,
		},
		{
			name: "is_dict function",
			expr: "is_dict(x)",
			vars: map[string]interface{}{"x": map[string]interface{}{"key": "value"}},
			want: true,
		},
		{
			name: "has_length function",
			expr: "has_length([1, 2, 3], 3)",
			vars: nil,
			want: true,
		},
		{
			name: "is_subset function",
			expr: "is_subset([1, 2], [1, 2, 3, 4])",
			vars: nil,
			want: true,
		},
		{
			name: "complex expression",
			expr: "greater_than(len(text), 10) and contains(text, 'hello')",
			vars: map[string]interface{}{"text": "hello world, this is a test"},
			want: true,
		},
		{
			name: "none value",
			expr: "None",
			vars: nil,
			want: false,
		},
		{
			name:    "syntax error",
			expr:    "x == ",
			vars:    nil,
			wantErr: true,
		},
		{
			name:    "undefined variable",
			expr:    "undefined_var == 42",
			vars:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			interp := NewInterpreter()
			got, err := interp.Eval(tt.expr, tt.vars)
			if (err != nil) != tt.wantErr {
				t.Errorf("Eval() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("Eval() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGoToStarlark(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		check   func(t *testing.T, v starlark.Value)
		wantErr bool
	}{
		{
			name:  "nil to None",
			input: nil,
			check: func(t *testing.T, v starlark.Value) {
				if v != starlark.None {
					t.Errorf("got %v, want None", v)
				}
			},
		},
		{
			name:  "bool true",
			input: true,
			check: func(t *testing.T, v starlark.Value) {
				if v != starlark.True {
					t.Errorf("got %v, want True", v)
				}
			},
		},
		{
			name:  "bool false",
			input: false,
			check: func(t *testing.T, v starlark.Value) {
				if v != starlark.False {
					t.Errorf("got %v, want False", v)
				}
			},
		},
		{
			name:  "int",
			input: 42,
			check: func(t *testing.T, v starlark.Value) {
				i, ok := v.(starlark.Int)
				if !ok {
					t.Errorf("got type %T, want starlark.Int", v)
					return
				}
				val, _ := i.Int64()
				if val != 42 {
					t.Errorf("got %d, want 42", val)
				}
			},
		},
		{
			name:  "int64",
			input: int64(123),
			check: func(t *testing.T, v starlark.Value) {
				i, ok := v.(starlark.Int)
				if !ok {
					t.Errorf("got type %T, want starlark.Int", v)
					return
				}
				val, _ := i.Int64()
				if val != 123 {
					t.Errorf("got %d, want 123", val)
				}
			},
		},
		{
			name:  "float64",
			input: 3.14,
			check: func(t *testing.T, v starlark.Value) {
				f, ok := v.(starlark.Float)
				if !ok {
					t.Errorf("got type %T, want starlark.Float", v)
					return
				}
				if float64(f) != 3.14 {
					t.Errorf("got %f, want 3.14", f)
				}
			},
		},
		{
			name:  "string",
			input: "hello",
			check: func(t *testing.T, v starlark.Value) {
				s, ok := v.(starlark.String)
				if !ok {
					t.Errorf("got type %T, want starlark.String", v)
					return
				}
				if string(s) != "hello" {
					t.Errorf("got %q, want %q", s, "hello")
				}
			},
		},
		{
			name:  "slice",
			input: []interface{}{1, "two", 3.0},
			check: func(t *testing.T, v starlark.Value) {
				list, ok := v.(*starlark.List)
				if !ok {
					t.Errorf("got type %T, want *starlark.List", v)
					return
				}
				if list.Len() != 3 {
					t.Errorf("got length %d, want 3", list.Len())
				}
			},
		},
		{
			name:  "map",
			input: map[string]interface{}{"key": "value", "num": 42},
			check: func(t *testing.T, v starlark.Value) {
				dict, ok := v.(*starlark.Dict)
				if !ok {
					t.Errorf("got type %T, want *starlark.Dict", v)
					return
				}
				if dict.Len() != 2 {
					t.Errorf("got length %d, want 2", dict.Len())
				}
			},
		},
		{
			name:    "unsupported type",
			input:   struct{ X int }{X: 1},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := goToStarlark(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("goToStarlark() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.check != nil {
				tt.check(t, got)
			}
		})
	}
}

func TestSetGlobal(t *testing.T) {
	interp := NewInterpreter()

	tests := []struct {
		name    string
		key     string
		value   interface{}
		wantErr bool
	}{
		{
			name:  "set string",
			key:   "myvar",
			value: "hello",
		},
		{
			name:  "set int",
			key:   "num",
			value: 42,
		},
		{
			name:  "set list",
			key:   "items",
			value: []interface{}{1, 2, 3},
		},
		{
			name:    "unsupported type",
			key:     "bad",
			value:   struct{}{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := interp.SetGlobal(tt.key, tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("SetGlobal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if _, ok := interp.globals[tt.key]; !ok {
					t.Errorf("global %q not set", tt.key)
				}
			}
		})
	}
}

func TestAddBuiltin(t *testing.T) {
	interp := NewInterpreter()

	// Add a custom builtin
	customFn := func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		return starlark.String("custom"), nil
	}

	interp.AddBuiltin("my_custom_func", customFn)

	// Verify it was added
	if _, ok := interp.globals["my_custom_func"]; !ok {
		t.Error("custom builtin not added")
	}

	// Test using it in eval
	result, err := interp.Eval("my_custom_func()", nil)
	if err != nil {
		t.Fatalf("Eval() error = %v", err)
	}
	if !result {
		t.Error("expected custom function to return truthy value")
	}
}

func TestBuiltinFunctions(t *testing.T) {
	t.Run("contains", func(t *testing.T) {
		interp := NewInterpreter()
		tests := []struct {
			haystack string
			needle   string
			want     bool
		}{
			{"hello world", "world", true},
			{"hello world", "foo", false},
			{"", "", true},
		}

		for _, tt := range tests {
			expr := "contains('" + tt.haystack + "', '" + tt.needle + "')"
			got, err := interp.Eval(expr, nil)
			if err != nil {
				t.Fatalf("Eval() error = %v", err)
			}
			if got != tt.want {
				t.Errorf("contains(%q, %q) = %v, want %v", tt.haystack, tt.needle, got, tt.want)
			}
		}
	})

	t.Run("equals", func(t *testing.T) {
		interp := NewInterpreter()
		tests := []struct {
			expr string
			want bool
		}{
			{"equals(42, 42)", true},
			{"equals(42, 43)", false},
			{"equals('hello', 'hello')", true},
			{"equals('hello', 'world')", false},
		}

		for _, tt := range tests {
			got, err := interp.Eval(tt.expr, nil)
			if err != nil {
				t.Fatalf("Eval() error = %v", err)
			}
			if got != tt.want {
				t.Errorf("%s = %v, want %v", tt.expr, got, tt.want)
			}
		}
	})

	t.Run("min_length", func(t *testing.T) {
		interp := NewInterpreter()
		tests := []struct {
			str    string
			minLen int
			want   bool
		}{
			{"hello", 3, true},
			{"hello", 5, true},
			{"hello", 6, false},
			{"", 0, true},
			{"", 1, false},
		}

		for _, tt := range tests {
			expr := fmt.Sprintf("min_length('%s', %d)", tt.str, tt.minLen)
			got, err := interp.Eval(expr, nil)
			if err != nil {
				t.Fatalf("Eval() error = %v", err)
			}
			if got != tt.want {
				t.Errorf("min_length(%q, %d) = %v, want %v", tt.str, tt.minLen, got, tt.want)
			}
		}
	})

	t.Run("max_length", func(t *testing.T) {
		interp := NewInterpreter()
		tests := []struct {
			str    string
			maxLen int
			want   bool
		}{
			{"hello", 10, true},
			{"hello", 5, true},
			{"hello", 4, false},
		}

		for _, tt := range tests {
			expr := fmt.Sprintf("max_length('%s', %d)", tt.str, tt.maxLen)
			got, err := interp.Eval(expr, nil)
			if err != nil {
				t.Fatalf("Eval() error = %v", err)
			}
			if got != tt.want {
				t.Errorf("max_length(%q, %d) = %v, want %v", tt.str, tt.maxLen, got, tt.want)
			}
		}
	})

	t.Run("regex_match", func(t *testing.T) {
		interp := NewInterpreter()
		tests := []struct {
			str     string
			pattern string
			want    bool
			wantErr bool
		}{
			{"test123", `^test\d+$`, true, false},
			{"test", `^test\d+$`, false, false},
			{"hello@example.com", `^[\w.]+@[\w.]+$`, true, false},
			{"invalid", `[`, false, true}, // Invalid regex
		}

		for _, tt := range tests {
			expr := "regex_match('" + tt.str + "', r'" + tt.pattern + "')"
			got, err := interp.Eval(expr, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("Eval() error = %v, wantErr %v", err, tt.wantErr)
				continue
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("regex_match(%q, %q) = %v, want %v", tt.str, tt.pattern, got, tt.want)
			}
		}
	})

	t.Run("numeric comparisons", func(t *testing.T) {
		interp := NewInterpreter()
		tests := []struct {
			expr string
			want bool
		}{
			{"greater_than(10, 5)", true},
			{"greater_than(5, 10)", false},
			{"less_than(5, 10)", true},
			{"less_than(10, 5)", false},
			{"between(5, 1, 10)", true},
			{"between(0, 1, 10)", false},
			{"between(11, 1, 10)", false},
		}

		for _, tt := range tests {
			got, err := interp.Eval(tt.expr, nil)
			if err != nil {
				t.Fatalf("Eval() error = %v", err)
			}
			if got != tt.want {
				t.Errorf("%s = %v, want %v", tt.expr, got, tt.want)
			}
		}
	})

	t.Run("has_length", func(t *testing.T) {
		interp := NewInterpreter()
		tests := []struct {
			expr string
			want bool
		}{
			{"has_length([1, 2, 3], 3)", true},
			{"has_length([1, 2, 3], 2)", false},
			{"has_length([], 0)", true},
			{"has_length('hello', 5)", true},
		}

		for _, tt := range tests {
			got, err := interp.Eval(tt.expr, nil)
			if err != nil {
				t.Fatalf("Eval() error = %v", err)
			}
			if got != tt.want {
				t.Errorf("%s = %v, want %v", tt.expr, got, tt.want)
			}
		}
	})

	t.Run("is_subset", func(t *testing.T) {
		interp := NewInterpreter()
		tests := []struct {
			expr string
			want bool
		}{
			{"is_subset([1, 2], [1, 2, 3])", true},
			{"is_subset([1, 4], [1, 2, 3])", false},
			{"is_subset([], [1, 2, 3])", true},
		}

		for _, tt := range tests {
			got, err := interp.Eval(tt.expr, nil)
			if err != nil {
				t.Fatalf("Eval() error = %v", err)
			}
			if got != tt.want {
				t.Errorf("%s = %v, want %v", tt.expr, got, tt.want)
			}
		}
	})

	t.Run("type checks", func(t *testing.T) {
		interp := NewInterpreter()
		tests := []struct {
			expr string
			want bool
		}{
			{"is_string('hello')", true},
			{"is_string(42)", false},
			{"is_int(42)", true},
			{"is_int('42')", false},
			{"is_float(3.14)", true},
			{"is_float(3)", false},
			{"is_list([1, 2, 3])", true},
			{"is_list(42)", false},
			{"is_dict({'key': 'value'})", true},
			{"is_dict([1, 2])", false},
		}

		for _, tt := range tests {
			got, err := interp.Eval(tt.expr, nil)
			if err != nil {
				t.Fatalf("Eval() error = %v", err)
			}
			if got != tt.want {
				t.Errorf("%s = %v, want %v", tt.expr, got, tt.want)
			}
		}
	})
}

func TestEvalWithVariables(t *testing.T) {
	interp := NewInterpreter()

	tests := []struct {
		name string
		expr string
		vars map[string]interface{}
		want bool
	}{
		{
			name: "simple variable",
			expr: "x > 10",
			vars: map[string]interface{}{"x": 20},
			want: true,
		},
		{
			name: "multiple variables",
			expr: "x + y == 10",
			vars: map[string]interface{}{"x": 3, "y": 7},
			want: true,
		},
		{
			name: "list variable",
			expr: "has_length(items, 3)",
			vars: map[string]interface{}{"items": []interface{}{1, 2, 3}},
			want: true,
		},
		{
			name: "dict variable",
			expr: "is_dict(data) and data['key'] == 'value'",
			vars: map[string]interface{}{"data": map[string]interface{}{"key": "value"}},
			want: true,
		},
		{
			name: "nested structures",
			expr: "users[0]['name'] == 'Alice'",
			vars: map[string]interface{}{
				"users": []interface{}{
					map[string]interface{}{"name": "Alice"},
				},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := interp.Eval(tt.expr, tt.vars)
			if err != nil {
				t.Fatalf("Eval() error = %v", err)
			}
			if got != tt.want {
				t.Errorf("Eval() = %v, want %v", got, tt.want)
			}
		})
	}
}
