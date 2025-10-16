package starlark

import (
	"testing"

	"go.starlark.net/starlark"
)

func TestNewEvaluator(t *testing.T) {
	e := NewEvaluator()
	if e == nil {
		t.Fatal("NewEvaluator returned nil")
	}
	if e.globals == nil {
		t.Error("globals is nil")
	}

	// Verify built-ins are registered
	builtins := []string{"contains", "min_length", "max_length", "equals", "regex_match", "word_count", "struct"}
	for _, name := range builtins {
		if _, ok := e.globals[name]; !ok {
			t.Errorf("built-in %q not registered", name)
		}
	}
}

func TestEvalFile(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		src      string
		wantErr  bool
		checkFn  func(t *testing.T, globals starlark.StringDict)
	}{
		{
			name:     "simple variable definition",
			filename: "test.star",
			src:      `x = 42`,
			wantErr:  false,
			checkFn: func(t *testing.T, globals starlark.StringDict) {
				if _, ok := globals["x"]; !ok {
					t.Error("variable 'x' not defined")
				}
			},
		},
		{
			name:     "function definition",
			filename: "test.star",
			src: `def test_func(response):
  return True`,
			wantErr: false,
			checkFn: func(t *testing.T, globals starlark.StringDict) {
				if _, ok := globals["test_func"]; !ok {
					t.Error("function 'test_func' not defined")
				}
			},
		},
		{
			name:     "using built-in function",
			filename: "test.star",
			src:      `result = contains("hello world", "world")`,
			wantErr:  false,
			checkFn: func(t *testing.T, globals starlark.StringDict) {
				if _, ok := globals["result"]; !ok {
					t.Error("variable 'result' not defined")
				}
			},
		},
		{
			name:     "syntax error",
			filename: "bad.star",
			src:      `def test_func(\n`,
			wantErr:  true,
		},
		{
			name:     "runtime error",
			filename: "error.star",
			src:      `x = undefined_var + 1`,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := NewEvaluator()
			globals, err := e.EvalFile(tt.filename, []byte(tt.src))
			if (err != nil) != tt.wantErr {
				t.Errorf("EvalFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.checkFn != nil {
				tt.checkFn(t, globals)
			}
		})
	}
}

func TestEvaluateTest(t *testing.T) {
	tests := []struct {
		name       string
		starlarkSrc string
		testName   string
		response   string
		wantPass   bool
		wantErr    bool
	}{
		{
			name: "test returns true",
			starlarkSrc: `
def test_simple(response):
    return True
`,
			testName: "test_simple",
			response: "any response",
			wantPass: true,
			wantErr:  false,
		},
		{
			name: "test returns false",
			starlarkSrc: `
def test_fail(response):
    return False
`,
			testName: "test_fail",
			response: "any response",
			wantPass: false,
			wantErr:  false,
		},
		{
			name: "test with dict result - pass",
			starlarkSrc: `
def test_dict(response):
    return {
        "pass": True,
        "score": 0.95,
        "reason": "Excellent response"
    }
`,
			testName: "test_dict",
			response: "test response",
			wantPass: true,
			wantErr:  false,
		},
		{
			name: "test with dict result - fail",
			starlarkSrc: `
def test_dict_fail(response):
    return {
        "pass": False,
        "score": 0.3,
        "reason": "Missing key information"
    }
`,
			testName: "test_dict_fail",
			response: "test response",
			wantPass: false,
			wantErr:  false,
		},
		{
			name: "test using contains builtin",
			starlarkSrc: `
def test_contains(response):
    return contains(response, "world")
`,
			testName: "test_contains",
			response: "hello world",
			wantPass: true,
			wantErr:  false,
		},
		{
			name: "test using word_count builtin",
			starlarkSrc: `
def test_word_count(response):
    count = word_count(response)
    return count > 5
`,
			testName: "test_word_count",
			response: "this is a test response with many words",
			wantPass: true,
			wantErr:  false,
		},
		{
			name: "non-existent test function",
			starlarkSrc: `
def test_exists(response):
    return True
`,
			testName: "test_nonexistent",
			response: "any response",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := NewEvaluator()
			globals, err := e.EvalFile("test.star", []byte(tt.starlarkSrc))
			if err != nil {
				t.Fatalf("EvalFile() failed: %v", err)
			}

			result, err := e.EvaluateTest(globals, tt.testName, tt.response)
			if (err != nil) != tt.wantErr {
				t.Errorf("EvaluateTest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && result.Pass != tt.wantPass {
				t.Errorf("EvaluateTest() pass = %v, want %v", result.Pass, tt.wantPass)
			}
		})
	}
}

func TestConvertToTestResult(t *testing.T) {
	tests := []struct {
		name    string
		val     starlark.Value
		wantErr bool
		check   func(t *testing.T, result *TestResult)
	}{
		{
			name:    "boolean true",
			val:     starlark.True,
			wantErr: false,
			check: func(t *testing.T, result *TestResult) {
				if !result.Pass {
					t.Error("expected Pass to be true")
				}
			},
		},
		{
			name:    "boolean false",
			val:     starlark.False,
			wantErr: false,
			check: func(t *testing.T, result *TestResult) {
				if result.Pass {
					t.Error("expected Pass to be false")
				}
			},
		},
		{
			name: "dict with pass field",
			val: func() starlark.Value {
				d := starlark.NewDict(1)
				d.SetKey(starlark.String("pass"), starlark.True)
				return d
			}(),
			wantErr: false,
			check: func(t *testing.T, result *TestResult) {
				if !result.Pass {
					t.Error("expected Pass to be true")
				}
			},
		},
		{
			name: "dict with score field",
			val: func() starlark.Value {
				d := starlark.NewDict(2)
				d.SetKey(starlark.String("pass"), starlark.True)
				d.SetKey(starlark.String("score"), starlark.Float(0.85))
				return d
			}(),
			wantErr: false,
			check: func(t *testing.T, result *TestResult) {
				if result.Score != 0.85 {
					t.Errorf("expected Score 0.85, got %f", result.Score)
				}
			},
		},
		{
			name: "dict with reason field",
			val: func() starlark.Value {
				d := starlark.NewDict(2)
				d.SetKey(starlark.String("pass"), starlark.False)
				d.SetKey(starlark.String("reason"), starlark.String("failed validation"))
				return d
			}(),
			wantErr: false,
			check: func(t *testing.T, result *TestResult) {
				if result.Reason != "failed validation" {
					t.Errorf("expected Reason 'failed validation', got %q", result.Reason)
				}
			},
		},
		{
			name: "dict with custom details",
			val: func() starlark.Value {
				d := starlark.NewDict(3)
				d.SetKey(starlark.String("pass"), starlark.True)
				d.SetKey(starlark.String("custom"), starlark.String("value"))
				d.SetKey(starlark.String("count"), starlark.MakeInt(42))
				return d
			}(),
			wantErr: false,
			check: func(t *testing.T, result *TestResult) {
				if result.Details == nil {
					t.Error("expected Details to be non-nil")
					return
				}
				if result.Details["custom"] != "value" {
					t.Errorf("expected Details['custom'] = 'value', got %v", result.Details["custom"])
				}
			},
		},
		{
			name:    "unsupported type",
			val:     starlark.String("invalid"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := convertToTestResult(tt.val)
			if (err != nil) != tt.wantErr {
				t.Errorf("convertToTestResult() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.check != nil {
				tt.check(t, result)
			}
		})
	}
}

func TestConvertStarlarkValue(t *testing.T) {
	tests := []struct {
		name string
		val  starlark.Value
		want interface{}
	}{
		{
			name: "string",
			val:  starlark.String("hello"),
			want: "hello",
		},
		{
			name: "int",
			val:  starlark.MakeInt(42),
			want: int64(42),
		},
		{
			name: "float",
			val:  starlark.Float(3.14),
			want: float64(3.14),
		},
		{
			name: "bool true",
			val:  starlark.True,
			want: true,
		},
		{
			name: "bool false",
			val:  starlark.False,
			want: false,
		},
		{
			name: "list",
			val: func() starlark.Value {
				return starlark.NewList([]starlark.Value{
					starlark.MakeInt(1),
					starlark.String("two"),
					starlark.Float(3.0),
				})
			}(),
			want: []interface{}{int64(1), "two", float64(3.0)},
		},
		{
			name: "dict",
			val: func() starlark.Value {
				d := starlark.NewDict(2)
				d.SetKey(starlark.String("key1"), starlark.String("value1"))
				d.SetKey(starlark.String("key2"), starlark.MakeInt(42))
				return d
			}(),
			want: map[string]interface{}{"key1": "value1", "key2": int64(42)},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := convertStarlarkValue(tt.val)
			// Simple equality check (not deep comparison for simplicity)
			switch want := tt.want.(type) {
			case string:
				if got != want {
					t.Errorf("convertStarlarkValue() = %v, want %v", got, want)
				}
			case int64:
				if got != want {
					t.Errorf("convertStarlarkValue() = %v, want %v", got, want)
				}
			case float64:
				if got != want {
					t.Errorf("convertStarlarkValue() = %v, want %v", got, want)
				}
			case bool:
				if got != want {
					t.Errorf("convertStarlarkValue() = %v, want %v", got, want)
				}
			case []interface{}:
				gotSlice, ok := got.([]interface{})
				if !ok {
					t.Errorf("convertStarlarkValue() returned %T, want []interface{}", got)
					return
				}
				if len(gotSlice) != len(want) {
					t.Errorf("convertStarlarkValue() length = %d, want %d", len(gotSlice), len(want))
				}
			case map[string]interface{}:
				gotMap, ok := got.(map[string]interface{})
				if !ok {
					t.Errorf("convertStarlarkValue() returned %T, want map[string]interface{}", got)
					return
				}
				if len(gotMap) != len(want) {
					t.Errorf("convertStarlarkValue() map size = %d, want %d", len(gotMap), len(want))
				}
			}
		})
	}
}

func TestBuiltinFunctions(t *testing.T) {
	e := NewEvaluator()

	t.Run("contains", func(t *testing.T) {
		src := `result = contains("hello world", "world")`
		globals, err := e.EvalFile("test.star", []byte(src))
		if err != nil {
			t.Fatalf("EvalFile() error = %v", err)
		}
		result, ok := globals["result"].(starlark.Bool)
		if !ok {
			t.Fatal("result is not a bool")
		}
		if !bool(result) {
			t.Error("expected contains to return true")
		}
	})

	t.Run("min_length", func(t *testing.T) {
		src := `result = min_length("hello", 3)`
		globals, err := e.EvalFile("test.star", []byte(src))
		if err != nil {
			t.Fatalf("EvalFile() error = %v", err)
		}
		result, ok := globals["result"].(starlark.Bool)
		if !ok {
			t.Fatal("result is not a bool")
		}
		if !bool(result) {
			t.Error("expected min_length to return true")
		}
	})

	t.Run("max_length", func(t *testing.T) {
		src := `result = max_length("hi", 5)`
		globals, err := e.EvalFile("test.star", []byte(src))
		if err != nil {
			t.Fatalf("EvalFile() error = %v", err)
		}
		result, ok := globals["result"].(starlark.Bool)
		if !ok {
			t.Fatal("result is not a bool")
		}
		if !bool(result) {
			t.Error("expected max_length to return true")
		}
	})

	t.Run("equals", func(t *testing.T) {
		src := `result = equals("hello", "hello")`
		globals, err := e.EvalFile("test.star", []byte(src))
		if err != nil {
			t.Fatalf("EvalFile() error = %v", err)
		}
		result, ok := globals["result"].(starlark.Bool)
		if !ok {
			t.Fatal("result is not a bool")
		}
		if !bool(result) {
			t.Error("expected equals to return true")
		}
	})

	t.Run("regex_match", func(t *testing.T) {
		src := `result = regex_match("test123", "^test\\d+$")`
		globals, err := e.EvalFile("test.star", []byte(src))
		if err != nil {
			t.Fatalf("EvalFile() error = %v", err)
		}
		result, ok := globals["result"].(starlark.Bool)
		if !ok {
			t.Fatal("result is not a bool")
		}
		if !bool(result) {
			t.Error("expected regex_match to return true")
		}
	})

	t.Run("word_count", func(t *testing.T) {
		src := `result = word_count("hello world test")`
		globals, err := e.EvalFile("test.star", []byte(src))
		if err != nil {
			t.Fatalf("EvalFile() error = %v", err)
		}
		result, ok := globals["result"].(starlark.Int)
		if !ok {
			t.Fatal("result is not an int")
		}
		count, _ := result.Int64()
		if count != 3 {
			t.Errorf("expected word_count = 3, got %d", count)
		}
	})

	t.Run("regex_match with invalid pattern", func(t *testing.T) {
		src := `result = regex_match("test", "[")`
		_, err := e.EvalFile("test.star", []byte(src))
		if err == nil {
			t.Error("expected error for invalid regex pattern")
		}
	})
}

func TestMakeBuiltins(t *testing.T) {
	builtins := makeBuiltins()

	required := []string{"contains", "min_length", "max_length", "equals", "regex_match", "word_count", "struct"}
	for _, name := range required {
		if _, ok := builtins[name]; !ok {
			t.Errorf("built-in %q not found", name)
		}
	}
}
