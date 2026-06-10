package evaluator

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Deterministic agent tool-call / trajectory assertions. They read the agent's
// recorded trajectory from metadata["_trajectory"] — a list of steps, each a
// {name, arguments} object (arguments may be an object or a JSON string). When
// no trajectory var is present, they fall back to parsing OpenAI tool_calls
// from the output, so a single tool-calling response is also gradable. With no
// usable trajectory at all they fail with a clear message rather than passing.

// toolStep is one recorded tool invocation in a trajectory.
type toolStep struct {
	Name string
	Args map[string]interface{}
}

// trajectorySteps extracts the ordered tool-call steps from metadata, falling
// back to OpenAI tool_calls embedded in the output.
func trajectorySteps(output string, metadata map[string]interface{}) ([]toolStep, bool) {
	if metadata != nil {
		if raw, ok := metadata["_trajectory"].([]interface{}); ok {
			return parseTrajectory(raw), true
		}
	}
	if steps, ok := parseOutputToolCalls(output); ok {
		return steps, true
	}
	return nil, false
}

// parseTrajectory converts a raw _trajectory list into tool steps. Each entry
// may carry the call directly ({name, arguments}) or nest it under "function".
func parseTrajectory(raw []interface{}) []toolStep {
	steps := make([]toolStep, 0, len(raw))
	for _, e := range raw {
		entry, ok := e.(map[string]interface{})
		if !ok {
			continue
		}
		call := entry
		if fn, ok := entry["function"].(map[string]interface{}); ok {
			call = fn
		}
		name, _ := call["name"].(string)
		if name == "" {
			continue
		}
		steps = append(steps, toolStep{Name: name, Args: coerceArgs(call["arguments"])})
	}
	return steps
}

// parseOutputToolCalls extracts tool steps from an output that is (or contains)
// an OpenAI tool_calls array.
func parseOutputToolCalls(output string) ([]toolStep, bool) {
	var raw interface{}
	if err := json.Unmarshal([]byte(output), &raw); err != nil {
		return nil, false
	}
	var calls []interface{}
	switch v := raw.(type) {
	case []interface{}:
		calls = v
	case map[string]interface{}:
		if tc, ok := v["tool_calls"].([]interface{}); ok {
			calls = tc
		} else {
			return nil, false
		}
	default:
		return nil, false
	}
	return parseTrajectory(calls), true
}

// coerceArgs normalizes a tool call's arguments into a map, accepting either an
// object or a stringified JSON object.
func coerceArgs(v interface{}) map[string]interface{} {
	switch a := v.(type) {
	case map[string]interface{}:
		return a
	case string:
		var m map[string]interface{}
		if json.Unmarshal([]byte(a), &m) == nil {
			return m
		}
	}
	return nil
}

func toolNames(steps []toolStep) []string {
	names := make([]string, len(steps))
	for i, s := range steps {
		names[i] = s.Name
	}
	return names
}

// evaluateToolCallF1 scores the F1 of the called tool-name set against the
// expected tool-name set (assertion.Value, a list), passing when F1 >=
// threshold (default 1.0 — every expected tool used and no extras).
func (ae *AssertionEvaluator) evaluateToolCallF1(assertion Assertion, output string, metadata map[string]interface{}) *AssertionResult {
	steps, ok := trajectorySteps(output, metadata)
	if !ok {
		return failResult(assertion, "tool-call-f1 requires a tool-call trajectory (_trajectory var or tool_calls output)")
	}
	expected := toStringSlice(assertion.Value)
	if len(expected) == 0 {
		return failResult(assertion, "tool-call-f1 requires the expected tool names as a list value")
	}
	actualSet := stringSet(toolNames(steps))
	expectedSet := stringSet(expected)
	tp := 0
	for name := range expectedSet {
		if actualSet[name] {
			tp++
		}
	}
	precision, recall := 0.0, 0.0
	if len(actualSet) > 0 {
		precision = float64(tp) / float64(len(actualSet))
	}
	if len(expectedSet) > 0 {
		recall = float64(tp) / float64(len(expectedSet))
	}
	f1 := fMeasure(precision, recall, 1.0)
	threshold := 1.0
	if assertion.Threshold != nil {
		threshold = *assertion.Threshold
	}
	return scoreResult(assertion, f1, threshold, fmt.Sprintf("tool-call F1 %.3f (threshold %.2f)", f1, threshold), map[string]interface{}{"method": "tool_call_f1", "called": toolNames(steps)})
}

// evaluateTrajectoryToolUsed passes when the trajectory invoked the named tool
// (assertion.Value, a string).
func (ae *AssertionEvaluator) evaluateTrajectoryToolUsed(assertion Assertion, output string, metadata map[string]interface{}) *AssertionResult {
	steps, ok := trajectorySteps(output, metadata)
	if !ok {
		return failResult(assertion, "trajectory:tool-used requires a tool-call trajectory")
	}
	want, ok := assertion.Value.(string)
	if !ok || want == "" {
		return failResult(assertion, "trajectory:tool-used requires the tool name as a string value")
	}
	used := stringSet(toolNames(steps))[want]
	return boolResult(assertion, used, fmt.Sprintf("tool %q %s", want, passOrNot(used, "was used", "was not used")))
}

// evaluateTrajectoryToolSequence passes when the expected ordered tool-name
// sequence (assertion.Value, a list) appears as a contiguous run in the actual
// tool sequence.
func (ae *AssertionEvaluator) evaluateTrajectoryToolSequence(assertion Assertion, output string, metadata map[string]interface{}) *AssertionResult {
	steps, ok := trajectorySteps(output, metadata)
	if !ok {
		return failResult(assertion, "trajectory:tool-sequence requires a tool-call trajectory")
	}
	want := toStringSlice(assertion.Value)
	if len(want) == 0 {
		return failResult(assertion, "trajectory:tool-sequence requires the expected sequence as a list value")
	}
	actual := toolNames(steps)
	matched := containsSubsequenceContiguous(actual, want)
	return boolResult(assertion, matched, fmt.Sprintf("tool sequence %v %s in %v", want, passOrNot(matched, "found", "not found"), actual))
}

// evaluateTrajectoryToolArgsMatch passes when some step called the expected
// tool with arguments that match the expected arguments (a subset match: every
// expected key/value must be present). assertion.Value is {name, arguments}.
func (ae *AssertionEvaluator) evaluateTrajectoryToolArgsMatch(assertion Assertion, output string, metadata map[string]interface{}) *AssertionResult {
	steps, ok := trajectorySteps(output, metadata)
	if !ok {
		return failResult(assertion, "trajectory:tool-args-match requires a tool-call trajectory")
	}
	spec, ok := assertion.Value.(map[string]interface{})
	if !ok {
		return failResult(assertion, "trajectory:tool-args-match requires a {name, arguments} object value")
	}
	wantName, _ := spec["name"].(string)
	wantArgs := coerceArgs(spec["arguments"])
	for _, s := range steps {
		if wantName != "" && s.Name != wantName {
			continue
		}
		if argsSubsetMatch(wantArgs, s.Args) {
			return boolResult(assertion, true, fmt.Sprintf("tool %q called with matching arguments", s.Name))
		}
	}
	return boolResult(assertion, false, fmt.Sprintf("no call to %q matched the expected arguments", wantName))
}

// evaluateTrajectoryStepCount checks the number of trajectory steps against
// min/max bounds or an exact value (assertion.Value).
func (ae *AssertionEvaluator) evaluateTrajectoryStepCount(assertion Assertion, output string, metadata map[string]interface{}) *AssertionResult {
	steps, ok := trajectorySteps(output, metadata)
	if !ok {
		return failResult(assertion, "trajectory:step-count requires a tool-call trajectory")
	}
	count := len(steps)
	var passed bool
	var message string
	switch {
	case assertion.Min != nil && assertion.Max != nil:
		passed = count >= int(*assertion.Min) && count <= int(*assertion.Max)
		message = fmt.Sprintf("step count %d (expected between %d and %d)", count, int(*assertion.Min), int(*assertion.Max))
	case assertion.Min != nil:
		passed = count >= int(*assertion.Min)
		message = fmt.Sprintf("step count %d (expected at least %d)", count, int(*assertion.Min))
	case assertion.Max != nil:
		passed = count <= int(*assertion.Max)
		message = fmt.Sprintf("step count %d (expected at most %d)", count, int(*assertion.Max))
	default:
		exact, ok := assertionInt(assertion.Value)
		if !ok {
			return failResult(assertion, "trajectory:step-count requires min/max bounds or an exact integer value")
		}
		passed = count == exact
		message = fmt.Sprintf("step count %d (expected %d)", count, exact)
	}
	return &AssertionResult{Type: assertion.Type, Passed: passed, Score: boolScore(passed), Actual: count, Message: message}
}

// --- shared helpers ---

func stringSet(items []string) map[string]bool {
	set := make(map[string]bool, len(items))
	for _, s := range items {
		set[s] = true
	}
	return set
}

// containsSubsequenceContiguous reports whether want appears as a contiguous
// run inside seq.
func containsSubsequenceContiguous(seq, want []string) bool {
	if len(want) == 0 || len(want) > len(seq) {
		return false
	}
	for i := 0; i+len(want) <= len(seq); i++ {
		if equalStrings(seq[i:i+len(want)], want) {
			return true
		}
	}
	return false
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// argsSubsetMatch reports whether every key/value in want is present and equal
// in got. Values are compared by their JSON-normalized form.
func argsSubsetMatch(want, got map[string]interface{}) bool {
	if len(want) == 0 {
		return true
	}
	if got == nil {
		return false
	}
	for k, wv := range want {
		gv, ok := got[k]
		if !ok || !jsonEqual(wv, gv) {
			return false
		}
	}
	return true
}

func jsonEqual(a, b interface{}) bool {
	ab, err1 := json.Marshal(a)
	bb, err2 := json.Marshal(b)
	if err1 != nil || err2 != nil {
		return fmt.Sprintf("%v", a) == fmt.Sprintf("%v", b)
	}
	return strings.EqualFold(string(ab), string(bb))
}
