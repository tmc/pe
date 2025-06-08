package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tmc/pe/internal/prompt"
	"gopkg.in/yaml.v3"
)

var evalPromptCmd = &cobra.Command{
	Use:   "eval-prompt [prompt-file]",
	Short: "Run evaluations defined in a prompt file",
	Long: `Run evaluations defined in the -- evals -- section of a prompt file.

The evals section should contain test cases in YAML format:

  -- evals --
  tests:
    - vars:
        input: "test input"
      assert:
        - type: contains
          value: "expected"
    - vars:
        input: "another test"
      assert:
        - type: llm_rubric
          value: "Should be concise"`,
	Args: cobra.ExactArgs(1),
	RunE: runEvalPrompt,
}

var (
	evalPromptProvider string
	evalPromptOutput   string
	evalPromptVariant  string
	evalPromptJSON     bool
)

func init() {
	evalPromptCmd.Flags().StringVar(&evalPromptProvider, "provider", "", "LLM provider to use")
	evalPromptCmd.Flags().StringVar(&evalPromptOutput, "output", "", "Output file for results")
	evalPromptCmd.Flags().StringVar(&evalPromptVariant, "variant", "", "Apply variant before running evals")
	evalPromptCmd.Flags().BoolVar(&evalPromptJSON, "json", false, "Output results as JSON")
}

type EvalConfig struct {
	Provider string     `yaml:"provider"`
	Tests    []TestCase `yaml:"tests"`
}

type TestCase struct {
	Name   string                 `yaml:"name,omitempty"`
	Vars   map[string]interface{} `yaml:"vars"`
	Assert []Assertion            `yaml:"assert"`
	Expect string                 `yaml:"expect,omitempty"`
}

type Assertion struct {
	Type      string      `yaml:"type"`
	Value     interface{} `yaml:"value"`
	Threshold float64     `yaml:"threshold,omitempty"`
}

func runEvalPrompt(cmd *cobra.Command, args []string) error {
	filename := args[0]

	// Read prompt file
	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("reading prompt file: %w", err)
	}

	// Parse prompt
	p, err := prompt.Parse(string(data))
	if err != nil {
		return fmt.Errorf("parsing prompt: %w", err)
	}

	// Apply variant if specified
	if evalPromptVariant != "" {
		p, err = p.ApplyVariant(evalPromptVariant)
		if err != nil {
			return fmt.Errorf("applying variant: %w", err)
		}
	}

	// Get evals section
	evalsContent, ok := p.Sections["evals"]
	if !ok {
		return fmt.Errorf("no evals section found in prompt file")
	}

	// Try to parse as scripttest format first
	scriptTests, err := prompt.ParseScriptTests(evalsContent)
	if err == nil && len(scriptTests) > 0 {
		// Convert scripttest to YAML format
		yamlTests := prompt.ConvertScriptTestsToYAML(scriptTests)

		// Create eval config with the converted tests
		evalConfig := EvalConfig{}
		if err := yaml.Unmarshal([]byte(yamlTests), &evalConfig); err != nil {
			return fmt.Errorf("converting scripttest format: %w", err)
		}

		return runEvalWithConfig(p, evalConfig, filename)
	}

	// Fall back to YAML format
	var evalConfig EvalConfig
	if err := yaml.Unmarshal([]byte(evalsContent), &evalConfig); err != nil {
		return fmt.Errorf("parsing evals section: %w", err)
	}

	return runEvalWithConfig(p, evalConfig, filename)
}

func runEvalWithConfig(p *prompt.Prompt, evalConfig EvalConfig, filename string) error {
	// Determine provider
	provider := evalPromptProvider
	if provider == "" {
		provider = evalConfig.Provider
	}
	if provider == "" {
		// Try to extract from shebang
		if p.Shebang != "" && strings.Contains(p.Shebang, "--provider") {
			parts := strings.Fields(p.Shebang)
			for i, part := range parts {
				if part == "--provider" && i+1 < len(parts) {
					provider = parts[i+1]
					break
				}
			}
		}
	}
	if provider == "" {
		provider = "openai:gpt-3.5-turbo" // Default
	}

	// Create evaluation config
	evalConfigYAML := generateEvalConfig(p, evalConfig, provider)

	// Write to temporary file
	tmpFile, err := os.CreateTemp("", "pe-eval-*.yaml")
	if err != nil {
		return fmt.Errorf("creating temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write([]byte(evalConfigYAML)); err != nil {
		return fmt.Errorf("writing eval config: %w", err)
	}
	tmpFile.Close()

	// Run evaluation
	evalArgs := []string{tmpFile.Name()}
	if evalPromptOutput != "" {
		evalArgs = append(evalArgs, "-o", evalPromptOutput)
	}
	if evalPromptJSON {
		evalArgs = append(evalArgs, "--json")
	}

	// Execute eval command
	evalCommand := evalCmd()
	evalCommand.SetArgs(evalArgs)
	return evalCommand.Execute()
}

func generateEvalConfig(p *prompt.Prompt, evalConfig EvalConfig, provider string) string {
	var b strings.Builder

	// Generate promptfoo-compatible config
	b.WriteString("prompts:\n")
	b.WriteString("  - id: main\n")
	b.WriteString("    raw: |\n")

	// Include system prompt if present
	if p.SystemPrompt != "" {
		b.WriteString(indent(p.SystemPrompt, "      "))
		b.WriteString("\n      \n")
		b.WriteString("      ---\n      \n")
	}

	b.WriteString(indent(p.Main, "      "))
	b.WriteString("\n\n")

	// Provider
	b.WriteString("providers:\n")
	b.WriteString(fmt.Sprintf("  - %s\n\n", provider))

	// Tests
	b.WriteString("tests:\n")
	for i, test := range evalConfig.Tests {
		if test.Name == "" {
			test.Name = fmt.Sprintf("test_%d", i+1)
		}

		b.WriteString(fmt.Sprintf("  - description: %s\n", test.Name))

		// Variables
		if len(test.Vars) > 0 {
			b.WriteString("    vars:\n")
			for k, v := range test.Vars {
				b.WriteString(fmt.Sprintf("      %s: %v\n", k, v))
			}
		}

		// Assertions
		if len(test.Assert) > 0 {
			b.WriteString("    assert:\n")
			for _, assert := range test.Assert {
				b.WriteString(fmt.Sprintf("      - type: %s\n", assert.Type))
				if assert.Value != nil {
					b.WriteString(fmt.Sprintf("        value: %v\n", assert.Value))
				}
				if assert.Threshold > 0 {
					b.WriteString(fmt.Sprintf("        threshold: %v\n", assert.Threshold))
				}
			}
		}

		// Expected output
		if test.Expect != "" {
			b.WriteString("    assert:\n")
			b.WriteString("      - type: equals\n")
			b.WriteString(fmt.Sprintf("        value: %s\n", test.Expect))
		}
	}

	return b.String()
}

func indent(text string, prefix string) string {
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		if line != "" {
			lines[i] = prefix + line
		}
	}
	return strings.Join(lines, "\n")
}
