package evaluator

import (
	"context"
	"testing"

	"github.com/tmc/pe/internal/promptfoo"
)

func TestContainsAnyAllAssertions(t *testing.T) {
	tests := []struct {
		name   string
		assert promptfoo.Assertion
		output string
		want   bool
	}{
		{"contains-any hit", promptfoo.Assertion{Type: "contains-any", Value: []interface{}{"cat", "dog"}}, "I have a dog", true},
		{"contains-any miss", promptfoo.Assertion{Type: "contains-any", Value: []interface{}{"cat", "fish"}}, "I have a dog", false},
		{"contains-all hit", promptfoo.Assertion{Type: "contains-all", Value: []interface{}{"dog", "cat"}}, "a dog and a cat", true},
		{"contains-all miss", promptfoo.Assertion{Type: "contains-all", Value: []interface{}{"dog", "bird"}}, "a dog and a cat", false},
		{"icontains-any case-insensitive", promptfoo.Assertion{Type: "icontains-any", Value: []interface{}{"DOG"}}, "a dog", true},
		{"icontains-all case-insensitive", promptfoo.Assertion{Type: "icontains-all", Value: []interface{}{"DOG", "Cat"}}, "a dog and a cat", true},
		{"not-contains-any pass", promptfoo.Assertion{Type: "not-contains-any", Value: []interface{}{"x", "y"}}, "a dog", true},
		{"not-contains-any fail", promptfoo.Assertion{Type: "not-contains-any", Value: []interface{}{"dog"}}, "a dog", false},
		{"contains-all empty value is false", promptfoo.Assertion{Type: "contains-all", Value: []interface{}{}}, "anything", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pass, _ := evaluateAssertions(tt.output, []promptfoo.Assertion{tt.assert})
			if pass != tt.want {
				t.Fatalf("%s: pass = %v, want %v", tt.name, pass, tt.want)
			}
		})
	}
}

// TestContainsListRichPath confirms the contains-* ids now resolve through the
// alias map and run via the rich EvaluateAssertion evaluator (producing scored
// AssertionResults), rather than only via the checkAssertion string fallback.
func TestContainsListRichPath(t *testing.T) {
	ae := NewAssertionEvaluator(nil)
	tests := []struct {
		name   string
		typ    AssertionType
		output string
		value  interface{}
		want   bool
	}{
		{"any-hit", AssertionContainsAny, "the quick brown fox", []interface{}{"cat", "fox"}, true},
		{"all-miss", AssertionContainsAll, "the quick brown fox", []interface{}{"quick", "dog"}, false},
		{"icontains-hit", AssertionIContains, "The Quick Brown Fox", "quick", true},
		{"iall-hit", AssertionIContainsAll, "The Quick Brown Fox", []interface{}{"QUICK", "fox"}, true},
		{"all-case-sensitive-miss", AssertionContainsAll, "The Quick Brown Fox", []interface{}{"quick", "fox"}, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			a := Assertion{Type: tc.typ, Value: tc.value}
			got, err := ae.EvaluateAssertion(context.Background(), a, tc.output, nil)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Passed != tc.want {
				t.Errorf("%s: Passed = %v, want %v (%s)", tc.typ, got.Passed, tc.want, got.Message)
			}
			if got.Message == "" {
				t.Errorf("%s: expected a non-empty rich message", tc.typ)
			}
		})
	}
}

func TestContainsListAliases(t *testing.T) {
	ids := []struct {
		id   string
		want AssertionType
	}{
		{"contains-any", AssertionContainsAny},
		{"contains-all", AssertionContainsAll},
		{"icontains", AssertionIContains},
		{"icontains-any", AssertionIContainsAny},
		{"icontains-all", AssertionIContainsAll},
	}
	for _, tc := range ids {
		canonical, negate, ok := normalizeAssertionType(tc.id)
		if !ok {
			t.Errorf("%q did not resolve through the alias map", tc.id)
			continue
		}
		if negate {
			t.Errorf("%q unexpectedly negated", tc.id)
		}
		if canonical != tc.want {
			t.Errorf("%q resolved to %q, want %q", tc.id, canonical, tc.want)
		}
	}
	// not-contains-any strips to contains-any with negate=true.
	if canonical, negate, ok := normalizeAssertionType("not-contains-any"); !ok || !negate || canonical != AssertionContainsAny {
		t.Errorf("not-contains-any => (%q, negate=%v, ok=%v), want (contains-any, true, true)", canonical, negate, ok)
	}
}

func TestToStringSlice(t *testing.T) {
	cases := []struct {
		in   interface{}
		want []string
	}{
		{[]interface{}{"a", 1, true}, []string{"a", "1", "true"}},
		{[]string{"x", "y"}, []string{"x", "y"}},
		{"solo", []string{"solo"}},
		{nil, nil},
	}
	for _, c := range cases {
		got := toStringSlice(c.in)
		if len(got) != len(c.want) {
			t.Fatalf("toStringSlice(%v) = %v, want %v", c.in, got, c.want)
		}
		for i := range got {
			if got[i] != c.want[i] {
				t.Fatalf("toStringSlice(%v)[%d] = %q, want %q", c.in, i, got[i], c.want[i])
			}
		}
	}
}
