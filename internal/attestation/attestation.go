package attestation

import (
	"crypto"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// RunInput captures all inputs to a prompt run
type RunInput struct {
	Prompt       string                 `json:"prompt"`
	Variables    map[string]interface{} `json:"variables,omitempty"`
	Provider     string                 `json:"provider"`
	Model        string                 `json:"model"`
	Temperature  float32                `json:"temperature"`
	SystemPrompt string                 `json:"system_prompt,omitempty"`
}

// RunOutput captures the results of a prompt run
type RunOutput struct {
	Response         string        `json:"response"`
	PromptTokens     int           `json:"prompt_tokens"`
	CompletionTokens int           `json:"completion_tokens"`
	TotalTokens      int           `json:"total_tokens"`
	Latency          time.Duration `json:"latency"`
	FinishReason     string        `json:"finish_reason"`
}

// RunAttestation is a cryptographic attestation of a prompt run
type RunAttestation struct {
	// Metadata
	ID        string    `json:"id"`        // Unique identifier
	Timestamp time.Time `json:"timestamp"` // When the run occurred
	Version   string    `json:"version"`   // Attestation format version
	
	// Input
	Prompt       string                 `json:"prompt"`        // The prompt template
	Variables    map[string]interface{} `json:"variables"`     // Template variables
	Provider     string                 `json:"provider"`      // LLM provider used
	Model        string                 `json:"model"`         // Model identifier
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
	keyStore   *KeyStore
}

// NewAttestationService creates a new attestation service
func NewAttestationService(dataDir string) (*AttestationService, error) {
	storageDir := filepath.Join(dataDir, "attestations")
	if err := os.MkdirAll(storageDir, 0700); err != nil {
		return nil, fmt.Errorf("creating attestation directory: %w", err)
	}
	
	// Initialize key store
	keyStore := NewKeyStore()
	
	// For non-macOS systems that have file-based storage, initialize with storage directory
	// This is handled internally by the platform-specific implementations
	
	var privateKey ed25519.PrivateKey
	var publicKey ed25519.PublicKey
	
	// Check if key exists in secure storage
	if keyStore.HasPrivateKey() {
		// Load from secure storage
		var err error
		privateKey, err = keyStore.RetrievePrivateKey()
		if err != nil {
			return nil, fmt.Errorf("retrieving key from secure storage: %w", err)
		}
		publicKey = privateKey.Public().(ed25519.PublicKey)
	} else {
		// Check for legacy file-based key
		keyFile := filepath.Join(storageDir, "signing.key")
		if keyData, err := os.ReadFile(keyFile); err == nil {
			// Migrate from file to secure storage
			privateKey = ed25519.PrivateKey(keyData)
			publicKey = privateKey.Public().(ed25519.PublicKey)
			
			fmt.Fprintln(os.Stderr, "Migrating signing key to secure storage...")
			if err := keyStore.MigrateFromFile(privateKey); err != nil {
				return nil, fmt.Errorf("migrating key to secure storage: %w", err)
			}
			
			// Remove old files after successful migration
			os.Remove(keyFile)
			os.Remove(filepath.Join(storageDir, "signing.pub"))
			fmt.Fprintln(os.Stderr, "✓ Key migrated to secure storage")
		} else {
			// Generate new key pair
			publicKey, privateKey, err = ed25519.GenerateKey(rand.Reader)
			if err != nil {
				return nil, fmt.Errorf("generating key pair: %w", err)
			}
			
			// Store in secure storage
			if err := keyStore.StorePrivateKey(privateKey); err != nil {
				return nil, fmt.Errorf("storing key in secure storage: %w", err)
			}
			fmt.Fprintln(os.Stderr, "✓ New signing key stored securely")
		}
	}
	
	return &AttestationService{
		privateKey: privateKey,
		publicKey:  publicKey,
		chainFile:  filepath.Join(storageDir, "chain.jsonl"),
		storageDir: storageDir,
		keyStore:   keyStore,
	}, nil
}

// AttestRun creates a cryptographic attestation for a prompt run
func (s *AttestationService) AttestRun(input RunInput, output RunOutput) (*RunAttestation, error) {
	// Generate unique ID
	id := generateID()
	
	// Get previous hash for chaining
	previousHash, err := s.getLatestHash()
	if err != nil {
		return nil, fmt.Errorf("getting previous hash: %w", err)
	}
	
	// Create attestation
	attestation := &RunAttestation{
		ID:           id,
		Timestamp:    time.Now().UTC(),
		Version:      "1.0",
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
	
	// Compute hashes
	attestation.InputHash = s.hashInput(input)
	attestation.OutputHash = s.hashOutput(output)
	
	// Sign the attestation
	signature, err := s.sign(attestation)
	if err != nil {
		return nil, fmt.Errorf("signing attestation: %w", err)
	}
	attestation.Signature = base64.StdEncoding.EncodeToString(signature)
	
	// Append to chain
	if err := s.appendToChain(attestation); err != nil {
		return nil, fmt.Errorf("appending to chain: %w", err)
	}
	
	return attestation, nil
}

// Verify verifies a single attestation
func (s *AttestationService) Verify(attestation *RunAttestation) error {
	// Decode public key
	publicKey, err := base64.StdEncoding.DecodeString(attestation.PublicKey)
	if err != nil {
		return fmt.Errorf("decoding public key: %w", err)
	}
	
	// Decode signature
	signature, err := base64.StdEncoding.DecodeString(attestation.Signature)
	if err != nil {
		return fmt.Errorf("decoding signature: %w", err)
	}
	
	// Verify signature
	message := s.getSigningMessage(attestation)
	if !ed25519.Verify(ed25519.PublicKey(publicKey), message, signature) {
		return fmt.Errorf("invalid signature")
	}
	
	// Verify hashes
	input := RunInput{
		Prompt:       attestation.Prompt,
		Variables:    attestation.Variables,
		Provider:     attestation.Provider,
		Model:        attestation.Model,
		Temperature:  attestation.Temperature,
		SystemPrompt: attestation.SystemPrompt,
	}
	
	if computedHash := s.hashInput(input); computedHash != attestation.InputHash {
		return fmt.Errorf("input hash mismatch")
	}
	
	output := RunOutput{
		Response:         attestation.Response,
		PromptTokens:     attestation.TokensUsed.Prompt,
		CompletionTokens: attestation.TokensUsed.Completion,
		TotalTokens:      attestation.TokensUsed.Total,
		Latency:          attestation.Latency,
		FinishReason:     attestation.FinishReason,
	}
	
	if computedHash := s.hashOutput(output); computedHash != attestation.OutputHash {
		return fmt.Errorf("output hash mismatch")
	}
	
	return nil
}

// VerifyChain verifies the entire attestation chain
func (s *AttestationService) VerifyChain() error {
	file, err := os.Open(s.chainFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // Empty chain is valid
		}
		return fmt.Errorf("opening chain file: %w", err)
	}
	defer file.Close()
	
	decoder := json.NewDecoder(file)
	var previousHash string
	lineNum := 0
	
	for decoder.More() {
		lineNum++
		var attestation RunAttestation
		if err := decoder.Decode(&attestation); err != nil {
			return fmt.Errorf("decoding attestation at line %d: %w", lineNum, err)
		}
		
		// Verify individual attestation
		if err := s.Verify(&attestation); err != nil {
			return fmt.Errorf("verifying attestation %s at line %d: %w", attestation.ID, lineNum, err)
		}
		
		// Verify chain linkage
		if attestation.PreviousHash != previousHash {
			return fmt.Errorf("broken chain at attestation %s: expected previous hash %s, got %s",
				attestation.ID, previousHash, attestation.PreviousHash)
		}
		
		// Update previous hash for next iteration
		previousHash = s.computeAttestationHash(&attestation)
	}
	
	return nil
}

// GetPublicKeyString returns the public key as a base64 string
func (s *AttestationService) GetPublicKeyString() string {
	return base64.StdEncoding.EncodeToString(s.publicKey)
}

// GenerateNewKeyPair generates a new key pair
func (s *AttestationService) GenerateNewKeyPair() error {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return fmt.Errorf("generating key pair: %w", err)
	}
	
	// Store in secure storage
	if err := s.keyStore.StorePrivateKey(privateKey); err != nil {
		return fmt.Errorf("storing key in secure storage: %w", err)
	}
	
	// Update service keys
	s.privateKey = privateKey
	s.publicKey = publicKey
	
	fmt.Fprintln(os.Stderr, "✓ New key pair generated and stored securely")
	return nil
}

// Helper methods

func (s *AttestationService) hashInput(input RunInput) string {
	// Canonicalize input
	data, _ := json.Marshal(map[string]interface{}{
		"prompt":        input.Prompt,
		"variables":     input.Variables,
		"provider":      input.Provider,
		"model":         input.Model,
		"temperature":   input.Temperature,
		"system_prompt": input.SystemPrompt,
	})
	
	hash := sha256.Sum256(data)
	return base64.StdEncoding.EncodeToString(hash[:])
}

func (s *AttestationService) hashOutput(output RunOutput) string {
	// Canonicalize output
	data, _ := json.Marshal(map[string]interface{}{
		"response":          output.Response,
		"prompt_tokens":     output.PromptTokens,
		"completion_tokens": output.CompletionTokens,
		"total_tokens":      output.TotalTokens,
		"latency":           output.Latency.Nanoseconds(),
		"finish_reason":     output.FinishReason,
	})
	
	hash := sha256.Sum256(data)
	return base64.StdEncoding.EncodeToString(hash[:])
}

func (s *AttestationService) sign(attestation *RunAttestation) ([]byte, error) {
	message := s.getSigningMessage(attestation)
	signature, err := s.privateKey.Sign(rand.Reader, message, crypto.Hash(0))
	if err != nil {
		return nil, err
	}
	return signature, nil
}

func (s *AttestationService) getSigningMessage(attestation *RunAttestation) []byte {
	// Create canonical message for signing
	data, _ := json.Marshal(map[string]interface{}{
		"id":            attestation.ID,
		"timestamp":     attestation.Timestamp.Unix(),
		"input_hash":    attestation.InputHash,
		"output_hash":   attestation.OutputHash,
		"previous_hash": attestation.PreviousHash,
	})
	return data
}

func (s *AttestationService) getLatestHash() (string, error) {
	// Read the last line of the chain file
	file, err := os.Open(s.chainFile)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil // Genesis attestation
		}
		return "", err
	}
	defer file.Close()
	
	var lastAttestation *RunAttestation
	decoder := json.NewDecoder(file)
	
	for decoder.More() {
		var attestation RunAttestation
		if err := decoder.Decode(&attestation); err != nil {
			continue
		}
		lastAttestation = &attestation
	}
	
	if lastAttestation == nil {
		return "", nil // Empty chain
	}
	
	return s.computeAttestationHash(lastAttestation), nil
}

func (s *AttestationService) computeAttestationHash(attestation *RunAttestation) string {
	data, _ := json.Marshal(attestation)
	hash := sha256.Sum256(data)
	return base64.StdEncoding.EncodeToString(hash[:])
}

func (s *AttestationService) appendToChain(attestation *RunAttestation) error {
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(s.chainFile), 0700); err != nil {
		return err
	}
	
	// Open file for appending
	file, err := os.OpenFile(s.chainFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer file.Close()
	
	// Write attestation as single line
	encoder := json.NewEncoder(file)
	encoder.SetEscapeHTML(false)
	return encoder.Encode(attestation)
}

func generateID() string {
	// Generate a random ID
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}