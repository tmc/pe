package main

import (
	"github.com/spf13/cobra"
)

var expCmd = &cobra.Command{
	Use:   "exp",
	Short: "Experimental features and commands",
	Long:  `This command groups experimental features and commands that are in prototype or development stage.`,
}

func init() {
	// Register experimental commands
	// These are defined in exp_commands.go
}
