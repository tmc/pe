package commands

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestCommandRegistryRegisterDiscoverAndHelp(t *testing.T) {
	registry := NewRegistry()
	registry.AddGroup("core", "everyday commands")
	if err := registry.Register("core", &cobra.Command{Use: "run", Short: "run text"}, Metadata{}); err != nil {
		t.Fatalf("Register run: %v", err)
	}
	if err := registry.Register("core", &cobra.Command{Use: "build", Short: "build text"}, Metadata{}); err != nil {
		t.Fatalf("Register build: %v", err)
	}
	got := registry.Discover()
	if len(got) != 2 || got[0].Name != "build" || got[1].Name != "run" {
		t.Fatalf("Discover = %+v", got)
	}
	help := registry.Help()
	for _, want := range []string{"core", "everyday commands", "build", "run"} {
		if !strings.Contains(help, want) {
			t.Fatalf("Help missing %q:\n%s", want, help)
		}
	}
}

func TestCommandRegistryRejectsDuplicates(t *testing.T) {
	registry := NewRegistry()
	if err := registry.Register("core", &cobra.Command{Use: "run"}, Metadata{}); err != nil {
		t.Fatalf("Register run: %v", err)
	}
	if err := registry.Register("core", &cobra.Command{Use: "run"}, Metadata{}); err == nil {
		t.Fatal("Register accepted duplicate command")
	}
}

func TestCommandRegistryAddTo(t *testing.T) {
	registry := NewRegistry()
	if err := registry.Register("core", &cobra.Command{Use: "run"}, Metadata{}); err != nil {
		t.Fatalf("Register run: %v", err)
	}
	root := &cobra.Command{Use: "pe"}
	registry.AddTo(root)
	if _, _, err := root.Find([]string{"run"}); err != nil {
		t.Fatalf("Find run: %v", err)
	}
}
