package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"os/exec"
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

func TestCacheManifestWorkflowDetectsTamper(t *testing.T) {
	dir := t.TempDir()
	cacheDir := filepath.Join(dir, "cache")
	root := filepath.Join(dir, "root")
	if err := os.Mkdir(root, 0755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(root, "prompt.txt"), "hello")

	manifest, err := buildUnsignedManifest(root, []string{"prompt.txt"})
	if err != nil {
		t.Fatal(err)
	}
	manifestData, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	manifestPath := filepath.Join(dir, "manifest.json")
	writeFile(t, manifestPath, string(manifestData))

	key, err := putManifestCacheFile(cacheDir, manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := verifyCachedManifest(cacheDir, root, key); err != nil {
		t.Fatalf("verify cached manifest before tamper failed: %v", err)
	}

	writeFile(t, filepath.Join(root, "prompt.txt"), "changed")
	if err := verifyCachedManifest(cacheDir, root, key); err == nil {
		t.Fatal("verify cached manifest accepted changed content")
	}
	writeFile(t, filepath.Join(root, "prompt.txt"), "hello")

	objectPath, err := cacheObjectPath(cacheDir, key)
	if err != nil {
		t.Fatal(err)
	}
	writeFile(t, objectPath, strings.Replace(string(manifestData), "prompt.txt", "other.txt", 1))
	if err := verifyCachedManifest(cacheDir, root, key); err == nil {
		t.Fatal("verify cached manifest accepted tampered manifest object")
	}
}

func TestCacheManifestRejectsUnsafePaths(t *testing.T) {
	dir := t.TempDir()
	cacheDir := filepath.Join(dir, "cache")
	root := filepath.Join(dir, "root")
	if err := os.Mkdir(root, 0755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(root, "prompt.txt"), "hello")

	manifest := unsignedFileManifest{
		Type:      unsignedManifestType,
		Algorithm: "sha256",
		Entries: []unsignedManifestEntry{
			{Path: "../prompt.txt", Size: 5, SHA256: strings.Repeat("a", 64)},
		},
	}
	manifestData, err := json.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}
	manifestPath := filepath.Join(dir, "manifest.json")
	writeFile(t, manifestPath, string(manifestData))
	key, err := putManifestCacheFile(cacheDir, manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := verifyCachedManifest(cacheDir, root, key); err == nil {
		t.Fatal("verify cached manifest accepted escaping path")
	}

	link := filepath.Join(root, "link.txt")
	if err := os.Symlink(filepath.Join(root, "prompt.txt"), link); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	linkManifest, err := buildUnsignedManifest(root, []string{"prompt.txt"})
	if err != nil {
		t.Fatal(err)
	}
	linkManifest.Entries[0].Path = "link.txt"
	linkData, err := json.Marshal(linkManifest)
	if err != nil {
		t.Fatal(err)
	}
	linkManifestPath := filepath.Join(dir, "link-manifest.json")
	writeFile(t, linkManifestPath, string(linkData))
	linkKey, err := putManifestCacheFile(cacheDir, linkManifestPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := verifyCachedManifest(cacheDir, root, linkKey); err == nil {
		t.Fatal("verify cached manifest accepted symlink path")
	}
}

func TestExpCacheManifestCommands(t *testing.T) {
	dir := t.TempDir()
	cacheDir := filepath.Join(dir, "cache")
	root := filepath.Join(dir, "root")
	if err := os.Mkdir(root, 0755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(root, "prompt.txt"), "hello")

	attestCmd := newExpAttestCmd()
	var manifestOut bytes.Buffer
	attestCmd.SetOut(&manifestOut)
	attestCmd.SetErr(&manifestOut)
	attestCmd.SetArgs([]string{"manifest", "--root", root, "prompt.txt"})
	if err := attestCmd.Execute(); err != nil {
		t.Fatal(err)
	}
	manifestPath := filepath.Join(dir, "manifest.json")
	writeFile(t, manifestPath, manifestOut.String())

	putCmd := newExpCacheCmd()
	var putOut bytes.Buffer
	putCmd.SetOut(&putOut)
	putCmd.SetErr(&putOut)
	putCmd.SetArgs([]string{"manifest", "put", "--cache-dir", cacheDir, manifestPath})
	if err := putCmd.Execute(); err != nil {
		t.Fatal(err)
	}
	key := strings.TrimSpace(putOut.String())

	verifyCmd := newExpCacheCmd()
	var verifyOut bytes.Buffer
	verifyCmd.SetOut(&verifyOut)
	verifyCmd.SetErr(&verifyOut)
	verifyCmd.SetArgs([]string{"manifest", "verify", "--cache-dir", cacheDir, "--root", root, key})
	if err := verifyCmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(verifyOut.String(), "cached manifest verified") {
		t.Fatalf("verify output = %q", verifyOut.String())
	}
}

func TestExpCacheManifestCLIWorkflow(t *testing.T) {
	dir := t.TempDir()
	pe := filepath.Join(dir, "pe")
	build := exec.Command("go", "build", "-o", pe, ".")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build pe: %v\n%s", err, out)
	}

	cacheDir := filepath.Join(dir, "cache")
	root := filepath.Join(dir, "root")
	if err := os.Mkdir(root, 0755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(root, "prompt.txt"), "hello")

	manifest := runPE(t, pe, "exp", "attest", "manifest", "--root", root, "prompt.txt")
	manifestPath := filepath.Join(dir, "manifest.json")
	writeFile(t, manifestPath, manifest)

	key := strings.TrimSpace(runPE(t, pe, "exp", "cache", "manifest", "put", "--cache-dir", cacheDir, manifestPath))
	verify := runPE(t, pe, "exp", "cache", "manifest", "verify", "--cache-dir", cacheDir, "--root", root, key)
	if !strings.Contains(verify, "cached manifest verified") {
		t.Fatalf("verify output = %q", verify)
	}

	writeFile(t, filepath.Join(root, "prompt.txt"), "changed")
	cmd := exec.Command(pe, "exp", "cache", "manifest", "verify", "--cache-dir", cacheDir, "--root", root, key)
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("tampered content verified:\n%s", out)
	}
	if !strings.Contains(string(out), "size mismatch for prompt.txt") {
		t.Fatalf("tamper output = %q", out)
	}

	help := runPE(t, pe, "exp", "cache", "manifest", "--help")
	for _, want := range []string{"unsigned", "local-only", "detects tamper", "does not prove identity", "origin", "freshness"} {
		if !strings.Contains(help, want) {
			t.Fatalf("help missing %q:\n%s", want, help)
		}
	}
}

func runPE(t *testing.T, pe string, args ...string) string {
	t.Helper()
	cmd := exec.Command(pe, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("pe %s: %v\n%s", strings.Join(args, " "), err, out)
	}
	return string(out)
}
