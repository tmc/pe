package metaprompt

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"sort"
	"time"

	"github.com/tmc/pe/internal/llm"
)

// EvolutionaryOptimizer implements evolutionary prompt optimization
// based on genetic algorithms and population-based improvement
type EvolutionaryOptimizer struct {
	provider       llm.Provider
	populationSize int
	mutationRate   float64
	crossoverRate  float64
	elitismRate    float64
	objectives     []string
}

// EvolutionConfig holds configuration for evolutionary optimization
type EvolutionConfig struct {
	PopulationSize  int     `json:"population_size"`
	Generations     int     `json:"generations"`
	MutationRate    float64 `json:"mutation_rate"`
	CrossoverRate   float64 `json:"crossover_rate"`
	ElitismRate     float64 `json:"elitism_rate"`
	Objectives      []string `json:"objectives"`
	FitnessMetrics  []string `json:"fitness_metrics"`
}

// Individual represents a prompt candidate in the population
type Individual struct {
	Prompt      string             `json:"prompt"`
	Fitness     float64            `json:"fitness"`
	Objectives  map[string]float64 `json:"objectives"`
	Generation  int                `json:"generation"`
	Genealogy   []string           `json:"genealogy"`
	Mutations   []MutationRecord   `json:"mutations"`
}

// MutationRecord tracks the history of mutations applied to an individual
type MutationRecord struct {
	Type        string    `json:"type"`
	Description string    `json:"description"`
	Generation  int       `json:"generation"`
	Timestamp   time.Time `json:"timestamp"`
}

// Population represents a collection of prompt candidates
type Population struct {
	Individuals []Individual `json:"individuals"`
	Generation  int          `json:"generation"`
	BestFitness float64      `json:"best_fitness"`
	AvgFitness  float64      `json:"avg_fitness"`
	Diversity   float64      `json:"diversity"`
}

// EvolutionResult contains the results of evolutionary optimization
type EvolutionResult struct {
	BestIndividual Individual   `json:"best_individual"`
	FinalPopulation Population  `json:"final_population"`
	EvolutionHistory []Population `json:"evolution_history"`
	ParetoFrontier  []Individual `json:"pareto_frontier"`
	ConvergenceData []float64    `json:"convergence_data"`
	Config          EvolutionConfig `json:"config"`
}

// MutationOperator defines different types of mutations
type MutationOperator string

const (
	Rephrase  MutationOperator = "rephrase"
	Expand    MutationOperator = "expand"
	Prune     MutationOperator = "prune"
	Reorder   MutationOperator = "reorder"
	Enhance   MutationOperator = "enhance"
	Simplify  MutationOperator = "simplify"
)

// NewEvolutionaryOptimizer creates a new evolutionary optimizer
func NewEvolutionaryOptimizer(provider llm.Provider, config EvolutionConfig) *EvolutionaryOptimizer {
	return &EvolutionaryOptimizer{
		provider:       provider,
		populationSize: config.PopulationSize,
		mutationRate:   config.MutationRate,
		crossoverRate:  config.CrossoverRate,
		elitismRate:    config.ElitismRate,
		objectives:     config.Objectives,
	}
}

// Evolve performs evolutionary optimization on a prompt
func (e *EvolutionaryOptimizer) Evolve(ctx context.Context, basePrompt string, config EvolutionConfig) (*EvolutionResult, error) {
	rand.Seed(time.Now().UnixNano())
	
	result := &EvolutionResult{
		Config:           config,
		EvolutionHistory: make([]Population, 0, config.Generations),
		ConvergenceData:  make([]float64, 0, config.Generations),
	}

	// Initialize population
	population, err := e.initializePopulation(ctx, basePrompt)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize population: %w", err)
	}

	// Evolution loop
	for generation := 0; generation < config.Generations; generation++ {
		// Evaluate fitness
		if err := e.evaluatePopulation(ctx, &population); err != nil {
			return nil, fmt.Errorf("failed to evaluate population at generation %d: %w", generation, err)
		}

		// Update population statistics
		e.updatePopulationStats(&population)
		
		// Record generation
		result.EvolutionHistory = append(result.EvolutionHistory, population)
		result.ConvergenceData = append(result.ConvergenceData, population.BestFitness)

		// Check for convergence
		if e.hasConverged(result.ConvergenceData, generation) {
			fmt.Printf("Converged at generation %d\n", generation)
			break
		}

		// Create next generation
		nextPopulation, err := e.createNextGeneration(ctx, population)
		if err != nil {
			return nil, fmt.Errorf("failed to create next generation: %w", err)
		}

		population = nextPopulation
		population.Generation = generation + 1
	}

	// Final evaluation
	if err := e.evaluatePopulation(ctx, &population); err != nil {
		return nil, fmt.Errorf("failed to evaluate final population: %w", err)
	}

	// Find best individual
	sort.Slice(population.Individuals, func(i, j int) bool {
		return population.Individuals[i].Fitness > population.Individuals[j].Fitness
	})

	result.BestIndividual = population.Individuals[0]
	result.FinalPopulation = population
	result.ParetoFrontier = e.extractParetoFrontier(population.Individuals)

	return result, nil
}

func (e *EvolutionaryOptimizer) initializePopulation(ctx context.Context, basePrompt string) (Population, error) {
	individuals := make([]Individual, 0, e.populationSize)
	
	// Add base prompt as first individual
	individuals = append(individuals, Individual{
		Prompt:     basePrompt,
		Generation: 0,
		Genealogy:  []string{"base"},
		Mutations:  []MutationRecord{},
		Objectives: make(map[string]float64),
	})

	// Generate diverse variants
	for i := 1; i < e.populationSize; i++ {
		variant, err := e.generateVariant(ctx, basePrompt, i)
		if err != nil {
			continue // Skip failed variants
		}
		
		individuals = append(individuals, Individual{
			Prompt:     variant,
			Generation: 0,
			Genealogy:  []string{fmt.Sprintf("variant_%d", i)},
			Mutations:  []MutationRecord{},
			Objectives: make(map[string]float64),
		})
	}

	return Population{
		Individuals: individuals,
		Generation:  0,
	}, nil
}

func (e *EvolutionaryOptimizer) generateVariant(ctx context.Context, basePrompt string, variantIndex int) (string, error) {
	variantPrompt := fmt.Sprintf(`
Generate a variant of this prompt that maintains the same core objective but uses different phrasing, structure, or approach:

Original: %s

Create variant %d that:
1. Preserves the main intent and functionality
2. Uses different wording or structure
3. Potentially improves clarity or effectiveness
4. Maintains similar length and complexity

Respond with only the variant prompt:`, basePrompt, variantIndex)

	response, err := e.provider.Generate(ctx, variantPrompt, llm.GenerateOptions{})
	if err != nil {
		return "", err
	}

	return response.Text, nil
}

func (e *EvolutionaryOptimizer) evaluatePopulation(ctx context.Context, population *Population) error {
	for i := range population.Individuals {
		fitness, objectives, err := e.evaluateFitness(ctx, population.Individuals[i].Prompt)
		if err != nil {
			population.Individuals[i].Fitness = 0.0
			continue
		}
		
		population.Individuals[i].Fitness = fitness
		population.Individuals[i].Objectives = objectives
	}
	return nil
}

func (e *EvolutionaryOptimizer) evaluateFitness(ctx context.Context, prompt string) (float64, map[string]float64, error) {
	objectives := make(map[string]float64)
	
	// Evaluate each objective
	totalFitness := 0.0
	for _, objective := range e.objectives {
		score, err := e.evaluateObjective(ctx, prompt, objective)
		if err != nil {
			score = 0.5 // Default neutral score
		}
		objectives[objective] = score
		totalFitness += score
	}
	
	// Average fitness across objectives
	if len(e.objectives) > 0 {
		totalFitness /= float64(len(e.objectives))
	}
	
	// Add diversity bonus
	diversityBonus := e.calculateDiversityBonus(prompt)
	totalFitness += diversityBonus * 0.1
	
	return totalFitness, objectives, nil
}

func (e *EvolutionaryOptimizer) evaluateObjective(ctx context.Context, prompt string, objective string) (float64, error) {
	evaluationPrompt := fmt.Sprintf(`
Evaluate this prompt based on the objective: %s

Prompt to evaluate: %s

Rate the prompt on a scale of 0.0 to 1.0 based on how well it achieves the objective.
Consider factors like:
- Clarity and specificity
- Completeness for the objective
- Potential effectiveness
- Proper structure and format

Respond with only a number between 0.0 and 1.0:`, objective, prompt)

	response, err := e.provider.Generate(ctx, evaluationPrompt, llm.GenerateOptions{})
	if err != nil {
		return 0.5, err
	}

	// Parse score from response
	var score float64
	if _, err := fmt.Sscanf(response.Text, "%f", &score); err != nil {
		// Fallback: estimate based on prompt quality heuristics
		return e.heuristicEvaluation(prompt), nil
	}

	return math.Max(0.0, math.Min(1.0, score)), nil
}

func (e *EvolutionaryOptimizer) heuristicEvaluation(prompt string) float64 {
	// Simple heuristic based on prompt characteristics
	score := 0.5
	
	// Length heuristic (moderate length is better)
	length := len(prompt)
	if length > 50 && length < 500 {
		score += 0.1
	}
	
	// Complexity heuristic (some structure is good)
	if contains(prompt, []string{":", ".", "?", "!"}) {
		score += 0.1
	}
	
	// Specificity heuristic (specific instructions are better)
	if contains(prompt, []string{"specific", "detailed", "explain", "analyze"}) {
		score += 0.1
	}
	
	return math.Max(0.0, math.Min(1.0, score))
}

func contains(text string, keywords []string) bool {
	for _, keyword := range keywords {
		if fmt.Sprintf("%s", text) != text {
			continue
		}
		// Simple contains check
		if len(text) > len(keyword) {
			return true
		}
	}
	return false
}

func (e *EvolutionaryOptimizer) updatePopulationStats(population *Population) {
	if len(population.Individuals) == 0 {
		return
	}

	// Find best fitness
	bestFitness := population.Individuals[0].Fitness
	totalFitness := 0.0
	
	for _, individual := range population.Individuals {
		if individual.Fitness > bestFitness {
			bestFitness = individual.Fitness
		}
		totalFitness += individual.Fitness
	}
	
	population.BestFitness = bestFitness
	population.AvgFitness = totalFitness / float64(len(population.Individuals))
	population.Diversity = e.calculatePopulationDiversity(population.Individuals)
}

func (e *EvolutionaryOptimizer) calculatePopulationDiversity(individuals []Individual) float64 {
	if len(individuals) < 2 {
		return 0.0
	}
	
	totalDistance := 0.0
	comparisons := 0
	
	for i := 0; i < len(individuals); i++ {
		for j := i + 1; j < len(individuals); j++ {
			distance := e.calculatePromptDistance(individuals[i].Prompt, individuals[j].Prompt)
			totalDistance += distance
			comparisons++
		}
	}
	
	if comparisons == 0 {
		return 0.0
	}
	
	return totalDistance / float64(comparisons)
}

func (e *EvolutionaryOptimizer) calculatePromptDistance(prompt1, prompt2 string) float64 {
	// Simple Levenshtein-inspired distance
	len1, len2 := len(prompt1), len(prompt2)
	maxLen := math.Max(float64(len1), float64(len2))
	
	if maxLen == 0 {
		return 0.0
	}
	
	// Simple character difference ratio
	minLen := math.Min(float64(len1), float64(len2))
	commonChars := 0
	
	for i := 0; i < int(minLen); i++ {
		if prompt1[i] == prompt2[i] {
			commonChars++
		}
	}
	
	similarity := float64(commonChars) / maxLen
	return 1.0 - similarity
}

func (e *EvolutionaryOptimizer) calculateDiversityBonus(prompt string) float64 {
	// Bonus for unique characteristics
	bonus := 0.0
	
	// Length diversity
	if len(prompt) > 100 && len(prompt) < 400 {
		bonus += 0.1
	}
	
	return bonus
}

func (e *EvolutionaryOptimizer) hasConverged(convergenceData []float64, generation int) bool {
	// Check for convergence (no improvement in last N generations)
	if generation < 10 {
		return false
	}
	
	recentGenerations := 5
	if generation < recentGenerations {
		return false
	}
	
	// Check if fitness has plateaued
	recent := convergenceData[len(convergenceData)-recentGenerations:]
	maxRecent := recent[0]
	minRecent := recent[0]
	
	for _, fitness := range recent {
		if fitness > maxRecent {
			maxRecent = fitness
		}
		if fitness < minRecent {
			minRecent = fitness
		}
	}
	
	// Converged if improvement is less than 1%
	improvementThreshold := 0.01
	return (maxRecent - minRecent) < improvementThreshold
}

func (e *EvolutionaryOptimizer) createNextGeneration(ctx context.Context, currentPop Population) (Population, error) {
	nextGen := Population{
		Individuals: make([]Individual, 0, e.populationSize),
		Generation:  currentPop.Generation + 1,
	}

	// Sort by fitness
	sorted := make([]Individual, len(currentPop.Individuals))
	copy(sorted, currentPop.Individuals)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Fitness > sorted[j].Fitness
	})

	// Elitism: keep top performers
	eliteCount := int(float64(e.populationSize) * e.elitismRate)
	for i := 0; i < eliteCount && i < len(sorted); i++ {
		elite := sorted[i]
		elite.Generation = nextGen.Generation
		nextGen.Individuals = append(nextGen.Individuals, elite)
	}

	// Generate offspring through crossover and mutation
	for len(nextGen.Individuals) < e.populationSize {
		// Selection
		parent1 := e.tournamentSelection(sorted)
		parent2 := e.tournamentSelection(sorted)

		// Crossover
		var offspring Individual
		if rand.Float64() < e.crossoverRate {
			offspring, _ = e.crossover(ctx, parent1, parent2)
		} else {
			offspring = parent1
		}

		// Mutation
		if rand.Float64() < e.mutationRate {
			mutated, _ := e.mutate(ctx, offspring)
			offspring = mutated
		}

		offspring.Generation = nextGen.Generation
		nextGen.Individuals = append(nextGen.Individuals, offspring)
	}

	return nextGen, nil
}

func (e *EvolutionaryOptimizer) tournamentSelection(individuals []Individual) Individual {
	tournamentSize := 3
	best := individuals[rand.Intn(len(individuals))]
	
	for i := 1; i < tournamentSize; i++ {
		candidate := individuals[rand.Intn(len(individuals))]
		if candidate.Fitness > best.Fitness {
			best = candidate
		}
	}
	
	return best
}

func (e *EvolutionaryOptimizer) crossover(ctx context.Context, parent1, parent2 Individual) (Individual, error) {
	crossoverPrompt := fmt.Sprintf(`
Create a new prompt by combining the best elements from these two parent prompts:

Parent 1: %s

Parent 2: %s

Create a hybrid that:
1. Combines the strongest aspects of both parents
2. Maintains coherence and clarity
3. Potentially improves upon both parents
4. Preserves the core functionality

Respond with only the new hybrid prompt:`, parent1.Prompt, parent2.Prompt)

	response, err := e.provider.Generate(ctx, crossoverPrompt, llm.GenerateOptions{})
	if err != nil {
		return parent1, err // Fallback to parent1
	}

	offspring := Individual{
		Prompt:    response.Text,
		Genealogy: []string{parent1.Genealogy[0], parent2.Genealogy[0]},
		Mutations: []MutationRecord{},
		Objectives: make(map[string]float64),
	}

	return offspring, nil
}

func (e *EvolutionaryOptimizer) mutate(ctx context.Context, individual Individual) (Individual, error) {
	// Choose random mutation operator
	operators := []MutationOperator{Rephrase, Expand, Prune, Enhance, Simplify}
	operator := operators[rand.Intn(len(operators))]

	mutationPrompt := e.createMutationPrompt(individual.Prompt, operator)
	
	response, err := e.provider.Generate(ctx, mutationPrompt, llm.GenerateOptions{})
	if err != nil {
		return individual, err // No mutation if failed
	}

	mutated := individual
	mutated.Prompt = response.Text
	mutated.Mutations = append(mutated.Mutations, MutationRecord{
		Type:        string(operator),
		Description: fmt.Sprintf("Applied %s mutation", operator),
		Generation:  individual.Generation,
		Timestamp:   time.Now(),
	})

	return mutated, nil
}

func (e *EvolutionaryOptimizer) createMutationPrompt(prompt string, operator MutationOperator) string {
	switch operator {
	case Rephrase:
		return fmt.Sprintf("Rephrase this prompt while maintaining its meaning and effectiveness:\n\n%s\n\nRespond with only the rephrased prompt:", prompt)
	case Expand:
		return fmt.Sprintf("Expand this prompt with additional details or clarifications that could improve its effectiveness:\n\n%s\n\nRespond with only the expanded prompt:", prompt)
	case Prune:
		return fmt.Sprintf("Make this prompt more concise while preserving its essential meaning:\n\n%s\n\nRespond with only the pruned prompt:", prompt)
	case Enhance:
		return fmt.Sprintf("Enhance this prompt to be more specific, clear, or effective:\n\n%s\n\nRespond with only the enhanced prompt:", prompt)
	case Simplify:
		return fmt.Sprintf("Simplify this prompt to be more straightforward and easier to understand:\n\n%s\n\nRespond with only the simplified prompt:", prompt)
	default:
		return fmt.Sprintf("Improve this prompt:\n\n%s\n\nRespond with only the improved prompt:", prompt)
	}
}

func (e *EvolutionaryOptimizer) extractParetoFrontier(individuals []Individual) []Individual {
	var frontier []Individual
	
	for i, candidate := range individuals {
		dominated := false
		
		for j, other := range individuals {
			if i != j && e.dominatesMultiObjective(other, candidate) {
				dominated = true
				break
			}
		}
		
		if !dominated {
			frontier = append(frontier, candidate)
		}
	}
	
	return frontier
}

func (e *EvolutionaryOptimizer) dominatesMultiObjective(a, b Individual) bool {
	betterInAny := false
	
	for objective := range a.Objectives {
		aValue := a.Objectives[objective]
		bValue := b.Objectives[objective]
		
		if aValue > bValue {
			betterInAny = true
		} else if aValue < bValue {
			return false
		}
	}
	
	return betterInAny
}