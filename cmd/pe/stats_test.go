package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestInteractiveCmdFlags(t *testing.T) {
	cmd := interactiveCmd()

	tests := []struct {
		name string
		def  string
		typ  string
	}{
		{name: "provider", def: "openai:gpt-4", typ: "string"},
		{name: "config", def: "", typ: "string"},
		{name: "temperature", def: "0.7", typ: "float64"},
	}

	for _, tt := range tests {
		flag := cmd.Flags().Lookup(tt.name)
		if flag == nil {
			t.Fatalf("flag %q not found", tt.name)
		}
		if flag.DefValue != tt.def {
			t.Errorf("%s default = %q, want %q", tt.name, flag.DefValue, tt.def)
		}
		if flag.Value.Type() != tt.typ {
			t.Errorf("%s type = %q, want %q", tt.name, flag.Value.Type(), tt.typ)
		}
	}
}

func TestInteractiveCmdRunsREPLWithFlags(t *testing.T) {
	cmd := interactiveCmd()
	var stdout, stderr bytes.Buffer
	cmd.SetIn(strings.NewReader(""))
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{
		"--provider", "mock:test",
		"--config", "config.yaml",
		"--temperature", "0.2",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}

	output := stdout.String()
	for _, want := range []string{
		"PE Interactive Mode",
		"Provider:    mock:test",
		"Temperature: 0.2",
		"Config:      config.yaml",
		"pe> ",
	} {
		if !strings.Contains(output, want) {
			t.Errorf("output missing %q:\n%s", want, output)
		}
	}
}

func TestInteractiveCmdRejectsInvalidTemperature(t *testing.T) {
	cmd := interactiveCmd()
	cmd.SetIn(strings.NewReader(""))
	cmd.SetArgs([]string{"--temperature", "3"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("Execute() error = nil, want temperature error")
	}
	if !strings.Contains(err.Error(), "temperature must be between 0.0 and 2.0") {
		t.Fatalf("Execute() error = %v", err)
	}
}
