package cli

import (
	"os"
	"strings"

	"github.com/tmc/pe/internal/prompt"
)

// VariableFlags handles custom variable flag parsing for prompt files
type VariableFlags struct {
	Variables map[string]string
}

// ParseVariableFlags extracts variable values from command-line flags
// It handles both --var key=value and direct --key=value syntax
func ParseVariableFlags(args []string, promptContent string) (*VariableFlags, error) {
	vf := &VariableFlags{
		Variables: make(map[string]string),
	}

	// Extract variables from the prompt
	promptVars := prompt.ExtractVariables(promptContent)
	if len(promptVars) == 0 {
		return vf, nil
	}

	// Create a map for case-insensitive lookup
	varLookup := make(map[string]string)
	for _, v := range promptVars {
		varLookup[strings.ToLower(v)] = v
	}

	// Parse os.Args for matching flags
	for i := 0; i < len(os.Args); i++ {
		arg := os.Args[i]

		// Handle --var key=value syntax
		if arg == "--var" && i+1 < len(os.Args) {
			if kv := os.Args[i+1]; strings.Contains(kv, "=") {
				parts := strings.SplitN(kv, "=", 2)
				if len(parts) == 2 {
					key := strings.TrimSpace(parts[0])
					value := strings.TrimSpace(parts[1])
					vf.Variables[key] = value
				}
			}
			i++ // Skip the next arg
			continue
		}

		// Handle direct --key=value or --key value syntax
		if strings.HasPrefix(arg, "--") {
			flagName := arg[2:]
			var value string

			// Check for --flag=value format
			if idx := strings.Index(flagName, "="); idx > 0 {
				value = flagName[idx+1:]
				flagName = flagName[:idx]
			} else if i+1 < len(os.Args) && !strings.HasPrefix(os.Args[i+1], "-") {
				// Check for --flag value format
				value = os.Args[i+1]
				i++ // Skip the next arg
			} else {
				continue
			}

			// Normalize flag name (handle hyphens -> underscores)
			normalizedFlag := strings.ReplaceAll(flagName, "-", "_")

			// Check if this flag matches a prompt variable (case-insensitive)
			if actualVar, exists := varLookup[strings.ToLower(normalizedFlag)]; exists {
				vf.Variables[actualVar] = value
			} else if actualVar, exists := varLookup[strings.ToLower(flagName)]; exists {
				vf.Variables[actualVar] = value
			}
		}
	}

	return vf, nil
}

// MergeWithDefaults merges parsed variables with defaults from the prompt
func (vf *VariableFlags) MergeWithDefaults(defaults map[string]string) map[string]string {
	result := make(map[string]string)

	// Start with defaults
	for k, v := range defaults {
		result[k] = v
	}

	// Override with parsed values
	for k, v := range vf.Variables {
		result[k] = v
	}

	return result
}

// GetMissingRequired returns variables that are required but not provided
func (vf *VariableFlags) GetMissingRequired(required []string, defaults map[string]string) []string {
	var missing []string

	for _, req := range required {
		// Check if we have a value or a default
		if _, hasValue := vf.Variables[req]; !hasValue {
			if _, hasDefault := defaults[req]; !hasDefault {
				missing = append(missing, req)
			}
		}
	}

	return missing
}

// FlagParser provides utilities for parsing custom flags
type FlagParser struct {
	KnownFlags map[string]bool // Standard flags to ignore
}

// NewFlagParser creates a new flag parser with standard pe run flags
func NewFlagParser() *FlagParser {
	return &FlagParser{
		KnownFlags: map[string]bool{
			"--help":        true,
			"-h":            true,
			"--model":       true,
			"-m":            true,
			"--provider":    true,
			"--temperature": true,
			"-t":            true,
			"--max-tokens":  true,
			"--system":      true,
			"-s":            true,
			"--stream":      true,
			"--json":        true,
			"--var":         true,
			"--example":     true,
			"-e":            true,
		},
	}
}

// IsKnownFlag checks if a flag is a standard pe run flag
func (fp *FlagParser) IsKnownFlag(flag string) bool {
	// Remove value if present
	if idx := strings.Index(flag, "="); idx > 0 {
		flag = flag[:idx]
	}
	return fp.KnownFlags[flag]
}

// ExtractUnknownFlags returns flags that aren't standard pe run flags
func (fp *FlagParser) ExtractUnknownFlags(args []string) []string {
	var unknown []string

	for i := 0; i < len(args); i++ {
		arg := args[i]

		if strings.HasPrefix(arg, "-") {
			if !fp.IsKnownFlag(arg) {
				unknown = append(unknown, arg)

				// If this flag takes a value (no = sign), skip the next arg
				if !strings.Contains(arg, "=") && i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
					i++
				}
			} else {
				// Skip known flag values
				if !strings.Contains(arg, "=") && i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
					i++
				}
			}
		}
	}

	return unknown
}
