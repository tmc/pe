package inference_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/tmc/pe/internal/inference"
	"github.com/tmc/pe/internal/llm"
	"github.com/tmc/pe/internal/promptfoo"
)

type legacyProvider struct {
	name   string
	model  string
	prompt string
	vars   map[string]interface{}
}

func (p *legacyProvider) EvaluatePrompt(ctx context.Context, prompt string, vars map[string]interface{}) (*promptfoo.ProviderResponse, error) {
	p.prompt = prompt
	p.vars = vars
	return &promptfoo.ProviderResponse{
		Output: "legacy output",
		TokenUsage: &promptfoo.TokenUsage{
			Prompt:     3,
			Completion: 5,
			Total:      8,
		},
		Cost:   0.25,
		Cached: true,
	}, nil
}

func (p *legacyProvider) Name() string  { return p.name }
func (p *legacyProvider) Model() string { return p.model }
func (p *legacyProvider) Generate(ctx context.Context, prompt string, options llm.GenerateOptions) (*llm.GenerateResponse, error) {
	return nil, errors.New("not used")
}
func (p *legacyProvider) SupportsStreaming() bool { return false }
func (p *legacyProvider) SupportsBatch() bool     { return false }

type modernProvider struct {
	name      string
	models    []string
	lastReq   inference.Request
	streamReq inference.Request
}

func (p *modernProvider) Name() string { return p.name }

func (p *modernProvider) Complete(ctx context.Context, req inference.Request) (*inference.Response, error) {
	p.lastReq = req
	return &inference.Response{
		Content: "modern output",
		Model:   req.Model,
		TokensUsed: inference.TokenUsage{
			PromptTokens:     7,
			CompletionTokens: 11,
			TotalTokens:      18,
		},
		Metadata: map[string]interface{}{
			"cost":       0.75,
			"cached":     true,
			"latency_ms": 42,
			"provider":   "test",
		},
	}, nil
}

func (p *modernProvider) Stream(ctx context.Context, req inference.Request) (<-chan inference.StreamChunk, error) {
	p.streamReq = req
	chunks := make(chan inference.StreamChunk, 3)
	chunks <- inference.StreamChunk{Delta: "one"}
	chunks <- inference.StreamChunk{Delta: "two"}
	chunks <- inference.StreamChunk{Done: true}
	close(chunks)
	return chunks, nil
}

func (p *modernProvider) Models(ctx context.Context) ([]string, error) { return p.models, nil }
func (p *modernProvider) Close() error                                 { return nil }

func TestLegacyAdapterComplete(t *testing.T) {
	legacy := &legacyProvider{name: "legacy", model: "legacy-model"}
	provider := inference.NewLegacyAdapter(legacy)

	resp, err := provider.Complete(context.Background(), inference.Request{
		Prompt:       "hello",
		Model:        "override-model",
		SystemPrompt: "system",
		Prefill:      "prefill",
		Temperature:  0.5,
		MaxTokens:    32,
	})
	if err != nil {
		t.Fatalf("Complete: %v", err)
	}
	if provider.Name() != "legacy" {
		t.Fatalf("Name = %q, want legacy", provider.Name())
	}
	if legacy.prompt != "system\n\nhello\n\nAssistant: prefill" {
		t.Fatalf("prompt = %q", legacy.prompt)
	}
	wantVars := map[string]interface{}{
		"temperature":   float32(0.5),
		"max_tokens":    32,
		"model":         "override-model",
		"system_prompt": "system",
	}
	if !reflect.DeepEqual(legacy.vars, wantVars) {
		t.Fatalf("vars = %#v, want %#v", legacy.vars, wantVars)
	}
	if resp.Content != "legacy output" || resp.Model != "legacy-model" {
		t.Fatalf("response content/model = %q/%q", resp.Content, resp.Model)
	}
	if resp.TokensUsed.PromptTokens != 3 || resp.TokensUsed.CompletionTokens != 5 || resp.TokensUsed.TotalTokens != 8 {
		t.Fatalf("tokens = %#v", resp.TokensUsed)
	}
	if resp.Metadata["cost"] != 0.25 || resp.Metadata["cached"] != true {
		t.Fatalf("metadata = %#v", resp.Metadata)
	}
}

func TestLegacyAdapterStreamAndModels(t *testing.T) {
	legacy := &legacyProvider{name: "other", model: "fallback-model"}
	provider := inference.NewLegacyAdapter(legacy)

	models, err := provider.Models(context.Background())
	if err != nil {
		t.Fatalf("Models: %v", err)
	}
	if !reflect.DeepEqual(models, []string{"fallback-model"}) {
		t.Fatalf("models = %#v", models)
	}

	chunks, err := provider.Stream(context.Background(), inference.Request{Prompt: "hello"})
	if err != nil {
		t.Fatalf("Stream: %v", err)
	}
	var got []inference.StreamChunk
	for chunk := range chunks {
		got = append(got, chunk)
	}
	if len(got) != 2 {
		t.Fatalf("got %d chunks, want 2", len(got))
	}
	if got[0].Delta != "legacy output" || got[0].Done {
		t.Fatalf("first chunk = %#v", got[0])
	}
	if !got[1].Done {
		t.Fatalf("final chunk = %#v", got[1])
	}
}

func TestModernAdapterGenerate(t *testing.T) {
	modern := &modernProvider{name: "modern", models: []string{"model-a"}}
	provider := inference.NewModernAdapter(modern, "adapter-model")
	temp := 0.2
	maxTokens := 64
	topP := 0.9
	topK := 40

	resp, err := provider.Generate(context.Background(), "hello", llm.GenerateOptions{
		Temperature: &temp,
		MaxTokens:   &maxTokens,
		TopP:        &topP,
		TopK:        &topK,
		Stop:        []string{"STOP"},
		ProviderOptions: map[string]interface{}{
			"top_p": 0.8,
			"extra": "value",
		},
	})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if provider.Name() != "modern" || provider.Model() != "adapter-model" {
		t.Fatalf("name/model = %q/%q", provider.Name(), provider.Model())
	}
	if modern.lastReq.Prompt != "hello" || modern.lastReq.Model != "adapter-model" {
		t.Fatalf("request prompt/model = %q/%q", modern.lastReq.Prompt, modern.lastReq.Model)
	}
	if modern.lastReq.Temperature != float32(temp) || modern.lastReq.MaxTokens != maxTokens {
		t.Fatalf("request temp/max = %v/%v", modern.lastReq.Temperature, modern.lastReq.MaxTokens)
	}
	if !reflect.DeepEqual(modern.lastReq.StopSequences, []string{"STOP"}) {
		t.Fatalf("stop = %#v", modern.lastReq.StopSequences)
	}
	if modern.lastReq.Options["top_p"] != 0.8 || modern.lastReq.Options["top_k"] != topK || modern.lastReq.Options["extra"] != "value" {
		t.Fatalf("options = %#v", modern.lastReq.Options)
	}
	if resp.Text != "modern output" || resp.Model != "adapter-model" || resp.FinishReason != "stop" {
		t.Fatalf("response text/model/finish = %q/%q/%q", resp.Text, resp.Model, resp.FinishReason)
	}
	if resp.PromptTokens != 7 || resp.CompletionTokens != 11 || resp.TotalTokens != 18 || resp.Cost != 0.75 {
		t.Fatalf("response usage/cost = %#v", resp)
	}
}

func TestModernAdapterEvaluatePromptAndStream(t *testing.T) {
	modern := &modernProvider{name: "modern", models: []string{"model-a"}}
	provider := inference.NewModernAdapter(modern, "adapter-model")

	resp, err := provider.EvaluatePrompt(context.Background(), "prompt", map[string]interface{}{
		"temperature":   0.4,
		"max_tokens":    12,
		"model":         "request-model",
		"system_prompt": "system",
		"custom":        "kept",
	})
	if err != nil {
		t.Fatalf("EvaluatePrompt: %v", err)
	}
	if modern.lastReq.Prompt != "prompt" || modern.lastReq.Model != "request-model" || modern.lastReq.SystemPrompt != "system" {
		t.Fatalf("request = %#v", modern.lastReq)
	}
	if modern.lastReq.Temperature != 0.4 || modern.lastReq.MaxTokens != 12 || modern.lastReq.Options["custom"] != "kept" {
		t.Fatalf("request options = %#v", modern.lastReq)
	}
	if resp.Output != "modern output" || resp.Cost != 0.75 || !resp.Cached || resp.LatencyMs != 42 {
		t.Fatalf("response = %#v", resp)
	}
	if resp.TokenUsage.Prompt != 7 || resp.TokenUsage.Completion != 11 || resp.TokenUsage.Total != 18 {
		t.Fatalf("token usage = %#v", resp.TokenUsage)
	}

	streaming, ok := provider.(llm.StreamingProvider)
	if !ok {
		t.Fatalf("modern adapter does not implement StreamingProvider")
	}
	chunks, err := streaming.GenerateStream(context.Background(), "stream", llm.GenerateOptions{})
	if err != nil {
		t.Fatalf("GenerateStream: %v", err)
	}
	var text string
	var done bool
	for chunk := range chunks {
		text += chunk.Text
		done = done || chunk.Done
	}
	if text != "onetwo" || !done {
		t.Fatalf("stream text/done = %q/%v", text, done)
	}
	if !modern.streamReq.Stream || modern.streamReq.Prompt != "stream" || modern.streamReq.Model != "adapter-model" {
		t.Fatalf("stream request = %#v", modern.streamReq)
	}
}

func TestAdapterUnwrapHelpers(t *testing.T) {
	legacy := &legacyProvider{name: "legacy", model: "legacy-model"}
	modern := inference.MigrateProvider(legacy)
	if inference.GetLegacyProvider(modern, "ignored") != legacy {
		t.Fatalf("GetLegacyProvider did not unwrap LegacyAdapter")
	}

	modernProvider := &modernProvider{name: "modern", models: []string{"model-a"}}
	legacyAdapter := inference.GetLegacyProvider(modernProvider, "model-a")
	if inference.MigrateProvider(legacyAdapter) != modernProvider {
		t.Fatalf("MigrateProvider did not unwrap ModernAdapter")
	}
}

func TestRegisterProviderSpec(t *testing.T) {
	const spec = "migration-test-provider"
	inference.Register(spec, func(config map[string]interface{}) (inference.Provider, error) {
		if config["model"] != "test-model" {
			t.Fatalf("config = %#v", config)
		}
		return &modernProvider{name: spec, models: []string{"test-model"}}, nil
	})

	client := inference.NewClient()
	provider, err := inference.RegisterProviderSpec(client, spec, map[string]interface{}{
		"model": "test-model",
	})
	if err != nil {
		t.Fatalf("RegisterProviderSpec: %v", err)
	}
	if provider.Name() != spec {
		t.Fatalf("provider name = %q, want %q", provider.Name(), spec)
	}
	models, err := client.Models(context.Background(), spec)
	if err != nil {
		t.Fatalf("Models: %v", err)
	}
	if !reflect.DeepEqual(models, []string{"test-model"}) {
		t.Fatalf("models = %#v", models)
	}
}

func TestRegisterProviderSpecRejectsNilClient(t *testing.T) {
	if _, err := inference.RegisterProviderSpec(nil, "unused", nil); err == nil {
		t.Fatalf("RegisterProviderSpec accepted nil client")
	}
}

func TestCompatibilityLayer(t *testing.T) {
	legacy := &legacyProvider{name: "legacy", model: "legacy-model"}
	modern, err := inference.AsProvider(legacy)
	if err != nil {
		t.Fatalf("AsProvider legacy: %v", err)
	}
	if modern.Name() != "legacy" {
		t.Fatalf("modern name = %q", modern.Name())
	}
	if got, err := inference.AsProvider(modern); err != nil || got != modern {
		t.Fatalf("AsProvider modern = %v, %v", got, err)
	}

	legacyAgain, err := inference.AsLegacyProvider(modern, "legacy-model")
	if err != nil {
		t.Fatalf("AsLegacyProvider modern: %v", err)
	}
	if legacyAgain != legacy {
		t.Fatalf("AsLegacyProvider did not unwrap legacy provider")
	}
	if got, err := inference.AsLegacyProvider(legacy, "ignored"); err != nil || got != legacy {
		t.Fatalf("AsLegacyProvider legacy = %v, %v", got, err)
	}
}

func TestCompatibilityLayerRejectsUnknownTypes(t *testing.T) {
	if _, err := inference.AsProvider("not a provider"); err == nil {
		t.Fatalf("AsProvider accepted string")
	}
	if _, err := inference.AsLegacyProvider("not a provider", ""); err == nil {
		t.Fatalf("AsLegacyProvider accepted string")
	}
}
