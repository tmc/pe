package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var workCmd = &cobra.Command{
	Use:   "work",
	Short: "Manage prompt workspaces",
	Long: `Work provides workspace support for developing multiple related prompts.

A pe.work file in the root of your workspace lets you develop multiple prompt
modules together, similar to go.work files.`,
}

// Workspace represents a pe.work file
type Workspace struct {
	Version string    `yaml:"version"`
	Use     []string  `yaml:"use"`     // Directories containing prompts
	Replace []Replace `yaml:"replace"` // Replace directives
}

type Replace struct {
	Old string `yaml:"old"`
	New string `yaml:"new"`
}

var workInitCmd = &cobra.Command{
	Use:   "init [directories...]",
	Short: "Initialize a workspace",
	Long: `Initialize a workspace in the current directory.

Creates a pe.work file that includes the specified directories.`,
	Example: `  pe work init ./prompts ./team-prompts
  pe work init  # Initialize empty workspace`,
	RunE: runWorkInit,
}

var workUseCmd = &cobra.Command{
	Use:   "use [directories...]",
	Short: "Add directories to workspace",
	Example: `  pe work use ./new-prompts
  pe work use -r ./old-prompts  # Remove directory`,
	RunE: runWorkUse,
}

var workEditCmd = &cobra.Command{
	Use:   "edit",
	Short: "Edit pe.work file programmatically",
	Long:  `Edit provides a command-line interface for editing pe.work files.`,
	RunE:  runWorkEdit,
}

var workSyncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync workspace prompt dependencies",
	Long:  `Sync ensures all prompts referenced in the workspace are available.`,
	RunE:  runWorkSync,
}

var workListCmd = &cobra.Command{
	Use:   "list",
	Short: "List workspace contents",
	RunE:  runWorkList,
}

// Flags
var (
	workRemove      bool
	workReplace     []string
	workDropReplace []string
)

func init() {
	workCmd.AddCommand(workInitCmd)
	workCmd.AddCommand(workUseCmd)
	workCmd.AddCommand(workEditCmd)
	workCmd.AddCommand(workSyncCmd)
	workCmd.AddCommand(workListCmd)

	// Flags for work use
	workUseCmd.Flags().BoolVarP(&workRemove, "remove", "r", false, "Remove directories from workspace")

	// Flags for work edit
	workEditCmd.Flags().StringSliceVar(&workReplace, "replace", []string{}, "Add replace directive (old=new)")
	workEditCmd.Flags().StringSliceVar(&workDropReplace, "dropreplace", []string{}, "Drop replace directive")
}

func runWorkInit(cmd *cobra.Command, args []string) error {
	// Check if go.work already exists
	if _, err := os.Stat("go.work"); err == nil {
		return fmt.Errorf("go.work already exists")
	}

	// Create workspace
	ws := &Workspace{
		Version: "1",
		Use:     args,
	}

	// Normalize paths
	for i, dir := range ws.Use {
		ws.Use[i] = filepath.Clean(dir)
	}

	// Write go.work file in go.work format
	var content strings.Builder
	content.WriteString("go 1.21\n")

	if len(ws.Use) > 0 {
		content.WriteString("\nuse (\n")
		for _, dir := range ws.Use {
			content.WriteString(fmt.Sprintf("\t%s\n", dir))
		}
		content.WriteString(")\n")
	}

	if err := os.WriteFile("go.work", []byte(content.String()), 0644); err != nil {
		return fmt.Errorf("failed to write go.work: %w", err)
	}

	fmt.Println("Created go.work")
	return nil
}

func runWorkUse(cmd *cobra.Command, args []string) error {
	// Load existing workspace
	ws, err := loadWorkspace()
	if err != nil {
		return err
	}

	// Normalize paths
	for i, dir := range args {
		args[i] = filepath.Clean(dir)
	}

	if workRemove {
		// Remove directories
		ws.Use = removeStrings(ws.Use, args)
	} else {
		// Add directories
		for _, dir := range args {
			if !containsString(ws.Use, dir) {
				ws.Use = append(ws.Use, dir)
			}
		}
	}

	// Save workspace
	return saveWorkspace(ws)
}

func runWorkEdit(cmd *cobra.Command, args []string) error {
	// Load existing workspace
	ws, err := loadWorkspace()
	if err != nil {
		return err
	}

	// Handle replace directives
	for _, replace := range workReplace {
		parts := strings.SplitN(replace, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid replace directive: %s (use old=new)", replace)
		}

		// Add or update replace
		found := false
		for i, r := range ws.Replace {
			if r.Old == parts[0] {
				ws.Replace[i].New = parts[1]
				found = true
				break
			}
		}
		if !found {
			ws.Replace = append(ws.Replace, Replace{
				Old: parts[0],
				New: parts[1],
			})
		}
	}

	// Handle drop replace
	for _, drop := range workDropReplace {
		var newReplace []Replace
		for _, r := range ws.Replace {
			if r.Old != drop {
				newReplace = append(newReplace, r)
			}
		}
		ws.Replace = newReplace
	}

	// Save workspace
	return saveWorkspace(ws)
}

func runWorkSync(cmd *cobra.Command, args []string) error {
	ws, err := loadWorkspace()
	if err != nil {
		return err
	}

	fmt.Println("Syncing workspace prompts...")

	// Check each directory
	for _, dir := range ws.Use {
		info, err := os.Stat(dir)
		if err != nil {
			fmt.Printf("  %s: %v\n", dir, err)
			continue
		}

		if !info.IsDir() {
			fmt.Printf("  %s: not a directory\n", dir)
			continue
		}

		// Count prompt files
		count := 0
		err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}

			if strings.HasSuffix(path, ".prompt") || strings.HasSuffix(path, ".pe") {
				count++
			}
			return nil
		})

		fmt.Printf("  %s: %d prompts\n", dir, count)
	}

	return nil
}

func runWorkList(cmd *cobra.Command, args []string) error {
	ws, err := loadWorkspace()
	if err != nil {
		return err
	}

	fmt.Println("Workspace directories:")
	for _, dir := range ws.Use {
		fmt.Printf("  %s\n", dir)

		// List prompts in directory
		var prompts []string
		filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}

			if strings.HasSuffix(path, ".prompt") || strings.HasSuffix(path, ".pe") {
				rel, _ := filepath.Rel(dir, path)
				prompts = append(prompts, rel)
			}
			return nil
		})

		if len(prompts) > 0 {
			for _, p := range prompts {
				fmt.Printf("    - %s\n", p)
			}
		}
	}

	if len(ws.Replace) > 0 {
		fmt.Println("\nReplace directives:")
		for _, r := range ws.Replace {
			fmt.Printf("  %s => %s\n", r.Old, r.New)
		}
	}

	return nil
}

// Helper functions

func loadWorkspace() (*Workspace, error) {
	data, err := os.ReadFile("go.work")
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("no go.work file found (run 'pe work init')")
		}
		return nil, fmt.Errorf("failed to read go.work: %w", err)
	}

	// Parse go.work format
	ws := &Workspace{
		Version: "1",
		Use:     []string{},
	}

	lines := strings.Split(string(data), "\n")
	inUse := false

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if line == "use (" {
			inUse = true
			continue
		}

		if inUse {
			if line == ")" {
				inUse = false
				continue
			}
			// Add directory from use block
			if line != "" {
				ws.Use = append(ws.Use, line)
			}
		}
	}

	return ws, nil
}

func saveWorkspace(ws *Workspace) error {
	// Write go.work file in go.work format
	var content strings.Builder
	content.WriteString("go 1.21\n")

	if len(ws.Use) > 0 {
		content.WriteString("\nuse (\n")
		for _, dir := range ws.Use {
			content.WriteString(fmt.Sprintf("\t%s\n", dir))
		}
		content.WriteString(")\n")
	}

	if err := os.WriteFile("go.work", []byte(content.String()), 0644); err != nil {
		return fmt.Errorf("failed to write go.work: %w", err)
	}

	return nil
}

func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

func removeStrings(slice []string, remove []string) []string {
	var result []string
	for _, s := range slice {
		if !containsString(remove, s) {
			result = append(result, s)
		}
	}
	return result
}
