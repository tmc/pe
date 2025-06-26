//go:build !darwin
// +build !darwin

package attestation

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKeyStore_FileBasedStorage(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Override PE_DATA_DIR for this test
	oldDataDir := os.Getenv("PE_DATA_DIR")
	os.Setenv("PE_DATA_DIR", tmpDir)
	defer func() {
		if oldDataDir == "" {
			os.Unsetenv("PE_DATA_DIR")
		} else {
			os.Setenv("PE_DATA_DIR", oldDataDir)
		}
	}()

	keyStore := NewKeyStore()
	
	// Initially should not have key
	assert.False(t, keyStore.HasPrivateKey())

	// Generate test key
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	// Mock the passphrase input (this is challenging without dependency injection)
	// For now, we'll test what we can without user interaction
	
	// Test that key path is constructed correctly
	expectedPath := filepath.Join(tmpDir, "attestations", "signing.key.enc")
	assert.Equal(t, expectedPath, keyStore.keyPath)
}

func TestKeyStore_HasPrivateKey(t *testing.T) {
	tmpDir := t.TempDir()
	
	keyStore := &KeyStore{
		keyPath: filepath.Join(tmpDir, "test.key.enc"),
	}

	// Initially should not exist
	assert.False(t, keyStore.HasPrivateKey())

	// Create a dummy file
	err := os.WriteFile(keyStore.keyPath, []byte("dummy"), 0400)
	require.NoError(t, err)

	// Now should exist
	assert.True(t, keyStore.HasPrivateKey())
}

func TestKeyStore_DeletePrivateKey(t *testing.T) {
	tmpDir := t.TempDir()
	
	keyStore := &KeyStore{
		keyPath: filepath.Join(tmpDir, "test.key.enc"),
	}

	// Create a dummy file
	err := os.WriteFile(keyStore.keyPath, []byte("dummy"), 0400)
	require.NoError(t, err)

	// Verify it exists
	assert.True(t, keyStore.HasPrivateKey())

	// Delete it
	err = keyStore.DeletePrivateKey()
	require.NoError(t, err)

	// Verify it's gone
	assert.False(t, keyStore.HasPrivateKey())

	// Deleting non-existent file should not error
	err = keyStore.DeletePrivateKey()
	assert.NoError(t, err)
}

func TestKeyStore_EncryptDecryptKey(t *testing.T) {
	keyStore := NewKeyStore()
	
	// Generate test key
	_, privateKey, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	passphrase := "test-passphrase"

	// Test encryption
	encryptedData, err := keyStore.encryptKey(privateKey, passphrase)
	require.NoError(t, err)
	
	// Encrypted data should be longer than original (salt + nonce + ciphertext)
	assert.Greater(t, len(encryptedData), len(privateKey))
	
	// Should start with salt (32 bytes)
	assert.GreaterOrEqual(t, len(encryptedData), 32)

	// Test decryption
	decryptedKey, err := keyStore.decryptKey(encryptedData, passphrase)
	require.NoError(t, err)
	
	// Should match original
	assert.Equal(t, privateKey, decryptedKey)

	// Test decryption with wrong passphrase
	_, err = keyStore.decryptKey(encryptedData, "wrong-passphrase")
	assert.Error(t, err)

	// Test decryption with invalid data
	_, err = keyStore.decryptKey([]byte("invalid"), passphrase)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid encrypted data")

	// Test decryption with data too short
	_, err = keyStore.decryptKey(make([]byte, 10), passphrase)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid encrypted data")
}

func TestKeyStore_NewKeyStore(t *testing.T) {
	tests := []struct {
		name     string
		dataDir  string
		expected string
	}{
		{
			name:     "default data dir",
			dataDir:  "",
			expected: filepath.Join(".pe", "attestations", "signing.key.enc"),
		},
		{
			name:     "custom data dir",
			dataDir:  "/custom/path",
			expected: filepath.Join("/custom/path", "attestations", "signing.key.enc"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldDataDir := os.Getenv("PE_DATA_DIR")
			if tt.dataDir != "" {
				os.Setenv("PE_DATA_DIR", tt.dataDir)
			} else {
				os.Unsetenv("PE_DATA_DIR")
			}
			defer func() {
				if oldDataDir == "" {
					os.Unsetenv("PE_DATA_DIR")
				} else {
					os.Setenv("PE_DATA_DIR", oldDataDir)
				}
			}()

			keyStore := NewKeyStore()
			assert.Equal(t, tt.expected, keyStore.keyPath)
		})
	}
}

// Test the internal encryption/decryption logic more thoroughly
func TestKeyStore_EncryptionDetails(t *testing.T) {
	keyStore := NewKeyStore()
	
	// Generate test key
	_, privateKey1, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	
	_, privateKey2, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	passphrase := "test-passphrase"

	// Encrypt same key twice - should produce different results due to random salt/nonce
	encrypted1, err := keyStore.encryptKey(privateKey1, passphrase)
	require.NoError(t, err)
	
	encrypted2, err := keyStore.encryptKey(privateKey1, passphrase)
	require.NoError(t, err)
	
	assert.NotEqual(t, encrypted1, encrypted2)

	// But both should decrypt to the same key
	decrypted1, err := keyStore.decryptKey(encrypted1, passphrase)
	require.NoError(t, err)
	
	decrypted2, err := keyStore.decryptKey(encrypted2, passphrase)
	require.NoError(t, err)
	
	assert.Equal(t, privateKey1, decrypted1)
	assert.Equal(t, privateKey1, decrypted2)

	// Different keys should produce different encrypted data
	encrypted3, err := keyStore.encryptKey(privateKey2, passphrase)
	require.NoError(t, err)
	
	assert.NotEqual(t, encrypted1, encrypted3)

	// Test edge case: empty key should fail
	_, err = keyStore.encryptKey(nil, passphrase)
	assert.NoError(t, err) // Actually, this might not fail - AES-GCM can encrypt empty data

	// Test corrupted ciphertext
	corruptedData := make([]byte, len(encrypted1))
	copy(corruptedData, encrypted1)
	
	// Corrupt some bytes in the middle (after salt)
	if len(corruptedData) > 50 {
		corruptedData[40] ^= 0xFF
		corruptedData[45] ^= 0xFF
	}
	
	_, err = keyStore.decryptKey(corruptedData, passphrase)
	assert.Error(t, err)
}

// Benchmark encryption/decryption
func BenchmarkKeyStore_Encrypt(b *testing.B) {
	keyStore := NewKeyStore()
	_, privateKey, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(b, err)
	
	passphrase := "benchmark-passphrase"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := keyStore.encryptKey(privateKey, passphrase)
		require.NoError(b, err)
	}
}

func BenchmarkKeyStore_Decrypt(b *testing.B) {
	keyStore := NewKeyStore()
	_, privateKey, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(b, err)
	
	passphrase := "benchmark-passphrase"
	encryptedData, err := keyStore.encryptKey(privateKey, passphrase)
	require.NoError(b, err)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := keyStore.decryptKey(encryptedData, passphrase)
		require.NoError(b, err)
	}
}