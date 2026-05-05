package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestApplyRootMetadata(t *testing.T) {
	root := &cobra.Command{Use: "pe"}
	root.AddCommand(&cobra.Command{Use: "eval", Short: "evaluate configs"})
	root.AddCommand(&cobra.Command{Use: "fmt", Short: "format configs"})
	root.AddCommand(&cobra.Command{Use: "security", Short: "security testing"})

	applyRootMetadata(root)

	eval, _, err := root.Find([]string{"evaluate"})
	if err != nil {
		t.Fatalf("Find evaluate: %v", err)
	}
	if eval.Name() != "eval" {
		t.Fatalf("evaluate resolved to %s, want eval", eval.Name())
	}
	if got := eval.Annotations[commandGroupAnnotation]; got != "evaluation" {
		t.Fatalf("eval group = %q, want evaluation", got)
	}
	security, _, err := root.Find([]string{"security"})
	if err != nil {
		t.Fatalf("Find security: %v", err)
	}
	if security.Deprecated == "" {
		t.Fatal("security command missing experimental warning")
	}
}

func TestRootGroupedHelp(t *testing.T) {
	root := &cobra.Command{Use: "pe"}
	root.AddCommand(&cobra.Command{Use: "run", Short: "run prompt"})
	root.AddCommand(&cobra.Command{Use: "stats", Short: "show stats"})
	applyRootMetadata(root)

	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetErr(&buf)
	root.SetArgs([]string{"--help"})
	if err := root.Execute(); err != nil {
		t.Fatalf("Execute help: %v", err)
	}
	out := buf.String()
	for _, want := range []string{"Command groups:", "core:", "run", "evaluation:", "stats"} {
		if !strings.Contains(out, want) {
			t.Fatalf("help missing %q:\n%s", want, out)
		}
	}
}
