package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"text/template"

	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"
)

// PromptComponents represents the parsed components of a prompt file
type PromptComponents struct {
	SystemPrompt string            `yaml:"system,omitempty"`
	UserPrompt   string            `yaml:"user,omitempty"`
	Messages     []Message         `yaml:"messages,omitempty"`
	Variables    map[string]string `yaml:"variables,omitempty"`
	Config       map[string]any    `yaml:"config,omitempty"`
	Metadata     map[string]any    `yaml:"metadata,omitempty"`
	RawContent   string            // For non-YAML files
}

// Message represents a conversation message
type Message struct {
	Role    string `yaml:"role"`
	Content string `yaml:"content"`
}

// catCmd returns a cobra.Command for the 'cat' subcommand
func catCmd() *cobra.Command {
	var (
		showSystem     bool
		showVariables  bool
		showComponents bool
		showMetadata   bool
		setVars        []string
		interactive    bool
		outputFormat   string
		raw            bool
	)

	cmd := &cobra.Command{
		Use:   "cat [prompt_file]",
		Short: "Display and inspect prompt files with variable substitution",
		Long: `Display the contents of prompt files with comprehensive inspection capabilities.

The cat command can read various prompt file formats including:
• Plain text prompts (.prompt, .txt)
• YAML-structured prompts (.yaml, .yml)
• JSON-structured prompts (.json)

Features:
• Variable substitution with interactive input or command-line values
• Component inspection (system prompt, user prompt, messages, metadata)
• Multiple output formats (text, yaml, json)
• Raw file display without processing

Variable Substitution:
Variables in prompt files can be specified using {{variable_name}} syntax.
Use --set key=value to provide values, or --interactive for guided input.`,
		Example: `  # Display a simple prompt file
  pe cat prompts/analyze.prompt

  # Show all components of a structured prompt
  pe cat --components prompts/chat.yaml

  # Substitute variables interactively
  pe cat --interactive prompts/template.prompt

  # Set variables via command line
  pe cat --set topic=AI --set style=formal prompts/template.prompt

  # Show only the system prompt component
  pe cat --system prompts/chat.yaml

  # Output as YAML with all metadata
  pe cat --components --format yaml prompts/complex.yaml

  # Raw display without any processing
  pe cat --raw prompts/template.prompt`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCat(args[0], catOptions{
				showSystem:     showSystem,
				showVariables:  showVariables,
				showComponents: showComponents,
				showMetadata:   showMetadata,
				setVars:        setVars,
				interactive:    interactive,
				outputFormat:   outputFormat,
				raw:            raw,
			})
		},
	}

	cmd.Flags().BoolVar(&showSystem, "system", false, "Show only the system prompt component")
	cmd.Flags().BoolVar(&showVariables, "variables", false, "Show available variables")
	cmd.Flags().BoolVar(&showComponents, "components", false, "Show all prompt components (system, user, messages, metadata)")
	cmd.Flags().BoolVar(&showMetadata, "metadata", false, "Show metadata information")
	cmd.Flags().StringSliceVar(&setVars, "set", []string{}, "Set variable values (key=value format)")
	cmd.Flags().BoolVar(&interactive, "interactive", false, "Interactively prompt for variable values")
	cmd.Flags().StringVar(&outputFormat, "format", "text", "Output format: text, yaml, json")
	cmd.Flags().BoolVar(&raw, "raw", false, "Display raw file content without processing")

	return cmd
}

// catOptions holds configuration for the cat command
type catOptions struct {
	showSystem     bool
	showVariables  bool
	showComponents bool
	showMetadata   bool
	setVars        []string
	interactive    bool
	outputFormat   string
	raw            bool
}

// runCat executes the cat command logic
func runCat(promptFile string, opts catOptions) error {
	// Check if file exists
	if _, err := os.Stat(promptFile); os.IsNotExist(err) {
		return fmt.Errorf("prompt file not found: %s", promptFile)
	}

	// Read the file
	content, err := os.ReadFile(promptFile)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// If raw mode, just output the content
	if opts.raw {
		fmt.Print(string(content))
		return nil
	}

	// Parse the prompt file
	components, err := parsePromptFileForCat(promptFile, content)
	if err != nil {
		return fmt.Errorf("failed to parse prompt file: %w", err)
	}

	// Parse variable assignments from command line
	variables, err := parseVariableAssignments(opts.setVars)
	if err != nil {
		return fmt.Errorf("failed to parse variable assignments: %w", err)
	}

	// Merge with file variables (command line takes precedence)
	if components.Variables == nil {
		components.Variables = make(map[string]string)
	}
	for k, v := range variables {
		components.Variables[k] = v
	}

	// Interactive variable input if requested
	if opts.interactive {
		if err := interactiveVariableInput(components); err != nil {
			return fmt.Errorf("interactive variable input failed: %w", err)
		}
	}

	// Handle specific display options
	if opts.showVariables {
		return displayVariables(components.Variables, opts.outputFormat)
	}

	if opts.showSystem {
		return displaySystemPrompt(components, opts.outputFormat)
	}

	if opts.showComponents {
		return displayAllComponents(components, opts.outputFormat)
	}

	if opts.showMetadata {
		return displayMetadata(components, opts.outputFormat)
	}

	// Default: display the main prompt with variable substitution
	return displayMainPrompt(components, opts.outputFormat)
}

// parsePromptFileForCat parses a prompt file and extracts its components
func parsePromptFileForCat(filename string, content []byte) (*PromptComponents, error) {
	ext := strings.ToLower(filepath.Ext(filename))
	
	switch ext {
	case ".yaml", ".yml":
		return parseYAMLPrompt(content)
	case ".json":
		return parseJSONPrompt(content)
	default:
		return parseTextPrompt(content), nil
	}
}

// parseYAMLPrompt parses a YAML-formatted prompt file
func parseYAMLPrompt(content []byte) (*PromptComponents, error) {
	var components PromptComponents
	if err := yaml.Unmarshal(content, &components); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}
	
	// Also store raw content for fallback
	components.RawContent = string(content)
	
	// If direct unmarshaling didn't work, try manual parsing
	if components.SystemPrompt == "" && components.UserPrompt == "" && len(components.Messages) == 0 {
		// Try parsing as a generic map and extract fields manually
		var data map[string]any
		if err := yaml.Unmarshal(content, &data); err == nil {
			if system, ok := data["system"].(string); ok {
				components.SystemPrompt = system
			}
			if user, ok := data["user"].(string); ok {
				components.UserPrompt = user
			}
			if messages, ok := data["messages"].([]any); ok {
				for _, msg := range messages {
					if msgMap, ok := msg.(map[string]any); ok {
						role, _ := msgMap["role"].(string)
						content, _ := msgMap["content"].(string)
						components.Messages = append(components.Messages, Message{
							Role:    role,
							Content: content,
						})
					}
				}
			}
			if variables, ok := data["variables"].(map[string]any); ok {
				if components.Variables == nil {
					components.Variables = make(map[string]string)
				}
				for k, v := range variables {
					components.Variables[k] = fmt.Sprintf("%v", v)
				}
			}
			if config, ok := data["config"].(map[string]any); ok {
				components.Config = config
			}
			if metadata, ok := data["metadata"].(map[string]any); ok {
				components.Metadata = metadata
			}
		}
	}
	
	return &components, nil
}

// parseJSONPrompt parses a JSON-formatted prompt file
func parseJSONPrompt(content []byte) (*PromptComponents, error) {
	// Convert JSON to YAML for unified processing
	var data map[string]any
	if err := yaml.Unmarshal(content, &data); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}
	
	yamlContent, err := yaml.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to convert JSON to YAML: %w", err)
	}
	
	return parseYAMLPrompt(yamlContent)
}

// parseTextPrompt parses a plain text prompt file
func parseTextPrompt(content []byte) *PromptComponents {
	text := string(content)
	
	// Extract variables from {{variable}} patterns
	variables := extractVariablesFromText(text)
	
	return &PromptComponents{
		UserPrompt: text,
		RawContent: text,
		Variables:  variables,
	}
}

// extractVariablesFromText finds all {{variable}} patterns in text
func extractVariablesFromText(text string) map[string]string {
	variables := make(map[string]string)
	re := regexp.MustCompile(`\{\{([^}]+)\}\}`)
	matches := re.FindAllStringSubmatch(text, -1)
	
	for _, match := range matches {
		if len(match) > 1 {
			varName := strings.TrimSpace(match[1])
			if _, exists := variables[varName]; !exists {
				variables[varName] = "" // Default empty value
			}
		}
	}
	
	return variables
}

// parseVariableAssignments parses --set key=value assignments
func parseVariableAssignments(assignments []string) (map[string]string, error) {
	variables := make(map[string]string)
	
	for _, assignment := range assignments {
		parts := strings.SplitN(assignment, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid variable assignment format: %s (expected key=value)", assignment)
		}
		
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		
		if key == "" {
			return nil, fmt.Errorf("empty variable name in assignment: %s", assignment)
		}
		
		variables[key] = value
	}
	
	return variables, nil
}

// interactiveVariableInput prompts for variable values interactively
func interactiveVariableInput(components *PromptComponents) error {
	if len(components.Variables) == 0 {
		fmt.Println("No variables found in prompt.")
		return nil
	}
	
	reader := bufio.NewReader(os.Stdin)
	
	// Sort variables for consistent order
	var varNames []string
	for name := range components.Variables {
		varNames = append(varNames, name)
	}
	sort.Strings(varNames)
	
	fmt.Printf("Found %d variable(s). Enter values (press Enter to keep current value):\n\n", len(varNames))
	
	for _, name := range varNames {
		currentValue := components.Variables[name]
		if currentValue != "" {
			fmt.Printf("%s [%s]: ", name, currentValue)
		} else {
			fmt.Printf("%s: ", name)
		}
		
		input, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read input for variable %s: %w", name, err)
		}
		
		input = strings.TrimSpace(input)
		if input != "" {
			components.Variables[name] = input
		}
	}
	
	return nil
}

// substituteVariables replaces {{variable}} patterns with actual values
func substituteVariables(text string, variables map[string]string) string {
	// Use standard Go template with map - the proper way
	tmpl, err := template.New("prompt").Parse(text)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: template parse error: %v\n", err)
		return text
	}
	
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, variables); err != nil {
		fmt.Fprintf(os.Stderr, "Error: template execution error: %v\n", err)
		return text
	}
	return buf.String()
}

// Display functions

func displayVariables(variables map[string]string, format string) error {
	if len(variables) == 0 {
		fmt.Println("No variables found.")
		return nil
	}
	
	switch format {
	case "yaml":
		data, err := yaml.Marshal(map[string]map[string]string{"variables": variables})
		if err != nil {
			return err
		}
		fmt.Print(string(data))
	case "json":
		data, err := yaml.Marshal(map[string]map[string]string{"variables": variables})
		if err != nil {
			return err
		}
		fmt.Print(string(data))
	default:
		fmt.Println("Variables:")
		var names []string
		for name := range variables {
			names = append(names, name)
		}
		sort.Strings(names)
		
		for _, name := range names {
			value := variables[name]
			if value == "" {
				fmt.Printf("  %s: (not set)\n", name)
			} else {
				fmt.Printf("  %s: %s\n", name, value)
			}
		}
	}
	
	return nil
}

func displaySystemPrompt(components *PromptComponents, format string) error {
	if components.SystemPrompt == "" {
		fmt.Println("No system prompt found.")
		return nil
	}
	
	systemPrompt := substituteVariables(components.SystemPrompt, components.Variables)
	
	switch format {
	case "yaml":
		data, err := yaml.Marshal(map[string]string{"system": systemPrompt})
		if err != nil {
			return err
		}
		fmt.Print(string(data))
	case "json":
		data, err := yaml.Marshal(map[string]string{"system": systemPrompt})
		if err != nil {
			return err
		}
		fmt.Print(string(data))
	default:
		fmt.Print(systemPrompt)
	}
	
	return nil
}

func displayAllComponents(components *PromptComponents, format string) error {
	// Substitute variables in all text components
	processedComponents := &PromptComponents{
		SystemPrompt: substituteVariables(components.SystemPrompt, components.Variables),
		UserPrompt:   substituteVariables(components.UserPrompt, components.Variables),
		Variables:    components.Variables,
		Config:       components.Config,
		Metadata:     components.Metadata,
	}
	
	// Process messages
	for _, msg := range components.Messages {
		processedComponents.Messages = append(processedComponents.Messages, Message{
			Role:    msg.Role,
			Content: substituteVariables(msg.Content, components.Variables),
		})
	}
	
	switch format {
	case "yaml":
		data, err := yaml.Marshal(processedComponents)
		if err != nil {
			return err
		}
		fmt.Print(string(data))
	case "json":
		data, err := yaml.Marshal(processedComponents)
		if err != nil {
			return err
		}
		fmt.Print(string(data))
	default:
		// Text format with clear sections
		if processedComponents.SystemPrompt != "" {
			fmt.Println("=== System Prompt ===")
			fmt.Println(processedComponents.SystemPrompt)
			fmt.Println()
		}
		
		if processedComponents.UserPrompt != "" {
			fmt.Println("=== User Prompt ===")
			fmt.Println(processedComponents.UserPrompt)
			fmt.Println()
		}
		
		if len(processedComponents.Messages) > 0 {
			fmt.Println("=== Messages ===")
			for i, msg := range processedComponents.Messages {
				fmt.Printf("[%d] %s: %s\n", i+1, msg.Role, msg.Content)
			}
			fmt.Println()
		}
		
		if len(processedComponents.Variables) > 0 {
			fmt.Println("=== Variables ===")
			displayVariables(processedComponents.Variables, "text")
			fmt.Println()
		}
		
		if len(processedComponents.Config) > 0 {
			fmt.Println("=== Config ===")
			for k, v := range processedComponents.Config {
				fmt.Printf("  %s: %v\n", k, v)
			}
			fmt.Println()
		}
		
		if len(processedComponents.Metadata) > 0 {
			fmt.Println("=== Metadata ===")
			for k, v := range processedComponents.Metadata {
				fmt.Printf("  %s: %v\n", k, v)
			}
		}
	}
	
	return nil
}

func displayMetadata(components *PromptComponents, format string) error {
	metadata := make(map[string]any)
	
	// Combine config and metadata
	for k, v := range components.Config {
		metadata[k] = v
	}
	for k, v := range components.Metadata {
		metadata[k] = v
	}
	
	if len(metadata) == 0 {
		fmt.Println("No metadata found.")
		return nil
	}
	
	switch format {
	case "yaml":
		data, err := yaml.Marshal(map[string]any{"metadata": metadata})
		if err != nil {
			return err
		}
		fmt.Print(string(data))
	case "json":
		data, err := yaml.Marshal(map[string]any{"metadata": metadata})
		if err != nil {
			return err
		}
		fmt.Print(string(data))
	default:
		fmt.Println("Metadata:")
		for k, v := range metadata {
			fmt.Printf("  %s: %v\n", k, v)
		}
	}
	
	return nil
}

func displayMainPrompt(components *PromptComponents, format string) error {
	// Determine the main prompt content
	var mainPrompt string
	
	if components.UserPrompt != "" {
		mainPrompt = components.UserPrompt
	} else if len(components.Messages) > 0 {
		// For conversation-style prompts, combine messages
		var parts []string
		for _, msg := range components.Messages {
			parts = append(parts, fmt.Sprintf("%s: %s", msg.Role, msg.Content))
		}
		mainPrompt = strings.Join(parts, "\n\n")
	} else if components.RawContent != "" {
		mainPrompt = components.RawContent
	} else {
		return fmt.Errorf("no prompt content found")
	}
	
	// Substitute variables
	processedPrompt := substituteVariables(mainPrompt, components.Variables)
	
	switch format {
	case "yaml":
		data, err := yaml.Marshal(map[string]string{"prompt": processedPrompt})
		if err != nil {
			return err
		}
		fmt.Print(string(data))
	case "json":
		data, err := yaml.Marshal(map[string]string{"prompt": processedPrompt})
		if err != nil {
			return err
		}
		fmt.Print(string(data))
	default:
		fmt.Print(processedPrompt)
	}
	
	return nil
}