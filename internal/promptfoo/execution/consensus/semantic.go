package consensus

import (
	"context"
	"fmt"
	"log"
	"math"
	"sort"
	"sync"
	"time"
)

// SemanticConsensus implements consensus through vector similarity analysis
type SemanticConsensus struct {
	embeddingService EmbeddingService
	providers        []ModelProvider
	thresholds       ConsensusThresholds
	mu               sync.RWMutex
}

// ConsensusThresholds define semantic similarity requirements
type ConsensusThresholds struct {
	MinSimilarity    float64 `json:"min_similarity"`    // 0.85 typical
	ClusterThreshold float64 `json:"cluster_threshold"` // 0.80 typical
	OutlierThreshold float64 `json:"outlier_threshold"` // 0.60 typical
	MinProviders     int     `json:"min_providers"`     // 3 minimum
}

// ProviderResponse represents a single provider's response
type ProviderResponse struct {
	ProviderID    string        `json:"provider_id"`
	Response      string        `json:"response"`
	Embedding     []float64     `json:"embedding"`
	ExecutionTime time.Duration `json:"execution_time"`
	TokenUsage    TokenUsage    `json:"token_usage"`
	Error         error         `json:"error,omitempty"`
}

// SemanticCluster groups similar responses
type SemanticCluster struct {
	Members    []int     `json:"members"`    // Provider indices
	Centroid   []float64 `json:"centroid"`   // Average embedding
	Cohesion   float64   `json:"cohesion"`   // Internal similarity
	Size       int       `json:"size"`       // Number of members
	Confidence float64   `json:"confidence"` // Consensus confidence
}

// ConsensusResult contains the final consensus analysis
type ConsensusResult struct {
	PromptHash           string                `json:"prompt_hash"`
	ConsensusResponse    string                `json:"consensus_response"`
	ConsensusEmbedding   []float64             `json:"consensus_embedding"`
	ProviderResponses    []ProviderResponse    `json:"provider_responses"`
	SimilarityMatrix     [][]float64           `json:"similarity_matrix"`
	Clusters             []SemanticCluster     `json:"clusters"`
	ManipulationAnalysis *ManipulationAnalysis `json:"manipulation_analysis"`
	ConsensusMetrics     *ConsensusMetrics     `json:"consensus_metrics"`
	Timestamp            time.Time             `json:"timestamp"`
}

// ManipulationAnalysis detects potential response manipulation
type ManipulationAnalysis struct {
	SuspiciousProviders []string `json:"suspicious_providers"`
	AnomalyScore        float64  `json:"anomaly_score"`
	ManipulationLikely  bool     `json:"manipulation_likely"`
	ConfidenceLevel     float64  `json:"confidence_level"`
	DetectionReasons    []string `json:"detection_reasons"`
}

// ConsensusMetrics provide statistical analysis
type ConsensusMetrics struct {
	AgreementRatio      float64 `json:"agreement_ratio"`      // % in majority cluster
	AverageSimilarity   float64 `json:"average_similarity"`   // Mean pairwise similarity
	ConsensusConfidence float64 `json:"consensus_confidence"` // Overall confidence
	SemanticCoherence   float64 `json:"semantic_coherence"`   // Cluster cohesion
	OutlierCount        int     `json:"outlier_count"`        // Number of outliers
}

// NewSemanticConsensus creates a new semantic consensus engine
func NewSemanticConsensus(embeddingService EmbeddingService, providers []ModelProvider) *SemanticConsensus {
	return &SemanticConsensus{
		embeddingService: embeddingService,
		providers:        providers,
		thresholds: ConsensusThresholds{
			MinSimilarity:    0.85,
			ClusterThreshold: 0.80,
			OutlierThreshold: 0.60,
			MinProviders:     3,
		},
	}
}

// ExecuteWithConsensus runs the same prompt across multiple providers and analyzes consensus
func (sc *SemanticConsensus) ExecuteWithConsensus(
	ctx context.Context,
	prompt string,
	params *ModelParameters,
) (*ConsensusResult, error) {
	start := time.Now()

	// 1. Execute across all providers in parallel
	responses, err := sc.executeAcrossProviders(ctx, prompt, params)
	if err != nil {
		return nil, fmt.Errorf("provider execution failed: %w", err)
	}

	if len(responses) < sc.thresholds.MinProviders {
		return nil, fmt.Errorf("insufficient providers responded: %d < %d",
			len(responses), sc.thresholds.MinProviders)
	}

	// 2. Generate embeddings for all responses
	if err := sc.generateEmbeddings(ctx, responses); err != nil {
		return nil, fmt.Errorf("embedding generation failed: %w", err)
	}

	// 3. Compute similarity matrix
	similarityMatrix := sc.computeSimilarityMatrix(responses)

	// 4. Perform clustering analysis
	clusters := sc.clusterResponses(responses, similarityMatrix)

	// 5. Detect potential manipulation
	manipulation := sc.detectManipulation(responses, clusters, similarityMatrix)

	// 6. Calculate consensus metrics
	metrics := sc.calculateConsensusMetrics(responses, clusters, similarityMatrix)

	// 7. Generate final result
	result := &ConsensusResult{
		PromptHash:           sc.generatePromptHash(prompt, params),
		ProviderResponses:    responses,
		SimilarityMatrix:     similarityMatrix,
		Clusters:             clusters,
		ManipulationAnalysis: manipulation,
		ConsensusMetrics:     metrics,
		Timestamp:            time.Now(),
	}

	// 8. Select consensus response
	if len(clusters) > 0 {
		result.ConsensusResponse = sc.selectConsensusResponse(responses, clusters[0])
		result.ConsensusEmbedding = clusters[0].Centroid
	}

	log.Printf("Semantic consensus completed in %v", time.Since(start))
	return result, nil
}

// executeAcrossProviders runs the prompt across all providers in parallel
func (sc *SemanticConsensus) executeAcrossProviders(
	ctx context.Context,
	prompt string,
	params *ModelParameters,
) ([]ProviderResponse, error) {
	type result struct {
		response ProviderResponse
		index    int
	}

	results := make(chan result, len(sc.providers))
	var wg sync.WaitGroup

	// Execute in parallel with timeout
	execCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	for i, provider := range sc.providers {
		wg.Add(1)
		go func(idx int, p ModelProvider) {
			defer wg.Done()

			start := time.Now()
			response, tokenUsage, err := p.Execute(execCtx, prompt, params)
			executionTime := time.Since(start)

			results <- result{
				response: ProviderResponse{
					ProviderID:    p.ID(),
					Response:      response,
					ExecutionTime: executionTime,
					TokenUsage:    tokenUsage,
					Error:         err,
				},
				index: idx,
			}
		}(i, provider)
	}

	// Wait for all to complete
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect successful responses
	var responses []ProviderResponse
	for res := range results {
		if res.response.Error == nil {
			responses = append(responses, res.response)
		} else {
			log.Printf("Provider %s failed: %v", res.response.ProviderID, res.response.Error)
		}
	}

	return responses, nil
}

// generateEmbeddings creates vector embeddings for all responses
func (sc *SemanticConsensus) generateEmbeddings(ctx context.Context, responses []ProviderResponse) error {
	for i := range responses {
		embedding, err := sc.embeddingService.Embed(ctx, responses[i].Response)
		if err != nil {
			return fmt.Errorf("embedding failed for provider %s: %w", responses[i].ProviderID, err)
		}
		responses[i].Embedding = embedding
	}
	return nil
}

// computeSimilarityMatrix calculates cosine similarity between all response pairs
func (sc *SemanticConsensus) computeSimilarityMatrix(responses []ProviderResponse) [][]float64 {
	n := len(responses)
	similarity := make([][]float64, n)

	for i := 0; i < n; i++ {
		similarity[i] = make([]float64, n)
		for j := 0; j < n; j++ {
			if i == j {
				similarity[i][j] = 1.0
			} else {
				similarity[i][j] = cosineSimilarity(responses[i].Embedding, responses[j].Embedding)
			}
		}
	}

	return similarity
}

// cosineSimilarity calculates cosine similarity between two vectors
func cosineSimilarity(a, b []float64) float64 {
	if len(a) != len(b) {
		return 0.0
	}

	var dotProduct, normA, normB float64

	for i := 0; i < len(a); i++ {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 0.0
	}

	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}

// clusterResponses groups similar responses using similarity thresholds
func (sc *SemanticConsensus) clusterResponses(
	responses []ProviderResponse,
	similarity [][]float64,
) []SemanticCluster {
	n := len(responses)
	visited := make([]bool, n)
	var clusters []SemanticCluster

	for i := 0; i < n; i++ {
		if visited[i] {
			continue
		}

		cluster := SemanticCluster{
			Members: []int{i},
		}
		visited[i] = true

		// Find all responses similar to this one
		for j := i + 1; j < n; j++ {
			if !visited[j] && similarity[i][j] >= sc.thresholds.ClusterThreshold {
				cluster.Members = append(cluster.Members, j)
				visited[j] = true
			}
		}

		cluster.Size = len(cluster.Members)
		cluster.Centroid = sc.computeClusterCentroid(responses, cluster.Members)
		cluster.Cohesion = sc.computeClusterCohesion(similarity, cluster.Members)
		cluster.Confidence = sc.computeClusterConfidence(cluster, similarity)

		clusters = append(clusters, cluster)
	}

	// Sort clusters by size (largest first)
	sort.Slice(clusters, func(i, j int) bool {
		return clusters[i].Size > clusters[j].Size
	})

	return clusters
}

// computeClusterCentroid calculates the average embedding for cluster members
func (sc *SemanticConsensus) computeClusterCentroid(
	responses []ProviderResponse,
	members []int,
) []float64 {
	if len(members) == 0 {
		return nil
	}

	embeddingSize := len(responses[members[0]].Embedding)
	centroid := make([]float64, embeddingSize)

	for _, memberIdx := range members {
		for i, val := range responses[memberIdx].Embedding {
			centroid[i] += val
		}
	}

	// Average the embeddings
	for i := range centroid {
		centroid[i] /= float64(len(members))
	}

	return centroid
}

// computeClusterCohesion calculates internal similarity within a cluster
func (sc *SemanticConsensus) computeClusterCohesion(
	similarity [][]float64,
	members []int,
) float64 {
	if len(members) <= 1 {
		return 1.0
	}

	var sum float64
	var count int

	for i := 0; i < len(members); i++ {
		for j := i + 1; j < len(members); j++ {
			sum += similarity[members[i]][members[j]]
			count++
		}
	}

	if count == 0 {
		return 1.0
	}

	return sum / float64(count)
}

// computeClusterConfidence calculates confidence score for a cluster
func (sc *SemanticConsensus) computeClusterConfidence(
	cluster SemanticCluster,
	similarity [][]float64,
) float64 {
	// Base confidence on cluster size and cohesion
	sizeScore := math.Min(float64(cluster.Size)/3.0, 1.0) // Normalize to max 3 providers
	cohesionScore := cluster.Cohesion

	return (sizeScore + cohesionScore) / 2.0
}

// detectManipulation identifies potentially manipulated responses
func (sc *SemanticConsensus) detectManipulation(
	responses []ProviderResponse,
	clusters []SemanticCluster,
	similarity [][]float64,
) *ManipulationAnalysis {
	analysis := &ManipulationAnalysis{
		SuspiciousProviders: []string{},
		DetectionReasons:    []string{},
	}

	if len(clusters) == 0 {
		analysis.ManipulationLikely = true
		analysis.DetectionReasons = append(analysis.DetectionReasons, "No valid clusters found")
		return analysis
	}

	majorityCluster := clusters[0] // Largest cluster

	// 1. Identify outlier responses
	for i, response := range responses {
		isInMajority := false
		for _, memberIdx := range majorityCluster.Members {
			if memberIdx == i {
				isInMajority = true
				break
			}
		}

		if !isInMajority {
			// Check if this response is a significant outlier
			maxSimilarity := 0.0
			for _, memberIdx := range majorityCluster.Members {
				if similarity[i][memberIdx] > maxSimilarity {
					maxSimilarity = similarity[i][memberIdx]
				}
			}

			if maxSimilarity < sc.thresholds.OutlierThreshold {
				analysis.SuspiciousProviders = append(analysis.SuspiciousProviders, response.ProviderID)
				analysis.DetectionReasons = append(analysis.DetectionReasons,
					fmt.Sprintf("Provider %s response semantically distant from consensus (similarity: %.2f)",
						response.ProviderID, maxSimilarity))
			}
		}
	}

	// 2. Calculate anomaly score
	analysis.AnomalyScore = sc.calculateAnomalyScore(clusters, similarity)

	// 3. Determine if manipulation is likely
	analysis.ManipulationLikely = len(analysis.SuspiciousProviders) > 0 || analysis.AnomalyScore > 0.5

	// 4. Calculate confidence level
	analysis.ConfidenceLevel = sc.calculateManipulationConfidence(clusters, analysis)

	return analysis
}

// calculateAnomalyScore computes overall anomaly score
func (sc *SemanticConsensus) calculateAnomalyScore(
	clusters []SemanticCluster,
	similarity [][]float64,
) float64 {
	if len(clusters) == 0 {
		return 1.0 // Maximum anomaly
	}

	majorityCluster := clusters[0]

	// Score based on consensus strength
	consensusStrength := float64(majorityCluster.Size) / float64(len(similarity))
	cohesionScore := majorityCluster.Cohesion

	// Higher consensus strength and cohesion = lower anomaly score
	anomalyScore := 1.0 - (consensusStrength+cohesionScore)/2.0

	return math.Max(0.0, math.Min(1.0, anomalyScore))
}

// calculateManipulationConfidence computes confidence in manipulation detection
func (sc *SemanticConsensus) calculateManipulationConfidence(
	clusters []SemanticCluster,
	analysis *ManipulationAnalysis,
) float64 {
	if len(clusters) == 0 {
		return 0.5 // Low confidence when no clusters
	}

	majorityCluster := clusters[0]

	// Base confidence on cluster strength and outlier clarity
	clusterConfidence := majorityCluster.Confidence
	outlierClarity := 1.0 - analysis.AnomalyScore

	return (clusterConfidence + outlierClarity) / 2.0
}

// calculateConsensusMetrics computes overall consensus statistics
func (sc *SemanticConsensus) calculateConsensusMetrics(
	responses []ProviderResponse,
	clusters []SemanticCluster,
	similarity [][]float64,
) *ConsensusMetrics {
	metrics := &ConsensusMetrics{}

	if len(clusters) == 0 {
		return metrics
	}

	majorityCluster := clusters[0]

	// Agreement ratio (percentage in majority cluster)
	metrics.AgreementRatio = float64(majorityCluster.Size) / float64(len(responses))

	// Average pairwise similarity
	var sum float64
	var count int
	for i := 0; i < len(similarity); i++ {
		for j := i + 1; j < len(similarity); j++ {
			sum += similarity[i][j]
			count++
		}
	}
	if count > 0 {
		metrics.AverageSimilarity = sum / float64(count)
	}

	// Consensus confidence (from majority cluster)
	metrics.ConsensusConfidence = majorityCluster.Confidence

	// Semantic coherence (majority cluster cohesion)
	metrics.SemanticCoherence = majorityCluster.Cohesion

	// Outlier count (responses not in majority cluster)
	metrics.OutlierCount = len(responses) - majorityCluster.Size

	return metrics
}

// selectConsensusResponse chooses the best response from the majority cluster
func (sc *SemanticConsensus) selectConsensusResponse(
	responses []ProviderResponse,
	majorityCluster SemanticCluster,
) string {
	if len(majorityCluster.Members) == 0 {
		return ""
	}

	// For now, select the first response in the majority cluster
	// TODO: Implement more sophisticated selection (e.g., closest to centroid)
	return responses[majorityCluster.Members[0]].Response
}

// generatePromptHash creates a deterministic hash for the prompt and parameters
func (sc *SemanticConsensus) generatePromptHash(prompt string, params *ModelParameters) string {
	// TODO: Implement proper canonical hashing
	return fmt.Sprintf("prompt_%x", time.Now().UnixNano())
}
