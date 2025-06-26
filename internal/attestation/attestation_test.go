package attestation

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAttestationService(t *testing.T) {
	tmpDir := t.TempDir()
	
	service, err := NewAttestationService(tmpDir)
	require.NoError(t, err)
	require.NotNil(t, service)

	// Check that directory was created
	attestDir := filepath.Join(tmpDir, "attestations")
	info, err := os.Stat(attestDir)
	require.NoError(t, err)
	assert.True(t, info.IsDir())
	assert.Equal(t, os.FileMode(0700), info.Mode().Perm())
}

func TestAttestRun(t *testing.T) {
	tmpDir := t.TempDir()
	service, err := NewAttestationService(tmpDir)
	require.NoError(t, err)

	input := RunInput{
		Prompt:      "test prompt",
		Variables:   map[string]interface{}{"key": "value"},
		Provider:    "openai",
		Model:       "gpt-4",
		Temperature: 0.7,
	}

	output := RunOutput{
		Response:         "test response",
		PromptTokens:     10,
		CompletionTokens: 20,
		TotalTokens:      30,
		Latency:          time.Second,
		FinishReason:     "stop",
	}

	attestation, err := service.AttestRun(input, output)
	require.NoError(t, err)
	require.NotNil(t, attestation)

	// Verify basic fields
	assert.NotEmpty(t, attestation.ID)
	assert.NotZero(t, attestation.Timestamp)
	assert.Equal(t, input.Prompt, attestation.Prompt)
	assert.Equal(t, output.Response, attestation.Response)
	assert.NotEmpty(t, attestation.InputHash)
	assert.NotEmpty(t, attestation.OutputHash)
	assert.NotEmpty(t, attestation.Signature)
	assert.NotEmpty(t, attestation.PublicKey)

	// Verify token usage
	assert.Equal(t, output.PromptTokens, attestation.TokensUsed.Prompt)
	assert.Equal(t, output.CompletionTokens, attestation.TokensUsed.Completion)
	assert.Equal(t, output.TotalTokens, attestation.TokensUsed.Total)
}

func TestVerifyAttestation(t *testing.T) {
	tmpDir := t.TempDir()
	service, err := NewAttestationService(tmpDir)
	require.NoError(t, err)

	input := RunInput{
		Prompt:   "test prompt",
		Provider: "openai",
		Model:    "gpt-4",
	}

	output := RunOutput{
		Response:    "test response",
		TotalTokens: 30,
	}

	// Create attestation
	attestation, err := service.AttestRun(input, output)
	require.NoError(t, err)

	// Verify should pass
	err = service.Verify(attestation)
	require.NoError(t, err)

	// Tamper with signature - should fail
	tamperedAttestation := *attestation
	tamperedAttestation.Signature = "invalid_signature"
	err = service.Verify(&tamperedAttestation)
	assert.Error(t, err)
}

func TestVerifyChain(t *testing.T) {
	tmpDir := t.TempDir()
	service, err := NewAttestationService(tmpDir)
	require.NoError(t, err)

	// Create first attestation
	input1 := RunInput{Prompt: "first prompt", Provider: "openai", Model: "gpt-4"}
	output1 := RunOutput{Response: "first response", TotalTokens: 10}
	_, err = service.AttestRun(input1, output1)
	require.NoError(t, err)

	// Create second attestation (should chain to first)
	input2 := RunInput{Prompt: "second prompt", Provider: "openai", Model: "gpt-4"}
	output2 := RunOutput{Response: "second response", TotalTokens: 20}
	attestation2, err := service.AttestRun(input2, output2)
	require.NoError(t, err)

	// Verify chain
	err = service.VerifyChain()
	require.NoError(t, err)

	// Verify that second attestation has a previous hash
	assert.NotEmpty(t, attestation2.PreviousHash)
}

func TestGetPublicKeyString(t *testing.T) {
	tmpDir := t.TempDir()
	service, err := NewAttestationService(tmpDir)
	require.NoError(t, err)

	pubKeyStr := service.GetPublicKeyString()
	assert.NotEmpty(t, pubKeyStr)

	// Should be base64 encoded
	assert.Greater(t, len(pubKeyStr), 40) // Ed25519 public keys are 32 bytes, base64 encoded ~44 chars
}

func TestGenerateNewKeyPair(t *testing.T) {
	tmpDir := t.TempDir()
	service, err := NewAttestationService(tmpDir)
	require.NoError(t, err)

	oldKey := service.GetPublicKeyString()

	// Generate new key pair
	err = service.GenerateNewKeyPair()
	require.NoError(t, err)

	newKey := service.GetPublicKeyString()
	assert.NotEqual(t, oldKey, newKey)
}

func TestHashConsistency(t *testing.T) {
	tmpDir := t.TempDir()
	service, err := NewAttestationService(tmpDir)
	require.NoError(t, err)

	input := RunInput{
		Prompt:       "test prompt",
		Variables:    map[string]interface{}{"var1": "value1"},
		Provider:     "openai",
		Model:        "gpt-4",
		Temperature:  0.7,
		SystemPrompt: "system",
	}

	// Same input should produce same hash
	hash1 := service.hashInput(input)
	hash2 := service.hashInput(input)
	assert.Equal(t, hash1, hash2)
	assert.NotEmpty(t, hash1)

	// Different input should produce different hash
	input.Prompt = "different prompt"
	hash3 := service.hashInput(input)
	assert.NotEqual(t, hash1, hash3)
}

func TestOutputHashConsistency(t *testing.T) {
	tmpDir := t.TempDir()
	service, err := NewAttestationService(tmpDir)
	require.NoError(t, err)

	output := RunOutput{
		Response:         "test response",
		PromptTokens:     10,
		CompletionTokens: 20,
		TotalTokens:      30,
		Latency:          time.Second,
		FinishReason:     "stop",
	}

	// Same output should produce same hash
	hash1 := service.hashOutput(output)
	hash2 := service.hashOutput(output)
	assert.Equal(t, hash1, hash2)
	assert.NotEmpty(t, hash1)

	// Different output should produce different hash
	output.Response = "different response"
	hash3 := service.hashOutput(output)
	assert.NotEqual(t, hash1, hash3)
}

func TestSignatureVerification(t *testing.T) {
	// Test direct signature verification
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	message := []byte("test message")
	signature := ed25519.Sign(priv, message)

	// Valid signature should verify
	valid := ed25519.Verify(pub, message, signature)
	assert.True(t, valid)

	// Invalid signature should not verify
	invalidSig := make([]byte, len(signature))
	copy(invalidSig, signature)
	invalidSig[0] ^= 1 // Flip a bit

	valid = ed25519.Verify(pub, message, invalidSig)
	assert.False(t, valid)
}

func TestAttestationSerialization(t *testing.T) {
	tmpDir := t.TempDir()
	service, err := NewAttestationService(tmpDir)
	require.NoError(t, err)

	input := RunInput{Prompt: "test", Provider: "openai", Model: "gpt-4"}
	output := RunOutput{Response: "response", TotalTokens: 10}
	
	original, err := service.AttestRun(input, output)
	require.NoError(t, err)

	// Serialize to JSON
	data, err := json.Marshal(original)
	require.NoError(t, err)

	// Deserialize
	var deserialized RunAttestation
	err = json.Unmarshal(data, &deserialized)
	require.NoError(t, err)

	// Compare key fields
	assert.Equal(t, original.ID, deserialized.ID)
	assert.Equal(t, original.Prompt, deserialized.Prompt)
	assert.Equal(t, original.Response, deserialized.Response)
	assert.Equal(t, original.InputHash, deserialized.InputHash)
	assert.Equal(t, original.OutputHash, deserialized.OutputHash)
	assert.Equal(t, original.Signature, deserialized.Signature)
	assert.Equal(t, original.PublicKey, deserialized.PublicKey)
}

func TestErrorCases(t *testing.T) {
	t.Run("invalid directory", func(t *testing.T) {
		// Try to create service in non-writable location
		service, err := NewAttestationService("/invalid/path/that/does/not/exist")
		assert.Error(t, err)
		assert.Nil(t, service)
	})

	t.Run("verify invalid attestation", func(t *testing.T) {
		tmpDir := t.TempDir()
		service, err := NewAttestationService(tmpDir)
		require.NoError(t, err)

		// Create invalid attestation
		invalidAttestation := &RunAttestation{
			ID:        "test",
			Signature: "invalid",
			PublicKey: "invalid",
		}

		err = service.Verify(invalidAttestation)
		assert.Error(t, err)
	})
}

func TestMultipleAttestations(t *testing.T) {
	tmpDir := t.TempDir()
	service, err := NewAttestationService(tmpDir)
	require.NoError(t, err)

	// Create multiple attestations
	attestations := make([]*RunAttestation, 3)
	for i := 0; i < 3; i++ {
		input := RunInput{
			Prompt:   fmt.Sprintf("prompt %d", i),
			Provider: "openai",
			Model:    "gpt-4",
		}
		output := RunOutput{
			Response:    fmt.Sprintf("response %d", i),
			TotalTokens: 10 + i,
		}
		attestations[i], err = service.AttestRun(input, output)
		require.NoError(t, err)
	}

	// Verify all attestations
	for i, attestation := range attestations {
		err := service.Verify(attestation)
		require.NoError(t, err, "attestation %d should verify", i)
	}

	// Verify chain
	err = service.VerifyChain()
	require.NoError(t, err)
}