package integration

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

func TestExecutableTextWorkflow(t *testing.T) {
	pe := buildPE(t)
	dir := t.TempDir()
	prompt := filepath.Join(dir, "review.prompt")
	if err := os.WriteFile(prompt, []byte(`---
kind: pe.text.v1
inputs:
  topic:
    type: string
---
Review {{ .topic }} safely.
`), 0o644); err != nil {
		t.Fatalf("write prompt: %v", err)
	}

	out := runPE(t, pe, "run-text", prompt, "--var", "topic=release")
	if strings.TrimSpace(out) != "Review release safely." {
		t.Fatalf("output = %q", out)
	}
}

func TestExecutableTextConcurrentStress(t *testing.T) {
	pe := buildPE(t)
	dir := t.TempDir()
	prompt := filepath.Join(dir, "plain.prompt")
	if err := os.WriteFile(prompt, []byte("plain text is valid"), 0o644); err != nil {
		t.Fatalf("write prompt: %v", err)
	}

	var wg sync.WaitGroup
	errs := make(chan string, 8)
	for i := 0; i < cap(errs); i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			out := runPE(t, pe, "run-text", prompt)
			if strings.TrimSpace(out) != "plain text is valid" {
				errs <- out
			}
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Fatalf("unexpected output: %q", err)
	}
}

func buildPE(t *testing.T) string {
	t.Helper()
	bin := filepath.Join(t.TempDir(), "pe")
	cmd := exec.Command("go", "build", "-o", bin, "../../cmd/pe")
	cmd.Env = append(os.Environ(), "GOTOOLCHAIN=local")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("build pe: %v\n%s", err, out)
	}
	return bin
}

func runPE(t *testing.T, pe string, args ...string) string {
	t.Helper()
	cmd := exec.Command(pe, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("pe %v: %v\n%s", args, err, out)
	}
	return string(out)
}
