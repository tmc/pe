//go:build ignore

package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"sigs.k8s.io/yaml"
)

// PromptVariation represents a named variation of a prompt
type PromptVariation struct {
	Name         string                 `yaml:"name" json:"name"`
	Description  string                 `yaml:"description" json:"description"`
	Base         string                 `yaml:"base" json:"base"` // Base variation to extend from
	SystemPrompt string                 `yaml:"system_prompt" json:"system_prompt"`
	UserPrompt   string                 `yaml:"user_prompt" json:"user_prompt"`
	Prefill      string                 `yaml:"prefill" json:"prefill"`
	Temperature  *float64               `yaml:"temperature" json:"temperature"`
	MaxTokens    *int                   `yaml:"max_tokens" json:"max_tokens"`
	Model        string                 `yaml:"model" json:"model"`
	Defaults     map[string]interface{} `yaml:"defaults" json:"defaults"`
	Transform    *PromptTransform       `yaml:"transform" json:"transform"`
	Examples     []Example              `yaml:"examples" json:"examples"`
}

// PromptTransform defines how to transform the base prompt
type PromptTransform struct {
	PrependSystem  string            `yaml:"prepend_system" json:"prepend_system"`
	AppendSystem   string            `yaml:"append_system" json:"append_system"`
	PrependUser    string            `yaml:"prepend_user" json:"prepend_user"`
	AppendUser     string            `yaml:"append_user" json:"append_user"`
	ReplacePattern map[string]string `yaml:"replace_pattern" json:"replace_pattern"`
	Variables      map[string]string `yaml:"variables" json:"variables"` // Variable transforms
}

// VariationRegistry manages prompt variations
type VariationRegistry struct {
	variations map[string]map[string]*PromptVariation // prompt_name -> variation_name -> variation
	plugins    []VariationPlugin
}

// VariationPlugin allows dynamic variation creation
type VariationPlugin interface {
	// MatchVariation checks if this plugin can handle a variation name
	MatchVariation(promptName, variationName string) bool

	// CreateVariation dynamically creates a variation
	CreateVariation(base *PromptCLI, variationName string) (*PromptVariation, error)

	// ListVariations returns available variations this plugin can create
	ListVariations(promptName string) []string
}

// NewVariationRegistry creates a new variation registry
func NewVariationRegistry() *VariationRegistry {
	return &VariationRegistry{
		variations: make(map[string]map[string]*PromptVariation),
		plugins:    []VariationPlugin{},
	}
}

// RegisterVariation registers a prompt variation
func (r *VariationRegistry) RegisterVariation(promptName string, variation *PromptVariation) {
	if r.variations[promptName] == nil {
		r.variations[promptName] = make(map[string]*PromptVariation)
	}
	r.variations[promptName][variation.Name] = variation
}

// RegisterPlugin registers a variation plugin
func (r *VariationRegistry) RegisterPlugin(plugin VariationPlugin) {
	r.plugins = append(r.plugins, plugin)
}

// GetVariation retrieves a variation, checking plugins if not found
func (r *VariationRegistry) GetVariation(promptName, variationName string) (*PromptVariation, error) {
	// Check registered variations
	if prompts, exists := r.variations[promptName]; exists {
		if variation, exists := prompts[variationName]; exists {
			return variation, nil
		}
	}

	// Check plugins
	for _, plugin := range r.plugins {
		if plugin.MatchVariation(promptName, variationName) {
			// Need the base prompt to create variation
			// This would be passed in a real implementation
			return nil, fmt.Errorf("plugin variations not fully implemented")
		}
	}

	return nil, fmt.Errorf("variation '%s' not found for prompt '%s'", variationName, promptName)
}

// ParsePromptWithVariations extends ParsePromptFile to handle variations
func ParsePromptWithVariations(content string, selectedVariation string) (*PromptCLI, error) {
	// Parse base prompt
	cli, err := ParsePromptFile(content)
	if err != nil {
		return nil, err
	}

	// Look for variations in frontmatter
	if cli.Metadata.Variations != nil {
		variations, ok := cli.Metadata.Variations.([]interface{})
		if ok {
			for _, v := range variations {
				varMap, ok := v.(map[string]interface{})
				if !ok {
					continue
				}

				var variation PromptVariation
				yamlData, _ := yaml.Marshal(varMap)
				if err := yaml.Unmarshal(yamlData, &variation); err != nil {
					continue
				}

				if variation.Name == selectedVariation {
					// Apply variation
					applyVariation(cli, &variation)
					break
				}
			}
		}
	}

	// Check for file-based variations
	if selectedVariation != "" && selectedVariation != "default" {
		if err := applyFileVariation(cli, selectedVariation); err == nil {
			// Successfully applied file variation
		}
	}

	return cli, nil
}

// applyVariation modifies the CLI based on a variation
func applyVariation(cli *PromptCLI, variation *PromptVariation) {
	// Apply system prompt changes
	if variation.SystemPrompt != "" {
		cli.Metadata.SystemPrompt = variation.SystemPrompt
	} else if variation.Transform != nil {
		if variation.Transform.PrependSystem != "" {
			cli.Metadata.SystemPrompt = variation.Transform.PrependSystem + "\n\n" + cli.Metadata.SystemPrompt
		}
		if variation.Transform.AppendSystem != "" {
			cli.Metadata.SystemPrompt = cli.Metadata.SystemPrompt + "\n\n" + variation.Transform.AppendSystem
		}
	}

	// Apply user prompt changes
	if variation.UserPrompt != "" {
		cli.Prompt = variation.UserPrompt
	} else if variation.Transform != nil {
		if variation.Transform.PrependUser != "" {
			cli.Prompt = variation.Transform.PrependUser + "\n\n" + cli.Prompt
		}
		if variation.Transform.AppendUser != "" {
			cli.Prompt = cli.Prompt + "\n\n" + variation.Transform.AppendUser
		}

		// Apply pattern replacements
		for pattern, replacement := range variation.Transform.ReplacePattern {
			re, err := regexp.Compile(pattern)
			if err == nil {
				cli.Prompt = re.ReplaceAllString(cli.Prompt, replacement)
			}
		}
	}

	// Apply prefill
	if variation.Prefill != "" {
		cli.Metadata.Prefill = variation.Prefill
	}

	// Apply settings
	if variation.Temperature != nil {
		cli.Metadata.Temperature = *variation.Temperature
	}
	if variation.MaxTokens != nil {
		cli.Metadata.MaxTokens = *variation.MaxTokens
	}
	if variation.Model != "" {
		cli.Metadata.Model = variation.Model
	}

	// Merge defaults
	if variation.Defaults != nil {
		if cli.Metadata.Defaults == nil {
			cli.Metadata.Defaults = make(map[string]interface{})
		}
		for k, v := range variation.Defaults {
			cli.Metadata.Defaults[k] = v
		}
	}

	// Update description
	if variation.Description != "" {
		cli.Description = fmt.Sprintf("%s (%s variation)", cli.Description, variation.Name)
	}
}

// applyFileVariation looks for variation files
func applyFileVariation(cli *PromptCLI, variationName string) error {
	// Look for files like prompt.academic.yaml or prompt_academic.yaml
	possibleFiles := []string{
		fmt.Sprintf("%s.%s.yaml", cli.Name, variationName),
		fmt.Sprintf("%s_%s.yaml", cli.Name, variationName),
		fmt.Sprintf("%s.%s.yml", cli.Name, variationName),
		fmt.Sprintf("%s_%s.yml", cli.Name, variationName),
	}

	for _, file := range possibleFiles {
		if content, err := os.ReadFile(file); err == nil {
			var variation PromptVariation
			if err := yaml.Unmarshal(content, &variation); err == nil {
				variation.Name = variationName
				applyVariation(cli, &variation)
				return nil
			}
		}
	}

	return fmt.Errorf("variation file not found")
}

// ResolvePromptPath resolves a prompt path with potential variation suffix
func ResolvePromptPath(path string) (promptPath string, variation string) {
	// Check if path ends with a known variation pattern
	base := strings.TrimSuffix(path, filepath.Ext(path))

	// Look for variation suffix after last dot or underscore
	if idx := strings.LastIndex(base, "."); idx > 0 {
		possibleVariation := base[idx+1:]
		possibleBase := base[:idx]

		// Check if base file exists
		for _, ext := range []string{".prompt", ".yaml", ".yml"} {
			if _, err := os.Stat(possibleBase + ext); err == nil {
				return possibleBase + ext, possibleVariation
			}
		}
	}

	// Check underscore pattern
	if idx := strings.LastIndex(base, "_"); idx > 0 {
		possibleVariation := base[idx+1:]
		possibleBase := base[:idx]

		// Check if base file exists
		for _, ext := range []string{".prompt", ".yaml", ".yml"} {
			if _, err := os.Stat(possibleBase + ext); err == nil {
				return possibleBase + ext, possibleVariation
			}
		}
	}

	// No variation detected
	return path, ""
}

// Built-in variation plugins

// StyleVariationPlugin creates style-based variations
type StyleVariationPlugin struct{}

func (p *StyleVariationPlugin) MatchVariation(promptName, variationName string) bool {
	styles := []string{"concise", "detailed", "academic", "casual", "formal", "technical"}
	for _, style := range styles {
		if variationName == style {
			return true
		}
	}
	return false
}

func (p *StyleVariationPlugin) CreateVariation(base *PromptCLI, variationName string) (*PromptVariation, error) {
	variation := &PromptVariation{
		Name:        variationName,
		Description: fmt.Sprintf("%s style variation", strings.Title(variationName)),
	}

	switch variationName {
	case "concise":
		variation.Transform = &PromptTransform{
			AppendUser: "\n\nBe extremely concise in your response.",
			Variables: map[string]string{
				"maxLength": "100",
			},
		}
		variation.Temperature = &[]float64{0.5}[0]

	case "detailed":
		variation.Transform = &PromptTransform{
			AppendUser: "\n\nProvide a comprehensive and detailed response with examples.",
		}
		variation.MaxTokens = &[]int{2000}[0]

	case "academic":
		variation.Transform = &PromptTransform{
			PrependSystem: "You are an academic expert. Use formal language and cite sources where appropriate.",
			AppendUser:    "\n\nStructure your response with clear sections and academic rigor.",
		}
		variation.Temperature = &[]float64{0.3}[0]

	case "casual":
		variation.Transform = &PromptTransform{
			PrependSystem: "Be friendly and conversational in your responses.",
			ReplacePattern: map[string]string{
				"please provide": "can you give me",
				"analyze":        "check out",
			},
		}
		variation.Temperature = &[]float64{0.8}[0]
	}

	return variation, nil
}

func (p *StyleVariationPlugin) ListVariations(promptName string) []string {
	return []string{"concise", "detailed", "academic", "casual", "formal", "technical"}
}

// LanguageVariationPlugin creates language-specific variations
type LanguageVariationPlugin struct{}

func (p *LanguageVariationPlugin) MatchVariation(promptName, variationName string) bool {
	// Check if variation name is a language code
	matched, _ := regexp.MatchString(`^[a-z]{2}(-[A-Z]{2})?$`, variationName)
	return matched
}

func (p *LanguageVariationPlugin) CreateVariation(base *PromptCLI, variationName string) (*PromptVariation, error) {
	languages := map[string]string{
		"es":    "Spanish",
		"fr":    "French",
		"de":    "German",
		"it":    "Italian",
		"pt":    "Portuguese",
		"ja":    "Japanese",
		"zh":    "Chinese",
		"ko":    "Korean",
		"ar":    "Arabic",
		"hi":    "Hindi",
		"ru":    "Russian",
		"es-MX": "Mexican Spanish",
		"pt-BR": "Brazilian Portuguese",
		"zh-CN": "Simplified Chinese",
		"zh-TW": "Traditional Chinese",
	}

	langName, exists := languages[variationName]
	if !exists {
		langName = variationName
	}

	variation := &PromptVariation{
		Name:        variationName,
		Description: fmt.Sprintf("%s language variation", langName),
		Transform: &PromptTransform{
			PrependSystem: fmt.Sprintf("Always respond in %s language.", langName),
			PrependUser:   fmt.Sprintf("Please respond in %s.\n\n", langName),
		},
	}

	return variation, nil
}

func (p *LanguageVariationPlugin) ListVariations(promptName string) []string {
	return []string{"es", "fr", "de", "it", "pt", "ja", "zh", "ko", "ar", "hi", "ru"}
}

// DomainVariationPlugin creates domain-specific variations
type DomainVariationPlugin struct{}

func (p *DomainVariationPlugin) MatchVariation(promptName, variationName string) bool {
	domains := []string{"medical", "legal", "finance", "engineering", "education", "marketing"}
	for _, domain := range domains {
		if variationName == domain {
			return true
		}
	}
	return false
}

func (p *DomainVariationPlugin) CreateVariation(base *PromptCLI, variationName string) (*PromptVariation, error) {
	domainPrompts := map[string]string{
		"medical":     "You are a medical professional. Use appropriate medical terminology and consider patient safety.",
		"legal":       "You are a legal expert. Provide legally sound advice with appropriate disclaimers.",
		"finance":     "You are a financial expert. Consider regulatory compliance and risk factors.",
		"engineering": "You are an engineering expert. Focus on technical accuracy and best practices.",
		"education":   "You are an educator. Explain concepts clearly and provide learning scaffolding.",
		"marketing":   "You are a marketing expert. Focus on audience engagement and brand messaging.",
	}

	systemPrompt, exists := domainPrompts[variationName]
	if !exists {
		return nil, fmt.Errorf("unknown domain: %s", variationName)
	}

	variation := &PromptVariation{
		Name:        variationName,
		Description: fmt.Sprintf("%s domain variation", strings.Title(variationName)),
		Transform: &PromptTransform{
			PrependSystem: systemPrompt,
		},
		Temperature: &[]float64{0.4}[0], // Lower temperature for domain expertise
	}

	return variation, nil
}

func (p *DomainVariationPlugin) ListVariations(promptName string) []string {
	return []string{"medical", "legal", "finance", "engineering", "education", "marketing"}
}

// RegisterBuiltinPlugins registers all built-in variation plugins
func (r *VariationRegistry) RegisterBuiltinPlugins() {
	r.RegisterPlugin(&StyleVariationPlugin{})
	r.RegisterPlugin(&LanguageVariationPlugin{})
	r.RegisterPlugin(&DomainVariationPlugin{})
}
