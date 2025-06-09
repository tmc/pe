//go:build !darwin
// +build !darwin

package attestation

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/crypto/pbkdf2"
	"golang.org/x/term"
)

// KeyStore provides encrypted file-based key storage for non-macOS systems
type KeyStore struct {
	keyPath string
}

// NewKeyStore creates a new file-based key store
func NewKeyStore() *KeyStore {
	// Get data directory from environment or default
	dataDir := os.Getenv("PE_DATA_DIR")
	if dataDir == "" {
		dataDir = ".pe"
	}

	return &KeyStore{
		keyPath: filepath.Join(dataDir, "attestations", "signing.key.enc"),
	}
}

// StorePrivateKey stores the private key in an encrypted file
func (ks *KeyStore) StorePrivateKey(privateKey ed25519.PrivateKey) error {
	// Get passphrase from user
	passphrase, err := ks.getPassphrase("Enter passphrase to encrypt signing key: ")
	if err != nil {
		return fmt.Errorf("getting passphrase: %w", err)
	}

	// Encrypt the key
	encryptedKey, err := ks.encryptKey(privateKey, passphrase)
	if err != nil {
		return fmt.Errorf("encrypting key: %w", err)
	}

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(ks.keyPath), 0700); err != nil {
		return fmt.Errorf("creating directory: %w", err)
	}

	// Write encrypted key with secure permissions
	if err := os.WriteFile(ks.keyPath, encryptedKey, 0400); err != nil {
		return fmt.Errorf("writing encrypted key: %w", err)
	}

	return nil
}

// RetrievePrivateKey retrieves and decrypts the private key
func (ks *KeyStore) RetrievePrivateKey() (ed25519.PrivateKey, error) {
	// Read encrypted key
	encryptedKey, err := os.ReadFile(ks.keyPath)
	if err != nil {
		return nil, fmt.Errorf("reading encrypted key: %w", err)
	}

	// Get passphrase from user
	passphrase, err := ks.getPassphrase("Enter passphrase to decrypt signing key: ")
	if err != nil {
		return nil, fmt.Errorf("getting passphrase: %w", err)
	}

	// Decrypt the key
	privateKey, err := ks.decryptKey(encryptedKey, passphrase)
	if err != nil {
		return nil, fmt.Errorf("decrypting key: %w", err)
	}

	return privateKey, nil
}

// DeletePrivateKey removes the encrypted key file
func (ks *KeyStore) DeletePrivateKey() error {
	err := os.Remove(ks.keyPath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// HasPrivateKey checks if an encrypted key file exists
func (ks *KeyStore) HasPrivateKey() bool {
	_, err := os.Stat(ks.keyPath)
	return err == nil
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

// MigrateFromFile migrates from unencrypted to encrypted storage
func (ks *KeyStore) MigrateFromFile(privateKey ed25519.PrivateKey) error {
	return ks.StorePrivateKey(privateKey)
}

// encryptKey encrypts the private key using AES-GCM with PBKDF2
func (ks *KeyStore) encryptKey(privateKey []byte, passphrase string) ([]byte, error) {
	// Generate salt
	salt := make([]byte, 32)
	if _, err := rand.Read(salt); err != nil {
		return nil, err
	}

	// Derive key using PBKDF2
	key := pbkdf2.Key([]byte(passphrase), salt, 100000, 32, sha256.New)

	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// Create GCM
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Generate nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}

	// Encrypt
	ciphertext := gcm.Seal(nonce, nonce, privateKey, nil)

	// Return salt + ciphertext
	return append(salt, ciphertext...), nil
}

// decryptKey decrypts the private key
func (ks *KeyStore) decryptKey(encryptedData []byte, passphrase string) (ed25519.PrivateKey, error) {
	if len(encryptedData) < 32 {
		return nil, fmt.Errorf("invalid encrypted data")
	}

	// Extract salt
	salt := encryptedData[:32]
	ciphertext := encryptedData[32:]

	// Derive key using PBKDF2
	key := pbkdf2.Key([]byte(passphrase), salt, 100000, 32, sha256.New)

	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// Create GCM
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Extract nonce
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	// Decrypt
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	if len(plaintext) != ed25519.PrivateKeySize {
		return nil, fmt.Errorf("invalid key size after decryption")
	}

	return ed25519.PrivateKey(plaintext), nil
}

// getPassphrase prompts for a passphrase
func (ks *KeyStore) getPassphrase(prompt string) (string, error) {
	fmt.Fprint(os.Stderr, prompt)
	bytePassword, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Fprintln(os.Stderr) // New line after password input
	if err != nil {
		return "", err
	}
	return string(bytePassword), nil
}
