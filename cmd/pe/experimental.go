package main

import (
	"github.com/spf13/cobra"
)

// experimentalCmd groups all experimental, academic, and research commands
// These are not part of the core PE workflow but are available for advanced users
func experimentalCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "experimental",
		Short: "Experimental and research commands (use with caution)",
		Long: `Experimental commands for advanced prompt engineering research.

These commands implement cutting-edge techniques from academic papers
and research projects. They may be unstable, slow, or produce 
inconsistent results. Not recommended for production use.

Examples:
  pe experimental optimize config.yaml
  pe experimental evolve --population 10
  pe experimental semantic analyze prompt.txt`,
	}

	// Add all experimental subcommands
	cmd.AddCommand(optimizeCmd())
	cmd.AddCommand(evolveCmd())
	cmd.AddCommand(semanticCmd())
	cmd.AddCommand(composeCmd)
	cmd.AddCommand(fusionCmd)
	cmd.AddCommand(synthesizeCmd)
	cmd.AddCommand(metricsCmd())
	cmd.AddCommand(playgroundCmd())
	
	// Pipeline and utility commands
	cmd.AddCommand(extractCmd())
	cmd.AddCommand(pluginCmd())
	cmd.AddCommand(streamCmd())
	cmd.AddCommand(filterCmd())
	cmd.AddCommand(analyzeCmd())
	
	// Demo commands
	cmd.AddCommand(&cobra.Command{
		Use:   "consensus-demo",
		Short: "Demo: Distributed consensus visualization",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runConsensusDemo()
		},
	})

	return cmd
}