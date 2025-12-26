package main

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

// Version of the PE tool
const Version = "v0.5.0"

// versionCmd returns a cobra.Command for the 'version' subcommand.
func versionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version number of pe",
		Long:  `All software has versions. This is pe's.`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("pe version %s %s/%s\n", Version, runtime.GOOS, runtime.GOARCH)
		},
	}
}
