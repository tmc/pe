// Package mocks provides configurable mock implementations for testing.
package mocks

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/tmc/pe/internal/inference"
)

// Compile-time interface compliance checks
var (
	_ inference.Provider = (*MockProvider)(nil)
	_ inference.Provider = (*AdvancedMockProvider)(nil)
	_ inference.Provider = (*ChaosProvider)(nil)
)

// MockProvider is a configurable mock implementation of inference.Provider for testing.
type MockProvider struct {
	mu                 sync.RWMutex
	name               string
	models             []string
	responses          map[string]*inference.Response // keyed by prompt
	streamResponses    map[string][]inference.StreamChunk
	errors             map[string]error // keyed by operation type
	latency            time.Duration
	callCounts         map[string]int
	lastRequest        *inference.Request
	shouldPanic        map[string]bool
	responseGenerator  func(req inference.Request) *inference.Response
	streamGenerator    func(req inference.Request) []inference.StreamChunk
	rateLimitBehavior  *RateLimitBehavior
	contextBehavior    *ContextBehavior
}

// RateLimitBehavior configures rate limiting simulation.
type RateLimitBehavior struct {
	MaxRequestsPerSecond int
	lastRequestTime      time.Time
	requestCount         int
}

// ContextBehavior configures context handling behavior.
type ContextBehavior struct {
	RespectCancellation bool
	BlockOnCancel       bool
}

// MockProviderConfig configures the mock provider behavior.
type MockProviderConfig struct {
	Name              string
	Models            []string
	DefaultLatency    time.Duration
	RespectContext    bool
	MaxRPS            int // Max requests per second for rate limiting
}

// NewMockProvider creates a new configurable mock provider.
func NewMockProvider(config MockProviderConfig) *MockProvider {
	if config.Name == "" {
		config.Name = "mock"
	}
	if config.Models == nil {
		config.Models = []string{"mock-model-1", "mock-model-2"}
	}

	mp := &MockProvider{
		name:            config.Name,
		models:          config.Models,
		responses:       make(map[string]*inference.Response),
		streamResponses: make(map[string][]inference.StreamChunk),
		errors:          make(map[string]error),
		latency:         config.DefaultLatency,
		callCounts:      make(map[string]int),
		shouldPanic:     make(map[string]bool),
		contextBehavior: &ContextBehavior{RespectCancellation: config.RespectContext},
	}

	if config.MaxRPS > 0 {
		mp.rateLimitBehavior = &RateLimitBehavior{
			MaxRequestsPerSecond: config.MaxRPS,
		}
	}

	return mp
}

// Name returns the provider name.
func (mp *MockProvider) Name() string {
	return mp.name
}

// SetResponse configures the response for a specific prompt.
func (mp *MockProvider) SetResponse(prompt string, response *inference.Response) {
	mp.mu.Lock()
	defer mp.mu.Unlock()
	mp.responses[prompt] = response
}

// SetStreamResponse configures the streaming response for a specific prompt.
func (mp *MockProvider) SetStreamResponse(prompt string, chunks []inference.StreamChunk) {
	mp.mu.Lock()
	defer mp.mu.Unlock()
	mp.streamResponses[prompt] = chunks
}

// SetError configures an error for a specific operation.
func (mp *MockProvider) SetError(operation string, err error) {
	mp.mu.Lock()
	defer mp.mu.Unlock()
	mp.errors[operation] = err
}

// SetLatency configures the simulated latency.
func (mp *MockProvider) SetLatency(latency time.Duration) {
	mp.mu.Lock()
	defer mp.mu.Unlock()
	mp.latency = latency
}

// SetPanic configures whether an operation should panic.
func (mp *MockProvider) SetPanic(operation string, shouldPanic bool) {
	mp.mu.Lock()
	defer mp.mu.Unlock()
	mp.shouldPanic[operation] = shouldPanic
}

// SetResponseGenerator configures a dynamic response generator.
func (mp *MockProvider) SetResponseGenerator(generator func(req inference.Request) *inference.Response) {
	mp.mu.Lock()
	defer mp.mu.Unlock()
	mp.responseGenerator = generator
}

// SetStreamGenerator configures a dynamic stream generator.
func (mp *MockProvider) SetStreamGenerator(generator func(req inference.Request) []inference.StreamChunk) {
	mp.mu.Lock()
	defer mp.mu.Unlock()
	mp.streamGenerator = generator
}

// Complete performs a mock non-streaming inference.
func (mp *MockProvider) Complete(ctx context.Context, req inference.Request) (*inference.Response, error) {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	mp.callCounts["Complete"]++
	mp.lastRequest = &req

	// Check for panic behavior
	if mp.shouldPanic["Complete"] {
		panic("mock provider panic in Complete")
	}

	// Check for configured error
	if err, exists := mp.errors["Complete"]; exists {
		return nil, err
	}

	// Simulate rate limiting
	if mp.rateLimitBehavior != nil {
		if err := mp.checkRateLimit(); err != nil {
			return nil, err
		}
	}

	// Simulate latency
	if mp.latency > 0 {
		select {
		case <-ctx.Done():
			if mp.contextBehavior.RespectCancellation {
				return nil, ctx.Err()
			}
		case <-time.After(mp.latency):
		}
	}

	// Check context cancellation
	if mp.contextBehavior.RespectCancellation {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
	}

	// Use response generator if available
	if mp.responseGenerator != nil {
		return mp.responseGenerator(req), nil
	}

	// Check for specific response
	if response, exists := mp.responses[req.Prompt]; exists {
		return response, nil
	}

	// Generate default response
	return mp.generateDefaultResponse(req), nil
}

// Stream performs a mock streaming inference.
func (mp *MockProvider) Stream(ctx context.Context, req inference.Request) (<-chan inference.StreamChunk, error) {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	mp.callCounts["Stream"]++
	mp.lastRequest = &req

	// Check for panic behavior
	if mp.shouldPanic["Stream"] {
		panic("mock provider panic in Stream")
	}

	// Check for configured error
	if err, exists := mp.errors["Stream"]; exists {
		return nil, err
	}

	// Simulate rate limiting
	if mp.rateLimitBehavior != nil {
		if err := mp.checkRateLimit(); err != nil {
			return nil, err
		}
	}

	chunks := make(chan inference.StreamChunk, 1)

	go func() {
		defer close(chunks)

		// Get chunks to stream
		var chunksToStream []inference.StreamChunk
		if mp.streamGenerator != nil {
			chunksToStream = mp.streamGenerator(req)
		} else if configuredChunks, exists := mp.streamResponses[req.Prompt]; exists {
			chunksToStream = configuredChunks
		} else {
			chunksToStream = mp.generateDefaultStreamChunks(req)
		}

		for _, chunk := range chunksToStream {
			// Check context cancellation
			if mp.contextBehavior.RespectCancellation {
				select {
				case <-ctx.Done():
					chunks <- inference.StreamChunk{Error: ctx.Err()}
					return
				default:
				}
			}

			// Simulate streaming delay
			if mp.latency > 0 {
				time.Sleep(mp.latency / time.Duration(len(chunksToStream)))
			}

			select {
			case <-ctx.Done():
				if mp.contextBehavior.RespectCancellation {
					chunks <- inference.StreamChunk{Error: ctx.Err()}
					return
				}
			case chunks <- chunk:
			}
		}

		// Send final done chunk
		chunks <- inference.StreamChunk{Done: true}
	}()

	return chunks, nil
}

// Models returns the available models.
func (mp *MockProvider) Models(ctx context.Context) ([]string, error) {
	mp.mu.RLock()
	defer mp.mu.RUnlock()

	mp.callCounts["Models"]++

	// Check for panic behavior
	if mp.shouldPanic["Models"] {
		panic("mock provider panic in Models")
	}

	// Check for configured error
	if err, exists := mp.errors["Models"]; exists {
		return nil, err
	}

	// Check context cancellation
	if mp.contextBehavior.RespectCancellation {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
	}

	return mp.models, nil
}

// Close cleans up resources.
func (mp *MockProvider) Close() error {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	mp.callCounts["Close"]++

	// Check for panic behavior
	if mp.shouldPanic["Close"] {
		panic("mock provider panic in Close")
	}

	// Check for configured error
	if err, exists := mp.errors["Close"]; exists {
		return err
	}

	return nil
}

// GetCallCount returns the number of times a method was called.
func (mp *MockProvider) GetCallCount(method string) int {
	mp.mu.RLock()
	defer mp.mu.RUnlock()
	return mp.callCounts[method]
}

// GetLastRequest returns the last inference request received.
func (mp *MockProvider) GetLastRequest() *inference.Request {
	mp.mu.RLock()
	defer mp.mu.RUnlock()
	return mp.lastRequest
}

// Reset clears all mock state.
func (mp *MockProvider) Reset() {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	mp.responses = make(map[string]*inference.Response)
	mp.streamResponses = make(map[string][]inference.StreamChunk)
	mp.errors = make(map[string]error)
	mp.callCounts = make(map[string]int)
	mp.shouldPanic = make(map[string]bool)
	mp.lastRequest = nil
	mp.responseGenerator = nil
	mp.streamGenerator = nil
}

// checkRateLimit simulates rate limiting behavior.
func (mp *MockProvider) checkRateLimit() error {
	if mp.rateLimitBehavior == nil {
		return nil
	}

	now := time.Now()
	if now.Sub(mp.rateLimitBehavior.lastRequestTime) > time.Second {
		mp.rateLimitBehavior.requestCount = 0
		mp.rateLimitBehavior.lastRequestTime = now
	}

	mp.rateLimitBehavior.requestCount++
	if mp.rateLimitBehavior.requestCount > mp.rateLimitBehavior.MaxRequestsPerSecond {
		return fmt.Errorf("rate limit exceeded: max %d requests per second", 
			mp.rateLimitBehavior.MaxRequestsPerSecond)
	}

	return nil
}

// generateDefaultResponse creates a default response for testing.
func (mp *MockProvider) generateDefaultResponse(req inference.Request) *inference.Response {
	return &inference.Response{
		Content: fmt.Sprintf("Mock response to: %s", req.Prompt),
		Model:   req.Model,
		TokensUsed: inference.TokenUsage{
			PromptTokens:     len(req.Prompt) / 4, // Rough estimate
			CompletionTokens: 50,
			TotalTokens:      (len(req.Prompt) / 4) + 50,
		},
		Metadata: map[string]interface{}{
			"provider":    mp.name,
			"mock":        true,
			"request_id":  fmt.Sprintf("mock_%d", time.Now().UnixNano()),
		},
	}
}

// generateDefaultStreamChunks creates default stream chunks for testing.
func (mp *MockProvider) generateDefaultStreamChunks(req inference.Request) []inference.StreamChunk {
	words := []string{"Mock", "streaming", "response", "to:", req.Prompt}
	
	chunks := make([]inference.StreamChunk, len(words))
	for i, word := range words {
		chunks[i] = inference.StreamChunk{
			Delta: word + " ",
		}
	}
	
	return chunks
}

// AdvancedMockProvider extends MockProvider with more sophisticated behaviors.
type AdvancedMockProvider struct {
	*MockProvider
	failureRate     float64
	randomResponses bool
	rand            *rand.Rand
	metrics         *ProviderMetrics
}

// ProviderMetrics tracks various provider metrics for testing.
type ProviderMetrics struct {
	mu                 sync.RWMutex
	TotalRequests      int
	SuccessfulRequests int
	FailedRequests     int
	AverageLatency     time.Duration
	TotalLatency       time.Duration
	MinLatency         time.Duration
	MaxLatency         time.Duration
}

// ProviderMetricsSnapshot is a copy of metrics without the mutex for safe return.
type ProviderMetricsSnapshot struct {
	TotalRequests      int
	SuccessfulRequests int
	FailedRequests     int
	AverageLatency     time.Duration
	TotalLatency       time.Duration
	MinLatency         time.Duration
	MaxLatency         time.Duration
}

// NewAdvancedMockProvider creates a mock provider with advanced behaviors.
func NewAdvancedMockProvider(config MockProviderConfig) *AdvancedMockProvider {
	return &AdvancedMockProvider{
		MockProvider: NewMockProvider(config),
		rand:        rand.New(rand.NewSource(time.Now().UnixNano())),
		metrics:     &ProviderMetrics{},
	}
}

// SetFailureRate sets the probability of requests failing (0.0 to 1.0).
func (amp *AdvancedMockProvider) SetFailureRate(rate float64) {
	amp.failureRate = rate
}

// SetRandomResponses enables/disables random response generation.
func (amp *AdvancedMockProvider) SetRandomResponses(enabled bool) {
	amp.randomResponses = enabled
}

// Complete overrides the base Complete method with advanced behaviors.
func (amp *AdvancedMockProvider) Complete(ctx context.Context, req inference.Request) (*inference.Response, error) {
	start := time.Now()
	
	// Record metrics
	amp.metrics.mu.Lock()
	amp.metrics.TotalRequests++
	amp.metrics.mu.Unlock()
	
	// Simulate random failures
	if amp.rand.Float64() < amp.failureRate {
		amp.metrics.mu.Lock()
		amp.metrics.FailedRequests++
		amp.metrics.mu.Unlock()
		return nil, fmt.Errorf("simulated random failure")
	}
	
	// Call base implementation
	resp, err := amp.MockProvider.Complete(ctx, req)
	
	// Update metrics
	latency := time.Since(start)
	amp.metrics.mu.Lock()
	if err != nil {
		amp.metrics.FailedRequests++
	} else {
		amp.metrics.SuccessfulRequests++
	}
	amp.metrics.TotalLatency += latency
	amp.metrics.AverageLatency = amp.metrics.TotalLatency / time.Duration(amp.metrics.TotalRequests)
	if amp.metrics.MinLatency == 0 || latency < amp.metrics.MinLatency {
		amp.metrics.MinLatency = latency
	}
	if latency > amp.metrics.MaxLatency {
		amp.metrics.MaxLatency = latency
	}
	amp.metrics.mu.Unlock()
	
	// Modify response if random responses enabled
	if resp != nil && amp.randomResponses {
		resp.Content = amp.generateRandomContent()
		resp.TokensUsed.CompletionTokens = amp.rand.Intn(1000) + 50
		resp.TokensUsed.TotalTokens = resp.TokensUsed.PromptTokens + resp.TokensUsed.CompletionTokens
	}
	
	return resp, err
}

// GetMetrics returns current provider metrics as a snapshot (without mutex).
func (amp *AdvancedMockProvider) GetMetrics() ProviderMetricsSnapshot {
	amp.metrics.mu.RLock()
	defer amp.metrics.mu.RUnlock()
	return ProviderMetricsSnapshot{
		TotalRequests:      amp.metrics.TotalRequests,
		SuccessfulRequests: amp.metrics.SuccessfulRequests,
		FailedRequests:     amp.metrics.FailedRequests,
		AverageLatency:     amp.metrics.AverageLatency,
		TotalLatency:       amp.metrics.TotalLatency,
		MinLatency:         amp.metrics.MinLatency,
		MaxLatency:         amp.metrics.MaxLatency,
	}
}

// generateRandomContent generates random content for testing.
func (amp *AdvancedMockProvider) generateRandomContent() string {
	responses := []string{
		"This is a random test response.",
		"Here's another random response for testing purposes.",
		"Random content generated by the advanced mock provider.",
		"Testing with dynamic random responses.",
		"Another random response to ensure variety in testing.",
	}
	return responses[amp.rand.Intn(len(responses))]
}

// ChaosProvider is a mock provider that simulates various failure modes.
type ChaosProvider struct {
	*AdvancedMockProvider
	chaosConfig ChaosConfig
}

// ChaosConfig configures chaos engineering behaviors.
type ChaosConfig struct {
	TimeoutProbability    float64
	PanicProbability     float64
	CorruptionProbability float64
	SlowResponseProbability float64
	SlowResponseDelay     time.Duration
}

// NewChaosProvider creates a chaos engineering mock provider.
func NewChaosProvider(config MockProviderConfig, chaosConfig ChaosConfig) *ChaosProvider {
	return &ChaosProvider{
		AdvancedMockProvider: NewAdvancedMockProvider(config),
		chaosConfig:         chaosConfig,
	}
}

// Complete implements chaos behaviors in the Complete method.
func (cp *ChaosProvider) Complete(ctx context.Context, req inference.Request) (*inference.Response, error) {
	// Simulate timeout
	if cp.rand.Float64() < cp.chaosConfig.TimeoutProbability {
		<-time.After(5 * time.Second) // Force timeout
		return nil, fmt.Errorf("chaos timeout")
	}
	
	// Simulate panic
	if cp.rand.Float64() < cp.chaosConfig.PanicProbability {
		panic("chaos panic")
	}
	
	// Simulate slow response
	if cp.rand.Float64() < cp.chaosConfig.SlowResponseProbability {
		time.Sleep(cp.chaosConfig.SlowResponseDelay)
	}
	
	// Get normal response
	resp, err := cp.AdvancedMockProvider.Complete(ctx, req)
	if err != nil {
		return resp, err
	}
	
	// Simulate corruption
	if resp != nil && cp.rand.Float64() < cp.chaosConfig.CorruptionProbability {
		resp.Content = "CORRUPTED_RESPONSE_" + resp.Content
		resp.TokensUsed.TotalTokens = -1 // Invalid token count
	}
	
	return resp, err
}