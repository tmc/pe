package attestation

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// RunAttestation represents a cryptographically signed record of a prompt execution
type RunAttestation struct {
	// Metadata
	ID        string    `json:"id"`         // Unique run ID
	Timestamp time.Time `json:"timestamp"`  // When the run occurred
	Version   string    `json:"version"`    // PE version
	
	// Input
	Prompt       string                 `json:"prompt"`        // The prompt template
	Variables    map[string]interface{} `json:"variables"`     // Template variables
	Provider     string                 `json:"provider"`      // LLM provider used
	Model        string                 `json:"model"`         // Specific model
	Temperature  float32                `json:"temperature"`   // Temperature setting
	SystemPrompt string                 `json:"system_prompt"` // System prompt if any
	
	// Output
	Response     string        `json:"response"`      // The LLM response
	TokensUsed   TokenUsage    `json:"tokens_used"`   // Token usage
	Latency      time.Duration `json:"latency"`       // Response time
	FinishReason string        `json:"finish_reason"` // Why generation stopped
	
	// Verification
	InputHash    string `json:"input_hash"`    // SHA-256 of canonicalized input
	OutputHash   string `json:"output_hash"`   // SHA-256 of response
	PreviousHash string `json:"previous_hash"` // Chain to previous run
	Signature    string `json:"signature"`     // Ed25519 signature
	PublicKey    string `json:"public_key"`    // Public key used for signing
}

// TokenUsage tracks token consumption
type TokenUsage struct {
	Prompt     int `json:"prompt"`
	Completion int `json:"completion"`
	Total      int `json:"total"`
}

// AttestationService handles cryptographic attestation of runs
type AttestationService struct {
	privateKey ed25519.PrivateKey
	publicKey  ed25519.PublicKey
	chainFile  string
	storageDir string
}

// NewAttestationService creates a new attestation service
func NewAttestationService(dataDir string) (*AttestationService, error) {
	storageDir := filepath.Join(dataDir, "attestations")
	if err := os.MkdirAll(storageDir, 0700); err != nil {
		return nil, fmt.Errorf("creating attestation directory: %w", err)
	}
	
	// Load or generate key pair
	keyFile := filepath.Join(storageDir, "signing.key")
	pubKeyFile := filepath.Join(storageDir, "signing.pub")
	
	var privateKey ed25519.PrivateKey
	var publicKey ed25519.PublicKey
	
	if keyData, err := os.ReadFile(keyFile); err == nil {
		// Load existing key
		privateKey = ed25519.PrivateKey(keyData)
		publicKey = privateKey.Public().(ed25519.PublicKey)
	} else {
		// Generate new key pair
		publicKey, privateKey, err = ed25519.GenerateKey(rand.Reader)
		if err != nil {
			return nil, fmt.Errorf("generating key pair: %w", err)
		}
		
		// Save keys
		if err := os.WriteFile(keyFile, privateKey, 0600); err != nil {
			return nil, fmt.Errorf("saving private key: %w", err)
		}
		if err := os.WriteFile(pubKeyFile, publicKey, 0644); err != nil {
			return nil, fmt.Errorf("saving public key: %w", err)
		}
	}
	
	return &AttestationService{
		privateKey: privateKey,
		publicKey:  publicKey,
		chainFile:  filepath.Join(storageDir, "chain.jsonl"),
		storageDir: storageDir,
	}, nil
}

// AttestRun creates a cryptographic attestation for a prompt run
func (s *AttestationService) AttestRun(input RunInput, output RunOutput) (*RunAttestation, error) {
	// Generate unique ID
	id := generateRunID()
	
	// Get previous hash for chaining
	previousHash, err := s.getLastHash()
	if err != nil {
		return nil, fmt.Errorf("getting previous hash: %w", err)
	}
	
	// Create attestation
	attestation := &RunAttestation{
		ID:           id,
		Timestamp:    time.Now().UTC(),
		Version:      getPEVersion(),
		Prompt:       input.Prompt,
		Variables:    input.Variables,
		Provider:     input.Provider,
		Model:        input.Model,
		Temperature:  input.Temperature,
		SystemPrompt: input.SystemPrompt,
		Response:     output.Response,
		TokensUsed: TokenUsage{
			Prompt:     output.PromptTokens,
			Completion: output.CompletionTokens,
			Total:      output.TotalTokens,
		},
		Latency:      output.Latency,
		FinishReason: output.FinishReason,
		PreviousHash: previousHash,
		PublicKey:    base64.StdEncoding.EncodeToString(s.publicKey),
	}
	
	// Calculate hashes
	attestation.InputHash = s.hashInput(input)
	attestation.OutputHash = s.hashOutput(output)
	
	// Sign the attestation
	signature, err := s.sign(attestation)
	if err != nil {
		return nil, fmt.Errorf("signing attestation: %w", err)
	}
	attestation.Signature = base64.StdEncoding.EncodeToString(signature)
	
	// Store attestation
	if err := s.store(attestation); err != nil {
		return nil, fmt.Errorf("storing attestation: %w", err)
	}
	
	return attestation, nil
}

// Verify checks if an attestation is valid
func (s *AttestationService) Verify(attestation *RunAttestation) error {
	// Decode public key
	pubKeyBytes, err := base64.StdEncoding.DecodeString(attestation.PublicKey)
	if err != nil {
		return fmt.Errorf("decoding public key: %w", err)
	}
	publicKey := ed25519.PublicKey(pubKeyBytes)
	
	// Decode signature
	signature, err := base64.StdEncoding.DecodeString(attestation.Signature)
	if err != nil {
		return fmt.Errorf("decoding signature: %w", err)
	}
	
	// Verify signature
	message := s.attestationMessage(attestation)
	if !ed25519.Verify(publicKey, message, signature) {
		return fmt.Errorf("invalid signature")
	}
	
	// Verify input hash
	input := RunInput{
		Prompt:       attestation.Prompt,
		Variables:    attestation.Variables,
		Provider:     attestation.Provider,
		Model:        attestation.Model,
		Temperature:  attestation.Temperature,
		SystemPrompt: attestation.SystemPrompt,
	}
	if s.hashInput(input) != attestation.InputHash {
		return fmt.Errorf("input hash mismatch")
	}
	
	// Verify output hash
	output := RunOutput{
		Response:         attestation.Response,
		PromptTokens:     attestation.TokensUsed.Prompt,
		CompletionTokens: attestation.TokensUsed.Completion,
		TotalTokens:      attestation.TokensUsed.Total,
		Latency:          attestation.Latency,
		FinishReason:     attestation.FinishReason,
	}
	if s.hashOutput(output) != attestation.OutputHash {
		return fmt.Errorf("output hash mismatch")
	}
	
	return nil
}

// VerifyChain verifies the entire chain of attestations
func (s *AttestationService) VerifyChain() error {
	file, err := os.Open(s.chainFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // Empty chain is valid
		}
		return fmt.Errorf("opening chain file: %w", err)
	}
	defer file.Close()
	
	var previousHash string
	decoder := json.NewDecoder(file)
	
	for decoder.More() {
		var attestation RunAttestation
		if err := decoder.Decode(&attestation); err != nil {
			return fmt.Errorf("decoding attestation: %w", err)
		}
		
		// Verify attestation
		if err := s.Verify(&attestation); err != nil {
			return fmt.Errorf("verifying attestation %s: %w", attestation.ID, err)
		}
		
		// Verify chain
		if attestation.PreviousHash != previousHash {
			return fmt.Errorf("chain broken at %s: expected previous hash %s, got %s",
				attestation.ID, previousHash, attestation.PreviousHash)
		}
		
		// Update previous hash for next iteration
		previousHash = s.attestationHash(&attestation)
	}
	
	return nil
}

// Helper methods

func (s *AttestationService) hashInput(input RunInput) string {
	// Canonicalize input for consistent hashing
	canonical := map[string]interface{}{
		"prompt":        input.Prompt,
		"variables":     input.Variables,
		"provider":      input.Provider,
		"model":         input.Model,
		"temperature":   input.Temperature,
		"system_prompt": input.SystemPrompt,
	}
	
	data, _ := json.Marshal(canonical)
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

func (s *AttestationService) hashOutput(output RunOutput) string {
	canonical := map[string]interface{}{
		"response":          output.Response,
		"prompt_tokens":     output.PromptTokens,
		"completion_tokens": output.CompletionTokens,
		"total_tokens":      output.TotalTokens,
		"latency_ns":        output.Latency.Nanoseconds(),
		"finish_reason":     output.FinishReason,
	}
	
	data, _ := json.Marshal(canonical)
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

func (s *AttestationService) attestationMessage(a *RunAttestation) []byte {
	// Create canonical message for signing
	message := map[string]interface{}{
		"id":            a.ID,
		"timestamp":     a.Timestamp.Unix(),
		"version":       a.Version,
		"input_hash":    a.InputHash,
		"output_hash":   a.OutputHash,
		"previous_hash": a.PreviousHash,
	}
	
	data, _ := json.Marshal(message)
	return data
}

func (s *AttestationService) attestationHash(a *RunAttestation) string {
	// Hash of the entire attestation for chaining
	data, _ := json.Marshal(a)
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

func (s *AttestationService) sign(attestation *RunAttestation) ([]byte, error) {
	message := s.attestationMessage(attestation)
	signature := ed25519.Sign(s.privateKey, message)
	return signature, nil
}

func (s *AttestationService) store(attestation *RunAttestation) error {
	// Append to chain file
	file, err := os.OpenFile(s.chainFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("opening chain file: %w", err)
	}
	defer file.Close()
	
	encoder := json.NewEncoder(file)
	if err := encoder.Encode(attestation); err != nil {
		return fmt.Errorf("encoding attestation: %w", err)
	}
	
	// Also store individual attestation
	attestFile := filepath.Join(s.storageDir, fmt.Sprintf("%s.json", attestation.ID))
	data, err := json.MarshalIndent(attestation, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling attestation: %w", err)
	}
	
	if err := os.WriteFile(attestFile, data, 0644); err != nil {
		return fmt.Errorf("writing attestation file: %w", err)
	}
	
	return nil
}

func (s *AttestationService) getLastHash() (string, error) {
	file, err := os.Open(s.chainFile)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil // First attestation
		}
		return "", fmt.Errorf("opening chain file: %w", err)
	}
	defer file.Close()
	
	var lastAttestation *RunAttestation
	decoder := json.NewDecoder(file)
	
	for decoder.More() {
		var attestation RunAttestation
		if err := decoder.Decode(&attestation); err != nil {
			return "", fmt.Errorf("decoding attestation: %w", err)
		}
		lastAttestation = &attestation
	}
	
	if lastAttestation == nil {
		return "", nil
	}
	
	return s.attestationHash(lastAttestation), nil
}

func generateRunID() string {
	// Generate a unique run ID
	timestamp := time.Now().UnixNano()
	randomBytes := make([]byte, 8)
	rand.Read(randomBytes)
	return fmt.Sprintf("run-%d-%x", timestamp, randomBytes)
}

func getPEVersion() string {
	// TODO: Get actual PE version
	return "0.1.0"
}

// Input/Output structs for attestation

type RunInput struct {
	Prompt       string
	Variables    map[string]interface{}
	Provider     string
	Model        string
	Temperature  float32
	SystemPrompt string
}

type RunOutput struct {
	Response         string
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
	Latency          time.Duration
	FinishReason     string
}