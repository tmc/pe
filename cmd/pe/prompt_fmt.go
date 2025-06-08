package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// promptFmtCmd returns the fmt command for prompt formatting
func promptFmtCmd() *cobra.Command {
	var write bool
	var check bool
	var style string
	var fix bool

	cmd := &cobra.Command{
		Use:   "fmt [files...]",
		Short: "Format prompts with consistent style",
		Long: `Format prompts according to best practices and style guidelines.

Supports different formatting styles and can automatically fix common issues.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// If no files specified, format all .txt files
			if len(args) == 0 {
				matches, _ := filepath.Glob("*.txt")
				args = matches
			}

			hasIssues := false
			for _, file := range args {
				if err := formatPromptFile(file, write, check, style, fix); err != nil {
					if check {
						hasIssues = true
						fmt.Fprintf(cmd.OutOrStderr(), "%s: %v\n", file, err)
					} else {
						return err
					}
				} else if !check {
					fmt.Fprintf(cmd.OutOrStdout(), "%s: formatted\n", file)
				}
			}

			if check && hasIssues {
				return fmt.Errorf("formatting issues found")
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&write, "write", "w", false, "Write result to file instead of stdout")
	cmd.Flags().BoolVar(&check, "check", false, "Check if files are formatted (exit with error if not)")
	cmd.Flags().StringVar(&style, "style", "anthropic", "Formatting style (anthropic, openai, standard)")
	cmd.Flags().BoolVar(&fix, "fix", false, "Automatically fix common issues")

	return cmd
}

func formatPromptFile(file string, write, check bool, style string, fix bool) error {
	content, err := os.ReadFile(file)
	if err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}

	formatted := formatPromptContent(string(content), style, fix)

	// In check mode, just verify if formatting is needed
	if check {
		if string(content) != formatted {
			return fmt.Errorf("needs formatting")
		}
		return nil
	}

	// Write formatted content
	if write {
		return os.WriteFile(file, []byte(formatted), 0644)
	}

	// Print to stdout
	fmt.Print(formatted)
	return nil
}

func formatPromptContent(content, style string, fix bool) string {
	lines := strings.Split(content, "\n")
	formatted := make([]string, 0, len(lines))

	inSystemSection := false
	inUserSection := false
	inAssistantSection := false
	inCodeBlock := false

	// For anthropic style, if no Human:/Assistant: markers, add them
	if style == "anthropic" && !strings.Contains(content, "Human:") && !strings.Contains(content, "Assistant:") {
		// Wrap entire content as a Human prompt
		formatted = append(formatted, "Human: "+strings.TrimSpace(lines[0]))
		for i := 1; i < len(lines); i++ {
			formatted = append(formatted, lines[i])
		}
		result := strings.Join(formatted, "\n")
		if !strings.HasSuffix(result, "\n") && len(result) > 0 {
			result += "\n"
		}
		return result
	}

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Handle code blocks
		if strings.HasPrefix(trimmed, "```") {
			inCodeBlock = !inCodeBlock
			formatted = append(formatted, line)
			continue
		}

		// Preserve code block content as-is
		if inCodeBlock {
			formatted = append(formatted, line)
			continue
		}

		// Format based on style
		switch style {
		case "anthropic":
			line = formatAnthropicStyle(line, &inSystemSection, &inUserSection, &inAssistantSection, fix)
		case "openai":
			line = formatOpenAIStyle(line, fix)
		default:
			line = formatStandardStyle(line, fix)
		}

		// Fix common issues
		if fix {
			line = fixCommonIssues(line, i == 0)
		}

		formatted = append(formatted, line)
	}

	result := strings.Join(formatted, "\n")

	// Ensure file ends with newline
	if !strings.HasSuffix(result, "\n") && len(result) > 0 {
		result += "\n"
	}

	return result
}

func formatAnthropicStyle(line string, inSystem, inUser, inAssistant *bool, fix bool) string {
	trimmed := strings.TrimSpace(line)

	// Handle section markers
	if strings.HasPrefix(trimmed, "Human:") {
		*inSystem = false
		*inUser = true
		*inAssistant = false
		return "Human: " + strings.TrimSpace(strings.TrimPrefix(trimmed, "Human:"))
	}

	if strings.HasPrefix(trimmed, "Assistant:") {
		*inSystem = false
		*inUser = false
		*inAssistant = true
		return "Assistant: " + strings.TrimSpace(strings.TrimPrefix(trimmed, "Assistant:"))
	}

	if strings.HasPrefix(trimmed, "System:") {
		*inSystem = true
		*inUser = false
		*inAssistant = false
		return "System: " + strings.TrimSpace(strings.TrimPrefix(trimmed, "System:"))
	}

	// Preserve empty lines between sections
	if trimmed == "" && (*inUser || *inAssistant || *inSystem) {
		return ""
	}

	// Ensure consistent spacing
	if *inUser || *inAssistant || *inSystem {
		if trimmed != "" {
			return trimmed
		}
	}

	return line
}

func formatOpenAIStyle(line string, fix bool) string {
	trimmed := strings.TrimSpace(line)

	// Format role indicators
	if strings.HasPrefix(trimmed, "system:") {
		return "system: " + strings.TrimSpace(strings.TrimPrefix(trimmed, "system:"))
	}
	if strings.HasPrefix(trimmed, "user:") {
		return "user: " + strings.TrimSpace(strings.TrimPrefix(trimmed, "user:"))
	}
	if strings.HasPrefix(trimmed, "assistant:") {
		return "assistant: " + strings.TrimSpace(strings.TrimPrefix(trimmed, "assistant:"))
	}

	return line
}

func formatStandardStyle(line string, fix bool) string {
	// Basic formatting for standard style
	return strings.TrimRight(line, " \t")
}

func fixCommonIssues(line string, isFirstLine bool) string {
	// Remove trailing whitespace
	line = strings.TrimRight(line, " \t")

	// Fix multiple consecutive spaces
	for strings.Contains(line, "  ") {
		line = strings.ReplaceAll(line, "  ", " ")
	}

	// Fix capitalization for first line or after period
	if isFirstLine && len(line) > 0 {
		// Capitalize first letter
		runes := []rune(line)
		if len(runes) > 0 && runes[0] >= 'a' && runes[0] <= 'z' {
			runes[0] = runes[0] - 32 // Convert to uppercase
			line = string(runes)
		}
	}

	// Fix multiple periods
	for strings.Contains(line, "...") && !strings.Contains(line, "....") {
		line = strings.ReplaceAll(line, "...", "…") // Use ellipsis character
	}

	// Fix smart quotes - ASCII only for now
	// TODO: Add proper Unicode smart quote handling

	return line
}
