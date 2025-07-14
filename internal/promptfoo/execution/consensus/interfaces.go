package consensus

import (
	"context"
	"time"
)

// EmbeddingService generates vector embeddings for text
type EmbeddingService interface {
	Embed(ctx context.Context, text string) ([]float64, error)
	ModelName() string
	Dimensions() int
}

// ModelProvider represents an AI model provider (OpenAI, Anthropic, etc.)
type ModelProvider interface {
	ID() string
	Execute(ctx context.Context, prompt string, params *ModelParameters) (response string, tokenUsage TokenUsage, err error)
	SupportsParameters(params *ModelParameters) bool
}

// ModelParameters define model execution parameters
type ModelParameters struct {
	Model            string   `json:"model"`
	Temperature      float64  `json:"temperature"`
	MaxTokens        int      `json:"max_tokens"`
	TopP             float64  `json:"top_p,omitempty"`
	FrequencyPenalty float64  `json:"frequency_penalty,omitempty"`
	PresencePenalty  float64  `json:"presence_penalty,omitempty"`
	Seed             *int     `json:"seed,omitempty"`
	StopSequences    []string `json:"stop_sequences,omitempty"`
}

// TokenUsage tracks token consumption
type TokenUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// ConsensusEngine defines the interface for consensus mechanisms
type ConsensusEngine interface {
	ExecuteWithConsensus(ctx context.Context, prompt string, params *ModelParameters) (*ConsensusResult, error)
	SetThresholds(thresholds ConsensusThresholds)
	AddProvider(provider ModelProvider)
	RemoveProvider(providerID string)
}

// AttestationService handles cryptographic attestation
type AttestationService interface {
	GenerateAttestation(result *ConsensusResult) (*Attestation, error)
	VerifyAttestation(attestation *Attestation) error
	SignResult(result *ConsensusResult) (string, error)
}

// Attestation represents a cryptographic proof of consensus
type Attestation struct {
	ConsensusHash     string    `json:"consensus_hash"`
	ProviderCount     int       `json:"provider_count"`
	AgreementRatio    float64   `json:"agreement_ratio"`
	ManipulationScore float64   `json:"manipulation_score"`
	Timestamp         time.Time `json:"timestamp"`
	Signature         string    `json:"signature"`
	CertificateChain  []string  `json:"certificate_chain,omitempty"`
}

// Logger interface for consensus operations
type Logger interface {
	Info(msg string, fields ...interface{})
	Warn(msg string, fields ...interface{})
	Error(msg string, err error, fields ...interface{})
	Debug(msg string, fields ...interface{})
}
