package main

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tmc/pe/internal/prompt"
)

var getCmd = &cobra.Command{
	Use:   "get [prompt-file] [field]",
	Short: "Get information from prompt files",
	Long: `Get specific fields or all information from prompt files.

Fields:
  prompt         - The main prompt text
  system-prompt  - The system prompt
  variables      - List of template variables
  defaults       - Default values for variables
  examples       - Get examples section
  variants       - List all variants
  all            - Get all information (default)

Examples:
  # Get the main prompt
  pe get summarize.txt prompt
  
  # Get system prompt
  pe get summarize.txt system-prompt
  
  # Get variables used in the prompt
  pe get summarize.txt variables
  
  # Get all information as JSON
  pe get summarize.txt --json
  
  # Get specific variant
  pe get summarize.txt --variant academic prompt`,
	Args: cobra.RangeArgs(0, 2),
	RunE: runGet,
}

var (
	getJSON    bool
	getVariant string
	getKeys    bool
)

func init() {
	getCmd.Flags().BoolVar(&getJSON, "json", false, "Output in JSON format")
	getCmd.Flags().StringVar(&getVariant, "variant", "", "Apply variant before getting field")
	getCmd.Flags().BoolVar(&getKeys, "keys", false, "List available section keys")
}

func runGet(cmd *cobra.Command, args []string) error {
	// If no args, check for go.mod
	if len(args) == 0 {
		return runGetModule(cmd)
	}

	filename := args[0]
	field := "all"
	if len(args) > 1 {
		field = args[1]
	}

	// Read prompt file
	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("reading prompt file: %w", err)
	}

	p, err := prompt.Parse(string(data))
	if err != nil {
		return fmt.Errorf("parsing prompt: %w", err)
	}

	// Apply variant if specified
	if getVariant != "" {
		p = applyVariant(p, getVariant)
	}

	// List keys if requested
	if getKeys {
		return listKeys(p)
	}

	// Handle JSON output for all fields
	if getJSON && field == "all" {
		info := getPromptInfo(p)
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(info)
	}

	// Get specific field
	switch field {
	case "prompt":
		fmt.Print(p.Main)
		if !strings.HasSuffix(p.Main, "\n") {
			fmt.Println()
		}

	case "system-prompt":
		if p.SystemPrompt != "" {
			fmt.Print(p.SystemPrompt)
			if !strings.HasSuffix(p.SystemPrompt, "\n") {
				fmt.Println()
			}
		}

	case "variables":
		vars := extractVariables(p)
		if getJSON {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			enc.Encode(vars)
		} else {
			for _, v := range vars {
				fmt.Println(v)
			}
		}

	case "defaults":
		if getJSON {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			enc.Encode(p.Defaults)
		} else {
			for k, v := range p.Defaults {
				fmt.Printf("%s: %s\n", k, v)
			}
		}

	case "examples":
		if examples, ok := p.Sections["examples"]; ok {
			fmt.Print(examples)
			if !strings.HasSuffix(examples, "\n") {
				fmt.Println()
			}
		}

	case "variants":
		variants := getVariants(p)
		if getJSON {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			enc.Encode(variants)
		} else {
			for _, v := range variants {
				fmt.Println(v)
			}
		}

	case "all":
		// Non-JSON output for all
		fmt.Println("Main prompt:")
		fmt.Println(p.Main)

		if p.SystemPrompt != "" {
			fmt.Println("\nSystem prompt:")
			fmt.Println(p.SystemPrompt)
		}

		vars := extractVariables(p)
		if len(vars) > 0 {
			fmt.Println("\nVariables:")
			for _, v := range vars {
				fmt.Printf("  - %s\n", v)
			}
		}

		if len(p.Defaults) > 0 {
			fmt.Println("\nDefaults:")
			for k, v := range p.Defaults {
				fmt.Printf("  %s: %s\n", k, v)
			}
		}

		variants := getVariants(p)
		if len(variants) > 0 {
			fmt.Println("\nVariants:")
			for _, v := range variants {
				fmt.Printf("  - %s\n", v)
			}
		}

		// Show other sections
		for name, content := range p.Sections {
			if !strings.HasPrefix(name, "variant:") && name != "system-prompt" {
				fmt.Printf("\n%s:\n", name)
				fmt.Println(content)
			}
		}

	default:
		// Check if it's a section name
		if content, ok := p.Sections[field]; ok {
			fmt.Print(content)
			if !strings.HasSuffix(content, "\n") {
				fmt.Println()
			}
		} else {
			return fmt.Errorf("unknown field: %s", field)
		}
	}

	return nil
}

func applyVariant(p *prompt.Prompt, variant string) *prompt.Prompt {
	// Clone the prompt
	result := &prompt.Prompt{
		Shebang:      p.Shebang,
		Main:         p.Main,
		SystemPrompt: p.SystemPrompt,
		Sections:     make(map[string]string),
	}

	// Copy sections
	for k, v := range p.Sections {
		result.Sections[k] = v
	}

	// Apply variant commands
	variantKey := "variant:" + variant
	if commands, ok := p.Sections[variantKey]; ok {
		// Parse and apply variant commands
		for _, line := range strings.Split(commands, "\n") {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}

			parts := strings.Fields(line)
			if len(parts) == 0 {
				continue
			}

			switch parts[0] {
			case "extend-system-prompt":
				if len(parts) > 1 {
					text := strings.Join(parts[1:], " ")
					text = strings.Trim(text, "'\"")
					if result.SystemPrompt != "" {
						result.SystemPrompt += "\n" + text
					} else {
						result.SystemPrompt = text
					}
				}
			case "set-flag":
				// Handle flag setting if needed
			case "append-prompt":
				if len(parts) > 1 {
					text := strings.Join(parts[1:], " ")
					text = strings.Trim(text, "'\"")
					result.Main += "\n" + text
				}
			case "prepend-prompt":
				if len(parts) > 1 {
					text := strings.Join(parts[1:], " ")
					text = strings.Trim(text, "'\"")
					result.Main = text + "\n" + result.Main
				}
			}
		}
	}

	return result
}

func extractVariables(p *prompt.Prompt) []string {
	// Regular expression to match template variables
	re := regexp.MustCompile(`\{\{\.(\w+)\}\}`)

	varMap := make(map[string]bool)

	// Extract from main prompt
	matches := re.FindAllStringSubmatch(p.Main, -1)
	for _, match := range matches {
		if len(match) > 1 {
			varMap[match[1]] = true
		}
	}

	// Extract from system prompt
	matches = re.FindAllStringSubmatch(p.SystemPrompt, -1)
	for _, match := range matches {
		if len(match) > 1 {
			varMap[match[1]] = true
		}
	}

	// Extract from sections
	for _, content := range p.Sections {
		matches = re.FindAllStringSubmatch(content, -1)
		for _, match := range matches {
			if len(match) > 1 {
				varMap[match[1]] = true
			}
		}
	}

	// Convert to sorted list
	var vars []string
	for v := range varMap {
		vars = append(vars, v)
	}

	return vars
}

func getVariants(p *prompt.Prompt) []string {
	var variants []string

	for name := range p.Sections {
		if strings.HasPrefix(name, "variant:") {
			variant := strings.TrimPrefix(name, "variant:")
			variants = append(variants, variant)
		}
	}

	return variants
}

func listKeys(p *prompt.Prompt) error {
	fmt.Println("Available keys:")
	fmt.Println("  prompt")
	fmt.Println("  system-prompt")
	fmt.Println("  variables")
	fmt.Println("  variants")

	// List section names
	for name := range p.Sections {
		if !strings.HasPrefix(name, "variant:") && name != "system-prompt" {
			fmt.Printf("  %s\n", name)
		}
	}

	return nil
}

type PromptInfo struct {
	Prompt       string            `json:"prompt"`
	SystemPrompt string            `json:"systemPrompt,omitempty"`
	Variables    []string          `json:"variables,omitempty"`
	Defaults     map[string]string `json:"defaults,omitempty"`
	Variants     []string          `json:"variants,omitempty"`
	Sections     map[string]string `json:"sections,omitempty"`
	Shebang      string            `json:"shebang,omitempty"`
}

func getPromptInfo(p *prompt.Prompt) PromptInfo {
	info := PromptInfo{
		Prompt:       p.Main,
		SystemPrompt: p.SystemPrompt,
		Variables:    extractVariables(p),
		Defaults:     p.Defaults,
		Variants:     getVariants(p),
		Shebang:      p.Shebang,
	}

	// Add non-variant sections
	info.Sections = make(map[string]string)
	for name, content := range p.Sections {
		if !strings.HasPrefix(name, "variant:") && name != "system-prompt" {
			info.Sections[name] = content
		}
	}

	if len(info.Sections) == 0 {
		info.Sections = nil
	}

	return info
}

func runGetModule(cmd *cobra.Command) error {
	// Read go.mod file
	data, err := os.ReadFile("go.mod")
	if err != nil {
		return fmt.Errorf("reading go.mod: %w", err)
	}

	// Simple parsing of go.mod
	lines := strings.Split(string(data), "\n")
	inRequire := false

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Print module line
		if strings.HasPrefix(line, "module ") {
			fmt.Println(line)
			continue
		}

		// Check for require block
		if line == "require (" {
			inRequire = true
			continue
		}

		if inRequire {
			if line == ")" {
				inRequire = false
				continue
			}
			// Print require lines (skip comments)
			if line != "" && !strings.HasPrefix(line, "//") {
				fmt.Printf("require %s\n", line)
			}
		}

		// Handle single-line require
		if strings.HasPrefix(line, "require ") {
			fmt.Println(line)
		}
	}

	return nil
}
