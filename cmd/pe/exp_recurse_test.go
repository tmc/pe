package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tmc/pe/internal/rlm"
)

func TestExpRecurseRegistered(t *testing.T) {
	found := false
	for _, c := range expCmd.Commands() {
		if c.Name() == "recurse" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("recurse command not registered under exp")
	}
}

func TestExpRecurseRequiresPrompt(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "input.txt")
	if err := os.WriteFile(target, []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}

	cmd := expRecurseCmd()
	cmd.SetArgs([]string{target})
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	if err := cmd.Execute(); err == nil {
		t.Fatal("Execute succeeded, want missing prompt error")
	}
}

func TestExpRecurseMissingTarget(t *testing.T) {
	cmd := expRecurseCmd()
	cmd.SetArgs([]string{"--prompt", "p", filepath.Join(t.TempDir(), "nope.txt")})
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	if err := cmd.Execute(); err == nil {
		t.Fatal("Execute succeeded, want read target error")
	}
}

func TestExpRecurseEndToEndMock(t *testing.T) {
	t.Setenv("PE_TEST_MODE", "true")

	dir := t.TempDir()
	target := filepath.Join(dir, "input.txt")
	// 10 bytes at chunk size 4 -> 3 depth-0 chunks -> 3 map calls, then a
	// reduce pass folds them at the default --max-depth of 1.
	if err := os.WriteFile(target, []byte("0123456789"), 0644); err != nil {
		t.Fatal(err)
	}

	cmd := expRecurseCmd()
	cmd.SetArgs([]string{
		"--prompt", "summarize",
		"--provider", "mock",
		"--chunk-size", "4",
		"--consensus", "concat",
		target,
	})
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute: %v\n%s", err, out.String())
	}

	var trace rlm.Trace
	if err := json.Unmarshal(out.Bytes(), &trace); err != nil {
		t.Fatalf("decode trace: %v\n%s", err, out.String())
	}
	if trace.Version != rlm.TraceVersion {
		t.Errorf("version = %q, want %q", trace.Version, rlm.TraceVersion)
	}
	if trace.Command != "pe exp recurse" {
		t.Errorf("command = %q", trace.Command)
	}

	// The target splits into exactly three depth-0 map calls; each goes through
	// the mock provider with its fixed token cost and a content-addressed
	// snippet. Reduce passes may add further calls at deeper levels.
	mapCalls := 0
	for i, call := range trace.Calls {
		if call.Depth != 0 {
			continue
		}
		mapCalls++
		if call.Operation != "map" {
			t.Errorf("call %d operation = %q, want map", i, call.Operation)
		}
		if call.Provider != "mock" {
			t.Errorf("call %d provider = %q, want mock", i, call.Provider)
		}
		if call.Cost.TotalTokens != 15 {
			t.Errorf("call %d total tokens = %d, want 15 (mock)", i, call.Cost.TotalTokens)
		}
		if call.Snippet.CacheKey == "" {
			t.Errorf("call %d missing snippet cache key", i)
		}
	}
	if mapCalls != 3 {
		t.Fatalf("depth-0 map calls = %d, want 3", mapCalls)
	}
	if trace.TerminationReason != rlm.TerminationCompleted {
		t.Errorf("termination = %q", trace.TerminationReason)
	}
}

func TestExpRecurseTraceWriteDeniedByPolicy(t *testing.T) {
	t.Setenv("PE_TEST_MODE", "true")

	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(oldWd)

	if err := os.WriteFile("input.txt", []byte("data"), 0644); err != nil {
		t.Fatal(err)
	}
	writeExpPolicyDenyWritePeMod(t)

	cmd := expRecurseCmd()
	cmd.SetArgs([]string{
		"--prompt", "p",
		"--provider", "mock",
		"--trace", "trace.json",
		"input.txt",
	})
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	err := cmd.Execute()
	if err == nil {
		t.Fatal("Execute succeeded, want write policy error")
	}
	if !strings.Contains(err.Error(), "tool write is denied by pe.mod") {
		t.Fatalf("error = %v, want write policy error", err)
	}
	if _, statErr := os.Stat("trace.json"); !os.IsNotExist(statErr) {
		t.Fatalf("trace.json stat error = %v, want not exist", statErr)
	}
}
