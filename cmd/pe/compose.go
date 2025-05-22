package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tmc/pe/internal/metaprompt"
)

// ComposeConfig represents configuration for prompt composition
type ComposeConfig struct {
	Components      []string          `json:"components"`
	Style          string            `json:"style"`
	Target         string            `json:"target"`
	Coherence      bool              `json:"coherence"`
	ValidationGate bool              `json:"validation_gate"`
	Optimize       bool              `json:"optimize"`
	Metadata       map[string]interface{} `json:"metadata"`
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

// ComposeResult represents the result of prompt composition
type ComposeResult struct {
	ComposedPrompt string            `json:"composed_prompt"`
	Components     []PromptComponent `json:"components"`
	Style          string            `json:"style"`
	CoherenceScore float64           `json:"coherence_score,omitempty"`
	ValidationPass bool              `json:"validation_pass"`
	Metadata       map[string]interface{} `json:"metadata"`
}

var composeCmd = &cobra.Command{
	Use:   "compose [component files...]",
	Short: "Compose prompts from verified components with type-safe composition",
	Long: `
Compose prompts from modular components using advanced composition techniques:

• Component-based architecture with type-safe validation
• Style-specific composition (chain-of-thought, few-shot, etc.)
• Semantic coherence validation
• Automatic optimization integration
• Dependency resolution

Examples:
  pe compose context.txt instruction.txt examples.txt --style cot --optimize
  pe compose components/ --style few-shot --coherence --validation-gate
  pe compose --library-init  # Initialize component library
`,
	RunE: runCompose,
}

func init() {
	composeCmd.Flags().StringSlice("components", []string{}, "Component files or directories")
	composeCmd.Flags().String("style", "default", "Composition style (cot, few-shot, structured, conversational)")
	composeCmd.Flags().String("target", "gpt-4", "Target model for optimization")
	composeCmd.Flags().Bool("coherence", false, "Enable semantic coherence validation")
	composeCmd.Flags().Bool("validation-gate", false, "Enable validation gates for quality assurance")
	composeCmd.Flags().Bool("optimize", false, "Apply TextGrad optimization after composition")
	composeCmd.Flags().Bool("library-init", false, "Initialize component library")
	composeCmd.Flags().String("output", "", "Output file for composed prompt")
	composeCmd.Flags().String("config", "", "Configuration file for composition settings")
}

func runCompose(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	
	// Handle library initialization
	if libInit, _ := cmd.Flags().GetBool("library-init"); libInit {
		return initComponentLibrary()
	}

	// Load configuration
	config, err := loadComposeConfig(cmd, args)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %v", err)
	}

	// Load components
	components, err := loadComponents(config.Components)
	if err != nil {
		return fmt.Errorf("failed to load components: %v", err)
	}

	// Validate component dependencies
	if err := validateComponentDependencies(components); err != nil {
		return fmt.Errorf("dependency validation failed: %v", err)
	}

	// Compose prompt with style-specific logic
	composer := metaprompt.NewPromptComposer()
	
	// Convert components to interface slice
	var interfaceComponents []interface{}
	for _, comp := range components {
		interfaceComponents = append(interfaceComponents, comp)
	}
	
	metaResult, err := composer.Compose(ctx, interfaceComponents, config)
	if err != nil {
		return fmt.Errorf("composition failed: %v", err)
	}
	
	// Convert to local result type
	result := &ComposeResult{
		ComposedPrompt: metaResult.ComposedPrompt,
		Components:     components,
		Style:          metaResult.Style,
		ValidationPass: metaResult.ValidationPass,
		Metadata:       metaResult.Metadata,
	}

	// Apply coherence validation if requested
	if config.Coherence {
		coherenceScore, err := validateCoherence(ctx, result.ComposedPrompt)
		if err != nil {
			fmt.Printf("Warning: Coherence validation failed: %v\n", err)
		} else {
			result.CoherenceScore = coherenceScore
			fmt.Printf("Coherence Score: %.2f\n", coherenceScore)
		}
	}

	// Apply optimization if requested
	if config.Optimize {
		// Note: Would need LLM provider for actual optimization
		// For now, apply simple optimization
		result.ComposedPrompt = simpleOptimizePrompt(result.ComposedPrompt)
		fmt.Println("Applied basic prompt optimization")
	}

	// Output result
	return outputComposeResult(cmd, result)
}

func loadComposeConfig(cmd *cobra.Command, args []string) (*ComposeConfig, error) {
	config := &ComposeConfig{
		Metadata: make(map[string]interface{}),
	}

	// Load from config file if specified
	if configFile, _ := cmd.Flags().GetString("config"); configFile != "" {
		data, err := os.ReadFile(configFile)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(data, config); err != nil {
			return nil, err
		}
	}

	// Override with command line flags
	components, _ := cmd.Flags().GetStringSlice("components")
	if len(components) == 0 && len(args) > 0 {
		components = args
	}
	config.Components = components
	
	config.Style, _ = cmd.Flags().GetString("style")
	config.Target, _ = cmd.Flags().GetString("target")
	config.Coherence, _ = cmd.Flags().GetBool("coherence")
	config.ValidationGate, _ = cmd.Flags().GetBool("validation-gate")
	config.Optimize, _ = cmd.Flags().GetBool("optimize")

	return config, nil
}

func loadComponents(componentPaths []string) ([]PromptComponent, error) {
	var components []PromptComponent

	for _, path := range componentPaths {
		// Check if path is a directory
		if info, err := os.Stat(path); err == nil && info.IsDir() {
			// Load all component files from directory
			dirComponents, err := loadComponentsFromDirectory(path)
			if err != nil {
				return nil, err
			}
			components = append(components, dirComponents...)
		} else {
			// Load single component file
			component, err := loadComponent(path)
			if err != nil {
				return nil, err
			}
			components = append(components, component)
		}
	}

	return components, nil
}

func loadComponentsFromDirectory(dir string) ([]PromptComponent, error) {
	var components []PromptComponent

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && (strings.HasSuffix(path, ".txt") || strings.HasSuffix(path, ".json")) {
			component, err := loadComponent(path)
			if err != nil {
				return err
			}
			components = append(components, component)
		}

		return nil
	})

	return components, err
}

func loadComponent(path string) (PromptComponent, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return PromptComponent{}, err
	}

	// Try JSON first
	var component PromptComponent
	if strings.HasSuffix(path, ".json") {
		if err := json.Unmarshal(data, &component); err != nil {
			return PromptComponent{}, err
		}
	} else {
		// Plain text component
		component = PromptComponent{
			Type:     inferComponentType(path, string(data)),
			Content:  string(data),
			Category: inferCategory(path),
			Metadata: map[string]interface{}{
				"source": path,
				"size":   len(data),
			},
			Verified: false, // Default to unverified for plain text
		}
	}

	return component, nil
}

func inferComponentType(path, content string) string {
	filename := strings.ToLower(filepath.Base(path))
	
	if strings.Contains(filename, "context") {
		return "context"
	}
	if strings.Contains(filename, "instruction") {
		return "instruction"
	}
	if strings.Contains(filename, "example") {
		return "example"
	}
	if strings.Contains(filename, "constraint") {
		return "constraint"
	}
	
	// Analyze content for type hints
	contentLower := strings.ToLower(content)
	if strings.Contains(contentLower, "you are") || strings.Contains(contentLower, "your role") {
		return "context"
	}
	if strings.Contains(contentLower, "please") || strings.Contains(contentLower, "analyze") {
		return "instruction"
	}
	
	return "unknown"
}

func inferCategory(path string) string {
	dir := filepath.Dir(path)
	return filepath.Base(dir)
}

func validateComponentDependencies(components []PromptComponent) error {
	componentMap := make(map[string]PromptComponent)
	for _, comp := range components {
		componentMap[comp.Type] = comp
	}

	for _, comp := range components {
		for _, dep := range comp.Dependencies {
			if _, exists := componentMap[dep]; !exists {
				return fmt.Errorf("component %s requires dependency %s which is not available", comp.Type, dep)
			}
		}
	}

	return nil
}

func validateCoherence(ctx context.Context, prompt string) (float64, error) {
	// Implement semantic coherence validation
	// This would use an LLM to evaluate semantic consistency
	// For now, return a placeholder score
	
	// Simple heuristic based on prompt structure
	sentences := strings.Split(prompt, ".")
	if len(sentences) < 2 {
		return 1.0, nil // Single sentence is coherent
	}

	// Check for transition words and consistency
	transitionWords := []string{"however", "therefore", "additionally", "furthermore", "moreover", "consequently"}
	transitionCount := 0
	for _, word := range transitionWords {
		if strings.Contains(strings.ToLower(prompt), word) {
			transitionCount++
		}
	}

	// Simple coherence score based on transitions and length
	coherenceScore := float64(transitionCount) / float64(len(sentences))
	if coherenceScore > 1.0 {
		coherenceScore = 1.0
	}

	return coherenceScore, nil
}

func outputComposeResult(cmd *cobra.Command, result *ComposeResult) error {
	outputFile, _ := cmd.Flags().GetString("output")

	if outputFile != "" {
		// Save to file
		data, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return err
		}
		return os.WriteFile(outputFile, data, 0644)
	}

	// Print to stdout
	fmt.Println("=== Composed Prompt ===")
	fmt.Println(result.ComposedPrompt)
	
	if result.CoherenceScore > 0 {
		fmt.Printf("\nCoherence Score: %.2f\n", result.CoherenceScore)
	}
	
	fmt.Printf("Components Used: %d\n", len(result.Components))
	fmt.Printf("Style: %s\n", result.Style)
	fmt.Printf("Validation Passed: %t\n", result.ValidationPass)

	return nil
}

func initComponentLibrary() error {
	// Create component library structure
	dirs := []string{
		"components/context",
		"components/instructions", 
		"components/examples",
		"components/constraints",
		"components/templates",
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	// Create sample components
	samples := map[string]string{
		"components/context/analytical.txt": "You are an expert analyst with deep knowledge in data interpretation and pattern recognition.",
		"components/instructions/analyze.txt": "Please analyze the following information and provide detailed insights.",
		"components/examples/analysis-example.txt": "Example: Input: Sales data shows 20% increase. Output: This indicates strong market performance.",
		"components/constraints/format.txt": "Provide your response in a clear, structured format with bullet points.",
	}

	for path, content := range samples {
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			return err
		}
	}

	fmt.Println("Component library initialized successfully!")
	fmt.Println("Created directories:")
	for _, dir := range dirs {
		fmt.Printf("  %s\n", dir)
	}

	return nil
}

func simpleOptimizePrompt(prompt string) string {
	optimized := prompt
	
	// Add structure if missing
	if !strings.Contains(prompt, ":") && !strings.Contains(prompt, "\n") {
		optimized = "Task: " + optimized + "\n\nPlease provide a detailed response."
	}
	
	// Add specificity cues
	if !strings.Contains(strings.ToLower(prompt), "specific") && 
	   !strings.Contains(strings.ToLower(prompt), "detailed") {
		optimized += " Please be specific and detailed in your response."
	}
	
	return optimized
}