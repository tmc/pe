package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestRunTextPlain(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "plain.prompt")
	if err := os.WriteFile(file, []byte("hello\n"), 0644); err != nil {
		t.Fatal(err)
	}
	cmd := runTextCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{file})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if out.String() != "hello\n" {
		t.Fatalf("output = %q", out.String())
	}
}

func TestRunTextTemplate(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "templated.prompt")
	data := []byte(`---
kind: pe.text.v1
inputs:
  name:
    type: string
---
hello {{ .name }}
`)
	if err := os.WriteFile(file, data, 0644); err != nil {
		t.Fatal(err)
	}
	cmd := runTextCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{file, "--var", "name=pe"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if out.String() != "hello pe\n" {
		t.Fatalf("output = %q", out.String())
	}
}

func TestRunTextImports(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "part.prompt"), []byte("part {{ .name }}\n"), 0644); err != nil {
		t.Fatal(err)
	}
	main := []byte(`---
kind: pe.text.v1
inputs:
  name:
    type: string
imports:
  part: part.prompt
---
{{ import "part" }}`)
	file := filepath.Join(dir, "main.prompt")
	if err := os.WriteFile(file, main, 0644); err != nil {
		t.Fatal(err)
	}
	cmd := runTextCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{file, "--var", "name=pe"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if out.String() != "part pe\n" {
		t.Fatalf("output = %q", out.String())
	}
}

func TestRunTextImportsRejectAbsolutePath(t *testing.T) {
	dir := t.TempDir()
	secret := filepath.Join(dir, "secret.txt")
	if err := os.WriteFile(secret, []byte("secret\n"), 0644); err != nil {
		t.Fatal(err)
	}
	main := []byte(`---
kind: pe.text.v1
imports:
  leak: ` + secret + `
---
{{ import "leak" }}`)
	file := filepath.Join(dir, "main.prompt")
	if err := os.WriteFile(file, main, 0644); err != nil {
		t.Fatal(err)
	}
	cmd := runTextCmd()
	cmd.SetArgs([]string{file})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("absolute import succeeded")
	}
}

func TestRunTextImportsRejectTraversal(t *testing.T) {
	parent := t.TempDir()
	dir := filepath.Join(parent, "prompts")
	if err := os.Mkdir(dir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(parent, "secret.txt"), []byte("secret\n"), 0644); err != nil {
		t.Fatal(err)
	}
	main := []byte(`---
kind: pe.text.v1
imports:
  leak: ../secret.txt
---
{{ import "leak" }}`)
	file := filepath.Join(dir, "main.prompt")
	if err := os.WriteFile(file, main, 0644); err != nil {
		t.Fatal(err)
	}
	cmd := runTextCmd()
	cmd.SetArgs([]string{file})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("traversal import succeeded")
	}
}
