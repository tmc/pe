package prompt

import (
	"strings"
)

// ExtractDefaults extracts default variable values from prompt content
func ExtractDefaults(content string) map[string]string {
	defaults := make(map[string]string)
	lines := strings.Split(content, "\n")
	inDefaults := false

	for _, line := range lines {
		if strings.HasPrefix(line, "---defaults---") {
			inDefaults = true
			continue
		} else if inDefaults && strings.HasPrefix(line, "---") {
			break
		}

		if inDefaults && strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				defaults[key] = value
			}
		}
	}

	return defaults
}

// ExtractDescription extracts description from prompt comments
func ExtractDescription(content string) string {
	lines := strings.Split(content, "\n")
	var description []string

	for i, line := range lines {
		// Skip shebang
		if i == 0 && strings.HasPrefix(line, "#!") {
			continue
		}

		// Collect comment lines at the top as description
		if strings.HasPrefix(line, "# ") {
			description = append(description, strings.TrimPrefix(line, "# "))
		} else if strings.TrimSpace(line) != "" && !strings.HasPrefix(line, "#") {
			// Stop at first non-comment, non-empty line
			break
		}
	}

	return strings.Join(description, "\n")
}

// ExtractSystemPrompt extracts system prompt section from content
func ExtractSystemPrompt(content string) string {
	lines := strings.Split(content, "\n")
	inSystem := false
	var systemLines []string

	for _, line := range lines {
		if strings.HasPrefix(line, "---system---") {
			inSystem = true
			continue
		} else if inSystem && strings.HasPrefix(line, "---") {
			break
		}

		if inSystem {
			systemLines = append(systemLines, line)
		}
	}

	return strings.TrimSpace(strings.Join(systemLines, "\n"))
}

// ExtractVariables extracts template variables from the main prompt content
func ExtractVariables(content string) []string {
	// First extract just the main prompt content
	mainContent := ExtractMainContent(content)
	
	// Find variables in the main content
	return FindVariables(mainContent)
}

// ExtractMainContent extracts the main prompt content (before any sections)
func ExtractMainContent(content string) string {
	lines := strings.Split(content, "\n")
	var mainContent []string
	inSection := false

	for i, line := range lines {
		// Skip shebang
		if i == 0 && strings.HasPrefix(line, "#!") {
			continue
		}

		// Check for section start
		if strings.HasPrefix(line, "---") && strings.HasSuffix(line, "---") {
			inSection = true
			continue
		}

		// Collect main content (before any section)
		if !inSection {
			mainContent = append(mainContent, line)
		}
	}

	return strings.Join(mainContent, "\n")
}

// FindVariables finds all {{.variable}} patterns in text
func FindVariables(text string) []string {
	var vars []string
	seen := make(map[string]bool)

	// Simple parser to find {{.variable}} patterns
	for i := 0; i < len(text); i++ {
		if i+3 < len(text) && text[i:i+3] == "{{." {
			end := strings.Index(text[i:], "}}")
			if end > 0 {
				varName := text[i+3 : i+end]
				// Clean up the variable name
				varName = strings.TrimSpace(varName)
				if idx := strings.IndexAny(varName, " |"); idx > 0 {
					varName = varName[:idx]
				}
				if varName != "" && !seen[varName] {
					vars = append(vars, varName)
					seen[varName] = true
				}
				i = i + end + 1
			}
		}
	}

	return vars
}

// ExtractSections returns all section names present in the prompt
func ExtractSections(content string) []string {
	lines := strings.Split(content, "\n")
	var sections []string

	for _, line := range lines {
		if strings.HasPrefix(line, "---") && strings.HasSuffix(line, "---") {
			section := strings.TrimSuffix(strings.TrimPrefix(line, "---"), "---")
			if section != "" {
				sections = append(sections, section)
			}
		}
	}

	return sections
}

// PromptMetadata contains all extracted metadata from a prompt file
type PromptMetadata struct {
	Description  string            `json:"description,omitempty"`
	Variables    []string          `json:"variables,omitempty"`
	Defaults     map[string]string `json:"defaults,omitempty"`
	SystemPrompt string            `json:"system_prompt,omitempty"`
	Sections     []string          `json:"sections,omitempty"`
	HasShebang   bool              `json:"has_shebang"`
	Provider     string            `json:"provider,omitempty"`
}

// ExtractMetadata extracts all metadata from a prompt file
func ExtractMetadata(content string) *PromptMetadata {
	meta := &PromptMetadata{
		Description:  ExtractDescription(content),
		Variables:    ExtractVariables(content),
		Defaults:     ExtractDefaults(content),
		SystemPrompt: ExtractSystemPrompt(content),
		Sections:     ExtractSections(content),
	}

	// Check for shebang and extract provider
	lines := strings.Split(content, "\n")
	if len(lines) > 0 && strings.HasPrefix(lines[0], "#!") {
		meta.HasShebang = true
		
		// Extract provider from shebang if present
		if strings.Contains(lines[0], "--provider=") {
			parts := strings.Fields(lines[0])
			for _, part := range parts {
				if strings.HasPrefix(part, "--provider=") {
					meta.Provider = strings.TrimPrefix(part, "--provider=")
					break
				}
			}
		}
	}

	return meta
}