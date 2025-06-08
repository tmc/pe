//go:build ignore

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/spf13/cobra"
	// "github.com/tmc/pe/internal/cli" // disabled due to build constraints
)

var toolCmd = &cobra.Command{
	Use:   "tool",
	Short: "Convert prompts into CLI tools",
	Long: `Convert prompt templates into standalone CLI tools with proper flag handling,
help text, and output formatting.

Prompts can include:
- Template variables ({{.VarName}}) that become CLI flags
- Frontmatter metadata for configuration
- Subcommands for modifying prompts
- Output format specifications`,
}

var toolGenCmd = &cobra.Command{
	Use:   "generate [prompt-file] [output-name]",
	Short: "Generate a CLI tool from a prompt file",
	Args:  cobra.ExactArgs(2),
	Example: `  # Generate a tool from a prompt
  pe tool generate summarize.prompt summarize-tool
  
  # The generated tool can then be used:
  ./summarize-tool --text "long article..." --style bullet-points
  
  # With subcommands:
  ./summarize-tool academic --text "research paper..."`,
	RunE: generateTool,
}

var toolRunCmd = &cobra.Command{
	Use:   "run [prompt-file] [flags]",
	Short: "Run a prompt as a CLI tool directly",
	Long: `Run a prompt file as if it were a CLI tool, without generating a binary.
	
This is useful for testing prompts before converting them to standalone tools.`,
	Example: `  # Run a prompt directly
  pe tool run summarize.prompt --text "article to summarize"
  
  # With output formatting
  pe tool run extract.prompt --file data.txt --format "{{.name}}: {{.value}}"
  
  # With subcommands
  pe tool run chat.prompt prefill --greeting "Hello!"`,
	DisableFlagParsing: true, // We'll handle flags ourselves
	RunE:               runTool,
}

var (
	toolOutputDir   string
	toolBinary      bool
	toolInstallPath string
)

func init() {
	toolCmd.AddCommand(toolGenCmd)
	toolCmd.AddCommand(toolRunCmd)

	toolGenCmd.Flags().StringVarP(&toolOutputDir, "output-dir", "o", ".", "Output directory for generated tool")
	toolGenCmd.Flags().BoolVarP(&toolBinary, "binary", "b", false, "Build as binary (requires Go)")
	toolGenCmd.Flags().StringVarP(&toolInstallPath, "install", "i", "", "Install to PATH location")
}

func generateTool(cmd *cobra.Command, args []string) error {
	promptFile := args[0]
	outputName := args[1]

	// Read prompt file
	content, err := os.ReadFile(promptFile)
	if err != nil {
		return fmt.Errorf("failed to read prompt file: %w", err)
	}

	// Parse prompt
	promptCLI, err := cli.ParsePromptFile(string(content))
	if err != nil {
		return fmt.Errorf("failed to parse prompt: %w", err)
	}

	// Set name and description
	if promptCLI.Name == "" {
		promptCLI.Name = outputName
	}
	if promptCLI.Description == "" {
		promptCLI.Description = fmt.Sprintf("CLI tool generated from %s", promptFile)
	}

	// Generate the tool code
	code := generateToolCode(promptCLI, outputName, content)

	// Write output
	outputPath := filepath.Join(toolOutputDir, outputName)
	if !toolBinary {
		outputPath += ".go"
	}

	if toolBinary {
		// Write to temp file and build
		tmpFile := filepath.Join(os.TempDir(), outputName+".go")
		if err := os.WriteFile(tmpFile, []byte(code), 0644); err != nil {
			return fmt.Errorf("failed to write temp file: %w", err)
		}
		defer os.Remove(tmpFile)

		// Build binary
		if err := buildBinary(tmpFile, outputPath); err != nil {
			return fmt.Errorf("failed to build binary: %w", err)
		}

		fmt.Printf("Generated binary: %s\n", outputPath)

		// Install if requested
		if toolInstallPath != "" {
			installPath := filepath.Join(toolInstallPath, outputName)
			if err := copyFile(outputPath, installPath); err != nil {
				return fmt.Errorf("failed to install: %w", err)
			}
			if err := os.Chmod(installPath, 0755); err != nil {
				return fmt.Errorf("failed to set permissions: %w", err)
			}
			fmt.Printf("Installed to: %s\n", installPath)
		}
	} else {
		// Write Go source
		if err := os.WriteFile(outputPath, []byte(code), 0644); err != nil {
			return fmt.Errorf("failed to write output: %w", err)
		}
		fmt.Printf("Generated Go source: %s\n", outputPath)
		fmt.Printf("\nTo build: go build -o %s %s\n", outputName, outputPath)
	}

	// Show example usage
	fmt.Printf("\nExample usage:\n")
	fmt.Printf("  ./%s --help\n", outputName)

	if len(promptCLI.Variables) > 0 {
		fmt.Printf("  ./%s", outputName)
		for _, v := range promptCLI.Variables {
			if v.Default != nil {
				continue
			}
			flag := "--" + toKebabCase(v.Name)
			value := fmt.Sprintf("<%s>", v.Name)
			fmt.Printf(" %s %s", flag, value)
		}
		fmt.Println()
	}

	return nil
}

func runTool(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("prompt file required")
	}

	promptFile := args[0]

	// Read prompt file
	content, err := os.ReadFile(promptFile)
	if err != nil {
		return fmt.Errorf("failed to read prompt file: %w", err)
	}

	// Parse prompt
	promptCLI, err := cli.ParsePromptFile(string(content))
	if err != nil {
		return fmt.Errorf("failed to parse prompt: %w", err)
	}

	// Set name from filename if not specified
	if promptCLI.Name == "" {
		base := filepath.Base(promptFile)
		promptCLI.Name = strings.TrimSuffix(base, filepath.Ext(base))
	}

	// Build command
	toolCmd := promptCLI.BuildCommand()

	// Set args (skip the prompt file name)
	toolCmd.SetArgs(args[1:])

	// Execute
	return toolCmd.Execute()
}

func generateToolCode(p *cli.PromptCLI, name string, promptContent []byte) string {
	// Generate standalone Go code for the tool
	code := fmt.Sprintf(`// Generated by PE tool generator
// Source: %s

package main

import (
	"fmt"
	"os"
	
	"github.com/spf13/cobra"
	"github.com/tmc/pe/internal/cli"
)

const promptContent = %s

func main() {
	// Parse the embedded prompt
	promptCLI, err := cli.ParsePromptFile(promptContent)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %%v\n", err)
		os.Exit(1)
	}
	
	// Build and execute command
	cmd := promptCLI.BuildCommand()
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
`, p.Name, "`"+string(promptContent)+"`")

	return code
}

func buildBinary(sourceFile, outputFile string) error {
	// Build Go binary
	cmd := exec.Command("go", "build", "-o", outputFile, sourceFile)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func copyFile(src, dst string) error {
	input, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, input, 0755)
}

func toKebabCase(s string) string {
	// Simple camelCase to kebab-case conversion
	var result []rune
	for i, r := range s {
		if i > 0 && 'A' <= r && r <= 'Z' {
			result = append(result, '-')
		}
		result = append(result, unicode.ToLower(r))
	}
	return string(result)
}
