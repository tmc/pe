package promptcmd

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/spf13/cobra"
	"github.com/tmc/pe/internal/prompt"
	"gopkg.in/yaml.v3"
)

// NewCommand creates the prompt command group
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "prompt",
		Short: "Manage prompt files",
		Long: `Manage prompt files with initialization, formatting, and metadata.

Commands for working with prompt files, similar to 'go mod' for Go modules.`,
	}

	cmd.AddCommand(newInitCommand())
	cmd.AddCommand(newEditCommand())
	cmd.AddCommand(newInfoCommand())
	cmd.AddCommand(newTidyCommand())
	cmd.AddCommand(newHelpCommand())

	return cmd
}

// newInitCommand creates a new prompt file with shebang and structure
func newInitCommand() *cobra.Command {
	var (
		provider     string
		system       string
		withDefaults bool
		withTests    bool
		withVariants bool
		force        bool
	)

	cmd := &cobra.Command{
		Use:   "init [file]",
		Short: "Initialize a new prompt file with shebang and structure",
		Long: `Initialize a new prompt file with proper structure.

Creates a prompt file with:
- Shebang line for direct execution
- Variable defaults section
- Optional test cases
- Optional variants

Examples:
  pe prompt init my-prompt.prompt --provider=openai
  pe prompt init analyzer.prompt --with-defaults --with-tests
  pe prompt init --provider=cgpt  # Creates prompt.prompt`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			filename := "prompt.prompt"
			if len(args) > 0 {
				filename = args[0]
			}

			// Ensure .prompt extension
			if !strings.HasSuffix(filename, ".prompt") {
				filename += ".prompt"
			}

			// Check if file exists
			if _, err := os.Stat(filename); err == nil && !force {
				return fmt.Errorf("file %s already exists (use --force to overwrite)", filename)
			}

			// Build the prompt file content
			var buf bytes.Buffer

			// Add shebang line
			buf.WriteString("#!/usr/bin/env pe run")
			if provider != "" {
				buf.WriteString(" --provider=" + provider)
			}
			if system != "" {
				buf.WriteString(" --system=\"" + system + "\"")
			}
			buf.WriteString("\n\n")

			// Add main prompt section
			buf.WriteString("# Main prompt\n")
			buf.WriteString("{{.task}}\n\n")

			// Add defaults section if requested
			if withDefaults {
				buf.WriteString("---defaults---\n")
				buf.WriteString("task: Explain this concept clearly\n")
				buf.WriteString("style: concise\n")
				buf.WriteString("audience: general\n\n")
			}

			// Add system prompt section
			if system != "" {
				buf.WriteString("---system---\n")
				buf.WriteString(system + "\n\n")
			}

			// Add tests section if requested
			if withTests {
				buf.WriteString("---tests---\n")
				buf.WriteString("- name: basic_test\n")
				buf.WriteString("  vars:\n")
				buf.WriteString("    task: \"What is 2+2?\"\n")
				buf.WriteString("  assert:\n")
				buf.WriteString("    - type: contains\n")
				buf.WriteString("      value: \"4\"\n\n")
			}

			// Add variants section if requested
			if withVariants {
				buf.WriteString("---variants---\n")
				buf.WriteString("verbose:\n")
				buf.WriteString("  Provide a detailed explanation of {{.task}}\n")
				buf.WriteString("  Include examples and context.\n\n")
				buf.WriteString("brief:\n")
				buf.WriteString("  {{.task}} (be very concise)\n\n")
			}

			// Write the file
			if err := os.WriteFile(filename, buf.Bytes(), 0755); err != nil {
				return fmt.Errorf("writing file: %w", err)
			}

			fmt.Printf("Created %s\n", filename)
			if provider == "" {
				fmt.Println("Tip: Set a default provider with --provider flag")
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&provider, "provider", "cgpt", "Default provider for the prompt")
	cmd.Flags().StringVar(&system, "system", "", "System prompt to include")
	cmd.Flags().BoolVar(&withDefaults, "with-defaults", true, "Include defaults section")
	cmd.Flags().BoolVar(&withTests, "with-tests", false, "Include test cases")
	cmd.Flags().BoolVar(&withVariants, "with-variants", false, "Include variants section")
	cmd.Flags().BoolVar(&force, "force", false, "Overwrite existing file")

	return cmd
}

// newEditCommand edits prompt file metadata and defaults
func newEditCommand() *cobra.Command {
	var (
		setDefault    map[string]string
		removeDefault []string
		setProvider   string
		addVariant    string
		removeVariant string
	)

	cmd := &cobra.Command{
		Use:   "edit [file]",
		Short: "Edit prompt file defaults and metadata",
		Long: `Edit prompt file defaults, variables, and metadata.

Examples:
  pe prompt edit my.prompt --set-default task="Analyze this code"
  pe prompt edit my.prompt --set-provider openai
  pe prompt edit my.prompt --remove-default task
  pe prompt edit my.prompt --add-variant detailed`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			filename := args[0]

			// Read the existing file
			content, err := os.ReadFile(filename)
			if err != nil {
				return fmt.Errorf("reading file: %w", err)
			}

			lines := strings.Split(string(content), "\n")
			var newLines []string
			inDefaults := false
			defaultsExist := false
			defaultsMap := make(map[string]string)

			// Parse the file
			for i, line := range lines {
				if strings.HasPrefix(line, "---defaults---") {
					inDefaults = true
					defaultsExist = true
					newLines = append(newLines, line)
					continue
				} else if inDefaults && strings.HasPrefix(line, "---") {
					inDefaults = false
				}

				if inDefaults {
					if strings.TrimSpace(line) != "" {
						parts := strings.SplitN(line, ":", 2)
						if len(parts) == 2 {
							key := strings.TrimSpace(parts[0])
							value := strings.TrimSpace(parts[1])
							defaultsMap[key] = value
						}
					}
				} else {
					// Handle shebang line provider update
					if i == 0 && strings.HasPrefix(line, "#!/usr/bin/env pe run") && setProvider != "" {
						// Parse and update the shebang line
						shebang := "#!/usr/bin/env pe run"
						if strings.Contains(line, "--provider=") {
							// Replace existing provider
							parts := strings.Fields(line)
							var newParts []string
							for _, part := range parts {
								if !strings.HasPrefix(part, "--provider=") {
									newParts = append(newParts, part)
								}
							}
							newParts = append(newParts, "--provider="+setProvider)
							shebang = strings.Join(newParts, " ")
						} else {
							// Add provider
							shebang = line + " --provider=" + setProvider
						}
						newLines = append(newLines, shebang)
					} else if !inDefaults {
						newLines = append(newLines, line)
					}
				}
			}

			// Apply changes to defaults
			for k, v := range setDefault {
				defaultsMap[k] = v
			}
			for _, k := range removeDefault {
				delete(defaultsMap, k)
			}

			// If we have defaults but no defaults section, add it
			if len(defaultsMap) > 0 && !defaultsExist {
				// Find where to insert defaults (after shebang and main prompt)
				insertIdx := 2
				for i, line := range newLines {
					if strings.TrimSpace(line) != "" && !strings.HasPrefix(line, "#") && i > 0 {
						insertIdx = i + 1
						break
					}
				}
				newLines = append(newLines[:insertIdx],
					append([]string{"", "---defaults---"}, newLines[insertIdx:]...)...)
				defaultsExist = true
			}

			// Rebuild the file with updated defaults
			if defaultsExist && len(defaultsMap) > 0 {
				var finalLines []string
				inDefaults = false
				defaultsWritten := false

				for _, line := range newLines {
					if strings.HasPrefix(line, "---defaults---") {
						inDefaults = true
						finalLines = append(finalLines, line)
						// Write all defaults here
						for k, v := range defaultsMap {
							finalLines = append(finalLines, fmt.Sprintf("%s: %s", k, v))
						}
						defaultsWritten = true
						continue
					} else if inDefaults && strings.HasPrefix(line, "---") {
						inDefaults = false
						finalLines = append(finalLines, line)
						continue
					}

					if !inDefaults || !defaultsWritten {
						finalLines = append(finalLines, line)
					}
				}
				newLines = finalLines
			}

			// Write the updated file
			output := strings.Join(newLines, "\n")
			if err := os.WriteFile(filename, []byte(output), 0755); err != nil {
				return fmt.Errorf("writing file: %w", err)
			}

			fmt.Printf("Updated %s\n", filename)
			return nil
		},
	}

	cmd.Flags().StringToStringVar(&setDefault, "set-default", nil, "Set default variable values")
	cmd.Flags().StringSliceVar(&removeDefault, "remove-default", nil, "Remove default variables")
	cmd.Flags().StringVar(&setProvider, "set-provider", "", "Set the default provider")
	cmd.Flags().StringVar(&addVariant, "add-variant", "", "Add a variant (interactive)")
	cmd.Flags().StringVar(&removeVariant, "remove-variant", "", "Remove a variant")

	return cmd
}

// newInfoCommand displays information about a prompt file
func newInfoCommand() *cobra.Command {
	var (
		jsonOutput bool
		verbose    bool
	)

	cmd := &cobra.Command{
		Use:   "info [file]",
		Short: "Display information about a prompt file",
		Long: `Display information about a prompt file including variables, defaults, and metadata.

Examples:
  pe prompt info my.prompt
  pe prompt info my.prompt --json
  pe prompt info my.prompt --verbose`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			filename := args[0]

			content, err := os.ReadFile(filename)
			if err != nil {
				return fmt.Errorf("reading file: %w", err)
			}

			// Use the new metadata extraction from internal package
			meta := prompt.ExtractMetadata(string(content))

			if jsonOutput {
				data, err := yaml.Marshal(meta)
				if err != nil {
					return err
				}
				fmt.Print(string(data))
			} else {
				fmt.Printf("File: %s\n", filename)
				fmt.Printf("Executable: %v\n", meta.HasShebang)
				
				if meta.Provider != "" {
					fmt.Printf("Provider: %s\n", meta.Provider)
				}

				if len(meta.Variables) > 0 {
					fmt.Printf("Variables: %s\n", strings.Join(meta.Variables, ", "))
				}

				if len(meta.Defaults) > 0 {
					fmt.Println("Defaults:")
					for k, v := range meta.Defaults {
						fmt.Printf("  %s: %s\n", k, v)
					}
				}

				if len(meta.Sections) > 0 {
					fmt.Printf("Sections: %s\n", strings.Join(meta.Sections, ", "))
				}

				if verbose {
					if meta.SystemPrompt != "" {
						fmt.Printf("System Prompt: %s\n", meta.SystemPrompt)
					}
				}
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&jsonOutput, "json", "j", false, "Output in JSON format")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show verbose information")

	return cmd
}

// newTidyCommand cleans up and validates prompt files
func newTidyCommand() *cobra.Command {
	var (
		removeUnused bool
		validate     bool
	)

	cmd := &cobra.Command{
		Use:   "tidy [file]",
		Short: "Clean up and validate prompt files",
		Long: `Clean up and validate prompt files.

Actions:
- Remove unused variables from defaults
- Validate template syntax
- Check for missing variables
- Ensure proper structure

Examples:
  pe prompt tidy my.prompt
  pe prompt tidy my.prompt --remove-unused
  pe prompt tidy *.prompt --validate`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			for _, pattern := range args {
				files, err := filepath.Glob(pattern)
				if err != nil {
					return fmt.Errorf("globbing pattern %s: %w", pattern, err)
				}

				for _, file := range files {
					if err := tidyPromptFile(file, removeUnused, validate); err != nil {
						fmt.Fprintf(os.Stderr, "Error tidying %s: %v\n", file, err)
					}
				}
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&removeUnused, "remove-unused", false, "Remove unused default variables")
	cmd.Flags().BoolVar(&validate, "validate", true, "Validate template syntax")

	return cmd
}

// tidyPromptFile tidies a single prompt file
func tidyPromptFile(filename string, removeUnused, validate bool) error {
	content, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	lines := strings.Split(string(content), "\n")
	
	// Extract all variables used in the prompt
	usedVars := make(map[string]bool)
	mainContent := prompt.ExtractMainContent(string(content))
	for _, v := range prompt.FindVariables(mainContent) {
		usedVars[v] = true
	}

	// Validate template syntax if requested
	if validate {
		tmpl := template.New("prompt")
		if _, err := tmpl.Parse(mainContent); err != nil {
			return fmt.Errorf("template syntax error: %w", err)
		}
		fmt.Printf("%s: valid\n", filename)
	}

	// Remove unused defaults if requested
	if removeUnused {
		lines = removeUnusedDefaults(lines, usedVars)
		if err := os.WriteFile(filename, []byte(strings.Join(lines, "\n")), 0755); err != nil {
			return err
		}
		fmt.Printf("%s: tidied\n", filename)
	}

	return nil
}

// newHelpCommand shows usage information for a prompt file
func newHelpCommand() *cobra.Command {
	var (
		asScript bool
	)
	
	cmd := &cobra.Command{
		Use:   "help [file]",
		Short: "Show usage information for a prompt file",
		Long: `Show usage information for a prompt file, including variables, defaults, and examples.

This command displays comprehensive help for a prompt file, showing:
- Description from comments
- Required and optional variables
- Default values
- Usage examples

Examples:
  pe prompt help my-script.prompt
  pe prompt help analyzer --as-script`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			filename := args[0]
			
			// Read the prompt file
			content, err := os.ReadFile(filename)
			if err != nil {
				return fmt.Errorf("reading file: %w", err)
			}
			
			// Use the metadata extraction from internal package
			meta := prompt.ExtractMetadata(string(content))
			
			// Display help
			scriptName := filepath.Base(filename)
			fmt.Printf("%s\n", scriptName)
			
			if meta.Description != "" {
				fmt.Printf("\n%s\n", meta.Description)
			}
			
			fmt.Printf("\nUsage:\n")
			if asScript {
				fmt.Printf("  %s", scriptName)
			} else {
				fmt.Printf("  pe run %s", filename)
			}
			
			// Show variable flags
			for _, v := range meta.Variables {
				if def, hasDefault := meta.Defaults[v]; hasDefault {
					fmt.Printf(" [--%s=value]", strings.ToLower(v))
					_ = def // We'll show the default in the Variables section
				} else {
					fmt.Printf(" --%s=value", strings.ToLower(v))
				}
			}
			fmt.Printf("\n")
			
			// Show variables section
			if len(meta.Variables) > 0 {
				fmt.Printf("\nVariables:\n")
				for _, v := range meta.Variables {
					if def, hasDefault := meta.Defaults[v]; hasDefault {
						fmt.Printf("  --%s string    %s (default: %s)\n", 
							strings.ToLower(v), v, def)
					} else {
						fmt.Printf("  --%s string    %s (required)\n", 
							strings.ToLower(v), v)
					}
				}
			}
			
			// Show system prompt if present
			if meta.SystemPrompt != "" {
				fmt.Printf("\nSystem Prompt:\n  %s\n", 
					strings.ReplaceAll(meta.SystemPrompt, "\n", "\n  "))
			}
			
			// Examples
			fmt.Printf("\nExamples:\n")
			
			// Basic example
			if asScript {
				fmt.Printf("  # Run with defaults\n")
				fmt.Printf("  ./%s\n", scriptName)
				
				if len(meta.Variables) > 0 {
					fmt.Printf("\n  # Override variables\n")
					fmt.Printf("  ./%s", scriptName)
					for _, v := range meta.Variables {
						fmt.Printf(" --%s=\"value\"", strings.ToLower(v))
					}
					fmt.Printf("\n")
				}
			} else {
				fmt.Printf("  # Run with defaults\n")
				fmt.Printf("  pe run %s\n", filename)
				
				if len(meta.Variables) > 0 {
					fmt.Printf("\n  # Override variables\n") 
					fmt.Printf("  pe run %s", filename)
					for _, v := range meta.Variables {
						fmt.Printf(" --var %s=\"value\"", v)
					}
					fmt.Printf("\n")
					
					fmt.Printf("\n  # Using direct flags (with custom parser)\n")
					fmt.Printf("  pe run %s", filename)
					for _, v := range meta.Variables {
						fmt.Printf(" --%s=\"value\"", strings.ToLower(v))
					}
					fmt.Printf("\n")
				}
			}
			
			return nil
		},
	}
	
	cmd.Flags().BoolVar(&asScript, "as-script", false, 
		"Show help as if running as standalone script")
	
	return cmd
}

// Helper function for removing unused defaults
func removeUnusedDefaults(lines []string, usedVars map[string]bool) []string {
	var result []string
	inDefaults := false
	
	for _, line := range lines {
		if strings.HasPrefix(line, "---defaults---") {
			inDefaults = true
			result = append(result, line)
			continue
		} else if inDefaults && strings.HasPrefix(line, "---") {
			inDefaults = false
		}
		
		if inDefaults && strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				if usedVars[key] {
					result = append(result, line)
				} else {
					fmt.Printf("  Removing unused default: %s\n", key)
				}
			}
		} else {
			result = append(result, line)
		}
	}
	
	return result
}