package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExpDistributedCmdPreservesOrder(t *testing.T) {
	taskFile := writeDistributedTaskFile(t, `{"tasks":[
		{"name":"slow","output":"first","delay_ms":20},
		{"name":"fast","output":"second"}
	]}`)

	cmd := newExpDistributedCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"--workers", "2", taskFile})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("distributed Execute() error = %v", err)
	}
	var got []distributedTaskResult
	if err := json.Unmarshal(out.Bytes(), &got); err != nil {
		t.Fatalf("distributed output did not decode: %v\n%s", err, out.String())
	}
	if len(got) != 2 {
		t.Fatalf("got %d results, want 2", len(got))
	}
	if got[0].Name != "slow" || got[0].Output != "first" || got[1].Name != "fast" || got[1].Output != "second" {
		t.Fatalf("results = %+v, want input order", got)
	}
}

func TestExpDistributedCmdTextOutput(t *testing.T) {
	taskFile := writeDistributedTaskFile(t, `{"tasks":[{"name":"one","output":"ok"}]}`)

	cmd := newExpDistributedCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"--format", "text", taskFile})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("distributed Execute() error = %v", err)
	}
	if got, want := out.String(), "one\tok\n"; got != want {
		t.Fatalf("text output = %q, want %q", got, want)
	}
}

func TestExpDistributedCmdValidatesWorkers(t *testing.T) {
	taskFile := writeDistributedTaskFile(t, `{"tasks":[{"name":"one","output":"ok"}]}`)

	cmd := newExpDistributedCmd()
	cmd.SetArgs([]string{"--workers", "0", taskFile})

	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "workers must be at least 1") {
		t.Fatalf("distributed error = %v, want worker validation", err)
	}
}

func TestExpDistributedCmdReturnsTaskError(t *testing.T) {
	taskFile := writeDistributedTaskFile(t, `{"tasks":[
		{"name":"ok","output":"done"},
		{"name":"bad","error":"boom"}
	]}`)

	cmd := newExpDistributedCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"--workers", "1", taskFile})

	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "boom") {
		t.Fatalf("distributed error = %v, want task error", err)
	}
	var got []distributedTaskResult
	if err := json.Unmarshal(out.Bytes(), &got); err != nil {
		t.Fatalf("distributed output did not decode: %v\n%s", err, out.String())
	}
	if len(got) != 2 || got[1].Name != "bad" || got[1].Error != "boom" {
		t.Fatalf("results = %+v, want indexed task error", got)
	}
}

func TestExpDistributedCmdTimeoutCancels(t *testing.T) {
	taskFile := writeDistributedTaskFile(t, `{"tasks":[{"name":"slow","output":"late","delay_ms":50}]}`)

	cmd := newExpDistributedCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"--timeout-ms", "1", taskFile})

	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "context deadline exceeded") {
		t.Fatalf("distributed error = %v, want context deadline exceeded", err)
	}
	var got []distributedTaskResult
	if err := json.Unmarshal(out.Bytes(), &got); err != nil {
		t.Fatalf("distributed output did not decode: %v\n%s", err, out.String())
	}
	if len(got) != 1 || got[0].Name != "slow" || !strings.Contains(got[0].Error, "context deadline exceeded") {
		t.Fatalf("results = %+v, want cancellation result", got)
	}
}

func TestExpDistributedCmdHelp(t *testing.T) {
	cmd := newExpDistributedCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"--help"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("distributed help error = %v", err)
	}
	for _, want := range []string{
		"Run local deterministic tasks",
		"--workers",
		"--timeout-ms",
		"local scheduler prototype",
	} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("help output missing %q:\n%s", want, out.String())
		}
	}
}

func writeDistributedTaskFile(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "tasks.json")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write task file: %v", err)
	}
	return path
}
