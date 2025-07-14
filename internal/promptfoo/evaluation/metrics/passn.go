package metrics

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/tmc/pe/internal/llm"
)

// PassNStore handles storage and retrieval of pass@n evaluation data
type PassNStore struct {
	mu       sync.RWMutex
	dataDir  string
	metadata map[string]*PassNMetadata
}

// PassNMetadata stores metadata about pass@n evaluations
type PassNMetadata struct {
	PromptID     string                 `json:"prompt_id"`
	PromptHash   string                 `json:"prompt_hash"`
	Prompt       string                 `json:"prompt"`
	ProviderID   string                 `json:"provider_id"`
	Task         string                 `json:"task,omitempty"`
	Language     string                 `json:"language,omitempty"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
	Evaluations  []PassNEvaluation      `json:"evaluations"`
	Tags         []string               `json:"tags,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	BestPassRate float64                `json:"best_pass_rate"`
	AvgPassRate  float64                `json:"avg_pass_rate"`
	TotalSamples int                    `json:"total_samples"`
}

// PassNEvaluation represents a single pass@n evaluation session
type PassNEvaluation struct {
	ID            string                 `json:"id"`
	Timestamp     time.Time              `json:"timestamp"`
	N             int                    `json:"n"`
	NumSamples    int                    `json:"num_samples"`
	NumPassed     int                    `json:"num_passed"`
	PassRate      float64                `json:"pass_rate"`
	Samples       []PassNSample          `json:"samples"`
	TestCases     []PassNTestCase        `json:"test_cases,omitempty"`
	Configuration map[string]interface{} `json:"configuration,omitempty"`
	Duration      time.Duration          `json:"duration"`
}

// PassNSample represents a single generated sample
type PassNSample struct {
	Index       int                    `json:"index"`
	Content     string                 `json:"content"`
	Passed      bool                   `json:"passed"`
	TestResults []PassNTestResult      `json:"test_results,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Error       string                 `json:"error,omitempty"`
}

// PassNTestCase represents a test case for evaluation
type PassNTestCase struct {
	ID          string  `json:"id"`
	Input       string  `json:"input"`
	Expected    string  `json:"expected"`
	Description string  `json:"description,omitempty"`
	Type        string  `json:"type,omitempty"` // exact, contains, regex, semantic
	Weight      float64 `json:"weight,omitempty"`
}

// PassNTestResult represents the result of a single test case
type PassNTestResult struct {
	TestCaseID string  `json:"test_case_id"`
	Passed     bool    `json:"passed"`
	Actual     string  `json:"actual,omitempty"`
	Score      float64 `json:"score,omitempty"`
	Error      string  `json:"error,omitempty"`
}

// NewPassNStore creates a new pass@n storage manager
func NewPassNStore(dataDir string) (*PassNStore, error) {
	// Ensure data directory exists
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	store := &PassNStore{
		dataDir:  dataDir,
		metadata: make(map[string]*PassNMetadata),
	}

	// Load existing metadata
	if err := store.loadMetadata(); err != nil {
		return nil, fmt.Errorf("failed to load metadata: %w", err)
	}

	return store, nil
}

// StoreEvaluation stores a new pass@n evaluation
func (s *PassNStore) StoreEvaluation(promptID string, prompt string, providerID string, evaluation PassNEvaluation) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Get or create metadata
	metadata, exists := s.metadata[promptID]
	if !exists {
		metadata = &PassNMetadata{
			PromptID:    promptID,
			PromptHash:  generateHash(prompt),
			Prompt:      prompt,
			ProviderID:  providerID,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			Evaluations: []PassNEvaluation{},
		}
		s.metadata[promptID] = metadata
	}

	// Add evaluation
	metadata.Evaluations = append(metadata.Evaluations, evaluation)
	metadata.UpdatedAt = time.Now()

	// Update statistics
	s.updateStatistics(metadata)

	// Save to disk
	return s.saveMetadata(promptID, metadata)
}

// GetEvaluations retrieves all evaluations for a prompt
func (s *PassNStore) GetEvaluations(promptID string) (*PassNMetadata, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	metadata, exists := s.metadata[promptID]
	if !exists {
		return nil, fmt.Errorf("no evaluations found for prompt ID: %s", promptID)
	}

	return metadata, nil
}

// GetBestEvaluation returns the evaluation with the highest pass rate
func (s *PassNStore) GetBestEvaluation(promptID string) (*PassNEvaluation, error) {
	metadata, err := s.GetEvaluations(promptID)
	if err != nil {
		return nil, err
	}

	if len(metadata.Evaluations) == 0 {
		return nil, fmt.Errorf("no evaluations found")
	}

	best := &metadata.Evaluations[0]
	for i := range metadata.Evaluations {
		if metadata.Evaluations[i].PassRate > best.PassRate {
			best = &metadata.Evaluations[i]
		}
	}

	return best, nil
}

// SearchByTags finds prompts with specific tags
func (s *PassNStore) SearchByTags(tags []string) ([]*PassNMetadata, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var results []*PassNMetadata

	for _, metadata := range s.metadata {
		if containsAllTags(metadata.Tags, tags) {
			results = append(results, metadata)
		}
	}

	return results, nil
}

// GeneratePassNReport generates a comprehensive report for a prompt
func (s *PassNStore) GeneratePassNReport(promptID string) (*PassNReport, error) {
	metadata, err := s.GetEvaluations(promptID)
	if err != nil {
		return nil, err
	}

	report := &PassNReport{
		PromptID:     metadata.PromptID,
		Prompt:       metadata.Prompt,
		ProviderID:   metadata.ProviderID,
		TotalEvals:   len(metadata.Evaluations),
		BestPassRate: metadata.BestPassRate,
		AvgPassRate:  metadata.AvgPassRate,
		TotalSamples: metadata.TotalSamples,
		CreatedAt:    metadata.CreatedAt,
		UpdatedAt:    metadata.UpdatedAt,
	}

	// Calculate pass rate distribution
	distribution := make(map[int][]float64)
	for _, eval := range metadata.Evaluations {
		distribution[eval.N] = append(distribution[eval.N], eval.PassRate)
	}
	report.PassRateByN = distribution

	// Find trends
	if len(metadata.Evaluations) > 1 {
		report.Trend = calculateTrend(metadata.Evaluations)
	}

	return report, nil
}

// PassNReport represents a comprehensive report for pass@n evaluations
type PassNReport struct {
	PromptID     string            `json:"prompt_id"`
	Prompt       string            `json:"prompt"`
	ProviderID   string            `json:"provider_id"`
	TotalEvals   int               `json:"total_evaluations"`
	BestPassRate float64           `json:"best_pass_rate"`
	AvgPassRate  float64           `json:"avg_pass_rate"`
	TotalSamples int               `json:"total_samples"`
	PassRateByN  map[int][]float64 `json:"pass_rate_by_n"`
	Trend        string            `json:"trend,omitempty"`
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
}

// PassNGenerator generates samples for pass@n evaluation
type PassNGenerator struct {
	provider llm.Provider
	config   PassNConfig
}

// PassNConfig configures pass@n sample generation
type PassNConfig struct {
	Temperature      float64                `json:"temperature"`
	TopP             float64                `json:"top_p"`
	MaxTokens        int                    `json:"max_tokens"`
	StopSequences    []string               `json:"stop_sequences,omitempty"`
	PresencePenalty  float64                `json:"presence_penalty,omitempty"`
	FrequencyPenalty float64                `json:"frequency_penalty,omitempty"`
	SamplingStrategy string                 `json:"sampling_strategy,omitempty"` // random, diverse, adaptive
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}

// NewPassNGenerator creates a new pass@n sample generator
func NewPassNGenerator(provider llm.Provider, config PassNConfig) *PassNGenerator {
	// Set defaults
	if config.Temperature == 0 {
		config.Temperature = 0.8 // Higher temperature for diversity
	}
	if config.TopP == 0 {
		config.TopP = 0.95
	}
	if config.SamplingStrategy == "" {
		config.SamplingStrategy = "random"
	}

	return &PassNGenerator{
		provider: provider,
		config:   config,
	}
}

// GenerateSamples generates n samples for a given prompt
func (g *PassNGenerator) GenerateSamples(ctx context.Context, prompt string, n int) ([]string, error) {
	samples := make([]string, 0, n)
	errors := make([]error, 0)

	// Use different sampling strategies
	switch g.config.SamplingStrategy {
	case "diverse":
		samples, errors = g.generateDiverseSamples(ctx, prompt, n)
	case "adaptive":
		samples, errors = g.generateAdaptiveSamples(ctx, prompt, n)
	default:
		samples, errors = g.generateRandomSamples(ctx, prompt, n)
	}

	// Check if we got enough samples
	if len(samples) < n/2 && len(errors) > 0 {
		return samples, fmt.Errorf("failed to generate sufficient samples: %v", errors[0])
	}

	return samples, nil
}

// generateRandomSamples uses random sampling with the configured temperature
func (g *PassNGenerator) generateRandomSamples(ctx context.Context, prompt string, n int) ([]string, []error) {
	var wg sync.WaitGroup
	samplesChan := make(chan string, n)
	errorsChan := make(chan error, n)

	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			opts := llm.GenerateOptions{
				Temperature: &g.config.Temperature,
				TopP:        &g.config.TopP,
				MaxTokens:   &g.config.MaxTokens,
				Stop:        g.config.StopSequences,
			}

			response, err := g.provider.Generate(ctx, prompt, opts)
			if err != nil {
				errorsChan <- err
				return
			}

			samplesChan <- response.Text
		}(i)
	}

	wg.Wait()
	close(samplesChan)
	close(errorsChan)

	// Collect results
	samples := make([]string, 0, n)
	errors := make([]error, 0)

	for sample := range samplesChan {
		samples = append(samples, sample)
	}

	for err := range errorsChan {
		errors = append(errors, err)
	}

	return samples, errors
}

// generateDiverseSamples uses temperature scheduling for diversity
func (g *PassNGenerator) generateDiverseSamples(ctx context.Context, prompt string, n int) ([]string, []error) {
	samples := make([]string, 0, n)
	errors := make([]error, 0)

	// Use different temperatures for diversity
	temperatures := []float64{0.5, 0.7, 0.9, 1.0, 1.2}

	for i := 0; i < n; i++ {
		temp := temperatures[i%len(temperatures)]

		opts := llm.GenerateOptions{
			Temperature: &temp,
			TopP:        &g.config.TopP,
			MaxTokens:   &g.config.MaxTokens,
			Stop:        g.config.StopSequences,
		}

		response, err := g.provider.Generate(ctx, prompt, opts)
		if err != nil {
			errors = append(errors, err)
			continue
		}

		samples = append(samples, response.Text)
	}

	return samples, errors
}

// generateAdaptiveSamples adjusts parameters based on initial results
func (g *PassNGenerator) generateAdaptiveSamples(ctx context.Context, prompt string, n int) ([]string, []error) {
	samples := make([]string, 0, n)
	errors := make([]error, 0)

	// Start with base temperature
	currentTemp := g.config.Temperature

	for i := 0; i < n; i++ {
		opts := llm.GenerateOptions{
			Temperature: &currentTemp,
			TopP:        &g.config.TopP,
			MaxTokens:   &g.config.MaxTokens,
			Stop:        g.config.StopSequences,
		}

		response, err := g.provider.Generate(ctx, prompt, opts)
		if err != nil {
			errors = append(errors, err)
			continue
		}

		samples = append(samples, response.Text)

		// Adapt temperature based on diversity
		if i > 0 && i%5 == 0 {
			diversity := calculateDiversity(samples)
			if diversity < 0.5 {
				newTemp := currentTemp * 1.1
				if newTemp > 1.5 {
					newTemp = 1.5
				}
				currentTemp = newTemp
			} else if diversity > 0.8 {
				newTemp := currentTemp * 0.9
				if newTemp < 0.3 {
					newTemp = 0.3
				}
				currentTemp = newTemp
			}
		}
	}

	return samples, errors
}

// Helper functions

func (s *PassNStore) loadMetadata() error {
	metadataFile := filepath.Join(s.dataDir, "metadata.json")

	data, err := os.ReadFile(metadataFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No metadata yet
		}
		return err
	}

	return json.Unmarshal(data, &s.metadata)
}

func (s *PassNStore) saveMetadata(promptID string, metadata *PassNMetadata) error {
	// Save individual metadata file
	promptFile := filepath.Join(s.dataDir, fmt.Sprintf("%s.json", promptID))
	data, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(promptFile, data, 0644); err != nil {
		return err
	}

	// Update global metadata index
	metadataFile := filepath.Join(s.dataDir, "metadata.json")
	allData, err := json.MarshalIndent(s.metadata, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(metadataFile, allData, 0644)
}

func (s *PassNStore) updateStatistics(metadata *PassNMetadata) {
	totalPassRate := 0.0
	totalSamples := 0
	bestPassRate := 0.0

	for _, eval := range metadata.Evaluations {
		totalPassRate += eval.PassRate
		totalSamples += eval.NumSamples
		if eval.PassRate > bestPassRate {
			bestPassRate = eval.PassRate
		}
	}

	if len(metadata.Evaluations) > 0 {
		metadata.AvgPassRate = totalPassRate / float64(len(metadata.Evaluations))
	}
	metadata.BestPassRate = bestPassRate
	metadata.TotalSamples = totalSamples
}

func generateHash(content string) string {
	// Simple hash generation (implement proper hashing)
	return fmt.Sprintf("%x", content[:min(16, len(content))])
}

func containsAllTags(metadataTags, searchTags []string) bool {
	tagSet := make(map[string]bool)
	for _, tag := range metadataTags {
		tagSet[tag] = true
	}

	for _, tag := range searchTags {
		if !tagSet[tag] {
			return false
		}
	}

	return true
}

func calculateTrend(evaluations []PassNEvaluation) string {
	if len(evaluations) < 2 {
		return ""
	}

	// Simple trend calculation based on recent evaluations
	recent := evaluations[len(evaluations)-3:]
	if len(recent) < 2 {
		recent = evaluations[len(evaluations)-2:]
	}

	avgRecent := 0.0
	for _, eval := range recent {
		avgRecent += eval.PassRate
	}
	avgRecent /= float64(len(recent))

	older := evaluations[:len(evaluations)-len(recent)]
	avgOlder := 0.0
	for _, eval := range older {
		avgOlder += eval.PassRate
	}
	avgOlder /= float64(len(older))

	if avgRecent > avgOlder*1.1 {
		return "improving"
	} else if avgRecent < avgOlder*0.9 {
		return "declining"
	}
	return "stable"
}

func calculateDiversity(samples []string) float64 {
	if len(samples) < 2 {
		return 1.0
	}

	// Simple diversity metric based on unique tokens
	allTokens := make(map[string]int)
	for _, sample := range samples {
		tokens := tokenize(sample)
		for _, token := range tokens {
			allTokens[token]++
		}
	}

	// Calculate diversity as ratio of unique tokens to total tokens
	uniqueTokens := len(allTokens)
	totalTokens := 0
	for _, count := range allTokens {
		totalTokens += count
	}

	if totalTokens == 0 {
		return 0.0
	}

	return float64(uniqueTokens) / float64(totalTokens)
}
