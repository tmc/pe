//go:build darwin
// +build darwin

package attestation

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKeyStore_StoreAndRetrieve(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping macOS Keychain integration test in short mode")
	}

	keyStore := NewKeyStore()
	
	// Clean up any existing test key
	keyStore.DeletePrivateKey()
	
	// Generate test key
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	// Test storing key
	err = keyStore.StorePrivateKey(privateKey)
	require.NoError(t, err)

	// Test checking if key exists
	assert.True(t, keyStore.HasPrivateKey())

	// Test retrieving key
	retrievedKey, err := keyStore.RetrievePrivateKey()
	require.NoError(t, err)
	assert.Equal(t, privateKey, retrievedKey)

	// Test getting public key string
	publicKeyString, err := keyStore.GetPublicKeyString()
	require.NoError(t, err)

	decodedPublicKey, err := base64.StdEncoding.DecodeString(publicKeyString)
	require.NoError(t, err)
	assert.Equal(t, publicKey, ed25519.PublicKey(decodedPublicKey))

	// Test updating existing key
	_, newPrivateKey, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	err = keyStore.StorePrivateKey(newPrivateKey)
	require.NoError(t, err)

	retrievedNewKey, err := keyStore.RetrievePrivateKey()
	require.NoError(t, err)
	assert.Equal(t, newPrivateKey, retrievedNewKey)

	// Test deleting key
	err = keyStore.DeletePrivateKey()
	require.NoError(t, err)

	assert.False(t, keyStore.HasPrivateKey())

	// Test retrieving after deletion should fail
	_, err = keyStore.RetrievePrivateKey()
	assert.Error(t, err)

	// Test deleting non-existent key (should not error)
	err = keyStore.DeletePrivateKey()
	assert.NoError(t, err)
}

func TestKeyStore_MigrateFromFile(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping macOS Keychain integration test in short mode")
	}

	keyStore := NewKeyStore()
	
	// Clean up any existing test key
	keyStore.DeletePrivateKey()

	// Generate test key
	_, privateKey, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	// Test migration
	err = keyStore.MigrateFromFile(privateKey)
	require.NoError(t, err)

	// Verify key was stored
	assert.True(t, keyStore.HasPrivateKey())

	retrievedKey, err := keyStore.RetrievePrivateKey()
	require.NoError(t, err)
	assert.Equal(t, privateKey, retrievedKey)

	// Clean up
	keyStore.DeletePrivateKey()
}

func TestKeyStore_ErrorCases(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping macOS Keychain integration test in short mode")
	}

	keyStore := NewKeyStore()
	
	// Clean up any existing test key
	keyStore.DeletePrivateKey()

	// Test retrieving non-existent key
	_, err := keyStore.RetrievePrivateKey()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no signing key found in keychain")

	// Test getting public key string when no key exists
	_, err = keyStore.GetPublicKeyString()
	assert.Error(t, err)

	// Test invalid key size (this is harder to test without mocking keychain)
	// Would require modifying the keychain directly, which is not recommended
}

func TestKeyStore_Constants(t *testing.T) {
	// Test that constants are set correctly
	assert.Equal(t, "com.github.tmc.pe", keychainService)
	assert.Equal(t, "attestation-signing-key", keychainAccount)
	assert.Equal(t, "PE Attestation Signing Key", keychainLabel)
}