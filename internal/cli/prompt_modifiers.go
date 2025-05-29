package cli

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"github.com/spf13/cobra"
)

// PromptModifier represents a modification that can be applied to a prompt
type PromptModifier interface {
	// Apply modifies the prompt context
	Apply(ctx *PromptContext) error
	// Description returns a description of what this modifier does
	Description() string
}

// PromptContext holds the full context for prompt execution
type PromptContext struct {
	SystemPrompt string                 // System prompt (for chat models)
	UserPrompt   string                 // Main user prompt
	Prefill      string                 // Assistant prefill text
	Messages     []Message              // Full message history
	Variables    map[string]interface{} // Template variables
	Metadata     map[string]interface{} // Additional metadata
	Temperature  float64
	MaxTokens    int
	Model        string
}

// Message represents a conversation message
type Message struct {
	Role    string `json:"role"`    // system, user, assistant
	Content string `json:"content"`
}

// SubcommandModifier wraps a prompt as a subcommand that modifies another prompt
type SubcommandModifier struct {
	Name        string
	Description string
	Modifier    PromptModifier
	Flags       []FlagDefinition
}

// FlagDefinition defines a flag for a subcommand
type FlagDefinition struct {
	Name        string
	Type        string
	Default     interface{}
	Description string
	Short       string
}

// Common modifiers

// PrefillModifier adds or modifies assistant prefill
type PrefillModifier struct {
	Template string
	Append   bool // If true, append to existing prefill
}

func (m *PrefillModifier) Apply(ctx *PromptContext) error {
	prefill, err := renderTemplate(m.Template, ctx.Variables)
	if err != nil {
		return fmt.Errorf("failed to render prefill template: %w", err)
	}
	
	if m.Append && ctx.Prefill != "" {
		ctx.Prefill = ctx.Prefill + "\n" + prefill
	} else {
		ctx.Prefill = prefill
	}
	
	return nil
}

func (m *PrefillModifier) Description() string {
	return "Set or modify assistant prefill text"
}

// SystemPromptModifier modifies the system prompt
type SystemPromptModifier struct {
	Template string
	Mode     string // replace, prepend, append
}

func (m *SystemPromptModifier) Apply(ctx *PromptContext) error {
	systemPrompt, err := renderTemplate(m.Template, ctx.Variables)
	if err != nil {
		return fmt.Errorf("failed to render system prompt template: %w", err)
	}
	
	switch m.Mode {
	case "prepend":
		ctx.SystemPrompt = systemPrompt + "\n\n" + ctx.SystemPrompt
	case "append":
		ctx.SystemPrompt = ctx.SystemPrompt + "\n\n" + systemPrompt
	default: // replace
		ctx.SystemPrompt = systemPrompt
	}
	
	return nil
}

func (m *SystemPromptModifier) Description() string {
	return fmt.Sprintf("Modify system prompt (%s mode)", m.Mode)
}

// WrapperModifier wraps the main prompt with additional context
type WrapperModifier struct {
	Before string
	After  string
}

func (m *WrapperModifier) Apply(ctx *PromptContext) error {
	before, err := renderTemplate(m.Before, ctx.Variables)
	if err != nil {
		return fmt.Errorf("failed to render before template: %w", err)
	}
	
	after, err := renderTemplate(m.After, ctx.Variables)
	if err != nil {
		return fmt.Errorf("failed to render after template: %w", err)
	}
	
	if before != "" {
		ctx.UserPrompt = before + "\n\n" + ctx.UserPrompt
	}
	if after != "" {
		ctx.UserPrompt = ctx.UserPrompt + "\n\n" + after
	}
	
	return nil
}

func (m *WrapperModifier) Description() string {
	return "Wrap prompt with additional context"
}

// ChainModifier adds context from previous interactions
type ChainModifier struct {
	Messages []Message
}

func (m *ChainModifier) Apply(ctx *PromptContext) error {
	// Prepend messages to the context
	ctx.Messages = append(m.Messages, ctx.Messages...)
	return nil
}

func (m *ChainModifier) Description() string {
	return "Add conversation history"
}

// StyleModifier applies a specific prompting style
type StyleModifier struct {
	Style string // cot, few-shot, step-by-step, etc.
}

func (m *StyleModifier) Apply(ctx *PromptContext) error {
	switch m.Style {
	case "cot", "chain-of-thought":
		ctx.UserPrompt += "\n\nLet's think step by step:"
		ctx.Prefill = "I'll work through this step-by-step.\n\n"
		
	case "step-by-step":
		ctx.UserPrompt = fmt.Sprintf(`Please complete the following task step by step:

%s

Format your response as:
Step 1: [First step]
Step 2: [Second step]
...`, ctx.UserPrompt)
		
	case "few-shot":
		// This would need examples passed in
		ctx.SystemPrompt += "\n\nYou will be shown examples before the actual task."
		
	case "structured":
		ctx.UserPrompt += "\n\nProvide your response in a clear, structured format with appropriate headings and sections."
		
	case "concise":
		ctx.SystemPrompt += "\n\nBe concise and direct in your responses. Avoid unnecessary elaboration."
		
	case "detailed":
		ctx.SystemPrompt += "\n\nProvide comprehensive, detailed responses with thorough explanations."
		
	case "code":
		ctx.SystemPrompt += "\n\nYou are an expert programmer. Format code examples properly and include comments."
		ctx.UserPrompt += "\n\nProvide complete, working code with explanations."
		
	default:
		return fmt.Errorf("unknown style: %s", m.Style)
	}
	
	return nil
}

func (m *StyleModifier) Description() string {
	return fmt.Sprintf("Apply %s prompting style", m.Style)
}

// ConstraintModifier adds constraints to the output
type ConstraintModifier struct {
	MaxLength   int
	Format      string
	MustInclude []string
	MustExclude []string
}

func (m *ConstraintModifier) Apply(ctx *PromptContext) error {
	var constraints []string
	
	if m.MaxLength > 0 {
		constraints = append(constraints, fmt.Sprintf("- Maximum length: %d characters", m.MaxLength))
	}
	
	if m.Format != "" {
		constraints = append(constraints, fmt.Sprintf("- Format: %s", m.Format))
	}
	
	if len(m.MustInclude) > 0 {
		constraints = append(constraints, fmt.Sprintf("- Must include: %s", strings.Join(m.MustInclude, ", ")))
	}
	
	if len(m.MustExclude) > 0 {
		constraints = append(constraints, fmt.Sprintf("- Must not include: %s", strings.Join(m.MustExclude, ", ")))
	}
	
	if len(constraints) > 0 {
		ctx.UserPrompt += fmt.Sprintf("\n\nConstraints:\n%s", strings.Join(constraints, "\n"))
	}
	
	return nil
}

func (m *ConstraintModifier) Description() string {
	return "Add output constraints"
}

// PersonaModifier sets a specific persona for the assistant
type PersonaModifier struct {
	Persona     string
	Expertise   []string
	Tone        string
	Perspective string
}

func (m *PersonaModifier) Apply(ctx *PromptContext) error {
	persona := fmt.Sprintf("You are %s", m.Persona)
	
	if len(m.Expertise) > 0 {
		persona += fmt.Sprintf(" with expertise in %s", strings.Join(m.Expertise, ", "))
	}
	
	if m.Tone != "" {
		persona += fmt.Sprintf(". Maintain a %s tone", m.Tone)
	}
	
	if m.Perspective != "" {
		persona += fmt.Sprintf(". Approach from a %s perspective", m.Perspective)
	}
	
	persona += "."
	
	// Prepend to system prompt
	if ctx.SystemPrompt != "" {
		ctx.SystemPrompt = persona + "\n\n" + ctx.SystemPrompt
	} else {
		ctx.SystemPrompt = persona
	}
	
	return nil
}

func (m *PersonaModifier) Description() string {
	return fmt.Sprintf("Set persona: %s", m.Persona)
}

// ExtendPromptCLI extends PromptCLI with subcommand support
func (p *PromptCLI) AddSubcommands(subcommands []SubcommandDef) {
	for _, sub := range subcommands {
		p.addSubcommand(sub)
	}
}

// SubcommandDef defines a subcommand in the metadata
type SubcommandDef struct {
	Name        string                 `yaml:"name"`
	Description string                 `yaml:"description"`
	Type        string                 `yaml:"type"` // prefill, system, wrap, style, etc.
	Config      map[string]interface{} `yaml:"config"`
	Flags       []FlagDefinition       `yaml:"flags"`
}

// addSubcommand adds a subcommand to the CLI
func (p *PromptCLI) addSubcommand(def SubcommandDef) {
	subCmd := &cobra.Command{
		Use:   def.Name,
		Short: def.Description,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Create modifier based on type
			modifier, err := createModifier(def.Type, def.Config, cmd)
			if err != nil {
				return err
			}
			
			// Apply modifier and execute
			return p.executeWithModifier(cmd, args, modifier)
		},
	}
	
	// Add subcommand-specific flags
	for _, flag := range def.Flags {
		addFlagToCommand(subCmd, flag)
	}
	
	// Inherit parent flags
	p.cmd.AddCommand(subCmd)
}

// createModifier creates a modifier based on type and config
func createModifier(modType string, config map[string]interface{}, cmd *cobra.Command) (PromptModifier, error) {
	switch modType {
	case "prefill":
		template, _ := config["template"].(string)
		append, _ := config["append"].(bool)
		return &PrefillModifier{Template: template, Append: append}, nil
		
	case "system":
		template, _ := config["template"].(string)
		mode, _ := config["mode"].(string)
		if mode == "" {
			mode = "replace"
		}
		return &SystemPromptModifier{Template: template, Mode: mode}, nil
		
	case "wrap":
		before, _ := config["before"].(string)
		after, _ := config["after"].(string)
		return &WrapperModifier{Before: before, After: after}, nil
		
	case "style":
		style, _ := cmd.Flags().GetString("style")
		if style == "" {
			style, _ = config["style"].(string)
		}
		return &StyleModifier{Style: style}, nil
		
	case "persona":
		persona, _ := config["persona"].(string)
		tone, _ := config["tone"].(string)
		return &PersonaModifier{Persona: persona, Tone: tone}, nil
		
	default:
		return nil, fmt.Errorf("unknown modifier type: %s", modType)
	}
}

// executeWithModifier executes the prompt with a modifier applied
func (p *PromptCLI) executeWithModifier(cmd *cobra.Command, args []string, modifier PromptModifier) error {
	// Build context from parent command
	ctx := &PromptContext{
		SystemPrompt: p.Metadata.SystemPrompt,
		Variables:    make(map[string]interface{}),
		Temperature:  p.Metadata.Temperature,
		MaxTokens:    p.Metadata.MaxTokens,
		Model:        p.Metadata.Model,
	}
	
	// Collect variables from parent flags
	parentCmd := cmd.Parent()
	for _, v := range p.Variables {
		value, err := p.getVariableValue(parentCmd, v)
		if err != nil {
			return err
		}
		ctx.Variables[v.Name] = value
	}
	
	// Render the base prompt
	tmpl, err := template.New("prompt").Parse(p.Prompt)
	if err != nil {
		return err
	}
	
	var promptBuf bytes.Buffer
	if err := tmpl.Execute(&promptBuf, ctx.Variables); err != nil {
		return err
	}
	ctx.UserPrompt = promptBuf.String()
	
	// Apply the modifier
	if err := modifier.Apply(ctx); err != nil {
		return err
	}
	
	// Execute the modified prompt
	return p.executeContext(cmd, ctx)
}

// executeContext executes a prompt context
func (p *PromptCLI) executeContext(cmd *cobra.Command, ctx *PromptContext) error {
	// Build the final prompt based on the context
	var finalPrompt string
	
	if len(ctx.Messages) > 0 {
		// Use message format
		// This would be formatted for the specific provider
		finalPrompt = formatMessages(ctx)
	} else {
		// Simple format
		if ctx.SystemPrompt != "" {
			finalPrompt = fmt.Sprintf("System: %s\n\nUser: %s", ctx.SystemPrompt, ctx.UserPrompt)
		} else {
			finalPrompt = ctx.UserPrompt
		}
		
		if ctx.Prefill != "" {
			finalPrompt += fmt.Sprintf("\n\nAssistant: %s", ctx.Prefill)
		}
	}
	
	// Check for dry run
	if dryRun, _ := cmd.Flags().GetBool("dry-run"); dryRun {
		fmt.Println("=== Final Prompt ===")
		fmt.Println(finalPrompt)
		fmt.Println("\n=== Context ===")
		fmt.Printf("Model: %s\n", ctx.Model)
		fmt.Printf("Temperature: %.2f\n", ctx.Temperature)
		fmt.Printf("Max Tokens: %d\n", ctx.MaxTokens)
		return nil
	}
	
	// Execute the prompt
	response, err := p.executePrompt(finalPrompt, ctx.Model, ctx.Temperature, ctx.MaxTokens)
	if err != nil {
		return err
	}
	
	// Format and output
	output, err := p.formatOutput(cmd, response)
	if err != nil {
		return err
	}
	
	fmt.Print(output)
	return nil
}

// Helper functions

func renderTemplate(tmpl string, vars map[string]interface{}) (string, error) {
	if tmpl == "" {
		return "", nil
	}
	
	t, err := template.New("modifier").Parse(tmpl)
	if err != nil {
		return "", err
	}
	
	var buf bytes.Buffer
	if err := t.Execute(&buf, vars); err != nil {
		return "", err
	}
	
	return buf.String(), nil
}

func formatMessages(ctx *PromptContext) string {
	// This would format messages according to the provider's requirements
	// For now, simple format
	var parts []string
	
	if ctx.SystemPrompt != "" {
		parts = append(parts, fmt.Sprintf("System: %s", ctx.SystemPrompt))
	}
	
	for _, msg := range ctx.Messages {
		parts = append(parts, fmt.Sprintf("%s: %s", strings.Title(msg.Role), msg.Content))
	}
	
	parts = append(parts, fmt.Sprintf("User: %s", ctx.UserPrompt))
	
	if ctx.Prefill != "" {
		parts = append(parts, fmt.Sprintf("Assistant: %s", ctx.Prefill))
	}
	
	return strings.Join(parts, "\n\n")
}

func addFlagToCommand(cmd *cobra.Command, flag FlagDefinition) {
	switch flag.Type {
	case "string":
		if flag.Short != "" {
			cmd.Flags().StringP(flag.Name, flag.Short, flag.Default.(string), flag.Description)
		} else {
			cmd.Flags().String(flag.Name, flag.Default.(string), flag.Description)
		}
	case "bool":
		if flag.Short != "" {
			cmd.Flags().BoolP(flag.Name, flag.Short, flag.Default.(bool), flag.Description)
		} else {
			cmd.Flags().Bool(flag.Name, flag.Default.(bool), flag.Description)
		}
	case "int":
		if flag.Short != "" {
			cmd.Flags().IntP(flag.Name, flag.Short, flag.Default.(int), flag.Description)
		} else {
			cmd.Flags().Int(flag.Name, flag.Default.(int), flag.Description)
		}
	}
}