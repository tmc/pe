//go:build darwin
// +build darwin

package attestation

import (
	"crypto/ed25519"
	"encoding/base64"
	"fmt"

	"github.com/keybase/go-keychain"
)

const (
	keychainService = "com.github.tmc.pe"
	keychainAccount = "attestation-signing-key"
	keychainLabel   = "PE Attestation Signing Key"
)

// KeyStore provides secure key storage using macOS Keychain
type KeyStore struct{}

// NewKeyStore creates a new macOS Keychain-based key store
func NewKeyStore() *KeyStore {
	return &KeyStore{}
}

// StorePrivateKey securely stores the Ed25519 private key in macOS Keychain
func (ks *KeyStore) StorePrivateKey(privateKey ed25519.PrivateKey) error {
	// First, try to delete any existing key
	ks.DeletePrivateKey()

	// Create keychain item
	item := keychain.NewItem()
	item.SetSecClass(keychain.SecClassGenericPassword)
	item.SetService(keychainService)
	item.SetAccount(keychainAccount)
	item.SetLabel(keychainLabel)
	item.SetData(privateKey)
	item.SetAccessible(keychain.AccessibleWhenUnlocked)
	item.SetSynchronizable(keychain.SynchronizableNo)

	// Add to keychain
	err := keychain.AddItem(item)
	if err == keychain.ErrorDuplicateItem {
		// Update existing item
		queryItem := keychain.NewItem()
		queryItem.SetSecClass(keychain.SecClassGenericPassword)
		queryItem.SetService(keychainService)
		queryItem.SetAccount(keychainAccount)

		updateItem := keychain.NewItem()
		updateItem.SetData(privateKey)

		return keychain.UpdateItem(queryItem, updateItem)
	}

	return err
}

// RetrievePrivateKey retrieves the Ed25519 private key from macOS Keychain
func (ks *KeyStore) RetrievePrivateKey() (ed25519.PrivateKey, error) {
	query := keychain.NewItem()
	query.SetSecClass(keychain.SecClassGenericPassword)
	query.SetService(keychainService)
	query.SetAccount(keychainAccount)
	query.SetMatchLimit(keychain.MatchLimitOne)
	query.SetReturnData(true)

	results, err := keychain.QueryItem(query)
	if err != nil {
		return nil, fmt.Errorf("querying keychain: %w", err)
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("no signing key found in keychain")
	}

	keyData := results[0].Data
	if len(keyData) != ed25519.PrivateKeySize {
		return nil, fmt.Errorf("invalid key size: expected %d, got %d", ed25519.PrivateKeySize, len(keyData))
	}

	return ed25519.PrivateKey(keyData), nil
}

// DeletePrivateKey removes the private key from keychain
func (ks *KeyStore) DeletePrivateKey() error {
	item := keychain.NewItem()
	item.SetSecClass(keychain.SecClassGenericPassword)
	item.SetService(keychainService)
	item.SetAccount(keychainAccount)

	err := keychain.DeleteItem(item)
	if err == keychain.ErrorItemNotFound {
		return nil // Not an error if key doesn't exist
	}
	return err
}

// HasPrivateKey checks if a private key exists in keychain
func (ks *KeyStore) HasPrivateKey() bool {
	query := keychain.NewItem()
	query.SetSecClass(keychain.SecClassGenericPassword)
	query.SetService(keychainService)
	query.SetAccount(keychainAccount)
	query.SetMatchLimit(keychain.MatchLimitOne)
	query.SetReturnData(true) // Add this to match RetrievePrivateKey

	results, err := keychain.QueryItem(query)
	return err == nil && len(results) > 0
}

// GetPublicKeyString retrieves the public key as a base64 string
func (ks *KeyStore) GetPublicKeyString() (string, error) {
	privateKey, err := ks.RetrievePrivateKey()
	if err != nil {
		return "", err
	}

	publicKey := privateKey.Public().(ed25519.PublicKey)
	return base64.StdEncoding.EncodeToString(publicKey), nil
}

// MigrateFromFile attempts to migrate a key from file storage to keychain
func (ks *KeyStore) MigrateFromFile(privateKey ed25519.PrivateKey) error {
	// Store in keychain
	if err := ks.StorePrivateKey(privateKey); err != nil {
		return fmt.Errorf("storing key in keychain: %w", err)
	}

	return nil
}
