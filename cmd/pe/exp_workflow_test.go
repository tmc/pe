package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tmc/pe/internal/workflow"
)

func writeWorkflowFile(t *testing.T, name, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
	return path
}

func runWorkflowCmd(t *testing.T, args ...string) (string, string, error) {
	t.Helper()
	cmd := expWorkflowRunCmd()
	var out, errOut bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errOut)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return out.String(), errOut.String(), err
}

func TestExpWorkflowValidate(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantErr string
	}{
		{"star", `result = generate("hi")`, ""},
		{
			"exec text",
			"---\nkind: pe.workflow.v1\nname: t\n---\nresult = parallel([\"a\"])\n",
			"",
		},
		{"syntax error", `result = (`, "workflow script"},
		{
			"wrong kind",
			"---\nkind: pe.text.v1\n---\nresult = 1\n",
			"require kind pe.workflow.v1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := writeWorkflowFile(t, "wf.star", tt.content)
			cmd := expWorkflowValidateCmd()
			var out bytes.Buffer
			cmd.SetOut(&out)
			cmd.SetArgs([]string{path})
			err := cmd.Execute()
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("validate: %v", err)
				}
				if !strings.Contains(out.String(), "OK: ") {
					t.Errorf("output = %q, want OK", out.String())
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("err = %v, want containing %q", err, tt.wantErr)
			}
		})
	}
}

func TestExpWorkflowRunMockProvider(t *testing.T) {
	t.Setenv("PE_TEST_MODE", "true")
	path := writeWorkflowFile(t, "wf.star", `result = len(parallel(["a", "b", "c"]))`)
	out, _, err := runWorkflowCmd(t, path, "--provider", "mock")
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if got := strings.TrimSpace(out); got != "3" {
		t.Errorf("output = %q, want 3", got)
	}
}

func TestExpWorkflowRunTraceFile(t *testing.T) {
	t.Setenv("PE_TEST_MODE", "true")
	path := writeWorkflowFile(t, "wf.star", `
phase("Fanout")
result = parallel(["a", "b"])
`)
	tracePath := filepath.Join(t.TempDir(), "trace.json")
	if _, _, err := runWorkflowCmd(t, path, "--provider", "mock", "--trace", tracePath); err != nil {
		t.Fatalf("run: %v", err)
	}
	data, err := os.ReadFile(tracePath)
	if err != nil {
		t.Fatalf("read trace: %v", err)
	}
	var trace workflow.Trace
	if err := json.Unmarshal(data, &trace); err != nil {
		t.Fatalf("parse trace: %v", err)
	}
	if trace.Version != workflow.TraceVersion {
		t.Errorf("trace version = %q, want %q", trace.Version, workflow.TraceVersion)
	}
	if len(trace.Calls) != 2 {
		t.Errorf("trace calls = %d, want 2", len(trace.Calls))
	}
	for _, c := range trace.Calls {
		if c.Phase != "Fanout" {
			t.Errorf("call %d phase = %q, want Fanout", c.ID, c.Phase)
		}
	}
}

func TestExpWorkflowRunFrontMatterDefaults(t *testing.T) {
	t.Setenv("PE_TEST_MODE", "true")
	path := writeWorkflowFile(t, "wf.pe.md", `---
kind: pe.workflow.v1
name: defaults
inputs:
  topic:
    type: string
    default: caching
---
result = args["topic"]
`)
	out, _, err := runWorkflowCmd(t, path, "--provider", "mock")
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if got := strings.TrimSpace(out); got != "caching" {
		t.Errorf("output = %q, want caching", got)
	}
}

func TestExpWorkflowRunVarOverridesDefault(t *testing.T) {
	t.Setenv("PE_TEST_MODE", "true")
	path := writeWorkflowFile(t, "wf.pe.md", `---
kind: pe.workflow.v1
inputs:
  topic:
    type: string
    default: caching
---
result = args["topic"]
`)
	out, _, err := runWorkflowCmd(t, path, "--provider", "mock", "--var", "topic=locking")
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if got := strings.TrimSpace(out); got != "locking" {
		t.Errorf("output = %q, want locking", got)
	}
}

func TestExpWorkflowRunSafetyProviderDeny(t *testing.T) {
	t.Setenv("PE_TEST_MODE", "true")
	path := writeWorkflowFile(t, "wf.pe.md", `---
kind: pe.workflow.v1
safety:
  providers:
    deny: [mock]
---
result = generate("hi")
`)
	_, _, err := runWorkflowCmd(t, path, "--provider", "mock")
	if err == nil || !strings.Contains(err.Error(), "denied by workflow safety policy") {
		t.Fatalf("err = %v, want workflow safety denial", err)
	}
}

func TestExpWorkflowRunSafetyProviderAllowList(t *testing.T) {
	t.Setenv("PE_TEST_MODE", "true")
	path := writeWorkflowFile(t, "wf.pe.md", `---
kind: pe.workflow.v1
safety:
  providers:
    allow: [openai]
---
result = generate("hi")
`)
	_, _, err := runWorkflowCmd(t, path, "--provider", "mock")
	if err == nil || !strings.Contains(err.Error(), "not in workflow safety allow list") {
		t.Fatalf("err = %v, want allow list denial", err)
	}
}

func TestExpWorkflowRunPeModProviderDeny(t *testing.T) {
	t.Setenv("PE_TEST_MODE", "true")
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	if err := os.WriteFile("pe.mod", []byte(`module example.com/app

pe 1

capability {
    providers deny mock
}
`), 0o644); err != nil {
		t.Fatalf("writing pe.mod: %v", err)
	}
	path := writeWorkflowFile(t, "wf.star", `result = generate("hi")`)
	_, _, err := runWorkflowCmd(t, path, "--provider", "mock")
	if err == nil || !strings.Contains(err.Error(), "denied by pe.mod") {
		t.Fatalf("err = %v, want pe.mod denial", err)
	}
}

func TestExpWorkflowRunMaxCalls(t *testing.T) {
	t.Setenv("PE_TEST_MODE", "true")
	path := writeWorkflowFile(t, "wf.star", `
for i in range(3):
    generate("p" + str(i))
result = True
`)
	_, _, err := runWorkflowCmd(t, path, "--provider", "mock", "--max-calls", "2")
	if err == nil || !strings.Contains(err.Error(), "call budget exceeded") {
		t.Fatalf("err = %v, want call budget exceeded", err)
	}
}
