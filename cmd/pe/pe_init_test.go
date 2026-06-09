package main

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestPEInitDeniedByWritePolicy(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	if err := os.WriteFile("pe.mod", []byte(peInitPolicyTestModule()), 0644); err != nil {
		t.Fatal(err)
	}

	cmd := peInitCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs(nil)

	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "tool write is denied") {
		t.Fatalf("pe init error = %v, want write policy denial", err)
	}
	if _, err := os.Stat(".pe"); !os.IsNotExist(err) {
		t.Fatalf("pe init created .pe despite write policy: %v", err)
	}
	if _, err := os.Stat(".peignore"); !os.IsNotExist(err) {
		t.Fatalf("pe init created .peignore despite write policy: %v", err)
	}
}

func TestPEInitCreatesProjectFiles(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	cmd := peInitCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs(nil)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("pe init: %v", err)
	}
	if _, err := os.Stat(".pe/config/pe.yaml"); err != nil {
		t.Fatalf("pe init config stat: %v", err)
	}
	if _, err := os.Stat(".peignore"); err != nil {
		t.Fatalf("pe init ignore stat: %v", err)
	}
}

func peInitPolicyTestModule() string {
	return `module example.com/prompts

pe 1

capability {
    tools deny write
}
`
}
