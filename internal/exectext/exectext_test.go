package exectext

import (
	"strings"
	"testing"
)

func TestParsePlainText(t *testing.T) {
	f, err := Parse(strings.NewReader("hello\n"))
	if err != nil {
		t.Fatal(err)
	}
	if f.Body != "hello\n" {
		t.Fatalf("Body = %q", f.Body)
	}
}

func TestRenderDeclaredInputs(t *testing.T) {
	f, err := Parse(strings.NewReader(`#!/usr/bin/env pe run-text
---
kind: pe.text.v1
run: pe run-text
inputs:
  topic:
    type: string
---
Review {{ .topic }}.
`))
	if err != nil {
		t.Fatal(err)
	}
	if err := f.Validate(); err != nil {
		t.Fatal(err)
	}
	got, err := f.Render(map[string]string{"topic": "release"})
	if err != nil {
		t.Fatal(err)
	}
	if got != "Review release.\n" {
		t.Fatalf("Render = %q", got)
	}
}

func TestRenderMissingInput(t *testing.T) {
	f, err := Parse(strings.NewReader(`---
kind: pe.text.v1
inputs:
  topic:
    type: string
---
{{ .topic }}
`))
	if err != nil {
		t.Fatal(err)
	}
	_, err = f.Render(nil)
	if err == nil || !strings.Contains(err.Error(), "missing input topic") {
		t.Fatalf("Render error = %v", err)
	}
}
