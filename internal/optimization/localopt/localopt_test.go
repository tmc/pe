package localopt

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestOptimizerAcceptsBestImprovingVariant(t *testing.T) {
	scorer := ScorerFunc(func(ctx context.Context, variant Variant) (Score, error) {
		score := float64(strings.Count(variant.Prompt, "specific"))
		score += float64(strings.Count(variant.Prompt, "example"))
		return Score{Value: score, Reason: variant.Source}, nil
	})
	refiner := RefinerFunc(func(ctx context.Context, current Variant) ([]Variant, error) {
		return []Variant{
			{Prompt: current.Prompt + " specific", Source: "add-specific"},
			{Prompt: current.Prompt + " example specific", Source: "add-example-specific"},
			{Prompt: current.Prompt, Source: "no-change"},
		}, nil
	})

	opt, err := New(scorer, refiner, Config{MaxRounds: 1, MinImprovement: 0.1})
	if err != nil {
		t.Fatal(err)
	}
	result, err := opt.Optimize(context.Background(), Variant{Prompt: "summarize", Source: "seed"})
	if err != nil {
		t.Fatal(err)
	}

	if got, want := result.Best.Variant.Source, "add-example-specific"; got != want {
		t.Fatalf("best source = %q, want %q", got, want)
	}
	if got, want := result.Best.Score.Value, 2.0; got != want {
		t.Fatalf("best score = %v, want %v", got, want)
	}
	if !result.Best.Accepted {
		t.Fatal("best variant was not marked accepted")
	}
}

func TestOptimizerStopsWhenNoCandidateImprovesEnough(t *testing.T) {
	scorer := ScorerFunc(func(ctx context.Context, variant Variant) (Score, error) {
		return Score{Value: float64(len(variant.Prompt))}, nil
	})
	refiner := RefinerFunc(func(ctx context.Context, current Variant) ([]Variant, error) {
		return []Variant{{Prompt: current.Prompt + "x", Source: "tiny"}}, nil
	})

	opt, err := New(scorer, refiner, Config{MaxRounds: 3, MinImprovement: 2})
	if err != nil {
		t.Fatal(err)
	}
	result, err := opt.Optimize(context.Background(), Variant{Prompt: "seed", Source: "seed"})
	if err != nil {
		t.Fatal(err)
	}

	if !result.Converged {
		t.Fatal("result did not converge")
	}
	if got, want := result.Best.Variant.Prompt, "seed"; got != want {
		t.Fatalf("best prompt = %q, want %q", got, want)
	}
}

func TestOptimizerRejectsTies(t *testing.T) {
	scorer := ScorerFunc(func(ctx context.Context, variant Variant) (Score, error) {
		return Score{Value: 1}, nil
	})
	refiner := RefinerFunc(func(ctx context.Context, current Variant) ([]Variant, error) {
		return []Variant{{Prompt: current.Prompt + " tied", Source: "tie"}}, nil
	})

	opt, err := New(scorer, refiner, Config{MaxRounds: 1})
	if err != nil {
		t.Fatal(err)
	}
	result, err := opt.Optimize(context.Background(), Variant{Prompt: "seed", Source: "seed"})
	if err != nil {
		t.Fatal(err)
	}
	if got, want := result.Best.Variant.Source, "seed"; got != want {
		t.Fatalf("best source = %q, want %q", got, want)
	}
	if !result.Converged {
		t.Fatal("tie should converge")
	}
}

func TestOptimizerReturnsContextualErrors(t *testing.T) {
	wantErr := errors.New("scorer unavailable")
	scorer := ScorerFunc(func(ctx context.Context, variant Variant) (Score, error) {
		return Score{}, wantErr
	})
	refiner := RefinerFunc(func(ctx context.Context, current Variant) ([]Variant, error) {
		t.Fatal("refiner should not be called")
		return nil, nil
	})

	opt, err := New(scorer, refiner, Config{})
	if err != nil {
		t.Fatal(err)
	}
	_, err = opt.Optimize(context.Background(), Variant{Prompt: "seed"})
	if !errors.Is(err, wantErr) {
		t.Fatalf("error = %v, want wrapping %v", err, wantErr)
	}
}

func TestFuncAdaptersRejectNilFunctions(t *testing.T) {
	if _, err := ScorerFunc(nil).Score(context.Background(), Variant{}); err == nil {
		t.Fatal("nil scorer func accepted")
	}
	if _, err := RefinerFunc(nil).Refine(context.Background(), Variant{}); err == nil {
		t.Fatal("nil refiner func accepted")
	}
}

func TestNewValidatesConfig(t *testing.T) {
	if _, err := New(nil, RefinerFunc(nil), Config{}); err == nil {
		t.Fatal("nil scorer accepted")
	}
	if _, err := New(ScorerFunc(nil), nil, Config{}); err == nil {
		t.Fatal("nil refiner accepted")
	}
	if _, err := New(ScorerFunc(nil), RefinerFunc(nil), Config{MaxRounds: -1}); err == nil {
		t.Fatal("negative max rounds accepted")
	}
	if _, err := New(ScorerFunc(nil), RefinerFunc(nil), Config{MinImprovement: -1}); err == nil {
		t.Fatal("negative min improvement accepted")
	}
}
