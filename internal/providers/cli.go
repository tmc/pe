package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/kballard/go-shellquote"
	"github.com/tmc/pe/internal/llm"
	"github.com/tmc/pe/internal/promptfoo"
)

// GenericCLIProvider implements the Provider interface for any CLI tool
type GenericCLIProvider struct {
	commandTemplate string
	executable      string
	argTemplates    []string
	promptStdin     bool
	parseJSON       bool
	model           string
	env             map[string]string
	options         map[string]interface{}
}

// NewGenericCLIProvider creates a new generic CLI provider
func NewGenericCLIProvider(model string, options map[string]interface{}) (*GenericCLIProvider, error) {
	if options == nil {
		options = map[string]interface{}{}
	}

	env := getStringMapOption(options, "env")
	executable := getStringOption(options, "executable", "")
	args := getStringSliceOption(options, "args")
	commandTemplate := ""

	if command, ok := options["command"]; ok {
		switch v := command.(type) {
		case string:
			commandTemplate = v
		case []string:
			if executable == "" && len(v) > 0 {
				executable = v[0]
				args = append([]string(nil), v[1:]...)
			}
		case []interface{}:
			if executable == "" && len(v) > 0 {
				executable = fmt.Sprintf("%v", v[0])
				args = make([]string, 0, len(v)-1)
				for _, arg := range v[1:] {
					args = append(args, fmt.Sprintf("%v", arg))
				}
			}
		}
	}

	if commandTemplate == "" && executable == "" {
		return nil, fmt.Errorf("generic cli provider requires command or executable")
	}

	return &GenericCLIProvider{
		commandTemplate: commandTemplate,
		executable:      executable,
		argTemplates:    args,
		promptStdin:     getBoolOption(options, "prompt_stdin", false),
		parseJSON:       getBoolOption(options, "parse_json_response", false),
		model:           model,
		env:             env,
		options:         options,
	}, nil
}

// Name returns the provider name
func (p *GenericCLIProvider) Name() string {
	return "cli"
}

// Model returns the model name
func (p *GenericCLIProvider) Model() string {
	return p.model
}

type templateData struct {
	Model          string
	HasModel       bool
	Prompt         string
	HasPrompt      bool
	Temperature    float64
	HasTemperature bool
	MaxTokens      int
	HasMaxTokens   bool
}

// Generate generates a response using the configured CLI command
func (p *GenericCLIProvider) Generate(ctx context.Context, prompt string, options llm.GenerateOptions) (*llm.GenerateResponse, error) {
	data := makeTemplateData(p.model, prompt, p.options, options)

	cmd, err := p.buildCommand(ctx, data)
	if err != nil {
		return nil, err
	}

	// Set environment variables
	cmd.Env = os.Environ()
	for k, v := range p.env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}

	var out bytes.Buffer
	var errBuf bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errBuf

	start := time.Now()
	if err := cmd.Run(); err != nil {
		if errBuf.Len() > 0 {
			return nil, fmt.Errorf("cli error: %s", errBuf.String())
		}
		return nil, fmt.Errorf("cli execution failed: %w", err)
	}
	latency := time.Since(start)

	stdout := strings.TrimSpace(out.String())
	if !p.parseJSON {
		return &llm.GenerateResponse{
			Text:    stdout,
			Model:   p.model,
			Latency: latency,
		}, nil
	}
	return parseCLIResponse(stdout, p.model, latency), nil
}

// EvaluatePrompt implements the legacy Provider interface method
func (p *GenericCLIProvider) EvaluatePrompt(ctx context.Context, prompt string, vars map[string]interface{}) (*promptfoo.ProviderResponse, error) {
	// Apply variables to prompt if needed
	finalPrompt := promptfoo.ApplyVars(prompt, vars)

	generateOptions := llm.GenerateOptions{}
	if temp, ok := lookupFloat64Option(vars, "temperature"); ok {
		generateOptions.Temperature = &temp
	} else if temp, ok := lookupFloat64Option(vars, "temp"); ok {
		generateOptions.Temperature = &temp
	}
	if maxTokens, ok := lookupIntOption(vars, "max_tokens"); ok {
		generateOptions.MaxTokens = &maxTokens
	} else if maxTokens, ok := lookupIntOption(vars, "num_predict"); ok {
		generateOptions.MaxTokens = &maxTokens
	}

	result, err := p.Generate(ctx, finalPrompt, generateOptions)
	if err != nil {
		return nil, err
	}

	return &promptfoo.ProviderResponse{
		Output: result.Text,
		TokenUsage: &promptfoo.TokenUsage{
			Total:      int32(result.TotalTokens),
			Prompt:     int32(result.PromptTokens),
			Completion: int32(result.CompletionTokens),
		},
		Cost:      result.Cost,
		LatencyMs: result.Latency.Milliseconds(),
		Metadata:  result.Metadata,
	}, nil
}

func (p *GenericCLIProvider) SupportsStreaming() bool {
	return false
}

func (p *GenericCLIProvider) SupportsBatch() bool {
	return false
}

func (p *GenericCLIProvider) buildCommand(ctx context.Context, data templateData) (*exec.Cmd, error) {
	if p.executable != "" {
		executable, err := renderTemplateString(p.executable, data)
		if err != nil {
			return nil, fmt.Errorf("render executable: %w", err)
		}
		executable = strings.TrimSpace(executable)
		if executable == "" {
			return nil, fmt.Errorf("empty executable")
		}

		args, err := renderTemplateArgs(p.argTemplates, data)
		if err != nil {
			return nil, fmt.Errorf("render args: %w", err)
		}
		if _, err := exec.LookPath(executable); err != nil {
			return nil, fmt.Errorf("executable not found: %s", executable)
		}
		cmd := exec.CommandContext(ctx, executable, args...)
		if p.promptStdin {
			cmd.Stdin = strings.NewReader(data.Prompt)
		}
		return cmd, nil
	}

	tmpl, err := template.New("cli").Parse(p.commandTemplate)
	if err != nil {
		return nil, fmt.Errorf("failed to parse command template: %w", err)
	}

	var cmdStr bytes.Buffer
	if err := tmpl.Execute(&cmdStr, data); err != nil {
		return nil, fmt.Errorf("failed to execute command template: %w", err)
	}

	parts, err := shellquote.Split(cmdStr.String())
	if err != nil {
		return nil, fmt.Errorf("failed to split command string: %w", err)
	}
	if len(parts) == 0 {
		return nil, fmt.Errorf("empty command resulted from template")
	}

	executable := parts[0]
	args := parts[1:]
	if _, err := exec.LookPath(executable); err != nil {
		return nil, fmt.Errorf("executable not found: %s", executable)
	}
	cmd := exec.CommandContext(ctx, executable, args...)
	if p.promptStdin {
		cmd.Stdin = strings.NewReader(data.Prompt)
	}
	return cmd, nil
}

func makeTemplateData(model, prompt string, providerOptions map[string]interface{}, options llm.GenerateOptions) templateData {
	data := templateData{
		Model:     model,
		HasModel:  model != "",
		Prompt:    prompt,
		HasPrompt: true,
	}

	if temp, ok := lookupFloat64Option(providerOptions, "temperature"); ok {
		data.Temperature = temp
		data.HasTemperature = true
	}
	if temp, ok := lookupFloat64Option(providerOptions, "temp"); ok && !data.HasTemperature {
		data.Temperature = temp
		data.HasTemperature = true
	}
	if maxTokens, ok := lookupIntOption(providerOptions, "max_tokens"); ok {
		data.MaxTokens = maxTokens
		data.HasMaxTokens = true
	}
	if maxTokens, ok := lookupIntOption(providerOptions, "num_predict"); ok && !data.HasMaxTokens {
		data.MaxTokens = maxTokens
		data.HasMaxTokens = true
	}

	if options.Temperature != nil {
		data.Temperature = *options.Temperature
		data.HasTemperature = true
	}
	if options.MaxTokens != nil {
		data.MaxTokens = *options.MaxTokens
		data.HasMaxTokens = true
	}

	return data
}

func renderTemplateArgs(templates []string, data templateData) ([]string, error) {
	args := make([]string, 0, len(templates))
	for _, argTemplate := range templates {
		rendered, err := renderTemplateString(argTemplate, data)
		if err != nil {
			return nil, err
		}
		rendered = strings.TrimSpace(rendered)
		if rendered == "" {
			continue
		}
		args = append(args, rendered)
	}
	return args, nil
}

func renderTemplateString(value string, data templateData) (string, error) {
	tmpl, err := template.New("cli-arg").Parse(value)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func parseCLIResponse(stdout string, model string, measuredLatency time.Duration) *llm.GenerateResponse {
	resp := &llm.GenerateResponse{
		Text:    stdout,
		Model:   model,
		Latency: measuredLatency,
	}

	var raw map[string]interface{}
	if err := json.Unmarshal([]byte(stdout), &raw); err != nil {
		return resp
	}
	if !looksLikeStructuredCLIResponse(raw) {
		return resp
	}

	if text := getFirstString(raw, "output", "text", "response"); text != "" {
		resp.Text = text
	}
	if promptTokens, ok := getInt(raw, "prompt_tokens"); ok {
		resp.PromptTokens = promptTokens
	}
	if completionTokens, ok := getInt(raw, "completion_tokens"); ok {
		resp.CompletionTokens = completionTokens
	}
	if totalTokens, ok := getInt(raw, "total_tokens"); ok {
		resp.TotalTokens = totalTokens
	}
	if tokenUsage, ok := getMap(raw, "tokenUsage"); ok {
		mergeTokenUsage(resp, tokenUsage)
	}
	if tokenUsage, ok := getMap(raw, "token_usage"); ok {
		mergeTokenUsage(resp, tokenUsage)
	}
	if latencyMs, ok := getInt64(raw, "latency_ms"); ok {
		resp.Latency = time.Duration(latencyMs) * time.Millisecond
	}
	if cost, ok := getFloat(raw, "cost"); ok {
		resp.Cost = cost
	}

	metadata := make(map[string]interface{})
	if nested, ok := getMap(raw, "metadata"); ok {
		for k, v := range nested {
			metadata[k] = v
		}
	}
	if nested, ok := getMap(raw, "metrics"); ok {
		for k, v := range nested {
			metadata[k] = v
		}
	}
	for k, v := range raw {
		switch k {
		case "output", "text", "response", "prompt_tokens", "completion_tokens", "total_tokens", "latency_ms", "cost", "tokenUsage", "token_usage", "metadata", "metrics":
			continue
		default:
			metadata[k] = v
		}
	}
	if len(metadata) > 0 {
		resp.Metadata = metadata
	}
	if resp.TotalTokens == 0 && (resp.PromptTokens > 0 || resp.CompletionTokens > 0) {
		resp.TotalTokens = resp.PromptTokens + resp.CompletionTokens
	}

	return resp
}

func looksLikeStructuredCLIResponse(raw map[string]interface{}) bool {
	for _, key := range []string{"output", "text", "response", "prompt_tokens", "completion_tokens", "total_tokens", "latency_ms", "tokenUsage", "token_usage", "metrics", "metadata"} {
		if _, ok := raw[key]; ok {
			return true
		}
	}
	return false
}

func mergeTokenUsage(resp *llm.GenerateResponse, raw map[string]interface{}) {
	if promptTokens, ok := getInt(raw, "prompt"); ok {
		resp.PromptTokens = promptTokens
	}
	if promptTokens, ok := getInt(raw, "prompt_tokens"); ok {
		resp.PromptTokens = promptTokens
	}
	if completionTokens, ok := getInt(raw, "completion"); ok {
		resp.CompletionTokens = completionTokens
	}
	if completionTokens, ok := getInt(raw, "completion_tokens"); ok {
		resp.CompletionTokens = completionTokens
	}
	if totalTokens, ok := getInt(raw, "total"); ok {
		resp.TotalTokens = totalTokens
	}
	if totalTokens, ok := getInt(raw, "total_tokens"); ok {
		resp.TotalTokens = totalTokens
	}
}

func getFirstString(raw map[string]interface{}, keys ...string) string {
	for _, key := range keys {
		if value, ok := raw[key].(string); ok {
			return value
		}
	}
	return ""
}

func getMap(raw map[string]interface{}, key string) (map[string]interface{}, bool) {
	value, ok := raw[key]
	if !ok {
		return nil, false
	}
	m, ok := value.(map[string]interface{})
	return m, ok
}

func getInt(raw map[string]interface{}, key string) (int, bool) {
	value, ok := raw[key]
	if !ok {
		return 0, false
	}
	switch v := value.(type) {
	case float64:
		return int(v), true
	case int:
		return v, true
	case int64:
		return int(v), true
	case json.Number:
		i, err := v.Int64()
		if err == nil {
			return int(i), true
		}
	case string:
		if i, err := strconv.Atoi(v); err == nil {
			return i, true
		}
	}
	return 0, false
}

func getInt64(raw map[string]interface{}, key string) (int64, bool) {
	value, ok := raw[key]
	if !ok {
		return 0, false
	}
	switch v := value.(type) {
	case float64:
		return int64(v), true
	case int:
		return int64(v), true
	case int64:
		return v, true
	case json.Number:
		i, err := v.Int64()
		if err == nil {
			return i, true
		}
	case string:
		if i, err := strconv.ParseInt(v, 10, 64); err == nil {
			return i, true
		}
	}
	return 0, false
}

func getFloat(raw map[string]interface{}, key string) (float64, bool) {
	value, ok := raw[key]
	if !ok {
		return 0, false
	}
	switch v := value.(type) {
	case float64:
		return v, true
	case int:
		return float64(v), true
	case json.Number:
		f, err := v.Float64()
		if err == nil {
			return f, true
		}
	case string:
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f, true
		}
	}
	return 0, false
}
