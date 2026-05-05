package starlark

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	starlarklib "go.starlark.net/starlark"
)

func captureStdout(t *testing.T, fn func() error) (string, error) {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	runErr := fn()
	w.Close()
	os.Stdout = old
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatal(err)
	}
	return buf.String(), runErr
}

func writeStar(t *testing.T, src string) string {
	t.Helper()
	file := filepath.Join(t.TempDir(), "test.star")
	if err := os.WriteFile(file, []byte(src), 0644); err != nil {
		t.Fatal(err)
	}
	return file
}

func TestStarlarkCommandsRunListSuiteValidateDiscover(t *testing.T) {
	file := writeStar(t, `
def test_pass(response):
    return {"pass": contains(response, "ok"), "score": 0.9, "reason": "checked"}

def test_fail(response):
    return False
`)

	out, err := captureStdout(t, func() error {
		return RunStarlarkTest(file, "ok response", "")
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, `"pass": true`) || !strings.Contains(findTestFunction(loadGlobals(t, file)), "test_") {
		t.Fatalf("run output = %s", out)
	}

	out, err = captureStdout(t, func() error { return ListTestFunctions(file) })
	if err != nil || !strings.Contains(out, "test_pass") || !strings.Contains(out, "test_fail") {
		t.Fatalf("list output = %q err=%v", out, err)
	}

	out, err = captureStdout(t, func() error { return RunStarlarkTestSuite(file, "ok response") })
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, `"total_tests": 2`) || !strings.Contains(out, `"failed_tests": 1`) {
		t.Fatalf("suite output = %s", out)
	}

	out, err = captureStdout(t, func() error { return ValidateStarlarkFile(file) })
	if err != nil || !strings.Contains(out, "is valid") {
		t.Fatalf("validate output = %q err=%v", out, err)
	}

	out, err = captureStdout(t, func() error { return StarlarkDiscoverCommand([]string{filepath.Dir(file)}) })
	if err != nil || !strings.Contains(out, "test.star") {
		t.Fatalf("discover output = %q err=%v", out, err)
	}
}

func TestStarlarkCommandErrors(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "missing.star")
	for _, fn := range []func() error{
		func() error { return RunStarlarkTest(missing, "x", "") },
		func() error { return ListTestFunctions(missing) },
		func() error { return RunStarlarkTestSuite(missing, "x") },
		func() error { return ValidateStarlarkFile(missing) },
		func() error { return StarlarkEvalCommand(nil) },
		func() error { return StarlarkListCommand(nil) },
		func() error { return StarlarkSuiteCommand(nil) },
		func() error { return StarlarkValidateCommand(nil) },
		func() error { return StarlarkDiscoverCommand([]string{missing}) },
	} {
		if err := fn(); err == nil {
			t.Fatal("expected error")
		}
	}

	noTests := writeStar(t, `x = 1`)
	out, err := captureStdout(t, func() error { return ListTestFunctions(noTests) })
	if err != nil || !strings.Contains(out, "No test functions") {
		t.Fatalf("no test output = %q err=%v", out, err)
	}
	if err := RunStarlarkTest(noTests, "x", ""); err == nil {
		t.Fatal("run without test function succeeded")
	}
	if err := RunStarlarkTestSuite(noTests, "x"); err == nil {
		t.Fatal("suite without test function succeeded")
	}

	bad := writeStar(t, `def test_bad(`)
	if err := ValidateStarlarkFile(bad); err == nil {
		t.Fatal("bad file validated")
	}
	if got := findTestFunction(starlarklib.StringDict{"x": starlarklib.String("y")}); got != "" {
		t.Fatalf("findTestFunction = %q", got)
	}
}

func TestStarlarkCommandWrappers(t *testing.T) {
	file := writeStar(t, `def test_ok(response): return True`)
	if err := StarlarkEvalCommand([]string{file, "response", "test_ok"}); err != nil {
		t.Fatal(err)
	}
	if err := StarlarkListCommand([]string{file}); err != nil {
		t.Fatal(err)
	}
	if err := StarlarkSuiteCommand([]string{file, "response"}); err != nil {
		t.Fatal(err)
	}
	if err := StarlarkValidateCommand([]string{file}); err != nil {
		t.Fatal(err)
	}
}

func loadGlobals(t *testing.T, file string) starlarklib.StringDict {
	t.Helper()
	content, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	globals, err := NewEvaluator().EvalFile(file, content)
	if err != nil {
		t.Fatal(err)
	}
	return globals
}
