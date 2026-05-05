package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestUnsignedManifestDeterministic(t *testing.T) {
	dir := t.TempDir()
	writeAttestFile(t, filepath.Join(dir, "b.txt"), "bravo")
	writeAttestFile(t, filepath.Join(dir, "a.txt"), "alpha")

	first, err := buildUnsignedManifest(dir, []string{"b.txt", "a.txt"})
	if err != nil {
		t.Fatal(err)
	}
	second, err := buildUnsignedManifest(dir, []string{"a.txt", "b.txt"})
	if err != nil {
		t.Fatal(err)
	}

	firstJSON, _ := json.Marshal(first)
	secondJSON, _ := json.Marshal(second)
	if string(firstJSON) != string(secondJSON) {
		t.Fatalf("manifest is not deterministic\nfirst:  %s\nsecond: %s", firstJSON, secondJSON)
	}
	if len(first.Entries) != 2 {
		t.Fatalf("len(entries) = %d, want 2", len(first.Entries))
	}
	if first.Entries[0].Path != "a.txt" || first.Entries[1].Path != "b.txt" {
		t.Fatalf("entries not sorted: %+v", first.Entries)
	}
}

func TestUnsignedManifestRejectsEscapesAndSymlinks(t *testing.T) {
	dir := t.TempDir()
	writeAttestFile(t, filepath.Join(dir, "ok.txt"), "ok")
	if _, err := buildUnsignedManifest(dir, []string{"../outside.txt"}); err == nil {
		t.Fatal("buildUnsignedManifest accepted escaping path")
	}

	link := filepath.Join(dir, "link.txt")
	if err := os.Symlink(filepath.Join(dir, "ok.txt"), link); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	if _, err := buildUnsignedManifest(dir, []string{"link.txt"}); err == nil {
		t.Fatal("buildUnsignedManifest accepted symlink")
	}
}

func TestUnsignedManifestVerifyDetectsTamper(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "prompt.txt")
	writeAttestFile(t, target, "original")

	manifest, err := buildUnsignedManifest(dir, []string{"prompt.txt"})
	if err != nil {
		t.Fatal(err)
	}
	manifestPath := filepath.Join(dir, "manifest.json")
	data, _ := json.Marshal(manifest)
	writeAttestFile(t, manifestPath, string(data))

	if err := verifyUnsignedManifestFile(dir, manifestPath); err != nil {
		t.Fatalf("verify before tamper failed: %v", err)
	}
	writeAttestFile(t, target, "changed")
	if err := verifyUnsignedManifestFile(dir, manifestPath); err == nil {
		t.Fatal("verify after tamper succeeded")
	}
}

func TestExpAttestManifestCommand(t *testing.T) {
	dir := t.TempDir()
	writeAttestFile(t, filepath.Join(dir, "prompt.txt"), "hello")

	cmd := newExpAttestCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"manifest", "--root", dir, "prompt.txt"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), `"type": "pe.unsigned_file_manifest.v1"`) {
		t.Fatalf("manifest output missing type: %s", out.String())
	}
	if strings.Contains(out.String(), "signature") {
		t.Fatalf("unsigned manifest output should not imply signing: %s", out.String())
	}
}

func writeAttestFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}
