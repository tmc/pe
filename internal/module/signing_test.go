package module

import (
	"crypto/ed25519"
	"testing"
)

func TestSignAndVerifyModule(t *testing.T) {
	publicKey, privateKey, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatal(err)
	}
	module := &Module{Name: "example.com/mod", Version: "v1.0.0", Checksum: "abc123"}
	sig, err := SignModule(module, privateKey)
	if err != nil {
		t.Fatalf("SignModule: %v", err)
	}
	trust := NewTrustStore()
	trust.Add(module.Name, publicKey)
	if err := VerifyModuleSignature(module, sig, publicKey, trust); err != nil {
		t.Fatalf("VerifyModuleSignature: %v", err)
	}
}

func TestVerifyModuleSignatureRejectsTamper(t *testing.T) {
	publicKey, privateKey, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatal(err)
	}
	module := &Module{Name: "example.com/mod", Version: "v1.0.0", Checksum: "abc123"}
	sig, err := SignModule(module, privateKey)
	if err != nil {
		t.Fatalf("SignModule: %v", err)
	}
	trust := NewTrustStore()
	trust.Add(module.Name, publicKey)
	tampered := &Module{Name: module.Name, Version: module.Version, Checksum: "changed"}
	if err := VerifyModuleSignature(tampered, sig, publicKey, trust); err == nil {
		t.Fatal("VerifyModuleSignature accepted tampered module")
	}
}

func TestVerifyModuleSignatureRejectsUntrustedKey(t *testing.T) {
	publicKey, privateKey, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatal(err)
	}
	module := &Module{Name: "example.com/mod", Version: "v1.0.0", Checksum: "abc123"}
	sig, err := SignModule(module, privateKey)
	if err != nil {
		t.Fatalf("SignModule: %v", err)
	}
	if err := VerifyModuleSignature(module, sig, publicKey, NewTrustStore()); err == nil {
		t.Fatal("VerifyModuleSignature accepted untrusted key")
	}
}

func TestTrustStoreFingerprint(t *testing.T) {
	publicKey, _, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatal(err)
	}
	trust := NewTrustStore()
	trust.Add("example.com/mod", publicKey)
	if !trust.Trusts("example.com/mod", publicKey) {
		t.Fatal("trusted key was not recognized")
	}
	if trust.Trusts("example.com/other", publicKey) {
		t.Fatal("key trusted for wrong module")
	}
	if PublicKeyFingerprint(publicKey) == "" {
		t.Fatal("empty fingerprint")
	}
}
