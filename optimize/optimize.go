package optimize

import (
	"context"
	"fmt"
	"math"
	"time"
)

// Variant is one prompt candidate in an optimization run.
type Variant struct {
	Prompt   string
	Source   string
	Metadata map[string]string
}

// Score is a scalar score plus optional diagnostic data.
type Score struct {
	Value   float64
	Reason  string
	Metrics map[string]float64
}

// Scorer assigns a score to a prompt variant.
type Scorer interface {
	Score(ctx context.Context, variant Variant) (Score, error)
}

// ScorerFunc adapts a function to [Scorer].
type ScorerFunc func(context.Context, Variant) (Score, error)

// Score calls f(ctx, variant).
func (f ScorerFunc) Score(ctx context.Context, variant Variant) (Score, error) {
	if f == nil {
		return Score{}, fmt.Errorf("optimize: nil scorer func")
	}
	return f(ctx, variant)
}

// Refiner returns candidate variants for the current best variant.
type Refiner interface {
	Refine(ctx context.Context, current Variant) ([]Variant, error)
}

// RefinerFunc adapts a function to [Refiner].
type RefinerFunc func(context.Context, Variant) ([]Variant, error)

// Refine calls f(ctx, current).
func (f RefinerFunc) Refine(ctx context.Context, current Variant) ([]Variant, error) {
	if f == nil {
		return nil, fmt.Errorf("optimize: nil refiner func")
	}
	return f(ctx, current)
}

// Config controls optimization.
type Config struct {
	MaxRounds      int
	MinImprovement float64
}

// Step records one scored variant.
type Step struct {
	Round       int
	Variant     Variant
	Score       Score
	Improvement float64
	Accepted    bool
}

// Result records the full optimization trajectory.
type Result struct {
	Initial   Step
	Best      Step
	Steps     []Step
	Converged bool
	Duration  time.Duration
}

// Optimizer runs deterministic local refinement loops.
type Optimizer struct {
	scorer  Scorer
	refiner Refiner
	config  Config
}

// New returns an Optimizer.
func New(scorer Scorer, refiner Refiner, config Config) (*Optimizer, error) {
	if scorer == nil {
		return nil, fmt.Errorf("optimize: nil scorer")
	}
	if refiner == nil {
		return nil, fmt.Errorf("optimize: nil refiner")
	}
	if config.MaxRounds < 0 {
		return nil, fmt.Errorf("optimize: negative max rounds")
	}
	if config.MaxRounds == 0 {
		config.MaxRounds = 1
	}
	if config.MinImprovement < 0 {
		return nil, fmt.Errorf("optimize: negative min improvement")
	}
	return &Optimizer{
		scorer:  scorer,
		refiner: refiner,
		config:  config,
	}, nil
}

// Optimize scores seed and repeatedly accepts the best improving refinement.
func (o *Optimizer) Optimize(ctx context.Context, seed Variant) (*Result, error) {
	start := time.Now()
	initialScore, err := o.score(ctx, seed)
	if err != nil {
		return nil, fmt.Errorf("score seed: %w", err)
	}

	initial := Step{
		Round:   0,
		Variant: seed,
		Score:   initialScore,
	}
	result := &Result{
		Initial: initial,
		Best:    initial,
		Steps:   []Step{initial},
	}

	current := initial
	for round := 1; round <= o.config.MaxRounds; round++ {
		candidates, err := o.refiner.Refine(ctx, current.Variant)
		if err != nil {
			return nil, fmt.Errorf("refine round %d: %w", round, err)
		}
		if len(candidates) == 0 {
			result.Converged = true
			break
		}

		bestRound := current
		for _, candidate := range candidates {
			score, err := o.score(ctx, candidate)
			if err != nil {
				return nil, fmt.Errorf("score round %d candidate %q: %w", round, candidate.Source, err)
			}
			step := Step{
				Round:       round,
				Variant:     candidate,
				Score:       score,
				Improvement: score.Value - current.Score.Value,
			}
			if score.Value > bestRound.Score.Value {
				bestRound = step
			}
			result.Steps = append(result.Steps, step)
		}

		improvement := bestRound.Score.Value - current.Score.Value
		if improvement <= o.config.MinImprovement {
			result.Converged = true
			break
		}
		bestRound.Accepted = true
		bestRound.Improvement = improvement
		result.Steps = append(result.Steps, bestRound)
		result.Best = bestRound
		current = bestRound
	}

	result.Duration = time.Since(start)
	return result, nil
}

func (o *Optimizer) score(ctx context.Context, variant Variant) (Score, error) {
	score, err := o.scorer.Score(ctx, variant)
	if err != nil {
		return Score{}, err
	}
	if math.IsNaN(score.Value) || math.IsInf(score.Value, 0) {
		return Score{}, fmt.Errorf("invalid score %v", score.Value)
	}
	return score, nil
}
