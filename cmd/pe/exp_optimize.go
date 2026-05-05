package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/tmc/pe/internal/optimization/localopt"
)

type expOptimizeVariant struct {
	Prompt   string            `json:"prompt"`
	Source   string            `json:"source,omitempty"`
	Score    *float64          `json:"score,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

type expOptimizeInput struct {
	Seed           expOptimizeVariant     `json:"seed"`
	Variants       []expOptimizeVariant   `json:"variants,omitempty"`
	Rounds         [][]expOptimizeVariant `json:"rounds,omitempty"`
	MaxRounds      int                    `json:"max_rounds,omitempty"`
	MinImprovement float64                `json:"min_improvement,omitempty"`
}

type expOptimizeOutput struct {
	Selected   expOptimizeOutputStep   `json:"selected"`
	Initial    expOptimizeOutputStep   `json:"initial"`
	Trajectory []expOptimizeOutputStep `json:"trajectory"`
	Converged  bool                    `json:"converged"`
	Iterations int                     `json:"iterations"`
}

type expOptimizeOutputStep struct {
	Round       int                `json:"round"`
	Source      string             `json:"source"`
	Prompt      string             `json:"prompt"`
	Score       float64            `json:"score"`
	Reason      string             `json:"reason,omitempty"`
	Metrics     map[string]float64 `json:"metrics,omitempty"`
	Improvement float64            `json:"improvement,omitempty"`
	Accepted    bool               `json:"accepted,omitempty"`
}

func expOptimizeCmd() *cobra.Command {
	var inputPath string
	var outputPath string
	var maxRounds int
	var minImprovement float64

	cmd := &cobra.Command{
		Use:   "optimize",
		Short: "Select prompt variants with local deterministic scores",
		RunE: func(cmd *cobra.Command, args []string) error {
			in, err := readExpOptimizeInput(cmd.InOrStdin(), inputPath)
			if err != nil {
				return err
			}
			if maxRounds > 0 {
				in.MaxRounds = maxRounds
			}
			if cmd.Flags().Changed("min-improvement") {
				in.MinImprovement = minImprovement
			}

			out, err := runExpOptimize(cmd.Context(), in)
			if err != nil {
				return err
			}
			return writeExpOptimizeOutput(cmd.OutOrStdout(), outputPath, out)
		},
	}
	cmd.Flags().StringVarP(&inputPath, "input", "i", "-", "input JSON file, or - for stdin")
	cmd.Flags().StringVarP(&outputPath, "output", "o", "-", "output JSON file, or - for stdout")
	cmd.Flags().IntVar(&maxRounds, "max-rounds", 0, "override input max_rounds")
	cmd.Flags().Float64Var(&minImprovement, "min-improvement", 0, "override input min_improvement")
	return cmd
}

func readExpOptimizeInput(stdin io.Reader, path string) (*expOptimizeInput, error) {
	var r io.Reader = stdin
	if path != "" && path != "-" {
		f, err := os.Open(path)
		if err != nil {
			return nil, fmt.Errorf("open optimize input: %w", err)
		}
		defer f.Close()
		r = f
	}

	var in expOptimizeInput
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&in); err != nil {
		return nil, fmt.Errorf("decode optimize input: %w", err)
	}
	return &in, nil
}

func writeExpOptimizeOutput(stdout io.Writer, path string, out *expOptimizeOutput) error {
	var w io.Writer = stdout
	if path != "" && path != "-" {
		f, err := os.Create(path)
		if err != nil {
			return fmt.Errorf("create optimize output: %w", err)
		}
		defer f.Close()
		w = f
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(out); err != nil {
		return fmt.Errorf("encode optimize output: %w", err)
	}
	return nil
}

func runExpOptimize(ctx context.Context, in *expOptimizeInput) (*expOptimizeOutput, error) {
	if in == nil {
		return nil, fmt.Errorf("optimize input is nil")
	}
	rounds, scores, err := expOptimizeRounds(in)
	if err != nil {
		return nil, err
	}
	seed, err := expOptimizeLocalVariant("seed", in.Seed)
	if err != nil {
		return nil, err
	}
	if _, ok := scores[expOptimizeVariantKey(seed)]; !ok {
		return nil, fmt.Errorf("seed score is required")
	}

	refiner := &expOptimizeRefiner{rounds: rounds}
	scorer := localopt.ScorerFunc(func(ctx context.Context, variant localopt.Variant) (localopt.Score, error) {
		score, ok := scores[expOptimizeVariantKey(variant)]
		if !ok {
			return localopt.Score{}, fmt.Errorf("score missing for %q", variant.Source)
		}
		return localopt.Score{
			Value:  score,
			Reason: "provided local score",
			Metrics: map[string]float64{
				"provided": score,
			},
		}, nil
	})
	opt, err := localopt.New(scorer, refiner, localopt.Config{
		MaxRounds:      in.MaxRounds,
		MinImprovement: in.MinImprovement,
	})
	if err != nil {
		return nil, err
	}
	result, err := opt.Optimize(ctx, seed)
	if err != nil {
		return nil, err
	}

	out := &expOptimizeOutput{
		Selected:   expOptimizeStep(result.Best),
		Initial:    expOptimizeStep(result.Initial),
		Trajectory: make([]expOptimizeOutputStep, 0, len(result.Steps)),
		Converged:  result.Converged,
		Iterations: refiner.calls,
	}
	for _, step := range result.Steps {
		out.Trajectory = append(out.Trajectory, expOptimizeStep(step))
	}
	return out, nil
}

func expOptimizeRounds(in *expOptimizeInput) ([][]localopt.Variant, map[string]float64, error) {
	scores := make(map[string]float64)
	if err := expOptimizeAddScore(scores, "seed", in.Seed); err != nil {
		return nil, nil, err
	}

	rawRounds := in.Rounds
	if len(rawRounds) == 0 && len(in.Variants) > 0 {
		rawRounds = [][]expOptimizeVariant{in.Variants}
	}
	if len(rawRounds) == 0 {
		return nil, nil, fmt.Errorf("at least one variant or round is required")
	}

	rounds := make([][]localopt.Variant, 0, len(rawRounds))
	for i, rawRound := range rawRounds {
		if len(rawRound) == 0 {
			return nil, nil, fmt.Errorf("round %d has no variants", i+1)
		}
		round := make([]localopt.Variant, 0, len(rawRound))
		for j, raw := range rawRound {
			source := raw.Source
			if source == "" {
				source = "round-" + strconv.Itoa(i+1) + "-variant-" + strconv.Itoa(j+1)
			}
			if err := expOptimizeAddScore(scores, source, raw); err != nil {
				return nil, nil, err
			}
			variant, err := expOptimizeLocalVariant(source, raw)
			if err != nil {
				return nil, nil, err
			}
			round = append(round, variant)
		}
		rounds = append(rounds, round)
	}
	return rounds, scores, nil
}

func expOptimizeAddScore(scores map[string]float64, source string, raw expOptimizeVariant) error {
	variant, err := expOptimizeLocalVariant(source, raw)
	if err != nil {
		return err
	}
	if raw.Score == nil {
		return fmt.Errorf("score missing for %q", variant.Source)
	}
	scores[expOptimizeVariantKey(variant)] = *raw.Score
	return nil
}

func expOptimizeLocalVariant(defaultSource string, raw expOptimizeVariant) (localopt.Variant, error) {
	if raw.Prompt == "" {
		return localopt.Variant{}, fmt.Errorf("prompt missing for %q", defaultSource)
	}
	source := raw.Source
	if source == "" {
		source = defaultSource
	}
	return localopt.Variant{
		Prompt:   raw.Prompt,
		Source:   source,
		Metadata: raw.Metadata,
	}, nil
}

func expOptimizeVariantKey(variant localopt.Variant) string {
	return variant.Source + "\x00" + variant.Prompt
}

func expOptimizeStep(step localopt.Step) expOptimizeOutputStep {
	return expOptimizeOutputStep{
		Round:       step.Round,
		Source:      step.Variant.Source,
		Prompt:      step.Variant.Prompt,
		Score:       step.Score.Value,
		Reason:      step.Score.Reason,
		Metrics:     step.Score.Metrics,
		Improvement: step.Improvement,
		Accepted:    step.Accepted,
	}
}

type expOptimizeRefiner struct {
	rounds [][]localopt.Variant
	calls  int
}

func (r *expOptimizeRefiner) Refine(ctx context.Context, current localopt.Variant) ([]localopt.Variant, error) {
	if r.calls >= len(r.rounds) {
		return nil, nil
	}
	round := r.rounds[r.calls]
	r.calls++
	return round, nil
}
