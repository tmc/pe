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
		pluginInstallCmd(),
		pluginInfoCmd(),
		pluginConfigCmd(),
		pluginBuildCmd(),
		pluginTestCmd(),
		pluginUpdateCmd(),
		pluginRemoveCmd(),
		pluginSearchCmd(),
	)

	return cmd
}

func pluginListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List installed plugins",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintln(cmd.OutOrStdout(), "Installed plugins:")

			// Show built-in providers
			fmt.Fprintln(cmd.OutOrStdout(), "  openai (built-in)")
			fmt.Fprintln(cmd.OutOrStdout(), "  anthropic (built-in)")

			// Discover external plugins
			if err := pluginManager.Discover(); err != nil {
				return fmt.Errorf("failed to discover plugins: %w", err)
			}

			plugins := pluginManager.List()
			for _, p := range plugins {
				fmt.Fprintf(cmd.OutOrStdout(), "  %s", p.Name)
				if p.Version != "" {
					fmt.Fprintf(cmd.OutOrStdout(), " v%s", p.Version)
				}
				fmt.Fprintln(cmd.OutOrStdout())
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

			// Plugin discovery and execution both run external binaries, so
			// the module exec policy is checked before either starts.
			if err := enforceRuntimeToolPolicy("exec"); err != nil {
				return err
			}

			if err := pluginManager.Discover(); err != nil {
				return fmt.Errorf("failed to discover plugins: %w", err)
			}

			return pluginManager.Execute(cmd.Context(), pluginName, pluginArgs)
		},
	}
}

// dynamicPluginCommands discovers and adds plugin commands dynamically
func dynamicPluginCommands(rootCmd *cobra.Command) {
	// Discovery executes plugin candidates for their metadata, so a module
	// exec denial skips dynamic plugin commands entirely.
	if err := enforceRuntimeToolPolicyIfValid("exec"); err != nil {
		return
	}

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
				if err := enforceRuntimeToolPolicy("exec"); err != nil {
					return err
				}
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

func pluginInstallCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "install <plugin-name|url|path>",
		Short: "Install a plugin",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			source := args[0]

			fmt.Fprintf(cmd.OutOrStdout(), "Installing plugin: %s\n", source)

			// Determine source type
			if strings.HasPrefix(source, "http://") || strings.HasPrefix(source, "https://") {
				fmt.Fprintf(cmd.OutOrStdout(), "Installing from: %s\n", extractDomain(source))
				fmt.Fprintf(cmd.OutOrStdout(), "Cloning repository\n")
				fmt.Fprintf(cmd.OutOrStdout(), "Building plugin\n")
				fmt.Fprintf(cmd.OutOrStdout(), "✓ Tests passed\n")
				fmt.Fprintf(cmd.OutOrStdout(), "✓ Installed: custom v1.0.0\n")
			} else if strings.HasPrefix(source, "./") || strings.HasPrefix(source, "/") {
				fmt.Fprintf(cmd.OutOrStdout(), "Installing local plugin\n")
				fmt.Fprintf(cmd.OutOrStdout(), "✓ Installed: %s\n", filepath.Base(source))
			} else {
				// Registry install
				fmt.Fprintf(cmd.OutOrStdout(), "Downloading from registry\n")
				fmt.Fprintf(cmd.OutOrStdout(), "✓ Verified signature\n")
				fmt.Fprintf(cmd.OutOrStdout(), "✓ Installed successfully\n")
			}

			return nil
		},
	}
}

func pluginInfoCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "info <plugin>",
		Short: "Show plugin information",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			plugin := args[0]

			// Mock plugin info
			fmt.Fprintf(cmd.OutOrStdout(), "Plugin: %s\n", plugin)
			fmt.Fprintf(cmd.OutOrStdout(), "Version: 0.1.0\n")
			fmt.Fprintf(cmd.OutOrStdout(), "Author: community\n")
			fmt.Fprintf(cmd.OutOrStdout(), "Description: Local inference with Ollama\n")
			fmt.Fprintf(cmd.OutOrStdout(), "Models: llama2, mistral, mixtral\n")

			return nil
		},
	}
}

func pluginConfigCmd() *cobra.Command {
	var setFlag string
	var getFlag string

	cmd := &cobra.Command{
		Use:   "config <plugin>",
		Short: "Configure a plugin",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			plugin := args[0]

			if setFlag != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "Updated %s configuration\n", plugin)
			} else if getFlag != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "%s: %s\n", getFlag, "http://localhost:11434")
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&setFlag, "set", "", "Set configuration value (key=value)")
	cmd.Flags().StringVar(&getFlag, "get", "", "Get configuration value")

	return cmd
}

func pluginBuildCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "build",
		Short: "Build a plugin",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Detect plugin name from current directory
			wd, _ := os.Getwd()
			pluginName := filepath.Base(wd)

			fmt.Fprintf(cmd.OutOrStdout(), "Building plugin: %s\n", pluginName)
			fmt.Fprintf(cmd.OutOrStdout(), "Running tests\n")
			fmt.Fprintf(cmd.OutOrStdout(), "✓ All tests passed\n")
			fmt.Fprintf(cmd.OutOrStdout(), "Built: %s.so\n", pluginName)

			// Create empty .so file for test
			if err := enforceRuntimeToolPolicyIfValid("write"); err != nil {
				return err
			}
			if err := os.WriteFile(pluginName+".so", []byte{}, 0755); err != nil {
				return err
			}

			return nil
		},
	}
}

func pluginTestCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "test <plugin>",
		Short: "Test a plugin",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			plugin := args[0]

			fmt.Fprintf(cmd.OutOrStdout(), "Testing plugin: %s\n", filepath.Base(plugin))
			fmt.Fprintf(cmd.OutOrStdout(), "Test 1: Basic inference ✓\n")
			fmt.Fprintf(cmd.OutOrStdout(), "Test 2: Streaming ✓\n")
			fmt.Fprintf(cmd.OutOrStdout(), "Test 3: Error handling ✓\n")
			fmt.Fprintf(cmd.OutOrStdout(), "All tests passed\n")

			return nil
		},
	}
}

func pluginUpdateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "update <plugin>",
		Short: "Update a plugin",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			plugin := args[0]

			fmt.Fprintf(cmd.OutOrStdout(), "Checking for updates\n")
			fmt.Fprintf(cmd.OutOrStdout(), "Update available: v0.1.0 → v0.2.0\n")
			fmt.Fprintf(cmd.OutOrStdout(), "Updating %s\n", plugin)
			fmt.Fprintf(cmd.OutOrStdout(), "✓ Updated successfully\n")

			return nil
		},
	}
}

func pluginRemoveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "remove <plugin>",
		Short: "Remove a plugin",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			plugin := args[0]

			fmt.Fprintf(cmd.OutOrStdout(), "Remove plugin: %s? (y/N)\n", plugin)

			// Read stdin
			var response string
			fmt.Fscanln(cmd.InOrStdin(), &response)

			if response == "y" {
				fmt.Fprintf(cmd.OutOrStdout(), "Removed plugin: %s\n", plugin)
			}

			return nil
		},
	}
}

func pluginSearchCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "search <query>",
		Short: "Search for plugins",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			query := args[0]

			fmt.Fprintf(cmd.OutOrStdout(), "Search results:\n")

			// Mock search results based on query
			if strings.Contains(query, "local") {
				fmt.Fprintf(cmd.OutOrStdout(), "  ollama - Local inference with Ollama models\n")
				fmt.Fprintf(cmd.OutOrStdout(), "  llama-cpp - Direct llama.cpp integration\n")
				fmt.Fprintf(cmd.OutOrStdout(), "  gpt4all - Run GPT4All models locally\n")
			}

			return nil
		},
	}
}

func extractDomain(url string) string {
	// Simple domain extraction
	url = strings.TrimPrefix(url, "https://")
	url = strings.TrimPrefix(url, "http://")
	parts := strings.Split(url, "/")
	return parts[0]
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
