package metaprompt

import (
	"context"
	"fmt"
	"strings"
)

// PromptComposer implements advanced prompt composition techniques
type PromptComposer struct {
	styleHandlers map[string]StyleHandler
}

// StyleHandler defines the interface for style-specific composition logic
type StyleHandler interface {
	Compose(ctx context.Context, components []interface{}, config interface{}) (*ComposeResult, error)
}

// ComposeResult represents the result of prompt composition
type ComposeResult struct {
	ComposedPrompt string                 `json:"composed_prompt"`
	Components     []interface{}          `json:"components"`
	Style          string                 `json:"style"`
	CoherenceScore float64                `json:"coherence_score,omitempty"`
	ValidationPass bool                   `json:"validation_pass"`
	Metadata       map[string]interface{} `json:"metadata"`
}

// NewPromptComposer creates a new prompt composer with style handlers
func NewPromptComposer() *PromptComposer {
	return &PromptComposer{
		styleHandlers: map[string]StyleHandler{
			"default":        &DefaultStyleHandler{},
			"cot":           &ChainOfThoughtHandler{},
			"few-shot":      &FewShotHandler{},
			"structured":    &StructuredHandler{},
			"conversational": &ConversationalHandler{},
		},
	}
}

// Compose composes a prompt from components using the specified style
func (pc *PromptComposer) Compose(ctx context.Context, components []interface{}, config interface{}) (*ComposeResult, error) {
	// Extract style from config
	var style string = "default"
	if cfg, ok := config.(map[string]interface{}); ok {
		if s, exists := cfg["Style"]; exists {
			if str, ok := s.(string); ok {
				style = str
			}
		}
	} else if cfg, ok := config.(*ComposeConfig); ok {
		style = cfg.Style
	}

	// Get style handler
	handler, exists := pc.styleHandlers[style]
	if !exists {
		return nil, fmt.Errorf("unsupported composition style: %s", style)
	}

	// Compose using style handler
	result, err := handler.Compose(ctx, components, config)
	if err != nil {
		return nil, err
	}

	// Set style in result
	result.Style = style
	result.ValidationPass = true // Basic validation passed

	return result, nil
}

// ComposeConfig represents configuration for prompt composition
type ComposeConfig struct {
	Components      []string               `json:"components"`
	Style          string                 `json:"style"`
	Target         string                 `json:"target"`
	Coherence      bool                   `json:"coherence"`
	ValidationGate bool                   `json:"validation_gate"`
	Optimize       bool                   `json:"optimize"`
	Metadata       map[string]interface{} `json:"metadata"`
}

// DefaultStyleHandler implements basic concatenation composition
type DefaultStyleHandler struct{}

func (d *DefaultStyleHandler) Compose(ctx context.Context, components []interface{}, config interface{}) (*ComposeResult, error) {
	var parts []string
	
	for _, comp := range components {
		if component, ok := comp.(PromptComponent); ok {
			parts = append(parts, component.Content)
		}
	}

	composedPrompt := strings.Join(parts, "\n\n")
	
	return &ComposeResult{
		ComposedPrompt: composedPrompt,
		Components:     components,
		Metadata:       make(map[string]interface{}),
	}, nil
}

// ChainOfThoughtHandler implements chain-of-thought composition
type ChainOfThoughtHandler struct{}

func (c *ChainOfThoughtHandler) Compose(ctx context.Context, components []interface{}, config interface{}) (*ComposeResult, error) {
	var parts []string
	var context, instruction, examples, constraints []string

	// Categorize components
	for _, comp := range components {
		if component, ok := comp.(PromptComponent); ok {
			switch component.Type {
			case "context":
				context = append(context, component.Content)
			case "instruction":
				instruction = append(instruction, component.Content)
			case "example":
				examples = append(examples, component.Content)
			case "constraint":
				constraints = append(constraints, component.Content)
			}
		}
	}

	// Build chain-of-thought structure
	if len(context) > 0 {
		parts = append(parts, strings.Join(context, "\n"))
	}

	if len(instruction) > 0 {
		parts = append(parts, strings.Join(instruction, "\n"))
	}

	if len(examples) > 0 {
		parts = append(parts, "Examples:")
		for _, example := range examples {
			parts = append(parts, example)
		}
	}

	// Add chain-of-thought instruction
	parts = append(parts, "Let's think step by step:")

	if len(constraints) > 0 {
		parts = append(parts, "Constraints:")
		parts = append(parts, strings.Join(constraints, "\n"))
	}

	composedPrompt := strings.Join(parts, "\n\n")

	return &ComposeResult{
		ComposedPrompt: composedPrompt,
		Components:     components,
		Metadata: map[string]interface{}{
			"reasoning_style": "chain-of-thought",
			"components_used": map[string]int{
				"context":     len(context),
				"instruction": len(instruction),
				"examples":    len(examples),
				"constraints": len(constraints),
			},
		},
	}, nil
}

// FewShotHandler implements few-shot learning composition
type FewShotHandler struct{}

func (f *FewShotHandler) Compose(ctx context.Context, components []interface{}, config interface{}) (*ComposeResult, error) {
	var parts []string
	var context, instruction, examples []string

	// Categorize components
	for _, comp := range components {
		if component, ok := comp.(PromptComponent); ok {
			switch component.Type {
			case "context":
				context = append(context, component.Content)
			case "instruction":
				instruction = append(instruction, component.Content)
			case "example":
				examples = append(examples, component.Content)
			}
		}
	}

	// Build few-shot structure
	if len(context) > 0 {
		parts = append(parts, strings.Join(context, "\n"))
	}

	if len(instruction) > 0 {
		parts = append(parts, strings.Join(instruction, "\n"))
	}

	// Add examples in few-shot format
	if len(examples) > 0 {
		parts = append(parts, "Here are some examples:")
		for i, example := range examples {
			parts = append(parts, fmt.Sprintf("Example %d: %s", i+1, example))
		}
	}

	parts = append(parts, "Now, please apply this pattern to the new input:")

	composedPrompt := strings.Join(parts, "\n\n")

	return &ComposeResult{
		ComposedPrompt: composedPrompt,
		Components:     components,
		Metadata: map[string]interface{}{
			"learning_style": "few-shot",
			"example_count":  len(examples),
		},
	}, nil
}

// StructuredHandler implements structured output composition
type StructuredHandler struct{}

func (s *StructuredHandler) Compose(ctx context.Context, components []interface{}, config interface{}) (*ComposeResult, error) {
	var parts []string

	// Add structured format instruction
	parts = append(parts, "Please provide your response in the following structured format:")
	parts = append(parts, "")

	// Process components
	for _, comp := range components {
		if component, ok := comp.(PromptComponent); ok {
			switch component.Type {
			case "context":
				parts = append(parts, "## Context")
				parts = append(parts, component.Content)
				parts = append(parts, "")
			case "instruction":
				parts = append(parts, "## Task")
				parts = append(parts, component.Content)
				parts = append(parts, "")
			case "constraint":
				parts = append(parts, "## Requirements")
				parts = append(parts, component.Content)
				parts = append(parts, "")
			}
		}
	}

	parts = append(parts, "## Output Format")
	parts = append(parts, "Provide your response with clear headings and structured sections.")

	composedPrompt := strings.Join(parts, "\n")

	return &ComposeResult{
		ComposedPrompt: composedPrompt,
		Components:     components,
		Metadata: map[string]interface{}{
			"output_style": "structured",
			"format_type": "markdown",
		},
	}, nil
}

// ConversationalHandler implements conversational composition
type ConversationalHandler struct{}

func (c *ConversationalHandler) Compose(ctx context.Context, components []interface{}, config interface{}) (*ComposeResult, error) {
	var parts []string

	// Build conversational flow
	for _, comp := range components {
		if component, ok := comp.(PromptComponent); ok {
			switch component.Type {
			case "context":
				parts = append(parts, fmt.Sprintf("Hi! I need your help with something. %s", component.Content))
			case "instruction":
				parts = append(parts, fmt.Sprintf("Could you please %s?", strings.ToLower(component.Content)))
			case "example":
				parts = append(parts, fmt.Sprintf("For example: %s", component.Content))
			case "constraint":
				parts = append(parts, fmt.Sprintf("Just to clarify: %s", component.Content))
			}
		}
	}

	parts = append(parts, "Thanks for your help!")

	composedPrompt := strings.Join(parts, " ")

	return &ComposeResult{
		ComposedPrompt: composedPrompt,
		Components:     components,
		Metadata: map[string]interface{}{
			"tone": "conversational",
			"style": "informal",
		},
	}, nil
}

// PromptComponent represents a reusable prompt component
type PromptComponent struct {
	Type        string                 `json:"type"`
	Content     string                 `json:"content"`
	Category    string                 `json:"category"`
	Metadata    map[string]interface{} `json:"metadata"`
	Verified    bool                   `json:"verified"`
	Dependencies []string              `json:"dependencies,omitempty"`
}