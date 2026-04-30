package providers

import "fmt"

// CLIPreset represents a predefined CLI configuration.
type CLIPreset struct {
	build func(options map[string]interface{}) (map[string]interface{}, error)
}

// Presets defines the built-in CLI tool configurations.
var Presets = map[string]CLIPreset{
	"mlx":       {build: buildMLXLMPreset},
	"mlx-lm":    {build: buildMLXLMPreset},
	"mlx-go":    {build: buildMLXGoPreset},
	"mlx-go-lm": {build: buildMLXGoPreset},
	"llama-cpp": {build: buildLlamaCPPPreset},
	"llama.cpp": {build: buildLlamaCPPPreset},
	"llm-tool":  {build: buildLLMToolPreset},
}

// GetPresetConfig returns the configuration options for a given preset.
func GetPresetConfig(presetName string, userOptions map[string]interface{}) (map[string]interface{}, error) {
	preset, ok := Presets[presetName]
	if !ok {
		return nil, fmt.Errorf("unknown preset: %s", presetName)
	}
	return preset.build(userOptions)
}

func buildMLXLMPreset(options map[string]interface{}) (map[string]interface{}, error) {
	return presetConfig(
		getStringOption(options, "executable", "mlx_lm.generate"),
		[]string{
			"{{if .HasModel}}--model{{end}}",
			"{{if .HasModel}}{{.Model}}{{end}}",
			"--prompt",
			"{{.Prompt}}",
			"{{if .HasMaxTokens}}--max-tokens{{end}}",
			"{{if .HasMaxTokens}}{{.MaxTokens}}{{end}}",
			"{{if .HasTemperature}}--temp{{end}}",
			"{{if .HasTemperature}}{{.Temperature}}{{end}}",
		},
		options,
	), nil
}

func buildMLXGoPreset(options map[string]interface{}) (map[string]interface{}, error) {
	return presetConfig(
		getStringOption(options, "executable", "mlx-lm-generate"),
		[]string{
			"{{if .HasModel}}--model{{end}}",
			"{{if .HasModel}}{{.Model}}{{end}}",
			"--prompt",
			"{{.Prompt}}",
			"{{if .HasMaxTokens}}--max-tokens{{end}}",
			"{{if .HasMaxTokens}}{{.MaxTokens}}{{end}}",
			"{{if .HasTemperature}}--temperature{{end}}",
			"{{if .HasTemperature}}{{.Temperature}}{{end}}",
		},
		options,
	), nil
}

func buildLlamaCPPPreset(options map[string]interface{}) (map[string]interface{}, error) {
	return presetConfig(
		getStringOption(options, "executable", "llama-cli"),
		[]string{
			"{{if .HasModel}}-m{{end}}",
			"{{if .HasModel}}{{.Model}}{{end}}",
			"-p",
			"{{.Prompt}}",
			"{{if .HasMaxTokens}}-n{{end}}",
			"{{if .HasMaxTokens}}{{.MaxTokens}}{{end}}",
			"{{if .HasTemperature}}--temp{{end}}",
			"{{if .HasTemperature}}{{.Temperature}}{{end}}",
		},
		options,
	), nil
}

func buildLLMToolPreset(options map[string]interface{}) (map[string]interface{}, error) {
	return presetConfig(
		getStringOption(options, "executable", "llm-tool"),
		[]string{
			"generate",
			"{{if .HasModel}}--model{{end}}",
			"{{if .HasModel}}{{.Model}}{{end}}",
			"--prompt",
			"{{.Prompt}}",
			"{{if .HasMaxTokens}}--max-tokens{{end}}",
			"{{if .HasMaxTokens}}{{.MaxTokens}}{{end}}",
			"{{if .HasTemperature}}--temperature{{end}}",
			"{{if .HasTemperature}}{{.Temperature}}{{end}}",
		},
		options,
	), nil
}

func presetConfig(executable string, baseArgs []string, userOptions map[string]interface{}) map[string]interface{} {
	config := make(map[string]interface{})
	for k, v := range userOptions {
		config[k] = v
	}
	if _, ok := userOptions["command"]; ok {
		return config
	}
	if _, hasExecutable := userOptions["executable"]; hasExecutable {
		if _, hasArgs := userOptions["args"]; hasArgs {
			return config
		}
	}
	config["executable"] = executable
	args := append([]string(nil), baseArgs...)
	args = append(args, getStringSliceOption(userOptions, "args")...)
	config["args"] = args
	return config
}
