package main

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestFmtCmd_WriteDeniedByPolicy(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	writeCommandsDenyWritePeMod(t)
	const original = "prompts:\n  - hello\n"
	if err := os.WriteFile("config.yaml", []byte(original), 0644); err != nil {
		t.Fatalf("writing config: %v", err)
	}

	cmd := fmtCmd()
	if err := cmd.Flags().Set("write", "true"); err != nil {
		t.Fatal(err)
	}
	err := cmd.RunE(cmd, []string{"config.yaml"})
	if err == nil {
		t.Fatal("fmt --write succeeded, want write policy error")
	}
	if !strings.Contains(err.Error(), "tool write is denied by pe.mod") {
		t.Fatalf("fmt --write error = %v, want write policy error", err)
	}
	data, err := os.ReadFile("config.yaml")
	if err != nil {
		t.Fatalf("reading config: %v", err)
	}
	if string(data) != original {
		t.Fatalf("config changed after denied fmt:\n%s", data)
	}
}

func TestFmtCmd_StdoutAllowedByPolicy(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	writeCommandsDenyWritePeMod(t)
	if err := os.WriteFile("config.yaml", []byte("prompts:\n- hello\n"), 0644); err != nil {
		t.Fatalf("writing config: %v", err)
	}

	var out bytes.Buffer
	cmd := fmtCmd()
	cmd.SetOut(&out)
	if err := cmd.RunE(cmd, []string{"config.yaml"}); err != nil {
		t.Fatalf("fmt stdout: %v", err)
	}
	if !strings.Contains(out.String(), "prompts:") {
		t.Fatalf("fmt stdout output = %q", out.String())
	}
}

func TestInitCmd_DeniedByPolicy(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	writeCommandsDenyWritePeMod(t)
	cmd := initCmd()
	err := cmd.RunE(cmd, []string{"blocked.yaml"})
	if err == nil {
		t.Fatal("init succeeded, want write policy error")
	}
	if !strings.Contains(err.Error(), "tool write is denied by pe.mod") {
		t.Fatalf("init error = %v, want write policy error", err)
	}
	if _, err := os.Stat("blocked.yaml"); !os.IsNotExist(err) {
		t.Fatalf("blocked.yaml stat error = %v, want not exist", err)
	}
}

func TestConvertCmd_DeniedByPolicy(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	writeCommandsDenyWritePeMod(t)
	if err := os.WriteFile("input.yaml", []byte("prompts:\n  - hello\n"), 0644); err != nil {
		t.Fatalf("writing input: %v", err)
	}
	cmd := convertCmd()
	err := cmd.RunE(cmd, []string{"input.yaml", "output.json"})
	if err == nil {
		t.Fatal("convert succeeded, want write policy error")
	}
	if !strings.Contains(err.Error(), "tool write is denied by pe.mod") {
		t.Fatalf("convert error = %v, want write policy error", err)
	}
	if _, err := os.Stat("output.json"); !os.IsNotExist(err) {
		t.Fatalf("output.json stat error = %v, want not exist", err)
	}
}

func writeCommandsDenyWritePeMod(t *testing.T) {
	t.Helper()
	if err := os.WriteFile("pe.mod", []byte(`module example.com/app

pe 1

capability {
    tools deny write
}
`), 0644); err != nil {
		t.Fatalf("writing pe.mod: %v", err)
	}
}
