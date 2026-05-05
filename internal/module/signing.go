package module

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

// Signature records an Ed25519 signature over module identity and checksum.
type Signature struct {
	Module      string `json:"module"`
	Version     string `json:"version"`
	Checksum    string `json:"checksum"`
	Algorithm   string `json:"algorithm"`
	KeyID       string `json:"key_id"`
	Signature   string `json:"signature"`
	Fingerprint string `json:"fingerprint"`
}

// TrustStore records trusted public-key fingerprints by module path.
type TrustStore struct {
	keys map[string]map[string]bool
}

// NewTrustStore creates an empty module trust store.
func NewTrustStore() *TrustStore {
	return &TrustStore{keys: make(map[string]map[string]bool)}
}

// Add trusts a public key for a module path.
func (s *TrustStore) Add(module string, publicKey ed25519.PublicKey) {
	if s.keys[module] == nil {
		s.keys[module] = make(map[string]bool)
	}
	s.keys[module][PublicKeyFingerprint(publicKey)] = true
}

// Trusts reports whether a public key is trusted for a module path.
func (s *TrustStore) Trusts(module string, publicKey ed25519.PublicKey) bool {
	if s == nil {
		return false
	}
	return s.keys[module][PublicKeyFingerprint(publicKey)]
}

// PublicKeyFingerprint returns the SHA-256 fingerprint of an Ed25519 public key.
func PublicKeyFingerprint(publicKey ed25519.PublicKey) string {
	sum := sha256.Sum256(publicKey)
	return hex.EncodeToString(sum[:])
}

// SignModule signs a module checksum with an Ed25519 private key.
func SignModule(module *Module, privateKey ed25519.PrivateKey) (*Signature, error) {
	if module == nil {
		return nil, fmt.Errorf("module is nil")
	}
	if module.Name == "" || module.Version == "" || module.Checksum == "" {
		return nil, fmt.Errorf("module name, version, and checksum are required")
	}
	publicKey, ok := privateKey.Public().(ed25519.PublicKey)
	if !ok {
		return nil, fmt.Errorf("invalid Ed25519 private key")
	}
	msg := moduleSigningMessage(module.Name, module.Version, module.Checksum)
	sig := ed25519.Sign(privateKey, msg)
	fp := PublicKeyFingerprint(publicKey)
	return &Signature{
		Module:      module.Name,
		Version:     module.Version,
		Checksum:    module.Checksum,
		Algorithm:   "ed25519",
		KeyID:       fp,
		Signature:   hex.EncodeToString(sig),
		Fingerprint: fp,
	}, nil
}

// VerifyModuleSignature verifies a module signature against a public key and trust store.
func VerifyModuleSignature(module *Module, sig *Signature, publicKey ed25519.PublicKey, trust *TrustStore) error {
	if module == nil {
		return fmt.Errorf("module is nil")
	}
	if sig == nil {
		return fmt.Errorf("signature is nil")
	}
	if sig.Algorithm != "ed25519" {
		return fmt.Errorf("unsupported signature algorithm %s", sig.Algorithm)
	}
	if sig.Module != module.Name || sig.Version != module.Version || sig.Checksum != module.Checksum {
		return fmt.Errorf("signature does not match module metadata")
	}
	fp := PublicKeyFingerprint(publicKey)
	if sig.Fingerprint != "" && sig.Fingerprint != fp {
		return fmt.Errorf("signature fingerprint does not match public key")
	}
	if sig.KeyID != "" && sig.KeyID != fp {
		return fmt.Errorf("signature key id does not match public key")
	}
	if !trust.Trusts(module.Name, publicKey) {
		return fmt.Errorf("untrusted signing key for module %s", module.Name)
	}
	raw, err := hex.DecodeString(sig.Signature)
	if err != nil {
		return fmt.Errorf("decode signature: %w", err)
	}
	if !ed25519.Verify(publicKey, moduleSigningMessage(module.Name, module.Version, module.Checksum), raw) {
		return fmt.Errorf("invalid module signature")
	}
	return nil
}

func moduleSigningMessage(name, version, checksum string) []byte {
	return []byte(strings.Join([]string{"pe module signature v1", name, version, checksum}, "\n"))
}
