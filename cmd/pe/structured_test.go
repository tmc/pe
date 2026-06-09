package main

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestOutputStructuredDeniedByWritePolicy(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	if err := os.WriteFile("pe.mod", []byte(structuredPolicyTestModule()), 0644); err != nil {
		t.Fatal(err)
	}

	cmd := &cobra.Command{}
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := outputStructured(cmd, "schema.json", `{"type":"object"}`)
	if err == nil || !strings.Contains(err.Error(), "tool write is denied") {
		t.Fatalf("structured output error = %v, want write policy denial", err)
	}
	if _, err := os.Stat("schema.json"); !os.IsNotExist(err) {
		t.Fatalf("structured wrote output despite write policy: %v", err)
	}
}

func TestOutputStructuredStdoutAllowedByWritePolicy(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	if err := os.WriteFile("pe.mod", []byte(structuredPolicyTestModule()), 0644); err != nil {
		t.Fatal(err)
	}

	cmd := &cobra.Command{}
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	if err := outputStructured(cmd, "", `{"type":"object"}`); err != nil {
		t.Fatalf("structured stdout: %v", err)
	}
	if !strings.Contains(buf.String(), `"type":"object"`) {
		t.Fatalf("structured stdout = %q, want schema content", buf.String())
	}
	if _, err := os.Stat("schema.json"); !os.IsNotExist(err) {
		t.Fatalf("structured created unexpected output: %v", err)
	}
}

func structuredPolicyTestModule() string {
	return `module example.com/prompts

pe 1

capability {
    tools deny write
}
`
}
