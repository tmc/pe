package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"
)

func expandCmd() *cobra.Command {
	var outputFile string

	cmd := &cobra.Command{
		Use:   "expand [config_file]",
		Short: "Expand a configuration file by resolving all file references and globs",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			configFile := args[0]

			// Load and expand
			config, err := loadAndExpandConfig(configFile)
			if err != nil {
				return err
			}

			// Marshal to JSON
			output, err := json.MarshalIndent(config, "", "  ")
			if err != nil {
				return fmt.Errorf("error marshaling expanded config: %v", err)
			}

			if outputFile != "" {
				return os.WriteFile(outputFile, output, 0644)
			}

			fmt.Println(string(output))
			return nil
		},
	}

	cmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file (JSON)")
	return cmd
}

// Structures for flexible parsing (Ported from tools/config-expander)

type RawConfig struct {
	Description string      `json:"description,omitempty"`
	Prompts     interface{} `json:"prompts"`   // string, []string, or map[string]string
	Providers   interface{} `json:"providers"` // string, []string, or []object
	Tests       interface{} `json:"tests"`     // string, []string, or []TestCase
	Scenarios   interface{} `json:"scenarios,omitempty"`
	DefaultTest interface{} `json:"defaultTest,omitempty"`
	Env         interface{} `json:"env,omitempty"`
}

type ExpandedConfig struct {
	Description string        `json:"description,omitempty"`
	Prompts     []interface{} `json:"prompts"`
	Providers   []interface{} `json:"providers"`
	Tests       []interface{} `json:"tests"`
	Scenarios   interface{}   `json:"scenarios,omitempty"`
	DefaultTest interface{}   `json:"defaultTest,omitempty"`
	Env         interface{}   `json:"env,omitempty"`
}

func loadAndExpandConfig(path string) (*ExpandedConfig, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var raw RawConfig
	if err := yaml.Unmarshal(content, &raw); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %v", err)
	}

	baseDir := filepath.Dir(path)

	expanded := &ExpandedConfig{
		Description: raw.Description,
		Scenarios:   raw.Scenarios,
		DefaultTest: raw.DefaultTest,
		Env:         raw.Env,
	}

	// Expand Prompts
	prompts, err := resolvePrompts(raw.Prompts, baseDir)
	if err != nil {
		return nil, fmt.Errorf("resolving prompts: %v", err)
	}
	expanded.Prompts = prompts

	// Expand Providers
	providers, err := resolveProviders(raw.Providers, baseDir)
	if err != nil {
		return nil, fmt.Errorf("resolving providers: %v", err)
	}
	expanded.Providers = providers

	// Expand Tests
	tests, err := resolveTests(raw.Tests, baseDir)
	if err != nil {
		return nil, fmt.Errorf("resolving tests: %v", err)
	}
	expanded.Tests = tests

	return expanded, nil
}

func resolvePrompts(prompts interface{}, baseDir string) ([]interface{}, error) {
	var results []interface{}

	if prompts == nil {
		return results, nil
	}

	switch v := prompts.(type) {
	case string:
		if isFileRef(v) {
			p, err := loadPromptsFromGlob(v, baseDir)
			if err != nil {
				return nil, err
			}
			for _, item := range p {
				results = append(results, item)
			}
		} else {
			results = append(results, v)
		}
	case []interface{}:
		for _, item := range v {
			if s, ok := item.(string); ok {
				if isFileRef(s) {
					p, err := loadPromptsFromGlob(s, baseDir)
					if err != nil {
						return nil, err
					}
					for _, item := range p {
						results = append(results, item)
					}
				} else {
					results = append(results, s)
				}
			} else {
				results = append(results, item)
			}
		}
	case map[string]interface{}:
		// Handle map format (id: prompt)
		results = append(results, v)
	}
	return results, nil
}

func loadPromptsFromGlob(pattern, baseDir string) ([]string, error) {
	pattern = strings.TrimPrefix(pattern, "file://")
	fullPattern := filepath.Join(baseDir, pattern)
	matches, err := filepath.Glob(fullPattern)
	if err != nil {
		return nil, err
	}
	if len(matches) == 0 {
		// Provide a fallback or warning? user might have just typed a path that doesn't exist yet
		return nil, fmt.Errorf("no files match prompt pattern: %s", fullPattern)
	}

	var prompts []string
	for _, match := range matches {
		content, err := os.ReadFile(match)
		if err != nil {
			return nil, err
		}
		prompts = append(prompts, string(content))
	}
	return prompts, nil
}

func resolveProviders(providers interface{}, baseDir string) ([]interface{}, error) {
	var results []interface{}
	if providers == nil {
		return results, nil
	}

	switch v := providers.(type) {
	case string:
		results = append(results, v) // Could be file ref? usually provider is a string name
	case []interface{}:
		results = append(results, v...)
	}
	// TODO: deep expansion if needed
	return results, nil
}

func resolveTests(tests interface{}, baseDir string) ([]interface{}, error) {
	var results []interface{}
	if tests == nil {
		return results, nil
	}

	switch v := tests.(type) {
	case string:
		t, err := loadTestsFromFile(v, baseDir)
		if err != nil {
			return nil, err
		}
		results = append(results, t...)
	case []interface{}:
		for _, item := range v {
			if s, ok := item.(string); ok {
				t, err := loadTestsFromFile(s, baseDir)
				if err != nil {
					return nil, err
				}
				results = append(results, t...)
			} else {
				results = append(results, item)
			}
		}
	}
	return results, nil
}

func loadTestsFromFile(pathStr, baseDir string) ([]interface{}, error) {
	pathStr = strings.TrimPrefix(pathStr, "file://")
	fullPath := filepath.Join(baseDir, pathStr)

	f, err := os.Open(fullPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	ext := strings.ToLower(filepath.Ext(fullPath))
	var tests []interface{}

	if ext == ".json" {
		dec := json.NewDecoder(f)
		if err := dec.Decode(&tests); err != nil {
			return nil, err
		}
	} else if ext == ".yaml" || ext == ".yml" {
		// YAML decoder in sigs.k8s.io/yaml handles JSON too mostly, but lets use ReadAll + Unmarshal
		data, err := os.ReadFile(fullPath)
		if err != nil {
			return nil, err
		}
		if err := yaml.Unmarshal(data, &tests); err != nil {
			return nil, err
		}
	} else if ext == ".csv" {
		// Minimal, skip or simple text
		return nil, fmt.Errorf("csv not supported in expand yet")
	}

	return tests, nil
}

func isFileRef(s string) bool {
	return strings.HasPrefix(s, "file://") ||
		strings.Contains(s, "*") ||
		strings.HasSuffix(s, ".txt") ||
		strings.HasSuffix(s, ".json") ||
		strings.HasSuffix(s, ".yaml") ||
		strings.HasSuffix(s, ".yml") ||
		strings.HasSuffix(s, ".py") ||
		strings.HasSuffix(s, ".js")
}
