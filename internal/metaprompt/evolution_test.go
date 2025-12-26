package metaprompt

import (
	"testing"
	"time"
)

func TestNewEvolutionaryOptimizer(t *testing.T) {
	provider := &mockProvider{}
	config := EvolutionConfig{
		PopulationSize: 20,
		MutationRate:   0.1,
		CrossoverRate:  0.7,
	}
	optimizer := NewEvolutionaryOptimizer(provider, config)

	if optimizer == nil {
		t.Fatal("NewEvolutionaryOptimizer returned nil")
	}
	if optimizer.provider == nil {
		t.Error("EvolutionaryOptimizer has nil provider")
	}
}

func TestEvolutionConfigStruct(t *testing.T) {
	config := EvolutionConfig{
		PopulationSize: 20,
		Generations:    10,
		MutationRate:   0.1,
		CrossoverRate:  0.7,
		ElitismRate:    0.1,
		Objectives:     []string{"clarity", "specificity"},
		FitnessMetrics: []string{"length", "structure"},
	}

	if config.PopulationSize != 20 {
		t.Error("EvolutionConfig.PopulationSize mismatch")
	}
	if len(config.Objectives) != 2 {
		t.Error("EvolutionConfig.Objectives length mismatch")
	}
}

func TestEvolutionResultStruct(t *testing.T) {
	result := EvolutionResult{
		BestIndividual: Individual{
			Prompt:  "optimized prompt",
			Fitness: 0.92,
		},
		FinalPopulation:  Population{},
		EvolutionHistory: []Population{},
		ParetoFrontier:   []Individual{},
		ConvergenceData:  []float64{0.5, 0.6, 0.7},
		Config:           EvolutionConfig{},
	}

	if result.BestIndividual.Fitness != 0.92 {
		t.Error("EvolutionResult.BestIndividual.Fitness mismatch")
	}
}

func TestIndividualStruct(t *testing.T) {
	individual := Individual{
		Prompt:     "test prompt",
		Fitness:    0.85,
		Objectives: map[string]float64{"clarity": 0.9},
		Generation: 5,
		Genealogy:  []string{"parent1"},
		Mutations:  []MutationRecord{},
	}

	if individual.Fitness != 0.85 {
		t.Error("Individual.Fitness mismatch")
	}
}

func TestPopulationStruct(t *testing.T) {
	pop := Population{
		Individuals: []Individual{
			{Prompt: "prompt1", Fitness: 0.8},
			{Prompt: "prompt2", Fitness: 0.7},
		},
		Generation:  5,
		BestFitness: 0.8,
		AvgFitness:  0.75,
		Diversity:   0.6,
	}

	if len(pop.Individuals) != 2 {
		t.Error("Population.Individuals length mismatch")
	}
}

func TestMutationRecordStruct(t *testing.T) {
	record := MutationRecord{
		Type:        string(Rephrase),
		Description: "Rephrased for clarity",
		Generation:  3,
		Timestamp:   time.Now(),
	}

	if record.Type != "rephrase" {
		t.Error("MutationRecord.Type mismatch")
	}
}

func TestMutationOperators(t *testing.T) {
	operators := []MutationOperator{
		Rephrase,
		Expand,
		Prune,
		Reorder,
		Enhance,
		Simplify,
	}

	for _, op := range operators {
		if string(op) == "" {
			t.Errorf("MutationOperator %v has empty string value", op)
		}
	}
}
