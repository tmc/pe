package optimize_test

import (
	"context"
	"fmt"
	"strings"

	"github.com/tmc/pe/optimize"
)

func Example() {
	scorer := optimize.ScorerFunc(func(ctx context.Context, v optimize.Variant) (optimize.Score, error) {
		return optimize.Score{Value: float64(strings.Count(v.Prompt, "specific"))}, nil
	})
	refiner := optimize.RefinerFunc(func(ctx context.Context, current optimize.Variant) ([]optimize.Variant, error) {
		return []optimize.Variant{{Prompt: current.Prompt + " specific", Source: "add-specific"}}, nil
	})

	opt, _ := optimize.New(scorer, refiner, optimize.Config{MaxRounds: 1})
	res, _ := opt.Optimize(context.Background(), optimize.Variant{Prompt: "summarize", Source: "seed"})
	fmt.Println(res.Best.Variant.Prompt)
	// Output: summarize specific
}
