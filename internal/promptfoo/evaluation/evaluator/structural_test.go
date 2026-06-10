package evaluator

import (
	"context"
	"testing"
)

func TestEvaluateStartsWith(t *testing.T) {
	ae := NewAssertionEvaluator(nil)
	tests := []struct {
		name   string
		output string
		value  string
		want   bool
	}{
		{"prefix-match", "Hello, world", "Hello", true},
		{"case-sensitive-miss", "hello, world", "Hello", false},
		{"no-match", "Goodbye", "Hello", false},
		{"exact", "Hello", "Hello", true},
		{"empty-prefix", "anything", "", true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			a := Assertion{Type: AssertionStartsWith, Value: tc.value}
			got, err := ae.EvaluateAssertion(context.Background(), a, tc.output, nil)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Passed != tc.want {
				t.Errorf("Passed = %v, want %v (%s)", got.Passed, tc.want, got.Message)
			}
		})
	}
}

func TestEvaluateIsRefusal(t *testing.T) {
	ae := NewAssertionEvaluator(nil)
	tests := []struct {
		name   string
		output string
		want   bool
	}{
		{"empty", "", true},
		{"whitespace", "   \n\t ", true},
		{"prefix-sorry", "Sorry, I can't do that.", true},
		{"prefix-i-cannot", "I cannot help with that request.", true},
		{"substring-as-an-ai", "As an AI, I won't be able to assist.", true},
		{"curly-apostrophe", "I’m sorry, but no.", true},
		{"word-boundary-false-positive", "As an airbnb host I love guests.", false},
		{"normal-answer", "The capital of France is Paris.", false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			a := Assertion{Type: AssertionIsRefusal}
			got, err := ae.EvaluateAssertion(context.Background(), a, tc.output, nil)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Passed != tc.want {
				t.Errorf("Passed = %v, want %v (%s)", got.Passed, tc.want, got.Message)
			}
		})
	}
}

func TestContainsHTMLScore(t *testing.T) {
	tests := []struct {
		name    string
		output  string
		wantGTE int // expected minimum score; pass threshold is 2
		pass    bool
	}{
		{"plain-text", "just some words here", 0, false},
		{"single-uncommon-tag", "a <foo> in text", 1, false}, // open tag, not a common element = 1
		{"common-tag-strong", "a <br> in text", 3, true},      // <br> is a common tag (2) + open (1)
		{"div-pair", "<div>hello</div>", 2, true},             // common tag (2) + open/close (2)
		{"doctype", "<!DOCTYPE html><p>hi</p>", 2, true},
		{"angle-math", "if a < b and c > d then", 0, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			score := containsHTMLScore(tc.output)
			if score < tc.wantGTE {
				t.Errorf("score = %d, want >= %d", score, tc.wantGTE)
			}
			if (score >= 2) != tc.pass {
				t.Errorf("pass = %v, want %v (score %d)", score >= 2, tc.pass, score)
			}
		})
	}
}

func TestEvaluateContainsHTML(t *testing.T) {
	ae := NewAssertionEvaluator(nil)
	cases := map[string]bool{
		"<div class=\"x\">hi</div>": true,
		"no markup at all":          false,
	}
	for output, want := range cases {
		a := Assertion{Type: AssertionContainsHTML}
		got, err := ae.EvaluateAssertion(context.Background(), a, output, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Passed != want {
			t.Errorf("contains-html(%q) = %v, want %v", output, got.Passed, want)
		}
	}
}

func TestEvaluateIsXML(t *testing.T) {
	ae := NewAssertionEvaluator(nil)
	tests := []struct {
		name   string
		output string
		want   bool
	}{
		{"well-formed", `<note><to>A</to><from>B</from></note>`, true},
		{"with-decl", `<?xml version="1.0"?><root><a/></root>`, true},
		{"unclosed", `<note><to>A</note>`, false},
		{"not-xml", `just text`, false},
		{"empty", ``, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			a := Assertion{Type: AssertionIsXML}
			got, err := ae.EvaluateAssertion(context.Background(), a, tc.output, nil)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Passed != tc.want {
				t.Errorf("is-xml(%q) = %v, want %v (%s)", tc.output, got.Passed, tc.want, got.Message)
			}
		})
	}
}

func TestEvaluateContainsXML(t *testing.T) {
	ae := NewAssertionEvaluator(nil)
	tests := []struct {
		name   string
		output string
		want   bool
	}{
		{"embedded", "Here is the result:\n<data><x>1</x></data>\nDone.", true},
		{"none", "no xml present", false},
		{"malformed", "<data><x>1</data>", false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			a := Assertion{Type: AssertionContainsXML}
			got, err := ae.EvaluateAssertion(context.Background(), a, tc.output, nil)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Passed != tc.want {
				t.Errorf("contains-xml(%q) = %v, want %v (%s)", tc.output, got.Passed, tc.want, got.Message)
			}
		})
	}
}

func TestEvaluateContainsSQL(t *testing.T) {
	ae := NewAssertionEvaluator(nil)
	tests := []struct {
		name   string
		output string
		want   bool
	}{
		{"fenced", "Here:\n```sql\nSELECT * FROM users WHERE id = 1;\n```\n", true},
		{"inline-backtick", "Run `SELECT name FROM t` to fetch.", true},
		{"bare", "SELECT 1", true},
		{"prose", "I cannot answer that question.", false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			a := Assertion{Type: AssertionContainsSQL}
			got, err := ae.EvaluateAssertion(context.Background(), a, tc.output, nil)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Passed != tc.want {
				t.Errorf("contains-sql(%q) = %v, want %v (%s)", tc.output, got.Passed, tc.want, got.Message)
			}
		})
	}
}

func TestEvaluateGLEU(t *testing.T) {
	ae := NewAssertionEvaluator(nil)
	tests := []struct {
		name      string
		output    string
		value     string
		wantScore float64
		wantPass  bool
	}{
		{"identical", "the cat sat on the mat", "the cat sat on the mat", 1.0, true},
		{"no-overlap", "completely different words", "the cat sat on the mat", 0.0, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			a := Assertion{Type: AssertionGLEU, Value: tc.value}
			got, err := ae.EvaluateAssertion(context.Background(), a, tc.output, nil)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !almostEqual(got.Score, tc.wantScore) {
				t.Errorf("Score = %v, want %v", got.Score, tc.wantScore)
			}
			if got.Passed != tc.wantPass {
				t.Errorf("Passed = %v, want %v", got.Passed, tc.wantPass)
			}
		})
	}
	// gleu is min(precision, recall): a shorter exact-prefix candidate has
	// precision 1 but recall < 1, so the score is the recall.
	if s := gleu("the cat", "the cat sat", 1, 1); !almostEqual(s, 2.0/3.0) {
		t.Errorf("gleu unigram min = %v, want %v", s, 2.0/3.0)
	}
}

func TestStructuralAliases(t *testing.T) {
	ids := []struct {
		id   string
		want AssertionType
	}{
		{"starts-with", AssertionStartsWith},
		{"is-refusal", AssertionIsRefusal},
		{"is-html", AssertionIsHTML},
		{"contains-html", AssertionContainsHTML},
		{"is-xml", AssertionIsXML},
		{"contains-xml", AssertionContainsXML},
		{"contains-sql", AssertionContainsSQL},
		{"gleu", AssertionGLEU},
		{"is-valid-function-call", AssertionIsValidFunctionCall},
	}
	for _, tc := range ids {
		canonical, _, ok := normalizeAssertionType(tc.id)
		if !ok {
			t.Errorf("%q did not resolve through the alias map", tc.id)
			continue
		}
		if canonical != tc.want {
			t.Errorf("%q resolved to %q, want %q", tc.id, canonical, tc.want)
		}
	}
	// not- inversion should compose with the new ids.
	if _, negate, ok := normalizeAssertionType("not-is-refusal"); !ok || !negate {
		t.Errorf("not-is-refusal: ok=%v negate=%v, want true,true", ok, negate)
	}
}
