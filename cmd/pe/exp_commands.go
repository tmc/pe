package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	expCmd.AddCommand(expDistributedCmd)
	expCmd.AddCommand(expAttestCmd)
	expCmd.AddCommand(expCacheCmd)
	expCmd.AddCommand(expComposeCmd)
}

func createStubCmd(use, short string) *cobra.Command {
	return &cobra.Command{
		Use:   use,
		Short: short,
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("experimental command %q is not yet implemented", use)
		},
	}
}

var expDistributedCmd = newExpDistributedCmd()
var expComposeCmd = createStubCmd("compose", "Compose prompts from verified components with type-safe composition")
