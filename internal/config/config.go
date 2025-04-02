// Package config provides utilities for working with prompt evaluation configurations.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/tmc/pe/internal/promptfoo"
	"sigs.k8s.io/yaml"
)

// Config represents a prompt evaluation configuration.
// This is the internal representation used by the application.
type Config struct {
	Prompts   []string                 `json:"prompts"`
	Providers []string                 `json:"providers"`
	Tests     []map[string]interface{} `json:"tests"`
	Options   map[string]interface{}   `json:"options,omitempty"`
	Metadata  map[string]interface{}   `json:"metadata,omitempty"`
}

// LoadConfig loads a configuration from a file in either YAML or JSON format.
func LoadConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	ext := strings.ToLower(filepath.Ext(filename))
	formatHint := ""
	if ext == ".yaml" || ext == ".yml" {
		formatHint = "yaml"
	} else if ext == ".json" {
		formatHint = "json"
	}

	return LoadConfigFromData(data, formatHint)
}

// LoadConfigFromData loads a configuration from a data byte slice in either YAML or JSON format.
func LoadConfigFromData(data []byte, formatHint string) (*Config, error) {
	var config Config
	var err error

	formatHint = strings.ToLower(formatHint)

	if formatHint == "yaml" || formatHint == "yml" {
		// Try parsing as promptfoo format first
		var pfConfig promptfoo.Config
		err = yaml.Unmarshal(data, &pfConfig)
		if err == nil {
			// Convert promptfoo.Config to config.Config
			config.Prompts = pfConfig.Prompts
			config.Providers = pfConfig.Providers
			config.Tests = make([]map[string]interface{}, len(pfConfig.Tests))
			for i, testCase := range pfConfig.Tests {
				testMap := make(map[string]interface{})
				testMap["vars"] = testCase.Vars
				// Convert []Assertion to []interface{} for compatibility with validation
				asserts := make([]interface{}, len(testCase.Assert))
				for j, assertion := range testCase.Assert {
					// Convert assertion to map[string]interface{}
					assertMap := map[string]interface{}{
						"type":  assertion.Type,
						"value": assertion.Value,
						// TODO: Add other assertion fields if they exist in promptfoo.Assertion
					}
					asserts[j] = assertMap
				}
				testMap["assert"] = asserts
				config.Tests[i] = testMap
			}
			// TODO: Handle Options and Metadata if they exist in promptfoo.Config
			return &config, nil
		}
		// If promptfoo parsing failed, fall through to try the generic config struct
		// (This handles cases where the YAML might not match the promptfoo structure but matches the generic one)
		fmt.Fprintf(os.Stderr, "Warning: Failed to parse YAML as promptfoo format, trying generic structure: %v\n", err)
	}

	// Handle JSON or fallback YAML parsing using the generic Config struct
	if formatHint == "json" {
		err = json.Unmarshal(data, &config)
		if err != nil {
			return nil, fmt.Errorf("error parsing JSON config: %w", err)
		}
	} else if formatHint == "yaml" || formatHint == "yml" {
		// This path is reached if promptfoo parsing failed for YAML
		err = yaml.Unmarshal(data, &config)
		if err != nil {
			return nil, fmt.Errorf("error parsing YAML config (generic structure): %w", err)
		}
	} else {
		// If no format hint, try YAML (promptfoo first), then JSON
		var pfConfig promptfoo.Config
		err = yaml.Unmarshal(data, &pfConfig)
		if err == nil {
			// Convert promptfoo.Config to config.Config (duplicate code, consider refactoring)
			config.Prompts = pfConfig.Prompts
			config.Providers = pfConfig.Providers
			config.Tests = make([]map[string]interface{}, len(pfConfig.Tests))
			for i, testCase := range pfConfig.Tests {
				testMap := make(map[string]interface{})
				testMap["vars"] = testCase.Vars
				asserts := make([]interface{}, len(testCase.Assert))
				for j, assertion := range testCase.Assert {
					assertMap := map[string]interface{}{
						"type":  assertion.Type,
						"value": assertion.Value,
					}
					asserts[j] = assertMap
				}
				testMap["assert"] = asserts
				config.Tests[i] = testMap
			}
			return &config, nil
		}

		// Try generic YAML
		err = yaml.Unmarshal(data, &config)
		if err != nil {
			// Try JSON as a last resort
			errJSON := json.Unmarshal(data, &config)
			if errJSON != nil {
				return nil, fmt.Errorf("error parsing config (tried promptfoo YAML, generic YAML, and JSON): %w, %w", err, errJSON)
			}
		}
	}

	return &config, nil
}

// Validate checks if the configuration is valid.
func (c *Config) Validate() error {
	if len(c.Prompts) == 0 {
		return fmt.Errorf("no prompts specified")
	}

	if len(c.Providers) == 0 {
		return fmt.Errorf("no providers specified")
	}

	if len(c.Tests) == 0 {
		return fmt.Errorf("no tests specified")
	}

	// Validate that each test has vars and at least one assertion
	for i, test := range c.Tests {
		if _, hasVars := test["vars"]; !hasVars {
			return fmt.Errorf("test %d is missing 'vars' field", i)
		}

		assertions, hasAssertions := test["assert"]
		if !hasAssertions {
			return fmt.Errorf("test %d is missing 'assert' field", i)
		}

		assertList, ok := assertions.([]interface{})
		if !ok || len(assertList) == 0 {
			return fmt.Errorf("test %d has invalid or empty 'assert' list", i)
		}

		// Validate individual assertions within the list
		for j, assertItem := range assertList {
			assertMap, ok := assertItem.(map[string]interface{})
			if !ok {
				return fmt.Errorf("test %d, assertion %d is not a valid map", i, j)
			}
			if _, hasType := assertMap["type"]; !hasType {
				return fmt.Errorf("test %d, assertion %d is missing 'type' field", i, j)
			}
			// Add more specific assertion validation if needed (e.g., check for 'value')
		}
	}

	return nil
}

// AsMap converts the Config to a generic map for use with evaluators.
func (c *Config) AsMap() map[string]interface{} {
	// Marshal to JSON and unmarshal to map - a bit inefficient but ensures correct types
	data, _ := json.Marshal(c)
	var result map[string]interface{}
	_ = json.Unmarshal(data, &result)
	return result
}

// SaveToFile saves the configuration to a file in the specified format.
func (c *Config) SaveToFile(filename string, formatHint string) error {
	var data []byte
	var err error

	formatHint = strings.ToLower(formatHint)
	if formatHint == "" {
		// Detect format from filename extension
		ext := strings.ToLower(filepath.Ext(filename))
		if ext == ".yaml" || ext == ".yml" {
			formatHint = "yaml"
		} else if ext == ".json" {
			formatHint = "json"
		} else {
			formatHint = "yaml" // Default to YAML
		}
	}

	if formatHint == "yaml" || formatHint == "yml" {
		data, err = yaml.Marshal(c)
		if err != nil {
			return fmt.Errorf("error marshaling to YAML: %w", err)
		}
	} else if formatHint == "json" {
		data, err = json.MarshalIndent(c, "", "  ")
		if err != nil {
			return fmt.Errorf("error marshaling to JSON: %w", err)
		}
	} else {
		return fmt.Errorf("unsupported format: %s", formatHint)
	}

	err = os.WriteFile(filename, data, 0644)
	if err != nil {
		return fmt.Errorf("error writing to file: %w", err)
	}

	return nil
}
