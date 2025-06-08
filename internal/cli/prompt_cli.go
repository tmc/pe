//go:build ignore

package cli

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"regexp"
	"strings"
	"text/template"

	"sigs.k8s.io/yaml"
)

// PromptCLI converts a prompt template into a CLI tool
type PromptCLI struct {
	Name        string
	Description string
	Prompt      string
	Metadata    PromptMetadata
	Variables   []Variable
	OutputVars  []OutputVariable
	flagSet     *flag.FlagSet
}

// PromptMetadata contains CLI configuration from prompt frontmatter
type PromptMetadata struct {
	Name         string                 `yaml:"name" json:"name"`
	Description  string                 `yaml:"description" json:"description"`
	Version      string                 `yaml:"version" json:"version"`
	Author       string                 `yaml:"author" json:"author"`
	Model        string                 `yaml:"model" json:"model"`
	Temperature  float64                `yaml:"temperature" json:"temperature"`
	MaxTokens    int                    `yaml:"max_tokens" json:"max_tokens"`
	Examples     []Example              `yaml:"examples" json:"examples"`
	OutputFormat string                 `yaml:"output_format" json:"output_format"`
	Defaults     map[string]interface{} `yaml:"defaults" json:"defaults"`
	SystemPrompt string                 `yaml:"system_prompt" json:"system_prompt"`
	Prefill      string                 `yaml:"prefill" json:"prefill"`
	Variations   interface{}            `yaml:"variations" json:"variations"`
	Subcommands  []SubcommandDef        `yaml:"subcommands" json:"subcommands"`
	Variables    []Variable             `yaml:"variables" json:"variables"`
	OutputVars   []OutputVariable       `yaml:"output_vars" json:"output_vars"`
}

// Variable represents a template variable in the prompt
type Variable struct {
	Name        string      `yaml:"name" json:"name"`
	Type        string      `yaml:"type" json:"type"` // string, int, float, bool, file
	Description string      `yaml:"description" json:"description"`
	Default     interface{} `yaml:"default" json:"default"`
	Required    bool        `yaml:"required" json:"required"`
	Short       string      `yaml:"short" json:"short"`     // short flag
	Enum        []string    `yaml:"enum" json:"enum"`       // allowed values
	Pattern     string      `yaml:"pattern" json:"pattern"` // regex validation
	Min         *float64    `yaml:"min" json:"min"`
	Max         *float64    `yaml:"max" json:"max"`
}

// OutputVariable represents a variable that can be extracted from output
type OutputVariable struct {
	Name     string `yaml:"name" json:"name"`
	Type     string `yaml:"type" json:"type"`           // string, list, etc.
	Pattern  string `yaml:"pattern" json:"pattern"`     // regex to extract
	XMLTag   string `yaml:"xml_tag" json:"xml_tag"`     // XML tag to extract
	JSONPath string `yaml:"json_path" json:"json_path"` // JSON path to extract
}

// Example shows usage examples
type Example struct {
	Name        string                 `yaml:"name" json:"name"`
	Description string                 `yaml:"description" json:"description"`
	Flags       map[string]interface{} `yaml:"flags" json:"flags"`
	Output      string                 `yaml:"output" json:"output"`
}

// NewPromptCLI creates a new PromptCLI from a prompt template
func NewPromptCLI(prompt string) (*PromptCLI, error) {
	return ParsePromptFile(prompt)
}

// ParsePromptFile parses a prompt file with frontmatter
func ParsePromptFile(content string) (*PromptCLI, error) {
	// Check for frontmatter
	parts := regexp.MustCompile(`(?s)^---\n(.+?)\n---\n(.*)$`).FindStringSubmatch(content)

	cli := &PromptCLI{}

	if len(parts) == 3 {
		// Parse frontmatter
		if err := yaml.Unmarshal([]byte(parts[1]), &cli.Metadata); err != nil {
			return nil, fmt.Errorf("failed to parse frontmatter: %w", err)
		}
		cli.Prompt = parts[2]
	} else {
		// No frontmatter, use the whole content as prompt
		cli.Prompt = content
	}

	// Use metadata-defined variables if available, otherwise extract from prompt
	if len(cli.Metadata.Variables) > 0 {
		cli.Variables = cli.Metadata.Variables
	} else {
		// Extract variables from prompt
		cli.Variables = extractVariables(cli.Prompt)
	}

	// Use metadata-defined output vars if available, otherwise extract from prompt
	if len(cli.Metadata.OutputVars) > 0 {
		cli.OutputVars = cli.Metadata.OutputVars
	} else {
		// Extract output variables from prompt
		cli.OutputVars = extractOutputVariables(cli.Prompt)
	}

	// Merge with metadata-defined variables for additional configuration
	cli.mergeVariableMetadata()

	// Set defaults
	if cli.Metadata.Temperature == 0 {
		cli.Metadata.Temperature = 0.7
	}
	if cli.Metadata.Model == "" {
		cli.Metadata.Model = "gpt-4"
	}

	// Copy metadata to top-level fields
	cli.Name = cli.Metadata.Name
	if cli.Name == "" {
		cli.Name = "prompt-cli"
	}
	cli.Description = cli.Metadata.Description
	if cli.Description == "" {
		cli.Description = "Execute a prompt template"
	}

	return cli, nil
}

// ToCommand creates a cobra command from the prompt CLI (alias for BuildCommand)
func (p *PromptCLI) ToCommand() *cobra.Command {
	return p.BuildCommand()
}

// BuildCommand creates a cobra command from the prompt CLI
func (p *PromptCLI) BuildCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     p.Name,
		Short:   p.Description,
		Long:    p.generateLongDescription(),
		Example: p.generateExamples(),
		RunE: func(cmd *cobra.Command, args []string) error {
			return p.execute(cmd, args)
		},
	}

	// Add variable flags
	for _, v := range p.Variables {
		p.addVariableFlag(cmd, v)
	}

	// Add standard flags
	cmd.Flags().StringP("model", "m", p.Metadata.Model, "LLM model to use")
	cmd.Flags().Float64P("temperature", "t", p.Metadata.Temperature, "Sampling temperature")
	cmd.Flags().IntP("max-tokens", "n", p.Metadata.MaxTokens, "Maximum tokens to generate")
	cmd.Flags().StringP("output", "o", "", "Output file (default: stdout)")
	cmd.Flags().StringP("format", "f", p.Metadata.OutputFormat, "Output format template")
	cmd.Flags().BoolP("json", "j", false, "Output as JSON")
	cmd.Flags().BoolP("quiet", "q", false, "Suppress non-essential output")
	cmd.Flags().Bool("dry-run", false, "Show the prompt without executing")

	// Add completion
	p.addCompletion(cmd)

	p.cmd = cmd
	return cmd
}

// execute runs the prompt with the provided flags
func (p *PromptCLI) execute(cmd *cobra.Command, args []string) error {
	// Collect variable values
	vars := make(map[string]interface{})

	for _, v := range p.Variables {
		value, err := p.getVariableValue(cmd, v)
		if err != nil {
			return err
		}
		vars[v.Name] = value
	}

	// Render the prompt
	tmpl, err := template.New("prompt").Parse(p.Prompt)
	if err != nil {
		return fmt.Errorf("failed to parse prompt template: %w", err)
	}

	var promptBuf bytes.Buffer
	if err := tmpl.Execute(&promptBuf, vars); err != nil {
		return fmt.Errorf("failed to render prompt: %w", err)
	}

	renderedPrompt := promptBuf.String()

	// Check for dry run
	if dryRun, _ := cmd.Flags().GetBool("dry-run"); dryRun {
		fmt.Println("=== Rendered Prompt ===")
		fmt.Println(renderedPrompt)
		fmt.Println("=== Variables ===")
		for k, v := range vars {
			fmt.Printf("%s: %v\n", k, v)
		}
		return nil
	}

	// Get execution parameters
	model, _ := cmd.Flags().GetString("model")
	temperature, _ := cmd.Flags().GetFloat64("temperature")
	maxTokens, _ := cmd.Flags().GetInt("max-tokens")

	// Execute the prompt (this would call PE's inference API)
	response, err := p.executePrompt(renderedPrompt, model, temperature, maxTokens)
	if err != nil {
		return fmt.Errorf("failed to execute prompt: %w", err)
	}

	// Format output
	output, err := p.formatOutput(cmd, response)
	if err != nil {
		return fmt.Errorf("failed to format output: %w", err)
	}

	// Write output
	outputFile, _ := cmd.Flags().GetString("output")
	if outputFile != "" {
		return os.WriteFile(outputFile, []byte(output), 0644)
	}

	fmt.Print(output)
	return nil
}

// extractVariables finds all {{VarName}} or {{.VarName}} patterns in the prompt
func extractVariables(prompt string) []Variable {
	re := regexp.MustCompile(`\{\{\.?(\w+)\}\}`)
	matches := re.FindAllStringSubmatch(prompt, -1)

	seen := make(map[string]bool)
	var variables []Variable

	for _, match := range matches {
		varName := match[1]
		if !seen[varName] {
			seen[varName] = true
			variables = append(variables, Variable{
				Name:     varName,
				Type:     "string", // default type
				Required: true,     // default to required
			})
		}
	}

	return variables
}

// extractOutputVariables finds output format specifications
func extractOutputVariables(prompt string) []OutputVariable {
	var outputVars []OutputVariable

	// Look for XML tags - simplified version without backreference
	xmlRe := regexp.MustCompile(`<(\w+)>.*?</(\w+)>`)
	matches := xmlRe.FindAllStringSubmatch(prompt, -1)
	seen := make(map[string]bool)
	for _, match := range matches {
		if match[1] == match[2] && !seen[match[1]] {
			seen[match[1]] = true
			outputVars = append(outputVars, OutputVariable{
				Name:   match[1],
				XMLTag: match[1],
			})
		}
	}

	// Look for JSON structure hints
	jsonRe := regexp.MustCompile(`"(\w+)":\s*(?:".*?"|[\d.]+|true|false|null)`)
	for _, match := range jsonRe.FindAllStringSubmatch(prompt, -1) {
		outputVars = append(outputVars, OutputVariable{
			Name:     match[1],
			JSONPath: "$." + match[1],
		})
	}

	return outputVars
}

// mergeVariableMetadata merges metadata-defined variables with extracted ones
func (p *PromptCLI) mergeVariableMetadata() {
	// Create a map of existing variables
	varMap := make(map[string]*Variable)
	for i := range p.Variables {
		varMap[p.Variables[i].Name] = &p.Variables[i]
	}

	// Apply defaults from metadata
	if p.Metadata.Defaults != nil {
		for name, value := range p.Metadata.Defaults {
			if v, exists := varMap[name]; exists {
				v.Default = value
				v.Required = false
			}
		}
	}
}

// addVariableFlag adds a flag for a variable
func (p *PromptCLI) addVariableFlag(cmd *cobra.Command, v Variable) {
	flags := cmd.Flags()

	// Generate flag name (convert camelCase to kebab-case)
	flagName := toKebabCase(v.Name)

	// Add description with type info
	description := v.Description
	if description == "" {
		description = fmt.Sprintf("%s value", v.Name)
	}
	if len(v.Enum) > 0 {
		description += fmt.Sprintf(" (allowed: %s)", strings.Join(v.Enum, ", "))
	}

	switch v.Type {
	case "int":
		defaultVal := 0
		if v.Default != nil {
			defaultVal = v.Default.(int)
		}
		if v.Short != "" {
			flags.IntP(flagName, v.Short, defaultVal, description)
		} else {
			flags.Int(flagName, defaultVal, description)
		}

	case "float", "float64":
		defaultVal := 0.0
		if v.Default != nil {
			defaultVal = v.Default.(float64)
		}
		if v.Short != "" {
			flags.Float64P(flagName, v.Short, defaultVal, description)
		} else {
			flags.Float64(flagName, defaultVal, description)
		}

	case "bool":
		defaultVal := false
		if v.Default != nil {
			defaultVal = v.Default.(bool)
		}
		if v.Short != "" {
			flags.BoolP(flagName, v.Short, defaultVal, description)
		} else {
			flags.Bool(flagName, defaultVal, description)
		}

	case "file":
		defaultVal := ""
		if v.Default != nil {
			defaultVal = v.Default.(string)
		}
		if v.Short != "" {
			flags.StringP(flagName, v.Short, defaultVal, description+" (file path)")
		} else {
			flags.String(flagName, defaultVal, description+" (file path)")
		}

	default: // string
		defaultVal := ""
		if v.Default != nil {
			defaultVal = fmt.Sprintf("%v", v.Default)
		}
		if v.Short != "" {
			flags.StringP(flagName, v.Short, defaultVal, description)
		} else {
			flags.String(flagName, defaultVal, description)
		}
	}

	// Mark as required if needed
	if v.Required && v.Default == nil {
		cmd.MarkFlagRequired(flagName)
	}
}

// getVariableValue retrieves the value of a variable from flags
func (p *PromptCLI) getVariableValue(cmd *cobra.Command, v Variable) (interface{}, error) {
	flagName := toKebabCase(v.Name)
	flags := cmd.Flags()

	// Check if flag was set
	if !flags.Changed(flagName) && v.Required && v.Default == nil {
		return nil, fmt.Errorf("required flag --%s not provided", flagName)
	}

	var value interface{}
	var err error

	switch v.Type {
	case "int":
		value, err = flags.GetInt(flagName)
	case "float", "float64":
		value, err = flags.GetFloat64(flagName)
	case "bool":
		value, err = flags.GetBool(flagName)
	case "file":
		path, err := flags.GetString(flagName)
		if err != nil {
			return nil, err
		}
		if path != "" {
			content, err := os.ReadFile(path)
			if err != nil {
				return nil, fmt.Errorf("failed to read file %s: %w", path, err)
			}
			value = string(content)
		}
	default:
		value, err = flags.GetString(flagName)
	}

	if err != nil {
		return nil, err
	}

	// Validate value
	if err := p.validateValue(v, value); err != nil {
		return nil, fmt.Errorf("invalid value for --%s: %w", flagName, err)
	}

	return value, nil
}

// validateValue validates a variable value against its constraints
func (p *PromptCLI) validateValue(v Variable, value interface{}) error {
	// Enum validation
	if len(v.Enum) > 0 {
		strVal := fmt.Sprintf("%v", value)
		valid := false
		for _, allowed := range v.Enum {
			if strVal == allowed {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("must be one of: %s", strings.Join(v.Enum, ", "))
		}
	}

	// Pattern validation
	if v.Pattern != "" && v.Type == "string" {
		strVal := value.(string)
		matched, err := regexp.MatchString(v.Pattern, strVal)
		if err != nil {
			return fmt.Errorf("invalid pattern: %w", err)
		}
		if !matched {
			return fmt.Errorf("does not match pattern: %s", v.Pattern)
		}
	}

	// Range validation
	if v.Min != nil || v.Max != nil {
		var numVal float64
		switch v := value.(type) {
		case int:
			numVal = float64(v)
		case float64:
			numVal = v
		default:
			return nil // skip range validation for non-numeric types
		}

		if v.Min != nil && numVal < *v.Min {
			return fmt.Errorf("must be >= %v", *v.Min)
		}
		if v.Max != nil && numVal > *v.Max {
			return fmt.Errorf("must be <= %v", *v.Max)
		}
	}

	return nil
}

// formatOutput formats the response according to the format flag
func (p *PromptCLI) formatOutput(cmd *cobra.Command, response string) (string, error) {
	formatStr, _ := cmd.Flags().GetString("format")
	jsonOutput, _ := cmd.Flags().GetBool("json")

	if jsonOutput {
		// Extract structured data and output as JSON
		data := p.extractStructuredData(response)
		output, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			return "", err
		}
		return string(output) + "\n", nil
	}

	if formatStr != "" {
		// Parse format template
		tmpl, err := template.New("output").Parse(formatStr)
		if err != nil {
			return "", fmt.Errorf("invalid format template: %w", err)
		}

		// Extract data for template
		data := p.extractStructuredData(response)
		data["_raw"] = response // Include raw response

		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, data); err != nil {
			return "", fmt.Errorf("failed to execute format template: %w", err)
		}

		return buf.String(), nil
	}

	// Default: return raw response
	return response, nil
}

// extractStructuredData extracts structured data from the response
func (p *PromptCLI) extractStructuredData(response string) map[string]interface{} {
	data := make(map[string]interface{})

	// Extract XML tags
	for _, ov := range p.OutputVars {
		if ov.XMLTag != "" {
			re := regexp.MustCompile(fmt.Sprintf(`<%s>(.*?)</%s>`, ov.XMLTag, ov.XMLTag))
			if match := re.FindStringSubmatch(response); len(match) > 1 {
				data[ov.Name] = match[1]
			}
		}
	}

	// Try to parse as JSON
	var jsonData map[string]interface{}
	if err := json.Unmarshal([]byte(response), &jsonData); err == nil {
		// Merge JSON data
		for k, v := range jsonData {
			data[k] = v
		}
	}

	return data
}

// generateLongDescription creates detailed help text
func (p *PromptCLI) generateLongDescription() string {
	desc := p.Metadata.Description
	if desc == "" {
		desc = "Execute a prompt with customizable variables"
	}

	if p.Metadata.Author != "" {
		desc += fmt.Sprintf("\n\nAuthor: %s", p.Metadata.Author)
	}
	if p.Metadata.Version != "" {
		desc += fmt.Sprintf("\nVersion: %s", p.Metadata.Version)
	}

	return desc
}

// generateExamples creates example usage text
func (p *PromptCLI) generateExamples() string {
	if len(p.Metadata.Examples) == 0 {
		return ""
	}

	var examples []string
	for _, ex := range p.Metadata.Examples {
		example := fmt.Sprintf("  # %s\n  %s", ex.Description, p.Name)
		for flag, value := range ex.Flags {
			example += fmt.Sprintf(" --%s %s", flag, value)
		}
		if ex.Output != "" {
			example += fmt.Sprintf("\n  # Output: %s", ex.Output)
		}
		examples = append(examples, example)
	}

	return strings.Join(examples, "\n\n")
}

// addCompletion adds shell completion
func (p *PromptCLI) addCompletion(cmd *cobra.Command) {
	// Add custom completion for enum values
	for _, v := range p.Variables {
		if len(v.Enum) > 0 {
			flagName := toKebabCase(v.Name)
			cmd.RegisterFlagCompletionFunc(flagName, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
				return v.Enum, cobra.ShellCompDirectiveDefault
			})
		}
	}

	// Add model completion
	cmd.RegisterFlagCompletionFunc("model", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"gpt-4", "gpt-3.5-turbo", "claude-3", "claude-2"}, cobra.ShellCompDirectiveDefault
	})
}

// executePrompt would call PE's inference API
func (p *PromptCLI) executePrompt(prompt string, model string, temperature float64, maxTokens int) (string, error) {
	// This would integrate with PE's inference API
	// For now, return a placeholder
	return fmt.Sprintf("Response from %s model with temperature %.1f", model, temperature), nil
}

// Helper functions

func toKebabCase(s string) string {
	// Convert camelCase to kebab-case
	re := regexp.MustCompile(`([a-z])([A-Z])`)
	kebab := re.ReplaceAllString(s, `${1}-${2}`)
	return strings.ToLower(kebab)
}

// parseMetadata extracts YAML frontmatter from a prompt
func parseMetadata(content string) (PromptMetadata, string, error) {
	var meta PromptMetadata

	// Check for frontmatter
	if !strings.HasPrefix(content, "---\n") {
		return meta, content, nil
	}

	// Find end of frontmatter
	parts := strings.SplitN(content[4:], "\n---\n", 2)
	if len(parts) != 2 {
		return meta, content, nil
	}

	// Parse YAML
	if err := yaml.Unmarshal([]byte(parts[0]), &meta); err != nil {
		return meta, "", fmt.Errorf("failed to parse metadata: %w", err)
	}

	return meta, parts[1], nil
}

// validateOutputFormat validates the output format
func validateOutputFormat(format string) error {
	if format == "" {
		return nil
	}

	validFormats := []string{"json", "yaml", "markdown", "csv", "tsv", "xml"}
	for _, valid := range validFormats {
		if format == valid {
			return nil
		}
	}

	return fmt.Errorf("invalid output format: %s (must be one of: %s)", format, strings.Join(validFormats, ", "))
}

// buildExamplesHelp builds the examples section for help text
func buildExamplesHelp(cmdName string, examples []Example) string {
	if len(examples) == 0 {
		return ""
	}

	var lines []string
	lines = append(lines, "Examples:")

	for _, ex := range examples {
		lines = append(lines, fmt.Sprintf("  # %s", ex.Name))

		// Build command line
		cmd := fmt.Sprintf("  %s", cmdName)
		for name, value := range ex.Flags {
			cmd += fmt.Sprintf(" --%s=\"%v\"", name, value)
		}
		lines = append(lines, cmd)

		if ex.Output != "" {
			lines = append(lines, fmt.Sprintf("  # Output: %s", ex.Output))
		}
		lines = append(lines, "")
	}

	return strings.Join(lines, "\n")
}

// addFlagsToCommand adds all variable flags to a command
func (p *PromptCLI) addFlagsToCommand(cmd *cobra.Command) {
	for _, v := range p.Variables {
		p.addVariableFlag(cmd, v)
	}
}
