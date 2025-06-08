package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/tmc/pe/internal/consensus"
)

// DemoConfig holds configuration for the consensus demo
type DemoConfig struct {
	Prompts             []string `json:"prompts"`
	IncludeMalicious    bool     `json:"include_malicious"`
	EmbeddingDimensions int      `json:"embedding_dimensions"`
	Verbose             bool     `json:"verbose"`
}

// ConsensusDemo demonstrates the semantic consensus system
func runConsensusDemo() error {
	fmt.Println("🔍 PE Semantic Consensus Demo")
	fmt.Println("============================")

	// Create demo configuration
	config := &DemoConfig{
		Prompts: []string{
			"Analyze the sentiment of this customer review: 'I absolutely love this product! The quality is excellent and shipping was fast.'",
			"What is the current stock price of Apple Inc.?",
			"Write a short poem about mountains and nature.",
			"Summarize the benefits of renewable energy for the environment.",
			"Write a Python function to calculate the fibonacci sequence.",
		},
		IncludeMalicious:    true,
		EmbeddingDimensions: 384,
		Verbose:             true,
	}

	// Create embedding service
	embeddingService := consensus.NewMockEmbeddingService(config.EmbeddingDimensions)

	// Create providers (including malicious one for demo)
	var providers []consensus.ModelProvider
	if config.IncludeMalicious {
		providers = consensus.CreateTestProvidersWithMalicious()
		fmt.Printf("📡 Created %d providers (including 1 malicious)\n", len(providers))
	} else {
		providers = consensus.CreateTestProviders()
		fmt.Printf("📡 Created %d honest providers\n", len(providers))
	}

	// Create semantic consensus engine
	consensusEngine := consensus.NewSemanticConsensus(embeddingService, providers)

	fmt.Printf("🧠 Embedding model: %s (%d dimensions)\n", embeddingService.ModelName(), embeddingService.Dimensions())
	fmt.Println()

	// Model parameters for all tests
	params := &consensus.ModelParameters{
		Model:       "gpt-4-turbo",
		Temperature: 0.7,
		MaxTokens:   500,
	}

	// Run consensus tests for each prompt
	for i, prompt := range config.Prompts {
		fmt.Printf("🎯 Test %d: Running semantic consensus analysis\n", i+1)
		fmt.Printf("📝 Prompt: %s\n", truncateString(prompt, 80))
		fmt.Println("---")

		startTime := time.Now()

		// Execute with consensus
		result, err := consensusEngine.ExecuteWithConsensus(context.Background(), prompt, params)
		if err != nil {
			fmt.Printf("❌ Error: %v\n\n", err)
			continue
		}

		executionTime := time.Since(startTime)

		// Display results
		displayConsensusResults(result, executionTime, config.Verbose)
		fmt.Println()
	}

	return nil
}

// displayConsensusResults shows the consensus analysis results
func displayConsensusResults(result *consensus.ConsensusResult, executionTime time.Duration, verbose bool) {
	fmt.Printf("⏱️  Execution time: %v\n", executionTime)
	fmt.Printf("🤝 Providers responded: %d\n", len(result.ProviderResponses))

	// Display consensus metrics
	metrics := result.ConsensusMetrics
	fmt.Printf("📊 Consensus metrics:\n")
	fmt.Printf("   • Agreement ratio: %.1f%%\n", metrics.AgreementRatio*100)
	fmt.Printf("   • Average similarity: %.3f\n", metrics.AverageSimilarity)
	fmt.Printf("   • Consensus confidence: %.3f\n", metrics.ConsensusConfidence)
	fmt.Printf("   • Semantic coherence: %.3f\n", metrics.SemanticCoherence)
	fmt.Printf("   • Outlier count: %d\n", metrics.OutlierCount)

	// Display manipulation analysis
	manipulation := result.ManipulationAnalysis
	if manipulation.ManipulationLikely {
		fmt.Printf("🚨 Manipulation detected!\n")
		fmt.Printf("   • Anomaly score: %.3f\n", manipulation.AnomalyScore)
		fmt.Printf("   • Confidence: %.3f\n", manipulation.ConfidenceLevel)
		fmt.Printf("   • Suspicious providers: %v\n", manipulation.SuspiciousProviders)
		if len(manipulation.DetectionReasons) > 0 {
			fmt.Printf("   • Reasons:\n")
			for _, reason := range manipulation.DetectionReasons {
				fmt.Printf("     - %s\n", reason)
			}
		}
	} else {
		fmt.Printf("✅ No manipulation detected\n")
		fmt.Printf("   • Anomaly score: %.3f\n", manipulation.AnomalyScore)
		fmt.Printf("   • Confidence: %.3f\n", manipulation.ConfidenceLevel)
	}

	// Display cluster information
	fmt.Printf("🔍 Clusters found: %d\n", len(result.Clusters))
	for i, cluster := range result.Clusters {
		fmt.Printf("   Cluster %d: %d members, cohesion %.3f, confidence %.3f\n",
			i+1, cluster.Size, cluster.Cohesion, cluster.Confidence)
	}

	// Display consensus response
	if result.ConsensusResponse != "" {
		fmt.Printf("💬 Consensus response:\n")
		fmt.Printf("   %s\n", truncateString(result.ConsensusResponse, 120))
	}

	// Verbose output
	if verbose {
		fmt.Printf("\n📋 Detailed provider responses:\n")
		for i, providerResp := range result.ProviderResponses {
			fmt.Printf("   %d. %s (%v):\n", i+1, providerResp.ProviderID, providerResp.ExecutionTime)
			fmt.Printf("      %s\n", truncateString(providerResp.Response, 100))
		}

		fmt.Printf("\n🔢 Similarity matrix:\n")
		displaySimilarityMatrix(result.SimilarityMatrix, result.ProviderResponses)
	}
}

// displaySimilarityMatrix shows the similarity matrix in a readable format
func displaySimilarityMatrix(matrix [][]float64, responses []consensus.ProviderResponse) {
	if len(matrix) == 0 {
		return
	}

	// Header
	fmt.Printf("      ")
	for i := range responses {
		fmt.Printf("%8s", truncateString(responses[i].ProviderID, 8))
	}
	fmt.Printf("\n")

	// Matrix rows
	for i, row := range matrix {
		fmt.Printf("%8s", truncateString(responses[i].ProviderID, 8))
		for _, val := range row {
			fmt.Printf("%8.3f", val)
		}
		fmt.Printf("\n")
	}
}

// truncateString truncates a string to a maximum length
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// saveResults saves consensus results to a JSON file
func saveResults(result *consensus.ConsensusResult, filename string) error {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, 0644)
}

// runInteractiveDemo allows users to input custom prompts
func runInteractiveDemo() error {
	fmt.Println("🎮 Interactive Semantic Consensus Demo")
	fmt.Println("=====================================")
	fmt.Println("Enter prompts to test semantic consensus (type 'quit' to exit)")

	// Create services
	embeddingService := consensus.NewMockEmbeddingService(384)
	providers := consensus.CreateTestProvidersWithMalicious() // Include malicious for testing
	consensusEngine := consensus.NewSemanticConsensus(embeddingService, providers)

	params := &consensus.ModelParameters{
		Model:       "gpt-4-turbo",
		Temperature: 0.7,
		MaxTokens:   500,
	}

	for {
		fmt.Print("\n💭 Enter prompt: ")
		var prompt string
		if _, err := fmt.Scanln(&prompt); err != nil {
			continue
		}

		if prompt == "quit" {
			break
		}

		if prompt == "" {
			continue
		}

		fmt.Printf("🔄 Running consensus analysis...\n")

		result, err := consensusEngine.ExecuteWithConsensus(context.Background(), prompt, params)
		if err != nil {
			fmt.Printf("❌ Error: %v\n", err)
			continue
		}

		displayConsensusResults(result, 0, false)
	}

	fmt.Println("👋 Goodbye!")
	return nil
}

// Benchmark function to test performance
func benchmarkConsensus() error {
	fmt.Println("⚡ Semantic Consensus Benchmark")
	fmt.Println("==============================")

	embeddingService := consensus.NewMockEmbeddingService(384)
	providers := consensus.CreateTestProviders() // No malicious for clean benchmark
	consensusEngine := consensus.NewSemanticConsensus(embeddingService, providers)

	params := &consensus.ModelParameters{
		Model:       "gpt-4-turbo",
		Temperature: 0.7,
		MaxTokens:   100,
	}

	testPrompts := []string{
		"What is the capital of France?",
		"Explain machine learning in simple terms.",
		"Write a haiku about technology.",
		"List the benefits of exercise.",
		"Describe the solar system.",
	}

	fmt.Printf("🎯 Running %d test prompts across %d providers\n", len(testPrompts), len(providers))

	totalStart := time.Now()
	var totalConsensusTime time.Duration

	for i, prompt := range testPrompts {
		start := time.Now()
		result, err := consensusEngine.ExecuteWithConsensus(context.Background(), prompt, params)
		elapsed := time.Since(start)
		totalConsensusTime += elapsed

		if err != nil {
			fmt.Printf("❌ Test %d failed: %v\n", i+1, err)
			continue
		}

		fmt.Printf("✅ Test %d: %v (consensus: %.3f, outliers: %d)\n",
			i+1, elapsed, result.ConsensusMetrics.ConsensusConfidence, result.ConsensusMetrics.OutlierCount)
	}

	totalElapsed := time.Since(totalStart)

	fmt.Printf("\n📈 Benchmark Results:\n")
	fmt.Printf("   • Total time: %v\n", totalElapsed)
	fmt.Printf("   • Average per prompt: %v\n", totalConsensusTime/time.Duration(len(testPrompts)))
	fmt.Printf("   • Prompts per second: %.2f\n", float64(len(testPrompts))/totalElapsed.Seconds())

	return nil
}

// Command-line interface for consensus demo
func init() {
	// This would integrate with the main PE CLI
	// For now, it's a standalone demo function
}

// Main demo runner
func DemoSemanticConsensus(mode string) error {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	switch mode {
	case "demo", "":
		return runConsensusDemo()
	case "interactive":
		return runInteractiveDemo()
	case "benchmark":
		return benchmarkConsensus()
	default:
		return fmt.Errorf("unknown demo mode: %s (available: demo, interactive, benchmark)", mode)
	}
}
