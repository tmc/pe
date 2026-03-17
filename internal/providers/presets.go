package providers

import (
	"fmt"
	"strconv"
	"strings"
)

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
	executable := getStringOption(options, "executable", "mlx_lm.generate")
	command := buildCommandTemplate(
		executable,
		[]string{
			"{{if .Model}}--model {{printf \"%q\" .Model}}{{end}}",
			"--prompt {{printf \"%q\" .Prompt}}",
			"--max-tokens {{.MaxTokens}}",
			"--temp {{.Temperature}}",
			"--verbose false",
		},
		getStringSliceOption(options, "args"),
	)
	return presetConfig(command, options), nil
}

func buildMLXGoPreset(options map[string]interface{}) (map[string]interface{}, error) {
	executable := getStringOption(options, "executable", "mlx-lm-generate")
	command := buildCommandTemplate(
		executable,
		[]string{
			"{{if .Model}}--model {{printf \"%q\" .Model}}{{end}}",
			"--prompt {{printf \"%q\" .Prompt}}",
			"--max-tokens {{.MaxTokens}}",
			"--temperature {{.Temperature}}",
			"--quiet",
		},
		getStringSliceOption(options, "args"),
	)
	return presetConfig(command, options), nil
}

func buildLlamaCPPPreset(options map[string]interface{}) (map[string]interface{}, error) {
	executable := getStringOption(options, "executable", "llama-cli")
	command := buildCommandTemplate(
		executable,
		[]string{
			"-m {{printf \"%q\" .Model}}",
			"-p {{printf \"%q\" .Prompt}}",
			"-n {{.MaxTokens}}",
			"--temp {{.Temperature}}",
		},
		getStringSliceOption(options, "args"),
	)
	return presetConfig(command, options), nil
}

func buildLLMToolPreset(options map[string]interface{}) (map[string]interface{}, error) {
	executable := getStringOption(options, "executable", "llm-tool")
	command := buildCommandTemplate(
		executable,
		[]string{
			"generate",
			"--model {{printf \"%q\" .Model}}",
			"--prompt {{printf \"%q\" .Prompt}}",
			"--max-tokens {{.MaxTokens}}",
			"--temperature {{.Temperature}}",
		},
		getStringSliceOption(options, "args"),
	)
	return presetConfig(command, options), nil
}

func buildCommandTemplate(executable string, baseArgs, extraArgs []string) string {
	parts := []string{shellEscape(executable)}
	parts = append(parts, baseArgs...)
	for _, arg := range extraArgs {
		parts = append(parts, shellEscape(arg))
	}
	return strings.Join(parts, " ")
}

func presetConfig(command string, userOptions map[string]interface{}) map[string]interface{} {
	config := make(map[string]interface{})
	for k, v := range userOptions {
		config[k] = v
	}
	config["command"] = command
	return config
}

func shellEscape(s string) string {
	if s == "" {
		return "''"
	}
	if !strings.ContainsAny(s, " \t\n\r'\"\\") {
		return s
	}
	return strconv.Quote(s)
}
