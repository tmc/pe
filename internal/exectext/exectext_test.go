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

func TestRenderWithImports(t *testing.T) {
	f, err := Parse(strings.NewReader(`---
kind: pe.text.v1
inputs:
  topic:
    type: string
imports:
  reviewer: reviewer.prompt
---
{{ import "reviewer" }}
`))
	if err != nil {
		t.Fatal(err)
	}
	got, err := f.RenderWithImports(map[string]string{"topic": "release"}, map[string]string{
		"reviewer": "Review {{ .topic }}.\n",
	})
	if err != nil {
		t.Fatal(err)
	}
	if got != "Review release.\n\n" {
		t.Fatalf("RenderWithImports = %q", got)
	}
}

func TestCheckComposition(t *testing.T) {
	parse := func(t *testing.T, text string) *File {
		t.Helper()
		f, err := Parse(strings.NewReader(text))
		if err != nil {
			t.Fatal(err)
		}
		return f
	}

	parent := parse(t, `---
kind: pe.text.v1
safety:
  providers:
    allow: [ollama, mock]
    deny: [remote]
  tools:
    deny: [write]
---
parent
`)

	tests := []struct {
		name    string
		child   string
		wantErr string
	}{
		{
			name:  "no child policy",
			child: "plain child\n",
		},
		{
			name: "child within parent allow",
			child: `---
safety:
  providers:
    allow: [mock]
---
child
`,
		},
		{
			name: "child tightens with more denies",
			child: `---
safety:
  providers:
    deny: [ollama]
  tools:
    deny: [write, network]
---
child
`,
		},
		{
			name: "child dimension parent does not declare",
			child: `---
safety:
  data:
    allow: [public]
---
child
`,
		},
		{
			name: "child allows parent-denied value",
			child: `---
safety:
  providers:
    allow: [remote]
---
child
`,
			wantErr: "safety providers allows remote denied by parent",
		},
		{
			name: "child allows tool parent denies",
			child: `---
safety:
  tools:
    allow: [write]
---
child
`,
			wantErr: "safety tools allows write denied by parent",
		},
		{
			name: "child allow outside parent allow list",
			child: `---
safety:
  providers:
    allow: [openai]
---
child
`,
			wantErr: "safety providers allows openai outside parent allow list",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CheckComposition(parent, parse(t, tt.child))
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("CheckComposition = %v, want nil", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("CheckComposition = %v, want error containing %q", err, tt.wantErr)
			}
		})
	}
}

func TestRenderWithImportsRejectsLoosenedPolicy(t *testing.T) {
	f, err := Parse(strings.NewReader(`---
kind: pe.text.v1
safety:
  providers:
    deny: [remote]
imports:
  child: child.prompt
---
{{ import "child" }}
`))
	if err != nil {
		t.Fatal(err)
	}
	_, err = f.RenderWithImports(nil, map[string]string{
		"child": "---\nsafety:\n  providers:\n    allow: [remote]\n---\nchild\n",
	})
	if err == nil || !strings.Contains(err.Error(), "denied by parent") {
		t.Fatalf("RenderWithImports error = %v, want composition denial", err)
	}
}
