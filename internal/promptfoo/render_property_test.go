package promptfoo

import (
	"strings"
	"testing"
	"testing/quick"
)

func TestApplyVarsPropertyReplacesBothForms(t *testing.T) {
	property := func(keyRaw, valueRaw string) bool {
		key := propertyIdent(keyRaw)
		value := propertyValue(valueRaw)
		prompt := "before {{" + key + "}} middle {{." + key + "}} after"
		got := ApplyVars(prompt, map[string]interface{}{key: value})
		return strings.Count(got, value) == 2 &&
			!strings.Contains(got, "{{"+key+"}}") &&
			!strings.Contains(got, "{{."+key+"}}")
	}
	if err := quick.Check(property, nil); err != nil {
		t.Fatal(err)
	}
}

func propertyIdent(s string) string {
	var b strings.Builder
	for _, r := range s {
		if ('a' <= r && r <= 'z') || ('A' <= r && r <= 'Z') || ('0' <= r && r <= '9') {
			b.WriteRune(r)
		}
	}
	if b.Len() == 0 {
		return "x"
	}
	return b.String()
}

func propertyValue(s string) string {
	s = strings.ReplaceAll(s, "{", "")
	s = strings.ReplaceAll(s, "}", "")
	if s == "" {
		return "value"
	}
	return s
}
