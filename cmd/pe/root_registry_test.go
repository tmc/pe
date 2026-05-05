package main

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestRegisterRootCommands(t *testing.T) {
	root := &cobra.Command{Use: "pe"}
	registerRootCommands(root)
	applyRootMetadata(root)

	tests := []struct {
		name  string
		group string
	}{
		{name: "run", group: "core"},
		{name: "build", group: "core"},
		{name: "test", group: "core"},
		{name: "ask", group: "core"},
		{name: "prompt", group: "core"},
		{name: "edit", group: "core"},
		{name: "work", group: "core"},
		{name: "serve", group: "core"},
		{name: "eval", group: "evaluation"},
		{name: "benchmark", group: "evaluation"},
		{name: "diff", group: "evaluation"},
		{name: "stats", group: "evaluation"},
		{name: "view", group: "evaluation"},
		{name: "optimize", group: "optimization"},
		{name: "semantic", group: "optimization"},
		{name: "evolve", group: "optimization"},
		{name: "mod", group: "module"},
		{name: "push", group: "module"},
		{name: "get", group: "module"},
		{name: "stream", group: "pipeline"},
		{name: "filter", group: "pipeline"},
		{name: "analyze", group: "pipeline"},
		{name: "collect", group: "pipeline"},
		{name: "reduce", group: "pipeline"},
		{name: "extract", group: "pipeline"},
		{name: "expand", group: "pipeline"},
		{name: "compose", group: "pipeline"},
		{name: "cat", group: "pipeline"},
		{name: "fmt", group: "utility"},
		{name: "vet", group: "utility"},
		{name: "convert", group: "utility"},
		{name: "template", group: "utility"},
		{name: "interactive", group: "utility"},
		{name: "watch", group: "utility"},
		{name: "security", group: "experimental"},
		{name: "profile", group: "experimental"},
		{name: "plugin", group: "plugin"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, _, err := root.Find([]string{tt.name})
			if err != nil {
				t.Fatalf("Find(%s): %v", tt.name, err)
			}
			if cmd.Name() != tt.name {
				t.Fatalf("Find(%s) = %s", tt.name, cmd.Name())
			}
			if got := cmd.Annotations[commandGroupAnnotation]; got != tt.group {
				t.Fatalf("group = %q, want %q", got, tt.group)
			}
		})
	}
}

func TestApplyRootMetadataInheritsSubcommandGroups(t *testing.T) {
	root := &cobra.Command{Use: "pe"}
	mod := &cobra.Command{Use: "mod"}
	mod.AddCommand(&cobra.Command{Use: "init"}, &cobra.Command{Use: "tidy"})
	root.AddCommand(mod)

	applyRootMetadata(root)

	for _, name := range []string{"init", "tidy"} {
		cmd, _, err := root.Find([]string{"mod", name})
		if err != nil {
			t.Fatalf("Find mod %s: %v", name, err)
		}
		if got, want := cmd.Annotations[commandGroupAnnotation], "module"; got != want {
			t.Fatalf("mod %s group = %q, want %q", name, got, want)
		}
	}
}
