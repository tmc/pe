package metrics

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/tmc/pe/internal/llm"
	"github.com/tmc/pe/internal/promptfoo"
)

func TestPassNStoreBasic(t *testing.T) {
	// Create temporary directory for testing
	tmpDir := t.TempDir()

	store, err := NewPassNStore(tmpDir)
	if err != nil {
		t.Fatalf("NewPassNStore() error = %v", err)
	}
	if store == nil {
		t.Fatal("NewPassNStore returned nil")
	}

	// Test storing an evaluation
	evaluation := PassNEvaluation{
		ID:         "eval-test",
		Timestamp:  time.Now(),
		N:          1,
		NumSamples: 10,
		NumPassed:  8,
		PassRate:   0.8,
	}

	err = store.StoreEvaluation("prompt-1", "Test prompt", "provider-1", evaluation)
	if err != nil {
		t.Errorf("StoreEvaluation() error = %v", err)
	}

	// Test retrieving evaluations
	metadata, err := store.GetEvaluations("prompt-1")
	if err != nil {
		t.Errorf("GetEvaluations() error = %v", err)
	}

	if metadata == nil {
		t.Fatal("GetEvaluations() returned nil")
	}

	if len(metadata.Evaluations) == 0 {
		t.Error("No evaluations found")
	}
}

func TestPassNReport(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := NewPassNStore(tmpDir)
	if err != nil {
		t.Fatalf("NewPassNStore() error = %v", err)
	}

	// Store some evaluations
	for i := 0; i < 3; i++ {
		evaluation := PassNEvaluation{
			ID:         fmt.Sprintf("eval-%d", i),
			Timestamp:  time.Now(),
			N:          1,
			NumSamples: 20,
			NumPassed:  15 + i,
			PassRate:   float64(15+i) / 20.0,
		}

		err = store.StoreEvaluation("prompt-report", "Test prompt", "provider-1", evaluation)
		if err != nil {
			t.Errorf("StoreEvaluation() error = %v", err)
		}
	}

	// Generate report
	report, err := store.GeneratePassNReport("prompt-report")
	if err != nil {
		t.Errorf("GeneratePassNReport() error = %v", err)
	}

	if report == nil {
		t.Fatal("GeneratePassNReport() returned nil")
	}

	if report.TotalEvals != 3 {
		t.Errorf("TotalEvals = %d, want 3", report.TotalEvals)
	}
}

func TestPassNStoreQueriesAndTrends(t *testing.T) {
	store, err := NewPassNStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	evals := []PassNEvaluation{
		{ID: "e1", N: 1, NumSamples: 10, PassRate: 0.4},
		{ID: "e2", N: 2, NumSamples: 10, PassRate: 0.7},
		{ID: "e3", N: 2, NumSamples: 10, PassRate: 0.9},
	}
	for _, eval := range evals {
		if err := store.StoreEvaluation("prompt", "short prompt", "provider", eval); err != nil {
			t.Fatal(err)
		}
	}
	meta, err := store.GetEvaluations("prompt")
	if err != nil {
		t.Fatal(err)
	}
	meta.Tags = []string{"go", "eval"}
	best, err := store.GetBestEvaluation("prompt")
	if err != nil || best.ID != "e3" {
		t.Fatalf("best = %#v err=%v", best, err)
	}
	found, err := store.SearchByTags([]string{"go"})
	if err != nil || len(found) != 1 {
		t.Fatalf("found = %#v err=%v", found, err)
	}
	found, err = store.SearchByTags([]string{"go", "missing"})
	if err != nil || len(found) != 0 {
		t.Fatalf("missing found = %#v err=%v", found, err)
	}
	report, err := store.GeneratePassNReport("prompt")
	if err != nil {
		t.Fatal(err)
	}
	if report.BestPassRate != 0.9 || report.AvgPassRate <= 0 || report.TotalSamples != 30 || report.Trend != "stable" {
		t.Fatalf("report = %#v", report)
	}
	if _, err := store.GetEvaluations("missing"); err == nil {
		t.Fatal("missing evaluations succeeded")
	}
	store.metadata["empty"] = &PassNMetadata{PromptID: "empty"}
	if _, err := store.GetBestEvaluation("empty"); err == nil {
		t.Fatal("empty best evaluation succeeded")
	}
	if _, err := store.GeneratePassNReport("missing"); err == nil {
		t.Fatal("missing report succeeded")
	}
}

func TestPassNStoreLoadAndSaveErrors(t *testing.T) {
	tmpDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmpDir, "metadata.json"), []byte("{"), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := NewPassNStore(tmpDir); err == nil {
		t.Fatal("bad metadata loaded")
	}
	blocker := filepath.Join(t.TempDir(), "blocker")
	if err := os.WriteFile(blocker, []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := NewPassNStore(filepath.Join(blocker, "child")); err == nil {
		t.Fatal("mkdir through file succeeded")
	}
	store := &PassNStore{dataDir: filepath.Join(t.TempDir(), "missing"), metadata: map[string]*PassNMetadata{}}
	if err := store.saveMetadata("prompt", &PassNMetadata{}); err == nil {
		t.Fatal("save to missing directory succeeded")
	}
}

func TestPassNGeneratorStrategies(t *testing.T) {
	provider := &passNProvider{text: "same words"}
	gen := NewPassNGenerator(provider, PassNConfig{MaxTokens: 8})
	if gen.config.Temperature != 0.8 || gen.config.TopP != 0.95 || gen.config.SamplingStrategy != "random" {
		t.Fatalf("defaults = %#v", gen.config)
	}
	samples, err := gen.GenerateSamples(context.Background(), "prompt", 4)
	if err != nil || len(samples) != 4 {
		t.Fatalf("random samples=%#v err=%v", samples, err)
	}
	gen = NewPassNGenerator(provider, PassNConfig{SamplingStrategy: "diverse", MaxTokens: 8})
	samples, err = gen.GenerateSamples(context.Background(), "prompt", 6)
	if err != nil || len(samples) != 6 {
		t.Fatalf("diverse samples=%#v err=%v", samples, err)
	}
	gen = NewPassNGenerator(&passNProvider{text: "unique content changes every time"}, PassNConfig{SamplingStrategy: "adaptive", Temperature: 0.8, MaxTokens: 8})
	samples, err = gen.GenerateSamples(context.Background(), "prompt", 7)
	if err != nil || len(samples) != 7 {
		t.Fatalf("adaptive samples=%#v err=%v", samples, err)
	}
	want := errors.New("generate failed")
	gen = NewPassNGenerator(&passNProvider{err: want}, PassNConfig{})
	samples, err = gen.GenerateSamples(context.Background(), "prompt", 4)
	if err == nil || !strings.Contains(err.Error(), want.Error()) || len(samples) != 0 {
		t.Fatalf("error samples=%#v err=%v", samples, err)
	}
}

func TestPassNHelpers(t *testing.T) {
	if got := generateHash("abcdef"); got != "bef57ec7f53a6d40beb640a780a639c83bc29ac8a9816f1fc6c5c6dcd93c4721" {
		t.Fatalf("hash short = %q", got)
	}
	if got := generateHash("abcdefghijklmnopq"); got != "918a954ac4dfb54ac39f068d9868227f69ab39bc362e2c9b0083bf6a109d6ad7" {
		t.Fatalf("hash long = %q", got)
	}
	if !containsAllTags([]string{"a", "b"}, []string{"a"}) || containsAllTags([]string{"a"}, []string{"b"}) {
		t.Fatal("tag containment mismatch")
	}
	if calculateTrend(nil) != "" {
		t.Fatal("empty trend should be empty")
	}
	if calculateTrend([]PassNEvaluation{{PassRate: 1}, {PassRate: 0.5}}) != "declining" {
		t.Fatal("declining trend mismatch")
	}
	if calculateTrend([]PassNEvaluation{{PassRate: 1}, {PassRate: 1.05}}) != "stable" {
		t.Fatal("stable trend mismatch")
	}
	if calculateDiversity([]string{"one"}) != 1 || calculateDiversity([]string{"", ""}) != 0 {
		t.Fatal("diversity edge mismatch")
	}
	if got := calculateDiversity([]string{"a b", "a c"}); got <= 0 || got >= 1 {
		t.Fatalf("diversity = %v", got)
	}
}

type passNProvider struct {
	text string
	err  error
}

func (p *passNProvider) Name() string { return "passn" }

func (p *passNProvider) Model() string { return "test" }

func (p *passNProvider) SupportsStreaming() bool { return false }

func (p *passNProvider) SupportsBatch() bool { return false }

func (p *passNProvider) EvaluatePrompt(context.Context, string, map[string]interface{}) (*promptfoo.ProviderResponse, error) {
	return &promptfoo.ProviderResponse{Output: p.text}, p.err
}

func (p *passNProvider) Generate(context.Context, string, llm.GenerateOptions) (*llm.GenerateResponse, error) {
	if p.err != nil {
		return nil, p.err
	}
	return &llm.GenerateResponse{Text: p.text}, nil
}
