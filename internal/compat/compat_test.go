package compat

import "testing"

func TestLayerResolve(t *testing.T) {
	layer := NewLayer(Alias{Old: "eval-prompt", New: "eval"})
	if got := layer.Resolve("eval-prompt"); got != "eval" {
		t.Fatalf("Resolve = %q, want eval", got)
	}
	if got := layer.Resolve("run"); got != "run" {
		t.Fatalf("Resolve passthrough = %q, want run", got)
	}
}

func TestDeprecatedWarning(t *testing.T) {
	if got := DeprecatedWarning("old", "new"); got != "old is deprecated; use new" {
		t.Fatalf("DeprecatedWarning = %q", got)
	}
}

func TestPlan(t *testing.T) {
	steps := Plan([]Alias{{Old: "old", New: "new"}})
	if len(steps) != 1 || steps[0].From != "old" || steps[0].To != "new" {
		t.Fatalf("Plan = %+v", steps)
	}
}
