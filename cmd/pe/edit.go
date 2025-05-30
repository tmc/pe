package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tmc/pe/internal/prompt"
)

var editCmd = &cobra.Command{
	Use:   "edit [prompt-file]",
	Short: "Edit prompt files programmatically",
	Long: `Edit prompt files programmatically, inspired by 'go mod edit'.

This command provides a programmatic interface for editing prompt files
without manual text editing. All modifications preserve the prompt format.

Examples:
  # Set the main prompt
  pe edit prompt.txt --set-prompt "Summarize this document"
  
  # Set the system prompt
  pe edit prompt.txt --set-system-prompt "You are a helpful assistant"
  
  # Add a variant
  pe edit prompt.txt --add-variant academic --variant-cmd "extend-system-prompt 'Use academic language'"
  
  # Add a section
  pe edit prompt.txt --add-section examples --section-content "Example 1: ..."
  
  # Output as JSON
  pe edit prompt.txt --json
  
  # Set examples
  pe edit prompt.txt --set-examples "Example 1: ..."
  
  # Set defaults
  pe edit prompt.txt --set-defaults "lang=en&model=gpt-4"
  pe edit prompt.txt --add-default lang=en
  
  # Format the file
  pe edit prompt.txt --fmt`,
	Args: cobra.RangeArgs(0, 1),
	RunE: runEdit,
}

var (
	editSetPrompt       string
	editSetSystemPrompt string
	editSetExamples     string
	editSetDefaults     string
	editAddDefault      string
	editRemoveDefault   string
	editAddVariant      string
	editVariantCmd      string
	editRemoveVariant   string
	editAddSection      string
	editSectionContent  string
	editRemoveSection   string
	editAppendPrompt    string
	editPrependPrompt   string
	editJSON            bool
	editPrint           bool
	editFmt             bool
	editModule          string
)

func init() {
	editCmd.Flags().StringVar(&editSetPrompt, "set-prompt", "", "Set the main prompt text")
	editCmd.Flags().StringVar(&editSetSystemPrompt, "set-system-prompt", "", "Set the system prompt")
	editCmd.Flags().StringVar(&editSetExamples, "set-examples", "", "Set the examples section")
	editCmd.Flags().StringVar(&editSetDefaults, "set-defaults", "", "Set defaults (format: key1=val1&key2=val2)")
	editCmd.Flags().StringVar(&editAddDefault, "add-default", "", "Add a default (format: key=value)")
	editCmd.Flags().StringVar(&editRemoveDefault, "remove-default", "", "Remove a default by key")
	editCmd.Flags().StringVar(&editAppendPrompt, "append-prompt", "", "Append text to the main prompt")
	editCmd.Flags().StringVar(&editPrependPrompt, "prepend-prompt", "", "Prepend text to the main prompt")
	editCmd.Flags().StringVar(&editAddVariant, "add-variant", "", "Add a variant")
	editCmd.Flags().StringVar(&editVariantCmd, "variant-cmd", "", "Commands for the variant (use with --add-variant)")
	editCmd.Flags().StringVar(&editRemoveVariant, "remove-variant", "", "Remove a variant")
	editCmd.Flags().StringVar(&editAddSection, "add-section", "", "Add a section")
	editCmd.Flags().StringVar(&editSectionContent, "section-content", "", "Content for the section (use with --add-section)")
	editCmd.Flags().StringVar(&editRemoveSection, "remove-section", "", "Remove a section")
	editCmd.Flags().BoolVar(&editJSON, "json", false, "Output the prompt in JSON format")
	editCmd.Flags().BoolVar(&editPrint, "print", false, "Print the result instead of writing to file")
	editCmd.Flags().BoolVar(&editFmt, "fmt", false, "Format the prompt file")
	editCmd.Flags().StringVar(&editModule, "module", "", "Add module dependency to go.mod")
}

func runEdit(cmd *cobra.Command, args []string) error {
	// Handle module dependency if specified and no file is provided
	if editModule != "" && len(args) == 0 {
		if err := addModuleDependency(editModule); err != nil {
			return fmt.Errorf("adding module dependency: %w", err)
		}
		// Format the output to match expected format
		parts := strings.Split(editModule, "@")
		if len(parts) == 2 {
			fmt.Printf("Added requirement: %s %s\n", parts[0], parts[1])
		} else {
			fmt.Printf("Added requirement: %s\n", editModule)
		}
		return nil
	}
	
	// Require a filename if not just editing module
	if len(args) == 0 {
		return fmt.Errorf("filename required")
	}
	
	filename := args[0]

	// Read existing prompt or create new one
	var p *prompt.Prompt
	data, err := os.ReadFile(filename)
	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("reading prompt file: %w", err)
		}
		// Create new prompt if file doesn't exist
		p = &prompt.Prompt{
			Sections: make(map[string]string),
			Defaults: make(map[string]string),
		}
	} else {
		p, err = prompt.Parse(string(data))
		if err != nil {
			return fmt.Errorf("parsing prompt: %w", err)
		}
	}

	// Apply modifications
	modified := false

	if editSetPrompt != "" {
		p.Main = editSetPrompt
		modified = true
	}

	if editSetSystemPrompt != "" {
		p.SystemPrompt = editSetSystemPrompt
		modified = true
	}

	if editSetExamples != "" {
		if p.Sections == nil {
			p.Sections = make(map[string]string)
		}
		p.Sections["examples"] = editSetExamples
		modified = true
	}

	if editSetDefaults != "" {
		// Replace all defaults
		p.Defaults = make(map[string]string)
		p.ParseDefaults(editSetDefaults)
		modified = true
	}

	if editAddDefault != "" {
		// Add or update a single default
		if p.Defaults == nil {
			p.Defaults = make(map[string]string)
		}
		if kv := strings.SplitN(editAddDefault, "=", 2); len(kv) == 2 {
			p.Defaults[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
			modified = true
		}
	}

	if editRemoveDefault != "" {
		// Remove a default by key
		if p.Defaults != nil {
			delete(p.Defaults, editRemoveDefault)
			modified = true
		}
	}

	if editAppendPrompt != "" {
		if p.Main != "" && !strings.HasSuffix(p.Main, "\n") {
			p.Main += "\n"
		}
		p.Main += editAppendPrompt
		modified = true
	}

	if editPrependPrompt != "" {
		if editPrependPrompt != "" && !strings.HasSuffix(editPrependPrompt, "\n") {
			editPrependPrompt += "\n"
		}
		p.Main = editPrependPrompt + p.Main
		modified = true
	}

	if editAddVariant != "" {
		if p.Sections == nil {
			p.Sections = make(map[string]string)
		}
		variantSection := "variant:" + editAddVariant
		if editVariantCmd != "" {
			p.Sections[variantSection] = editVariantCmd
		} else {
			p.Sections[variantSection] = ""
		}
		modified = true
	}

	if editRemoveVariant != "" {
		variantSection := "variant:" + editRemoveVariant
		delete(p.Sections, variantSection)
		modified = true
	}

	if editAddSection != "" {
		if p.Sections == nil {
			p.Sections = make(map[string]string)
		}
		p.Sections[editAddSection] = editSectionContent
		modified = true
	}

	if editRemoveSection != "" {
		delete(p.Sections, editRemoveSection)
		modified = true
	}


	// Output
	if editJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(p)
	}

	if !modified && !editPrint && !editFmt {
		// Nothing to do
		return nil
	}

	// Format the prompt back to text
	output := formatPrompt(p)

	if editPrint {
		fmt.Print(output)
		return nil
	}

	// Write back to file
	return os.WriteFile(filename, []byte(output), 0644)
}

func formatPrompt(p *prompt.Prompt) string {
	var b strings.Builder

	// Shebang
	if p.Shebang != "" {
		b.WriteString(p.Shebang)
		b.WriteString("\n")
		if p.Main != "" || p.SystemPrompt != "" || len(p.Sections) > 0 {
			b.WriteString("\n")
		}
	}

	// Main prompt
	b.WriteString(p.Main)
	if p.Main != "" && !strings.HasSuffix(p.Main, "\n") {
		b.WriteString("\n")
	}

	// Defaults
	if len(p.Defaults) > 0 {
		if p.Main != "" {
			b.WriteString("\n")
		}
		b.WriteString("-- defaults --\n")
		
		// Format as key1=value1&key2=value2
		var defaults []string
		for k, v := range p.Defaults {
			defaults = append(defaults, fmt.Sprintf("%s=%s", k, v))
		}
		b.WriteString(strings.Join(defaults, "&"))
		b.WriteString("\n")
	}

	// System prompt
	if p.SystemPrompt != "" {
		if p.Main != "" || len(p.Defaults) > 0 {
			b.WriteString("\n")
		}
		b.WriteString("-- system-prompt --\n")
		b.WriteString(p.SystemPrompt)
		if !strings.HasSuffix(p.SystemPrompt, "\n") {
			b.WriteString("\n")
		}
	}

	// Other sections
	for name, content := range p.Sections {
		if name == "system-prompt" {
			continue // Already handled
		}
		if p.Main != "" || p.SystemPrompt != "" {
			b.WriteString("\n")
		}
		b.WriteString(fmt.Sprintf("-- %s --\n", name))
		b.WriteString(content)
		if content != "" && !strings.HasSuffix(content, "\n") {
			b.WriteString("\n")
		}
	}

	return b.String()
}

func addModuleDependency(module string) error {
	// Read existing go.mod
	data, err := os.ReadFile("go.mod")
	if err != nil {
		return fmt.Errorf("reading go.mod: %w", err)
	}

	content := string(data)
	
	// Parse module and version
	parts := strings.Split(module, "@")
	if len(parts) != 2 {
		return fmt.Errorf("module must be in format: name@version")
	}
	moduleName := parts[0]
	version := parts[1]
	
	// Check if already has a require section
	requireLine := fmt.Sprintf("\t%s %s", moduleName, version)
	
	if strings.Contains(content, "require (") {
		// Insert into existing require block
		lines := strings.Split(content, "\n")
		for i, line := range lines {
			if strings.TrimSpace(line) == "require (" {
				// Find the closing ) and handle comments
				for j := i + 1; j < len(lines); j++ {
					trimmed := strings.TrimSpace(lines[j])
					if trimmed == ")" {
						// Insert before the closing )
						newLines := append(lines[:j], append([]string{requireLine}, lines[j:]...)...)
						lines = newLines
						content = strings.Join(lines, "\n")
						break
					} else if strings.HasPrefix(trimmed, "//") && strings.Contains(trimmed, "will be added here") {
						// Replace the placeholder comment
						lines[j] = requireLine
						content = strings.Join(lines, "\n")
						break
					}
				}
				break
			}
		}
	} else {
		// Add new require block
		content = strings.TrimRight(content, "\n")
		content += fmt.Sprintf("\n\nrequire (\n%s\n)\n", requireLine)
	}
	
	// Write back
	return os.WriteFile("go.mod", []byte(content), 0644)
}