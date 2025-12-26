package providers

import (
	"fmt"
)

// CLIPreset represents a predefined CLI configuration
type CLIPreset struct {
	CommandTemplate string
	DefaultOptions  map[string]interface{}
}

// Presets defines the built-in CLI tool configurations
var Presets = map[string]CLIPreset{
	"ollama": {
		CommandTemplate: "ollama run {{.Model}} {{.Prompt}}",
	},
	"mlx": {
		// Assumes python environment with mlx_lm installed
		CommandTemplate: "python -m mlx_lm.generate --model {{.Model}} --prompt {{.Prompt}} --max-tokens {{.MaxTokens}} --temp {{.Temperature}}",
	},
	"llama-cpp": {
		// Requires user to likely specify binary path via options or have 'llama-cli' in PATH
		// This is a common name, but 'main' is also common for built source
		CommandTemplate: "llama-cli -m {{.Model}} -p {{.Prompt}} -n {{.MaxTokens}} --temp {{.Temperature}}",
	},
	"llm-tool": {
		// MLX Swift example tool
		CommandTemplate: "llm-tool generate --model {{.Model}} --prompt {{.Prompt}} --max-tokens {{.MaxTokens}} --temperature {{.Temperature}}",
	},
	// Keeps 'llm' separate via LLMCLIProvider? Or migrate 'llm' here?
	// For backward compatibility, we can keep using NewLLMCLIProvider or make a preset here.
	// Simon Willison's llm tool has specific flag structure.
}

// GetPresetConfig returns the configuration options for a given preset
func GetPresetConfig(presetName string, userOptions map[string]interface{}) (map[string]interface{}, error) {
	preset, ok := Presets[presetName]
	if !ok {
		return nil, fmt.Errorf("unknown preset: %s", presetName)
	}

	config := make(map[string]interface{})

	// Copy default options if any
	for k, v := range preset.DefaultOptions {
		config[k] = v
	}

	// Set the command template
	config["command"] = preset.CommandTemplate

	// Merge user options
	for k, v := range userOptions {
		config[k] = v
	}

	return config, nil
}
