package evaluator

import (
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
