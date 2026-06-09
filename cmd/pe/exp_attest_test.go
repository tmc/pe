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

func TestSignedManifestEnvelope(t *testing.T) {
	dir := t.TempDir()
	writeAttestFile(t, filepath.Join(dir, "prompt.txt"), "hello")
	manifest, err := buildUnsignedManifest(dir, []string{"prompt.txt"})
	if err != nil {
		t.Fatal(err)
	}
	manifestPath := filepath.Join(dir, "manifest.json")
	data, err := json.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}
	writeAttestFile(t, manifestPath, string(data))

	key, err := generateAttestKeyPair()
	if err != nil {
		t.Fatal(err)
	}
	envelope, err := signUnsignedManifestFile(manifestPath, key.PrivateKey)
	if err != nil {
		t.Fatalf("signUnsignedManifestFile: %v", err)
	}
	if envelope.Type != signedManifestType {
		t.Fatalf("envelope type = %q", envelope.Type)
	}
	if err := verifySignedManifestEnvelope(envelope, key.PublicKey); err != nil {
		t.Fatalf("verifySignedManifestEnvelope: %v", err)
	}

	envelopePath := filepath.Join(dir, "signed.json")
	envelopeData, err := json.Marshal(envelope)
	if err != nil {
		t.Fatal(err)
	}
	writeAttestFile(t, envelopePath, string(envelopeData))
	if err := verifySignedManifestFile(dir, envelopePath, key.PublicKey); err != nil {
		t.Fatalf("verifySignedManifestFile: %v", err)
	}
}

func TestSignedManifestRejectsFailures(t *testing.T) {
	dir := t.TempDir()
	writeAttestFile(t, filepath.Join(dir, "prompt.txt"), "hello")
	manifest, err := buildUnsignedManifest(dir, []string{"prompt.txt"})
	if err != nil {
		t.Fatal(err)
	}
	manifestPath := filepath.Join(dir, "manifest.json")
	data, err := json.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}
	writeAttestFile(t, manifestPath, string(data))

	key, err := generateAttestKeyPair()
	if err != nil {
		t.Fatal(err)
	}
	otherKey, err := generateAttestKeyPair()
	if err != nil {
		t.Fatal(err)
	}
	envelope, err := signUnsignedManifestFile(manifestPath, key.PrivateKey)
	if err != nil {
		t.Fatalf("signUnsignedManifestFile: %v", err)
	}
	if err := verifySignedManifestEnvelope(envelope, otherKey.PublicKey); err == nil {
		t.Fatal("verifySignedManifestEnvelope accepted wrong expected key")
	}

	missingSig := *envelope
	missingSig.Signature = ""
	if err := verifySignedManifestEnvelope(&missingSig, key.PublicKey); err == nil {
		t.Fatal("verifySignedManifestEnvelope accepted missing signature")
	}

	tampered := *envelope
	tampered.Payload.Entries[0].SHA256 = strings.Repeat("0", 64)
	if err := verifySignedManifestEnvelope(&tampered, key.PublicKey); err == nil {
		t.Fatal("verifySignedManifestEnvelope accepted tampered payload")
	}

	envelopePath := filepath.Join(dir, "signed.json")
	envelopeData, err := json.Marshal(envelope)
	if err != nil {
		t.Fatal(err)
	}
	writeAttestFile(t, envelopePath, string(envelopeData))
	writeAttestFile(t, filepath.Join(dir, "prompt.txt"), "changed")
	if err := verifySignedManifestFile(dir, envelopePath, key.PublicKey); err == nil {
		t.Fatal("verifySignedManifestFile accepted changed file")
	}
}

func TestExpAttestSignedCommands(t *testing.T) {
	dir := t.TempDir()
	writeAttestFile(t, filepath.Join(dir, "prompt.txt"), "hello")

	keyCmd := newExpAttestCmd()
	var keyOut bytes.Buffer
	keyCmd.SetOut(&keyOut)
	keyCmd.SetErr(&keyOut)
	keyCmd.SetArgs([]string{"keygen"})
	if err := keyCmd.Execute(); err != nil {
		t.Fatalf("keygen: %v", err)
	}
	var key attestKeyPair
	if err := json.Unmarshal(keyOut.Bytes(), &key); err != nil {
		t.Fatalf("parse keygen: %v", err)
	}
	if key.PrivateKey == "" || key.PublicKey == "" || key.Fingerprint == "" {
		t.Fatalf("incomplete key: %+v", key)
	}

	manifestPath := filepath.Join(dir, "manifest.json")
	manifestCmd := newExpAttestCmd()
	manifestFile, err := os.Create(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	manifestCmd.SetOut(manifestFile)
	manifestCmd.SetErr(manifestFile)
	manifestCmd.SetArgs([]string{"manifest", "--root", dir, "prompt.txt"})
	if err := manifestCmd.Execute(); err != nil {
		t.Fatalf("manifest: %v", err)
	}
	if err := manifestFile.Close(); err != nil {
		t.Fatal(err)
	}

	signCmd := newExpAttestCmd()
	var signedOut bytes.Buffer
	signCmd.SetOut(&signedOut)
	signCmd.SetErr(&signedOut)
	signCmd.SetArgs([]string{"sign", "--private-key", key.PrivateKey, manifestPath})
	if err := signCmd.Execute(); err != nil {
		t.Fatalf("sign: %v", err)
	}
	signedPath := filepath.Join(dir, "signed.json")
	writeAttestFile(t, signedPath, signedOut.String())

	verifyCmd := newExpAttestCmd()
	var verifyOut bytes.Buffer
	verifyCmd.SetOut(&verifyOut)
	verifyCmd.SetErr(&verifyOut)
	verifyCmd.SetArgs([]string{"verify-signed", "--root", dir, "--public-key", key.PublicKey, signedPath})
	if err := verifyCmd.Execute(); err != nil {
		t.Fatalf("verify-signed: %v\n%s", err, verifyOut.String())
	}
	if !strings.Contains(verifyOut.String(), "signed manifest verified") {
		t.Fatalf("verify output = %q", verifyOut.String())
	}
}

func writeAttestFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}
