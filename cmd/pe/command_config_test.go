package main

import (
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/tmc/pe/internal/config"
)

func TestApplyCommandConfigDefaults(t *testing.T) {
	root := &cobra.Command{Use: "pe"}
	run := &cobra.Command{Use: "run"}
	run.Flags().String("provider", "openai", "")
	run.Flags().String("timeout", "", "")
	run.Flags().Int("iterations", 1, "")
	root.AddCommand(run)

	cfg := config.DefaultConfig()
	cfg.Commands["run"] = config.CommandConfig{
		Provider: "mock",
		Timeout:  2 * time.Second,
		Options: map[string]interface{}{
			"iterations": 3,
		},
	}

	applyCommandConfigDefaults(root, cfg)

	if got := run.Flags().Lookup("provider").DefValue; got != "mock" {
		t.Fatalf("provider default = %q, want mock", got)
	}
	if got := run.Flags().Lookup("timeout").DefValue; got != "2s" {
		t.Fatalf("timeout default = %q, want 2s", got)
	}
	if got := run.Annotations[commandTimeoutAnnotation]; got != "2s" {
		t.Fatalf("timeout annotation = %q, want 2s", got)
	}
	if got := run.Flags().Lookup("iterations").DefValue; got != "3" {
		t.Fatalf("iterations default = %q, want 3", got)
	}
}
