package main

import (
	"github.com/spf13/cobra"
)

// experimentalCmd groups commands that are outside the core PE workflow.
func experimentalCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "experimental",
		Short: "Experimental and research commands (use with caution)",
		Long: `Experimental commands for prompt engineering research and prototypes.

These commands may be unstable, slow, or produce inconsistent results. Do not
depend on them in production automation without pinning and testing the exact
release.

Examples:
  pe experimental optimize config.yaml
  pe experimental evolve --population 10
  pe experimental semantic analyze prompt.txt`,
	}

	// Add all experimental subcommands
	cmd.AddCommand(optimizeCmd())
	cmd.AddCommand(evolveCmd())
	cmd.AddCommand(semanticCmd())
	cmd.AddCommand(newComposeCmd())
	// fusionCmd moved to advanced-features branch
	cmd.AddCommand(synthesizeCmd)
	cmd.AddCommand(metricsCmd())
	cmd.AddCommand(playgroundCmd())

	// Pipeline and utility commands
	cmd.AddCommand(extractCmd())
	cmd.AddCommand(pluginCmd())
	cmd.AddCommand(streamCmd())
	cmd.AddCommand(filterCmd())
	cmd.AddCommand(analyzeCmd())

	// consensus-demo moved to advanced-features branch

	return cmd
}
