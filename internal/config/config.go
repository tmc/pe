// Package config provides utilities for working with prompt evaluation configurations.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"sigs.k8s.io/yaml"
)

// Config represents a prompt evaluation configuration.
type Config struct {
	Prompts   []string               `json:"prompts"`
	Providers []string               `json:"providers"`
	Tests     []map[string]interface{} `json:"tests"`
	Options   map[string]interface{} `json:"options,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// LoadConfig loads a configuration from a file in either YAML or JSON format.
func LoadConfig(filename string) (*Config, error) {
	// Read config file
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}
	
	// Detect format based on extension
	var config Config
	ext := strings.ToLower(filepath.Ext(filename))
	
	if ext == ".yaml" || ext == ".yml" {
		err = yaml.Unmarshal(data, &config)
		if err != nil {
			return nil, fmt.Errorf("error parsing YAML config: %w", err)
		}
	} else if ext == ".json" {
		err = json.Unmarshal(data, &config)
		if err != nil {
			return nil, fmt.Errorf("error parsing JSON config: %w", err)
		}
	} else {
		// Try YAML first, then JSON if that fails
		err = yaml.Unmarshal(data, &config)
		if err != nil {
			err = json.Unmarshal(data, &config)
			if err != nil {
				return nil, fmt.Errorf("error parsing config (tried YAML and JSON): %w", err)
			}
		}
	}
	
	return &config, nil
}

// LoadConfigFromData loads a configuration from a data byte slice in either YAML or JSON format.
func LoadConfigFromData(data []byte, formatHint string) (*Config, error) {
	var config Config
	var err error
	
	formatHint = strings.ToLower(formatHint)
	
	if formatHint == "yaml" || formatHint == "yml" {
		err = yaml.Unmarshal(data, &config)
		if err != nil {
			return nil, fmt.Errorf("error parsing YAML config: %w", err)
		}
	} else if formatHint == "json" {
		err = json.Unmarshal(data, &config)
		if err != nil {
			return nil, fmt.Errorf("error parsing JSON config: %w", err)
		}
	} else {
		// Try YAML first, then JSON if that fails
		err = yaml.Unmarshal(data, &config)
		if err != nil {
			err = json.Unmarshal(data, &config)
			if err != nil {
				return nil, fmt.Errorf("error parsing config (tried YAML and JSON): %w", err)
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