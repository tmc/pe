package main

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

// Version is the PE release version.
var Version = "v0.5.0"

// Commit and Date are set by release builds with -ldflags.
var (
	Commit string
	Date   string
)

func versionLine() string {
	s := fmt.Sprintf("pe version %s %s/%s", Version, runtime.GOOS, runtime.GOARCH)
	if Commit != "" {
		s += " " + Commit
	}
	if Date != "" {
		s += " " + Date
	}
	return s
}

// versionCmd returns a cobra.Command for the 'version' subcommand.
func versionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version number of pe",
		Long:  `All software has versions. This is pe's.`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(versionLine())
		},
	}
}
