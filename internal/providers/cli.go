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
	model           string
	env             map[string]string
	options         map[string]interface{}
}

// NewGenericCLIProvider creates a new generic CLI provider
func NewGenericCLIProvider(model string, options map[string]interface{}) (*GenericCLIProvider, error) {
	cmdTemplate := getStringOption(options, "command", "")
	if cmdTemplate == "" {
		return nil, fmt.Errorf("generic cli provider requires 'command' option")
	}

	env := make(map[string]string)
	if envMap, ok := options["env"].(map[string]interface{}); ok {
		for k, v := range envMap {
			env[k] = fmt.Sprintf("%v", v)
		}
	}

	return &GenericCLIProvider{
		commandTemplate: cmdTemplate,
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
	Model       string
	Prompt      string
	Temperature float64
	MaxTokens   int
}

// Generate generates a response using the configured CLI command
func (p *GenericCLIProvider) Generate(ctx context.Context, prompt string, options llm.GenerateOptions) (*llm.GenerateResponse, error) {
	// Prepare template data
	data := templateData{
		Model:       p.model,
		Prompt:      prompt,
		Temperature: getFloat64Option(p.options, "temperature", 0.7),
		MaxTokens:   getIntOption(p.options, "max_tokens", getIntOption(p.options, "num_predict", 1000)),
	}

	if options.Temperature != nil {
		data.Temperature = *options.Temperature
	}
	if options.MaxTokens != nil {
		data.MaxTokens = *options.MaxTokens
	}

	// Parse and execute template
	tmpl, err := template.New("cli").Parse(p.commandTemplate)
	if err != nil {
		return nil, fmt.Errorf("failed to parse command template: %w", err)
	}

	var cmdStr bytes.Buffer
	if err := tmpl.Execute(&cmdStr, data); err != nil {
		return nil, fmt.Errorf("failed to execute command template: %w", err)
	}

	// Parse command string into executable and arguments
	// We use shellquote to handle quoted arguments correctly
	parts, err := shellquote.Split(cmdStr.String())
	if err != nil {
		return nil, fmt.Errorf("failed to split command string: %w", err)
	}
	if len(parts) == 0 {
		return nil, fmt.Errorf("empty command resulted from template")
	}

	executable := parts[0]
	args := parts[1:]

	// Look for executable
	if _, err := exec.LookPath(executable); err != nil {
		return nil, fmt.Errorf("executable not found: %s", executable)
	}

	cmd := exec.CommandContext(ctx, executable, args...)

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

	return parseCLIResponse(strings.TrimSpace(out.String()), p.model, latency), nil
}

// EvaluatePrompt implements the legacy Provider interface method
func (p *GenericCLIProvider) EvaluatePrompt(ctx context.Context, prompt string, vars map[string]interface{}) (*promptfoo.ProviderResponse, error) {
	// Apply variables to prompt if needed
	finalPrompt := prompt
	if len(vars) > 0 {
		for key, value := range vars {
			placeholder := fmt.Sprintf("{{%s}}", key)
			finalPrompt = strings.ReplaceAll(finalPrompt, placeholder, fmt.Sprintf("%v", value))
			placeholder = fmt.Sprintf("{{.%s}}", key)
			finalPrompt = strings.ReplaceAll(finalPrompt, placeholder, fmt.Sprintf("%v", value))
		}
	}

	result, err := p.Generate(ctx, finalPrompt, llm.GenerateOptions{})
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
