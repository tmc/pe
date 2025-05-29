package prompt

import (
	"bufio"
	"fmt"
	"strings"
)

// Prompt represents a parsed prompt file
type Prompt struct {
	Shebang      string            // #!/usr/bin/env pe run --flags
	Main         string            // The main prompt text
	SystemPrompt string            // Optional system prompt
	Sections     map[string]string // Named sections like variants, tests, etc.
	Flags        map[string]string // Flags from shebang
	Defaults     map[string]string // Default values for variables
}

// Parse parses a prompt file in txtar-inspired format
func Parse(content string) (*Prompt, error) {
	p := &Prompt{
		Sections: make(map[string]string),
		Flags:    make(map[string]string),
		Defaults: make(map[string]string),
	}

	scanner := bufio.NewScanner(strings.NewReader(content))
	var currentSection string
	var sectionContent strings.Builder
	lineNum := 0

	for scanner.Scan() {
		line := scanner.Text()
		lineNum++

		// Handle shebang on first line
		if lineNum == 1 && strings.HasPrefix(line, "#!") {
			p.Shebang = line
			p.parseShebang(line)
			continue
		}

		// Check for section marker: -- section-name --
		if strings.HasPrefix(line, "-- ") && strings.HasSuffix(line, " --") {
			// Save previous section
			if currentSection != "" {
				p.Sections[currentSection] = strings.TrimSpace(sectionContent.String())
			} else if sectionContent.Len() > 0 {
				p.Main = strings.TrimSpace(sectionContent.String())
			}

			// Start new section
			currentSection = strings.TrimSpace(line[3 : len(line)-3])
			sectionContent.Reset()
			continue
		}

		// Add line to current section
		if sectionContent.Len() > 0 {
			sectionContent.WriteByte('\n')
		}
		sectionContent.WriteString(line)
	}

	// Save final section
	if currentSection != "" {
		p.Sections[currentSection] = strings.TrimSpace(sectionContent.String())
	} else if sectionContent.Len() > 0 {
		p.Main = strings.TrimSpace(sectionContent.String())
	}

	// Handle special sections
	if systemPrompt, ok := p.Sections["system-prompt"]; ok {
		p.SystemPrompt = systemPrompt
		delete(p.Sections, "system-prompt")
	}

	// Handle defaults section
	if defaults, ok := p.Sections["defaults"]; ok {
		p.ParseDefaults(defaults)
		delete(p.Sections, "defaults")
	}

	return p, nil
}

// ParseDefaults parses the defaults section
func (p *Prompt) ParseDefaults(content string) {
	scanner := bufio.NewScanner(strings.NewReader(content))
	
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		
		// Handle single-line format: key1=value1&key2=value2
		if strings.Contains(line, "&") || strings.Contains(line, "=") {
			// Remove quotes if the entire line is quoted
			line = strings.Trim(line, "'\"")
			
			// Split by & to get key=value pairs
			pairs := strings.Split(line, "&")
			for _, pair := range pairs {
				pair = strings.TrimSpace(pair)
				if kv := strings.SplitN(pair, "=", 2); len(kv) == 2 {
					key := strings.TrimSpace(kv[0])
					value := strings.TrimSpace(kv[1])
					// Remove quotes from individual values
					value = strings.Trim(value, "'\"")
					p.Defaults[key] = value
				}
			}
		} else {
			// Handle multi-line format: key: value
			if strings.Contains(line, ":") {
				kv := strings.SplitN(line, ":", 2)
				if len(kv) == 2 {
					key := strings.TrimSpace(kv[0])
					value := strings.TrimSpace(kv[1])
					value = strings.Trim(value, "'\"")
					p.Defaults[key] = value
				}
			}
		}
	}
}

// parseShebang extracts flags from shebang line
func (p *Prompt) parseShebang(shebang string) {
	// Parse: #!/usr/bin/env pe run --flag=value --flag2=value2
	parts := strings.Fields(shebang)
	
	for i, part := range parts {
		if strings.HasPrefix(part, "--") {
			if strings.Contains(part, "=") {
				kv := strings.SplitN(part[2:], "=", 2)
				p.Flags[kv[0]] = kv[1]
			} else {
				// Flag without value, might be in next part
				if i+1 < len(parts) && !strings.HasPrefix(parts[i+1], "--") {
					p.Flags[part[2:]] = parts[i+1]
				} else {
					p.Flags[part[2:]] = "true"
				}
			}
		}
	}
}

// GetVariant returns a variant section content
func (p *Prompt) GetVariant(name string) (string, bool) {
	content, ok := p.Sections["variants/"+name]
	return content, ok
}

// GetTest returns a test section content  
func (p *Prompt) GetTest(name string) (string, bool) {
	content, ok := p.Sections["tests/"+name]
	return content, ok
}

// ApplyVariant applies variant modifications to the prompt
func (p *Prompt) ApplyVariant(variant string) (*Prompt, error) {
	variantContent, ok := p.GetVariant(variant)
	if !ok {
		return nil, fmt.Errorf("variant %q not found", variant)
	}

	// Parse variant commands
	result := &Prompt{
		Main:         p.Main,
		SystemPrompt: p.SystemPrompt,
		Sections:     make(map[string]string),
		Flags:        make(map[string]string),
		Defaults:     make(map[string]string),
	}

	// Copy existing data
	for k, v := range p.Sections {
		result.Sections[k] = v
	}
	for k, v := range p.Flags {
		result.Flags[k] = v
	}
	for k, v := range p.Defaults {
		result.Defaults[k] = v
	}

	// Apply variant commands
	scanner := bufio.NewScanner(strings.NewReader(variantContent))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}

		cmd := parts[0]
		args := strings.Join(parts[1:], " ")

		switch cmd {
		case "extend-system-prompt":
			result.SystemPrompt += "\n" + strings.Trim(args, "'\"")
		case "prepend-system-prompt":
			result.SystemPrompt = strings.Trim(args, "'\"") + "\n" + result.SystemPrompt
		case "set-system-prompt":
			result.SystemPrompt = strings.Trim(args, "'\"")
		case "extend-prompt":
			result.Main += "\n" + strings.Trim(args, "'\"")
		case "prepend-prompt":
			result.Main = strings.Trim(args, "'\"") + "\n" + result.Main
		case "set-flag":
			if len(parts) >= 3 {
				result.Flags[parts[1]] = strings.Join(parts[2:], " ")
			}
		}
	}

	return result, nil
}

// Format returns the prompt formatted for execution
func (p *Prompt) Format() string {
	var parts []string

	if p.SystemPrompt != "" {
		parts = append(parts, fmt.Sprintf("System: %s", p.SystemPrompt))
	}

	parts = append(parts, fmt.Sprintf("User: %s", p.Main))

	return strings.Join(parts, "\n\n")
}

// Minimal returns just the main prompt text
func (p *Prompt) Minimal() string {
	return p.Main
}