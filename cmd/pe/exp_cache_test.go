package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestContentKeyDeterministic(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "input.txt")
	writeFile(t, path, "cache me")

	first, err := contentKeyFile(path)
	if err != nil {
		t.Fatal(err)
	}
	second, err := contentKeyFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if first != second {
		t.Fatalf("keys differ: %s != %s", first, second)
	}
	sum := sha256.Sum256([]byte("cache me"))
	if first != hex.EncodeToString(sum[:]) {
		t.Fatalf("key = %s, want raw sha256", first)
	}
}

func TestCachePutGetVerifyAndTamper(t *testing.T) {
	dir := t.TempDir()
	cacheDir := filepath.Join(dir, "cache")
	input := filepath.Join(dir, "input.txt")
	writeFile(t, input, "cache me")

	key, err := putCacheFile(cacheDir, input)
	if err != nil {
		t.Fatal(err)
	}
	if err := verifyCacheKey(cacheDir, key); err != nil {
		t.Fatalf("verify before tamper failed: %v", err)
	}
	got, err := getCacheBytes(cacheDir, key)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "cache me" {
		t.Fatalf("cached content = %q", got)
	}

	objectPath, err := cacheObjectPath(cacheDir, key)
	if err != nil {
		t.Fatal(err)
	}
	writeFile(t, objectPath, "changed")
	if err := verifyCacheKey(cacheDir, key); err == nil {
		t.Fatal("verify after tamper succeeded")
	}
}

func TestCacheRejectsUnsafeKeysAndSymlinks(t *testing.T) {
	dir := t.TempDir()
	if _, err := cacheObjectPath(dir, "../"+strings.Repeat("a", 64)); err == nil {
		t.Fatal("cacheObjectPath accepted traversal key")
	}
	if _, err := getCacheBytes(dir, strings.Repeat("a", 64)); err == nil {
		t.Fatal("getCacheBytes accepted missing entry")
	}

	target := filepath.Join(dir, "target.txt")
	writeFile(t, target, "ok")
	link := filepath.Join(dir, "link.txt")
	if err := os.Symlink(target, link); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	if _, err := contentKeyFile(link); err == nil {
		t.Fatal("contentKeyFile accepted symlink")
	}
}

func TestManifestDigestCanonical(t *testing.T) {
	dir := t.TempDir()
	manifest := unsignedFileManifest{
		Type:      unsignedManifestType,
		Algorithm: "sha256",
		Entries: []unsignedManifestEntry{
			{Path: "a.txt", Size: 1, SHA256: strings.Repeat("a", 64)},
		},
	}
	pretty, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	compact, err := json.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}
	prettyPath := filepath.Join(dir, "pretty.json")
	compactPath := filepath.Join(dir, "compact.json")
	writeFile(t, prettyPath, string(pretty))
	writeFile(t, compactPath, string(compact))

	prettyKey, err := unsignedManifestDigestFile(prettyPath)
	if err != nil {
		t.Fatal(err)
	}
	compactKey, err := unsignedManifestDigestFile(compactPath)
	if err != nil {
		t.Fatal(err)
	}
	if prettyKey != compactKey {
		t.Fatalf("manifest keys differ: %s != %s", prettyKey, compactKey)
	}
}

func TestExpCacheCommands(t *testing.T) {
	dir := t.TempDir()
	cacheDir := filepath.Join(dir, "cache")
	input := filepath.Join(dir, "input.txt")
	writeFile(t, input, "cache me")

	putCmd := newExpCacheCmd()
	var putOut bytes.Buffer
	putCmd.SetOut(&putOut)
	putCmd.SetErr(&putOut)
	putCmd.SetArgs([]string{"put", "--cache-dir", cacheDir, input})
	if err := putCmd.Execute(); err != nil {
		t.Fatal(err)
	}
	key := strings.TrimSpace(putOut.String())

	verifyCmd := newExpCacheCmd()
	var verifyOut bytes.Buffer
	verifyCmd.SetOut(&verifyOut)
	verifyCmd.SetErr(&verifyOut)
	verifyCmd.SetArgs([]string{"verify", "--cache-dir", cacheDir, key})
	if err := verifyCmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(verifyOut.String(), "cache entry verified") {
		t.Fatalf("verify output = %q", verifyOut.String())
	}
}
