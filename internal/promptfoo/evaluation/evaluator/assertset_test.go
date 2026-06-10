package evaluator

import (
	"testing"

	"github.com/tmc/pe/internal/promptfoo"
)

func TestAssertSet(t *testing.T) {
	tests := []struct {
		name   string
		assert promptfoo.Assertion
		output string
		want   bool
	}{
		{
			name: "all-children-pass",
			assert: promptfoo.Assertion{Type: "assert-set", Value: []interface{}{
				map[string]interface{}{"type": "contains", "value": "cat"},
				map[string]interface{}{"type": "contains", "value": "dog"},
			}},
			output: "a cat and a dog",
			want:   true,
		},
		{
			name: "one-child-fails-no-threshold",
			assert: promptfoo.Assertion{Type: "assert-set", Value: []interface{}{
				map[string]interface{}{"type": "contains", "value": "cat"},
				map[string]interface{}{"type": "contains", "value": "bird"},
			}},
			output: "a cat and a dog",
			want:   false, // default: all must pass
		},
		{
			name: "threshold-tolerates-one-failure",
			assert: promptfoo.Assertion{Type: "assert-set", Threshold: 0.5, Value: []interface{}{
				map[string]interface{}{"type": "contains", "value": "cat"},
				map[string]interface{}{"type": "contains", "value": "bird"},
			}},
			output: "a cat and a dog",
			want:   true, // 1/2 = 0.5 >= 0.5
		},
		{
			name: "weighted-threshold",
			assert: promptfoo.Assertion{Type: "assert-set", Threshold: 0.6, Value: []interface{}{
				map[string]interface{}{"type": "contains", "value": "cat", "weight": float64(3)},
				map[string]interface{}{"type": "contains", "value": "bird", "weight": float64(1)},
			}},
			output: "a cat and a dog",
			want:   true, // weighted score 3/4 = 0.75 >= 0.6
		},
		{
			name:   "empty-set-fails",
			assert: promptfoo.Assertion{Type: "assert-set", Value: []interface{}{}},
			output: "anything",
			want:   false,
		},
		{
			name: "child-from-config",
			assert: promptfoo.Assertion{Type: "assert-set", Config: map[string]interface{}{
				"assert": []interface{}{
					map[string]interface{}{"type": "contains", "value": "cat"},
				},
			}},
			output: "a cat",
			want:   true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			pass, _ := evaluateAssertions(tc.output, []promptfoo.Assertion{tc.assert})
			if pass != tc.want {
				t.Errorf("assert-set pass = %v, want %v", pass, tc.want)
			}
		})
	}
}

func TestAssertSetNested(t *testing.T) {
	// An assert-set inside an assert-set should recurse.
	inner := map[string]interface{}{
		"type": "assert-set",
		"value": []interface{}{
			map[string]interface{}{"type": "contains", "value": "cat"},
		},
	}
	outer := promptfoo.Assertion{Type: "assert-set", Value: []interface{}{
		inner,
		map[string]interface{}{"type": "contains", "value": "dog"},
	}}
	pass, _ := evaluateAssertions("a cat and a dog", []promptfoo.Assertion{outer})
	if !pass {
		t.Errorf("nested assert-set should pass")
	}
}

func TestNormalizeAssertSetType(t *testing.T) {
	for _, id := range []string{"assert-set", "assert_set", "ASSERT-SET", " assert-set "} {
		if !normalizeAssertSetType(id) {
			t.Errorf("%q should be recognized as assert-set", id)
		}
	}
	if normalizeAssertSetType("contains") {
		t.Errorf("contains should not be assert-set")
	}
}
