package config

import (
	"bytes"
	"strings"
	"testing"
)

func TestWriteDocumentation(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteDocumentation(&buf); err != nil {
		t.Fatalf("WriteDocumentation: %v", err)
	}
	out := buf.String()
	for _, want := range []string{
		"# PE Configuration",
		"## `providers.openai.api_key`",
		"Environment: `OPENAI_API_KEY`",
		"## `app.log_level`",
		"Default: `info`",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("documentation missing %q\n%s", want, out)
		}
	}
}
