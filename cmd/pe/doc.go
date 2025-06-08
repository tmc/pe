package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tmc/pe/internal/prompt"
)

var (
	docAll      bool
	docShort    bool
	docExamples bool
)

// docCmd implements the 'pe doc' command, similar to 'go doc'
func docCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "doc [prompt] [variable]",
		Short: "Show documentation for prompts and variables",
		Long: `Show documentation extracted from prompt files.

Similar to 'go doc', this command displays documentation embedded in prompts.

Examples:
  pe doc                          # List all documented prompts
  pe doc math-solver              # Show documentation for math-solver.prompt
  pe doc math-solver.EXPRESSION   # Show variable documentation
  pe doc -all                     # Show all prompts with full documentation`,
		RunE: runDoc,
	}

	cmd.Flags().BoolVar(&docAll, "all", false, "Show all documentation")
	cmd.Flags().BoolVar(&docShort, "short", false, "Show only brief descriptions")
	cmd.Flags().BoolVar(&docExamples, "examples", true, "Show examples")

	return cmd
}

func runDoc(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		// List all documented prompts
		return listDocumentedPrompts()
	}

	// Parse the argument: could be "prompt" or "prompt.VARIABLE"
	arg := args[0]
	parts := strings.Split(arg, ".")

	if len(parts) == 1 {
		// Show documentation for a prompt
		return showPromptDoc(parts[0])
	} else if len(parts) == 2 {
		// Show documentation for a specific variable
		return showVariableDoc(parts[0], parts[1])
	}

	return fmt.Errorf("invalid documentation path: %s", arg)
}

func listDocumentedPrompts() error {
	// Find all .prompt files in current directory and subdirectories
	var prompts []string

	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if strings.HasSuffix(path, ".prompt") {
			prompts = append(prompts, path)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("walking directory: %w", err)
	}

	if len(prompts) == 0 {
		fmt.Println("No prompt files found.")
		return nil
	}

	// Sort prompts by name
	sort.Strings(prompts)

	if docShort {
		// Just list the prompt files
		for _, p := range prompts {
			fmt.Println(p)
		}
		return nil
	}

	// Show brief documentation for each
	fmt.Println("PROMPTS")
	fmt.Println()

	for _, promptPath := range prompts {
		content, err := os.ReadFile(promptPath)
		if err != nil {
			continue
		}

		p, err := prompt.Parse(string(content))
		if err != nil {
			continue
		}

		// Show prompt name and summary
		name := strings.TrimSuffix(filepath.Base(promptPath), ".prompt")
		fmt.Printf("%-20s", name)

		if p.PromptSummary != "" {
			// Take first line of summary
			lines := strings.Split(p.PromptSummary, "\n")
			summary := strings.TrimSpace(lines[0])
			if len(summary) > 60 {
				summary = summary[:57] + "..."
			}
			fmt.Printf(" %s", summary)
		}
		fmt.Println()
	}

	return nil
}

func showPromptDoc(promptName string) error {
	// Try to find the prompt file
	promptPath := promptName
	if !strings.HasSuffix(promptPath, ".prompt") {
		promptPath += ".prompt"
	}

	content, err := os.ReadFile(promptPath)
	if err != nil {
		// Try looking in common directories
		for _, dir := range []string{"prompts", "examples", "."} {
			tryPath := filepath.Join(dir, promptPath)
			if content, err = os.ReadFile(tryPath); err == nil {
				promptPath = tryPath
				break
			}
		}
		if err != nil {
			return fmt.Errorf("prompt not found: %s", promptName)
		}
	}

	// Parse the prompt
	p, err := prompt.Parse(string(content))
	if err != nil {
		return fmt.Errorf("parsing prompt: %w", err)
	}

	// Display documentation
	fmt.Printf("PROMPT %s\n", promptName)
	fmt.Printf("(%s)\n\n", promptPath)

	if p.PromptSummary != "" {
		fmt.Println("SUMMARY")
		fmt.Println(indentText(p.PromptSummary, "    "))
		fmt.Println()
	}

	// Show the main prompt
	if !docAll {
		mainPrompt := p.Main
		lines := strings.Split(mainPrompt, "\n")
		if len(lines) > 5 {
			mainPrompt = strings.Join(lines[:5], "\n") + "\n    ..."
		}
		fmt.Println("PROMPT")
		fmt.Println(indentText(mainPrompt, "    "))
		fmt.Println()
	} else {
		fmt.Println("PROMPT")
		fmt.Println(indentText(p.Main, "    "))
		fmt.Println()
	}

	// Show variables
	if len(p.VariableDescriptions) > 0 || len(extractTemplateVars(p.Main)) > 0 {
		fmt.Println("VARIABLES")

		// Get all variables (from template and descriptions)
		varMap := make(map[string]bool)
		for _, v := range extractTemplateVars(p.Main) {
			varMap[v] = true
		}
		for v := range p.VariableDescriptions {
			varMap[v] = true
		}

		// Sort variables
		var vars []string
		for v := range varMap {
			vars = append(vars, v)
		}
		sort.Strings(vars)

		// Display each variable
		for _, v := range vars {
			fmt.Printf("    %s", v)
			if desc, ok := p.VariableDescriptions[v]; ok && !docShort {
				lines := strings.Split(desc, "\n")
				if len(lines) > 0 {
					fmt.Printf(" - %s", strings.TrimSpace(lines[0]))
				}
			}
			fmt.Println()
		}
		fmt.Println()
	}

	// Show defaults
	if len(p.Defaults) > 0 {
		fmt.Println("DEFAULTS")
		for k, v := range p.Defaults {
			fmt.Printf("    %s = %s\n", k, v)
		}
		fmt.Println()
	}

	// Show examples
	if docExamples && len(p.Examples) > 0 {
		fmt.Println("EXAMPLES")

		// Sort example names
		var exampleNames []string
		for name := range p.Examples {
			exampleNames = append(exampleNames, name)
		}
		sort.Strings(exampleNames)

		for _, name := range exampleNames {
			example := p.Examples[name]
			fmt.Printf("    %s\n", name)

			// Show variables for this example
			for k, v := range example {
				if k != "ideal-output" {
					lines := strings.Split(v, "\n")
					if len(lines) == 1 && len(v) < 60 {
						fmt.Printf("        %s: %s\n", k, v)
					} else {
						fmt.Printf("        %s: <%d lines>\n", k, len(lines))
					}
				}
			}
		}
		fmt.Println()
	}

	// Show available variants
	var variants []string
	for name := range p.Sections {
		if strings.HasPrefix(name, "variant:") {
			variants = append(variants, strings.TrimPrefix(name, "variant:"))
		}
	}
	if len(variants) > 0 {
		sort.Strings(variants)
		fmt.Println("VARIANTS")
		for _, v := range variants {
			fmt.Printf("    %s\n", v)
		}
		fmt.Println()
	}

	return nil
}

func showVariableDoc(promptName, variableName string) error {
	// Try to find the prompt file
	promptPath := promptName
	if !strings.HasSuffix(promptPath, ".prompt") {
		promptPath += ".prompt"
	}

	content, err := os.ReadFile(promptPath)
	if err != nil {
		// Try looking in common directories
		for _, dir := range []string{"prompts", "examples", "."} {
			tryPath := filepath.Join(dir, promptPath)
			if content, err = os.ReadFile(tryPath); err == nil {
				promptPath = tryPath
				break
			}
		}
		if err != nil {
			return fmt.Errorf("prompt not found: %s", promptName)
		}
	}

	// Parse the prompt
	p, err := prompt.Parse(string(content))
	if err != nil {
		return fmt.Errorf("parsing prompt: %w", err)
	}

	// Check if variable exists
	desc, hasDesc := p.VariableDescriptions[variableName]

	// Check if variable is used in template
	varsInTemplate := extractTemplateVars(p.Main)
	isUsed := false
	for _, v := range varsInTemplate {
		if v == variableName {
			isUsed = true
			break
		}
	}

	if !hasDesc && !isUsed {
		return fmt.Errorf("variable %s not found in %s", variableName, promptName)
	}

	// Display variable documentation
	fmt.Printf("var %s.%s\n\n", promptName, variableName)

	if hasDesc {
		fmt.Println(desc)
	} else {
		fmt.Printf("Variable %s is used in the prompt but has no documentation.\n", variableName)
		fmt.Println("\nAdd documentation with:")
		fmt.Printf("-- variable-description/%s --\n", variableName)
		fmt.Println("Description goes here")
	}

	// Show default value if exists
	if defaultVal, ok := p.Defaults[variableName]; ok {
		fmt.Printf("\nDefault: %s\n", defaultVal)
	}

	// Show example values
	if docExamples {
		var examplesShown bool
		for exName, example := range p.Examples {
			if val, ok := example[variableName]; ok {
				if !examplesShown {
					fmt.Println("\nExample values:")
					examplesShown = true
				}
				fmt.Printf("    %s: %s\n", exName, val)
			}
		}
	}

	return nil
}

func indentText(text, prefix string) string {
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		if line != "" {
			lines[i] = prefix + line
		}
	}
	return strings.Join(lines, "\n")
}
