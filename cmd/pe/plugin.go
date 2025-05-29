package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tmc/pe/internal/plugin"
)

var pluginManager *plugin.Manager

func init() {
	pluginManager = plugin.NewManager()
}

func pluginCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "plugin",
		Short: "Manage PE plugins",
		Long:  `List, install, and manage PE plugins.`,
	}

	cmd.AddCommand(
		pluginListCmd(),
		pluginRunCmd(),
	)

	return cmd
}

func pluginListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List installed plugins",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := pluginManager.Discover(); err != nil {
				return fmt.Errorf("failed to discover plugins: %w", err)
			}

			plugins := pluginManager.List()
			if len(plugins) == 0 {
				fmt.Println("No plugins found.")
				fmt.Println("\nTo install plugins, place them in your PATH with names like 'pe-<plugin>'")
				return nil
			}

			fmt.Println("Installed plugins:")
			for _, p := range plugins {
				fmt.Printf("  %s - %s\n", p.Name, p.Description)
				if p.Version != "" {
					fmt.Printf("    Version: %s\n", p.Version)
				}
			}

			return nil
		},
	}
}

func pluginRunCmd() *cobra.Command {
	return &cobra.Command{
		Use:                "run <plugin> [args...]",
		Short:              "Run a plugin command",
		DisableFlagParsing: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return fmt.Errorf("plugin name required")
			}

			pluginName := args[0]
			pluginArgs := args[1:]

			if err := pluginManager.Discover(); err != nil {
				return fmt.Errorf("failed to discover plugins: %w", err)
			}

			return pluginManager.Execute(cmd.Context(), pluginName, pluginArgs)
		},
	}
}

// dynamicPluginCommands discovers and adds plugin commands dynamically
func dynamicPluginCommands(rootCmd *cobra.Command) {
	// Discover plugins silently
	if err := pluginManager.Discover(); err != nil {
		return
	}

	// Add each plugin as a top-level command
	for _, p := range pluginManager.List() {
		plugin := p // capture for closure
		cmd := &cobra.Command{
			Use:                plugin.Name,
			Short:              plugin.Description,
			DisableFlagParsing: true,
			Hidden:             false,
			RunE: func(cmd *cobra.Command, args []string) error {
				return pluginManager.Execute(cmd.Context(), plugin.Name, args)
			},
		}

		// Check if command already exists (don't override built-ins)
		existing := false
		for _, c := range rootCmd.Commands() {
			if c.Name() == plugin.Name {
				existing = true
				break
			}
		}

		if !existing {
			rootCmd.AddCommand(cmd)
		}
	}
}

// HandlePluginExecution checks if PE was invoked as a plugin and handles it
func HandlePluginExecution() {
	// Check if we're being invoked with --pe-plugin-info
	if len(os.Args) > 1 && os.Args[1] == "--pe-plugin-info" {
		// This would be implemented by actual plugins
		// For now, just exit as we're not a plugin
		os.Exit(1)
	}

	// Check if we're being invoked as a plugin (pe-*)
	progName := filepath.Base(os.Args[0])
	if strings.HasPrefix(progName, "pe-") && progName != "pe" {
		// We're being invoked as a plugin, but the main PE binary isn't a plugin
		fmt.Fprintf(os.Stderr, "Error: %s is not a valid PE plugin\n", progName)
		os.Exit(1)
	}
}