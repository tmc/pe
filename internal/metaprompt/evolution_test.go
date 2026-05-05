package metaprompt

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/tmc/pe/internal/llm"
)

type evolutionProvider struct {
	err   error
	calls []string
}

func (p *evolutionProvider) Generate(ctx context.Context, prompt string, options llm.GenerateOptions) (*llm.GenerateResponse, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if p.err != nil {
		return nil, p.err
	}
	p.calls = append(p.calls, prompt)
	text := "improved prompt with details"
	if strings.Contains(prompt, "Rate the prompt") {
		text = "0.8"
	}
	if strings.Contains(prompt, "variant") {
		text = "variant prompt with specific structure"
	}
	if strings.Contains(prompt, "hybrid") {
		text = "hybrid prompt"
	}
	if strings.Contains(prompt, "Rephrase") || strings.Contains(prompt, "Expand") || strings.Contains(prompt, "Make this prompt") || strings.Contains(prompt, "Enhance") || strings.Contains(prompt, "Simplify") {
		text = "mutated prompt"
	}
	return &llm.GenerateResponse{Text: text}, nil
}

func TestEvolutionaryOptimizerEvolve(t *testing.T) {
	provider := &evolutionProvider{}
	cfg := EvolutionConfig{
		PopulationSize: 3,
		Generations:    1,
		MutationRate:   1,
		CrossoverRate:  1,
		ElitismRate:    0.34,
		Objectives:     []string{"clarity", "specificity"},
	}
	optimizer := NewEvolutionaryOptimizer(provider, cfg)
	result, err := optimizer.Evolve(context.Background(), "base prompt", cfg)
	if err != nil {
		t.Fatal(err)
	}
	if result.BestIndividual.Prompt == "" || result.FinalPopulation.Generation != 1 {
		t.Fatalf("result = %#v", result)
	}
	if len(result.EvolutionHistory) != 1 || len(result.ParetoFrontier) == 0 {
		t.Fatalf("history/frontier = %#v", result)
	}
	if len(provider.calls) == 0 {
		t.Fatal("provider was not called")
	}
}

func TestEvolutionaryOptimizerHelpers(t *testing.T) {
	optimizer := NewEvolutionaryOptimizer(&evolutionProvider{}, EvolutionConfig{PopulationSize: 2, Objectives: []string{"clarity"}})
	pop, err := optimizer.initializePopulation(context.Background(), "base prompt")
	if err != nil {
		t.Fatal(err)
	}
	if len(pop.Individuals) != 2 || pop.Individuals[0].Genealogy[0] != "base" {
		t.Fatalf("population = %#v", pop)
	}
	fitness, objectives, err := optimizer.evaluateFitness(context.Background(), "Analyze this detailed prompt.")
	if err != nil {
		t.Fatal(err)
	}
	if fitness <= 0 || objectives["clarity"] == 0 {
		t.Fatalf("fitness = %v objectives = %#v", fitness, objectives)
	}
	if score := optimizer.heuristicEvaluation("Analyze this detailed prompt."); score <= 0.5 {
		t.Fatalf("heuristic score = %v", score)
	}
	if !contains("long enough text", []string{"x"}) {
		t.Fatal("contains helper returned false")
	}
	if optimizer.calculatePromptDistance("abc", "abc") != 0 {
		t.Fatal("same prompt distance not zero")
	}
	if optimizer.calculatePromptDistance("", "") != 0 {
		t.Fatal("empty prompt distance not zero")
	}
	if optimizer.calculatePopulationDiversity([]Individual{{Prompt: "a"}}) != 0 {
		t.Fatal("single population diversity not zero")
	}
	pop.Individuals[0].Fitness = 0.2
	pop.Individuals[1].Fitness = 0.9
	optimizer.updatePopulationStats(&pop)
	if pop.BestFitness != 0.9 || pop.AvgFitness <= 0 || pop.Diversity < 0 {
		t.Fatalf("population stats = %#v", pop)
	}
	if !optimizer.hasConverged([]float64{.9, .9, .9, .9, .9}, 10) {
		t.Fatal("expected convergence")
	}
	if optimizer.hasConverged([]float64{.1, .2, .3}, 3) {
		t.Fatal("early convergence")
	}
}

func TestEvolutionaryOptimizerOperators(t *testing.T) {
	optimizer := NewEvolutionaryOptimizer(&evolutionProvider{}, EvolutionConfig{PopulationSize: 4, MutationRate: 1, CrossoverRate: 1, ElitismRate: 0.25})
	parent1 := Individual{Prompt: "parent one", Fitness: 0.9, Genealogy: []string{"p1"}, Objectives: map[string]float64{"a": 0.9}}
	parent2 := Individual{Prompt: "parent two", Fitness: 0.8, Genealogy: []string{"p2"}, Objectives: map[string]float64{"a": 0.8}}
	offspring, err := optimizer.crossover(context.Background(), parent1, parent2)
	if err != nil || offspring.Prompt == "" || len(offspring.Genealogy) != 2 {
		t.Fatalf("offspring = %#v err=%v", offspring, err)
	}
	mutated, err := optimizer.mutate(context.Background(), parent1)
	if err != nil || mutated.Prompt == parent1.Prompt || len(mutated.Mutations) != 1 {
		t.Fatalf("mutated = %#v err=%v", mutated, err)
	}
	for _, op := range []MutationOperator{Rephrase, Expand, Prune, Enhance, Simplify, MutationOperator("other")} {
		if prompt := optimizer.createMutationPrompt("base", op); !strings.Contains(prompt, "base") {
			t.Fatalf("mutation prompt for %s = %q", op, prompt)
		}
	}
	next, err := optimizer.createNextGeneration(context.Background(), Population{Individuals: []Individual{parent1, parent2}})
	if err != nil {
		t.Fatal(err)
	}
	if len(next.Individuals) != 4 || next.Generation != 1 {
		t.Fatalf("next = %#v", next)
	}
	frontier := optimizer.extractParetoFrontier([]Individual{parent1, parent2})
	if len(frontier) != 1 || frontier[0].Prompt != parent1.Prompt {
		t.Fatalf("frontier = %#v", frontier)
	}
	if !optimizer.dominatesMultiObjective(parent1, parent2) || optimizer.dominatesMultiObjective(parent2, parent1) {
		t.Fatal("dominance failed")
	}
}

func TestEvolutionaryOptimizerErrors(t *testing.T) {
	boom := errors.New("boom")
	optimizer := NewEvolutionaryOptimizer(&evolutionProvider{err: boom}, EvolutionConfig{PopulationSize: 1, Objectives: []string{"clarity"}})
	if _, err := optimizer.generateVariant(context.Background(), "base", 1); !errors.Is(err, boom) {
		t.Fatalf("variant error = %v", err)
	}
	if score, err := optimizer.evaluateObjective(context.Background(), "base", "clarity"); !errors.Is(err, boom) || score != 0.5 {
		t.Fatalf("objective = %v err=%v", score, err)
	}
	individual := Individual{Prompt: "x", Genealogy: []string{"x"}, Objectives: map[string]float64{}}
	if got, err := optimizer.crossover(context.Background(), individual, individual); !errors.Is(err, boom) || got.Prompt != individual.Prompt {
		t.Fatalf("crossover fallback = %#v err=%v", got, err)
	}
	if got, err := optimizer.mutate(context.Background(), individual); !errors.Is(err, boom) || got.Prompt != individual.Prompt {
		t.Fatalf("mutate fallback = %#v err=%v", got, err)
	}
}
