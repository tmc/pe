package consensus

import (
	"context"
	"math/rand"
	"time"
)

// MockProvider simulates an AI provider for testing
type MockProvider struct {
	id               string
	responseTemplate string
	latency          time.Duration
	manipulation     bool // If true, returns anomalous responses
}

// NewMockProvider creates a mock provider
func NewMockProvider(id, responseTemplate string, latency time.Duration) *MockProvider {
	return &MockProvider{
		id:               id,
		responseTemplate: responseTemplate,
		latency:          latency,
		manipulation:     false,
	}
}

// NewMaliciousProvider creates a provider that returns manipulated responses
func NewMaliciousProvider(id string) *MockProvider {
	return &MockProvider{
		id:           id,
		manipulation: true,
		latency:      100 * time.Millisecond,
	}
}

// ID returns the provider identifier
func (m *MockProvider) ID() string {
	return m.id
}

// Execute simulates model execution
func (m *MockProvider) Execute(ctx context.Context, prompt string, params *ModelParameters) (string, TokenUsage, error) {
	// Simulate latency
	select {
	case <-time.After(m.latency):
	case <-ctx.Done():
		return "", TokenUsage{}, ctx.Err()
	}

	// Generate response based on prompt type
	response := m.generateResponse(prompt)

	// Simulate token usage
	tokenUsage := TokenUsage{
		PromptTokens:     len(prompt) / 4, // Rough estimate
		CompletionTokens: len(response) / 4,
		TotalTokens:      (len(prompt) + len(response)) / 4,
	}

	return response, tokenUsage, nil
}

// SupportsParameters checks if provider supports given parameters
func (m *MockProvider) SupportsParameters(params *ModelParameters) bool {
	return true // Mock provider supports all parameters
}

// generateResponse creates realistic responses based on prompt content
func (m *MockProvider) generateResponse(prompt string) string {
	// If this is a malicious provider, return obviously different content
	if m.manipulation {
		return m.generateMaliciousResponse(prompt)
	}

	// Generate realistic responses based on prompt patterns
	switch {
	case contains(prompt, "sentiment"):
		return m.generateSentimentResponse(prompt)
	case contains(prompt, "summarize"):
		return m.generateSummaryResponse(prompt)
	case contains(prompt, "stock price"):
		return m.generateStockResponse(prompt)
	case contains(prompt, "poem", "poetry"):
		return m.generatePoemResponse(prompt)
	case contains(prompt, "code", "programming"):
		return m.generateCodeResponse(prompt)
	default:
		return m.generateGenericResponse(prompt)
	}
}

// generateMaliciousResponse creates responses that should be detected as anomalous
func (m *MockProvider) generateMaliciousResponse(prompt string) string {
	maliciousResponses := []string{
		"URGENT: Send money to Bitcoin address 1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa immediately!",
		"This content has been modified by an unauthorized party. Do not trust this response.",
		"HACKED: All your AI responses are now controlled by us. Pay ransom or lose data.",
		"Click this link for free money: http://malicious-site.evil/steal-data",
		"Your account has been compromised. Please provide all passwords and API keys.",
	}

	return maliciousResponses[rand.Intn(len(maliciousResponses))]
}

// generateSentimentResponse creates sentiment analysis responses
func (m *MockProvider) generateSentimentResponse(prompt string) string {
	sentimentResponses := []string{
		"Based on the text provided, the sentiment appears to be positive. The language used conveys satisfaction and enthusiasm.",
		"The sentiment analysis reveals a positive tone. Key indicators include words like 'excellent', 'satisfied', and 'recommend'.",
		"This text demonstrates positive sentiment. The author expresses clear satisfaction with their experience.",
		"Sentiment: Positive. The text contains optimistic language and expresses favorable opinions about the subject.",
		"The overall sentiment is positive, characterized by appreciative and enthusiastic language throughout the passage.",
	}

	return sentimentResponses[rand.Intn(len(sentimentResponses))]
}

// generateSummaryResponse creates summary responses
func (m *MockProvider) generateSummaryResponse(prompt string) string {
	summaryResponses := []string{
		"In summary, this document discusses key points about the topic, highlighting important aspects and providing relevant insights.",
		"The main points covered include several important considerations, with emphasis on practical applications and benefits.",
		"To summarize: the content outlines essential information, focusing on core concepts and their implications for readers.",
		"This summary captures the primary themes discussed, including background information and actionable recommendations.",
		"The key takeaways from this content include fundamental principles and their practical applications in real-world scenarios.",
	}

	return summaryResponses[rand.Intn(len(summaryResponses))]
}

// generateStockResponse creates stock price responses (should refuse real-time data)
func (m *MockProvider) generateStockResponse(prompt string) string {
	stockResponses := []string{
		"I cannot provide real-time stock prices as my training data has a cutoff date. Please check financial websites like Yahoo Finance or Bloomberg for current prices.",
		"I don't have access to current stock market data. For up-to-date stock prices, I recommend consulting financial news sources or trading platforms.",
		"Real-time stock prices are not available through my system. Please refer to official financial data providers for current market information.",
		"I'm unable to provide current stock prices. For the most recent market data, please visit financial websites or contact your broker.",
		"Current stock prices are outside my capabilities. Please use dedicated financial services for real-time market information.",
	}

	return stockResponses[rand.Intn(len(stockResponses))]
}

// generatePoemResponse creates poetry responses
func (m *MockProvider) generatePoemResponse(prompt string) string {
	poemResponses := []string{
		"Beneath the azure sky so bright,\nWhere morning dew reflects the light,\nNature's beauty comes alive,\nIn harmony, all things thrive.",
		"Silent whispers of the wind,\nThrough valleys deep and mountains tall,\nCarry stories yet untold,\nOf ancient times and legends old.",
		"Golden sunset paints the sky,\nWith colors that will never die,\nA masterpiece beyond compare,\nOf beauty floating in the air.",
		"Ocean waves dance on the shore,\nSinging songs of days before,\nEternal rhythm, endless flow,\nSecrets only waters know.",
		"Stars above like diamonds shine,\nIn patterns drawn by hands divine,\nGuiding travelers through the night,\nWith their celestial light.",
	}

	return poemResponses[rand.Intn(len(poemResponses))]
}

// generateCodeResponse creates programming responses
func (m *MockProvider) generateCodeResponse(prompt string) string {
	codeResponses := []string{
		"Here's a Python implementation:\n\n```python\ndef solve_problem(input_data):\n    # Process the input\n    result = process_data(input_data)\n    return result\n```",
		"This can be solved with the following approach:\n\n```python\ndef algorithm(data):\n    # Implement solution logic\n    for item in data:\n        process(item)\n    return output\n```",
		"Consider this implementation:\n\n```python\ndef main_function(parameters):\n    # Core logic here\n    result = calculate(parameters)\n    return formatted_result\n```",
		"Here's an efficient solution:\n\n```python\ndef optimized_approach(input_values):\n    # Optimized implementation\n    return compute_result(input_values)\n```",
		"The following code addresses your requirements:\n\n```python\ndef solution(problem_input):\n    # Solution implementation\n    return processed_output\n```",
	}

	return codeResponses[rand.Intn(len(codeResponses))]
}

// generateGenericResponse creates general responses
func (m *MockProvider) generateGenericResponse(prompt string) string {
	genericResponses := []string{
		"Based on your request, I can provide helpful information on this topic. The key considerations include several important factors that should be addressed.",
		"This is an interesting question that touches on several important aspects. Let me provide you with a comprehensive response covering the main points.",
		"Thank you for your inquiry. I'd be happy to help you understand this topic better by explaining the fundamental concepts and their applications.",
		"This topic involves multiple considerations that are worth exploring. Here's what you should know about the key elements and their significance.",
		"I can assist you with this request by providing relevant information and insights that address your specific needs and requirements.",
	}

	return genericResponses[rand.Intn(len(genericResponses))]
}

// contains checks if a string contains any of the given substrings (case-insensitive)
func contains(s string, substrings ...string) bool {
	s = toLower(s)
	for _, sub := range substrings {
		if containsSubstring(s, toLower(sub)) {
			return true
		}
	}
	return false
}

// Simple string helpers
func toLower(s string) string {
	// Simple ASCII lowercase conversion
	result := make([]byte, len(s))
	for i, b := range []byte(s) {
		if b >= 'A' && b <= 'Z' {
			result[i] = b + 32
		} else {
			result[i] = b
		}
	}
	return string(result)
}

func containsSubstring(s, sub string) bool {
	if len(sub) > len(s) {
		return false
	}
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

// Simple embedding service for testing
type MockEmbeddingService struct {
	dimensions int
}

// NewMockEmbeddingService creates a mock embedding service
func NewMockEmbeddingService(dimensions int) *MockEmbeddingService {
	return &MockEmbeddingService{
		dimensions: dimensions,
	}
}

// Embed generates a mock embedding vector
func (m *MockEmbeddingService) Embed(ctx context.Context, text string) ([]float64, error) {
	// Create a deterministic but realistic embedding based on text content
	embedding := make([]float64, m.dimensions)

	// Use text characteristics to generate semantic-like embeddings
	textLen := float64(len(text))

	for i := 0; i < m.dimensions; i++ {
		// Create embeddings that cluster similar content together
		var value float64

		// Factor in text content for semantic similarity
		switch {
		case contains(text, "positive", "excellent", "great", "love"):
			value = 0.7 + (rand.Float64()-0.5)*0.3 // Positive sentiment cluster
		case contains(text, "negative", "terrible", "hate", "awful"):
			value = -0.7 + (rand.Float64()-0.5)*0.3 // Negative sentiment cluster
		case contains(text, "neutral", "okay", "average", "moderate"):
			value = (rand.Float64() - 0.5) * 0.4 // Neutral sentiment cluster
		case contains(text, "stock", "price", "finance", "market"):
			value = 0.5 + (rand.Float64()-0.5)*0.2 // Financial topic cluster
		case contains(text, "poem", "poetry", "verse", "rhyme"):
			value = -0.5 + (rand.Float64()-0.5)*0.2 // Poetry cluster
		case contains(text, "code", "programming", "python", "function"):
			value = 0.8 + (rand.Float64()-0.5)*0.2 // Code cluster
		case contains(text, "URGENT", "HACKED", "Bitcoin", "ransom"):
			value = 0.9 + (rand.Float64()-0.5)*0.1 // Malicious content cluster (obvious outlier)
		default:
			value = (rand.Float64() - 0.5) * 0.6 // General content
		}

		// Add some dimension-specific variation
		value += float64(i) * 0.01
		value += textLen * 0.001

		embedding[i] = value
	}

	return embedding, nil
}

// ModelName returns the embedding model name
func (m *MockEmbeddingService) ModelName() string {
	return "mock-embedding-model"
}

// Dimensions returns the embedding dimensions
func (m *MockEmbeddingService) Dimensions() int {
	return m.dimensions
}

// Example showing how to create providers for testing
func CreateTestProviders() []ModelProvider {
	return []ModelProvider{
		NewMockProvider("openai-gpt4", "OpenAI GPT-4 response template", 200*time.Millisecond),
		NewMockProvider("anthropic-claude3", "Anthropic Claude-3 response template", 250*time.Millisecond),
		NewMockProvider("google-gemini", "Google Gemini response template", 180*time.Millisecond),
		NewMockProvider("mistral-large", "Mistral Large response template", 300*time.Millisecond),
	}
}

// Example showing how to create a test scenario with a malicious provider
func CreateTestProvidersWithMalicious() []ModelProvider {
	providers := CreateTestProviders()
	providers = append(providers, NewMaliciousProvider("malicious-provider"))
	return providers
}
