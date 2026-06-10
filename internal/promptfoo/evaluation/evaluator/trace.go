package evaluator

import "fmt"

// Deterministic trace assertions. They read a recorded trace from
// metadata["_trace"] — a list of spans, each a {name, durationMs, error}
// object. promptfoo collects these from an OpenTelemetry exporter; pe has no
// in-process trace collector, so it grades whatever trace the config supplies
// as a var, failing clearly when none is present.

// traceSpans extracts the recorded spans from metadata.
func traceSpans(metadata map[string]interface{}) ([]map[string]interface{}, bool) {
	if metadata == nil {
		return nil, false
	}
	raw, ok := metadata["_trace"].([]interface{})
	if !ok {
		return nil, false
	}
	spans := make([]map[string]interface{}, 0, len(raw))
	for _, s := range raw {
		if span, ok := s.(map[string]interface{}); ok {
			spans = append(spans, span)
		}
	}
	return spans, true
}

// spanDuration reads a span's duration in milliseconds from durationMs or
// duration.
func spanDuration(span map[string]interface{}) (float64, bool) {
	for _, key := range []string{"durationMs", "duration"} {
		if v, ok := span[key]; ok {
			if f, ok := asFloat(v); ok {
				return f, true
			}
		}
	}
	return 0, false
}

// spanHasError reports whether a span is marked as errored (error: true, or a
// non-empty status/statusCode of "error").
func spanHasError(span map[string]interface{}) bool {
	if e, ok := span["error"].(bool); ok && e {
		return true
	}
	for _, key := range []string{"status", "statusCode"} {
		if s, ok := span[key].(string); ok {
			switch s {
			case "error", "ERROR", "Error":
				return true
			}
		}
	}
	return false
}

func asFloat(v interface{}) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	}
	return 0, false
}

// evaluateTraceSpanCount checks the number of recorded spans against min/max
// bounds or an exact value (assertion.Value).
func (ae *AssertionEvaluator) evaluateTraceSpanCount(assertion Assertion, metadata map[string]interface{}) *AssertionResult {
	spans, ok := traceSpans(metadata)
	if !ok {
		return failResult(assertion, "trace-span-count requires a recorded trace (_trace var)")
	}
	count := len(spans)
	var passed bool
	var message string
	switch {
	case assertion.Min != nil && assertion.Max != nil:
		passed = count >= int(*assertion.Min) && count <= int(*assertion.Max)
		message = fmt.Sprintf("span count %d (expected between %d and %d)", count, int(*assertion.Min), int(*assertion.Max))
	case assertion.Min != nil:
		passed = count >= int(*assertion.Min)
		message = fmt.Sprintf("span count %d (expected at least %d)", count, int(*assertion.Min))
	case assertion.Max != nil:
		passed = count <= int(*assertion.Max)
		message = fmt.Sprintf("span count %d (expected at most %d)", count, int(*assertion.Max))
	default:
		exact, ok := assertionInt(assertion.Value)
		if !ok {
			return failResult(assertion, "trace-span-count requires min/max bounds or an exact integer value")
		}
		passed = count == exact
		message = fmt.Sprintf("span count %d (expected %d)", count, exact)
	}
	return &AssertionResult{Type: assertion.Type, Passed: passed, Score: boolScore(passed), Actual: count, Message: message}
}

// evaluateTraceSpanDuration passes when the maximum span duration is within the
// max bound (assertion.Max, in milliseconds; or assertion.Value as an exact
// upper bound). This catches a single slow span in the trace.
func (ae *AssertionEvaluator) evaluateTraceSpanDuration(assertion Assertion, metadata map[string]interface{}) *AssertionResult {
	spans, ok := traceSpans(metadata)
	if !ok {
		return failResult(assertion, "trace-span-duration requires a recorded trace (_trace var)")
	}
	maxDur := 0.0
	for _, s := range spans {
		if d, ok := spanDuration(s); ok && d > maxDur {
			maxDur = d
		}
	}
	bound := 0.0
	switch {
	case assertion.Max != nil:
		bound = *assertion.Max
	case assertion.Value != nil:
		if f, ok := asFloat(assertion.Value); ok {
			bound = f
		} else {
			return failResult(assertion, "trace-span-duration requires a max bound (max or numeric value)")
		}
	default:
		return failResult(assertion, "trace-span-duration requires a max bound (max or numeric value)")
	}
	passed := maxDur <= bound
	return &AssertionResult{
		Type:    assertion.Type,
		Passed:  passed,
		Score:   boolScore(passed),
		Actual:  maxDur,
		Message: fmt.Sprintf("max span duration %.0fms (expected <= %.0fms)", maxDur, bound),
	}
}

// evaluateTraceErrorSpans passes when the number of errored spans is within the
// allowed maximum (assertion.Max, or assertion.Value as an exact maximum;
// default 0 — no errored spans permitted).
func (ae *AssertionEvaluator) evaluateTraceErrorSpans(assertion Assertion, metadata map[string]interface{}) *AssertionResult {
	spans, ok := traceSpans(metadata)
	if !ok {
		return failResult(assertion, "trace-error-spans requires a recorded trace (_trace var)")
	}
	errors := 0
	for _, s := range spans {
		if spanHasError(s) {
			errors++
		}
	}
	maxErrors := 0
	if assertion.Max != nil {
		maxErrors = int(*assertion.Max)
	} else if n, ok := assertionInt(assertion.Value); ok {
		maxErrors = n
	}
	passed := errors <= maxErrors
	return &AssertionResult{
		Type:    assertion.Type,
		Passed:  passed,
		Score:   boolScore(passed),
		Actual:  errors,
		Message: fmt.Sprintf("%d errored span(s) (expected at most %d)", errors, maxErrors),
	}
}
