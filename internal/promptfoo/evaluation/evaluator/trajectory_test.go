package evaluator

import (
	"context"
	"testing"
)

// sampleTrajectory builds a _trajectory metadata var from {name, args} pairs.
func sampleTrajectory(steps ...map[string]interface{}) map[string]interface{} {
	raw := make([]interface{}, len(steps))
	for i, s := range steps {
		raw[i] = s
	}
	return map[string]interface{}{"_trajectory": raw}
}

func step(name string, args map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{"name": name, "arguments": args}
}

func TestEvaluateToolCallF1(t *testing.T) {
	ae := NewAssertionEvaluator(nil)
	meta := sampleTrajectory(step("search", nil), step("fetch", nil))
	tests := []struct {
		name      string
		expected  interface{}
		threshold *float64
		want      bool
	}{
		{"exact-set", []interface{}{"search", "fetch"}, nil, true},         // F1 = 1.0
		{"missing-one", []interface{}{"search", "fetch", "x"}, nil, false}, // recall 2/3
		{"partial-low-threshold", []interface{}{"search", "fetch", "x"}, floatPtr(0.7), true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			a := Assertion{Type: AssertionToolCallF1, Value: tc.expected, Threshold: tc.threshold}
			got, err := ae.EvaluateAssertion(context.Background(), a, "", meta)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Passed != tc.want {
				t.Errorf("Passed = %v, want %v (score %v, %s)", got.Passed, tc.want, got.Score, got.Message)
			}
		})
	}
}

func TestEvaluateTrajectoryToolUsed(t *testing.T) {
	ae := NewAssertionEvaluator(nil)
	meta := sampleTrajectory(step("search", nil), step("fetch", nil))
	for tool, want := range map[string]bool{"search": true, "delete": false} {
		a := Assertion{Type: AssertionTrajToolUsed, Value: tool}
		got, _ := ae.EvaluateAssertion(context.Background(), a, "", meta)
		if got.Passed != want {
			t.Errorf("tool-used(%q) = %v, want %v", tool, got.Passed, want)
		}
	}
}

func TestEvaluateTrajectoryToolSequence(t *testing.T) {
	ae := NewAssertionEvaluator(nil)
	meta := sampleTrajectory(step("a", nil), step("b", nil), step("c", nil))
	tests := []struct {
		seq  []interface{}
		want bool
	}{
		{[]interface{}{"a", "b"}, true},
		{[]interface{}{"b", "c"}, true},
		{[]interface{}{"a", "c"}, false}, // not contiguous
		{[]interface{}{"c", "a"}, false},
	}
	for _, tc := range tests {
		a := Assertion{Type: AssertionTrajToolSequence, Value: tc.seq}
		got, _ := ae.EvaluateAssertion(context.Background(), a, "", meta)
		if got.Passed != tc.want {
			t.Errorf("sequence %v = %v, want %v", tc.seq, got.Passed, tc.want)
		}
	}
}

func TestEvaluateTrajectoryToolArgsMatch(t *testing.T) {
	ae := NewAssertionEvaluator(nil)
	meta := sampleTrajectory(step("search", map[string]interface{}{"q": "cats", "limit": float64(5)}))
	tests := []struct {
		name string
		spec map[string]interface{}
		want bool
	}{
		{"subset-match", map[string]interface{}{"name": "search", "arguments": map[string]interface{}{"q": "cats"}}, true},
		{"full-match", map[string]interface{}{"name": "search", "arguments": map[string]interface{}{"q": "cats", "limit": float64(5)}}, true},
		{"wrong-value", map[string]interface{}{"name": "search", "arguments": map[string]interface{}{"q": "dogs"}}, false},
		{"wrong-tool", map[string]interface{}{"name": "fetch", "arguments": map[string]interface{}{"q": "cats"}}, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			a := Assertion{Type: AssertionTrajToolArgsMatch, Value: tc.spec}
			got, _ := ae.EvaluateAssertion(context.Background(), a, "", meta)
			if got.Passed != tc.want {
				t.Errorf("args-match = %v, want %v (%s)", got.Passed, tc.want, got.Message)
			}
		})
	}
}

func TestEvaluateTrajectoryStepCount(t *testing.T) {
	ae := NewAssertionEvaluator(nil)
	meta := sampleTrajectory(step("a", nil), step("b", nil), step("c", nil))
	tests := []struct {
		name     string
		value    interface{}
		min, max *float64
		want     bool
	}{
		{"exact-pass", float64(3), nil, nil, true},
		{"exact-fail", float64(2), nil, nil, false},
		{"max-pass", nil, nil, floatPtr(5), true},
		{"max-fail", nil, nil, floatPtr(2), false},
		{"min-pass", nil, floatPtr(2), nil, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			a := Assertion{Type: AssertionTrajStepCount, Value: tc.value, Min: tc.min, Max: tc.max}
			got, _ := ae.EvaluateAssertion(context.Background(), a, "", meta)
			if got.Passed != tc.want {
				t.Errorf("step-count = %v, want %v (%s)", got.Passed, tc.want, got.Message)
			}
		})
	}
}

func TestTrajectoryFallbackToToolCalls(t *testing.T) {
	ae := NewAssertionEvaluator(nil)
	// No _trajectory var: parse OpenAI tool_calls from the output instead.
	output := `{"tool_calls":[{"type":"function","function":{"name":"search","arguments":"{}"}}]}`
	a := Assertion{Type: AssertionTrajToolUsed, Value: "search"}
	got, err := ae.EvaluateAssertion(context.Background(), a, output, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got.Passed {
		t.Errorf("expected tool-used to read tool_calls from output, got fail (%s)", got.Message)
	}
}

func TestTrajectoryMissingFails(t *testing.T) {
	ae := NewAssertionEvaluator(nil)
	a := Assertion{Type: AssertionToolCallF1, Value: []interface{}{"x"}}
	got, _ := ae.EvaluateAssertion(context.Background(), a, "not json", nil)
	if got.Passed {
		t.Error("expected fail when no trajectory is available")
	}
}

func TestTrajectoryTraceAliases(t *testing.T) {
	ids := map[string]AssertionType{
		"tool-call-f1":               AssertionToolCallF1,
		"trajectory:tool-used":       AssertionTrajToolUsed,
		"trajectory:tool-sequence":   AssertionTrajToolSequence,
		"trajectory:tool-args-match": AssertionTrajToolArgsMatch,
		"trajectory:step-count":      AssertionTrajStepCount,
		"trace-span-count":           AssertionTraceSpanCount,
		"trace-span-duration":        AssertionTraceSpanDuration,
		"trace-error-spans":          AssertionTraceErrorSpans,
	}
	for id, want := range ids {
		canonical, _, ok := normalizeAssertionType(id)
		if !ok || canonical != want {
			t.Errorf("%q => (%q, ok=%v), want %q", id, canonical, ok, want)
		}
	}
}

func TestEvaluateTraceAssertions(t *testing.T) {
	ae := NewAssertionEvaluator(nil)
	meta := map[string]interface{}{"_trace": []interface{}{
		map[string]interface{}{"name": "root", "durationMs": float64(120), "error": false},
		map[string]interface{}{"name": "db", "durationMs": float64(40), "error": true},
	}}
	// span-count exact 2
	if got, _ := ae.EvaluateAssertion(context.Background(), Assertion{Type: AssertionTraceSpanCount, Value: float64(2)}, "", meta); !got.Passed {
		t.Errorf("span-count=2 should pass (%s)", got.Message)
	}
	// max span duration <= 200ms passes; <= 100ms fails (root is 120)
	if got, _ := ae.EvaluateAssertion(context.Background(), Assertion{Type: AssertionTraceSpanDuration, Max: floatPtr(200)}, "", meta); !got.Passed {
		t.Errorf("span-duration<=200 should pass (%s)", got.Message)
	}
	if got, _ := ae.EvaluateAssertion(context.Background(), Assertion{Type: AssertionTraceSpanDuration, Max: floatPtr(100)}, "", meta); got.Passed {
		t.Errorf("span-duration<=100 should fail (root span 120ms)")
	}
	// error-spans: default max 0 fails (one errored span); max 1 passes
	if got, _ := ae.EvaluateAssertion(context.Background(), Assertion{Type: AssertionTraceErrorSpans}, "", meta); got.Passed {
		t.Errorf("error-spans default (max 0) should fail with one errored span")
	}
	if got, _ := ae.EvaluateAssertion(context.Background(), Assertion{Type: AssertionTraceErrorSpans, Max: floatPtr(1)}, "", meta); !got.Passed {
		t.Errorf("error-spans max 1 should pass (%s)", got.Message)
	}
	// missing _trace fails clearly
	if got, _ := ae.EvaluateAssertion(context.Background(), Assertion{Type: AssertionTraceSpanCount, Value: float64(1)}, "", nil); got.Passed {
		t.Errorf("trace-span-count with no _trace should fail")
	}
}
